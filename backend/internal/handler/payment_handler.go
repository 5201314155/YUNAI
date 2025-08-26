package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/service"
	"yunai/pkg/auth"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService service.PaymentService
	logger         *logrus.Logger
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService service.PaymentService, logger *logrus.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// BindPaymentCard 绑定支付卡片
// @Summary 绑定支付卡片
// @Description 用户绑定新的支付卡片
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param request body domain.BindCardRequest true "绑卡请求"
// @Success 200 {object} domain.PaymentCard
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/payment/cards [post]
func (h *PaymentHandler) BindPaymentCard(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	var req domain.BindCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	card, err := h.paymentService.BindPaymentCard(c.Request.Context(), userID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to bind payment card")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "绑卡失败",
		})
		return
	}

	c.JSON(http.StatusOK, card)
}

// GetPaymentCards 获取支付卡片列表
// @Summary 获取支付卡片列表
// @Description 获取用户的支付卡片列表
// @Tags 支付管理
// @Produce json
// @Success 200 {array} domain.PaymentCard
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/payment/cards [get]
func (h *PaymentHandler) GetPaymentCards(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	cards, err := h.paymentService.GetUserPaymentCards(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get payment cards")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "获取卡片列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, cards)
}

// SetDefaultPaymentCard 设置默认支付卡片
// @Summary 设置默认支付卡片
// @Description 设置用户的默认支付卡片
// @Tags 支付管理
// @Param card_id path string true "卡片ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/payment/cards/{card_id}/default [put]
func (h *PaymentHandler) SetDefaultPaymentCard(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	cardIDStr := c.Param("card_id")
	cardID, err := uuid.Parse(cardIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "卡片ID格式无效",
		})
		return
	}

	if err := h.paymentService.SetDefaultPaymentCard(c.Request.Context(), userID, cardID); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to set default payment card")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "设置默认卡片失败",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "默认卡片设置成功",
	})
}

// DeletePaymentCard 删除支付卡片
// @Summary 删除支付卡片
// @Description 删除用户的支付卡片
// @Tags 支付管理
// @Param card_id path string true "卡片ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/payment/cards/{card_id} [delete]
func (h *PaymentHandler) DeletePaymentCard(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	cardIDStr := c.Param("card_id")
	cardID, err := uuid.Parse(cardIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "卡片ID格式无效",
		})
		return
	}

	if err := h.paymentService.DeletePaymentCard(c.Request.Context(), userID, cardID); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to delete payment card")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "删除卡片失败",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "卡片删除成功",
	})
}

// SetPaymentPassword 设置支付密码
// @Summary 设置支付密码
// @Description 用户设置支付密码
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param request body domain.SetPaymentPasswordRequest true "设置支付密码请求"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/payment/password [post]
func (h *PaymentHandler) SetPaymentPassword(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	var req domain.SetPaymentPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	if err := h.paymentService.SetPaymentPassword(c.Request.Context(), userID, &req); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to set payment password")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "设置支付密码失败",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "支付密码设置成功",
	})
}

// VerifyPaymentPassword 验证支付密码
// @Summary 验证支付密码
// @Description 验证用户的支付密码
// @Tags 支付管理
// @Accept json
// @Produce json
// @Param request body domain.VerifyPaymentPasswordRequest true "验证支付密码请求"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/payment/password/verify [post]
func (h *PaymentHandler) VerifyPaymentPassword(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	var req domain.VerifyPaymentPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	if err := h.paymentService.VerifyPaymentPassword(c.Request.Context(), userID, &req); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to verify payment password")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "验证支付密码失败",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "支付密码验证成功",
	})
}

// GetRechargePackages 获取充值套餐
// @Summary 获取充值套餐
// @Description 获取可用的充值套餐列表
// @Tags 充值管理
// @Produce json
// @Success 200 {array} domain.RechargePackage
// @Router /api/v1/recharge/packages [get]
func (h *PaymentHandler) GetRechargePackages(c *gin.Context) {
	packages, err := h.paymentService.GetRechargePackages(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get recharge packages")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "获取充值套餐失败",
		})
		return
	}

	c.JSON(http.StatusOK, packages)
}

// CreateRechargeOrder 创建充值订单
// @Summary 创建充值订单
// @Description 创建新的充值订单
// @Tags 充值管理
// @Accept json
// @Produce json
// @Param request body domain.CreateRechargeOrderRequest true "创建充值订单请求"
// @Success 200 {object} domain.PaymentOrderResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/recharge/orders [post]
func (h *PaymentHandler) CreateRechargeOrder(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	var req domain.CreateRechargeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	response, err := h.paymentService.CreateRechargeOrder(c.Request.Context(), userID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to create recharge order")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "创建充值订单失败",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRechargeOrder 获取充值订单详情
// @Summary 获取充值订单详情
// @Description 获取指定充值订单的详细信息
// @Tags 充值管理
// @Param order_id path string true "订单ID"
// @Success 200 {object} domain.RechargeOrder
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/recharge/orders/{order_id} [get]
func (h *PaymentHandler) GetRechargeOrder(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "订单ID格式无效",
		})
		return
	}

	order, err := h.paymentService.GetRechargeOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to get recharge order")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "获取订单详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetUserRechargeOrders 获取用户充值订单列表
// @Summary 获取用户充值订单列表
// @Description 获取用户的充值订单历史
// @Tags 充值管理
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {array} domain.RechargeOrder
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/recharge/orders [get]
func (h *PaymentHandler) GetUserRechargeOrders(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	orders, err := h.paymentService.GetUserRechargeOrders(c.Request.Context(), userID, offset, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user recharge orders")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "获取充值记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// ProcessPayment 处理支付
// @Summary 处理支付
// @Description 处理充值订单的支付
// @Tags 充值管理
// @Accept json
// @Param order_id path string true "订单ID"
// @Param request body object{payment_password=string} true "支付密码"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/recharge/orders/{order_id}/pay [post]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "订单ID格式无效",
		})
		return
	}

	var req struct {
		PaymentPassword string `json:"payment_password" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    domain.CodeInvalidRequest,
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	if err := h.paymentService.ProcessPayment(c.Request.Context(), orderID, req.PaymentPassword); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.HTTPStatus(), domain.ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to process payment")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    domain.CodeInternalError,
			Message: "支付处理失败",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "支付成功",
	})
}
