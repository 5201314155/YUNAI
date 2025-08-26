package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Character 角色
type Character struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Personality *string   `json:"personality" db:"personality"`

	// 角色两图
	BgImageURL     *string `json:"bg_image_url" db:"bg_image_url"`         // 背景图URL
	CutoutImageURL *string `json:"cutout_image_url" db:"cutout_image_url"` // 抠图URL

	// 图片元数据
	BgImageWidth      *int `json:"bg_image_width" db:"bg_image_width"`
	BgImageHeight     *int `json:"bg_image_height" db:"bg_image_height"`
	CutoutImageWidth  *int `json:"cutout_image_width" db:"cutout_image_width"`
	CutoutImageHeight *int `json:"cutout_image_height" db:"cutout_image_height"`

	// 模型配置
	DefaultModelID *uuid.UUID      `json:"default_model_id" db:"default_model_id"`
	ModelParams    json.RawMessage `json:"model_params" db:"model_params"`
	SystemPrompt   *string         `json:"system_prompt" db:"system_prompt"`

	// 可见性设置
	Visibility string `json:"visibility" db:"visibility"`
	IsFeatured bool   `json:"is_featured" db:"is_featured"`

	// 社交设置
	AllowChat      bool `json:"allow_chat" db:"allow_chat"`
	AllowGroupChat bool `json:"allow_group_chat" db:"allow_group_chat"`
	AllowCalls     bool `json:"allow_calls" db:"allow_calls"`

	// 统计信息
	ChatCount int `json:"chat_count" db:"chat_count"`
	LikeCount int `json:"like_count" db:"like_count"`
	ViewCount int `json:"view_count" db:"view_count"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CharacterTag 角色标签
type CharacterTag struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	Tag         string    `json:"tag" db:"tag"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// GroupChat 群聊
type GroupChat struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatorUserID uuid.UUID `json:"creator_user_id" db:"creator_user_id"`
	Name          string    `json:"name" db:"name"`
	Description   *string   `json:"description" db:"description"`

	// 背景设置
	BackgroundImageURL *string `json:"background_image_url" db:"background_image_url"`
	BackgroundMusicURL *string `json:"background_music_url" db:"background_music_url"`
	BackgroundSfxURL   *string `json:"background_sfx_url" db:"background_sfx_url"`

	// 群聊设置
	MaxMembers    int  `json:"max_members" db:"max_members"`
	IsPublic      bool `json:"is_public" db:"is_public"`
	AllowAIInvite bool `json:"allow_ai_invite" db:"allow_ai_invite"`

	// 剧情设置
	WorldSetting *string `json:"world_setting" db:"world_setting"`
	CurrentScene *string `json:"current_scene" db:"current_scene"`
	SceneStyle   *string `json:"scene_style" db:"scene_style"`

	// 统计信息
	MemberCount  int `json:"member_count" db:"member_count"`
	MessageCount int `json:"message_count" db:"message_count"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GroupChatMember 群聊成员
type GroupChatMember struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	GroupChatID uuid.UUID  `json:"group_chat_id" db:"group_chat_id"`
	CharacterID *uuid.UUID `json:"character_id" db:"character_id"`
	UserID      *uuid.UUID `json:"user_id" db:"user_id"`
	MemberType  string     `json:"member_type" db:"member_type"`
	Role        string     `json:"role" db:"role"`
	CanInvite   bool       `json:"can_invite" db:"can_invite"`
	CanKick     bool       `json:"can_kick" db:"can_kick"`
	JoinedAt    time.Time  `json:"joined_at" db:"joined_at"`
	InvitedBy   *uuid.UUID `json:"invited_by" db:"invited_by"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	GroupChatID       *uuid.UUID `json:"group_chat_id" db:"group_chat_id"`
	SenderCharacterID *uuid.UUID `json:"sender_character_id" db:"sender_character_id"`
	SenderUserID      *uuid.UUID `json:"sender_user_id" db:"sender_user_id"`
	SenderType        string     `json:"sender_type" db:"sender_type"`
	Content           string     `json:"content" db:"content"`
	MessageType       string     `json:"message_type" db:"message_type"`

	// 媒体附件
	MediaURLs     pq.StringArray  `json:"media_urls" db:"media_urls"`
	MediaMetadata json.RawMessage `json:"media_metadata" db:"media_metadata"`

	// AI生成信息
	ModelUsed      *string  `json:"model_used" db:"model_used"`
	GenerationCost *float64 `json:"generation_cost" db:"generation_cost"`
	TokensUsed     *int     `json:"tokens_used" db:"tokens_used"`

	// 消息状态
	IsEdited  bool `json:"is_edited" db:"is_edited"`
	IsDeleted bool `json:"is_deleted" db:"is_deleted"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// 请求和响应结构

// CreateCharacterRequest 创建角色请求
type CreateCharacterRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Personality *string `json:"personality,omitempty" validate:"omitempty,max=2000"`

	// 角色两图
	BgImageURL     *string `json:"bg_image_url,omitempty"`
	CutoutImageURL *string `json:"cutout_image_url,omitempty"`

	// 模型配置
	DefaultModelID *uuid.UUID      `json:"default_model_id,omitempty"`
	ModelParams    json.RawMessage `json:"model_params,omitempty"`
	SystemPrompt   *string         `json:"system_prompt,omitempty" validate:"omitempty,max=5000"`

	// 可见性设置
	Visibility string `json:"visibility" validate:"required,oneof=private public friends"`

	// 社交设置
	AllowChat      bool `json:"allow_chat"`
	AllowGroupChat bool `json:"allow_group_chat"`
	AllowCalls     bool `json:"allow_calls"`

	// 标签
	Tags []string `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=50"`
}

