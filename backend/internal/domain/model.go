package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AIModel AI模型
type AIModel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	InternalKey string    `json:"internal_key" db:"internal_key"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Provider    string    `json:"provider" db:"provider"`
	ModelType   string    `json:"model_type" db:"model_type"`

	// 能力配置
	Capabilities json.RawMessage `json:"capabilities" db:"capabilities"`
	ParamsSchema json.RawMessage `json:"params_schema" db:"params_schema"`

	// 系统配置
	ModelSystemPrompt *string `json:"model_system_prompt" db:"model_system_prompt"`
	BaseURL           *string `json:"base_url" db:"base_url"`
	APIKeyEncrypted   *string `json:"-" db:"api_key_encrypted"`

	// 定价配置
	Pricing json.RawMessage `json:"pricing" db:"pricing"`

	// 可见性和权限
	Visibility  string  `json:"visibility" db:"visibility"`
	MinUserType string  `json:"min_user_type" db:"min_user_type"`
	MinBalance  float64 `json:"min_balance" db:"min_balance"`
	DailyLimit  int     `json:"daily_limit" db:"daily_limit"`
	UserLimit   int     `json:"user_limit" db:"user_limit"`

	// 健康和路由
	HealthStatus             string          `json:"health_status" db:"health_status"`
	Weight                   int             `json:"weight" db:"weight"`
	FallbackChain            json.RawMessage `json:"fallback_chain" db:"fallback_chain"`
	ConnectivityTestEndpoint *string         `json:"connectivity_test_endpoint" db:"connectivity_test_endpoint"`

	// 状态
	IsActive   bool `json:"is_active" db:"is_active"`
	IsFeatured bool `json:"is_featured" db:"is_featured"`

	// 时间戳
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	LastHealthCheck *time.Time `json:"last_health_check" db:"last_health_check"`
}

// ModelCapability 模型能力
type ModelCapability struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ModelID    uuid.UUID `json:"model_id" db:"model_id"`
	Capability string    `json:"capability" db:"capability"`
	IsEnabled  bool      `json:"is_enabled" db:"is_enabled"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ModelHealthLog 模型健康日志
type ModelHealthLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ModelID        uuid.UUID `json:"model_id" db:"model_id"`
	Status         string    `json:"status" db:"status"`
	ResponseTimeMs *int      `json:"response_time_ms" db:"response_time_ms"`
	ErrorMessage   *string   `json:"error_message" db:"error_message"`
	CheckedAt      time.Time `json:"checked_at" db:"checked_at"`
}

