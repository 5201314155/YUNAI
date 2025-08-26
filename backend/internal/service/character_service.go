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
}

type characterService struct {
	characterRepo repository.CharacterRepository
	modelService  ModelService
	logger        *logrus.Logger
}

// NewCharacterService 创建角色服务
func NewCharacterService(characterRepo repository.CharacterRepository, modelService ModelService, logger *logrus.Logger) CharacterService {
	return &characterService{
		characterRepo: characterRepo,
		modelService:  modelService,
		logger:        logger,
	}
}

// CreateCharacter 创建角色
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

	s.logger.WithFields(logrus.Fields{
		"character_id": character.ID,
		"user_id":      userID,
		"name":         character.Name,
	}).Info("Character created successfully")

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
