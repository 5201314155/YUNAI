package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// RelationshipType 关系类型定义
type RelationshipType struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      *uuid.UUID `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	DisplayName string     `json:"display_name" db:"display_name"`
	Description *string    `json:"description" db:"description"`
	Category    string     `json:"category" db:"category"`

	// 关系属性
	IsMutual        bool    `json:"is_mutual" db:"is_mutual"`
	DefaultStrength float64 `json:"default_strength" db:"default_strength"`

	// 情感维度默认值
	DefaultTrust     float64 `json:"default_trust" db:"default_trust"`
	DefaultAffection float64 `json:"default_affection" db:"default_affection"`
	DefaultRespect   float64 `json:"default_respect" db:"default_respect"`
	DefaultIntimacy  float64 `json:"default_intimacy" db:"default_intimacy"`

	// 行为设置
	DefaultTone         string `json:"default_tone" db:"default_tone"`
	DefaultAddressStyle string `json:"default_address_style" db:"default_address_style"`

	// 使用统计
	UsageCount int  `json:"usage_count" db:"usage_count"`
	IsFeatured bool `json:"is_featured" db:"is_featured"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CharacterRelationshipEnhanced 增强的角色关系
type CharacterRelationshipEnhanced struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	SourceCharacterID  uuid.UUID  `json:"source_character_id" db:"source_character_id"`
	TargetCharacterID  uuid.UUID  `json:"target_character_id" db:"target_character_id"`
	RelationshipTypeID *uuid.UUID `json:"relationship_type_id" db:"relationship_type_id"`
	CustomTypeName     *string    `json:"custom_type_name" db:"custom_type_name"`

	// 关系强度和状态
	Strength float64 `json:"strength" db:"strength"`
	Status   string  `json:"status" db:"status"`

	// 多维情感 (0-1 scale)
	Trust      float64 `json:"trust" db:"trust"`
	Affection  float64 `json:"affection" db:"affection"`
	Respect    float64 `json:"respect" db:"respect"`
	Intimacy   float64 `json:"intimacy" db:"intimacy"`
	Jealousy   float64 `json:"jealousy" db:"jealousy"`
	Dependency float64 `json:"dependency" db:"dependency"`

	// 沟通风格
	AddressName    *string `json:"address_name" db:"address_name"`
	Tone           string  `json:"tone" db:"tone"`
	FormalityLevel float64 `json:"formality_level" db:"formality_level"`

	// 行为修饰符
	SpeakingFrequency float64 `json:"speaking_frequency" db:"speaking_frequency"`
	InitiativeLevel   float64 `json:"initiative_level" db:"initiative_level"`
	ConflictTendency  float64 `json:"conflict_tendency" db:"conflict_tendency"`

	// 触发条件
	TriggerKeywords  pq.StringArray `json:"trigger_keywords" db:"trigger_keywords"`
	TriggerEmotions  pq.StringArray `json:"trigger_emotions" db:"trigger_emotions"`
	TriggerScenarios pq.StringArray `json:"trigger_scenarios" db:"trigger_scenarios"`

	// 记忆和历史
	RelationshipHistory json.RawMessage `json:"relationship_history" db:"relationship_history"`
	SharedMemories      pq.StringArray  `json:"shared_memories" db:"shared_memories"`
	PrivateNotes        *string         `json:"private_notes" db:"private_notes"`

	// 约束和边界
	ForbiddenTopics   pq.StringArray  `json:"forbidden_topics" db:"forbidden_topics"`
	PreferredTopics   pq.StringArray  `json:"preferred_topics" db:"preferred_topics"`
	InteractionLimits json.RawMessage `json:"interaction_limits" db:"interaction_limits"`

	// 时间戳和跟踪
	LastInteractionAt         *time.Time `json:"last_interaction_at" db:"last_interaction_at"`
	RelationshipEstablishedAt time.Time  `json:"relationship_established_at" db:"relationship_established_at"`
	CreatedAt                 time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at" db:"updated_at"`
}

