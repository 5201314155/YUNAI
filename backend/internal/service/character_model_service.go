package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// CharacterModelService 角色模型配置服务
// 每个角色都可以选择不同的模型，不选择就默认自动获取模型
type CharacterModelService interface {
	// 设置角色专用模型
	SetCharacterModel(ctx context.Context, req *SetCharacterModelRequest) error

	// 获取角色使用的模型（优先级：角色专用 > 用户默认 > 系统默认）
	GetCharacterModel(ctx context.Context, characterID uuid.UUID, functionType string) (*domain.UniversalModelConfig, error)

	// 获取角色所有模型配置
	GetCharacterModelConfigs(ctx context.Context, characterID uuid.UUID) (*CharacterModelConfigs, error)

	// 删除角色模型配置（回退到默认）
	RemoveCharacterModel(ctx context.Context, characterID uuid.UUID, functionType string) error

	// 批量设置角色模型
	BatchSetCharacterModels(ctx context.Context, req *BatchSetCharacterModelsRequest) error

	// 获取角色模型使用统计
	GetCharacterModelStats(ctx context.Context, characterID uuid.UUID) (*CharacterModelStats, error)
}

type characterModelService struct {
	universalModelService UniversalModelService
	// userModelService      UserModelConfigService // 暂时注释掉
	characterRepo repository.CharacterRepository
	logger        *logrus.Logger
}

// UserModelConfigService 用户模型配置服务占位接口（用于类型对齐）
// 如后续需要可补充具体方法签名
type UserModelConfigService interface{}

// NewCharacterModelService 创建角色模型配置服务
func NewCharacterModelService(
	universalModelService UniversalModelService,
	userModelService UserModelConfigService,
	characterRepo repository.CharacterRepository,
	logger *logrus.Logger,
) CharacterModelService {
	return &characterModelService{
		universalModelService: universalModelService,
		// userModelService:      userModelService, // 暂时注释掉
		characterRepo: characterRepo,
		logger:        logger,
	}
}

// SetCharacterModelRequest 设置角色模型请求
type SetCharacterModelRequest struct {
	CharacterID  uuid.UUID `json:"character_id"`
	FunctionType string    `json:"function_type"` // chat, moments, story, emotion, voice
	ModelID      uuid.UUID `json:"model_id"`
	Priority     int       `json:"priority"`    // 优先级
	AutoSwitch   bool      `json:"auto_switch"` // 是否自动切换到最优模型
	Reason       string    `json:"reason"`      // 选择原因
}

// CharacterModelConfigs 角色模型配置
type CharacterModelConfigs struct {
	CharacterID    uuid.UUID                               `json:"character_id"`
	CharacterName  string                                  `json:"character_name"`
	FunctionModels map[string]*domain.UniversalModelConfig `json:"function_models"`
	DefaultModels  map[string]*domain.UniversalModelConfig `json:"default_models"`
	LastUpdated    time.Time                               `json:"last_updated"`
	ConfigSource   map[string]string                       `json:"config_source"` // character, user, system
}

// BatchSetCharacterModelsRequest 批量设置角色模型请求
type BatchSetCharacterModelsRequest struct {
	CharacterID uuid.UUID               `json:"character_id"`
	Models      map[string]uuid.UUID    `json:"models"` // function_type -> model_id
	Settings    *CharacterModelSettings `json:"settings"`
}

// CharacterModelSettings 角色模型设置
type CharacterModelSettings struct {
	AutoOptimize     bool                   `json:"auto_optimize"`     // 自动优化模型选择
	CostLimit        float64                `json:"cost_limit"`        // 成本限制
	PerformanceLevel string                 `json:"performance_level"` // high, medium, low
	Preferences      map[string]interface{} `json:"preferences"`
}

// CharacterModelStats 角色模型使用统计
type CharacterModelStats struct {
	CharacterID     uuid.UUID                    `json:"character_id"`
	TotalRequests   int64                        `json:"total_requests"`
	ModelUsage      map[string]*ModelUsageDetail `json:"model_usage"`
	CostAnalysis    *ModelCostAnalysis           `json:"cost_analysis"`
	Performance     *ModelPerformanceStats       `json:"performance"`
	Recommendations []string                     `json:"recommendations"`
}

// ModelUsageDetail 模型使用详情
type ModelUsageDetail struct {
	ModelName       string    `json:"model_name"`
	Provider        string    `json:"provider"`
	RequestCount    int64     `json:"request_count"`
	SuccessRate     float64   `json:"success_rate"`
	AvgResponseTime int64     `json:"avg_response_time"`
	TotalCost       float64   `json:"total_cost"`
	LastUsed        time.Time `json:"last_used"`
	Quality         float64   `json:"quality"` // 生成质量评分
}

