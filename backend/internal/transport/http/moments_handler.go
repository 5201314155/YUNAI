package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/service"
)

// MomentsHandler 朋友圈HTTP处理器
type MomentsHandler struct {
	momentsService service.MomentsService
	logger         *logrus.Logger
}

// NewMomentsHandler 创建朋友圈处理器
func NewMomentsHandler(momentsService service.MomentsService, logger *logrus.Logger) *MomentsHandler {
	return &MomentsHandler{
		momentsService: momentsService,
		logger:         logger,
	}
}

// RegisterRoutes 注册路由
func (h *MomentsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/moments", func(r chi.Router) {
		// 朋友圈动态管理
		r.Post("/", h.CreateMoment)
		r.Get("/", h.ListMoments)
		r.Get("/{id}", h.GetMoment)
		r.Put("/{id}", h.UpdateMoment)
		r.Delete("/{id}", h.DeleteMoment)

		// 朋友圈草稿管理
		r.Post("/drafts/generate", h.GenerateMomentDrafts)
		r.Get("/drafts", h.GetMomentDrafts)
		r.Post("/drafts/{id}/publish", h.PublishMomentDraft)
		r.Delete("/drafts/{id}", h.DeleteMomentDraft)

		// 朋友圈互动
		r.Post("/{id}/like", h.LikeMoment)
		r.Delete("/{id}/like", h.UnlikeMoment)
		r.Post("/{id}/comment", h.CommentMoment)
		r.Post("/{id}/share", h.ShareMoment)

		// 自动生成配置
		r.Get("/config/{characterId}", h.GetAutoGenerationConfig)
		r.Put("/config/{characterId}", h.UpdateAutoGenerationConfig)

		// 自动生成任务
		r.Post("/auto-generate/{characterId}", h.AutoGenerateMoments)

		// 统计分析
		r.Get("/analytics/{characterId}", h.GetMomentAnalytics)

		// 通知管理
		r.Get("/notifications", h.GetUnreadNotifications)
		r.Put("/notifications/{id}/read", h.MarkNotificationAsRead)
	})
}

// CreateMoment 创建朋友圈动态
func (h *MomentsHandler) CreateMoment(w http.ResponseWriter, r *http.Request) {
	var moment domain.Moment
	if err := json.NewDecoder(r.Body).Decode(&moment); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 从上下文获取用户ID
	userID := getUserIDFromContext(r.Context())
	moment.UserID = userID

	err := h.momentsService.CreateMoment(r.Context(), &moment)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create moment", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"moment":  moment,
	})
}

// GetMoment 获取朋友圈动态
func (h *MomentsHandler) GetMoment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	response, err := h.momentsService.GetMoment(r.Context(), id, userID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Moment not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ListMoments 获取朋友圈动态列表
func (h *MomentsHandler) ListMoments(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	req := &domain.MomentListRequest{
		UserID: userID,
		Page:   1,
		Limit:  20,
	}

	// 解析查询参数
	if characterIDStr := r.URL.Query().Get("character_id"); characterIDStr != "" {
		characterID, err := uuid.Parse(characterIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
			return
		}
		req.CharacterID = &characterID
	}

	if visibility := r.URL.Query().Get("visibility"); visibility != "" {
		req.Visibility = &visibility
	}

	if contentType := r.URL.Query().Get("content_type"); contentType != "" {
		req.ContentType = &contentType
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			h.writeError(w, http.StatusBadRequest, "Invalid page number", err)
			return
		}
		req.Page = page
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 50 {
			h.writeError(w, http.StatusBadRequest, "Invalid limit", err)
			return
		}
		req.Limit = limit
	}

	moments, total, err := h.momentsService.ListMoments(r.Context(), req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list moments", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"moments": moments,
		"total":   total,
		"page":    req.Page,
		"limit":   req.Limit,
	})
}

// UpdateMoment 更新朋友圈动态
func (h *MomentsHandler) UpdateMoment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	var moment domain.Moment
	if err := json.NewDecoder(r.Body).Decode(&moment); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	moment.ID = id
	moment.UserID = getUserIDFromContext(r.Context())

	err = h.momentsService.UpdateMoment(r.Context(), &moment)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update moment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"moment":  moment,
	})
}

// DeleteMoment 删除朋友圈动态
func (h *MomentsHandler) DeleteMoment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	err = h.momentsService.DeleteMoment(r.Context(), id, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete moment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// GenerateMomentDrafts 生成朋友圈草稿
func (h *MomentsHandler) GenerateMomentDrafts(w http.ResponseWriter, r *http.Request) {
	var req domain.MomentGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	req.UserID = getUserIDFromContext(r.Context())

	result, err := h.momentsService.GenerateMomentDrafts(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate drafts", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetMomentDrafts 获取朋友圈草稿
func (h *MomentsHandler) GetMomentDrafts(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	characterIDStr := r.URL.Query().Get("character_id")
	if characterIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "Missing character_id parameter", nil)
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
		return
	}

	drafts, err := h.momentsService.GetMomentDrafts(r.Context(), characterID, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get drafts", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"drafts": drafts,
	})
}

