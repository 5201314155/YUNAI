package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// RelationshipRepository 关系仓库接口
type RelationshipRepository interface {
	// 关系类型管理
	CreateRelationshipType(ctx context.Context, relType *domain.RelationshipType) error
	GetRelationshipTypeByID(ctx context.Context, id uuid.UUID) (*domain.RelationshipType, error)
	UpdateRelationshipType(ctx context.Context, relType *domain.RelationshipType) error
	DeleteRelationshipType(ctx context.Context, id uuid.UUID) error
	ListRelationshipTypes(ctx context.Context, userID *uuid.UUID, category *string, page, limit int) ([]*domain.RelationshipType, int, error)

	// 角色关系管理
	CreateCharacterRelationship(ctx context.Context, relationship *domain.CharacterRelationshipEnhanced) error
	GetCharacterRelationshipByID(ctx context.Context, id uuid.UUID) (*domain.CharacterRelationshipEnhanced, error)
	GetCharacterRelationship(ctx context.Context, sourceID, targetID uuid.UUID) (*domain.CharacterRelationshipEnhanced, error)
	UpdateCharacterRelationship(ctx context.Context, relationship *domain.CharacterRelationshipEnhanced) error
	DeleteCharacterRelationship(ctx context.Context, id uuid.UUID) error
	ListCharacterRelationships(ctx context.Context, characterID uuid.UUID, status *string) ([]*domain.CharacterRelationshipEnhanced, error)
	GetUserRelationships(ctx context.Context, userID string) ([]domain.Relationship, error)

	// 关系事件管理
	CreateRelationshipEvent(ctx context.Context, event *domain.RelationshipEvent) error
	GetRelationshipEvents(ctx context.Context, relationshipID uuid.UUID, limit int) ([]*domain.RelationshipEvent, error)

	// 触发规则管理
	CreateTriggerRule(ctx context.Context, rule *domain.TriggerRule) error
	GetTriggerRuleByID(ctx context.Context, id uuid.UUID) (*domain.TriggerRule, error)
	UpdateTriggerRule(ctx context.Context, rule *domain.TriggerRule) error
	DeleteTriggerRule(ctx context.Context, id uuid.UUID) error
	ListTriggerRules(ctx context.Context, groupChatID *uuid.UUID, ruleType *string, isActive *bool) ([]*domain.TriggerRule, error)

	// 关系查询和分析
	GetRelationshipNetwork(ctx context.Context, characterID uuid.UUID, maxDepth int) (map[string]interface{}, error)
	GetStrongestRelationships(ctx context.Context, characterID uuid.UUID, limit int) ([]*domain.CharacterRelationshipEnhanced, error)
	GetRelationshipsByType(ctx context.Context, characterID uuid.UUID, relationshipTypeID uuid.UUID) ([]*domain.CharacterRelationshipEnhanced, error)
}

type relationshipRepository struct {
	db *sqlx.DB
}

// NewRelationshipRepository 创建关系仓库
func NewRelationshipRepository(db *sqlx.DB) RelationshipRepository {
	return &relationshipRepository{db: db}
}

// CreateRelationshipType 创建关系类型
func (r *relationshipRepository) CreateRelationshipType(ctx context.Context, relType *domain.RelationshipType) error {
	query := `
		INSERT INTO relationship_types (
			id, user_id, name, display_name, description, category, is_mutual,
			default_strength, default_trust, default_affection, default_respect,
			default_intimacy, default_tone, default_address_style, usage_count, is_featured
		) VALUES (
			:id, :user_id, :name, :display_name, :description, :category, :is_mutual,
			:default_strength, :default_trust, :default_affection, :default_respect,
			:default_intimacy, :default_tone, :default_address_style, :usage_count, :is_featured
		)`

	_, err := r.db.NamedExecContext(ctx, query, relType)
	return err
}

