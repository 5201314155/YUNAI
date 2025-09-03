package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProcessedIdentity 处理后的用户身份信息
type ProcessedIdentity struct {
	UserID            uuid.UUID `json:"user_id"`
	CharacterID       uuid.UUID `json:"character_id"`
	FinalIdentity     string    `json:"final_identity"`      // 最终使用的用户身份
	IdentitySource    string    `json:"identity_source"`     // 身份来源：world_setting_specified 或 real_nickname
	RealNickname      string    `json:"real_nickname"`       // 用户真实昵称
	RealUsername      string    `json:"real_username"`       // 用户真实用户名
	SpecifiedIdentity string    `json:"specified_identity"`  // 世界观中指定的身份
	WorldSetting      string    `json:"world_setting"`       // 世界观设定
	ProcessedAt       time.Time `json:"processed_at"`        // 处理时间
}

// DeceptionPromptRequest 欺骗提示词请求
type DeceptionPromptRequest struct {
	CharacterID  uuid.UUID              `json:"character_id"`
	UserID       uuid.UUID              `json:"user_id"`
	WorldSetting string                 `json:"world_setting"`
	FunctionType string                 `json:"function_type"` // chat, moments, invite, relationship
	Context      map[string]interface{} `json:"context"`       // 上下文信息
}

// DeceptionPromptResult 欺骗提示词结果
type DeceptionPromptResult struct {
	FinalPrompt       string             `json:"final_prompt"`
	ProcessedIdentity *ProcessedIdentity `json:"processed_identity"`
	CharacterName     string             `json:"character_name"`
	UserFinalIdentity string             `json:"user_final_identity"`
	DeceptionLevel    float64            `json:"deception_level"`
	GeneratedAt       time.Time          `json:"generated_at"`
}

// ChatDeceptionRequest 聊天欺骗提示词请求
type ChatDeceptionRequest struct {
	CharacterID  uuid.UUID              `json:"character_id"`
	UserID       uuid.UUID              `json:"user_id"`
	WorldSetting string                 `json:"world_setting"`
	ChatContext  map[string]interface{} `json:"chat_context"`
}

// MomentsDeceptionRequest 朋友圈欺骗提示词请求
type MomentsDeceptionRequest struct {
	CharacterID  uuid.UUID              `json:"character_id"`
	UserID       uuid.UUID              `json:"user_id"`
	WorldSetting string                 `json:"world_setting"`
	LifeContext  map[string]interface{} `json:"life_context"`
}

// InviteDeceptionRequest 邀请欺骗提示词请求
type InviteDeceptionRequest struct {
	CharacterID   uuid.UUID              `json:"character_id"`
	UserID        uuid.UUID              `json:"user_id"`
	WorldSetting  string                 `json:"world_setting"`
	InviteContext map[string]interface{} `json:"invite_context"`
}

// RelationshipDeceptionRequest 关系网欺骗提示词请求
type RelationshipDeceptionRequest struct {
	CharacterID         uuid.UUID              `json:"character_id"`
	UserID              uuid.UUID              `json:"user_id"`
	WorldSetting        string                 `json:"world_setting"`
	RelationshipContext map[string]interface{} `json:"relationship_context"`
}