// CharacterModelCostAnalysis 角色模型成本分析
type CharacterModelCostAnalysis struct {
	TotalCost      float64            `json:"total_cost"`
	CostByFunction map[string]float64 `json:"cost_by_function"`
	CostByModel    map[string]float64 `json:"cost_by_model"`
	EstimatedDaily float64            `json:"estimated_daily"`
	CostTrend      []float64          `json:"cost_trend"`
}

// ModelPerformanceStats 模型性能统计
type ModelPerformanceStats struct {
	AvgResponseTime   int64              `json:"avg_response_time"`
	SuccessRate       float64            `json:"success_rate"`
	QualityScore      float64            `json:"quality_score"`
	UserSatisfaction  float64            `json:"user_satisfaction"`
	PerformanceByFunc map[string]float64 `json:"performance_by_func"`
}

// SetCharacterModel 设置角色专用模型
func (s *characterModelService) SetCharacterModel(ctx context.Context, req *SetCharacterModelRequest) error {
	s.logger.WithFields(logrus.Fields{
		"character_id":  req.CharacterID,
		"function_type": req.FunctionType,
		"model_id":      req.ModelID,
	}).Info("Setting character-specific model")

	// 1. 验证角色是否存在
	character, err := s.characterRepo.GetCharacterByID(ctx, req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// 2. 验证模型是否存在且可用
	model, err := s.universalModelService.GetModel(ctx, req.ModelID)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	if model.Status != "active" {
		return fmt.Errorf("model %s is not active", model.Name)
	}

	// 3. 验证模型是否适合该功能
	if !s.isModelSuitableForFunction(model, req.FunctionType) {
		return fmt.Errorf("model %s is not suitable for function %s", model.Name, req.FunctionType)
	}

	// 4. 保存角色模型配置
	err = s.saveCharacterModelConfig(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to save character model config: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"character_name": character.Name,
		"function_type":  req.FunctionType,
		"model_name":     model.Name,
		"reason":         req.Reason,
	}).Info("Character model configuration updated successfully")

	return nil
}

// GetCharacterModel 获取角色使用的模型（智能优先级选择）
func (s *characterModelService) GetCharacterModel(ctx context.Context, characterID uuid.UUID, functionType string) (*domain.UniversalModelConfig, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id":  characterID,
		"function_type": functionType,
	}).Debug("Getting character model with priority selection")

	// 优先级1: 角色专用模型配置
	characterModel, err := s.getCharacterSpecificModel(ctx, characterID, functionType)
	if err == nil && characterModel != nil {
		s.logger.WithField("source", "character_specific").Debug("Using character-specific model")
		return characterModel, nil
	}

	// 优先级2: 获取角色的创建者（用户）的默认模型
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err == nil && character.UserID != uuid.Nil {
		// userModel, err := s.userModelService.GetUserFunctionModel(ctx, character.UserID, functionType)
		// if err == nil && userModel != nil {
		//	s.logger.WithField("source", "user_default").Debug("Using user default model")
		//	return userModel, nil
		// }
	}

	// 优先级3: 系统推荐的最优模型
	recommendations, err := s.getSystemRecommendedModels(ctx, characterID, functionType)
	if err == nil && len(recommendations) > 0 {
		s.logger.WithField("source", "system_recommended").Debug("Using system recommended model")
		return recommendations[0].Model, nil
	}

	// 优先级4: 系统默认模型
	defaultModel, err := s.getSystemDefaultModel(ctx, functionType)
	if err == nil && defaultModel != nil {
		s.logger.WithField("source", "system_default").Debug("Using system default model")
		return defaultModel, nil
	}

	// 优先级5: 任意可用的合适模型
	availableModel, err := s.getAnyAvailableModelForFunction(ctx, functionType)
	if err != nil {
		return nil, fmt.Errorf("no available model found for character %s function %s: %w", characterID, functionType, err)
	}

	s.logger.WithField("source", "fallback").Debug("Using fallback available model")
	return availableModel, nil
}

