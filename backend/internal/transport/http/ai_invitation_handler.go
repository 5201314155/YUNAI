package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"yunai/internal/domain"
	"yunai/internal/service"
	"yunai/pkg/response"
)

// AIInvitationHandler AI智能邀请处理器
type AIInvitationHandler struct {
	aiInvitationService service.AIInvitationService
}

// NewAIInvitationHandler 创建AI智能邀请处理器
func NewAIInvitationHandler(aiInvitationService service.AIInvitationService) *AIInvitationHandler {
	return &AIInvitationHandler{
		aiInvitationService: aiInvitationService,
	}
}

// RegisterRoutes 注册路由
func (h *AIInvitationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/ai-invitation", func(r chi.Router) {
		r.Post("/analyze", h.AnalyzeInvitationIntent)
		r.Post("/find-matches", h.FindMatchingCharacters)
		r.Post("/suggestions", h.GenerateInvitationSuggestions)
		r.Post("/execute", h.ExecuteSmartInvitation)
	})
}

// AnalyzeInvitationIntentRequest 分析邀请意图请求
type AnalyzeInvitationIntentRequest struct {
	UserMessage string    `json:"user_message" validate:"required"`
	GroupChatID uuid.UUID `json:"group_chat_id" validate:"required"`
}

// AnalyzeInvitationIntent 分析邀请意图
func (h *AIInvitationHandler) AnalyzeInvitationIntent(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeInvitationIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 获取用户ID（从认证中间件）- 测试环境下使用默认用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		// 测试环境下使用默认用户ID
		userID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	}

	analysis, err := h.aiInvitationService.AnalyzeInvitationIntent(r.Context(), req.UserMessage, req.GroupChatID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to analyze invitation intent")
		return
	}

	response.Success(w, analysis)
}

// FindMatchingCharactersRequest 查找匹配角色请求
type FindMatchingCharactersRequest struct {
	Intent *domain.InvitationAnalysis `json:"intent" validate:"required"`
}

// FindMatchingCharacters 查找匹配角色
func (h *AIInvitationHandler) FindMatchingCharacters(w http.ResponseWriter, r *http.Request) {
	var req FindMatchingCharactersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 获取用户ID - 测试环境下使用默认用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		// 测试环境下使用默认用户ID
		userID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	}

	candidates, err := h.aiInvitationService.FindMatchingCharacters(r.Context(), req.Intent, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to find matching characters")
		return
	}

	response.Success(w, map[string]interface{}{
		"candidates": candidates,
		"count":      len(candidates),
	})
}

// GenerateInvitationSuggestionsRequest 生成邀请建议请求
type GenerateInvitationSuggestionsRequest struct {
	Candidates  []*domain.InvitationCandidate `json:"candidates" validate:"required"`
	GroupChatID uuid.UUID                     `json:"group_chat_id" validate:"required"`
}

// GenerateInvitationSuggestions 生成邀请建议
func (h *AIInvitationHandler) GenerateInvitationSuggestions(w http.ResponseWriter, r *http.Request) {
	var req GenerateInvitationSuggestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	suggestions, err := h.aiInvitationService.GenerateInvitationSuggestions(r.Context(), req.Candidates, req.GroupChatID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate invitation suggestions")
		return
	}

	response.Success(w, map[string]interface{}{
		"suggestions": suggestions,
		"count":       len(suggestions),
	})
}

// ExecuteSmartInvitationRequest 执行智能邀请请求
type ExecuteSmartInvitationRequest struct {
	UserMessage string    `json:"user_message" validate:"required"`
	GroupChatID uuid.UUID `json:"group_chat_id" validate:"required"`
}

// ExecuteSmartInvitation 执行智能邀请
func (h *AIInvitationHandler) ExecuteSmartInvitation(w http.ResponseWriter, r *http.Request) {
	var req ExecuteSmartInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 获取用户ID - 测试环境下使用默认用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		// 测试环境下使用默认用户ID
		userID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	}

	result, err := h.aiInvitationService.ExecuteSmartInvitation(r.Context(), req.UserMessage, req.GroupChatID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to execute smart invitation")
		return
	}

	response.Success(w, result)
}

// 简化的邀请接口，用于快速测试

// QuickInviteRequest 快速邀请请求
type QuickInviteRequest struct {
	Message     string    `json:"message" validate:"required"`
	GroupChatID uuid.UUID `json:"group_chat_id" validate:"required"`
}

// QuickInviteResponse 快速邀请响应
type QuickInviteResponse struct {
	Success     bool                           `json:"success"`
	Message     string                         `json:"message"`
	Suggestions []*domain.InvitationSuggestion `json:"suggestions,omitempty"`
	Analysis    *domain.InvitationAnalysis     `json:"analysis,omitempty"`
}

// QuickInvite 快速邀请（一步完成所有操作）
func (h *AIInvitationHandler) QuickInvite(w http.ResponseWriter, r *http.Request) {
	var req QuickInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 获取用户ID - 测试环境下使用默认用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		// 测试环境下使用默认用户ID
		userID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	}

	// 执行智能邀请
	result, err := h.aiInvitationService.ExecuteSmartInvitation(r.Context(), req.Message, req.GroupChatID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to execute smart invitation")
		return
	}

	// 如果需要分析详情，也可以返回
	analysis, _ := h.aiInvitationService.AnalyzeInvitationIntent(r.Context(), req.Message, req.GroupChatID, userID) // 忽略错误

	quickResponse := &QuickInviteResponse{
		Success:     result.Success,
		Message:     result.Message,
		Suggestions: result.Suggestions,
		Analysis:    analysis,
	}

	response.Success(w, quickResponse)
}

// 添加快速邀请路由
func (h *AIInvitationHandler) RegisterQuickRoutes(r chi.Router) {
	r.Post("/quick-invite", h.QuickInvite)
}

// 获取邀请历史和统计

// GetInvitationStats 获取邀请统计
func (h *AIInvitationHandler) GetInvitationStats(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID - 测试环境下使用默认用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		// 测试环境下使用默认用户ID
		userID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	}

	// 获取查询参数
	days := 7 // 默认7天
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	// 这里可以实现邀请统计逻辑
	// 暂时返回模拟数据
	stats := map[string]interface{}{
		"user_id":                   userID,
		"days":                      days,
		"total_invitations":         0,
		"successful_invites":        0,
		"most_invited_characters":   []string{},
		"common_relationship_types": []string{},
	}

	response.Success(w, stats)
}

// RegisterStatsRoutes 注册统计路由
func (h *AIInvitationHandler) RegisterStatsRoutes(r chi.Router) {
	r.Get("/stats", h.GetInvitationStats)
}
