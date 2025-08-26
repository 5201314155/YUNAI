package domain

import (
	"time"

	"github.com/google/uuid"
)

// Moment 朋友圈动态
type Moment struct {
	ID           uuid.UUID `json:"id"`
	CharacterID  uuid.UUID `json:"character_id"`
	UserID       uuid.UUID `json:"user_id"`
	Content      string    `json:"content"`
	ContentType  string    `json:"content_type"` // text, image, video, talking_head, 3d
	MediaURL     *string   `json:"media_url,omitempty"`
	MediaType    *string   `json:"media_type,omitempty"`
	Visibility   string    `json:"visibility"`         // public, friends, private
	IsGenerated  bool      `json:"is_generated"`       // AI生成还是用户创建
	Mood         *string   `json:"mood,omitempty"`     // 心情标签
	Location     *string   `json:"location,omitempty"` // 位置信息
	Tags         []string  `json:"tags,omitempty"`     // 标签
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	ShareCount   int       `json:"share_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MomentDraft 朋友圈草稿
type MomentDraft struct {
	ID              uuid.UUID  `json:"id"`
	CharacterID     uuid.UUID  `json:"character_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Content         string     `json:"content"`
	ContentType     string     `json:"content_type"`
	MediaPrompt     *string    `json:"media_prompt,omitempty"` // 媒体生成提示词
	Visibility      string     `json:"visibility"`
	Mood            *string    `json:"mood,omitempty"`
	Tags            []string   `json:"tags,omitempty"`
	Priority        int        `json:"priority"`         // 优先级，用于排序
	SimilarityScore float64    `json:"similarity_score"` // 与历史内容的相似度
	GeneratedAt     time.Time  `json:"generated_at"`
	IsPublished     bool       `json:"is_published"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
}

// MomentGenerationRequest 朋友圈生成请求
type MomentGenerationRequest struct {
	CharacterID  uuid.UUID `json:"character_id" validate:"required"`
	UserID       uuid.UUID `json:"user_id" validate:"required"`
	Count        int       `json:"count" validate:"min=1,max=10"` // 生成数量，默认3-5条
	ContentTypes []string  `json:"content_types,omitempty"`       // 指定内容类型
	Mood         *string   `json:"mood,omitempty"`                // 指定心情
	Context      *string   `json:"context,omitempty"`             // 额外上下文
	AutoInteract bool      `json:"auto_interact"`                 // 是否自动触发关系网角色互动
	BasedOnChat  bool      `json:"based_on_chat"`                 // 是否基于最新聊天记录生成
}

// MomentGenerationResult 朋友圈生成结果
type MomentGenerationResult struct {
	CharacterID   uuid.UUID      `json:"character_id"`
	UserID        uuid.UUID      `json:"user_id"`
	GeneratedAt   time.Time      `json:"generated_at"`
	Success       bool           `json:"success"`
	Message       string         `json:"message"`
	Drafts        []*MomentDraft `json:"drafts"`
	TotalCount    int            `json:"total_count"`
	FilteredCount int            `json:"filtered_count"` // 被相似度过滤的数量
}

// MomentPublishRequest 朋友圈发布请求
type MomentPublishRequest struct {
	DraftID       uuid.UUID `json:"draft_id" validate:"required"`
	UserID        uuid.UUID `json:"user_id" validate:"required"`
	Visibility    *string   `json:"visibility,omitempty"` // 可选择修改可见性
	GenerateMedia bool      `json:"generate_media"`       // 是否生成媒体内容
}

// MomentPublishResult 朋友圈发布结果
type MomentPublishResult struct {
	MomentID    uuid.UUID `json:"moment_id"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	MediaURL    *string   `json:"media_url,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}

// MomentListRequest 朋友圈列表请求
type MomentListRequest struct {
	UserID      uuid.UUID  `json:"user_id" validate:"required"`
	CharacterID *uuid.UUID `json:"character_id,omitempty"` // 指定角色，为空则获取所有
	Visibility  *string    `json:"visibility,omitempty"`   // 可见性过滤
	ContentType *string    `json:"content_type,omitempty"` // 内容类型过滤
	Page        int        `json:"page" validate:"min=1"`
	Limit       int        `json:"limit" validate:"min=1,max=50"`
}

// MomentResponse 朋友圈响应
type MomentResponse struct {
	Moment    *Moment            `json:"moment"`
	Character *CharacterResponse `json:"character"`
	IsLiked   bool               `json:"is_liked"`   // 用户是否点赞
	CanEdit   bool               `json:"can_edit"`   // 是否可编辑
	CanDelete bool               `json:"can_delete"` // 是否可删除
}

// MomentInteraction 朋友圈互动
type MomentInteraction struct {
	ID          uuid.UUID  `json:"id"`
	MomentID    uuid.UUID  `json:"moment_id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	CharacterID *uuid.UUID `json:"character_id,omitempty"`
	Type        string     `json:"type"`              // like, comment, share
	Content     *string    `json:"content,omitempty"` // 评论内容
	CreatedAt   time.Time  `json:"created_at"`
}

// MomentGenerationContext 朋友圈生成上下文
type MomentGenerationContext struct {
	Character           *CharacterResponse `json:"character"`
	RecentChats         []string           `json:"recent_chats"`                   // 最近聊天内容
	LongTermMemory      []string           `json:"long_term_memory"`               // 长期记忆
	UserInteractions    []string           `json:"user_interactions"`              // 与用户的互动历史
	CurrentMood         *string            `json:"current_mood,omitempty"`         // 当前心情
	RecentMoments       []*Moment          `json:"recent_moments"`                 // 最近的朋友圈，用于去重
	RelationshipContext *string            `json:"relationship_context,omitempty"` // 关系上下文
}

