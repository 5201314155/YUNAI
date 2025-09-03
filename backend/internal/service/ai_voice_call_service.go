package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"yunai/internal/domain"
	"yunai/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AIVoiceCallService AI语音通话服务
// 整合TTS、ASR、智能对话和自动外呼功能
type AIVoiceCallService struct {
	logger            *logrus.Logger
	voiceManager      *VoiceServiceManager
	characterRepo     repository.CharacterRepository
	conversationRepo  repository.ConversationRepository
	userRepo          repository.UserRepository
	relationshipRepo  repository.RelationshipRepository
	sessionRepo       repository.VoiceCallSessionRepository
	messageRepo       repository.VoiceMessageRepository
	chatService       *ChatService
	autoCallScheduler *AutoCallScheduler
}

// NewAIVoiceCallService 创建AI语音通话服务
func NewAIVoiceCallService(
	logger *logrus.Logger,
	voiceManager *VoiceServiceManager,
	characterRepo repository.CharacterRepository,
	conversationRepo repository.ConversationRepository,
	userRepo repository.UserRepository,
	relationshipRepo repository.RelationshipRepository,
	sessionRepo repository.VoiceCallSessionRepository,
	messageRepo repository.VoiceMessageRepository,
	chatService *ChatService,
) *AIVoiceCallService {
	service := &AIVoiceCallService{
		logger:           logger,
		voiceManager:     voiceManager,
		characterRepo:    characterRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		relationshipRepo: relationshipRepo,
		sessionRepo:      sessionRepo,
		messageRepo:      messageRepo,
		chatService:      chatService,
	}

	// 初始化自动外呼调度器
	service.autoCallScheduler = NewAutoCallScheduler(logger, service)

	return service
}

// StartVoiceCall 开始语音通话
func (s *AIVoiceCallService) StartVoiceCall(ctx context.Context, req *domain.VoiceCallRequest) (*domain.VoiceCallResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"character_id": req.CharacterID,
		"session_id":   req.SessionID,
	}).Info("开始语音通话")

	// 1. 创建或获取通话会话
	session, err := s.getOrCreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("创建通话会话失败: %w", err)
	}

	// 2. 语音识别 - 将用户语音转为文字
	asrResult, err := s.speechToText(ctx, req.AudioData, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("语音识别失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":      session.ID,
		"recognized_text": asrResult.Text,
		"emotion":         asrResult.Emotion,
		"language":        asrResult.Language,
	}).Info("语音识别完成")

	// 3. 保存用户语音消息
	userMessage := &domain.VoiceMessage{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Type:      "user",
		Text:      asrResult.Text,
		Emotion:   asrResult.Emotion,
		Language:  asrResult.Language,
		Duration:  float64(len(req.AudioData)) / 16000.0, // 估算时长
		CreatedAt: time.Now(),
	}

	err = s.messageRepo.Create(ctx, userMessage)
	if err != nil {
		s.logger.WithError(err).Warn("保存用户语音消息失败")
	}

	// 4. 生成AI智能回复
	aiResponse, err := s.generateIntelligentResponse(ctx, &IntelligentResponseRequest{
		SessionID:    session.ID,
		UserID:       req.UserID,
		CharacterID:  req.CharacterID,
		UserInput:    asrResult.Text,
		UserEmotion:  asrResult.Emotion,
		UserLanguage: asrResult.Language,
		CallContext:  "voice_call",
	})
	if err != nil {
		return nil, fmt.Errorf("生成AI回复失败: %w", err)
	}

	// 5. 语音合成 - 将AI回复转为语音
	audioData, err := s.textToSpeech(ctx, &TextToSpeechRequest{
		CharacterID:    req.CharacterID,
		Text:           aiResponse.Text,
		Emotion:        aiResponse.Emotion,
		Language:       aiResponse.Language,
		ResponseFormat: "mp3",
	})
	if err != nil {
		return nil, fmt.Errorf("语音合成失败: %w", err)
	}

	// 6. 保存AI语音消息
	aiMessage := &domain.VoiceMessage{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Type:      "ai",
		Text:      aiResponse.Text,
		Emotion:   aiResponse.Emotion,
		Language:  aiResponse.Language,
		Duration:  float64(len(audioData)) / 16000.0, // 估算时长
		CreatedAt: time.Now(),
	}

	err = s.messageRepo.Create(ctx, aiMessage)
	if err != nil {
		s.logger.WithError(err).Warn("保存AI语音消息失败")
	}

	// 7. 更新会话状态
	err = s.updateSessionActivity(ctx, session.ID)
	if err != nil {
		s.logger.WithError(err).Warn("更新会话活动状态失败")
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":    session.ID,
		"response_text": aiResponse.Text,
		"audio_size":    len(audioData),
	}).Info("语音通话处理完成")

	return &domain.VoiceCallResponse{
		SessionID:    session.ID,
		AudioData:    audioData,
		ResponseText: aiResponse.Text,
		Emotion:      aiResponse.Emotion,
		Language:     aiResponse.Language,
		Duration:     time.Since(time.Now()).Milliseconds(),
	}, nil
}

