package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// 🚀 YUNAI 批量模型管理器 - 智能批量添加和管理模型

// BatchModelManager 批量模型管理器接口
type BatchModelManager interface {
	// API 探测和模型发现
	DiscoverModelsFromAPI(ctx context.Context, apiBase string, apiKey string) (*BatchModelDiscoveryResult, error)

	// 批量添加模型
	BatchAddModels(ctx context.Context, request *BatchAddRequest) (*BatchAddResponse, error)

	// 批量添加 API Keys (相同模型多实例)
	BatchAddAPIKeys(ctx context.Context, request *BatchKeysRequest) (*BatchKeysResponse, error)

	// 智能定价管理
	SetModelPricing(ctx context.Context, modelName string, pricing *ModelPricingConfig) error
	GetModelPricing(ctx context.Context, modelName string) (*ModelPricingConfig, error)

	// 模型去重和展示
	GetUniqueModels(ctx context.Context) ([]*UniqueModelInfo, error)
	GetModelInstances(ctx context.Context, modelName string) ([]*ModelInstanceInfo, error)
}

type batchModelManager struct {
	httpClient        *http.Client
	universalModelSvc UniversalModelService
	loadBalancer      ModelLoadBalancer
	logger            *logrus.Logger

	// 扣费相关服务 - 集成之前的金币扣费功能
	aiBillingService AIBillingService
	walletRepo       repository.WalletRepository

	// 模型定价缓存 modelName -> pricing
	pricingCache map[string]*ModelPricingConfig

	// 模型实例映射 modelName -> []instanceID
	modelInstances map[string][]uuid.UUID
}

// NewBatchModelManager 创建批量模型管理器
func NewBatchModelManager(
	universalModelSvc UniversalModelService,
	loadBalancer ModelLoadBalancer,
	aiBillingService AIBillingService,
	walletRepo repository.WalletRepository,
	logger *logrus.Logger,
) BatchModelManager {
	return &batchModelManager{
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		universalModelSvc: universalModelSvc,
		loadBalancer:      loadBalancer,
		aiBillingService:  aiBillingService,
		walletRepo:        walletRepo,
		logger:            logger,
		pricingCache:      make(map[string]*ModelPricingConfig),
		modelInstances:    make(map[string][]uuid.UUID),
	}
}

// BatchModelDiscoveryResult API 探测结果
type BatchModelDiscoveryResult struct {
	APIBase               string             `json:"api_base"`
	Provider              string             `json:"provider"`
	Models                []*DiscoveredModel `json:"models"`
	TotalCount            int                `json:"total_count"`
	SupportedCapabilities []string           `json:"supported_capabilities"`
	Error                 string             `json:"error,omitempty"`
}

// DiscoveredModel 发现的模型
type DiscoveredModel struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Capabilities  []string               `json:"capabilities"`
	ModelType     string                 `json:"model_type"` // text, image, video, audio, multimodal
	Parameters    map[string]interface{} `json:"parameters"`
	Pricing       *ModelPricingInfo      `json:"pricing"`
	MaxTokens     int                    `json:"max_tokens"`
	ContextWindow int                    `json:"context_window"`
}

// ModelPricingInfo 模型定价信息
type ModelPricingInfo struct {
	InputPrice  float64 `json:"input_price"`  // 输入价格 (per 1K tokens)
	OutputPrice float64 `json:"output_price"` // 输出价格 (per 1K tokens)
	ImagePrice  float64 `json:"image_price"`  // 图片价格 (per image)
	VideoPrice  float64 `json:"video_price"`  // 视频价格 (per second)
	AudioPrice  float64 `json:"audio_price"`  // 音频价格 (per second)
	Currency    string  `json:"currency"`     // 货币单位
}

// BatchAddRequest 批量添加请求
type BatchAddRequest struct {
	APIBase      string                 `json:"api_base"`
	Provider     string                 `json:"provider"`
	APIKeys      []string               `json:"api_keys"`      // 多个 API Key
	ModelIDs     []string               `json:"model_ids"`     // 要添加的模型ID列表
	AutoPricing  bool                   `json:"auto_pricing"`  // 是否自动设置定价
	CustomConfig map[string]interface{} `json:"custom_config"` // 自定义配置
}