// UpdateCharacterRequest 更新角色请求
type UpdateCharacterRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Personality *string `json:"personality,omitempty" validate:"omitempty,max=2000"`

	// 角色两图
	BgImageURL     *string `json:"bg_image_url,omitempty"`
	CutoutImageURL *string `json:"cutout_image_url,omitempty"`

	// 模型配置
	DefaultModelID *uuid.UUID      `json:"default_model_id,omitempty"`
	ModelParams    json.RawMessage `json:"model_params,omitempty"`
	SystemPrompt   *string         `json:"system_prompt,omitempty" validate:"omitempty,max=5000"`

	// 可见性设置
	Visibility *string `json:"visibility,omitempty" validate:"omitempty,oneof=private public friends"`

	// 社交设置
	AllowChat      *bool `json:"allow_chat,omitempty"`
	AllowGroupChat *bool `json:"allow_group_chat,omitempty"`
	AllowCalls     *bool `json:"allow_calls,omitempty"`

	// 标签
	Tags []string `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=50"`
}

// CharacterListRequest 角色列表请求
type CharacterListRequest struct {
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Visibility *string    `json:"visibility,omitempty" validate:"omitempty,oneof=private public friends"`
	Tag        *string    `json:"tag,omitempty"`
	Search     *string    `json:"search,omitempty"`
	IsFeatured *bool      `json:"is_featured,omitempty"`
	Page       int        `json:"page" validate:"min=1"`
	Limit      int        `json:"limit" validate:"min=1,max=100"`
}

// CharacterResponse 角色响应
type CharacterResponse struct {
	*Character
	Tags      []string `json:"tags"`
	IsOwner   bool     `json:"is_owner"`
	CanEdit   bool     `json:"can_edit"`
	ModelName *string  `json:"model_name,omitempty"`
}

