package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// MediaService 媒体服务接口
type MediaService interface {
	// 媒体生成任务
	CreateMediaTask(ctx context.Context, userID uuid.UUID, req *domain.CreateMediaTaskRequest) (*domain.MediaTaskResponse, error)
	GetMediaTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.MediaTaskResponse, error)
	UpdateMediaTask(ctx context.Context, id uuid.UUID, req *domain.UpdateMediaTaskRequest) (*domain.MediaTaskResponse, error)
	CancelMediaTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ListMediaTasks(ctx context.Context, req *domain.MediaTaskListRequest, viewerUserID uuid.UUID) (*domain.MediaTaskListResponse, error)

	// 媒体模板
	CreateMediaTemplate(ctx context.Context, userID uuid.UUID, req *domain.CreateMediaTemplateRequest) (*domain.MediaTemplate, error)
	GetMediaTemplate(ctx context.Context, id uuid.UUID) (*domain.MediaTemplate, error)
	ListMediaTemplates(ctx context.Context, templateType *string, page, limit int) ([]*domain.MediaTemplate, int, error)

	// 媒体资产
	GetMediaAsset(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.MediaAssetResponse, error)
	ListMediaAssets(ctx context.Context, req *domain.MediaAssetListRequest, viewerUserID *uuid.UUID) (*domain.MediaAssetListResponse, error)
	UpdateMediaAsset(ctx context.Context, id uuid.UUID, userID uuid.UUID, title, description *string, tags []string, isFavorite, isPublic *bool) (*domain.MediaAsset, error)
	DeleteMediaAsset(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// 角色自动生成
	AutoGenerateCharacterImages(ctx context.Context, userID uuid.UUID, req *domain.AutoGenerateCharacterImagesRequest) (*domain.AutoGenerateCharacterImagesResponse, error)

	// 媒体处理（内部方法）
	ProcessMediaTask(ctx context.Context, taskID uuid.UUID) error
}

type mediaService struct {
	mediaRepo        repository.MediaRepository
	characterRepo    repository.CharacterRepository
	modelService     ModelService
	aiBillingService AIBillingService
	logger           *logrus.Logger
}

// NewMediaService 创建媒体服务
func NewMediaService(
	mediaRepo repository.MediaRepository,
	characterRepo repository.CharacterRepository,
	modelService ModelService,
	aiBillingService AIBillingService,
	logger *logrus.Logger,
) MediaService {
	return &mediaService{
		mediaRepo:        mediaRepo,
		characterRepo:    characterRepo,
		modelService:     modelService,
		aiBillingService: aiBillingService,
		logger:           logger,
	}
}

// CreateMediaTask 创建媒体生成任务
func (s *mediaService) CreateMediaTask(ctx context.Context, userID uuid.UUID, req *domain.CreateMediaTaskRequest) (*domain.MediaTaskResponse, error) {
	// 验证模型ID
	var modelResp *domain.ModelResponse
	if req.ModelID != nil {
		var err error
		modelResp, err = s.modelService.GetModel(ctx, *req.ModelID)
		if err != nil {
			return nil, fmt.Errorf("invalid model ID: %w", err)
		}

		// 检查模型是否支持该任务类型
		if !s.modelSupportsTaskType(modelResp, req.TaskType) {
			return nil, domain.NewAppError(domain.CodeBadRequest, "模型不支持该任务类型", nil)
		}
	}

	// 验证输入参数
	if err := s.validateTaskInput(req); err != nil {
		return nil, err
	}

	// 估算成本
	estimatedCost, err := s.estimateTaskCost(ctx, req, modelResp)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate cost: %w", err)
	}

	// 预扣费检查
	if modelResp != nil {
		pricingReq := &domain.PricingCalculationRequest{
			ModelID:  *req.ModelID,
			UserType: "basic", // TODO: 从用户信息获取
		}

		preCheck, err := s.aiBillingService.PreCheckBalance(ctx, userID, *req.ModelID, pricingReq)
		if err != nil {
			return nil, fmt.Errorf("failed to check balance: %w", err)
		}

		if !preCheck.Sufficient {
			return nil, domain.NewAppError(domain.CodePaymentRequired, preCheck.ErrorMessage, nil)
		}
	}

	// 创建任务对象
	task := &domain.MediaGenerationTask{
		ID:                uuid.New(),
		UserID:            userID,
		TaskType:          req.TaskType,
		Status:            domain.TaskStatusPending,
		ModelID:           req.ModelID,
		ModelParams:       req.ModelParams,
		TextPrompt:        req.TextPrompt,
		NegativePrompt:    req.NegativePrompt,
		ReferenceImageURL: req.ReferenceImageURL,
		ReferenceVideoURL: req.ReferenceVideoURL,
		Progress:          0,
		EstimatedCost:     &estimatedCost,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// 如果没有提供ModelParams，设置默认值
	if task.ModelParams == nil {
		task.ModelParams = json.RawMessage("{}")
	}

	// 确保ResultMetadata有默认值
	if task.ResultMetadata == nil {
		task.ResultMetadata = json.RawMessage("{}")
	}

	// 保存任务
	err = s.mediaRepo.CreateMediaTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create media task: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"task_id":   task.ID,
		"user_id":   userID,
		"task_type": task.TaskType,
	}).Info("Media generation task created")

	// 异步处理任务
	go func() {
		if err := s.ProcessMediaTask(context.Background(), task.ID); err != nil {
			s.logger.WithError(err).WithField("task_id", task.ID).Error("Failed to process media task")
		}
	}()

	return s.buildMediaTaskResponse(ctx, task)
}

