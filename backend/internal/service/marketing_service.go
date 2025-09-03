package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// MarketingService 营销服务接口
type MarketingService interface {
	// 营销活动管理
	CreateCampaign(ctx context.Context, campaign *domain.MarketingCampaign) error
	UpdateCampaign(ctx context.Context, campaign *domain.MarketingCampaign) error
	DeleteCampaign(ctx context.Context, campaignID uuid.UUID) error
	GetCampaign(ctx context.Context, campaignID uuid.UUID) (*domain.MarketingCampaign, error)
	ListCampaigns(ctx context.Context, status string) ([]*domain.MarketingCampaign, error)
	GetActiveCampaigns(ctx context.Context) ([]*domain.MarketingCampaign, error)

	// 用户推荐
	GetRecommendedUsers(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.User, error)
	GetRecommendedCharacters(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Character, error)
	GetPersonalizedContent(ctx context.Context, userID uuid.UUID) (*domain.PersonalizedContent, error)

	// 邀请奖励系统
	CreateReferralCode(ctx context.Context, userID uuid.UUID) (*domain.ReferralCode, error)
	ProcessReferral(ctx context.Context, referralCode string, refereeID uuid.UUID) (*domain.ReferralReward, error)
	GetReferralStats(ctx context.Context, userID uuid.UUID) (*domain.ReferralStats, error)
	ProcessReferralReward(ctx context.Context, referrerID, refereeID uuid.UUID) error

	// A/B测试框架
	CreateABTest(ctx context.Context, test *domain.ABTest) error
	GetUserABTestGroup(ctx context.Context, userID uuid.UUID, testID string) (string, error)
	RecordABTestEvent(ctx context.Context, userID uuid.UUID, testID, event string) error
	GetABTestResults(ctx context.Context, testID string) (*domain.ABTestResults, error)

	// 用户分群
	CreateUserSegment(ctx context.Context, segment *domain.UserSegment) error
	GetUsersInSegment(ctx context.Context, segmentID uuid.UUID) ([]*domain.User, error)
	UpdateUserSegments(ctx context.Context) error

	// 营销统计
	GetMarketingStats(ctx context.Context) (*domain.MarketingStats, error)
	GetCampaignPerformance(ctx context.Context, campaignID uuid.UUID) (*domain.CampaignPerformance, error)
}

// marketingService 营销服务实现
type marketingService struct {
	marketingRepo repository.MarketingRepository
	userRepo      repository.UserRepository
	characterRepo repository.CharacterRepository
	redis         *redis.Client
	logger        *logrus.Logger

	// 推荐算法配置
	recommendationConfig *RecommendationConfig
}

// RecommendationConfig 推荐算法配置
type RecommendationConfig struct {
	UserSimilarityWeight      float64 `json:"user_similarity_weight"`
	ContentSimilarityWeight   float64 `json:"content_similarity_weight"`
	PopularityWeight          float64 `json:"popularity_weight"`
	RecencyWeight             float64 `json:"recency_weight"`
	DiversityFactor           float64 `json:"diversity_factor"`
	MinRecommendationScore    float64 `json:"min_recommendation_score"`
	MaxRecommendationsPerType int     `json:"max_recommendations_per_type"`
}

// NewMarketingService 创建营销服务
func NewMarketingService(
	marketingRepo repository.MarketingRepository,
	userRepo repository.UserRepository,
	characterRepo repository.CharacterRepository,
	redis *redis.Client,
	logger *logrus.Logger,
) MarketingService {
	config := &RecommendationConfig{
		UserSimilarityWeight:      0.3,
		ContentSimilarityWeight:   0.25,
		PopularityWeight:          0.2,
		RecencyWeight:             0.15,
		DiversityFactor:           0.1,
		MinRecommendationScore:    0.3,
		MaxRecommendationsPerType: 20,
	}

	return &marketingService{
		marketingRepo:        marketingRepo,
		userRepo:             userRepo,
		characterRepo:        characterRepo,
		redis:                redis,
		logger:               logger,
		recommendationConfig: config,
	}
}

