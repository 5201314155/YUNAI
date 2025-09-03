package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// AnalyticsService 分析服务接口
type AnalyticsService interface {
	// 用户行为分析
	TrackEvent(ctx context.Context, userID uuid.UUID, event string, properties map[string]interface{}) error
	TrackUserEvent(ctx context.Context, userID uuid.UUID, event string, properties map[string]interface{}) error
	GetUserBehaviorStats(ctx context.Context, userID uuid.UUID, timeRange string) (*domain.UserStats, error)

	// 系统性能监控
	RecordAPIMetrics(method, path string, duration time.Duration, statusCode int)
	GetSystemHealthMetrics(ctx context.Context) (*domain.SystemHealth, error)

	// 业务指标统计
	GetDAU(ctx context.Context, date time.Time) (int64, error)
	GetMAU(ctx context.Context, date time.Time) (int64, error)
	GetRevenue(ctx context.Context, timeRange string) (*domain.RevenueStats, error)
	GetUserRetention(ctx context.Context, cohortDate time.Time) (*domain.RetentionStats, error)

	// 实时监控
	GetRealTimeMetrics(ctx context.Context) (*domain.RealTimeMetrics, error)
	GetDashboard(ctx context.Context) (*domain.AnalyticsDashboard, error)
	GetSystemStats(ctx context.Context) (*domain.SystemStats, error)

	// 定期任务
	UpdateUserStats(ctx context.Context) error
	CollectSystemMetrics(ctx context.Context) error
}

// analyticsService 分析服务实现
type analyticsService struct {
	userRepo    repository.UserRepository
	redis       *redis.Client
	logger      *logrus.Logger
	apiMetrics  map[string]*domain.APIMetrics
	startTime   time.Time
}

// NewAnalyticsService 创建分析服务
func NewAnalyticsService(
	userRepo repository.UserRepository,
	redis *redis.Client,
	logger *logrus.Logger,
) AnalyticsService {
	return &analyticsService{
		userRepo:   userRepo,
		redis:      redis,
		logger:     logger,
		apiMetrics: make(map[string]*domain.APIMetrics),
		startTime:  time.Now(),
	}
}

// TrackEvent 追踪事件
func (s *analyticsService) TrackEvent(ctx context.Context, userID uuid.UUID, event string, properties map[string]interface{}) error {
	return s.TrackUserEvent(ctx, userID, event, properties)
}

// TrackUserEvent 追踪用户事件
func (s *analyticsService) TrackUserEvent(ctx context.Context, userID uuid.UUID, event string, properties map[string]interface{}) error {
	// 构建事件数据
	eventData := map[string]interface{}{
		"user_id":    userID.String(),
		"event":      event,
		"properties": properties,
		"timestamp":  time.Now().Unix(),
	}

	// 序列化事件数据
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// 存储到Redis
	eventKey := fmt.Sprintf("events:%s:%s", time.Now().Format("2006-01-02"), event)
	if err := s.redis.LPush(ctx, eventKey, eventJSON).Err(); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// 设置过期时间（保留30天）
	s.redis.Expire(ctx, eventKey, 30*24*time.Hour)

	// 更新实时计数器
	dailyKey := fmt.Sprintf("daily_events:%s", time.Now().Format("2006-01-02"))
	s.redis.HIncrBy(ctx, dailyKey, event, 1)
	s.redis.Expire(ctx, dailyKey, 30*24*time.Hour)

	// 更新用户活跃度
	userActiveKey := fmt.Sprintf("user_active:%s", time.Now().Format("2006-01-02"))
	s.redis.SAdd(ctx, userActiveKey, userID.String())
	s.redis.Expire(ctx, userActiveKey, 30*24*time.Hour)

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"event":   event,
	}).Debug("Event tracked")

	return nil
}

