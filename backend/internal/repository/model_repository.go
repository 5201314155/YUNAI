package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// ModelRepository 模型仓库接口
type ModelRepository interface {
	// 模型管理
	CreateModel(ctx context.Context, model *domain.AIModel) error
	GetModelByID(ctx context.Context, id uuid.UUID) (*domain.AIModel, error)
	GetModelByInternalKey(ctx context.Context, internalKey string) (*domain.AIModel, error)
	UpdateModel(ctx context.Context, model *domain.AIModel) error
	DeleteModel(ctx context.Context, id uuid.UUID) error
	ListModels(ctx context.Context, req *domain.ModelListRequest) ([]*domain.AIModel, int, error)

	// 能力管理
	CreateModelCapability(ctx context.Context, capability *domain.ModelCapability) error
	GetModelCapabilities(ctx context.Context, modelID uuid.UUID) ([]*domain.ModelCapability, error)
	DeleteModelCapabilities(ctx context.Context, modelID uuid.UUID) error

	// 健康检查
	CreateHealthLog(ctx context.Context, log *domain.ModelHealthLog) error
	GetRecentHealthLogs(ctx context.Context, modelID uuid.UUID, limit int) ([]*domain.ModelHealthLog, error)
	UpdateModelHealth(ctx context.Context, modelID uuid.UUID, status string) error

	// 使用统计
	CreateOrUpdateUsageStats(ctx context.Context, stats *domain.ModelUsageStats) error
	GetUsageStats(ctx context.Context, modelID uuid.UUID, userID *uuid.UUID, startDate, endDate time.Time) ([]*domain.ModelUsageStats, error)
}

type modelRepository struct {
	db *sqlx.DB
}

// NewModelRepository 创建模型仓库
func NewModelRepository(db *sqlx.DB) ModelRepository {
	return &modelRepository{db: db}
}

// CreateModel 创建模型
func (r *modelRepository) CreateModel(ctx context.Context, model *domain.AIModel) error {
	query := `
		INSERT INTO ai_models (
			id, internal_key, display_name, provider, model_type,
			capabilities, params_schema, model_system_prompt, base_url, api_key_encrypted,
			pricing, visibility, min_user_type, min_balance, daily_limit, user_limit,
			health_status, weight, fallback_chain, connectivity_test_endpoint,
			is_active, is_featured
		) VALUES (
			:id, :internal_key, :display_name, :provider, :model_type,
			:capabilities, :params_schema, :model_system_prompt, :base_url, :api_key_encrypted,
			:pricing, :visibility, :min_user_type, :min_balance, :daily_limit, :user_limit,
			:health_status, :weight, :fallback_chain, :connectivity_test_endpoint,
			:is_active, :is_featured
		)`

	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}

// GetModelByID 根据ID获取模型
func (r *modelRepository) GetModelByID(ctx context.Context, id uuid.UUID) (*domain.AIModel, error) {
	query := `
		SELECT id, internal_key, display_name, provider, model_type,
			   capabilities, params_schema, model_system_prompt, base_url, api_key_encrypted,
			   pricing, visibility, min_user_type, min_balance, daily_limit, user_limit,
			   health_status, weight, fallback_chain, connectivity_test_endpoint,
			   is_active, is_featured, created_at, updated_at, last_health_check
		FROM ai_models WHERE id = $1`

	var model domain.AIModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "模型不存在", err)
		}
		return nil, err
	}

	return &model, nil
}

// GetModelByInternalKey 根据内部键获取模型
func (r *modelRepository) GetModelByInternalKey(ctx context.Context, internalKey string) (*domain.AIModel, error) {
	query := `
		SELECT id, internal_key, display_name, provider, model_type,
			   capabilities, params_schema, model_system_prompt, base_url, api_key_encrypted,
			   pricing, visibility, min_user_type, min_balance, daily_limit, user_limit,
			   health_status, weight, fallback_chain, connectivity_test_endpoint,
			   is_active, is_featured, created_at, updated_at, last_health_check
		FROM ai_models WHERE internal_key = $1`

	var model domain.AIModel
	err := r.db.GetContext(ctx, &model, query, internalKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "模型不存在", err)
		}
		return nil, err
	}

	return &model, nil
}

