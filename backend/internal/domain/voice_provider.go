package domain

import (
	"encoding/json"
	"time"
)

// VoiceProvider 语音服务提供商
type VoiceProvider struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	DisplayName string                 `json:"display_name" db:"display_name"`
	Type        string                 `json:"type" db:"type"` // tts, asr, both
	BaseURL     string                 `json:"base_url" db:"base_url"`
	APIKey      string                 `json:"api_key" db:"api_key"`
	Config      json.RawMessage        `json:"config" db:"config"`
	Status      string                 `json:"status" db:"status"` // active, inactive
	Priority    int                    `json:"priority" db:"priority"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// VoiceModel 语音模型配置
type VoiceModel struct {
	ID           string          `json:"id" db:"id"`
	ProviderID   string          `json:"provider_id" db:"provider_id"`
	ModelKey     string          `json:"model_key" db:"model_key"`
	DisplayName  string          `json:"display_name" db:"display_name"`
	Type         string          `json:"type" db:"type"` // tts, asr
	Capabilities json.RawMessage `json:"capabilities" db:"capabilities"`
	Pricing      json.RawMessage `json:"pricing" db:"pricing"`
	Config       json.RawMessage `json:"config" db:"config"`
	Status       string          `json:"status" db:"status"`
	Weight       int             `json:"weight" db:"weight"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// VoiceTemplate 音色模板
type VoiceTemplate struct {
	ID           string          `json:"id" db:"id"`
	ProviderID   string          `json:"provider_id" db:"provider_id"`
	VoiceKey     string          `json:"voice_key" db:"voice_key"`
	DisplayName  string          `json:"display_name" db:"display_name"`
	Language     string          `json:"language" db:"language"`
	Gender       string          `json:"gender" db:"gender"`
	Age          string          `json:"age" db:"age"`
	Style        string          `json:"style" db:"style"`
	Description  string          `json:"description" db:"description"`
	SampleURL    string          `json:"sample_url" db:"sample_url"`
	CoverImage   string          `json:"cover_image" db:"cover_image"`
	Config       json.RawMessage `json:"config" db:"config"`
	IsDefault    bool            `json:"is_default" db:"is_default"`
	Status       string          `json:"status" db:"status"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// CustomVoice 用户自定义音色
type CustomVoice struct {
	ID            string          `json:"id" db:"id"`
	UserID        string          `json:"user_id" db:"user_id"`
	ProviderID    string          `json:"provider_id" db:"provider_id"`
	VoiceName     string          `json:"voice_name" db:"voice_name"`
	VoiceID       string          `json:"voice_id" db:"voice_id"` // 提供商返回的ID
	CoverImage    string          `json:"cover_image" db:"cover_image"`
	ReferenceText string          `json:"reference_text" db:"reference_text"`
	AudioURL      string          `json:"audio_url" db:"audio_url"`
	Config        json.RawMessage `json:"config" db:"config"`
	Status        string          `json:"status" db:"status"` // processing, ready, failed
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// VoiceCallSession 语音通话会话
type VoiceCallSession struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	CharacterID string    `json:"character_id" db:"character_id"`
	SessionType string    `json:"session_type" db:"session_type"` // incoming, outgoing, auto
	Status      string    `json:"status" db:"status"` // active, ended, failed
	StartTime   time.Time `json:"start_time" db:"start_time"`
	EndTime     *time.Time `json:"end_time" db:"end_time"`
	Duration    int       `json:"duration" db:"duration"` // 秒
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// VoiceMessage 语音消息
type VoiceMessage struct {
	ID        string          `json:"id" db:"id"`
	SessionID string          `json:"session_id" db:"session_id"`
	Type      string          `json:"type" db:"type"` // user, ai
	AudioURL  string          `json:"audio_url" db:"audio_url"`
	Text      string          `json:"text" db:"text"`
	Emotion   string          `json:"emotion" db:"emotion"`
	Language  string          `json:"language" db:"language"`
	Duration  float64         `json:"duration" db:"duration"`
	Config    json.RawMessage `json:"config" db:"config"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// 请求和响应结构

// CreateProviderRequest 创建提供商请求
type CreateProviderRequest struct {
	Name        string                 `json:"name" validate:"required"`
	DisplayName string                 `json:"display_name" validate:"required"`
	Type        string                 `json:"type" validate:"required,oneof=tts asr both"`
	BaseURL     string                 `json:"base_url" validate:"required,url"`
	APIKey      string                 `json:"api_key" validate:"required"`
	Config      map[string]interface{} `json:"config"`
	Priority    int                    `json:"priority"`
}

// CreateVoiceModelRequest 创建语音模型请求
type CreateVoiceModelRequest struct {
	ProviderID   string                 `json:"provider_id" validate:"required"`
	ModelKey     string                 `json:"model_key" validate:"required"`
	DisplayName  string                 `json:"display_name" validate:"required"`
	Type         string                 `json:"type" validate:"required,oneof=tts asr"`
	Capabilities []string               `json:"capabilities"`
	Pricing      map[string]interface{} `json:"pricing"`
	Config       map[string]interface{} `json:"config"`
	Weight       int                    `json:"weight"`
}

// CreateVoiceTemplateRequest 创建音色模板请求
type CreateVoiceTemplateRequest struct {
	ProviderID  string                 `json:"provider_id" validate:"required"`
	VoiceKey    string                 `json:"voice_key" validate:"required"`
	DisplayName string                 `json:"display_name" validate:"required"`
	Language    string                 `json:"language" validate:"required"`
	Gender      string                 `json:"gender" validate:"oneof=male female neutral"`
	Age         string                 `json:"age"`
	Style       string                 `json:"style"`
	Description string                 `json:"description"`
	SampleURL   string                 `json:"sample_url"`
	CoverImage  string                 `json:"cover_image"`
	Config      map[string]interface{} `json:"config"`
	IsDefault   bool                   `json:"is_default"`
}

// VoiceCallRequest 语音通话请求
type VoiceCallRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	CharacterID string `json:"character_id" validate:"required"`
	AudioData   []byte `json:"audio_data" validate:"required"`
	SessionID   string `json:"session_id"`
}

