package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"yunai/internal/domain"
)

// MomentsRepository 朋友圈仓库接口
type MomentsRepository interface {
	// 朋友圈动态
	CreateMoment(ctx context.Context, moment *domain.Moment) error
	GetMoment(ctx context.Context, id uuid.UUID) (*domain.Moment, error)
	ListMoments(ctx context.Context, req *domain.MomentListRequest) ([]*domain.Moment, int, error)
	UpdateMoment(ctx context.Context, moment *domain.Moment) error
	DeleteMoment(ctx context.Context, id uuid.UUID) error

	// 朋友圈草稿
	CreateMomentDraft(ctx context.Context, draft *domain.MomentDraft) error
	GetMomentDraft(ctx context.Context, id uuid.UUID) (*domain.MomentDraft, error)
	ListMomentDrafts(ctx context.Context, characterID, userID uuid.UUID, limit int) ([]*domain.MomentDraft, error)
	UpdateMomentDraft(ctx context.Context, draft *domain.MomentDraft) error
	DeleteMomentDraft(ctx context.Context, id uuid.UUID) error
	PublishDraft(ctx context.Context, draftID uuid.UUID) (*domain.Moment, error)

	// 朋友圈互动
	CreateInteraction(ctx context.Context, interaction *domain.MomentInteraction) error
	GetInteractions(ctx context.Context, momentID uuid.UUID, interactionType string) ([]*domain.MomentInteraction, error)
	DeleteInteraction(ctx context.Context, id uuid.UUID) error

	// 自动生成配置
	GetAutoGenerationConfig(ctx context.Context, characterID uuid.UUID) (*domain.MomentAutoGenerationConfig, error)
	UpsertAutoGenerationConfig(ctx context.Context, config *domain.MomentAutoGenerationConfig) error

	// 相似度检查
	CheckSimilarity(ctx context.Context, characterID uuid.UUID, content string, threshold float64) ([]*domain.MomentSimilarityCheck, error)

	// 统计分析
	GetMomentAnalytics(ctx context.Context, characterID, userID uuid.UUID) (*domain.MomentAnalytics, error)

	// 通知
	CreateNotification(ctx context.Context, notification *domain.MomentNotification) error
	GetUnreadNotifications(ctx context.Context, userID uuid.UUID) ([]*domain.MomentNotification, error)
	MarkNotificationAsRead(ctx context.Context, id uuid.UUID) error
}

type momentsRepository struct {
	db *sql.DB
}

// NewMomentsRepository 创建朋友圈仓库
func NewMomentsRepository(db *sql.DB) MomentsRepository {
	return &momentsRepository{db: db}
}