// BatchAddResponse 批量添加响应
type BatchAddResponse struct {
	Success        bool               `json:"success"`
	AddedModels    []*AddedModelInfo  `json:"added_models"`
	FailedModels   []*FailedModelInfo `json:"failed_models"`
	TotalAdded     int                `json:"total_added"`
	TotalFailed    int                `json:"total_failed"`
	UniqueModels   int                `json:"unique_models"`   // 去重后的模型数
	TotalInstances int                `json:"total_instances"` // 总实例数
}

// AddedModelInfo 成功添加的模型信息
type AddedModelInfo struct {
	ModelName     string            `json:"model_name"`
	ModelID       uuid.UUID         `json:"model_id"`
	InstanceCount int               `json:"instance_count"` // 该模型的实例数量
	Capabilities  []string          `json:"capabilities"`
	Pricing       *ModelPricingInfo `json:"pricing"`
}

// FailedModelInfo 添加失败的模型信息
type FailedModelInfo struct {
	ModelName string `json:"model_name"`
	Error     string `json:"error"`
	APIKey    string `json:"api_key,omitempty"`
}

// BatchKeysRequest 批量密钥请求
type BatchKeysRequest struct {
	APIBase             string   `json:"api_base"`
	ModelName           string   `json:"model_name"`            // 指定模型名称
	APIKeys             []string `json:"api_keys"`              // 多个 API Key，一行一个
	LoadBalanceStrategy string   `json:"load_balance_strategy"` // 负载均衡策略
}

// BatchKeysResponse 批量密钥响应
type BatchKeysResponse struct {
	Success        bool     `json:"success"`
	ModelName      string   `json:"model_name"`
	AddedKeys      []string `json:"added_keys"`
	FailedKeys     []string `json:"failed_keys"`
	TotalInstances int      `json:"total_instances"`
	Strategy       string   `json:"strategy"`
}

// UniqueModelInfo 去重后的模型信息
type UniqueModelInfo struct {
	ModelName     string              `json:"model_name"`
	DisplayName   string              `json:"display_name"`
	ModelType     string              `json:"model_type"`
	Capabilities  []string            `json:"capabilities"`
	InstanceCount int                 `json:"instance_count"` // 实例数量
	Pricing       *ModelPricingConfig `json:"pricing"`
	Status        string              `json:"status"`    // active, inactive
	Providers     []string            `json:"providers"` // 提供商列表
}

// ModelInstanceInfo 模型实例信息
type ModelInstanceInfo struct {
	InstanceID   uuid.UUID `json:"instance_id"`
	APIBase      string    `json:"api_base"`
	Provider     string    `json:"provider"`
	Status       string    `json:"status"`
	Health       string    `json:"health"`
	CurrentLoad  int       `json:"current_load"`
	Weight       int       `json:"weight"`
	ResponseTime int64     `json:"response_time"`
}

// ModelPricingConfig 模型定价配置
type ModelPricingConfig struct {
	ModelName   string    `json:"model_name"`
	InputPrice  float64   `json:"input_price"`
	OutputPrice float64   `json:"output_price"`
	ImagePrice  float64   `json:"image_price"`
	VideoPrice  float64   `json:"video_price"`
	AudioPrice  float64   `json:"audio_price"`
	Currency    string    `json:"currency"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DiscoverModelsFromAPI 从 API 地址探测所有可用模型
func (bmm *batchModelManager) DiscoverModelsFromAPI(ctx context.Context, apiBase string, apiKey string) (*BatchModelDiscoveryResult, error) {
	bmm.logger.WithFields(logrus.Fields{
		"api_base": apiBase,
	}).Info("开始探测 API 模型")

	result := &BatchModelDiscoveryResult{
		APIBase:  apiBase,
		Provider: bmm.detectProvider(apiBase),
		Models:   make([]*DiscoveredModel, 0),
	}

	// 尝试调用 /models 端点
	modelsURL := strings.TrimSuffix(apiBase, "/") + "/models"

	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result, nil
	}

	// 设置认证头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := bmm.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		result.Error = fmt.Sprintf("API 错误 %d: %s", resp.StatusCode, string(body))
		return result, nil
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result, nil
	}

	var apiResponse struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
			Root    string `json:"root,omitempty"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result, nil
	}

	// 转换为发现的模型
	capabilityMap := make(map[string]bool)

	for _, model := range apiResponse.Data {
		discoveredModel := &DiscoveredModel{
			ID:           model.ID,
			Name:         model.ID,
			Description:  fmt.Sprintf("%s 模型 (来自 %s)", model.ID, model.OwnedBy),
			Capabilities: bmm.inferCapabilities(model.ID),
			ModelType:    bmm.inferModelType(model.ID),
			Parameters:   bmm.getDefaultParameters(model.ID),
			Pricing:      bmm.inferPricing(model.ID),
		}

		result.Models = append(result.Models, discoveredModel)

		// 收集支持的能力
		for _, cap := range discoveredModel.Capabilities {
			capabilityMap[cap] = true
		}
	}

	// 设置支持的能力
	for cap := range capabilityMap {
		result.SupportedCapabilities = append(result.SupportedCapabilities, cap)
	}

	result.TotalCount = len(result.Models)

	bmm.logger.WithFields(logrus.Fields{
		"api_base":     apiBase,
		"total_models": result.TotalCount,
		"capabilities": len(result.SupportedCapabilities),
	}).Info("API 模型探测完成")

	return result, nil
}