// CreateCampaign 创建营销活动
func (s *marketingService) CreateCampaign(ctx context.Context, campaign *domain.MarketingCampaign) error {
	campaign.ID = uuid.New()
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()

	// 验证活动配置
	if err := s.validateCampaign(campaign); err != nil {
		return fmt.Errorf("invalid campaign configuration: %w", err)
	}

	if err := s.marketingRepo.CreateCampaign(ctx, campaign); err != nil {
		return fmt.Errorf("failed to create campaign: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"campaign_id":   campaign.ID,
		"campaign_name": campaign.Name,
		"campaign_type": campaign.Type,
		"start_time":    campaign.StartTime,
		"end_time":      campaign.EndTime,
	}).Info("Marketing campaign created")

	return nil
}

// UpdateCampaign 更新营销活动
func (s *marketingService) UpdateCampaign(ctx context.Context, campaign *domain.MarketingCampaign) error {
	campaign.UpdatedAt = time.Now()

	if err := s.validateCampaign(campaign); err != nil {
		return fmt.Errorf("invalid campaign configuration: %w", err)
	}

	if err := s.marketingRepo.UpdateCampaign(ctx, campaign); err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	s.logger.WithField("campaign_id", campaign.ID).Info("Marketing campaign updated")
	return nil
}

// DeleteCampaign 删除营销活动
func (s *marketingService) DeleteCampaign(ctx context.Context, campaignID uuid.UUID) error {
	if err := s.marketingRepo.DeleteCampaign(ctx, campaignID); err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	s.logger.WithField("campaign_id", campaignID).Info("Marketing campaign deleted")
	return nil
}

// GetCampaign 获取营销活动
func (s *marketingService) GetCampaign(ctx context.Context, campaignID uuid.UUID) (*domain.MarketingCampaign, error) {
	campaign, err := s.marketingRepo.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}
	return campaign, nil
}

// ListCampaigns 列出营销活动
func (s *marketingService) ListCampaigns(ctx context.Context, status string) ([]*domain.MarketingCampaign, error) {
	campaigns, err := s.marketingRepo.ListCampaigns(ctx, status, 100, 0) // 默认限制100条
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}
	return campaigns, nil
}

// GetActiveCampaigns 获取活跃的营销活动
func (s *marketingService) GetActiveCampaigns(ctx context.Context) ([]*domain.MarketingCampaign, error) {
	// 使用ListCampaigns获取活跃状态的活动
	campaigns, err := s.marketingRepo.ListCampaigns(ctx, "active", 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get active campaigns: %w", err)
	}
	return campaigns, nil
}

// GetRecommendedUsers 获取推荐用户
func (s *marketingService) GetRecommendedUsers(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.User, error) {
	// 获取用户信息
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取用户的兴趣标签和行为数据
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user profile, using basic recommendation")
		return s.getBasicUserRecommendations(ctx, userID, limit)
	}

	// 获取候选用户
	candidates, err := s.userRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate users: %w", err)
	}

	// 过滤掉自己
	var filteredCandidates []*domain.User
	for _, candidate := range candidates {
		if candidate.ID != userID {
			filteredCandidates = append(filteredCandidates, candidate)
		}
	}

	// 计算推荐分数
	recommendations := s.calculateUserRecommendationScores(userProfile, filteredCandidates)

	// 排序并返回前N个
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	var result []*domain.User
	for i, rec := range recommendations {
		if i >= limit || rec.Score < s.recommendationConfig.MinRecommendationScore {
			break
		}
		result = append(result, rec.User)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":               userID,
		"candidates_count":      len(filteredCandidates),
		"recommendations_count": len(result),
	}).Debug("User recommendations generated")

	return result, nil
}

// GetRecommendedCharacters 获取推荐角色
func (s *marketingService) GetRecommendedCharacters(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Character, error) {
	// 获取用户偏好
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get user profile, using basic recommendation")
		return s.getBasicCharacterRecommendations(ctx, userID, limit)
	}

	// 获取候选角色 - 使用现有的方法
	publicVisibility := "public"
	candidates, _, err := s.characterRepo.ListCharacters(ctx, &domain.CharacterListRequest{
		Visibility: &publicVisibility,
		Page:       1,
		Limit:      100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate characters: %w", err)
	}

	// 过滤掉用户已有的角色 - 使用现有的方法
	userCharacters, _, _ := s.characterRepo.ListCharacters(ctx, &domain.CharacterListRequest{
		UserID: &userID,
		Page:   1,
		Limit:  100,
	})
	userCharacterIDs := make(map[uuid.UUID]bool)
	for _, char := range userCharacters {
		userCharacterIDs[char.ID] = true
	}

	var filteredCandidates []*domain.Character
	for _, candidate := range candidates {
		if !userCharacterIDs[candidate.ID] {
			filteredCandidates = append(filteredCandidates, candidate)
		}
	}

	// 计算推荐分数
	recommendations := s.calculateCharacterRecommendationScores(userProfile, filteredCandidates)

	// 排序并返回前N个
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	var result []*domain.Character
	for i, rec := range recommendations {
		if i >= limit || rec.Score < s.recommendationConfig.MinRecommendationScore {
			break
		}
		result = append(result, rec.Character)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":               userID,
		"candidates_count":      len(filteredCandidates),
		"recommendations_count": len(result),
	}).Debug("Character recommendations generated")

	return result, nil
}

