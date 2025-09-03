package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"yunai/internal/handler"
	"yunai/internal/service"
)

// SetupPaymentRoutes 设置支付路由
func SetupPaymentRoutes(
	router *gin.Engine,
	paymentService service.PaymentService,
	logger *logrus.Logger,
) {
	// 创建处理器
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	// API路由组
	api := router.Group("/api/v1")
	{
		// 支付管理
		payment := api.Group("/payment")
		{
			// 公开路由（不需要认证）
			payment.GET("/packages", paymentHandler.GetRechargePackages)

			// 需要认证的路由（暂时不使用认证中间件，直接开放所有路由用于测试）
			payment.POST("/cards", paymentHandler.BindPaymentCard)
			payment.GET("/cards", paymentHandler.GetPaymentCards)
			payment.POST("/cards/:id/default", paymentHandler.SetDefaultPaymentCard)
			payment.DELETE("/cards/:id", paymentHandler.DeletePaymentCard)
			payment.POST("/password/set", paymentHandler.SetPaymentPassword)
			payment.POST("/password/verify", paymentHandler.VerifyPaymentPassword)
			payment.POST("/orders/recharge", paymentHandler.CreateRechargeOrder)
		}
	}
}
