package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims JWT 声明
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	UserType string    `json:"user_type"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewManager 创建 JWT 管理器
func NewManager(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:            secretKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, username, userType string) (accessToken, refreshToken string, err error) {
	// 生成访问令牌
	accessToken, err = m.GenerateAccessToken(userID, username, userType)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err = m.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken 生成访问令牌
func (m *JWTManager) GenerateAccessToken(userID uuid.UUID, username, userType string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID.String(),
			Audience:  []string{"yunai"},
			Issuer:    "yunai-auth",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成刷新令牌（随机字符串）
func (m *JWTManager) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateAccessToken 验证访问令牌
func (m *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ExtractTokenFromHeader 从 Authorization 头中提取令牌
func ExtractTokenFromHeader(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return "", fmt.Errorf("invalid authorization header format")
	}

	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetAccessTokenDuration 获取访问令牌有效期
func (m *JWTManager) GetAccessTokenDuration() time.Duration {
	return m.accessTokenDuration
}

// GetRefreshTokenDuration 获取刷新令牌有效期
func (m *JWTManager) GetRefreshTokenDuration() time.Duration {
	return m.refreshTokenDuration
}

// TokenInfo 令牌信息
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// CreateTokenInfo 创建令牌信息
func (m *JWTManager) CreateTokenInfo(accessToken, refreshToken string) *TokenInfo {
	now := time.Now()
	return &TokenInfo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.accessTokenDuration.Seconds()),
		ExpiresAt:    now.Add(m.accessTokenDuration),
	}
}

// IsTokenExpired 检查令牌是否过期
func IsTokenExpired(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为过期错误 (JWT v5)
	return err == jwt.ErrTokenExpired
}

// GetTokenRemainingTime 获取令牌剩余时间
func GetTokenRemainingTime(claims *JWTClaims) time.Duration {
	if claims.ExpiresAt == nil {
		return 0
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0
	}

	return remaining
}

// ShouldRefreshToken 判断是否应该刷新令牌
func ShouldRefreshToken(claims *JWTClaims, threshold time.Duration) bool {
	remaining := GetTokenRemainingTime(claims)
	return remaining <= threshold
}

// GenerateAPIKey 生成 API 密钥
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return "yunai_" + base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateAPIKey 验证 API 密钥格式
func ValidateAPIKey(apiKey string) bool {
	const prefix = "yunai_"
	if len(apiKey) < len(prefix) {
		return false
	}

	if apiKey[:len(prefix)] != prefix {
		return false
	}

	// 验证 base64 编码部分
	encoded := apiKey[len(prefix):]
	if _, err := base64.URLEncoding.DecodeString(encoded); err != nil {
		return false
	}

	return true
}
