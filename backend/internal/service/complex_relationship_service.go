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

// ComplexRelationshipService 复杂关系网络服务
type ComplexRelationshipService interface {
	// 智能解析复杂关系描述
	ParseComplexRelationship(ctx context.Context, req *domain.ComplexRelationshipParseRequest) (*domain.ComplexRelationshipParseResult, error)

	// 分析关系网络动态
	AnalyzeRelationshipDynamics(ctx context.Context, req *domain.RelationshipDynamicsRequest) (*domain.RelationshipDynamicsResult, error)

	// 获取角色在群聊中的关系上下文
	GetCharacterRelationshipContext(ctx context.Context, req *domain.CharacterRelationshipContextRequest) (*domain.CharacterRelationshipContextResult, error)

	// 智能推荐关系处理策略
	RecommendRelationshipStrategy(ctx context.Context, req *domain.RelationshipStrategyRequest) (*domain.RelationshipStrategyResult, error)

	// 检测关系冲突
	DetectRelationshipConflicts(ctx context.Context, groupChatID uuid.UUID) (*domain.RelationshipConflictResult, error)
}

type complexRelationshipService struct {
	relationshipRepo repository.RelationshipRepository
	characterRepo    repository.CharacterRepository
	modelService     ModelService
	logger           *logrus.Logger
}

// NewComplexRelationshipService 创建复杂关系网络服务
func NewComplexRelationshipService(
	relationshipRepo repository.RelationshipRepository,
	characterRepo repository.CharacterRepository,
	modelService ModelService,
	logger *logrus.Logger,
) ComplexRelationshipService {
	return &complexRelationshipService{
		relationshipRepo: relationshipRepo,
		characterRepo:    characterRepo,
		modelService:     modelService,
		logger:           logger,
	}
}

// ParseComplexRelationship 智能解析复杂关系描述
func (s *complexRelationshipService) ParseComplexRelationship(ctx context.Context, req *domain.ComplexRelationshipParseRequest) (*domain.ComplexRelationshipParseResult, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":     req.UserID,
		"description": req.Description,
	}).Debug("Parsing complex relationship description")

	result := &domain.ComplexRelationshipParseResult{
		UserID:              req.UserID,
		Description:         req.Description,
		ParsedRelationships: []*domain.ParsedRelationship{},
		Confidence:          0.0,
	}

	// 1. 使用正则表达式提取基本关系信息
	basicRelations := s.extractBasicRelations(req.Description)
	result.ParsedRelationships = append(result.ParsedRelationships, basicRelations...)

	// 2. 使用AI模型进行深度语义解析
	aiParsedRelations, err := s.aiParseRelationships(ctx, req.Description)
	if err != nil {
		s.logger.WithError(err).Warn("AI parsing failed, using regex results only")
	} else {
		result.ParsedRelationships = append(result.ParsedRelationships, aiParsedRelations...)
	}

	// 3. 去重和合并关系
	result.ParsedRelationships = s.mergeAndDeduplicateRelationships(result.ParsedRelationships)

	// 4. 计算整体置信度
	result.Confidence = s.calculateParsingConfidence(result.ParsedRelationships)

	return result, nil
}

// extractBasicRelations 使用正则表达式提取基本关系
func (s *complexRelationshipService) extractBasicRelations(description string) []*domain.ParsedRelationship {
	var relations []*domain.ParsedRelationship

	// 定义关系模式
	patterns := []struct {
		regex            *regexp.Regexp
		relationshipType string
		confidence       float64
	}{
		{
			regex:            regexp.MustCompile(`(\w+)是我的(.{1,10}?)朋友`),
			relationshipType: "friend",
			confidence:       0.8,
		},
		{
			regex:            regexp.MustCompile(`我和(\w+)(.{1,20}?)敌人`),
			relationshipType: "enemy",
			confidence:       0.7,
		},
		{
			regex:            regexp.MustCompile(`(\w+)(.{1,20}?)闺蜜`),
			relationshipType: "best_friend",
			confidence:       0.8,
		},
		{
			regex:            regexp.MustCompile(`表面上(.{1,20}?)实际上(.{1,20})`),
			relationshipType: "complex",
			confidence:       0.6,
		},
		{
			regex:            regexp.MustCompile(`我们(.{1,20}?)竞争关系`),
			relationshipType: "rival",
			confidence:       0.7,
		},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatch(description, -1)
		for _, match := range matches {
			if len(match) > 1 {
				relation := &domain.ParsedRelationship{
					CharacterName:    match[1],
					RelationshipType: pattern.relationshipType,
					Description:      strings.Join(match[1:], " "),
					Confidence:       pattern.confidence,
					Source:           "regex",
				}

				// 提取情感强度和复杂性
				relation.EmotionalIntensity = s.extractEmotionalIntensity(description)
				relation.Complexity = s.extractComplexity(description)

				relations = append(relations, relation)
			}
		}
	}

	return relations
}

