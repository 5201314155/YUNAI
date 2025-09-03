package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// notificationRepository 通知推送数据访问实现
type notificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository 创建通知推送数据访问实例
func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

// CreateNotification 创建通知
func (r *notificationRepository) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, channel, priority, status, title, content, 
		                          data, template_id, scheduled_at, sent_at, read_at, expires_at, 
		                          retry_count, max_retries, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	// 处理空的Data字段
	var dataToStore interface{}
	if len(notification.Data) == 0 {
		dataToStore = nil
	} else {
		dataToStore = notification.Data
	}

	_, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.UserID, notification.Type, notification.Channel,
		notification.Priority, notification.Status, notification.Title, notification.Content,
		dataToStore, notification.TemplateID, notification.ScheduledAt, notification.SentAt,
		notification.ReadAt, notification.ExpiresAt, notification.RetryCount, notification.MaxRetries,
		notification.CreatedAt, notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetNotification 获取通知
func (r *notificationRepository) GetNotification(ctx context.Context, notificationID uuid.UUID) (*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, channel, priority, status, title, content, data, template_id,
		       scheduled_at, sent_at, read_at, expires_at, retry_count, max_retries, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	var notification domain.Notification
	err := r.db.GetContext(ctx, &notification, query, notificationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

// UpdateNotification 更新通知
func (r *notificationRepository) UpdateNotification(ctx context.Context, notification *domain.Notification) error {
	query := `
		UPDATE notifications 
		SET user_id = $2, type = $3, channel = $4, priority = $5, status = $6, title = $7, 
		    content = $8, data = $9, template_id = $10, scheduled_at = $11, sent_at = $12, 
		    read_at = $13, expires_at = $14, retry_count = $15, max_retries = $16, updated_at = $17
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.UserID, notification.Type, notification.Channel,
		notification.Priority, notification.Status, notification.Title, notification.Content,
		notification.Data, notification.TemplateID, notification.ScheduledAt, notification.SentAt,
		notification.ReadAt, notification.ExpiresAt, notification.RetryCount, notification.MaxRetries,
		notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationNotFound
	}

	return nil
}

// DeleteNotification 删除通知
func (r *notificationRepository) DeleteNotification(ctx context.Context, notificationID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationNotFound
	}

	return nil
}

// ListNotifications 获取通知列表
func (r *notificationRepository) ListNotifications(ctx context.Context, req *domain.NotificationListRequest) ([]*domain.Notification, int, error) {
	// 构建查询条件
	whereClause := "WHERE user_id = $1"
	args := []interface{}{req.UserID}
	argIndex := 2

	if req.Type != nil {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *req.Type)
		argIndex++
	}

	if req.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *req.Status)
		argIndex++
	}

	if req.Channel != nil {
		whereClause += fmt.Sprintf(" AND channel = $%d", argIndex)
		args = append(args, *req.Channel)
		argIndex++
	}

	if req.StartTime != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *req.StartTime)
		argIndex++
	}

	if req.EndTime != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *req.EndTime)
		argIndex++
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// 获取数据
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, type, channel, priority, status, title, content, data, template_id,
		       scheduled_at, sent_at, read_at, expires_at, retry_count, max_retries, created_at, updated_at
		FROM notifications %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, req.Offset)

	var notifications []*domain.Notification
	err = r.db.SelectContext(ctx, &notifications, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	return notifications, total, nil
}

// CreateUserPreferences 创建用户通知偏好
func (r *notificationRepository) CreateUserPreferences(ctx context.Context, prefs *domain.NotificationPreferences) error {
	query := `
		INSERT INTO notification_preferences (id, user_id, chat_notifications, moment_notifications, 
		                                     system_notifications, activity_notifications, email_enabled, 
		                                     push_enabled, sms_enabled, quiet_hours, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	// 将QuietHours转换为JSON
	var quietHoursJSON []byte
	var err error
	if prefs.QuietHours != nil {
		quietHoursJSON, err = json.Marshal(prefs.QuietHours)
		if err != nil {
			return fmt.Errorf("failed to marshal quiet hours: %w", err)
		}
	}

	_, err = r.db.ExecContext(ctx, query,
		prefs.ID, prefs.UserID, prefs.ChatNotifications, prefs.MomentNotifications,
		prefs.SystemNotifications, prefs.ActivityNotifications, prefs.EmailEnabled,
		prefs.PushEnabled, prefs.SMSEnabled, quietHoursJSON, prefs.CreatedAt, prefs.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user preferences: %w", err)
	}

	return nil
}

// GetUserPreferences 获取用户通知偏好
func (r *notificationRepository) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	query := `
		SELECT id, user_id, chat_notifications, moment_notifications, system_notifications, 
		       activity_notifications, email_enabled, push_enabled, sms_enabled, quiet_hours, 
		       created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	var prefs domain.NotificationPreferences
	var quietHoursJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.ID, &prefs.UserID, &prefs.ChatNotifications, &prefs.MomentNotifications,
		&prefs.SystemNotifications, &prefs.ActivityNotifications, &prefs.EmailEnabled,
		&prefs.PushEnabled, &prefs.SMSEnabled, &quietHoursJSON, &prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotificationPreferencesNotFound
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// 解析QuietHours
	if len(quietHoursJSON) > 0 {
		var quietHours domain.QuietHours
		err = json.Unmarshal(quietHoursJSON, &quietHours)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal quiet hours: %w", err)
		}
		prefs.QuietHours = &quietHours
	}

	return &prefs, nil
}

// UpdateUserPreferences 更新用户通知偏好
func (r *notificationRepository) UpdateUserPreferences(ctx context.Context, prefs *domain.NotificationPreferences) error {
	query := `
		UPDATE notification_preferences 
		SET chat_notifications = $2, moment_notifications = $3, system_notifications = $4, 
		    activity_notifications = $5, email_enabled = $6, push_enabled = $7, sms_enabled = $8, 
		    quiet_hours = $9, updated_at = $10
		WHERE user_id = $1
	`

	// 将QuietHours转换为JSON
	var quietHoursJSON []byte
	var err error
	if prefs.QuietHours != nil {
		quietHoursJSON, err = json.Marshal(prefs.QuietHours)
		if err != nil {
			return fmt.Errorf("failed to marshal quiet hours: %w", err)
		}
	}

	result, err := r.db.ExecContext(ctx, query,
		prefs.UserID, prefs.ChatNotifications, prefs.MomentNotifications,
		prefs.SystemNotifications, prefs.ActivityNotifications, prefs.EmailEnabled,
		prefs.PushEnabled, prefs.SMSEnabled, quietHoursJSON, prefs.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationPreferencesNotFound
	}

	return nil
}

// DeleteUserPreferences 删除用户通知偏好
func (r *notificationRepository) DeleteUserPreferences(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM notification_preferences WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user preferences: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationPreferencesNotFound
	}

	return nil
}

// CreateTemplate 创建通知模板
func (r *notificationRepository) CreateTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	query := `
		INSERT INTO notification_templates (id, name, type, channel, title, content, variables,
		                                   config, enabled, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	// 将Variables转换为JSON
	var variablesJSON interface{}
	if len(template.Variables) > 0 {
		variablesData, err := json.Marshal(template.Variables)
		if err != nil {
			return fmt.Errorf("failed to marshal variables: %w", err)
		}
		variablesJSON = variablesData
	} else {
		variablesJSON = nil
	}

	// 处理Config字段
	var configJSON interface{}
	if len(template.Config) > 0 {
		configJSON = template.Config
	} else {
		configJSON = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Type, template.Channel, template.Title,
		template.Content, variablesJSON, configJSON, template.Enabled,
		template.Description, template.CreatedAt, template.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification template: %w", err)
	}

	return nil
}

// GetTemplate 获取通知模板
func (r *notificationRepository) GetTemplate(ctx context.Context, templateID uuid.UUID) (*domain.NotificationTemplate, error) {
	query := `
		SELECT id, name, type, channel, title, content, variables, config, enabled,
		       description, created_at, updated_at
		FROM notification_templates
		WHERE id = $1
	`

	var template domain.NotificationTemplate
	var variablesJSON []byte
	var configJSON []byte

	err := r.db.QueryRowContext(ctx, query, templateID).Scan(
		&template.ID, &template.Name, &template.Type, &template.Channel,
		&template.Title, &template.Content, &variablesJSON, &configJSON,
		&template.Enabled, &template.Description, &template.CreatedAt, &template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotificationTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}

	// 解析Variables
	if len(variablesJSON) > 0 {
		err = json.Unmarshal(variablesJSON, &template.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}
	}

	// 解析Config
	if len(configJSON) > 0 {
		template.Config = json.RawMessage(configJSON)
	}

	return &template, nil
}

// UpdateTemplate 更新通知模板
func (r *notificationRepository) UpdateTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	query := `
		UPDATE notification_templates
		SET name = $2, type = $3, channel = $4, title = $5, content = $6, variables = $7,
		    config = $8, enabled = $9, description = $10, updated_at = $11
		WHERE id = $1
	`

	// 将Variables转换为JSON
	var variablesJSON interface{}
	if len(template.Variables) > 0 {
		variablesData, err := json.Marshal(template.Variables)
		if err != nil {
			return fmt.Errorf("failed to marshal variables: %w", err)
		}
		variablesJSON = variablesData
	} else {
		variablesJSON = nil
	}

	// 处理Config字段
	var configJSON interface{}
	if len(template.Config) > 0 {
		configJSON = template.Config
	} else {
		configJSON = nil
	}

	result, err := r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Type, template.Channel, template.Title,
		template.Content, variablesJSON, configJSON, template.Enabled,
		template.Description, template.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationTemplateNotFound
	}

	return nil
}

// DeleteTemplate 删除通知模板
func (r *notificationRepository) DeleteTemplate(ctx context.Context, templateID uuid.UUID) error {
	query := `DELETE FROM notification_templates WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete notification template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationTemplateNotFound
	}

	return nil
}

// ListTemplates 列出通知模板
func (r *notificationRepository) ListTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error) {
	query := `
		SELECT id, name, type, channel, title, content, variables, config, enabled,
		       description, created_at, updated_at
		FROM notification_templates
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification templates: %w", err)
	}
	defer rows.Close()

	var templates []*domain.NotificationTemplate
	for rows.Next() {
		var template domain.NotificationTemplate
		var variablesJSON []byte

		var configJSON []byte
		err := rows.Scan(
			&template.ID, &template.Name, &template.Type, &template.Channel,
			&template.Title, &template.Content, &variablesJSON, &configJSON,
			&template.Enabled, &template.Description, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification template: %w", err)
		}

		// 解析Variables
		if len(variablesJSON) > 0 {
			err = json.Unmarshal(variablesJSON, &template.Variables)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
			}
		}

		// 解析Config
		if len(configJSON) > 0 {
			template.Config = json.RawMessage(configJSON)
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate notification templates: %w", err)
	}

	return templates, nil
}