// MomentContentTemplate 朋友圈内容模板
type MomentContentTemplate struct {
	Type        string   `json:"type"`        // daily_life, emotion, memory, interaction, complaint
	Templates   []string `json:"templates"`   // 模板列表
	Mood        string   `json:"mood"`        // 适用心情
	Personality []string `json:"personality"` // 适用性格
}

// MomentAutoGenerationConfig 自动生成配置
type MomentAutoGenerationConfig struct {
	CharacterID     uuid.UUID `json:"character_id"`
	UserID          uuid.UUID `json:"user_id"`
	Enabled         bool      `json:"enabled"`            // 是否启用自动生成
	Frequency       string    `json:"frequency"`          // daily, weekly, custom
	MaxDraftsPerDay int       `json:"max_drafts_per_day"` // 每天最大草稿数
	AutoPublish     bool      `json:"auto_publish"`       // 是否自动发布
	ContentTypes    []string  `json:"content_types"`      // 允许的内容类型
	Visibility      string    `json:"visibility"`         // 默认可见性
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// MomentSimilarityCheck 相似度检查结果
type MomentSimilarityCheck struct {
	Content         string  `json:"content"`
	SimilarContent  string  `json:"similar_content"`
	SimilarityScore float64 `json:"similarity_score"`
	IsDuplicate     bool    `json:"is_duplicate"` // 是否为重复内容
	Threshold       float64 `json:"threshold"`    // 相似度阈值
}

// MomentAnalytics 朋友圈分析数据
type MomentAnalytics struct {
	CharacterID       uuid.UUID  `json:"character_id"`
	UserID            uuid.UUID  `json:"user_id"`
	TotalMoments      int        `json:"total_moments"`
	TotalLikes        int        `json:"total_likes"`
	TotalComments     int        `json:"total_comments"`
	TotalShares       int        `json:"total_shares"`
	AvgLikesPerMoment float64    `json:"avg_likes_per_moment"`
	MostPopularMood   string     `json:"most_popular_mood"`
	MostActiveTime    string     `json:"most_active_time"`
	EngagementRate    float64    `json:"engagement_rate"`
	LastMomentAt      *time.Time `json:"last_moment_at,omitempty"`
}

// MomentNotification 朋友圈通知
type MomentNotification struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	CharacterID uuid.UUID `json:"character_id"`
	MomentID    uuid.UUID `json:"moment_id"`
	Type        string    `json:"type"` // new_moment, like, comment, share, mention
	Content     string    `json:"content"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

// MomentMention 朋友圈@提及
type MomentMention struct {
	ID          uuid.UUID `json:"id"`
	MomentID    uuid.UUID `json:"moment_id"`
	MentionedID uuid.UUID `json:"mentioned_id"` // 被@的用户或角色ID
	MentionType string    `json:"mention_type"` // user, character
	Position    int       `json:"position"`     // 在内容中的位置
	CreatedAt   time.Time `json:"created_at"`
}

// MomentAutoInteraction 朋友圈自动互动配置
type MomentAutoInteraction struct {
	ID                 uuid.UUID `json:"id"`
	CharacterID        uuid.UUID `json:"character_id"`
	UserID             uuid.UUID `json:"user_id"`
	AutoLike           bool      `json:"auto_like"`           // 自动点赞
	AutoComment        bool      `json:"auto_comment"`        // 自动评论
	AutoShare          bool      `json:"auto_share"`          // 自动分享
	LikeProbability    float64   `json:"like_probability"`    // 点赞概率 0-1
	CommentProbability float64   `json:"comment_probability"` // 评论概率 0-1
	ShareProbability   float64   `json:"share_probability"`   // 分享概率 0-1
	InteractionDelay   int       `json:"interaction_delay"`   // 互动延迟(秒)
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SmartMomentGeneration 智能朋友圈生成上下文
type SmartMomentGeneration struct {
	Character         *CharacterResponse       `json:"character"`           // 全局共享角色
	SelectedModel     *AIModel                 `json:"selected_model"`      // 角色选择的聊天模型
	AvailableModels   []*AIModel               `json:"available_models"`    // 可用的聊天模型
	UserRelationships *UserRelationshipMap     `json:"user_relationships"`  // 用户私有关系网络
	UserChats         []*ChatMessage           `json:"user_chats"`          // 用户私有聊天记录
	RecentChats       []*ChatMessage           `json:"recent_chats"`        // 最近聊天记录
	ChatBasedContent  bool                     `json:"chat_based_content"`  // 是否基于聊天生成
	MentionCandidates []*RelationshipCharacter `json:"mention_candidates"`  // 用户关系网中可@的角色
	AutoInteractions  []*MomentAutoInteraction `json:"auto_interactions"`   // 用户私有自动互动配置
	UserNickname      string                   `json:"user_nickname"`       // 用户昵称
	IsSharedCharacter bool                     `json:"is_shared_character"` // 是否为共享角色
}

// MomentInteractionTask 朋友圈互动任务
type MomentInteractionTask struct {
	ID              uuid.UUID  `json:"id"`
	MomentID        uuid.UUID  `json:"moment_id"`
	CharacterID     uuid.UUID  `json:"character_id"`
	InteractionType string     `json:"interaction_type"`  // like, comment, share
	Content         *string    `json:"content,omitempty"` // 评论内容
	ScheduledAt     time.Time  `json:"scheduled_at"`      // 计划执行时间
	ExecutedAt      *time.Time `json:"executed_at,omitempty"`
	Status          string     `json:"status"` // pending, executed, failed
	CreatedAt       time.Time  `json:"created_at"`
}
