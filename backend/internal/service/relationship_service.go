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

// RelationshipService 关系服务接口
type RelationshipService interface {
	// 关系类型管理
	CreateRelationshipType(ctx context.Context, userID uuid.UUID, req *domain.CreateRelationshipTypeRequest) (*domain.RelationshipType, error)
	GetRelationshipType(ctx context.Context, id uuid.UUID) (*domain.RelationshipType, error)
	ListRelationshipTypes(ctx context.Context, userID *uuid.UUID, category *string, page, limit int) ([]*domain.RelationshipType, int, error)

	// 角色关系管理
	CreateCharacterRelationship(ctx context.Context, sourceCharacterID uuid.UUID, req *domain.CreateCharacterRelationshipRequest) (*domain.RelationshipResponse, error)
	GetCharacterRelationship(ctx context.Context, sourceID, targetID uuid.UUID) (*domain.RelationshipResponse, error)
	UpdateCharacterRelationship(ctx context.Context, relationshipID uuid.UUID, updates map[string]interface{}) (*domain.RelationshipResponse, error)
	DeleteCharacterRelationship(ctx context.Context, relationshipID uuid.UUID) error
	ListCharacterRelationships(ctx context.Context, characterID uuid.UUID, status *string) ([]*domain.RelationshipResponse, error)

	// 关系网络分析
	GetRelationshipNetwork(ctx context.Context, characterID uuid.UUID, maxDepth int) (map[string]interface{}, error)
	GetStrongestRelationships(ctx context.Context, characterID uuid.UUID, limit int) ([]*domain.RelationshipResponse, error)
	GetRelationshipsByType(ctx context.Context, characterID uuid.UUID, relationshipTypeID uuid.UUID) ([]*domain.RelationshipResponse, error)

	// 关系事件和变化
	CreateRelationshipEvent(ctx context.Context, relationshipID uuid.UUID, eventType, description string, changes map[string]float64) (*domain.RelationshipEvent, error)
	GetRelationshipEvents(ctx context.Context, relationshipID uuid.UUID, limit int) ([]*domain.RelationshipEvent, error)
	UpdateRelationshipFromInteraction(ctx context.Context, sourceID, targetID uuid.UUID, interactionType string, intensity float64) error

	// 触发规则管理
	CreateTriggerRule(ctx context.Context, userID uuid.UUID, req *domain.CreateTriggerRuleRequest) (*domain.TriggerRule, error)
	GetTriggerRule(ctx context.Context, id uuid.UUID) (*domain.TriggerRule, error)
	ListTriggerRules(ctx context.Context, groupChatID *uuid.UUID, ruleType *string, isActive *bool) ([]*domain.TriggerRule, error)
	EvaluateTriggers(ctx context.Context, groupChatID uuid.UUID, context map[string]interface{}) ([]domain.TriggerRule, error)

	// 关系推荐和建议
	SuggestRelationships(ctx context.Context, characterID uuid.UUID, limit int) ([]*domain.RelationshipType, error)
	AnalyzeRelationshipCompatibility(ctx context.Context, sourceID, targetID uuid.UUID) (float64, error)
}

type relationshipService struct {
	relationshipRepo repository.RelationshipRepository
	characterRepo    repository.CharacterRepository
	logger           *logrus.Logger
}

// NewRelationshipService 创建关系服务
func NewRelationshipService(
	relationshipRepo repository.RelationshipRepository,
	characterRepo repository.CharacterRepository,
	logger *logrus.Logger,
) RelationshipService {
	return &relationshipService{
		relationshipRepo: relationshipRepo,
		characterRepo:    characterRepo,
		logger:           logger,
	}
}

// CreateRelationshipType 创建关系类型
func (s *relationshipService) CreateRelationshipType(ctx context.Context, userID uuid.UUID, req *domain.CreateRelationshipTypeRequest) (*domain.RelationshipType, error) {
	relType := &domain.RelationshipType{
		ID:                  uuid.New(),
		UserID:              &userID,
		Name:                req.Name,
		DisplayName:         req.DisplayName,
		Description:         req.Description,
		Category:            domain.RelationshipCategoryCustom,
		IsMutual:            req.IsMutual,
		DefaultStrength:     req.DefaultStrength,
		DefaultTrust:        req.DefaultTrust,
		DefaultAffection:    req.DefaultAffection,
		DefaultRespect:      req.DefaultRespect,
		DefaultIntimacy:     req.DefaultIntimacy,
		DefaultTone:         req.DefaultTone,
		DefaultAddressStyle: req.DefaultAddressStyle,
		UsageCount:          0,
		IsFeatured:          false,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err := s.relationshipRepo.CreateRelationshipType(ctx, relType)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship type: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"relationship_type_id": relType.ID,
		"user_id":              userID,
		"name":                 relType.Name,
	}).Info("Relationship type created")

	return relType, nil
}