// CharacterListResponse 角色列表响应
type CharacterListResponse struct {
	Characters []CharacterResponse `json:"characters"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
}

// CreateGroupChatRequest 创建群聊请求
type CreateGroupChatRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`

	// 背景设置
	BackgroundImageURL *string `json:"background_image_url,omitempty"`
	BackgroundMusicURL *string `json:"background_music_url,omitempty"`
	BackgroundSfxURL   *string `json:"background_sfx_url,omitempty"`

	// 群聊设置
	MaxMembers    int  `json:"max_members" validate:"min=2,max=50"`
	IsPublic      bool `json:"is_public"`
	AllowAIInvite bool `json:"allow_ai_invite"`

	// 剧情设置
	WorldSetting *string `json:"world_setting,omitempty" validate:"omitempty,max=5000"`
	CurrentScene *string `json:"current_scene,omitempty" validate:"omitempty,max=1000"`
	SceneStyle   *string `json:"scene_style,omitempty" validate:"omitempty,max=100"`

	// 初始成员
	InitialCharacters []uuid.UUID `json:"initial_characters,omitempty"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	GroupChatID uuid.UUID `json:"group_chat_id" validate:"required"`
	Content     string    `json:"content" validate:"required,min=1,max=10000"`
	MessageType string    `json:"message_type" validate:"required,oneof=text image audio video"`
	MediaURLs   []string  `json:"media_urls,omitempty"`
}

// 角色相关常量定义
const (
	// 角色可见性
	VisibilityPrivate = "private"
	VisibilityFriends = "friends"
	// VisibilityPublic 在 model.go 中已定义

	// 成员类型
	MemberTypeUser      = "user"
	MemberTypeCharacter = "character"

	// 成员角色
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"

	// 消息类型
	MessageTypeText   = "text"
	MessageTypeImage  = "image"
	MessageTypeAudio  = "audio"
	MessageTypeVideo  = "video"
	MessageTypeSystem = "system"

	// 发送者类型
	SenderTypeUser      = "user"
	SenderTypeCharacter = "character"

	// Moments相关常量
	// 内容类型
	MomentContentTypeText        = "text"
	MomentContentTypeImage       = "image"
	MomentContentTypeVideo       = "video"
	MomentContentTypeTalkingHead = "talking_head"
	MomentContentType3D          = "3d"

	// 可见性
	MomentVisibilityPublic  = "public"
	MomentVisibilityFriends = "friends"
	MomentVisibilityPrivate = "private"

	// 互动类型
	InteractionTypeLike    = "like"
	InteractionTypeComment = "comment"
	InteractionTypeShare   = "share"
)

// 剧情触发相关结构体

// TriggerAnalysisRequest 触发分析请求
type TriggerAnalysisRequest struct {
	GroupChatID uuid.UUID `json:"group_chat_id"`
	Message     string    `json:"message"`
	UserID      uuid.UUID `json:"user_id"`
}

// TriggerAnalysisResult 触发分析结果
type TriggerAnalysisResult struct {
	GroupChatID     uuid.UUID         `json:"group_chat_id"`
	Message         string            `json:"message"`
	CurrentChapter  *StoryChapter     `json:"current_chapter,omitempty"`
	TriggeredEvents []*TriggeredEvent `json:"triggered_events"`
}

// TriggeredEvent 触发事件
type TriggeredEvent struct {
	Type           string        `json:"type"`         // keyword, emotion, semantic, custom
	TriggerType    string        `json:"trigger_type"` // mystery, battle, celebration, etc.
	MatchedKeyword string        `json:"matched_keyword,omitempty"`
	TargetChapter  *StoryChapter `json:"target_chapter,omitempty"`
	Confidence     float64       `json:"confidence"`
}

// StoryTriggerRequest 剧情触发请求
type StoryTriggerRequest struct {
	GroupChatID     uuid.UUID  `json:"group_chat_id"`
	TriggerType     string     `json:"trigger_type"`
	TriggerMessage  string     `json:"trigger_message"`
	TargetChapterID *uuid.UUID `json:"target_chapter_id,omitempty"`
}

// StoryTriggerResult 剧情触发结果
type StoryTriggerResult struct {
	GroupChatID uuid.UUID      `json:"group_chat_id"`
	TriggerType string         `json:"trigger_type"`
	Success     bool           `json:"success"`
	NewChapter  *StoryChapter  `json:"new_chapter,omitempty"`
	Changes     []*StoryChange `json:"changes"`
}

// ChapterSwitchRequest 章节切换请求
type ChapterSwitchRequest struct {
	GroupChatID     uuid.UUID `json:"group_chat_id"`
	TargetChapterID uuid.UUID `json:"target_chapter_id"`
	TriggerMessage  string    `json:"trigger_message"`
}

// ChapterSwitchResult 章节切换结果
type ChapterSwitchResult struct {
	GroupChatID uuid.UUID      `json:"group_chat_id"`
	NewChapter  *StoryChapter  `json:"new_chapter"`
	Success     bool           `json:"success"`
	Changes     []*StoryChange `json:"changes"`
}

// StoryChange 剧情变化
type StoryChange struct {
	Type     string `json:"type"` // background_image, background_music, sound_effects, performance_mode
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// TriggerCondition 触发条件
type TriggerCondition struct {
	Type       string  `json:"type"`       // keyword, regex, semantic, emotion
	Value      string  `json:"value"`      // 触发值
	Confidence float64 `json:"confidence"` // 置信度
}

// StoryChapter 故事章节
type StoryChapter struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	GroupChatID         uuid.UUID `json:"group_chat_id" db:"group_chat_id"`
	Title               string    `json:"title" db:"title"`
	Description         *string   `json:"description,omitempty" db:"description"`
	Order               int       `json:"order" db:"order"`
	BackgroundImageURL  *string   `json:"background_image_url,omitempty" db:"background_image_url"`
	BackgroundMusic     *string   `json:"background_music,omitempty" db:"background_music"`
	SoundEffects        *string   `json:"sound_effects,omitempty" db:"sound_effects"`
	PerformanceSettings *string   `json:"performance_settings,omitempty" db:"performance_settings"`
	TriggerConditions   *string   `json:"trigger_conditions,omitempty" db:"trigger_conditions"`
	Tags                *string   `json:"tags,omitempty" db:"tags"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// 两图系统相关结构体

// TwoImageValidationRequest 两图验证请求
type TwoImageValidationRequest struct {
	CharacterID    uuid.UUID `json:"character_id"`
	BgImageURL     string    `json:"bg_image_url"`     // 背景图URL
	CutoutImageURL string    `json:"cutout_image_url"` // 抠图URL
}

// TwoImageValidationResult 两图验证结果
type TwoImageValidationResult struct {
	CharacterID uuid.UUID `json:"character_id"`
	IsValid     bool      `json:"is_valid"`
	Issues      []string  `json:"issues"`      // 问题列表
	Suggestions []string  `json:"suggestions"` // 建议列表
}

// ChatBackgroundResult 聊天背景结果
type ChatBackgroundResult struct {
	CharacterID        uuid.UUID `json:"character_id"`
	CharacterName      string    `json:"character_name"`
	ChatType           string    `json:"chat_type"` // single, group
	BackgroundImageURL string    `json:"background_image_url"`
	BackgroundType     string    `json:"background_type"` // character_bg, default, custom
	Description        string    `json:"description"`
}

// GroupChatCutoutRequest 群聊抠图请求
type GroupChatCutoutRequest struct {
	GroupChatID  uuid.UUID   `json:"group_chat_id"`
	CharacterIDs []uuid.UUID `json:"character_ids"`
}

// GroupChatCutoutResult 群聊抠图结果
type GroupChatCutoutResult struct {
	GroupChatID         uuid.UUID          `json:"group_chat_id"`
	GlobalBackgroundURL string             `json:"global_background_url"` // 群聊全局背景
	CharacterCutouts    []*CharacterCutout `json:"character_cutouts"`
}

// CharacterCutout 角色抠图配置
type CharacterCutout struct {
	CharacterID       uuid.UUID `json:"character_id"`
	CharacterName     string    `json:"character_name"`
	CutoutImageURL    string    `json:"cutout_image_url"`
	HasCutout         bool      `json:"has_cutout"`
	FallbackText      string    `json:"fallback_text"`      // 无抠图时显示的文字
	IsVisible         bool      `json:"is_visible"`         // 当前是否可见
	AnimationType     string    `json:"animation_type"`     // fade_in, slide_in, etc.
	AnimationDuration int       `json:"animation_duration"` // 动画时长(ms)
}

// CharacterSpeakingRequest 角色发言请求
type CharacterSpeakingRequest struct {
	GroupChatID         uuid.UUID   `json:"group_chat_id"`
	SpeakingCharacterID uuid.UUID   `json:"speaking_character_id"`
	PreviousCharacterID *uuid.UUID  `json:"previous_character_id,omitempty"`
	NextCharacterID     *uuid.UUID  `json:"next_character_id,omitempty"`
	OtherCharacterIDs   []uuid.UUID `json:"other_character_ids"`
}

// CharacterSpeakingResult 角色发言结果
type CharacterSpeakingResult struct {
	GroupChatID         uuid.UUID       `json:"group_chat_id"`
	SpeakingCharacterID uuid.UUID       `json:"speaking_character_id"`
	CutoutChanges       []*CutoutChange `json:"cutout_changes"`
}

// CutoutChange 抠图变化
type CutoutChange struct {
	CharacterID uuid.UUID `json:"character_id"`
	Action      string    `json:"action"` // show, hide, fade, preview
	CutoutURL   string    `json:"cutout_url"`
	Animation   string    `json:"animation"` // fade_in, fade_out, slide_in, etc.
	Duration    int       `json:"duration"`  // 动画时长(ms)
	ZIndex      int       `json:"z_index"`   // 层级
	Opacity     float64   `json:"opacity"`   // 透明度 0.0-1.0
}

// ImageGenerationRequest 图像生成请求
type ImageGenerationRequest struct {
	CharacterID        uuid.UUID `json:"character_id"`
	GenerateBackground bool      `json:"generate_background"`
	GenerateCutout     bool      `json:"generate_cutout"`
	BackgroundStyle    string    `json:"background_style"` // 背景风格
	CutoutStyle        string    `json:"cutout_style"`     // 抠图风格
}

// ImageGenerationResult 图像生成结果
type ImageGenerationResult struct {
	CharacterID     uuid.UUID         `json:"character_id"`
	Success         bool              `json:"success"`
	GeneratedImages []*GeneratedImage `json:"generated_images"`
}

// GeneratedImage 生成的图像
type GeneratedImage struct {
	Type   string `json:"type"`   // background, cutout
	URL    string `json:"url"`    // 生成的图像URL
	Prompt string `json:"prompt"` // 生成提示词
	Style  string `json:"style"`  // 风格
}

// 复杂关系网络相关结构体

// ComplexRelationshipParseRequest 复杂关系解析请求
type ComplexRelationshipParseRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	Description string    `json:"description"` // 用户输入的关系描述
}