// GetRelationshipTypeByID 根据ID获取关系类型
func (r *relationshipRepository) GetRelationshipTypeByID(ctx context.Context, id uuid.UUID) (*domain.RelationshipType, error) {
	query := `
		SELECT id, user_id, name, display_name, description, category, is_mutual,
			   default_strength, default_trust, default_affection, default_respect,
			   default_intimacy, default_tone, default_address_style, usage_count,
			   is_featured, created_at, updated_at
		FROM relationship_types WHERE id = $1`

	var relType domain.RelationshipType
	err := r.db.GetContext(ctx, &relType, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "关系类型不存在", err)
		}
		return nil, err
	}

	return &relType, nil
}

// UpdateRelationshipType 更新关系类型
func (r *relationshipRepository) UpdateRelationshipType(ctx context.Context, relType *domain.RelationshipType) error {
	query := `
		UPDATE relationship_types SET
			name = :name, display_name = :display_name, description = :description,
			is_mutual = :is_mutual, default_strength = :default_strength,
			default_trust = :default_trust, default_affection = :default_affection,
			default_respect = :default_respect, default_intimacy = :default_intimacy,
			default_tone = :default_tone, default_address_style = :default_address_style,
			usage_count = :usage_count, is_featured = :is_featured, updated_at = :updated_at
		WHERE id = :id`

	relType.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, relType)
	return err
}

// DeleteRelationshipType 删除关系类型
func (r *relationshipRepository) DeleteRelationshipType(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM relationship_types WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListRelationshipTypes 获取关系类型列表
func (r *relationshipRepository) ListRelationshipTypes(ctx context.Context, userID *uuid.UUID, category *string, page, limit int) ([]*domain.RelationshipType, int, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("(user_id = $%d OR user_id IS NULL)", argIndex))
		args = append(args, *userID)
		argIndex++
	} else {
		conditions = append(conditions, "user_id IS NULL")
	}

	if category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *category)
		argIndex++
	}

	// 构建WHERE子句
	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM relationship_types %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, name, display_name, description, category, is_mutual,
			   default_strength, default_trust, default_affection, default_respect,
			   default_intimacy, default_tone, default_address_style, usage_count,
			   is_featured, created_at, updated_at
		FROM relationship_types %s
		ORDER BY is_featured DESC, usage_count DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	var relTypes []*domain.RelationshipType
	err = r.db.SelectContext(ctx, &relTypes, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return relTypes, total, nil
}

