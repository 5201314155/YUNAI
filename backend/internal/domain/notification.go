package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeChat     NotificationType = "chat"     // 聊天消息
	NotificationTypeMoment   NotificationType = "moment"   // 朋友圈动态
	NotificationTypeSystem   NotificationType = "system"   // 系统通知
	NotificationTypeActivity NotificationType = "activity" // 活动通知
	NotificationTypeReminder NotificationType = "reminder" // 提醒通知
	NotificationTypeInvite   NotificationType = "invite"   // 邀请通知
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"   // 待发送
	NotificationStatusSent      NotificationStatus = "sent"      // 已发送
	NotificationStatusDelivered NotificationStatus = "delivered" // 已送达
	NotificationStatusRead      NotificationStatus = "read"      // 已读
	NotificationStatusFailed    NotificationStatus = "failed"    // 发送失败
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	NotificationChannelInApp     NotificationChannel = "in_app"     // 应用内通知
	NotificationChannelPush      NotificationChannel = "push"       // 推送通知
	NotificationChannelEmail     NotificationChannel = "email"      // 邮件通知
	NotificationChannelSMS       NotificationChannel = "sms"        // 短信通知
	NotificationChannelWebSocket NotificationChannel = "websocket"  // WebSocket实时通知
)

// NotificationPriority 通知优先级
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"      // 低优先级
	NotificationPriorityNormal   NotificationPriority = "normal"   // 普通优先级
	NotificationPriorityHigh     NotificationPriority = "high"     // 高优先级
	NotificationPriorityCritical NotificationPriority = "critical" // 紧急优先级
)

// Notification 通知
type Notification struct {
	ID          uuid.UUID            `json:"id" db:"id"`
	UserID      uuid.UUID            `json:"user_id" db:"user_id"`
	Type        NotificationType     `json:"type" db:"type"`
	Channel     NotificationChannel  `json:"channel" db:"channel"`
	Priority    NotificationPriority `json:"priority" db:"priority"`
	Status      NotificationStatus   `json:"status" db:"status"`
	Title       string               `json:"title" db:"title"`
	Content     string               `json:"content" db:"content"`
	Data        json.RawMessage      `json:"data" db:"data"`         // 额外数据
	TemplateID  *uuid.UUID           `json:"template_id" db:"template_id"`
	ScheduledAt *time.Time           `json:"scheduled_at" db:"scheduled_at"` // 定时发送
	SentAt      *time.Time           `json:"sent_at" db:"sent_at"`
	ReadAt      *time.Time           `json:"read_at" db:"read_at"`
	ExpiresAt   *time.Time           `json:"expires_at" db:"expires_at"`
	RetryCount  int                  `json:"retry_count" db:"retry_count"`
	MaxRetries  int                  `json:"max_retries" db:"max_retries"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          uuid.UUID            `json:"id" db:"id"`
	Name        string               `json:"name" db:"name"`
	Type        NotificationType     `json:"type" db:"type"`
	Channel     NotificationChannel  `json:"channel" db:"channel"`
	Title       string               `json:"title" db:"title"`           // 支持模板变量
	Content     string               `json:"content" db:"content"`       // 支持模板变量
	Variables   []string             `json:"variables" db:"variables"`   // 模板变量列表
	Config      json.RawMessage      `json:"config" db:"config"`         // 渠道特定配置
	Enabled     bool                 `json:"enabled" db:"enabled"`
	Description string               `json:"description" db:"description"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

// NotificationPreferences 用户通知偏好
type NotificationPreferences struct {
	ID                uuid.UUID                       `json:"id" db:"id"`
	UserID            uuid.UUID                       `json:"user_id" db:"user_id"`
	ChatNotifications bool                            `json:"chat_notifications" db:"chat_notifications"`
	MomentNotifications bool                          `json:"moment_notifications" db:"moment_notifications"`
	SystemNotifications bool                          `json:"system_notifications" db:"system_notifications"`
	ActivityNotifications bool                        `json:"activity_notifications" db:"activity_notifications"`
	EmailEnabled      bool                            `json:"email_enabled" db:"email_enabled"`
	PushEnabled       bool                            `json:"push_enabled" db:"push_enabled"`
	SMSEnabled        bool                            `json:"sms_enabled" db:"sms_enabled"`
	QuietHours        *QuietHours                     `json:"quiet_hours" db:"quiet_hours"`
	Channels          map[NotificationType][]NotificationChannel `json:"channels" db:"channels"`
	CreatedAt         time.Time                       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time                       `json:"updated_at" db:"updated_at"`
}

// QuietHours 免打扰时间
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // HH:MM格式
	EndTime   string `json:"end_time"`   // HH:MM格式
	Timezone  string `json:"timezone"`   // 时区
}

