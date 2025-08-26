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

// MediaRepository 媒体仓库接口
type MediaRepository interface {
	// 媒体生成任务
	CreateMediaTask(ctx context.Context, task *domain.MediaGenerationTask) error
	GetMediaTaskByID(ctx context.Context, id uuid.UUID) (*domain.MediaGenerationTask, error)
	UpdateMediaTask(ctx context.Context, task *domain.MediaGenerationTask) error
	DeleteMediaTask(ctx context.Context, id uuid.UUID) error
	ListMediaTasks(ctx context.Context, req *domain.MediaTaskListRequest) ([]*domain.MediaGenerationTask, int, error)

	// 媒体模板
	CreateMediaTemplate(ctx context.Context, template *domain.MediaTemplate) error
	GetMediaTemplateByID(ctx context.Context, id uuid.UUID) (*domain.MediaTemplate, error)
	UpdateMediaTemplate(ctx context.Context, template *domain.MediaTemplate) error
	DeleteMediaTemplate(ctx context.Context, id uuid.UUID) error
	ListMediaTemplates(ctx context.Context, templateType *string, isPublic *bool, page, limit int) ([]*domain.MediaTemplate, int, error)

	// 媒体资产
	CreateMediaAsset(ctx context.Context, asset *domain.MediaAsset) error
	GetMediaAssetByID(ctx context.Context, id uuid.UUID) (*domain.MediaAsset, error)
	UpdateMediaAsset(ctx context.Context, asset *domain.MediaAsset) error
	DeleteMediaAsset(ctx context.Context, id uuid.UUID) error
	ListMediaAssets(ctx context.Context, req *domain.MediaAssetListRequest) ([]*domain.MediaAsset, int, error)

	// 角色媒体关联
	CreateCharacterMedia(ctx context.Context, characterMedia *domain.CharacterMedia) error
	GetCharacterMedia(ctx context.Context, characterID uuid.UUID) ([]*domain.CharacterMedia, error)
	DeleteCharacterMedia(ctx context.Context, characterID uuid.UUID, mediaRole string) error
}

type mediaRepository struct {
	db *sqlx.DB
}

// NewMediaRepository 创建媒体仓库
func NewMediaRepository(db *sqlx.DB) MediaRepository {
	return &mediaRepository{db: db}
}

// CreateMediaTask 创建媒体生成任务
func (r *mediaRepository) CreateMediaTask(ctx context.Context, task *domain.MediaGenerationTask) error {
	query := `
		INSERT INTO media_generation_tasks (
			id, user_id, task_type, status, model_id, model_params,
			text_prompt, negative_prompt, reference_image_url, reference_video_url,
			result_urls, result_metadata, workflow_id, external_task_id, progress,
			error_message, estimated_cost, actual_cost, tokens_used, processing_time_seconds
		) VALUES (
			:id, :user_id, :task_type, :status, :model_id, :model_params,
			:text_prompt, :negative_prompt, :reference_image_url, :reference_video_url,
			:result_urls, :result_metadata, :workflow_id, :external_task_id, :progress,
			:error_message, :estimated_cost, :actual_cost, :tokens_used, :processing_time_seconds
		)`

	_, err := r.db.NamedExecContext(ctx, query, task)
	return err
}

// GetMediaTaskByID 根据ID获取媒体生成任务
func (r *mediaRepository) GetMediaTaskByID(ctx context.Context, id uuid.UUID) (*domain.MediaGenerationTask, error) {
	query := `
		SELECT id, user_id, task_type, status, model_id, model_params,
			   text_prompt, negative_prompt, reference_image_url, reference_video_url,
			   result_urls, result_metadata, workflow_id, external_task_id, progress,
			   error_message, estimated_cost, actual_cost, tokens_used, processing_time_seconds,
			   created_at, started_at, completed_at, updated_at
		FROM media_generation_tasks WHERE id = $1`

	var task domain.MediaGenerationTask
	err := r.db.GetContext(ctx, &task, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "媒体生成任务不存在", err)
		}
		return nil, err
	}

	return &task, nil
}

// UpdateMediaTask 更新媒体生成任务
func (r *mediaRepository) UpdateMediaTask(ctx context.Context, task *domain.MediaGenerationTask) error {
	query := `
		UPDATE media_generation_tasks SET
			status = :status, result_urls = :result_urls, result_metadata = :result_metadata,
			workflow_id = :workflow_id, external_task_id = :external_task_id, progress = :progress,
			error_message = :error_message, actual_cost = :actual_cost, tokens_used = :tokens_used,
			processing_time_seconds = :processing_time_seconds, started_at = :started_at,
			completed_at = :completed_at, updated_at = :updated_at
		WHERE id = :id`

	task.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, task)
	return err
}

