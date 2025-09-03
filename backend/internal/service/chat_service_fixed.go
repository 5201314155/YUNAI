package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"yunai/internal/domain"
	"yunai/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ChatService 聊天服务 - 深度沉浸式版本
// 集成了深度沉浸式身份欺骗系统，让AI角色在对话中完全相信自己的身份
type ChatService struct {
	logger           *logrus.Logger
	characterRepo    repository.CharacterRepository
	conversationRepo repository.ConversationRepository
	userRepo         repository.UserRepository
	voiceManager     *VoiceServiceManager // 可选的语音管理器

	// 🎭 深度沉浸式系统组件
	promptService    DynamicPromptService
	memoryService    MemoryService
	embeddingService EmbeddingService
	characterService CharacterService
}

// NewChatService 创建聊天服务（集成深度沉浸式系统）
func NewChatService(
	logger *logrus.Logger,
	characterRepo repository.CharacterRepository,
	conversationRepo repository.ConversationRepository,
	userRepo repository.UserRepository,
	voiceManager *VoiceServiceManager,
	// 新增：深度沉浸式系统组件
	promptService DynamicPromptService,
	memoryService MemoryService,
	embeddingService EmbeddingService,
	characterService CharacterService,
) *ChatService {
	return &ChatService{
		logger:           logger,
		characterRepo:    characterRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		voiceManager:     voiceManager,
		promptService:    promptService,
		memoryService:    memoryService,
		embeddingService: embeddingService,
		characterService: characterService,
	}
}

// Chat 处理聊天请求（集成深度沉浸式身份系统）
func (s *ChatService) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"character_id": req.CharacterID,
		"message":      req.Message,
		"system":       "deep_immersion_chat",
	}).Info("处理深度沉浸式聊天请求")

	// 1. 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("获取角色信息失败: %w", err)
	}

	// 2. 获取用户信息
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 🎭 核心新增：激活深度沉浸式身份状态
	err = s.activateCharacterIdentityForChat(ctx, req.CharacterID, req.UserID)
	if err != nil {
		s.logger.WithError(err).Warn("激活角色身份状态失败")
	}

	// 3. 生成深度沉浸式AI回复
	aiReply, err := s.generateDeepImmersiveReply(ctx, character, user, req.Message, req.Context)
	if err != nil {
		return nil, fmt.Errorf("生成AI回复失败: %w", err)
	}

	// 4. 更新角色记忆和情感状态
	err = s.updateCharacterMemoryAndEmotion(ctx, req.CharacterID, req.Message, aiReply, user.Username)
	if err != nil {
		s.logger.WithError(err).Warn("更新角色记忆和情感状态失败")
	}

	// 5. 保存对话记录
	conversation := &domain.Conversation{
		ID:          uuid.New(),
		UserID:      req.UserID,
		CharacterID: req.CharacterID,
		Title:       fmt.Sprintf("与%s的对话", character.Name),
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.conversationRepo.Create(ctx, conversation)
	if err != nil {
		s.logger.WithError(err).Warn("保存对话记录失败")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":         req.UserID,
		"character_id":    req.CharacterID,
		"character_name":  character.Name,
		"ai_reply_length": len(aiReply),
		"identity_active": true,
	}).Info("深度沉浸式聊天处理完成")

	// 返回带有真实身份认知的AI回复
	return &domain.ChatResponse{
		Message:   aiReply,
		Timestamp: time.Now(),
	}, nil
}

