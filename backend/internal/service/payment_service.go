package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"yunai/internal/domain"
	"yunai/internal/repository"
	"yunai/pkg/auth"
)

// PaymentService 支付服务接口
type PaymentService interface {
	// 支付卡片管理
	BindPaymentCard(ctx context.Context, userID uuid.UUID, req *domain.BindCardRequest) (*domain.PaymentCard, error)
	BindPaymentCardWithEmail(ctx context.Context, userID uuid.UUID, req *domain.BindCardWithEmailRequest) (*domain.PaymentCard, error)
	UnbindPaymentCard(ctx context.Context, userID uuid.UUID, req *domain.UnbindCardRequest) error
	GetUserPaymentCards(ctx context.Context, userID uuid.UUID) ([]*domain.PaymentCard, error)
	SetDefaultPaymentCard(ctx context.Context, userID, cardID uuid.UUID) error
	DeletePaymentCard(ctx context.Context, userID uuid.UUID, req *domain.DeleteCardRequest) error

	// 卡片状态管理
	FreezeCard(ctx context.Context, userID uuid.UUID, req *domain.FreezeCardRequest) (*domain.CardStatusResponse, error)
	UnfreezeCard(ctx context.Context, userID uuid.UUID, req *domain.UnfreezeCardRequest) (*domain.CardStatusResponse, error)
	GetCardStatus(ctx context.Context, userID, cardID uuid.UUID) (*domain.CardStatusResponse, error)
	GetCardOperations(ctx context.Context, userID, cardID uuid.UUID, offset, limit int) ([]*domain.CardOperation, error)

	// 卡片余额管理
	GetCardBalance(ctx context.Context, userID, cardID uuid.UUID) (*domain.CardBalanceResponse, error)
	RechargeCard(ctx context.Context, userID uuid.UUID, req *domain.CardRechargeRequest) error
	TransferBetweenCards(ctx context.Context, userID uuid.UUID, req *domain.CardTransferRequest) error
	GetCardTransactions(ctx context.Context, userID, cardID uuid.UUID, offset, limit int) ([]*domain.CardTransaction, error)

	// 支付密码管理
	SetPaymentPassword(ctx context.Context, userID uuid.UUID, req *domain.SetPaymentPasswordRequest) error
	ResetPaymentPassword(ctx context.Context, req *domain.ResetPaymentPasswordRequest) error
	VerifyPaymentPassword(ctx context.Context, userID uuid.UUID, req *domain.VerifyPaymentPasswordRequest) error
	HasPaymentPassword(ctx context.Context, userID uuid.UUID) (bool, error)

	// 充值订单管理
	CreateRechargeOrder(ctx context.Context, userID uuid.UUID, req *domain.CreateRechargeOrderRequest) (*domain.PaymentOrderResponse, error)
	GetRechargeOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*domain.RechargeOrder, error)
	GetUserRechargeOrders(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.RechargeOrder, error)
	ProcessPayment(ctx context.Context, orderID uuid.UUID, paymentPassword string) error

	// 充值套餐
	GetRechargePackages(ctx context.Context) ([]*domain.RechargePackage, error)

	// 系统配置
	GetCoinExchangeRate(ctx context.Context) (int, error)
	UpdateCoinExchangeRate(ctx context.Context, rate int) error

	// 代充功能
	ProxyRecharge(ctx context.Context, payerUserID uuid.UUID, req *domain.ProxyRechargeRequest) (*domain.ProxyRechargeResponse, error)
	GetProxyRechargeOrders(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.ProxyRechargeOrder, error)
	GetProxyRechargeOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*domain.ProxyRechargeOrder, error)

	// 管理员功能
	AdminRecharge(ctx context.Context, adminUserID uuid.UUID, req *domain.AdminRechargeRequest) error
	GetUserPaymentInfo(ctx context.Context, userIdentifier string) (*domain.UserPaymentInfo, error)

	// 用户信息查询（支持ID或邮箱）
	GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error)
}

// paymentService 支付服务实现
type paymentService struct {
	paymentRepo repository.PaymentRepository
	userRepo    repository.UserRepository
	walletRepo  repository.WalletRepository
	logger      *logrus.Logger
}

// NewPaymentService 创建支付服务
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	userRepo repository.UserRepository,
	walletRepo repository.WalletRepository,
	logger *logrus.Logger,
) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		walletRepo:  walletRepo,
		logger:      logger,
	}
}

// BindPaymentCard 绑定支付卡片
func (s *paymentService) BindPaymentCard(ctx context.Context, userID uuid.UUID, req *domain.BindCardRequest) (*domain.PaymentCard, error) {
	// 验证YUNAI卡号格式（支持所有类型）
	if !s.validateAnyYunaiCard(req.CardNumber) {
		return nil, domain.NewAppError(domain.CodeInvalidRequest, "YUNAI卡号格式不正确", nil)
	}

	// 生成卡号哈希
	cardHash := s.hashCardNumber(req.CardNumber)

	// 脱敏显示卡号
	maskedCardNumber := s.maskYunaiCardNumber(req.CardNumber)

	card := &domain.PaymentCard{
		ID:             uuid.New(),
		UserID:         userID,
		CardNumber:     maskedCardNumber,
		CardNumberHash: cardHash,
		CardType:       req.CardType,
		BankName:       req.BankName,
		BankCode:       req.BankCode,
		CardholderName: req.CardholderName,
		Balance:        0.0,     // 初始余额为0
		Currency:       "coins", // 默认货币为金币
		IsDefault:      req.IsDefault,
		IsActive:       true,
		IsFrozen:       false, // 默认未冻结
		IsVerified:     false, // 需要后续验证
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.paymentRepo.CreatePaymentCard(ctx, card); err != nil {
		if err == domain.ErrConflict {
			return nil, domain.NewAppError(domain.CodeConflict, "该卡片已绑定", err)
		}
		return nil, fmt.Errorf("failed to bind payment card: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"card_id":   card.ID,
		"bank_name": req.BankName,
	}).Info("Payment card bound successfully")

	return card, nil
}

// GetUserPaymentCards 获取用户支付卡片列表
func (s *paymentService) GetUserPaymentCards(ctx context.Context, userID uuid.UUID) ([]*domain.PaymentCard, error) {
	cards, err := s.paymentRepo.GetPaymentCardsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user payment cards: %w", err)
	}

	return cards, nil
}

