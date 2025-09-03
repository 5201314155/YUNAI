package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"yunai/internal/domain"
	"yunai/internal/repository"
	"yunai/pkg/auth"
	"yunai/pkg/email"
)

// AuthService 认证服务接口
type AuthService interface {
	// 用户注册和登录
	Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error)
	Login(ctx context.Context, req *domain.LoginRequest, clientIP net.IP) (*domain.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// 用户管理
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *domain.ChangePasswordRequest) error

	// TOTP 双因子认证
	SetupTOTP(ctx context.Context, userID uuid.UUID) (*auth.TOTPSetup, error)
	EnableTOTP(ctx context.Context, userID uuid.UUID, req *domain.EnableTOTPRequest) error
	DisableTOTP(ctx context.Context, userID uuid.UUID, req *domain.DisableTOTPRequest) error

	// 令牌验证
	ValidateAccessToken(tokenString string) (*auth.JWTClaims, error)

	// 权限管理
	GrantPermission(ctx context.Context, userID uuid.UUID, permission string, grantedBy uuid.UUID, expiresAt *time.Time) error
	RevokePermission(ctx context.Context, userID uuid.UUID, permission string) error
	HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)

	// 邮箱验证
	SendVerificationEmail(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error
	ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error

	// 密码重置
	SendPasswordResetEmail(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ValidateResetToken(ctx context.Context, token string) (*domain.User, error)
}

// authService 认证服务实现
type authService struct {
	userRepo     repository.UserRepository
	authRepo     repository.AuthRepository
	jwtManager   *auth.JWTManager
	totpManager  *auth.TOTPManager
	emailService email.EmailService
	logger       *logrus.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo repository.UserRepository,
	authRepo repository.AuthRepository,
	jwtManager *auth.JWTManager,
	logger *logrus.Logger,
) AuthService {
	// 创建邮件服务（使用Mock服务用于开发测试）
	emailService := email.NewMockEmailService(logger)
	totpManager := auth.NewTOTPManager("YUNAI")

	return &authService{
		userRepo:     userRepo,
		authRepo:     authRepo,
		jwtManager:   jwtManager,
		totpManager:  totpManager,
		emailService: emailService,
		logger:       logger,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username exists: %w", err)
	}
	if exists {
		return nil, domain.AppErrUsernameExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email exists: %w", err)
	}
	if exists {
		return nil, domain.AppErrEmailExists
	}

	// 加密密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &domain.User{
		ID:            uuid.New(),
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  string(passwordHash),
		UserType:      req.UserType,
		Nickname:      req.Nickname,
		EmailVerified: false, // 默认未验证
		IsActive:      true,
		IsBanned:      false,
		LoginCount:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 设置默认用户类型
	if user.UserType == "" {
		user.UserType = domain.UserTypeBasic
	}

	// 保存用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("User registered successfully")

	return user, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *domain.LoginRequest, clientIP net.IP) (*domain.LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.AppErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if !user.CanLogin() {
		if !user.IsUserActive() {
			return nil, domain.AppErrAccountLocked
		}
		if user.IsBanned {
			return nil, domain.AppErrAccountBanned
		}
		if !user.EmailVerified {
			return nil, domain.AppErrEmailNotVerified
		}
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.AppErrInvalidCredentials
	}

	// 如果启用了 TOTP，验证验证码
	if user.TOTPEnabled {
		if req.TOTPCode == nil || *req.TOTPCode == "" {
			return nil, domain.AppErrTOTPRequired
		}

		if !s.totpManager.ValidateCurrentCode(*user.TOTPSecret, *req.TOTPCode) {
			return nil, domain.AppErrInvalidTOTP
		}
	}

	// 生成令牌对
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(
		user.ID, user.Username, string(user.UserType))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 创建会话
	session := &domain.UserSession{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IPAddress:    &clientIP,
		ExpiresAt:    time.Now().Add(s.jwtManager.GetRefreshTokenDuration()),
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 更新登录信息
	user.UpdateLoginInfo(clientIP)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Warn("Failed to update login info")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"ip":       clientIP,
	}).Info("User logged in successfully")

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenDuration().Seconds()),
		User:         user,
	}, nil
}

// RefreshToken 刷新令牌
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error) {
	// 获取会话
	session, err := s.userRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.AppErrInvalidToken
	}

	// 检查会话是否过期
	if session.IsExpired() {
		// 删除过期会话
		s.userRepo.DeleteSession(ctx, session.ID)
		return nil, domain.AppErrTokenExpired
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if !user.CanLogin() {
		return nil, domain.AppErrAccountLocked
	}

	// 生成新的访问令牌
	accessToken, err := s.jwtManager.GenerateAccessToken(
		user.ID, user.Username, string(user.UserType))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // 刷新令牌保持不变
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenDuration().Seconds()),
		User:         user,
	}, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	// 获取会话
	session, err := s.userRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil // 会话不存在或已过期，视为成功
	}

	// 验证会话属于指定用户
	if session.UserID != userID {
		return domain.AppErrUnauthorized
	}

	// 删除会话
	if err := s.userRepo.DeleteSession(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": session.ID,
	}).Info("User logged out successfully")

	return nil
}

