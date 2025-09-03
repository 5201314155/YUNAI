package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
)

// 🔄 YUNAI 模型负载均衡器 - 支持相同模型轮询调用

// ModelLoadBalancer 模型负载均衡器接口
type ModelLoadBalancer interface {
	// 选择最佳实例
	SelectInstance(ctx context.Context, modelID uuid.UUID, sessionID string) (*domain.ModelInstance, error)
	
	// 更新实例状态
	UpdateInstanceHealth(instanceID uuid.UUID, health *domain.InstanceHealth) error
	UpdateInstanceLoad(instanceID uuid.UUID, currentLoad int) error
	
	// 实例管理
	AddInstance(modelID uuid.UUID, instance *domain.ModelInstance) error
	RemoveInstance(modelID uuid.UUID, instanceID uuid.UUID) error
	GetInstances(modelID uuid.UUID) ([]*domain.ModelInstance, error)
	
	// 负载均衡策略
	SetStrategy(modelID uuid.UUID, strategy string) error
	GetStrategy(modelID uuid.UUID) string
	
	// 熔断器管理
	RecordSuccess(instanceID uuid.UUID) error
	RecordFailure(instanceID uuid.UUID, err error) error
	IsCircuitOpen(instanceID uuid.UUID) bool
}

type modelLoadBalancer struct {
	// 模型实例映射 modelID -> instances
	modelInstances map[uuid.UUID][]*domain.ModelInstance
	
	// 负载均衡策略 modelID -> strategy
	strategies map[uuid.UUID]string
	
	// 轮询计数器 modelID -> counter
	roundRobinCounters map[uuid.UUID]*int64
	
	// 会话粘性 sessionID -> instanceID
	stickySessions map[string]uuid.UUID
	
	// 熔断器状态 instanceID -> circuitState
	circuitStates map[uuid.UUID]*CircuitState
	
	// 读写锁
	mu sync.RWMutex
	
	logger *logrus.Logger
}

// CircuitState 熔断器状态
type CircuitState struct {
	State             string    // closed, open, half_open
	FailureCount      int       // 失败次数
	SuccessCount      int       // 成功次数
	LastFailureTime   time.Time // 最后失败时间
	NextRetryTime     time.Time // 下次重试时间
	HalfOpenCallCount int       // 半开状态调用次数
}

// NewModelLoadBalancer 创建模型负载均衡器
func NewModelLoadBalancer(logger *logrus.Logger) ModelLoadBalancer {
	return &modelLoadBalancer{
		modelInstances:     make(map[uuid.UUID][]*domain.ModelInstance),
		strategies:         make(map[uuid.UUID]string),
		roundRobinCounters: make(map[uuid.UUID]*int64),
		stickySessions:     make(map[string]uuid.UUID),
		circuitStates:      make(map[uuid.UUID]*CircuitState),
		logger:             logger,
	}
}

// SelectInstance 选择最佳实例 - 支持多种负载均衡策略
func (lb *modelLoadBalancer) SelectInstance(ctx context.Context, modelID uuid.UUID, sessionID string) (*domain.ModelInstance, error) {
	lb.mu.RLock()
	instances, exists := lb.modelInstances[modelID]
	if !exists || len(instances) == 0 {
		lb.mu.RUnlock()
		return nil, fmt.Errorf("no instances available for model %s", modelID)
	}
	
	strategy := lb.strategies[modelID]
	if strategy == "" {
		strategy = "round_robin" // 默认轮询策略
	}
	lb.mu.RUnlock()
	
	// 过滤健康的实例
	healthyInstances := lb.filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instances available for model %s", modelID)
	}
	
	// 根据策略选择实例
	switch strategy {
	case "round_robin":
		return lb.selectRoundRobin(modelID, healthyInstances), nil
	case "weighted":
		return lb.selectWeighted(healthyInstances), nil
	case "least_connections":
		return lb.selectLeastConnections(healthyInstances), nil
	case "random":
		return lb.selectRandom(healthyInstances), nil
	case "cost_optimized":
		return lb.selectCostOptimized(healthyInstances), nil
	case "response_time":
		return lb.selectFastestResponse(healthyInstances), nil
	case "sticky_session":
		return lb.selectStickySession(sessionID, healthyInstances), nil
	default:
		return lb.selectRoundRobin(modelID, healthyInstances), nil
	}
}

// filterHealthyInstances 过滤健康的实例
func (lb *modelLoadBalancer) filterHealthyInstances(instances []*domain.ModelInstance) []*domain.ModelInstance {
	var healthy []*domain.ModelInstance
	
	for _, instance := range instances {
		// 检查实例状态
		if instance.Status != "active" {
			continue
		}
		
		// 检查熔断器状态
		if lb.IsCircuitOpen(instance.ID) {
			continue
		}
		
		// 检查健康状态
		if instance.Health != nil && instance.Health.Status == "unhealthy" {
			continue
		}
		
		// 检查并发限制
		if instance.MaxConcurrency > 0 && instance.CurrentLoad >= instance.MaxConcurrency {
			continue
		}
		
		healthy = append(healthy, instance)
	}
	
	return healthy
}

