package domain

import (
	"time"

	"github.com/google/uuid"
)

// MarketingCampaign 营销活动
type MarketingCampaign struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string                 `json:"name" gorm:"not null"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" gorm:"not null"` // email, push, banner, discount
	Status      string                 `json:"status" gorm:"default:'draft'"` // draft, active, paused, completed
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Budget      float64                `json:"budget" gorm:"default:0"`
	Spent       float64                `json:"spent" gorm:"default:0"`
	TargetUsers int                    `json:"target_users" gorm:"default:0"`
	ReachedUsers int                   `json:"reached_users" gorm:"default:0"`
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ReferralCode 邀请码
type ReferralCode struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Code      string    `json:"code" gorm:"unique;not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	UsedCount int       `json:"used_count" gorm:"default:0"`
	MaxUses   int       `json:"max_uses" gorm:"default:0"` // 0表示无限制
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Referral 邀请记录
type Referral struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ReferrerID   uuid.UUID `json:"referrer_id" gorm:"type:uuid;not null"`
	RefereeID    uuid.UUID `json:"referee_id" gorm:"type:uuid;not null"`
	ReferralCode string    `json:"referral_code" gorm:"not null"`
	Status       string    `json:"status" gorm:"default:'pending'"` // pending, completed, failed
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReferralReward 邀请奖励
type ReferralReward struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ReferralID uuid.UUID `json:"referral_id" gorm:"type:uuid"`
	ReferrerID uuid.UUID `json:"referrer_id" gorm:"type:uuid;not null"`
	RefereeID  uuid.UUID `json:"referee_id" gorm:"type:uuid;not null"`
	RewardType string    `json:"reward_type" gorm:"not null"` // coins, vip_days, credits
	Amount     int       `json:"amount" gorm:"not null"`
	Status     string    `json:"status" gorm:"default:'pending'"` // pending, completed, failed
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ReferralStats 邀请统计
type ReferralStats struct {
	UserID           uuid.UUID `json:"user_id"`
	TotalReferrals   int       `json:"total_referrals"`
	SuccessReferrals int       `json:"success_referrals"`
	TotalRewards     int       `json:"total_rewards"`
	PendingRewards   int       `json:"pending_rewards"`
}

// ABTest A/B测试
type ABTest struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Groups      []string  `json:"groups" gorm:"type:jsonb"` // ["control", "variant_a", "variant_b"]
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ABTestAssignment A/B测试分配
type ABTestAssignment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TestID    string    `json:"test_id" gorm:"not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Group     string    `json:"group" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// ABTestEvent A/B测试事件
type ABTestEvent struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TestID    string    `json:"test_id" gorm:"not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Group     string    `json:"group" gorm:"not null"`
	Event     string    `json:"event" gorm:"not null"`
	Timestamp time.Time `json:"timestamp"`
}

// ABTestResults A/B测试结果
type ABTestResults struct {
	TestID      string                 `json:"test_id"`
	Groups      map[string]int         `json:"groups"`      // 各组用户数
	Events      map[string]int         `json:"events"`      // 各组事件数
	Conversions map[string]float64     `json:"conversions"` // 各组转化率
	Statistics  map[string]interface{} `json:"statistics"`  // 统计显著性等
}

// UserSegment 用户分群
type UserSegment struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string                 `json:"name" gorm:"not null"`
	Description string                 `json:"description"`
	Criteria    map[string]interface{} `json:"criteria" gorm:"type:jsonb"`
	UserCount   int                    `json:"user_count" gorm:"default:0"`
	IsActive    bool                   `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PersonalizedContent 个性化内容
type PersonalizedContent struct {
	UserID                uuid.UUID            `json:"user_id"`
	RecommendedUsers      []*User              `json:"recommended_users"`
	RecommendedCharacters []*Character         `json:"recommended_characters"`
	ActiveCampaigns       []*MarketingCampaign `json:"active_campaigns"`
	Timestamp             time.Time            `json:"timestamp"`
}

// MarketingStats 营销统计
type MarketingStats struct {
	TotalCampaigns    int     `json:"total_campaigns"`
	ActiveCampaigns   int     `json:"active_campaigns"`
	TotalBudget       float64 `json:"total_budget"`
	TotalSpent        float64 `json:"total_spent"`
	TotalReferrals    int     `json:"total_referrals"`
	SuccessReferrals  int     `json:"success_referrals"`
	TotalRewards      int     `json:"total_rewards"`
	ActiveABTests     int     `json:"active_ab_tests"`
	TotalUserSegments int     `json:"total_user_segments"`
}

// CampaignPerformance 活动表现
type CampaignPerformance struct {
	CampaignID    uuid.UUID `json:"campaign_id"`
	Impressions   int       `json:"impressions"`
	Clicks        int       `json:"clicks"`
	Conversions   int       `json:"conversions"`
	CTR           float64   `json:"ctr"`           // 点击率
	ConversionRate float64  `json:"conversion_rate"` // 转化率
	CostPerClick  float64   `json:"cost_per_click"`
	ROI           float64   `json:"roi"` // 投资回报率
}
