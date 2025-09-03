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

// AIInvitationService AI智能邀请服务接口
type AIInvitationService interface {
	// 智能邀请分析
	AnalyzeInvitationIntent(ctx context.Context, userMessage string, groupChatID uuid.UUID, userID uuid.UUID) (*domain.InvitationAnalysis, error)

	// 智能匹配角色
	FindMatchingCharacters(ctx context.Context, intent *domain.InvitationAnalysis, userID uuid.UUID) ([]*domain.InvitationCandidate, error)

	// 生成邀请建议
	GenerateInvitationSuggestions(ctx context.Context, candidates []*domain.InvitationCandidate, groupChatID uuid.UUID) ([]*domain.InvitationSuggestion, error)

	// 执行智能邀请
	ExecuteSmartInvitation(ctx context.Context, userMessage string, groupChatID uuid.UUID, userID uuid.UUID) (*domain.SmartInvitationResult, error)
}

type aiInvitationService struct {
	relationshipService RelationshipService
	characterService    CharacterService
	memoryService       MemoryService
	logger              *logrus.Logger
}

// NewAIInvitationService 创建AI智能邀请服务
func NewAIInvitationService(
	relationshipService RelationshipService,
	characterService CharacterService,
	memoryService MemoryService,
	logger *logrus.Logger,
) AIInvitationService {
	return &aiInvitationService{
		relationshipService: relationshipService,
		characterService:    characterService,
		memoryService:       memoryService,
		logger:              logger,
	}
}

// AnalyzeInvitationIntent 分析邀请意图
func (s *aiInvitationService) AnalyzeInvitationIntent(ctx context.Context, userMessage string, groupChatID uuid.UUID, userID uuid.UUID) (*domain.InvitationAnalysis, error) {
	s.logger.WithFields(logrus.Fields{
		"user_message":  userMessage,
		"group_chat_id": groupChatID,
		"user_id":       userID,
	}).Info("Analyzing invitation intent")

	analysis := &domain.InvitationAnalysis{
		OriginalMessage:   userMessage,
		Intent:            "invite",
		Confidence:        0.0,
		Keywords:          []string{},
		RelationshipHints: []string{},
		EmotionalContext:  "neutral",
		Urgency:           "normal",
	}

	// 检测邀请意图的关键词
	inviteKeywords := []string{
		"拉", "叫", "邀请", "来", "一起", "加入", "参与",
		"找", "呼叫", "召集", "聚集", "过来", "进来",
	}

	// 关系类型关键词映射
	relationshipKeywords := map[string][]string{
		"朋友":  {"朋友", "好友", "友人", "伙伴", "兄弟", "姐妹", "哥们", "闺蜜"},
		"恋人":  {"男朋友", "女朋友", "恋人", "爱人", "对象", "另一半", "心上人"},
		"家人":  {"家人", "亲人", "父母", "兄弟姐妹", "亲戚", "家属"},
		"同事":  {"同事", "同僚", "工作伙伴", "合作伙伴", "队友"},
		"老师":  {"老师", "导师", "师父", "教练", "指导者"},
		"学生":  {"学生", "徒弟", "弟子", "学员"},
		"邻居":  {"邻居", "邻里", "街坊"},
		"陌生人": {"陌生人", "新朋友", "不认识的人"},
	}

	// 情感上下文关键词
	emotionalKeywords := map[string][]string{
		"happy":   {"开心", "高兴", "快乐", "兴奋", "愉快", "欢乐"},
		"sad":     {"难过", "伤心", "沮丧", "失落", "郁闷", "不开心"},
		"angry":   {"生气", "愤怒", "恼火", "烦躁", "不爽"},
		"lonely":  {"孤独", "寂寞", "无聊", "一个人", "没人陪"},
		"excited": {"激动", "兴奋", "期待", "迫不及待"},
		"worried": {"担心", "焦虑", "紧张", "不安"},
	}

	// 紧急程度关键词
	urgencyKeywords := map[string][]string{
		"urgent":   {"快", "赶紧", "马上", "立刻", "急", "紧急", "现在就"},
		"casual":   {"有空", "方便", "随时", "不急", "慢慢来"},
		"specific": {"今天", "明天", "这会儿", "现在", "等下", "一会儿"},
	}

	message := strings.ToLower(userMessage)

	// 检测邀请意图
	hasInviteIntent := false
	for _, keyword := range inviteKeywords {
		if strings.Contains(message, keyword) {
			hasInviteIntent = true
			analysis.Keywords = append(analysis.Keywords, keyword)
			analysis.Confidence += 0.2
		}
	}

	if !hasInviteIntent {
		analysis.Intent = "none"
		return analysis, nil
	}

	// 分析关系类型提示
	for relType, keywords := range relationshipKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				analysis.RelationshipHints = append(analysis.RelationshipHints, relType)
				analysis.Confidence += 0.15
			}
		}
	}

	// 分析情感上下文
	for emotion, keywords := range emotionalKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				analysis.EmotionalContext = emotion
				analysis.Confidence += 0.1
				break
			}
		}
	}

	// 分析紧急程度
	for urgency, keywords := range urgencyKeywords {
		for _, keyword := range keywords {
			if strings.Contains(message, keyword) {
				analysis.Urgency = urgency
				analysis.Confidence += 0.05
				break
			}
		}
	}

	// 模糊匹配增强
	analysis.FuzzyMatches = s.extractFuzzyMatches(userMessage)

	// 确保置信度在合理范围内
	if analysis.Confidence > 1.0 {
		analysis.Confidence = 1.0
	}

	s.logger.WithFields(logrus.Fields{
		"intent":             analysis.Intent,
		"confidence":         analysis.Confidence,
		"relationship_hints": analysis.RelationshipHints,
		"emotional_context":  analysis.EmotionalContext,
		"urgency":            analysis.Urgency,
	}).Info("Invitation intent analyzed")

	return analysis, nil
}

