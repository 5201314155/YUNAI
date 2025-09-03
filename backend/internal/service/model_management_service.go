package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"yunai/internal/domain"
)

// ModelManagementService 模型管理服务
// 负责模型的启用/禁用、权限控制、健康检查等
type ModelManagementService interface {
	// 模型状态管理
	EnableModel(ctx context.Context, modelID uuid.UUID, adminID uuid.UUID, reason string) error
	DisableModel(ctx context.Context, modelID uuid.UUID, adminID uuid.UUID, reason string) error
	UpdateModelStatus(ctx context.Context, modelID uuid.UUID, status string, adminID uuid.UUID) error

	// 模型查询
	GetAvailableModels(ctx context.Context, userID *uuid.UUID, modelType string) ([]*domain.AIModel, error)
	GetAllModels(ctx context.Context, includeDisabled bool) ([]*domain.AIModel, error)
	GetModelByID(ctx context.Context, modelID uuid.UUID) (*domain.AIModel, error)
	GetModelByInternalKey(ctx context.Context, internalKey string) (*domain.AIModel, error)

	// 权限管理
	CheckModelAccess(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) (bool, error)
	AddUserToModelWhitelist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error
	RemoveUserFromModelWhitelist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error
	AddUserToModelBlacklist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error
	RemoveUserFromModelBlacklist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error

	// 健康检查
	PerformHealthCheck(ctx context.Context, modelID uuid.UUID) (*ModelHealthResult, error)
	PerformBatchHealthCheck(ctx context.Context) ([]*ModelHealthResult, error)
	GetModelHealthHistory(ctx context.Context, modelID uuid.UUID, limit int) ([]*ModelHealthResult, error)

	// 使用统计
	RecordModelUsage(ctx context.Context, usage *ModelUsageRecord) error
	GetModelUsageStats(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelUsageStats, error)

	// 模型分析
	GetModelUsageAnalytics(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelUsageAnalytics, error)
	GetModelCostAnalysis(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelCostAnalysis, error)
	GetModelPerformanceMetrics(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelPerformanceMetrics, error)

	// 模型同步
	SyncModelsFromProvider(ctx context.Context, provider string) error
	RefreshModelCapabilities(ctx context.Context, modelID uuid.UUID) error
}

// ModelHealthResult 模型健康检查结果
type ModelHealthResult struct {
	ModelID      uuid.UUID              `json:"model_id"`
	ModelName    string                 `json:"model_name"`
	Status       string                 `json:"status"`        // healthy, unhealthy, timeout, error
	ResponseTime int64                  `json:"response_time"` // 响应时间(ms)
	ErrorMessage string                 `json:"error_message"`
	CheckedAt    time.Time              `json:"checked_at"`
	Details      map[string]interface{} `json:"details"`
}

// ModelUsageRecord 模型使用记录
type ModelUsageRecord struct {
	ModelID      uuid.UUID              `json:"model_id"`
	UserID       uuid.UUID              `json:"user_id"`
	CharacterID  *uuid.UUID             `json:"character_id"`
	FunctionType string                 `json:"function_type"`
	InputTokens  int                    `json:"input_tokens"`
	OutputTokens int                    `json:"output_tokens"`
	ResponseTime int64                  `json:"response_time"`
	Cost         float64                `json:"cost"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message"`
	RequestData  map[string]interface{} `json:"request_data"`
	ResponseData map[string]interface{} `json:"response_data"`
}

// ModelUsageStats 模型使用统计
type ModelUsageStats struct {
	ModelID         uuid.UUID `json:"model_id"`
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	FailureRequests int64     `json:"failure_requests"`
	SuccessRate     float64   `json:"success_rate"`
	AvgResponseTime int64     `json:"avg_response_time"`
	TotalCost       float64   `json:"total_cost"`
	TotalTokens     int64     `json:"total_tokens"`
	TimeRange       string    `json:"time_range"`
}

// ModelUsageAnalytics 模型使用分析
type ModelUsageAnalytics struct {
	ModelID     uuid.UUID       `json:"model_id"`
	TimeRange   string          `json:"time_range"`
	TotalStats  ModelUsageStats `json:"total_stats"`
	HourlyUsage []HourlyUsage   `json:"hourly_usage"`
}

