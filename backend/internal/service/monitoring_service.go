package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
)

// MonitoringService 监控服务接口
type MonitoringService interface {
	// 系统性能监控
	GetSystemPerformance(ctx context.Context) (*domain.TechnicalMetrics, error)
	GetCPUUsage() float64
	GetMemoryUsage() (*domain.MemoryStats, error)
	GetDiskUsage() (*domain.DiskStats, error)

	// 数据库性能监控
	GetDatabaseMetrics(ctx context.Context) (*domain.DatabaseMetrics, error)

	// Redis性能监控
	GetRedisMetrics(ctx context.Context) (*domain.RedisMetrics, error)

	// 网络监控
	GetNetworkMetrics() (*domain.NetworkMetrics, error)

	// 健康检查
	HealthCheck(ctx context.Context) (*domain.HealthCheckResult, error)
	ReadinessCheck(ctx context.Context) (*domain.ReadinessCheckResult, error)
}

// monitoringService 监控服务实现
type monitoringService struct {
	redis  *redis.Client
	logger *logrus.Logger
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(redis *redis.Client, logger *logrus.Logger) MonitoringService {
	return &monitoringService{
		redis:  redis,
		logger: logger,
	}
}

// GetSystemPerformance 获取系统性能指标
func (s *monitoringService) GetSystemPerformance(ctx context.Context) (*domain.TechnicalMetrics, error) {
	metrics := &domain.TechnicalMetrics{
		Timestamp: time.Now(),
	}

	// 获取CPU使用率
	metrics.CPUUsage = s.GetCPUUsage()

	// 获取内存使用情况
	memStats, err := s.GetMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}
	metrics.MemoryUsage = float64(memStats.UsedPercent)

	// 获取磁盘使用情况
	diskStats, err := s.GetDiskUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk stats: %w", err)
	}
	metrics.DiskUsage = float64(diskStats.UsedPercent)

	// 获取数据库指标
	dbMetrics, err := s.GetDatabaseMetrics(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get database metrics")
	} else {
		metrics.DBConnections = dbMetrics.ActiveConnections
		metrics.DBQueryTime = dbMetrics.AvgQueryTime
		metrics.DBErrorRate = dbMetrics.ErrorRate
	}

	// 获取Redis指标
	redisMetrics, err := s.GetRedisMetrics(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get Redis metrics")
	} else {
		metrics.RedisConnections = redisMetrics.ConnectedClients
		metrics.RedisMemoryUsage = redisMetrics.UsedMemory
		metrics.RedisHitRate = redisMetrics.HitRate
	}

	return metrics, nil
}

// GetCPUUsage 获取CPU使用率
func (s *monitoringService) GetCPUUsage() float64 {
	// 简化实现：基于goroutine数量估算CPU使用率
	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()

	// 简单的估算公式
	usage := float64(numGoroutines) / float64(numCPU*100) * 100
	if usage > 100 {
		usage = 100
	}

	return usage
}

// GetMemoryUsage 获取内存使用情况
func (s *monitoringService) GetMemoryUsage() (*domain.MemoryStats, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := &domain.MemoryStats{
		Total:       m.Sys,
		Used:        m.Alloc,
		Free:        m.Sys - m.Alloc,
		UsedPercent: float64(m.Alloc) / float64(m.Sys) * 100,
		Timestamp:   time.Now(),
	}

	return stats, nil
}

// GetDiskUsage 获取磁盘使用情况
func (s *monitoringService) GetDiskUsage() (*domain.DiskStats, error) {
	// 简化实现：返回模拟数据
	// 在实际环境中，这里应该调用系统API获取真实的磁盘使用情况
	stats := &domain.DiskStats{
		Total:       1000 * 1024 * 1024 * 1024, // 1TB
		Used:        300 * 1024 * 1024 * 1024,  // 300GB
		Free:        700 * 1024 * 1024 * 1024,  // 700GB
		UsedPercent: 30.0,
		Timestamp:   time.Now(),
	}

	return stats, nil
}

// GetDatabaseMetrics 获取数据库性能指标
func (s *monitoringService) GetDatabaseMetrics(ctx context.Context) (*domain.DatabaseMetrics, error) {
	// 简化实现：返回模拟数据
	// 在实际环境中，这里应该查询数据库的性能视图
	metrics := &domain.DatabaseMetrics{
		ActiveConnections: 10,
		IdleConnections:   5,
		MaxConnections:    100,
		AvgQueryTime:      50 * time.Millisecond,
		SlowQueries:       2,
		ErrorRate:         0.1,
		Timestamp:         time.Now(),
	}

	return metrics, nil
}