// GetMediaTask 获取媒体生成任务
func (s *mediaService) GetMediaTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.MediaTaskResponse, error) {
	task, err := s.mediaRepo.GetMediaTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if task.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权查看此任务", nil)
	}

	return s.buildMediaTaskResponse(ctx, task)
}

// UpdateMediaTask 更新媒体生成任务
func (s *mediaService) UpdateMediaTask(ctx context.Context, id uuid.UUID, req *domain.UpdateMediaTaskRequest) (*domain.MediaTaskResponse, error) {
	task, err := s.mediaRepo.GetMediaTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Status != nil {
		task.Status = *req.Status
		if *req.Status == domain.TaskStatusProcessing && task.StartedAt == nil {
			now := time.Now()
			task.StartedAt = &now
		}
		if *req.Status == domain.TaskStatusCompleted || *req.Status == domain.TaskStatusFailed {
			now := time.Now()
			task.CompletedAt = &now
		}
	}

	if req.Progress != nil {
		task.Progress = *req.Progress
	}

	if req.ResultURLs != nil {
		task.ResultURLs = req.ResultURLs
	}

	if req.ResultMetadata != nil {
		task.ResultMetadata = req.ResultMetadata
	}

	if req.ErrorMessage != nil {
		task.ErrorMessage = req.ErrorMessage
	}

	if req.ActualCost != nil {
		task.ActualCost = req.ActualCost
	}

	if req.TokensUsed != nil {
		task.TokensUsed = req.TokensUsed
	}

	// 计算处理时间
	if task.StartedAt != nil && task.CompletedAt != nil {
		processingTime := int(task.CompletedAt.Sub(*task.StartedAt).Seconds())
		task.ProcessingTimeSeconds = &processingTime
	}

	// 保存更新
	err = s.mediaRepo.UpdateMediaTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to update media task: %w", err)
	}

	// 如果任务完成，创建媒体资产
	if task.Status == domain.TaskStatusCompleted && len(task.ResultURLs) > 0 {
		err = s.createMediaAssetsFromTask(ctx, task)
		if err != nil {
			s.logger.WithError(err).WithField("task_id", task.ID).Warn("Failed to create media assets")
		}
	}

	return s.buildMediaTaskResponse(ctx, task)
}

// CancelMediaTask 取消媒体生成任务
func (s *mediaService) CancelMediaTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	task, err := s.mediaRepo.GetMediaTaskByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限
	if task.UserID != userID {
		return domain.NewAppError(domain.CodeForbidden, "无权取消此任务", nil)
	}

	// 只能取消待处理或处理中的任务
	if task.Status != domain.TaskStatusPending && task.Status != domain.TaskStatusProcessing {
		return domain.NewAppError(domain.CodeBadRequest, "任务状态不允许取消", nil)
	}

	// 更新状态
	task.Status = domain.TaskStatusCancelled
	now := time.Now()
	task.CompletedAt = &now

	err = s.mediaRepo.UpdateMediaTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to cancel media task: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"task_id": task.ID,
		"user_id": userID,
	}).Info("Media generation task cancelled")

	return nil
}