// ComplexRelationshipParseResult 复杂关系解析结果
type ComplexRelationshipParseResult struct {
	UserID              uuid.UUID             `json:"user_id"`
	Description         string                `json:"description"`
	ParsedRelationships []*ParsedRelationship `json:"parsed_relationships"`
	Confidence          float64               `json:"confidence"` // 整体解析置信度
}

// ParsedRelationship 解析出的关系
type ParsedRelationship struct {
	CharacterName      string  `json:"character_name"`
	RelationshipType   string  `json:"relationship_type"` // friend, enemy, rival, complex, etc.
	Description        string  `json:"description"`
	Confidence         float64 `json:"confidence"`          // 解析置信度
	Source             string  `json:"source"`              // regex, ai, merged
	EmotionalIntensity float64 `json:"emotional_intensity"` // 情感强度 0-1
	Complexity         float64 `json:"complexity"`          // 关系复杂度 0-1
}

// RelationshipDynamicsRequest 关系动态分析请求
type RelationshipDynamicsRequest struct {
	UserID      uuid.UUID  `json:"user_id"`
	GroupChatID *uuid.UUID `json:"group_chat_id,omitempty"`
}

// RelationshipDynamicsResult 关系动态分析结果
type RelationshipDynamicsResult struct {
	UserID      uuid.UUID               `json:"user_id"`
	GroupChatID *uuid.UUID              `json:"group_chat_id,omitempty"`
	Dynamics    []*RelationshipDynamic  `json:"dynamics"`
	Conflicts   []*RelationshipConflict `json:"conflicts"`
	Suggestions []string                `json:"suggestions"`
}