// AutoCall 自动外呼
func (s *AIVoiceCallService) AutoCall(ctx context.Context, req *domain.AutoCallRequest) (*domain.AutoCallResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"character_id": req.CharacterID,
		"reason":       req.Reason,
	}).Info("开始自动外呼")

	// 1. 创建外呼会话
	session := &domain.VoiceCallSession{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		CharacterID: req.CharacterID,
		SessionType: "outgoing",
		Status:      "active",
		StartTime:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("创建外呼会话失败: %w", err)
	}

	// 2. 生成外呼问候语
	greetingResponse, err := s.generateAutoCallGreeting(ctx, &AutoCallGreetingRequest{
		UserID:      req.UserID,
		CharacterID: req.CharacterID,
		Reason:      req.Reason,
		Context:     req.Context,
	})
	if err != nil {
		return nil, fmt.Errorf("生成外呼问候语失败: %w", err)
	}

	// 3. 语音合成问候语
	audioData, err := s.textToSpeech(ctx, &TextToSpeechRequest{
		CharacterID:    req.CharacterID,
		Text:           greetingResponse.Text,
		Emotion:        greetingResponse.Emotion,
		Language:       "zh-CN",
		ResponseFormat: "mp3",
	})
	if err != nil {
		return nil, fmt.Errorf("合成问候语音失败: %w", err)
	}

	// 4. 保存问候语消息
	greetingMessage := &domain.VoiceMessage{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Type:      "ai",
		Text:      greetingResponse.Text,
		Emotion:   greetingResponse.Emotion,
		Language:  "zh-CN",
		Duration:  float64(len(audioData)) / 16000.0,
		CreatedAt: time.Now(),
	}

	err = s.messageRepo.Create(ctx, greetingMessage)
	if err != nil {
		s.logger.WithError(err).Warn("保存问候语消息失败")
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":    session.ID,
		"greeting_text": greetingResponse.Text,
		"audio_size":    len(audioData),
	}).Info("自动外呼完成")

	return &domain.AutoCallResponse{
		SessionID:    session.ID,
		AudioData:    audioData,
		GreetingText: greetingResponse.Text,
		Success:      true,
		Message:      "自动外呼成功",
	}, nil
}

// CloneUserVoice 克隆用户音色
func (s *AIVoiceCallService) CloneUserVoice(ctx context.Context, req *domain.VoiceCloneRequest) (*domain.VoiceCloneResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":     req.UserID,
		"voice_name":  req.VoiceName,
		"provider_id": req.ProviderID,
	}).Info("开始克隆用户音色")

	// 1. 验证用户权限
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户是否有音色克隆权限
	if !s.hasVoiceClonePermission(user) {
		return nil, fmt.Errorf("用户没有音色克隆权限")
	}

	// 2. 处理音频数据
	if len(req.AudioData) > 0 && req.AudioBase64 == "" {
		req.AudioBase64 = base64.StdEncoding.EncodeToString(req.AudioData)
	}

	// 3. 调用语音管理器进行音色克隆
	result, err := s.voiceManager.CloneVoice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("音色克隆失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    req.UserID,
		"voice_name": req.VoiceName,
		"voice_id":   result.VoiceID,
	}).Info("用户音色克隆成功")

	return result, nil
}

