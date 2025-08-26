package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
)

// InvitationResponseService 邀请响应服务接口
type InvitationResponseService interface {
	// 处理邀请响应
	ProcessInvitationResponse(ctx context.Context, req *domain.InvitationResponseRequest) (*domain.InvitationResponseResult, error)

	// 生成邀请弹窗
	GenerateInvitationPopup(ctx context.Context, suggestion *domain.InvitationSuggestion, groupChatID uuid.UUID, userID uuid.UUID) (*domain.InvitationPopup, error)

	// 处理接受邀请
	HandleAcceptInvitation(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, userID uuid.UUID) (*domain.AcceptInvitationResult, error)

	// 处理拒绝邀请
	HandleRejectInvitation(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, userID uuid.UUID, reason string) (*domain.RejectInvitationResult, error)

	// 生成角色真实反应
	GenerateCharacterReaction(ctx context.Context, characterID uuid.UUID, action string, context map[string]interface{}) (*domain.CharacterReaction, error)

	// 发送被拒绝后的私信
	SendRejectionPrivateMessage(ctx context.Context, characterID uuid.UUID, userID uuid.UUID, rejectionReason string) (*domain.PrivateMessageResult, error)
}

type invitationResponseService struct {
	characterService    CharacterService
	relationshipService RelationshipService
	orchestratorService OrchestratorService
	logger              *logrus.Logger
}

// NewInvitationResponseService 创建邀请响应服务
func NewInvitationResponseService(
	characterService CharacterService,
	relationshipService RelationshipService,
	orchestratorService OrchestratorService,
	logger *logrus.Logger,
) InvitationResponseService {
	return &invitationResponseService{
		characterService:    characterService,
		relationshipService: relationshipService,
		orchestratorService: orchestratorService,
		logger:              logger,
	}
}

// ProcessInvitationResponse 处理邀请响应
func (s *invitationResponseService) ProcessInvitationResponse(ctx context.Context, req *domain.InvitationResponseRequest) (*domain.InvitationResponseResult, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id":  req.CharacterID,
		"group_chat_id": req.GroupChatID,
		"user_id":       req.UserID,
		"action":        req.Action,
	}).Info("Processing invitation response")

	result := &domain.InvitationResponseResult{
		ProcessedAt: time.Now(),
		Success:     false,
		CharacterID: req.CharacterID,
		Action:      req.Action,
	}

	switch req.Action {
	case "accept":
		acceptResult, err := s.HandleAcceptInvitation(ctx, req.CharacterID, req.GroupChatID, req.UserID)
		if err != nil {
			result.Message = fmt.Sprintf("处理接受邀请失败: %v", err)
			return result, err
		}
		result.Success = acceptResult.Success
		result.Message = acceptResult.Message
		result.CharacterReaction = acceptResult.CharacterReaction

	case "reject":
		rejectResult, err := s.HandleRejectInvitation(ctx, req.CharacterID, req.GroupChatID, req.UserID, req.Reason)
		if err != nil {
			result.Message = fmt.Sprintf("处理拒绝邀请失败: %v", err)
			return result, err
		}
		result.Success = rejectResult.Success
		result.Message = rejectResult.Message
		result.CharacterReaction = rejectResult.CharacterReaction
		result.PrivateMessage = rejectResult.PrivateMessage

	default:
		result.Message = "未知的邀请响应动作"
		return result, fmt.Errorf("unknown action: %s", req.Action)
	}

	return result, nil
}

// GenerateInvitationPopup 生成邀请弹窗
func (s *invitationResponseService) GenerateInvitationPopup(ctx context.Context, suggestion *domain.InvitationSuggestion, groupChatID uuid.UUID, userID uuid.UUID) (*domain.InvitationPopup, error) {
	// 获取群聊信息
	groupChat, err := s.characterService.GetGroupChat(ctx, groupChatID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group chat: %w", err)
	}

	// 获取角色与用户的关系
	var relationshipDesc string
	relationships, err := s.relationshipService.ListCharacterRelationships(ctx, suggestion.Character.ID, nil)
	if err == nil && len(relationships) > 0 {
		// 找到与用户相关的关系描述
		relationshipDesc = fmt.Sprintf("你们是%s", relationships[0].CustomTypeName)
	} else {
		relationshipDesc = "你们认识"
	}

	popup := &domain.InvitationPopup{
		CharacterID:      suggestion.Character.ID,
		CharacterName:    suggestion.Character.Name,
		CharacterAvatar:  nil, // suggestion.Character.AvatarURL,
		GroupChatID:      groupChatID,
		GroupChatName:    groupChat.Name,
		InviteReason:     suggestion.InviteReason,
		RelationshipDesc: relationshipDesc,
		ExpectedReaction: suggestion.ExpectedReaction,
		MatchScore:       suggestion.MatchScore,
		Priority:         suggestion.Priority,
		PopupTitle:       fmt.Sprintf("邀请 %s 加入群聊", suggestion.Character.Name),
		PopupMessage:     fmt.Sprintf("%s，%s。\n\n%s", relationshipDesc, suggestion.InviteReason, suggestion.ExpectedReaction),
		AcceptButtonText: "邀请加入",
		RejectButtonText: "暂时不邀请",
		CreatedAt:        time.Now(),
	}

	return popup, nil
}

