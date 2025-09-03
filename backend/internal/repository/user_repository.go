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

// UserRepository 用户仓库接口
type UserRepository interface {
	// 用户基本操作
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 用户查询
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)
	ListActive(ctx context.Context) ([]*domain.User, error)
	Count(ctx context.Context) (int64, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// 会话管理
	CreateSession(ctx context.Context, session *domain.UserSession) error
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*domain.UserSession, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanExpiredSessions(ctx context.Context) error

	// 权限管理
	GrantPermission(ctx context.Context, permission *domain.UserPermission) error
	RevokePermission(ctx context.Context, userID uuid.UUID, permission string) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.UserPermission, error)
	HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

// userRepository 用户仓库实现
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, user_type, nickname, avatar_url, bio,
			email_verified, totp_secret, totp_enabled, is_active, is_banned
		) VALUES (
			:id, :username, :email, :password_hash, :user_type, :nickname, :avatar_url, :bio,
			:email_verified, :totp_secret, :totp_enabled, :is_active, :is_banned
		)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_username_key" {
					return domain.ErrUsernameExists
				}
				if pqErr.Constraint == "users_email_key" {
					return domain.ErrEmailExists
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, username, email, password_hash, user_type, nickname, avatar_url, bio,
			   email_verified, totp_secret, totp_enabled, is_active, is_banned, ban_reason, ban_expires_at,
			   last_login_at, last_login_ip, login_count, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, username, email, password_hash, user_type, nickname, avatar_url, bio,
			   email_verified, totp_secret, totp_enabled, is_active, is_banned, ban_reason, ban_expires_at,
			   last_login_at, last_login_ip, login_count, created_at, updated_at
		FROM users WHERE username = $1`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, username, email, password_hash, user_type, nickname, avatar_url, bio,
			   email_verified, totp_secret, totp_enabled, is_active, is_banned, ban_reason, ban_expires_at,
			   last_login_at, last_login_ip, login_count, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			username = :username, email = :email, password_hash = :password_hash,
			user_type = :user_type, nickname = :nickname, avatar_url = :avatar_url, bio = :bio,
			email_verified = :email_verified, totp_secret = :totp_secret, totp_enabled = :totp_enabled,
			is_active = :is_active, is_banned = :is_banned, ban_reason = :ban_reason, ban_expires_at = :ban_expires_at,
			last_login_at = :last_login_at, last_login_ip = :last_login_ip, login_count = :login_count,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	var users []*domain.User
	query := `
		SELECT id, username, email, user_type, nickname, avatar_url, bio,
			   email_verified, totp_enabled, is_active, is_banned,
			   last_login_at, login_count, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// ListActive 获取活跃用户列表
func (r *userRepository) ListActive(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	query := `
		SELECT id, username, email, user_type, nickname, avatar_url, bio,
			   email_verified, totp_enabled, is_active, is_banned,
			   last_login_at, login_count, created_at, updated_at
		FROM users
		WHERE is_active = true AND is_banned = false
		ORDER BY last_login_at DESC`

	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active users: %w", err)
	}

	return users, nil
}

// Count 获取用户总数
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	err := r.db.GetContext(ctx, &exists, query, username)
	if err != nil {
		return false, fmt.Errorf("failed to check username exists: %w", err)
	}

	return exists, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email exists: %w", err)
	}

	return exists, nil
}

// CreateSession 创建用户会话
func (r *userRepository) CreateSession(ctx context.Context, session *domain.UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, refresh_token, device_info, ip_address, user_agent, expires_at)
		VALUES (:id, :user_id, :refresh_token, :device_info, :ip_address, :user_agent, :expires_at)`

	_, err := r.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByRefreshToken 根据刷新令牌获取会话
func (r *userRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*domain.UserSession, error) {
	var session domain.UserSession
	query := `
		SELECT id, user_id, refresh_token, device_info, ip_address, user_agent, expires_at, created_at
		FROM user_sessions WHERE refresh_token = $1`

	err := r.db.GetContext(ctx, &session, query, refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}

	return &session, nil
}

// DeleteSession 删除会话
func (r *userRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM user_sessions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteUserSessions 删除用户的所有会话
func (r *userRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// CleanExpiredSessions 清理过期会话
func (r *userRepository) CleanExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM user_sessions WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to clean expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// 可以记录日志
	}

	return nil
}

// GrantPermission 授予权限
func (r *userRepository) GrantPermission(ctx context.Context, permission *domain.UserPermission) error {
	query := `
		INSERT INTO user_permissions (id, user_id, permission, granted_by, granted_at, expires_at)
		VALUES (:id, :user_id, :permission, :granted_by, :granted_at, :expires_at)
		ON CONFLICT (user_id, permission) DO UPDATE SET
			granted_by = EXCLUDED.granted_by,
			granted_at = EXCLUDED.granted_at,
			expires_at = EXCLUDED.expires_at`

	_, err := r.db.NamedExecContext(ctx, query, permission)
	if err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	return nil
}

// RevokePermission 撤销权限
func (r *userRepository) RevokePermission(ctx context.Context, userID uuid.UUID, permission string) error {
	query := `DELETE FROM user_permissions WHERE user_id = $1 AND permission = $2`

	_, err := r.db.ExecContext(ctx, query, userID, permission)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	return nil
}

// GetUserPermissions 获取用户权限列表
func (r *userRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.UserPermission, error) {
	var permissions []*domain.UserPermission
	query := `
		SELECT id, user_id, permission, granted_by, granted_at, expires_at
		FROM user_permissions
		WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	err := r.db.SelectContext(ctx, &permissions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// HasPermission 检查用户是否有指定权限
func (r *userRepository) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_permissions
			WHERE user_id = $1 AND permission = $2
			AND (expires_at IS NULL OR expires_at > NOW())
		)`

	err := r.db.GetContext(ctx, &exists, query, userID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	return exists, nil
}
