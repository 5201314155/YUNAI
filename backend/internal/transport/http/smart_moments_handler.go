package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/service"
	"yunai/pkg/response"
)

// SmartMomentsHandler 智能朋友圈处理器
type SmartMomentsHandler struct {
	smartMomentsService service.SmartMomentsService
	logger              *logrus.Logger
}

// NewSmartMomentsHandler 创建智能朋友圈处理器
func NewSmartMomentsHandler(smartMomentsService service.SmartMomentsService, logger *logrus.Logger) *SmartMomentsHandler {
	return &SmartMomentsHandler{
		smartMomentsService: smartMomentsService,
		logger:              logger,
	}
}

// RegisterRoutes 注册路由
func (h *SmartMomentsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/smart-moments", func(r chi.Router) {
		r.Post("/generate", h.SmartGenerateMoments)
		r.Post("/{momentID}/auto-interact", h.TriggerAutoInteractions)
		r.Post("/{momentID}/process-mentions", h.ProcessMentions)
		r.Get("/generation-context/{characterID}", h.GetGenerationContext)
	})
}

// SmartGenerateMomentsRequest 智能生成朋友圈请求
type SmartGenerateMomentsRequest struct {
	CharacterID  uuid.UUID `json:"character_id" validate:"required"`
	Count        int       `json:"count" validate:"min=1,max=10"`
	ContentTypes []string  `json:"content_types,omitempty"`
	Mood         *string   `json:"mood,omitempty"`
	Context      *string   `json:"context,omitempty"`
	AutoInteract bool      `json:"auto_interact"`
	BasedOnChat  bool      `json:"based_on_chat"`
}

// SmartGenerateMomentsResponse 智能生成朋友圈响应
type SmartGenerateMomentsResponse struct {
	Success      bool                    `json:"success"`
	Message      string                  `json:"message"`
	TotalCount   int                     `json:"total_count"`
	Drafts       []*domain.MomentDraft   `json:"drafts"`
	GeneratedAt  string                  `json:"generated_at"`
	SmartFeatures SmartFeaturesInfo      `json:"smart_features"`
}

// SmartFeaturesInfo 智能功能信息
type SmartFeaturesInfo struct {
	ModelUsed        string   `json:"model_used"`
	ChatBased        bool     `json:"chat_based"`
	MentionsDetected []string `json:"mentions_detected"`
	AutoInteractEnabled bool  `json:"auto_interact_enabled"`
	RelationshipCount int    `json:"relationship_count"`
}

// TriggerAutoInteractionsRequest 触发自动互动请求
type TriggerAutoInteractionsRequest struct {
	Force bool `json:"force"` // 是否强制触发
}

// ProcessMentionsRequest 处理@提及请求
type ProcessMentionsRequest struct {
	Content string `json:"content" validate:"required"`
}

// GetGenerationContextResponse 获取生成上下文响应
type GetGenerationContextResponse struct {
	Character         *domain.CharacterResponse    `json:"character"`
	SelectedModel     *domain.AIModel              `json:"selected_model"`
	AvailableModels   []*domain.AIModel            `json:"available_models"`
	RelationshipCount int                          `json:"relationship_count"`
	RecentChatsCount  int                          `json:"recent_chats_count"`
	MentionCandidates []*MentionCandidate          `json:"mention_candidates"`
}

// MentionCandidate @提及候选
type MentionCandidate struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Relationship string    `json:"relationship"`
	Strength     int       `json:"strength"`
}

