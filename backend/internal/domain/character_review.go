package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CharacterReviewStatus 角色审核状态
type CharacterReviewStatus string

const (
	ReviewStatusPending  CharacterReviewStatus = "pending"  // 待审核
	ReviewStatusApproved CharacterReviewStatus = "approved" // 已通过
	ReviewStatusRejected CharacterReviewStatus = "rejected" // 已拒绝
	ReviewStatusRevision CharacterReviewStatus = "revision" // 需修改
)

// CharacterReview 角色审核记录
type CharacterReview struct {
	ID          uuid.UUID             `json:"id" db:"id"`
	CharacterID uuid.UUID             `json:"character_id" db:"character_id"`
	ReviewerID  *uuid.UUID            `json:"reviewer_id,omitempty" db:"reviewer_id"`
	Status      CharacterReviewStatus `json:"status" db:"status"`
	ReviewType  string                `json:"review_type" db:"review_type"` // content, avatar, name
	Score       int                   `json:"score" db:"score"`             // 审核评分 0-100
	Reason      string                `json:"reason" db:"reason"`           // 审核原因/建议
	AutoReview  bool                  `json:"auto_review" db:"auto_review"` // 是否自动审核
	ReviewData  *json.RawMessage      `json:"review_data" db:"review_data"` // 审核详细数据
	CreatedAt   time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at" db:"updated_at"`
	ReviewedAt  *time.Time            `json:"reviewed_at,omitempty" db:"reviewed_at"`
}

// ReviewRule 审核规则
type ReviewRule struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	Name        string           `json:"name" db:"name"`
	Type        string           `json:"type" db:"type"`         // content, avatar, name
	Category    string           `json:"category" db:"category"` // sensitive_words, inappropriate_content, etc.
	Pattern     string           `json:"pattern" db:"pattern"`   // 匹配模式
	Action      string           `json:"action" db:"action"`     // reject, flag, warn
	Severity    int              `json:"severity" db:"severity"` // 严重程度 1-10
	Enabled     bool             `json:"enabled" db:"enabled"`
	Description string           `json:"description" db:"description"`
	Config      *json.RawMessage `json:"config" db:"config"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// SensitiveWord 敏感词
type SensitiveWord struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Word      string    `json:"word" db:"word"`
	Category  string    `json:"category" db:"category"` // political, violence, adult, etc.
	Level     int       `json:"level" db:"level"`       // 敏感级别 1-5
	Action    string    `json:"action" db:"action"`     // block, replace, warn
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ReviewResult 审核结果
type ReviewResult struct {
	Passed       bool                   `json:"passed"`
	Score        int                    `json:"score"`
	Reason       string                 `json:"reason"`
	Suggestions  []string               `json:"suggestions"`
	Violations   []ReviewViolation      `json:"violations"`
	AutoReviewed bool                   `json:"auto_reviewed"`
	Details      map[string]interface{} `json:"details"`
}

// ReviewViolation 审核违规项
type ReviewViolation struct {
	Type       string `json:"type"`       // sensitive_word, inappropriate_content
	Content    string `json:"content"`    // 违规内容
	Position   int    `json:"position"`   // 位置
	Severity   int    `json:"severity"`   // 严重程度
	Suggestion string `json:"suggestion"` // 修改建议
	RuleID     string `json:"rule_id"`    // 触发的规则ID
}

// CharacterReviewRequest 角色审核请求
type CharacterReviewRequest struct {
	CharacterID uuid.UUID `json:"character_id" binding:"required"`
	ReviewType  string    `json:"review_type" binding:"required"` // content, avatar, name
	Priority    int       `json:"priority"`                       // 优先级 1-5
	Reason      string    `json:"reason"`                         // 审核原因
}

// CharacterReviewResponse 角色审核响应
type CharacterReviewResponse struct {
	ReviewID uuid.UUID    `json:"review_id"`
	Status   string       `json:"status"`
	Result   ReviewResult `json:"result"`
	Message  string       `json:"message"`
}

// ReviewStatistics 审核统计
type ReviewStatistics struct {
	TotalReviews    int64                           `json:"total_reviews"`
	PendingReviews  int64                           `json:"pending_reviews"`
	ApprovedReviews int64                           `json:"approved_reviews"`
	RejectedReviews int64                           `json:"rejected_reviews"`
	AutoReviews     int64                           `json:"auto_reviews"`
	ManualReviews   int64                           `json:"manual_reviews"`
	AvgScore        float64                         `json:"avg_score"`
	StatusBreakdown map[CharacterReviewStatus]int64 `json:"status_breakdown"`
	TypeBreakdown   map[string]int64                `json:"type_breakdown"`
	UpdatedAt       time.Time                       `json:"updated_at"`
}

// ReviewConfig 审核配置
type ReviewConfig struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	AutoReviewEnabled bool                   `json:"auto_review_enabled" db:"auto_review_enabled"`
	PassThreshold     int                    `json:"pass_threshold" db:"pass_threshold"`     // 通过阈值
	RejectThreshold   int                    `json:"reject_threshold" db:"reject_threshold"` // 拒绝阈值
	RequireManual     bool                   `json:"require_manual" db:"require_manual"`     // 是否需要人工审核
	ReviewTimeout     int                    `json:"review_timeout" db:"review_timeout"`     // 审核超时时间(小时)
	NotifyReviewer    bool                   `json:"notify_reviewer" db:"notify_reviewer"`   // 是否通知审核员
	Settings          map[string]interface{} `json:"settings" db:"settings"`                 // 其他设置
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// ReviewerAssignment 审核员分配
type ReviewerAssignment struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ReviewerID uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	ReviewID   uuid.UUID `json:"review_id" db:"review_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	Status     string    `json:"status" db:"status"` // assigned, completed, timeout
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// ReviewLog 审核日志
type ReviewLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	ReviewID  uuid.UUID              `json:"review_id" db:"review_id"`
	Action    string                 `json:"action" db:"action"`         // created, assigned, reviewed, approved, rejected
	ActorID   *uuid.UUID             `json:"actor_id" db:"actor_id"`     // 操作者ID
	ActorType string                 `json:"actor_type" db:"actor_type"` // user, system, admin
	Details   map[string]interface{} `json:"details" db:"details"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// ContentAnalysis 内容分析结果
type ContentAnalysis struct {
	TextLength     int                    `json:"text_length"`
	SentimentScore float64                `json:"sentiment_score"` // 情感分数 -1 到 1
	ToxicityScore  float64                `json:"toxicity_score"`  // 毒性分数 0 到 1
	AdultScore     float64                `json:"adult_score"`     // 成人内容分数 0 到 1
	ViolenceScore  float64                `json:"violence_score"`  // 暴力内容分数 0 到 1
	Keywords       []string               `json:"keywords"`
	Topics         []string               `json:"topics"`
	Language       string                 `json:"language"`
	Confidence     float64                `json:"confidence"`
	Flags          []string               `json:"flags"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ImageAnalysis 图像分析结果
