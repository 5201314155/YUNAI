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

// RelationshipDrivenInvitationService 基于关系网的智能邀请服务
type RelationshipDrivenInvitationService interface {
	// 分析聊天内容并基于关系网推荐邀请
	AnalyzeChatAndSuggestInvitation(ctx context.Context, req *domain.ChatAnalysisRequest) (*domain.RelationshipBasedInvitationResult, error)

	// 获取用户的关系网络图谱
	GetUserRelationshipMap(ctx context.Context, userID uuid.UUID) (*domain.UserRelationshipMap, error)

	// 基于关系描述匹配合适的角色
	MatchCharactersByRelationship(ctx context.Context, userID uuid.UUID, relationshipHint string, chatContext string) ([]*domain.RelationshipMatchResult, error)

	// 生成基于关系的邀请建议
	GenerateRelationshipBasedSuggestion(ctx context.Context, match *domain.RelationshipMatchResult, chatContext string) (*domain.RelationshipInvitationSuggestion, error)
}

type relationshipDrivenInvitationService struct {
	relationshipService RelationshipService
	characterService    CharacterService
	memoryService       MemoryService
	logger              *logrus.Logger
}

// NewRelationshipDrivenInvitationService 创建基于关系网的邀请服务
func NewRelationshipDrivenInvitationService(
	relationshipService RelationshipService,
	characterService CharacterService,
	memoryService MemoryService,
	logger *logrus.Logger,
) RelationshipDrivenInvitationService {
	return &relationshipDrivenInvitationService{
		relationshipService: relationshipService,
		characterService:    characterService,
		memoryService:       memoryService,
		logger:              logger,
	}
}

// AnalyzeChatAndSuggestInvitation 分析聊天内容并基于关系网推荐邀请
func (s *relationshipDrivenInvitationService) AnalyzeChatAndSuggestInvitation(ctx context.Context, req *domain.ChatAnalysisRequest) (*domain.RelationshipBasedInvitationResult, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":       req.UserID,
		"group_chat_id": req.GroupChatID,
		"message":       req.ChatMessage,
	}).Info("Analyzing chat for relationship-based invitation")

	result := &domain.RelationshipBasedInvitationResult{
		AnalyzedAt:  time.Now(),
		ChatMessage: req.ChatMessage,
		UserID:      req.UserID,
		GroupChatID: req.GroupChatID,
		Success:     false,
		Suggestions: []*domain.RelationshipInvitationSuggestion{},
	}

	// 1. 获取用户的关系网络图谱
	relationshipMap, err := s.GetUserRelationshipMap(ctx, req.UserID)
	if err != nil {
		result.Message = fmt.Sprintf("获取关系网络失败: %v", err)
		return result, err
	}

	if len(relationshipMap.RelationshipGroups) == 0 {
		result.Message = "用户还没有建立任何关系网络"
		return result, nil
	}

	// 2. 分析聊天内容，提取关系提示
	relationshipHints := s.extractRelationshipHints(req.ChatMessage)
	if len(relationshipHints) == 0 {
		result.Message = "聊天内容中没有检测到关系相关的提示"
		return result, nil
	}

	s.logger.WithField("relationship_hints", relationshipHints).Info("Extracted relationship hints from chat")

	// 3. 基于关系提示匹配角色
	allMatches := []*domain.RelationshipMatchResult{}
	for _, hint := range relationshipHints {
		matches, err := s.MatchCharactersByRelationship(ctx, req.UserID, hint, req.ChatMessage)
		if err != nil {
			s.logger.WithError(err).WithField("hint", hint).Warn("Failed to match characters by relationship")
			continue
		}
		allMatches = append(allMatches, matches...)
	}

	if len(allMatches) == 0 {
		result.Message = "在关系网络中没有找到匹配的角色"
		return result, nil
	}

	// 4. 过滤已在群聊中的角色
	filteredMatches, err := s.filterExistingGroupMembers(ctx, allMatches, req.GroupChatID, req.UserID)
	if err != nil {
		result.Message = fmt.Sprintf("过滤群聊成员失败: %v", err)
		return result, err
	}

	if len(filteredMatches) == 0 {
		result.Message = "所有匹配的角色都已经在群聊中了"
		return result, nil
	}

	// 5. 生成邀请建议
	for _, match := range filteredMatches {
		suggestion, err := s.GenerateRelationshipBasedSuggestion(ctx, match, req.ChatMessage)
		if err != nil {
			s.logger.WithError(err).WithField("character_id", match.Character.ID).Warn("Failed to generate suggestion")
			continue
		}
		result.Suggestions = append(result.Suggestions, suggestion)
	}

	// 6. 按匹配度排序并限制数量
	if len(result.Suggestions) > 3 {
		result.Suggestions = result.Suggestions[:3]
	}

	result.Success = len(result.Suggestions) > 0
	if result.Success {
		result.Message = fmt.Sprintf("基于关系网络找到了 %d 个邀请建议", len(result.Suggestions))
	} else {
		result.Message = "没有生成有效的邀请建议"
	}

	s.logger.WithFields(logrus.Fields{
		"suggestions_count": len(result.Suggestions),
		"success":           result.Success,
	}).Info("Relationship-based invitation analysis completed")

	return result, nil
}