// DeleteMediaTask 删除媒体生成任务
func (r *mediaRepository) DeleteMediaTask(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM media_generation_tasks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListMediaTasks 获取媒体生成任务列表
func (r *mediaRepository) ListMediaTasks(ctx context.Context, req *domain.MediaTaskListRequest) ([]*domain.MediaGenerationTask, int, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.TaskType != nil {
		conditions = append(conditions, fmt.Sprintf("task_type = $%d", argIndex))
		args = append(args, *req.TaskType)
		argIndex++
	}

	if req.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *req.DateFrom)
		argIndex++
	}

	if req.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *req.DateTo)
		argIndex++
	}

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM media_generation_tasks %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (req.Page - 1) * req.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, task_type, status, model_id, model_params,
			   text_prompt, negative_prompt, reference_image_url, reference_video_url,
			   result_urls, result_metadata, workflow_id, external_task_id, progress,
			   error_message, estimated_cost, actual_cost, tokens_used, processing_time_seconds,
			   created_at, started_at, completed_at, updated_at
		FROM media_generation_tasks %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	var tasks []*domain.MediaGenerationTask
	err = r.db.SelectContext(ctx, &tasks, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// CreateMediaTemplate 创建媒体模板
func (r *mediaRepository) CreateMediaTemplate(ctx context.Context, template *domain.MediaTemplate) error {
	query := `
		INSERT INTO media_templates (
			id, user_id, name, description, template_type, model_id,
			default_params, prompt_template, is_public, is_featured, usage_count
		) VALUES (
			:id, :user_id, :name, :description, :template_type, :model_id,
			:default_params, :prompt_template, :is_public, :is_featured, :usage_count
		)`

	_, err := r.db.NamedExecContext(ctx, query, template)
	return err
}

// GetMediaTemplateByID 根据ID获取媒体模板
func (r *mediaRepository) GetMediaTemplateByID(ctx context.Context, id uuid.UUID) (*domain.MediaTemplate, error) {
	query := `
		SELECT id, user_id, name, description, template_type, model_id,
			   default_params, prompt_template, is_public, is_featured, usage_count,
			   created_at, updated_at
		FROM media_templates WHERE id = $1`

	var template domain.MediaTemplate
	err := r.db.GetContext(ctx, &template, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "媒体模板不存在", err)
		}
		return nil, err
	}

	return &template, nil
}

// UpdateMediaTemplate 更新媒体模板
func (r *mediaRepository) UpdateMediaTemplate(ctx context.Context, template *domain.MediaTemplate) error {
	query := `
		UPDATE media_templates SET
			name = :name, description = :description, model_id = :model_id,
			default_params = :default_params, prompt_template = :prompt_template,
			is_public = :is_public, is_featured = :is_featured, usage_count = :usage_count,
			updated_at = :updated_at
		WHERE id = :id`

	template.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, template)
	return err
}

// DeleteMediaTemplate 删除媒体模板
func (r *mediaRepository) DeleteMediaTemplate(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM media_templates WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListMediaTemplates 获取媒体模板列表
func (r *mediaRepository) ListMediaTemplates(ctx context.Context, templateType *string, isPublic *bool, page, limit int) ([]*domain.MediaTemplate, int, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if templateType != nil {
		conditions = append(conditions, fmt.Sprintf("template_type = $%d", argIndex))
		args = append(args, *templateType)
		argIndex++
	}

	if isPublic != nil {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *isPublic)
		argIndex++
	}

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM media_templates %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, name, description, template_type, model_id,
			   default_params, prompt_template, is_public, is_featured, usage_count,
			   created_at, updated_at
		FROM media_templates %s
		ORDER BY is_featured DESC, usage_count DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	var templates []*domain.MediaTemplate
	err = r.db.SelectContext(ctx, &templates, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// CreateMediaAsset 创建媒体资产
func (r *mediaRepository) CreateMediaAsset(ctx context.Context, asset *domain.MediaAsset) error {
	query := `
		INSERT INTO media_assets (
			id, user_id, task_id, asset_type, file_url, thumbnail_url,
			file_size, width, height, duration_seconds, format,
			generation_params, model_used, prompt_used, title, description,
			tags, is_favorite, is_public, download_count, view_count
		) VALUES (
			:id, :user_id, :task_id, :asset_type, :file_url, :thumbnail_url,
			:file_size, :width, :height, :duration_seconds, :format,
			:generation_params, :model_used, :prompt_used, :title, :description,
			:tags, :is_favorite, :is_public, :download_count, :view_count
		)`

	_, err := r.db.NamedExecContext(ctx, query, asset)
	return err
}

// GetMediaAssetByID 根据ID获取媒体资产
func (r *mediaRepository) GetMediaAssetByID(ctx context.Context, id uuid.UUID) (*domain.MediaAsset, error) {
	query := `
		SELECT id, user_id, task_id, asset_type, file_url, thumbnail_url,
			   file_size, width, height, duration_seconds, format,
			   generation_params, model_used, prompt_used, title, description,
			   tags, is_favorite, is_public, download_count, view_count,
			   created_at, updated_at
		FROM media_assets WHERE id = $1`

	var asset domain.MediaAsset
	err := r.db.GetContext(ctx, &asset, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "媒体资产不存在", err)
		}
		return nil, err
	}

	return &asset, nil
}