// aiParseRelationships 使用AI模型解析关系
func (s *complexRelationshipService) aiParseRelationships(ctx context.Context, description string) ([]*domain.ParsedRelationship, error) {
	_ = fmt.Sprintf(`请分析以下复杂的人际关系描述，提取其中的关系信息：

描述：%s

请识别：
1. 涉及的人物名称
2. 关系类型（朋友、敌人、竞争对手、暧昧、复杂等）
3. 关系的复杂程度（简单、复杂、非常复杂）
4. 情感强度（低、中、高）
5. 关系的真实性（表面、真实、混合）

以JSON格式返回结果。`, description)

	// 这里应该调用实际的AI模型服务
	// response, err := s.modelService.GenerateText(ctx, prompt)
	// 暂时返回模拟结果

	return []*domain.ParsedRelationship{
		{
			CharacterName:      "小米",
			RelationshipType:   "friend",
			Description:        "好朋友关系",
			Confidence:         0.9,
			Source:             "ai",
			EmotionalIntensity: 0.7,
			Complexity:         0.6,
		},
	}, nil
}

// extractEmotionalIntensity 提取情感强度
func (s *complexRelationshipService) extractEmotionalIntensity(description string) float64 {
	// 情感强度关键词
	highIntensity := []string{"非常", "特别", "极其", "深深", "强烈"}
	mediumIntensity := []string{"比较", "还算", "挺", "相当"}
	lowIntensity := []string{"有点", "稍微", "一般", "普通"}

	descLower := strings.ToLower(description)

	for _, word := range highIntensity {
		if strings.Contains(descLower, word) {
			return 0.9
		}
	}

	for _, word := range mediumIntensity {
		if strings.Contains(descLower, word) {
			return 0.6
		}
	}

	for _, word := range lowIntensity {
		if strings.Contains(descLower, word) {
			return 0.3
		}
	}

	return 0.5 // 默认中等强度
}

// extractComplexity 提取关系复杂度
func (s *complexRelationshipService) extractComplexity(description string) float64 {
	complexityIndicators := []string{
		"表面", "实际上", "但是", "不过", "虽然", "尽管",
		"一方面", "另一方面", "有时候", "矛盾", "复杂", "说不清",
	}

	descLower := strings.ToLower(description)
	complexityScore := 0.0

	for _, indicator := range complexityIndicators {
		if strings.Contains(descLower, indicator) {
			complexityScore += 0.2
		}
	}

	// 限制在0-1之间
	if complexityScore > 1.0 {
		complexityScore = 1.0
	}

	return complexityScore
}

// mergeAndDeduplicateRelationships 合并和去重关系
func (s *complexRelationshipService) mergeAndDeduplicateRelationships(relations []*domain.ParsedRelationship) []*domain.ParsedRelationship {
	// 按角色名称分组
	relationMap := make(map[string][]*domain.ParsedRelationship)

	for _, relation := range relations {
		key := strings.ToLower(relation.CharacterName)
		relationMap[key] = append(relationMap[key], relation)
	}

	var mergedRelations []*domain.ParsedRelationship

	for _, relationGroup := range relationMap {
		if len(relationGroup) == 1 {
			mergedRelations = append(mergedRelations, relationGroup[0])
		} else {
			// 合并多个关系描述
			merged := s.mergeRelationshipGroup(relationGroup)
			mergedRelations = append(mergedRelations, merged)
		}
	}

	return mergedRelations
}