// CreateCharacterRelationship 创建角色关系
func (r *relationshipRepository) CreateCharacterRelationship(ctx context.Context, relationship *domain.CharacterRelationshipEnhanced) error {
	query := `
		INSERT INTO character_relationships_enhanced (
			id, source_character_id, target_character_id, relationship_type_id, custom_type_name,
			strength, status, trust, affection, respect, intimacy, jealousy, dependency,
			address_name, tone, formality_level, speaking_frequency, initiative_level,
			conflict_tendency, trigger_keywords, trigger_emotions, trigger_scenarios,
			relationship_history, shared_memories, private_notes, forbidden_topics,
			preferred_topics, interaction_limits, relationship_established_at
		) VALUES (
			:id, :source_character_id, :target_character_id, :relationship_type_id, :custom_type_name,
			:strength, :status, :trust, :affection, :respect, :intimacy, :jealousy, :dependency,
			:address_name, :tone, :formality_level, :speaking_frequency, :initiative_level,
			:conflict_tendency, :trigger_keywords, :trigger_emotions, :trigger_scenarios,
			:relationship_history, :shared_memories, :private_notes, :forbidden_topics,
			:preferred_topics, :interaction_limits, :relationship_established_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, relationship)
	return err
}

// GetCharacterRelationshipByID 根据ID获取角色关系
func (r *relationshipRepository) GetCharacterRelationshipByID(ctx context.Context, id uuid.UUID) (*domain.CharacterRelationshipEnhanced, error) {
	query := `
		SELECT id, source_character_id, target_character_id, relationship_type_id, custom_type_name,
			   strength, status, trust, affection, respect, intimacy, jealousy, dependency,
			   address_name, tone, formality_level, speaking_frequency, initiative_level,
			   conflict_tendency, trigger_keywords, trigger_emotions, trigger_scenarios,
			   relationship_history, shared_memories, private_notes, forbidden_topics,
			   preferred_topics, interaction_limits, last_interaction_at,
			   relationship_established_at, created_at, updated_at
		FROM character_relationships_enhanced WHERE id = $1`

	var relationship domain.CharacterRelationshipEnhanced
	err := r.db.GetContext(ctx, &relationship, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "角色关系不存在", err)
		}
		return nil, err
	}

	return &relationship, nil
}

// GetCharacterRelationship 获取两个角色之间的关系
func (r *relationshipRepository) GetCharacterRelationship(ctx context.Context, sourceID, targetID uuid.UUID) (*domain.CharacterRelationshipEnhanced, error) {
	query := `
		SELECT id, source_character_id, target_character_id, relationship_type_id, custom_type_name,
			   strength, status, trust, affection, respect, intimacy, jealousy, dependency,
			   address_name, tone, formality_level, speaking_frequency, initiative_level,
			   conflict_tendency, trigger_keywords, trigger_emotions, trigger_scenarios,
			   relationship_history, shared_memories, private_notes, forbidden_topics,
			   preferred_topics, interaction_limits, last_interaction_at,
			   relationship_established_at, created_at, updated_at
		FROM character_relationships_enhanced 
		WHERE source_character_id = $1 AND target_character_id = $2`

	var relationship domain.CharacterRelationshipEnhanced
	err := r.db.GetContext(ctx, &relationship, query, sourceID, targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "角色关系不存在", err)
		}
		return nil, err
	}

	return &relationship, nil
}

// UpdateCharacterRelationship 更新角色关系
func (r *relationshipRepository) UpdateCharacterRelationship(ctx context.Context, relationship *domain.CharacterRelationshipEnhanced) error {
	query := `
		UPDATE character_relationships_enhanced SET
			relationship_type_id = :relationship_type_id, custom_type_name = :custom_type_name,
			strength = :strength, status = :status, trust = :trust, affection = :affection,
			respect = :respect, intimacy = :intimacy, jealousy = :jealousy, dependency = :dependency,
			address_name = :address_name, tone = :tone, formality_level = :formality_level,
			speaking_frequency = :speaking_frequency, initiative_level = :initiative_level,
			conflict_tendency = :conflict_tendency, trigger_keywords = :trigger_keywords,
			trigger_emotions = :trigger_emotions, trigger_scenarios = :trigger_scenarios,
			relationship_history = :relationship_history, shared_memories = :shared_memories,
			private_notes = :private_notes, forbidden_topics = :forbidden_topics,
			preferred_topics = :preferred_topics, interaction_limits = :interaction_limits,
			last_interaction_at = :last_interaction_at, updated_at = :updated_at
		WHERE id = :id`

	relationship.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, relationship)
	return err
}

// DeleteCharacterRelationship 删除角色关系
func (r *relationshipRepository) DeleteCharacterRelationship(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM character_relationships_enhanced WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListCharacterRelationships 获取角色的所有关系
func (r *relationshipRepository) ListCharacterRelationships(ctx context.Context, characterID uuid.UUID, status *string) ([]*domain.CharacterRelationshipEnhanced, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, fmt.Sprintf("source_character_id = $%d", argIndex))
	args = append(args, characterID)
	argIndex++

	if status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, source_character_id, target_character_id, relationship_type_id, custom_type_name,
			   strength, status, trust, affection, respect, intimacy, jealousy, dependency,
			   address_name, tone, formality_level, speaking_frequency, initiative_level,
			   conflict_tendency, trigger_keywords, trigger_emotions, trigger_scenarios,
			   relationship_history, shared_memories, private_notes, forbidden_topics,
			   preferred_topics, interaction_limits, last_interaction_at,
			   relationship_established_at, created_at, updated_at
		FROM character_relationships_enhanced %s
		ORDER BY strength DESC, created_at DESC
	`, whereClause)

	var relationships []*domain.CharacterRelationshipEnhanced
	err := r.db.SelectContext(ctx, &relationships, query, args...)
	return relationships, err
}

