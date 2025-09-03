package adapter

import (
	"context"
	"yunai/internal/domain"
)

// VoiceAdapter 语音服务适配器接口
type VoiceAdapter interface {
	// 基础信息
	GetProviderName() string
	GetSupportedCapabilities() domain.ProviderCapabilities
	
	// TTS 文本转语音
	TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error)
	
	// ASR 语音转文本
	SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error)
	
	// 音色克隆
	CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error)
	
	// 获取音色列表
	ListVoices(ctx context.Context) ([]domain.VoiceTemplate, error)
	
	// 删除音色
	DeleteVoice(ctx context.Context, voiceID string) error
	
	// 健康检查
	HealthCheck(ctx context.Context) error
	
	// 配置验证
	ValidateConfig(config map[string]interface{}) error
}

// VoiceAdapterFactory 语音适配器工厂
type VoiceAdapterFactory interface {
	CreateAdapter(provider *domain.VoiceProvider) (VoiceAdapter, error)
	GetSupportedProviders() []string
}

// BaseVoiceAdapter 基础语音适配器
type BaseVoiceAdapter struct {
	Provider *domain.VoiceProvider
	Config   *domain.VoiceProviderConfig
}

// NewBaseVoiceAdapter 创建基础语音适配器
func NewBaseVoiceAdapter(provider *domain.VoiceProvider) *BaseVoiceAdapter {
	return &BaseVoiceAdapter{
		Provider: provider,
	}
}

// GetProviderName 获取提供商名称
func (a *BaseVoiceAdapter) GetProviderName() string {
	return a.Provider.Name
}

// ValidateConfig 验证配置
func (a *BaseVoiceAdapter) ValidateConfig(config map[string]interface{}) error {
	// 基础验证逻辑
	return nil
}

// VoiceAdapterManager 语音适配器管理器
type VoiceAdapterManager struct {
	adapters map[string]VoiceAdapter
	factory  VoiceAdapterFactory
}

// NewVoiceAdapterManager 创建语音适配器管理器
func NewVoiceAdapterManager(factory VoiceAdapterFactory) *VoiceAdapterManager {
	return &VoiceAdapterManager{
		adapters: make(map[string]VoiceAdapter),
		factory:  factory,
	}
}

// RegisterAdapter 注册适配器
func (m *VoiceAdapterManager) RegisterAdapter(providerID string, adapter VoiceAdapter) {
	m.adapters[providerID] = adapter
}

// GetAdapter 获取适配器
func (m *VoiceAdapterManager) GetAdapter(providerID string) (VoiceAdapter, bool) {
	adapter, exists := m.adapters[providerID]
	return adapter, exists
}

// CreateAdapter 创建适配器
func (m *VoiceAdapterManager) CreateAdapter(provider *domain.VoiceProvider) (VoiceAdapter, error) {
	return m.factory.CreateAdapter(provider)
}

// RemoveAdapter 移除适配器
func (m *VoiceAdapterManager) RemoveAdapter(providerID string) {
	delete(m.adapters, providerID)
}

// ListAdapters 列出所有适配器
func (m *VoiceAdapterManager) ListAdapters() map[string]VoiceAdapter {
	return m.adapters
}

// DefaultVoiceAdapterFactory 默认语音适配器工厂
type DefaultVoiceAdapterFactory struct{}

// NewDefaultVoiceAdapterFactory 创建默认语音适配器工厂
func NewDefaultVoiceAdapterFactory() *DefaultVoiceAdapterFactory {
	return &DefaultVoiceAdapterFactory{}
}

// CreateAdapter 创建适配器
func (f *DefaultVoiceAdapterFactory) CreateAdapter(provider *domain.VoiceProvider) (VoiceAdapter, error) {
	switch provider.Name {
	case "siliconflow":
		return NewSiliconFlowAdapter(provider), nil
	case "openai":
		return NewOpenAIAdapter(provider), nil
	case "azure":
		return NewAzureAdapter(provider), nil
	case "elevenlabs":
		return NewElevenLabsAdapter(provider), nil
	case "custom":
		return NewCustomAdapter(provider), nil
	default:
		return NewGenericAdapter(provider), nil
	}
}

