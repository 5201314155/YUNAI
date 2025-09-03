package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
	"yunai/pkg/email"
)

// AlertService 告警服务接口
type AlertService interface {
	// 告警规则管理
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	UpdateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	DeleteAlertRule(ctx context.Context, ruleID uuid.UUID) error
	GetAlertRule(ctx context.Context, ruleID uuid.UUID) (*domain.AlertRule, error)
	ListAlertRules(ctx context.Context) ([]*domain.AlertRule, error)

	// 告警检查和触发
	CheckAlerts(ctx context.Context) error
	TriggerAlert(ctx context.Context, ruleID uuid.UUID, value float64) error
	ResolveAlert(ctx context.Context, alertID uuid.UUID) error

	// 告警历史
	GetAlerts(ctx context.Context, status string, limit int) ([]*domain.Alert, error)
	GetAlertHistory(ctx context.Context, ruleID uuid.UUID) ([]*domain.Alert, error)

	// 告警通知
	SendAlertNotification(ctx context.Context, alert *domain.Alert) error
}

// alertService 告警服务实现
type alertService struct {
	alertRepo        repository.AlertRepository
	analyticsService AnalyticsService
	emailService     email.EmailService
	logger           *logrus.Logger
}

// NewAlertService 创建告警服务
func NewAlertService(
	alertRepo repository.AlertRepository,
	analyticsService AnalyticsService,
	emailService email.EmailService,
	logger *logrus.Logger,
) AlertService {
	return &alertService{
		alertRepo:        alertRepo,
		analyticsService: analyticsService,
		emailService:     emailService,
		logger:           logger,
	}
}

// CreateAlertRule 创建告警规则
func (s *alertService) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	if err := s.alertRepo.CreateAlertRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"metric":    rule.Metric,
		"threshold": rule.Threshold,
	}).Info("Alert rule created")

	return nil
}

// UpdateAlertRule 更新告警规则
func (s *alertService) UpdateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	rule.UpdatedAt = time.Now()

	if err := s.alertRepo.UpdateAlertRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
	}).Info("Alert rule updated")

	return nil
}

// DeleteAlertRule 删除告警规则
func (s *alertService) DeleteAlertRule(ctx context.Context, ruleID uuid.UUID) error {
	if err := s.alertRepo.DeleteAlertRule(ctx, ruleID); err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	s.logger.WithField("rule_id", ruleID).Info("Alert rule deleted")
	return nil
}

// GetAlertRule 获取告警规则
func (s *alertService) GetAlertRule(ctx context.Context, ruleID uuid.UUID) (*domain.AlertRule, error) {
	rule, err := s.alertRepo.GetAlertRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}
	return rule, nil
}

// ListAlertRules 列出所有告警规则
func (s *alertService) ListAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	rules, err := s.alertRepo.ListAlertRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert rules: %w", err)
	}
	return rules, nil
}

// CheckAlerts 检查所有告警规则
func (s *alertService) CheckAlerts(ctx context.Context) error {
	rules, err := s.ListAlertRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if err := s.checkSingleAlert(ctx, rule); err != nil {
			s.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to check alert rule")
		}
	}

	return nil
}

// checkSingleAlert 检查单个告警规则
func (s *alertService) checkSingleAlert(ctx context.Context, rule *domain.AlertRule) error {
	// 获取指标值
	value, err := s.getMetricValue(ctx, rule.Metric)
	if err != nil {
		return fmt.Errorf("failed to get metric value for %s: %w", rule.Metric, err)
	}

	// 检查是否触发告警
	triggered := s.evaluateAlertCondition(rule, value)

	if triggered {
		// 检查是否已经有活跃的告警
		activeAlert, err := s.alertRepo.GetActiveAlert(ctx, rule.ID)
		if err != nil && err != domain.ErrAlertNotFound {
			return fmt.Errorf("failed to check active alert: %w", err)
		}

		if activeAlert == nil {
			// 触发新告警
			return s.TriggerAlert(ctx, rule.ID, value)
		}
	} else {
		// 检查是否需要解决告警
		activeAlert, err := s.alertRepo.GetActiveAlert(ctx, rule.ID)
		if err != nil && err != domain.ErrAlertNotFound {
			return fmt.Errorf("failed to check active alert: %w", err)
		}

		if activeAlert != nil {
			// 解决告警
			return s.ResolveAlert(ctx, activeAlert.ID)
		}
	}

	return nil
}