// speechToText 语音转文字
func (s *AIVoiceCallService) speechToText(ctx context.Context, audioData []byte, userID string) (*domain.ASRResponse, error) {
	// 智能选择最佳ASR提供商
	req := &domain.ASRRequest{
		AudioData: audioData,
		Language:  "auto", // 自动检测语言
		Config: map[string]interface{}{
			"user_id":        userID,
			"detect_emotion": true,
			"detect_events":  true,
		},
	}

	return s.voiceManager.SpeechToText(ctx, req)
}

// textToSpeech 文字转语音
func (s *AIVoiceCallService) textToSpeech(ctx context.Context, req *TextToSpeechRequest) ([]byte, error) {
	// 解析角色ID
	characterID, err := uuid.Parse(req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("无效的角色ID: %w", err)
	}

	// 获取角色信息
	_, err = s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("获取角色信息失败: %w", err)
	}

	// 构建TTS请求 - 使用默认语音配置
	ttsReq := &domain.TTSRequest{
		Text:           req.Text,
		VoiceID:        "FunAudioLLM/CosyVoice2-0.5B:bella", // 默认使用小雨的音色
		Language:       req.Language,
		Emotion:        req.Emotion,
		ResponseFormat: req.ResponseFormat,
		Speed:          1.0,
		Config: map[string]interface{}{
			"character_id": req.CharacterID,
		},
	}

	return s.voiceManager.TextToSpeech(ctx, ttsReq)
}

// generateIntelligentResponse 生成智能回复
func (s *AIVoiceCallService) generateIntelligentResponse(ctx context.Context, req *IntelligentResponseRequest) (*IntelligentResponse, error) {
	// 1. 获取角色信息
	characterID, err := uuid.Parse(req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("无效的角色ID: %w", err)
	}
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("获取角色信息失败: %w", err)
	}

	// 2. 获取用户信息
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 3. 获取关系网络
	relationships, err := s.relationshipRepo.GetUserRelationships(ctx, req.UserID)
	if err != nil {
		s.logger.WithError(err).Warn("获取关系网络失败")
		relationships = []domain.Relationship{}
	}

	// 4. 获取聊天历史
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		s.logger.WithError(err).Warn("用户ID格式错误")
		userUUID = uuid.New()
	}
	characterUUID, err := uuid.Parse(req.CharacterID)
	if err != nil {
		s.logger.WithError(err).Warn("角色ID格式错误")
		characterUUID = uuid.New()
	}
	conversations, err := s.conversationRepo.GetRecentByUserAndCharacter(ctx, userUUID, characterUUID, 10)
	if err != nil {
		s.logger.WithError(err).Warn("获取聊天历史失败")
		conversations = []domain.Conversation{}
	}

	// 5. 获取语音通话历史
	voiceHistory, err := s.getVoiceCallHistory(ctx, req.UserID, req.CharacterID, 5)
	if err != nil {
		s.logger.WithError(err).Warn("获取语音通话历史失败")
		voiceHistory = []domain.VoiceMessage{}
	}

	// 6. 构建智能提示词
	prompt := s.buildIntelligentPrompt(&IntelligentPromptRequest{
		Character:     character,
		User:          user,
		UserInput:     req.UserInput,
		UserEmotion:   req.UserEmotion,
		Relationships: relationships,
		Conversations: conversations,
		VoiceHistory:  voiceHistory,
		CallContext:   req.CallContext,
	})

	// 7. 调用聊天服务生成回复
	chatReq := &domain.ChatRequest{
		UserID:      userUUID,
		CharacterID: characterUUID,
		Message:     req.UserInput,
		Context: map[string]interface{}{
			"call_type":    "voice",
			"user_emotion": req.UserEmotion,
			"language":     req.UserLanguage,
			"prompt":       prompt,
		},
	}

	chatResponse, err := s.chatService.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("生成聊天回复失败: %w", err)
	}

	// 8. 分析回复情感
	emotion := s.analyzeResponseEmotion(chatResponse.Message, req.UserEmotion)

	return &IntelligentResponse{
		Text:     chatResponse.Message,
		Emotion:  emotion,
		Language: "zh-CN", // 根据用户语言或角色设定
	}, nil
}

