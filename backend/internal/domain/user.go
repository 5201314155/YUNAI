package domain

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// UserType 用户类型
type UserType string

const (
	UserTypeBasic   UserType = "basic"
	UserTypeVip     UserType = "vip"
	UserTypeCreator UserType = "creator"
	UserTypeAdmin   UserType = "admin"
)

// User 用户领域模型
type User struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Username string    `json:"username" db:"username"`
	Email    string    `json:"email" db:"email"`

	// 密码相关（不在 JSON 中暴露）
	PasswordHash string `json:"-" db:"password_hash"`

	// 用户类型
	UserType UserType `json:"user_type" db:"user_type"`

	// 个人信息
	Nickname  *string `json:"nickname" db:"nickname"`
	AvatarURL *string `json:"avatar_url" db:"avatar_url"`
	Bio       *string `json:"bio" db:"bio"`

	// 认证相关
	EmailVerified bool    `json:"email_verified" db:"email_verified"`
	TOTPSecret    *string `json:"-" db:"totp_secret"` // 不在 JSON 中暴露
	TOTPEnabled   bool    `json:"totp_enabled" db:"totp_enabled"`

	// 状态管理
	IsActive     bool       `json:"is_active" db:"is_active"`
	IsBanned     bool       `json:"is_banned" db:"is_banned"`
	BanReason    *string    `json:"ban_reason,omitempty" db:"ban_reason"`
	BanExpiresAt *time.Time `json:"ban_expires_at,omitempty" db:"ban_expires_at"`

	// 登录信息
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	LastLoginIP *net.IP    `json:"last_login_ip" db:"last_login_ip"`
	LoginCount  int        `json:"login_count" db:"login_count"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// IdentityType 身份类型
type IdentityType string

const (
	IdentityTypeReal      IdentityType = "real"      // 真实昵称
	IdentityTypeSpecified IdentityType = "specified" // 指定身份
)

// UserIdentity 用户身份
type UserIdentity struct {
	Type        IdentityType `json:"type"`              // 身份类型
	DisplayName string       `json:"display_name"`      // 显示名称
	Source      string       `json:"source"`            // 身份来源
	UserID      *uuid.UUID   `json:"user_id,omitempty"` // 关联的用户ID
	Context     string       `json:"context,omitempty"` // 身份上下文
	CreatedAt   time.Time    `json:"created_at"`        // 创建时间
}

// UserIdentityContext 用户身份上下文
type UserIdentityContext struct {
	UserID            uuid.UUID     `json:"user_id"`                       // 用户ID
	RealIdentity      *UserIdentity `json:"real_identity"`                 // 真实身份
	CharacterIdentity *UserIdentity `json:"character_identity,omitempty"`  // 角色上下文身份
	GroupChatIdentity *UserIdentity `json:"group_chat_identity,omitempty"` // 群聊上下文身份
	CurrentIdentity   *UserIdentity `json:"current_identity"`              // 当前使用的身份
}

// RegisterUserRequest 用户注册请求
type RegisterUserRequest struct {
	Username     string  `json:"username" validate:"required,min=3,max=50"`
	Email        string  `json:"email" validate:"required,email"`
	PasswordHash string  `json:"password_hash" validate:"required"`
	UserType     string  `json:"user_type" validate:"required,oneof=basic vip creator admin"`
	Nickname     *string `json:"nickname,omitempty" validate:"omitempty,max=100"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	Bio          *string `json:"bio,omitempty" validate:"omitempty,max=500"`
}

// IsValidUserType 检查用户类型是否有效
func IsValidUserType(userType string) bool {
	switch UserType(userType) {
	case UserTypeBasic, UserTypeVip, UserTypeCreator, UserTypeAdmin:
		return true
	default:
		return false
	}
}

// HasPermission 检查用户是否有指定权限
func (u *User) HasPermission(requiredType UserType) bool {
	userTypeLevel := map[UserType]int{
		UserTypeBasic:   0,
		UserTypeVip:     1,
		UserTypeCreator: 2,
		UserTypeAdmin:   3,
	}

	userLevel, exists := userTypeLevel[u.UserType]
	if !exists {
		return false
	}

	requiredLevel, exists := userTypeLevel[requiredType]
	if !exists {
		return false
	}

	return userLevel >= requiredLevel
}