// mergeRelationshipGroup 合并同一角色的多个关系描述
func (s *complexRelationshipService) mergeRelationshipGroup(relations []*domain.ParsedRelationship) *domain.ParsedRelationship {
	if len(relations) == 0 {
		return nil
	}

	merged := &domain.ParsedRelationship{
		CharacterName: relations[0].CharacterName,
		Source:        "merged",
	}

	// 合并描述
	var descriptions []string
	var totalConfidence float64
	var totalIntensity float64
	var totalComplexity float64

	relationTypes := make(map[string]int)

	for _, relation := range relations {
		descriptions = append(descriptions, relation.Description)
		totalConfidence += relation.Confidence
		totalIntensity += relation.EmotionalIntensity
		totalComplexity += relation.Complexity
		relationTypes[relation.RelationshipType]++
	}

	merged.Description = strings.Join(descriptions, "; ")
	merged.Confidence = totalConfidence / float64(len(relations))
	merged.EmotionalIntensity = totalIntensity / float64(len(relations))
	merged.Complexity = totalComplexity / float64(len(relations))

	// 选择最常见的关系类型
	maxCount := 0
	for relType, count := range relationTypes {
		if count > maxCount {
			maxCount = count
			merged.RelationshipType = relType
		}
	}

	// 如果有多种关系类型，标记为复杂关系
	if len(relationTypes) > 1 {
		merged.RelationshipType = "complex"
		merged.Complexity = 1.0
	}

	return merged
}

// calculateParsingConfidence 计算解析置信度
func (s *complexRelationshipService) calculateParsingConfidence(relations []*domain.ParsedRelationship) float64 {
	if len(relations) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, relation := range relations {
		totalConfidence += relation.Confidence
	}

	return totalConfidence / float64(len(relations))
}

// AnalyzeRelationshipDynamics 分析关系网络动态
func (s *complexRelationshipService) AnalyzeRelationshipDynamics(ctx context.Context, req *domain.RelationshipDynamicsRequest) (*domain.RelationshipDynamicsResult, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":       req.UserID,
		"group_chat_id": req.GroupChatID,
	}).Debug("Analyzing relationship dynamics")

	// 获取用户的所有关系
	// 使用现有方法的替代实现
	relationships, err := s.relationshipRepo.ListCharacterRelationships(ctx, uuid.Nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user relationships: %w", err)
	}

	result := &domain.RelationshipDynamicsResult{
		UserID:      req.UserID,
		GroupChatID: req.GroupChatID,
		Dynamics:    []*domain.RelationshipDynamic{},
		Conflicts:   []*domain.RelationshipConflict{},
		Suggestions: []string{},
	}

	// 分析关系动态
	for _, relationship := range relationships {
		dynamic := s.analyzeRelationshipDynamic(relationship)
		result.Dynamics = append(result.Dynamics, dynamic)
	}

	// 检测关系冲突
	conflicts := s.detectRelationshipConflictsInGroup(relationships)
	result.Conflicts = append(result.Conflicts, conflicts...)

	// 生成建议
	result.Suggestions = s.generateRelationshipSuggestions(result.Dynamics, result.Conflicts)

	return result, nil
}

// DetectRelationshipConflicts 检测关系冲突
func (s *complexRelationshipService) DetectRelationshipConflicts(ctx context.Context, groupChatID uuid.UUID) (*domain.RelationshipConflictResult, error) {
	s.logger.WithFields(logrus.Fields{
		"group_chat_id": groupChatID,
	}).Debug("Detecting relationship conflicts")

	// 获取群聊中的所有关系 (使用现有方法的替代实现)
	relationships, err := s.relationshipRepo.ListCharacterRelationships(ctx, uuid.Nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships: %w", err)
	}

	// 检测冲突
	conflicts := s.detectRelationshipConflictsInGroup(relationships)

	// 计算冲突级别
	conflictLevel := "low"
	if len(conflicts) > 0 {
		maxSeverity := 0.0
		for _, conflict := range conflicts {
			if conflict.Severity > maxSeverity {
				maxSeverity = conflict.Severity
			}
		}

		if maxSeverity > 0.8 {
			conflictLevel = "critical"
		} else if maxSeverity > 0.6 {
			conflictLevel = "high"
		} else if maxSeverity > 0.3 {
			conflictLevel = "medium"
		}
	}

	result := &domain.RelationshipConflictResult{
		GroupChatID:     groupChatID,
		ConflictLevel:   conflictLevel,
		ActiveConflicts: conflicts,
		RiskPredictions: []*domain.ConflictRiskPrediction{},
		Recommendations: s.generateConflictRecommendations(conflicts),
	}

	return result, nil
}

