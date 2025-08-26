package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// MemoryRepository 记忆仓库接口
type MemoryRepository interface {
	// 记忆片段管理
	CreateMemoryFragment(ctx context.Context, memory *domain.MemoryFragment) error
	GetMemoryFragmentByID(ctx context.Context, id uuid.UUID) (*domain.MemoryFragment, error)
	UpdateMemoryFragment(ctx context.Context, memory *domain.MemoryFragment) error
	DeleteMemoryFragment(ctx context.Context, id uuid.UUID) error
	ListMemoryFragments(ctx context.Context, characterID *uuid.UUID, memoryTypes []string, page, limit int) ([]*domain.MemoryFragment, int, error)

	// 记忆搜索
	SearchMemoriesByContent(ctx context.Context, characterID *uuid.UUID, query string, memoryTypes []string, limit int) ([]*domain.MemoryFragment, error)
	SearchMemoriesByEmbedding(ctx context.Context, characterID *uuid.UUID, embedding json.RawMessage, memoryTypes []string, limit int, minScore float64) ([]*domain.MemoryFragment, error)
	GetRecentMemories(ctx context.Context, characterID uuid.UUID, hours int, limit int) ([]*domain.MemoryFragment, error)
	GetImportantMemories(ctx context.Context, characterID uuid.UUID, minImportance float64, limit int) ([]*domain.MemoryFragment, error)

	// 记忆访问跟踪
	IncrementMemoryAccess(ctx context.Context, memoryID uuid.UUID) error
	CleanupExpiredMemories(ctx context.Context) (int, error)

	// 对话上下文管理
	CreateConversationContext(ctx context.Context, context *domain.ConversationContext) error
	GetConversationContextByGroupID(ctx context.Context, groupChatID uuid.UUID) (*domain.ConversationContext, error)
	UpdateConversationContext(ctx context.Context, context *domain.ConversationContext) error
	DeleteConversationContext(ctx context.Context, groupChatID uuid.UUID) error

	// 记忆统计
	GetMemoryStats(ctx context.Context, characterID uuid.UUID) (map[string]interface{}, error)
}

type memoryRepository struct {
	db *sqlx.DB
}

// NewMemoryRepository 创建记忆仓库
func NewMemoryRepository(db *sqlx.DB) MemoryRepository {
	return &memoryRepository{db: db}
}