// CreateMoment 创建朋友圈动态
func (r *momentsRepository) CreateMoment(ctx context.Context, moment *domain.Moment) error {
	query := `
		INSERT INTO moments (
			id, character_id, user_id, content, content_type, media_url, media_type,
			visibility, is_generated, mood, location, tags, like_count, comment_count, share_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		moment.ID, moment.CharacterID, moment.UserID, moment.Content, moment.ContentType,
		moment.MediaURL, moment.MediaType, moment.Visibility, moment.IsGenerated,
		moment.Mood, moment.Location, pq.Array(moment.Tags),
		moment.LikeCount, moment.CommentCount, moment.ShareCount,
	)

	return err
}

// GetMoment 获取朋友圈动态
func (r *momentsRepository) GetMoment(ctx context.Context, id uuid.UUID) (*domain.Moment, error) {
	query := `
		SELECT id, character_id, user_id, content, content_type, media_url, media_type,
			   visibility, is_generated, mood, location, tags, like_count, comment_count, share_count,
			   created_at, updated_at
		FROM moments WHERE id = $1
	`

	moment := &domain.Moment{}
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&moment.ID, &moment.CharacterID, &moment.UserID, &moment.Content, &moment.ContentType,
		&moment.MediaURL, &moment.MediaType, &moment.Visibility, &moment.IsGenerated,
		&moment.Mood, &moment.Location, &tags, &moment.LikeCount, &moment.CommentCount, &moment.ShareCount,
		&moment.CreatedAt, &moment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	moment.Tags = []string(tags)
	return moment, nil
}

// ListMoments 获取朋友圈动态列表
func (r *momentsRepository) ListMoments(ctx context.Context, req *domain.MomentListRequest) ([]*domain.Moment, int, error) {
	// 构建查询条件
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if req.CharacterID != nil {
		conditions = append(conditions, fmt.Sprintf("character_id = $%d", argIndex))
		args = append(args, *req.CharacterID)
		argIndex++
	}

	if req.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, *req.Visibility)
		argIndex++
	}

	if req.ContentType != nil {
		conditions = append(conditions, fmt.Sprintf("content_type = $%d", argIndex))
		args = append(args, *req.ContentType)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM moments WHERE %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (req.Page - 1) * req.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, character_id, user_id, content, content_type, media_url, media_type,
			   visibility, is_generated, mood, location, tags, like_count, comment_count, share_count,
			   created_at, updated_at
		FROM moments 
		WHERE %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var moments []*domain.Moment
	for rows.Next() {
		moment := &domain.Moment{}
		var tags pq.StringArray

		err := rows.Scan(
			&moment.ID, &moment.CharacterID, &moment.UserID, &moment.Content, &moment.ContentType,
			&moment.MediaURL, &moment.MediaType, &moment.Visibility, &moment.IsGenerated,
			&moment.Mood, &moment.Location, &tags, &moment.LikeCount, &moment.CommentCount, &moment.ShareCount,
			&moment.CreatedAt, &moment.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		moment.Tags = []string(tags)
		moments = append(moments, moment)
	}

	return moments, total, nil
}

// UpdateMoment 更新朋友圈动态
func (r *momentsRepository) UpdateMoment(ctx context.Context, moment *domain.Moment) error {
	query := `
		UPDATE moments SET 
			content = $2, content_type = $3, media_url = $4, media_type = $5,
			visibility = $6, mood = $7, location = $8, tags = $9,
			like_count = $10, comment_count = $11, share_count = $12,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		moment.ID, moment.Content, moment.ContentType, moment.MediaURL, moment.MediaType,
		moment.Visibility, moment.Mood, moment.Location, pq.Array(moment.Tags),
		moment.LikeCount, moment.CommentCount, moment.ShareCount,
	)

	return err
}

// DeleteMoment 删除朋友圈动态
func (r *momentsRepository) DeleteMoment(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM moments WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CreateMomentDraft 创建朋友圈草稿
func (r *momentsRepository) CreateMomentDraft(ctx context.Context, draft *domain.MomentDraft) error {
	query := `
		INSERT INTO moment_drafts (
			id, character_id, user_id, content, content_type, media_prompt,
			visibility, mood, tags, priority, similarity_score
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		draft.ID, draft.CharacterID, draft.UserID, draft.Content, draft.ContentType,
		draft.MediaPrompt, draft.Visibility, draft.Mood, pq.Array(draft.Tags),
		draft.Priority, draft.SimilarityScore,
	)

	return err
}

// GetMomentDraft 获取朋友圈草稿
func (r *momentsRepository) GetMomentDraft(ctx context.Context, id uuid.UUID) (*domain.MomentDraft, error) {
	query := `
		SELECT id, character_id, user_id, content, content_type, media_prompt,
			   visibility, mood, tags, priority, similarity_score,
			   generated_at, is_published, published_at
		FROM moment_drafts WHERE id = $1
	`

	draft := &domain.MomentDraft{}
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&draft.ID, &draft.CharacterID, &draft.UserID, &draft.Content, &draft.ContentType,
		&draft.MediaPrompt, &draft.Visibility, &draft.Mood, &tags,
		&draft.Priority, &draft.SimilarityScore, &draft.GeneratedAt,
		&draft.IsPublished, &draft.PublishedAt,
	)

	if err != nil {
		return nil, err
	}

	draft.Tags = []string(tags)
	return draft, nil
}

// ListMomentDrafts 获取朋友圈草稿列表
func (r *momentsRepository) ListMomentDrafts(ctx context.Context, characterID, userID uuid.UUID, limit int) ([]*domain.MomentDraft, error) {
	query := `
		SELECT id, character_id, user_id, content, content_type, media_prompt,
			   visibility, mood, tags, priority, similarity_score,
			   generated_at, is_published, published_at
		FROM moment_drafts 
		WHERE character_id = $1 AND user_id = $2 AND is_published = false
		ORDER BY priority DESC, generated_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, characterID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []*domain.MomentDraft
	for rows.Next() {
		draft := &domain.MomentDraft{}
		var tags pq.StringArray

		err := rows.Scan(
			&draft.ID, &draft.CharacterID, &draft.UserID, &draft.Content, &draft.ContentType,
			&draft.MediaPrompt, &draft.Visibility, &draft.Mood, &tags,
			&draft.Priority, &draft.SimilarityScore, &draft.GeneratedAt,
			&draft.IsPublished, &draft.PublishedAt,
		)
		if err != nil {
			return nil, err
		}

		draft.Tags = []string(tags)
		drafts = append(drafts, draft)
	}

	return drafts, nil
}

// UpdateMomentDraft 更新朋友圈草稿
func (r *momentsRepository) UpdateMomentDraft(ctx context.Context, draft *domain.MomentDraft) error {
	query := `
		UPDATE moment_drafts SET 
			content = $2, content_type = $3, media_prompt = $4,
			visibility = $5, mood = $6, tags = $7, priority = $8,
			similarity_score = $9, is_published = $10, published_at = $11
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		draft.ID, draft.Content, draft.ContentType, draft.MediaPrompt,
		draft.Visibility, draft.Mood, pq.Array(draft.Tags), draft.Priority,
		draft.SimilarityScore, draft.IsPublished, draft.PublishedAt,
	)

	return err
}

// DeleteMomentDraft 删除朋友圈草稿
func (r *momentsRepository) DeleteMomentDraft(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM moment_drafts WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// PublishDraft 发布草稿为正式朋友圈
func (r *momentsRepository) PublishDraft(ctx context.Context, draftID uuid.UUID) (*domain.Moment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 获取草稿
	draft, err := r.GetMomentDraft(ctx, draftID)
	if err != nil {
		return nil, err
	}

	// 创建朋友圈动态
	moment := &domain.Moment{
		ID:          uuid.New(),
		CharacterID: draft.CharacterID,
		UserID:      draft.UserID,
		Content:     draft.Content,
		ContentType: draft.ContentType,
		Visibility:  draft.Visibility,
		IsGenerated: true,
		Mood:        draft.Mood,
		Tags:        draft.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = r.CreateMoment(ctx, moment)
	if err != nil {
		return nil, err
	}

	// 标记草稿为已发布
	now := time.Now()
	draft.IsPublished = true
	draft.PublishedAt = &now
	err = r.UpdateMomentDraft(ctx, draft)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return moment, nil
}

// CreateInteraction 创建朋友圈互动
func (r *momentsRepository) CreateInteraction(ctx context.Context, interaction *domain.MomentInteraction) error {
	query := `
		INSERT INTO moment_interactions (id, moment_id, user_id, character_id, type, content)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		interaction.ID, interaction.MomentID, interaction.UserID,
		interaction.CharacterID, interaction.Type, interaction.Content,
	)

	return err
}

