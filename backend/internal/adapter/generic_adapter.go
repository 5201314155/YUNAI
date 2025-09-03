package adapter

import (
	"context"
	"net/http"
	"time"

	"yunai/internal/domain"
)

// GenericAdapter 通用语音服务适配器
// 用于支持自定义API格式的提供商
type GenericAdapter struct {
	*BaseVoiceAdapter
	client *http.Client
}

// NewGenericAdapter 创建通用适配器
func NewGenericAdapter(provider *domain.VoiceProvider) *GenericAdapter {
	return &GenericAdapter{
		BaseVoiceAdapter: NewBaseVoiceAdapter(provider),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetSupportedCapabilities 获取支持的能力
func (a *GenericAdapter) GetSupportedCapabilities() domain.ProviderCapabilities {
	// 从配置中读取能力，或使用默认值
	return domain.ProviderCapabilities{
		TTS:        true,
		ASR:        true,
		VoiceClone: false,
		Emotions:   []string{"neutral"},
		Languages:  []string{"en-US", "zh-CN"},
		AudioFormats: []string{"mp3", "wav"},
	}
}

// TextToSpeech 文本转语音
func (a *GenericAdapter) TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error) {
	// 通用实现，可以通过配置自定义请求格式
	return nil, NewAdapterError(a.GetProviderName(), "通用适配器需要自定义实现", "NOT_IMPLEMENTED", nil)
}

// SpeechToText 语音转文本
func (a *GenericAdapter) SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error) {
	// 通用实现，可以通过配置自定义请求格式
	return nil, NewAdapterError(a.GetProviderName(), "通用适配器需要自定义实现", "NOT_IMPLEMENTED", nil)
}

// CloneVoice 音色克隆
func (a *GenericAdapter) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	return nil, NewAdapterError(a.GetProviderName(), "通用适配器不支持音色克隆", "NOT_SUPPORTED", nil)
}

// ListVoices 获取音色列表
func (a *GenericAdapter) ListVoices(ctx context.Context) ([]domain.VoiceTemplate, error) {
	// 返回空列表或从配置中读取
	return []domain.VoiceTemplate{}, nil
}

// DeleteVoice 删除音色
func (a *GenericAdapter) DeleteVoice(ctx context.Context, voiceID string) error {
	return NewAdapterError(a.GetProviderName(), "通用适配器不支持删除音色", "NOT_SUPPORTED", nil)
}

// HealthCheck 健康检查
func (a *GenericAdapter) HealthCheck(ctx context.Context) error {
	return NewAdapterError(a.GetProviderName(), "通用适配器需要自定义健康检查", "NOT_IMPLEMENTED", nil)
}

// AzureAdapter Azure语音服务适配器
type AzureAdapter struct {
	*BaseVoiceAdapter
	client *http.Client
}

// NewAzureAdapter 创建Azure适配器
func NewAzureAdapter(provider *domain.VoiceProvider) *AzureAdapter {
	return &AzureAdapter{
		BaseVoiceAdapter: NewBaseVoiceAdapter(provider),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetSupportedCapabilities 获取支持的能力
func (a *AzureAdapter) GetSupportedCapabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{
		TTS:        true,
		ASR:        true,
		VoiceClone: true, // Azure支持自定义神经语音
		Emotions:   []string{"neutral", "cheerful", "sad", "angry", "excited", "friendly"},
		Languages:  []string{"en-US", "zh-CN", "ja-JP", "ko-KR", "es-ES", "fr-FR", "de-DE"},
		AudioFormats: []string{"mp3", "wav", "opus", "pcm"},
	}
}

// TextToSpeech 文本转语音
func (a *AzureAdapter) TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error) {
	// Azure TTS实现
	return nil, NewAdapterError(a.GetProviderName(), "Azure适配器待实现", "NOT_IMPLEMENTED", nil)
}

// SpeechToText 语音转文本
func (a *AzureAdapter) SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error) {
	// Azure STT实现
	return nil, NewAdapterError(a.GetProviderName(), "Azure适配器待实现", "NOT_IMPLEMENTED", nil)
}

// CloneVoice 音色克隆
func (a *AzureAdapter) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	// Azure自定义神经语音实现
	return nil, NewAdapterError(a.GetProviderName(), "Azure适配器待实现", "NOT_IMPLEMENTED", nil)
}

// ListVoices 获取音色列表
func (a *AzureAdapter) ListVoices(ctx context.Context) ([]domain.VoiceTemplate, error) {
	return []domain.VoiceTemplate{}, nil
}

// DeleteVoice 删除音色
func (a *AzureAdapter) DeleteVoice(ctx context.Context, voiceID string) error {
	return NewAdapterError(a.GetProviderName(), "Azure适配器待实现", "NOT_IMPLEMENTED", nil)
}

// HealthCheck 健康检查
func (a *AzureAdapter) HealthCheck(ctx context.Context) error {
	return NewAdapterError(a.GetProviderName(), "Azure适配器待实现", "NOT_IMPLEMENTED", nil)
}