// IsUserActive 检查用户是否处于活跃状态
func (u *User) IsUserActive() bool {
	if !u.IsActive || u.IsBanned {
		return false
	}

	// 检查封禁是否过期
	if u.BanExpiresAt != nil && time.Now().After(*u.BanExpiresAt) {
		return true // 封禁已过期，用户可以活跃
	}

	return !u.IsBanned
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.IsUserActive() && u.EmailVerified
}

// UpdateLoginInfo 更新登录信息
func (u *User) UpdateLoginInfo(ip net.IP) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = &ip
	u.LoginCount++
	u.UpdatedAt = now
}

// UserSession 用户会话
type UserSession struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"-" db:"refresh_token"` // 不在 JSON 中暴露
	DeviceInfo   *string   `json:"device_info" db:"device_info"`
	IPAddress    *net.IP   `json:"ip_address" db:"ip_address"`
	UserAgent    *string   `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// IsExpired 检查会话是否过期
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UserPermission 用户权限
type UserPermission struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	Permission string     `json:"permission" db:"permission"`
	GrantedBy  *uuid.UUID `json:"granted_by" db:"granted_by"`
	GrantedAt  time.Time  `json:"granted_at" db:"granted_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
}

// IsExpired 检查权限是否过期
func (p *UserPermission) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false // 永久权限
	}
	return time.Now().After(*p.ExpiresAt)
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username   string   `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email      string   `json:"email" validate:"required,email"`
	Password   string   `json:"password" validate:"required,min=8"`
	UserType   UserType `json:"user_type,omitempty"`
	Nickname   *string  `json:"nickname,omitempty" validate:"omitempty,max=100"`
	InviteCode *string  `json:"invite_code,omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string  `json:"username" validate:"required"`
	Password string  `json:"password" validate:"required"`
	TOTPCode *string `json:"totp_code,omitempty" validate:"omitempty,len=6,numeric"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *User  `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname  *string `json:"nickname,omitempty" validate:"omitempty,max=100"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,max=500"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string  `json:"old_password" validate:"required"`
	NewPassword string  `json:"new_password" validate:"required,min=8"`
	TOTPCode    *string `json:"totp_code,omitempty" validate:"omitempty,len=6,numeric"`
}

// EnableTOTPRequest 启用 TOTP 请求
type EnableTOTPRequest struct {
	TOTPCode string `json:"totp_code" validate:"required,len=6,numeric"`
}

// DisableTOTPRequest 禁用 TOTP 请求
type DisableTOTPRequest struct {
	Password string `json:"password" validate:"required"`
	TOTPCode string `json:"totp_code" validate:"required,len=6,numeric"`
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

// RechargeRecord 充值记录
type RechargeRecord struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	WalletID        uuid.UUID  `json:"wallet_id" db:"wallet_id"`
	Amount          float64    `json:"amount" db:"amount"`
	Currency        string     `json:"currency" db:"currency"`
	PaymentAmount   float64    `json:"payment_amount" db:"payment_amount"`
	PaymentCurrency string     `json:"payment_currency" db:"payment_currency"`
	PaymentMethod   string     `json:"payment_method" db:"payment_method"`
	PaymentID       *string    `json:"payment_id" db:"payment_id"`
	TransactionID   *string    `json:"transaction_id" db:"transaction_id"`
	Status          string     `json:"status" db:"status"`
	DiscountRate    float64    `json:"discount_rate" db:"discount_rate"`
	BonusAmount     float64    `json:"bonus_amount" db:"bonus_amount"`
	CardCode        *string    `json:"card_code" db:"card_code"`
	Remark          *string    `json:"remark" db:"remark"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
}

// ConsumptionRecord 消费记录
type ConsumptionRecord struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	WalletID        uuid.UUID  `json:"wallet_id" db:"wallet_id"`
	Amount          float64    `json:"amount" db:"amount"`
	Currency        string     `json:"currency" db:"currency"`
	ServiceType     string     `json:"service_type" db:"service_type"`
	ServiceID       *uuid.UUID `json:"service_id" db:"service_id"`
	ModelName       *string    `json:"model_name" db:"model_name"`
	BillingUnit     *string    `json:"billing_unit" db:"billing_unit"`
	BillingQuantity *float64   `json:"billing_quantity" db:"billing_quantity"`
	UnitPrice       *float64   `json:"unit_price" db:"unit_price"`
	Description     *string    `json:"description" db:"description"`
	Metadata        *string    `json:"metadata" db:"metadata"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// TransactionRecord 交易记录
type TransactionRecord struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Type        string    `json:"type" db:"type"` // recharge, consumption
	Amount      float64   `json:"amount" db:"amount"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
