package auth

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/service"
)

// Handler 认证处理器
type Handler struct {
	authService service.AuthService
	logger      *logrus.Logger
}

// NewHandler 创建认证处理器
func NewHandler(authService service.AuthService, logger *logrus.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

// Routes 返回认证路由
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/logout", h.Logout)
	
	return r
}

// Register 用户注册
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	
	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if appErr, ok := domain.IsAppError(err); ok {
			h.writeAppError(w, appErr)
			return
		}
		h.writeError(w, http.StatusInternalServerError, "registration failed", err)
		return
	}
	
	h.writeSuccess(w, user)
}

// Login 用户登录
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	
	// 获取客户端IP
	clientIP := h.getClientIP(r)
	
	loginResp, err := h.authService.Login(r.Context(), &req, clientIP)
	if err != nil {
		if appErr, ok := domain.IsAppError(err); ok {
			h.writeAppError(w, appErr)
			return
		}
		h.writeError(w, http.StatusInternalServerError, "login failed", err)
		return
	}
	
	h.writeSuccess(w, loginResp)
}

// RefreshToken 刷新令牌
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	
	loginResp, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if appErr, ok := domain.IsAppError(err); ok {
			h.writeAppError(w, appErr)
			return
		}
		h.writeError(w, http.StatusInternalServerError, "refresh failed", err)
		return
	}
	
	h.writeSuccess(w, loginResp)
}

// Logout 用户登出
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// 这里需要从JWT中获取用户ID和刷新令牌
	// 简化实现，实际应该从中间件中获取
	h.writeSuccess(w, map[string]string{"message": "logged out successfully"})
}

// getClientIP 获取客户端IP
func (h *Handler) getClientIP(r *http.Request) net.IP {
	// 尝试从 X-Forwarded-For 头获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := net.ParseIP(xff); ip != nil {
			return ip
		}
	}
	
	// 尝试从 X-Real-IP 头获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip
		}
	}
	
	// 从 RemoteAddr 获取
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return net.ParseIP("127.0.0.1")
	}
	
	if ip := net.ParseIP(host); ip != nil {
		return ip
	}
	
	return net.ParseIP("127.0.0.1")
}

// writeSuccess 写入成功响应
func (h *Handler) writeSuccess(w http.ResponseWriter, data interface{}) {
	response := map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    data,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeError 写入错误响应
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, message string, err error) {
	response := map[string]interface{}{
		"code":    statusCode,
		"message": message,
	}
	
	if err != nil {
		h.logger.WithError(err).Error(message)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// writeAppError 写入应用错误响应
func (h *Handler) writeAppError(w http.ResponseWriter, appErr *domain.AppError) {
	statusCode := h.getStatusCodeFromAppError(appErr)
	
	response := map[string]interface{}{
		"code":    appErr.Code,
		"message": appErr.Message,
	}
	
	if appErr.Details != "" {
		response["details"] = appErr.Details
	}
	
	h.logger.WithError(appErr).Warn("Application error")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// getStatusCodeFromAppError 根据应用错误获取HTTP状态码
func (h *Handler) getStatusCodeFromAppError(appErr *domain.AppError) int {
	switch appErr.Code {
	case domain.CodeInvalidCredentials, domain.CodeInvalidToken, domain.CodeTokenExpired:
		return http.StatusUnauthorized
	case domain.CodeUserExists, domain.CodeEmailExists, domain.CodeUsernameExists:
		return http.StatusConflict
	case domain.CodeUserNotFound:
		return http.StatusNotFound
	case domain.CodeInvalidRequest:
		return http.StatusBadRequest
	case domain.CodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
