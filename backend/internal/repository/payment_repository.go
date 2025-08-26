package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"yunai/internal/domain"
)

// PaymentRepository 支付仓库接口
type PaymentRepository interface {
	// 支付卡片管理
	CreatePaymentCard(ctx context.Context, card *domain.PaymentCard) error
	GetPaymentCardsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.PaymentCard, error)
	GetPaymentCardByID(ctx context.Context, id uuid.UUID) (*domain.PaymentCard, error)
	UpdatePaymentCard(ctx context.Context, card *domain.PaymentCard) error
	DeletePaymentCard(ctx context.Context, id uuid.UUID) error
	SetDefaultPaymentCard(ctx context.Context, userID, cardID uuid.UUID) error

	// 支付密码管理
	CreatePaymentPassword(ctx context.Context, password *domain.PaymentPassword) error
	GetPaymentPasswordByUserID(ctx context.Context, userID uuid.UUID) (*domain.PaymentPassword, error)
	UpdatePaymentPassword(ctx context.Context, password *domain.PaymentPassword) error
	IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error
	ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error
	LockPaymentPassword(ctx context.Context, userID uuid.UUID, lockDuration time.Duration) error

	// 充值订单管理
	CreateRechargeOrder(ctx context.Context, order *domain.RechargeOrder) error
	GetRechargeOrderByID(ctx context.Context, id uuid.UUID) (*domain.RechargeOrder, error)
	GetRechargeOrderByOrderNo(ctx context.Context, orderNo string) (*domain.RechargeOrder, error)
	UpdateRechargeOrder(ctx context.Context, order *domain.RechargeOrder) error
	GetUserRechargeOrders(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.RechargeOrder, error)
	GetExpiredOrders(ctx context.Context) ([]*domain.RechargeOrder, error)

	// 充值套餐管理
	GetActiveRechargePackages(ctx context.Context) ([]*domain.RechargePackage, error)
	GetRechargePackageByID(ctx context.Context, id uuid.UUID) (*domain.RechargePackage, error)

	// 管理员代充记录
	CreateAdminRechargeRecord(ctx context.Context, record *domain.AdminRechargeRecord) error
	GetAdminRechargeRecords(ctx context.Context, offset, limit int) ([]*domain.AdminRechargeRecord, error)
}

// paymentRepository 支付仓库实现
type paymentRepository struct {
	db *sqlx.DB
}

// NewPaymentRepository 创建支付仓库
func NewPaymentRepository(db *sqlx.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// CreatePaymentCard 创建支付卡片
func (r *paymentRepository) CreatePaymentCard(ctx context.Context, card *domain.PaymentCard) error {
	query := `
		INSERT INTO payment_cards (
			id, user_id, card_number, card_number_hash, card_type, bank_name, bank_code,
			cardholder_name, bound_email, bound_at, email_verified,
			balance, currency, is_default, is_active, is_frozen, is_verified
		) VALUES (
			:id, :user_id, :card_number, :card_number_hash, :card_type, :bank_name, :bank_code,
			:cardholder_name, :bound_email, :bound_at, :email_verified,
			:balance, :currency, :is_default, :is_active, :is_frozen, :is_verified
		)`

	_, err := r.db.NamedExecContext(ctx, query, card)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return domain.ErrConflict
			}
		}
		return fmt.Errorf("failed to create payment card: %w", err)
	}

	return nil
}

// GetPaymentCardsByUserID 获取用户的支付卡片列表
func (r *paymentRepository) GetPaymentCardsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.PaymentCard, error) {
	var cards []*domain.PaymentCard
	query := `
		SELECT id, user_id, card_number, card_type, bank_name, bank_code,
			   cardholder_name, bound_email, bound_at, email_verified,
			   balance, currency, is_default, is_active, is_frozen, is_verified, verified_at,
			   created_at, updated_at
		FROM payment_cards
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY is_default DESC, created_at DESC`

	err := r.db.SelectContext(ctx, &cards, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment cards: %w", err)
	}

	return cards, nil
}

