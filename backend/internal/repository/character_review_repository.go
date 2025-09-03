package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// CharacterReviewRepository 角色审核仓库接口
type CharacterReviewRepository interface {
	// 审核记录管理
	CreateReview(ctx context.Context, review *domain.CharacterReview) error
	UpdateReview(ctx context.Context, review *domain.CharacterReview) error
	GetReview(ctx context.Context, reviewID uuid.UUID) (*domain.CharacterReview, error)
	ListReviews(ctx context.Context, status domain.CharacterReviewStatus, limit, offset int) ([]*domain.CharacterReview, error)
	GetReviewsByCharacter(ctx context.Context, characterID uuid.UUID) ([]*domain.CharacterReview, error)
	DeleteReview(ctx context.Context, reviewID uuid.UUID) error

	// 审核规则管理
	CreateReviewRule(ctx context.Context, rule *domain.ReviewRule) error
	UpdateReviewRule(ctx context.Context, rule *domain.ReviewRule) error
	GetReviewRule(ctx context.Context, ruleID uuid.UUID) (*domain.ReviewRule, error)
	ListReviewRules(ctx context.Context) ([]*domain.ReviewRule, error)
	DeleteReviewRule(ctx context.Context, ruleID uuid.UUID) error

	// 敏感词管理
	CreateSensitiveWord(ctx context.Context, word *domain.SensitiveWord) error
	UpdateSensitiveWord(ctx context.Context, word *domain.SensitiveWord) error
	GetSensitiveWord(ctx context.Context, wordID uuid.UUID) (*domain.SensitiveWord, error)
	ListSensitiveWords(ctx context.Context) ([]*domain.SensitiveWord, error)
	DeleteSensitiveWord(ctx context.Context, wordID uuid.UUID) error

	// 审核配置管理
	CreateReviewConfig(ctx context.Context, config *domain.ReviewConfig) error
	UpdateReviewConfig(ctx context.Context, config *domain.ReviewConfig) error
	GetReviewConfig(ctx context.Context) (*domain.ReviewConfig, error)

	// 审核统计
	GetReviewStatistics(ctx context.Context) (*domain.ReviewStatistics, error)
	GetReviewCountByStatus(ctx context.Context, status domain.CharacterReviewStatus) (int64, error)
	GetReviewCountByType(ctx context.Context, reviewType string) (int64, error)

	// 审核队列管理
	AddToQueue(ctx context.Context, queueItem *domain.ReviewQueue) error
	GetNextFromQueue(ctx context.Context) (*domain.ReviewQueue, error)
	UpdateQueueItem(ctx context.Context, queueItem *domain.ReviewQueue) error
	RemoveFromQueue(ctx context.Context, queueID uuid.UUID) error

	// 审核日志
	CreateReviewLog(ctx context.Context, log *domain.ReviewLog) error
	GetReviewLogs(ctx context.Context, reviewID uuid.UUID) ([]*domain.ReviewLog, error)
}

// characterReviewRepository 角色审核仓库实现
type characterReviewRepository struct {
	db *sqlx.DB
}

// NewCharacterReviewRepository 创建角色审核仓库
func NewCharacterReviewRepository(db *sqlx.DB) CharacterReviewRepository {
	return &characterReviewRepository{db: db}
}

// CreateReview 创建审核记录
func (r *characterReviewRepository) CreateReview(ctx context.Context, review *domain.CharacterReview) error {
	query := `
		INSERT INTO character_reviews (id, character_id, reviewer_id, status, review_type,
		                              score, reason, auto_review, review_data, created_at,
		                              updated_at, reviewed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		review.ID, review.CharacterID, review.ReviewerID, review.Status,
		review.ReviewType, review.Score, review.Reason, review.AutoReview,
		review.ReviewData, review.CreatedAt, review.UpdatedAt, review.ReviewedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create character review: %w", err)
	}

	return nil
}

// UpdateReview 更新审核记录
func (r *characterReviewRepository) UpdateReview(ctx context.Context, review *domain.CharacterReview) error {
	query := `
		UPDATE character_reviews 
		SET reviewer_id = $2, status = $3, score = $4, reason = $5, 
		    review_data = $6, updated_at = $7, reviewed_at = $8
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		review.ID, review.ReviewerID, review.Status, review.Score,
		review.Reason, review.ReviewData, review.UpdatedAt, review.ReviewedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update character review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCharacterReviewNotFound
	}

	return nil
}

