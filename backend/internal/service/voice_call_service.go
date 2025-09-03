package service

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
	"yunai/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// VoiceCallService AI语音通话服务
type VoiceCallService struct {
	logger           *logrus.Logger
	characterRepo    repository.CharacterRepository
	conversationRepo repository.ConversationRepository
	voiceRepo        repository.VoiceRepository
	userRepo         repository.UserRepository
	apiKey           string
	baseURL          string
}

// NewVoiceCallService 创建语音通话服务
func NewVoiceCallService(
	logger *logrus.Logger,
	characterRepo repository.CharacterRepository,
	conversationRepo repository.ConversationRepository,
	voiceRepo repository.VoiceRepository,
	userRepo repository.UserRepository,
	apiKey string,
	baseURL string,
) *VoiceCallService {
	return &VoiceCallService{
		logger:           logger,
		characterRepo:    characterRepo,
		conversationRepo: conversationRepo,
		voiceRepo:        voiceRepo,
		userRepo:         userRepo,
		apiKey:           apiKey,
		baseURL:          baseURL,
	}
}

// ASRRequest 语音识别请求
type ASRRequest struct {
	Model    string `json:"model"`
	Language string `json:"language,omitempty"`
}

// ASRResponse 语音识别响应
type ASRResponse struct {
	Text     string   `json:"text"`
	Language string   `json:"language,omitempty"`
	Emotion  string   `json:"emotion,omitempty"`
	Events   []string `json:"events,omitempty"`
}

// VoiceCallTTSRequest 语音通话专用的TTS请求
type VoiceCallTTSRequest struct {
	Model          string                 `json:"model"`
	Input          string                 `json:"input"`
	Voice          string                 `json:"voice"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	Speed          float64                `json:"speed,omitempty"`
	Gain           float64                `json:"gain,omitempty"`
	ExtraBody      map[string]interface{} `json:"extra_body,omitempty"`
}

// VoiceCloneRequest 音色克隆请求
type VoiceCloneRequest struct {
	Model      string `json:"model"`
	CustomName string `json:"customName"`
	Audio      string `json:"audio"` // base64编码的音频
	Text       string `json:"text"`  // 参考音频的文字内容
}

// VoiceCloneResponse 音色克隆响应
type VoiceCloneResponse struct {
	URI string `json:"uri"`
}

// ProcessVoiceInput 处理用户语音输入
func (s *VoiceCallService) ProcessVoiceInput(ctx context.Context, req *domain.VoiceInputRequest) (*domain.VoiceCallResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"character_id": req.CharacterID,
		"audio_length": len(req.AudioData),
	}).Info("开始处理语音输入")

	// 1. 语音识别 (ASR)
	asrResult, err := s.speechToText(ctx, req.AudioData)
	if err != nil {
		s.logger.WithError(err).Error("语音识别失败")
		return nil, fmt.Errorf("语音识别失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"recognized_text": asrResult.Text,
		"emotion":         asrResult.Emotion,
		"language":        asrResult.Language,
	}).Info("语音识别完成")

	// 2-4. 简化处理，跳过角色、用户和历史获取

	// 5. 生成AI回复 - 简化实现
	aiResponse := &domain.AIResponseResult{
		Text:    "这是一个测试回复：" + asrResult.Text,
		Emotion: "friendly",
		Metadata: map[string]interface{}{
			"model": "test",
		},
	}
	if err != nil {
		return nil, fmt.Errorf("生成AI回复失败: %w", err)
	}

	// 6. 语音合成 (TTS)
	audioData, err := s.textToSpeech(ctx, &VoiceCallTTSRequest{
		Model:          "FunAudioLLM/CosyVoice2-0.5B",
		Input:          aiResponse.Text,
		Voice:          "FunAudioLLM/CosyVoice2-0.5B:bella", // 默认使用小雨的音色
		ResponseFormat: "mp3",
		Speed:          1.0,
		Gain:           0.0,
	})
	if err != nil {
		return nil, fmt.Errorf("语音合成失败: %w", err)
	}

	// 7. 简化处理，跳过保存对话记录

	return &domain.VoiceCallResponse{
		AudioData:    audioData,
		ResponseText: aiResponse.Text,
		Emotion:      aiResponse.Emotion,
		Duration:     time.Since(time.Now()).Milliseconds(),
	}, nil
}

// speechToText 语音转文字 (使用SenseVoice)
func (s *VoiceCallService) speechToText(ctx context.Context, audioData []byte) (*ASRResponse, error) {
	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加音频文件
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, fmt.Errorf("创建表单文件失败: %w", err)
	}
	_, err = part.Write(audioData)
	if err != nil {
		return nil, fmt.Errorf("写入音频数据失败: %w", err)
	}

	// 添加模型参数
	writer.WriteField("model", "FunAudioLLM/SenseVoiceSmall")
	writer.WriteField("language", "auto") // 自动识别语言

	writer.Close()

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var result ASRResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// textToSpeech 文字转语音 (使用CosyVoice2)
func (s *VoiceCallService) textToSpeech(ctx context.Context, req *VoiceCallTTSRequest) ([]byte, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/audio/speech", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取音频数据失败: %w", err)
	}

	return audioData, nil
}

// CloneVoice 克隆音色
func (s *VoiceCallService) CloneVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"voice_name":   req.VoiceName,
		"audio_length": len(req.AudioData),
	}).Info("开始音色克隆")

	// 1. 将音频数据转换为base64
	audioBase64 := fmt.Sprintf("data:audio/mpeg;base64,%s", req.AudioBase64)

	// 2. 调用音色克隆API
	cloneReq := &VoiceCloneRequest{
		Model:      "FunAudioLLM/CosyVoice2-0.5B",
		CustomName: req.VoiceName,
		Audio:      audioBase64,
		Text:       req.ReferenceText,
	}

	reqBody, err := json.Marshal(cloneReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/uploads/audio/voice", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var result VoiceCloneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 3. 保存音色信息到数据库
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}

	voice := &domain.Voice{
		ID:          uuid.New(),
		UserID:      userID,
		VoiceName:   req.VoiceName,
		VoiceID:     result.URI,
		Description: "用户自定义音色",
		Status:      "ready",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.voiceRepo.Create(ctx, voice)
	if err != nil {
		return nil, fmt.Errorf("保存音色信息失败: %w", err)
	}

	return &domain.VoiceCloneResponse{
		VoiceID:    result.URI,
		VoiceName:  req.VoiceName,
		CoverImage: req.CoverImage,
		Success:    true,
		Message:    "音色克隆成功",
	}, nil
}
