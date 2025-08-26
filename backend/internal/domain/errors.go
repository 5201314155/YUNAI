package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// 通用错误
var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("resource conflict")
	ErrInternalError = errors.New("internal server error")
)

// 认证相关错误
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidTOTP        = errors.New("invalid TOTP code")
	ErrTOTPRequired       = errors.New("TOTP code required")
	ErrTOTPNotEnabled     = errors.New("TOTP not enabled")
	ErrAccountLocked      = errors.New("account locked")
	ErrAccountBanned      = errors.New("account banned")
	ErrEmailNotVerified   = errors.New("email not verified")
)

// 钱包相关错误
var (
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrTransactionFailed    = errors.New("transaction failed")
	ErrDailyLimitExceeded   = errors.New("daily limit exceeded")
	ErrMonthlyLimitExceeded = errors.New("monthly limit exceeded")
)

// 卡密相关错误
var (
	ErrCardNotFound           = errors.New("gift card not found")
	ErrCardExpired            = errors.New("gift card expired")
	ErrCardAlreadyUsed        = errors.New("gift card already used")
	ErrCardInactive           = errors.New("gift card inactive")
	ErrInvalidCardCode        = errors.New("invalid card code")
	ErrCardUsageLimitExceeded = errors.New("card usage limit exceeded")
	ErrInsufficientUserType   = errors.New("insufficient user type for card")
)

// 功能开关相关错误
var (
	ErrFeatureDisabled        = errors.New("feature disabled")
	ErrFeatureNotFound        = errors.New("feature flag not found")
	ErrInsufficientPermission = errors.New("insufficient permission for feature")
)

// 模型相关错误
var (
	ErrModelNotFound      = errors.New("model not found")
	ErrModelUnavailable   = errors.New("model unavailable")
	ErrModelNotSupported  = errors.New("model not supported")
	ErrInvalidModelParams = errors.New("invalid model parameters")
	ErrModelQuotaExceeded = errors.New("model quota exceeded")
)

// AppError 应用错误结构
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 返回包装的错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建新的应用错误
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewAppErrorWithDetails 创建带详情的应用错误
func NewAppErrorWithDetails(code, message, details string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
		Err:     err,
	}
}

// 错误码常量
const (
	// 通用错误码
	CodeSuccess         = "SUCCESS"
	CodeInvalidRequest  = "INVALID_REQUEST"
	CodeBadRequest      = "BAD_REQUEST"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeRateLimited     = "RATE_LIMITED"
	CodePaymentRequired = "PAYMENT_REQUIRED"

	// 认证错误码
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserExists         = "USER_EXISTS"
	CodeEmailExists        = "EMAIL_EXISTS"
	CodeUsernameExists     = "USERNAME_EXISTS"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeInvalidTOTP        = "INVALID_TOTP"
	CodeTOTPRequired       = "TOTP_REQUIRED"
	CodeTOTPNotEnabled     = "TOTP_NOT_ENABLED"
	CodeAccountLocked      = "ACCOUNT_LOCKED"
	CodeAccountBanned      = "ACCOUNT_BANNED"
	CodeEmailNotVerified   = "EMAIL_NOT_VERIFIED"

	// 钱包错误码
	CodeInsufficientBalance  = "INSUFFICIENT_BALANCE"
	CodeInvalidAmount        = "INVALID_AMOUNT"
	CodeWalletNotFound       = "WALLET_NOT_FOUND"
	CodeTransactionFailed    = "TRANSACTION_FAILED"
	CodeDailyLimitExceeded   = "DAILY_LIMIT_EXCEEDED"
	CodeMonthlyLimitExceeded = "MONTHLY_LIMIT_EXCEEDED"

	// 卡密错误码
	CodeCardNotFound           = "CARD_NOT_FOUND"
	CodeCardExpired            = "CARD_EXPIRED"
	CodeCardAlreadyUsed        = "CARD_ALREADY_USED"
	CodeCardInactive           = "CARD_INACTIVE"
	CodeInvalidCardCode        = "INVALID_CARD_CODE"
	CodeCardUsageLimitExceeded = "CARD_USAGE_LIMIT_EXCEEDED"
	CodeInsufficientUserType   = "INSUFFICIENT_USER_TYPE"

	// 功能开关错误码
	CodeFeatureDisabled        = "FEATURE_DISABLED"
	CodeFeatureNotFound        = "FEATURE_NOT_FOUND"
	CodeInsufficientPermission = "INSUFFICIENT_PERMISSION"

	// 模型错误码
	CodeModelNotFound      = "MODEL_NOT_FOUND"
	CodeModelUnavailable   = "MODEL_UNAVAILABLE"
	CodeModelNotSupported  = "MODEL_NOT_SUPPORTED"
	CodeInvalidModelParams = "INVALID_MODEL_PARAMS"
	CodeModelQuotaExceeded = "MODEL_QUOTA_EXCEEDED"
)