// GetReview 获取审核记录
func (r *characterReviewRepository) GetReview(ctx context.Context, reviewID uuid.UUID) (*domain.CharacterReview, error) {
	query := `
		SELECT id, character_id, reviewer_id, status, review_type, score, reason, 
		       auto_review, review_data, created_at, updated_at, reviewed_at
		FROM character_reviews 
		WHERE id = $1
	`

	review := &domain.CharacterReview{}
	err := r.db.GetContext(ctx, review, query, reviewID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCharacterReviewNotFound
		}
		return nil, fmt.Errorf("failed to get character review: %w", err)
	}

	return review, nil
}

// ListReviews 列出审核记录
func (r *characterReviewRepository) ListReviews(ctx context.Context, status domain.CharacterReviewStatus, limit, offset int) ([]*domain.CharacterReview, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, character_id, reviewer_id, status, review_type, score, reason, 
			       auto_review, review_data, created_at, updated_at, reviewed_at
			FROM character_reviews 
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
			SELECT id, character_id, reviewer_id, status, review_type, score, reason, 
			       auto_review, review_data, created_at, updated_at, reviewed_at
			FROM character_reviews 
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	var reviews []*domain.CharacterReview
	err := r.db.SelectContext(ctx, &reviews, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list character reviews: %w", err)
	}

	return reviews, nil
}

// GetReviewsByCharacter 获取角色的审核记录
func (r *characterReviewRepository) GetReviewsByCharacter(ctx context.Context, characterID uuid.UUID) ([]*domain.CharacterReview, error) {
	query := `
		SELECT id, character_id, reviewer_id, status, review_type, score, reason, 
		       auto_review, review_data, created_at, updated_at, reviewed_at
		FROM character_reviews 
		WHERE character_id = $1
		ORDER BY created_at DESC
	`

	var reviews []*domain.CharacterReview
	err := r.db.SelectContext(ctx, &reviews, query, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character reviews: %w", err)
	}

	return reviews, nil
}

// DeleteReview 删除审核记录
func (r *characterReviewRepository) DeleteReview(ctx context.Context, reviewID uuid.UUID) error {
	query := `DELETE FROM character_reviews WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, reviewID)
	if err != nil {
		return fmt.Errorf("failed to delete character review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCharacterReviewNotFound
	}

	return nil
}

// CreateReviewRule 创建审核规则
func (r *characterReviewRepository) CreateReviewRule(ctx context.Context, rule *domain.ReviewRule) error {
	query := `
		INSERT INTO review_rules (id, name, type, category, pattern, action, severity,
		                         enabled, description, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Type, rule.Category, rule.Pattern,
		rule.Action, rule.Severity, rule.Enabled, rule.Description,
		rule.Config, rule.CreatedAt, rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create review rule: %w", err)
	}

	return nil
}