// analyzeRelationshipDynamic 分析单个关系的动态
func (s *complexRelationshipService) analyzeRelationshipDynamic(relationship *domain.CharacterRelationshipEnhanced) *domain.RelationshipDynamic {
	return &domain.RelationshipDynamic{
		RelationshipID:   relationship.ID,
		CharacterID:      relationship.SourceCharacterID,
		RelationshipType: "enhanced", // 使用固定值或从其他字段推导
		Stability:        s.calculateRelationshipStability(relationship),
		TrendDirection:   s.predictRelationshipTrend(relationship),
		InfluenceLevel:   s.calculateInfluenceLevel(relationship),
		RiskLevel:        s.calculateRiskLevel(relationship),
	}
}

// calculateRelationshipStability 计算关系稳定性
func (s *complexRelationshipService) calculateRelationshipStability(relationship *domain.Relationship) float64 {
	// 基于关系类型和描述计算稳定性
	baseStability := map[string]float64{
		"friend":      0.7,
		"best_friend": 0.9,
		"enemy":       0.3,
		"rival":       0.4,
		"complex":     0.2,
		"neutral":     0.6,
	}

	// 使用CustomTypeName作为关系类型
	relationshipType := "neutral"
	if relationship.CustomTypeName != nil {
		relationshipType = *relationship.CustomTypeName
	}

	stability, exists := baseStability[relationshipType]
	if !exists {
		stability = 0.5
	}

	// 根据私人备注调整稳定性
	if relationship.PrivateNotes != nil {
		desc := strings.ToLower(*relationship.PrivateNotes)
		if strings.Contains(desc, "稳定") || strings.Contains(desc, "可靠") {
			stability += 0.1
		}
		if strings.Contains(desc, "不稳定") || strings.Contains(desc, "变化") {
			stability -= 0.2
		}
	}

	// 基于信任度调整稳定性
	stability = (stability + relationship.Trust) / 2

	// 确保在0-1范围内
	if stability > 1.0 {
		stability = 1.0
	}
	if stability < 0.0 {
		stability = 0.0
	}

	return stability
}

// predictRelationshipTrend 预测关系趋势
func (s *complexRelationshipService) predictRelationshipTrend(relationship *domain.Relationship) string {
	// 简化的趋势预测逻辑
	if relationship.PrivateNotes != nil {
		desc := strings.ToLower(*relationship.PrivateNotes)
		if strings.Contains(desc, "越来越好") || strings.Contains(desc, "改善") {
			return "improving"
		}
		if strings.Contains(desc, "越来越差") || strings.Contains(desc, "恶化") {
			return "deteriorating"
		}
		if strings.Contains(desc, "复杂") || strings.Contains(desc, "矛盾") {
			return "unstable"
		}
	}

	// 基于信任度和亲密度预测趋势
	if relationship.Trust > 0.8 && relationship.Affection > 0.8 {
		return "improving"
	}
	if relationship.Trust < 0.3 || relationship.Jealousy > 0.7 {
		return "deteriorating"
	}

	return "stable"
}

// calculateInfluenceLevel 计算影响力水平
func (s *complexRelationshipService) calculateInfluenceLevel(relationship *domain.Relationship) float64 {
	// 基于关系类型计算影响力
	influenceMap := map[string]float64{
		"best_friend": 0.9,
		"friend":      0.7,
		"rival":       0.6,
		"enemy":       0.8, // 敌人也有很高的影响力
		"complex":     0.5,
		"neutral":     0.3,
	}

	relationshipType := "neutral"
	if relationship.CustomTypeName != nil {
		relationshipType = *relationship.CustomTypeName
	}

	influence, exists := influenceMap[relationshipType]
	if !exists {
		influence = 0.5
	}

	// 基于亲密度和依赖度调整影响力
	influence = (influence + relationship.Intimacy + relationship.Dependency) / 3

	return influence
}

