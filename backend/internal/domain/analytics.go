package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserStats 用户统计数据
type UserStats struct {
	UserID      uuid.UUID `json:"user_id"`
	TimeRange   string    `json:"time_range"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	LoginCount  int64     `json:"login_count"`
	ChatCount   int64     `json:"chat_count"`
	MomentCount int64     `json:"moment_count"`
	ActiveDays  int64     `json:"active_days"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SystemHealth 系统健康状态
type SystemHealth struct {
	Status     string                 `json:"status"` // healthy, degraded, unhealthy
	Timestamp  time.Time              `json:"timestamp"`
	Uptime     time.Duration          `json:"uptime"`
	Services   map[string]string      `json:"services"` // 各服务状态
	UserCount  int64                  `json:"user_count"`
	APIMetrics map[string]*APIMetrics `json:"api_metrics"`
}

// APIMetrics API指标
type APIMetrics struct {
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	RequestCount int64         `json:"request_count"`
	SuccessCount int64         `json:"success_count"`
	ErrorCount   int64         `json:"error_count"`
	TotalTime    time.Duration `json:"total_time"`
	AvgTime      time.Duration `json:"avg_time"`
	MinTime      time.Duration `json:"min_time"`
	MaxTime      time.Duration `json:"max_time"`
	SuccessRate  float64       `json:"success_rate"`
}

// RevenueStats 收入统计
type RevenueStats struct {
	TimeRange    string    `json:"time_range"`
	TotalRevenue float64   `json:"total_revenue"`
	OrderCount   int64     `json:"order_count"`
	ARPU         float64   `json:"arpu"` // 用户平均收入
	UpdatedAt    time.Time `json:"updated_at"`
}

// RetentionStats 留存率统计
type RetentionStats struct {
	CohortDate time.Time `json:"cohort_date"`
	Day1       float64   `json:"day1"`  // 次日留存
	Day7       float64   `json:"day7"`  // 7日留存
	Day30      float64   `json:"day30"` // 30日留存
	UpdatedAt  time.Time `json:"updated_at"`
}

// RealTimeMetrics 实时指标
type RealTimeMetrics struct {
	Timestamp    time.Time         `json:"timestamp"`
	OnlineUsers  int64             `json:"online_users"`
	EventCounts  map[string]string `json:"event_counts"`
	SystemStatus string            `json:"system_status"`
}

// AnalyticsDashboard 分析面板
type AnalyticsDashboard struct {
	TotalUsers         int64            `json:"total_users"`
	DailyActiveUsers   int64            `json:"daily_active_users"`
	MonthlyActiveUsers int64            `json:"monthly_active_users"`
	RealTimeMetrics    *RealTimeMetrics `json:"real_time_metrics"`
	SystemHealth       *SystemHealth    `json:"system_health"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// SystemStats 系统统计
type SystemStats struct {
	TotalUsers  int64                  `json:"total_users"`
	ActiveUsers int64                  `json:"active_users"`
	Uptime      time.Duration          `json:"uptime"`
	APIMetrics  map[string]*APIMetrics `json:"api_metrics"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// UserEvent 用户事件
type UserEvent struct {
	ID         uuid.UUID              `json:"id"`
	UserID     uuid.UUID              `json:"user_id"`
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Metric      string    `json:"metric" db:"metric"`       // 监控指标
	Operator    string    `json:"operator" db:"operator"`   // 比较操作符 (>, <, =, >=, <=)
	Threshold   float64   `json:"threshold" db:"threshold"` // 阈值
	Duration    int       `json:"duration" db:"duration"`   // 持续时间(秒)
	Severity    string    `json:"severity" db:"severity"`   // 严重程度 (low, medium, high, critical)
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Alert 告警记录
type Alert struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	RuleID     uuid.UUID  `json:"rule_id" db:"rule_id"`
	RuleName   string     `json:"rule_name" db:"rule_name"`
	Metric     string     `json:"metric" db:"metric"`
	Value      float64    `json:"value" db:"value"`
	Threshold  float64    `json:"threshold" db:"threshold"`
	Severity   string     `json:"severity" db:"severity"`
	Status     string     `json:"status" db:"status"` // firing, resolved
	Message    string     `json:"message" db:"message"`
	FiredAt    time.Time  `json:"fired_at" db:"fired_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"` // counter, gauge, histogram
	Help      string            `json:"help"`
	Labels    map[string]string `json:"labels"`
	Value     float64           `json:"value"`
	Buckets   []float64         `json:"buckets,omitempty"` // histogram用
	Samples   []MetricSample    `json:"samples,omitempty"` // 样本数据
	UpdatedAt time.Time         `json:"updated_at"`
}

// MetricSample 指标样本
type MetricSample struct {
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// BusinessMetrics 业务指标
type BusinessMetrics struct {
	// 用户指标
	TotalUsers    int64   `json:"total_users"`
	NewUsers      int64   `json:"new_users"`      // 新增用户
	ActiveUsers   int64   `json:"active_users"`   // 活跃用户
	RetentionRate float64 `json:"retention_rate"` // 留存率
	ChurnRate     float64 `json:"churn_rate"`     // 流失率

	// 内容指标
	TotalCharacters int64 `json:"total_characters"`
	NewCharacters   int64 `json:"new_characters"`
	TotalChats      int64 `json:"total_chats"`
	TotalMoments    int64 `json:"total_moments"`

	// 收入指标
	TotalRevenue   float64 `json:"total_revenue"`
	ARPU           float64 `json:"arpu"`            // 用户平均收入
	ConversionRate float64 `json:"conversion_rate"` // 转化率

	// 时间戳
	Date      time.Time `json:"date"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TechnicalMetrics 技术指标
type TechnicalMetrics struct {
	// 系统性能
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   int64   `json:"network_io"`

	// 数据库性能
	DBConnections int64         `json:"db_connections"`
	DBQueryTime   time.Duration `json:"db_query_time"`
	DBErrorRate   float64       `json:"db_error_rate"`

	// Redis性能
	RedisConnections int64   `json:"redis_connections"`
	RedisMemoryUsage int64   `json:"redis_memory_usage"`
	RedisHitRate     float64 `json:"redis_hit_rate"`

	// API性能
	APIResponseTime time.Duration `json:"api_response_time"`
	APIThroughput   int64         `json:"api_throughput"`
	APIErrorRate    float64       `json:"api_error_rate"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// AIMetrics AI模型指标
type AIMetrics struct {
	// 模型性能
	ModelResponseTime time.Duration `json:"model_response_time"`
	ModelSuccessRate  float64       `json:"model_success_rate"`
	ModelErrorRate    float64       `json:"model_error_rate"`
	TokensUsed        int64         `json:"tokens_used"`
	ModelCost         float64       `json:"model_cost"`

	// 内容质量
	ContentQuality    float64 `json:"content_quality"`    // 内容质量评分
	UserSatisfaction  float64 `json:"user_satisfaction"`  // 用户满意度
	ResponseRelevance float64 `json:"response_relevance"` // 回应相关性

	// 使用统计
	ChatRequests   int64 `json:"chat_requests"`
	MomentRequests int64 `json:"moment_requests"`
	ImageRequests  int64 `json:"image_requests"`

	// 时间戳
	Date      time.Time `json:"date"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	Total       uint64    `json:"total"`
	Used        uint64    `json:"used"`
	Free        uint64    `json:"free"`
	UsedPercent float64   `json:"used_percent"`
	Timestamp   time.Time `json:"timestamp"`
}

// DiskStats 磁盘统计
type DiskStats struct {
	Total       uint64    `json:"total"`
	Used        uint64    `json:"used"`
	Free        uint64    `json:"free"`
	UsedPercent float64   `json:"used_percent"`
	Timestamp   time.Time `json:"timestamp"`
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	ActiveConnections int64         `json:"active_connections"`
	IdleConnections   int64         `json:"idle_connections"`
	MaxConnections    int64         `json:"max_connections"`
	AvgQueryTime      time.Duration `json:"avg_query_time"`
	SlowQueries       int64         `json:"slow_queries"`
	ErrorRate         float64       `json:"error_rate"`
	Timestamp         time.Time     `json:"timestamp"`
}

// RedisMetrics Redis指标
type RedisMetrics struct {
	ConnectedClients int64     `json:"connected_clients"`
	UsedMemory       int64     `json:"used_memory"`
	MaxMemory        int64     `json:"max_memory"`
	HitRate          float64   `json:"hit_rate"`
	KeyspaceHits     int64     `json:"keyspace_hits"`
	KeyspaceMisses   int64     `json:"keyspace_misses"`
	Timestamp        time.Time `json:"timestamp"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	BytesReceived   int64     `json:"bytes_received"`
	BytesSent       int64     `json:"bytes_sent"`
	PacketsReceived int64     `json:"packets_received"`
	PacketsSent     int64     `json:"packets_sent"`
	Timestamp       time.Time `json:"timestamp"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status    string                  `json:"status"` // healthy, degraded, unhealthy
	Timestamp time.Time               `json:"timestamp"`
	Checks    map[string]*HealthCheck `json:"checks"`
}

// HealthCheck 单项健康检查
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // healthy, degraded, unhealthy
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ReadinessCheckResult 就绪检查结果
type ReadinessCheckResult struct {
	Ready     bool                       `json:"ready"`
	Timestamp time.Time                  `json:"timestamp"`
	Checks    map[string]*ReadinessCheck `json:"checks"`
}

// ReadinessCheck 单项就绪检查
type ReadinessCheck struct {
	Name    string `json:"name"`
	Ready   bool   `json:"ready"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