// GetPersonalizedContent 获取个性化内容
func (s *marketingService) GetPersonalizedContent(ctx context.Context, userID uuid.UUID) (*domain.PersonalizedContent, error) {
	_, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	content := &domain.PersonalizedContent{
		UserID:    userID,
		Timestamp: time.Now(),
	}

	// 获取推荐用户
	recommendedUsers, err := s.GetRecommendedUsers(ctx, userID, 5)
	if err == nil {
		content.RecommendedUsers = recommendedUsers
	}

	// 获取推荐角色
	recommendedCharacters, err := s.GetRecommendedCharacters(ctx, userID, 10)
	if err == nil {
		content.RecommendedCharacters = recommendedCharacters
	}

	// 获取活跃营销活动
	activeCampaigns, err := s.GetActiveCampaigns(ctx)
	if err == nil {
		content.ActiveCampaigns = activeCampaigns
	}

	return content, nil
}

// CreateReferralCode 创建邀请码
func (s *marketingService) CreateReferralCode(ctx context.Context, userID uuid.UUID) (*domain.ReferralCode, error) {
	// 检查用户是否已有邀请码
	existingCode, err := s.marketingRepo.GetReferralCodeByUserID(ctx, userID)
	if err == nil && existingCode != nil {
		return existingCode, nil
	}

	// 生成唯一邀请码
	code := s.generateReferralCode()

	referralCode := &domain.ReferralCode{
		ID:        uuid.New(),
		UserID:    userID,
		Code:      code,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.marketingRepo.CreateReferralCode(ctx, referralCode); err != nil {
		return nil, fmt.Errorf("failed to create referral code: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"referral_code": code,
	}).Info("Referral code created")

	return referralCode, nil
}

// ProcessReferral 处理邀请
func (s *marketingService) ProcessReferral(ctx context.Context, referralCode string, refereeID uuid.UUID) (*domain.ReferralReward, error) {
	// 获取邀请码信息
	refCode, err := s.marketingRepo.GetReferralCode(ctx, referralCode)
	if err != nil {
		return nil, fmt.Errorf("invalid referral code: %w", err)
	}

	if !refCode.IsActive {
		return nil, fmt.Errorf("referral code is inactive")
	}

	if refCode.UserID == refereeID {
		return nil, fmt.Errorf("cannot refer yourself")
	}

	// 检查是否已经被邀请过 - 简化实现，暂时跳过检查
	// TODO: 实现检查用户是否已被邀请的逻辑

	// 创建邀请记录
	referral := &domain.Referral{
		ID:           uuid.New(),
		ReferrerID:   refCode.UserID,
		RefereeID:    refereeID,
		ReferralCode: referralCode,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.marketingRepo.CreateReferral(ctx, referral); err != nil {
		return nil, fmt.Errorf("failed to create referral: %w", err)
	}

	// 处理邀请奖励
	err = s.ProcessReferralReward(ctx, refCode.UserID, refereeID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process referral reward")
	}

	s.logger.WithFields(logrus.Fields{
		"referrer_id":   refCode.UserID,
		"referee_id":    refereeID,
		"referral_code": referralCode,
	}).Info("Referral processed")

	return &domain.ReferralReward{
		ReferralID: referral.ID,
		ReferrerID: refCode.UserID,
		RefereeID:  refereeID,
		RewardType: "coins",
		Amount:     100, // 默认奖励100金币
		Status:     "pending",
		CreatedAt:  time.Now(),
	}, nil
}

// GetReferralStats 获取邀请统计
func (s *marketingService) GetReferralStats(ctx context.Context, userID uuid.UUID) (*domain.ReferralStats, error) {
	stats, err := s.marketingRepo.GetReferralStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get referral stats: %w", err)
	}
	return stats, nil
}

// ProcessReferralReward 处理邀请奖励
func (s *marketingService) ProcessReferralReward(ctx context.Context, referrerID, refereeID uuid.UUID) error {
	// 给邀请者奖励
	referrerReward := &domain.ReferralReward{
		ID:         uuid.New(),
		ReferrerID: referrerID,
		RefereeID:  refereeID,
		RewardType: "coins",
		Amount:     100, // 邀请者获得100金币
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.marketingRepo.CreateReferralReward(ctx, referrerReward); err != nil {
		return fmt.Errorf("failed to create referrer reward: %w", err)
	}

	// 给被邀请者奖励
	refereeReward := &domain.ReferralReward{
		ID:         uuid.New(),
		ReferrerID: referrerID,
		RefereeID:  refereeID,
		RewardType: "coins",
		Amount:     50, // 被邀请者获得50金币
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.marketingRepo.CreateReferralReward(ctx, refereeReward); err != nil {
		return fmt.Errorf("failed to create referee reward: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"referrer_id":     referrerID,
		"referee_id":      refereeID,
		"referrer_reward": referrerReward.Amount,
		"referee_reward":  refereeReward.Amount,
	}).Info("Referral rewards processed")

	return nil
}

// validateCampaign 验证营销活动配置
func (s *marketingService) validateCampaign(campaign *domain.MarketingCampaign) error {
	if campaign.Name == "" {
		return fmt.Errorf("campaign name is required")
	}

	if campaign.StartTime.After(campaign.EndTime) {
		return fmt.Errorf("start time must be before end time")
	}

	if campaign.Budget < 0 {
		return fmt.Errorf("budget cannot be negative")
	}

	return nil
}

// A/B测试相关方法

// CreateABTest 创建A/B测试
func (s *marketingService) CreateABTest(ctx context.Context, test *domain.ABTest) error {
	test.ID = uuid.New()
	test.CreatedAt = time.Now()
	test.UpdatedAt = time.Now()

	if err := s.marketingRepo.CreateABTest(ctx, test); err != nil {
		return fmt.Errorf("failed to create AB test: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"test_id":   test.ID,
		"test_name": test.Name,
	}).Info("A/B test created")

	return nil
}

// GetUserABTestGroup 获取用户的A/B测试分组
func (s *marketingService) GetUserABTestGroup(ctx context.Context, userID uuid.UUID, testID string) (string, error) {
	// 先从缓存中查找
	cacheKey := fmt.Sprintf("ab_test:%s:user:%s", testID, userID.String())
	group, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return group, nil
	}

	// 从数据库查找
	assignment, err := s.marketingRepo.GetABTestAssignment(ctx, testID, userID)
	if err == nil {
		// 缓存结果
		s.redis.Set(ctx, cacheKey, assignment.Group, time.Hour*24)
		return assignment.Group, nil
	}

	// 如果没有分配，进行新分配
	testUUID, _ := uuid.Parse(testID)
	test, err := s.marketingRepo.GetABTest(ctx, testUUID)
	if err != nil {
		return "", fmt.Errorf("failed to get AB test: %w", err)
	}

	if !test.IsActive {
		return "control", nil // 测试不活跃时返回对照组
	}

	// 基于用户ID进行一致性哈希分组
	group = s.assignUserToABTestGroup(userID, test.Groups)

	// 保存分配结果
	assignment = &domain.ABTestAssignment{
		ID:        uuid.New(),
		TestID:    testID,
		UserID:    userID,
		Group:     group,
		CreatedAt: time.Now(),
	}

	if err := s.marketingRepo.CreateABTestAssignment(ctx, assignment); err != nil {
		s.logger.WithError(err).Error("Failed to save AB test assignment")
	}

	// 缓存结果
	s.redis.Set(ctx, cacheKey, group, time.Hour*24)

	return group, nil
}

// RecordABTestEvent 记录A/B测试事件
func (s *marketingService) RecordABTestEvent(ctx context.Context, userID uuid.UUID, testID, event string) error {
	group, err := s.GetUserABTestGroup(ctx, userID, testID)
	if err != nil {
		return fmt.Errorf("failed to get user AB test group: %w", err)
	}

	eventRecord := &domain.ABTestEvent{
		ID:        uuid.New(),
		TestID:    testID,
		UserID:    userID,
		Group:     group,
		Event:     event,
		Timestamp: time.Now(),
	}

	if err := s.marketingRepo.CreateABTestEvent(ctx, eventRecord); err != nil {
		return fmt.Errorf("failed to record AB test event: %w", err)
	}

	return nil
}

// GetABTestResults 获取A/B测试结果
func (s *marketingService) GetABTestResults(ctx context.Context, testID string) (*domain.ABTestResults, error) {
	results, err := s.marketingRepo.GetABTestResults(ctx, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AB test results: %w", err)
	}
	return results, nil
}

// CreateUserSegment 创建用户分群
func (s *marketingService) CreateUserSegment(ctx context.Context, segment *domain.UserSegment) error {
	segment.ID = uuid.New()
	segment.CreatedAt = time.Now()
	segment.UpdatedAt = time.Now()

	if err := s.marketingRepo.CreateUserSegment(ctx, segment); err != nil {
		return fmt.Errorf("failed to create user segment: %w", err)
	}

	// 异步更新分群用户
	go func() {
		if err := s.updateSegmentUsers(context.Background(), segment); err != nil {
			s.logger.WithError(err).WithField("segment_id", segment.ID).Error("Failed to update segment users")
		}
	}()

	return nil
}

// GetUsersInSegment 获取分群中的用户
func (s *marketingService) GetUsersInSegment(ctx context.Context, segmentID uuid.UUID) ([]*domain.User, error) {
	// TODO: 实现获取分群用户的逻辑
	// 暂时返回空列表
	return []*domain.User{}, nil
}

// UpdateUserSegments 更新用户分群
func (s *marketingService) UpdateUserSegments(ctx context.Context) error {
	// 获取活跃的用户分群
	segments, err := s.marketingRepo.ListUserSegments(ctx)
	if err != nil {
		return fmt.Errorf("failed to get segments: %w", err)
	}

	for _, segment := range segments {
		if err := s.updateSegmentUsers(ctx, segment); err != nil {
			s.logger.WithError(err).WithField("segment_id", segment.ID).Error("Failed to update segment users")
		}
	}

	return nil
}

// GetMarketingStats 获取营销统计
func (s *marketingService) GetMarketingStats(ctx context.Context) (*domain.MarketingStats, error) {
	stats, err := s.marketingRepo.GetMarketingStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get marketing stats: %w", err)
	}
	return stats, nil
}

// GetCampaignPerformance 获取活动表现
func (s *marketingService) GetCampaignPerformance(ctx context.Context, campaignID uuid.UUID) (*domain.CampaignPerformance, error) {
	performance, err := s.marketingRepo.GetCampaignPerformance(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign performance: %w", err)
	}
	return performance, nil
}

// 辅助方法

// generateReferralCode 生成邀请码
func (s *marketingService) generateReferralCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// assignUserToABTestGroup 分配用户到A/B测试分组
func (s *marketingService) assignUserToABTestGroup(userID uuid.UUID, groups []string) string {
	if len(groups) == 0 {
		return "control"
	}

	// 使用用户ID的哈希值进行一致性分组
	hash := 0
	for _, b := range userID.String() {
		hash = hash*31 + int(b)
	}

	groupIndex := hash % len(groups)
	if groupIndex < 0 {
		groupIndex = -groupIndex
	}

	return groups[groupIndex]
}

// getUserProfile 获取用户画像
func (s *marketingService) getUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	// 从缓存获取
	cacheKey := fmt.Sprintf("user_profile:%s", userID.String())
	profileData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var profile UserProfile
		if json.Unmarshal([]byte(profileData), &profile) == nil {
			return &profile, nil
		}
	}

	// 从数据库构建用户画像
	profile := &UserProfile{
		UserID:    userID,
		UpdatedAt: time.Now(),
	}

	// 获取用户基本信息
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 设置默认值，因为User结构体中没有这些字段
	profile.Age = 25             // 默认年龄
	profile.Gender = "unknown"   // 默认性别
	profile.Location = "unknown" // 默认位置

	// 获取用户兴趣标签（基于角色和聊天内容）
	profile.InterestTags = s.extractUserInterests(ctx, userID)

	// 获取活跃度指标
	profile.ActivityLevel = s.calculateActivityLevel(ctx, userID)

	// 缓存用户画像
	if profileJSON, err := json.Marshal(profile); err == nil {
		s.redis.Set(ctx, cacheKey, profileJSON, time.Hour*6)
	}

	return profile, nil
}