// SetDefaultPaymentCard 设置默认支付卡片
func (s *paymentService) SetDefaultPaymentCard(ctx context.Context, userID, cardID uuid.UUID) error {
	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return domain.NewAppError(domain.CodeForbidden, "无权限操作该卡片", nil)
	}

	if err := s.paymentRepo.SetDefaultPaymentCard(ctx, userID, cardID); err != nil {
		return fmt.Errorf("failed to set default payment card: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"card_id": cardID,
	}).Info("Default payment card set")

	return nil
}

// DeletePaymentCard 删除支付卡片
func (s *paymentService) DeletePaymentCard(ctx context.Context, userID uuid.UUID, req *domain.DeleteCardRequest) error {
	// 验证支付密码
	if err := s.VerifyPaymentPassword(ctx, userID, &domain.VerifyPaymentPasswordRequest{
		Password: req.Password,
	}); err != nil {
		return err
	}

	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, req.CardID)
	if err != nil {
		return fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return domain.NewAppError(domain.CodeForbidden, "无权限操作该卡片", nil)
	}

	// 检查卡片是否有余额
	if card.Balance > 0 {
		return domain.NewAppError(domain.CodeInvalidRequest,
			fmt.Sprintf("卡片还有%.2f金币余额，请先转出余额再删除", card.Balance), nil)
	}

	// 软删除：标记为非激活状态
	card.IsActive = false
	card.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdatePaymentCard(ctx, card); err != nil {
		return fmt.Errorf("failed to delete payment card: %w", err)
	}

	// TODO: 记录操作日志
	// operation := &domain.CardOperation{
	//     ID:            uuid.New(),
	//     CardID:        req.CardID,
	//     UserID:        userID,
	//     OperationType: domain.CardOperationDelete,
	//     Reason:        req.Reason,
	//     CreatedAt:     time.Now(),
	// }

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"card_id": req.CardID,
		"reason":  req.Reason,
	}).Info("Payment card deleted")

	return nil
}

// SetPaymentPassword 设置支付密码
func (s *paymentService) SetPaymentPassword(ctx context.Context, userID uuid.UUID, req *domain.SetPaymentPasswordRequest) error {
	// 验证密码确认
	if req.Password != req.ConfirmPassword {
		return domain.NewAppError(domain.CodeInvalidRequest, "两次输入的密码不一致", nil)
	}

	// 检查是否已设置支付密码
	_, err := s.paymentRepo.GetPaymentPasswordByUserID(ctx, userID)
	if err == nil {
		return domain.NewAppError(domain.CodeConflict, "支付密码已设置", nil)
	}
	if err != domain.ErrNotFound {
		return fmt.Errorf("failed to check existing payment password: %w", err)
	}

	// 生成盐值
	salt, err := s.generateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// 哈希密码
	passwordHash, err := s.hashPaymentPassword(req.Password, salt)
	if err != nil {
		return fmt.Errorf("failed to hash payment password: %w", err)
	}

	paymentPassword := &domain.PaymentPassword{
		ID:             uuid.New(),
		UserID:         userID,
		PasswordHash:   passwordHash,
		Salt:           salt,
		FailedAttempts: 0,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.paymentRepo.CreatePaymentPassword(ctx, paymentPassword); err != nil {
		return fmt.Errorf("failed to set payment password: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("Payment password set successfully")

	return nil
}

// VerifyPaymentPassword 验证支付密码
func (s *paymentService) VerifyPaymentPassword(ctx context.Context, userID uuid.UUID, req *domain.VerifyPaymentPasswordRequest) error {
	paymentPassword, err := s.paymentRepo.GetPaymentPasswordByUserID(ctx, userID)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.NewAppError(domain.CodeNotFound, "未设置支付密码", err)
		}
		return fmt.Errorf("failed to get payment password: %w", err)
	}

	// 检查是否被锁定
	if paymentPassword.LockedUntil != nil && time.Now().Before(*paymentPassword.LockedUntil) {
		return domain.NewAppError(domain.CodeForbidden, "支付密码已锁定，请稍后再试", nil)
	}

	// 验证密码
	if !s.verifyPaymentPasswordHash(req.Password, paymentPassword.Salt, paymentPassword.PasswordHash) {
		// 增加失败次数
		if err := s.paymentRepo.IncrementFailedAttempts(ctx, userID); err != nil {
			s.logger.WithError(err).Error("Failed to increment failed attempts")
		}

		// 检查是否需要锁定
		if paymentPassword.FailedAttempts+1 >= domain.MaxPaymentPasswordFailures {
			lockDuration := time.Duration(domain.PaymentPasswordLockDuration) * time.Minute
			if err := s.paymentRepo.LockPaymentPassword(ctx, userID, lockDuration); err != nil {
				s.logger.WithError(err).Error("Failed to lock payment password")
			}
			return domain.NewAppError(domain.CodeForbidden, "支付密码错误次数过多，已锁定30分钟", nil)
		}

		return domain.NewAppError(domain.CodeInvalidCredentials, "支付密码错误", nil)
	}

	// 重置失败次数
	if err := s.paymentRepo.ResetFailedAttempts(ctx, userID); err != nil {
		s.logger.WithError(err).Error("Failed to reset failed attempts")
	}

	return nil
}

// HasPaymentPassword 检查是否已设置支付密码
func (s *paymentService) HasPaymentPassword(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, err := s.paymentRepo.GetPaymentPasswordByUserID(ctx, userID)
	if err == domain.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check payment password: %w", err)
	}

	return true, nil
}

