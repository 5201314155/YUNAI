package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// AuthRepository 认证仓库接口
type AuthRepository interface {
	// 登录尝试限制
	IncrementLoginAttempts(ctx context.Context, identifier string) (int, error)
	GetLoginAttempts(ctx context.Context, identifier string) (int, error)
	ResetLoginAttempts(ctx context.Context, identifier string) error
	
	// 黑名单管理
	AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	
	// 验证码缓存
	SetVerificationCode(ctx context.Context, key, code string, expiration time.Duration) error
	GetVerificationCode(ctx context.Context, key string) (string, error)
	DeleteVerificationCode(ctx context.Context, key string) error
	
	// 会话缓存
	SetUserSession(ctx context.Context, userID, sessionData string, expiration time.Duration) error
	GetUserSession(ctx context.Context, userID string) (string, error)
	DeleteUserSession(ctx context.Context, userID string) error
}

// authRepository 认证仓库实现
type authRepository struct {
	rdb *redis.Client
}

// NewAuthRepository 创建认证仓库
func NewAuthRepository(rdb *redis.Client) AuthRepository {
	return &authRepository{rdb: rdb}
}

// IncrementLoginAttempts 增加登录尝试次数
func (r *authRepository) IncrementLoginAttempts(ctx context.Context, identifier string) (int, error) {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	
	// 使用管道操作
	pipe := r.rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 15*time.Minute) // 15分钟过期
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment login attempts: %w", err)
	}
	
	return int(incrCmd.Val()), nil
}

// GetLoginAttempts 获取登录尝试次数
func (r *authRepository) GetLoginAttempts(ctx context.Context, identifier string) (int, error) {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	
	val, err := r.rdb.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // 键不存在，返回0
		}
		return 0, fmt.Errorf("failed to get login attempts: %w", err)
	}
	
	return val, nil
}

// ResetLoginAttempts 重置登录尝试次数
func (r *authRepository) ResetLoginAttempts(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	
	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to reset login attempts: %w", err)
	}
	
	return nil
}

// AddToBlacklist 添加令牌到黑名单
func (r *authRepository) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	
	err := r.rdb.Set(ctx, key, "1", expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}
	
	return nil
}

// IsBlacklisted 检查令牌是否在黑名单中
func (r *authRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	
	return exists > 0, nil
}

// SetVerificationCode 设置验证码
func (r *authRepository) SetVerificationCode(ctx context.Context, key, code string, expiration time.Duration) error {
	redisKey := fmt.Sprintf("verification:%s", key)
	
	err := r.rdb.Set(ctx, redisKey, code, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set verification code: %w", err)
	}
	
	return nil
}

// GetVerificationCode 获取验证码
func (r *authRepository) GetVerificationCode(ctx context.Context, key string) (string, error) {
	redisKey := fmt.Sprintf("verification:%s", key)
	
	code, err := r.rdb.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("verification code not found or expired")
		}
		return "", fmt.Errorf("failed to get verification code: %w", err)
	}
	
	return code, nil
}

// DeleteVerificationCode 删除验证码
func (r *authRepository) DeleteVerificationCode(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("verification:%s", key)
	
	err := r.rdb.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete verification code: %w", err)
	}
	
	return nil
}

// SetUserSession 设置用户会话缓存
func (r *authRepository) SetUserSession(ctx context.Context, userID, sessionData string, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)
	
	err := r.rdb.Set(ctx, key, sessionData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set user session: %w", err)
	}
	
	return nil
}

// GetUserSession 获取用户会话缓存
func (r *authRepository) GetUserSession(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("session:%s", userID)
	
	sessionData, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("user session not found")
		}
		return "", fmt.Errorf("failed to get user session: %w", err)
	}
	
	return sessionData, nil
}

// DeleteUserSession 删除用户会话缓存
func (r *authRepository) DeleteUserSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	
	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user session: %w", err)
	}
	
	return nil
}

// BatchDeleteUserSessions 批量删除用户会话（支持模式匹配）
func (r *authRepository) BatchDeleteUserSessions(ctx context.Context, pattern string) error {
	// 使用 SCAN 命令查找匹配的键
	iter := r.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}
	
	// 批量删除
	if len(keys) > 0 {
		err := r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete sessions: %w", err)
		}
	}
	
	return nil
}

// GetStats 获取认证相关统计信息
func (r *authRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 获取登录尝试相关统计
	loginAttemptsPattern := "login_attempts:*"
	loginAttemptKeys := r.rdb.Keys(ctx, loginAttemptsPattern).Val()
	stats["active_login_attempts"] = len(loginAttemptKeys)
	
	// 获取黑名单相关统计
	blacklistPattern := "blacklist:*"
	blacklistKeys := r.rdb.Keys(ctx, blacklistPattern).Val()
	stats["blacklisted_tokens"] = len(blacklistKeys)
	
	// 获取验证码相关统计
	verificationPattern := "verification:*"
	verificationKeys := r.rdb.Keys(ctx, verificationPattern).Val()
	stats["active_verification_codes"] = len(verificationKeys)
	
	// 获取会话相关统计
	sessionPattern := "session:*"
	sessionKeys := r.rdb.Keys(ctx, sessionPattern).Val()
	stats["active_sessions"] = len(sessionKeys)
	
	return stats, nil
}

// CleanupExpiredData 清理过期数据（定期任务）
func (r *authRepository) CleanupExpiredData(ctx context.Context) error {
	// Redis 会自动清理过期键，这里可以添加额外的清理逻辑
	// 比如清理一些没有设置过期时间但应该过期的数据
	
	// 示例：清理超过24小时的登录尝试记录
	pattern := "login_attempts:*"
	iter := r.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	
	for iter.Next(ctx) {
		key := iter.Val()
		ttl := r.rdb.TTL(ctx, key).Val()
		
		// 如果TTL为-1（永不过期）且键存在超过24小时，则删除
		if ttl == -1 {
			r.rdb.Expire(ctx, key, 24*time.Hour)
		}
	}
	
	return iter.Err()
}