// getBasicUserRecommendations 获取基础用户推荐
func (s *marketingService) getBasicUserRecommendations(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.User, error) {
	// 简单的基于活跃度的推荐 - 暂时返回空列表
	// TODO: 实现基于活跃度的用户推荐算法
	return []*domain.User{}, nil
}

// getBasicCharacterRecommendations 获取基础角色推荐
func (s *marketingService) getBasicCharacterRecommendations(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Character, error) {
	// 简单的基于热度的推荐 - 暂时返回空列表
	// TODO: 实现基于热度的角色推荐算法
	return []*domain.Character{}, nil
}

// calculateUserRecommendationScores 计算用户推荐分数
func (s *marketingService) calculateUserRecommendationScores(userProfile *UserProfile, candidates []*domain.User) []*UserRecommendation {
	var recommendations []*UserRecommendation

	for _, candidate := range candidates {
		score := s.calculateUserSimilarityScore(userProfile, candidate)
		if score >= s.recommendationConfig.MinRecommendationScore {
			recommendations = append(recommendations, &UserRecommendation{
				User:  candidate,
				Score: score,
			})
		}
	}

	return recommendations
}

// calculateCharacterRecommendationScores 计算角色推荐分数
func (s *marketingService) calculateCharacterRecommendationScores(userProfile *UserProfile, candidates []*domain.Character) []*CharacterRecommendation {
	var recommendations []*CharacterRecommendation

	for _, candidate := range candidates {
		score := s.calculateCharacterSimilarityScore(userProfile, candidate)
		if score >= s.recommendationConfig.MinRecommendationScore {
			recommendations = append(recommendations, &CharacterRecommendation{
				Character: candidate,
				Score:     score,
			})
		}
	}

	return recommendations
}