// UpdateMediaAsset 更新媒体资产
func (r *mediaRepository) UpdateMediaAsset(ctx context.Context, asset *domain.MediaAsset) error {
	query := `
		UPDATE media_assets SET
			title = :title, description = :description, tags = :tags,
			is_favorite = :is_favorite, is_public = :is_public,
			download_count = :download_count, view_count = :view_count,
			updated_at = :updated_at
		WHERE id = :id`

	asset.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, asset)
	return err
}

// DeleteMediaAsset 删除媒体资产
func (r *mediaRepository) DeleteMediaAsset(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM media_assets WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListMediaAssets 获取媒体资产列表
func (r *mediaRepository) ListMediaAssets(ctx context.Context, req *domain.MediaAssetListRequest) ([]*domain.MediaAsset, int, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.AssetType != nil {
		conditions = append(conditions, fmt.Sprintf("asset_type = $%d", argIndex))
		args = append(args, *req.AssetType)
		argIndex++
	}

	if req.IsPublic != nil {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	if req.IsFavorite != nil {
		conditions = append(conditions, fmt.Sprintf("is_favorite = $%d", argIndex))
		args = append(args, *req.IsFavorite)
		argIndex++
	}

	if req.Tag != nil && *req.Tag != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(tags)", argIndex))
		args = append(args, *req.Tag)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		searchTerm := "%" + *req.Search + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM media_assets %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (req.Page - 1) * req.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, task_id, asset_type, file_url, thumbnail_url,
			   file_size, width, height, duration_seconds, format,
			   generation_params, model_used, prompt_used, title, description,
			   tags, is_favorite, is_public, download_count, view_count,
			   created_at, updated_at
		FROM media_assets %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	var assets []*domain.MediaAsset
	err = r.db.SelectContext(ctx, &assets, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// CreateCharacterMedia 创建角色媒体关联
func (r *mediaRepository) CreateCharacterMedia(ctx context.Context, characterMedia *domain.CharacterMedia) error {
	query := `
		INSERT INTO character_media (
			id, character_id, media_asset_id, media_role, is_auto_generated, generation_prompt
		) VALUES (
			:id, :character_id, :media_asset_id, :media_role, :is_auto_generated, :generation_prompt
		)`

	_, err := r.db.NamedExecContext(ctx, query, characterMedia)
	return err
}

// GetCharacterMedia 获取角色媒体关联
func (r *mediaRepository) GetCharacterMedia(ctx context.Context, characterID uuid.UUID) ([]*domain.CharacterMedia, error) {
	query := `
		SELECT id, character_id, media_asset_id, media_role, is_auto_generated, generation_prompt, created_at
		FROM character_media
		WHERE character_id = $1
		ORDER BY created_at DESC`

	var characterMedia []*domain.CharacterMedia
	err := r.db.SelectContext(ctx, &characterMedia, query, characterID)
	return characterMedia, err
}

// DeleteCharacterMedia 删除角色媒体关联
func (r *mediaRepository) DeleteCharacterMedia(ctx context.Context, characterID uuid.UUID, mediaRole string) error {
	query := `DELETE FROM character_media WHERE character_id = $1 AND media_role = $2`
	_, err := r.db.ExecContext(ctx, query, characterID, mediaRole)
	return err
}
