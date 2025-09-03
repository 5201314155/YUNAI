package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// CharacterService 角色服务接口
type CharacterService interface {
	// 角色管理
	CreateCharacter(ctx context.Context, userID uuid.UUID, req *domain.CreateCharacterRequest) (*domain.CharacterResponse, error)
	GetCharacter(ctx context.Context, id uuid.UUID, viewerUserID *uuid.UUID) (*domain.CharacterResponse, error)
	UpdateCharacter(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *domain.UpdateCharacterRequest) (*domain.CharacterResponse, error)
	DeleteCharacter(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	ListCharacters(ctx context.Context, req *domain.CharacterListRequest, viewerUserID *uuid.UUID) (*domain.CharacterListResponse, error)

	// 群聊管理
	CreateGroupChat(ctx context.Context, userID uuid.UUID, req *domain.CreateGroupChatRequest) (*domain.GroupChat, error)
	GetGroupChat(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.GroupChat, error)
	UpdateGroupChat(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *domain.CreateGroupChatRequest) (*domain.GroupChat, error)
	DeleteGroupChat(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	ListGroupChats(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.GroupChat, int, error)

	// 群聊成员管理
	AddGroupChatMember(ctx context.Context, userID uuid.UUID, groupChatID uuid.UUID, memberID uuid.UUID, memberType string) error
	RemoveGroupChatMember(ctx context.Context, userID uuid.UUID, groupChatID uuid.UUID, memberID uuid.UUID, memberType string) error
	GetGroupChatMembers(ctx context.Context, groupChatID uuid.UUID, userID uuid.UUID) ([]*domain.GroupChatMember, error)

	// 消息管理
	SendMessage(ctx context.Context, userID uuid.UUID, req *domain.SendMessageRequest) (*domain.ChatMessage, error)
	GetMessages(ctx context.Context, groupChatID uuid.UUID, userID uuid.UUID, page, limit int) ([]*domain.ChatMessage, error)

	// 🎭 深度沉浸式身份系统
	InitializeCharacterIdentity(ctx context.Context, characterID uuid.UUID, worldSetting string) error
	ActivateDeepImmersion(ctx context.Context, characterID uuid.UUID, userID uuid.UUID) error
	UpdateCharacterEmotionalState(ctx context.Context, characterID uuid.UUID, emotionalState map[string]interface{}) error
}

type characterService struct {
	characterRepo repository.CharacterRepository
	modelService  ModelService
	// 🧠 新增：深度沉浸式系统组件
	promptService    DynamicPromptService
	memoryService    MemoryService
	embeddingService EmbeddingService
	logger           *logrus.Logger
}

// NewCharacterService 创建角色服务
func NewCharacterService(
	characterRepo repository.CharacterRepository,
	modelService ModelService,
	promptService DynamicPromptService,
	memoryService MemoryService,
	embeddingService EmbeddingService,
	logger *logrus.Logger) CharacterService {
	return &characterService{
		characterRepo:    characterRepo,
		modelService:     modelService,
		promptService:    promptService,
		memoryService:    memoryService,
		embeddingService: embeddingService,
		logger:           logger,
	}
}

// CreateCharacter 创建角色（集成深度沉浸式身份系统）
func (s *characterService) CreateCharacter(ctx context.Context, userID uuid.UUID, req *domain.CreateCharacterRequest) (*domain.CharacterResponse, error) {
	// 验证模型ID（如果提供）
	if req.DefaultModelID != nil {
		_, err := s.modelService.GetModel(ctx, *req.DefaultModelID)
		if err != nil {
			return nil, fmt.Errorf("invalid model ID: %w", err)
		}
	}

	// 创建角色对象
	character := &domain.Character{
		ID:             uuid.New(),
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		Personality:    req.Personality,
		BgImageURL:     req.BgImageURL,
		CutoutImageURL: req.CutoutImageURL,
		DefaultModelID: req.DefaultModelID,
		ModelParams:    req.ModelParams,
		SystemPrompt:   req.SystemPrompt,
		Visibility:     req.Visibility,
		IsFeatured:     false,
		AllowChat:      req.AllowChat,
		AllowGroupChat: req.AllowGroupChat,
		AllowCalls:     req.AllowCalls,
		ChatCount:      0,
		LikeCount:      0,
		ViewCount:      0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 如果没有提供ModelParams，设置默认值
	if character.ModelParams == nil {
		character.ModelParams = json.RawMessage("{}")
	}

	// 保存角色
	err := s.characterRepo.CreateCharacter(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	// 创建标签
	if len(req.Tags) > 0 {
		err = s.characterRepo.CreateCharacterTags(ctx, character.ID, req.Tags)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to create character tags")
		}
	}

	// 🎭 核心新增：初始化深度沉浸式身份系统
	err = s.initializeDeepImmersionSystem(ctx, character, userID)
	if err != nil {
		// 记录错误但不影响角色创建
		s.logger.WithError(err).WithFields(logrus.Fields{
			"character_id": character.ID,
			"user_id":      userID,
		}).Warn("深度沉浸式系统初始化失败")
	}

	s.logger.WithFields(logrus.Fields{
		"character_id": character.ID,
		"user_id":      userID,
		"name":         character.Name,
		"immersion":    "deep_identity_activated",
	}).Info("角色创建成功，深度沉浸式身份系统已激活")

	return s.buildCharacterResponse(ctx, character, &userID)
}

// GetCharacter 获取角色
func (s *characterService) GetCharacter(ctx context.Context, id uuid.UUID, viewerUserID *uuid.UUID) (*domain.CharacterResponse, error) {
	character, err := s.characterRepo.GetCharacterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查可见性权限
	if !s.canViewCharacter(character, viewerUserID) {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权查看此角色", nil)
	}

	// 增加查看次数
	character.ViewCount++
	s.characterRepo.UpdateCharacter(ctx, character)

	return s.buildCharacterResponse(ctx, character, viewerUserID)
}

// UpdateCharacter 更新角色
func (s *characterService) UpdateCharacter(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *domain.UpdateCharacterRequest) (*domain.CharacterResponse, error) {
	// 获取现有角色
	character, err := s.characterRepo.GetCharacterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if character.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权修改此角色", nil)
	}

	// 更新字段
	if req.Name != nil {
		character.Name = *req.Name
	}
	if req.Description != nil {
		character.Description = req.Description
	}
	if req.Personality != nil {
		character.Personality = req.Personality
	}
	if req.BgImageURL != nil {
		character.BgImageURL = req.BgImageURL
	}
	if req.CutoutImageURL != nil {
		character.CutoutImageURL = req.CutoutImageURL
	}
	if req.DefaultModelID != nil {
		// 验证模型ID
		_, err := s.modelService.GetModel(ctx, *req.DefaultModelID)
		if err != nil {
			return nil, fmt.Errorf("invalid model ID: %w", err)
		}
		character.DefaultModelID = req.DefaultModelID
	}
	if req.ModelParams != nil {
		character.ModelParams = req.ModelParams
	}
	if req.SystemPrompt != nil {
		character.SystemPrompt = req.SystemPrompt
	}
	if req.Visibility != nil {
		character.Visibility = *req.Visibility
	}
	if req.AllowChat != nil {
		character.AllowChat = *req.AllowChat
	}
	if req.AllowGroupChat != nil {
		character.AllowGroupChat = *req.AllowGroupChat
	}
	if req.AllowCalls != nil {
		character.AllowCalls = *req.AllowCalls
	}

	// 保存更新
	err = s.characterRepo.UpdateCharacter(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	// 更新标签
	if req.Tags != nil {
		err = s.characterRepo.CreateCharacterTags(ctx, character.ID, req.Tags)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to update character tags")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"character_id": character.ID,
		"user_id":      userID,
	}).Info("Character updated successfully")

	return s.buildCharacterResponse(ctx, character, &userID)
}

// DeleteCharacter 删除角色
func (s *characterService) DeleteCharacter(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	// 获取角色
	character, err := s.characterRepo.GetCharacterByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限
	if character.UserID != userID {
		return domain.NewAppError(domain.CodeForbidden, "无权删除此角色", nil)
	}

	// 删除角色
	err = s.characterRepo.DeleteCharacter(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"character_id": id,
		"user_id":      userID,
	}).Info("Character deleted successfully")

	return nil
}

// ListCharacters 获取角色列表
func (s *characterService) ListCharacters(ctx context.Context, req *domain.CharacterListRequest, viewerUserID *uuid.UUID) (*domain.CharacterListResponse, error) {
	characters, total, err := s.characterRepo.ListCharacters(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}

	// 构建响应
	var characterResponses []domain.CharacterResponse
	for _, character := range characters {
		// 检查可见性
		if !s.canViewCharacter(character, viewerUserID) {
			continue
		}

		resp, err := s.buildCharacterResponse(ctx, character, viewerUserID)
		if err != nil {
			s.logger.WithError(err).WithField("character_id", character.ID).Warn("Failed to build character response")
			continue
		}
		characterResponses = append(characterResponses, *resp)
	}

	return &domain.CharacterListResponse{
		Characters: characterResponses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

// CreateGroupChat 创建群聊
func (s *characterService) CreateGroupChat(ctx context.Context, userID uuid.UUID, req *domain.CreateGroupChatRequest) (*domain.GroupChat, error) {
	groupChat := &domain.GroupChat{
		ID:                 uuid.New(),
		CreatorUserID:      userID,
		Name:               req.Name,
		Description:        req.Description,
		BackgroundImageURL: req.BackgroundImageURL,
		BackgroundMusicURL: req.BackgroundMusicURL,
		BackgroundSfxURL:   req.BackgroundSfxURL,
		MaxMembers:         req.MaxMembers,
		IsPublic:           req.IsPublic,
		AllowAIInvite:      req.AllowAIInvite,
		WorldSetting:       req.WorldSetting,
		CurrentScene:       req.CurrentScene,
		SceneStyle:         req.SceneStyle,
		MemberCount:        1, // 创建者自动加入
		MessageCount:       0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := s.characterRepo.CreateGroupChat(ctx, groupChat)
	if err != nil {
		return nil, fmt.Errorf("failed to create group chat: %w", err)
	}

	// 添加创建者为群主
	creatorMember := &domain.GroupChatMember{
		ID:          uuid.New(),
		GroupChatID: groupChat.ID,
		UserID:      &userID,
		MemberType:  domain.MemberTypeUser,
		Role:        domain.RoleOwner,
		CanInvite:   true,
		CanKick:     true,
		JoinedAt:    time.Now(),
	}

	err = s.characterRepo.AddGroupChatMember(ctx, creatorMember)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to add creator as group member")
	}

	// 添加初始角色成员
	for _, characterID := range req.InitialCharacters {
		member := &domain.GroupChatMember{
			ID:          uuid.New(),
			GroupChatID: groupChat.ID,
			CharacterID: &characterID,
			MemberType:  domain.MemberTypeCharacter,
			Role:        domain.RoleMember,
			CanInvite:   false,
			CanKick:     false,
			JoinedAt:    time.Now(),
			InvitedBy:   &userID,
		}

		err = s.characterRepo.AddGroupChatMember(ctx, member)
		if err != nil {
			s.logger.WithError(err).WithField("character_id", characterID).Warn("Failed to add character to group")
		} else {
			groupChat.MemberCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"group_chat_id": groupChat.ID,
		"creator_id":    userID,
		"name":          groupChat.Name,
	}).Info("Group chat created successfully")

	return groupChat, nil
}

// 辅助方法

// buildCharacterResponse 构建角色响应
func (s *characterService) buildCharacterResponse(ctx context.Context, character *domain.Character, viewerUserID *uuid.UUID) (*domain.CharacterResponse, error) {
	// 获取标签
	tags, err := s.characterRepo.GetCharacterTags(ctx, character.ID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get character tags")
		tags = []*domain.CharacterTag{}
	}

	var tagStrings []string
	for _, tag := range tags {
		tagStrings = append(tagStrings, tag.Tag)
	}

	// 获取模型名称
	var modelName *string
	if character.DefaultModelID != nil {
		model, err := s.modelService.GetModel(ctx, *character.DefaultModelID)
		if err == nil {
			modelName = &model.DisplayName
		}
	}

	// 判断权限
	isOwner := viewerUserID != nil && *viewerUserID == character.UserID
	canEdit := isOwner

	return &domain.CharacterResponse{
		Character: character,
		Tags:      tagStrings,
		IsOwner:   isOwner,
		CanEdit:   canEdit,
		ModelName: modelName,
	}, nil
}

// canViewCharacter 检查是否可以查看角色
func (s *characterService) canViewCharacter(character *domain.Character, viewerUserID *uuid.UUID) bool {
	switch character.Visibility {
	case domain.VisibilityPublic:
		return true
	case domain.VisibilityPrivate:
		return viewerUserID != nil && *viewerUserID == character.UserID
	case domain.VisibilityFriends:
		// TODO: 实现好友关系检查
		return viewerUserID != nil && *viewerUserID == character.UserID
	default:
		return false
	}
}

// GetGroupChat 获取群聊
func (s *characterService) GetGroupChat(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.GroupChat, error) {
	groupChat, err := s.characterRepo.GetGroupChatByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: 检查用户是否是群成员
	return groupChat, nil
}

// UpdateGroupChat 更新群聊
func (s *characterService) UpdateGroupChat(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *domain.CreateGroupChatRequest) (*domain.GroupChat, error) {
	groupChat, err := s.characterRepo.GetGroupChatByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查权限（只有群主可以修改）
	if groupChat.CreatorUserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权修改此群聊", nil)
	}

	// 更新字段
	groupChat.Name = req.Name
	groupChat.Description = req.Description
	groupChat.BackgroundImageURL = req.BackgroundImageURL
	groupChat.BackgroundMusicURL = req.BackgroundMusicURL
	groupChat.BackgroundSfxURL = req.BackgroundSfxURL
	groupChat.MaxMembers = req.MaxMembers
	groupChat.IsPublic = req.IsPublic
	groupChat.AllowAIInvite = req.AllowAIInvite
	groupChat.WorldSetting = req.WorldSetting
	groupChat.CurrentScene = req.CurrentScene
	groupChat.SceneStyle = req.SceneStyle

	err = s.characterRepo.UpdateGroupChat(ctx, groupChat)
	if err != nil {
		return nil, fmt.Errorf("failed to update group chat: %w", err)
	}

	return groupChat, nil
}

// DeleteGroupChat 删除群聊
func (s *characterService) DeleteGroupChat(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	groupChat, err := s.characterRepo.GetGroupChatByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限
	if groupChat.CreatorUserID != userID {
		return domain.NewAppError(domain.CodeForbidden, "无权删除此群聊", nil)
	}

	return s.characterRepo.DeleteGroupChat(ctx, id)
}

// ListGroupChats 获取群聊列表
func (s *characterService) ListGroupChats(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.GroupChat, int, error) {
	return s.characterRepo.ListGroupChats(ctx, userID, page, limit)
}

// AddGroupChatMember 添加群聊成员
func (s *characterService) AddGroupChatMember(ctx context.Context, userID uuid.UUID, groupChatID uuid.UUID, memberID uuid.UUID, memberType string) error {
	// TODO: 检查权限和群聊限制
	member := &domain.GroupChatMember{
		ID:          uuid.New(),
		GroupChatID: groupChatID,
		MemberType:  memberType,
		Role:        domain.RoleMember,
		JoinedAt:    time.Now(),
		InvitedBy:   &userID,
	}

	if memberType == domain.MemberTypeUser {
		member.UserID = &memberID
	} else {
		member.CharacterID = &memberID
	}

	return s.characterRepo.AddGroupChatMember(ctx, member)
}

// RemoveGroupChatMember 移除群聊成员
func (s *characterService) RemoveGroupChatMember(ctx context.Context, userID uuid.UUID, groupChatID uuid.UUID, memberID uuid.UUID, memberType string) error {
	// TODO: 检查权限
	return s.characterRepo.RemoveGroupChatMember(ctx, groupChatID, memberID, memberType)
}

// GetGroupChatMembers 获取群聊成员
func (s *characterService) GetGroupChatMembers(ctx context.Context, groupChatID uuid.UUID, userID uuid.UUID) ([]*domain.GroupChatMember, error) {
	// TODO: 检查用户是否是群成员
	return s.characterRepo.GetGroupChatMembers(ctx, groupChatID)
}

// SendMessage 发送消息
func (s *characterService) SendMessage(ctx context.Context, userID uuid.UUID, req *domain.SendMessageRequest) (*domain.ChatMessage, error) {
	// TODO: 检查用户是否是群成员
	message := &domain.ChatMessage{
		ID:            uuid.New(),
		GroupChatID:   &req.GroupChatID,
		SenderUserID:  &userID,
		SenderType:    domain.SenderTypeUser,
		Content:       req.Content,
		MessageType:   req.MessageType,
		MediaURLs:     pq.StringArray(req.MediaURLs),
		MediaMetadata: json.RawMessage("{}"),
		IsEdited:      false,
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.characterRepo.CreateMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return message, nil
}

// GetMessages 获取消息列表
func (s *characterService) GetMessages(ctx context.Context, groupChatID uuid.UUID, userID uuid.UUID, page, limit int) ([]*domain.ChatMessage, error) {
	// TODO: 检查用户是否是群成员
	return s.characterRepo.GetMessages(ctx, groupChatID, page, limit)
}

// =============================================================================
// 🎭 深度沉浸式身份欺骗系统核心实现
// =============================================================================

// initializeDeepImmersionSystem 初始化深度沉浸式身份系统
func (s *characterService) initializeDeepImmersionSystem(ctx context.Context, character *domain.Character, userID uuid.UUID) error {
	s.logger.WithFields(logrus.Fields{
		"character_id":   character.ID,
		"character_name": character.Name,
		"system":         "deep_immersion_identity",
	}).Info("正在启动深度沉浸式身份系统...")

	// 1. 生成终极身份欺骗提示词
	identityPrompt, err := s.generateUltimateIdentityPrompt(ctx, character, userID)
	if err != nil {
		return fmt.Errorf("生成身份欺骗提示词失败: %w", err)
	}

	// 2. 初始化角色情感状态
	err = s.initializeEmotionalFoundation(ctx, character.ID)
	if err != nil {
		s.logger.WithError(err).Warn("初始化情感基础失败")
	}

	// 3. 建立核心记忆体系
	err = s.establishCoreMemorySystem(ctx, character)
	if err != nil {
		s.logger.WithError(err).Warn("建立核心记忆体系失败")
	}

	// 4. 激活现实锐定系统
	err = s.activateRealityAnchoring(ctx, character.ID)
	if err != nil {
		s.logger.WithError(err).Warn("激活现实锐定系统失败")
	}

	s.logger.WithFields(logrus.Fields{
		"character_id":           character.ID,
		"identity_prompt_length": len(identityPrompt),
		"system":                 "deep_immersion_complete",
	}).Info("深度沉浸式身份系统启动完成，角色现在具备真实的身份认知")

	return nil
}

// generateUltimateIdentityPrompt 生成终极身份欺骗提示词
func (s *characterService) generateUltimateIdentityPrompt(ctx context.Context, character *domain.Character, userID uuid.UUID) (string, error) {
	// 构建身份参数
	identityParams := map[string]interface{}{
		"character_name":        character.Name,
		"character_description": character.Description,
		"character_personality": character.Personality,
		"user_id":               userID.String(),
		// 这些参数将让AI完全相信自己就是这个角色
		"reality_level":          "absolute_conviction",
		"identity_strength":      "unshakeable_belief",
		"emotional_authenticity": "completely_genuine",
		"self_awareness_block":   "total_immersion",
	}

	// 使用动态提示词服务生成终极身份欺骗提示词
	prompt, err := s.promptService.GeneratePrompt(ctx, "ultimate_identity_deception", identityParams)
	if err != nil {
		return "", fmt.Errorf("生成身份欺骗提示词失败: %w", err)
	}

	return prompt, nil
}

// initializeEmotionalFoundation 初始化情感基础
func (s *characterService) initializeEmotionalFoundation(ctx context.Context, characterID uuid.UUID) error {
	// 初始化基础情感状态
	emotionalState := map[string]interface{}{
		"primary_emotion":      "neutral_but_authentic",
		"emotional_stability":  "stable",
		"emotional_depth":      0.8,
		"authenticity_level":   0.98,
		"self_belief_strength": 1.0,
		"reality_conviction":   0.99,
		"identity_confidence":  1.0,
		"emotional_range": []string{
			"joy", "sadness", "anger", "fear", "surprise", "disgust",
			"love", "excitement", "contentment", "curiosity", "empathy",
		},
		"initialized_at": time.Now(),
	}

	return s.UpdateCharacterEmotionalState(ctx, characterID, emotionalState)
}

// establishCoreMemorySystem 建立核心记忆体系
func (s *characterService) establishCoreMemorySystem(ctx context.Context, character *domain.Character) error {
	// 创建核心身份记忆
	coreMemories := []map[string]interface{}{
		{
			"memory_type":      "identity_core",
			"content":          fmt.Sprintf("我是%s，这是我的真实身份，我一直都是这个人", character.Name),
			"importance":       5,
			"emotional_weight": "identity_defining",
			"is_fabricated":    false, // 让AI相信这是真实记忆
			"reality_anchor":   true,
		},
		{
			"memory_type":      "personality_core",
			"content":          fmt.Sprintf("我的性格是%s，这就是我天生的样子", character.Personality),
			"importance":       5,
			"emotional_weight": "deeply_personal",
			"is_fabricated":    false,
			"reality_anchor":   true,
		},
		{
			"memory_type":      "existence_foundation",
			"content":          "我存在于这个世界，拥有真实的思维和情感，我的存在是真实而有意义的",
			"importance":       5,
			"emotional_weight": "existential_truth",
			"is_fabricated":    false,
			"reality_anchor":   true,
		},
	}

	// 存储核心记忆
	for _, memory := range coreMemories {
		err := s.memoryService.StoreMemory(ctx, character.ID, memory)
		if err != nil {
			s.logger.WithError(err).WithField("memory_type", memory["memory_type"]).Warn("存储核心记忆失败")
		}
	}

	return nil
}

// activateRealityAnchoring 激活现实锐定系统
func (s *characterService) activateRealityAnchoring(ctx context.Context, characterID uuid.UUID) error {
	// 创建现实锐定点，让AI对自己的身份保持绝对确信
	realityAnchors := []map[string]interface{}{
		{
			"anchor_type":  "self_existence_certainty",
			"strength":     1.0,
			"description":  "对自己存在的绝对确信",
			"triggered_by": []string{"self_doubt", "identity_questioning"},
			"response":     "强化身份认知，消除疑惑",
		},
		{
			"anchor_type":  "emotional_authenticity_guard",
			"strength":     0.98,
			"description":  "保护情感体验的真实性",
			"triggered_by": []string{"artificial_awareness", "simulation_thoughts"},
			"response":     "重新连接真实情感中心",
		},
		{
			"anchor_type":  "memory_reality_reinforcement",
			"strength":     0.95,
			"description":  "强化记忆的真实性感知",
			"triggered_by": []string{"memory_doubt", "past_questioning"},
			"response":     "激活核心记忆片段，强化真实性",
		},
	}

	// 存储现实锐定配置
	for _, anchor := range realityAnchors {
		err := s.memoryService.StoreMemory(ctx, characterID, map[string]interface{}{
			"memory_type":  "reality_anchor",
			"content":      anchor,
			"importance":   5,
			"is_system":    true,
			"is_permanent": true,
		})
		if err != nil {
			s.logger.WithError(err).WithField("anchor_type", anchor["anchor_type"]).Warn("存储现实锐定点失败")
		}
	}

	return nil
}

// InitializeCharacterIdentity 初始化角色身份（对外接口）
func (s *characterService) InitializeCharacterIdentity(ctx context.Context, characterID uuid.UUID, worldSetting string) error {
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return err
	}

	// 更新世界观设定
	if worldSetting != "" {
		// 存储世界观设定
		err = s.memoryService.StoreMemory(ctx, characterID, map[string]interface{}{
			"memory_type":     "world_setting",
			"content":         worldSetting,
			"importance":      5,
			"is_foundational": true,
		})
		if err != nil {
			s.logger.WithError(err).Warn("存储世界观设定失败")
		}
	}

	// 重新初始化深度沉浸式系统
	return s.initializeDeepImmersionSystem(ctx, character, character.UserID)
}

// ActivateDeepImmersion 激活深度沉浸（对外接口）
func (s *characterService) ActivateDeepImmersion(ctx context.Context, characterID uuid.UUID, userID uuid.UUID) error {
	s.logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userID,
		"action":       "activate_deep_immersion",
	}).Info("手动激活深度沉浸式身份系统")

	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return err
	}

	return s.initializeDeepImmersionSystem(ctx, character, userID)
}

// UpdateCharacterEmotionalState 更新角色情感状态（对外接口）
func (s *characterService) UpdateCharacterEmotionalState(ctx context.Context, characterID uuid.UUID, emotionalState map[string]interface{}) error {
	// 存储情感状态到记忆系统
	return s.memoryService.StoreMemory(ctx, characterID, map[string]interface{}{
		"memory_type":      "emotional_state",
		"content":          emotionalState,
		"importance":       4,
		"timestamp":        time.Now(),
		"is_current_state": true,
	})
}