// ModelUsageStats 模型使用统计
type ModelUsageStats struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ModelID      uuid.UUID  `json:"model_id" db:"model_id"`
	UserID       *uuid.UUID `json:"user_id" db:"user_id"`
	Date         time.Time  `json:"date" db:"date"`
	RequestCount int        `json:"request_count" db:"request_count"`
	SuccessCount int        `json:"success_count" db:"success_count"`
	ErrorCount   int        `json:"error_count" db:"error_count"`
	TotalTokens  int64      `json:"total_tokens" db:"total_tokens"`
	TotalCost    float64    `json:"total_cost" db:"total_cost"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// 请求和响应结构

// CreateModelRequest 创建模型请求
type CreateModelRequest struct {
	InternalKey              string          `json:"internal_key" validate:"required,min=1,max=100"`
	DisplayName              string          `json:"display_name" validate:"required,min=1,max=200"`
	Provider                 string          `json:"provider" validate:"required,min=1,max=100"`
	ModelType                string          `json:"model_type" validate:"required,min=1,max=50"`
	Capabilities             json.RawMessage `json:"capabilities" validate:"required"`
	ParamsSchema             json.RawMessage `json:"params_schema" validate:"required"`
	ModelSystemPrompt        *string         `json:"model_system_prompt,omitempty"`
	BaseURL                  *string         `json:"base_url,omitempty"`
	APIKey                   *string         `json:"api_key,omitempty"`
	Pricing                  json.RawMessage `json:"pricing" validate:"required"`
	Visibility               string          `json:"visibility" validate:"required,oneof=public vip_only hidden"`
	MinUserType              string          `json:"min_user_type" validate:"required,oneof=basic vip creator admin"`
	MinBalance               float64         `json:"min_balance" validate:"min=0"`
	DailyLimit               int             `json:"daily_limit" validate:"min=-1"`
	UserLimit                int             `json:"user_limit" validate:"min=-1"`
	Weight                   int             `json:"weight" validate:"min=0,max=1000"`
	ConnectivityTestEndpoint *string         `json:"connectivity_test_endpoint,omitempty"`
}

// UpdateModelRequest 更新模型请求
type UpdateModelRequest struct {
	DisplayName              *string         `json:"display_name,omitempty" validate:"omitempty,min=1,max=200"`
	ModelSystemPrompt        *string         `json:"model_system_prompt,omitempty"`
	BaseURL                  *string         `json:"base_url,omitempty"`
	APIKey                   *string         `json:"api_key,omitempty"`
	Pricing                  json.RawMessage `json:"pricing,omitempty"`
	Visibility               *string         `json:"visibility,omitempty" validate:"omitempty,oneof=public vip_only hidden"`
	MinUserType              *string         `json:"min_user_type,omitempty" validate:"omitempty,oneof=basic vip creator admin"`
	MinBalance               *float64        `json:"min_balance,omitempty" validate:"omitempty,min=0"`
	DailyLimit               *int            `json:"daily_limit,omitempty" validate:"omitempty,min=-1"`
	UserLimit                *int            `json:"user_limit,omitempty" validate:"omitempty,min=-1"`
	Weight                   *int            `json:"weight,omitempty" validate:"omitempty,min=0,max=1000"`
	ConnectivityTestEndpoint *string         `json:"connectivity_test_endpoint,omitempty"`
	IsActive                 *bool           `json:"is_active,omitempty"`
	IsFeatured               *bool           `json:"is_featured,omitempty"`
}

// ModelListRequest 模型列表请求
type ModelListRequest struct {
	Provider    *string `json:"provider,omitempty"`
	ModelType   *string `json:"model_type,omitempty"`
	Capability  *string `json:"capability,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
	UserType    string  `json:"user_type" validate:"required,oneof=basic vip creator admin"`
	UserBalance float64 `json:"user_balance" validate:"min=0"`
	IsActive    *bool   `json:"is_active,omitempty"`
	IsFeatured  *bool   `json:"is_featured,omitempty"`
	Page        int     `json:"page" validate:"min=1"`
	Limit       int     `json:"limit" validate:"min=1,max=100"`
}

// ModelResponse 模型响应
type ModelResponse struct {
	*AIModel
	CapabilityList []string `json:"capability_list"`
	IsAvailable    bool     `json:"is_available"`
}

// ModelListResponse 模型列表响应
type ModelListResponse struct {
	Models []ModelResponse `json:"models"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// ConnectivityTestRequest 连接测试请求
type ConnectivityTestRequest struct {
	ModelID uuid.UUID `json:"model_id" validate:"required"`
}

// ConnectivityTestResponse 连接测试响应
type ConnectivityTestResponse struct {
	ModelID        uuid.UUID `json:"model_id"`
	Status         string    `json:"status"`
	ResponseTimeMs int       `json:"response_time_ms"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	TestedAt       time.Time `json:"tested_at"`
}

// ModelHealthResponse 模型健康响应
type ModelHealthResponse struct {
	ModelID         uuid.UUID        `json:"model_id"`
	HealthStatus    string           `json:"health_status"`
	LastHealthCheck *time.Time       `json:"last_health_check"`
	RecentLogs      []ModelHealthLog `json:"recent_logs"`
}