// UpdateReviewRule 更新审核规则
func (r *characterReviewRepository) UpdateReviewRule(ctx context.Context, rule *domain.ReviewRule) error {
	query := `
		UPDATE review_rules 
		SET name = $2, type = $3, category = $4, pattern = $5, action = $6, 
		    severity = $7, enabled = $8, description = $9, config = $10, updated_at = $11
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Type, rule.Category, rule.Pattern,
		rule.Action, rule.Severity, rule.Enabled, rule.Description,
		rule.Config, rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update review rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrReviewRuleNotFound
	}

	return nil
}

// GetReviewRule 获取审核规则
func (r *characterReviewRepository) GetReviewRule(ctx context.Context, ruleID uuid.UUID) (*domain.ReviewRule, error) {
	query := `
		SELECT id, name, type, category, pattern, action, severity, enabled, 
		       description, config, created_at, updated_at
		FROM review_rules 
		WHERE id = $1
	`

	rule := &domain.ReviewRule{}
	err := r.db.GetContext(ctx, rule, query, ruleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrReviewRuleNotFound
		}
		return nil, fmt.Errorf("failed to get review rule: %w", err)
	}

	return rule, nil
}

// ListReviewRules 列出审核规则
func (r *characterReviewRepository) ListReviewRules(ctx context.Context) ([]*domain.ReviewRule, error) {
	query := `
		SELECT id, name, type, category, pattern, action, severity, enabled, 
		       description, config, created_at, updated_at
		FROM review_rules 
		ORDER BY created_at DESC
	`

	var rules []*domain.ReviewRule
	err := r.db.SelectContext(ctx, &rules, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list review rules: %w", err)
	}

	return rules, nil
}

// DeleteReviewRule 删除审核规则
func (r *characterReviewRepository) DeleteReviewRule(ctx context.Context, ruleID uuid.UUID) error {
	query := `DELETE FROM review_rules WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete review rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrReviewRuleNotFound
	}

	return nil
}

// CreateSensitiveWord 创建敏感词
func (r *characterReviewRepository) CreateSensitiveWord(ctx context.Context, word *domain.SensitiveWord) error {
	query := `
		INSERT INTO sensitive_words (id, word, category, level, action, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		word.ID, word.Word, word.Category, word.Level,
		word.Action, word.Enabled, word.CreatedAt, word.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create sensitive word: %w", err)
	}

	return nil
}

// UpdateSensitiveWord 更新敏感词
func (r *characterReviewRepository) UpdateSensitiveWord(ctx context.Context, word *domain.SensitiveWord) error {
	query := `
		UPDATE sensitive_words 
		SET word = $2, category = $3, level = $4, action = $5, enabled = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		word.ID, word.Word, word.Category, word.Level,
		word.Action, word.Enabled, word.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update sensitive word: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSensitiveWordNotFound
	}

	return nil
}

// GetSensitiveWord 获取敏感词
func (r *characterReviewRepository) GetSensitiveWord(ctx context.Context, wordID uuid.UUID) (*domain.SensitiveWord, error) {
	query := `
		SELECT id, word, category, level, action, enabled, created_at, updated_at
		FROM sensitive_words 
		WHERE id = $1
	`

	word := &domain.SensitiveWord{}
	err := r.db.GetContext(ctx, word, query, wordID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSensitiveWordNotFound
		}
		return nil, fmt.Errorf("failed to get sensitive word: %w", err)
	}

	return word, nil
}

// ListSensitiveWords 列出敏感词
func (r *characterReviewRepository) ListSensitiveWords(ctx context.Context) ([]*domain.SensitiveWord, error) {
	query := `
		SELECT id, word, category, level, action, enabled, created_at, updated_at
		FROM sensitive_words 
		ORDER BY created_at DESC
	`

	var words []*domain.SensitiveWord
	err := r.db.SelectContext(ctx, &words, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sensitive words: %w", err)
	}

	return words, nil
}