// HandleAcceptInvitation 处理接受邀请
func (s *invitationResponseService) HandleAcceptInvitation(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, userID uuid.UUID) (*domain.AcceptInvitationResult, error) {
	result := &domain.AcceptInvitationResult{
		Success:     false,
		CharacterID: characterID,
		GroupChatID: groupChatID,
	}

	// 1. 添加角色到群聊
	err := s.characterService.AddGroupChatMember(ctx, userID, groupChatID, characterID, "character")
	if err != nil {
		result.Message = fmt.Sprintf("添加角色到群聊失败: %v", err)
		return result, err
	}

	// 2. 生成角色的接受反应
	reaction, err := s.GenerateCharacterReaction(ctx, characterID, "accept_invitation", map[string]interface{}{
		"group_chat_id": groupChatID,
		"user_id":       userID,
	})
	if err != nil {
		s.logger.WithError(err).Warn("Failed to generate character reaction")
	} else {
		result.CharacterReaction = reaction
	}

	result.Success = true
	result.Message = "角色已成功加入群聊"

	s.logger.WithFields(logrus.Fields{
		"character_id":  characterID,
		"group_chat_id": groupChatID,
		"user_id":       userID,
	}).Info("Character accepted invitation and joined group chat")

	return result, nil
}

// HandleRejectInvitation 处理拒绝邀请
func (s *invitationResponseService) HandleRejectInvitation(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, userID uuid.UUID, reason string) (*domain.RejectInvitationResult, error) {
	result := &domain.RejectInvitationResult{
		Success:     false,
		CharacterID: characterID,
		GroupChatID: groupChatID,
	}

	// 1. 生成角色的拒绝反应
	reaction, err := s.GenerateCharacterReaction(ctx, characterID, "reject_invitation", map[string]interface{}{
		"group_chat_id": groupChatID,
		"user_id":       userID,
		"reason":        reason,
	})
	if err != nil {
		s.logger.WithError(err).Warn("Failed to generate character reaction")
	} else {
		result.CharacterReaction = reaction
	}

	// 2. 发送被拒绝后的私信
	privateMsg, err := s.SendRejectionPrivateMessage(ctx, characterID, userID, reason)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to send rejection private message")
	} else {
		result.PrivateMessage = privateMsg
	}

	result.Success = true
	result.Message = "角色拒绝了邀请"

	s.logger.WithFields(logrus.Fields{
		"character_id":  characterID,
		"group_chat_id": groupChatID,
		"user_id":       userID,
		"reason":        reason,
	}).Info("Character rejected invitation")

	return result, nil
}

// GenerateCharacterReaction 生成角色真实反应
func (s *invitationResponseService) GenerateCharacterReaction(ctx context.Context, characterID uuid.UUID, action string, context map[string]interface{}) (*domain.CharacterReaction, error) {
	// 获取角色信息
	character, err := s.characterService.GetCharacter(ctx, characterID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	reaction := &domain.CharacterReaction{
		CharacterID:   characterID,
		CharacterName: character.Name,
		Action:        action,
		GeneratedAt:   time.Now(),
	}

	// 根据角色性格和动作生成反应
	switch action {
	case "accept_invitation":
		reaction.Content = s.generateAcceptReaction(character)
		reaction.Emotion = "happy"
		reaction.Intensity = 0.8

	case "reject_invitation":
		reaction.Content = s.generateRejectReaction(character, context)
		reaction.Emotion = "apologetic"
		reaction.Intensity = 0.6

	default:
		reaction.Content = "角色做出了反应"
		reaction.Emotion = "neutral"
		reaction.Intensity = 0.5
	}

	return reaction, nil
}

// SendRejectionPrivateMessage 发送被拒绝后的私信
func (s *invitationResponseService) SendRejectionPrivateMessage(ctx context.Context, characterID uuid.UUID, userID uuid.UUID, rejectionReason string) (*domain.PrivateMessageResult, error) {
	// 获取角色信息
	character, err := s.characterService.GetCharacter(ctx, characterID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 生成私信内容
	messageContent := s.generateRejectionPrivateMessage(character, rejectionReason)

	result := &domain.PrivateMessageResult{
		CharacterID:   characterID,
		CharacterName: character.Name,
		UserID:        userID,
		Content:       messageContent,
		MessageType:   "text",
		SentAt:        time.Now(),
		Success:       true,
	}

	s.logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"user_id":      userID,
		"content":      messageContent,
	}).Info("Sent rejection private message")

	return result, nil
}

