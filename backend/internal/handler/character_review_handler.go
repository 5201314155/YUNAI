package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/service"
)

// CharacterReviewHandler 角色审核处理器
type CharacterReviewHandler struct {
	reviewService service.CharacterReviewService
	logger        *logrus.Logger
}

// NewCharacterReviewHandler 创建角色审核处理器
func NewCharacterReviewHandler(
	reviewService service.CharacterReviewService,
	logger *logrus.Logger,
) *CharacterReviewHandler {
	return &CharacterReviewHandler{
		reviewService: reviewService,
		logger:        logger,
	}
}

// SubmitReview 提交角色审核
func (h *CharacterReviewHandler) SubmitReview(c *gin.Context) {
	var req domain.CharacterReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证角色ID
	if req.CharacterID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID不能为空"})
		return
	}

	// 验证审核类型
	validTypes := map[string]bool{
		"content": true,
		"avatar":  true,
		"name":    true,
		"all":     true,
	}
	if !validTypes[req.ReviewType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核类型"})
		return
	}

	// 提交审核
	response, err := h.reviewService.SubmitReview(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to submit character review")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交审核失败"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"character_id": req.CharacterID,
		"review_type":  req.ReviewType,
		"review_id":    response.ReviewID,
		"status":       response.Status,
	}).Info("Character review submitted")

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "审核请求已提交",
		"data":    response,
	})
}

// GetReview 获取审核记录
func (h *CharacterReviewHandler) GetReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核ID"})
		return
	}

	review, err := h.reviewService.GetReview(c.Request.Context(), reviewID)
	if err != nil {
		if err == domain.ErrCharacterReviewNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "审核记录不存在"})
			return
		}
		h.logger.WithError(err).Error("Failed to get character review")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审核记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    review,
	})
}

// ListReviews 列出审核记录
func (h *CharacterReviewHandler) ListReviews(c *gin.Context) {
	status := domain.CharacterReviewStatus(c.Query("status"))
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	reviews, err := h.reviewService.ListReviews(c.Request.Context(), status, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list character reviews")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审核列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"reviews": reviews,
			"limit":   limit,
			"offset":  offset,
			"total":   len(reviews),
		},
	})
}

// ApproveReview 通过审核
func (h *CharacterReviewHandler) ApproveReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取审核员ID（从JWT中获取）
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	uid, ok := reviewerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	err = h.reviewService.ApproveReview(c.Request.Context(), reviewID, uid, req.Reason)
	if err != nil {
		if err == domain.ErrCharacterReviewNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "审核记录不存在"})
			return
		}
		h.logger.WithError(err).Error("Failed to approve character review")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核通过失败"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"review_id":   reviewID,
		"reviewer_id": uid,
		"reason":      req.Reason,
	}).Info("Character review approved")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "审核已通过",
	})
}

// RejectReview 拒绝审核
func (h *CharacterReviewHandler) RejectReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误，拒绝原因不能为空"})
		return
	}

	// 获取审核员ID
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	uid, ok := reviewerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	err = h.reviewService.RejectReview(c.Request.Context(), reviewID, uid, req.Reason)
	if err != nil {
		if err == domain.ErrCharacterReviewNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "审核记录不存在"})
			return
		}
		h.logger.WithError(err).Error("Failed to reject character review")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核拒绝失败"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"review_id":   reviewID,
		"reviewer_id": uid,
		"reason":      req.Reason,
	}).Info("Character review rejected")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "审核已拒绝",
	})
}

// AutoReviewCharacter 自动审核角色
func (h *CharacterReviewHandler) AutoReviewCharacter(c *gin.Context) {
	characterIDStr := c.Param("character_id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	result, err := h.reviewService.AutoReviewCharacter(c.Request.Context(), characterID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to auto review character")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "自动审核失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "自动审核完成",
		"data":    result,
	})
}

// ReviewContent 审核文本内容
func (h *CharacterReviewHandler) ReviewContent(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	result, err := h.reviewService.ReviewContent(c.Request.Context(), req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "内容审核完成",
		"data":    result,
	})
}

// GetReviewStatistics 获取审核统计
func (h *CharacterReviewHandler) GetReviewStatistics(c *gin.Context) {
	stats, err := h.reviewService.GetReviewStatistics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get review statistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审核统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// CreateReviewRule 创建审核规则
func (h *CharacterReviewHandler) CreateReviewRule(c *gin.Context) {
	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Type        string                 `json:"type" binding:"required"`
		Category    string                 `json:"category" binding:"required"`
		Pattern     string                 `json:"pattern"`
		Action      string                 `json:"action" binding:"required"`
		Severity    int                    `json:"severity" binding:"required"`
		Description string                 `json:"description"`
		Config      map[string]interface{} `json:"config"`
		Enabled     bool                   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 转换Config为JSON
	var configJSON *json.RawMessage
	if req.Config != nil {
		configData, marshalErr := json.Marshal(req.Config)
		if marshalErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "配置格式错误"})
			return
		}
		rawMsg := json.RawMessage(configData)
		configJSON = &rawMsg
	}

	rule := &domain.ReviewRule{
		ID:          uuid.New(),
		Name:        req.Name,
		Type:        req.Type,
		Category:    req.Category,
		Pattern:     req.Pattern,
		Action:      req.Action,
		Severity:    req.Severity,
		Description: req.Description,
		Config:      configJSON,
		Enabled:     req.Enabled,
	}

	err := h.reviewService.CreateReviewRule(c.Request.Context(), rule)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create review rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建审核规则失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "审核规则创建成功",
		"data":    rule,
	})
}

// ListReviewRules 列出审核规则
func (h *CharacterReviewHandler) ListReviewRules(c *gin.Context) {
	rules, err := h.reviewService.ListReviewRules(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list review rules")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审核规则失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    rules,
	})
}

// AddSensitiveWord 添加敏感词
func (h *CharacterReviewHandler) AddSensitiveWord(c *gin.Context) {
	var req struct {
		Word     string `json:"word" binding:"required"`
		Category string `json:"category" binding:"required"`
		Level    int    `json:"level" binding:"required"`
		Action   string `json:"action" binding:"required"`
		Enabled  bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	word := &domain.SensitiveWord{
		ID:       uuid.New(),
		Word:     req.Word,
		Category: req.Category,
		Level:    req.Level,
		Action:   req.Action,
		Enabled:  req.Enabled,
	}

	err := h.reviewService.AddSensitiveWord(c.Request.Context(), word)
	if err != nil {
		h.logger.WithError(err).Error("Failed to add sensitive word")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加敏感词失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "敏感词添加成功",
		"data":    word,
	})
}

// ListSensitiveWords 列出敏感词
func (h *CharacterReviewHandler) ListSensitiveWords(c *gin.Context) {
	words, err := h.reviewService.ListSensitiveWords(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list sensitive words")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取敏感词列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    words,
	})
}