// UpdateModel 更新模型
func (r *modelRepository) UpdateModel(ctx context.Context, model *domain.AIModel) error {
	query := `
		UPDATE ai_models SET
			display_name = :display_name, model_system_prompt = :model_system_prompt,
			base_url = :base_url, api_key_encrypted = :api_key_encrypted,
			pricing = :pricing, visibility = :visibility, min_user_type = :min_user_type,
			min_balance = :min_balance, daily_limit = :daily_limit, user_limit = :user_limit,
			health_status = :health_status, weight = :weight, fallback_chain = :fallback_chain,
			connectivity_test_endpoint = :connectivity_test_endpoint,
			is_active = :is_active, is_featured = :is_featured, updated_at = :updated_at
		WHERE id = :id`

	model.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}

// DeleteModel 删除模型
func (r *modelRepository) DeleteModel(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ai_models WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListModels 获取模型列表
func (r *modelRepository) ListModels(ctx context.Context, req *domain.ModelListRequest) ([]*domain.AIModel, int, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 基础条件
	if req.Provider != nil {
		conditions = append(conditions, fmt.Sprintf("provider = $%d", argIndex))
		args = append(args, *req.Provider)
		argIndex++
	}

	if req.ModelType != nil {
		conditions = append(conditions, fmt.Sprintf("model_type = $%d", argIndex))
		args = append(args, *req.ModelType)
		argIndex++
	}

	if req.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, *req.Visibility)
		argIndex++
	}

	if req.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if req.IsFeatured != nil {
		conditions = append(conditions, fmt.Sprintf("is_featured = $%d", argIndex))
		args = append(args, *req.IsFeatured)
		argIndex++
	}

	// 权限过滤
	userTypeOrder := map[string]int{
		"basic": 1, "vip": 2, "creator": 3, "admin": 4,
	}
	currentUserLevel := userTypeOrder[req.UserType]
	
	var visibilityConditions []string
	if currentUserLevel >= userTypeOrder["basic"] {
		visibilityConditions = append(visibilityConditions, "visibility = 'public'")
	}
	if currentUserLevel >= userTypeOrder["vip"] {
		visibilityConditions = append(visibilityConditions, "visibility = 'vip_only'")
	}
	if currentUserLevel >= userTypeOrder["admin"] {
		visibilityConditions = append(visibilityConditions, "visibility = 'hidden'")
	}
	
	if len(visibilityConditions) > 0 {
		conditions = append(conditions, "("+strings.Join(visibilityConditions, " OR ")+")")
	}

	// 余额过滤
	conditions = append(conditions, fmt.Sprintf("min_balance <= $%d", argIndex))
	args = append(args, req.UserBalance)
	argIndex++

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_models %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (req.Page - 1) * req.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, internal_key, display_name, provider, model_type,
			   capabilities, params_schema, model_system_prompt, base_url, api_key_encrypted,
			   pricing, visibility, min_user_type, min_balance, daily_limit, user_limit,
			   health_status, weight, fallback_chain, connectivity_test_endpoint,
			   is_active, is_featured, created_at, updated_at, last_health_check
		FROM ai_models %s
		ORDER BY is_featured DESC, weight DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	var models []*domain.AIModel
	err = r.db.SelectContext(ctx, &models, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return models, total, nil
}

// CreateModelCapability 创建模型能力
func (r *modelRepository) CreateModelCapability(ctx context.Context, capability *domain.ModelCapability) error {
	query := `
		INSERT INTO model_capabilities (id, model_id, capability, is_enabled)
		VALUES (:id, :model_id, :capability, :is_enabled)`

	_, err := r.db.NamedExecContext(ctx, query, capability)
	return err
}

