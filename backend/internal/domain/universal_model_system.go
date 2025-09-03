package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// 🌍 YUNAI 超级模型管理平台 - 支持全球所有模型的完全自定义系统

// UniversalModelConfig 通用模型配置 - 支持任何厂商任何模型
type UniversalModelConfig struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`                 // 自定义显示名称
	InternalKey string    `json:"internal_key" db:"internal_key"` // 内部唯一标识

	// 🔗 基础连接配置 - 完全自定义
	Provider    string `json:"provider" db:"provider"`         // 提供商标识 (可自定义)
	ModelID     string `json:"model_id" db:"model_id"`         // 原始模型ID
	APIEndpoint string `json:"api_endpoint" db:"api_endpoint"` // 完整API端点URL
	APIKey      string `json:"api_key" db:"api_key"`           // API密钥
	APISecret   string `json:"api_secret" db:"api_secret"`     // API密钥(如果需要)

	// 🎯 模型能力定义 - 完全自定义
	ModelType        string   `json:"model_type" db:"model_type"`               // 模型类型 (text, image, video, audio, multimodal)
	Capabilities     []string `json:"capabilities" db:"capabilities"`           // 能力列表 (完全自定义)
	SupportedFormats []string `json:"supported_formats" db:"supported_formats"` // 支持的格式

	// 🔧 请求配置 - 完全自定义
	HTTPMethod      string                 `json:"http_method" db:"http_method"`           // HTTP方法 (GET, POST, PUT, etc.)
	RequestHeaders  map[string]string      `json:"request_headers" db:"request_headers"`   // 自定义请求头
	RequestTemplate *json.RawMessage       `json:"request_template" db:"request_template"` // 请求体模板
	AuthType        string                 `json:"auth_type" db:"auth_type"`               // 认证类型 (bearer, basic, custom)
	AuthConfig      map[string]interface{} `json:"auth_config" db:"auth_config"`           // 认证配置

	// 📊 响应处理 - 完全自定义
	ResponseMapping   *json.RawMessage `json:"response_mapping" db:"response_mapping"`     // 响应字段映射
	ErrorMapping      *json.RawMessage `json:"error_mapping" db:"error_mapping"`           // 错误字段映射
	SuccessIndicators []string         `json:"success_indicators" db:"success_indicators"` // 成功标识
	StreamingConfig   *StreamingConfig `json:"streaming_config" db:"streaming_config"`     // 流式配置

	// 🎮 参数配置 - 超级自定义
	ParameterSchema   *json.RawMessage          `json:"parameter_schema" db:"parameter_schema"`     // 参数JSON Schema
	DefaultParameters map[string]interface{}    `json:"default_parameters" db:"default_parameters"` // 默认参数
	ParameterMappings map[string]string         `json:"parameter_mappings" db:"parameter_mappings"` // 参数名映射
	ParameterRanges   map[string]ParameterRange `json:"parameter_ranges" db:"parameter_ranges"`     // 参数范围

	// 🎨 AI 高级参数 - 全面支持
	AIParameters *AIParameterConfig `json:"ai_parameters" db:"ai_parameters"` // AI参数配置

	// 💰 定价配置 - 完全自定义
	PricingConfig *CustomPricingConfig `json:"pricing_config" db:"pricing_config"` // 定价配置

	// 🏥 健康监控 - 完全自定义
	HealthCheck *CustomHealthCheck `json:"health_check" db:"health_check"` // 健康检查配置

	// 🔄 高级功能
	LoadBalancing *LoadBalancingConfig `json:"load_balancing" db:"load_balancing"` // 负载均衡
	Fallback      *FallbackConfig      `json:"fallback" db:"fallback"`             // 降级配置
	Caching       *CachingConfig       `json:"caching" db:"caching"`               // 缓存配置
	RateLimit     *RateLimitConfig     `json:"rate_limit" db:"rate_limit"`         // 限流配置

	// 📝 元数据
	Categories    []string `json:"categories" db:"categories"`       // 自定义分类
	Tags          []string `json:"tags" db:"tags"`                   // 标签
	Description   string   `json:"description" db:"description"`     // 描述
	Documentation string   `json:"documentation" db:"documentation"` // 文档链接
	Version       string   `json:"version" db:"version"`             // 版本
	Status        string   `json:"status" db:"status"`               // 状态
	Visibility    string   `json:"visibility" db:"visibility"`       // 可见性

	// 🔐 权限控制
	AccessControl *AccessControlConfig `json:"access_control" db:"access_control"` // 访问控制

	// 📊 统计信息
	UsageStats *UsageStatsConfig `json:"usage_stats" db:"usage_stats"` // 使用统计

	// 🕐 时间戳
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ParameterRange 参数范围定义
type ParameterRange struct {
	Type          string        `json:"type"`           // 参数类型 (number, string, boolean, array, object)
	MinValue      interface{}   `json:"min_value"`      // 最小值
	MaxValue      interface{}   `json:"max_value"`      // 最大值
	AllowedValues []interface{} `json:"allowed_values"` // 允许的值
	DefaultValue  interface{}   `json:"default_value"`  // 默认值
	Required      bool          `json:"required"`       // 是否必需
	Description   string        `json:"description"`    // 参数描述
}

// AIParameterConfig AI参数配置 - 支持所有AI模型参数
type AIParameterConfig struct {
	// 🌡️ 生成控制参数
	Temperature *float64 `json:"temperature"` // 温度 (0.0-2.0)
	TopP        *float64 `json:"top_p"`       // Top-p 采样
	TopK        *int     `json:"top_k"`       // Top-k 采样
	MaxTokens   *int     `json:"max_tokens"`  // 最大输出token
	MinTokens   *int     `json:"min_tokens"`  // 最小输出token

	// 🎯 惩罚参数
	FrequencyPenalty  *float64 `json:"frequency_penalty"`  // 频率惩罚
	PresencePenalty   *float64 `json:"presence_penalty"`   // 存在惩罚
	RepetitionPenalty *float64 `json:"repetition_penalty"` // 重复惩罚

	// 🛑 停止控制
	StopSequences []string `json:"stop_sequences"` // 停止序列
	StopTokens    []string `json:"stop_tokens"`    // 停止token

	// 🎭 提示词配置
	SystemPrompt     string `json:"system_prompt"`      // 系统提示词
	UserPromptPrefix string `json:"user_prompt_prefix"` // 用户提示词前缀
	UserPromptSuffix string `json:"user_prompt_suffix"` // 用户提示词后缀

	// 🎨 创意参数
	Creativity *float64 `json:"creativity"` // 创意度
	Coherence  *float64 `json:"coherence"`  // 连贯性
	Diversity  *float64 `json:"diversity"`  // 多样性

	// 🖼️ 图像参数
	ImageSize     string   `json:"image_size"`     // 图像尺寸
	ImageQuality  string   `json:"image_quality"`  // 图像质量
	ImageStyle    string   `json:"image_style"`    // 图像风格
	ImageSteps    *int     `json:"image_steps"`    // 生成步数
	ImageGuidance *float64 `json:"image_guidance"` // 引导强度

	// 🎬 视频参数
	VideoLength     *int   `json:"video_length"`     // 视频长度(秒)
	VideoFPS        *int   `json:"video_fps"`        // 帧率
	VideoResolution string `json:"video_resolution"` // 分辨率
	VideoStyle      string `json:"video_style"`      // 视频风格

	// 🎵 音频参数
	AudioFormat  string   `json:"audio_format"`  // 音频格式
	AudioQuality string   `json:"audio_quality"` // 音频质量
	VoiceID      string   `json:"voice_id"`      // 音色ID
	SpeechSpeed  *float64 `json:"speech_speed"`  // 语速
	SpeechPitch  *float64 `json:"speech_pitch"`  // 音调

	// 🔧 自定义参数 - 支持任何未来参数
	CustomParameters map[string]interface{} `json:"custom_parameters"` // 完全自定义参数
}

// StreamingConfig 流式配置
type StreamingConfig struct {
	Supported         bool   `json:"supported"`          // 是否支持流式
	StreamingEndpoint string `json:"streaming_endpoint"` // 流式端点
	StreamingMethod   string `json:"streaming_method"`   // 流式方法
	ChunkDelimiter    string `json:"chunk_delimiter"`    // 块分隔符
	DataPrefix        string `json:"data_prefix"`        // 数据前缀
	EventType         string `json:"event_type"`         // 事件类型
}

// CustomPricingConfig 自定义定价配置
type CustomPricingConfig struct {
	BillingType       string           `json:"billing_type"`       // 计费类型 (token, request, time, custom)
	PricePerUnit      float64          `json:"price_per_unit"`     // 单价
	Currency          string           `json:"currency"`           // 货币
	FreeQuota         int64            `json:"free_quota"`         // 免费额度
	BillingRules      []BillingRule    `json:"billing_rules"`      // 计费规则
	CustomCalculation *json.RawMessage `json:"custom_calculation"` // 自定义计算公式
}

// BillingRule 计费规则
type BillingRule struct {
	Condition string  `json:"condition"` // 条件
	Price     float64 `json:"price"`     // 价格
	Unit      string  `json:"unit"`      // 单位
}

// CustomHealthCheck 自定义健康检查
type CustomHealthCheck struct {
	Enabled           bool              `json:"enabled"`            // 是否启用
	CheckInterval     int               `json:"check_interval"`     // 检查间隔(秒)
	HealthEndpoint    string            `json:"health_endpoint"`    // 健康检查端点
	HealthMethod      string            `json:"health_method"`      // 检查方法
	HealthHeaders     map[string]string `json:"health_headers"`     // 检查请求头
	HealthBody        string            `json:"health_body"`        // 检查请求体
	SuccessConditions []string          `json:"success_conditions"` // 成功条件
	TimeoutSeconds    int               `json:"timeout_seconds"`    // 超时时间
}

// LoadBalancingConfig 负载均衡配置 - 支持相同模型轮询
type LoadBalancingConfig struct {
	Enabled         bool                  `json:"enabled"`          // 是否启用负载均衡
	Strategy        string                `json:"strategy"`         // 策略 (round_robin, weighted, least_connections, random)
	ModelInstances  []*ModelInstance      `json:"model_instances"`  // 模型实例列表
	HealthCheck     bool                  `json:"health_check"`     // 是否健康检查
	FailoverEnabled bool                  `json:"failover_enabled"` // 是否启用故障转移
	MaxRetries      int                   `json:"max_retries"`      // 最大重试次数
	RetryDelay      int                   `json:"retry_delay"`      // 重试延迟(秒)
	CircuitBreaker  *CircuitBreakerConfig `json:"circuit_breaker"`  // 熔断器配置
	StickySession   bool                  `json:"sticky_session"`   // 是否启用会话粘性
	SessionTTL      int                   `json:"session_ttl"`      // 会话TTL(秒)
}

// ModelInstance 模型实例 - 支持相同模型的多个实例
type ModelInstance struct {
	ID             uuid.UUID              `json:"id"`              // 实例ID
	Name           string                 `json:"name"`            // 实例名称
	APIEndpoint    string                 `json:"api_endpoint"`    // API端点
	APIKey         string                 `json:"api_key"`         // API密钥
	APISecret      string                 `json:"api_secret"`      // API密钥(如果需要)
	Weight         int                    `json:"weight"`          // 权重 (1-100)
	Priority       int                    `json:"priority"`        // 优先级 (1-10, 1最高)
	MaxConcurrency int                    `json:"max_concurrency"` // 最大并发数
	CurrentLoad    int                    `json:"current_load"`    // 当前负载
	Status         string                 `json:"status"`          // 状态 (active, inactive, maintenance)
	Health         *InstanceHealth        `json:"health"`          // 健康状态
	RateLimit      *InstanceRateLimit     `json:"rate_limit"`      // 限流配置
	CostConfig     *InstanceCostConfig    `json:"cost_config"`     // 成本配置
	Region         string                 `json:"region"`          // 地区
	Provider       string                 `json:"provider"`        // 提供商
	CustomHeaders  map[string]string      `json:"custom_headers"`  // 自定义请求头
	CustomConfig   map[string]interface{} `json:"custom_config"`   // 自定义配置
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// InstanceHealth 实例健康状态
type InstanceHealth struct {
	Status            string    `json:"status"`             // healthy, unhealthy, unknown
	LastCheckTime     time.Time `json:"last_check_time"`    // 最后检查时间
	ResponseTime      int64     `json:"response_time"`      // 平均响应时间(ms)
	SuccessRate       float64   `json:"success_rate"`       // 成功率 (0.0-1.0)
	ErrorCount        int       `json:"error_count"`        // 错误次数
	ConsecutiveErrors int       `json:"consecutive_errors"` // 连续错误次数
	LastError         string    `json:"last_error"`         // 最后错误信息
	Uptime            float64   `json:"uptime"`             // 可用时间百分比
}

// InstanceRateLimit 实例限流配置
type InstanceRateLimit struct {
	RequestsPerSecond int `json:"requests_per_second"` // 每秒请求数
	RequestsPerMinute int `json:"requests_per_minute"` // 每分钟请求数
	RequestsPerHour   int `json:"requests_per_hour"`   // 每小时请求数
	RequestsPerDay    int `json:"requests_per_day"`    // 每天请求数
	TokensPerMinute   int `json:"tokens_per_minute"`   // 每分钟token数
	BurstSize         int `json:"burst_size"`          // 突发大小
}

// InstanceCostConfig 实例成本配置
type InstanceCostConfig struct {
	InputTokenPrice  float64 `json:"input_token_price"`  // 输入token价格
	OutputTokenPrice float64 `json:"output_token_price"` // 输出token价格
	RequestPrice     float64 `json:"request_price"`      // 请求价格
	Currency         string  `json:"currency"`           // 货币
	CostPriority     int     `json:"cost_priority"`      // 成本优先级 (1-10, 1最便宜)
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Enabled                  bool `json:"enabled"`                     // 是否启用熔断器
	FailureThreshold         int  `json:"failure_threshold"`           // 失败阈值
	SuccessThreshold         int  `json:"success_threshold"`           // 成功阈值
	Timeout                  int  `json:"timeout"`                     // 超时时间(秒)
	HalfOpenMaxCalls         int  `json:"half_open_max_calls"`         // 半开状态最大调用数
	HalfOpenSuccessThreshold int  `json:"half_open_success_threshold"` // 半开状态成功阈值
}

// FallbackConfig 降级配置
type FallbackConfig struct {
	Enabled           bool        `json:"enabled"`            // 是否启用降级
	FallbackModels    []uuid.UUID `json:"fallback_models"`    // 降级模型列表
	TriggerConditions []string    `json:"trigger_conditions"` // 触发条件
	MaxRetries        int         `json:"max_retries"`        // 最大重试次数
}

// CachingConfig 缓存配置
type CachingConfig struct {
	Enabled       bool   `json:"enabled"`        // 是否启用缓存
	TTLSeconds    int    `json:"ttl_seconds"`    // 缓存时间
	CacheKey      string `json:"cache_key"`      // 缓存键模板
	CacheStrategy string `json:"cache_strategy"` // 缓存策略
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerSecond int `json:"requests_per_second"` // 每秒请求数
	RequestsPerMinute int `json:"requests_per_minute"` // 每分钟请求数
	RequestsPerHour   int `json:"requests_per_hour"`   // 每小时请求数
	RequestsPerDay    int `json:"requests_per_day"`    // 每天请求数
	BurstSize         int `json:"burst_size"`          // 突发大小
}

// AccessControlConfig 访问控制配置
type AccessControlConfig struct {
	RequiredUserLevel string      `json:"required_user_level"` // 所需用户等级
	RequiredBalance   float64     `json:"required_balance"`    // 所需余额
	AllowedUsers      []uuid.UUID `json:"allowed_users"`       // 允许的用户
	BlockedUsers      []uuid.UUID `json:"blocked_users"`       // 禁止的用户
	AllowedRegions    []string    `json:"allowed_regions"`     // 允许的地区
	BlockedRegions    []string    `json:"blocked_regions"`     // 禁止的地区
}

// UsageStatsConfig 使用统计配置
type UsageStatsConfig struct {
	TrackUsage       bool     `json:"track_usage"`       // 是否跟踪使用
	TrackPerformance bool     `json:"track_performance"` // 是否跟踪性能
	TrackErrors      bool     `json:"track_errors"`      // 是否跟踪错误
	RetentionDays    int      `json:"retention_days"`    // 数据保留天数
	MetricsToTrack   []string `json:"metrics_to_track"`  // 要跟踪的指标
}

// 🎯 扩展的模型能力定义 - 避免与现有定义冲突
var (
	// 扩展文本能力
	CapabilityUniversalSummarization = "summarization"

	// 扩展图像能力
	CapabilityUniversalImageAnalysis    = "image_analysis"
	CapabilityUniversalRemoveBackground = "remove_background"

	// 视频能力
	CapabilityUniversalVideoEdit     = "video_edit"
	CapabilityUniversalVideoAnalysis = "video_analysis"

	// 音频能力
	CapabilityUniversalAudioGeneration = "audio_generation"
	CapabilityUniversalMusicGeneration = "music_generation"

	// 嵌入和检索
	CapabilityUniversalSemanticSearch = "semantic_search"

	// 多模态能力
	CapabilityUniversalVisionLanguage = "vision_language"
	CapabilityUniversalMultimodal     = "multimodal"
)

// 🌍 预定义的提供商 - 可扩展
var (
	ProviderOpenAI      = "openai"
	ProviderAnthropic   = "anthropic"
	ProviderGoogle      = "google"
	ProviderMicrosoft   = "microsoft"
	ProviderAmazon      = "amazon"
	ProviderSiliconFlow = "siliconflow"
	ProviderBaidu       = "baidu"
	ProviderAlibaba     = "alibaba"
	ProviderTencent     = "tencent"
	ProviderByteDance   = "bytedance"
	ProviderMiniMax     = "minimax"
	ProviderZhipu       = "zhipu"
	ProviderMoonshot    = "moonshot"
	ProviderCustom      = "custom"
	ProviderSelfHosted  = "self_hosted"
)