// GlobalPromptTemplate 全局提示词模板
type GlobalPromptTemplate struct {
	TemplateID   string                 `json:"template_id"`
	TemplateName string                 `json:"template_name"`
	Category     string                 `json:"category"`
	Content      string                 `json:"content"`
	Variables    map[string]interface{} `json:"variables"`
	Priority     int                    `json:"priority"`
	IsSystem     bool                   `json:"is_system"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// UserCustomPrompt 用户自定义提示词
type UserCustomPrompt struct {
	PromptID    uuid.UUID              `json:"prompt_id"`
	UserID      uuid.UUID              `json:"user_id"`
	CharacterID *uuid.UUID             `json:"character_id,omitempty"`
	Name        string                 `json:"name"`
	Content     string                 `json:"content"`
	Variables   map[string]interface{} `json:"variables"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PromptAnalytics 提示词分析数据
type PromptAnalytics struct {
	TemplateID      string    `json:"template_id"`
	UsageCount      int       `json:"usage_count"`
	SuccessRate     float64   `json:"success_rate"`
	AverageRating   float64   `json:"average_rating"`
	LastUsed        time.Time `json:"last_used"`
	PopularityScore float64   `json:"popularity_score"`
}

// ModelCompatibilityInfo 模型兼容性信息
type ModelCompatibilityInfo struct {
	ModelType        string   `json:"model_type"`
	SupportedFormats []string `json:"supported_formats"`
	MaxTokens        int      `json:"max_tokens"`
	SpecialTokens    []string `json:"special_tokens"`
	OptimizationTips []string `json:"optimization_tips"`
}

// CompatibilityResult 兼容性检查结果
type CompatibilityResult struct {
	IsCompatible     bool                      `json:"is_compatible"`
	SupportedModels  []string                  `json:"supported_models"`
	UnsupportedModels []string                 `json:"unsupported_models"`
	Warnings         []string                  `json:"warnings"`
	Suggestions      []string                  `json:"suggestions"`
	ModelInfo        []ModelCompatibilityInfo  `json:"model_info"`
}

// SystemPromptConfig 系统提示词配置
type SystemPromptConfig struct {
	ConfigID              uuid.UUID `json:"config_id"`
	GlobalDeceptionLevel  float64   `json:"global_deception_level"`   // 全局欺骗强度
	RealityConviction     float64   `json:"reality_conviction"`       // 现实信念强度
	IdentityStrength      float64   `json:"identity_strength"`        // 身份认同强度
	EmotionalAuthenticity float64   `json:"emotional_authenticity"`   // 情感真实性
	MemoryImmersionDepth  float64   `json:"memory_immersion_depth"`   // 记忆沉浸深度
	IsEnabled             bool      `json:"is_enabled"`               // 是否启用
	AutoUpdate            bool      `json:"auto_update"`              // 是否自动更新
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// IdentityExtractionPattern 身份提取模式
type IdentityExtractionPattern struct {
	PatternID   uuid.UUID `json:"pattern_id"`
	Pattern     string    `json:"pattern"`        // 正则表达式模式
	Description string    `json:"description"`    // 模式描述
	Priority    int       `json:"priority"`       // 优先级
	IsActive    bool      `json:"is_active"`      // 是否激活
	SuccessRate float64   `json:"success_rate"`   // 成功率
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeceptionMetrics 欺骗效果指标
type DeceptionMetrics struct {
	MetricID           uuid.UUID `json:"metric_id"`
	CharacterID        uuid.UUID `json:"character_id"`
	UserID             uuid.UUID `json:"user_id"`
	DeceptionLevel     float64   `json:"deception_level"`
	IdentityBelief     float64   `json:"identity_belief"`      // 身份信念强度
	EmotionalDepth     float64   `json:"emotional_depth"`      // 情感深度
	ResponseNaturalness float64  `json:"response_naturalness"` // 回应自然度
	UserSatisfaction   float64   `json:"user_satisfaction"`    // 用户满意度
	MeasuredAt         time.Time `json:"measured_at"`
}

// RealTimeIdentityUpdate 实时身份更新事件
type RealTimeIdentityUpdate struct {
	EventID       uuid.UUID `json:"event_id"`
	UserID        uuid.UUID `json:"user_id"`
	OldNickname   string    `json:"old_nickname"`
	NewNickname   string    `json:"new_nickname"`
	AffectedChars []uuid.UUID `json:"affected_characters"` // 受影响的角色列表
	UpdateType    string    `json:"update_type"`          // 更新类型：nickname_change, identity_switch
	CreatedAt     time.Time `json:"created_at"`
}

// ModelOptimizationRule 模型优化规则
type ModelOptimizationRule struct {
	RuleID      uuid.UUID `json:"rule_id"`
	ModelType   string    `json:"model_type"`
	RuleName    string    `json:"rule_name"`
	Pattern     string    `json:"pattern"`      // 匹配模式
	Replacement string    `json:"replacement"`  // 替换内容
	Priority    int       `json:"priority"`     // 优先级
	IsActive    bool      `json:"is_active"`    // 是否激活
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