// GetUserRelationships 获取用户的关系网络
func (r *relationshipRepository) GetUserRelationships(ctx context.Context, userID string) ([]domain.Relationship, error) {
	// 简化实现：返回空的关系列表
	// 在实际实现中，这里应该查询用户与各个角色的关系
	return []domain.Relationship{}, nil
}

// CreateRelationshipEvent 创建关系事件
func (r *relationshipRepository) CreateRelationshipEvent(ctx context.Context, event *domain.RelationshipEvent) error {
	query := `
		INSERT INTO relationship_events (
			id, relationship_id, event_type, event_description, strength_change,
			emotion_changes, triggered_by_message_id, triggered_by_scenario,
			trigger_keywords, automatic, created_by_user_id
		) VALUES (
			:id, :relationship_id, :event_type, :event_description, :strength_change,
			:emotion_changes, :triggered_by_message_id, :triggered_by_scenario,
			:trigger_keywords, :automatic, :created_by_user_id
		)`

	_, err := r.db.NamedExecContext(ctx, query, event)
	return err
}

// GetRelationshipEvents 获取关系事件
func (r *relationshipRepository) GetRelationshipEvents(ctx context.Context, relationshipID uuid.UUID, limit int) ([]*domain.RelationshipEvent, error) {
	query := `
		SELECT id, relationship_id, event_type, event_description, strength_change,
			   emotion_changes, triggered_by_message_id, triggered_by_scenario,
			   trigger_keywords, automatic, created_by_user_id, created_at
		FROM relationship_events
		WHERE relationship_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	var events []*domain.RelationshipEvent
	err := r.db.SelectContext(ctx, &events, query, relationshipID, limit)
	return events, err
}

// CreateTriggerRule 创建触发规则
func (r *relationshipRepository) CreateTriggerRule(ctx context.Context, rule *domain.TriggerRule) error {
	query := `
		INSERT INTO trigger_rules (
			id, user_id, group_chat_id, name, description, rule_type,
			trigger_conditions, actions, cooldown_minutes, max_triggers_per_day,
			priority, is_active, trigger_count
		) VALUES (
			:id, :user_id, :group_chat_id, :name, :description, :rule_type,
			:trigger_conditions, :actions, :cooldown_minutes, :max_triggers_per_day,
			:priority, :is_active, :trigger_count
		)`

	_, err := r.db.NamedExecContext(ctx, query, rule)
	return err
}

// GetTriggerRuleByID 根据ID获取触发规则
func (r *relationshipRepository) GetTriggerRuleByID(ctx context.Context, id uuid.UUID) (*domain.TriggerRule, error) {
	query := `
		SELECT id, user_id, group_chat_id, name, description, rule_type,
			   trigger_conditions, actions, cooldown_minutes, max_triggers_per_day,
			   priority, is_active, last_triggered_at, trigger_count, created_at, updated_at
		FROM trigger_rules WHERE id = $1`

	var rule domain.TriggerRule
	err := r.db.GetContext(ctx, &rule, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "触发规则不存在", err)
		}
		return nil, err
	}

	return &rule, nil
}

// UpdateTriggerRule 更新触发规则
func (r *relationshipRepository) UpdateTriggerRule(ctx context.Context, rule *domain.TriggerRule) error {
	query := `
		UPDATE trigger_rules SET
			name = :name, description = :description, trigger_conditions = :trigger_conditions,
			actions = :actions, cooldown_minutes = :cooldown_minutes,
			max_triggers_per_day = :max_triggers_per_day, priority = :priority,
			is_active = :is_active, last_triggered_at = :last_triggered_at,
			trigger_count = :trigger_count, updated_at = :updated_at
		WHERE id = :id`

	rule.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, rule)
	return err
}

// DeleteTriggerRule 删除触发规则
func (r *relationshipRepository) DeleteTriggerRule(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM trigger_rules WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListTriggerRules 获取触发规则列表
func (r *relationshipRepository) ListTriggerRules(ctx context.Context, groupChatID *uuid.UUID, ruleType *string, isActive *bool) ([]*domain.TriggerRule, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if groupChatID != nil {
		conditions = append(conditions, fmt.Sprintf("group_chat_id = $%d", argIndex))
		args = append(args, *groupChatID)
		argIndex++
	}

	if ruleType != nil {
		conditions = append(conditions, fmt.Sprintf("rule_type = $%d", argIndex))
		args = append(args, *ruleType)
		argIndex++
	}

	if isActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *isActive)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, group_chat_id, name, description, rule_type,
			   trigger_conditions, actions, cooldown_minutes, max_triggers_per_day,
			   priority, is_active, last_triggered_at, trigger_count, created_at, updated_at
		FROM trigger_rules %s
		ORDER BY priority DESC, created_at DESC
	`, whereClause)

	var rules []*domain.TriggerRule
	err := r.db.SelectContext(ctx, &rules, query, args...)
	return rules, err
}

