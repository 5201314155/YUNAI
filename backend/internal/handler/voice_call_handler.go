package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"yunai/internal/domain"
	"yunai/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// VoiceCallHandler 语音通话处理器
type VoiceCallHandler struct {
	logger              *logrus.Logger
	voiceCallService    *service.AIVoiceCallService
	voiceServiceManager *service.VoiceServiceManager
}

// NewVoiceCallHandler 创建语音通话处理器
func NewVoiceCallHandler(
	logger *logrus.Logger,
	voiceCallService *service.AIVoiceCallService,
	voiceServiceManager *service.VoiceServiceManager,
) *VoiceCallHandler {
	return &VoiceCallHandler{
		logger:              logger,
		voiceCallService:    voiceCallService,
		voiceServiceManager: voiceServiceManager,
	}
}

// StartVoiceCall 开始语音通话
// @Summary 开始语音通话
// @Description 用户发送语音，AI角色回复语音
// @Tags 语音通话
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData string true "用户ID"
// @Param character_id formData string true "角色ID"
// @Param session_id formData string false "会话ID"
// @Param audio formData file true "音频文件"
// @Success 200 {object} domain.VoiceCallResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/voice/call [post]
func (h *VoiceCallHandler) StartVoiceCall(c *gin.Context) {
	h.logger.Info("收到语音通话请求")

	// 解析表单参数
	userID := c.PostForm("user_id")
	characterID := c.PostForm("character_id")
	sessionID := c.PostForm("session_id")

	if userID == "" || characterID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "参数错误",
			Message: "user_id和character_id不能为空",
		})
		return
	}

	// 获取音频文件
	file, err := c.FormFile("audio")
	if err != nil {
		h.logger.WithError(err).Error("获取音频文件失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "音频文件错误",
			Message: "无法获取音频文件",
		})
		return
	}

	// 读取音频数据
	audioFile, err := file.Open()
	if err != nil {
		h.logger.WithError(err).Error("打开音频文件失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "文件处理错误",
			Message: "无法打开音频文件",
		})
		return
	}
	defer audioFile.Close()

	audioData, err := io.ReadAll(audioFile)
	if err != nil {
		h.logger.WithError(err).Error("读取音频数据失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "文件处理错误",
			Message: "无法读取音频数据",
		})
		return
	}

	// 构建请求
	req := &domain.VoiceCallRequest{
		UserID:      userID,
		CharacterID: characterID,
		SessionID:   sessionID,
		AudioData:   audioData,
	}

	// 调用语音通话服务
	response, err := h.voiceCallService.StartVoiceCall(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("语音通话处理失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "语音通话失败",
			Message: err.Error(),
		})
		return
	}

	// 将音频数据转换为base64返回
	response.AudioURL = fmt.Sprintf("data:audio/mp3;base64,%s", base64.StdEncoding.EncodeToString(response.AudioData))
	response.AudioData = nil // 清空原始数据，避免重复传输

	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"character_id": characterID,
		"session_id":   response.SessionID,
	}).Info("语音通话处理成功")

	c.JSON(http.StatusOK, response)
}

// AutoCall 自动外呼
// @Summary 自动外呼
// @Description AI角色主动给用户打电话
// @Tags 语音通话
// @Accept json
// @Produce json
// @Param request body domain.AutoCallRequest true "自动外呼请求"
// @Success 200 {object} domain.AutoCallResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/voice/auto-call [post]
func (h *VoiceCallHandler) AutoCall(c *gin.Context) {
	h.logger.Info("收到自动外呼请求")

	var req domain.AutoCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析自动外呼请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "请求格式错误",
			Message: err.Error(),
		})
		return
	}

	// 调用自动外呼服务
	response, err := h.voiceCallService.AutoCall(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("自动外呼失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "自动外呼失败",
			Message: err.Error(),
		})
		return
	}

	// 将音频数据转换为base64返回
	response.AudioURL = fmt.Sprintf("data:audio/mp3;base64,%s", base64.StdEncoding.EncodeToString(response.AudioData))
	response.AudioData = nil // 清空原始数据

	h.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"character_id": req.CharacterID,
		"session_id":   response.SessionID,
	}).Info("自动外呼成功")

	c.JSON(http.StatusOK, response)
}