// GetPaymentCardByID 根据ID获取支付卡片
func (r *paymentRepository) GetPaymentCardByID(ctx context.Context, id uuid.UUID) (*domain.PaymentCard, error) {
	var card domain.PaymentCard
	query := `
		SELECT id, user_id, card_number, card_number_hash, card_type, bank_name, bank_code,
			   cardholder_name, bound_email, bound_at, email_verified,
			   balance, currency, is_default, is_active, is_frozen, is_verified, verified_at,
			   created_at, updated_at
		FROM payment_cards WHERE id = $1`

	err := r.db.GetContext(ctx, &card, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	return &card, nil
}

// UpdatePaymentCard 更新支付卡片
func (r *paymentRepository) UpdatePaymentCard(ctx context.Context, card *domain.PaymentCard) error {
	card.UpdatedAt = time.Now()

	query := `
		UPDATE payment_cards SET
			card_number = :card_number, card_type = :card_type, bank_name = :bank_name,
			bank_code = :bank_code, cardholder_name = :cardholder_name,
			bound_email = :bound_email, bound_at = :bound_at, email_verified = :email_verified,
			balance = :balance, currency = :currency,
			is_default = :is_default, is_active = :is_active, is_frozen = :is_frozen, is_verified = :is_verified,
			verified_at = :verified_at, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, card)
	if err != nil {
		return fmt.Errorf("failed to update payment card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// DeletePaymentCard 删除支付卡片
func (r *paymentRepository) DeletePaymentCard(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE payment_cards SET is_active = FALSE, updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// SetDefaultPaymentCard 设置默认支付卡片
func (r *paymentRepository) SetDefaultPaymentCard(ctx context.Context, userID, cardID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 取消用户所有卡片的默认状态
	_, err = tx.ExecContext(ctx,
		`UPDATE payment_cards SET is_default = FALSE, updated_at = NOW() WHERE user_id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("failed to unset default cards: %w", err)
	}

	// 设置指定卡片为默认
	result, err := tx.ExecContext(ctx,
		`UPDATE payment_cards SET is_default = TRUE, updated_at = NOW() WHERE id = $1 AND user_id = $2`,
		cardID, userID)
	if err != nil {
		return fmt.Errorf("failed to set default card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit()
}

// CreatePaymentPassword 创建支付密码
func (r *paymentRepository) CreatePaymentPassword(ctx context.Context, password *domain.PaymentPassword) error {
	query := `
		INSERT INTO payment_passwords (
			id, user_id, password_hash, salt, failed_attempts, is_active
		) VALUES (
			:id, :user_id, :password_hash, :salt, :failed_attempts, :is_active
		)`

	_, err := r.db.NamedExecContext(ctx, query, password)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return domain.ErrConflict
			}
		}
		return fmt.Errorf("failed to create payment password: %w", err)
	}

	return nil
}

// GetPaymentPasswordByUserID 根据用户ID获取支付密码
func (r *paymentRepository) GetPaymentPasswordByUserID(ctx context.Context, userID uuid.UUID) (*domain.PaymentPassword, error) {
	var password domain.PaymentPassword
	query := `
		SELECT id, user_id, password_hash, salt, failed_attempts, locked_until,
			   last_failed_at, is_active, created_at, updated_at
		FROM payment_passwords WHERE user_id = $1`

	err := r.db.GetContext(ctx, &password, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get payment password: %w", err)
	}

	return &password, nil
}

// UpdatePaymentPassword 更新支付密码
func (r *paymentRepository) UpdatePaymentPassword(ctx context.Context, password *domain.PaymentPassword) error {
	password.UpdatedAt = time.Now()

	query := `
		UPDATE payment_passwords SET
			password_hash = :password_hash, salt = :salt, failed_attempts = :failed_attempts,
			locked_until = :locked_until, last_failed_at = :last_failed_at,
			is_active = :is_active, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, password)
	if err != nil {
		return fmt.Errorf("failed to update payment password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// IncrementFailedAttempts 增加失败尝试次数
func (r *paymentRepository) IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE payment_passwords
		SET failed_attempts = failed_attempts + 1,
			last_failed_at = NOW(),
			updated_at = NOW()
		WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to increment failed attempts: %w", err)
	}

	return nil
}

// ResetFailedAttempts 重置失败尝试次数
func (r *paymentRepository) ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE payment_passwords
		SET failed_attempts = 0,
			locked_until = NULL,
			last_failed_at = NULL,
			updated_at = NOW()
		WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}

	return nil
}

// LockPaymentPassword 锁定支付密码
func (r *paymentRepository) LockPaymentPassword(ctx context.Context, userID uuid.UUID, lockDuration time.Duration) error {
	lockUntil := time.Now().Add(lockDuration)

	query := `
		UPDATE payment_passwords
		SET locked_until = $2,
			updated_at = NOW()
		WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID, lockUntil)
	if err != nil {
		return fmt.Errorf("failed to lock payment password: %w", err)
	}

	return nil
}
