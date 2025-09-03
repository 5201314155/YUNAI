package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// UserIdentityService 用户身份识别服务
type UserIdentityService interface {
	// 从设定中提取用户身份
	ExtractUserIdentityFromSetting(ctx context.Context, setting string) (*domain.UserIdentity, error)

	// 获取用户在特定上下文中的身份
	GetUserIdentityInContext(ctx context.Context, userID uuid.UUID, contextType string, contextID uuid.UUID) (*domain.UserIdentity, error)

	// 解析@提及中的用户身份
	ParseMentionIdentity(ctx context.Context, mention string, userID uuid.UUID) (*domain.UserIdentity, error)

	// 替换内容中的用户身份引用
	ReplaceUserReferences(ctx context.Context, content string, userID uuid.UUID, targetIdentity *domain.UserIdentity) (string, error)

	// 获取用户真实昵称
	GetUserRealNickname(ctx context.Context, userID uuid.UUID) (string, error)

	// 构建用户身份上下文
	BuildUserIdentityContext(ctx context.Context, userID uuid.UUID, characterID *uuid.UUID, groupChatID *uuid.UUID) (*domain.UserIdentityContext, error)
}

type userIdentityService struct {
	userRepo      repository.UserRepository
	characterRepo repository.CharacterRepository
	logger        *logrus.Logger
}

// NewUserIdentityService 创建用户身份识别服务
func NewUserIdentityService(
	userRepo repository.UserRepository,
	characterRepo repository.CharacterRepository,
	logger *logrus.Logger,
) UserIdentityService {
	return &userIdentityService{
		userRepo:      userRepo,
		characterRepo: characterRepo,
		logger:        logger,
	}
}

// ExtractUserIdentityFromSetting 从设定中提取用户身份
func (s *userIdentityService) ExtractUserIdentityFromSetting(ctx context.Context, setting string) (*domain.UserIdentity, error) {
	s.logger.WithField("setting", setting).Debug("Extracting user identity from setting")

	// 常见的用户身份提取模式
	patterns := []struct {
		regex   *regexp.Regexp
		extract func([]string) *domain.UserIdentity
	}{
		{
			// 匹配 "用户（李明）" 或 "用户(李明)"
			regex: regexp.MustCompile(`用户[（(]([^）)]+)[）)]`),
			extract: func(matches []string) *domain.UserIdentity {
				return &domain.UserIdentity{
					Type:        domain.IdentityTypeSpecified,
					DisplayName: matches[1],
					Source:      "setting_parentheses",
				}
			},
		},
		{
			// 匹配 "你是李明" 或 "你叫李明"
			regex: regexp.MustCompile(`你[是叫]([^\s，。！？,.\!?]+)`),
			extract: func(matches []string) *domain.UserIdentity {
				return &domain.UserIdentity{
					Type:        domain.IdentityTypeSpecified,
					DisplayName: matches[1],
					Source:      "setting_direct",
				}
			},
		},
		{
			// 匹配 "李明说" 或 "李明问"
			regex: regexp.MustCompile(`([^\s，。！？,.\!?]+)[说问道]`),
			extract: func(matches []string) *domain.UserIdentity {
				return &domain.UserIdentity{
					Type:        domain.IdentityTypeSpecified,
					DisplayName: matches[1],
					Source:      "setting_speech",
				}
			},
		},
	}

	// 尝试匹配各种模式
	for _, pattern := range patterns {
		if matches := pattern.regex.FindStringSubmatch(setting); len(matches) > 1 {
			identity := pattern.extract(matches)
			s.logger.WithFields(logrus.Fields{
				"identity_type": identity.Type,
				"display_name":  identity.DisplayName,
				"source":        identity.Source,
			}).Info("Extracted user identity from setting")
			return identity, nil
		}
	}

	// 如果没有找到指定身份，返回需要使用真实昵称的标识
	return &domain.UserIdentity{
		Type:   domain.IdentityTypeReal,
		Source: "setting_fallback",
	}, nil
}

// GetUserIdentityInContext 获取用户在特定上下文中的身份
func (s *userIdentityService) GetUserIdentityInContext(ctx context.Context, userID uuid.UUID, contextType string, contextID uuid.UUID) (*domain.UserIdentity, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"context_type": contextType,
		"context_id":   contextID,
	}).Debug("Getting user identity in context")

	switch contextType {
	case "character":
		return s.getUserIdentityForCharacter(ctx, userID, contextID)
	case "group_chat":
		return s.getUserIdentityForGroupChat(ctx, userID, contextID)
	case "moment":
		return s.getUserIdentityForMoment(ctx, userID, contextID)
	default:
		// 默认使用真实昵称
		return s.getRealUserIdentity(ctx, userID)
	}
}

// getUserIdentityForCharacter 获取用户在角色上下文中的身份
func (s *userIdentityService) getUserIdentityForCharacter(ctx context.Context, userID uuid.UUID, characterID uuid.UUID) (*domain.UserIdentity, error) {
	// 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return s.getRealUserIdentity(ctx, userID)
	}

	// 从角色设定中提取用户身份
	if character.SystemPrompt != nil {
		identity, err := s.ExtractUserIdentityFromSetting(ctx, *character.SystemPrompt)
		if err == nil && identity.Type == domain.IdentityTypeSpecified {
			return identity, nil
		}
	}

	// 从角色描述中提取用户身份
	if character.Description != nil {
		identity, err := s.ExtractUserIdentityFromSetting(ctx, *character.Description)
		if err == nil && identity.Type == domain.IdentityTypeSpecified {
			return identity, nil
		}
	}

	// 如果没有找到指定身份，使用真实昵称
	return s.getRealUserIdentity(ctx, userID)
}