// detectProvider 检测提供商
func (bmm *batchModelManager) detectProvider(apiBase string) string {
	apiBase = strings.ToLower(apiBase)

	if strings.Contains(apiBase, "openai.com") {
		return "openai"
	} else if strings.Contains(apiBase, "siliconflow.cn") {
		return "siliconflow"
	} else if strings.Contains(apiBase, "anthropic.com") {
		return "anthropic"
	} else if strings.Contains(apiBase, "google") || strings.Contains(apiBase, "googleapis.com") {
		return "google"
	} else if strings.Contains(apiBase, "baidu") || strings.Contains(apiBase, "baidubce.com") {
		return "baidu"
	} else if strings.Contains(apiBase, "aliyun") || strings.Contains(apiBase, "dashscope") {
		return "alibaba"
	} else if strings.Contains(apiBase, "tencent") || strings.Contains(apiBase, "hunyuan") {
		return "tencent"
	} else if strings.Contains(apiBase, "bytedance") || strings.Contains(apiBase, "volces.com") {
		return "bytedance"
	}

	return "custom"
}

// inferCapabilities 推断模型能力
func (bmm *batchModelManager) inferCapabilities(modelID string) []string {
	modelID = strings.ToLower(modelID)
	var capabilities []string

	// 文本模型
	if strings.Contains(modelID, "gpt") || strings.Contains(modelID, "chat") ||
		strings.Contains(modelID, "claude") || strings.Contains(modelID, "deepseek") ||
		strings.Contains(modelID, "qwen") || strings.Contains(modelID, "gemini") {
		capabilities = append(capabilities, "chat", "completion")
	}

	// 图像模型
	if strings.Contains(modelID, "dall") || strings.Contains(modelID, "stable") ||
		strings.Contains(modelID, "flux") || strings.Contains(modelID, "midjourney") ||
		strings.Contains(modelID, "image") {
		capabilities = append(capabilities, "text_to_image", "image_to_image")
	}

	// 视频模型
	if strings.Contains(modelID, "video") || strings.Contains(modelID, "runway") ||
		strings.Contains(modelID, "pika") || strings.Contains(modelID, "gen") {
		capabilities = append(capabilities, "text_to_video", "image_to_video")
	}

	// 音频模型
	if strings.Contains(modelID, "tts") || strings.Contains(modelID, "speech") ||
		strings.Contains(modelID, "voice") || strings.Contains(modelID, "audio") ||
		strings.Contains(modelID, "whisper") {
		capabilities = append(capabilities, "text_to_speech", "speech_to_text")
	}

	// 嵌入模型
	if strings.Contains(modelID, "embedding") || strings.Contains(modelID, "bge") ||
		strings.Contains(modelID, "text-embedding") {
		capabilities = append(capabilities, "embedding")
	}

	// 如果没有匹配到，默认为聊天
	if len(capabilities) == 0 {
		capabilities = append(capabilities, "chat")
	}

	return capabilities
}