// GetUserRelationshipMap 获取用户的关系网络图谱
func (s *relationshipDrivenInvitationService) GetUserRelationshipMap(ctx context.Context, userID uuid.UUID) (*domain.UserRelationshipMap, error) {
	// 获取用户的所有关系类型
	relationshipTypes, _, err := s.relationshipService.ListRelationshipTypes(ctx, &userID, nil, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship types: %w", err)
	}

	relationshipMap := &domain.UserRelationshipMap{
		UserID:             userID,
		RelationshipGroups: make(map[string]*domain.RelationshipGroup),
		TotalCharacters:    0,
	}

	// 获取用户的所有角色
	characterList, err := s.characterService.ListCharacters(ctx, &domain.CharacterListRequest{
		UserID: &userID,
		Page:   1,
		Limit:  100,
	}, &userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user characters: %w", err)
	}

	// 为每个关系类型收集相关角色
	for _, relType := range relationshipTypes {
		group := &domain.RelationshipGroup{
			RelationshipType: relType,
			Characters:       []*domain.RelationshipCharacter{},
			Description:      relType.Description,
			CustomPrompt:     nil,
		}

		// 遍历用户的所有角色，查找该关系类型的关系
		for _, character := range characterList.Characters {
			relationships, err := s.relationshipService.GetRelationshipsByType(ctx, character.ID, relType.ID)
			if err != nil {
				continue
			}

			// 为每个关系创建RelationshipCharacter
			for _, rel := range relationships {
				relChar := &domain.RelationshipCharacter{
					Character:    &character,
					Relationship: rel.CharacterRelationshipEnhanced,
					MatchScore:   s.calculateRelationshipStrength(rel.CharacterRelationshipEnhanced),
				}

				group.Characters = append(group.Characters, relChar)
				relationshipMap.TotalCharacters++
			}
		}

		if len(group.Characters) > 0 {
			relationshipMap.RelationshipGroups[relType.Name] = group
		}
	}

	return relationshipMap, nil
}

// MatchCharactersByRelationship 基于关系描述匹配合适的角色
func (s *relationshipDrivenInvitationService) MatchCharactersByRelationship(ctx context.Context, userID uuid.UUID, relationshipHint string, chatContext string) ([]*domain.RelationshipMatchResult, error) {
	relationshipMap, err := s.GetUserRelationshipMap(ctx, userID)
	if err != nil {
		return nil, err
	}

	matches := []*domain.RelationshipMatchResult{}

	// 遍历所有关系组，寻找匹配的角色
	for relationshipName, group := range relationshipMap.RelationshipGroups {
		// 检查关系名称是否匹配
		description := ""
		if group.Description != nil {
			description = *group.Description
		}
		if s.isRelationshipMatch(relationshipName, description, relationshipHint) {
			// 为该关系组中的每个角色创建匹配结果
			for _, relChar := range group.Characters {
				match := &domain.RelationshipMatchResult{
					Character:        relChar.Character,
					Relationship:     relChar.Relationship,
					RelationshipType: group.RelationshipType,
					MatchScore:       s.calculateContextualMatchScore(relChar, relationshipHint, chatContext),
					MatchReason:      s.generateMatchReason(relChar, relationshipHint, chatContext),
				}
				matches = append(matches, match)
			}
		}
	}

	// 按匹配度排序
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].MatchScore < matches[j].MatchScore {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches, nil
}