// ModelPricing 模型定价结构
type ModelPricing struct {
	// 基础定价
	InputTokenPrice  *float64 `json:"input_token_price,omitempty"`  // 输入token价格
	OutputTokenPrice *float64 `json:"output_token_price,omitempty"` // 输出token价格
	ImagePrice       *float64 `json:"image_price,omitempty"`        // 图片生成价格
	VideoPrice       *float64 `json:"video_price,omitempty"`        // 视频生成价格
	AudioPrice       *float64 `json:"audio_price,omitempty"`        // 音频处理价格

	// 计费单位
	Unit     string `json:"unit"`     // 计费单位：1k_tokens, image, second, minute
	Currency string `json:"currency"` // 货币类型：USD, CNY

	// 倍率设置
	PriceMultiplier float64 `json:"price_multiplier"` // 价格倍率，默认1.0

	// 高级定价选项
	TierPricing map[string]float64 `json:"tier_pricing,omitempty"` // 分层定价

	// 最小消费
	MinCharge float64 `json:"min_charge"` // 最小消费金额
}

// ModelPricingConfig 模型定价配置请求
type ModelPricingConfigRequest struct {
	ModelID          uuid.UUID          `json:"model_id" validate:"required"`
	InputTokenPrice  *float64           `json:"input_token_price,omitempty" validate:"omitempty,min=0"`
	OutputTokenPrice *float64           `json:"output_token_price,omitempty" validate:"omitempty,min=0"`
	ImagePrice       *float64           `json:"image_price,omitempty" validate:"omitempty,min=0"`
	VideoPrice       *float64           `json:"video_price,omitempty" validate:"omitempty,min=0"`
	AudioPrice       *float64           `json:"audio_price,omitempty" validate:"omitempty,min=0"`
	Unit             string             `json:"unit" validate:"required,oneof=1k_tokens image second minute"`
	Currency         string             `json:"currency" validate:"required,oneof=USD CNY"`
	PriceMultiplier  float64            `json:"price_multiplier" validate:"min=0.1,max=10.0"`
	TierPricing      map[string]float64 `json:"tier_pricing,omitempty"`
	MinCharge        float64            `json:"min_charge" validate:"min=0"`
}

// ModelPricingResponse 模型定价响应
type ModelPricingResponse struct {
	ModelID         uuid.UUID           `json:"model_id"`
	ModelName       string              `json:"model_name"`
	Provider        string              `json:"provider"`
	OriginalPricing ModelPricing        `json:"original_pricing"` // 官方原价
	CurrentPricing  ModelPricing        `json:"current_pricing"`  // 当前价格
	PriceHistory    []ModelPriceHistory `json:"price_history"`    // 价格历史
	EstimatedCost   map[string]float64  `json:"estimated_cost"`   // 预估成本
}