// LogoutAll 用户全部登出
func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.DeleteUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("User logged out from all devices")

	return nil
}

// GetUserByID 根据ID获取用户
func (s *authService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *authService) UpdateUser(ctx context.Context, userID uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 更新字段
	if req.Nickname != nil {
		user.Nickname = req.Nickname
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("User updated successfully")

	return user, nil
}

// ChangePassword 修改密码
func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, req *domain.ChangePasswordRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return domain.AppErrInvalidCredentials
	}

	// 如果启用了 TOTP，验证验证码
	if user.TOTPEnabled {
		if req.TOTPCode == nil || *req.TOTPCode == "" {
			return domain.AppErrTOTPRequired
		}

		if !s.totpManager.ValidateCurrentCode(*user.TOTPSecret, *req.TOTPCode) {
			return domain.AppErrInvalidTOTP
		}
	}

	// 加密新密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	user.PasswordHash = string(passwordHash)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 删除所有会话，强制重新登录
	if err := s.userRepo.DeleteUserSessions(ctx, userID); err != nil {
		s.logger.WithError(err).Warn("Failed to delete user sessions after password change")
	}

	s.logger.WithField("user_id", userID).Info("Password changed successfully")

	return nil
}

// SetupTOTP 设置TOTP双因子认证
func (s *authService) SetupTOTP(ctx context.Context, userID uuid.UUID) (*auth.TOTPSetup, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 生成 TOTP 设置
	setup, err := s.totpManager.SetupTOTP(user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to setup TOTP: %w", err)
	}

	// 临时保存密钥（未启用）
	user.TOTPSecret = &setup.Secret
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("TOTP setup initiated")

	return setup, nil
}

// EnableTOTP 启用TOTP双因子认证
func (s *authService) EnableTOTP(ctx context.Context, userID uuid.UUID, req *domain.EnableTOTPRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 检查是否已设置密钥
	if user.TOTPSecret == nil {
		return fmt.Errorf("TOTP not setup")
	}

	// 验证验证码
	if !s.totpManager.ValidateCurrentCode(*user.TOTPSecret, req.TOTPCode) {
		return domain.AppErrInvalidTOTP
	}

	// 启用 TOTP
	user.TOTPEnabled = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("TOTP enabled successfully")

	return nil
}

// DisableTOTP 禁用TOTP双因子认证
func (s *authService) DisableTOTP(ctx context.Context, userID uuid.UUID, req *domain.DisableTOTPRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 检查是否启用了 TOTP
	if !user.TOTPEnabled {
		return domain.AppErrTOTPNotEnabled
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return domain.AppErrInvalidCredentials
	}

	// 验证 TOTP 验证码
	if !s.totpManager.ValidateCurrentCode(*user.TOTPSecret, req.TOTPCode) {
		return domain.AppErrInvalidTOTP
	}

	// 禁用 TOTP
	user.TOTPEnabled = false
	user.TOTPSecret = nil
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to disable TOTP: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("TOTP disabled successfully")

	return nil
}

// ValidateAccessToken 验证访问令牌
func (s *authService) ValidateAccessToken(tokenString string) (*auth.JWTClaims, error) {
	return s.jwtManager.ValidateAccessToken(tokenString)
}

// GrantPermission 授予权限
func (s *authService) GrantPermission(ctx context.Context, userID uuid.UUID, permission string, grantedBy uuid.UUID, expiresAt *time.Time) error {
	perm := &domain.UserPermission{
		ID:         uuid.New(),
		UserID:     userID,
		Permission: permission,
		GrantedBy:  &grantedBy,
		GrantedAt:  time.Now(),
		ExpiresAt:  expiresAt,
	}

	if err := s.userRepo.GrantPermission(ctx, perm); err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"permission": permission,
		"granted_by": grantedBy,
	}).Info("Permission granted")

	return nil
}

// RevokePermission 撤销权限
func (s *authService) RevokePermission(ctx context.Context, userID uuid.UUID, permission string) error {
	if err := s.userRepo.RevokePermission(ctx, userID, permission); err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"permission": permission,
	}).Info("Permission revoked")

	return nil
}

// HasPermission 检查权限
func (s *authService) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	return s.userRepo.HasPermission(ctx, userID, permission)
}