// calculateUserSimilarityScore 计算用户相似度分数
func (s *marketingService) calculateUserSimilarityScore(userProfile *UserProfile, candidate *domain.User) float64 {
	score := 0.0

	// 年龄相似度 - 暂时跳过，因为User结构体没有Birthday字段
	// TODO: 实现年龄相似度计算

	// 地理位置相似度 - 暂时跳过，因为User结构体没有Location字段
	// TODO: 实现地理位置相似度计算

	// 活跃度相似度
	activityScore := 1.0 - math.Abs(userProfile.ActivityLevel-0.7)/1.0 // 假设候选用户活跃度为0.7
	score += activityScore * 0.3

	// 兴趣标签相似度
	interestScore := s.calculateInterestSimilarity(userProfile.InterestTags, []string{"AI", "聊天", "社交"}) // 简化处理
	score += interestScore * 0.2

	return math.Min(score, 1.0)
}

// calculateCharacterSimilarityScore 计算角色相似度分数
func (s *marketingService) calculateCharacterSimilarityScore(userProfile *UserProfile, candidate *domain.Character) float64 {
	score := 0.0

	// 基于用户兴趣标签匹配角色标签
	if len(userProfile.InterestTags) > 0 {
		characterTags := s.extractCharacterTags(candidate)
		interestScore := s.calculateInterestSimilarity(userProfile.InterestTags, characterTags)
		score += interestScore * 0.4
	}

	// 角色热度分数
	popularityScore := math.Min(float64(candidate.ChatCount)/1000.0, 1.0) // 1000次聊天为满分
	score += popularityScore * s.recommendationConfig.PopularityWeight

	// 角色新鲜度分数
	daysSinceCreated := time.Since(candidate.CreatedAt).Hours() / 24
	recencyScore := math.Max(0, 1.0-daysSinceCreated/30.0) // 30天内为新角色
	score += recencyScore * s.recommendationConfig.RecencyWeight

	return math.Min(score, 1.0)
}