// CreateQueueItem 创建队列项
func (r *notificationRepository) CreateQueueItem(ctx context.Context, queueItem *domain.NotificationQueue) error {
	query := `
		INSERT INTO notification_queue (id, notification_id, priority, scheduled_at, processed_at,
		                               status, retry_count, max_retries, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		queueItem.ID, queueItem.NotificationID, queueItem.Priority, queueItem.ScheduledAt,
		queueItem.ProcessedAt, queueItem.Status, queueItem.RetryCount, queueItem.MaxRetries,
		queueItem.ErrorMessage, queueItem.CreatedAt, queueItem.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create queue item: %w", err)
	}

	return nil
}

// GetQueueItem 获取队列项
func (r *notificationRepository) GetQueueItem(ctx context.Context, queueItemID uuid.UUID) (*domain.NotificationQueue, error) {
	query := `
		SELECT id, notification_id, priority, scheduled_at, processed_at, status,
		       retry_count, max_retries, error_message, created_at, updated_at
		FROM notification_queue
		WHERE id = $1
	`

	var queueItem domain.NotificationQueue
	err := r.db.GetContext(ctx, &queueItem, query, queueItemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotificationQueueNotFound
		}
		return nil, fmt.Errorf("failed to get queue item: %w", err)
	}

	return &queueItem, nil
}

// UpdateQueueItem 更新队列项
func (r *notificationRepository) UpdateQueueItem(ctx context.Context, queueItem *domain.NotificationQueue) error {
	query := `
		UPDATE notification_queue
		SET notification_id = $2, priority = $3, scheduled_at = $4, processed_at = $5,
		    status = $6, retry_count = $7, max_retries = $8, error_message = $9, updated_at = $10
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		queueItem.ID, queueItem.NotificationID, queueItem.Priority, queueItem.ScheduledAt,
		queueItem.ProcessedAt, queueItem.Status, queueItem.RetryCount, queueItem.MaxRetries,
		queueItem.ErrorMessage, queueItem.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update queue item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationQueueNotFound
	}

	return nil
}