// SendVerificationEmail 发送邮箱验证码
func (s *authService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 生成验证码
	code, err := s.generateVerificationCode()
	if err != nil {
		return fmt.Errorf("failed to generate verification code: %w", err)
	}

	// 保存验证码到Redis缓存
	verificationKey := fmt.Sprintf("%s:email_verification", userID.String())
	if err := s.authRepo.SetVerificationCode(ctx, verificationKey, code, 10*time.Minute); err != nil {
		return fmt.Errorf("failed to save verification code: %w", err)
	}

	// 发送邮件
	if err := s.emailService.SendVerificationEmail(user.Email, user.Username, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   user.Email,
	}).Info("Verification email sent")

	return nil
}

// VerifyEmail 验证邮箱
func (s *authService) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证验证码
	verificationKey := fmt.Sprintf("%s:email_verification", userID.String())
	storedCode, err := s.authRepo.GetVerificationCode(ctx, verificationKey)
	if err != nil {
		return domain.AppErrInvalidVerificationCode
	}

	// 检查验证码
	if storedCode != code {
		return domain.AppErrInvalidVerificationCode
	}

	// 删除验证码（标记为已使用）
	if err := s.authRepo.DeleteVerificationCode(ctx, verificationKey); err != nil {
		s.logger.WithError(err).Warn("Failed to delete verification code")
	}

	// 标记邮箱为已验证
	user.EmailVerified = true
	user.EmailVerifiedAt = &time.Time{}
	*user.EmailVerifiedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user email verification status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   user.Email,
	}).Info("Email verified successfully")

	return nil
}

// ResendVerificationEmail 重新发送验证邮件
func (s *authService) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// 检查是否已经验证
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.EmailVerified {
		return domain.AppErrEmailAlreadyVerified
	}

	// 删除旧的验证码
	verificationKey := fmt.Sprintf("%s:email_verification", userID.String())
	if err := s.authRepo.DeleteVerificationCode(ctx, verificationKey); err != nil {
		s.logger.WithError(err).Warn("Failed to delete old verification codes")
	}

	// 发送新的验证码
	return s.SendVerificationEmail(ctx, userID)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *authService) SendPasswordResetEmail(ctx context.Context, email string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// 为了安全，即使用户不存在也返回成功
			s.logger.WithField("email", email).Warn("Password reset requested for non-existent user")
			return nil
		}
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	// 生成重置Token
	resetToken, err := s.generateResetToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// 保存重置Token到Redis缓存
	resetKey := fmt.Sprintf("%s:password_reset", resetToken)
	if err := s.authRepo.SetVerificationCode(ctx, resetKey, user.ID.String(), 30*time.Minute); err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// 构建重置URL
	resetURL := fmt.Sprintf("https://yunai.com/reset-password?token=%s", resetToken)

	// 发送重置邮件
	if err := s.emailService.SendPasswordResetEmail(email, user.Username, resetURL); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   email,
	}).Info("Password reset email sent")

	return nil
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// 验证重置Token
	user, err := s.ValidateResetToken(ctx, token)
	if err != nil {
		return err
	}

	// 加密新密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	user.PasswordHash = string(passwordHash)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 标记重置Token为已使用
	if err := s.markResetTokenAsUsed(ctx, token); err != nil {
		s.logger.WithError(err).Warn("Failed to mark reset token as used")
	}

	// 删除所有会话，强制重新登录
	if err := s.userRepo.DeleteUserSessions(ctx, user.ID); err != nil {
		s.logger.WithError(err).Warn("Failed to delete user sessions after password reset")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Password reset successfully")

	return nil
}

// ValidateResetToken 验证重置Token
func (s *authService) ValidateResetToken(ctx context.Context, token string) (*domain.User, error) {
	// 从Redis获取用户ID
	resetKey := fmt.Sprintf("%s:password_reset", token)
	userIDStr, err := s.authRepo.GetVerificationCode(ctx, resetKey)
	if err != nil {
		return nil, domain.AppErrInvalidResetToken
	}

	// 解析用户ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.AppErrInvalidResetToken
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// generateVerificationCode 生成验证码
func (s *authService) generateVerificationCode() (string, error) {
	// 生成6位数字验证码
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// 转换为6位数字
	code := ""
	for _, b := range bytes {
		code += fmt.Sprintf("%02d", int(b)%100)
	}

	return code[:6], nil
}

// generateResetToken 生成重置Token
func (s *authService) generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// markResetTokenAsUsed 标记重置Token为已使用
func (s *authService) markResetTokenAsUsed(ctx context.Context, token string) error {
	// 删除Redis中的重置Token
	resetKey := fmt.Sprintf("%s:password_reset", token)
	return s.authRepo.DeleteVerificationCode(ctx, resetKey)
}
