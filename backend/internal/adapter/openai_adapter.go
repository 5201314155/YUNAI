package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"yunai/internal/domain"
)

// OpenAIAdapter OpenAI语音服务适配器
type OpenAIAdapter struct {
	*BaseVoiceAdapter
	client *http.Client
}

// NewOpenAIAdapter 创建OpenAI适配器
func NewOpenAIAdapter(provider *domain.VoiceProvider) *OpenAIAdapter {
	return &OpenAIAdapter{
		BaseVoiceAdapter: NewBaseVoiceAdapter(provider),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetSupportedCapabilities 获取支持的能力
func (a *OpenAIAdapter) GetSupportedCapabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{
		TTS:        true,
		ASR:        true,
		VoiceClone: false, // OpenAI不支持音色克隆
		Emotions:   []string{"neutral"},
		Languages:  []string{"en-US", "zh-CN", "ja-JP", "ko-KR", "es-ES", "fr-FR", "de-DE"},
		AudioFormats: []string{"mp3", "opus", "aac", "flac"},
	}
}

// TextToSpeech 文本转语音
func (a *OpenAIAdapter) TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error) {
	// OpenAI TTS API请求体
	requestBody := map[string]interface{}{
		"model": req.ModelKey, // tts-1 或 tts-1-hd
		"input": req.Text,
		"voice": a.mapVoiceID(req.VoiceID), // alloy, echo, fable, onyx, nova, shimmer
	}

	// 添加可选参数
	if req.ResponseFormat != "" {
		requestBody["response_format"] = req.ResponseFormat
	} else {
		requestBody["response_format"] = "mp3"
	}

	if req.Speed > 0 {
		requestBody["speed"] = req.Speed
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "序列化请求失败", "MARSHAL_ERROR", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.Provider.BaseURL+"/audio/speech", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "创建请求失败", "REQUEST_ERROR", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.Provider.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "发送请求失败", "HTTP_ERROR", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, NewAdapterError(a.GetProviderName(), 
			fmt.Sprintf("API请求失败: %d, %s", resp.StatusCode, string(body)), 
			"API_ERROR", nil)
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "读取音频数据失败", "READ_ERROR", err)
	}

	return audioData, nil
}

// SpeechToText 语音转文本
func (a *OpenAIAdapter) SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error) {
	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加音频文件
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "创建表单文件失败", "FORM_ERROR", err)
	}
	_, err = part.Write(req.AudioData)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "写入音频数据失败", "WRITE_ERROR", err)
	}

	// 添加模型参数
	writer.WriteField("model", req.ModelKey) // whisper-1

	// 添加语言参数
	if req.Language != "" && req.Language != "auto" {
		writer.WriteField("language", req.Language)
	}

	// 添加响应格式
	writer.WriteField("response_format", "json")

	writer.Close()

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.Provider.BaseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "创建请求失败", "REQUEST_ERROR", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.Provider.APIKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "发送请求失败", "HTTP_ERROR", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, NewAdapterError(a.GetProviderName(), 
			fmt.Sprintf("API请求失败: %d, %s", resp.StatusCode, string(body)), 
			"API_ERROR", nil)
	}

	var result struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "解析响应失败", "PARSE_ERROR", err)
	}

	return &domain.ASRResponse{
		Text:     result.Text,
		Language: req.Language,
		Metadata: map[string]interface{}{
			"provider": a.GetProviderName(),
			"model":    req.ModelKey,
		},
	}, nil
}

// CloneVoice 音色克隆 (OpenAI不支持)
func (a *OpenAIAdapter) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	return nil, NewAdapterError(a.GetProviderName(), "OpenAI不支持音色克隆", "NOT_SUPPORTED", nil)
}

// ListVoices 获取音色列表
func (a *OpenAIAdapter) ListVoices(ctx context.Context) ([]domain.VoiceTemplate, error) {
	// OpenAI的预定义音色
	voices := []domain.VoiceTemplate{
		{
			VoiceKey:    "alloy",
			DisplayName: "Alloy",
			Language:    "en-US",
			Gender:      "neutral",
			Description: "A balanced and versatile voice",
			ProviderID:  a.Provider.ID,
			IsDefault:   true,
			Status:      "active",
		},
		{
			VoiceKey:    "echo",
			DisplayName: "Echo",
			Language:    "en-US",
			Gender:      "male",
			Description: "A clear and articulate male voice",
			ProviderID:  a.Provider.ID,
			Status:      "active",
		},
		{
			VoiceKey:    "fable",
			DisplayName: "Fable",
			Language:    "en-US",
			Gender:      "male",
			Description: "A warm and expressive male voice",
			ProviderID:  a.Provider.ID,
			Status:      "active",
		},
		{
			VoiceKey:    "onyx",
			DisplayName: "Onyx",
			Language:    "en-US",
			Gender:      "male",
			Description: "A deep and resonant male voice",
			ProviderID:  a.Provider.ID,
			Status:      "active",
		},
		{
			VoiceKey:    "nova",
			DisplayName: "Nova",
			Language:    "en-US",
			Gender:      "female",
			Description: "A bright and energetic female voice",
			ProviderID:  a.Provider.ID,
			Status:      "active",
		},
		{
			VoiceKey:    "shimmer",
			DisplayName: "Shimmer",
			Language:    "en-US",
			Gender:      "female",
			Description: "A soft and gentle female voice",
			ProviderID:  a.Provider.ID,
			Status:      "active",
		},
	}

	return voices, nil
}

// DeleteVoice 删除音色 (OpenAI不支持)
func (a *OpenAIAdapter) DeleteVoice(ctx context.Context, voiceID string) error {
	return NewAdapterError(a.GetProviderName(), "OpenAI不支持删除音色", "NOT_SUPPORTED", nil)
}

// HealthCheck 健康检查
func (a *OpenAIAdapter) HealthCheck(ctx context.Context) error {
	// 发送简单的TTS请求来检查服务状态
	req := &domain.TTSRequest{
		ModelKey:       "tts-1",
		Text:           "Health check",
		VoiceID:        "alloy",
		ResponseFormat: "mp3",
	}

	_, err := a.TextToSpeech(ctx, req)
	if err != nil {
		return NewAdapterError(a.GetProviderName(), "健康检查失败", "HEALTH_CHECK_ERROR", err)
	}

	return nil
}

// ValidateConfig 验证配置
func (a *OpenAIAdapter) ValidateConfig(config map[string]interface{}) error {
	// 验证必需的配置项
	if a.Provider.BaseURL == "" {
		return NewAdapterError(a.GetProviderName(), "BaseURL不能为空", "CONFIG_ERROR", nil)
	}
	if a.Provider.APIKey == "" {
		return NewAdapterError(a.GetProviderName(), "APIKey不能为空", "CONFIG_ERROR", nil)
	}

	return nil
}

// mapVoiceID 映射音色ID
func (a *OpenAIAdapter) mapVoiceID(voiceID string) string {
	// 如果是完整的音色ID，提取实际的音色名称
	if voiceID == "" {
		return "alloy" // 默认音色
	}

	// 支持的OpenAI音色
	validVoices := map[string]bool{
		"alloy":   true,
		"echo":    true,
		"fable":   true,
		"onyx":    true,
		"nova":    true,
		"shimmer": true,
	}

	if validVoices[voiceID] {
		return voiceID
	}

	// 如果不是有效的音色，返回默认音色
	return "alloy"
}
