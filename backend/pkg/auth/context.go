package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserIDFromContext 从 Gin 上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) uuid.UUID {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// GetUserFromContext 从 Gin 上下文中获取用户信息
func GetUserFromContext(c *gin.Context) *JWTClaims {
	if user, exists := c.Get("user"); exists {
		if claims, ok := user.(*JWTClaims); ok {
			return claims
		}
	}
	return nil
}

// SetUserInContext 在 Gin 上下文中设置用户信息
func SetUserInContext(c *gin.Context, claims *JWTClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("user", claims)
	c.Set("username", claims.Username)
	c.Set("user_type", claims.UserType)
}