// GetSupportedProviders 获取支持的提供商
func (f *DefaultVoiceAdapterFactory) GetSupportedProviders() []string {
	return []string{
		"siliconflow",
		"openai", 
		"azure",
		"elevenlabs",
		"custom",
		"generic",
	}
}

// AdapterError 适配器错误
type AdapterError struct {
	Provider string
	Message  string
	Code     string
	Err      error
}

// Error 实现error接口
func (e *AdapterError) Error() string {
	if e.Err != nil {
		return e.Provider + ": " + e.Message + " - " + e.Err.Error()
	}
	return e.Provider + ": " + e.Message
}

// Unwrap 解包错误
func (e *AdapterError) Unwrap() error {
	return e.Err
}

// NewAdapterError 创建适配器错误
func NewAdapterError(provider, message, code string, err error) *AdapterError {
	return &AdapterError{
		Provider: provider,
		Message:  message,
		Code:     code,
		Err:      err,
	}
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	Timeout    int                    `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
	RateLimit  map[string]int         `json:"rate_limit"`
	Headers    map[string]string      `json:"headers"`
	Params     map[string]interface{} `json:"params"`
}

// LoadAdapterConfig 加载适配器配置
func LoadAdapterConfig(provider *domain.VoiceProvider) (*AdapterConfig, error) {
	config := &AdapterConfig{
		Timeout:    30,
		RetryCount: 3,
		RateLimit:  make(map[string]int),
		Headers:    make(map[string]string),
		Params:     make(map[string]interface{}),
	}
	
	// 从provider.Config中解析配置
	// 这里可以添加具体的配置解析逻辑
	
	return config, nil
}

// AdapterMetrics 适配器指标
type AdapterMetrics struct {
	RequestCount  int64   `json:"request_count"`
	SuccessCount  int64   `json:"success_count"`
	ErrorCount    int64   `json:"error_count"`
	AvgLatency    float64 `json:"avg_latency"`
	LastUsed      int64   `json:"last_used"`
	HealthStatus  string  `json:"health_status"`
}

// AdapterRegistry 适配器注册表
type AdapterRegistry struct {
	adapters map[string]VoiceAdapter
	metrics  map[string]*AdapterMetrics
}

// NewAdapterRegistry 创建适配器注册表
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]VoiceAdapter),
		metrics:  make(map[string]*AdapterMetrics),
	}
}

// Register 注册适配器
func (r *AdapterRegistry) Register(name string, adapter VoiceAdapter) {
	r.adapters[name] = adapter
	r.metrics[name] = &AdapterMetrics{
		HealthStatus: "unknown",
	}
}

// Get 获取适配器
func (r *AdapterRegistry) Get(name string) (VoiceAdapter, bool) {
	adapter, exists := r.adapters[name]
	return adapter, exists
}

// List 列出所有适配器
func (r *AdapterRegistry) List() []string {
	var names []string
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// GetMetrics 获取适配器指标
func (r *AdapterRegistry) GetMetrics(name string) (*AdapterMetrics, bool) {
	metrics, exists := r.metrics[name]
	return metrics, exists
}

// UpdateMetrics 更新适配器指标
func (r *AdapterRegistry) UpdateMetrics(name string, success bool, latency float64) {
	if metrics, exists := r.metrics[name]; exists {
		metrics.RequestCount++
		if success {
			metrics.SuccessCount++
		} else {
			metrics.ErrorCount++
		}
		
		// 计算平均延迟
		if metrics.RequestCount == 1 {
			metrics.AvgLatency = latency
		} else {
			metrics.AvgLatency = (metrics.AvgLatency*float64(metrics.RequestCount-1) + latency) / float64(metrics.RequestCount)
		}
		
		metrics.LastUsed = getCurrentTimestamp()
	}
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return 0 // 实现获取当前时间戳的逻辑
}