// GenerateRelationshipBasedSuggestion 生成基于关系的邀请建议
func (s *relationshipDrivenInvitationService) GenerateRelationshipBasedSuggestion(ctx context.Context, match *domain.RelationshipMatchResult, chatContext string) (*domain.RelationshipInvitationSuggestion, error) {
	suggestion := &domain.RelationshipInvitationSuggestion{
		Character:        match.Character,
		Relationship:     match.Relationship,
		RelationshipType: match.RelationshipType,
		MatchScore:       match.MatchScore,
		InviteReason:     s.generateInviteReason(match, chatContext),
		ExpectedReaction: s.generateExpectedReaction(match),
		RelationshipDesc: s.generateRelationshipDescription(match),
		Priority:         s.calculatePriority(match),
		CreatedAt:        time.Now(),
	}

	return suggestion, nil
}

// 辅助方法

// extractRelationshipHints 从聊天内容中提取关系提示
func (s *relationshipDrivenInvitationService) extractRelationshipHints(message string) []string {
	hints := []string{}
	message = strings.ToLower(message)

	// 关系关键词映射
	relationshipKeywords := map[string][]string{
		"朋友":  {"朋友", "好友", "兄弟", "姐妹", "伙伴", "buddy", "friend"},
		"恋人":  {"恋人", "男友", "女友", "爱人", "情人", "boyfriend", "girlfriend", "lover"},
		"家人":  {"家人", "亲人", "父母", "兄弟姐妹", "family"},
		"同事":  {"同事", "同僚", "工作伙伴", "colleague"},
		"老师":  {"老师", "导师", "师父", "teacher", "mentor"},
		"学生":  {"学生", "徒弟", "student"},
		"敌人":  {"敌人", "对手", "竞争对手", "enemy", "rival"},
		"陌生人": {"陌生人", "不认识", "stranger"},
	}

	for relationship, keywords := range relationshipKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				hints = append(hints, relationship)
				break
			}
		}
	}

	return hints
}

// isRelationshipMatch 检查关系是否匹配
func (s *relationshipDrivenInvitationService) isRelationshipMatch(relationshipName, description, hint string) bool {
	relationshipName = strings.ToLower(relationshipName)
	description = strings.ToLower(description)
	hint = strings.ToLower(hint)

	// 直接名称匹配
	if strings.Contains(relationshipName, hint) || strings.Contains(hint, relationshipName) {
		return true
	}

	// 描述匹配
	if description != "" && (strings.Contains(description, hint) || strings.Contains(hint, description)) {
		return true
	}

	return false
}

// calculateRelationshipStrength 计算关系强度
func (s *relationshipDrivenInvitationService) calculateRelationshipStrength(rel *domain.CharacterRelationshipEnhanced) float64 {
	// 基于多维度计算关系强度
	strength := (rel.Strength + rel.Trust + rel.Affection + rel.Respect + rel.Intimacy) / 5.0
	return strength
}