// RelationshipEvent 关系事件
type RelationshipEvent struct {
	ID             uuid.UUID `json:"id" db:"id"`
	RelationshipID uuid.UUID `json:"relationship_id" db:"relationship_id"`

	// 事件信息
	EventType        string `json:"event_type" db:"event_type"`
	EventDescription string `json:"event_description" db:"event_description"`

	// 变化
	StrengthChange *float64        `json:"strength_change" db:"strength_change"`
	EmotionChanges json.RawMessage `json:"emotion_changes" db:"emotion_changes"`

	// 上下文
	TriggeredByMessageID *uuid.UUID     `json:"triggered_by_message_id" db:"triggered_by_message_id"`
	TriggeredByScenario  *string        `json:"triggered_by_scenario" db:"triggered_by_scenario"`
	TriggerKeywords      pq.StringArray `json:"trigger_keywords" db:"trigger_keywords"`

	// 元数据
	Automatic     bool       `json:"automatic" db:"automatic"`
	CreatedByUser *uuid.UUID `json:"created_by_user_id" db:"created_by_user_id"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MemoryFragment 记忆片段
type MemoryFragment struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`

	// 记忆内容
	Content    string  `json:"content" db:"content"`
	Summary    *string `json:"summary" db:"summary"`
	MemoryType string  `json:"memory_type" db:"memory_type"`

	// 重要性和相关性
	ImportanceScore    float64 `json:"importance_score" db:"importance_score"`
	EmotionalIntensity float64 `json:"emotional_intensity" db:"emotional_intensity"`

	// 关联
	RelatedCharacters pq.StringArray `json:"related_characters" db:"related_characters"`
	RelatedTopics     pq.StringArray `json:"related_topics" db:"related_topics"`
	RelatedEmotions   pq.StringArray `json:"related_emotions" db:"related_emotions"`

	// 上下文
	SourceMessageID   *uuid.UUID      `json:"source_message_id" db:"source_message_id"`
	SourceGroupChatID *uuid.UUID      `json:"source_group_chat_id" db:"source_group_chat_id"`
	ContextMetadata   json.RawMessage `json:"context_metadata" db:"context_metadata"`

	// 向量嵌入 (存储为JSON)
	Embedding json.RawMessage `json:"embedding" db:"embedding"`

	// 生命周期
	AccessCount    int        `json:"access_count" db:"access_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at" db:"last_accessed_at"`
	ExpiresAt      *time.Time `json:"expires_at" db:"expires_at"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ConversationContext 对话上下文
type ConversationContext struct {
	ID          uuid.UUID `json:"id" db:"id"`
	GroupChatID uuid.UUID `json:"group_chat_id" db:"group_chat_id"`

	// 上下文信息
	ContextName  *string        `json:"context_name" db:"context_name"`
	CurrentScene *string        `json:"current_scene" db:"current_scene"`
	Mood         *string        `json:"mood" db:"mood"`
	ActiveThemes pq.StringArray `json:"active_themes" db:"active_themes"`

	// 参与者和状态
	ActiveCharacters pq.StringArray `json:"active_characters" db:"active_characters"`
	SpeakingQueue    pq.StringArray `json:"speaking_queue" db:"speaking_queue"`
	CurrentSpeakerID *uuid.UUID     `json:"current_speaker_id" db:"current_speaker_id"`
	NextSpeakerID    *uuid.UUID     `json:"next_speaker_id" db:"next_speaker_id"`

	// 对话流程
	TurnCount      int       `json:"turn_count" db:"turn_count"`
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`

	// 记忆上下文
	RecentMemoryIDs pq.StringArray `json:"recent_memory_ids" db:"recent_memory_ids"`
	ContextSummary  *string        `json:"context_summary" db:"context_summary"`

	// 编排设置
	OrchestrationMode string  `json:"orchestration_mode" db:"orchestration_mode"`
	InterventionLevel float64 `json:"intervention_level" db:"intervention_level"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TriggerRule 触发规则
type TriggerRule struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      *uuid.UUID `json:"user_id" db:"user_id"`
	GroupChatID *uuid.UUID `json:"group_chat_id" db:"group_chat_id"`

	// 规则信息
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	RuleType    string  `json:"rule_type" db:"rule_type"`

	// 触发条件和动作
	TriggerConditions json.RawMessage `json:"trigger_conditions" db:"trigger_conditions"`
	Actions           json.RawMessage `json:"actions" db:"actions"`

	// 约束
	CooldownMinutes   int `json:"cooldown_minutes" db:"cooldown_minutes"`
	MaxTriggersPerDay int `json:"max_triggers_per_day" db:"max_triggers_per_day"`
	Priority          int `json:"priority" db:"priority"`

	// 状态
	IsActive        bool       `json:"is_active" db:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at" db:"last_triggered_at"`
	TriggerCount    int        `json:"trigger_count" db:"trigger_count"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// 请求和响应结构

// CreateRelationshipTypeRequest 创建关系类型请求
type CreateRelationshipTypeRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	DisplayName string  `json:"display_name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`

	IsMutual        bool    `json:"is_mutual"`
	DefaultStrength float64 `json:"default_strength" validate:"min=0,max=1"`

	DefaultTrust     float64 `json:"default_trust" validate:"min=0,max=1"`
	DefaultAffection float64 `json:"default_affection" validate:"min=0,max=1"`
	DefaultRespect   float64 `json:"default_respect" validate:"min=0,max=1"`
	DefaultIntimacy  float64 `json:"default_intimacy" validate:"min=0,max=1"`

	DefaultTone         string `json:"default_tone" validate:"required"`
	DefaultAddressStyle string `json:"default_address_style" validate:"required"`
}