// ModelPriceHistory 模型价格历史
type ModelPriceHistory struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	ModelID     uuid.UUID       `json:"model_id" db:"model_id"`
	PricingData json.RawMessage `json:"pricing_data" db:"pricing_data"`
	Reason      string          `json:"reason" db:"reason"`
	OperatorID  *uuid.UUID      `json:"operator_id" db:"operator_id"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// PricingCalculationRequest 价格计算请求
type PricingCalculationRequest struct {
	ModelID      uuid.UUID `json:"model_id" validate:"required"`
	InputTokens  *int64    `json:"input_tokens,omitempty" validate:"omitempty,min=0"`
	OutputTokens *int64    `json:"output_tokens,omitempty" validate:"omitempty,min=0"`
	ImageCount   *int      `json:"image_count,omitempty" validate:"omitempty,min=0"`
	VideoSeconds *int      `json:"video_seconds,omitempty" validate:"omitempty,min=0"`
	AudioSeconds *int      `json:"audio_seconds,omitempty" validate:"omitempty,min=0"`
	UserType     string    `json:"user_type" validate:"required,oneof=basic vip creator admin"`
}

// PricingCalculationResponse 价格计算响应
type PricingCalculationResponse struct {
	ModelID       uuid.UUID          `json:"model_id"`
	OriginalCost  float64            `json:"original_cost"`  // 官方原价成本
	PlatformCost  float64            `json:"platform_cost"`  // 平台价格成本
	UserCost      float64            `json:"user_cost"`      // 用户实际支付
	CostBreakdown map[string]float64 `json:"cost_breakdown"` // 成本明细
	Currency      string             `json:"currency"`
	Multiplier    float64            `json:"multiplier"` // 应用的倍率
	Discount      float64            `json:"discount"`   // 用户折扣
}

// 常量定义
const (
	// 模型类型
	ModelTypeChat  = "chat"
	ModelTypeImage = "image"
	ModelTypeVideo = "video"
	ModelTypeAudio = "audio"
	ModelTypeCode  = "code"

	// 可见性
	VisibilityPublic  = "public"
	VisibilityVIPOnly = "vip_only"
	VisibilityHidden  = "hidden"

	// 健康状态
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"

	// 能力类型
	CapabilityChat           = "chat"
	CapabilityTextGeneration = "text_generation"
	CapabilityCodeGeneration = "code_generation"
	CapabilityTxt2Img        = "txt2img"
	CapabilityImg2Img        = "img2img"
	CapabilityImg2Video      = "img2video"
	CapabilityTxt2Video      = "txt2video"
	CapabilityTTS            = "tts"
	CapabilitySTT            = "stt"
	CapabilityRemoveBG       = "remove_bg"
	CapabilityAnimation      = "animation"

	// 计费单位
	UnitTokens1K = "1k_tokens"
	UnitImage    = "image"
	UnitSecond   = "second"
	UnitMinute   = "minute"

	// 货币类型
	CurrencyUSD = "USD"
	CurrencyCNY = "CNY"

	// 默认倍率
	DefaultPriceMultiplier = 1.0
)

// AI调用计费相关结构

// PreCheckResponse 预扣费检查响应
type PreCheckResponse struct {
	UserID         uuid.UUID                   `json:"user_id"`
	ModelID        uuid.UUID                   `json:"model_id"`
	RequiredCoins  float64                     `json:"required_coins"`
	CurrentBalance float64                     `json:"current_balance"`
	Sufficient     bool                        `json:"sufficient"`
	ErrorMessage   string                      `json:"error_message,omitempty"`
	CostDetails    *PricingCalculationResponse `json:"cost_details"`
	ExchangeRate   float64                     `json:"exchange_rate"`
}

// AICallChargeRequest AI调用扣费请求
type AICallChargeRequest struct {
	RequestID                 uuid.UUID `json:"request_id" validate:"required"`
	ModelID                   uuid.UUID `json:"model_id" validate:"required"`
	ModelName                 string    `json:"model_name" validate:"required"`
	TotalTokens               int64     `json:"total_tokens" validate:"min=0"`
	PricingCalculationRequest           // 嵌入定价计算请求
}

// AICallChargeResponse AI调用扣费响应
type AICallChargeResponse struct {
	Success       bool              `json:"success"`
	TransactionID uuid.UUID         `json:"transaction_id,omitempty"`
	CoinsCharged  float64           `json:"coins_charged,omitempty"`
	NewBalance    float64           `json:"new_balance,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
	PreCheck      *PreCheckResponse `json:"pre_check,omitempty"`
}

// AICallRecord AI调用记录
type AICallRecord struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ModelID      uuid.UUID `json:"model_id"`
	ModelName    string    `json:"model_name"`
	RequestID    uuid.UUID `json:"request_id"`
	CoinsCharged float64   `json:"coins_charged"`
	TokensUsed   int64     `json:"tokens_used"`
	CallType     string    `json:"call_type"` // chat, image, video, audio
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// AICallHistoryResponse AI调用历史响应
type AICallHistoryResponse struct {
	Records []AICallRecord `json:"records"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
}