// NotificationQueue 通知队列项
type NotificationQueue struct {
	ID            uuid.UUID            `json:"id" db:"id"`
	NotificationID uuid.UUID           `json:"notification_id" db:"notification_id"`
	Priority      NotificationPriority `json:"priority" db:"priority"`
	ScheduledAt   time.Time            `json:"scheduled_at" db:"scheduled_at"`
	ProcessedAt   *time.Time           `json:"processed_at" db:"processed_at"`
	Status        string               `json:"status" db:"status"` // queued, processing, completed, failed
	RetryCount    int                  `json:"retry_count" db:"retry_count"`
	MaxRetries    int                  `json:"max_retries" db:"max_retries"`
	ErrorMessage  *string              `json:"error_message" db:"error_message"`
	CreatedAt     time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" db:"updated_at"`
}

// WebSocketConnection WebSocket连接信息
type WebSocketConnection struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	SessionID  string    `json:"session_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPingAt time.Time `json:"last_ping_at"`
	UserAgent  string    `json:"user_agent"`
	IPAddress  string    `json:"ip_address"`
}

// NotificationEvent 通知事件
type NotificationEvent struct {
	ID          uuid.UUID       `json:"id"`
	Type        string          `json:"type"`        // notification.sent, notification.read, etc.
	UserID      uuid.UUID       `json:"user_id"`
	NotificationID uuid.UUID    `json:"notification_id"`
	Data        json.RawMessage `json:"data"`
	Timestamp   time.Time       `json:"timestamp"`
}

// PushNotificationRequest 推送通知请求
type PushNotificationRequest struct {
	UserID   uuid.UUID            `json:"user_id"`
	Type     NotificationType     `json:"type"`
	Priority NotificationPriority `json:"priority"`
	Title    string               `json:"title"`
	Content  string               `json:"content"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Channels []NotificationChannel `json:"channels,omitempty"`
	ScheduleAt *time.Time          `json:"schedule_at,omitempty"`
	ExpiresAt  *time.Time          `json:"expires_at,omitempty"`
}

// PushNotificationResponse 推送通知响应
type PushNotificationResponse struct {
	NotificationID uuid.UUID `json:"notification_id"`
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	SentChannels   []NotificationChannel `json:"sent_channels"`
	FailedChannels []NotificationChannel `json:"failed_channels"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalSent      int64                            `json:"total_sent"`
	TotalDelivered int64                            `json:"total_delivered"`
	TotalRead      int64                            `json:"total_read"`
	TotalFailed    int64                            `json:"total_failed"`
	ByType         map[NotificationType]int64       `json:"by_type"`
	ByChannel      map[NotificationChannel]int64    `json:"by_channel"`
	ByStatus       map[NotificationStatus]int64     `json:"by_status"`
	DeliveryRate   float64                          `json:"delivery_rate"`
	ReadRate       float64                          `json:"read_rate"`
}

// WebSocketMessage WebSocket消息
type WebSocketMessage struct {
	Type      string          `json:"type"`      // notification, ping, pong, etc.
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	MessageID string          `json:"message_id"`
}

// NotificationListRequest 通知列表请求
type NotificationListRequest struct {
	UserID     uuid.UUID         `json:"user_id"`
	Type       *NotificationType `json:"type,omitempty"`
	Status     *NotificationStatus `json:"status,omitempty"`
	Channel    *NotificationChannel `json:"channel,omitempty"`
	StartTime  *time.Time        `json:"start_time,omitempty"`
	EndTime    *time.Time        `json:"end_time,omitempty"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// NotificationListResponse 通知列表响应
type NotificationListResponse struct {
	Notifications []*Notification `json:"notifications"`
	Total         int             `json:"total"`
	Page          int             `json:"page"`
	Limit         int             `json:"limit"`
	HasMore       bool            `json:"has_more"`
}