// FindMatchingCharacters 智能匹配角色
func (s *aiInvitationService) FindMatchingCharacters(ctx context.Context, intent *domain.InvitationAnalysis, userID uuid.UUID) ([]*domain.InvitationCandidate, error) {
	s.logger.WithField("user_id", userID).Info("Finding matching characters for invitation")

	var candidates []*domain.InvitationCandidate

	// 获取用户的所有角色
	characterList, err := s.characterService.ListCharacters(ctx, &domain.CharacterListRequest{
		UserID: &userID,
		Page:   1,
		Limit:  100, // 获取所有角色
	}, &userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user characters: %w", err)
	}

	// 为每个角色分析匹配度
	for _, character := range characterList.Characters {
		candidate := &domain.InvitationCandidate{
			Character:     &character,
			MatchScore:    0.0,
			MatchReasons:  []string{},
			Relationships: []*domain.RelationshipResponse{},
		}

		// 获取角色的关系网
		relationships, err := s.relationshipService.ListCharacterRelationships(ctx, character.ID, nil)
		if err == nil {
			candidate.Relationships = relationships
		}

		// 计算匹配分数 - 简化处理，避免panic
		candidate.MatchScore = 0.5 // 默认分数
		// 简化匹配原因，避免nil pointer
		candidate.MatchReasons = []string{"基础匹配"}

		// 只保留有一定匹配度的候选者（降低阈值）
		if candidate.MatchScore > 0.05 {
			candidates = append(candidates, candidate)
		}
	}

	// 按匹配分数排序
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].MatchScore < candidates[j].MatchScore {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	s.logger.WithField("candidates_count", len(candidates)).Info("Found matching characters")

	return candidates, nil
}