// CreateCharacterRelationshipRequest 创建角色关系请求
type CreateCharacterRelationshipRequest struct {
	TargetCharacterID  uuid.UUID  `json:"target_character_id" validate:"required"`
	RelationshipTypeID *uuid.UUID `json:"relationship_type_id,omitempty"`
	CustomTypeName     *string    `json:"custom_type_name,omitempty"`

	Strength float64 `json:"strength" validate:"min=0,max=1"`

	Trust      float64 `json:"trust" validate:"min=0,max=1"`
	Affection  float64 `json:"affection" validate:"min=0,max=1"`
	Respect    float64 `json:"respect" validate:"min=0,max=1"`
	Intimacy   float64 `json:"intimacy" validate:"min=0,max=1"`
	Jealousy   float64 `json:"jealousy" validate:"min=0,max=1"`
	Dependency float64 `json:"dependency" validate:"min=0,max=1"`

	AddressName    *string `json:"address_name,omitempty"`
	Tone           string  `json:"tone"`
	FormalityLevel float64 `json:"formality_level" validate:"min=0,max=1"`

	TriggerKeywords  []string `json:"trigger_keywords,omitempty"`
	TriggerEmotions  []string `json:"trigger_emotions,omitempty"`
	TriggerScenarios []string `json:"trigger_scenarios,omitempty"`

	SharedMemories  []string `json:"shared_memories,omitempty"`
	PrivateNotes    *string  `json:"private_notes,omitempty"`
	ForbiddenTopics []string `json:"forbidden_topics,omitempty"`
	PreferredTopics []string `json:"preferred_topics,omitempty"`
}