// CreateMemoryFragment 创建记忆片段
func (r *memoryRepository) CreateMemoryFragment(ctx context.Context, memory *domain.MemoryFragment) error {
	query := `
		INSERT INTO memory_fragments (
			id, character_id, content, summary, memory_type, importance_score,
			emotional_intensity, related_characters, related_topics, related_emotions,
			source_message_id, source_group_chat_id, context_metadata, embedding,
			access_count, expires_at
		) VALUES (
			:id, :character_id, :content, :summary, :memory_type, :importance_score,
			:emotional_intensity, :related_characters, :related_topics, :related_emotions,
			:source_message_id, :source_group_chat_id, :context_metadata, :embedding,
			:access_count, :expires_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, memory)
	return err
}

// GetMemoryFragmentByID 根据ID获取记忆片段
func (r *memoryRepository) GetMemoryFragmentByID(ctx context.Context, id uuid.UUID) (*domain.MemoryFragment, error) {
	query := `
		SELECT id, character_id, content, summary, memory_type, importance_score,
			   emotional_intensity, related_characters, related_topics, related_emotions,
			   source_message_id, source_group_chat_id, context_metadata, embedding,
			   access_count, last_accessed_at, expires_at, created_at, updated_at
		FROM memory_fragments WHERE id = $1`

	var memory domain.MemoryFragment
	err := r.db.GetContext(ctx, &memory, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "记忆片段不存在", err)
		}
		return nil, err
	}

	return &memory, nil
}

// UpdateMemoryFragment 更新记忆片段
func (r *memoryRepository) UpdateMemoryFragment(ctx context.Context, memory *domain.MemoryFragment) error {
	query := `
		UPDATE memory_fragments SET
			content = :content, summary = :summary, importance_score = :importance_score,
			emotional_intensity = :emotional_intensity, related_characters = :related_characters,
			related_topics = :related_topics, related_emotions = :related_emotions,
			context_metadata = :context_metadata, access_count = :access_count,
			last_accessed_at = :last_accessed_at, expires_at = :expires_at,
			updated_at = :updated_at
		WHERE id = :id`

	memory.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, memory)
	return err
}

// DeleteMemoryFragment 删除记忆片段
func (r *memoryRepository) DeleteMemoryFragment(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM memory_fragments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListMemoryFragments 获取记忆片段列表
func (r *memoryRepository) ListMemoryFragments(ctx context.Context, characterID *uuid.UUID, memoryTypes []string, page, limit int) ([]*domain.MemoryFragment, int, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if characterID != nil {
		conditions = append(conditions, fmt.Sprintf("character_id = $%d", argIndex))
		args = append(args, *characterID)
		argIndex++
	}

	if len(memoryTypes) > 0 {
		placeholders := make([]string, len(memoryTypes))
		for i, memoryType := range memoryTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, memoryType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("memory_type IN (%s)", strings.Join(placeholders, ",")))
	}

	// 过滤未过期的记忆
	conditions = append(conditions, "(expires_at IS NULL OR expires_at > NOW())")

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM memory_fragments %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`
		SELECT id, character_id, content, summary, memory_type, importance_score,
			   emotional_intensity, related_characters, related_topics, related_emotions,
			   source_message_id, source_group_chat_id, context_metadata, embedding,
			   access_count, last_accessed_at, expires_at, created_at, updated_at
		FROM memory_fragments %s
		ORDER BY importance_score DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	var memories []*domain.MemoryFragment
	err = r.db.SelectContext(ctx, &memories, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return memories, total, nil
}

// SearchMemoriesByContent 根据内容搜索记忆
func (r *memoryRepository) SearchMemoriesByContent(ctx context.Context, characterID *uuid.UUID, query string, memoryTypes []string, limit int) ([]*domain.MemoryFragment, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if characterID != nil {
		conditions = append(conditions, fmt.Sprintf("character_id = $%d", argIndex))
		args = append(args, *characterID)
		argIndex++
	}

	if len(memoryTypes) > 0 {
		placeholders := make([]string, len(memoryTypes))
		for i, memoryType := range memoryTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, memoryType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("memory_type IN (%s)", strings.Join(placeholders, ",")))
	}

	// 全文搜索条件
	conditions = append(conditions, fmt.Sprintf("(content ILIKE $%d OR summary ILIKE $%d)", argIndex, argIndex))
	searchTerm := "%" + query + "%"
	args = append(args, searchTerm)
	argIndex++

	// 过滤未过期的记忆
	conditions = append(conditions, "(expires_at IS NULL OR expires_at > NOW())")

	// 构建WHERE子句
	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	dataQuery := fmt.Sprintf(`
		SELECT id, character_id, content, summary, memory_type, importance_score,
			   emotional_intensity, related_characters, related_topics, related_emotions,
			   source_message_id, source_group_chat_id, context_metadata, embedding,
			   access_count, last_accessed_at, expires_at, created_at, updated_at
		FROM memory_fragments %s
		ORDER BY importance_score DESC, created_at DESC
		LIMIT $%d
	`, whereClause, argIndex)

	args = append(args, limit)

	var memories []*domain.MemoryFragment
	err := r.db.SelectContext(ctx, &memories, dataQuery, args...)
	return memories, err
}

