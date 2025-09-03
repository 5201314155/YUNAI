package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// NotificationService 通知推送服务接口
type NotificationService interface {
	// 基础通知功能
	SendNotification(ctx context.Context, req *domain.PushNotificationRequest) (*domain.PushNotificationResponse, error)
	GetNotification(ctx context.Context, notificationID uuid.UUID) (*domain.Notification, error)
	ListNotifications(ctx context.Context, req *domain.NotificationListRequest) (*domain.NotificationListResponse, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error

	// WebSocket实时推送
	RegisterWebSocketConnection(userID uuid.UUID, conn *websocket.Conn, sessionID string) error
	UnregisterWebSocketConnection(userID uuid.UUID, sessionID string) error
	BroadcastToUser(userID uuid.UUID, message *domain.WebSocketMessage) error
	BroadcastToAll(message *domain.WebSocketMessage) error

	// 推送策略
	ScheduleNotification(ctx context.Context, notification *domain.Notification, scheduleTime time.Time) error
	GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	UpdateUserNotificationPreferences(ctx context.Context, prefs *domain.NotificationPreferences) error

	// 消息队列
	QueueNotification(ctx context.Context, notification *domain.Notification) error
	ProcessNotificationQueue(ctx context.Context) error
	GetQueueStatus(ctx context.Context) (map[string]int, error)

	// 模板管理
	CreateNotificationTemplate(ctx context.Context, template *domain.NotificationTemplate) error
	GetNotificationTemplate(ctx context.Context, templateID uuid.UUID) (*domain.NotificationTemplate, error)
	UpdateNotificationTemplate(ctx context.Context, template *domain.NotificationTemplate) error
	DeleteNotificationTemplate(ctx context.Context, templateID uuid.UUID) error
	ListNotificationTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error)
	RenderNotification(ctx context.Context, templateID uuid.UUID, data map[string]interface{}) (*domain.Notification, error)

	// 统计分析
	GetNotificationStats(ctx context.Context, userID *uuid.UUID, startTime, endTime time.Time) (*domain.NotificationStats, error)
	GetDeliveryReport(ctx context.Context, notificationID uuid.UUID) (*domain.NotificationEvent, error)
}

// notificationService 通知推送服务实现
type notificationService struct {
	notificationRepo repository.NotificationRepository
	logger           *logrus.Logger

	// WebSocket连接管理
	connections map[uuid.UUID]map[string]*websocket.Conn // userID -> sessionID -> connection
	connMutex   sync.RWMutex

	// 消息队列
	messageQueue chan *domain.Notification
	queueWorkers int

	// 推送策略配置
	config *NotificationConfig
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	QueueSize        int           `json:"queue_size"`
	WorkerCount      int           `json:"worker_count"`
	WebSocketTimeout time.Duration `json:"websocket_timeout"`
	PingInterval     time.Duration `json:"ping_interval"`
	MaxConnections   int           `json:"max_connections"`
	EnableRateLimit  bool          `json:"enable_rate_limit"`
	RateLimitPerHour int           `json:"rate_limit_per_hour"`
}

// NewNotificationService 创建通知推送服务
func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	logger *logrus.Logger,
) NotificationService {
	config := &NotificationConfig{
		MaxRetries:       3,
		RetryDelay:       time.Minute * 5,
		QueueSize:        10000,
		WorkerCount:      10,
		WebSocketTimeout: time.Second * 30,
		PingInterval:     time.Second * 30,
		MaxConnections:   1000,
		EnableRateLimit:  true,
		RateLimitPerHour: 100,
	}

	service := &notificationService{
		notificationRepo: notificationRepo,
		logger:           logger,
		connections:      make(map[uuid.UUID]map[string]*websocket.Conn),
		messageQueue:     make(chan *domain.Notification, config.QueueSize),
		queueWorkers:     config.WorkerCount,
		config:           config,
	}

	// 启动队列处理器
	service.startQueueWorkers()

	return service
}