// CreateMemoryFragmentRequest 创建记忆片段请求
type CreateMemoryFragmentRequest struct {
	CharacterID uuid.UUID `json:"character_id" validate:"required"`
	Content     string    `json:"content" validate:"required,min=1,max=5000"`
	MemoryType  string    `json:"memory_type" validate:"required,oneof=conversation event fact emotion relationship"`

	ImportanceScore    float64 `json:"importance_score" validate:"min=0,max=1"`
	EmotionalIntensity float64 `json:"emotional_intensity" validate:"min=0,max=1"`

	RelatedCharacters []string `json:"related_characters,omitempty"`
	RelatedTopics     []string `json:"related_topics,omitempty"`
	RelatedEmotions   []string `json:"related_emotions,omitempty"`

	SourceMessageID   *uuid.UUID `json:"source_message_id,omitempty"`
	SourceGroupChatID *uuid.UUID `json:"source_group_chat_id,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// CreateTriggerRuleRequest 创建触发规则请求
type CreateTriggerRuleRequest struct {
	GroupChatID *uuid.UUID `json:"group_chat_id,omitempty"`
	Name        string     `json:"name" validate:"required,min=1,max=200"`
	Description *string    `json:"description,omitempty"`
	RuleType    string     `json:"rule_type" validate:"required,oneof=keyword emotion relationship time event"`

	TriggerConditions json.RawMessage `json:"trigger_conditions" validate:"required"`
	Actions           json.RawMessage `json:"actions" validate:"required"`

	CooldownMinutes   int `json:"cooldown_minutes" validate:"min=0"`
	MaxTriggersPerDay int `json:"max_triggers_per_day" validate:"min=0"`
	Priority          int `json:"priority" validate:"min=0,max=100"`
}

// RelationshipResponse 关系响应
type RelationshipResponse struct {
	*CharacterRelationshipEnhanced
	RelationshipType *RelationshipType   `json:"relationship_type,omitempty"`
	SourceCharacter  *Character          `json:"source_character,omitempty"`
	TargetCharacter  *Character          `json:"target_character,omitempty"`
	RecentEvents     []RelationshipEvent `json:"recent_events,omitempty"`
	CanEdit          bool                `json:"can_edit"`
}

// MemorySearchRequest 记忆搜索请求
type MemorySearchRequest struct {
	CharacterID *uuid.UUID `json:"character_id,omitempty"`
	Query       string     `json:"query" validate:"required,min=1"`
	MemoryTypes []string   `json:"memory_types,omitempty"`
	Limit       int        `json:"limit" validate:"min=1,max=50"`
	MinScore    float64    `json:"min_score" validate:"min=0,max=1"`
}

// MemorySearchResponse 记忆搜索响应
type MemorySearchResponse struct {
	Memories []MemoryFragment `json:"memories"`
	Query    string           `json:"query"`
	Total    int              `json:"total"`
}

// 常量定义
const (
	// 关系类型分类
	RelationshipCategorySystem = "system"
	RelationshipCategoryPublic = "public"
	RelationshipCategoryCustom = "custom"

	// 关系状态
	RelationshipStatusActive  = "active"
	RelationshipStatusDormant = "dormant"
	RelationshipStatusBroken  = "broken"
	RelationshipStatusHidden  = "hidden"

	// 记忆类型
	MemoryTypeConversation = "conversation"
	MemoryTypeEvent        = "event"
	MemoryTypeFact         = "fact"
	MemoryTypeEmotion      = "emotion"
	MemoryTypeRelationship = "relationship"

	// 触发规则类型
	TriggerRuleTypeKeyword      = "keyword"
	TriggerRuleTypeEmotion      = "emotion"
	TriggerRuleTypeRelationship = "relationship"
	TriggerRuleTypeTime         = "time"
	TriggerRuleTypeEvent        = "event"

	// 编排模式
	OrchestrationModeAuto   = "auto"
	OrchestrationModeManual = "manual"
	OrchestrationModeGuided = "guided"

	// 语调类型
	ToneNeutral  = "neutral"
	ToneFormal   = "formal"
	ToneCasual   = "casual"
	ToneIntimate = "intimate"
	ToneHostile  = "hostile"
	TonePlayful  = "playful"
	ToneCaring   = "caring"

	// 称呼风格
	AddressStyleName     = "name"
	AddressStyleTitle    = "title"
	AddressStyleNickname = "nickname"
	AddressStylePetName  = "pet_name"
)

// 编排相关结构

// ProcessMessageRequest 处理消息请求
type ProcessMessageRequest struct {
	GroupChatID uuid.UUID `json:"group_chat_id" validate:"required"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Content     string    `json:"content" validate:"required"`
	MessageType string    `json:"message_type" validate:"required"`
}

// ProcessMessageResponse 处理消息响应
type ProcessMessageResponse struct {
	ProcessedAt    time.Time            `json:"processed_at"`
	TriggersCount  int                  `json:"triggers_count"`
	NextSpeakerID  *uuid.UUID           `json:"next_speaker_id,omitempty"`
	Responses      []*GeneratedResponse `json:"responses,omitempty"`
	ContextUpdated bool                 `json:"context_updated"`
}

// GeneratedResponse 生成的响应
type GeneratedResponse struct {
	CharacterID   uuid.UUID `json:"character_id"`
	Content       string    `json:"content"`
	ModelUsed     string    `json:"model_used"`
	TokensUsed    int       `json:"tokens_used"`
	Cost          float64   `json:"cost"`
	GeneratedAt   time.Time `json:"generated_at"`
	ContextLength int       `json:"context_length"`
}

// TriggerAction 触发动作
type TriggerAction struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// PerformanceScenario 演出场景
type PerformanceScenario struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Acts        []Act     `json:"acts"`
}

// Act 演出幕
type Act struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    int    `json:"duration"` // 分钟
}