// RelationshipDynamic 关系动态
type RelationshipDynamic struct {
	RelationshipID   uuid.UUID `json:"relationship_id"`
	CharacterID      uuid.UUID `json:"character_id"`
	RelationshipType string    `json:"relationship_type"`
	Stability        float64   `json:"stability"`       // 稳定性 0-1
	TrendDirection   string    `json:"trend_direction"` // improving, deteriorating, stable, unstable
	InfluenceLevel   float64   `json:"influence_level"` // 影响力水平 0-1
	RiskLevel        float64   `json:"risk_level"`      // 风险水平 0-1
}

// RelationshipConflict 关系冲突
type RelationshipConflict struct {
	ConflictType    string    `json:"conflict_type"` // opposing_relationships, triangular_conflict, etc.
	Relationship1ID uuid.UUID `json:"relationship1_id"`
	Relationship2ID uuid.UUID `json:"relationship2_id"`
	Severity        float64   `json:"severity"` // 冲突严重程度 0-1
	Description     string    `json:"description"`
}

// CharacterRelationshipContextRequest 角色关系上下文请求
type CharacterRelationshipContextRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	CharacterID uuid.UUID `json:"character_id"`
	GroupChatID uuid.UUID `json:"group_chat_id"`
}

// CharacterRelationshipContextResult 角色关系上下文结果
type CharacterRelationshipContextResult struct {
	UserID            uuid.UUID                      `json:"user_id"`
	CharacterID       uuid.UUID                      `json:"character_id"`
	GroupChatID       uuid.UUID                      `json:"group_chat_id"`
	DirectRelation    *RelationshipContext           `json:"direct_relation,omitempty"`
	IndirectRelations []*IndirectRelationshipContext `json:"indirect_relations"`
	ContextSummary    string                         `json:"context_summary"`
}