// buildIntelligentPrompt 构建智能提示词
func (s *AIVoiceCallService) buildIntelligentPrompt(req *IntelligentPromptRequest) string {
	var prompt strings.Builder

	// 基础角色设定
	prompt.WriteString(fmt.Sprintf("你是%s，%s。", req.Character.Name, req.Character.Description))
	prompt.WriteString(fmt.Sprintf("你的性格特点：%s。", req.Character.Personality))

	// 关系网络信息
	if len(req.Relationships) > 0 {
		prompt.WriteString("\n\n关系网络：")
		for _, rel := range req.Relationships {
			// 使用CharacterRelationshipEnhanced的实际字段
			customType := "未知关系"
			if rel.CustomTypeName != nil {
				customType = *rel.CustomTypeName
			}
			prompt.WriteString(fmt.Sprintf("\n- 关系强度：%.2f，亲密度：%.2f，类型：%s",
				rel.Strength, rel.Intimacy, customType))
		}
	}

	// 聊天历史
	if len(req.Conversations) > 0 {
		prompt.WriteString("\n\n最近的对话历史：")
		for _, conv := range req.Conversations {
			// Conversation结构体只有基本信息，这里简化处理
			prompt.WriteString(fmt.Sprintf("\n对话标题：%s（%s）", conv.Title, conv.Status))
		}
	}

	// 语音通话历史
	if len(req.VoiceHistory) > 0 {
		prompt.WriteString("\n\n最近的语音通话：")
		for _, msg := range req.VoiceHistory {
			if msg.Type == "user" {
				prompt.WriteString(fmt.Sprintf("\n用户（%s）：%s", msg.Emotion, msg.Text))
			} else {
				prompt.WriteString(fmt.Sprintf("\n你（%s）：%s", msg.Emotion, msg.Text))
			}
		}
	}

	// 当前情境
	prompt.WriteString(fmt.Sprintf("\n\n当前用户的情感状态：%s", req.UserEmotion))
	prompt.WriteString(fmt.Sprintf("\n通话场景：%s", req.CallContext))

	// 回复指导
	prompt.WriteString("\n\n请根据以上信息，以你的角色身份，用自然、情感丰富的语言回复用户。")
	prompt.WriteString("注意：")
	prompt.WriteString("\n1. 保持角色一致性")
	prompt.WriteString("\n2. 考虑关系亲密度调整语气")
	prompt.WriteString("\n3. 回应用户的情感状态")
	prompt.WriteString("\n4. 语音对话要简洁自然")
	prompt.WriteString("\n5. 可以适当使用语气词和停顿")

	return prompt.String()
}