// PerformanceStatus 演出状态
type PerformanceStatus struct {
	IsActive   bool       `json:"is_active"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	CurrentAct int        `json:"current_act"`
	TotalActs  int        `json:"total_acts"`
	ScenarioID *uuid.UUID `json:"scenario_id,omitempty"`
}

// AI智能邀请相关结构

// InvitationAnalysis 邀请意图分析
type InvitationAnalysis struct {
	OriginalMessage   string   `json:"original_message"`
	Intent            string   `json:"intent"` // invite, none
	Confidence        float64  `json:"confidence"`
	Keywords          []string `json:"keywords"`
	RelationshipHints []string `json:"relationship_hints"`
	EmotionalContext  string   `json:"emotional_context"`
	Urgency           string   `json:"urgency"`
	FuzzyMatches      []string `json:"fuzzy_matches"`
}

// InvitationCandidate 邀请候选者
type InvitationCandidate struct {
	Character     *CharacterResponse      `json:"character"`
	MatchScore    float64                 `json:"match_score"`
	MatchReasons  []string                `json:"match_reasons"`
	Relationships []*RelationshipResponse `json:"relationships"`
}

// InvitationSuggestion 邀请建议
type InvitationSuggestion struct {
	Character        *CharacterResponse `json:"character"`
	MatchScore       float64            `json:"match_score"`
	InviteReason     string             `json:"invite_reason"`
	ExpectedReaction string             `json:"expected_reaction"`
	Confidence       float64            `json:"confidence"`
	Priority         int                `json:"priority"` // 1=高, 2=中, 3=低
}

// SmartInvitationResult 智能邀请结果
type SmartInvitationResult struct {
	ProcessedAt     time.Time               `json:"processed_at"`
	OriginalMessage string                  `json:"original_message"`
	Success         bool                    `json:"success"`
	Suggestions     []*InvitationSuggestion `json:"suggestions"`
	Message         string                  `json:"message"`
}

// 邀请响应相关结构

// InvitationResponseRequest 邀请响应请求
type InvitationResponseRequest struct {
	CharacterID uuid.UUID `json:"character_id" validate:"required"`
	GroupChatID uuid.UUID `json:"group_chat_id" validate:"required"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Action      string    `json:"action" validate:"required"` // accept, reject
	Reason      string    `json:"reason,omitempty"`           // 拒绝原因
}

// InvitationResponseResult 邀请响应结果
type InvitationResponseResult struct {
	ProcessedAt       time.Time             `json:"processed_at"`
	Success           bool                  `json:"success"`
	CharacterID       uuid.UUID             `json:"character_id"`
	Action            string                `json:"action"`
	Message           string                `json:"message"`
	CharacterReaction *CharacterReaction    `json:"character_reaction,omitempty"`
	PrivateMessage    *PrivateMessageResult `json:"private_message,omitempty"`
}

// InvitationPopup 邀请弹窗
type InvitationPopup struct {
	CharacterID      uuid.UUID `json:"character_id"`
	CharacterName    string    `json:"character_name"`
	CharacterAvatar  *string   `json:"character_avatar,omitempty"`
	GroupChatID      uuid.UUID `json:"group_chat_id"`
	GroupChatName    string    `json:"group_chat_name"`
	InviteReason     string    `json:"invite_reason"`
	RelationshipDesc string    `json:"relationship_desc"`
	ExpectedReaction string    `json:"expected_reaction"`
	MatchScore       float64   `json:"match_score"`
	Priority         int       `json:"priority"`
	PopupTitle       string    `json:"popup_title"`
	PopupMessage     string    `json:"popup_message"`
	AcceptButtonText string    `json:"accept_button_text"`
	RejectButtonText string    `json:"reject_button_text"`
	CreatedAt        time.Time `json:"created_at"`
}

// AcceptInvitationResult 接受邀请结果
type AcceptInvitationResult struct {
	Success           bool               `json:"success"`
	CharacterID       uuid.UUID          `json:"character_id"`
	GroupChatID       uuid.UUID          `json:"group_chat_id"`
	Message           string             `json:"message"`
	CharacterReaction *CharacterReaction `json:"character_reaction,omitempty"`
}

// RejectInvitationResult 拒绝邀请结果
type RejectInvitationResult struct {
	Success           bool                  `json:"success"`
	CharacterID       uuid.UUID             `json:"character_id"`
	GroupChatID       uuid.UUID             `json:"group_chat_id"`
	Message           string                `json:"message"`
	CharacterReaction *CharacterReaction    `json:"character_reaction,omitempty"`
	PrivateMessage    *PrivateMessageResult `json:"private_message,omitempty"`
}