// SmartGenerateMoments 智能生成朋友圈
func (h *SmartMomentsHandler) SmartGenerateMoments(w http.ResponseWriter, r *http.Request) {
	var req SmartGenerateMomentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 获取用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 3
	}

	// 构建服务请求
	serviceReq := &domain.MomentGenerationRequest{
		CharacterID:  req.CharacterID,
		UserID:       userID,
		Count:        req.Count,
		ContentTypes: req.ContentTypes,
		Mood:         req.Mood,
		Context:      req.Context,
		AutoInteract: req.AutoInteract,
		BasedOnChat:  req.BasedOnChat,
	}

	// 调用智能生成服务
	result, err := h.smartMomentsService.SmartGenerateMoments(r.Context(), serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to smart generate moments")
		response.Error(w, http.StatusInternalServerError, "Failed to generate moments", err)
		return
	}

	// 构建响应
	resp := &SmartGenerateMomentsResponse{
		Success:     result.Success,
		Message:     result.Message,
		TotalCount:  result.TotalCount,
		Drafts:      result.Drafts,
		GeneratedAt: result.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"),
		SmartFeatures: SmartFeaturesInfo{
			ModelUsed:           "deepseek测试专用",
			ChatBased:           req.BasedOnChat,
			MentionsDetected:    []string{}, // TODO: 从结果中提取
			AutoInteractEnabled: req.AutoInteract,
			RelationshipCount:   0, // TODO: 从上下文中获取
		},
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"character_id":  req.CharacterID,
		"generated_count": len(result.Drafts),
		"chat_based":    req.BasedOnChat,
		"auto_interact": req.AutoInteract,
	}).Info("Smart moments generated successfully")

	response.Success(w, resp)
}

// TriggerAutoInteractions 触发自动互动
func (h *SmartMomentsHandler) TriggerAutoInteractions(w http.ResponseWriter, r *http.Request) {
	momentIDStr := chi.URLParam(r, "momentID")
	momentID, err := uuid.Parse(momentIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	var req TriggerAutoInteractionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 获取用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 触发自动互动
	err = h.smartMomentsService.TriggerAutoInteractions(r.Context(), momentID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to trigger auto interactions")
		response.Error(w, http.StatusInternalServerError, "Failed to trigger auto interactions", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"moment_id": momentID,
		"force":     req.Force,
	}).Info("Auto interactions triggered")

	response.Success(w, map[string]interface{}{
		"message": "Auto interactions triggered successfully",
		"moment_id": momentID,
	})
}

// ProcessMentions 处理@提及
func (h *SmartMomentsHandler) ProcessMentions(w http.ResponseWriter, r *http.Request) {
	momentIDStr := chi.URLParam(r, "momentID")
	momentID, err := uuid.Parse(momentIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid moment ID", err)
		return
	}

	var req ProcessMentionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 获取用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 处理@提及
	err = h.smartMomentsService.ProcessMentions(r.Context(), momentID, req.Content)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process mentions")
		response.Error(w, http.StatusInternalServerError, "Failed to process mentions", err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"moment_id": momentID,
		"content":   req.Content,
	}).Info("Mentions processed")

	response.Success(w, map[string]interface{}{
		"message": "Mentions processed successfully",
		"moment_id": momentID,
	})
}

// GetGenerationContext 获取生成上下文
func (h *SmartMomentsHandler) GetGenerationContext(w http.ResponseWriter, r *http.Request) {
	characterIDStr := chi.URLParam(r, "characterID")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid character ID", err)
		return
	}

	// 获取用户ID
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 获取查询参数
	basedOnChat := r.URL.Query().Get("based_on_chat") == "true"

	// 构建智能生成上下文
	context, err := h.smartMomentsService.BuildSmartGenerationContext(r.Context(), characterID, userID, basedOnChat)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build generation context")
		response.Error(w, http.StatusInternalServerError, "Failed to build generation context", err)
		return
	}

	// 构建@提及候选列表
	var mentionCandidates []*MentionCandidate
	for _, candidate := range context.MentionCandidates {
		mentionCandidates = append(mentionCandidates, &MentionCandidate{
			ID:           candidate.Character.ID,
			Name:         candidate.Character.Name,
			Relationship: func() string {
				if candidate.Relationship.CustomTypeName != nil {
					return *candidate.Relationship.CustomTypeName
				}
				return "朋友"
			}(),
			Strength: candidate.Relationship.Strength,
		})
	}

	// 构建响应
	resp := &GetGenerationContextResponse{
		Character:         context.Character,
		SelectedModel:     context.SelectedModel,
		AvailableModels:   context.AvailableModels,
		RelationshipCount: len(context.MentionCandidates),
		RecentChatsCount:  len(context.RecentChats),
		MentionCandidates: mentionCandidates,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":            userID,
		"character_id":       characterID,
		"based_on_chat":      basedOnChat,
		"relationship_count": len(context.MentionCandidates),
		"recent_chats_count": len(context.RecentChats),
	}).Info("Generation context retrieved")

	response.Success(w, resp)
}
