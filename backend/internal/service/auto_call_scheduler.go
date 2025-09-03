package service

import (
	"context"
	"time"

	"yunai/internal/domain"
	"yunai/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AutoCallScheduler 自动外呼调度器
type AutoCallScheduler struct {
	logger           *logrus.Logger
	voiceCallService *AIVoiceCallService
	userRepo         repository.UserRepository
	characterRepo    repository.CharacterRepository
	conversationRepo repository.ConversationRepository
	sessionRepo      repository.VoiceCallSessionRepository
	ticker           *time.Ticker
	stopChan         chan bool
	running          bool
}

// NewAutoCallScheduler 创建自动外呼调度器
func NewAutoCallScheduler(logger *logrus.Logger, voiceCallService *AIVoiceCallService) *AutoCallScheduler {
	return &AutoCallScheduler{
		logger:           logger,
		voiceCallService: voiceCallService,
		stopChan:         make(chan bool),
		running:          false,
	}
}

// Start 启动调度器
func (s *AutoCallScheduler) Start(ctx context.Context, interval time.Duration) {
	if s.running {
		s.logger.Warn("自动外呼调度器已在运行")
		return
	}

	s.logger.WithField("interval", interval).Info("启动自动外呼调度器")
	s.ticker = time.NewTicker(interval)
	s.running = true

	go s.run(ctx)
}

// Stop 停止调度器
func (s *AutoCallScheduler) Stop() {
	if !s.running {
		return
	}

	s.logger.Info("停止自动外呼调度器")
	s.running = false
	s.ticker.Stop()
	s.stopChan <- true
}

// run 运行调度器
func (s *AutoCallScheduler) run(ctx context.Context) {
	for {
		select {
		case <-s.ticker.C:
			s.checkAndTriggerAutoCalls(ctx)
		case <-s.stopChan:
			s.logger.Info("自动外呼调度器已停止")
			return
		case <-ctx.Done():
			s.logger.Info("自动外呼调度器因上下文取消而停止")
			return
		}
	}
}

// checkAndTriggerAutoCalls 检查并触发自动外呼
func (s *AutoCallScheduler) checkAndTriggerAutoCalls(ctx context.Context) {
	s.logger.Debug("开始检查自动外呼条件")

	// 获取所有活跃用户
	users, err := s.userRepo.ListActive(ctx)
	if err != nil {
		s.logger.WithError(err).Error("获取活跃用户失败")
		return
	}

	for _, user := range users {
		s.checkUserForAutoCall(ctx, user)
	}
}

// checkUserForAutoCall 检查单个用户是否需要自动外呼
func (s *AutoCallScheduler) checkUserForAutoCall(ctx context.Context, user *domain.User) {
	// 获取用户的角色
	characters, err := s.characterRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Warn("获取用户角色失败")
		return
	}

	for _, character := range characters {
		s.checkUserCharacterForAutoCall(ctx, user, character)
	}
}

// checkUserCharacterForAutoCall 检查用户和角色的自动外呼条件
func (s *AutoCallScheduler) checkUserCharacterForAutoCall(ctx context.Context, user *domain.User, character *domain.Character) {
	// 1. 检查长时间未活跃
	if s.shouldCallForInactivity(ctx, user.ID.String(), character.ID.String()) {
		s.triggerAutoCall(ctx, user.ID.String(), character.ID.String(), "long_inactive", "用户很久没有聊天了")
		return
	}

	// 2. 检查特殊事件
	if s.shouldCallForSpecialEvent(ctx, user.ID.String(), character.ID.String()) {
		s.triggerAutoCall(ctx, user.ID.String(), character.ID.String(), "special_event", "今天是特殊的日子")
		return
	}

	// 3. 检查定时提醒
	if s.shouldCallForScheduledReminder(ctx, user.ID.String(), character.ID.String()) {
		s.triggerAutoCall(ctx, user.ID.String(), character.ID.String(), "scheduled_reminder", "定时关怀提醒")
		return
	}

	// 4. 检查情感支持需求
	if s.shouldCallForEmotionalSupport(ctx, user.ID.String(), character.ID.String()) {
		s.triggerAutoCall(ctx, user.ID.String(), character.ID.String(), "emotional_support", "检测到用户可能需要情感支持")
		return
	}
}

// shouldCallForInactivity 检查是否因长时间未活跃需要外呼
func (s *AutoCallScheduler) shouldCallForInactivity(ctx context.Context, userID, characterID string) bool {
	// 获取最后一次对话时间
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	characterUUID, err := uuid.Parse(characterID)
	if err != nil {
		return false
	}

	lastConversation, err := s.conversationRepo.GetLastByUserAndCharacter(ctx, userUUID, characterUUID)
	if err != nil {
		// 如果没有对话记录，说明是新用户，不需要外呼
		return false
	}

	// 检查是否超过3天没有聊天
	inactiveThreshold := 3 * 24 * time.Hour
	if time.Since(lastConversation.CreatedAt) > inactiveThreshold {
		// 检查是否已经有最近的外呼记录
		recentCall, err := s.sessionRepo.GetRecentOutgoingCall(ctx, userID, characterID)
		if err != nil || recentCall == nil {
			return true
		}
	}

	return false
}