// GetRelationshipNetwork 获取关系网络
func (r *relationshipRepository) GetRelationshipNetwork(ctx context.Context, characterID uuid.UUID, maxDepth int) (map[string]interface{}, error) {
	// 简化实现：获取直接关系
	relationships, err := r.ListCharacterRelationships(ctx, characterID, nil)
	if err != nil {
		return nil, err
	}

	network := map[string]interface{}{
		"center_character_id": characterID,
		"relationships":       relationships,
		"depth":               1, // 当前只实现深度1
	}

	return network, nil
}

// GetStrongestRelationships 获取最强的关系
func (r *relationshipRepository) GetStrongestRelationships(ctx context.Context, characterID uuid.UUID, limit int) ([]*domain.CharacterRelationshipEnhanced, error) {
	query := `
		SELECT id, source_character_id, target_character_id, relationship_type_id, custom_type_name,
			   strength, status, trust, affection, respect, intimacy, jealousy, dependency,
			   address_name, tone, formality_level, speaking_frequency, initiative_level,
			   conflict_tendency, trigger_keywords, trigger_emotions, trigger_scenarios,
			   relationship_history, shared_memories, private_notes, forbidden_topics,
			   preferred_topics, interaction_limits, last_interaction_at,
			   relationship_established_at, created_at, updated_at
		FROM character_relationships_enhanced
		WHERE source_character_id = $1 AND status = 'active'
		ORDER BY strength DESC
		LIMIT $2`

	var relationships []*domain.CharacterRelationshipEnhanced
	err := r.db.SelectContext(ctx, &relationships, query, characterID, limit)
	return relationships, err
}

// GetRelationshipsByType 根据类型获取关系
func (r *relationshipRepository) GetRelationshipsByType(ctx context.Context, characterID uuid.UUID, relationshipTypeID uuid.UUID) ([]*domain.CharacterRelationshipEnhanced, error) {
	query := `
		SELECT id, source_character_id, target_character_id, relationship_type_id, custom_type_name,
			   strength, status, trust, affection, respect, intimacy, jealousy, dependency,
			   address_name, tone, formality_level, speaking_frequency, initiative_level,
			   conflict_tendency, trigger_keywords, trigger_emotions, trigger_scenarios,
			   relationship_history, shared_memories, private_notes, forbidden_topics,
			   preferred_topics, interaction_limits, last_interaction_at,
			   relationship_established_at, created_at, updated_at
		FROM character_relationships_enhanced
		WHERE source_character_id = $1 AND relationship_type_id = $2 AND status = 'active'
		ORDER BY strength DESC`

	var relationships []*domain.CharacterRelationshipEnhanced
	err := r.db.SelectContext(ctx, &relationships, query, characterID, relationshipTypeID)
	return relationships, err
}