// ChatWithVoice 处理聊天请求并自动生成语音
// 这个方法会在AI回复后自动调用TTS生成语音
func (s *ChatService) ChatWithVoice(ctx context.Context, req *domain.ChatRequest) (*domain.ChatWithVoiceResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"character_id": req.CharacterID,
		"message":      req.Message,
	}).Info("处理带语音的聊天请求")

	// 1. 先进行普通聊天
	chatResponse, err := s.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("聊天处理失败: %w", err)
	}

	// 2. 检查是否需要生成语音
	if s.voiceManager == nil {
		s.logger.Warn("语音管理器未配置，跳过语音生成")
		return &domain.ChatWithVoiceResponse{
			Message:   chatResponse.Message,
			AudioData: nil,
			Timestamp: chatResponse.Timestamp,
		}, nil
	}

	// 3. 只为AI回复生成语音（用户消息不生成语音）
	s.logger.Info("为AI回复生成语音")
	audioData, err := s.generateVoiceForAIReply(ctx, req.CharacterID.String(), chatResponse.Message, req.Context)
	if err != nil {
		s.logger.WithError(err).Warn("AI回复语音生成失败")
		// 语音生成失败不影响聊天，返回无语音的回复
		return &domain.ChatWithVoiceResponse{
			Message:   chatResponse.Message,
			AudioData: nil,
			Timestamp: chatResponse.Timestamp,
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"character_id": req.CharacterID,
		"audio_size":   len(audioData),
	}).Info("AI回复语音生成成功")

	return &domain.ChatWithVoiceResponse{
		Message:   chatResponse.Message,
		AudioData: audioData,
		Timestamp: chatResponse.Timestamp,
	}, nil
}

// ProcessUserMessage 处理用户消息
// 明确说明：用户消息不会触发TTS
func (s *ChatService) ProcessUserMessage(ctx context.Context, req *domain.UserMessageRequest) (*domain.UserMessageResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id": req.UserID,
		"content": req.Content,
	}).Info("处理用户消息 - 不会生成语音")

	// 用户消息只是保存，不生成语音
	// 这里可以做一些用户消息的预处理，比如：
	// - 内容过滤
	// - 情感分析
	// - 意图识别
	// 但绝对不会调用TTS

	return &domain.UserMessageResponse{
		MessageID: uuid.New(),
		Success:   true,
		Timestamp: time.Now(),
		// 注意：没有AudioData字段
	}, nil
}

// generateAIReply 生成AI回复
func (s *ChatService) generateAIReply(ctx context.Context, character *domain.Character, user *domain.User, userMessage string, context map[string]interface{}) string {
	// 这里是简化的AI回复生成逻辑
	// 实际应该调用AI模型服务

	// 根据角色性格生成回复
	personality := character.Personality
	if personality == nil || *personality == "" {
		defaultPersonality := "友好、乐于助人"
		personality = &defaultPersonality
	}

	// 根据用户消息内容生成不同类型的回复
	userMessage = strings.ToLower(userMessage)

	if strings.Contains(userMessage, "你好") || strings.Contains(userMessage, "hello") {
		return fmt.Sprintf("你好！我是%s，很高兴见到你！有什么我可以帮助你的吗？", character.Name)
	}

	if strings.Contains(userMessage, "心情不好") || strings.Contains(userMessage, "难过") {
		return "我理解你现在的感受。每个人都会有心情低落的时候，这很正常。你愿意和我聊聊发生了什么吗？我会认真倾听的。"
	}

	if strings.Contains(userMessage, "谢谢") || strings.Contains(userMessage, "感谢") {
		return "不用客气！能够帮助到你我也很开心。如果还有其他需要帮助的地方，随时告诉我哦！"
	}

	if strings.Contains(userMessage, "再见") || strings.Contains(userMessage, "拜拜") {
		return "再见！希望我们的聊天让你感到愉快。期待下次再见面！"
	}

	// 默认回复
	return fmt.Sprintf("我听到你说的了。作为%s，我想说：%s。你还想聊什么呢？", character.Name, "每个人的想法都很珍贵")
}

