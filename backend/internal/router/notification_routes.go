package router

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/handler"
	"yunai/internal/service"
)

// SetupNotificationRoutes 设置通知推送路由
func SetupNotificationRoutes(
	router *gin.Engine,
	notificationService service.NotificationService,
	logger *logrus.Logger,
) {
	// 创建处理器
	notificationHandler := handler.NewNotificationHandler(notificationService, logger)
	webSocketHandler := handler.NewWebSocketHandler(notificationService, logger)

	// API路由组
	api := router.Group("/api/v1")
	{
		// 通知管理
		notifications := api.Group("/notifications")
		{
			notifications.POST("/send", notificationHandler.SendNotification)    // 发送通知
			notifications.GET("/:id", notificationHandler.GetNotification)       // 获取通知详情
			notifications.GET("", notificationHandler.ListNotifications)         // 获取通知列表
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)       // 标记为已读
			notifications.DELETE("/:id", notificationHandler.DeleteNotification) // 删除通知
		}

		// 用户偏好管理
		preferences := api.Group("/users/:user_id/notification-preferences")
		{
			preferences.GET("", notificationHandler.GetUserPreferences)    // 获取用户偏好
			preferences.PUT("", notificationHandler.UpdateUserPreferences) // 更新用户偏好
		}

		// 通知统计
		stats := api.Group("/notifications/stats")
		{
			stats.GET("", notificationHandler.GetNotificationStats) // 获取通知统计
		}
	}

	// WebSocket路由
	router.GET("/ws/notifications", webSocketHandler.HandleWebSocket) // WebSocket连接

	// 管理员路由组
	admin := router.Group("/api/v1/admin")
	{
		// 通知模板管理
		templates := admin.Group("/notification-templates")
		{
			templates.POST("", createNotificationTemplate(notificationService, logger))       // 创建模板
			templates.GET("", listNotificationTemplates(notificationService, logger))         // 列出模板
			templates.GET("/:id", getNotificationTemplate(notificationService, logger))       // 获取模板详情
			templates.PUT("/:id", updateNotificationTemplate(notificationService, logger))    // 更新模板
			templates.DELETE("/:id", deleteNotificationTemplate(notificationService, logger)) // 删除模板
		}

		// 系统管理
		system := admin.Group("/notifications/system")
		{
			system.GET("/queue-status", getQueueStatus(notificationService, logger))      // 获取队列状态
			system.POST("/process-queue", processQueue(notificationService, logger))      // 处理队列
			system.POST("/broadcast", broadcastNotification(notificationService, logger)) // 系统广播
		}
	}
}

// 管理员模板管理处理器

// createNotificationTemplate 创建通知模板
func createNotificationTemplate(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var template domain.NotificationTemplate
		if err := c.ShouldBindJSON(&template); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		err := notificationService.CreateNotificationTemplate(c.Request.Context(), &template)
		if err != nil {
			logger.WithError(err).Error("Failed to create notification template")
			c.JSON(500, gin.H{"error": "Failed to create template"})
			return
		}

		c.JSON(201, gin.H{"message": "Template created successfully", "template_id": template.ID})
	}
}

// listNotificationTemplates 列出通知模板
func listNotificationTemplates(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		templates, err := notificationService.ListNotificationTemplates(c.Request.Context())
		if err != nil {
			logger.WithError(err).Error("Failed to list notification templates")
			c.JSON(500, gin.H{"error": "Failed to list templates"})
			return
		}

		c.JSON(200, gin.H{"templates": templates, "count": len(templates)})
	}
}

// getNotificationTemplate 获取通知模板详情
func getNotificationTemplate(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		templateIDStr := c.Param("id")
		templateID, err := uuid.Parse(templateIDStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid template ID format"})
			return
		}

		template, err := notificationService.GetNotificationTemplate(c.Request.Context(), templateID)
		if err != nil {
			if err == domain.ErrNotificationTemplateNotFound {
				c.JSON(404, gin.H{"error": "Template not found"})
				return
			}
			logger.WithError(err).Error("Failed to get notification template")
			c.JSON(500, gin.H{"error": "Failed to get template"})
			return
		}

		c.JSON(200, template)
	}
}

// updateNotificationTemplate 更新通知模板
func updateNotificationTemplate(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		templateIDStr := c.Param("id")
		templateID, err := uuid.Parse(templateIDStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid template ID format"})
			return
		}

		var template domain.NotificationTemplate
		if err := c.ShouldBindJSON(&template); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		template.ID = templateID
		err = notificationService.UpdateNotificationTemplate(c.Request.Context(), &template)
		if err != nil {
			if err == domain.ErrNotificationTemplateNotFound {
				c.JSON(404, gin.H{"error": "Template not found"})
				return
			}
			logger.WithError(err).Error("Failed to update notification template")
			c.JSON(500, gin.H{"error": "Failed to update template"})
			return
		}

		c.JSON(200, gin.H{"message": "Template updated successfully"})
	}
}

// deleteNotificationTemplate 删除通知模板
func deleteNotificationTemplate(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		templateIDStr := c.Param("id")
		templateID, err := uuid.Parse(templateIDStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid template ID format"})
			return
		}

		err = notificationService.DeleteNotificationTemplate(c.Request.Context(), templateID)
		if err != nil {
			if err == domain.ErrNotificationTemplateNotFound {
				c.JSON(404, gin.H{"error": "Template not found"})
				return
			}
			logger.WithError(err).Error("Failed to delete notification template")
			c.JSON(500, gin.H{"error": "Failed to delete template"})
			return
		}

		c.JSON(200, gin.H{"message": "Template deleted successfully"})
	}
}

// 系统管理处理器

// getQueueStatus 获取队列状态
func getQueueStatus(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := notificationService.GetQueueStatus(c.Request.Context())
		if err != nil {
			logger.WithError(err).Error("Failed to get queue status")
			c.JSON(500, gin.H{"error": "Failed to get queue status"})
			return
		}

		c.JSON(200, gin.H{"queue_status": status})
	}
}

// processQueue 处理队列
func processQueue(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := notificationService.ProcessNotificationQueue(c.Request.Context())
		if err != nil {
			logger.WithError(err).Error("Failed to process notification queue")
			c.JSON(500, gin.H{"error": "Failed to process queue"})
			return
		}

		c.JSON(200, gin.H{"message": "Queue processed successfully"})
	}
}

// broadcastNotification 系统广播
func broadcastNotification(notificationService service.NotificationService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Type    string                 `json:"type"`
			Title   string                 `json:"title"`
			Content string                 `json:"content"`
			Data    map[string]interface{} `json:"data,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		if req.Title == "" || req.Content == "" {
			c.JSON(400, gin.H{"error": "Title and content are required"})
			return
		}

		// 创建广播消息
		message := &domain.WebSocketMessage{
			Type:      req.Type,
			Timestamp: time.Now(),
			MessageID: uuid.New().String(),
		}

		if req.Data != nil {
			dataJSON, err := json.Marshal(map[string]interface{}{
				"title":   req.Title,
				"content": req.Content,
				"data":    req.Data,
			})
			if err != nil {
				c.JSON(400, gin.H{"error": "Invalid data format"})
				return
			}
			message.Data = json.RawMessage(dataJSON)
		}

		err := notificationService.BroadcastToAll(message)
		if err != nil {
			logger.WithError(err).Error("Failed to broadcast notification")
			c.JSON(500, gin.H{"error": "Failed to broadcast notification"})
			return
		}

		c.JSON(200, gin.H{"message": "Notification broadcasted successfully"})
	}
}