// selectRoundRobin 轮询选择 - 经典轮询算法
func (lb *modelLoadBalancer) selectRoundRobin(modelID uuid.UUID, instances []*domain.ModelInstance) *domain.ModelInstance {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	counter, exists := lb.roundRobinCounters[modelID]
	if !exists {
		var c int64 = 0
		counter = &c
		lb.roundRobinCounters[modelID] = counter
	}
	
	index := atomic.AddInt64(counter, 1) % int64(len(instances))
	return instances[index]
}

// selectWeighted 加权选择 - 根据权重分配请求
func (lb *modelLoadBalancer) selectWeighted(instances []*domain.ModelInstance) *domain.ModelInstance {
	totalWeight := 0
	for _, instance := range instances {
		totalWeight += instance.Weight
	}
	
	if totalWeight == 0 {
		// 如果没有权重，使用随机选择
		return instances[rand.Intn(len(instances))]
	}
	
	randomWeight := rand.Intn(totalWeight)
	currentWeight := 0
	
	for _, instance := range instances {
		currentWeight += instance.Weight
		if randomWeight < currentWeight {
			return instance
		}
	}
	
	return instances[0] // 兜底
}

// selectLeastConnections 最少连接选择 - 选择当前负载最小的实例
func (lb *modelLoadBalancer) selectLeastConnections(instances []*domain.ModelInstance) *domain.ModelInstance {
	minLoad := instances[0].CurrentLoad
	selectedInstance := instances[0]
	
	for _, instance := range instances {
		if instance.CurrentLoad < minLoad {
			minLoad = instance.CurrentLoad
			selectedInstance = instance
		}
	}
	
	return selectedInstance
}

// selectRandom 随机选择
func (lb *modelLoadBalancer) selectRandom(instances []*domain.ModelInstance) *domain.ModelInstance {
	return instances[rand.Intn(len(instances))]
}

// selectCostOptimized 成本优化选择 - 选择成本最低的实例
func (lb *modelLoadBalancer) selectCostOptimized(instances []*domain.ModelInstance) *domain.ModelInstance {
	bestCostPriority := 11 // 最大优先级是10
	var selectedInstance *domain.ModelInstance
	
	for _, instance := range instances {
		if instance.CostConfig != nil && instance.CostConfig.CostPriority < bestCostPriority {
			bestCostPriority = instance.CostConfig.CostPriority
			selectedInstance = instance
		}
	}
	
	if selectedInstance == nil {
		return instances[0] // 兜底
	}
	
	return selectedInstance
}

// selectFastestResponse 最快响应选择 - 选择响应时间最短的实例
func (lb *modelLoadBalancer) selectFastestResponse(instances []*domain.ModelInstance) *domain.ModelInstance {
	minResponseTime := int64(999999)
	selectedInstance := instances[0]
	
	for _, instance := range instances {
		if instance.Health != nil && instance.Health.ResponseTime < minResponseTime {
			minResponseTime = instance.Health.ResponseTime
			selectedInstance = instance
		}
	}
	
	return selectedInstance
}

// selectStickySession 会话粘性选择 - 同一会话使用相同实例
func (lb *modelLoadBalancer) selectStickySession(sessionID string, instances []*domain.ModelInstance) *domain.ModelInstance {
	lb.mu.RLock()
	instanceID, exists := lb.stickySessions[sessionID]
	lb.mu.RUnlock()
	
	if exists {
		// 查找对应的实例
		for _, instance := range instances {
			if instance.ID == instanceID {
				return instance
			}
		}
	}
	
	// 如果没有找到或者是新会话，随机选择一个并记录
	selectedInstance := lb.selectRandom(instances)
	
	lb.mu.Lock()
	lb.stickySessions[sessionID] = selectedInstance.ID
	lb.mu.Unlock()
	
	return selectedInstance
}

// UpdateInstanceHealth 更新实例健康状态
func (lb *modelLoadBalancer) UpdateInstanceHealth(instanceID uuid.UUID, health *domain.InstanceHealth) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	// 查找并更新实例健康状态
	for _, instances := range lb.modelInstances {
		for _, instance := range instances {
			if instance.ID == instanceID {
				instance.Health = health
				return nil
			}
		}
	}
	
	return fmt.Errorf("instance %s not found", instanceID)
}

// UpdateInstanceLoad 更新实例负载
func (lb *modelLoadBalancer) UpdateInstanceLoad(instanceID uuid.UUID, currentLoad int) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	// 查找并更新实例负载
	for _, instances := range lb.modelInstances {
		for _, instance := range instances {
			if instance.ID == instanceID {
				instance.CurrentLoad = currentLoad
				return nil
			}
		}
	}
	
	return fmt.Errorf("instance %s not found", instanceID)
}