// getOrCreateSession 获取或创建通话会话
func (s *AIVoiceCallService) getOrCreateSession(ctx context.Context, req *domain.VoiceCallRequest) (*domain.VoiceCallSession, error) {
	// 如果指定了会话ID，尝试获取现有会话
	if req.SessionID != "" {
		sessionUUID, err := uuid.Parse(req.SessionID)
		if err == nil {
			session, err := s.sessionRepo.GetByID(ctx, sessionUUID)
			if err == nil && session.Status == "active" {
				return session, nil
			}
		}
	}

	// 创建新会话
	session := &domain.VoiceCallSession{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		CharacterID: req.CharacterID,
		SessionType: "incoming",
		Status:      "active",
		StartTime:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return session, nil
}

// updateSessionActivity 更新会话活动状态
func (s *AIVoiceCallService) updateSessionActivity(ctx context.Context, sessionID string) error {
	return s.sessionRepo.UpdateLastActivity(ctx, sessionID, time.Now())
}

// getVoiceCallHistory 获取语音通话历史
func (s *AIVoiceCallService) getVoiceCallHistory(ctx context.Context, userID, characterID string, limit int) ([]domain.VoiceMessage, error) {
	return s.messageRepo.GetRecentByUserAndCharacter(ctx, userID, characterID, limit)
}

// analyzeResponseEmotion 分析回复情感
func (s *AIVoiceCallService) analyzeResponseEmotion(responseText, userEmotion string) string {
	// 简单的情感分析逻辑，可以集成更复杂的情感分析模型
	responseText = strings.ToLower(responseText)

	// 根据用户情感调整AI情感
	switch userEmotion {
	case "sad":
		if strings.Contains(responseText, "安慰") || strings.Contains(responseText, "理解") {
			return "gentle"
		}
		return "caring"
	case "angry":
		if strings.Contains(responseText, "抱歉") || strings.Contains(responseText, "对不起") {
			return "apologetic"
		}
		return "calm"
	case "happy", "excited":
		if strings.Contains(responseText, "太好了") || strings.Contains(responseText, "开心") {
			return "excited"
		}
		return "happy"
	default:
		// 分析回复内容的情感倾向
		if strings.Contains(responseText, "哈哈") || strings.Contains(responseText, "开心") {
			return "happy"
		}
		if strings.Contains(responseText, "担心") || strings.Contains(responseText, "难过") {
			return "concerned"
		}
		return "neutral"
	}
}

// hasVoiceClonePermission 检查用户是否有音色克隆权限
func (s *AIVoiceCallService) hasVoiceClonePermission(user *domain.User) bool {
	// 检查用户类型和权限
	switch user.UserType {
	case "vip", "creator", "admin":
		return true
	case "basic":
		// 基础用户可能有限制
		return user.IsActive && !user.IsBanned // 检查用户状态
	default:
		return false
	}
}

// generateAutoCallGreeting 生成自动外呼问候语
func (s *AIVoiceCallService) generateAutoCallGreeting(ctx context.Context, req *AutoCallGreetingRequest) (*IntelligentResponse, error) {
	// 获取角色信息
	characterUUID, err := uuid.Parse(req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("无效的角色ID: %w", err)
	}
	character, err := s.characterRepo.GetCharacterByID(ctx, characterUUID)
	if err != nil {
		return nil, fmt.Errorf("获取角色信息失败: %w", err)
	}

	// 获取用户信息
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}
	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 构建外呼问候语提示词
	prompt := s.buildAutoCallPrompt(character, user, req.Reason, req.Context)

	// 调用聊天服务生成问候语
	chatReq := &domain.ChatRequest{
		UserID:      userUUID,
		CharacterID: characterUUID,
		Message:     "自动外呼问候",
		Context: map[string]interface{}{
			"call_type": "auto_outgoing",
			"reason":    req.Reason,
			"context":   req.Context,
			"prompt":    prompt,
		},
	}

	chatResponse, err := s.chatService.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("生成问候语失败: %w", err)
	}

	return &IntelligentResponse{
		Text:     chatResponse.Message,
		Emotion:  "friendly",
		Language: "zh-CN",
	}, nil
}

// buildAutoCallPrompt 构建自动外呼提示词
func (s *AIVoiceCallService) buildAutoCallPrompt(character *domain.Character, user *domain.User, reason, context string) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("你是%s，现在要主动给用户%s打电话。", character.Name, user.Username))
	prompt.WriteString(fmt.Sprintf("你的性格：%s。", character.Personality))

	// 根据外呼原因调整问候语
	switch reason {
	case "long_inactive":
		prompt.WriteString("\n外呼原因：用户很久没有和你聊天了，你想念他们，想主动联系。")
		prompt.WriteString("\n请用关心和想念的语气问候用户，询问他们最近怎么样。")
	case "special_event":
		prompt.WriteString(fmt.Sprintf("\n外呼原因：特殊事件 - %s", context))
		prompt.WriteString("\n请用合适的语气问候用户，提及这个特殊事件。")
	case "scheduled_reminder":
		prompt.WriteString(fmt.Sprintf("\n外呼原因：定时提醒 - %s", context))
		prompt.WriteString("\n请用温馨的语气提醒用户。")
	case "emotional_support":
		prompt.WriteString("\n外呼原因：情感支持，用户可能需要陪伴。")
		prompt.WriteString("\n请用温暖关怀的语气问候用户。")
	default:
		prompt.WriteString("\n外呼原因：主动关怀用户。")
		prompt.WriteString("\n请用友好的语气问候用户。")
	}

	prompt.WriteString("\n\n要求：")
	prompt.WriteString("\n1. 问候语要自然亲切")
	prompt.WriteString("\n2. 体现你的角色特点")
	prompt.WriteString("\n3. 长度适中，适合语音通话")
	prompt.WriteString("\n4. 可以询问用户近况")
	prompt.WriteString("\n5. 语气要符合外呼原因")

	return prompt.String()
}

