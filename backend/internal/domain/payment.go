package domain

import (
	"time"

	"github.com/google/uuid"
)

// PaymentCard 支付卡片
type PaymentCard struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// 卡片信息
	CardNumber     string `json:"card_number" db:"card_number"` // 卡号（脱敏显示）
	CardNumberHash string `json:"-" db:"card_number_hash"`      // 卡号哈希（存储）
	CardType       string `json:"card_type" db:"card_type"`     // 卡片类型：debit, credit
	BankName       string `json:"bank_name" db:"bank_name"`     // 银行名称
	BankCode       string `json:"bank_code" db:"bank_code"`     // 银行代码

	// 持卡人信息
	CardholderName string `json:"cardholder_name" db:"cardholder_name"` // 持卡人姓名

	// 邮箱绑定（安全功能）
	BoundEmail    *string    `json:"bound_email" db:"bound_email"`       // 绑定的邮箱
	BoundAt       *time.Time `json:"bound_at" db:"bound_at"`             // 绑定时间
	EmailVerified bool       `json:"email_verified" db:"email_verified"` // 邮箱是否验证

	// 卡片余额
	Balance  float64 `json:"balance" db:"balance"`   // 卡片余额（金币）
	Currency string  `json:"currency" db:"currency"` // 货币类型，默认 "coins"

	// 状态管理
	IsDefault bool `json:"is_default" db:"is_default"` // 是否默认卡片
	IsActive  bool `json:"is_active" db:"is_active"`   // 是否激活
	IsFrozen  bool `json:"is_frozen" db:"is_frozen"`   // 是否冻结

	// 验证信息
	IsVerified bool       `json:"is_verified" db:"is_verified"` // 是否已验证
	VerifiedAt *time.Time `json:"verified_at" db:"verified_at"` // 验证时间

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentPassword 支付密码
type PaymentPassword struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// 密码信息
	PasswordHash string `json:"-" db:"password_hash"` // 支付密码哈希
	Salt         string `json:"-" db:"salt"`          // 盐值

	// 安全设置
	FailedAttempts int        `json:"failed_attempts" db:"failed_attempts"` // 失败次数
	LockedUntil    *time.Time `json:"locked_until" db:"locked_until"`       // 锁定到期时间
	LastFailedAt   *time.Time `json:"last_failed_at" db:"last_failed_at"`   // 最后失败时间

	// 状态管理
	IsActive bool `json:"is_active" db:"is_active"` // 是否激活

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RechargeOrder 充值订单
type RechargeOrder struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// 订单信息
	OrderNo      string  `json:"order_no" db:"order_no"`           // 订单号
	Amount       float64 `json:"amount" db:"amount"`               // 充值金额（人民币）
	CoinsAmount  int64   `json:"coins_amount" db:"coins_amount"`   // 金币数量
	ExchangeRate int     `json:"exchange_rate" db:"exchange_rate"` // 汇率（1元=N金币）

	// 支付信息
	PaymentMethod string     `json:"payment_method" db:"payment_method"`   // 支付方式：card, alipay, wechat
	PaymentCardID *uuid.UUID `json:"payment_card_id" db:"payment_card_id"` // 支付卡片ID

	// 折扣信息
	OriginalAmount float64 `json:"original_amount" db:"original_amount"` // 原始金额
	DiscountAmount float64 `json:"discount_amount" db:"discount_amount"` // 折扣金额
	BonusCoins     int64   `json:"bonus_coins" db:"bonus_coins"`         // 赠送金币
	CardCode       *string `json:"card_code" db:"card_code"`             // 使用的卡密

	// 订单状态
	Status        string `json:"status" db:"status"`                 // pending, processing, completed, failed, cancelled
	PaymentStatus string `json:"payment_status" db:"payment_status"` // unpaid, paid, refunded

	// 第三方支付信息
	ThirdPartyOrderNo *string `json:"third_party_order_no" db:"third_party_order_no"` // 第三方订单号
	PaymentURL        *string `json:"payment_url" db:"payment_url"`                   // 支付链接

	// 失败信息
	FailureReason *string `json:"failure_reason" db:"failure_reason"` // 失败原因

	// 时间戳
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	ExpiredAt   *time.Time `json:"expired_at" db:"expired_at"` // 订单过期时间
}

// 请求和响应结构

// BindCardRequest 绑卡请求
type BindCardRequest struct {
	CardNumber     string `json:"card_number" validate:"required,len=16"`
	CardholderName string `json:"cardholder_name" validate:"required,min=2,max=50"`
	BankName       string `json:"bank_name" validate:"required,min=2,max=100"`
	BankCode       string `json:"bank_code" validate:"required,len=4"`
	CardType       string `json:"card_type" validate:"required,oneof=debit credit"`
	IsDefault      bool   `json:"is_default"`
	Email          string `json:"email" validate:"required,email"` // 绑定邮箱
}