// calculateContextualMatchScore 计算上下文匹配分数
func (s *relationshipDrivenInvitationService) calculateContextualMatchScore(relChar *domain.RelationshipCharacter, hint string, context string) float64 {
	baseScore := relChar.MatchScore

	// 基于角色性格和聊天上下文调整分数
	if relChar.Character.Personality != nil {
		personality := strings.ToLower(*relChar.Character.Personality)
		context = strings.ToLower(context)

		// 情感需求匹配
		if strings.Contains(context, "安慰") || strings.Contains(context, "难过") {
			if strings.Contains(personality, "温柔") || strings.Contains(personality, "体贴") {
				baseScore += 0.2
			}
		}

		// 活动需求匹配
		if strings.Contains(context, "热闹") || strings.Contains(context, "开心") {
			if strings.Contains(personality, "开朗") || strings.Contains(personality, "活泼") {
				baseScore += 0.2
			}
		}

		// 智力需求匹配
		if strings.Contains(context, "讨论") || strings.Contains(context, "思考") {
			if strings.Contains(personality, "聪明") || strings.Contains(personality, "智慧") {
				baseScore += 0.2
			}
		}
	}

	// 确保分数在0-1范围内
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

// generateMatchReason 生成匹配原因
func (s *relationshipDrivenInvitationService) generateMatchReason(relChar *domain.RelationshipCharacter, hint string, context string) string {
	relationshipName := "朋友"
	if relChar.Relationship.CustomTypeName != nil && *relChar.Relationship.CustomTypeName != "" {
		relationshipName = *relChar.Relationship.CustomTypeName
	}

	characterName := relChar.Character.Name

	reasons := []string{
		fmt.Sprintf("%s是你的%s", characterName, relationshipName),
		fmt.Sprintf("你们的%s关系很好", relationshipName),
	}

	// 基于上下文添加具体原因
	if relChar.Character.Personality != nil {
		personality := *relChar.Character.Personality
		if strings.Contains(strings.ToLower(context), "安慰") && strings.Contains(strings.ToLower(personality), "温柔") {
			reasons = append(reasons, fmt.Sprintf("%s很温柔，擅长安慰人", characterName))
		}
		if strings.Contains(strings.ToLower(context), "讨论") && strings.Contains(strings.ToLower(personality), "聪明") {
			reasons = append(reasons, fmt.Sprintf("%s很聪明，适合深度讨论", characterName))
		}
	}

	return strings.Join(reasons, "，")
}

// generateInviteReason 生成邀请理由
func (s *relationshipDrivenInvitationService) generateInviteReason(match *domain.RelationshipMatchResult, context string) string {
	relationshipName := "朋友"
	if match.Relationship.CustomTypeName != nil && *match.Relationship.CustomTypeName != "" {
		relationshipName = *match.Relationship.CustomTypeName
	}
	return fmt.Sprintf("基于你们的%s关系，%s", relationshipName, match.MatchReason)
}

// generateExpectedReaction 生成预期反应
func (s *relationshipDrivenInvitationService) generateExpectedReaction(match *domain.RelationshipMatchResult) string {
	strength := s.calculateRelationshipStrength(match.Relationship)

	if strength >= 0.8 {
		return "应该会很高兴地同意"
	} else if strength >= 0.6 {
		return "可能会考虑一下，但应该会同意"
	} else if strength >= 0.4 {
		return "可能会有些犹豫"
	} else {
		return "可能需要一些说服"
	}
}

// generateRelationshipDescription 生成关系描述
func (s *relationshipDrivenInvitationService) generateRelationshipDescription(match *domain.RelationshipMatchResult) string {
	relationshipName := "朋友"
	if match.Relationship.CustomTypeName != nil && *match.Relationship.CustomTypeName != "" {
		relationshipName = *match.Relationship.CustomTypeName
	}
	return fmt.Sprintf("你们是%s关系，关系强度: %.1f", relationshipName, s.calculateRelationshipStrength(match.Relationship))
}

// calculatePriority 计算优先级
func (s *relationshipDrivenInvitationService) calculatePriority(match *domain.RelationshipMatchResult) int {
	if match.MatchScore >= 0.8 {
		return 1 // 高优先级
	} else if match.MatchScore >= 0.6 {
		return 2 // 中优先级
	} else {
		return 3 // 低优先级
	}
}

// filterExistingGroupMembers 过滤已在群聊中的角色
func (s *relationshipDrivenInvitationService) filterExistingGroupMembers(ctx context.Context, matches []*domain.RelationshipMatchResult, groupChatID uuid.UUID, userID uuid.UUID) ([]*domain.RelationshipMatchResult, error) {
	// 获取群聊成员
	members, err := s.characterService.GetGroupChatMembers(ctx, groupChatID, userID)
	if err != nil {
		return nil, err
	}

	// 创建已存在角色的映射
	existingCharacters := make(map[uuid.UUID]bool)
	for _, member := range members {
		if member.CharacterID != nil {
			existingCharacters[*member.CharacterID] = true
		}
	}

	// 过滤已存在的角色
	filtered := []*domain.RelationshipMatchResult{}
	for _, match := range matches {
		if !existingCharacters[match.Character.ID] {
			filtered = append(filtered, match)
		}
	}

	return filtered, nil
}