// GetRelationshipType 获取关系类型
func (s *relationshipService) GetRelationshipType(ctx context.Context, id uuid.UUID) (*domain.RelationshipType, error) {
	return s.relationshipRepo.GetRelationshipTypeByID(ctx, id)
}

// ListRelationshipTypes 获取关系类型列表
func (s *relationshipService) ListRelationshipTypes(ctx context.Context, userID *uuid.UUID, category *string, page, limit int) ([]*domain.RelationshipType, int, error) {
	return s.relationshipRepo.ListRelationshipTypes(ctx, userID, category, page, limit)
}

// CreateCharacterRelationship 创建角色关系
func (s *relationshipService) CreateCharacterRelationship(ctx context.Context, sourceCharacterID uuid.UUID, req *domain.CreateCharacterRelationshipRequest) (*domain.RelationshipResponse, error) {
	// 验证角色存在
	sourceChar, err := s.characterRepo.GetCharacterByID(ctx, sourceCharacterID)
	if err != nil {
		return nil, fmt.Errorf("source character not found: %w", err)
	}

	targetChar, err := s.characterRepo.GetCharacterByID(ctx, req.TargetCharacterID)
	if err != nil {
		return nil, fmt.Errorf("target character not found: %w", err)
	}

	// 检查关系是否已存在
	existing, _ := s.relationshipRepo.GetCharacterRelationship(ctx, sourceCharacterID, req.TargetCharacterID)
	if existing != nil {
		return nil, domain.NewAppError(domain.CodeConflict, "关系已存在", nil)
	}

	// 创建关系对象
	relationship := &domain.CharacterRelationshipEnhanced{
		ID:                        uuid.New(),
		SourceCharacterID:         sourceCharacterID,
		TargetCharacterID:         req.TargetCharacterID,
		RelationshipTypeID:        req.RelationshipTypeID,
		CustomTypeName:            req.CustomTypeName,
		Strength:                  req.Strength,
		Status:                    domain.RelationshipStatusActive,
		Trust:                     req.Trust,
		Affection:                 req.Affection,
		Respect:                   req.Respect,
		Intimacy:                  req.Intimacy,
		Jealousy:                  req.Jealousy,
		Dependency:                req.Dependency,
		AddressName:               req.AddressName,
		Tone:                      req.Tone,
		FormalityLevel:            req.FormalityLevel,
		SpeakingFrequency:         0.5, // 默认值
		InitiativeLevel:           0.5, // 默认值
		ConflictTendency:          0.2, // 默认值
		TriggerKeywords:           pq.StringArray(req.TriggerKeywords),
		TriggerEmotions:           pq.StringArray(req.TriggerEmotions),
		TriggerScenarios:          pq.StringArray(req.TriggerScenarios),
		RelationshipHistory:       json.RawMessage("[]"),
		SharedMemories:            pq.StringArray(req.SharedMemories),
		PrivateNotes:              req.PrivateNotes,
		ForbiddenTopics:           pq.StringArray(req.ForbiddenTopics),
		PreferredTopics:           pq.StringArray(req.PreferredTopics),
		InteractionLimits:         json.RawMessage("{}"),
		RelationshipEstablishedAt: time.Now(),
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = s.relationshipRepo.CreateCharacterRelationship(ctx, relationship)
	if err != nil {
		return nil, fmt.Errorf("failed to create character relationship: %w", err)
	}

	// 创建关系建立事件
	_, err = s.CreateRelationshipEvent(ctx, relationship.ID, "relationship_established", "关系建立", nil)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to create relationship event")
	}

	s.logger.WithFields(logrus.Fields{
		"relationship_id":     relationship.ID,
		"source_character_id": sourceCharacterID,
		"target_character_id": req.TargetCharacterID,
		"strength":            relationship.Strength,
	}).Info("Character relationship created")

	return s.buildRelationshipResponse(ctx, relationship, sourceChar, targetChar)
}