// RelationshipContext 关系上下文
type RelationshipContext struct {
	RelationshipType   string   `json:"relationship_type"`
	Description        string   `json:"description"`
	EmotionalIntensity float64  `json:"emotional_intensity"`
	Complexity         float64  `json:"complexity"`
	RecentInteractions []string `json:"recent_interactions"`
}

// IndirectRelationshipContext 间接关系上下文
type IndirectRelationshipContext struct {
	ThroughCharacterID   uuid.UUID `json:"through_character_id"`
	ThroughCharacterName string    `json:"through_character_name"`
	RelationshipChain    []string  `json:"relationship_chain"` // ["我->朋友->小米", "小米->闺蜜->小明"]
	InfluenceLevel       float64   `json:"influence_level"`
}

// RelationshipStrategyRequest 关系策略请求
type RelationshipStrategyRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	CharacterID uuid.UUID `json:"character_id"`
	GroupChatID uuid.UUID `json:"group_chat_id"`
	Scenario    string    `json:"scenario"` // 当前场景描述
}

// RelationshipStrategyResult 关系策略结果
type RelationshipStrategyResult struct {
	UserID                uuid.UUID                   `json:"user_id"`
	CharacterID           uuid.UUID                   `json:"character_id"`
	GroupChatID           uuid.UUID                   `json:"group_chat_id"`
	RecommendedStrategy   *RelationshipStrategy       `json:"recommended_strategy"`
	AlternativeStrategies []*RelationshipStrategy     `json:"alternative_strategies"`
	RiskAssessment        *RelationshipRiskAssessment `json:"risk_assessment"`
}

// RelationshipStrategy 关系策略
type RelationshipStrategy struct {
	StrategyType    string   `json:"strategy_type"` // direct, indirect, avoidance, mediation
	Description     string   `json:"description"`
	Actions         []string `json:"actions"`          // 具体行动建议
	ExpectedOutcome string   `json:"expected_outcome"` // 预期结果
	Confidence      float64  `json:"confidence"`       // 策略置信度
}

// RelationshipRiskAssessment 关系风险评估
type RelationshipRiskAssessment struct {
	OverallRisk     float64  `json:"overall_risk"`     // 整体风险 0-1
	RiskFactors     []string `json:"risk_factors"`     // 风险因素
	MitigationSteps []string `json:"mitigation_steps"` // 缓解措施
	WarningSignals  []string `json:"warning_signals"`  // 警告信号
}

// RelationshipConflictResult 关系冲突检测结果
type RelationshipConflictResult struct {
	GroupChatID     uuid.UUID                 `json:"group_chat_id"`
	ConflictLevel   string                    `json:"conflict_level"` // low, medium, high, critical
	ActiveConflicts []*RelationshipConflict   `json:"active_conflicts"`
	RiskPredictions []*ConflictRiskPrediction `json:"risk_predictions"`
	Recommendations []string                  `json:"recommendations"`
}

// ConflictRiskPrediction 冲突风险预测
type ConflictRiskPrediction struct {
	Character1ID     uuid.UUID `json:"character1_id"`
	Character2ID     uuid.UUID `json:"character2_id"`
	ConflictRisk     float64   `json:"conflict_risk"`     // 冲突风险 0-1
	TriggerScenarios []string  `json:"trigger_scenarios"` // 可能触发冲突的场景
	PreventionTips   []string  `json:"prevention_tips"`   // 预防建议
}