// getMetricValue 获取指标值
func (s *alertService) getMetricValue(ctx context.Context, metric string) (float64, error) {
	switch metric {
	case "api_response_time":
		health, err := s.analyticsService.GetSystemHealthMetrics(ctx)
		if err != nil {
			return 0, err
		}
		// 计算平均响应时间
		var totalTime time.Duration
		var count int64
		for _, apiMetric := range health.APIMetrics {
			totalTime += apiMetric.AvgTime
			count++
		}
		if count > 0 {
			return float64(totalTime/time.Duration(count)) / float64(time.Millisecond), nil
		}
		return 0, nil

	case "api_error_rate":
		health, err := s.analyticsService.GetSystemHealthMetrics(ctx)
		if err != nil {
			return 0, err
		}
		// 计算平均错误率
		var totalErrorRate float64
		var count int64
		for _, apiMetric := range health.APIMetrics {
			totalErrorRate += (100 - apiMetric.SuccessRate)
			count++
		}
		if count > 0 {
			return totalErrorRate / float64(count), nil
		}
		return 0, nil

	case "daily_active_users":
		dau, err := s.analyticsService.GetDAU(ctx, time.Now())
		if err != nil {
			return 0, err
		}
		return float64(dau), nil

	case "system_health":
		health, err := s.analyticsService.GetSystemHealthMetrics(ctx)
		if err != nil {
			return 0, err
		}
		// 将健康状态转换为数值
		switch health.Status {
		case "healthy":
			return 1, nil
		case "degraded":
			return 0.5, nil
		case "unhealthy":
			return 0, nil
		default:
			return 0, nil
		}

	default:
		return 0, fmt.Errorf("unknown metric: %s", metric)
	}
}

// evaluateAlertCondition 评估告警条件
func (s *alertService) evaluateAlertCondition(rule *domain.AlertRule, value float64) bool {
	switch rule.Operator {
	case ">":
		return value > rule.Threshold
	case ">=":
		return value >= rule.Threshold
	case "<":
		return value < rule.Threshold
	case "<=":
		return value <= rule.Threshold
	case "=", "==":
		return value == rule.Threshold
	case "!=":
		return value != rule.Threshold
	default:
		s.logger.WithField("operator", rule.Operator).Warn("Unknown alert operator")
		return false
	}
}

// TriggerAlert 触发告警
func (s *alertService) TriggerAlert(ctx context.Context, ruleID uuid.UUID, value float64) error {
	rule, err := s.GetAlertRule(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("failed to get alert rule: %w", err)
	}

	alert := &domain.Alert{
		ID:        uuid.New(),
		RuleID:    ruleID,
		RuleName:  rule.Name,
		Metric:    rule.Metric,
		Value:     value,
		Threshold: rule.Threshold,
		Severity:  rule.Severity,
		Status:    "firing",
		Message:   s.buildAlertMessage(rule, value),
		FiredAt:   time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// 发送告警通知
	if err := s.SendAlertNotification(ctx, alert); err != nil {
		s.logger.WithError(err).Error("Failed to send alert notification")
	}

	s.logger.WithFields(logrus.Fields{
		"alert_id":  alert.ID,
		"rule_name": rule.Name,
		"metric":    rule.Metric,
		"value":     value,
		"threshold": rule.Threshold,
		"severity":  rule.Severity,
	}).Warn("Alert triggered")

	return nil
}

// ResolveAlert 解决告警
func (s *alertService) ResolveAlert(ctx context.Context, alertID uuid.UUID) error {
	alert, err := s.alertRepo.GetAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now
	alert.UpdatedAt = now

	if err := s.alertRepo.UpdateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"alert_id":  alert.ID,
		"rule_name": alert.RuleName,
		"duration":  now.Sub(alert.FiredAt),
	}).Info("Alert resolved")

	return nil
}

// GetAlerts 获取告警列表
func (s *alertService) GetAlerts(ctx context.Context, status string, limit int) ([]*domain.Alert, error) {
	alerts, err := s.alertRepo.GetAlerts(ctx, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	return alerts, nil
}

// GetAlertHistory 获取告警历史
func (s *alertService) GetAlertHistory(ctx context.Context, ruleID uuid.UUID) ([]*domain.Alert, error) {
	alerts, err := s.alertRepo.GetAlertsByRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert history: %w", err)
	}
	return alerts, nil
}

// SendAlertNotification 发送告警通知
func (s *alertService) SendAlertNotification(ctx context.Context, alert *domain.Alert) error {
	// 构建邮件主题和内容
	subject := fmt.Sprintf("🚨 YUNAI Alert: %s", alert.RuleName)
	content := fmt.Sprintf(`
告警详情：
- 规则名称: %s
- 监控指标: %s
- 当前值: %.2f
- 阈值: %.2f
- 严重程度: %s
- 触发时间: %s
- 告警消息: %s

请及时处理此告警。

YUNAI 监控系统
`, alert.RuleName, alert.Metric, alert.Value, alert.Threshold, 
   alert.Severity, alert.FiredAt.Format("2006-01-02 15:04:05"), alert.Message)

	// 发送邮件通知（这里应该配置管理员邮箱列表）
	adminEmails := []string{"admin@yunai.com"} // 应该从配置中读取
	
	for _, email := range adminEmails {
		if err := s.emailService.SendNotificationEmail(email, subject, content); err != nil {
			s.logger.WithError(err).WithField("email", email).Error("Failed to send alert email")
		}
	}

	return nil
}

// buildAlertMessage 构建告警消息
func (s *alertService) buildAlertMessage(rule *domain.AlertRule, value float64) string {
	return fmt.Sprintf("指标 %s 的值 %.2f %s %.2f，触发了告警规则 %s",
		rule.Metric, value, rule.Operator, rule.Threshold, rule.Name)
}