// GetCharacterRelationship 获取角色关系
func (s *relationshipService) GetCharacterRelationship(ctx context.Context, sourceID, targetID uuid.UUID) (*domain.RelationshipResponse, error) {
	relationship, err := s.relationshipRepo.GetCharacterRelationship(ctx, sourceID, targetID)
	if err != nil {
		return nil, err
	}

	sourceChar, _ := s.characterRepo.GetCharacterByID(ctx, sourceID)
	targetChar, _ := s.characterRepo.GetCharacterByID(ctx, targetID)

	return s.buildRelationshipResponse(ctx, relationship, sourceChar, targetChar)
}

// UpdateCharacterRelationship 更新角色关系
func (s *relationshipService) UpdateCharacterRelationship(ctx context.Context, relationshipID uuid.UUID, updates map[string]interface{}) (*domain.RelationshipResponse, error) {
	relationship, err := s.relationshipRepo.GetCharacterRelationshipByID(ctx, relationshipID)
	if err != nil {
		return nil, err
	}

	// 记录变化
	changes := make(map[string]float64)

	// 应用更新
	for key, value := range updates {
		switch key {
		case "strength":
			if v, ok := value.(float64); ok {
				changes["strength"] = v - relationship.Strength
				relationship.Strength = v
			}
		case "trust":
			if v, ok := value.(float64); ok {
				changes["trust"] = v - relationship.Trust
				relationship.Trust = v
			}
		case "affection":
			if v, ok := value.(float64); ok {
				changes["affection"] = v - relationship.Affection
				relationship.Affection = v
			}
		case "respect":
			if v, ok := value.(float64); ok {
				changes["respect"] = v - relationship.Respect
				relationship.Respect = v
			}
		case "intimacy":
			if v, ok := value.(float64); ok {
				changes["intimacy"] = v - relationship.Intimacy
				relationship.Intimacy = v
			}
		case "status":
			if v, ok := value.(string); ok {
				relationship.Status = v
			}
		}
	}

	relationship.LastInteractionAt = &time.Time{}
	*relationship.LastInteractionAt = time.Now()

	err = s.relationshipRepo.UpdateCharacterRelationship(ctx, relationship)
	if err != nil {
		return nil, fmt.Errorf("failed to update character relationship: %w", err)
	}

	// 创建变化事件
	if len(changes) > 0 {
		_, err = s.CreateRelationshipEvent(ctx, relationshipID, "relationship_change", "关系发生变化", changes)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to create relationship change event")
		}
	}

	sourceChar, _ := s.characterRepo.GetCharacterByID(ctx, relationship.SourceCharacterID)
	targetChar, _ := s.characterRepo.GetCharacterByID(ctx, relationship.TargetCharacterID)

	return s.buildRelationshipResponse(ctx, relationship, sourceChar, targetChar)
}

// DeleteCharacterRelationship 删除角色关系
func (s *relationshipService) DeleteCharacterRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	return s.relationshipRepo.DeleteCharacterRelationship(ctx, relationshipID)
}