// GetUserBehaviorStats 获取用户行为统计
func (s *analyticsService) GetUserBehaviorStats(ctx context.Context, userID uuid.UUID, timeRange string) (*domain.UserStats, error) {
	// 计算时间范围
	endTime := time.Now()
	var startTime time.Time

	switch timeRange {
	case "today":
		startTime = time.Now().Truncate(24 * time.Hour)
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		startTime = time.Now().AddDate(0, 0, -7) // 默认一周
	}

	// 从Redis获取用户事件
	stats := &domain.UserStats{
		UserID:    userID,
		TimeRange: timeRange,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// 计算各种统计数据
	stats.LoginCount = s.getUserEventCount(ctx, userID, "user_login", startTime, endTime)
	stats.ChatCount = s.getUserEventCount(ctx, userID, "chat_message", startTime, endTime)
	stats.MomentCount = s.getUserEventCount(ctx, userID, "moment_created", startTime, endTime)

	// 计算活跃天数
	stats.ActiveDays = s.getUserActiveDays(ctx, userID, startTime, endTime)

	return stats, nil
}

// RecordAPIMetrics 记录API指标
func (s *analyticsService) RecordAPIMetrics(method, path string, duration time.Duration, statusCode int) {
	key := fmt.Sprintf("%s %s", method, path)

	if s.apiMetrics[key] == nil {
		s.apiMetrics[key] = &domain.APIMetrics{
			Method:       method,
			Path:         path,
			RequestCount: 0,
			TotalTime:    0,
			MinTime:      duration,
			MaxTime:      duration,
			SuccessCount: 0,
			ErrorCount:   0,
		}
	}

	metrics := s.apiMetrics[key]
	metrics.RequestCount++
	metrics.TotalTime += duration

	if duration < metrics.MinTime {
		metrics.MinTime = duration
	}
	if duration > metrics.MaxTime {
		metrics.MaxTime = duration
	}

	if statusCode >= 200 && statusCode < 400 {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	metrics.AvgTime = time.Duration(int64(metrics.TotalTime) / metrics.RequestCount)
	if metrics.RequestCount > 0 {
		metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100
	}
}

// GetSystemHealthMetrics 获取系统健康指标
func (s *analyticsService) GetSystemHealthMetrics(ctx context.Context) (*domain.SystemHealth, error) {
	health := &domain.SystemHealth{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(s.startTime),
		Services:  make(map[string]string),
	}

	// 检查Redis连接
	if _, err := s.redis.Ping(ctx).Result(); err != nil {
		health.Services["redis"] = "unhealthy"
		health.Status = "degraded"
	} else {
		health.Services["redis"] = "healthy"
	}

	// 检查数据库连接
	if count, err := s.userRepo.Count(ctx); err != nil {
		health.Services["database"] = "unhealthy"
		health.Status = "degraded"
	} else {
		health.Services["database"] = "healthy"
		health.UserCount = count
	}

	// 获取API指标
	health.APIMetrics = make(map[string]*domain.APIMetrics)
	for key, metrics := range s.apiMetrics {
		health.APIMetrics[key] = metrics
	}

	return health, nil
}

// GetDAU 获取日活跃用户数
func (s *analyticsService) GetDAU(ctx context.Context, date time.Time) (int64, error) {
	userActiveKey := fmt.Sprintf("user_active:%s", date.Format("2006-01-02"))
	count, err := s.redis.SCard(ctx, userActiveKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get DAU: %w", err)
	}
	return count, nil
}

// GetMAU 获取月活跃用户数
func (s *analyticsService) GetMAU(ctx context.Context, date time.Time) (int64, error) {
	// 获取过去30天的活跃用户
	activeUsers := make(map[string]bool)

	for i := 0; i < 30; i++ {
		day := date.AddDate(0, 0, -i)
		userActiveKey := fmt.Sprintf("user_active:%s", day.Format("2006-01-02"))
		
		users, err := s.redis.SMembers(ctx, userActiveKey).Result()
		if err != nil {
			continue
		}

		for _, userID := range users {
			activeUsers[userID] = true
		}
	}

	return int64(len(activeUsers)), nil
}

// GetRevenue 获取收入统计
func (s *analyticsService) GetRevenue(ctx context.Context, timeRange string) (*domain.RevenueStats, error) {
	// 这里需要从支付系统获取收入数据
	// 暂时返回模拟数据
	return &domain.RevenueStats{
		TimeRange:    timeRange,
		TotalRevenue: 0,
		OrderCount:   0,
		ARPU:         0,
		UpdatedAt:    time.Now(),
	}, nil
}

// GetUserRetention 获取用户留存率
func (s *analyticsService) GetUserRetention(ctx context.Context, cohortDate time.Time) (*domain.RetentionStats, error) {
	// 这里需要实现留存率计算逻辑
	// 暂时返回模拟数据
	return &domain.RetentionStats{
		CohortDate: cohortDate,
		Day1:       0.8,
		Day7:       0.6,
		Day30:      0.4,
		UpdatedAt:  time.Now(),
	}, nil
}

// GetRealTimeMetrics 获取实时指标
func (s *analyticsService) GetRealTimeMetrics(ctx context.Context) (*domain.RealTimeMetrics, error) {
	metrics := &domain.RealTimeMetrics{
		Timestamp: time.Now(),
	}

	// 获取今日活跃用户数
	today := time.Now().Format("2006-01-02")
	dau, _ := s.GetDAU(ctx, time.Now())
	metrics.OnlineUsers = dau

	// 获取今日事件统计
	dailyKey := fmt.Sprintf("daily_events:%s", today)
	eventCounts, err := s.redis.HGetAll(ctx, dailyKey).Result()
	if err == nil {
		metrics.EventCounts = eventCounts
	}

	// 获取系统健康状态
	health, err := s.GetSystemHealthMetrics(ctx)
	if err == nil {
		metrics.SystemStatus = health.Status
	}

	return metrics, nil
}

// GetDashboard 获取分析面板
func (s *analyticsService) GetDashboard(ctx context.Context) (*domain.AnalyticsDashboard, error) {
	dashboard := &domain.AnalyticsDashboard{
		UpdatedAt: time.Now(),
	}

	// 获取用户统计
	totalUsers, _ := s.userRepo.Count(ctx)
	dashboard.TotalUsers = totalUsers

	// 获取今日活跃用户
	dau, _ := s.GetDAU(ctx, time.Now())
	dashboard.DailyActiveUsers = dau

	// 获取月活跃用户
	mau, _ := s.GetMAU(ctx, time.Now())
	dashboard.MonthlyActiveUsers = mau

	// 获取实时指标
	realTimeMetrics, _ := s.GetRealTimeMetrics(ctx)
	dashboard.RealTimeMetrics = realTimeMetrics

	// 获取系统健康状态
	systemHealth, _ := s.GetSystemHealthMetrics(ctx)
	dashboard.SystemHealth = systemHealth

	return dashboard, nil
}

// GetSystemStats 获取系统统计
func (s *analyticsService) GetSystemStats(ctx context.Context) (*domain.SystemStats, error) {
	stats := &domain.SystemStats{
		UpdatedAt: time.Now(),
		Uptime:    time.Since(s.startTime),
	}

	// 获取用户统计
	totalUsers, _ := s.userRepo.Count(ctx)
	stats.TotalUsers = totalUsers

	// 获取活跃用户统计
	activeUsers, _ := s.userRepo.ListActive(ctx)
	stats.ActiveUsers = int64(len(activeUsers))

	// 获取API指标
	stats.APIMetrics = make(map[string]*domain.APIMetrics)
	for key, metrics := range s.apiMetrics {
		stats.APIMetrics[key] = metrics
	}

	return stats, nil
}

// UpdateUserStats 更新用户统计
func (s *analyticsService) UpdateUserStats(ctx context.Context) error {
	s.logger.Info("📊 更新用户统计...")

	// 获取今日活跃用户数
	dau, err := s.GetDAU(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get DAU: %w", err)
	}

	// 存储DAU到Redis
	dauKey := fmt.Sprintf("stats:dau:%s", time.Now().Format("2006-01-02"))
	if err := s.redis.Set(ctx, dauKey, dau, 30*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store DAU: %w", err)
	}

	// 获取月活跃用户数
	mau, err := s.GetMAU(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get MAU: %w", err)
	}

	// 存储MAU到Redis
	mauKey := fmt.Sprintf("stats:mau:%s", time.Now().Format("2006-01"))
	if err := s.redis.Set(ctx, mauKey, mau, 30*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store MAU: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"dau": dau,
		"mau": mau,
	}).Info("📊 用户统计更新完成")

	return nil
}

// CollectSystemMetrics 收集系统指标
func (s *analyticsService) CollectSystemMetrics(ctx context.Context) error {
	s.logger.Debug("📊 收集系统指标...")

	// 获取系统健康状态
	health, err := s.GetSystemHealthMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system health: %w", err)
	}

	// 存储系统健康状态到Redis
	healthKey := fmt.Sprintf("system:health:%d", time.Now().Unix())
	healthJSON, _ := json.Marshal(health)
	s.redis.Set(ctx, healthKey, healthJSON, time.Hour)

	// 清理旧的健康检查数据（保留24小时）
	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	pattern := "system:health:*"
	keys, _ := s.redis.Keys(ctx, pattern).Result()
	for _, key := range keys {
		// 解析时间戳
		var timestamp int64
		if _, err := fmt.Sscanf(key, "system:health:%d", &timestamp); err == nil {
			if timestamp < cutoff {
				s.redis.Del(ctx, key)
			}
		}
	}

	return nil
}