// BindCardWithEmailRequest 邮箱验证绑卡请求
type BindCardWithEmailRequest struct {
	CardNumber     string `json:"card_number" validate:"required,len=16"`
	CardholderName string `json:"cardholder_name" validate:"required,min=2,max=50"`
	BankName       string `json:"bank_name" validate:"required,min=2,max=100"`
	BankCode       string `json:"bank_code" validate:"required,len=4"`
	CardType       string `json:"card_type" validate:"required,oneof=debit credit"`
	IsDefault      bool   `json:"is_default"`
	Email          string `json:"email" validate:"required,email"`
	EmailCode      string `json:"email_code" validate:"required,len=6"` // 邮箱验证码
}

// UnbindCardRequest 解绑卡片请求
type UnbindCardRequest struct {
	CardID    uuid.UUID `json:"card_id" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	EmailCode string    `json:"email_code" validate:"required,len=6"`
}

// ResetPaymentPasswordRequest 重置支付密码请求
type ResetPaymentPasswordRequest struct {
	Email           string `json:"email" validate:"required,email"`
	EmailCode       string `json:"email_code" validate:"required,len=6"`
	NewPassword     string `json:"new_password" validate:"required,len=6,numeric"`
	ConfirmPassword string `json:"confirm_password" validate:"required,len=6,numeric"`
}

// CardRechargeRequest 卡片充值请求
type CardRechargeRequest struct {
	CardID   uuid.UUID `json:"card_id" validate:"required"`
	Amount   float64   `json:"amount" validate:"required,min=1"`
	CardCode *string   `json:"card_code,omitempty"` // 可选的卡密
}

// CardTransferRequest 卡片转账请求
type CardTransferRequest struct {
	FromCardID uuid.UUID `json:"from_card_id" validate:"required"`
	ToCardID   uuid.UUID `json:"to_card_id" validate:"required"`
	Amount     float64   `json:"amount" validate:"required,min=0.01"`
	Password   string    `json:"password" validate:"required,len=6,numeric"`
}

// CardBalanceResponse 卡片余额响应
type CardBalanceResponse struct {
	CardID      uuid.UUID `json:"card_id"`
	CardNumber  string    `json:"card_number"`
	Balance     float64   `json:"balance"`
	Currency    string    `json:"currency"`
	LastUpdated time.Time `json:"last_updated"`
}

// FreezeCardRequest 冻结卡片请求
type FreezeCardRequest struct {
	CardID   uuid.UUID `json:"card_id" validate:"required"`
	Reason   string    `json:"reason" validate:"required,min=1,max=200"`
	Password string    `json:"password" validate:"required,len=6,numeric"`
}

// UnfreezeCardRequest 解冻卡片请求
type UnfreezeCardRequest struct {
	CardID   uuid.UUID `json:"card_id" validate:"required"`
	Password string    `json:"password" validate:"required,len=6,numeric"`
}

// DeleteCardRequest 删除卡片请求
type DeleteCardRequest struct {
	CardID   uuid.UUID `json:"card_id" validate:"required"`
	Password string    `json:"password" validate:"required,len=6,numeric"`
	Reason   string    `json:"reason" validate:"required,min=1,max=200"`
}

// CardStatusResponse 卡片状态响应
type CardStatusResponse struct {
	CardID     uuid.UUID `json:"card_id"`
	CardNumber string    `json:"card_number"`
	IsActive   bool      `json:"is_active"`
	IsFrozen   bool      `json:"is_frozen"`
	Status     string    `json:"status"` // active, frozen, deleted
	Message    string    `json:"message"`
}

// SetPaymentPasswordRequest 设置支付密码请求
type SetPaymentPasswordRequest struct {
	Password        string `json:"password" validate:"required,len=6,numeric"`
	ConfirmPassword string `json:"confirm_password" validate:"required,len=6,numeric"`
}

// VerifyPaymentPasswordRequest 验证支付密码请求
type VerifyPaymentPasswordRequest struct {
	Password string `json:"password" validate:"required,len=6,numeric"`
}

// CreateRechargeOrderRequest 创建充值订单请求
type CreateRechargeOrderRequest struct {
	Amount        float64    `json:"amount" validate:"required,min=1,max=10000"`
	PaymentMethod string     `json:"payment_method" validate:"required,oneof=card alipay wechat"`
	PaymentCardID *uuid.UUID `json:"payment_card_id,omitempty"`
	CardCode      *string    `json:"card_code,omitempty" validate:"omitempty,len=16,numeric"`
}

// AdminRechargeRequest 管理员代充请求
type AdminRechargeRequest struct {
	UserIdentifier  string `json:"user_identifier" validate:"required"` // 用户ID或邮箱
	CoinsAmount     int64  `json:"coins_amount" validate:"required,min=1"`
	Reason          string `json:"reason" validate:"required,min=5,max=200"`
	PaymentPassword string `json:"payment_password" validate:"required,len=6,numeric"`
}

// ProxyRechargeRequest 代充请求
type ProxyRechargeRequest struct {
	TargetUserID    uuid.UUID `json:"target_user_id" validate:"required"`                 // 目标用户ID
	TargetEmail     *string   `json:"target_email,omitempty"`                             // 目标用户邮箱（可选）
	Amount          float64   `json:"amount" validate:"required,min=1"`                   // 充值金额
	PaymentCardID   uuid.UUID `json:"payment_card_id" validate:"required"`                // 支付卡片ID
	Message         string    `json:"message,omitempty"`                                  // 代充留言
	PaymentPassword string    `json:"payment_password" validate:"required,len=6,numeric"` // 支付密码
}

// ProxyRechargeResponse 代充响应
type ProxyRechargeResponse struct {
	OrderID        uuid.UUID `json:"order_id"`
	OrderNo        string    `json:"order_no"`
	TargetUserID   uuid.UUID `json:"target_user_id"`
	TargetUsername string    `json:"target_username"`
	PayerUserID    uuid.UUID `json:"payer_user_id"`
	PayerUsername  string    `json:"payer_username"`
	OriginalAmount float64   `json:"original_amount"` // 原始金额
	ActualAmount   float64   `json:"actual_amount"`   // 实际支付金额
	DiscountAmount float64   `json:"discount_amount"` // 折扣金额
	CoinsAmount    int64     `json:"coins_amount"`    // 基础金币数量
	BonusCoins     int64     `json:"bonus_coins"`     // 返利金币
	TotalCoins     int64     `json:"total_coins"`     // 总金币数量
	CardType       string    `json:"card_type"`       // 卡片类型
	Message        string    `json:"message"`         // 代充留言
	CreatedAt      time.Time `json:"created_at"`
}

// PaymentOrderResponse 支付订单响应
type PaymentOrderResponse struct {
	OrderID     uuid.UUID `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	Amount      float64   `json:"amount"`
	CoinsAmount int64     `json:"coins_amount"`
	PaymentURL  *string   `json:"payment_url,omitempty"`
	QRCode      *string   `json:"qr_code,omitempty"`
	ExpiredAt   time.Time `json:"expired_at"`
}