// PublishMomentDraft 发布朋友圈草稿
func (h *MomentsHandler) PublishMomentDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid draft ID", err)
		return
	}

	var req domain.MomentPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	req.DraftID = draftID
	req.UserID = getUserIDFromContext(r.Context())

	result, err := h.momentsService.PublishMomentDraft(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to publish draft", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// DeleteMomentDraft 删除朋友圈草稿
func (h *MomentsHandler) DeleteMomentDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid draft ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	err = h.momentsService.DeleteMomentDraft(r.Context(), draftID, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete draft", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// LikeMoment 点赞朋友圈
func (h *MomentsHandler) LikeMoment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	momentID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	// 可选的角色ID
	var characterID *uuid.UUID
	if characterIDStr := r.URL.Query().Get("character_id"); characterIDStr != "" {
		cID, err := uuid.Parse(characterIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
			return
		}
		characterID = &cID
	}

	err = h.momentsService.LikeMoment(r.Context(), momentID, userID, characterID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to like moment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// UnlikeMoment 取消点赞朋友圈
func (h *MomentsHandler) UnlikeMoment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	momentID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	// 可选的角色ID
	var characterID *uuid.UUID
	if characterIDStr := r.URL.Query().Get("character_id"); characterIDStr != "" {
		cID, err := uuid.Parse(characterIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
			return
		}
		characterID = &cID
	}

	err = h.momentsService.UnlikeMoment(r.Context(), momentID, userID, characterID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to unlike moment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// CommentMoment 评论朋友圈
func (h *MomentsHandler) CommentMoment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	momentID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	var req struct {
		Content     string     `json:"content"`
		CharacterID *uuid.UUID `json:"character_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Content == "" {
		h.writeError(w, http.StatusBadRequest, "Comment content is required", nil)
		return
	}

	userID := getUserIDFromContext(r.Context())

	err = h.momentsService.CommentMoment(r.Context(), momentID, userID, req.CharacterID, req.Content)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to comment moment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// ShareMoment 分享朋友圈
func (h *MomentsHandler) ShareMoment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	momentID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	// 可选的角色ID
	var characterID *uuid.UUID
	if characterIDStr := r.URL.Query().Get("character_id"); characterIDStr != "" {
		cID, err := uuid.Parse(characterIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
			return
		}
		characterID = &cID
	}

	err = h.momentsService.ShareMoment(r.Context(), momentID, userID, characterID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to share moment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// GetAutoGenerationConfig 获取自动生成配置
func (h *MomentsHandler) GetAutoGenerationConfig(w http.ResponseWriter, r *http.Request) {
	characterIDStr := chi.URLParam(r, "characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
		return
	}

	config, err := h.momentsService.GetAutoGenerationConfig(r.Context(), characterID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get config", err)
		return
	}

	h.writeJSON(w, http.StatusOK, config)
}

// UpdateAutoGenerationConfig 更新自动生成配置
func (h *MomentsHandler) UpdateAutoGenerationConfig(w http.ResponseWriter, r *http.Request) {
	characterIDStr := chi.URLParam(r, "characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
		return
	}

	var config domain.MomentAutoGenerationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	config.CharacterID = characterID
	config.UserID = getUserIDFromContext(r.Context())

	err = h.momentsService.UpdateAutoGenerationConfig(r.Context(), &config)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update config", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"config":  config,
	})
}

// AutoGenerateMoments 自动生成朋友圈
func (h *MomentsHandler) AutoGenerateMoments(w http.ResponseWriter, r *http.Request) {
	characterIDStr := chi.URLParam(r, "characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
		return
	}

	result, err := h.momentsService.AutoGenerateMoments(r.Context(), characterID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to auto generate moments", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetMomentAnalytics 获取朋友圈分析数据
func (h *MomentsHandler) GetMomentAnalytics(w http.ResponseWriter, r *http.Request) {
	characterIDStr := chi.URLParam(r, "characterId")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid character ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	analytics, err := h.momentsService.GetMomentAnalytics(r.Context(), characterID, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get analytics", err)
		return
	}

	h.writeJSON(w, http.StatusOK, analytics)
}

// GetUnreadNotifications 获取未读通知
func (h *MomentsHandler) GetUnreadNotifications(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	notifications, err := h.momentsService.GetUnreadNotifications(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get notifications", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
	})
}

// MarkNotificationAsRead 标记通知为已读
func (h *MomentsHandler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid notification ID", err)
		return
	}

	userID := getUserIDFromContext(r.Context())

	err = h.momentsService.MarkNotificationAsRead(r.Context(), notificationID, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to mark notification as read", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// 辅助方法

// writeJSON 写入JSON响应
func (h *MomentsHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeError 写入错误响应
func (h *MomentsHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		h.logger.WithError(err).WithField("message", message).Error("HTTP handler error")
	}

	h.writeJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(ctx context.Context) uuid.UUID {
	// 这里应该从JWT token或session中获取用户ID
	// 暂时返回一个固定的UUID用于测试
	userID, _ := uuid.Parse("daa19f83-c430-41d1-9aba-cc73176b5582")
	return userID
}