// getUserIdentityForGroupChat 获取用户在群聊上下文中的身份
func (s *userIdentityService) getUserIdentityForGroupChat(ctx context.Context, userID uuid.UUID, groupChatID uuid.UUID) (*domain.UserIdentity, error) {
	// 获取群聊信息
	groupChat, err := s.characterRepo.GetGroupChatByID(ctx, groupChatID)
	if err != nil {
		return s.getRealUserIdentity(ctx, userID)
	}

	// 从群聊设定中提取用户身份
	if groupChat.WorldSetting != nil {
		identity, err := s.ExtractUserIdentityFromSetting(ctx, *groupChat.WorldSetting)
		if err == nil && identity.Type == domain.IdentityTypeSpecified {
			return identity, nil
		}
	}

	// 如果没有找到指定身份，使用真实昵称
	return s.getRealUserIdentity(ctx, userID)
}

// getUserIdentityForMoment 获取用户在朋友圈上下文中的身份
func (s *userIdentityService) getUserIdentityForMoment(ctx context.Context, userID uuid.UUID, momentID uuid.UUID) (*domain.UserIdentity, error) {
	// 朋友圈上下文中默认使用真实昵称，除非有特殊设定
	return s.getRealUserIdentity(ctx, userID)
}

// getRealUserIdentity 获取用户真实身份
func (s *userIdentityService) getRealUserIdentity(ctx context.Context, userID uuid.UUID) (*domain.UserIdentity, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	displayName := user.Username
	if user.Nickname != nil && *user.Nickname != "" {
		displayName = *user.Nickname
	}

	return &domain.UserIdentity{
		Type:        domain.IdentityTypeReal,
		DisplayName: displayName,
		Source:      "user_profile",
		UserID:      &userID,
	}, nil
}

// ParseMentionIdentity 解析@提及中的用户身份
func (s *userIdentityService) ParseMentionIdentity(ctx context.Context, mention string, userID uuid.UUID) (*domain.UserIdentity, error) {
	// 移除@符号
	cleanMention := strings.TrimPrefix(mention, "@")

	// 如果@的是"用户"，则使用真实昵称
	if cleanMention == "用户" {
		return s.getRealUserIdentity(ctx, userID)
	}

	// 否则使用指定的身份名称
	return &domain.UserIdentity{
		Type:        domain.IdentityTypeSpecified,
		DisplayName: cleanMention,
		Source:      "mention",
	}, nil
}

// ReplaceUserReferences 替换内容中的用户身份引用
func (s *userIdentityService) ReplaceUserReferences(ctx context.Context, content string, userID uuid.UUID, targetIdentity *domain.UserIdentity) (string, error) {
	if targetIdentity == nil {
		var err error
		targetIdentity, err = s.getRealUserIdentity(ctx, userID)
		if err != nil {
			return content, err
		}
	}

	// 替换常见的用户引用
	replacements := map[string]string{
		"用户":  targetIdentity.DisplayName,
		"@用户": "@" + targetIdentity.DisplayName,
		"你":   targetIdentity.DisplayName,
		"您":   targetIdentity.DisplayName,
	}

	result := content
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result, nil
}

// GetUserRealNickname 获取用户真实昵称
func (s *userIdentityService) GetUserRealNickname(ctx context.Context, userID uuid.UUID) (string, error) {
	identity, err := s.getRealUserIdentity(ctx, userID)
	if err != nil {
		return "", err
	}
	return identity.DisplayName, nil
}

// BuildUserIdentityContext 构建用户身份上下文
func (s *userIdentityService) BuildUserIdentityContext(ctx context.Context, userID uuid.UUID, characterID *uuid.UUID, groupChatID *uuid.UUID) (*domain.UserIdentityContext, error) {
	context := &domain.UserIdentityContext{
		UserID: userID,
	}

	// 获取真实身份
	realIdentity, err := s.getRealUserIdentity(ctx, userID)
	if err != nil {
		return nil, err
	}
	context.RealIdentity = realIdentity

	// 获取角色上下文身份
	if characterID != nil {
		charIdentity, err := s.getUserIdentityForCharacter(ctx, userID, *characterID)
		if err == nil {
			context.CharacterIdentity = charIdentity
		}
	}

	// 获取群聊上下文身份
	if groupChatID != nil {
		groupIdentity, err := s.getUserIdentityForGroupChat(ctx, userID, *groupChatID)
		if err == nil {
			context.GroupChatIdentity = groupIdentity
		}
	}

	// 确定当前使用的身份
	if context.GroupChatIdentity != nil && context.GroupChatIdentity.Type == domain.IdentityTypeSpecified {
		context.CurrentIdentity = context.GroupChatIdentity
	} else if context.CharacterIdentity != nil && context.CharacterIdentity.Type == domain.IdentityTypeSpecified {
		context.CurrentIdentity = context.CharacterIdentity
	} else {
		context.CurrentIdentity = context.RealIdentity
	}

	return context, nil
}