// DeleteQueueItem 删除队列项
func (r *notificationRepository) DeleteQueueItem(ctx context.Context, queueItemID uuid.UUID) error {
	query := `DELETE FROM notification_queue WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, queueItemID)
	if err != nil {
		return fmt.Errorf("failed to delete queue item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotificationQueueNotFound
	}

	return nil
}

// GetPendingQueueItems 获取待处理的队列项
func (r *notificationRepository) GetPendingQueueItems(ctx context.Context, limit int) ([]*domain.NotificationQueue, error) {
	query := `
		SELECT id, notification_id, priority, scheduled_at, processed_at, status,
		       retry_count, max_retries, error_message, created_at, updated_at
		FROM notification_queue
		WHERE status IN ('queued', 'failed') AND scheduled_at <= NOW()
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT $1
	`

	var queueItems []*domain.NotificationQueue
	err := r.db.SelectContext(ctx, &queueItems, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending queue items: %w", err)
	}

	return queueItems, nil
}

// GetQueueStats 获取队列统计
func (r *notificationRepository) GetQueueStats(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM notification_queue
		GROUP BY status
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan queue stats: %w", err)
		}
		stats[status] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate queue stats: %w", err)
	}

	return stats, nil
}

// GetStats 获取通知统计
func (r *notificationRepository) GetStats(ctx context.Context, userID *uuid.UUID, startTime, endTime time.Time) (*domain.NotificationStats, error) {
	whereClause := "WHERE created_at >= $1 AND created_at <= $2"
	args := []interface{}{startTime, endTime}

	if userID != nil {
		whereClause += " AND user_id = $3"
		args = append(args, *userID)
	}

	// 基础统计
	basicQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_sent,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as total_delivered,
			COUNT(CASE WHEN status = 'read' THEN 1 END) as total_read,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as total_failed
		FROM notifications %s
	`, whereClause)

	var stats domain.NotificationStats
	err := r.db.QueryRowContext(ctx, basicQuery, args...).Scan(
		&stats.TotalSent, &stats.TotalDelivered, &stats.TotalRead, &stats.TotalFailed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	// 计算比率
	if stats.TotalSent > 0 {
		stats.DeliveryRate = float64(stats.TotalDelivered) / float64(stats.TotalSent) * 100
		stats.ReadRate = float64(stats.TotalRead) / float64(stats.TotalSent) * 100
	}

	// 按类型统计
	typeQuery := fmt.Sprintf(`
		SELECT type, COUNT(*) as count
		FROM notifications %s
		GROUP BY type
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, typeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}
	defer rows.Close()

	stats.ByType = make(map[domain.NotificationType]int64)
	for rows.Next() {
		var notType domain.NotificationType
		var count int64
		err := rows.Scan(&notType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan type stats: %w", err)
		}
		stats.ByType[notType] = count
	}

	// 按渠道统计
	channelQuery := fmt.Sprintf(`
		SELECT channel, COUNT(*) as count
		FROM notifications %s
		GROUP BY channel
	`, whereClause)

	rows, err = r.db.QueryContext(ctx, channelQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel stats: %w", err)
	}
	defer rows.Close()

	stats.ByChannel = make(map[domain.NotificationChannel]int64)
	for rows.Next() {
		var channel domain.NotificationChannel
		var count int64
		err := rows.Scan(&channel, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel stats: %w", err)
		}
		stats.ByChannel[channel] = count
	}

	// 按状态统计
	statusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*) as count
		FROM notifications %s
		GROUP BY status
	`, whereClause)

	rows, err = r.db.QueryContext(ctx, statusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}
	defer rows.Close()

	stats.ByStatus = make(map[domain.NotificationStatus]int64)
	for rows.Next() {
		var status domain.NotificationStatus
		var count int64
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status stats: %w", err)
		}
		stats.ByStatus[status] = count
	}

	return &stats, nil
}

