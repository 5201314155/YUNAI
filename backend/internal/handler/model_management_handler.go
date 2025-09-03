package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/service"
)

// ModelManagementHandler 模型管理处理器
type ModelManagementHandler struct {
	modelService service.ModelManagementService
	logger       *logrus.Logger
}

// NewModelManagementHandler 创建模型管理处理器
func NewModelManagementHandler(modelService service.ModelManagementService, logger *logrus.Logger) *ModelManagementHandler {
	return &ModelManagementHandler{
		modelService: modelService,
		logger:       logger,
	}
}

// GetAvailableModels 获取可用模型列表
// GET /api/v1/models/available
func (h *ModelManagementHandler) GetAvailableModels(c *gin.Context) {
	modelType := c.Query("type")
	userIDStr := c.Query("user_id")

	var userID *uuid.UUID
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			userID = &uid
		}
	}

	models, err := h.modelService.GetAvailableModels(c.Request.Context(), userID, modelType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get available models")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取可用模型失败"})
		return
	}

	// 过滤敏感信息
	var publicModels []map[string]interface{}
	for _, model := range models {
		publicModel := map[string]interface{}{
			"id":            model.ID,
			"internal_key":  model.InternalKey,
			"display_name":  model.DisplayName,
			"provider":      model.Provider,
			"model_type":    model.ModelType,
			"health_status": model.HealthStatus,
			"is_featured":   model.IsFeatured,
			"capabilities":  model.Capabilities,
			"pricing":       model.Pricing,
		}
		publicModels = append(publicModels, publicModel)
	}

	c.JSON(http.StatusOK, gin.H{
		"models": publicModels,
		"total":  len(publicModels),
	})
}

// GetAllModels 获取所有模型（管理员接口）
// GET /api/v1/admin/models
func (h *ModelManagementHandler) GetAllModels(c *gin.Context) {
	includeDisabled := c.Query("include_disabled") == "true"

	models, err := h.modelService.GetAllModels(c.Request.Context(), includeDisabled)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all models")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  len(models),
	})
}

// EnableModel 启用模型
// POST /api/v1/admin/models/:id/enable
func (h *ModelManagementHandler) EnableModel(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
		return
	}

	var req struct {
		AdminID uuid.UUID `json:"admin_id" binding:"required"`
		Reason  string    `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.modelService.EnableModel(c.Request.Context(), modelID, req.AdminID, req.Reason); err != nil {
		h.logger.WithError(err).Error("Failed to enable model")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "启用模型失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "模型启用成功",
		"model_id": modelID,
	})
}

// DisableModel 禁用模型
// POST /api/v1/admin/models/:id/disable
func (h *ModelManagementHandler) DisableModel(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
		return
	}

	var req struct {
		AdminID uuid.UUID `json:"admin_id" binding:"required"`
		Reason  string    `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.modelService.DisableModel(c.Request.Context(), modelID, req.AdminID, req.Reason); err != nil {
		h.logger.WithError(err).Error("Failed to disable model")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "禁用模型失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "模型禁用成功",
		"model_id": modelID,
		"reason":   req.Reason,
	})
}

// GetModelDetails 获取模型详情
// GET /api/v1/models/:id
func (h *ModelManagementHandler) GetModelDetails(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
		return
	}

	model, err := h.modelService.GetModelByID(c.Request.Context(), modelID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get model details")
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}

	c.JSON(http.StatusOK, model)
}

// PerformHealthCheck 执行健康检查
// POST /api/v1/admin/models/:id/health-check
func (h *ModelManagementHandler) PerformHealthCheck(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
		return
	}

	result, err := h.modelService.PerformHealthCheck(c.Request.Context(), modelID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to perform health check")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "健康检查失败"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CheckModelAccess 检查模型访问权限
// GET /api/v1/models/:id/access
func (h *ModelManagementHandler) CheckModelAccess(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
		return
	}

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	accessible, err := h.modelService.CheckModelAccess(c.Request.Context(), modelID, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check model access")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查访问权限失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessible": accessible,
		"model_id":   modelID,
		"user_id":    userID,
	})
}

// GetModelsByType 按类型获取模型
// GET /api/v1/models/by-type/:type
func (h *ModelManagementHandler) GetModelsByType(c *gin.Context) {
	modelType := c.Param("type")
	userIDStr := c.Query("user_id")

	var userID *uuid.UUID
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			userID = &uid
		}
	}

	models, err := h.modelService.GetAvailableModels(c.Request.Context(), userID, modelType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get models by type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型失败"})
		return
	}

	// 按功能分组
	groupedModels := make(map[string][]map[string]interface{})

	for _, model := range models {
		publicModel := map[string]interface{}{
			"id":            model.ID,
			"internal_key":  model.InternalKey,
			"display_name":  model.DisplayName,
			"provider":      model.Provider,
			"health_status": model.HealthStatus,
			"is_featured":   model.IsFeatured,
			"weight":        model.Weight,
		}

		groupedModels[model.Provider] = append(groupedModels[model.Provider], publicModel)
	}

	c.JSON(http.StatusOK, gin.H{
		"model_type": modelType,
		"providers":  groupedModels,
		"total":      len(models),
	})
}

// GetModelStats 获取模型统计信息
// GET /api/v1/admin/models/:id/stats
func (h *ModelManagementHandler) GetModelStats(c *gin.Context) {
	modelIDStr := c.Param("id")
	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
		return
	}

	timeRange := c.DefaultQuery("time_range", "24h")

	stats, err := h.modelService.GetModelUsageStats(c.Request.Context(), modelID, timeRange)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get model stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BatchHealthCheck 批量健康检查
// POST /api/v1/admin/models/batch-health-check
func (h *ModelManagementHandler) BatchHealthCheck(c *gin.Context) {
	results, err := h.modelService.PerformBatchHealthCheck(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to perform batch health check")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量健康检查失败"})
		return
	}

	// 统计结果
	var healthyCount, unhealthyCount, errorCount int
	for _, result := range results {
		switch result.Status {
		case "healthy":
			healthyCount++
		case "unhealthy":
			unhealthyCount++
		default:
			errorCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"summary": gin.H{
			"total":     len(results),
			"healthy":   healthyCount,
			"unhealthy": unhealthyCount,
			"error":     errorCount,
		},
	})
}

// RegisterRoutes 注册路由
func (h *ModelManagementHandler) RegisterRoutes(router *gin.RouterGroup) {
	// 公开接口
	models := router.Group("/models")
	{
		models.GET("/available", h.GetAvailableModels)
		models.GET("/:id", h.GetModelDetails)
		models.GET("/:id/access", h.CheckModelAccess)
		models.GET("/by-type/:type", h.GetModelsByType)
	}

	// 管理员接口
	admin := router.Group("/admin/models")
	{
		admin.GET("", h.GetAllModels)
		admin.POST("/:id/enable", h.EnableModel)
		admin.POST("/:id/disable", h.DisableModel)
		admin.POST("/:id/health-check", h.PerformHealthCheck)
		admin.GET("/:id/stats", h.GetModelStats)
		admin.POST("/batch-health-check", h.BatchHealthCheck)
	}
}