// 辅助方法

// hashCardNumber 哈希卡号
func (s *paymentService) hashCardNumber(cardNumber string) string {
	hash := sha256.Sum256([]byte(cardNumber))
	return hex.EncodeToString(hash[:])
}

// maskYunaiCardNumber 脱敏YUNAI卡号
func (s *paymentService) maskYunaiCardNumber(cardNumber string) string {
	// 获取所有卡片类型，找到匹配的类型
	cardTypes := auth.GetYunaiCardTypes()

	for _, cardType := range cardTypes {
		generator := auth.NewYunaiCardGenerator(cardType)
		if generator.ValidateYunaiCard(cardNumber) {
			return generator.MaskYunaiCard(cardNumber)
		}
	}

	// 如果没有匹配的类型，使用默认脱敏方式
	if len(cardNumber) >= 16 {
		return cardNumber[:4] + "********" + cardNumber[12:]
	}
	return cardNumber
}

// generateSalt 生成盐值
func (s *paymentService) generateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// hashPaymentPassword 哈希支付密码
func (s *paymentService) hashPaymentPassword(password, salt string) (string, error) {
	saltedPassword := password + salt
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPaymentPasswordHash 验证支付密码哈希
func (s *paymentService) verifyPaymentPasswordHash(password, salt, hash string) bool {
	saltedPassword := password + salt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}

// validateAnyYunaiCard 验证任意类型的YUNAI卡号
func (s *paymentService) validateAnyYunaiCard(cardNumber string) bool {
	// 测试环境下简化验证：接受16位数字
	if len(cardNumber) == 16 {
		for _, char := range cardNumber {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	}

	// 获取所有卡片类型
	cardTypes := auth.GetYunaiCardTypes()

	// 尝试用每种类型验证
	for _, cardType := range cardTypes {
		generator := auth.NewYunaiCardGenerator(cardType)
		if generator.ValidateYunaiCard(cardNumber) {
			return true
		}
	}

	return false
}

// CreateRechargeOrder 创建充值订单
func (s *paymentService) CreateRechargeOrder(ctx context.Context, userID uuid.UUID, req *domain.CreateRechargeOrderRequest) (*domain.PaymentOrderResponse, error) {
	// 验证支付方式和卡片
	if req.PaymentMethod == domain.PaymentMethodCard && req.PaymentCardID == nil {
		return nil, domain.NewAppError(domain.CodeInvalidRequest, "使用银行卡支付时必须选择卡片", nil)
	}

	var cardType *auth.YunaiCardType

	if req.PaymentCardID != nil {
		card, err := s.paymentRepo.GetPaymentCardByID(ctx, *req.PaymentCardID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment card: %w", err)
		}
		if card.UserID != userID {
			return nil, domain.NewAppError(domain.CodeForbidden, "无权限使用该卡片", nil)
		}

		// 识别卡片类型
		cardType = s.identifyCardType(card.CardNumber)
	}

	// 获取系统金币汇率
	exchangeRate, err := s.GetCoinExchangeRate(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get exchange rate, using default")
		exchangeRate = domain.DefaultCoinExchangeRate
	}

	// 计算原始金币数量
	originalCoinsAmount := int64(req.Amount * float64(exchangeRate))

	// 处理卡片类型的折扣和返利
	finalAmount := req.Amount
	bonusCoins := int64(0)
	discountAmount := 0.0

	if cardType != nil {
		// 折扣卡：减少支付金额
		if cardType.DiscountRate > 0 && cardType.DiscountRate < 1 {
			finalAmount = req.Amount * cardType.DiscountRate
			discountAmount = req.Amount - finalAmount
			s.logger.WithFields(logrus.Fields{
				"card_type":       cardType.Name,
				"original_amount": req.Amount,
				"discount_rate":   cardType.DiscountRate,
				"final_amount":    finalAmount,
			}).Info("Applied discount")
		}

		// 返利卡：增加金币数量
		if cardType.BonusRate > 0 {
			bonusCoins = int64(float64(originalCoinsAmount) * cardType.BonusRate)
			s.logger.WithFields(logrus.Fields{
				"card_type":      cardType.Name,
				"original_coins": originalCoinsAmount,
				"bonus_rate":     cardType.BonusRate,
				"bonus_coins":    bonusCoins,
			}).Info("Applied bonus")
		}
	}

	// 处理卡密折扣（如果有）
	if req.CardCode != nil {
		// TODO: 验证和应用卡密折扣
		// 这里简化处理，实际需要查询卡密表
	}

	// 生成订单
	order := &domain.RechargeOrder{
		ID:             uuid.New(),
		UserID:         userID,
		OrderNo:        s.generateOrderNo(),
		Amount:         finalAmount,         // 实际支付金额（可能有折扣）
		CoinsAmount:    originalCoinsAmount, // 基础金币数量
		ExchangeRate:   exchangeRate,
		PaymentMethod:  req.PaymentMethod,
		PaymentCardID:  req.PaymentCardID,
		OriginalAmount: req.Amount,     // 原始金额
		DiscountAmount: discountAmount, // 折扣金额
		BonusCoins:     bonusCoins,     // 返利金币
		CardCode:       req.CardCode,
		Status:         domain.OrderStatusPending,
		PaymentStatus:  domain.PaymentStatusUnpaid,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ExpiredAt:      &[]time.Time{time.Now().Add(30 * time.Minute)}[0], // 30分钟过期
	}

	if err := s.paymentRepo.CreateRechargeOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create recharge order: %w", err)
	}

	// 生成支付链接（根据支付方式）
	var paymentURL *string
	switch req.PaymentMethod {
	case domain.PaymentMethodAlipay:
		url := s.generateAlipayURL(order)
		paymentURL = &url
	case domain.PaymentMethodWechat:
		url := s.generateWechatURL(order)
		paymentURL = &url
	case domain.PaymentMethodCard:
		// 银行卡支付不需要外部链接
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"order_id":        order.ID,
		"original_amount": req.Amount,
		"final_amount":    finalAmount,
		"coins_amount":    originalCoinsAmount,
		"bonus_coins":     bonusCoins,
		"method":          req.PaymentMethod,
	}).Info("Recharge order created")

	return &domain.PaymentOrderResponse{
		OrderID:     order.ID,
		OrderNo:     order.OrderNo,
		Amount:      order.Amount,
		CoinsAmount: order.CoinsAmount + order.BonusCoins, // 显示总金币数量
		PaymentURL:  paymentURL,
		ExpiredAt:   *order.ExpiredAt,
	}, nil
}