// GetCharacterModelConfigs 获取角色所有模型配置
func (s *characterModelService) GetCharacterModelConfigs(ctx context.Context, characterID uuid.UUID) (*CharacterModelConfigs, error) {
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	configs := &CharacterModelConfigs{
		CharacterID:    characterID,
		CharacterName:  character.Name,
		FunctionModels: make(map[string]*domain.UniversalModelConfig),
		DefaultModels:  make(map[string]*domain.UniversalModelConfig),
		ConfigSource:   make(map[string]string),
		LastUpdated:    time.Now(),
	}

	// 获取所有功能类型的模型配置
	functionTypes := []string{"chat", "moments", "story", "emotion", "voice", "embedding"}

	for _, funcType := range functionTypes {
		// 获取角色专用模型
		characterModel, err := s.getCharacterSpecificModel(ctx, characterID, funcType)
		if err == nil && characterModel != nil {
			configs.FunctionModels[funcType] = characterModel
			configs.ConfigSource[funcType] = "character"
		} else {
			// 获取默认模型
			defaultModel, err := s.GetCharacterModel(ctx, characterID, funcType)
			if err == nil && defaultModel != nil {
				configs.DefaultModels[funcType] = defaultModel
				configs.ConfigSource[funcType] = s.determineModelSource(ctx, characterID, funcType)
			}
		}
	}

	return configs, nil
}

// 辅助方法
func (s *characterModelService) isModelSuitableForFunction(model *domain.UniversalModelConfig, functionType string) bool {
	switch functionType {
	case "chat", "moments", "story", "emotion":
		return model.ModelType == "chat" || model.ModelType == "text"
	case "voice":
		return model.ModelType == "tts" || model.ModelType == "audio"
	case "embedding":
		return model.ModelType == "embedding"
	case "image":
		return model.ModelType == "image"
	default:
		return false
	}
}

func (s *characterModelService) saveCharacterModelConfig(ctx context.Context, req *SetCharacterModelRequest) error {
	// TODO: 实现数据库保存逻辑
	// 这里应该保存到 character_model_configs 表
	return nil
}

func (s *characterModelService) getCharacterSpecificModel(ctx context.Context, characterID uuid.UUID, functionType string) (*domain.UniversalModelConfig, error) {
	// TODO: 实现从数据库查询角色专用模型配置
	return nil, fmt.Errorf("no character-specific model configured")
}

type ModelRecommendation struct {
	Model  *domain.UniversalModelConfig
	Reason string
}

func (s *characterModelService) getSystemRecommendedModels(ctx context.Context, characterID uuid.UUID, functionType string) ([]*ModelRecommendation, error) {
	// TODO: 实现基于角色特点的智能模型推荐
	return nil, fmt.Errorf("no recommendations available")
}

func (s *characterModelService) getSystemDefaultModel(ctx context.Context, functionType string) (*domain.UniversalModelConfig, error) {
	models, err := s.universalModelService.ListModels(ctx, &ModelFilter{
		ModelType: s.getModelTypeForFunction(functionType),
		Status:    "active",
	})
	if err != nil {
		return nil, err
	}

	if len(models) > 0 {
		return models[0], nil
	}

	return nil, fmt.Errorf("no default model found for function %s", functionType)
}

func (s *characterModelService) getAnyAvailableModelForFunction(ctx context.Context, functionType string) (*domain.UniversalModelConfig, error) {
	return s.getSystemDefaultModel(ctx, functionType)
}

func (s *characterModelService) getModelTypeForFunction(functionType string) string {
	switch functionType {
	case "chat", "moments", "story", "emotion":
		return "chat"
	case "voice":
		return "tts"
	case "embedding":
		return "embedding"
	case "image":
		return "image"
	default:
		return "chat"
	}
}

func (s *characterModelService) determineModelSource(ctx context.Context, characterID uuid.UUID, functionType string) string {
	// 判断模型来源：character, user, system
	if _, err := s.getCharacterSpecificModel(ctx, characterID, functionType); err == nil {
		return "character"
	}

	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err == nil && character.UserID != uuid.Nil {
		// if _, err := s.userModelService.GetUserFunctionModel(ctx, character.UserID, functionType); err == nil {
		//	return "user"
		// }
	}

	return "system"
}

// RemoveCharacterModel 删除角色模型配置
func (s *characterModelService) RemoveCharacterModel(ctx context.Context, characterID uuid.UUID, functionType string) error {
	// TODO: 实现删除角色模型配置
	return nil
}

// BatchSetCharacterModels 批量设置角色模型
func (s *characterModelService) BatchSetCharacterModels(ctx context.Context, req *BatchSetCharacterModelsRequest) error {
	// TODO: 实现批量设置
	return nil
}

// GetCharacterModelStats 获取角色模型使用统计
func (s *characterModelService) GetCharacterModelStats(ctx context.Context, characterID uuid.UUID) (*CharacterModelStats, error) {
	// TODO: 实现统计查询
	return &CharacterModelStats{
		CharacterID:   characterID,
		TotalRequests: 0,
		ModelUsage:    make(map[string]*ModelUsageDetail),
	}, nil
}