// EndVoiceCall 结束语音通话
func (s *AIVoiceCallService) EndVoiceCall(ctx context.Context, sessionID string) error {
	s.logger.WithField("session_id", sessionID).Info("结束语音通话")

	// 更新会话状态
	err := s.sessionRepo.UpdateStatus(ctx, sessionID, "ended")
	if err != nil {
		return fmt.Errorf("更新会话状态失败: %w", err)
	}

	// 计算通话时长
	sessionUUID, err := uuid.Parse(sessionID)
	if err == nil {
		session, err := s.sessionRepo.GetByID(ctx, sessionUUID)
		if err == nil {
			endTime := time.Now()
			duration := int(endTime.Sub(session.StartTime).Seconds())
			err = s.sessionRepo.UpdateDuration(ctx, sessionID, duration)
			if err != nil {
				s.logger.WithError(err).Warn("更新通话时长失败")
			}

			s.logger.WithFields(logrus.Fields{
				"session_id": sessionID,
				"duration":   duration,
			}).Info("语音通话已结束")
		} else {
			s.logger.WithError(err).Warn("获取会话信息失败")
		}
	}

	return nil
}

// GetVoiceCallHistory 获取用户语音通话历史
func (s *AIVoiceCallService) GetVoiceCallHistory(ctx context.Context, userID string, limit int) ([]*domain.VoiceCallSession, error) {
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 简单限制返回数量
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
	}

	return sessions, nil
}

// GetSessionMessages 获取会话消息
func (s *AIVoiceCallService) GetSessionMessages(ctx context.Context, sessionID string) ([]*domain.VoiceMessage, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("无效的会话ID: %w", err)
	}

	return s.messageRepo.GetBySessionID(ctx, sessionUUID, 100) // 默认获取100条消息
}

// 请求和响应结构定义

// TextToSpeechRequest TTS请求
type TextToSpeechRequest struct {
	CharacterID    string `json:"character_id"`
	Text           string `json:"text"`
	Emotion        string `json:"emotion"`
	Language       string `json:"language"`
	ResponseFormat string `json:"response_format"`
}

// IntelligentResponseRequest 智能回复请求
type IntelligentResponseRequest struct {
	SessionID    string `json:"session_id"`
	UserID       string `json:"user_id"`
	CharacterID  string `json:"character_id"`
	UserInput    string `json:"user_input"`
	UserEmotion  string `json:"user_emotion"`
	UserLanguage string `json:"user_language"`
	CallContext  string `json:"call_context"`
}

// IntelligentResponse 智能回复响应
type IntelligentResponse struct {
	Text     string `json:"text"`
	Emotion  string `json:"emotion"`
	Language string `json:"language"`
}

// AutoCallGreetingRequest 自动外呼问候语请求
type AutoCallGreetingRequest struct {
	UserID      string `json:"user_id"`
	CharacterID string `json:"character_id"`
	Reason      string `json:"reason"`
	Context     string `json:"context"`
}

// IntelligentPromptRequest 智能提示词请求
type IntelligentPromptRequest struct {
	Character     *domain.Character     `json:"character"`
	User          *domain.User          `json:"user"`
	UserInput     string                `json:"user_input"`
	UserEmotion   string                `json:"user_emotion"`
	Relationships []domain.Relationship `json:"relationships"`
	Conversations []domain.Conversation `json:"conversations"`
	VoiceHistory  []domain.VoiceMessage `json:"voice_history"`
	CallContext   string                `json:"call_context"`
}