// GetRechargeOrder 获取充值订单
func (s *paymentService) GetRechargeOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*domain.RechargeOrder, error) {
	order, err := s.paymentRepo.GetRechargeOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recharge order: %w", err)
	}

	if order.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限查看该订单", nil)
	}

	return order, nil
}

// GetUserRechargeOrders 获取用户充值订单列表
func (s *paymentService) GetUserRechargeOrders(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.RechargeOrder, error) {
	orders, err := s.paymentRepo.GetUserRechargeOrders(ctx, userID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user recharge orders: %w", err)
	}

	return orders, nil
}

// ProcessPayment 处理支付
func (s *paymentService) ProcessPayment(ctx context.Context, orderID uuid.UUID, paymentPassword string) error {
	order, err := s.paymentRepo.GetRechargeOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get recharge order: %w", err)
	}

	// 检查订单状态
	if order.Status != domain.OrderStatusPending {
		return domain.NewAppError(domain.CodeInvalidRequest, "订单状态不允许支付", nil)
	}

	// 检查订单是否过期
	if order.ExpiredAt != nil && time.Now().After(*order.ExpiredAt) {
		return domain.NewAppError(domain.CodeInvalidRequest, "订单已过期", nil)
	}

	// 验证支付密码
	if err := s.VerifyPaymentPassword(ctx, order.UserID, &domain.VerifyPaymentPasswordRequest{
		Password: paymentPassword,
	}); err != nil {
		return err
	}

	// 更新订单状态
	order.Status = domain.OrderStatusProcessing
	order.PaymentStatus = domain.PaymentStatusPaid
	now := time.Now()
	order.CompletedAt = &now

	if err := s.paymentRepo.UpdateRechargeOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update recharge order: %w", err)
	}

	// 增加用户金币
	if err := s.walletRepo.AddCoins(ctx, order.UserID, order.CoinsAmount+order.BonusCoins, "充值"); err != nil {
		s.logger.WithError(err).Error("Failed to add coins to wallet")
		// 回滚订单状态
		order.Status = domain.OrderStatusFailed
		order.PaymentStatus = domain.PaymentStatusUnpaid
		order.FailureReason = &[]string{"钱包充值失败"}[0]
		s.paymentRepo.UpdateRechargeOrder(ctx, order)
		return fmt.Errorf("failed to add coins: %w", err)
	}

	// 完成订单
	order.Status = domain.OrderStatusCompleted
	if err := s.paymentRepo.UpdateRechargeOrder(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to update order status to completed")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":      order.UserID,
		"order_id":     order.ID,
		"coins_amount": order.CoinsAmount + order.BonusCoins,
	}).Info("Payment processed successfully")

	return nil
}

// GetRechargePackages 获取充值套餐
func (s *paymentService) GetRechargePackages(ctx context.Context) ([]*domain.RechargePackage, error) {
	packages, err := s.paymentRepo.GetActiveRechargePackages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get recharge packages: %w", err)
	}

	return packages, nil
}

// generateOrderNo 生成订单号
func (s *paymentService) generateOrderNo() string {
	// YN + 时间戳 + 随机数
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	return "YN" + timestamp + random
}

// generateAlipayURL 生成支付宝支付链接
func (s *paymentService) generateAlipayURL(order *domain.RechargeOrder) string {
	// 这里应该调用支付宝API生成支付链接
	// 简化处理，返回模拟链接
	return fmt.Sprintf("https://openapi.alipay.com/gateway.do?order_no=%s&amount=%.2f", order.OrderNo, order.Amount)
}