// HourlyUsage 小时使用量
type HourlyUsage struct {
	Hour            time.Time `json:"hour"`
	RequestCount    int64     `json:"request_count"`
	SuccessCount    int64     `json:"success_count"`
	AvgResponseTime float64   `json:"avg_response_time"`
	TotalCost       float64   `json:"total_cost"`
}

// ModelCostAnalysis 模型成本分析
type ModelCostAnalysis struct {
	ModelID            uuid.UUID          `json:"model_id"`
	TimeRange          string             `json:"time_range"`
	TotalCost          float64            `json:"total_cost"`
	CostByFunction     map[string]float64 `json:"cost_by_function"`
	EstimatedDailyCost float64            `json:"estimated_daily_cost"`
	AnalyzedAt         time.Time          `json:"analyzed_at"`
}

// ModelPerformanceMetrics 模型性能指标
type ModelPerformanceMetrics struct {
	ModelID           uuid.UUID        `json:"model_id"`
	TimeRange         string           `json:"time_range"`
	AvgResponseTime   int64            `json:"avg_response_time"`
	SuccessRate       float64          `json:"success_rate"`
	TotalRequests     int64            `json:"total_requests"`
	SuccessRequests   int64            `json:"success_requests"`
	FailureRequests   int64            `json:"failure_requests"`
	ErrorDistribution map[string]int64 `json:"error_distribution"`
	MeasuredAt        time.Time        `json:"measured_at"`
}

type modelManagementService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewModelManagementService 创建模型管理服务
func NewModelManagementService(db *gorm.DB, logger *logrus.Logger) ModelManagementService {
	return &modelManagementService{
		db:     db,
		logger: logger,
	}
}

// EnableModel 启用模型
func (s *modelManagementService) EnableModel(ctx context.Context, modelID uuid.UUID, adminID uuid.UUID, reason string) error {
	s.logger.WithFields(logrus.Fields{
		"model_id": modelID,
		"admin_id": adminID,
		"reason":   reason,
	}).Info("Enabling model")

	// 检查模型是否存在
	var model domain.AIModel
	if err := s.db.First(&model, modelID).Error; err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	// 更新模型状态
	updates := map[string]interface{}{
		"is_active":     true,
		"health_status": "unknown",
		"updated_at":    time.Now(),
	}

	if err := s.db.Model(&model).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to enable model: %w", err)
	}

	// 记录操作日志
	s.logModelOperation(ctx, modelID, adminID, "enable", reason)

	// 执行健康检查
	go func() {
		if _, err := s.PerformHealthCheck(context.Background(), modelID); err != nil {
			s.logger.WithError(err).Error("Failed to perform health check after enabling model")
		}
	}()

	s.logger.WithField("model_id", modelID).Info("Model enabled successfully")
	return nil
}

// DisableModel 禁用模型
func (s *modelManagementService) DisableModel(ctx context.Context, modelID uuid.UUID, adminID uuid.UUID, reason string) error {
	s.logger.WithFields(logrus.Fields{
		"model_id": modelID,
		"admin_id": adminID,
		"reason":   reason,
	}).Info("Disabling model")

	// 检查模型是否存在
	var model domain.AIModel
	if err := s.db.First(&model, modelID).Error; err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	// 更新模型状态
	updates := map[string]interface{}{
		"is_active":     false,
		"health_status": "disabled",
		"updated_at":    time.Now(),
	}

	if err := s.db.Model(&model).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to disable model: %w", err)
	}

	// 记录操作日志
	s.logModelOperation(ctx, modelID, adminID, "disable", reason)

	s.logger.WithField("model_id", modelID).Info("Model disabled successfully")
	return nil
}