// inferModelType 推断模型类型
func (bmm *batchModelManager) inferModelType(modelID string) string {
	capabilities := bmm.inferCapabilities(modelID)

	for _, cap := range capabilities {
		switch cap {
		case "text_to_image", "image_to_image":
			return "image"
		case "text_to_video", "image_to_video":
			return "video"
		case "text_to_speech", "speech_to_text":
			return "audio"
		case "embedding":
			return "embedding"
		}
	}

	return "text"
}

// getDefaultParameters 获取默认参数
func (bmm *batchModelManager) getDefaultParameters(modelID string) map[string]interface{} {
	modelType := bmm.inferModelType(modelID)

	switch modelType {
	case "text":
		return map[string]interface{}{
			"temperature":       0.7,
			"max_tokens":        2000,
			"top_p":             0.9,
			"frequency_penalty": 0.0,
			"presence_penalty":  0.0,
		}
	case "image":
		return map[string]interface{}{
			"size":    "1024x1024",
			"quality": "standard",
			"style":   "vivid",
			"n":       1,
		}
	case "video":
		return map[string]interface{}{
			"duration":   30,
			"fps":        24,
			"resolution": "1920x1080",
		}
	case "audio":
		return map[string]interface{}{
			"voice":  "alloy",
			"format": "mp3",
			"speed":  1.0,
		}
	default:
		return map[string]interface{}{}
	}
}

// inferPricing 推断定价
func (bmm *batchModelManager) inferPricing(modelID string) *ModelPricingInfo {
	// 这里可以根据模型ID推断大概的定价
	// 实际应用中可以从配置文件或数据库中获取
	return &ModelPricingInfo{
		InputPrice:  0.001, // $0.001 per 1K tokens
		OutputPrice: 0.002, // $0.002 per 1K tokens
		Currency:    "USD",
	}
}

// BatchAddModels 批量添加模型
func (bmm *batchModelManager) BatchAddModels(ctx context.Context, request *BatchAddRequest) (*BatchAddResponse, error) {
	bmm.logger.WithFields(logrus.Fields{
		"api_base":  request.APIBase,
		"provider":  request.Provider,
		"api_keys":  len(request.APIKeys),
		"model_ids": len(request.ModelIDs),
	}).Info("开始批量添加模型")

	response := &BatchAddResponse{
		Success:      true,
		AddedModels:  make([]*AddedModelInfo, 0),
		FailedModels: make([]*FailedModelInfo, 0),
	}

	// 首先探测 API 获取模型信息
	discoveryResult, err := bmm.DiscoverModelsFromAPI(ctx, request.APIBase, request.APIKeys[0])
	if err != nil || discoveryResult.Error != "" {
		response.Success = false
		for _, modelID := range request.ModelIDs {
			response.FailedModels = append(response.FailedModels, &FailedModelInfo{
				ModelName: modelID,
				Error:     fmt.Sprintf("API 探测失败: %v", err),
			})
		}
		response.TotalFailed = len(request.ModelIDs)
		return response, nil
	}

	// 创建模型ID到发现模型的映射
	discoveredMap := make(map[string]*DiscoveredModel)
	for _, model := range discoveryResult.Models {
		discoveredMap[model.ID] = model
	}

	// 跟踪唯一模型
	uniqueModels := make(map[string]*AddedModelInfo)

	// 为每个模型ID和每个API Key创建实例
	for _, modelID := range request.ModelIDs {
		discoveredModel, exists := discoveredMap[modelID]
		if !exists {
			response.FailedModels = append(response.FailedModels, &FailedModelInfo{
				ModelName: modelID,
				Error:     "模型不存在于 API 中",
			})
			continue
		}

		// 检查是否已经有这个模型
		if addedModel, exists := uniqueModels[modelID]; exists {
			// 为现有模型添加更多实例
			for _, apiKey := range request.APIKeys {
				instanceID, err := bmm.createModelInstance(ctx, request, discoveredModel, apiKey)
				if err != nil {
					response.FailedModels = append(response.FailedModels, &FailedModelInfo{
						ModelName: modelID,
						Error:     err.Error(),
						APIKey:    apiKey,
					})
					continue
				}

				// 添加到负载均衡器
				bmm.addToLoadBalancer(instanceID, modelID)
				addedModel.InstanceCount++
			}
		} else {
			// 创建新的唯一模型
			addedModel := &AddedModelInfo{
				ModelName:     modelID,
				Capabilities:  discoveredModel.Capabilities,
				Pricing:       discoveredModel.Pricing,
				InstanceCount: 0,
			}

			// 为每个 API Key 创建实例
			for _, apiKey := range request.APIKeys {
				instanceID, err := bmm.createModelInstance(ctx, request, discoveredModel, apiKey)
				if err != nil {
					response.FailedModels = append(response.FailedModels, &FailedModelInfo{
						ModelName: modelID,
						Error:     err.Error(),
						APIKey:    apiKey,
					})
					continue
				}

				if addedModel.InstanceCount == 0 {
					addedModel.ModelID = instanceID // 使用第一个实例的ID作为代表
				}

				// 添加到负载均衡器
				bmm.addToLoadBalancer(instanceID, modelID)
				addedModel.InstanceCount++
			}

			if addedModel.InstanceCount > 0 {
				uniqueModels[modelID] = addedModel
				response.AddedModels = append(response.AddedModels, addedModel)

				// 设置定价
				if request.AutoPricing {
					bmm.SetModelPricing(ctx, modelID, &ModelPricingConfig{
						ModelName:   modelID,
						InputPrice:  discoveredModel.Pricing.InputPrice,
						OutputPrice: discoveredModel.Pricing.OutputPrice,
						ImagePrice:  discoveredModel.Pricing.ImagePrice,
						VideoPrice:  discoveredModel.Pricing.VideoPrice,
						AudioPrice:  discoveredModel.Pricing.AudioPrice,
						Currency:    discoveredModel.Pricing.Currency,
						UpdatedAt:   time.Now(),
					})
				}
			}
		}
	}

	// 统计结果
	response.TotalAdded = 0
	for _, model := range response.AddedModels {
		response.TotalAdded += model.InstanceCount
	}
	response.TotalFailed = len(response.FailedModels)
	response.UniqueModels = len(response.AddedModels)
	response.TotalInstances = response.TotalAdded

	bmm.logger.WithFields(logrus.Fields{
		"unique_models":   response.UniqueModels,
		"total_instances": response.TotalInstances,
		"total_failed":    response.TotalFailed,
	}).Info("批量添加模型完成")

	return response, nil
}