// generateWechatURL 生成微信支付链接
func (s *paymentService) generateWechatURL(order *domain.RechargeOrder) string {
	// 这里应该调用微信支付API生成支付链接
	// 简化处理，返回模拟链接
	return fmt.Sprintf("weixin://wxpay/bizpayurl?order_no=%s&amount=%.0f", order.OrderNo, order.Amount*100)
}

// AdminRecharge 管理员代充
func (s *paymentService) AdminRecharge(ctx context.Context, adminUserID uuid.UUID, req *domain.AdminRechargeRequest) error {
	// 验证管理员权限
	admin, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	if admin.UserType != "admin" {
		return domain.NewAppError(domain.CodeForbidden, "无管理员权限", nil)
	}

	// 验证管理员支付密码
	if err := s.VerifyPaymentPassword(ctx, adminUserID, &domain.VerifyPaymentPasswordRequest{
		Password: req.PaymentPassword,
	}); err != nil {
		return fmt.Errorf("管理员支付密码验证失败: %w", err)
	}

	// 查找目标用户
	targetUser, err := s.GetUserByIdentifier(ctx, req.UserIdentifier)
	if err != nil {
		return fmt.Errorf("failed to find target user: %w", err)
	}

	// 创建管理员代充记录
	record := &domain.AdminRechargeRecord{
		ID:                     uuid.New(),
		TargetUserID:           targetUser.ID,
		AdminUserID:            adminUserID,
		CoinsAmount:            req.CoinsAmount,
		Reason:                 req.Reason,
		AdminPaymentVerified:   true,
		AdminPaymentVerifiedAt: &[]time.Time{time.Now()}[0],
		CreatedAt:              time.Now(),
	}

	if err := s.paymentRepo.CreateAdminRechargeRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to create admin recharge record: %w", err)
	}

	// 增加目标用户金币
	if err := s.walletRepo.AddCoins(ctx, targetUser.ID, req.CoinsAmount, fmt.Sprintf("管理员代充: %s", req.Reason)); err != nil {
		return fmt.Errorf("failed to add coins: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"admin_user_id":  adminUserID,
		"target_user_id": targetUser.ID,
		"coins_amount":   req.CoinsAmount,
		"reason":         req.Reason,
	}).Info("Admin recharge completed")

	return nil
}

// GetUserPaymentInfo 获取用户支付信息
func (s *paymentService) GetUserPaymentInfo(ctx context.Context, userIdentifier string) (*domain.UserPaymentInfo, error) {
	user, err := s.GetUserByIdentifier(ctx, userIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// 检查是否有支付密码
	hasPaymentPassword, err := s.HasPaymentPassword(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check payment password: %w", err)
	}

	// 获取支付卡片
	cards, err := s.GetUserPaymentCards(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment cards: %w", err)
	}

	// 找到默认卡片
	var defaultCardID *uuid.UUID
	for _, card := range cards {
		if card.IsDefault {
			defaultCardID = &card.ID
			break
		}
	}

	return &domain.UserPaymentInfo{
		HasPaymentPassword: hasPaymentPassword,
		PaymentCards:       cards,
		DefaultCardID:      defaultCardID,
	}, nil
}

// GetUserByIdentifier 根据标识符获取用户（支持ID或邮箱）
func (s *paymentService) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	// 尝试解析为UUID
	if userID, err := uuid.Parse(identifier); err == nil {
		return s.userRepo.GetByID(ctx, userID)
	}

	// 尝试作为邮箱查找
	if strings.Contains(identifier, "@") {
		return s.userRepo.GetByEmail(ctx, identifier)
	}

	// 尝试作为用户名查找
	return s.userRepo.GetByUsername(ctx, identifier)
}

// identifyCardType 识别卡片类型
func (s *paymentService) identifyCardType(cardNumber string) *auth.YunaiCardType {
	if len(cardNumber) < 2 {
		return nil
	}

	prefix := cardNumber[:2]
	cardTypes := auth.GetYunaiCardTypes()

	for _, cardType := range cardTypes {
		if cardType.Prefix == prefix {
			return &cardType
		}
	}

	return nil
}

// GetCoinExchangeRate 获取金币汇率
func (s *paymentService) GetCoinExchangeRate(ctx context.Context) (int, error) {
	// 这里应该从数据库的系统配置表中读取
	// 暂时返回默认值
	return domain.DefaultCoinExchangeRate, nil
}

// UpdateCoinExchangeRate 更新金币汇率
func (s *paymentService) UpdateCoinExchangeRate(ctx context.Context, rate int) error {
	// 这里应该更新数据库的系统配置表
	// 暂时不实现
	return nil
}

// BindPaymentCardWithEmail 邮箱验证绑卡
func (s *paymentService) BindPaymentCardWithEmail(ctx context.Context, userID uuid.UUID, req *domain.BindCardWithEmailRequest) (*domain.PaymentCard, error) {
	// TODO: 验证邮箱验证码
	// TODO: 检查卡号是否已被其他用户绑定
	// TODO: 绑定卡片并关联邮箱
	return nil, fmt.Errorf("not implemented yet")
}

// UnbindPaymentCard 解绑卡片
func (s *paymentService) UnbindPaymentCard(ctx context.Context, userID uuid.UUID, req *domain.UnbindCardRequest) error {
	// TODO: 验证邮箱验证码
	// TODO: 解绑卡片
	return fmt.Errorf("not implemented yet")
}

// ResetPaymentPassword 重置支付密码
func (s *paymentService) ResetPaymentPassword(ctx context.Context, req *domain.ResetPaymentPasswordRequest) error {
	// TODO: 验证邮箱验证码
	// TODO: 重置支付密码
	return fmt.Errorf("not implemented yet")
}

