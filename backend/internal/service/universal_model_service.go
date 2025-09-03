package service

import (
	"bytes"
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
)

// 🌍 YUNAI 超级模型管理服务 - 支持全球所有模型的完全自定义系统

// UniversalModelService 通用模型服务接口
type UniversalModelService interface {
	// 模型管理
	CreateModel(ctx context.Context, config *domain.UniversalModelConfig) (*domain.UniversalModelConfig, error)
	UpdateModel(ctx context.Context, id uuid.UUID, config *domain.UniversalModelConfig) (*domain.UniversalModelConfig, error)
	DeleteModel(ctx context.Context, id uuid.UUID) error
	GetModel(ctx context.Context, id uuid.UUID) (*domain.UniversalModelConfig, error)
	ListModels(ctx context.Context, filter *ModelFilter) ([]*domain.UniversalModelConfig, error)
	
	// 模型调用 - 通用接口
	CallModel(ctx context.Context, request *UniversalModelRequest) (*UniversalModelResponse, error)
	CallModelStreaming(ctx context.Context, request *UniversalModelRequest) (<-chan *StreamingChunk, error)
	
	// 模型测试
	TestModel(ctx context.Context, id uuid.UUID, testRequest *ModelTestRequest) (*ModelTestResponse, error)
	
	// 健康检查
	CheckModelHealth(ctx context.Context, id uuid.UUID) (*ModelHealthStatus, error)
	
	// 批量操作
	BatchCallModels(ctx context.Context, requests []*UniversalModelRequest) ([]*UniversalModelResponse, error)
	
	// 模型发现和导入
	DiscoverModels(ctx context.Context, provider string, apiKey string) ([]*ModelDiscoveryResult, error)
	ImportModel(ctx context.Context, discoveryResult *ModelDiscoveryResult) (*domain.UniversalModelConfig, error)
}