// 预定义的应用错误
var (
	AppErrInvalidCredentials = NewAppError(CodeInvalidCredentials, "用户名或密码错误", ErrInvalidCredentials)
	AppErrUserNotFound       = NewAppError(CodeUserNotFound, "用户不存在", ErrUserNotFound)
	AppErrUserExists         = NewAppError(CodeUserExists, "用户已存在", ErrUserExists)
	AppErrEmailExists        = NewAppError(CodeEmailExists, "邮箱已存在", ErrEmailExists)
	AppErrUsernameExists     = NewAppError(CodeUsernameExists, "用户名已存在", ErrUsernameExists)
	AppErrInvalidToken       = NewAppError(CodeInvalidToken, "无效的令牌", ErrInvalidToken)
	AppErrTokenExpired       = NewAppError(CodeTokenExpired, "令牌已过期", ErrTokenExpired)
	AppErrInvalidTOTP        = NewAppError(CodeInvalidTOTP, "无效的验证码", ErrInvalidTOTP)
	AppErrTOTPRequired       = NewAppError(CodeTOTPRequired, "需要双因子认证", ErrTOTPRequired)
	AppErrTOTPNotEnabled     = NewAppError(CodeTOTPNotEnabled, "双因子认证未启用", ErrTOTPNotEnabled)
	AppErrAccountLocked      = NewAppError(CodeAccountLocked, "账户已锁定", ErrAccountLocked)
	AppErrAccountBanned      = NewAppError(CodeAccountBanned, "账户已封禁", ErrAccountBanned)
	AppErrEmailNotVerified   = NewAppError(CodeEmailNotVerified, "邮箱未验证", ErrEmailNotVerified)

	AppErrInsufficientBalance = NewAppError(CodeInsufficientBalance, "余额不足", ErrInsufficientBalance)
	AppErrInvalidAmount       = NewAppError(CodeInvalidAmount, "无效的金额", ErrInvalidAmount)
	AppErrWalletNotFound      = NewAppError(CodeWalletNotFound, "钱包不存在", ErrWalletNotFound)

	AppErrCardNotFound    = NewAppError(CodeCardNotFound, "卡密不存在", ErrCardNotFound)
	AppErrCardExpired     = NewAppError(CodeCardExpired, "卡密已过期", ErrCardExpired)
	AppErrCardAlreadyUsed = NewAppError(CodeCardAlreadyUsed, "卡密已使用", ErrCardAlreadyUsed)
	AppErrInvalidCardCode = NewAppError(CodeInvalidCardCode, "无效的卡密", ErrInvalidCardCode)

	AppErrFeatureDisabled = NewAppError(CodeFeatureDisabled, "功能已禁用", ErrFeatureDisabled)
	AppErrModelNotFound   = NewAppError(CodeModelNotFound, "模型不存在", ErrModelNotFound)

	AppErrUnauthorized = NewAppError(CodeUnauthorized, "未授权访问", ErrUnauthorized)
	AppErrForbidden    = NewAppError(CodeForbidden, "权限不足", ErrForbidden)
)

// IsAppError 检查是否为应用错误
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeInvalidRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeInvalidCredentials:
		return http.StatusUnauthorized
	case CodeInsufficientBalance:
		return http.StatusBadRequest
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