// GetAvailableModels 获取可用模型列表
func (s *modelManagementService) GetAvailableModels(ctx context.Context, userID *uuid.UUID, modelType string) ([]*domain.AIModel, error) {
	query := s.db.Where("is_active = ?", true)

	// 按模型类型过滤
	if modelType != "" {
		query = query.Where("model_type = ?", modelType)
	}

	// 按健康状态过滤
	query = query.Where("health_status IN (?)", []string{"healthy", "unknown"})

	var models []*domain.AIModel
	if err := query.Order("weight DESC, display_name ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}

	// 如果指定了用户ID，进行权限过滤
	if userID != nil {
		var filteredModels []*domain.AIModel
		for _, model := range models {
			if accessible, err := s.CheckModelAccess(ctx, model.ID, *userID); err == nil && accessible {
				filteredModels = append(filteredModels, model)
			}
		}
		models = filteredModels
	}

	return models, nil
}

// GetAllModels 获取所有模型
func (s *modelManagementService) GetAllModels(ctx context.Context, includeDisabled bool) ([]*domain.AIModel, error) {
	query := s.db

	if !includeDisabled {
		query = query.Where("is_active = ?", true)
	}

	var models []*domain.AIModel
	if err := query.Order("provider ASC, display_name ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get all models: %w", err)
	}

	return models, nil
}

// CheckModelAccess 检查模型访问权限
func (s *modelManagementService) CheckModelAccess(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) (bool, error) {
	var model domain.AIModel
	if err := s.db.First(&model, modelID).Error; err != nil {
		return false, fmt.Errorf("model not found: %w", err)
	}

	// 检查模型是否激活
	if !model.IsActive {
		return false, nil
	}

	// 检查健康状态
	if model.HealthStatus == "unhealthy" || model.HealthStatus == "error" {
		return false, nil
	}

	// TODO: 实现更复杂的权限检查逻辑
	// - 检查用户类型
	// - 检查用户余额
	// - 检查白名单/黑名单
	// - 检查使用限制

	return true, nil
}