// ElevenLabsAdapter ElevenLabs语音服务适配器
type ElevenLabsAdapter struct {
	*BaseVoiceAdapter
	client *http.Client
}

// NewElevenLabsAdapter 创建ElevenLabs适配器
func NewElevenLabsAdapter(provider *domain.VoiceProvider) *ElevenLabsAdapter {
	return &ElevenLabsAdapter{
		BaseVoiceAdapter: NewBaseVoiceAdapter(provider),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetSupportedCapabilities 获取支持的能力
func (a *ElevenLabsAdapter) GetSupportedCapabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{
		TTS:        true,
		ASR:        false, // ElevenLabs主要专注于TTS
		VoiceClone: true,  // ElevenLabs的核心功能
		Emotions:   []string{"neutral", "excited", "sad", "angry", "whispering", "shouting"},
		Languages:  []string{"en-US", "zh-CN", "ja-JP", "ko-KR", "es-ES", "fr-FR", "de-DE"},
		AudioFormats: []string{"mp3", "wav", "opus", "pcm"},
	}
}

// TextToSpeech 文本转语音
func (a *ElevenLabsAdapter) TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error) {
	// ElevenLabs TTS实现
	return nil, NewAdapterError(a.GetProviderName(), "ElevenLabs适配器待实现", "NOT_IMPLEMENTED", nil)
}

// SpeechToText 语音转文本
func (a *ElevenLabsAdapter) SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error) {
	return nil, NewAdapterError(a.GetProviderName(), "ElevenLabs不支持语音识别", "NOT_SUPPORTED", nil)
}

// CloneVoice 音色克隆
func (a *ElevenLabsAdapter) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	// ElevenLabs音色克隆实现
	return nil, NewAdapterError(a.GetProviderName(), "ElevenLabs适配器待实现", "NOT_IMPLEMENTED", nil)
}

// ListVoices 获取音色列表
func (a *ElevenLabsAdapter) ListVoices(ctx context.Context) ([]domain.VoiceTemplate, error) {
	return []domain.VoiceTemplate{}, nil
}

// DeleteVoice 删除音色
func (a *ElevenLabsAdapter) DeleteVoice(ctx context.Context, voiceID string) error {
	return NewAdapterError(a.GetProviderName(), "ElevenLabs适配器待实现", "NOT_IMPLEMENTED", nil)
}

// HealthCheck 健康检查
func (a *ElevenLabsAdapter) HealthCheck(ctx context.Context) error {
	return NewAdapterError(a.GetProviderName(), "ElevenLabs适配器待实现", "NOT_IMPLEMENTED", nil)
}

// CustomAdapter 自定义适配器
// 允许用户完全自定义API调用方式
type CustomAdapter struct {
	*BaseVoiceAdapter
	client *http.Client
}

// NewCustomAdapter 创建自定义适配器
func NewCustomAdapter(provider *domain.VoiceProvider) *CustomAdapter {
	return &CustomAdapter{
		BaseVoiceAdapter: NewBaseVoiceAdapter(provider),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetSupportedCapabilities 获取支持的能力
func (a *CustomAdapter) GetSupportedCapabilities() domain.ProviderCapabilities {
	// 从配置中读取能力
	return domain.ProviderCapabilities{
		TTS:        true,
		ASR:        true,
		VoiceClone: true,
		Emotions:   []string{"neutral"},
		Languages:  []string{"en-US", "zh-CN"},
		AudioFormats: []string{"mp3", "wav"},
	}
}

// TextToSpeech 文本转语音
func (a *CustomAdapter) TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error) {
	// 自定义实现，通过配置文件定义API调用方式
	return nil, NewAdapterError(a.GetProviderName(), "自定义适配器需要配置实现", "NOT_CONFIGURED", nil)
}

// SpeechToText 语音转文本
func (a *CustomAdapter) SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error) {
	// 自定义实现，通过配置文件定义API调用方式
	return nil, NewAdapterError(a.GetProviderName(), "自定义适配器需要配置实现", "NOT_CONFIGURED", nil)
}

// CloneVoice 音色克隆
func (a *CustomAdapter) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	// 自定义实现，通过配置文件定义API调用方式
	return nil, NewAdapterError(a.GetProviderName(), "自定义适配器需要配置实现", "NOT_CONFIGURED", nil)
}

// ListVoices 获取音色列表
func (a *CustomAdapter) ListVoices(ctx context.Context) ([]domain.VoiceTemplate, error) {
	return []domain.VoiceTemplate{}, nil
}

// DeleteVoice 删除音色
func (a *CustomAdapter) DeleteVoice(ctx context.Context, voiceID string) error {
	return NewAdapterError(a.GetProviderName(), "自定义适配器需要配置实现", "NOT_CONFIGURED", nil)
}

// HealthCheck 健康检查
func (a *CustomAdapter) HealthCheck(ctx context.Context) error {
	return NewAdapterError(a.GetProviderName(), "自定义适配器需要配置实现", "NOT_CONFIGURED", nil)
}
