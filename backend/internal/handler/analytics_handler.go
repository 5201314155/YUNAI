package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/service"
)

// AnalyticsHandler 分析处理器
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
	alertService     service.AlertService
	logger           *logrus.Logger
}

// NewAnalyticsHandler 创建分析处理器
func NewAnalyticsHandler(
	analyticsService service.AnalyticsService,
	alertService service.AlertService,
	logger *logrus.Logger,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		alertService:     alertService,
		logger:           logger,
	}
}

// GetDashboard 获取分析面板
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	dashboard, err := h.analyticsService.GetDashboard(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get analytics dashboard")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分析面板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    dashboard,
	})
}

// GetSystemHealth 获取系统健康状态
func (h *AnalyticsHandler) GetSystemHealth(c *gin.Context) {
	health, err := h.analyticsService.GetSystemHealthMetrics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system health")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统健康状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    health,
	})
}

// GetRealTimeMetrics 获取实时指标
func (h *AnalyticsHandler) GetRealTimeMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetRealTimeMetrics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get real-time metrics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取实时指标失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    metrics,
	})
}

// GetUserStats 获取用户统计
func (h *AnalyticsHandler) GetUserStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	timeRange := c.DefaultQuery("time_range", "week")

	stats, err := h.analyticsService.GetUserBehaviorStats(c.Request.Context(), userID, timeRange)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// TrackEvent 追踪事件
func (h *AnalyticsHandler) TrackEvent(c *gin.Context) {
	var req struct {
		Event      string                 `json:"event" binding:"required"`
		Properties map[string]interface{} `json:"properties"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 追踪事件
	if err := h.analyticsService.TrackEvent(c.Request.Context(), uid, req.Event, req.Properties); err != nil {
		h.logger.WithError(err).Error("Failed to track event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "事件追踪失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "事件追踪成功",
	})
}

// CreateAlertRule 创建告警规则
func (h *AnalyticsHandler) CreateAlertRule(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Metric      string  `json:"metric" binding:"required"`
		Operator    string  `json:"operator" binding:"required"`
		Threshold   float64 `json:"threshold" binding:"required"`
		Duration    int     `json:"duration"`
		Severity    string  `json:"severity" binding:"required"`
		Enabled     bool    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	rule := &domain.AlertRule{
		Name:        req.Name,
		Description: req.Description,
		Metric:      req.Metric,
		Operator:    req.Operator,
		Threshold:   req.Threshold,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Enabled:     req.Enabled,
	}

	if err := h.alertService.CreateAlertRule(c.Request.Context(), rule); err != nil {
		h.logger.WithError(err).Error("Failed to create alert rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建告警规则失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "告警规则创建成功",
		"data":    rule,
	})
}

// ListAlertRules 列出告警规则
func (h *AnalyticsHandler) ListAlertRules(c *gin.Context) {
	rules, err := h.alertService.ListAlertRules(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list alert rules")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取告警规则失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    rules,
	})
}

// GetAlerts 获取告警列表
func (h *AnalyticsHandler) GetAlerts(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	alerts, err := h.alertService.GetAlerts(c.Request.Context(), status, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get alerts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取告警列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alerts,
	})
}

// ResolveAlert 解决告警
func (h *AnalyticsHandler) ResolveAlert(c *gin.Context) {
	alertIDStr := c.Param("alert_id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警ID"})
		return
	}

	if err := h.alertService.ResolveAlert(c.Request.Context(), alertID); err != nil {
		h.logger.WithError(err).Error("Failed to resolve alert")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解决告警失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "告警已解决",
	})
}

// GetDAU 获取日活跃用户数
func (h *AnalyticsHandler) GetDAU(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日期格式"})
		return
	}

	dau, err := h.analyticsService.GetDAU(c.Request.Context(), date)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get DAU")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取DAU失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"date": dateStr,
			"dau":  dau,
		},
	})
}

// GetMAU 获取月活跃用户数
func (h *AnalyticsHandler) GetMAU(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日期格式"})
		return
	}

	mau, err := h.analyticsService.GetMAU(c.Request.Context(), date)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get MAU")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取MAU失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"date": dateStr,
			"mau":  mau,
		},
	})
}

// GetSystemStats 获取系统统计
func (h *AnalyticsHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.analyticsService.GetSystemStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// UpdateUserStats 更新用户统计
func (h *AnalyticsHandler) UpdateUserStats(c *gin.Context) {
	if err := h.analyticsService.UpdateUserStats(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to update user stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户统计更新成功",
	})
}

// CheckAlerts 手动检查告警
func (h *AnalyticsHandler) CheckAlerts(c *gin.Context) {
	if err := h.alertService.CheckAlerts(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to check alerts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查告警失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "告警检查完成",
	})
}