// ListCharacterRelationships 获取角色关系列表
func (s *relationshipService) ListCharacterRelationships(ctx context.Context, characterID uuid.UUID, status *string) ([]*domain.RelationshipResponse, error) {
	relationships, err := s.relationshipRepo.ListCharacterRelationships(ctx, characterID, status)
	if err != nil {
		return nil, err
	}

	var responses []*domain.RelationshipResponse
	for _, rel := range relationships {
		sourceChar, _ := s.characterRepo.GetCharacterByID(ctx, rel.SourceCharacterID)
		targetChar, _ := s.characterRepo.GetCharacterByID(ctx, rel.TargetCharacterID)

		response, err := s.buildRelationshipResponse(ctx, rel, sourceChar, targetChar)
		if err != nil {
			s.logger.WithError(err).WithField("relationship_id", rel.ID).Warn("Failed to build relationship response")
			continue
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetRelationshipNetwork 获取关系网络
func (s *relationshipService) GetRelationshipNetwork(ctx context.Context, characterID uuid.UUID, maxDepth int) (map[string]interface{}, error) {
	return s.relationshipRepo.GetRelationshipNetwork(ctx, characterID, maxDepth)
}

// GetStrongestRelationships 获取最强关系
func (s *relationshipService) GetStrongestRelationships(ctx context.Context, characterID uuid.UUID, limit int) ([]*domain.RelationshipResponse, error) {
	relationships, err := s.relationshipRepo.GetStrongestRelationships(ctx, characterID, limit)
	if err != nil {
		return nil, err
	}

	var responses []*domain.RelationshipResponse
	for _, rel := range relationships {
		sourceChar, _ := s.characterRepo.GetCharacterByID(ctx, rel.SourceCharacterID)
		targetChar, _ := s.characterRepo.GetCharacterByID(ctx, rel.TargetCharacterID)

		response, err := s.buildRelationshipResponse(ctx, rel, sourceChar, targetChar)
		if err != nil {
			continue
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetRelationshipsByType 根据类型获取关系
func (s *relationshipService) GetRelationshipsByType(ctx context.Context, characterID uuid.UUID, relationshipTypeID uuid.UUID) ([]*domain.RelationshipResponse, error) {
	relationships, err := s.relationshipRepo.GetRelationshipsByType(ctx, characterID, relationshipTypeID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.RelationshipResponse
	for _, rel := range relationships {
		sourceChar, _ := s.characterRepo.GetCharacterByID(ctx, rel.SourceCharacterID)
		targetChar, _ := s.characterRepo.GetCharacterByID(ctx, rel.TargetCharacterID)

		response, err := s.buildRelationshipResponse(ctx, rel, sourceChar, targetChar)
		if err != nil {
			continue
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// CreateRelationshipEvent 创建关系事件
func (s *relationshipService) CreateRelationshipEvent(ctx context.Context, relationshipID uuid.UUID, eventType, description string, changes map[string]float64) (*domain.RelationshipEvent, error) {
	emotionChanges, _ := json.Marshal(changes)

	event := &domain.RelationshipEvent{
		ID:               uuid.New(),
		RelationshipID:   relationshipID,
		EventType:        eventType,
		EventDescription: description,
		EmotionChanges:   emotionChanges,
		Automatic:        true,
		CreatedAt:        time.Now(),
	}

	err := s.relationshipRepo.CreateRelationshipEvent(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship event: %w", err)
	}

	return event, nil
}

// GetRelationshipEvents 获取关系事件
func (s *relationshipService) GetRelationshipEvents(ctx context.Context, relationshipID uuid.UUID, limit int) ([]*domain.RelationshipEvent, error) {
	return s.relationshipRepo.GetRelationshipEvents(ctx, relationshipID, limit)
}

// UpdateRelationshipFromInteraction 根据互动更新关系
func (s *relationshipService) UpdateRelationshipFromInteraction(ctx context.Context, sourceID, targetID uuid.UUID, interactionType string, intensity float64) error {
	relationship, err := s.relationshipRepo.GetCharacterRelationship(ctx, sourceID, targetID)
	if err != nil {
		// 如果关系不存在，可以选择创建默认关系
		return nil
	}

	// 根据互动类型调整关系参数
	updates := make(map[string]interface{})

	switch interactionType {
	case "positive":
		updates["affection"] = relationship.Affection + intensity*0.1
		updates["trust"] = relationship.Trust + intensity*0.05
	case "negative":
		updates["affection"] = relationship.Affection - intensity*0.1
		updates["trust"] = relationship.Trust - intensity*0.05
	case "conflict":
		updates["respect"] = relationship.Respect - intensity*0.1
		updates["trust"] = relationship.Trust - intensity*0.15
	case "intimate":
		updates["intimacy"] = relationship.Intimacy + intensity*0.1
		updates["affection"] = relationship.Affection + intensity*0.05
	}

	// 确保值在0-1范围内
	for key, value := range updates {
		if v, ok := value.(float64); ok {
			if v < 0 {
				updates[key] = 0.0
			} else if v > 1 {
				updates[key] = 1.0
			}
		}
	}

	_, err = s.UpdateCharacterRelationship(ctx, relationship.ID, updates)
	return err
}

// CreateTriggerRule 创建触发规则
func (s *relationshipService) CreateTriggerRule(ctx context.Context, userID uuid.UUID, req *domain.CreateTriggerRuleRequest) (*domain.TriggerRule, error) {
	rule := &domain.TriggerRule{
		ID:                uuid.New(),
		UserID:            &userID,
		GroupChatID:       req.GroupChatID,
		Name:              req.Name,
		Description:       req.Description,
		RuleType:          req.RuleType,
		TriggerConditions: req.TriggerConditions,
		Actions:           req.Actions,
		CooldownMinutes:   req.CooldownMinutes,
		MaxTriggersPerDay: req.MaxTriggersPerDay,
		Priority:          req.Priority,
		IsActive:          true,
		TriggerCount:      0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := s.relationshipRepo.CreateTriggerRule(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger rule: %w", err)
	}

	return rule, nil
}

// GetTriggerRule 获取触发规则
func (s *relationshipService) GetTriggerRule(ctx context.Context, id uuid.UUID) (*domain.TriggerRule, error) {
	return s.relationshipRepo.GetTriggerRuleByID(ctx, id)
}

// ListTriggerRules 获取触发规则列表
func (s *relationshipService) ListTriggerRules(ctx context.Context, groupChatID *uuid.UUID, ruleType *string, isActive *bool) ([]*domain.TriggerRule, error) {
	return s.relationshipRepo.ListTriggerRules(ctx, groupChatID, ruleType, isActive)
}

// EvaluateTriggers 评估触发器
func (s *relationshipService) EvaluateTriggers(ctx context.Context, groupChatID uuid.UUID, context map[string]interface{}) ([]domain.TriggerRule, error) {
	// 获取活跃的触发规则
	isActive := true
	rules, err := s.relationshipRepo.ListTriggerRules(ctx, &groupChatID, nil, &isActive)
	if err != nil {
		return nil, err
	}

	var triggeredRules []domain.TriggerRule

	for _, rule := range rules {
		// 简化的触发条件评估
		if s.evaluateTriggerCondition(rule, context) {
			triggeredRules = append(triggeredRules, *rule)
		}
	}

	return triggeredRules, nil
}

// SuggestRelationships 建议关系
func (s *relationshipService) SuggestRelationships(ctx context.Context, characterID uuid.UUID, limit int) ([]*domain.RelationshipType, error) {
	// 获取系统关系类型
	category := domain.RelationshipCategorySystem
	relTypes, _, err := s.relationshipRepo.ListRelationshipTypes(ctx, nil, &category, 1, limit)
	if err != nil {
		return nil, err
	}

	return relTypes, nil
}

// AnalyzeRelationshipCompatibility 分析关系兼容性
func (s *relationshipService) AnalyzeRelationshipCompatibility(ctx context.Context, sourceID, targetID uuid.UUID) (float64, error) {
	// 简化的兼容性分析
	// 实际实现可以基于角色属性、现有关系等进行复杂分析
	return 0.7, nil
}

// 辅助方法

// buildRelationshipResponse 构建关系响应
func (s *relationshipService) buildRelationshipResponse(ctx context.Context, relationship *domain.CharacterRelationshipEnhanced, sourceChar, targetChar *domain.Character) (*domain.RelationshipResponse, error) {
	response := &domain.RelationshipResponse{
		CharacterRelationshipEnhanced: relationship,
		SourceCharacter:               sourceChar,
		TargetCharacter:               targetChar,
		CanEdit:                       true, // 简化实现
	}

	// 获取关系类型
	if relationship.RelationshipTypeID != nil {
		relType, err := s.relationshipRepo.GetRelationshipTypeByID(ctx, *relationship.RelationshipTypeID)
		if err == nil {
			response.RelationshipType = relType
		}
	}

	// 获取最近事件
	events, err := s.relationshipRepo.GetRelationshipEvents(ctx, relationship.ID, 5)
	if err == nil {
		// 转换为正确的类型
		var eventList []domain.RelationshipEvent
		for _, e := range events {
			eventList = append(eventList, *e)
		}
		response.RecentEvents = eventList
	}

	return response, nil
}

// evaluateTriggerCondition 评估触发条件
func (s *relationshipService) evaluateTriggerCondition(rule *domain.TriggerRule, context map[string]interface{}) bool {
	// 简化的条件评估实现
	// 实际实现应该解析JSON条件并进行复杂匹配

	// 检查冷却时间
	if rule.LastTriggeredAt != nil {
		cooldownDuration := time.Duration(rule.CooldownMinutes) * time.Minute
		if time.Since(*rule.LastTriggeredAt) < cooldownDuration {
			return false
		}
	}

	// 检查每日触发次数限制
	// 这里需要更复杂的逻辑来跟踪每日触发次数

	return true // 简化实现，总是返回true
}