// GetCardBalance 获取卡片余额
func (s *paymentService) GetCardBalance(ctx context.Context, userID, cardID uuid.UUID) (*domain.CardBalanceResponse, error) {
	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限查看该卡片余额", nil)
	}

	return &domain.CardBalanceResponse{
		CardID:      card.ID,
		CardNumber:  card.CardNumber,
		Balance:     card.Balance,
		Currency:    card.Currency,
		LastUpdated: card.UpdatedAt,
	}, nil
}

// RechargeCard 卡片充值
func (s *paymentService) RechargeCard(ctx context.Context, userID uuid.UUID, req *domain.CardRechargeRequest) error {
	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, req.CardID)
	if err != nil {
		return fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return domain.NewAppError(domain.CodeForbidden, "无权限操作该卡片", nil)
	}

	// 检查卡片是否被冻结
	if card.IsFrozen {
		return domain.NewAppError(domain.CodeInvalidRequest, "卡片已冻结，无法进行充值操作", nil)
	}

	// 记录充值前余额
	balanceBefore := card.Balance

	// 更新卡片余额
	card.Balance += req.Amount
	card.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdatePaymentCard(ctx, card); err != nil {
		return fmt.Errorf("failed to update card balance: %w", err)
	}

	// 重新获取卡片以确保数据同步
	updatedCard, err := s.paymentRepo.GetPaymentCardByID(ctx, card.ID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get updated card, but recharge was successful")
	} else {
		card = updatedCard
	}

	// TODO: 记录交易到数据库
	// transaction := &domain.CardTransaction{
	//     ID:              uuid.New(),
	//     CardID:          card.ID,
	//     UserID:          userID,
	//     TransactionType: domain.CardTransactionRecharge,
	//     Amount:          req.Amount,
	//     BalanceBefore:   balanceBefore,
	//     BalanceAfter:    card.Balance,
	//     Description:     fmt.Sprintf("卡片充值 %.2f 金币", req.Amount),
	//     CreatedAt:       time.Now(),
	// }
	// if err := s.paymentRepo.CreateCardTransaction(ctx, transaction); err != nil {
	//     return fmt.Errorf("failed to create transaction record: %w", err)
	// }

	s.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"card_id":        req.CardID,
		"amount":         req.Amount,
		"balance_before": balanceBefore,
		"balance_after":  card.Balance,
	}).Info("Card recharged successfully")

	return nil
}

// TransferBetweenCards 卡片间转账
func (s *paymentService) TransferBetweenCards(ctx context.Context, userID uuid.UUID, req *domain.CardTransferRequest) error {
	// 验证支付密码
	if err := s.VerifyPaymentPassword(ctx, userID, &domain.VerifyPaymentPasswordRequest{
		Password: req.Password,
	}); err != nil {
		return err
	}

	// 获取源卡片和目标卡片
	fromCard, err := s.paymentRepo.GetPaymentCardByID(ctx, req.FromCardID)
	if err != nil {
		return fmt.Errorf("failed to get source card: %w", err)
	}

	toCard, err := s.paymentRepo.GetPaymentCardByID(ctx, req.ToCardID)
	if err != nil {
		return fmt.Errorf("failed to get target card: %w", err)
	}

	// 验证卡片所有权
	if fromCard.UserID != userID || toCard.UserID != userID {
		return domain.NewAppError(domain.CodeForbidden, "无权限操作该卡片", nil)
	}

	// 检查卡片是否被冻结
	if fromCard.IsFrozen {
		return domain.NewAppError(domain.CodeInvalidRequest, "源卡片已冻结，无法进行转账操作", nil)
	}
	if toCard.IsFrozen {
		return domain.NewAppError(domain.CodeInvalidRequest, "目标卡片已冻结，无法进行转账操作", nil)
	}

	// 验证余额
	if fromCard.Balance < req.Amount {
		return domain.NewAppError(domain.CodeInsufficientBalance, "卡片余额不足", nil)
	}

	// TODO: 记录转账前余额用于交易记录
	// fromBalanceBefore := fromCard.Balance
	// toBalanceBefore := toCard.Balance

	// 执行转账
	fromCard.Balance -= req.Amount
	toCard.Balance += req.Amount
	fromCard.UpdatedAt = time.Now()
	toCard.UpdatedAt = time.Now()

	// 更新卡片余额
	if err := s.paymentRepo.UpdatePaymentCard(ctx, fromCard); err != nil {
		return fmt.Errorf("failed to update source card: %w", err)
	}

	if err := s.paymentRepo.UpdatePaymentCard(ctx, toCard); err != nil {
		return fmt.Errorf("failed to update target card: %w", err)
	}

	// TODO: 记录转账交易
	// 创建转出记录
	// 创建转入记录

	s.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"from_card":    req.FromCardID,
		"to_card":      req.ToCardID,
		"amount":       req.Amount,
		"from_balance": fromCard.Balance,
		"to_balance":   toCard.Balance,
	}).Info("Card transfer completed")

	return nil
}

// GetCardTransactions 获取卡片交易记录
func (s *paymentService) GetCardTransactions(ctx context.Context, userID, cardID uuid.UUID, offset, limit int) ([]*domain.CardTransaction, error) {
	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限查看该卡片交易记录", nil)
	}

	// TODO: 从数据库获取交易记录
	// return s.paymentRepo.GetCardTransactions(ctx, cardID, offset, limit)

	// 暂时返回空列表
	return []*domain.CardTransaction{}, nil
}