// DeleteSensitiveWord 删除敏感词
func (r *characterReviewRepository) DeleteSensitiveWord(ctx context.Context, wordID uuid.UUID) error {
	query := `DELETE FROM sensitive_words WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, wordID)
	if err != nil {
		return fmt.Errorf("failed to delete sensitive word: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSensitiveWordNotFound
	}

	return nil
}

// GetReviewStatistics 获取审核统计
func (r *characterReviewRepository) GetReviewStatistics(ctx context.Context) (*domain.ReviewStatistics, error) {
	stats := &domain.ReviewStatistics{
		StatusBreakdown: make(map[domain.CharacterReviewStatus]int64),
		TypeBreakdown:   make(map[string]int64),
		UpdatedAt:       time.Now(),
	}

	// 获取总数
	err := r.db.GetContext(ctx, &stats.TotalReviews, "SELECT COUNT(*) FROM character_reviews")
	if err != nil {
		return nil, fmt.Errorf("failed to get total reviews: %w", err)
	}

	// 获取各状态统计
	statusQuery := `
		SELECT status, COUNT(*) as count 
		FROM character_reviews 
		GROUP BY status
	`
	rows, err := r.db.QueryContext(ctx, statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get status breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status domain.CharacterReviewStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		stats.StatusBreakdown[status] = count

		switch status {
		case domain.ReviewStatusPending:
			stats.PendingReviews = count
		case domain.ReviewStatusApproved:
			stats.ApprovedReviews = count
		case domain.ReviewStatusRejected:
			stats.RejectedReviews = count
		}
	}

	// 获取自动/手动审核统计
	err = r.db.GetContext(ctx, &stats.AutoReviews, "SELECT COUNT(*) FROM character_reviews WHERE auto_review = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get auto reviews: %w", err)
	}

	stats.ManualReviews = stats.TotalReviews - stats.AutoReviews

	// 获取平均分数
	err = r.db.GetContext(ctx, &stats.AvgScore, "SELECT COALESCE(AVG(score), 0) FROM character_reviews WHERE score > 0")
	if err != nil {
		return nil, fmt.Errorf("failed to get average score: %w", err)
	}

	return stats, nil
}

// GetReviewCountByStatus 按状态获取审核数量
func (r *characterReviewRepository) GetReviewCountByStatus(ctx context.Context, status domain.CharacterReviewStatus) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM character_reviews WHERE status = $1`

	err := r.db.GetContext(ctx, &count, query, status)
	if err != nil {
		return 0, fmt.Errorf("failed to get review count by status: %w", err)
	}

	return count, nil
}

// GetReviewCountByType 按类型获取审核数量
func (r *characterReviewRepository) GetReviewCountByType(ctx context.Context, reviewType string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM character_reviews WHERE review_type = $1`

	err := r.db.GetContext(ctx, &count, query, reviewType)
	if err != nil {
		return 0, fmt.Errorf("failed to get review count by type: %w", err)
	}

	return count, nil
}

// 其他方法的简化实现...
func (r *characterReviewRepository) CreateReviewConfig(ctx context.Context, config *domain.ReviewConfig) error {
	// 简化实现
	return nil
}

func (r *characterReviewRepository) UpdateReviewConfig(ctx context.Context, config *domain.ReviewConfig) error {
	// 简化实现
	return nil
}

func (r *characterReviewRepository) GetReviewConfig(ctx context.Context) (*domain.ReviewConfig, error) {
	// 简化实现
	return &domain.ReviewConfig{
		AutoReviewEnabled: true,
		PassThreshold:     70,
		RejectThreshold:   30,
		RequireManual:     false,
		ReviewTimeout:     24,
		NotifyReviewer:    true,
	}, nil
}

func (r *characterReviewRepository) AddToQueue(ctx context.Context, queueItem *domain.ReviewQueue) error {
	// 简化实现
	return nil
}

func (r *characterReviewRepository) GetNextFromQueue(ctx context.Context) (*domain.ReviewQueue, error) {
	// 简化实现
	return nil, nil
}

func (r *characterReviewRepository) UpdateQueueItem(ctx context.Context, queueItem *domain.ReviewQueue) error {
	// 简化实现
	return nil
}

func (r *characterReviewRepository) RemoveFromQueue(ctx context.Context, queueID uuid.UUID) error {
	// 简化实现
	return nil
}

func (r *characterReviewRepository) CreateReviewLog(ctx context.Context, log *domain.ReviewLog) error {
	// 简化实现
	return nil
}

func (r *characterReviewRepository) GetReviewLogs(ctx context.Context, reviewID uuid.UUID) ([]*domain.ReviewLog, error) {
	// 简化实现
	return []*domain.ReviewLog{}, nil
}