// SendNotification 发送通知
func (s *notificationService) SendNotification(ctx context.Context, req *domain.PushNotificationRequest) (*domain.PushNotificationResponse, error) {
	// 创建通知记录
	notification := &domain.Notification{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Type:        req.Type,
		Priority:    req.Priority,
		Status:      domain.NotificationStatusPending,
		Title:       req.Title,
		Content:     req.Content,
		ScheduledAt: req.ScheduleAt,
		ExpiresAt:   req.ExpiresAt,
		MaxRetries:  s.config.MaxRetries,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 设置额外数据
	if req.Data != nil {
		dataJSON, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal notification data: %w", err)
		}
		notification.Data = json.RawMessage(dataJSON)
	} else {
		// 设置空的JSON对象
		notification.Data = json.RawMessage("{}")
	}

	// 保存到数据库
	err := s.notificationRepo.CreateNotification(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// 获取用户通知偏好
	prefs, err := s.GetUserNotificationPreferences(ctx, req.UserID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user notification preferences")
		// 使用默认偏好
		prefs = s.getDefaultPreferences(req.UserID)
	}

	// 确定发送渠道
	channels := s.determineChannels(req, prefs)

	var sentChannels []domain.NotificationChannel
	var failedChannels []domain.NotificationChannel

	// 发送到各个渠道
	for _, channel := range channels {
		notification.Channel = channel

		if req.ScheduleAt != nil && req.ScheduleAt.After(time.Now()) {
			// 定时发送，加入队列
			err = s.QueueNotification(ctx, notification)
		} else {
			// 立即发送
			err = s.sendToChannel(ctx, notification, channel)
		}

		if err != nil {
			s.logger.WithError(err).WithField("channel", channel).Error("Failed to send notification")
			failedChannels = append(failedChannels, channel)
		} else {
			sentChannels = append(sentChannels, channel)
		}
	}

	// 更新通知状态
	if len(sentChannels) > 0 {
		notification.Status = domain.NotificationStatusSent
		notification.SentAt = &[]time.Time{time.Now()}[0]
	} else {
		notification.Status = domain.NotificationStatusFailed
	}
	notification.UpdatedAt = time.Now()

	err = s.notificationRepo.UpdateNotification(ctx, notification)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update notification status")
	}

	return &domain.PushNotificationResponse{
		NotificationID: notification.ID,
		Status:         string(notification.Status),
		Message:        "Notification processed",
		SentChannels:   sentChannels,
		FailedChannels: failedChannels,
	}, nil
}

// GetNotification 获取通知
func (s *notificationService) GetNotification(ctx context.Context, notificationID uuid.UUID) (*domain.Notification, error) {
	return s.notificationRepo.GetNotification(ctx, notificationID)
}

// ListNotifications 获取通知列表
func (s *notificationService) ListNotifications(ctx context.Context, req *domain.NotificationListRequest) (*domain.NotificationListResponse, error) {
	notifications, total, err := s.notificationRepo.ListNotifications(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	page := req.Offset/req.Limit + 1
	hasMore := req.Offset+req.Limit < total

	return &domain.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		Limit:         req.Limit,
		HasMore:       hasMore,
	}, nil
}

// MarkAsRead 标记为已读
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	notification, err := s.notificationRepo.GetNotification(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	if notification.Status == domain.NotificationStatusRead {
		return nil // 已经是已读状态
	}

	notification.Status = domain.NotificationStatusRead
	now := time.Now()
	notification.ReadAt = &now
	notification.UpdatedAt = now

	return s.notificationRepo.UpdateNotification(ctx, notification)
}

// DeleteNotification 删除通知
func (s *notificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	notification, err := s.notificationRepo.GetNotification(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	return s.notificationRepo.DeleteNotification(ctx, notificationID)
}

// RegisterWebSocketConnection 注册WebSocket连接
func (s *notificationService) RegisterWebSocketConnection(userID uuid.UUID, conn *websocket.Conn, sessionID string) error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if s.connections[userID] == nil {
		s.connections[userID] = make(map[string]*websocket.Conn)
	}

	// 检查连接数限制
	if len(s.connections[userID]) >= s.config.MaxConnections {
		return fmt.Errorf("maximum connections exceeded for user")
	}

	s.connections[userID][sessionID] = conn

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": sessionID,
	}).Info("WebSocket connection registered")

	// 启动连接心跳检测
	go s.handleWebSocketConnection(userID, sessionID, conn)

	return nil
}