// AddInstance 添加实例
func (lb *modelLoadBalancer) AddInstance(modelID uuid.UUID, instance *domain.ModelInstance) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	if lb.modelInstances[modelID] == nil {
		lb.modelInstances[modelID] = make([]*domain.ModelInstance, 0)
	}
	
	lb.modelInstances[modelID] = append(lb.modelInstances[modelID], instance)
	
	// 初始化熔断器状态
	lb.circuitStates[instance.ID] = &CircuitState{
		State: "closed",
	}
	
	lb.logger.WithFields(logrus.Fields{
		"model_id":    modelID,
		"instance_id": instance.ID,
		"instance_name": instance.Name,
	}).Info("Added model instance")
	
	return nil
}

// RemoveInstance 移除实例
func (lb *modelLoadBalancer) RemoveInstance(modelID uuid.UUID, instanceID uuid.UUID) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	instances, exists := lb.modelInstances[modelID]
	if !exists {
		return fmt.Errorf("model %s not found", modelID)
	}
	
	for i, instance := range instances {
		if instance.ID == instanceID {
			// 移除实例
			lb.modelInstances[modelID] = append(instances[:i], instances[i+1:]...)
			
			// 清理熔断器状态
			delete(lb.circuitStates, instanceID)
			
			// 清理会话粘性
			for sessionID, stickyInstanceID := range lb.stickySessions {
				if stickyInstanceID == instanceID {
					delete(lb.stickySessions, sessionID)
				}
			}
			
			lb.logger.WithFields(logrus.Fields{
				"model_id":    modelID,
				"instance_id": instanceID,
			}).Info("Removed model instance")
			
			return nil
		}
	}
	
	return fmt.Errorf("instance %s not found in model %s", instanceID, modelID)
}

// GetInstances 获取模型的所有实例
func (lb *modelLoadBalancer) GetInstances(modelID uuid.UUID) ([]*domain.ModelInstance, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	instances, exists := lb.modelInstances[modelID]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelID)
	}
	
	// 返回副本以避免并发修改
	result := make([]*domain.ModelInstance, len(instances))
	copy(result, instances)
	
	return result, nil
}

// SetStrategy 设置负载均衡策略
func (lb *modelLoadBalancer) SetStrategy(modelID uuid.UUID, strategy string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.strategies[modelID] = strategy
	
	lb.logger.WithFields(logrus.Fields{
		"model_id": modelID,
		"strategy": strategy,
	}).Info("Set load balancing strategy")
	
	return nil
}

// GetStrategy 获取负载均衡策略
func (lb *modelLoadBalancer) GetStrategy(modelID uuid.UUID) string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	strategy, exists := lb.strategies[modelID]
	if !exists {
		return "round_robin" // 默认策略
	}
	
	return strategy
}

// RecordSuccess 记录成功调用
func (lb *modelLoadBalancer) RecordSuccess(instanceID uuid.UUID) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	state, exists := lb.circuitStates[instanceID]
	if !exists {
		return fmt.Errorf("circuit state not found for instance %s", instanceID)
	}
	
	state.SuccessCount++
	state.FailureCount = 0 // 重置失败计数
	
	// 如果是半开状态，检查是否可以关闭熔断器
	if state.State == "half_open" {
		state.HalfOpenCallCount++
		// 这里可以根据配置决定何时关闭熔断器
		if state.SuccessCount >= 3 { // 连续3次成功就关闭
			state.State = "closed"
			state.HalfOpenCallCount = 0
		}
	}
	
	return nil
}

// RecordFailure 记录失败调用
func (lb *modelLoadBalancer) RecordFailure(instanceID uuid.UUID, err error) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	state, exists := lb.circuitStates[instanceID]
	if !exists {
		return fmt.Errorf("circuit state not found for instance %s", instanceID)
	}
	
	state.FailureCount++
	state.LastFailureTime = time.Now()
	
	// 检查是否需要打开熔断器
	if state.State == "closed" && state.FailureCount >= 5 { // 连续5次失败就打开
		state.State = "open"
		state.NextRetryTime = time.Now().Add(30 * time.Second) // 30秒后重试
	}
	
	return nil
}

// IsCircuitOpen 检查熔断器是否打开
func (lb *modelLoadBalancer) IsCircuitOpen(instanceID uuid.UUID) bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	state, exists := lb.circuitStates[instanceID]
	if !exists {
		return false // 默认关闭
	}
	
	switch state.State {
	case "open":
		// 检查是否可以进入半开状态
		if time.Now().After(state.NextRetryTime) {
			state.State = "half_open"
			state.HalfOpenCallCount = 0
			return false
		}
		return true
	case "half_open":
		// 半开状态限制调用次数
		return state.HalfOpenCallCount >= 3
	default:
		return false
	}
}