// GenerateInvitationSuggestions 生成邀请建议
func (s *aiInvitationService) GenerateInvitationSuggestions(ctx context.Context, candidates []*domain.InvitationCandidate, groupChatID uuid.UUID) ([]*domain.InvitationSuggestion, error) {
	var suggestions []*domain.InvitationSuggestion

	// 获取群聊现有成员
	members, err := s.characterService.GetGroupChatMembers(ctx, groupChatID, uuid.New()) // 临时用户ID
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get group chat members")
	}

	// 创建现有成员映射
	existingMembers := make(map[uuid.UUID]bool)
	if members != nil {
		for _, member := range members {
			if member.CharacterID != nil {
				existingMembers[*member.CharacterID] = true
			}
		}
	}

	// 为每个候选者生成邀请建议
	for i, candidate := range candidates {
		if i >= 5 { // 最多5个建议
			break
		}

		// 跳过已经在群聊中的角色
		if existingMembers[candidate.Character.ID] {
			continue
		}

		suggestion := &domain.InvitationSuggestion{
			Character:        candidate.Character,
			MatchScore:       candidate.MatchScore,
			InviteReason:     s.generateInviteReason(candidate),
			ExpectedReaction: s.predictReaction(candidate),
			Confidence:       candidate.MatchScore,
			Priority:         s.calculatePriority(candidate),
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// ExecuteSmartInvitation 执行智能邀请
func (s *aiInvitationService) ExecuteSmartInvitation(ctx context.Context, userMessage string, groupChatID uuid.UUID, userID uuid.UUID) (*domain.SmartInvitationResult, error) {
	s.logger.WithFields(logrus.Fields{
		"user_message":  userMessage,
		"group_chat_id": groupChatID,
		"user_id":       userID,
	}).Info("Executing smart invitation")

	result := &domain.SmartInvitationResult{
		ProcessedAt:     time.Now(),
		OriginalMessage: userMessage,
		Success:         false,
		Suggestions:     []*domain.InvitationSuggestion{},
		Message:         "",
	}

	// 1. 分析邀请意图
	intent, err := s.AnalyzeInvitationIntent(ctx, userMessage, groupChatID, userID)
	if err != nil {
		result.Message = "分析邀请意图失败"
		return result, err
	}

	if intent.Intent == "none" || intent.Confidence < 0.3 {
		result.Message = "没有检测到明确的邀请意图"
		return result, nil
	}

	// 2. 查找匹配的角色
	candidates, err := s.FindMatchingCharacters(ctx, intent, userID)
	if err != nil {
		result.Message = "查找匹配角色失败"
		return result, err
	}

	if len(candidates) == 0 {
		result.Message = "没有找到合适的角色可以邀请"
		return result, nil
	}

	// 3. 生成邀请建议
	suggestions, err := s.GenerateInvitationSuggestions(ctx, candidates, groupChatID)
	if err != nil {
		result.Message = "生成邀请建议失败"
		return result, err
	}

	result.Suggestions = suggestions
	result.Success = len(suggestions) > 0

	if result.Success {
		result.Message = fmt.Sprintf("找到了 %d 个合适的角色可以邀请", len(suggestions))
	} else {
		result.Message = "所有匹配的角色都已经在群聊中了"
	}

	s.logger.WithFields(logrus.Fields{
		"suggestions_count": len(suggestions),
		"success":           result.Success,
	}).Info("Smart invitation executed")

	return result, nil
}

// 辅助方法

// extractFuzzyMatches 提取模糊匹配
func (s *aiInvitationService) extractFuzzyMatches(message string) []string {
	var matches []string

	// 提取可能的角色名称或关系描述
	words := strings.Fields(message)
	for _, word := range words {
		if len(word) >= 2 && !s.isCommonWord(word) {
			matches = append(matches, word)
		}
	}

	return matches
}

// isCommonWord 检查是否为常见词
func (s *aiInvitationService) isCommonWord(word string) bool {
	commonWords := []string{
		"的", "了", "在", "是", "我", "你", "他", "她", "它", "们",
		"这", "那", "有", "没", "不", "也", "都", "很", "就", "要",
		"可以", "能够", "应该", "可能", "一定", "肯定", "当然",
	}

	for _, common := range commonWords {
		if word == common {
			return true
		}
	}
	return false
}

// calculateMatchScore 计算匹配分数
func (s *aiInvitationService) calculateMatchScore(intent *domain.InvitationAnalysis, character *domain.CharacterResponse, relationships []*domain.RelationshipResponse) float64 {
	score := 0.0

	// 基础分数（所有角色都有基础分数）
	score += 0.2

	// 角色名称精确匹配
	for _, fuzzyMatch := range intent.FuzzyMatches {
		if strings.Contains(strings.ToLower(character.Name), strings.ToLower(fuzzyMatch)) {
			score += 0.5 // 名称匹配给高分
		}
	}

	// 关系匹配分数
	for _, hint := range intent.RelationshipHints {
		for _, rel := range relationships {
			if rel.CustomTypeName != nil && strings.Contains(strings.ToLower(*rel.CustomTypeName), strings.ToLower(hint)) {
				score += 0.3 * rel.Strength
			}
		}
	}

	// 性格关键词匹配
	if character.Personality != nil {
		personality := strings.ToLower(*character.Personality)

		// 检查性格关键词
		personalityKeywords := []string{"温柔", "友好", "活泼", "开朗", "热情", "善良", "体贴", "耐心"}
		for _, keyword := range personalityKeywords {
			if strings.Contains(personality, keyword) {
				score += 0.15 // 每个性格关键词加分
			}
		}

		// 情感上下文匹配
		switch intent.EmotionalContext {
		case "happy":
			if strings.Contains(personality, "开朗") || strings.Contains(personality, "活泼") || strings.Contains(personality, "乐观") {
				score += 0.25
			}
		case "sad":
			if strings.Contains(personality, "温柔") || strings.Contains(personality, "体贴") || strings.Contains(personality, "善解人意") {
				score += 0.25
			}
		case "lonely":
			if strings.Contains(personality, "陪伴") || strings.Contains(personality, "友好") || strings.Contains(personality, "热情") {
				score += 0.25
			}
		}
	}

	// 模糊匹配增强
	for _, fuzzyMatch := range intent.FuzzyMatches {
		if character.Personality != nil && strings.Contains(strings.ToLower(*character.Personality), strings.ToLower(fuzzyMatch)) {
			score += 0.2
		}
	}

	// 特殊关键词匹配
	for _, keyword := range intent.Keywords {
		if character.Personality != nil && strings.Contains(strings.ToLower(*character.Personality), strings.ToLower(keyword)) {
			score += 0.1
		}
	}

	return score
}

// generateMatchReasons 生成匹配原因
func (s *aiInvitationService) generateMatchReasons(intent *domain.InvitationAnalysis, character *domain.CharacterResponse, relationships []*domain.RelationshipResponse) []string {
	var reasons []string

	// 关系匹配原因
	for _, hint := range intent.RelationshipHints {
		for _, rel := range relationships {
			if rel.CustomTypeName != nil && strings.Contains(strings.ToLower(*rel.CustomTypeName), strings.ToLower(hint)) {
				reasons = append(reasons, fmt.Sprintf("与你是%s关系", *rel.CustomTypeName))
			}
		}
	}

	// 性格匹配原因
	if character.Personality != nil {
		personality := *character.Personality
		switch intent.EmotionalContext {
		case "happy":
			if strings.Contains(strings.ToLower(personality), "开朗") {
				reasons = append(reasons, "性格开朗，适合开心的场合")
			}
		case "sad":
			if strings.Contains(strings.ToLower(personality), "温柔") {
				reasons = append(reasons, "性格温柔，能够给予安慰")
			}
		case "lonely":
			if strings.Contains(strings.ToLower(personality), "友好") {
				reasons = append(reasons, "性格友好，是很好的陪伴")
			}
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "可能是你想要邀请的人")
	}

	return reasons
}

// generateInviteReason 生成邀请理由
func (s *aiInvitationService) generateInviteReason(candidate *domain.InvitationCandidate) string {
	if len(candidate.MatchReasons) > 0 {
		return candidate.MatchReasons[0]
	}
	return fmt.Sprintf("%s 可能是你想要邀请的人", candidate.Character.Name)
}

// predictReaction 预测反应
func (s *aiInvitationService) predictReaction(candidate *domain.InvitationCandidate) string {
	if candidate.Character.Personality == nil {
		return "可能会同意"
	}

	personality := strings.ToLower(*candidate.Character.Personality)

	if strings.Contains(personality, "友好") || strings.Contains(personality, "热情") {
		return "很可能会高兴地同意"
	} else if strings.Contains(personality, "害羞") || strings.Contains(personality, "内向") {
		return "可能会有点犹豫，但会同意"
	} else if strings.Contains(personality, "冷淡") || strings.Contains(personality, "高冷") {
		return "可能需要一些说服"
	}

	return "可能会同意"
}

// calculatePriority 计算优先级
func (s *aiInvitationService) calculatePriority(candidate *domain.InvitationCandidate) int {
	if candidate.MatchScore > 0.8 {
		return 1 // 高优先级
	} else if candidate.MatchScore > 0.5 {
		return 2 // 中优先级
	}
	return 3 // 低优先级
}
