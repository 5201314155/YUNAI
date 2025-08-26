package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// ModelService 模型服务接口
type ModelService interface {
	// 模型管理
	CreateModel(ctx context.Context, req *domain.CreateModelRequest) (*domain.AIModel, error)
	GetModel(ctx context.Context, id uuid.UUID) (*domain.ModelResponse, error)
	GetModelByKey(ctx context.Context, internalKey string) (*domain.ModelResponse, error)
	UpdateModel(ctx context.Context, id uuid.UUID, req *domain.UpdateModelRequest) (*domain.AIModel, error)
	DeleteModel(ctx context.Context, id uuid.UUID) error
	ListModels(ctx context.Context, req *domain.ModelListRequest) (*domain.ModelListResponse, error)

	// 健康检查
	TestConnectivity(ctx context.Context, req *domain.ConnectivityTestRequest) (*domain.ConnectivityTestResponse, error)
	GetModelHealth(ctx context.Context, modelID uuid.UUID) (*domain.ModelHealthResponse, error)
	PerformHealthCheck(ctx context.Context, modelID uuid.UUID) error

	// 能力管理
	GetModelCapabilities(ctx context.Context, modelID uuid.UUID) ([]string, error)
	UpdateModelCapabilities(ctx context.Context, modelID uuid.UUID, capabilities []string) error

	// 使用统计
	RecordUsage(ctx context.Context, modelID uuid.UUID, userID *uuid.UUID, success bool, tokens int64, cost float64) error
	GetUsageStats(ctx context.Context, modelID uuid.UUID, userID *uuid.UUID, days int) ([]*domain.ModelUsageStats, error)

	// 定价管理
	ConfigureModelPricing(ctx context.Context, operatorID uuid.UUID, req *domain.ModelPricingConfigRequest) (*domain.ModelPricingResponse, error)
	GetModelPricing(ctx context.Context, modelID uuid.UUID) (*domain.ModelPricingResponse, error)
	CalculatePrice(ctx context.Context, req *domain.PricingCalculationRequest) (*domain.PricingCalculationResponse, error)
	GetPriceHistory(ctx context.Context, modelID uuid.UUID, limit int) ([]domain.ModelPriceHistory, error)
}

type modelService struct {
	modelRepo repository.ModelRepository
	logger    *logrus.Logger
	encKey    []byte // 用于加密API密钥
}

// NewModelService 创建模型服务
func NewModelService(modelRepo repository.ModelRepository, logger *logrus.Logger) ModelService {
	// 在实际应用中，这个密钥应该从环境变量或配置文件中读取
	encKey := []byte("yunai-model-service-encrypt-key32") // 32字节密钥

	return &modelService{
		modelRepo: modelRepo,
		logger:    logger,
		encKey:    encKey,
	}
}