type ImageAnalysis struct {
	Width         int                    `json:"width"`
	Height        int                    `json:"height"`
	Format        string                 `json:"format"`
	Size          int64                  `json:"size"`
	AdultScore    float64                `json:"adult_score"`    // 成人内容分数
	ViolenceScore float64                `json:"violence_score"` // 暴力内容分数
	RacyScore     float64                `json:"racy_score"`     // 性感内容分数
	Objects       []string               `json:"objects"`        // 识别的物体
	Faces         int                    `json:"faces"`          // 人脸数量
	Quality       float64                `json:"quality"`        // 图像质量分数
	Flags         []string               `json:"flags"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ReviewQueue 审核队列项
type ReviewQueue struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ReviewID    uuid.UUID  `json:"review_id" db:"review_id"`
	Priority    int        `json:"priority" db:"priority"` // 优先级 1-5
	QueuedAt    time.Time  `json:"queued_at" db:"queued_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	Status      string     `json:"status" db:"status"` // queued, processing, completed
	RetryCount  int        `json:"retry_count" db:"retry_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// ReviewTemplate 审核模板
type ReviewTemplate struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        string                 `json:"type" db:"type"` // content, avatar, name
	Description string                 `json:"description" db:"description"`
	Rules       []string               `json:"rules" db:"rules"` // 规则ID列表
	Config      map[string]interface{} `json:"config" db:"config"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ReviewNotification 审核通知
type ReviewNotification struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	ReviewID  uuid.UUID              `json:"review_id" db:"review_id"`
	UserID    uuid.UUID              `json:"user_id" db:"user_id"`
	Type      string                 `json:"type" db:"type"` // approved, rejected, revision_required
	Title     string                 `json:"title" db:"title"`
	Content   string                 `json:"content" db:"content"`
	Data      map[string]interface{} `json:"data" db:"data"`
	Read      bool                   `json:"read" db:"read"`
	SentAt    *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}