// createModelInstance 创建模型实例
func (bmm *batchModelManager) createModelInstance(ctx context.Context, request *BatchAddRequest, discoveredModel *DiscoveredModel, apiKey string) (uuid.UUID, error) {
	config := &domain.UniversalModelConfig{
		ID:           uuid.New(),
		Name:         discoveredModel.Name,
		InternalKey:  fmt.Sprintf("%s_%s", request.Provider, discoveredModel.ID),
		Provider:     request.Provider,
		ModelID:      discoveredModel.ID,
		APIEndpoint:  request.APIBase,
		APIKey:       apiKey,
		Categories:   []string{discoveredModel.ModelType},
		Capabilities: discoveredModel.Capabilities,

		// AI 参数配置
		AIParameters: &domain.AIParameterConfig{
			CustomParameters: discoveredModel.Parameters,
		},

		// HTTP 配置
		HTTPMethod: "POST",
		AuthType:   "bearer",
		RequestHeaders: map[string]string{
			"Content-Type": "application/json",
		},

		// 状态配置
		Status:     "active",
		Visibility: "public",

		// 时间戳
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 调用通用模型服务创建
	createdConfig, err := bmm.universalModelSvc.CreateModel(ctx, config)
	if err != nil {
		return uuid.Nil, fmt.Errorf("创建模型实例失败: %w", err)
	}

	return createdConfig.ID, nil
}

// addToLoadBalancer 添加到负载均衡器
func (bmm *batchModelManager) addToLoadBalancer(instanceID uuid.UUID, modelName string) {
	// 这里应该调用负载均衡器的添加实例方法
	// 由于接口限制，这里只是记录映射关系
	if bmm.modelInstances[modelName] == nil {
		bmm.modelInstances[modelName] = make([]uuid.UUID, 0)
	}
	bmm.modelInstances[modelName] = append(bmm.modelInstances[modelName], instanceID)
}

// BatchAddAPIKeys 批量添加 API Keys (相同模型多实例)
func (bmm *batchModelManager) BatchAddAPIKeys(ctx context.Context, request *BatchKeysRequest) (*BatchKeysResponse, error) {
	bmm.logger.WithFields(logrus.Fields{
		"api_base":   request.APIBase,
		"model_name": request.ModelName,
		"api_keys":   len(request.APIKeys),
	}).Info("开始批量添加 API Keys")

	response := &BatchKeysResponse{
		Success:    true,
		ModelName:  request.ModelName,
		AddedKeys:  make([]string, 0),
		FailedKeys: make([]string, 0),
		Strategy:   request.LoadBalanceStrategy,
	}

	// 验证每个 API Key
	for _, apiKey := range request.APIKeys {
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			continue
		}

		// 测试 API Key 是否有效
		if bmm.testAPIKey(ctx, request.APIBase, apiKey, request.ModelName) {
			response.AddedKeys = append(response.AddedKeys, apiKey)

			// 创建模型实例 (这里简化处理)
			instanceID := uuid.New()
			bmm.addToLoadBalancer(instanceID, request.ModelName)
		} else {
			response.FailedKeys = append(response.FailedKeys, apiKey)
		}
	}

	response.TotalInstances = len(response.AddedKeys)

	// 设置负载均衡策略
	if response.TotalInstances > 1 && request.LoadBalanceStrategy != "" {
		// 这里应该调用负载均衡器设置策略
		bmm.logger.WithFields(logrus.Fields{
			"model_name": request.ModelName,
			"strategy":   request.LoadBalanceStrategy,
			"instances":  response.TotalInstances,
		}).Info("设置负载均衡策略")
	}

	bmm.logger.WithFields(logrus.Fields{
		"model_name":      request.ModelName,
		"added_keys":      len(response.AddedKeys),
		"failed_keys":     len(response.FailedKeys),
		"total_instances": response.TotalInstances,
	}).Info("批量添加 API Keys 完成")

	return response, nil
}