// updateSegmentUsers 更新分群用户
func (s *marketingService) updateSegmentUsers(ctx context.Context, segment *domain.UserSegment) error {
	// 根据分群条件查询用户
	users, err := s.queryUsersBySegmentCriteria(ctx, segment.Criteria)
	if err != nil {
		return fmt.Errorf("failed to query users by segment criteria: %w", err)
	}

	// 更新分群用户关系 - 暂时跳过
	// TODO: 实现更新分群用户关系的逻辑

	s.logger.WithFields(logrus.Fields{
		"segment_id":   segment.ID,
		"segment_name": segment.Name,
		"user_count":   len(users),
	}).Info("User segment updated")

	return nil
}

// queryUsersBySegmentCriteria 根据分群条件查询用户
func (s *marketingService) queryUsersBySegmentCriteria(ctx context.Context, criteria map[string]interface{}) ([]*domain.User, error) {
	// 简化实现，实际应该根据复杂条件查询
	users, err := s.userRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	var result []*domain.User
	for _, user := range users {
		if s.matchesSegmentCriteria(user, criteria) {
			result = append(result, user)
		}
	}

	return result, nil
}

// matchesSegmentCriteria 检查用户是否匹配分群条件
func (s *marketingService) matchesSegmentCriteria(user *domain.User, criteria map[string]interface{}) bool {
	// 简化实现，实际应该支持复杂的条件匹配
	// 暂时跳过年龄、性别、位置检查，因为User结构体中没有这些字段
	// TODO: 实现完整的分群条件匹配逻辑

	// 检查用户类型
	if userType, ok := criteria["user_type"].(string); ok {
		if string(user.UserType) != userType {
			return false
		}
	}

	// 检查是否活跃
	if isActive, ok := criteria["is_active"].(bool); ok {
		if user.IsActive != isActive {
			return false
		}
	}

	return true
}

