package domain

import (
	"time"

	"github.com/google/uuid"
)

// VoiceInputRequest 语音输入请求
type VoiceInputRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	CharacterID uuid.UUID `json:"character_id"`
	AudioData   []byte    `json:"audio_data"`
	AudioFormat string    `json:"audio_format"` // mp3, wav, etc.
	Language    string    `json:"language,omitempty"`
}

// 注意：VoiceCallResponse 已在 voice_provider.go 中定义

// AIResponseRequest AI回复请求
type AIResponseRequest struct {
	UserID        uuid.UUID              `json:"user_id"`
	CharacterID   uuid.UUID              `json:"character_id"`
	UserInput     string                 `json:"user_input"`
	Conversations []Conversation         `json:"conversations"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

// AIResponseResult AI回复结果
type AIResponseResult struct {
	Text     string                 `json:"text"`
	Emotion  string                 `json:"emotion"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Conversation 对话记录
type Conversation struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	Title       string    `json:"title" db:"title"`
	Status      string    `json:"status" db:"status"` // active, archived, deleted
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// 注意：VoiceModel, VoiceTemplate, CustomVoice, VoiceCallSession, VoiceMessage
// 已在 voice_provider.go 中定义，这里不重复定义

// VoiceRecord 语音记录（兼容性）
type VoiceRecord struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	AudioURL  string    `json:"audio_url" db:"audio_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// 注意：TTSRequest, ASRRequest, ASRResponse, VoiceCloneRequest, VoiceCloneResponse
// 已在 voice_provider.go 中定义，这里不重复定义

// 注意：ProviderCapabilities, VoiceProviderConfig 已在 voice_provider.go 中定义

// ChatRequest 聊天请求
type ChatRequest struct {
	UserID      uuid.UUID              `json:"user_id"`
	CharacterID uuid.UUID              `json:"character_id"`
	Message     string                 `json:"message"`
	Model       string                 `json:"model,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	MessageID uuid.UUID `json:"message_id"`
	Message   string    `json:"message"`
	Emotion   string    `json:"emotion"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatWithVoiceResponse 带语音的聊天响应
type ChatWithVoiceResponse struct {
	MessageID   uuid.UUID `json:"message_id"`
	Message     string    `json:"message"`
	AudioData   []byte    `json:"audio_data,omitempty"`
	AudioFormat string    `json:"audio_format,omitempty"`
	Emotion     string    `json:"emotion"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// UserMessageRequest 用户消息请求
type UserMessageRequest struct {
	UserID      uuid.UUID              `json:"user_id"`
	CharacterID uuid.UUID              `json:"character_id"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"` // text, voice, image
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UserMessageResponse 用户消息响应
type UserMessageResponse struct {
	MessageID uuid.UUID `json:"message_id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// SaveConversationRequest 保存对话请求
type SaveConversationRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	CharacterID uuid.UUID `json:"character_id"`
	UserText    string    `json:"user_text"`
	AIText      string    `json:"ai_text"`
	UserEmotion string    `json:"user_emotion,omitempty"`
	AIEmotion   string    `json:"ai_emotion,omitempty"`
}

// Voice 音色信息
type Voice struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	VoiceName   string    `json:"voice_name" db:"voice_name"`
	VoiceID     string    `json:"voice_id" db:"voice_id"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
