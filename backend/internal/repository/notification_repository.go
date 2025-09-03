package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"yunai/internal/domain"
)

// NotificationRepository 通知推送数据访问接口
type NotificationRepository interface {
	// 基础通知操作
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	GetNotification(ctx context.Context, notificationID uuid.UUID) (*domain.Notification, error)
	UpdateNotification(ctx context.Context, notification *domain.Notification) error
	DeleteNotification(ctx context.Context, notificationID uuid.UUID) error
	ListNotifications(ctx context.Context, req *domain.NotificationListRequest) ([]*domain.Notification, int, error)

	// 用户偏好管理
	CreateUserPreferences(ctx context.Context, prefs *domain.NotificationPreferences) error
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	UpdateUserPreferences(ctx context.Context, prefs *domain.NotificationPreferences) error
	DeleteUserPreferences(ctx context.Context, userID uuid.UUID) error

	// 模板管理
	CreateTemplate(ctx context.Context, template *domain.NotificationTemplate) error
	GetTemplate(ctx context.Context, templateID uuid.UUID) (*domain.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, template *domain.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, templateID uuid.UUID) error
	ListTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error)

	// 队列管理
	CreateQueueItem(ctx context.Context, queueItem *domain.NotificationQueue) error
	GetQueueItem(ctx context.Context, queueItemID uuid.UUID) (*domain.NotificationQueue, error)
	UpdateQueueItem(ctx context.Context, queueItem *domain.NotificationQueue) error
	DeleteQueueItem(ctx context.Context, queueItemID uuid.UUID) error
	GetPendingQueueItems(ctx context.Context, limit int) ([]*domain.NotificationQueue, error)
	GetQueueStats(ctx context.Context) (map[string]int, error)

	// 统计分析
	GetStats(ctx context.Context, userID *uuid.UUID, startTime, endTime time.Time) (*domain.NotificationStats, error)
	GetDeliveryReport(ctx context.Context, notificationID uuid.UUID) (*domain.NotificationEvent, error)
	CreateEvent(ctx context.Context, event *domain.NotificationEvent) error

	// 批量操作
	BatchCreateNotifications(ctx context.Context, notifications []*domain.Notification) error
	BatchUpdateNotifications(ctx context.Context, notifications []*domain.Notification) error
	BatchDeleteNotifications(ctx context.Context, notificationIDs []uuid.UUID) error

	// 清理操作
	CleanupExpiredNotifications(ctx context.Context) (int, error)
	CleanupOldNotifications(ctx context.Context, olderThan time.Time) (int, error)
}
