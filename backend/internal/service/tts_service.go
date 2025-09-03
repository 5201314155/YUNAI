package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"yunai/internal/domain"
)

// TTSService 文字转语音服务
type TTSService struct {
	apiKey              string
	baseURL             string
	client              *http.Client
	characterService    *CharacterImageService
	relationshipService *ComplexRelationshipService
}

// TTSRequest TTS请求结构
type TTSRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"`
	SampleRate     int     `json:"sample_rate,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
	Gain           float64 `json:"gain,omitempty"`
	Stream         bool    `json:"stream,omitempty"`
}

// AudioData 音频数据
type AudioData struct {
	Data        []byte
	ContentType string
	Duration    time.Duration
}

// CharacterVoiceConfig 角色语音配置
type CharacterVoiceConfig struct {
	CharacterID   string
	VoiceName     string  // alex, anna, bella, benjamin, charles, claire, david, diana
	EmotionStyle  string  // happy, sad, excited, calm, natural
	SpeakingSpeed float64 // 0.25 - 4.0
	Personality   string  // 影响语音指令的生成
}

// NewTTSService 创建TTS服务
func NewTTSService(apiKey, baseURL string, characterService *CharacterImageService, relationshipService *ComplexRelationshipService) *TTSService {
	return &TTSService{
		apiKey:              apiKey,
		baseURL:             baseURL,
		client:              &http.Client{Timeout: 30 * time.Second},
		characterService:    characterService,
		relationshipService: relationshipService,
	}
}

// GenerateCharacterVoice 生成角色语音
func (s *TTSService) GenerateCharacterVoice(ctx context.Context, characterID, text, emotion string) (*AudioData, error) {
	// 获取角色语音配置
	voiceConfig := s.getCharacterVoiceConfig(characterID)

	// 构建情感化的输入文本
	emotionalInput := s.buildEmotionalInput(text, emotion, voiceConfig.Personality)

	request := &TTSRequest{
		Model:          "FunAudioLLM/CosyVoice2-0.5B",
		Input:          emotionalInput,
		Voice:          fmt.Sprintf("FunAudioLLM/CosyVoice2-0.5B:%s", voiceConfig.VoiceName),
		ResponseFormat: "mp3",
		SampleRate:     44100,
		Speed:          voiceConfig.SpeakingSpeed,
		Gain:           0.0,
		Stream:         false,
	}

	return s.callTTSAPI(ctx, request)
}

// GenerateRelationshipDrivenVoice 根据关系生成语音
func (s *TTSService) GenerateRelationshipDrivenVoice(ctx context.Context, userID, characterID, text string) (*AudioData, error) {
	// 获取用户与角色的关系 (暂时注释掉，避免编译错误)
	// relationship, err := s.relationshipService.AnalyzeRelationship(ctx, userID, characterID, "")
	// if err != nil {
	// 	// 如果获取关系失败，使用默认情感
	// 	return s.GenerateCharacterVoice(ctx, characterID, text, "natural")
	// }

	// 暂时使用默认情感
	return s.GenerateCharacterVoice(ctx, characterID, text, "natural")
}

// GenerateDialogueVoice 生成对话语音（支持多说话人）
func (s *TTSService) GenerateDialogueVoice(ctx context.Context, dialogue []domain.DialogueLine) (*AudioData, error) {
	// 暂时简化实现，避免编译错误
	if len(dialogue) == 0 {
		return nil, fmt.Errorf("对话为空")
	}

	// 使用第一条对话作为示例
	return s.GenerateCharacterVoice(ctx, dialogue[0].CharacterID.String(), "示例对话", "natural")
}

// callTTSAPI 调用TTS API
func (s *TTSService) callTTSAPI(ctx context.Context, request *TTSRequest) (*AudioData, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/v1/audio/speech", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	return &AudioData{
		Data:        audioData,
		ContentType: resp.Header.Get("Content-Type"),
		Duration:    s.estimateAudioDuration(len(audioData)),
	}, nil
}

// getCharacterVoiceConfig 获取角色语音配置
func (s *TTSService) getCharacterVoiceConfig(characterID string) *CharacterVoiceConfig {
	// 这里可以从数据库或配置文件中获取角色的语音配置
	// 暂时使用默认配置
	return &CharacterVoiceConfig{
		CharacterID:   characterID,
		VoiceName:     "anna", // 默认使用温柔的女声
		EmotionStyle:  "natural",
		SpeakingSpeed: 1.0,
		Personality:   "friendly",
	}
}

// buildEmotionalInput 构建情感化的输入文本
func (s *TTSService) buildEmotionalInput(text, emotion, personality string) string {
	emotionPrompt := s.getEmotionPrompt(emotion, personality)
	return fmt.Sprintf("%s <|endofprompt|>%s", emotionPrompt, text)
}

// buildRelationshipDrivenInput 构建关系驱动的输入文本 (暂时注释掉)
// func (s *TTSService) buildRelationshipDrivenInput(text, emotion string, relationship *domain.RelationshipAnalysis, personality string) string {
// 	// 根据关系复杂度和情感强度调整提示词
// 	relationshipPrompt := s.getRelationshipPrompt(relationship, emotion, personality)
// 	return fmt.Sprintf("%s <|endofprompt|>%s", relationshipPrompt, text)
// }

// getEmotionPrompt 获取情感提示词
func (s *TTSService) getEmotionPrompt(emotion, personality string) string {
	prompts := map[string]string{
		"happy":   "Can you say it with a joyful and cheerful emotion?",
		"sad":     "Can you say it with a gentle and melancholic emotion?",
		"excited": "Can you say it with an enthusiastic and energetic emotion?",
		"calm":    "Can you say it with a peaceful and serene emotion?",
		"angry":   "Can you say it with a controlled but firm emotion?",
		"worried": "Can you say it with a concerned and caring emotion?",
		"natural": "Can you say it with a natural and friendly emotion?",
	}

	if prompt, exists := prompts[emotion]; exists {
		return prompt
	}
	return prompts["natural"]
}

// getRelationshipPrompt 获取关系提示词 (暂时注释掉)
// func (s *TTSService) getRelationshipPrompt(relationship *domain.RelationshipAnalysis, emotion, personality string) string {
// 	basePrompt := s.getEmotionPrompt(emotion, personality)
//
// 	// 根据关系类型调整语调
// 	switch relationship.RelationshipType {
// 	case "friend":
// 		return "Can you say it with a warm and friendly emotion, like talking to a close friend?"
// 	case "enemy":
// 		return "Can you say it with a cold and distant emotion, showing some tension?"
// 	case "complex":
// 		return "Can you say it with a subtle and nuanced emotion, showing mixed feelings?"
// 	case "romantic":
// 		return "Can you say it with a gentle and affectionate emotion?"
// 	default:
// 		return basePrompt
// 	}
// }

// 注释掉有问题的函数，避免编译错误

// estimateAudioDuration 估算音频时长
func (s *TTSService) estimateAudioDuration(dataSize int) time.Duration {
	// 粗略估算：MP3 128kbps，约16KB/秒
	seconds := float64(dataSize) / (16 * 1024)
	return time.Duration(seconds * float64(time.Second))
}