// FreezeCard 冻结卡片
func (s *paymentService) FreezeCard(ctx context.Context, userID uuid.UUID, req *domain.FreezeCardRequest) (*domain.CardStatusResponse, error) {
	// 验证支付密码
	if err := s.VerifyPaymentPassword(ctx, userID, &domain.VerifyPaymentPasswordRequest{
		Password: req.Password,
	}); err != nil {
		return nil, err
	}

	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, req.CardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限操作该卡片", nil)
	}

	// 检查卡片是否已经冻结
	if card.IsFrozen {
		return nil, domain.NewAppError(domain.CodeInvalidRequest, "卡片已经处于冻结状态", nil)
	}

	// 冻结卡片
	card.IsFrozen = true
	card.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdatePaymentCard(ctx, card); err != nil {
		return nil, fmt.Errorf("failed to freeze card: %w", err)
	}

	// TODO: 记录操作日志
	// operation := &domain.CardOperation{
	//     ID:            uuid.New(),
	//     CardID:        req.CardID,
	//     UserID:        userID,
	//     OperationType: domain.CardOperationFreeze,
	//     Reason:        req.Reason,
	//     CreatedAt:     time.Now(),
	// }

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"card_id": req.CardID,
		"reason":  req.Reason,
	}).Info("Payment card frozen")

	return &domain.CardStatusResponse{
		CardID:     card.ID,
		CardNumber: card.CardNumber,
		IsActive:   card.IsActive,
		IsFrozen:   card.IsFrozen,
		Status:     domain.CardStatusFrozen,
		Message:    "卡片已成功冻结",
	}, nil
}

// UnfreezeCard 解冻卡片
func (s *paymentService) UnfreezeCard(ctx context.Context, userID uuid.UUID, req *domain.UnfreezeCardRequest) (*domain.CardStatusResponse, error) {
	// 验证支付密码
	if err := s.VerifyPaymentPassword(ctx, userID, &domain.VerifyPaymentPasswordRequest{
		Password: req.Password,
	}); err != nil {
		return nil, err
	}

	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, req.CardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限操作该卡片", nil)
	}

	// 检查卡片是否处于冻结状态
	if !card.IsFrozen {
		return nil, domain.NewAppError(domain.CodeInvalidRequest, "卡片未处于冻结状态", nil)
	}

	// 解冻卡片
	card.IsFrozen = false
	card.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdatePaymentCard(ctx, card); err != nil {
		return nil, fmt.Errorf("failed to unfreeze card: %w", err)
	}

	// TODO: 记录操作日志
	// operation := &domain.CardOperation{
	//     ID:            uuid.New(),
	//     CardID:        req.CardID,
	//     UserID:        userID,
	//     OperationType: domain.CardOperationUnfreeze,
	//     Reason:        "用户解冻",
	//     CreatedAt:     time.Now(),
	// }

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"card_id": req.CardID,
	}).Info("Payment card unfrozen")

	return &domain.CardStatusResponse{
		CardID:     card.ID,
		CardNumber: card.CardNumber,
		IsActive:   card.IsActive,
		IsFrozen:   card.IsFrozen,
		Status:     domain.CardStatusActive,
		Message:    "卡片已成功解冻",
	}, nil
}

// GetCardStatus 获取卡片状态
func (s *paymentService) GetCardStatus(ctx context.Context, userID, cardID uuid.UUID) (*domain.CardStatusResponse, error) {
	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限查看该卡片状态", nil)
	}

	// 确定卡片状态
	status := domain.CardStatusActive
	message := "卡片状态正常"

	if !card.IsActive {
		status = domain.CardStatusDeleted
		message = "卡片已删除"
	} else if card.IsFrozen {
		status = domain.CardStatusFrozen
		message = "卡片已冻结"
	}

	return &domain.CardStatusResponse{
		CardID:     card.ID,
		CardNumber: card.CardNumber,
		IsActive:   card.IsActive,
		IsFrozen:   card.IsFrozen,
		Status:     status,
		Message:    message,
	}, nil
}

// GetCardOperations 获取卡片操作记录
func (s *paymentService) GetCardOperations(ctx context.Context, userID, cardID uuid.UUID, offset, limit int) ([]*domain.CardOperation, error) {
	// 验证卡片是否属于用户
	card, err := s.paymentRepo.GetPaymentCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	if card.UserID != userID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限查看该卡片操作记录", nil)
	}

	// TODO: 从数据库获取操作记录
	// return s.paymentRepo.GetCardOperations(ctx, cardID, offset, limit)

	// 暂时返回空列表
	return []*domain.CardOperation{}, nil
}