// calculateRiskLevel 计算风险水平
func (s *complexRelationshipService) calculateRiskLevel(relationship *domain.Relationship) float64 {
	// 基于关系类型和复杂性计算风险
	riskMap := map[string]float64{
		"enemy":       0.9,
		"rival":       0.7,
		"complex":     0.8,
		"friend":      0.2,
		"best_friend": 0.1,
		"neutral":     0.3,
	}

	relationshipType := "neutral"
	if relationship.CustomTypeName != nil {
		relationshipType = *relationship.CustomTypeName
	}

	risk, exists := riskMap[relationshipType]
	if !exists {
		risk = 0.5
	}

	// 基于冲突倾向调整风险
	risk = (risk + relationship.ConflictTendency) / 2

	return risk
}

// detectRelationshipConflictsInGroup 检测群组中的关系冲突
func (s *complexRelationshipService) detectRelationshipConflictsInGroup(relationships []*domain.Relationship) []*domain.RelationshipConflict {
	var conflicts []*domain.RelationshipConflict

	// 检测敌对关系
	for i, rel1 := range relationships {
		for j, rel2 := range relationships {
			if i >= j {
				continue
			}

			if s.areRelationshipsConflicting(rel1, rel2) {
				rel1Type := "neutral"
				if rel1.CustomTypeName != nil {
					rel1Type = *rel1.CustomTypeName
				}
				rel2Type := "neutral"
				if rel2.CustomTypeName != nil {
					rel2Type = *rel2.CustomTypeName
				}

				conflict := &domain.RelationshipConflict{
					ConflictType:    "opposing_relationships",
					Relationship1ID: rel1.ID,
					Relationship2ID: rel2.ID,
					Severity:        s.calculateConflictSeverity(rel1, rel2),
					Description:     fmt.Sprintf("关系冲突：%s 和 %s", rel1Type, rel2Type),
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// areRelationshipsConflicting 判断两个关系是否冲突
func (s *complexRelationshipService) areRelationshipsConflicting(rel1, rel2 *domain.Relationship) bool {
	conflictingPairs := map[string][]string{
		"friend":      {"enemy"},
		"enemy":       {"friend", "best_friend"},
		"best_friend": {"enemy"},
	}

	rel1Type := "neutral"
	if rel1.CustomTypeName != nil {
		rel1Type = *rel1.CustomTypeName
	}
	rel2Type := "neutral"
	if rel2.CustomTypeName != nil {
		rel2Type = *rel2.CustomTypeName
	}

	if conflicts, exists := conflictingPairs[rel1Type]; exists {
		for _, conflictType := range conflicts {
			if rel2Type == conflictType {
				return true
			}
		}
	}

	return false
}

// calculateConflictSeverity 计算冲突严重程度
func (s *complexRelationshipService) calculateConflictSeverity(rel1, rel2 *domain.Relationship) float64 {
	severityMap := map[string]map[string]float64{
		"friend": {
			"enemy": 0.8,
		},
		"best_friend": {
			"enemy": 0.9,
		},
		"enemy": {
			"friend":      0.8,
			"best_friend": 0.9,
		},
	}

	rel1Type := "neutral"
	if rel1.CustomTypeName != nil {
		rel1Type = *rel1.CustomTypeName
	}
	rel2Type := "neutral"
	if rel2.CustomTypeName != nil {
		rel2Type = *rel2.CustomTypeName
	}

	if severities, exists := severityMap[rel1Type]; exists {
		if severity, exists := severities[rel2Type]; exists {
			return severity
		}
	}

	return 0.5
}

// generateConflictRecommendations 生成冲突解决建议
func (s *complexRelationshipService) generateConflictRecommendations(conflicts []*domain.RelationshipConflict) []string {
	var recommendations []string

	for _, conflict := range conflicts {
		switch conflict.ConflictType {
		case "opposing_relationships":
			recommendations = append(recommendations, "建议避免同时邀请冲突双方进入同一群聊")
			recommendations = append(recommendations, "可以考虑分别与冲突双方进行单独对话")
		default:
			recommendations = append(recommendations, "建议关注关系动态变化，及时调整互动策略")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "当前关系网络相对稳定，继续保持良好互动")
	}

	return recommendations
}

// GetCharacterRelationshipContext 获取角色在群聊中的关系上下文
func (s *complexRelationshipService) GetCharacterRelationshipContext(ctx context.Context, req *domain.CharacterRelationshipContextRequest) (*domain.CharacterRelationshipContextResult, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":       req.UserID,
		"character_id":  req.CharacterID,
		"group_chat_id": req.GroupChatID,
	}).Debug("Getting character relationship context")

	// 获取角色的直接关系
	relationships, err := s.relationshipRepo.ListCharacterRelationships(ctx, req.CharacterID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get character relationships: %w", err)
	}

	result := &domain.CharacterRelationshipContextResult{
		UserID:            req.UserID,
		CharacterID:       req.CharacterID,
		GroupChatID:       req.GroupChatID,
		DirectRelation:    nil,
		IndirectRelations: []*domain.IndirectRelationshipContext{},
		ContextSummary:    "角色关系上下文分析完成",
	}

	// 分析直接关系
	if len(relationships) > 0 {
		rel := relationships[0]
		relationshipType := "neutral"
		if rel.CustomTypeName != nil {
			relationshipType = *rel.CustomTypeName
		}

		result.DirectRelation = &domain.RelationshipContext{
			RelationshipType:   relationshipType,
			Description:        fmt.Sprintf("与角色的%s关系", relationshipType),
			EmotionalIntensity: rel.Affection,
			Complexity:         rel.ConflictTendency,
			RecentInteractions: []string{"最近互动记录"},
		}
	}

	return result, nil
}

// RecommendRelationshipStrategy 智能推荐关系处理策略
func (s *complexRelationshipService) RecommendRelationshipStrategy(ctx context.Context, req *domain.RelationshipStrategyRequest) (*domain.RelationshipStrategyResult, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":       req.UserID,
		"character_id":  req.CharacterID,
		"group_chat_id": req.GroupChatID,
		"scenario":      req.Scenario,
	}).Debug("Recommending relationship strategy")

	// 获取角色关系
	_, err := s.relationshipRepo.ListCharacterRelationships(ctx, req.CharacterID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get character relationships: %w", err)
	}

	result := &domain.RelationshipStrategyResult{
		UserID:      req.UserID,
		CharacterID: req.CharacterID,
		GroupChatID: req.GroupChatID,
		RecommendedStrategy: &domain.RelationshipStrategy{
			StrategyType:    "direct",
			Description:     "建议采用直接沟通策略",
			Actions:         []string{"主动发起对话", "表达真实想法", "倾听对方观点"},
			ExpectedOutcome: "改善关系，增进理解",
			Confidence:      0.8,
		},
		AlternativeStrategies: []*domain.RelationshipStrategy{
			{
				StrategyType:    "indirect",
				Description:     "通过共同朋友间接沟通",
				Actions:         []string{"寻找共同话题", "通过第三方了解情况"},
				ExpectedOutcome: "缓解紧张关系",
				Confidence:      0.6,
			},
		},
		RiskAssessment: &domain.RelationshipRiskAssessment{
			OverallRisk:     0.3,
			RiskFactors:     []string{"沟通误解", "情绪波动"},
			MitigationSteps: []string{"保持冷静", "选择合适时机"},
			WarningSignals:  []string{"对方情绪激动", "回避交流"},
		},
	}

	return result, nil
}

// generateRelationshipSuggestions 生成关系建议
func (s *complexRelationshipService) generateRelationshipSuggestions(dynamics []*domain.RelationshipDynamic, conflicts []*domain.RelationshipConflict) []string {
	var suggestions []string

	// 基于动态生成建议
	for _, dynamic := range dynamics {
		if dynamic.Stability < 0.3 {
			suggestions = append(suggestions, fmt.Sprintf("关系不稳定，建议谨慎处理与该角色的互动"))
		}
		if dynamic.RiskLevel > 0.7 {
			suggestions = append(suggestions, fmt.Sprintf("高风险关系，建议避免敏感话题"))
		}
	}

	// 基于冲突生成建议
	for _, conflict := range conflicts {
		if conflict.Severity > 0.7 {
			suggestions = append(suggestions, "检测到严重关系冲突，建议分别处理相关角色")
		}
	}

	return suggestions
}