// GetInteractions 获取朋友圈互动
func (r *momentsRepository) GetInteractions(ctx context.Context, momentID uuid.UUID, interactionType string) ([]*domain.MomentInteraction, error) {
	query := `
		SELECT id, moment_id, user_id, character_id, type, content, created_at
		FROM moment_interactions
		WHERE moment_id = $1 AND type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, momentID, interactionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interactions []*domain.MomentInteraction
	for rows.Next() {
		interaction := &domain.MomentInteraction{}

		err := rows.Scan(
			&interaction.ID, &interaction.MomentID, &interaction.UserID,
			&interaction.CharacterID, &interaction.Type, &interaction.Content,
			&interaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		interactions = append(interactions, interaction)
	}

	return interactions, nil
}

// DeleteInteraction 删除朋友圈互动
func (r *momentsRepository) DeleteInteraction(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM moment_interactions WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetAutoGenerationConfig 获取自动生成配置
func (r *momentsRepository) GetAutoGenerationConfig(ctx context.Context, characterID uuid.UUID) (*domain.MomentAutoGenerationConfig, error) {
	query := `
		SELECT character_id, user_id, enabled, frequency, max_drafts_per_day,
			   auto_publish, content_types, visibility, created_at, updated_at
		FROM moment_auto_generation_configs
		WHERE character_id = $1
	`

	config := &domain.MomentAutoGenerationConfig{}
	var contentTypes pq.StringArray

	err := r.db.QueryRowContext(ctx, query, characterID).Scan(
		&config.CharacterID, &config.UserID, &config.Enabled, &config.Frequency,
		&config.MaxDraftsPerDay, &config.AutoPublish, &contentTypes, &config.Visibility,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	config.ContentTypes = []string(contentTypes)
	return config, nil
}

// UpsertAutoGenerationConfig 创建或更新自动生成配置
func (r *momentsRepository) UpsertAutoGenerationConfig(ctx context.Context, config *domain.MomentAutoGenerationConfig) error {
	query := `
		INSERT INTO moment_auto_generation_configs (
			character_id, user_id, enabled, frequency, max_drafts_per_day,
			auto_publish, content_types, visibility
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (character_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			frequency = EXCLUDED.frequency,
			max_drafts_per_day = EXCLUDED.max_drafts_per_day,
			auto_publish = EXCLUDED.auto_publish,
			content_types = EXCLUDED.content_types,
			visibility = EXCLUDED.visibility,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		config.CharacterID, config.UserID, config.Enabled, config.Frequency,
		config.MaxDraftsPerDay, config.AutoPublish, pq.Array(config.ContentTypes),
		config.Visibility,
	)

	return err
}

// CheckSimilarity 检查内容相似度
func (r *momentsRepository) CheckSimilarity(ctx context.Context, characterID uuid.UUID, content string, threshold float64) ([]*domain.MomentSimilarityCheck, error) {
	// 获取该角色最近的朋友圈内容
	query := `
		SELECT content FROM moments
		WHERE character_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := r.db.QueryContext(ctx, query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*domain.MomentSimilarityCheck
	for rows.Next() {
		var existingContent string
		err := rows.Scan(&existingContent)
		if err != nil {
			continue
		}

		// 简单的相似度计算（实际应用中可以使用更复杂的算法）
		similarity := calculateTextSimilarity(content, existingContent)

		if similarity >= threshold {
			check := &domain.MomentSimilarityCheck{
				Content:         content,
				SimilarContent:  existingContent,
				SimilarityScore: similarity,
				IsDuplicate:     similarity > 0.8,
				Threshold:       threshold,
			}
			checks = append(checks, check)
		}
	}

	return checks, nil
}

// GetMomentAnalytics 获取朋友圈分析数据
func (r *momentsRepository) GetMomentAnalytics(ctx context.Context, characterID, userID uuid.UUID) (*domain.MomentAnalytics, error) {
	query := `
		SELECT character_id, user_id, total_moments, total_likes, total_comments,
			   total_shares, avg_likes_per_moment, most_popular_mood,
			   most_active_hour, engagement_rate, last_moment_at
		FROM moment_analytics
		WHERE character_id = $1 AND user_id = $2
	`

	analytics := &domain.MomentAnalytics{}
	var mostActiveHour sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, characterID, userID).Scan(
		&analytics.CharacterID, &analytics.UserID, &analytics.TotalMoments,
		&analytics.TotalLikes, &analytics.TotalComments, &analytics.TotalShares,
		&analytics.AvgLikesPerMoment, &analytics.MostPopularMood,
		&mostActiveHour, &analytics.EngagementRate, &analytics.LastMomentAt,
	)

	if err != nil {
		return nil, err
	}

	if mostActiveHour.Valid {
		analytics.MostActiveTime = fmt.Sprintf("%02d:00", int(mostActiveHour.Float64))
	}

	return analytics, nil
}

// CreateNotification 创建朋友圈通知
func (r *momentsRepository) CreateNotification(ctx context.Context, notification *domain.MomentNotification) error {
	query := `
		INSERT INTO moment_notifications (id, user_id, character_id, moment_id, type, content)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.UserID, notification.CharacterID,
		notification.MomentID, notification.Type, notification.Content,
	)

	return err
}

// GetUnreadNotifications 获取未读通知
func (r *momentsRepository) GetUnreadNotifications(ctx context.Context, userID uuid.UUID) ([]*domain.MomentNotification, error) {
	query := `
		SELECT id, user_id, character_id, moment_id, type, content, is_read, created_at
		FROM moment_notifications
		WHERE user_id = $1 AND is_read = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.MomentNotification
	for rows.Next() {
		notification := &domain.MomentNotification{}

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.CharacterID,
			&notification.MomentID, &notification.Type, &notification.Content,
			&notification.IsRead, &notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkNotificationAsRead 标记通知为已读
func (r *momentsRepository) MarkNotificationAsRead(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE moment_notifications SET is_read = true WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// calculateTextSimilarity 计算文本相似度（简单实现）
func calculateTextSimilarity(text1, text2 string) float64 {
	if text1 == text2 {
		return 1.0
	}

	// 简单的字符级相似度计算
	// 实际应用中可以使用更复杂的算法，如编辑距离、余弦相似度等
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// 计算交集
	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}

	intersection := 0
	for _, word := range words2 {
		if wordSet1[word] {
			intersection++
		}
	}

	// 计算Jaccard相似度
	union := len(words1) + len(words2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