// 辅助方法

// getUserEventCount 获取用户事件计数
func (s *analyticsService) getUserEventCount(ctx context.Context, userID uuid.UUID, event string, startTime, endTime time.Time) int64 {
	count := int64(0)

	// 遍历时间范围内的每一天
	for d := startTime; d.Before(endTime) || d.Equal(endTime.Truncate(24*time.Hour)); d = d.AddDate(0, 0, 1) {
		eventKey := fmt.Sprintf("events:%s:%s", d.Format("2006-01-02"), event)
		
		// 获取该天的事件列表
		events, err := s.redis.LRange(ctx, eventKey, 0, -1).Result()
		if err != nil {
			continue
		}

		// 统计该用户的事件
		for _, eventJSON := range events {
			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(eventJSON), &eventData); err != nil {
				continue
			}

			if eventData["user_id"] == userID.String() {
				count++
			}
		}
	}

	return count
}

// getUserActiveDays 获取用户活跃天数
func (s *analyticsService) getUserActiveDays(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) int64 {
	activeDays := int64(0)

	// 遍历时间范围内的每一天
	for d := startTime; d.Before(endTime) || d.Equal(endTime.Truncate(24*time.Hour)); d = d.AddDate(0, 0, 1) {
		userActiveKey := fmt.Sprintf("user_active:%s", d.Format("2006-01-02"))
		
		// 检查用户是否在该天活跃
		isActive, err := s.redis.SIsMember(ctx, userActiveKey, userID.String()).Result()
		if err == nil && isActive {
			activeDays++
		}
	}

	return activeDays
}
