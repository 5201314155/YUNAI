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

// SiliconFlowAdapter SiliconFlow语音服务适配器
type SiliconFlowAdapter struct {
	*BaseVoiceAdapter
	client *http.Client
}

// NewSiliconFlowAdapter 创建SiliconFlow适配器
func NewSiliconFlowAdapter(provider *domain.VoiceProvider) *SiliconFlowAdapter {
	return &SiliconFlowAdapter{
		BaseVoiceAdapter: NewBaseVoiceAdapter(provider),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetSupportedCapabilities 获取支持的能力
func (a *SiliconFlowAdapter) GetSupportedCapabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{
		TTS:        true,
		ASR:        true,
		VoiceClone: true,
		Emotions:   []string{"neutral", "happy", "sad", "angry", "excited"},
		Languages:  []string{"zh-CN", "en-US", "ja-JP", "ko-KR"},
		AudioFormats: []string{"mp3", "wav", "opus", "pcm"},
	}
}

// TextToSpeech 文本转语音
func (a *SiliconFlowAdapter) TextToSpeech(ctx context.Context, req *domain.TTSRequest) ([]byte, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":           req.ModelKey,
		"input":           req.Text,
		"voice":           req.VoiceID,
		"response_format": req.ResponseFormat,
	}

	// 添加可选参数
	if req.Speed > 0 {
		requestBody["speed"] = req.Speed
	}
	if req.Volume > 0 {
		requestBody["gain"] = req.Volume
	}

	// 处理情感和风格
	if req.Emotion != "" || req.Style != "" {
		input := req.Text
		if req.Emotion != "" {
			input = fmt.Sprintf("你能用%s的情感说吗？<|endofprompt|>%s", req.Emotion, req.Text)
		}
		requestBody["input"] = input
	}

	// 添加自定义配置
	if len(req.Config) > 0 {
		for key, value := range req.Config {
			requestBody[key] = value
		}
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
func (a *SiliconFlowAdapter) SpeechToText(ctx context.Context, req *domain.ASRRequest) (*domain.ASRResponse, error) {
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
	writer.WriteField("model", req.ModelKey)
	if req.Language != "" {
		writer.WriteField("language", req.Language)
	} else {
		writer.WriteField("language", "auto")
	}

	// 添加自定义配置
	if len(req.Config) > 0 {
		for key, value := range req.Config {
			if strValue, ok := value.(string); ok {
				writer.WriteField(key, strValue)
			}
		}
	}

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
		Text     string `json:"text"`
		Language string `json:"language,omitempty"`
		// SenseVoice特有字段
		Emotion string   `json:"emotion,omitempty"`
		Events  []string `json:"events,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "解析响应失败", "PARSE_ERROR", err)
	}

	return &domain.ASRResponse{
		Text:     result.Text,
		Language: result.Language,
		Emotion:  result.Emotion,
		Events:   result.Events,
		Metadata: map[string]interface{}{
			"provider": a.GetProviderName(),
			"model":    req.ModelKey,
		},
	}, nil
}

// CloneVoice 音色克隆
func (a *SiliconFlowAdapter) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":      "FunAudioLLM/CosyVoice2-0.5B",
		"customName": req.VoiceName,
		"text":       req.ReferenceText,
	}

	// 处理音频数据
	if req.AudioBase64 != "" {
		requestBody["audio"] = fmt.Sprintf("data:audio/mpeg;base64,%s", req.AudioBase64)
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "序列化请求失败", "MARSHAL_ERROR", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.Provider.BaseURL+"/uploads/audio/voice", bytes.NewBuffer(reqBody))
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

	var result struct {
		URI string `json:"uri"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "解析响应失败", "PARSE_ERROR", err)
	}

	return &domain.VoiceCloneResponse{
		VoiceID:   result.URI,
		VoiceName: req.VoiceName,
		Status:    "ready",
		Success:   true,
		Message:   "音色克隆成功",
	}, nil
}

// ListVoices 获取音色列表
func (a *SiliconFlowAdapter) ListVoices(ctx context.Context) ([]domain.VoiceTemplate, error) {
	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", a.Provider.BaseURL+"/audio/voice/list", nil)
	if err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "创建请求失败", "REQUEST_ERROR", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.Provider.APIKey)

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
		Data []struct {
			URI  string `json:"uri"`
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewAdapterError(a.GetProviderName(), "解析响应失败", "PARSE_ERROR", err)
	}

	var voices []domain.VoiceTemplate
	for _, voice := range result.Data {
		voices = append(voices, domain.VoiceTemplate{
			VoiceKey:    voice.URI,
			DisplayName: voice.Name,
			ProviderID:  a.Provider.ID,
			Status:      "active",
		})
	}

	return voices, nil
}

// DeleteVoice 删除音色
func (a *SiliconFlowAdapter) DeleteVoice(ctx context.Context, voiceID string) error {
	requestBody := map[string]interface{}{
		"uri": voiceID,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return NewAdapterError(a.GetProviderName(), "序列化请求失败", "MARSHAL_ERROR", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.Provider.BaseURL+"/audio/voice/deletions", bytes.NewBuffer(reqBody))
	if err != nil {
		return NewAdapterError(a.GetProviderName(), "创建请求失败", "REQUEST_ERROR", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.Provider.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return NewAdapterError(a.GetProviderName(), "发送请求失败", "HTTP_ERROR", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return NewAdapterError(a.GetProviderName(), 
			fmt.Sprintf("API请求失败: %d, %s", resp.StatusCode, string(body)), 
			"API_ERROR", nil)
	}

	return nil
}

// HealthCheck 健康检查
func (a *SiliconFlowAdapter) HealthCheck(ctx context.Context) error {
	// 发送简单的TTS请求来检查服务状态
	req := &domain.TTSRequest{
		ModelKey:       "FunAudioLLM/CosyVoice2-0.5B",
		Text:           "健康检查",
		VoiceID:        "FunAudioLLM/CosyVoice2-0.5B:alex",
		ResponseFormat: "mp3",
	}

	_, err := a.TextToSpeech(ctx, req)
	if err != nil {
		return NewAdapterError(a.GetProviderName(), "健康检查失败", "HEALTH_CHECK_ERROR", err)
	}

	return nil
}

// ValidateConfig 验证配置
func (a *SiliconFlowAdapter) ValidateConfig(config map[string]interface{}) error {
	// 验证必需的配置项
	if a.Provider.BaseURL == "" {
		return NewAdapterError(a.GetProviderName(), "BaseURL不能为空", "CONFIG_ERROR", nil)
	}
	if a.Provider.APIKey == "" {
		return NewAdapterError(a.GetProviderName(), "APIKey不能为空", "CONFIG_ERROR", nil)
	}

	return nil
}
