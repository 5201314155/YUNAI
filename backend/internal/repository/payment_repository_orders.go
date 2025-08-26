package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"yunai/internal/domain"
)

// CreateRechargeOrder 创建充值订单
func (r *paymentRepository) CreateRechargeOrder(ctx context.Context, order *domain.RechargeOrder) error {
	query := `
		INSERT INTO recharge_orders (
			id, user_id, order_no, amount, coins_amount, exchange_rate,
			payment_method, payment_card_id, original_amount, discount_amount,
			bonus_coins, card_code, status, payment_status, expired_at
		) VALUES (
			:id, :user_id, :order_no, :amount, :coins_amount, :exchange_rate,
			:payment_method, :payment_card_id, :original_amount, :discount_amount,
			:bonus_coins, :card_code, :status, :payment_status, :expired_at
		)`
	
	_, err := r.db.NamedExecContext(ctx, query, order)
	if err != nil {
		return fmt.Errorf("failed to create recharge order: %w", err)
	}
	
	return nil
}

// GetRechargeOrderByID 根据ID获取充值订单
func (r *paymentRepository) GetRechargeOrderByID(ctx context.Context, id uuid.UUID) (*domain.RechargeOrder, error) {
	var order domain.RechargeOrder
	query := `
		SELECT id, user_id, order_no, amount, coins_amount, exchange_rate,
			   payment_method, payment_card_id, original_amount, discount_amount,
			   bonus_coins, card_code, status, payment_status, third_party_order_no,
			   payment_url, failure_reason, created_at, updated_at, completed_at, expired_at
		FROM recharge_orders WHERE id = $1`
	
	err := r.db.GetContext(ctx, &order, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get recharge order: %w", err)
	}
	
	return &order, nil
}

// GetRechargeOrderByOrderNo 根据订单号获取充值订单
func (r *paymentRepository) GetRechargeOrderByOrderNo(ctx context.Context, orderNo string) (*domain.RechargeOrder, error) {
	var order domain.RechargeOrder
	query := `
		SELECT id, user_id, order_no, amount, coins_amount, exchange_rate,
			   payment_method, payment_card_id, original_amount, discount_amount,
			   bonus_coins, card_code, status, payment_status, third_party_order_no,
			   payment_url, failure_reason, created_at, updated_at, completed_at, expired_at
		FROM recharge_orders WHERE order_no = $1`
	
	err := r.db.GetContext(ctx, &order, query, orderNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get recharge order: %w", err)
	}
	
	return &order, nil
}

// UpdateRechargeOrder 更新充值订单
func (r *paymentRepository) UpdateRechargeOrder(ctx context.Context, order *domain.RechargeOrder) error {
	order.UpdatedAt = time.Now()
	
	query := `
		UPDATE recharge_orders SET
			status = :status, payment_status = :payment_status,
			third_party_order_no = :third_party_order_no, payment_url = :payment_url,
			failure_reason = :failure_reason, updated_at = :updated_at,
			completed_at = :completed_at
		WHERE id = :id`
	
	result, err := r.db.NamedExecContext(ctx, query, order)
	if err != nil {
		return fmt.Errorf("failed to update recharge order: %w", err)
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

// GetUserRechargeOrders 获取用户充值订单列表
func (r *paymentRepository) GetUserRechargeOrders(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.RechargeOrder, error) {
	var orders []*domain.RechargeOrder
	query := `
		SELECT id, user_id, order_no, amount, coins_amount, exchange_rate,
			   payment_method, payment_card_id, original_amount, discount_amount,
			   bonus_coins, card_code, status, payment_status, third_party_order_no,
			   payment_url, failure_reason, created_at, updated_at, completed_at, expired_at
		FROM recharge_orders 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	err := r.db.SelectContext(ctx, &orders, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user recharge orders: %w", err)
	}
	
	return orders, nil
}

// GetExpiredOrders 获取过期订单
func (r *paymentRepository) GetExpiredOrders(ctx context.Context) ([]*domain.RechargeOrder, error) {
	var orders []*domain.RechargeOrder
	query := `
		SELECT id, user_id, order_no, amount, coins_amount, exchange_rate,
			   payment_method, payment_card_id, original_amount, discount_amount,
			   bonus_coins, card_code, status, payment_status, third_party_order_no,
			   payment_url, failure_reason, created_at, updated_at, completed_at, expired_at
		FROM recharge_orders 
		WHERE expired_at < NOW() AND status IN ('pending', 'processing')
		ORDER BY expired_at ASC`
	
	err := r.db.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired orders: %w", err)
	}
	
	return orders, nil
}

// GetActiveRechargePackages 获取活跃的充值套餐
func (r *paymentRepository) GetActiveRechargePackages(ctx context.Context) ([]*domain.RechargePackage, error) {
	var packages []*domain.RechargePackage
	query := `
		SELECT id, name, amount, coins_amount, bonus_coins, discount_rate,
			   is_popular, is_active, sort_order, created_at, updated_at
		FROM recharge_packages 
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, created_at ASC`
	
	err := r.db.SelectContext(ctx, &packages, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get recharge packages: %w", err)
	}
	
	return packages, nil
}

// GetRechargePackageByID 根据ID获取充值套餐
func (r *paymentRepository) GetRechargePackageByID(ctx context.Context, id uuid.UUID) (*domain.RechargePackage, error) {
	var pkg domain.RechargePackage
	query := `
		SELECT id, name, amount, coins_amount, bonus_coins, discount_rate,
			   is_popular, is_active, sort_order, created_at, updated_at
		FROM recharge_packages WHERE id = $1`
	
	err := r.db.GetContext(ctx, &pkg, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get recharge package: %w", err)
	}
	
	return &pkg, nil
}

// CreateAdminRechargeRecord 创建管理员代充记录
func (r *paymentRepository) CreateAdminRechargeRecord(ctx context.Context, record *domain.AdminRechargeRecord) error {
	query := `
		INSERT INTO admin_recharge_records (
			id, target_user_id, admin_user_id, coins_amount, reason,
			admin_payment_verified, admin_payment_verified_at
		) VALUES (
			:id, :target_user_id, :admin_user_id, :coins_amount, :reason,
			:admin_payment_verified, :admin_payment_verified_at
		)`
	
	_, err := r.db.NamedExecContext(ctx, query, record)
	if err != nil {
		return fmt.Errorf("failed to create admin recharge record: %w", err)
	}
	
	return nil
}

// GetAdminRechargeRecords 获取管理员代充记录列表
func (r *paymentRepository) GetAdminRechargeRecords(ctx context.Context, offset, limit int) ([]*domain.AdminRechargeRecord, error) {
	var records []*domain.AdminRechargeRecord
	query := `
		SELECT id, target_user_id, admin_user_id, coins_amount, reason,
			   admin_payment_verified, admin_payment_verified_at, created_at
		FROM admin_recharge_records
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	
	err := r.db.SelectContext(ctx, &records, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin recharge records: %w", err)
	}
	
	return records, nil
}