// GetDeliveryReport 获取投递报告
func (r *notificationRepository) GetDeliveryReport(ctx context.Context, notificationID uuid.UUID) (*domain.NotificationEvent, error) {
	query := `
		SELECT id, type, user_id, notification_id, data, timestamp
		FROM notification_events
		WHERE notification_id = $1 AND type = 'notification.delivered'
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var event domain.NotificationEvent
	err := r.db.GetContext(ctx, &event, query, notificationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("delivery report not found")
		}
		return nil, fmt.Errorf("failed to get delivery report: %w", err)
	}

	return &event, nil
}

// CreateEvent 创建通知事件
func (r *notificationRepository) CreateEvent(ctx context.Context, event *domain.NotificationEvent) error {
	query := `
		INSERT INTO notification_events (id, type, user_id, notification_id, data, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.Type, event.UserID, event.NotificationID, event.Data, event.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification event: %w", err)
	}

	return nil
}

// BatchCreateNotifications 批量创建通知
func (r *notificationRepository) BatchCreateNotifications(ctx context.Context, notifications []*domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := `
		INSERT INTO notifications (id, user_id, type, channel, priority, status, title, content,
		                          data, template_id, scheduled_at, sent_at, read_at, expires_at,
		                          retry_count, max_retries, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, notification := range notifications {
		_, err = tx.ExecContext(ctx, query,
			notification.ID, notification.UserID, notification.Type, notification.Channel,
			notification.Priority, notification.Status, notification.Title, notification.Content,
			notification.Data, notification.TemplateID, notification.ScheduledAt, notification.SentAt,
			notification.ReadAt, notification.ExpiresAt, notification.RetryCount, notification.MaxRetries,
			notification.CreatedAt, notification.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create notification %s: %w", notification.ID, err)
		}
	}

	return tx.Commit()
}

// BatchUpdateNotifications 批量更新通知
func (r *notificationRepository) BatchUpdateNotifications(ctx context.Context, notifications []*domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := `
		UPDATE notifications
		SET user_id = $2, type = $3, channel = $4, priority = $5, status = $6, title = $7,
		    content = $8, data = $9, template_id = $10, scheduled_at = $11, sent_at = $12,
		    read_at = $13, expires_at = $14, retry_count = $15, max_retries = $16, updated_at = $17
		WHERE id = $1
	`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, notification := range notifications {
		_, err = tx.ExecContext(ctx, query,
			notification.ID, notification.UserID, notification.Type, notification.Channel,
			notification.Priority, notification.Status, notification.Title, notification.Content,
			notification.Data, notification.TemplateID, notification.ScheduledAt, notification.SentAt,
			notification.ReadAt, notification.ExpiresAt, notification.RetryCount, notification.MaxRetries,
			notification.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to update notification %s: %w", notification.ID, err)
		}
	}

	return tx.Commit()
}

// BatchDeleteNotifications 批量删除通知
func (r *notificationRepository) BatchDeleteNotifications(ctx context.Context, notificationIDs []uuid.UUID) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	query := `DELETE FROM notifications WHERE id = ANY($1)`

	_, err := r.db.ExecContext(ctx, query, notificationIDs)
	if err != nil {
		return fmt.Errorf("failed to batch delete notifications: %w", err)
	}

	return nil
}

// CleanupExpiredNotifications 清理过期通知
func (r *notificationRepository) CleanupExpiredNotifications(ctx context.Context) (int, error) {
	query := `DELETE FROM notifications WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// CleanupOldNotifications 清理旧通知
func (r *notificationRepository) CleanupOldNotifications(ctx context.Context, olderThan time.Time) (int, error) {
	query := `DELETE FROM notifications WHERE created_at < $1 AND status = 'read'`

	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