// shouldCallForSpecialEvent 检查是否因特殊事件需要外呼
func (s *AutoCallScheduler) shouldCallForSpecialEvent(ctx context.Context, userID, characterID string) bool {
	// 检查今天是否是特殊日期
	now := time.Now()

	// 检查是否是用户生日
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err == nil {
		// 简化生日检查 - 使用创建日期作为生日
		if s.isSameMonthDay(now, user.CreatedAt) {
			return true
		}
	}

	// 检查是否是节假日
	if s.isHoliday(now) {
		return true
	}

	// 检查是否是周末且用户喜欢周末聊天
	if s.isWeekend(now) && s.userLikesWeekendChats(ctx, userID) {
		return true
	}

	return false
}

// shouldCallForScheduledReminder 检查是否因定时提醒需要外呼
func (s *AutoCallScheduler) shouldCallForScheduledReminder(ctx context.Context, userID, characterID string) bool {
	// 检查是否是用户设定的提醒时间
	// 这里可以实现用户自定义的提醒逻辑

	now := time.Now()

	// 例如：每周一早上9点的工作提醒
	if now.Weekday() == time.Monday && now.Hour() == 9 && now.Minute() < 30 {
		// 检查是否已经提醒过
		recentCall, err := s.sessionRepo.GetRecentCallByReason(ctx, userID, characterID, "scheduled_reminder")
		if err != nil || recentCall == nil {
			return true
		}
	}

	return false
}

// shouldCallForEmotionalSupport 检查是否因情感支持需要外呼
func (s *AutoCallScheduler) shouldCallForEmotionalSupport(ctx context.Context, userID, characterID string) bool {
	// 分析用户最近的情感状态
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	characterUUID, err := uuid.Parse(characterID)
	if err != nil {
		return false
	}

	recentConversations, err := s.conversationRepo.GetRecentByUserAndCharacter(ctx, userUUID, characterUUID, 5)
	if err != nil || len(recentConversations) == 0 {
		return false
	}

	// 简化情感检查 - Conversation结构体没有UserEmotion字段
	// 这里可以基于对话标题或状态进行简单判断
	negativeCount := 0
	for _, conv := range recentConversations {
		if conv.Status == "archived" || conv.Status == "deleted" {
			negativeCount++
		}
	}

	// 如果最近对话中有超过一半是负面状态，考虑主动关怀
	if float64(negativeCount)/float64(len(recentConversations)) > 0.5 {
		// 检查是否已经有最近的情感支持外呼
		recentCall, err := s.sessionRepo.GetRecentCallByReason(ctx, userID, characterID, "emotional_support")
		if err != nil || recentCall == nil {
			return true
		}
	}

	return false
}

// triggerAutoCall 触发自动外呼
func (s *AutoCallScheduler) triggerAutoCall(ctx context.Context, userID, characterID, reason, context string) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"character_id": characterID,
		"reason":       reason,
		"context":      context,
	}).Info("触发自动外呼")

	req := &domain.AutoCallRequest{
		UserID:      userID,
		CharacterID: characterID,
		Reason:      reason,
		Context:     context,
	}

	_, err := s.voiceCallService.AutoCall(ctx, req)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      userID,
			"character_id": characterID,
			"reason":       reason,
		}).Error("自动外呼失败")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"character_id": characterID,
		"reason":       reason,
	}).Info("自动外呼成功")
}

// 辅助方法

// isSameMonthDay 检查两个时间是否是同一个月日
func (s *AutoCallScheduler) isSameMonthDay(t1, t2 time.Time) bool {
	return t1.Month() == t2.Month() && t1.Day() == t2.Day()
}

// isHoliday 检查是否是节假日
func (s *AutoCallScheduler) isHoliday(t time.Time) bool {
	// 简单的节假日检查，可以扩展为更完整的节假日数据库
	month := t.Month()
	day := t.Day()

	// 新年
	if month == time.January && day == 1 {
		return true
	}

	// 情人节
	if month == time.February && day == 14 {
		return true
	}

	// 圣诞节
	if month == time.December && day == 25 {
		return true
	}

	// 可以添加更多节假日
	return false
}

// isWeekend 检查是否是周末
func (s *AutoCallScheduler) isWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// userLikesWeekendChats 检查用户是否喜欢周末聊天
func (s *AutoCallScheduler) userLikesWeekendChats(ctx context.Context, userID string) bool {
	// 分析用户的聊天模式，这里简化为返回true
	// 实际实现可以分析用户的历史聊天时间分布
	return true
}

// isNegativeEmotion 检查是否是负面情感
func (s *AutoCallScheduler) isNegativeEmotion(emotion string) bool {
	negativeEmotions := map[string]bool{
		"sad":      true,
		"angry":    true,
		"worried":  true,
		"lonely":   true,
		"tired":    true,
		"stressed": true,
	}

	return negativeEmotions[emotion]
}

// GetSchedulerStatus 获取调度器状态
func (s *AutoCallScheduler) GetSchedulerStatus() map[string]interface{} {
	return map[string]interface{}{
		"running":     s.running,
		"next_check":  s.getNextCheckTime(),
		"total_calls": s.getTotalAutoCalls(),
	}
}

// getNextCheckTime 获取下次检查时间
func (s *AutoCallScheduler) getNextCheckTime() *time.Time {
	if !s.running || s.ticker == nil {
		return nil
	}

	// 这里简化处理，实际可以更精确计算
	nextTime := time.Now().Add(time.Hour)
	return &nextTime
}

// getTotalAutoCalls 获取总自动外呼次数
func (s *AutoCallScheduler) getTotalAutoCalls() int {
	// 这里可以从数据库查询统计信息
	return 0
}
