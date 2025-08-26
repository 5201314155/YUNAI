package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// AIBillingService AI调用计费服务接口
type AIBillingService interface {
	// 预扣费检查
	PreCheckBalance(ctx context.Context, userID uuid.UUID, modelID uuid.UUID, req *domain.PricingCalculationRequest) (*domain.PreCheckResponse, error)

	// 执行扣费
	ChargeForAICall(ctx context.Context, userID uuid.UUID, req *domain.AICallChargeRequest) (*domain.AICallChargeResponse, error)

	// 退费（调用失败时）
	RefundAICall(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID, reason string) error

	// 获取用户AI调用消费记录
	GetAICallHistory(ctx context.Context, userID uuid.UUID, days int, page, limit int) (*domain.AICallHistoryResponse, error)
}

type aiBillingService struct {
	walletRepo   repository.WalletRepository
	modelService ModelService
	logger       *logrus.Logger
}

// NewAIBillingService 创建AI计费服务
func NewAIBillingService(walletRepo repository.WalletRepository, modelService ModelService, logger *logrus.Logger) AIBillingService {
	return &aiBillingService{
		walletRepo:   walletRepo,
		modelService: modelService,
		logger:       logger,
	}
}

// PreCheckBalance 预扣费检查
func (s *aiBillingService) PreCheckBalance(ctx context.Context, userID uuid.UUID, modelID uuid.UUID, req *domain.PricingCalculationRequest) (*domain.PreCheckResponse, error) {
	// 获取用户钱包
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user wallet: %w", err)
	}

	// 计算调用成本
	costResp, err := s.modelService.CalculatePrice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %w", err)
	}

	// 转换为金币（假设汇率为1美元=10金币）
	exchangeRate := float64(wallet.ExchangeRate)
	requiredCoins := costResp.UserCost * exchangeRate

	// 检查余额是否充足
	sufficient := wallet.Balance >= requiredCoins

	response := &domain.PreCheckResponse{
		UserID:         userID,
		ModelID:        modelID,
		RequiredCoins:  requiredCoins,
		CurrentBalance: wallet.Balance,
		Sufficient:     sufficient,
		CostDetails:    costResp,
		ExchangeRate:   exchangeRate,
	}

	if !sufficient {
		response.ErrorMessage = fmt.Sprintf("余额不足，需要 %.2f 金币，当前余额 %.2f 金币", requiredCoins, wallet.Balance)
	}

	return response, nil
}

// ChargeForAICall 执行AI调用扣费
func (s *aiBillingService) ChargeForAICall(ctx context.Context, userID uuid.UUID, req *domain.AICallChargeRequest) (*domain.AICallChargeResponse, error) {
	// 预检查余额
	preCheck, err := s.PreCheckBalance(ctx, userID, req.ModelID, &req.PricingCalculationRequest)
	if err != nil {
		return nil, err
	}

	if !preCheck.Sufficient {
		return &domain.AICallChargeResponse{
			Success:      false,
			ErrorMessage: preCheck.ErrorMessage,
			PreCheck:     preCheck,
		}, nil
	}

	// 准备交易描述
	transactionDescription := fmt.Sprintf("AI调用扣费 - %s", req.ModelName)

	// 扣除金币（转换为int64）
	coinsToDeduct := int64(preCheck.RequiredCoins)
	err = s.walletRepo.DeductCoins(ctx, userID, coinsToDeduct, transactionDescription)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct coins: %w", err)
	}

	// 记录使用统计
	err = s.modelService.RecordUsage(ctx, req.ModelID, &userID, true, req.TotalTokens, preCheck.CostDetails.UserCost)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to record model usage")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"model_id":      req.ModelID,
		"coins_charged": preCheck.RequiredCoins,
		"request_id":    req.RequestID,
	}).Info("AI call charged successfully")

	// 获取更新后的钱包余额
	updatedWallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get updated wallet balance")
	}

	transactionID := uuid.New()
	newBalance := preCheck.CurrentBalance - preCheck.RequiredCoins
	if updatedWallet != nil {
		newBalance = updatedWallet.Balance
	}

	return &domain.AICallChargeResponse{
		Success:       true,
		TransactionID: transactionID,
		CoinsCharged:  preCheck.RequiredCoins,
		NewBalance:    newBalance,
		PreCheck:      preCheck,
	}, nil
}

// RefundAICall 退费（调用失败时）
func (s *aiBillingService) RefundAICall(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID, reason string) error {
	// TODO: 实现退费逻辑
	// 1. 查找原始交易记录
	// 2. 创建退费交易
	// 3. 增加用户金币余额

	s.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": transactionID,
		"reason":         reason,
	}).Info("AI call refund requested")

	return nil
}

// GetAICallHistory 获取AI调用历史
func (s *aiBillingService) GetAICallHistory(ctx context.Context, userID uuid.UUID, days int, page, limit int) (*domain.AICallHistoryResponse, error) {
	// TODO: 实现调用历史查询
	// 从wallet_transactions表查询ai_call类型的交易记录

	return &domain.AICallHistoryResponse{
		Records: []domain.AICallRecord{},
		Total:   0,
		Page:    page,
		Limit:   limit,
	}, nil
}