// VoiceCallResponse 语音通话响应
type VoiceCallResponse struct {
	SessionID    string `json:"session_id"`
	AudioData    []byte `json:"audio_data"`
	AudioURL     string `json:"audio_url"`
	ResponseText string `json:"response_text"`
	Emotion      string `json:"emotion"`
	Language     string `json:"language"`
	Duration     int64  `json:"duration"` // 处理时间(毫秒)
}

// VoiceCloneRequest 音色克隆请求
type VoiceCloneRequest struct {
	UserID        string `json:"user_id" validate:"required"`
	ProviderID    string `json:"provider_id" validate:"required"`
	VoiceName     string `json:"voice_name" validate:"required"`
	AudioData     []byte `json:"audio_data" validate:"required"`
	AudioBase64   string `json:"audio_base64"`
	ReferenceText string `json:"reference_text" validate:"required"`
	CoverImage    string `json:"cover_image"`
}

// VoiceCloneResponse 音色克隆响应
type VoiceCloneResponse struct {
	VoiceID    string `json:"voice_id"`
	VoiceName  string `json:"voice_name"`
	CoverImage string `json:"cover_image"`
	Status     string `json:"status"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// AutoCallRequest 自动外呼请求
type AutoCallRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	CharacterID string `json:"character_id" validate:"required"`
	Reason      string `json:"reason"` // 外呼原因
	Context     string `json:"context"` // 上下文信息
}

// AutoCallResponse 自动外呼响应
type AutoCallResponse struct {
	SessionID   string `json:"session_id"`
	AudioData   []byte `json:"audio_data"`
	AudioURL    string `json:"audio_url"`
	GreetingText string `json:"greeting_text"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
}

// TTSRequest 通用TTS请求
type TTSRequest struct {
	ProviderID     string                 `json:"provider_id"`
	ModelKey       string                 `json:"model_key"`
	Text           string                 `json:"text" validate:"required"`
	VoiceID        string                 `json:"voice_id"`
	Language       string                 `json:"language"`
	Speed          float64                `json:"speed"`
	Pitch          float64                `json:"pitch"`
	Volume         float64                `json:"volume"`
	Emotion        string                 `json:"emotion"`
	Style          string                 `json:"style"`
	ResponseFormat string                 `json:"response_format"`
	Config         map[string]interface{} `json:"config"`
}

// ASRRequest 通用ASR请求
type ASRRequest struct {
	ProviderID string                 `json:"provider_id"`
	ModelKey   string                 `json:"model_key"`
	AudioData  []byte                 `json:"audio_data" validate:"required"`
	Language   string                 `json:"language"`
	Config     map[string]interface{} `json:"config"`
}

// ASRResponse 通用ASR响应
type ASRResponse struct {
	Text      string            `json:"text"`
	Language  string            `json:"language"`
	Emotion   string            `json:"emotion"`
	Events    []string          `json:"events"`
	Segments  []ASRSegment      `json:"segments"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ASRSegment ASR分段结果
type ASRSegment struct {
	Text      string  `json:"text"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Confidence float64 `json:"confidence"`
}

// ProviderCapabilities 提供商能力
type ProviderCapabilities struct {
	TTS         bool     `json:"tts"`
	ASR         bool     `json:"asr"`
	VoiceClone  bool     `json:"voice_clone"`
	Emotions    []string `json:"emotions"`
	Languages   []string `json:"languages"`
	AudioFormats []string `json:"audio_formats"`
}

// VoiceProviderConfig 提供商配置
type VoiceProviderConfig struct {
	Timeout         int                    `json:"timeout"`
	RetryCount      int                    `json:"retry_count"`
	RateLimit       map[string]int         `json:"rate_limit"`
	DefaultParams   map[string]interface{} `json:"default_params"`
	Authentication  map[string]string      `json:"authentication"`
	Endpoints       map[string]string      `json:"endpoints"`
	Capabilities    ProviderCapabilities   `json:"capabilities"`
}