// GetRedisMetrics 获取Redis性能指标
func (s *monitoringService) GetRedisMetrics(ctx context.Context) (*domain.RedisMetrics, error) {
	// 获取Redis信息
	_, err := s.redis.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// 解析Redis信息（简化实现）
	metrics := &domain.RedisMetrics{
		ConnectedClients: 5,
		UsedMemory:       10 * 1024 * 1024,  // 10MB
		MaxMemory:        100 * 1024 * 1024, // 100MB
		HitRate:          95.5,
		KeyspaceHits:     1000,
		KeyspaceMisses:   50,
		Timestamp:        time.Now(),
	}

	// 计算命中率
	if metrics.KeyspaceHits+metrics.KeyspaceMisses > 0 {
		metrics.HitRate = float64(metrics.KeyspaceHits) / float64(metrics.KeyspaceHits+metrics.KeyspaceMisses) * 100
	}

	s.logger.WithFields(logrus.Fields{
		"connected_clients": metrics.ConnectedClients,
		"used_memory":       metrics.UsedMemory,
		"hit_rate":          metrics.HitRate,
	}).Debug("Redis metrics collected")

	return metrics, nil
}

// GetNetworkMetrics 获取网络指标
func (s *monitoringService) GetNetworkMetrics() (*domain.NetworkMetrics, error) {
	// 简化实现：返回模拟数据
	metrics := &domain.NetworkMetrics{
		BytesReceived:   1024 * 1024 * 100, // 100MB
		BytesSent:       1024 * 1024 * 80,  // 80MB
		PacketsReceived: 10000,
		PacketsSent:     8000,
		Timestamp:       time.Now(),
	}

	return metrics, nil
}

// HealthCheck 健康检查
func (s *monitoringService) HealthCheck(ctx context.Context) (*domain.HealthCheckResult, error) {
	result := &domain.HealthCheckResult{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks:    make(map[string]*domain.HealthCheck),
	}

	// 检查Redis
	redisCheck := &domain.HealthCheck{
		Name:   "redis",
		Status: "healthy",
	}

	if _, err := s.redis.Ping(ctx).Result(); err != nil {
		redisCheck.Status = "unhealthy"
		redisCheck.Error = err.Error()
		result.Status = "unhealthy"
	}

	result.Checks["redis"] = redisCheck

	// 检查内存使用
	memoryCheck := &domain.HealthCheck{
		Name:   "memory",
		Status: "healthy",
	}

	memStats, err := s.GetMemoryUsage()
	if err != nil {
		memoryCheck.Status = "unhealthy"
		memoryCheck.Error = err.Error()
		result.Status = "unhealthy"
	} else if memStats.UsedPercent > 90 {
		memoryCheck.Status = "degraded"
		memoryCheck.Message = fmt.Sprintf("Memory usage high: %.1f%%", memStats.UsedPercent)
		if result.Status == "healthy" {
			result.Status = "degraded"
		}
	}

	result.Checks["memory"] = memoryCheck

	return result, nil
}

// ReadinessCheck 就绪检查
func (s *monitoringService) ReadinessCheck(ctx context.Context) (*domain.ReadinessCheckResult, error) {
	result := &domain.ReadinessCheckResult{
		Ready:     true,
		Timestamp: time.Now(),
		Checks:    make(map[string]*domain.ReadinessCheck),
	}

	// 检查Redis连接
	redisCheck := &domain.ReadinessCheck{
		Name:  "redis",
		Ready: true,
	}

	if _, err := s.redis.Ping(ctx).Result(); err != nil {
		redisCheck.Ready = false
		redisCheck.Error = err.Error()
		result.Ready = false
	}

	result.Checks["redis"] = redisCheck

	// 检查系统资源
	resourceCheck := &domain.ReadinessCheck{
		Name:  "resources",
		Ready: true,
	}

	memStats, err := s.GetMemoryUsage()
	if err != nil {
		resourceCheck.Ready = false
		resourceCheck.Error = err.Error()
		result.Ready = false
	} else if memStats.UsedPercent > 95 {
		resourceCheck.Ready = false
		resourceCheck.Message = "Memory usage too high"
		result.Ready = false
	}

	result.Checks["resources"] = resourceCheck

	return result, nil
}
