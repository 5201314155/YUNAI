package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/service"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	notificationService service.NotificationService
	logger              *logrus.Logger
	upgrader            websocket.Upgrader
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(
	notificationService service.NotificationService,
	logger *logrus.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		notificationService: notificationService,
		logger:              logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 在生产环境中应该检查Origin
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 获取用户ID
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	// 生成会话ID
	sessionID := uuid.New().String()

	// 升级HTTP连接为WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	// 注册连接
	err = h.notificationService.RegisterWebSocketConnection(userID, conn, sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to register WebSocket connection")
		return
	}
	defer h.notificationService.UnregisterWebSocketConnection(userID, sessionID)

	h.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"session_id":  sessionID,
		"remote_addr": c.Request.RemoteAddr,
	}).Info("WebSocket connection established")

	// 发送连接成功消息
	welcomeMsg := &domain.WebSocketMessage{
		Type:      "connected",
		Data:      json.RawMessage(`{"message": "WebSocket connected successfully"}`),
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}

	err = conn.WriteJSON(welcomeMsg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to send welcome message")
		return
	}

	// 发送未读通知
	h.sendUnreadNotifications(userID, conn)

	// 保持连接活跃
	h.keepConnectionAlive(userID, sessionID, conn)
}

// sendUnreadNotifications 发送未读通知
func (h *WebSocketHandler) sendUnreadNotifications(userID uuid.UUID, conn *websocket.Conn) {
	ctx := context.Background()

	// 获取未读通知
	req := &domain.NotificationListRequest{
		UserID: userID,
		Status: func() *domain.NotificationStatus { s := domain.NotificationStatusSent; return &s }(),
		Limit:  50,
		Offset: 0,
	}

	response, err := h.notificationService.ListNotifications(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get unread notifications")
		return
	}

	if len(response.Notifications) == 0 {
		return
	}

	// 发送未读通知列表
	unreadMsg := &domain.WebSocketMessage{
		Type:      "unread_notifications",
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}

	data := map[string]interface{}{
		"notifications": response.Notifications,
		"total":         response.Total,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal unread notifications")
		return
	}

	unreadMsg.Data = json.RawMessage(dataJSON)

	err = conn.WriteJSON(unreadMsg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to send unread notifications")
	}
}

// keepConnectionAlive 保持连接活跃
func (h *WebSocketHandler) keepConnectionAlive(userID uuid.UUID, sessionID string, conn *websocket.Conn) {
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 设置pong处理器
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动ping定时器
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 启动ping goroutine
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					h.logger.WithError(err).WithFields(logrus.Fields{
						"user_id":    userID,
						"session_id": sessionID,
					}).Debug("Failed to send ping, connection likely closed")
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
				h.logger.WithError(err).WithFields(logrus.Fields{
					"user_id":    userID,
					"session_id": sessionID,
				}).Error("WebSocket connection error")
			}
			break
		}

		// 处理客户端消息
		h.handleClientMessage(userID, sessionID, conn, &msg)
	}
}

// handleClientMessage 处理客户端消息
func (h *WebSocketHandler) handleClientMessage(userID uuid.UUID, sessionID string, conn *websocket.Conn, msg *domain.WebSocketMessage) {
	ctx := context.Background()

	switch msg.Type {
	case "ping":
		// 响应ping
		pongMsg := &domain.WebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
			MessageID: msg.MessageID,
		}
		conn.WriteJSON(pongMsg)

	case "mark_read":
		// 标记通知为已读
		var data struct {
			NotificationID uuid.UUID `json:"notification_id"`
		}
		if err := json.Unmarshal(msg.Data, &data); err == nil {
			err = h.notificationService.MarkAsRead(ctx, data.NotificationID, userID)
			if err != nil {
				h.logger.WithError(err).Error("Failed to mark notification as read")
			} else {
				// 发送确认消息
				ackMsg := &domain.WebSocketMessage{
					Type:      "read_ack",
					Data:      json.RawMessage(fmt.Sprintf(`{"notification_id": "%s", "status": "read"}`, data.NotificationID)),
					Timestamp: time.Now(),
					MessageID: uuid.New().String(),
				}
				conn.WriteJSON(ackMsg)
			}
		}

	case "get_notifications":
		// 获取通知列表
		var data struct {
			Type   *domain.NotificationType   `json:"type,omitempty"`
			Status *domain.NotificationStatus `json:"status,omitempty"`
			Limit  int                        `json:"limit"`
			Offset int                        `json:"offset"`
		}

		if err := json.Unmarshal(msg.Data, &data); err == nil {
			req := &domain.NotificationListRequest{
				UserID: userID,
				Type:   data.Type,
				Status: data.Status,
				Limit:  data.Limit,
				Offset: data.Offset,
			}

			if req.Limit == 0 {
				req.Limit = 20
			}

			response, err := h.notificationService.ListNotifications(ctx, req)
			if err != nil {
				h.logger.WithError(err).Error("Failed to get notifications")
			} else {
				responseMsg := &domain.WebSocketMessage{
					Type:      "notifications_list",
					Timestamp: time.Now(),
					MessageID: uuid.New().String(),
				}

				responseData, _ := json.Marshal(response)
				responseMsg.Data = json.RawMessage(responseData)
				conn.WriteJSON(responseMsg)
			}
		}

	case "update_preferences":
		// 更新通知偏好
		var prefs domain.NotificationPreferences
		if err := json.Unmarshal(msg.Data, &prefs); err == nil {
			prefs.UserID = userID
			prefs.UpdatedAt = time.Now()

			err = h.notificationService.UpdateUserNotificationPreferences(ctx, &prefs)
			if err != nil {
				h.logger.WithError(err).Error("Failed to update notification preferences")
			} else {
				// 发送确认消息
				ackMsg := &domain.WebSocketMessage{
					Type:      "preferences_updated",
					Data:      json.RawMessage(`{"status": "success"}`),
					Timestamp: time.Now(),
					MessageID: uuid.New().String(),
				}
				conn.WriteJSON(ackMsg)
			}
		}

	default:
		h.logger.WithFields(logrus.Fields{
			"user_id":    userID,
			"session_id": sessionID,
			"msg_type":   msg.Type,
		}).Debug("Received unknown WebSocket message type")
	}
}