// UnregisterWebSocketConnection 注销WebSocket连接
func (s *notificationService) UnregisterWebSocketConnection(userID uuid.UUID, sessionID string) error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if s.connections[userID] != nil {
		delete(s.connections[userID], sessionID)

		// 如果用户没有其他连接，删除用户条目
		if len(s.connections[userID]) == 0 {
			delete(s.connections, userID)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": sessionID,
	}).Info("WebSocket connection unregistered")

	return nil
}

// BroadcastToUser 向特定用户广播消息
func (s *notificationService) BroadcastToUser(userID uuid.UUID, message *domain.WebSocketMessage) error {
	s.connMutex.RLock()
	userConnections := s.connections[userID]
	s.connMutex.RUnlock()

	if userConnections == nil || len(userConnections) == 0 {
		return fmt.Errorf("no active connections for user %s", userID)
	}

	var errors []error
	for sessionID, conn := range userConnections {
		err := conn.WriteJSON(message)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":    userID,
				"session_id": sessionID,
			}).Error("Failed to send WebSocket message")

			// 移除失效连接
			s.UnregisterWebSocketConnection(userID, sessionID)
			errors = append(errors, err)
		}
	}

	if len(errors) == len(userConnections) {
		return fmt.Errorf("failed to send to all connections")
	}

	return nil
}

// BroadcastToAll 向所有用户广播消息
func (s *notificationService) BroadcastToAll(message *domain.WebSocketMessage) error {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	var totalConnections int
	var failedConnections int

	for userID, userConnections := range s.connections {
		for sessionID, conn := range userConnections {
			totalConnections++
			err := conn.WriteJSON(message)
			if err != nil {
				failedConnections++
				s.logger.WithError(err).WithFields(logrus.Fields{
					"user_id":    userID,
					"session_id": sessionID,
				}).Error("Failed to broadcast WebSocket message")

				// 移除失效连接
				go s.UnregisterWebSocketConnection(userID, sessionID)
			}
		}
	}

	s.logger.WithFields(logrus.Fields{
		"total_connections":  totalConnections,
		"failed_connections": failedConnections,
		"success_rate":       float64(totalConnections-failedConnections) / float64(totalConnections) * 100,
	}).Info("Broadcast completed")

	return nil
}

// handleWebSocketConnection 处理WebSocket连接
func (s *notificationService) handleWebSocketConnection(userID uuid.UUID, sessionID string, conn *websocket.Conn) {
	defer s.UnregisterWebSocketConnection(userID, sessionID)

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(s.config.WebSocketTimeout))

	// 设置pong处理器
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(s.config.WebSocketTimeout))
		return nil
	})

	// 启动ping定时器
	ticker := time.NewTicker(s.config.PingInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// 读取消息循环
	for {
		var msg domain.WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.WithError(err).Error("WebSocket connection error")
			}
			break
		}

		// 处理客户端消息
		s.handleWebSocketMessage(userID, sessionID, &msg)
	}
}

// handleWebSocketMessage 处理WebSocket消息
func (s *notificationService) handleWebSocketMessage(userID uuid.UUID, sessionID string, msg *domain.WebSocketMessage) {
	switch msg.Type {
	case "ping":
		// 响应ping
		pongMsg := &domain.WebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
			MessageID: msg.MessageID,
		}
		s.BroadcastToUser(userID, pongMsg)

	case "mark_read":
		// 标记通知为已读
		var data struct {
			NotificationID uuid.UUID `json:"notification_id"`
		}
		if err := json.Unmarshal(msg.Data, &data); err == nil {
			s.MarkAsRead(context.Background(), data.NotificationID, userID)
		}

	default:
		s.logger.WithFields(logrus.Fields{
			"user_id":    userID,
			"session_id": sessionID,
			"msg_type":   msg.Type,
		}).Debug("Received WebSocket message")
	}
}