// CreateModel 创建模型
func (s *modelService) CreateModel(ctx context.Context, req *domain.CreateModelRequest) (*domain.AIModel, error) {
	// 检查内部键是否已存在
	existing, err := s.modelRepo.GetModelByInternalKey(ctx, req.InternalKey)
	if err == nil && existing != nil {
		return nil, domain.NewAppError(domain.CodeConflict, "模型内部键已存在", nil)
	}

	// 加密API密钥
	var encryptedAPIKey *string
	if req.APIKey != nil && *req.APIKey != "" {
		encrypted, err := s.encryptAPIKey(*req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		encryptedAPIKey = &encrypted
	}

	// 创建模型对象
	model := &domain.AIModel{
		ID:                       uuid.New(),
		InternalKey:              req.InternalKey,
		DisplayName:              req.DisplayName,
		Provider:                 req.Provider,
		ModelType:                req.ModelType,
		Capabilities:             req.Capabilities,
		ParamsSchema:             req.ParamsSchema,
		ModelSystemPrompt:        req.ModelSystemPrompt,
		BaseURL:                  req.BaseURL,
		APIKeyEncrypted:          encryptedAPIKey,
		Pricing:                  req.Pricing,
		Visibility:               req.Visibility,
		MinUserType:              req.MinUserType,
		MinBalance:               req.MinBalance,
		DailyLimit:               req.DailyLimit,
		UserLimit:                req.UserLimit,
		HealthStatus:             domain.HealthStatusUnknown,
		Weight:                   req.Weight,
		FallbackChain:            json.RawMessage("[]"),
		ConnectivityTestEndpoint: req.ConnectivityTestEndpoint,
		IsActive:                 true,
		IsFeatured:               false,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	// 保存到数据库
	if err := s.modelRepo.CreateModel(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// 创建能力记录
	if err := s.createCapabilitiesFromJSON(ctx, model.ID, req.Capabilities); err != nil {
		s.logger.WithError(err).Warn("Failed to create model capabilities")
	}

	s.logger.WithFields(logrus.Fields{
		"model_id":     model.ID,
		"internal_key": model.InternalKey,
		"provider":     model.Provider,
	}).Info("Model created successfully")

	return model, nil
}

// GetModel 获取模型
func (s *modelService) GetModel(ctx context.Context, id uuid.UUID) (*domain.ModelResponse, error) {
	model, err := s.modelRepo.GetModelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildModelResponse(ctx, model)
}

// GetModelByKey 根据内部键获取模型
func (s *modelService) GetModelByKey(ctx context.Context, internalKey string) (*domain.ModelResponse, error) {
	model, err := s.modelRepo.GetModelByInternalKey(ctx, internalKey)
	if err != nil {
		return nil, err
	}

	return s.buildModelResponse(ctx, model)
}

// UpdateModel 更新模型
func (s *modelService) UpdateModel(ctx context.Context, id uuid.UUID, req *domain.UpdateModelRequest) (*domain.AIModel, error) {
	// 获取现有模型
	model, err := s.modelRepo.GetModelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.DisplayName != nil {
		model.DisplayName = *req.DisplayName
	}
	if req.ModelSystemPrompt != nil {
		model.ModelSystemPrompt = req.ModelSystemPrompt
	}
	if req.BaseURL != nil {
		model.BaseURL = req.BaseURL
	}
	if req.APIKey != nil && *req.APIKey != "" {
		encrypted, err := s.encryptAPIKey(*req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		model.APIKeyEncrypted = &encrypted
	}
	if req.Pricing != nil {
		model.Pricing = req.Pricing
	}
	if req.Visibility != nil {
		model.Visibility = *req.Visibility
	}
	if req.MinUserType != nil {
		model.MinUserType = *req.MinUserType
	}
	if req.MinBalance != nil {
		model.MinBalance = *req.MinBalance
	}
	if req.DailyLimit != nil {
		model.DailyLimit = *req.DailyLimit
	}
	if req.UserLimit != nil {
		model.UserLimit = *req.UserLimit
	}
	if req.Weight != nil {
		model.Weight = *req.Weight
	}
	if req.ConnectivityTestEndpoint != nil {
		model.ConnectivityTestEndpoint = req.ConnectivityTestEndpoint
	}
	if req.IsActive != nil {
		model.IsActive = *req.IsActive
	}
	if req.IsFeatured != nil {
		model.IsFeatured = *req.IsFeatured
	}

	// 保存更新
	if err := s.modelRepo.UpdateModel(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to update model: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id":     model.ID,
		"internal_key": model.InternalKey,
	}).Info("Model updated successfully")

	return model, nil
}

// DeleteModel 删除模型
func (s *modelService) DeleteModel(ctx context.Context, id uuid.UUID) error {
	// 检查模型是否存在
	_, err := s.modelRepo.GetModelByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除模型
	if err := s.modelRepo.DeleteModel(ctx, id); err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	s.logger.WithField("model_id", id).Info("Model deleted successfully")
	return nil
}

// ListModels 获取模型列表
func (s *modelService) ListModels(ctx context.Context, req *domain.ModelListRequest) (*domain.ModelListResponse, error) {
	models, total, err := s.modelRepo.ListModels(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	// 构建响应
	var modelResponses []domain.ModelResponse
	for _, model := range models {
		resp, err := s.buildModelResponse(ctx, model)
		if err != nil {
			s.logger.WithError(err).WithField("model_id", model.ID).Warn("Failed to build model response")
			continue
		}
		modelResponses = append(modelResponses, *resp)
	}

	return &domain.ModelListResponse{
		Models: modelResponses,
		Total:  total,
		Page:   req.Page,
		Limit:  req.Limit,
	}, nil
}

// TestConnectivity 测试连接
func (s *modelService) TestConnectivity(ctx context.Context, req *domain.ConnectivityTestRequest) (*domain.ConnectivityTestResponse, error) {
	model, err := s.modelRepo.GetModelByID(ctx, req.ModelID)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	var errorMessage *string

	// 执行连接测试
	if model.ConnectivityTestEndpoint != nil && *model.ConnectivityTestEndpoint != "" {
		err = s.performConnectivityTest(*model.ConnectivityTestEndpoint)
		if err != nil {
			errMsg := err.Error()
			errorMessage = &errMsg
		}
	} else {
		errMsg := "未配置连接测试端点"
		errorMessage = &errMsg
	}

	responseTime := int(time.Since(startTime).Milliseconds())
	status := domain.HealthStatusHealthy
	if errorMessage != nil {
		status = domain.HealthStatusUnhealthy
	}

	// 记录健康日志
	healthLog := &domain.ModelHealthLog{
		ID:             uuid.New(),
		ModelID:        req.ModelID,
		Status:         status,
		ResponseTimeMs: &responseTime,
		ErrorMessage:   errorMessage,
		CheckedAt:      time.Now(),
	}

	if err := s.modelRepo.CreateHealthLog(ctx, healthLog); err != nil {
		s.logger.WithError(err).Warn("Failed to create health log")
	}

	// 更新模型健康状态
	if err := s.modelRepo.UpdateModelHealth(ctx, req.ModelID, status); err != nil {
		s.logger.WithError(err).Warn("Failed to update model health status")
	}

	return &domain.ConnectivityTestResponse{
		ModelID:        req.ModelID,
		Status:         status,
		ResponseTimeMs: responseTime,
		ErrorMessage:   errorMessage,
		TestedAt:       time.Now(),
	}, nil
}

// 辅助方法

// buildModelResponse 构建模型响应
func (s *modelService) buildModelResponse(ctx context.Context, model *domain.AIModel) (*domain.ModelResponse, error) {
	// 获取能力列表
	capabilities, err := s.GetModelCapabilities(ctx, model.ID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get model capabilities")
		capabilities = []string{}
	}

	// 判断是否可用
	isAvailable := model.IsActive && model.HealthStatus == domain.HealthStatusHealthy

	return &domain.ModelResponse{
		AIModel:        model,
		CapabilityList: capabilities,
		IsAvailable:    isAvailable,
	}, nil
}

// createCapabilitiesFromJSON 从JSON创建能力记录
func (s *modelService) createCapabilitiesFromJSON(ctx context.Context, modelID uuid.UUID, capabilitiesJSON json.RawMessage) error {
	var capabilities []string
	if err := json.Unmarshal(capabilitiesJSON, &capabilities); err != nil {
		return err
	}

	for _, capability := range capabilities {
		cap := &domain.ModelCapability{
			ID:         uuid.New(),
			ModelID:    modelID,
			Capability: capability,
			IsEnabled:  true,
			CreatedAt:  time.Now(),
		}

		if err := s.modelRepo.CreateModelCapability(ctx, cap); err != nil {
			s.logger.WithError(err).WithField("capability", capability).Warn("Failed to create model capability")
		}
	}

	return nil
}

// encryptAPIKey 加密API密钥
func (s *modelService) encryptAPIKey(apiKey string) (string, error) {
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// performConnectivityTest 执行连接测试
func (s *modelService) performConnectivityTest(endpoint string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// GetModelCapabilities 获取模型能力
func (s *modelService) GetModelCapabilities(ctx context.Context, modelID uuid.UUID) ([]string, error) {
	capabilities, err := s.modelRepo.GetModelCapabilities(ctx, modelID)
	if err != nil {
		return nil, err
	}

	var capList []string
	for _, cap := range capabilities {
		capList = append(capList, cap.Capability)
	}

	return capList, nil
}

// UpdateModelCapabilities 更新模型能力
func (s *modelService) UpdateModelCapabilities(ctx context.Context, modelID uuid.UUID, capabilities []string) error {
	// 删除现有能力
	if err := s.modelRepo.DeleteModelCapabilities(ctx, modelID); err != nil {
		return fmt.Errorf("failed to delete existing capabilities: %w", err)
	}

	// 创建新能力
	for _, capability := range capabilities {
		cap := &domain.ModelCapability{
			ID:         uuid.New(),
			ModelID:    modelID,
			Capability: capability,
			IsEnabled:  true,
			CreatedAt:  time.Now(),
		}

		if err := s.modelRepo.CreateModelCapability(ctx, cap); err != nil {
			return fmt.Errorf("failed to create capability %s: %w", capability, err)
		}
	}

	return nil
}

// PerformHealthCheck 执行健康检查
func (s *modelService) PerformHealthCheck(ctx context.Context, modelID uuid.UUID) error {
	req := &domain.ConnectivityTestRequest{ModelID: modelID}
	_, err := s.TestConnectivity(ctx, req)
	return err
}

// GetModelHealth 获取模型健康状态
func (s *modelService) GetModelHealth(ctx context.Context, modelID uuid.UUID) (*domain.ModelHealthResponse, error) {
	model, err := s.modelRepo.GetModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// 获取最近的健康日志
	logs, err := s.modelRepo.GetRecentHealthLogs(ctx, modelID, 10)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get recent health logs")
		logs = []*domain.ModelHealthLog{}
	}

	// 转换指针切片为值切片
	var recentLogs []domain.ModelHealthLog
	for _, log := range logs {
		if log != nil {
			recentLogs = append(recentLogs, *log)
		}
	}

	return &domain.ModelHealthResponse{
		ModelID:         modelID,
		HealthStatus:    model.HealthStatus,
		LastHealthCheck: model.LastHealthCheck,
		RecentLogs:      recentLogs,
	}, nil
}

// RecordUsage 记录使用统计
func (s *modelService) RecordUsage(ctx context.Context, modelID uuid.UUID, userID *uuid.UUID, success bool, tokens int64, cost float64) error {
	stats := &domain.ModelUsageStats{
		ID:           uuid.New(),
		ModelID:      modelID,
		UserID:       userID,
		Date:         time.Now().Truncate(24 * time.Hour),
		RequestCount: 1,
		SuccessCount: 0,
		ErrorCount:   0,
		TotalTokens:  tokens,
		TotalCost:    cost,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if success {
		stats.SuccessCount = 1
	} else {
		stats.ErrorCount = 1
	}

	return s.modelRepo.CreateOrUpdateUsageStats(ctx, stats)
}

// GetUsageStats 获取使用统计
func (s *modelService) GetUsageStats(ctx context.Context, modelID uuid.UUID, userID *uuid.UUID, days int) ([]*domain.ModelUsageStats, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	return s.modelRepo.GetUsageStats(ctx, modelID, userID, startDate, endDate)
}

// ConfigureModelPricing 配置模型定价
func (s *modelService) ConfigureModelPricing(ctx context.Context, operatorID uuid.UUID, req *domain.ModelPricingConfigRequest) (*domain.ModelPricingResponse, error) {
	// 获取模型
	model, err := s.modelRepo.GetModelByID(ctx, req.ModelID)
	if err != nil {
		return nil, err
	}

	// 解析当前定价
	var currentPricing domain.ModelPricing
	if err := json.Unmarshal(model.Pricing, &currentPricing); err != nil {
		s.logger.WithError(err).Warn("Failed to parse current pricing, using defaults")
		currentPricing = domain.ModelPricing{
			Unit:            domain.UnitTokens1K,
			Currency:        domain.CurrencyUSD,
			PriceMultiplier: domain.DefaultPriceMultiplier,
		}
	}

	// 保存原始定价作为历史记录
	originalPricingData, _ := json.Marshal(currentPricing)
	_ = originalPricingData // TODO: 保存到价格历史表
	// priceHistory := &domain.ModelPriceHistory{
	//     ID:          uuid.New(),
	//     ModelID:     req.ModelID,
	//     PricingData: originalPricingData,
	//     Reason:      "价格配置更新",
	//     OperatorID:  &operatorID,
	//     CreatedAt:   time.Now(),
	// }

	// 更新定价配置
	newPricing := domain.ModelPricing{
		InputTokenPrice:  req.InputTokenPrice,
		OutputTokenPrice: req.OutputTokenPrice,
		ImagePrice:       req.ImagePrice,
		VideoPrice:       req.VideoPrice,
		AudioPrice:       req.AudioPrice,
		Unit:             req.Unit,
		Currency:         req.Currency,
		PriceMultiplier:  req.PriceMultiplier,
		TierPricing:      req.TierPricing,
		MinCharge:        req.MinCharge,
	}

	// 序列化新定价
	newPricingData, err := json.Marshal(newPricing)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new pricing: %w", err)
	}

	// 更新模型定价
	model.Pricing = newPricingData
	model.UpdatedAt = time.Now()

	if err := s.modelRepo.UpdateModel(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to update model pricing: %w", err)
	}

	// TODO: 保存价格历史记录到数据库
	// if err := s.modelRepo.CreatePriceHistory(ctx, priceHistory); err != nil {
	//     s.logger.WithError(err).Warn("Failed to save price history")
	// }

	s.logger.WithFields(logrus.Fields{
		"model_id":    req.ModelID,
		"operator_id": operatorID,
		"multiplier":  req.PriceMultiplier,
	}).Info("Model pricing updated")

	// 构建响应
	return s.buildPricingResponse(ctx, model, currentPricing, newPricing)
}

// GetModelPricing 获取模型定价
func (s *modelService) GetModelPricing(ctx context.Context, modelID uuid.UUID) (*domain.ModelPricingResponse, error) {
	model, err := s.modelRepo.GetModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// 解析当前定价
	var currentPricing domain.ModelPricing
	if err := json.Unmarshal(model.Pricing, &currentPricing); err != nil {
		return nil, fmt.Errorf("failed to parse model pricing: %w", err)
	}

	// 构建原始定价（假设倍率为1.0时的价格）
	originalPricing := currentPricing
	if currentPricing.PriceMultiplier != 0 {
		if originalPricing.InputTokenPrice != nil {
			originalPrice := *originalPricing.InputTokenPrice / currentPricing.PriceMultiplier
			originalPricing.InputTokenPrice = &originalPrice
		}
		if originalPricing.OutputTokenPrice != nil {
			originalPrice := *originalPricing.OutputTokenPrice / currentPricing.PriceMultiplier
			originalPricing.OutputTokenPrice = &originalPrice
		}
		if originalPricing.ImagePrice != nil {
			originalPrice := *originalPricing.ImagePrice / currentPricing.PriceMultiplier
			originalPricing.ImagePrice = &originalPrice
		}
		if originalPricing.VideoPrice != nil {
			originalPrice := *originalPricing.VideoPrice / currentPricing.PriceMultiplier
			originalPricing.VideoPrice = &originalPrice
		}
		if originalPricing.AudioPrice != nil {
			originalPrice := *originalPricing.AudioPrice / currentPricing.PriceMultiplier
			originalPricing.AudioPrice = &originalPrice
		}
	}
	originalPricing.PriceMultiplier = 1.0

	return s.buildPricingResponse(ctx, model, originalPricing, currentPricing)
}

// CalculatePrice 计算价格
func (s *modelService) CalculatePrice(ctx context.Context, req *domain.PricingCalculationRequest) (*domain.PricingCalculationResponse, error) {
	// 获取模型定价
	pricingResp, err := s.GetModelPricing(ctx, req.ModelID)
	if err != nil {
		return nil, err
	}

	pricing := pricingResp.CurrentPricing
	originalPricing := pricingResp.OriginalPricing

	// 计算原始成本
	var originalCost float64
	var platformCost float64
	costBreakdown := make(map[string]float64)

	// Token 成本计算
	if req.InputTokens != nil && originalPricing.InputTokenPrice != nil {
		tokenCost := float64(*req.InputTokens) / 1000.0 * *originalPricing.InputTokenPrice
		originalCost += tokenCost
		costBreakdown["input_tokens"] = tokenCost
	}

	if req.OutputTokens != nil && originalPricing.OutputTokenPrice != nil {
		tokenCost := float64(*req.OutputTokens) / 1000.0 * *originalPricing.OutputTokenPrice
		originalCost += tokenCost
		costBreakdown["output_tokens"] = tokenCost
	}

	// 图片成本计算
	if req.ImageCount != nil && originalPricing.ImagePrice != nil {
		imageCost := float64(*req.ImageCount) * *originalPricing.ImagePrice
		originalCost += imageCost
		costBreakdown["images"] = imageCost
	}

	// 视频成本计算
	if req.VideoSeconds != nil && originalPricing.VideoPrice != nil {
		videoCost := float64(*req.VideoSeconds) * *originalPricing.VideoPrice
		originalCost += videoCost
		costBreakdown["video"] = videoCost
	}

	// 音频成本计算
	if req.AudioSeconds != nil && originalPricing.AudioPrice != nil {
		audioCost := float64(*req.AudioSeconds) * *originalPricing.AudioPrice
		originalCost += audioCost
		costBreakdown["audio"] = audioCost
	}

	// 应用平台倍率
	platformCost = originalCost * pricing.PriceMultiplier

	// 应用最小消费
	if platformCost < pricing.MinCharge {
		platformCost = pricing.MinCharge
	}

	// AI调用直接按平台价格扣除金币，无折扣
	userCost := platformCost

	return &domain.PricingCalculationResponse{
		ModelID:       req.ModelID,
		OriginalCost:  originalCost,
		PlatformCost:  platformCost,
		UserCost:      userCost,
		CostBreakdown: costBreakdown,
		Currency:      pricing.Currency,
		Multiplier:    pricing.PriceMultiplier,
		Discount:      0.0, // AI调用无折扣
	}, nil
}

// GetPriceHistory 获取价格历史
func (s *modelService) GetPriceHistory(ctx context.Context, modelID uuid.UUID, limit int) ([]domain.ModelPriceHistory, error) {
	// TODO: 从数据库获取价格历史
	// return s.modelRepo.GetPriceHistory(ctx, modelID, limit)

	// 暂时返回空列表
	return []domain.ModelPriceHistory{}, nil
}

// 辅助方法

// buildPricingResponse 构建定价响应
func (s *modelService) buildPricingResponse(ctx context.Context, model *domain.AIModel, originalPricing, currentPricing domain.ModelPricing) (*domain.ModelPricingResponse, error) {
	// 获取价格历史
	priceHistory, err := s.GetPriceHistory(ctx, model.ID, 10)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get price history")
		priceHistory = []domain.ModelPriceHistory{}
	}

	// 计算预估成本
	estimatedCost := make(map[string]float64)

	// 常见使用场景的预估成本
	if currentPricing.InputTokenPrice != nil && currentPricing.OutputTokenPrice != nil {
		// 1000 tokens 对话成本
		chatCost := *currentPricing.InputTokenPrice + *currentPricing.OutputTokenPrice
		estimatedCost["1k_tokens_chat"] = chatCost

		// 10000 tokens 长文本成本
		longTextCost := (*currentPricing.InputTokenPrice + *currentPricing.OutputTokenPrice) * 10
		estimatedCost["10k_tokens_text"] = longTextCost
	}

	if currentPricing.ImagePrice != nil {
		estimatedCost["single_image"] = *currentPricing.ImagePrice
		estimatedCost["10_images"] = *currentPricing.ImagePrice * 10
	}

	if currentPricing.VideoPrice != nil {
		estimatedCost["1_minute_video"] = *currentPricing.VideoPrice * 60
		estimatedCost["10_minute_video"] = *currentPricing.VideoPrice * 600
	}

	return &domain.ModelPricingResponse{
		ModelID:         model.ID,
		ModelName:       model.DisplayName,
		Provider:        model.Provider,
		OriginalPricing: originalPricing,
		CurrentPricing:  currentPricing,
		PriceHistory:    priceHistory,
		EstimatedCost:   estimatedCost,
	}, nil
}

// AI调用统一按平台价格扣除金币，无用户等级折扣
// 折扣只适用于充值购买，不适用于AI服务调用