// SearchMemoriesByEmbedding 根据向量嵌入搜索记忆
func (r *memoryRepository) SearchMemoriesByEmbedding(ctx context.Context, characterID *uuid.UUID, embedding json.RawMessage, memoryTypes []string, limit int, minScore float64) ([]*domain.MemoryFragment, error) {
	// 注意：这里需要使用向量数据库扩展，如pgvector
	// 目前使用简化实现，实际部署时需要安装pgvector扩展

	var conditions []string
	var args []interface{}
	argIndex := 1

	if characterID != nil {
		conditions = append(conditions, fmt.Sprintf("character_id = $%d", argIndex))
		args = append(args, *characterID)
		argIndex++
	}

	if len(memoryTypes) > 0 {
		placeholders := make([]string, len(memoryTypes))
		for i, memoryType := range memoryTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, memoryType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("memory_type IN (%s)", strings.Join(placeholders, ",")))
	}

	// 过滤未过期的记忆
	conditions = append(conditions, "(expires_at IS NULL OR expires_at > NOW())")

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 简化实现：按重要性和时间排序
	// 实际实现应该使用: ORDER BY embedding <-> $embedding LIMIT $limit
	dataQuery := fmt.Sprintf(`
		SELECT id, character_id, content, summary, memory_type, importance_score,
			   emotional_intensity, related_characters, related_topics, related_emotions,
			   source_message_id, source_group_chat_id, context_metadata, embedding,
			   access_count, last_accessed_at, expires_at, created_at, updated_at
		FROM memory_fragments %s
		ORDER BY importance_score DESC, created_at DESC
		LIMIT $%d
	`, whereClause, argIndex)

	args = append(args, limit)

	var memories []*domain.MemoryFragment
	err := r.db.SelectContext(ctx, &memories, dataQuery, args...)
	return memories, err
}

// GetRecentMemories 获取最近的记忆
func (r *memoryRepository) GetRecentMemories(ctx context.Context, characterID uuid.UUID, hours int, limit int) ([]*domain.MemoryFragment, error) {
	query := `
		SELECT id, character_id, content, summary, memory_type, importance_score,
			   emotional_intensity, related_characters, related_topics, related_emotions,
			   source_message_id, source_group_chat_id, context_metadata, embedding,
			   access_count, last_accessed_at, expires_at, created_at, updated_at
		FROM memory_fragments
		WHERE character_id = $1 
		  AND created_at > NOW() - INTERVAL '%d hours'
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
		LIMIT $2`

	query = fmt.Sprintf(query, hours)

	var memories []*domain.MemoryFragment
	err := r.db.SelectContext(ctx, &memories, query, characterID, limit)
	return memories, err
}

// GetImportantMemories 获取重要记忆
func (r *memoryRepository) GetImportantMemories(ctx context.Context, characterID uuid.UUID, minImportance float64, limit int) ([]*domain.MemoryFragment, error) {
	query := `
		SELECT id, character_id, content, summary, memory_type, importance_score,
			   emotional_intensity, related_characters, related_topics, related_emotions,
			   source_message_id, source_group_chat_id, context_metadata, embedding,
			   access_count, last_accessed_at, expires_at, created_at, updated_at
		FROM memory_fragments
		WHERE character_id = $1 
		  AND importance_score >= $2
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY importance_score DESC, access_count DESC
		LIMIT $3`

	var memories []*domain.MemoryFragment
	err := r.db.SelectContext(ctx, &memories, query, characterID, minImportance, limit)
	return memories, err
}

// IncrementMemoryAccess 增加记忆访问次数
func (r *memoryRepository) IncrementMemoryAccess(ctx context.Context, memoryID uuid.UUID) error {
	query := `
		UPDATE memory_fragments 
		SET access_count = access_count + 1, last_accessed_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, memoryID)
	return err
}

// CleanupExpiredMemories 清理过期记忆
func (r *memoryRepository) CleanupExpiredMemories(ctx context.Context) (int, error) {
	query := `DELETE FROM memory_fragments WHERE expires_at IS NOT NULL AND expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

// CreateConversationContext 创建对话上下文
func (r *memoryRepository) CreateConversationContext(ctx context.Context, context *domain.ConversationContext) error {
	query := `
		INSERT INTO conversation_contexts (
			id, group_chat_id, context_name, current_scene, mood, active_themes,
			active_characters, speaking_queue, current_speaker_id, next_speaker_id,
			turn_count, last_activity_at, recent_memory_ids, context_summary,
			orchestration_mode, intervention_level
		) VALUES (
			:id, :group_chat_id, :context_name, :current_scene, :mood, :active_themes,
			:active_characters, :speaking_queue, :current_speaker_id, :next_speaker_id,
			:turn_count, :last_activity_at, :recent_memory_ids, :context_summary,
			:orchestration_mode, :intervention_level
		)`

	_, err := r.db.NamedExecContext(ctx, query, context)
	return err
}

// GetConversationContextByGroupID 根据群聊ID获取对话上下文
func (r *memoryRepository) GetConversationContextByGroupID(ctx context.Context, groupChatID uuid.UUID) (*domain.ConversationContext, error) {
	query := `
		SELECT id, group_chat_id, context_name, current_scene, mood, active_themes,
			   active_characters, speaking_queue, current_speaker_id, next_speaker_id,
			   turn_count, last_activity_at, recent_memory_ids, context_summary,
			   orchestration_mode, intervention_level, created_at, updated_at
		FROM conversation_contexts WHERE group_chat_id = $1`

	var context domain.ConversationContext
	err := r.db.GetContext(ctx, &context, query, groupChatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "对话上下文不存在", err)
		}
		return nil, err
	}

	return &context, nil
}

