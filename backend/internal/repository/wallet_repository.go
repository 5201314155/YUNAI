package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// WalletRepository 钱包仓库接口
type WalletRepository interface {
	// 钱包基本操作
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	UpdateWalletBalance(ctx context.Context, userID uuid.UUID, amount float64) error
	AddCoins(ctx context.Context, userID uuid.UUID, coins int64, reason string) error
	DeductCoins(ctx context.Context, userID uuid.UUID, coins int64, reason string) error
	
	// 交易记录
	CreateRechargeRecord(ctx context.Context, record *domain.RechargeRecord) error
	CreateConsumptionRecord(ctx context.Context, record *domain.ConsumptionRecord) error
	GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.TransactionRecord, error)
}

// Wallet 钱包领域模型
type Wallet struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Balance      float64   `json:"balance" db:"balance"`
	Currency     string    `json:"currency" db:"currency"`
	ExchangeRate int       `json:"exchange_rate" db:"exchange_rate"`
	
	// 限制
	DailyLimit    float64 `json:"daily_limit" db:"daily_limit"`
	MonthlyLimit  float64 `json:"monthly_limit" db:"monthly_limit"`
	DailySpent    float64 `json:"daily_spent" db:"daily_spent"`
	MonthlySpent  float64 `json:"monthly_spent" db:"monthly_spent"`
	LastResetDate string  `json:"last_reset_date" db:"last_reset_date"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionRecord 交易记录
type TransactionRecord struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Type        string    `json:"type" db:"type"`        // recharge, consumption
	Amount      float64   `json:"amount" db:"amount"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// walletRepository 钱包仓库实现
type walletRepository struct {
	db *sqlx.DB
}

// NewWalletRepository 创建钱包仓库
func NewWalletRepository(db *sqlx.DB) WalletRepository {
	return &walletRepository{db: db}
}

// GetWalletByUserID 根据用户ID获取钱包
func (r *walletRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	var wallet domain.Wallet
	query := `
		SELECT id, user_id, balance, currency, exchange_rate,
			   daily_limit, monthly_limit, daily_spent, monthly_spent, last_reset_date,
			   created_at, updated_at
		FROM wallets WHERE user_id = $1`
	
	err := r.db.GetContext(ctx, &wallet, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	
	return &wallet, nil
}

// UpdateWalletBalance 更新钱包余额
func (r *walletRepository) UpdateWalletBalance(ctx context.Context, userID uuid.UUID, amount float64) error {
	query := `
		UPDATE wallets 
		SET balance = balance + $2, updated_at = NOW()
		WHERE user_id = $1 AND balance + $2 >= 0`
	
	result, err := r.db.ExecContext(ctx, query, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		if amount < 0 {
			return domain.ErrInsufficientBalance
		}
		return domain.ErrNotFound
	}
	
	return nil
}

// AddCoins 增加金币
func (r *walletRepository) AddCoins(ctx context.Context, userID uuid.UUID, coins int64, reason string) error {
	// 根据汇率计算金额
	wallet, err := r.GetWalletByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	
	amount := float64(coins) / float64(wallet.ExchangeRate)
	
	return r.UpdateWalletBalance(ctx, userID, amount)
}

// DeductCoins 扣除金币
func (r *walletRepository) DeductCoins(ctx context.Context, userID uuid.UUID, coins int64, reason string) error {
	// 根据汇率计算金额
	wallet, err := r.GetWalletByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	
	amount := float64(coins) / float64(wallet.ExchangeRate)
	
	return r.UpdateWalletBalance(ctx, userID, -amount)
}

// CreateRechargeRecord 创建充值记录
func (r *walletRepository) CreateRechargeRecord(ctx context.Context, record *domain.RechargeRecord) error {
	query := `
		INSERT INTO recharge_records (
			id, user_id, wallet_id, amount, currency, payment_amount, payment_currency,
			payment_method, payment_id, transaction_id, status, discount_rate, bonus_amount,
			card_code, remark, completed_at
		) VALUES (
			:id, :user_id, :wallet_id, :amount, :currency, :payment_amount, :payment_currency,
			:payment_method, :payment_id, :transaction_id, :status, :discount_rate, :bonus_amount,
			:card_code, :remark, :completed_at
		)`
	
	_, err := r.db.NamedExecContext(ctx, query, record)
	if err != nil {
		return fmt.Errorf("failed to create recharge record: %w", err)
	}
	
	return nil
}

// CreateConsumptionRecord 创建消费记录
func (r *walletRepository) CreateConsumptionRecord(ctx context.Context, record *domain.ConsumptionRecord) error {
	query := `
		INSERT INTO consumption_records (
			id, user_id, wallet_id, amount, currency, service_type, service_id,
			model_name, billing_unit, billing_quantity, unit_price, description, metadata
		) VALUES (
			:id, :user_id, :wallet_id, :amount, :currency, :service_type, :service_id,
			:model_name, :billing_unit, :billing_quantity, :unit_price, :description, :metadata
		)`
	
	_, err := r.db.NamedExecContext(ctx, query, record)
	if err != nil {
		return fmt.Errorf("failed to create consumption record: %w", err)
	}
	
	return nil
}

// GetUserTransactionHistory 获取用户交易历史
func (r *walletRepository) GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.TransactionRecord, error) {
	var records []*domain.TransactionRecord
	
	// 联合查询充值和消费记录
	query := `
		(SELECT id, user_id, 'recharge' as type, payment_amount as amount, 
		        CONCAT('充值 ', amount, ' 金币') as description, created_at
		 FROM recharge_records 
		 WHERE user_id = $1 AND status = 'completed')
		UNION ALL
		(SELECT id, user_id, 'consumption' as type, amount, description, created_at
		 FROM consumption_records 
		 WHERE user_id = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	err := r.db.SelectContext(ctx, &records, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	
	return records, nil
}
