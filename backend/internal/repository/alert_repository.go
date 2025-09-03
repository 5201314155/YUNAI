package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// AlertRepository 告警仓库接口
type AlertRepository interface {
	// 告警规则管理
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	UpdateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	DeleteAlertRule(ctx context.Context, ruleID uuid.UUID) error
	GetAlertRule(ctx context.Context, ruleID uuid.UUID) (*domain.AlertRule, error)
	ListAlertRules(ctx context.Context) ([]*domain.AlertRule, error)

	// 告警记录管理
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	GetAlert(ctx context.Context, alertID uuid.UUID) (*domain.Alert, error)
	GetActiveAlert(ctx context.Context, ruleID uuid.UUID) (*domain.Alert, error)
	GetAlerts(ctx context.Context, status string, limit int) ([]*domain.Alert, error)
	GetAlertsByRule(ctx context.Context, ruleID uuid.UUID) ([]*domain.Alert, error)
}

// alertRepository 告警仓库实现
type alertRepository struct {
	db *sqlx.DB
}

// NewAlertRepository 创建告警仓库
func NewAlertRepository(db *sqlx.DB) AlertRepository {
	return &alertRepository{db: db}
}

// CreateAlertRule 创建告警规则
func (r *alertRepository) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	query := `
		INSERT INTO alert_rules (id, name, description, metric, operator, threshold, 
		                        duration, severity, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Metric, rule.Operator,
		rule.Threshold, rule.Duration, rule.Severity, rule.Enabled,
		rule.CreatedAt, rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	return nil
}

// UpdateAlertRule 更新告警规则
func (r *alertRepository) UpdateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	query := `
		UPDATE alert_rules 
		SET name = $2, description = $3, metric = $4, operator = $5, 
		    threshold = $6, duration = $7, severity = $8, enabled = $9, updated_at = $10
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Metric, rule.Operator,
		rule.Threshold, rule.Duration, rule.Severity, rule.Enabled, rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAlertRuleNotFound
	}

	return nil
}

// DeleteAlertRule 删除告警规则
func (r *alertRepository) DeleteAlertRule(ctx context.Context, ruleID uuid.UUID) error {
	query := `DELETE FROM alert_rules WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAlertRuleNotFound
	}

	return nil
}

// GetAlertRule 获取告警规则
func (r *alertRepository) GetAlertRule(ctx context.Context, ruleID uuid.UUID) (*domain.AlertRule, error) {
	query := `
		SELECT id, name, description, metric, operator, threshold, duration, 
		       severity, enabled, created_at, updated_at
		FROM alert_rules 
		WHERE id = $1
	`

	rule := &domain.AlertRule{}
	err := r.db.GetContext(ctx, rule, query, ruleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAlertRuleNotFound
		}
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	return rule, nil
}

// ListAlertRules 列出所有告警规则
func (r *alertRepository) ListAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	query := `
		SELECT id, name, description, metric, operator, threshold, duration, 
		       severity, enabled, created_at, updated_at
		FROM alert_rules 
		ORDER BY created_at DESC
	`

	var rules []*domain.AlertRule
	err := r.db.SelectContext(ctx, &rules, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert rules: %w", err)
	}

	return rules, nil
}

// CreateAlert 创建告警记录
func (r *alertRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		INSERT INTO alerts (id, rule_id, rule_name, metric, value, threshold, 
		                   severity, status, message, fired_at, resolved_at, 
		                   created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.RuleID, alert.RuleName, alert.Metric, alert.Value,
		alert.Threshold, alert.Severity, alert.Status, alert.Message,
		alert.FiredAt, alert.ResolvedAt, alert.CreatedAt, alert.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// UpdateAlert 更新告警记录
func (r *alertRepository) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		UPDATE alerts 
		SET status = $2, resolved_at = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.Status, alert.ResolvedAt, alert.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAlertNotFound
	}

	return nil
}

// GetAlert 获取告警记录
func (r *alertRepository) GetAlert(ctx context.Context, alertID uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, rule_id, rule_name, metric, value, threshold, severity, 
		       status, message, fired_at, resolved_at, created_at, updated_at
		FROM alerts 
		WHERE id = $1
	`

	alert := &domain.Alert{}
	err := r.db.GetContext(ctx, alert, query, alertID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAlertNotFound
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return alert, nil
}

// GetActiveAlert 获取活跃的告警记录
func (r *alertRepository) GetActiveAlert(ctx context.Context, ruleID uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, rule_id, rule_name, metric, value, threshold, severity, 
		       status, message, fired_at, resolved_at, created_at, updated_at
		FROM alerts 
		WHERE rule_id = $1 AND status = 'firing'
		ORDER BY fired_at DESC
		LIMIT 1
	`

	alert := &domain.Alert{}
	err := r.db.GetContext(ctx, alert, query, ruleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAlertNotFound
		}
		return nil, fmt.Errorf("failed to get active alert: %w", err)
	}

	return alert, nil
}

// GetAlerts 获取告警列表
func (r *alertRepository) GetAlerts(ctx context.Context, status string, limit int) ([]*domain.Alert, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, rule_id, rule_name, metric, value, threshold, severity, 
			       status, message, fired_at, resolved_at, created_at, updated_at
			FROM alerts 
			WHERE status = $1
			ORDER BY fired_at DESC
			LIMIT $2
		`
		args = []interface{}{status, limit}
	} else {
		query = `
			SELECT id, rule_id, rule_name, metric, value, threshold, severity, 
			       status, message, fired_at, resolved_at, created_at, updated_at
			FROM alerts 
			ORDER BY fired_at DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	var alerts []*domain.Alert
	err := r.db.SelectContext(ctx, &alerts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, nil
}

// GetAlertsByRule 获取指定规则的告警历史
func (r *alertRepository) GetAlertsByRule(ctx context.Context, ruleID uuid.UUID) ([]*domain.Alert, error) {
	query := `
		SELECT id, rule_id, rule_name, metric, value, threshold, severity, 
		       status, message, fired_at, resolved_at, created_at, updated_at
		FROM alerts 
		WHERE rule_id = $1
		ORDER BY fired_at DESC
	`

	var alerts []*domain.Alert
	err := r.db.SelectContext(ctx, &alerts, query, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts by rule: %w", err)
	}

	return alerts, nil
}