// UpdateConversationContext 更新对话上下文
func (r *memoryRepository) UpdateConversationContext(ctx context.Context, context *domain.ConversationContext) error {
	query := `
		UPDATE conversation_contexts SET
			context_name = :context_name, current_scene = :current_scene, mood = :mood,
			active_themes = :active_themes, active_characters = :active_characters,
			speaking_queue = :speaking_queue, current_speaker_id = :current_speaker_id,
			next_speaker_id = :next_speaker_id, turn_count = :turn_count,
			last_activity_at = :last_activity_at, recent_memory_ids = :recent_memory_ids,
			context_summary = :context_summary, orchestration_mode = :orchestration_mode,
			intervention_level = :intervention_level, updated_at = :updated_at
		WHERE id = :id`

	context.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, context)
	return err
}

// DeleteConversationContext 删除对话上下文
func (r *memoryRepository) DeleteConversationContext(ctx context.Context, groupChatID uuid.UUID) error {
	query := `DELETE FROM conversation_contexts WHERE group_chat_id = $1`
	_, err := r.db.ExecContext(ctx, query, groupChatID)
	return err
}

// GetMemoryStats 获取记忆统计
func (r *memoryRepository) GetMemoryStats(ctx context.Context, characterID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_memories,
			COUNT(CASE WHEN memory_type = 'conversation' THEN 1 END) as conversation_memories,
			COUNT(CASE WHEN memory_type = 'event' THEN 1 END) as event_memories,
			COUNT(CASE WHEN memory_type = 'fact' THEN 1 END) as fact_memories,
			COUNT(CASE WHEN memory_type = 'emotion' THEN 1 END) as emotion_memories,
			COUNT(CASE WHEN memory_type = 'relationship' THEN 1 END) as relationship_memories,
			AVG(importance_score) as avg_importance,
			MAX(importance_score) as max_importance,
			SUM(access_count) as total_accesses
		FROM memory_fragments 
		WHERE character_id = $1 
		  AND (expires_at IS NULL OR expires_at > NOW())`

	var stats struct {
		TotalMemories        int     `db:"total_memories"`
		ConversationMemories int     `db:"conversation_memories"`
		EventMemories        int     `db:"event_memories"`
		FactMemories         int     `db:"fact_memories"`
		EmotionMemories      int     `db:"emotion_memories"`
		RelationshipMemories int     `db:"relationship_memories"`
		AvgImportance        float64 `db:"avg_importance"`
		MaxImportance        float64 `db:"max_importance"`
		TotalAccesses        int     `db:"total_accesses"`
	}

	err := r.db.GetContext(ctx, &stats, query, characterID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_memories":        stats.TotalMemories,
		"conversation_memories": stats.ConversationMemories,
		"event_memories":        stats.EventMemories,
		"fact_memories":         stats.FactMemories,
		"emotion_memories":      stats.EmotionMemories,
		"relationship_memories": stats.RelationshipMemories,
		"avg_importance":        stats.AvgImportance,
		"max_importance":        stats.MaxImportance,
		"total_accesses":        stats.TotalAccesses,
	}

	return result, nil
}