type universalModelService struct {
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewUniversalModelService 创建通用模型服务
func NewUniversalModelService(logger *logrus.Logger) UniversalModelService {
	return &universalModelService{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// ModelFilter 模型过滤器
type ModelFilter struct {
	Provider     string   `json:"provider"`
	ModelType    string   `json:"model_type"`
	Capabilities []string `json:"capabilities"`
	Categories   []string `json:"categories"`
	Status       string   `json:"status"`
	Search       string   `json:"search"`
	Page         int      `json:"page"`
	PageSize     int      `json:"page_size"`
}

// UniversalModelRequest 通用模型请求
type UniversalModelRequest struct {
	ModelID    uuid.UUID              `json:"model_id"`
	Parameters map[string]interface{} `json:"parameters"`
	Input      interface{}            `json:"input"`      // 可以是文本、图片URL、文件等
	Options    *RequestOptions        `json:"options"`
}

// RequestOptions 请求选项
type RequestOptions struct {
	Streaming     bool              `json:"streaming"`
	Timeout       int               `json:"timeout"`
	RetryCount    int               `json:"retry_count"`
	CustomHeaders map[string]string `json:"custom_headers"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// UniversalModelResponse 通用模型响应
type UniversalModelResponse struct {
	ModelID      uuid.UUID              `json:"model_id"`
	RequestID    string                 `json:"request_id"`
	Output       interface{}            `json:"output"`
	Usage        *UsageInfo             `json:"usage"`
	Metadata     map[string]interface{} `json:"metadata"`
	ResponseTime int64                  `json:"response_time"`
	Status       string                 `json:"status"`
	Error        string                 `json:"error,omitempty"`
}

// StreamingChunk 流式响应块
type StreamingChunk struct {
	ModelID   uuid.UUID   `json:"model_id"`
	RequestID string      `json:"request_id"`
	Delta     interface{} `json:"delta"`
	Finished  bool        `json:"finished"`
	Error     string      `json:"error,omitempty"`
}

// UsageInfo 使用信息
type UsageInfo struct {
	InputTokens    int     `json:"input_tokens"`
	OutputTokens   int     `json:"output_tokens"`
	TotalTokens    int     `json:"total_tokens"`
	ImageCount     int     `json:"image_count"`
	VideoSeconds   float64 `json:"video_seconds"`
	AudioSeconds   float64 `json:"audio_seconds"`
	RequestCount   int     `json:"request_count"`
	Cost           float64 `json:"cost"`
	Currency       string  `json:"currency"`
}

// ModelTestRequest 模型测试请求
type ModelTestRequest struct {
	TestType   string                 `json:"test_type"`   // connectivity, functionality, performance
	TestInput  interface{}            `json:"test_input"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ModelTestResponse 模型测试响应
type ModelTestResponse struct {
	Success      bool                   `json:"success"`
	ResponseTime int64                  `json:"response_time"`
	Output       interface{}            `json:"output"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ModelHealthStatus 模型健康状态
type ModelHealthStatus struct {
	Status       string    `json:"status"`        // healthy, unhealthy, unknown
	ResponseTime int64     `json:"response_time"` // 响应时间(ms)
	LastCheck    time.Time `json:"last_check"`
	Error        string    `json:"error,omitempty"`
}

// ModelDiscoveryResult 模型发现结果
type ModelDiscoveryResult struct {
	ModelID      string                 `json:"model_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"`
	Parameters   map[string]interface{} `json:"parameters"`
	Pricing      map[string]interface{} `json:"pricing"`
}

// CallModel 调用模型 - 通用接口支持所有模型
func (s *universalModelService) CallModel(ctx context.Context, request *UniversalModelRequest) (*UniversalModelResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"model_id": request.ModelID,
		"input":    request.Input,
	}).Debug("Calling universal model")

	// 1. 获取模型配置
	modelConfig, err := s.GetModel(ctx, request.ModelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}

	// 2. 构建请求
	httpRequest, err := s.buildHTTPRequest(ctx, modelConfig, request)
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// 3. 发送请求
	startTime := time.Now()
	httpResponse, err := s.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResponse.Body.Close()
	responseTime := time.Since(startTime).Milliseconds()

	// 4. 解析响应
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 5. 映射响应
	response, err := s.mapResponse(modelConfig, responseBody, httpResponse.StatusCode)
	if err != nil {
		return nil, fmt.Errorf("failed to map response: %w", err)
	}

	response.ModelID = request.ModelID
	response.ResponseTime = responseTime
	response.RequestID = uuid.New().String()

	return response, nil
}

// buildHTTPRequest 构建HTTP请求 - 支持完全自定义
func (s *universalModelService) buildHTTPRequest(ctx context.Context, config *domain.UniversalModelConfig, request *UniversalModelRequest) (*http.Request, error) {
	// 1. 构建请求体
	requestBody, err := s.buildRequestBody(config, request)
	if err != nil {
		return nil, fmt.Errorf("failed to build request body: %w", err)
	}

	// 2. 创建HTTP请求
	httpRequest, err := http.NewRequestWithContext(ctx, config.HTTPMethod, config.APIEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 3. 设置请求头
	s.setRequestHeaders(httpRequest, config, request)

	// 4. 设置认证
	err = s.setAuthentication(httpRequest, config)
	if err != nil {
		return nil, fmt.Errorf("failed to set authentication: %w", err)
	}

	return httpRequest, nil
}

// buildRequestBody 构建请求体 - 支持完全自定义模板
func (s *universalModelService) buildRequestBody(config *domain.UniversalModelConfig, request *UniversalModelRequest) ([]byte, error) {
	if config.RequestTemplate == nil {
		// 默认请求体构建
		body := map[string]interface{}{
			"input": request.Input,
		}
		
		// 合并参数
		for k, v := range request.Parameters {
			// 检查参数映射
			if mappedKey, exists := config.ParameterMappings[k]; exists {
				body[mappedKey] = v
			} else {
				body[k] = v
			}
		}
		
		// 合并默认参数
		for k, v := range config.DefaultParameters {
			if _, exists := body[k]; !exists {
				body[k] = v
			}
		}
		
		return json.Marshal(body)
	}

	// 使用自定义模板
	templateStr := string(*config.RequestTemplate)
	
	// 替换模板变量
	templateStr = s.replaceTemplateVariables(templateStr, request)
	
	return []byte(templateStr), nil
}

// replaceTemplateVariables 替换模板变量
func (s *universalModelService) replaceTemplateVariables(template string, request *UniversalModelRequest) string {
	// 替换输入
	if inputStr, ok := request.Input.(string); ok {
		template = strings.ReplaceAll(template, "{{input}}", inputStr)
	}
	
	// 替换参数
	for key, value := range request.Parameters {
		placeholder := fmt.Sprintf("{{%s}}", key)
		if valueStr, ok := value.(string); ok {
			template = strings.ReplaceAll(template, placeholder, valueStr)
		} else {
			valueBytes, _ := json.Marshal(value)
			template = strings.ReplaceAll(template, placeholder, string(valueBytes))
		}
	}
	
	return template
}

// setRequestHeaders 设置请求头
func (s *universalModelService) setRequestHeaders(httpRequest *http.Request, config *domain.UniversalModelConfig, request *UniversalModelRequest) {
	// 设置默认请求头
	httpRequest.Header.Set("Content-Type", "application/json")
	
	// 设置模型配置的请求头
	for key, value := range config.RequestHeaders {
		httpRequest.Header.Set(key, value)
	}
	
	// 设置请求选项的自定义请求头
	if request.Options != nil && request.Options.CustomHeaders != nil {
		for key, value := range request.Options.CustomHeaders {
			httpRequest.Header.Set(key, value)
		}
	}
}

// setAuthentication 设置认证 - 支持多种认证方式
func (s *universalModelService) setAuthentication(httpRequest *http.Request, config *domain.UniversalModelConfig) error {
	switch config.AuthType {
	case "bearer":
		httpRequest.Header.Set("Authorization", "Bearer "+config.APIKey)
	case "basic":
		httpRequest.SetBasicAuth(config.APIKey, config.APISecret)
	case "api_key":
		if keyName, exists := config.AuthConfig["key_name"].(string); exists {
			httpRequest.Header.Set(keyName, config.APIKey)
		} else {
			httpRequest.Header.Set("X-API-Key", config.APIKey)
		}
	case "custom":
		// 自定义认证逻辑
		for key, value := range config.AuthConfig {
			if valueStr, ok := value.(string); ok {
				// 替换API密钥占位符
				valueStr = strings.ReplaceAll(valueStr, "{{api_key}}", config.APIKey)
				valueStr = strings.ReplaceAll(valueStr, "{{api_secret}}", config.APISecret)
				httpRequest.Header.Set(key, valueStr)
			}
		}
	default:
		// 默认使用Bearer认证
		httpRequest.Header.Set("Authorization", "Bearer "+config.APIKey)
	}
	
	return nil
}

// mapResponse 映射响应 - 支持完全自定义响应映射
func (s *universalModelService) mapResponse(config *domain.UniversalModelConfig, responseBody []byte, statusCode int) (*UniversalModelResponse, error) {
	response := &UniversalModelResponse{
		Status: "success",
	}
	
	// 检查HTTP状态码
	if statusCode >= 400 {
		response.Status = "error"
		response.Error = fmt.Sprintf("HTTP %d: %s", statusCode, string(responseBody))
		return response, nil
	}
	
	// 解析JSON响应
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &jsonResponse); err != nil {
		// 如果不是JSON，直接返回原始响应
		response.Output = string(responseBody)
		return response, nil
	}
	
	// 使用响应映射
	if config.ResponseMapping != nil {
		var mapping map[string]string
		if err := json.Unmarshal(*config.ResponseMapping, &mapping); err == nil {
			mappedResponse := make(map[string]interface{})
			for sourceKey, targetKey := range mapping {
				if value, exists := jsonResponse[sourceKey]; exists {
					mappedResponse[targetKey] = value
				}
			}
			response.Output = mappedResponse
		}
	} else {
		// 默认映射
		response.Output = jsonResponse
	}
	
	// 提取使用信息
	response.Usage = s.extractUsageInfo(jsonResponse)
	
	return response, nil
}

// extractUsageInfo 提取使用信息
func (s *universalModelService) extractUsageInfo(jsonResponse map[string]interface{}) *UsageInfo {
	usage := &UsageInfo{}
	
	// 尝试从常见字段提取使用信息
	if usageData, exists := jsonResponse["usage"]; exists {
		if usageMap, ok := usageData.(map[string]interface{}); ok {
			if inputTokens, ok := usageMap["input_tokens"].(float64); ok {
				usage.InputTokens = int(inputTokens)
			}
			if outputTokens, ok := usageMap["output_tokens"].(float64); ok {
				usage.OutputTokens = int(outputTokens)
			}
			if totalTokens, ok := usageMap["total_tokens"].(float64); ok {
				usage.TotalTokens = int(totalTokens)
			}
		}
	}
	
	return usage
}

// 实现其他接口方法的占位符
func (s *universalModelService) CreateModel(ctx context.Context, config *domain.UniversalModelConfig) (*domain.UniversalModelConfig, error) {
	// TODO: 实现模型创建逻辑
	return config, nil
}

func (s *universalModelService) UpdateModel(ctx context.Context, id uuid.UUID, config *domain.UniversalModelConfig) (*domain.UniversalModelConfig, error) {
	// TODO: 实现模型更新逻辑
	return config, nil
}

func (s *universalModelService) DeleteModel(ctx context.Context, id uuid.UUID) error {
	// TODO: 实现模型删除逻辑
	return nil
}

func (s *universalModelService) GetModel(ctx context.Context, id uuid.UUID) (*domain.UniversalModelConfig, error) {
	// TODO: 实现从数据库获取模型配置
	// 这里返回一个示例配置
	return &domain.UniversalModelConfig{
		ID:           id,
		Name:         "示例模型",
		APIEndpoint:  "https://api.example.com/v1/chat/completions",
		HTTPMethod:   "POST",
		AuthType:     "bearer",
		APIKey:       "example-key",
		DefaultParameters: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  1000,
		},
		ParameterMappings: map[string]string{
			"input": "messages",
		},
	}, nil
}

func (s *universalModelService) ListModels(ctx context.Context, filter *ModelFilter) ([]*domain.UniversalModelConfig, error) {
	// TODO: 实现模型列表查询
	return []*domain.UniversalModelConfig{}, nil
}

func (s *universalModelService) CallModelStreaming(ctx context.Context, request *UniversalModelRequest) (<-chan *StreamingChunk, error) {
	// TODO: 实现流式调用
	ch := make(chan *StreamingChunk)
	close(ch)
	return ch, nil
}

func (s *universalModelService) TestModel(ctx context.Context, id uuid.UUID, testRequest *ModelTestRequest) (*ModelTestResponse, error) {
	// TODO: 实现模型测试
	return &ModelTestResponse{Success: true}, nil
}

func (s *universalModelService) CheckModelHealth(ctx context.Context, id uuid.UUID) (*ModelHealthStatus, error) {
	// TODO: 实现健康检查
	return &ModelHealthStatus{Status: "healthy"}, nil
}

func (s *universalModelService) BatchCallModels(ctx context.Context, requests []*UniversalModelRequest) ([]*UniversalModelResponse, error) {
	// TODO: 实现批量调用
	return []*UniversalModelResponse{}, nil
}

func (s *universalModelService) DiscoverModels(ctx context.Context, provider string, apiKey string) ([]*ModelDiscoveryResult, error) {
	// TODO: 实现模型发现
	return []*ModelDiscoveryResult{}, nil
}

func (s *universalModelService) ImportModel(ctx context.Context, discoveryResult *ModelDiscoveryResult) (*domain.UniversalModelConfig, error) {
	// TODO: 实现模型导入
	return &domain.UniversalModelConfig{}, nil
}