// PerformHealthCheck 执行健康检查
func (s *modelManagementService) PerformHealthCheck(ctx context.Context, modelID uuid.UUID) (*ModelHealthResult, error) {
	var model domain.AIModel
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	result := &ModelHealthResult{
		ModelID:   modelID,
		ModelName: model.DisplayName,
		CheckedAt: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// 执行健康检查
	startTime := time.Now()

	// TODO: 实现具体的健康检查逻辑
	// 这里可以调用模型的测试端点或发送测试请求

	// 模拟健康检查
	if model.IsActive {
		result.Status = "healthy"
		result.ResponseTime = time.Since(startTime).Milliseconds()
	} else {
		result.Status = "disabled"
		result.ErrorMessage = "Model is disabled"
	}

	// 更新模型健康状态
	updates := map[string]interface{}{
		"health_status":     result.Status,
		"last_health_check": &result.CheckedAt,
		"updated_at":        time.Now(),
	}

	if err := s.db.Model(&model).Updates(updates).Error; err != nil {
		s.logger.WithError(err).Error("Failed to update model health status")
	}

	return result, nil
}

// RecordModelUsage 记录模型使用
func (s *modelManagementService) RecordModelUsage(ctx context.Context, usage *ModelUsageRecord) error {
	// 创建使用记录
	usageRecord := &domain.ModelUsageRecord{
		ID:           uuid.New(),
		ModelID:      usage.ModelID,
		UserID:       usage.UserID,
		FunctionType: usage.FunctionType,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		Success:      usage.Success,
		Cost:         usage.Cost,
		Duration:     usage.ResponseTime,
		ErrorMessage: usage.ErrorMessage,
		CreatedAt:    time.Now(),
	}

	// 存储到数据库
	if err := s.db.Create(usageRecord).Error; err != nil {
		s.logger.WithError(err).Error("Failed to save model usage record")
		return fmt.Errorf("failed to save model usage record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id":      usage.ModelID,
		"user_id":       usage.UserID,
		"function_type": usage.FunctionType,
		"success":       usage.Success,
		"cost":          usage.Cost,
	}).Debug("Model usage recorded successfully")

	return nil
}

// logModelOperation 记录模型操作日志
func (s *modelManagementService) logModelOperation(ctx context.Context, modelID uuid.UUID, adminID uuid.UUID, operation string, reason string) {
	// 创建操作日志记录
	logRecord := &domain.ModelOperationLog{
		ID:        uuid.New(),
		ModelID:   modelID,
		AdminID:   adminID,
		Operation: operation,
		Reason:    reason,
		CreatedAt: time.Now(),
	}

	// 存储到数据库
	if err := s.db.Create(logRecord).Error; err != nil {
		s.logger.WithError(err).Error("Failed to save model operation log")
	}

	s.logger.WithFields(logrus.Fields{
		"model_id":  modelID,
		"admin_id":  adminID,
		"operation": operation,
		"reason":    reason,
		"timestamp": time.Now(),
	}).Info("Model operation logged")
}

// UpdateModelStatus 更新模型状态
func (s *modelManagementService) UpdateModelStatus(ctx context.Context, modelID uuid.UUID, status string, adminID uuid.UUID) error {
	// 获取模型
	var model domain.AIModel
	if err := s.db.First(&model, modelID).Error; err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	// 更新状态
	oldIsActive := model.IsActive
	newIsActive := (status == "active" || status == "enabled")
	model.IsActive = newIsActive
	model.UpdatedAt = time.Now()

	if err := s.db.Save(&model).Error; err != nil {
		return fmt.Errorf("failed to update model status: %w", err)
	}

	// 记录操作日志
	oldStatusStr := "inactive"
	if oldIsActive {
		oldStatusStr = "active"
	}
	reason := fmt.Sprintf("Status changed from %s to %s", oldStatusStr, status)
	s.logModelOperation(ctx, modelID, adminID, "update_status", reason)

	s.logger.WithFields(logrus.Fields{
		"model_id":   modelID,
		"old_status": oldStatusStr,
		"new_status": status,
		"admin_id":   adminID,
	}).Info("Model status updated")

	return nil
}

func (s *modelManagementService) GetModelByID(ctx context.Context, modelID uuid.UUID) (*domain.AIModel, error) {
	var model domain.AIModel
	if err := s.db.First(&model, modelID).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *modelManagementService) GetModelByInternalKey(ctx context.Context, internalKey string) (*domain.AIModel, error) {
	var model domain.AIModel
	if err := s.db.Where("internal_key = ?", internalKey).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *modelManagementService) AddUserToModelWhitelist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error {
	// 创建白名单记录
	whitelist := &domain.ModelUserWhitelist{
		ID:        uuid.New(),
		ModelID:   modelID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(whitelist).Error; err != nil {
		return fmt.Errorf("failed to add user to model whitelist: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id": modelID,
		"user_id":  userID,
	}).Info("User added to model whitelist")

	return nil
}

func (s *modelManagementService) RemoveUserFromModelWhitelist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error {
	if err := s.db.Where("model_id = ? AND user_id = ?", modelID, userID).Delete(&domain.ModelUserWhitelist{}).Error; err != nil {
		return fmt.Errorf("failed to remove user from model whitelist: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id": modelID,
		"user_id":  userID,
	}).Info("User removed from model whitelist")

	return nil
}

func (s *modelManagementService) AddUserToModelBlacklist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error {
	// 创建黑名单记录
	blacklist := &domain.ModelUserBlacklist{
		ID:        uuid.New(),
		ModelID:   modelID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(blacklist).Error; err != nil {
		return fmt.Errorf("failed to add user to model blacklist: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id": modelID,
		"user_id":  userID,
	}).Info("User added to model blacklist")

	return nil
}

func (s *modelManagementService) RemoveUserFromModelBlacklist(ctx context.Context, modelID uuid.UUID, userID uuid.UUID) error {
	if err := s.db.Where("model_id = ? AND user_id = ?", modelID, userID).Delete(&domain.ModelUserBlacklist{}).Error; err != nil {
		return fmt.Errorf("failed to remove user from model blacklist: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id": modelID,
		"user_id":  userID,
	}).Info("User removed from model blacklist")

	return nil
}

func (s *modelManagementService) PerformBatchHealthCheck(ctx context.Context) ([]*ModelHealthResult, error) {
	// 获取所有活跃模型
	var models []domain.AIModel
	if err := s.db.Where("is_active = ?", true).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get active models: %w", err)
	}

	var results []*ModelHealthResult
	for _, model := range models {
		result, err := s.PerformHealthCheck(ctx, model.ID)
		if err != nil {
			s.logger.WithError(err).WithField("model_id", model.ID).Error("Health check failed")
			// 创建失败结果
			result = &ModelHealthResult{
				ModelID:      model.ID,
				Status:       "unhealthy",
				CheckedAt:    time.Now(),
				ErrorMessage: err.Error(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *modelManagementService) GetModelHealthHistory(ctx context.Context, modelID uuid.UUID, limit int) ([]*ModelHealthResult, error) {
	// 从数据库获取健康检查历史
	var records []domain.ModelHealthRecord
	if err := s.db.Where("model_id = ?", modelID).Order("created_at DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get model health history: %w", err)
	}

	var results []*ModelHealthResult
	for _, record := range records {
		result := &ModelHealthResult{
			ModelID:      record.ModelID,
			Status:       record.Status,
			ResponseTime: record.ResponseTime,
			CheckedAt:    record.CreatedAt,
			ErrorMessage: record.ErrorMessage,
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *modelManagementService) GetModelUsageStats(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelUsageStats, error) {
	// 解析时间范围
	var startTime time.Time
	switch timeRange {
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	case "7d":
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = time.Now().Add(-30 * 24 * time.Hour)
	default:
		startTime = time.Now().Add(-24 * time.Hour)
	}

	// 查询使用记录
	var totalRequests int64
	var successRequests int64
	var totalCost float64
	var totalTokens int64
	var avgResponseTime float64

	// 总请求数
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ?", modelID, startTime).Count(&totalRequests)

	// 成功请求数
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ? AND success = ?", modelID, startTime, true).Count(&successRequests)

	// 总成本
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ?", modelID, startTime).Select("COALESCE(SUM(cost), 0)").Scan(&totalCost)

	// 总Token数
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ?", modelID, startTime).Select("COALESCE(SUM(input_tokens + output_tokens), 0)").Scan(&totalTokens)

	// 平均响应时间
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ? AND success = ?", modelID, startTime, true).Select("COALESCE(AVG(duration), 0)").Scan(&avgResponseTime)

	// 计算成功率
	var successRate float64
	if totalRequests > 0 {
		successRate = float64(successRequests) / float64(totalRequests)
	}

	stats := &ModelUsageStats{
		ModelID:         modelID,
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailureRequests: totalRequests - successRequests,
		SuccessRate:     successRate,
		AvgResponseTime: int64(avgResponseTime),
		TotalCost:       totalCost,
		TotalTokens:     totalTokens,
		TimeRange:       timeRange,
	}

	return stats, nil
}

// GetModelUsageAnalytics 获取模型使用分析
func (s *modelManagementService) GetModelUsageAnalytics(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelUsageAnalytics, error) {
	// 解析时间范围
	startTime := s.parseTimeRange(timeRange)

	// 获取使用统计
	stats, err := s.GetModelUsageStats(ctx, modelID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	// 获取按时间分组的使用数据
	var hourlyUsage []HourlyUsage
	query := `
		SELECT
			DATE_TRUNC('hour', created_at) as hour,
			COUNT(*) as request_count,
			SUM(CASE WHEN success THEN 1 ELSE 0 END) as success_count,
			AVG(duration) as avg_response_time,
			SUM(cost) as total_cost
		FROM model_usage_records
		WHERE model_id = $1 AND created_at >= $2
		GROUP BY DATE_TRUNC('hour', created_at)
		ORDER BY hour`

	rows, err := s.db.Raw(query, modelID, startTime).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly usage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var usage HourlyUsage
		if err := rows.Scan(&usage.Hour, &usage.RequestCount, &usage.SuccessCount, &usage.AvgResponseTime, &usage.TotalCost); err != nil {
			continue
		}
		hourlyUsage = append(hourlyUsage, usage)
	}

	analytics := &ModelUsageAnalytics{
		ModelID:     modelID,
		TimeRange:   timeRange,
		TotalStats:  *stats,
		HourlyUsage: hourlyUsage,
	}

	return analytics, nil
}

// GetModelCostAnalysis 获取模型成本分析
func (s *modelManagementService) GetModelCostAnalysis(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelCostAnalysis, error) {
	startTime := s.parseTimeRange(timeRange)

	// 获取总成本
	var totalCost float64
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ?", modelID, startTime).Select("COALESCE(SUM(cost), 0)").Scan(&totalCost)

	// 按功能类型分组的成本
	var costByFunction []struct {
		FunctionType string  `json:"function_type"`
		TotalCost    float64 `json:"total_cost"`
	}
	s.db.Model(&domain.ModelUsageRecord{}).
		Select("function_type, SUM(cost) as total_cost").
		Where("model_id = ? AND created_at >= ?", modelID, startTime).
		Group("function_type").
		Scan(&costByFunction)

	costByFunctionMap := make(map[string]float64)
	for _, item := range costByFunction {
		costByFunctionMap[item.FunctionType] = item.TotalCost
	}

	// 计算预估日成本
	days := time.Since(startTime).Hours() / 24
	estimatedDailyCost := totalCost / days

	analysis := &ModelCostAnalysis{
		ModelID:            modelID,
		TimeRange:          timeRange,
		TotalCost:          totalCost,
		CostByFunction:     costByFunctionMap,
		EstimatedDailyCost: estimatedDailyCost,
		AnalyzedAt:         time.Now(),
	}

	return analysis, nil
}

// GetModelPerformanceMetrics 获取模型性能指标
func (s *modelManagementService) GetModelPerformanceMetrics(ctx context.Context, modelID uuid.UUID, timeRange string) (*ModelPerformanceMetrics, error) {
	startTime := s.parseTimeRange(timeRange)

	// 获取性能指标
	var avgResponseTime float64
	var successRate float64
	var totalRequests int64
	var successRequests int64

	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ?", modelID, startTime).Count(&totalRequests)
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ? AND success = ?", modelID, startTime, true).Count(&successRequests)
	s.db.Model(&domain.ModelUsageRecord{}).Where("model_id = ? AND created_at >= ? AND success = ?", modelID, startTime, true).Select("COALESCE(AVG(duration), 0)").Scan(&avgResponseTime)

	if totalRequests > 0 {
		successRate = float64(successRequests) / float64(totalRequests)
	}

	// 获取错误分布
	var errorDistribution []struct {
		ErrorType string `json:"error_type"`
		Count     int64  `json:"count"`
	}
	s.db.Model(&domain.ModelUsageRecord{}).
		Select("COALESCE(error_message, 'unknown') as error_type, COUNT(*) as count").
		Where("model_id = ? AND created_at >= ? AND success = ?", modelID, startTime, false).
		Group("error_message").
		Scan(&errorDistribution)

	errorDistMap := make(map[string]int64)
	for _, item := range errorDistribution {
		errorDistMap[item.ErrorType] = item.Count
	}

	metrics := &ModelPerformanceMetrics{
		ModelID:           modelID,
		TimeRange:         timeRange,
		AvgResponseTime:   int64(avgResponseTime),
		SuccessRate:       successRate,
		TotalRequests:     totalRequests,
		SuccessRequests:   successRequests,
		FailureRequests:   totalRequests - successRequests,
		ErrorDistribution: errorDistMap,
		MeasuredAt:        time.Now(),
	}

	return metrics, nil
}

// parseTimeRange 解析时间范围
func (s *modelManagementService) parseTimeRange(timeRange string) time.Time {
	switch timeRange {
	case "24h":
		return time.Now().Add(-24 * time.Hour)
	case "7d":
		return time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		return time.Now().Add(-30 * 24 * time.Hour)
	default:
		return time.Now().Add(-24 * time.Hour)
	}
}

func (s *modelManagementService) SyncModelsFromProvider(ctx context.Context, provider string) error {
	return fmt.Errorf("not implemented")
}

func (s *modelManagementService) RefreshModelCapabilities(ctx context.Context, modelID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