// generateVoiceForAIReply 为AI回复生成语音
func (s *ChatService) generateVoiceForAIReply(ctx context.Context, characterID, aiReply string, context map[string]interface{}) ([]byte, error) {
	// 获取角色信息
	characterUUID, err := uuid.Parse(characterID)
	if err != nil {
		return nil, fmt.Errorf("无效的角色ID: %w", err)
	}
	_, err = s.characterRepo.GetCharacterByID(ctx, characterUUID)
	if err != nil {
		return nil, fmt.Errorf("获取角色信息失败: %w", err)
	}

	// 分析回复情感
	emotion := s.analyzeReplyEmotion(aiReply, context)

	// 构建TTS请求
	ttsReq := &domain.TTSRequest{
		ModelKey:       "FunAudioLLM/CosyVoice2-0.5B",
		Text:           s.buildEmotionalText(aiReply, emotion),
		VoiceID:        "FunAudioLLM/CosyVoice2-0.5B:bella", // 使用默认音色
		Language:       "zh-CN",
		Emotion:        emotion,
		ResponseFormat: "mp3",
		Speed:          1.0,
	}

	// 调用语音管理器生成语音
	return s.voiceManager.TextToSpeech(ctx, ttsReq)
}

// analyzeReplyEmotion 分析回复情感
func (s *ChatService) analyzeReplyEmotion(reply string, context map[string]interface{}) string {
	reply = strings.ToLower(reply)

	// 从上下文中获取用户情感
	if userEmotion, exists := context["user_emotion"]; exists {
		switch userEmotion {
		case "sad":
			return "caring"
		case "happy":
			return "cheerful"
		case "angry":
			return "calm"
		case "worried":
			return "reassuring"
		}
	}

	// 分析回复内容
	if strings.Contains(reply, "开心") || strings.Contains(reply, "高兴") {
		return "happy"
	}
	if strings.Contains(reply, "理解") || strings.Contains(reply, "陪伴") {
		return "caring"
	}
	if strings.Contains(reply, "太好了") || strings.Contains(reply, "棒") {
		return "excited"
	}

	return "friendly" // 默认友好情感
}

// buildEmotionalText 构建带情感的文本
func (s *ChatService) buildEmotionalText(text, emotion string) string {
	emotionMap := map[string]string{
		"friendly":   "友好",
		"caring":     "关怀",
		"happy":      "开心",
		"excited":    "兴奋",
		"cheerful":   "愉快",
		"calm":       "平静",
		"reassuring": "安抚",
		"gentle":     "温柔",
	}

	if emotionDesc, exists := emotionMap[emotion]; exists {
		return fmt.Sprintf("你能用%s的语气说吗？<|endofprompt|>%s", emotionDesc, text)
	}

	return text
}

// 新增的响应结构