// ListMediaTasks 获取媒体生成任务列表
func (s *mediaService) ListMediaTasks(ctx context.Context, req *domain.MediaTaskListRequest, viewerUserID uuid.UUID) (*domain.MediaTaskListResponse, error) {
	// 如果没有指定用户ID，默认查看自己的任务
	if req.UserID == nil {
		req.UserID = &viewerUserID
	}

	// 检查权限：只能查看自己的任务
	if *req.UserID != viewerUserID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权查看其他用户的任务", nil)
	}

	tasks, total, err := s.mediaRepo.ListMediaTasks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list media tasks: %w", err)
	}

	// 构建响应
	var taskResponses []domain.MediaTaskResponse
	for _, task := range tasks {
		resp, err := s.buildMediaTaskResponse(ctx, task)
		if err != nil {
			s.logger.WithError(err).WithField("task_id", task.ID).Warn("Failed to build task response")
			continue
		}
		taskResponses = append(taskResponses, *resp)
	}

	return &domain.MediaTaskListResponse{
		Tasks: taskResponses,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

// CreateMediaTemplate 创建媒体模板
func (s *mediaService) CreateMediaTemplate(ctx context.Context, userID uuid.UUID, req *domain.CreateMediaTemplateRequest) (*domain.MediaTemplate, error) {
	template := &domain.MediaTemplate{
		ID:             uuid.New(),
		UserID:         &userID,
		Name:           req.Name,
		Description:    req.Description,
		TemplateType:   req.TemplateType,
		ModelID:        req.ModelID,
		DefaultParams:  req.DefaultParams,
		PromptTemplate: req.PromptTemplate,
		IsPublic:       req.IsPublic,
		IsFeatured:     false,
		UsageCount:     0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if template.DefaultParams == nil {
		template.DefaultParams = json.RawMessage("{}")
	}

	err := s.mediaRepo.CreateMediaTemplate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to create media template: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"template_id": template.ID,
		"user_id":     userID,
		"name":        template.Name,
	}).Info("Media template created")

	return template, nil
}

// GetMediaTemplate 获取媒体模板
func (s *mediaService) GetMediaTemplate(ctx context.Context, id uuid.UUID) (*domain.MediaTemplate, error) {
	return s.mediaRepo.GetMediaTemplateByID(ctx, id)
}

// ListMediaTemplates 获取媒体模板列表
func (s *mediaService) ListMediaTemplates(ctx context.Context, templateType *string, page, limit int) ([]*domain.MediaTemplate, int, error) {
	isPublic := true
	return s.mediaRepo.ListMediaTemplates(ctx, templateType, &isPublic, page, limit)
}

// 辅助方法

// buildMediaTaskResponse 构建媒体任务响应
func (s *mediaService) buildMediaTaskResponse(ctx context.Context, task *domain.MediaGenerationTask) (*domain.MediaTaskResponse, error) {
	resp := &domain.MediaTaskResponse{
		MediaGenerationTask: task,
	}

	// 获取模型名称
	if task.ModelID != nil {
		model, err := s.modelService.GetModel(ctx, *task.ModelID)
		if err == nil {
			resp.ModelName = &model.DisplayName
		}
	}

	// 获取生成的媒体资产
	if task.Status == domain.TaskStatusCompleted {
		// TODO: 获取关联的媒体资产
	}

	return resp, nil
}

// modelSupportsTaskType 检查模型是否支持任务类型
func (s *mediaService) modelSupportsTaskType(model *domain.ModelResponse, taskType string) bool {
	for _, capability := range model.CapabilityList {
		if capability == taskType {
			return true
		}
	}
	return false
}

// validateTaskInput 验证任务输入
func (s *mediaService) validateTaskInput(req *domain.CreateMediaTaskRequest) error {
	switch req.TaskType {
	case domain.TaskTypeTxt2Img, domain.TaskTypeTxt2Video:
		if req.TextPrompt == nil || strings.TrimSpace(*req.TextPrompt) == "" {
			return domain.NewAppError(domain.CodeBadRequest, "文本提示词不能为空", nil)
		}
	case domain.TaskTypeImg2Img, domain.TaskTypeImg2Video, domain.TaskTypeRemoveBG:
		if req.ReferenceImageURL == nil || strings.TrimSpace(*req.ReferenceImageURL) == "" {
			return domain.NewAppError(domain.CodeBadRequest, "参考图片不能为空", nil)
		}
	}
	return nil
}