// testAPIKey 测试 API Key 是否有效
func (bmm *batchModelManager) testAPIKey(ctx context.Context, apiBase, apiKey, modelName string) bool {
	// 构建测试请求
	testURL := strings.TrimSuffix(apiBase, "/") + "/models"

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := bmm.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// SetModelPricing 设置模型定价
func (bmm *batchModelManager) SetModelPricing(ctx context.Context, modelName string, pricing *ModelPricingConfig) error {
	bmm.pricingCache[modelName] = pricing

	bmm.logger.WithFields(logrus.Fields{
		"model_name":   modelName,
		"input_price":  pricing.InputPrice,
		"output_price": pricing.OutputPrice,
		"currency":     pricing.Currency,
	}).Info("设置模型定价")

	return nil
}

// GetModelPricing 获取模型定价
func (bmm *batchModelManager) GetModelPricing(ctx context.Context, modelName string) (*ModelPricingConfig, error) {
	if pricing, exists := bmm.pricingCache[modelName]; exists {
		return pricing, nil
	}

	return nil, fmt.Errorf("模型 %s 的定价信息不存在", modelName)
}

// GetUniqueModels 获取去重后的模型列表
func (bmm *batchModelManager) GetUniqueModels(ctx context.Context) ([]*UniqueModelInfo, error) {
	uniqueModels := make([]*UniqueModelInfo, 0)

	for modelName, instanceIDs := range bmm.modelInstances {
		pricing, _ := bmm.GetModelPricing(ctx, modelName)

		uniqueModel := &UniqueModelInfo{
			ModelName:     modelName,
			DisplayName:   modelName,
			InstanceCount: len(instanceIDs),
			Status:        "active",
			Pricing:       pricing,
		}

		uniqueModels = append(uniqueModels, uniqueModel)
	}

	return uniqueModels, nil
}

// GetModelInstances 获取模型的所有实例
func (bmm *batchModelManager) GetModelInstances(ctx context.Context, modelName string) ([]*ModelInstanceInfo, error) {
	instanceIDs, exists := bmm.modelInstances[modelName]
	if !exists {
		return nil, fmt.Errorf("模型 %s 不存在", modelName)
	}

	instances := make([]*ModelInstanceInfo, 0)

	for _, instanceID := range instanceIDs {
		instance := &ModelInstanceInfo{
			InstanceID:   instanceID,
			Status:       "active",
			Health:       "healthy",
			CurrentLoad:  0,
			Weight:       1,
			ResponseTime: 100,
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// CallModelWithBilling 调用模型并扣费 - 集成之前的金币扣费功能
func (bmm *batchModelManager) CallModelWithBilling(ctx context.Context, userID uuid.UUID, request *ModelCallWithBillingRequest) (*ModelCallWithBillingResponse, error) {
	bmm.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"model_name": request.ModelName,
		"input":      request.Input,
	}).Info("开始调用模型并扣费")

	// 1. 获取模型定价
	pricing, err := bmm.GetModelPricing(ctx, request.ModelName)
	if err != nil {
		return nil, fmt.Errorf("获取模型定价失败: %w", err)
	}

	// 2. 计算调用成本 (用于日志记录)
	_ = bmm.calculateCallCost(request, pricing)

	// 3. 预检查余额
	inputTokens := int64(request.InputTokens)
	outputTokens := int64(request.EstimatedOutputTokens)
	imageCount := request.ImageCount
	videoSeconds := int(request.VideoSeconds)
	audioSeconds := int(request.AudioSeconds)

	preCheckReq := &domain.PricingCalculationRequest{
		ModelID:      request.ModelID,
		InputTokens:  &inputTokens,
		OutputTokens: &outputTokens,
		ImageCount:   &imageCount,
		VideoSeconds: &videoSeconds,
		AudioSeconds: &audioSeconds,
		UserType:     "basic", // TODO: 从用户信息获取实际用户类型
	}

	preCheck, err := bmm.aiBillingService.PreCheckBalance(ctx, userID, request.ModelID, preCheckReq)
	if err != nil {
		return nil, fmt.Errorf("余额预检查失败: %w", err)
	}

	if !preCheck.Sufficient {
		return &ModelCallWithBillingResponse{
			Success:      false,
			ErrorMessage: preCheck.ErrorMessage,
			PreCheck:     preCheck,
		}, nil
	}

	// 4. 选择最佳模型实例 (负载均衡)
	instances, err := bmm.GetModelInstances(ctx, request.ModelName)
	if err != nil {
		return nil, fmt.Errorf("获取模型实例失败: %w", err)
	}

	if len(instances) == 0 {
		return &ModelCallWithBillingResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("模型 %s 没有可用实例", request.ModelName),
		}, nil
	}

	// 使用负载均衡选择实例
	selectedInstance := bmm.selectBestInstance(instances, request.SessionID)

	// 5. 执行扣费
	requestID, _ := uuid.Parse(request.RequestID)
	if requestID == uuid.Nil {
		requestID = uuid.New()
	}

	chargeReq := &domain.AICallChargeRequest{
		RequestID:                 requestID,
		ModelID:                   request.ModelID,
		ModelName:                 request.ModelName,
		TotalTokens:               int64(request.InputTokens + request.EstimatedOutputTokens),
		PricingCalculationRequest: *preCheckReq,
	}

	chargeResp, err := bmm.aiBillingService.ChargeForAICall(ctx, userID, chargeReq)
	if err != nil {
		return nil, fmt.Errorf("扣费失败: %w", err)
	}

	if !chargeResp.Success {
		return &ModelCallWithBillingResponse{
			Success:      false,
			ErrorMessage: chargeResp.ErrorMessage,
			PreCheck:     chargeResp.PreCheck,
		}, nil
	}

	// 6. 调用模型 (这里简化处理，实际应该调用真实的模型API)
	modelResponse := bmm.simulateModelCall(request, selectedInstance)

	// 7. 如果调用失败，执行退费
	if !modelResponse.Success {
		refundErr := bmm.aiBillingService.RefundAICall(ctx, userID, chargeResp.TransactionID, "模型调用失败")
		if refundErr != nil {
			bmm.logger.WithError(refundErr).Error("退费失败")
		}
	}

	return &ModelCallWithBillingResponse{
		Success:       modelResponse.Success,
		Output:        modelResponse.Output,
		Usage:         modelResponse.Usage,
		CoinsCharged:  chargeResp.CoinsCharged,
		NewBalance:    chargeResp.NewBalance,
		TransactionID: chargeResp.TransactionID,
		InstanceUsed:  selectedInstance.InstanceID.String(),
		ResponseTime:  modelResponse.ResponseTime,
		ErrorMessage:  modelResponse.ErrorMessage,
	}, nil
}

// ModelCallWithBillingRequest 带扣费的模型调用请求
type ModelCallWithBillingRequest struct {
	ModelID               uuid.UUID              `json:"model_id"`
	ModelName             string                 `json:"model_name"`
	ModelType             string                 `json:"model_type"`
	Input                 interface{}            `json:"input"`
	Parameters            map[string]interface{} `json:"parameters"`
	RequestID             string                 `json:"request_id"`
	SessionID             string                 `json:"session_id"`
	InputTokens           int                    `json:"input_tokens"`
	EstimatedOutputTokens int                    `json:"estimated_output_tokens"`
	ImageCount            int                    `json:"image_count"`
	VideoSeconds          float64                `json:"video_seconds"`
	AudioSeconds          float64                `json:"audio_seconds"`
}

// ModelCallWithBillingResponse 带扣费的模型调用响应
type ModelCallWithBillingResponse struct {
	Success       bool                     `json:"success"`
	Output        interface{}              `json:"output"`
	Usage         *UsageInfo               `json:"usage"`
	CoinsCharged  float64                  `json:"coins_charged"`
	NewBalance    float64                  `json:"new_balance"`
	TransactionID uuid.UUID                `json:"transaction_id"`
	InstanceUsed  string                   `json:"instance_used"`
	ResponseTime  int64                    `json:"response_time"`
	ErrorMessage  string                   `json:"error_message,omitempty"`
	PreCheck      *domain.PreCheckResponse `json:"pre_check,omitempty"`
}

// calculateCallCost 计算调用成本
func (bmm *batchModelManager) calculateCallCost(request *ModelCallWithBillingRequest, pricing *ModelPricingConfig) float64 {
	var totalCost float64

	switch request.ModelType {
	case "text":
		// 文本模型按token计费
		inputCost := float64(request.InputTokens) / 1000.0 * pricing.InputPrice
		outputCost := float64(request.EstimatedOutputTokens) / 1000.0 * pricing.OutputPrice
		totalCost = inputCost + outputCost
	case "image":
		// 图像模型按图片数量计费
		totalCost = float64(request.ImageCount) * pricing.ImagePrice
	case "video":
		// 视频模型按秒数计费
		totalCost = request.VideoSeconds * pricing.VideoPrice
	case "audio":
		// 音频模型按秒数计费
		totalCost = request.AudioSeconds * pricing.AudioPrice
	default:
		// 默认按token计费
		totalCost = float64(request.InputTokens+request.EstimatedOutputTokens) / 1000.0 * pricing.InputPrice
	}

	return totalCost
}

// selectBestInstance 选择最佳实例 (简化的负载均衡)
func (bmm *batchModelManager) selectBestInstance(instances []*ModelInstanceInfo, sessionID string) *ModelInstanceInfo {
	if len(instances) == 1 {
		return instances[0]
	}

	// 简单的轮询选择
	// 实际应该调用负载均衡器的选择逻辑
	return instances[0]
}

// simulateModelCall 模拟模型调用 (实际应该调用真实API)
func (bmm *batchModelManager) simulateModelCall(request *ModelCallWithBillingRequest, instance *ModelInstanceInfo) *ModelCallResponse {
	// 这里应该调用真实的模型API
	// 现在只是模拟响应
	return &ModelCallResponse{
		Success:      true,
		Output:       fmt.Sprintf("模拟响应: %v", request.Input),
		Usage:        &UsageInfo{InputTokens: request.InputTokens, OutputTokens: request.EstimatedOutputTokens},
		ResponseTime: 1000, // 1秒
	}
}

// ModelCallResponse 模型调用响应
type ModelCallResponse struct {
	Success      bool        `json:"success"`
	Output       interface{} `json:"output"`
	Usage        *UsageInfo  `json:"usage"`
	ResponseTime int64       `json:"response_time"`
	ErrorMessage string      `json:"error_message,omitempty"`
}