// calculateAge 计算年龄
func (s *marketingService) calculateAge(birthday *time.Time) int {
	if birthday == nil {
		return 0
	}

	now := time.Now()
	age := now.Year() - birthday.Year()

	if now.YearDay() < birthday.YearDay() {
		age--
	}

	return age
}

// extractUserInterests 提取用户兴趣标签
func (s *marketingService) extractUserInterests(ctx context.Context, userID uuid.UUID) []string {
	// 简化实现，实际应该基于用户行为分析
	interests := []string{"AI", "聊天", "社交", "娱乐"}

	// 可以基于用户的角色、聊天内容等进行分析
	// 这里返回默认兴趣标签
	return interests
}

// calculateActivityLevel 计算活跃度
func (s *marketingService) calculateActivityLevel(ctx context.Context, userID uuid.UUID) float64 {
	// 简化实现，实际应该基于用户行为数据计算
	// 可以考虑登录频率、聊天次数、创建角色数等指标
	return 0.8 // 返回默认活跃度
}

// extractCharacterTags 提取角色标签
func (s *marketingService) extractCharacterTags(character *domain.Character) []string {
	// 简化实现，实际应该基于角色描述、设定等提取标签
	tags := []string{}

	// 可以基于角色名称、描述、设定等提取关键词
	if character.Name != "" {
		tags = append(tags, "角色")
	}

	return tags
}

// calculateInterestSimilarity 计算兴趣相似度
func (s *marketingService) calculateInterestSimilarity(interests1, interests2 []string) float64 {
	if len(interests1) == 0 || len(interests2) == 0 {
		return 0.0
	}

	// 计算交集
	intersection := 0
	interestSet := make(map[string]bool)
	for _, interest := range interests1 {
		interestSet[interest] = true
	}

	for _, interest := range interests2 {
		if interestSet[interest] {
			intersection++
		}
	}

	// 计算Jaccard相似度
	union := len(interests1) + len(interests2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// 辅助结构体

// UserProfile 用户画像
type UserProfile struct {
	UserID        uuid.UUID `json:"user_id"`
	Age           int       `json:"age"`
	Gender        string    `json:"gender"`
	Location      string    `json:"location"`
	InterestTags  []string  `json:"interest_tags"`
	ActivityLevel float64   `json:"activity_level"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserRecommendation 用户推荐
type UserRecommendation struct {
	User  *domain.User `json:"user"`
	Score float64      `json:"score"`
}

// CharacterRecommendation 角色推荐
type CharacterRecommendation struct {
	Character *domain.Character `json:"character"`
	Score     float64           `json:"score"`
}