// UserPaymentInfo 用户支付信息
type UserPaymentInfo struct {
	HasPaymentPassword bool           `json:"has_payment_password"`
	PaymentCards       []*PaymentCard `json:"payment_cards"`
	DefaultCardID      *uuid.UUID     `json:"default_card_id"`
}

// RechargePackage 充值套餐
type RechargePackage struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Amount       float64   `json:"amount" db:"amount"`               // 人民币金额
	CoinsAmount  int64     `json:"coins_amount" db:"coins_amount"`   // 金币数量
	BonusCoins   int64     `json:"bonus_coins" db:"bonus_coins"`     // 赠送金币
	DiscountRate float64   `json:"discount_rate" db:"discount_rate"` // 折扣率
	IsPopular    bool      `json:"is_popular" db:"is_popular"`       // 是否热门
	IsActive     bool      `json:"is_active" db:"is_active"`         // 是否激活
	SortOrder    int       `json:"sort_order" db:"sort_order"`       // 排序
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// AdminRechargeRecord 管理员代充记录
type AdminRechargeRecord struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	TargetUserID           uuid.UUID  `json:"target_user_id" db:"target_user_id"`
	AdminUserID            uuid.UUID  `json:"admin_user_id" db:"admin_user_id"`
	CoinsAmount            int64      `json:"coins_amount" db:"coins_amount"`
	Reason                 string     `json:"reason" db:"reason"`
	AdminPaymentVerified   bool       `json:"admin_payment_verified" db:"admin_payment_verified"`
	AdminPaymentVerifiedAt *time.Time `json:"admin_payment_verified_at" db:"admin_payment_verified_at"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ConfigKey   string    `json:"config_key" db:"config_key"`
	ConfigValue string    `json:"config_value" db:"config_value"`
	Description string    `json:"description" db:"description"`
	ConfigType  string    `json:"config_type" db:"config_type"` // int, float, string, bool
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CoinExchangeConfig 金币汇率配置
type CoinExchangeConfig struct {
	Rate        int     `json:"rate"`       // 1元 = N金币
	MinAmount   float64 `json:"min_amount"` // 最小充值金额
	MaxAmount   float64 `json:"max_amount"` // 最大充值金额
	Description string  `json:"description"`
}

// 常量定义
const (
	// 支付方式
	PaymentMethodCard   = "card"
	PaymentMethodAlipay = "alipay"
	PaymentMethodWechat = "wechat"

	// 订单状态
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusCompleted  = "completed"
	OrderStatusFailed     = "failed"
	OrderStatusCancelled  = "cancelled"

	// 支付状态
	PaymentStatusUnpaid   = "unpaid"
	PaymentStatusPaid     = "paid"
	PaymentStatusRefunded = "refunded"

	// 卡片类型
	CardTypeDebit  = "debit"
	CardTypeCredit = "credit"

	// 支付密码最大失败次数
	MaxPaymentPasswordFailures = 5

	// 支付密码锁定时间（分钟）
	PaymentPasswordLockDuration = 30

	// 系统配置键
	ConfigKeyCoinExchangeRate  = "coin_exchange_rate"
	ConfigKeyMinRechargeAmount = "min_recharge_amount"
	ConfigKeyMaxRechargeAmount = "max_recharge_amount"

	// 默认金币汇率
	DefaultCoinExchangeRate = 10 // 1元 = 10金币

	// 卡片交易类型
	CardTransactionRecharge    = "recharge"     // 充值
	CardTransactionTransferIn  = "transfer_in"  // 转入
	CardTransactionTransferOut = "transfer_out" // 转出
	CardTransactionConsume     = "consume"      // 消费

	// 卡片操作类型
	CardOperationFreeze   = "freeze"   // 冻结
	CardOperationUnfreeze = "unfreeze" // 解冻
	CardOperationDelete   = "delete"   // 删除
	CardOperationActivate = "activate" // 激活

	// 卡片状态
	CardStatusActive  = "active"  // 正常
	CardStatusFrozen  = "frozen"  // 冻结
	CardStatusDeleted = "deleted" // 已删除
)

// CardTransaction 卡片交易记录
type CardTransaction struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	CardID          uuid.UUID  `json:"card_id" db:"card_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	TransactionType string     `json:"transaction_type" db:"transaction_type"` // recharge, transfer_in, transfer_out, consume
	Amount          float64    `json:"amount" db:"amount"`
	BalanceBefore   float64    `json:"balance_before" db:"balance_before"`
	BalanceAfter    float64    `json:"balance_after" db:"balance_after"`
	RelatedCardID   *uuid.UUID `json:"related_card_id" db:"related_card_id"` // 转账相关的另一张卡
	OrderID         *uuid.UUID `json:"order_id" db:"order_id"`               // 关联的订单ID
	Description     string     `json:"description" db:"description"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// CardOperation 卡片操作记录
type CardOperation struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	CardID        uuid.UUID  `json:"card_id" db:"card_id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	OperationType string     `json:"operation_type" db:"operation_type"` // freeze, unfreeze, delete, activate
	Reason        string     `json:"reason" db:"reason"`
	OperatorID    *uuid.UUID `json:"operator_id" db:"operator_id"` // 操作员ID（管理员操作时）
	IPAddress     string     `json:"ip_address" db:"ip_address"`
	UserAgent     string     `json:"user_agent" db:"user_agent"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// ProxyRechargeOrder 代充订单记录
type ProxyRechargeOrder struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrderNo        string     `json:"order_no" db:"order_no"`
	PayerUserID    uuid.UUID  `json:"payer_user_id" db:"payer_user_id"`     // 付款人ID
	TargetUserID   uuid.UUID  `json:"target_user_id" db:"target_user_id"`   // 目标用户ID
	PaymentCardID  uuid.UUID  `json:"payment_card_id" db:"payment_card_id"` // 支付卡片ID
	OriginalAmount float64    `json:"original_amount" db:"original_amount"` // 原始金额
	ActualAmount   float64    `json:"actual_amount" db:"actual_amount"`     // 实际支付金额
	DiscountAmount float64    `json:"discount_amount" db:"discount_amount"` // 折扣金额
	CoinsAmount    int64      `json:"coins_amount" db:"coins_amount"`       // 基础金币数量
	BonusCoins     int64      `json:"bonus_coins" db:"bonus_coins"`         // 返利金币
	TotalCoins     int64      `json:"total_coins" db:"total_coins"`         // 总金币数量
	ExchangeRate   int        `json:"exchange_rate" db:"exchange_rate"`     // 汇率
	CardType       string     `json:"card_type" db:"card_type"`             // 卡片类型
	Message        string     `json:"message" db:"message"`                 // 代充留言
	Status         string     `json:"status" db:"status"`                   // 订单状态
	PaymentStatus  string     `json:"payment_status" db:"payment_status"`   // 支付状态
	CompletedAt    *time.Time `json:"completed_at" db:"completed_at"`       // 完成时间
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}