// ScheduleNotification 定时发送通知
func (s *notificationService) ScheduleNotification(ctx context.Context, notification *domain.Notification, scheduleTime time.Time) error {
	notification.ScheduledAt = &scheduleTime
	notification.Status = domain.NotificationStatusPending

	return s.QueueNotification(ctx, notification)
}

// GetUserNotificationPreferences 获取用户通知偏好
func (s *notificationService) GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	prefs, err := s.notificationRepo.GetUserPreferences(ctx, userID)
	if err != nil {
		// 如果没有找到偏好设置，返回默认设置
		return s.getDefaultPreferences(userID), nil
	}
	return prefs, nil
}

// UpdateUserNotificationPreferences 更新用户通知偏好
func (s *notificationService) UpdateUserNotificationPreferences(ctx context.Context, prefs *domain.NotificationPreferences) error {
	prefs.UpdatedAt = time.Now()

	// 尝试更新，如果不存在则创建
	err := s.notificationRepo.UpdateUserPreferences(ctx, prefs)
	if err == domain.ErrNotificationPreferencesNotFound {
		// 偏好不存在，创建新的
		prefs.ID = uuid.New()
		prefs.CreatedAt = time.Now()
		return s.notificationRepo.CreateUserPreferences(ctx, prefs)
	}

	return err
}

// QueueNotification 将通知加入队列
func (s *notificationService) QueueNotification(ctx context.Context, notification *domain.Notification) error {
	// 创建队列项
	queueItem := &domain.NotificationQueue{
		ID:             uuid.New(),
		NotificationID: notification.ID,
		Priority:       notification.Priority,
		ScheduledAt:    time.Now(),
		Status:         "queued",
		MaxRetries:     notification.MaxRetries,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if notification.ScheduledAt != nil {
		queueItem.ScheduledAt = *notification.ScheduledAt
	}

	// 保存队列项到数据库
	err := s.notificationRepo.CreateQueueItem(ctx, queueItem)
	if err != nil {
		return fmt.Errorf("failed to create queue item: %w", err)
	}

	// 如果是立即发送，加入内存队列
	if notification.ScheduledAt == nil || notification.ScheduledAt.Before(time.Now().Add(time.Minute)) {
		select {
		case s.messageQueue <- notification:
			s.logger.WithField("notification_id", notification.ID).Debug("Notification queued for immediate processing")
		default:
			s.logger.WithField("notification_id", notification.ID).Warn("Message queue is full, notification will be processed by scheduled workers")
		}
	}

	return nil
}

// ProcessNotificationQueue 处理通知队列
func (s *notificationService) ProcessNotificationQueue(ctx context.Context) error {
	// 获取待处理的队列项
	queueItems, err := s.notificationRepo.GetPendingQueueItems(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending queue items: %w", err)
	}

	for _, item := range queueItems {
		// 检查是否到了处理时间
		if item.ScheduledAt.After(time.Now()) {
			continue
		}

		// 获取通知详情
		notification, err := s.notificationRepo.GetNotification(ctx, item.NotificationID)
		if err != nil {
			s.logger.WithError(err).WithField("notification_id", item.NotificationID).Error("Failed to get notification")
			continue
		}

		// 处理通知
		err = s.processQueuedNotification(ctx, notification, item)
		if err != nil {
			s.logger.WithError(err).WithField("notification_id", notification.ID).Error("Failed to process queued notification")
		}
	}

	return nil
}

// GetQueueStatus 获取队列状态
func (s *notificationService) GetQueueStatus(ctx context.Context) (map[string]int, error) {
	stats, err := s.notificationRepo.GetQueueStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	status := map[string]int{
		"queued":     stats["queued"],
		"processing": stats["processing"],
		"completed":  stats["completed"],
		"failed":     stats["failed"],
		"in_memory":  len(s.messageQueue),
	}

	return status, nil
}

// CreateNotificationTemplate 创建通知模板
func (s *notificationService) CreateNotificationTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	return s.notificationRepo.CreateTemplate(ctx, template)
}

// GetNotificationTemplate 获取通知模板
func (s *notificationService) GetNotificationTemplate(ctx context.Context, templateID uuid.UUID) (*domain.NotificationTemplate, error) {
	return s.notificationRepo.GetTemplate(ctx, templateID)
}