// CharacterReaction 角色反应
type CharacterReaction struct {
	CharacterID   uuid.UUID `json:"character_id"`
	CharacterName string    `json:"character_name"`
	Action        string    `json:"action"`
	Content       string    `json:"content"`
	Emotion       string    `json:"emotion"`
	Intensity     float64   `json:"intensity"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// PrivateMessageResult 私信结果
type PrivateMessageResult struct {
	CharacterID   uuid.UUID `json:"character_id"`
	CharacterName string    `json:"character_name"`
	UserID        uuid.UUID `json:"user_id"`
	Content       string    `json:"content"`
	MessageType   string    `json:"message_type"`
	SentAt        time.Time `json:"sent_at"`
	Success       bool      `json:"success"`
}

// 基于关系网的邀请相关结构

// ChatAnalysisRequest 聊天分析请求
type ChatAnalysisRequest struct {
	UserID      uuid.UUID  `json:"user_id" validate:"required"`
	GroupChatID uuid.UUID  `json:"group_chat_id" validate:"required"`
	ChatMessage string     `json:"chat_message" validate:"required"`
	SenderID    *uuid.UUID `json:"sender_id,omitempty"`    // 发送者ID（可能是角色）
	MessageType string     `json:"message_type,omitempty"` // 消息类型
}

// RelationshipBasedInvitationResult 基于关系网的邀请结果
type RelationshipBasedInvitationResult struct {
	AnalyzedAt  time.Time                           `json:"analyzed_at"`
	ChatMessage string                              `json:"chat_message"`
	UserID      uuid.UUID                           `json:"user_id"`
	GroupChatID uuid.UUID                           `json:"group_chat_id"`
	Success     bool                                `json:"success"`
	Message     string                              `json:"message"`
	Suggestions []*RelationshipInvitationSuggestion `json:"suggestions"`
}

// UserRelationshipMap 用户关系网络图谱
type UserRelationshipMap struct {
	UserID             uuid.UUID                     `json:"user_id"`
	RelationshipGroups map[string]*RelationshipGroup `json:"relationship_groups"`
	TotalCharacters    int                           `json:"total_characters"`
}

// RelationshipGroup 关系组
type RelationshipGroup struct {
	RelationshipType *RelationshipType        `json:"relationship_type"`
	Characters       []*RelationshipCharacter `json:"characters"`
	Description      *string                  `json:"description,omitempty"`
	CustomPrompt     *string                  `json:"custom_prompt,omitempty"`
}

// RelationshipCharacter 关系中的角色
type RelationshipCharacter struct {
	Character    *CharacterResponse             `json:"character"`
	Relationship *CharacterRelationshipEnhanced `json:"relationship"`
	MatchScore   float64                        `json:"match_score"`
}

// RelationshipMatchResult 关系匹配结果
type RelationshipMatchResult struct {
	Character        *CharacterResponse             `json:"character"`
	Relationship     *CharacterRelationshipEnhanced `json:"relationship"`
	RelationshipType *RelationshipType              `json:"relationship_type"`
	MatchScore       float64                        `json:"match_score"`
	MatchReason      string                         `json:"match_reason"`
}

// RelationshipInvitationSuggestion 基于关系的邀请建议
type RelationshipInvitationSuggestion struct {
	Character        *CharacterResponse             `json:"character"`
	Relationship     *CharacterRelationshipEnhanced `json:"relationship"`
	RelationshipType *RelationshipType              `json:"relationship_type"`
	MatchScore       float64                        `json:"match_score"`
	InviteReason     string                         `json:"invite_reason"`
	ExpectedReaction string                         `json:"expected_reaction"`
	RelationshipDesc string                         `json:"relationship_desc"`
	Priority         int                            `json:"priority"`
	CreatedAt        time.Time                      `json:"created_at"`
}

// 类型别名，用于向后兼容
type Relationship = CharacterRelationshipEnhanced

// DialogueLine 对话行
type DialogueLine struct {
	ID          uuid.UUID `json:"id"`
	CharacterID uuid.UUID `json:"character_id"`
	Content     string    `json:"content"`
	Emotion     *string   `json:"emotion,omitempty"`
	Tone        *string   `json:"tone,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// RelationshipAnalysis 关系分析结果
type RelationshipAnalysis struct {
	ID                uuid.UUID                        `json:"id"`
	AnalyzedAt        time.Time                        `json:"analyzed_at"`
	Relationships     []*CharacterRelationshipEnhanced `json:"relationships"`
	Dynamics          []*RelationshipDynamic           `json:"dynamics"`
	Conflicts         []*RelationshipConflict          `json:"conflicts"`
	Suggestions       []string                         `json:"suggestions"`
	OverallStability  float64                          `json:"overall_stability"`
	NetworkComplexity float64                          `json:"network_complexity"`
}