// GetModelCapabilities 获取模型能力
func (r *modelRepository) GetModelCapabilities(ctx context.Context, modelID uuid.UUID) ([]*domain.ModelCapability, error) {
	query := `
		SELECT id, model_id, capability, is_enabled, created_at
		FROM model_capabilities
		WHERE model_id = $1 AND is_enabled = TRUE
		ORDER BY created_at`

	var capabilities []*domain.ModelCapability
	err := r.db.SelectContext(ctx, &capabilities, query, modelID)
	return capabilities, err
}

// DeleteModelCapabilities 删除模型能力
func (r *modelRepository) DeleteModelCapabilities(ctx context.Context, modelID uuid.UUID) error {
	query := `DELETE FROM model_capabilities WHERE model_id = $1`
	_, err := r.db.ExecContext(ctx, query, modelID)
	return err
}

// CreateHealthLog 创建健康日志
func (r *modelRepository) CreateHealthLog(ctx context.Context, log *domain.ModelHealthLog) error {
	query := `
		INSERT INTO model_health_logs (id, model_id, status, response_time_ms, error_message)
		VALUES (:id, :model_id, :status, :response_time_ms, :error_message)`

	_, err := r.db.NamedExecContext(ctx, query, log)
	return err
}

// GetRecentHealthLogs 获取最近的健康日志
func (r *modelRepository) GetRecentHealthLogs(ctx context.Context, modelID uuid.UUID, limit int) ([]*domain.ModelHealthLog, error) {
	query := `
		SELECT id, model_id, status, response_time_ms, error_message, checked_at
		FROM model_health_logs
		WHERE model_id = $1
		ORDER BY checked_at DESC
		LIMIT $2`

	var logs []*domain.ModelHealthLog
	err := r.db.SelectContext(ctx, &logs, query, modelID, limit)
	return logs, err
}

// UpdateModelHealth 更新模型健康状态
func (r *modelRepository) UpdateModelHealth(ctx context.Context, modelID uuid.UUID, status string) error {
	query := `
		UPDATE ai_models 
		SET health_status = $1, last_health_check = NOW(), updated_at = NOW()
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, status, modelID)
	return err
}

// CreateOrUpdateUsageStats 创建或更新使用统计
func (r *modelRepository) CreateOrUpdateUsageStats(ctx context.Context, stats *domain.ModelUsageStats) error {
	query := `
		INSERT INTO model_usage_stats (
			id, model_id, user_id, date, request_count, success_count, error_count, total_tokens, total_cost
		) VALUES (
			:id, :model_id, :user_id, :date, :request_count, :success_count, :error_count, :total_tokens, :total_cost
		) ON CONFLICT (model_id, user_id, date) DO UPDATE SET
			request_count = model_usage_stats.request_count + EXCLUDED.request_count,
			success_count = model_usage_stats.success_count + EXCLUDED.success_count,
			error_count = model_usage_stats.error_count + EXCLUDED.error_count,
			total_tokens = model_usage_stats.total_tokens + EXCLUDED.total_tokens,
			total_cost = model_usage_stats.total_cost + EXCLUDED.total_cost,
			updated_at = NOW()`

	_, err := r.db.NamedExecContext(ctx, query, stats)
	return err
}

// GetUsageStats 获取使用统计
func (r *modelRepository) GetUsageStats(ctx context.Context, modelID uuid.UUID, userID *uuid.UUID, startDate, endDate time.Time) ([]*domain.ModelUsageStats, error) {
	query := `
		SELECT id, model_id, user_id, date, request_count, success_count, error_count, total_tokens, total_cost, created_at, updated_at
		FROM model_usage_stats
		WHERE model_id = $1 AND date BETWEEN $2 AND $3`

	args := []interface{}{modelID, startDate, endDate}

	if userID != nil {
		query += " AND user_id = $4"
		args = append(args, *userID)
	}

	query += " ORDER BY date DESC"

	var stats []*domain.ModelUsageStats
	err := r.db.SelectContext(ctx, &stats, query, args...)
	return stats, err
}