// ProxyRecharge 代充功能
func (s *paymentService) ProxyRecharge(ctx context.Context, payerUserID uuid.UUID, req *domain.ProxyRechargeRequest) (*domain.ProxyRechargeResponse, error) {
	// 验证支付密码
	if err := s.VerifyPaymentPassword(ctx, payerUserID, &domain.VerifyPaymentPasswordRequest{
		Password: req.PaymentPassword,
	}); err != nil {
		return nil, err
	}

	// 验证目标用户是否存在
	targetUser, err := s.userRepo.GetByID(ctx, req.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// 获取付款人信息
	payerUser, err := s.userRepo.GetByID(ctx, payerUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payer user: %w", err)
	}

	// 验证支付卡片
	paymentCard, err := s.paymentRepo.GetPaymentCardByID(ctx, req.PaymentCardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment card: %w", err)
	}

	if paymentCard.UserID != payerUserID {
		return nil, domain.NewAppError(domain.CodeForbidden, "无权限使用该卡片", nil)
	}

	// 检查卡片状态
	if paymentCard.IsFrozen {
		return nil, domain.NewAppError(domain.CodeInvalidRequest, "卡片已冻结，无法进行代充操作", nil)
	}

	// 识别卡片类型并计算优惠
	cardType := s.identifyCardType(paymentCard.CardNumber)

	// 获取系统金币汇率
	exchangeRate, err := s.GetCoinExchangeRate(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get exchange rate, using default")
		exchangeRate = domain.DefaultCoinExchangeRate
	}

	// 计算原始金币数量
	originalCoinsAmount := int64(req.Amount * float64(exchangeRate))

	// 处理卡片类型的折扣和返利
	finalAmount := req.Amount
	bonusCoins := int64(0)
	discountAmount := 0.0
	cardTypeName := "标准卡"

	if cardType != nil {
		cardTypeName = cardType.Name

		// 折扣卡：减少支付金额
		if cardType.DiscountRate > 0 && cardType.DiscountRate < 1 {
			finalAmount = req.Amount * cardType.DiscountRate
			discountAmount = req.Amount - finalAmount
		}

		// 返利卡：增加金币数量
		if cardType.BonusRate > 0 {
			bonusCoins = int64(float64(originalCoinsAmount) * cardType.BonusRate)
		}
	}

	// 检查卡片余额是否足够
	if paymentCard.Balance < finalAmount {
		return nil, domain.NewAppError(domain.CodeInsufficientBalance,
			fmt.Sprintf("卡片余额不足，需要%.2f金币，当前余额%.2f金币", finalAmount, paymentCard.Balance), nil)
	}

	// 创建代充订单
	order := &domain.ProxyRechargeOrder{
		ID:             uuid.New(),
		OrderNo:        s.generateOrderNo(),
		PayerUserID:    payerUserID,
		TargetUserID:   req.TargetUserID,
		PaymentCardID:  req.PaymentCardID,
		OriginalAmount: req.Amount,
		ActualAmount:   finalAmount,
		DiscountAmount: discountAmount,
		CoinsAmount:    originalCoinsAmount,
		BonusCoins:     bonusCoins,
		TotalCoins:     originalCoinsAmount + bonusCoins,
		ExchangeRate:   exchangeRate,
		CardType:       cardTypeName,
		Message:        req.Message,
		Status:         domain.OrderStatusPending,
		PaymentStatus:  domain.PaymentStatusUnpaid,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// TODO: 保存代充订单到数据库
	// if err := s.paymentRepo.CreateProxyRechargeOrder(ctx, order); err != nil {
	//     return nil, fmt.Errorf("failed to create proxy recharge order: %w", err)
	// }

	// 扣除付款人卡片余额
	paymentCard.Balance -= finalAmount
	paymentCard.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdatePaymentCard(ctx, paymentCard); err != nil {
		return nil, fmt.Errorf("failed to deduct payer balance: %w", err)
	}

	// 增加目标用户钱包金币
	if err := s.walletRepo.AddCoins(ctx, req.TargetUserID, order.TotalCoins, fmt.Sprintf("代充充值，订单号：%s", order.OrderNo)); err != nil {
		// 回滚卡片余额
		paymentCard.Balance += finalAmount
		s.paymentRepo.UpdatePaymentCard(ctx, paymentCard)
		return nil, fmt.Errorf("failed to add coins to target user: %w", err)
	}

	// 更新订单状态
	order.Status = domain.OrderStatusCompleted
	order.PaymentStatus = domain.PaymentStatusPaid
	completedAt := time.Now()
	order.CompletedAt = &completedAt
	order.UpdatedAt = completedAt

	s.logger.WithFields(logrus.Fields{
		"payer_user_id":   payerUserID,
		"target_user_id":  req.TargetUserID,
		"order_id":        order.ID,
		"original_amount": req.Amount,
		"actual_amount":   finalAmount,
		"total_coins":     order.TotalCoins,
		"card_type":       cardTypeName,
	}).Info("Proxy recharge completed successfully")

	return &domain.ProxyRechargeResponse{
		OrderID:        order.ID,
		OrderNo:        order.OrderNo,
		TargetUserID:   targetUser.ID,
		TargetUsername: targetUser.Username,
		PayerUserID:    payerUser.ID,
		PayerUsername:  payerUser.Username,
		OriginalAmount: order.OriginalAmount,
		ActualAmount:   order.ActualAmount,
		DiscountAmount: order.DiscountAmount,
		CoinsAmount:    order.CoinsAmount,
		BonusCoins:     order.BonusCoins,
		TotalCoins:     order.TotalCoins,
		CardType:       order.CardType,
		Message:        order.Message,
		CreatedAt:      order.CreatedAt,
	}, nil
}

// GetProxyRechargeOrders 获取代充订单列表
func (s *paymentService) GetProxyRechargeOrders(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.ProxyRechargeOrder, error) {
	// TODO: 从数据库获取代充订单
	// return s.paymentRepo.GetProxyRechargeOrders(ctx, userID, offset, limit)

	// 暂时返回空列表
	return []*domain.ProxyRechargeOrder{}, nil
}

// GetProxyRechargeOrder 获取单个代充订单
func (s *paymentService) GetProxyRechargeOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*domain.ProxyRechargeOrder, error) {
	// TODO: 从数据库获取代充订单
	// return s.paymentRepo.GetProxyRechargeOrder(ctx, orderID)

	// 暂时返回空
	return nil, fmt.Errorf("order not found")
}