// UpdateNotificationTemplate 更新通知模板
func (s *notificationService) UpdateNotificationTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	template.UpdatedAt = time.Now()
	return s.notificationRepo.UpdateTemplate(ctx, template)
}

// DeleteNotificationTemplate 删除通知模板
func (s *notificationService) DeleteNotificationTemplate(ctx context.Context, templateID uuid.UUID) error {
	return s.notificationRepo.DeleteTemplate(ctx, templateID)
}

// ListNotificationTemplates 列出通知模板
func (s *notificationService) ListNotificationTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error) {
	return s.notificationRepo.ListTemplates(ctx)
}

// RenderNotification 渲染通知模板
func (s *notificationService) RenderNotification(ctx context.Context, templateID uuid.UUID, data map[string]interface{}) (*domain.Notification, error) {
	template, err := s.notificationRepo.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if !template.Enabled {
		return nil, fmt.Errorf("template is disabled")
	}

	// 渲染标题和内容
	title := s.renderTemplate(template.Title, data)
	content := s.renderTemplate(template.Content, data)

	notification := &domain.Notification{
		ID:         uuid.New(),
		Type:       template.Type,
		Channel:    template.Channel,
		Priority:   domain.NotificationPriorityNormal,
		Status:     domain.NotificationStatusPending,
		Title:      title,
		Content:    content,
		TemplateID: &templateID,
		MaxRetries: s.config.MaxRetries,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 设置额外数据
	if len(data) > 0 {
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal template data: %w", err)
		}
		notification.Data = json.RawMessage(dataJSON)
	} else {
		// 设置空的JSON对象
		notification.Data = json.RawMessage("{}")
	}

	return notification, nil
}

// GetNotificationStats 获取通知统计
func (s *notificationService) GetNotificationStats(ctx context.Context, userID *uuid.UUID, startTime, endTime time.Time) (*domain.NotificationStats, error) {
	return s.notificationRepo.GetStats(ctx, userID, startTime, endTime)
}

// GetDeliveryReport 获取投递报告
func (s *notificationService) GetDeliveryReport(ctx context.Context, notificationID uuid.UUID) (*domain.NotificationEvent, error) {
	return s.notificationRepo.GetDeliveryReport(ctx, notificationID)
}

// 辅助方法