// CloneVoice 克隆音色
// @Summary 克隆音色
// @Description 用户上传音频样本克隆音色
// @Tags 语音通话
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData string true "用户ID"
// @Param provider_id formData string true "提供商ID"
// @Param voice_name formData string true "音色名称"
// @Param reference_text formData string true "参考文本"
// @Param cover_image formData string false "封面图片URL"
// @Param audio formData file true "音频样本文件"
// @Success 200 {object} domain.VoiceCloneResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/voice/clone [post]
func (h *VoiceCallHandler) CloneVoice(c *gin.Context) {
	h.logger.Info("收到音色克隆请求")

	// 解析表单参数
	userID := c.PostForm("user_id")
	providerID := c.PostForm("provider_id")
	voiceName := c.PostForm("voice_name")
	referenceText := c.PostForm("reference_text")
	coverImage := c.PostForm("cover_image")

	if userID == "" || providerID == "" || voiceName == "" || referenceText == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "参数错误",
			Message: "user_id、provider_id、voice_name和reference_text不能为空",
		})
		return
	}

	// 获取音频文件
	file, err := c.FormFile("audio")
	if err != nil {
		h.logger.WithError(err).Error("获取音频文件失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "音频文件错误",
			Message: "无法获取音频文件",
		})
		return
	}

	// 读取音频数据
	audioFile, err := file.Open()
	if err != nil {
		h.logger.WithError(err).Error("打开音频文件失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "文件处理错误",
			Message: "无法打开音频文件",
		})
		return
	}
	defer audioFile.Close()

	audioData, err := io.ReadAll(audioFile)
	if err != nil {
		h.logger.WithError(err).Error("读取音频数据失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "文件处理错误",
			Message: "无法读取音频数据",
		})
		return
	}

	// 构建请求
	req := &domain.VoiceCloneRequest{
		UserID:        userID,
		ProviderID:    providerID,
		VoiceName:     voiceName,
		AudioData:     audioData,
		AudioBase64:   base64.StdEncoding.EncodeToString(audioData),
		ReferenceText: referenceText,
		CoverImage:    coverImage,
	}

	// 调用音色克隆服务
	response, err := h.voiceCallService.CloneUserVoice(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("音色克隆失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "音色克隆失败",
			Message: err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"voice_name": voiceName,
		"voice_id":   response.VoiceID,
	}).Info("音色克隆成功")

	c.JSON(http.StatusOK, response)
}

// EndVoiceCall 结束语音通话
// @Summary 结束语音通话
// @Description 结束当前语音通话会话
// @Tags 语音通话
// @Accept json
// @Produce json
// @Param session_id path string true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/voice/call/{session_id}/end [post]
func (h *VoiceCallHandler) EndVoiceCall(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "参数错误",
			Message: "session_id不能为空",
		})
		return
	}

	err := h.voiceCallService.EndVoiceCall(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("结束语音通话失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "结束通话失败",
			Message: err.Error(),
		})
		return
	}

	h.logger.WithField("session_id", sessionID).Info("语音通话已结束")

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "语音通话已结束",
	})
}

// GetVoiceCallHistory 获取语音通话历史
// @Summary 获取语音通话历史
// @Description 获取用户的语音通话历史记录
// @Tags 语音通话
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param limit query int false "限制数量" default(20)
// @Success 200 {array} domain.VoiceCallSession
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/voice/history/{user_id} [get]
func (h *VoiceCallHandler) GetVoiceCallHistory(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "参数错误",
			Message: "user_id不能为空",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	history, err := h.voiceCallService.GetVoiceCallHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.WithError(err).Error("获取语音通话历史失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "获取历史记录失败",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetSessionMessages 获取会话消息
// @Summary 获取会话消息
// @Description 获取指定会话的所有消息
// @Tags 语音通话
// @Accept json
// @Produce json
// @Param session_id path string true "会话ID"
// @Success 200 {array} domain.VoiceMessage
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/voice/session/{session_id}/messages [get]
func (h *VoiceCallHandler) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "参数错误",
			Message: "session_id不能为空",
		})
		return
	}

	messages, err := h.voiceCallService.GetSessionMessages(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("获取会话消息失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "获取消息失败",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// 响应结构

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