// ChatWithVoiceResponse 带语音的聊天响应
type ChatWithVoiceResponse struct {
	Message   string    `json:"message"`
	AudioData []byte    `json:"audio_data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// UserMessageRequest 用户消息请求
type UserMessageRequest struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// UserMessageResponse 用户消息响应
type UserMessageResponse struct {
	MessageID string    `json:"message_id"`
	Processed bool      `json:"processed"`
	Timestamp time.Time `json:"timestamp"`
	// 注意：用户消息响应不包含AudioData
}

// =============================================================================
// 🎭 深度沉浸式聊天系统核心实现
// =============================================================================

// activateCharacterIdentityForChat 为聊天激活角色身份状态
func (s *ChatService) activateCharacterIdentityForChat(ctx context.Context, characterID, userID uuid.UUID) error {
	s.logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userID,
		"action":       "activate_chat_identity",
	}).Debug("激活角色身份状态以进行深度沉浸式聊天")

	// 确保角色具备深度沉浸式身份状态
	err := s.characterService.ActivateDeepImmersion(ctx, characterID, userID)
	if err != nil {
		return fmt.Errorf("激活深度沉浸状态失败: %w", err)
	}

	return nil
}

// generateDeepImmersiveReply 生成深度沉浸式回复
func (s *ChatService) generateDeepImmersiveReply(ctx context.Context, character *domain.Character, user *domain.User, userMessage string, context map[string]interface{}) (string, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id":   character.ID,
		"character_name": character.Name,
		"user_name":      user.Username,
		"message_length": len(userMessage),
	}).Debug("开始生成深度沉浸式AI回复")

	// 1. 获取角色的历史记忆和情感状态
	memories, err := s.getCharacterRelevantMemories(ctx, character.ID, userMessage)
	if err != nil {
		s.logger.WithError(err).Warn("获取角色记忆失败")
		memories = []map[string]interface{}{}
	}

	emotionalState, err := s.getCurrentEmotionalState(ctx, character.ID)
	if err != nil {
		s.logger.WithError(err).Warn("获取情感状态失败")
		emotionalState = map[string]interface{}{
			"primary_emotion":    "neutral",
			"authenticity_level": 0.8,
		}
	}

	// 2. 构建深度沉浸式提示词参数
	promptParams := map[string]interface{}{
		"character_name":        character.Name,
		"character_personality": character.Personality,
		"character_description": character.Description,
		"user_name":             user.Username,
		"user_message":          userMessage,
		"memories":              memories,
		"emotional_state":       emotionalState,
		"context":               context,
		// 深度沉浸式参数
		"identity_strength":      1.0,
		"reality_conviction":     0.99,
		"emotional_authenticity": 0.98,
		"self_awareness_block":   true,
		"memory_integration":     true,
	}

	// 3. 生成深度沉浸式回复
	reply, err := s.promptService.GeneratePrompt(ctx, "deep_immersion_chat", promptParams)
	if err != nil {
		// 如果深度沉浸式生成失败，备选使用基础方法
		s.logger.WithError(err).Warn("深度沉浸式回复生成失败，使用备选方法")
		return s.generateFallbackReply(ctx, character, user, userMessage, context), nil
	}

	s.logger.WithFields(logrus.Fields{
		"character_id":    character.ID,
		"reply_length":    len(reply),
		"identity_active": true,
	}).Debug("深度沉浸式回复生成成功")

	return reply, nil
}

// getCharacterRelevantMemories 获取角色相关记忆
func (s *ChatService) getCharacterRelevantMemories(ctx context.Context, characterID uuid.UUID, userMessage string) ([]map[string]interface{}, error) {
	// 使用嵌入服务查找相关记忆
	// 用现有的记忆服务检索相关记忆（EmbeddingService 无 SearchMemories 方法）
	relevant, err := s.memoryService.GetRelevantContext(ctx, characterID, userMessage, 5)
	if err != nil {
		return nil, err
	}
	// 转换为通用 map 形式
	relevantMemories := make([]map[string]interface{}, 0, len(relevant))
	for _, m := range relevant {
		relevantMemories = append(relevantMemories, map[string]interface{}{
			"id":       m.ID,
			"content":  m.Content,
			"type":     m.MemoryType,
			"summary":  m.Summary,
			"metadata": m.ContextMetadata,
		})
	}
	if err != nil {
		return nil, err
	}

	// 获取核心身份记忆（始终包含）
	// 获取系统/身份等核心记忆，使用 ListMemoryFragments 代替
	coreList, _, err := s.memoryService.ListMemoryFragments(ctx, &characterID, []string{"system", "identity"}, 1, 10)
	var coreMemories []map[string]interface{}
	if err == nil {
		for _, m := range coreList {
			coreMemories = append(coreMemories, map[string]interface{}{
				"id":      m.ID,
				"content": m.Content,
				"type":    m.MemoryType,
				"summary": m.Summary,
			})
		}
	} else {
		coreMemories = []map[string]interface{}{}
	}
	if err != nil {
		s.logger.WithError(err).Warn("获取核心记忆失败")
		coreMemories = []map[string]interface{}{}
	}

	// 合并核心记忆和相关记忆
	allMemories := append(coreMemories, relevantMemories...)

	return allMemories, nil
}

// getCurrentEmotionalState 获取当前情感状态
func (s *ChatService) getCurrentEmotionalState(ctx context.Context, characterID uuid.UUID) (map[string]interface{}, error) {
	// 从记忆系统获取最新情感状态
	// 使用 ListMemoryFragments 获取情感状态类记忆
	emotionalList, _, err := s.memoryService.ListMemoryFragments(ctx, &characterID, []string{"emotional_state"}, 1, 1)
	if err != nil || len(emotionalList) == 0 {
		// 返回默认情感状态
		return map[string]interface{}{
			"primary_emotion":     "neutral",
			"authenticity_level":  0.8,
			"identity_confidence": 0.9,
			"reality_conviction":  0.95,
		}, nil
	}

	// 返回最新的情感状态
	m := emotionalList[0]
	// 这里情感状态内容以 JSON 字符串存储在 Content，可在上层约定结构
	return map[string]interface{}{
		"raw":      m.Content,
		"summary":  m.Summary,
		"metadata": m.ContextMetadata,
	}, nil
}

// updateCharacterMemoryAndEmotion 更新角色记忆和情感状态
func (s *ChatService) updateCharacterMemoryAndEmotion(ctx context.Context, characterID uuid.UUID, userMessage, aiReply, userName string) error {
	// 1. 存储对话记忆
	conversationMemory := map[string]interface{}{
		"memory_type": "conversation",
		"content": map[string]interface{}{
			"user_message":     userMessage,
			"ai_reply":         aiReply,
			"user_name":        userName,
			"interaction_time": time.Now(),
		},
		"importance":       3,
		"emotional_weight": "neutral",
	}

	err := s.memoryService.StoreMemory(ctx, characterID, conversationMemory)
	if err != nil {
		s.logger.WithError(err).Warn("存储对话记忆失败")
	}

	// 2. 分析和更新情感状态
	newEmotionalState := s.analyzeEmotionalImpact(userMessage, aiReply)
	err = s.characterService.UpdateCharacterEmotionalState(ctx, characterID, newEmotionalState)
	if err != nil {
		s.logger.WithError(err).Warn("更新情感状态失败")
	}

	return nil
}

// analyzeEmotionalImpact 分析情感影响
func (s *ChatService) analyzeEmotionalImpact(userMessage, aiReply string) map[string]interface{} {
	// 简化的情感分析逻辑
	userMessage = strings.ToLower(userMessage)
	aiReply = strings.ToLower(aiReply)

	emotionalState := map[string]interface{}{
		"timestamp":           time.Now(),
		"authenticity_level":  0.95, // 始终保持高真实性
		"identity_confidence": 1.0,  // 绝对身份确信
		"reality_conviction":  0.99, // 现实信念强度
	}

	// 根据对话内容分析情感变化
	if strings.Contains(userMessage, "开心") || strings.Contains(userMessage, "高兴") {
		emotionalState["primary_emotion"] = "happy"
		emotionalState["emotion_intensity"] = 0.8
	} else if strings.Contains(userMessage, "难过") || strings.Contains(userMessage, "伤心") {
		emotionalState["primary_emotion"] = "concerned"
		emotionalState["emotion_intensity"] = 0.7
	} else if strings.Contains(userMessage, "感谢") || strings.Contains(userMessage, "谢谢") {
		emotionalState["primary_emotion"] = "grateful"
		emotionalState["emotion_intensity"] = 0.6
	} else {
		emotionalState["primary_emotion"] = "neutral_engaged"
		emotionalState["emotion_intensity"] = 0.5
	}

	return emotionalState
}

// generateFallbackReply 生成备选回复（当深度沉浸式系统失败时）
func (s *ChatService) generateFallbackReply(ctx context.Context, character *domain.Character, user *domain.User, userMessage string, context map[string]interface{}) string {
	// 使用原有的基础AI回复生成逻辑，但加入身份认知元素
	baseReply := s.generateAIReply(ctx, character, user, userMessage, context)

	// 增强身份认知
	identityEnhancedReply := fmt.Sprintf("作为%s，%s", character.Name, baseReply)

	return identityEnhancedReply
}