// getDefaultPreferences 获取默认通知偏好
func (s *notificationService) getDefaultPreferences(userID uuid.UUID) *domain.NotificationPreferences {
	return &domain.NotificationPreferences{
		ID:                    uuid.New(),
		UserID:                userID,
		ChatNotifications:     true,
		MomentNotifications:   true,
		SystemNotifications:   true,
		ActivityNotifications: true,
		EmailEnabled:          false,
		PushEnabled:           true,
		SMSEnabled:            false,
		QuietHours: &domain.QuietHours{
			Enabled:   false,
			StartTime: "22:00",
			EndTime:   "08:00",
			Timezone:  "Asia/Shanghai",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// determineChannels 确定发送渠道
func (s *notificationService) determineChannels(req *domain.PushNotificationRequest, prefs *domain.NotificationPreferences) []domain.NotificationChannel {
	var channels []domain.NotificationChannel

	// 如果请求中指定了渠道，使用指定的渠道
	if len(req.Channels) > 0 {
		for _, channel := range req.Channels {
			if s.isChannelEnabled(channel, prefs) {
				channels = append(channels, channel)
			}
		}
		return channels
	}

	// 根据通知类型和用户偏好确定渠道
	switch req.Type {
	case domain.NotificationTypeChat:
		if prefs.ChatNotifications {
			channels = append(channels, domain.NotificationChannelWebSocket)
			if prefs.PushEnabled {
				channels = append(channels, domain.NotificationChannelPush)
			}
		}
	case domain.NotificationTypeMoment:
		if prefs.MomentNotifications {
			channels = append(channels, domain.NotificationChannelInApp)
			if prefs.PushEnabled {
				channels = append(channels, domain.NotificationChannelPush)
			}
		}
	case domain.NotificationTypeSystem:
		if prefs.SystemNotifications {
			channels = append(channels, domain.NotificationChannelInApp)
			if prefs.EmailEnabled {
				channels = append(channels, domain.NotificationChannelEmail)
			}
		}
	case domain.NotificationTypeActivity:
		if prefs.ActivityNotifications {
			channels = append(channels, domain.NotificationChannelInApp)
			if prefs.PushEnabled {
				channels = append(channels, domain.NotificationChannelPush)
			}
		}
	default:
		channels = append(channels, domain.NotificationChannelInApp)
	}

	// 检查免打扰时间
	if s.isQuietTime(prefs.QuietHours) {
		// 在免打扰时间内，只发送紧急通知
		if req.Priority != domain.NotificationPriorityCritical {
			return []domain.NotificationChannel{domain.NotificationChannelInApp}
		}
	}

	return channels
}

// sendToChannel 发送到指定渠道
func (s *notificationService) sendToChannel(ctx context.Context, notification *domain.Notification, channel domain.NotificationChannel) error {
	switch channel {
	case domain.NotificationChannelWebSocket:
		return s.sendWebSocketNotification(notification)
	case domain.NotificationChannelInApp:
		return s.sendInAppNotification(ctx, notification)
	case domain.NotificationChannelPush:
		return s.sendPushNotification(ctx, notification)
	case domain.NotificationChannelEmail:
		return s.sendEmailNotification(ctx, notification)
	case domain.NotificationChannelSMS:
		return s.sendSMSNotification(ctx, notification)
	default:
		return fmt.Errorf("unsupported notification channel: %s", channel)
	}
}

// sendWebSocketNotification 发送WebSocket通知
func (s *notificationService) sendWebSocketNotification(notification *domain.Notification) error {
	message := &domain.WebSocketMessage{
		Type:      "notification",
		Data:      notification.Data,
		Timestamp: time.Now(),
		MessageID: notification.ID.String(),
	}

	return s.BroadcastToUser(notification.UserID, message)
}

// sendInAppNotification 发送应用内通知
func (s *notificationService) sendInAppNotification(ctx context.Context, notification *domain.Notification) error {
	// 应用内通知只需要保存到数据库即可
	notification.Status = domain.NotificationStatusDelivered
	now := time.Now()
	notification.SentAt = &now
	return s.notificationRepo.UpdateNotification(ctx, notification)
}

// sendPushNotification 发送推送通知
func (s *notificationService) sendPushNotification(ctx context.Context, notification *domain.Notification) error {
	// 简化实现，实际应该调用推送服务API
	s.logger.WithFields(logrus.Fields{
		"user_id": notification.UserID,
		"title":   notification.Title,
		"content": notification.Content,
	}).Info("📱 Push notification sent (mock)")

	notification.Status = domain.NotificationStatusSent
	now := time.Now()
	notification.SentAt = &now
	return s.notificationRepo.UpdateNotification(ctx, notification)
}

// sendEmailNotification 发送邮件通知
func (s *notificationService) sendEmailNotification(ctx context.Context, notification *domain.Notification) error {
	// 简化实现，实际应该调用邮件服务
	s.logger.WithFields(logrus.Fields{
		"user_id": notification.UserID,
		"title":   notification.Title,
		"content": notification.Content,
	}).Info("📧 Email notification sent (mock)")

	notification.Status = domain.NotificationStatusSent
	now := time.Now()
	notification.SentAt = &now
	return s.notificationRepo.UpdateNotification(ctx, notification)
}

// sendSMSNotification 发送短信通知
func (s *notificationService) sendSMSNotification(ctx context.Context, notification *domain.Notification) error {
	// 简化实现，实际应该调用短信服务
	s.logger.WithFields(logrus.Fields{
		"user_id": notification.UserID,
		"title":   notification.Title,
		"content": notification.Content,
	}).Info("📱 SMS notification sent (mock)")

	notification.Status = domain.NotificationStatusSent
	now := time.Now()
	notification.SentAt = &now
	return s.notificationRepo.UpdateNotification(ctx, notification)
}

// renderTemplate 渲染模板
func (s *notificationService) renderTemplate(template string, data map[string]interface{}) string {
	// 简化的模板渲染实现
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.Replace(result, placeholder, fmt.Sprintf("%v", value), -1)
	}
	return result
}

// isChannelEnabled 检查渠道是否启用
func (s *notificationService) isChannelEnabled(channel domain.NotificationChannel, prefs *domain.NotificationPreferences) bool {
	switch channel {
	case domain.NotificationChannelEmail:
		return prefs.EmailEnabled
	case domain.NotificationChannelPush:
		return prefs.PushEnabled
	case domain.NotificationChannelSMS:
		return prefs.SMSEnabled
	default:
		return true // 应用内和WebSocket默认启用
	}
}

// isQuietTime 检查是否在免打扰时间
func (s *notificationService) isQuietTime(quietHours *domain.QuietHours) bool {
	if quietHours == nil || !quietHours.Enabled {
		return false
	}

	// 简化实现，实际应该考虑时区
	now := time.Now()
	currentTime := now.Format("15:04")

	return currentTime >= quietHours.StartTime && currentTime <= quietHours.EndTime
}

// processQueuedNotification 处理队列中的通知
func (s *notificationService) processQueuedNotification(ctx context.Context, notification *domain.Notification, queueItem *domain.NotificationQueue) error {
	// 更新队列项状态
	queueItem.Status = "processing"
	queueItem.ProcessedAt = &[]time.Time{time.Now()}[0]
	queueItem.UpdatedAt = time.Now()

	err := s.notificationRepo.UpdateQueueItem(ctx, queueItem)
	if err != nil {
		return fmt.Errorf("failed to update queue item: %w", err)
	}

	// 发送通知
	err = s.sendToChannel(ctx, notification, notification.Channel)
	if err != nil {
		// 处理失败，增加重试次数
		queueItem.RetryCount++
		queueItem.Status = "failed"
		errMsg := err.Error()
		queueItem.ErrorMessage = &errMsg

		if queueItem.RetryCount < queueItem.MaxRetries {
			// 重新调度
			queueItem.ScheduledAt = time.Now().Add(s.config.RetryDelay)
			queueItem.Status = "queued"
		}

		s.notificationRepo.UpdateQueueItem(ctx, queueItem)
		return err
	}

	// 处理成功
	queueItem.Status = "completed"
	queueItem.UpdatedAt = time.Now()

	return s.notificationRepo.UpdateQueueItem(ctx, queueItem)
}

// startQueueWorkers 启动队列处理器
func (s *notificationService) startQueueWorkers() {
	for i := 0; i < s.queueWorkers; i++ {
		go s.queueWorker(i)
	}

	s.logger.WithField("worker_count", s.queueWorkers).Info("Notification queue workers started")
}

// queueWorker 队列处理器
func (s *notificationService) queueWorker(workerID int) {
	s.logger.WithField("worker_id", workerID).Info("Notification queue worker started")

	for notification := range s.messageQueue {
		ctx := context.Background()

		s.logger.WithFields(logrus.Fields{
			"worker_id":       workerID,
			"notification_id": notification.ID,
			"user_id":         notification.UserID,
			"type":            notification.Type,
		}).Debug("Processing notification")

		err := s.sendToChannel(ctx, notification, notification.Channel)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"worker_id":       workerID,
				"notification_id": notification.ID,
			}).Error("Failed to process notification")

			// 增加重试次数
			notification.RetryCount++
			if notification.RetryCount < notification.MaxRetries {
				// 重新加入队列
				time.AfterFunc(s.config.RetryDelay, func() {
					select {
					case s.messageQueue <- notification:
					default:
						s.logger.WithField("notification_id", notification.ID).Warn("Failed to requeue notification")
					}
				})
			} else {
				// 达到最大重试次数，标记为失败
				notification.Status = domain.NotificationStatusFailed
				s.notificationRepo.UpdateNotification(ctx, notification)
			}
		} else {
			s.logger.WithFields(logrus.Fields{
				"worker_id":       workerID,
				"notification_id": notification.ID,
			}).Debug("Notification processed successfully")
		}
	}
}