// 辅助方法

// generateAcceptReaction 生成接受邀请的反应
func (s *invitationResponseService) generateAcceptReaction(character *domain.CharacterResponse) string {
	personality := ""
	if character.Personality != nil {
		personality = strings.ToLower(*character.Personality)
	}

	if strings.Contains(personality, "开朗") || strings.Contains(personality, "活泼") {
		return fmt.Sprintf("(开心地跳起来) 太好了！我很乐意加入！谢谢邀请我~ %s 最喜欢和大家一起聊天了！", character.Name)
	} else if strings.Contains(personality, "温柔") || strings.Contains(personality, "体贴") {
		return fmt.Sprintf("(温柔地笑) 谢谢你的邀请，我很高兴能加入大家。希望能和大家好好相处呢。")
	} else if strings.Contains(personality, "害羞") || strings.Contains(personality, "内向") {
		return fmt.Sprintf("(脸红) 啊...真的可以吗？谢谢你想到我...我会努力不给大家添麻烦的。")
	} else if strings.Contains(personality, "豪爽") || strings.Contains(personality, "义气") {
		return fmt.Sprintf("哈哈！兄弟够意思！我 %s 最喜欢热闹了，这就来！", character.Name)
	} else if strings.Contains(personality, "聪明") || strings.Contains(personality, "文静") {
		return fmt.Sprintf("谢谢邀请。我想这会是一次很有趣的交流，我很期待和大家的对话。")
	}

	return fmt.Sprintf("谢谢邀请！%s 很高兴能加入大家。", character.Name)
}

// generateRejectReaction 生成拒绝邀请的反应
func (s *invitationResponseService) generateRejectReaction(character *domain.CharacterResponse, context map[string]interface{}) string {
	personality := ""
	if character.Personality != nil {
		personality = strings.ToLower(*character.Personality)
	}

	reason, _ := context["reason"].(string)
	_ = reason // 避免未使用变量警告

	if strings.Contains(personality, "害羞") || strings.Contains(personality, "内向") {
		return fmt.Sprintf("(低头) 对不起...我现在还不太适应群聊，可能会给大家添麻烦...下次有机会再说吧。")
	} else if strings.Contains(personality, "温柔") || strings.Contains(personality, "体贴") {
		return fmt.Sprintf("(歉意地) 真的很感谢你的邀请，但是我现在有些事情要处理...希望你能理解。")
	} else if strings.Contains(personality, "高冷") || strings.Contains(personality, "冷淡") {
		return fmt.Sprintf("抱歉，我现在没有兴趣参加群聊。")
	} else if strings.Contains(personality, "直率") || strings.Contains(personality, "豪爽") {
		return fmt.Sprintf("兄弟，不是我不给面子，实在是现在有点忙。改天再说吧！")
	}

	return fmt.Sprintf("谢谢邀请，但是我现在不太方便加入。希望你能理解。")
}

// generateRejectionPrivateMessage 生成拒绝后的私信内容
func (s *invitationResponseService) generateRejectionPrivateMessage(character *domain.CharacterResponse, rejectionReason string) string {
	personality := ""
	if character.Personality != nil {
		personality = strings.ToLower(*character.Personality)
	}

	if strings.Contains(personality, "害羞") || strings.Contains(personality, "内向") {
		return fmt.Sprintf("(私信) 刚才真的很抱歉...其实我很想和大家一起聊天的，只是我在群里会很紧张。如果你不介意的话，我们可以私下聊聊天吗？")
	} else if strings.Contains(personality, "温柔") || strings.Contains(personality, "体贴") {
		return fmt.Sprintf("(私信) 刚才拒绝了你的邀请，真的很不好意思。我担心在群里会打扰到大家，但是我很珍惜我们的友谊。有什么事情可以随时找我哦。")
	} else if strings.Contains(personality, "开朗") || strings.Contains(personality, "活泼") {
		return fmt.Sprintf("(私信) 嘿嘿，刚才是不是有点失望？其实我也想去的，只是现在真的有点忙！等我忙完了一定找你玩，到时候我们好好聊聊！")
	} else if strings.Contains(personality, "豪爽") || strings.Contains(personality, "义气") {
		return fmt.Sprintf("(私信) 兄弟，刚才的事别放在心上。不是我不给你面子，实在是有急事要处理。改天我请你喝酒，咱们好好聊聊！")
	} else if strings.Contains(personality, "聪明") || strings.Contains(personality, "文静") {
		return fmt.Sprintf("(私信) 关于刚才的邀请，我想解释一下。我比较喜欢一对一的深度交流，群聊对我来说有点嘈杂。如果你愿意，我们可以经常私下交流想法。")
	}

	return fmt.Sprintf("(私信) 刚才拒绝了你的邀请，希望你不要介意。我们还是好朋友，有什么事情可以随时找我。")
}