// estimateTaskCost 估算任务成本
func (s *mediaService) estimateTaskCost(ctx context.Context, req *domain.CreateMediaTaskRequest, model *domain.ModelResponse) (float64, error) {
	// 基础成本估算
	baseCost := 0.01 // 默认基础成本

	switch req.TaskType {
	case domain.TaskTypeTxt2Img:
		baseCost = 0.02
	case domain.TaskTypeImg2Img:
		baseCost = 0.03
	case domain.TaskTypeImg2Video:
		baseCost = 0.10
	case domain.TaskTypeTxt2Video:
		baseCost = 0.15
	case domain.TaskTypeRemoveBG:
		baseCost = 0.01
	case domain.TaskTypeAnimation:
		baseCost = 0.05
	}

	// 如果有模型信息，使用模型定价
	if model != nil && req.ModelID != nil {
		pricingReq := &domain.PricingCalculationRequest{
			ModelID:    *req.ModelID,
			ImageCount: intPtr(1), // 假设生成1张图片
			UserType:   "basic",
		}

		pricingResp, err := s.modelService.CalculatePrice(ctx, pricingReq)
		if err == nil {
			return pricingResp.UserCost, nil
		}
	}

	return baseCost, nil
}

// createMediaAssetsFromTask 从任务创建媒体资产
func (s *mediaService) createMediaAssetsFromTask(ctx context.Context, task *domain.MediaGenerationTask) error {
	for i, url := range task.ResultURLs {
		asset := &domain.MediaAsset{
			ID:               uuid.New(),
			UserID:           task.UserID,
			TaskID:           &task.ID,
			AssetType:        s.getAssetTypeFromTaskType(task.TaskType),
			FileURL:          url,
			GenerationParams: task.ModelParams,
			PromptUsed:       task.TextPrompt,
			Title:            stringPtr(fmt.Sprintf("Generated %s %d", task.TaskType, i+1)),
			IsPublic:         false,
			IsFavorite:       false,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if task.ModelID != nil {
			model, err := s.modelService.GetModel(ctx, *task.ModelID)
			if err == nil {
				asset.ModelUsed = &model.DisplayName
			}
		}

		err := s.mediaRepo.CreateMediaAsset(ctx, asset)
		if err != nil {
			s.logger.WithError(err).WithField("task_id", task.ID).Warn("Failed to create media asset")
		}
	}

	return nil
}

// getAssetTypeFromTaskType 从任务类型获取资产类型
func (s *mediaService) getAssetTypeFromTaskType(taskType string) string {
	switch taskType {
	case domain.TaskTypeTxt2Img, domain.TaskTypeImg2Img, domain.TaskTypeRemoveBG:
		return domain.AssetTypeImage
	case domain.TaskTypeImg2Video, domain.TaskTypeTxt2Video, domain.TaskTypeAnimation:
		return domain.AssetTypeVideo
	default:
		return domain.AssetTypeImage
	}
}

// ProcessMediaTask 处理媒体任务（模拟实现）
func (s *mediaService) ProcessMediaTask(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.mediaRepo.GetMediaTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	// 更新状态为处理中
	task.Status = domain.TaskStatusProcessing
	now := time.Now()
	task.StartedAt = &now
	task.Progress = 10

	err = s.mediaRepo.UpdateMediaTask(ctx, task)
	if err != nil {
		return err
	}

	// 模拟处理过程
	time.Sleep(2 * time.Second)

	// 模拟生成结果
	resultURL := fmt.Sprintf("https://example.com/generated/%s.jpg", task.ID)
	task.ResultURLs = []string{resultURL}
	task.Status = domain.TaskStatusCompleted
	task.Progress = 100
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.ActualCost = task.EstimatedCost

	err = s.mediaRepo.UpdateMediaTask(ctx, task)
	if err != nil {
		return err
	}

	// 创建媒体资产
	return s.createMediaAssetsFromTask(ctx, task)
}

// 其他方法的实现...

func (s *mediaService) GetMediaAsset(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.MediaAssetResponse, error) {
	// TODO: 实现
	return nil, fmt.Errorf("not implemented")
}

func (s *mediaService) ListMediaAssets(ctx context.Context, req *domain.MediaAssetListRequest, viewerUserID *uuid.UUID) (*domain.MediaAssetListResponse, error) {
	// TODO: 实现
	return nil, fmt.Errorf("not implemented")
}

func (s *mediaService) UpdateMediaAsset(ctx context.Context, id uuid.UUID, userID uuid.UUID, title, description *string, tags []string, isFavorite, isPublic *bool) (*domain.MediaAsset, error) {
	// TODO: 实现
	return nil, fmt.Errorf("not implemented")
}

func (s *mediaService) DeleteMediaAsset(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// TODO: 实现
	return fmt.Errorf("not implemented")
}

func (s *mediaService) AutoGenerateCharacterImages(ctx context.Context, userID uuid.UUID, req *domain.AutoGenerateCharacterImagesRequest) (*domain.AutoGenerateCharacterImagesResponse, error) {
	// TODO: 实现
	return nil, fmt.Errorf("not implemented")
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
