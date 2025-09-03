package middleware

import (
	"context"
	"net/http"
	"strings"

	"yunai/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	// 可以添加JWT验证、Redis会话等
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// RequireAuth Gin认证中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单的测试认证 - 检查Authorization头或者直接通过
		authHeader := c.GetHeader("Authorization")

		// 如果没有Authorization头，创建一个测试用户
		if authHeader == "" {
			// 为测试目的，创建一个默认用户
			testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
			nickname := "测试用户"
			c.Set("user_id", testUserID)
			c.Set("user", &domain.User{
				ID:       testUserID,
				Username: "test_user",
				Email:    "test@yunai.com",
				Nickname: &nickname,
			})
			c.Next()
			return
		}

		// 简单的Bearer token验证
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// 简单验证 - 在实际项目中应该验证JWT
			if len(token) > 0 {
				testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				nickname := "测试用户"
				c.Set("user_id", testUserID)
				c.Set("user", &domain.User{
					ID:       testUserID,
					Username: "test_user",
					Email:    "test@yunai.com",
					Nickname: &nickname,
				})
				c.Next()
				return
			}
		}

		// 认证失败
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Code:    domain.CodeUnauthorized,
			Message: "User not authenticated",
		})
		c.Abort()
	}
}

// OptionalAuth 可选认证中间件
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if len(token) > 0 {
				testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				nickname := "测试用户"
				c.Set("user_id", testUserID)
				c.Set("user", &domain.User{
					ID:       testUserID,
					Username: "test_user",
					Email:    "test@yunai.com",
					Nickname: &nickname,
				})
			}
		}

		c.Next()
	}
}

// Chi认证中间件
func (m *AuthMiddleware) ChiRequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// 如果没有Authorization头，创建一个测试用户
		if authHeader == "" {
			testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
			nickname := "测试用户"
			ctx := context.WithValue(r.Context(), "user_id", testUserID)
			ctx = context.WithValue(ctx, "user", &domain.User{
				ID:       testUserID,
				Username: "test_user",
				Email:    "test@yunai.com",
				Nickname: &nickname,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 简单的Bearer token验证
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			if len(token) > 0 {
				testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				nickname := "测试用户"
				ctx := context.WithValue(r.Context(), "user_id", testUserID)
				ctx = context.WithValue(ctx, "user", &domain.User{
					ID:       testUserID,
					Username: "test_user",
					Email:    "test@yunai.com",
					Nickname: &nickname,
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// 认证失败
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":401,"message":"User not authenticated"}`))
	})
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(c *gin.Context) (*domain.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	domainUser, ok := user.(*domain.User)
	return domainUser, ok
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	id, ok := userID.(uuid.UUID)
	return id, ok
}
