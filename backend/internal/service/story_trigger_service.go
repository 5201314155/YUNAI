package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// StoryTriggerService 剧情触发服务
type StoryTriggerService interface {
	// 分析消息是否触发剧情
	AnalyzeMessageForTriggers(ctx context.Context, req *domain.TriggerAnalysisRequest) (*domain.TriggerAnalysisResult, error)

	// 执行剧情触发
	ExecuteStoryTrigger(ctx context.Context, req *domain.StoryTriggerRequest) (*domain.StoryTriggerResult, error)

	// 获取当前章节信息
	GetCurrentChapter(ctx context.Context, groupChatID uuid.UUID) (*domain.StoryChapter, error)

	// 切换章节
	SwitchChapter(ctx context.Context, req *domain.ChapterSwitchRequest) (*domain.ChapterSwitchResult, error)

	// 智能解析触发条件
	ParseTriggerConditions(ctx context.Context, content string) ([]*domain.TriggerCondition, error)
}

type storyTriggerService struct {
	characterRepo repository.CharacterRepository
	storyRepo     repository.StoryRepository
	modelService  ModelService
	logger        *logrus.Logger
}

// NewStoryTriggerService 创建剧情触发服务
func NewStoryTriggerService(
	characterRepo repository.CharacterRepository,
	storyRepo repository.StoryRepository,
	modelService ModelService,
	logger *logrus.Logger,
) StoryTriggerService {
	return &storyTriggerService{
		characterRepo: characterRepo,
		storyRepo:     storyRepo,
		modelService:  modelService,
		logger:        logger,
	}
}

// AnalyzeMessageForTriggers 分析消息是否触发剧情
func (s *storyTriggerService) AnalyzeMessageForTriggers(ctx context.Context, req *domain.TriggerAnalysisRequest) (*domain.TriggerAnalysisResult, error) {
	s.logger.WithFields(logrus.Fields{
		"group_chat_id": req.GroupChatID,
		"message":       req.Message,
	}).Debug("Analyzing message for story triggers")

	// 获取群聊的剧情设定
	groupChat, err := s.characterRepo.GetGroupChatByID(ctx, req.GroupChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group chat: %w", err)
	}

	// 获取当前章节
	currentChapter, err := s.GetCurrentChapter(ctx, req.GroupChatID)
	if err != nil {
		s.logger.WithError(err).Debug("No current chapter found, using default analysis")
		currentChapter = nil
	}

	// 获取所有可用章节
	chapters, err := s.storyRepo.GetChaptersByGroupChatID(ctx, req.GroupChatID)
	if err != nil {
		s.logger.WithError(err).Debug("No chapters found, using basic trigger analysis")
		chapters = []*domain.StoryChapter{}
	}

	result := &domain.TriggerAnalysisResult{
		GroupChatID:     req.GroupChatID,
		Message:         req.Message,
		CurrentChapter:  currentChapter,
		TriggeredEvents: []*domain.TriggeredEvent{},
	}

	// 1. 关键词触发分析
	keywordTriggers := s.analyzeKeywordTriggers(req.Message, chapters)
	result.TriggeredEvents = append(result.TriggeredEvents, keywordTriggers...)

	// 2. 情绪触发分析
	emotionTriggers := s.analyzeEmotionTriggers(req.Message, chapters)
	result.TriggeredEvents = append(result.TriggeredEvents, emotionTriggers...)

	// 3. 语义触发分析（超级模糊触发）
	semanticTriggers := s.analyzeSemanticTriggers(ctx, req.Message, chapters, groupChat)
	result.TriggeredEvents = append(result.TriggeredEvents, semanticTriggers...)

	// 4. 自定义触发条件分析
	customTriggers := s.analyzeCustomTriggers(req.Message, chapters)
	result.TriggeredEvents = append(result.TriggeredEvents, customTriggers...)

	s.logger.WithField("triggered_events_count", len(result.TriggeredEvents)).Debug("Trigger analysis completed")
	return result, nil
}

// analyzeKeywordTriggers 分析关键词触发
func (s *storyTriggerService) analyzeKeywordTriggers(message string, chapters []*domain.StoryChapter) []*domain.TriggeredEvent {
	var triggers []*domain.TriggeredEvent
	messageLower := strings.ToLower(message)

	// 预定义的关键词触发规则
	keywordRules := map[string][]string{
		"mystery":     {"神秘", "秘密", "隐藏", "奇怪", "不对劲", "诡异", "可疑"},
		"battle":      {"战斗", "打架", "敌人", "危险", "紧张", "攻击", "防御"},
		"celebration": {"开心", "庆祝", "成功", "胜利", "快乐", "欢乐", "喜悦"},
		"romance":     {"爱情", "喜欢", "心动", "表白", "约会", "浪漫"},
		"adventure":   {"冒险", "探索", "发现", "寻找", "旅行", "未知"},
	}

	for triggerType, keywords := range keywordRules {
		for _, keyword := range keywords {
			if strings.Contains(messageLower, keyword) {
				// 查找匹配的章节
				for _, chapter := range chapters {
					if s.chapterMatchesTriggerType(chapter, triggerType) {
						triggers = append(triggers, &domain.TriggeredEvent{
							Type:           "keyword",
							TriggerType:    triggerType,
							MatchedKeyword: keyword,
							TargetChapter:  chapter,
							Confidence:     0.8,
						})
					}
				}
				break // 找到一个关键词就够了
			}
		}
	}

	return triggers
}

// analyzeEmotionTriggers 分析情绪触发
func (s *storyTriggerService) analyzeEmotionTriggers(message string, chapters []*domain.StoryChapter) []*domain.TriggeredEvent {
	var triggers []*domain.TriggeredEvent

	// 情绪检测规则
	emotionPatterns := map[string]*regexp.Regexp{
		"happy":   regexp.MustCompile(`(开心|快乐|高兴|兴奋|愉快|满足|幸福)`),
		"sad":     regexp.MustCompile(`(难过|伤心|沮丧|失落|痛苦|悲伤)`),
		"angry":   regexp.MustCompile(`(生气|愤怒|恼火|烦躁|不爽|火大)`),
		"scared":  regexp.MustCompile(`(害怕|恐惧|紧张|担心|焦虑|不安)`),
		"excited": regexp.MustCompile(`(激动|兴奋|期待|热血|冲动)`),
	}

	for emotion, pattern := range emotionPatterns {
		if pattern.MatchString(message) {
			// 查找匹配情绪的章节
			for _, chapter := range chapters {
				if s.chapterMatchesEmotion(chapter, emotion) {
					triggers = append(triggers, &domain.TriggeredEvent{
						Type:          "emotion",
						TriggerType:   emotion,
						TargetChapter: chapter,
						Confidence:    0.7,
					})
				}
			}
		}
	}

	return triggers
}

// analyzeSemanticTriggers 分析语义触发（超级模糊触发）
func (s *storyTriggerService) analyzeSemanticTriggers(ctx context.Context, message string, chapters []*domain.StoryChapter, groupChat *domain.GroupChat) []*domain.TriggeredEvent {
	var triggers []*domain.TriggeredEvent

	// 语义模式匹配
	semanticPatterns := []struct {
		pattern     *regexp.Regexp
		triggerType string
		confidence  float64
	}{
		{
			pattern:     regexp.MustCompile(`(想.*看看|去.*地方|探索|寻找|发现.*什么)`),
			triggerType: "exploration",
			confidence:  0.6,
		},
		{
			pattern:     regexp.MustCompile(`(感觉.*不对|有点.*奇怪|似乎.*问题)`),
			triggerType: "mystery",
			confidence:  0.7,
		},
		{
			pattern:     regexp.MustCompile(`(准备.*战斗|面临.*挑战|迎接.*困难)`),
			triggerType: "battle",
			confidence:  0.8,
		},
		{
			pattern:     regexp.MustCompile(`(一起.*庆祝|值得.*高兴|真是.*太好了)`),
			triggerType: "celebration",
			confidence:  0.7,
		},
	}

	for _, semantic := range semanticPatterns {
		if semantic.pattern.MatchString(message) {
			// 查找匹配的章节
			for _, chapter := range chapters {
				if s.chapterMatchesTriggerType(chapter, semantic.triggerType) {
					triggers = append(triggers, &domain.TriggeredEvent{
						Type:          "semantic",
						TriggerType:   semantic.triggerType,
						TargetChapter: chapter,
						Confidence:    semantic.confidence,
					})
				}
			}
		}
	}

	return triggers
}

// analyzeCustomTriggers 分析自定义触发条件
func (s *storyTriggerService) analyzeCustomTriggers(message string, chapters []*domain.StoryChapter) []*domain.TriggeredEvent {
	var triggers []*domain.TriggeredEvent

	for _, chapter := range chapters {
		if chapter.TriggerConditions != nil {
			conditions, err := s.parseTriggerConditionsFromJSON(*chapter.TriggerConditions)
			if err != nil {
				s.logger.WithError(err).Warn("Failed to parse trigger conditions")
				continue
			}

			for _, condition := range conditions {
				if s.evaluateTriggerCondition(condition, message) {
					triggers = append(triggers, &domain.TriggeredEvent{
						Type:          "custom",
						TriggerType:   condition.Type,
						TargetChapter: chapter,
						Confidence:    condition.Confidence,
					})
				}
			}
		}
	}

	return triggers
}

// chapterMatchesTriggerType 检查章节是否匹配触发类型
func (s *storyTriggerService) chapterMatchesTriggerType(chapter *domain.StoryChapter, triggerType string) bool {
	if chapter.Tags == nil {
		return false
	}

	// 解析章节标签
	tags := strings.Split(*chapter.Tags, ",")
	for _, tag := range tags {
		if strings.TrimSpace(strings.ToLower(tag)) == triggerType {
			return true
		}
	}

	// 检查章节标题和描述
	titleLower := strings.ToLower(chapter.Title)
	descLower := ""
	if chapter.Description != nil {
		descLower = strings.ToLower(*chapter.Description)
	}

	triggerKeywords := map[string][]string{
		"mystery":     {"神秘", "秘密", "调查", "mystery"},
		"battle":      {"战斗", "大战", "冲突", "battle"},
		"celebration": {"庆祝", "胜利", "成功", "celebration"},
		"romance":     {"爱情", "浪漫", "约会", "romance"},
		"adventure":   {"冒险", "探索", "旅程", "adventure"},
	}

	if keywords, exists := triggerKeywords[triggerType]; exists {
		for _, keyword := range keywords {
			if strings.Contains(titleLower, keyword) || strings.Contains(descLower, keyword) {
				return true
			}
		}
	}

	return false
}

// chapterMatchesEmotion 检查章节是否匹配情绪
func (s *storyTriggerService) chapterMatchesEmotion(chapter *domain.StoryChapter, emotion string) bool {
	// 简化的情绪匹配逻辑
	emotionChapterMap := map[string][]string{
		"happy":   {"庆祝", "欢乐", "成功"},
		"sad":     {"悲伤", "失落", "告别"},
		"angry":   {"战斗", "冲突", "愤怒"},
		"scared":  {"恐怖", "神秘", "危险"},
		"excited": {"冒险", "刺激", "挑战"},
	}

	if keywords, exists := emotionChapterMap[emotion]; exists {
		titleLower := strings.ToLower(chapter.Title)
		for _, keyword := range keywords {
			if strings.Contains(titleLower, keyword) {
				return true
			}
		}
	}

	return false
}

// parseTriggerConditionsFromJSON 从JSON解析触发条件
func (s *storyTriggerService) parseTriggerConditionsFromJSON(jsonStr string) ([]*domain.TriggerCondition, error) {
	// 这里应该实现JSON解析逻辑
	// 暂时返回空切片
	return []*domain.TriggerCondition{}, nil
}

// evaluateTriggerCondition 评估触发条件
func (s *storyTriggerService) evaluateTriggerCondition(condition *domain.TriggerCondition, message string) bool {
	switch condition.Type {
	case "keyword":
		return strings.Contains(strings.ToLower(message), strings.ToLower(condition.Value))
	case "regex":
		if regex, err := regexp.Compile(condition.Value); err == nil {
			return regex.MatchString(message)
		}
	case "semantic":
		// 语义匹配，支持模糊匹配
		return s.semanticMatch(message, condition.Value)
	}
	return false
}

// semanticMatch 语义匹配（支持超级模糊触发）
func (s *storyTriggerService) semanticMatch(message, pattern string) bool {
	messageLower := strings.ToLower(message)
	patternLower := strings.ToLower(pattern)

	// 直接包含
	if strings.Contains(messageLower, patternLower) {
		return true
	}

	// 同义词匹配
	synonyms := map[string][]string{
		"神秘": {"奇怪", "诡异", "可疑", "不对劲", "秘密"},
		"战斗": {"打架", "冲突", "对战", "决斗", "较量"},
		"开心": {"快乐", "高兴", "愉快", "兴奋", "满足"},
		"探索": {"寻找", "发现", "调查", "研究", "查看"},
	}

	for key, syns := range synonyms {
		if strings.Contains(patternLower, key) {
			for _, syn := range syns {
				if strings.Contains(messageLower, syn) {
					return true
				}
			}
		}
	}

	return false
}

// ExecuteStoryTrigger 执行剧情触发
func (s *storyTriggerService) ExecuteStoryTrigger(ctx context.Context, req *domain.StoryTriggerRequest) (*domain.StoryTriggerResult, error) {
	s.logger.WithFields(logrus.Fields{
		"group_chat_id":  req.GroupChatID,
		"trigger_type":   req.TriggerType,
		"target_chapter": req.TargetChapterID,
	}).Info("Executing story trigger")

	result := &domain.StoryTriggerResult{
		GroupChatID: req.GroupChatID,
		TriggerType: req.TriggerType,
		Success:     false,
		Changes:     []*domain.StoryChange{},
	}

	// 获取目标章节
	if req.TargetChapterID != nil {
		_, err := s.storyRepo.GetChapterByID(ctx, *req.TargetChapterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get target chapter: %w", err)
		}

		// 切换章节
		switchReq := &domain.ChapterSwitchRequest{
			GroupChatID:     req.GroupChatID,
			TargetChapterID: *req.TargetChapterID,
			TriggerMessage:  req.TriggerMessage,
		}

		switchResult, err := s.SwitchChapter(ctx, switchReq)
		if err != nil {
			return nil, fmt.Errorf("failed to switch chapter: %w", err)
		}

		result.Success = switchResult.Success
		result.NewChapter = switchResult.NewChapter
		result.Changes = append(result.Changes, switchResult.Changes...)
	}

	return result, nil
}

// GetCurrentChapter 获取当前章节
func (s *storyTriggerService) GetCurrentChapter(ctx context.Context, groupChatID uuid.UUID) (*domain.StoryChapter, error) {
	return s.storyRepo.GetCurrentChapterByGroupChatID(ctx, groupChatID)
}

// SwitchChapter 切换章节
func (s *storyTriggerService) SwitchChapter(ctx context.Context, req *domain.ChapterSwitchRequest) (*domain.ChapterSwitchResult, error) {
	s.logger.WithFields(logrus.Fields{
		"group_chat_id":     req.GroupChatID,
		"target_chapter_id": req.TargetChapterID,
	}).Info("Switching story chapter")

	// 获取新章节
	newChapter, err := s.storyRepo.GetChapterByID(ctx, req.TargetChapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get new chapter: %w", err)
	}

	// 更新群聊的当前章节
	err = s.characterRepo.UpdateCurrentChapter(ctx, req.GroupChatID, req.TargetChapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to update current chapter: %w", err)
	}

	result := &domain.ChapterSwitchResult{
		GroupChatID: req.GroupChatID,
		NewChapter:  newChapter,
		Success:     true,
		Changes:     []*domain.StoryChange{},
	}

	// 记录背景图变化
	if newChapter.BackgroundImageURL != nil {
		result.Changes = append(result.Changes, &domain.StoryChange{
			Type:     "background_image",
			OldValue: "", // 需要从当前状态获取
			NewValue: *newChapter.BackgroundImageURL,
		})
	}

	// 记录背景音乐变化（可选）
	if newChapter.BackgroundMusic != nil {
		result.Changes = append(result.Changes, &domain.StoryChange{
			Type:     "background_music",
			OldValue: "",
			NewValue: *newChapter.BackgroundMusic,
		})
	}

	// 记录音效变化（可选）
	if newChapter.SoundEffects != nil {
		result.Changes = append(result.Changes, &domain.StoryChange{
			Type:     "sound_effects",
			OldValue: "",
			NewValue: *newChapter.SoundEffects,
		})
	}

	// 记录演绎模式变化（可选）
	if newChapter.PerformanceSettings != nil {
		result.Changes = append(result.Changes, &domain.StoryChange{
			Type:     "performance_mode",
			OldValue: "",
			NewValue: *newChapter.PerformanceSettings,
		})
	}

	s.logger.WithField("changes_count", len(result.Changes)).Info("Chapter switch completed")
	return result, nil
}

// ParseTriggerConditions 智能解析触发条件
func (s *storyTriggerService) ParseTriggerConditions(ctx context.Context, content string) ([]*domain.TriggerCondition, error) {
	var conditions []*domain.TriggerCondition

	// 解析自然语言触发条件
	// 例如："当用户提到神秘或秘密时触发"
	patterns := []struct {
		regex       *regexp.Regexp
		triggerType string
	}{
		{
			regex:       regexp.MustCompile(`当.*提到.*([^时]+)时?触发`),
			triggerType: "keyword",
		},
		{
			regex:       regexp.MustCompile(`如果.*感到.*([^，。]+)[，。]?触发`),
			triggerType: "emotion",
		},
		{
			regex:       regexp.MustCompile(`用户说.*([^的]+)的?时候触发`),
			triggerType: "semantic",
		},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				conditions = append(conditions, &domain.TriggerCondition{
					Type:       pattern.triggerType,
					Value:      match[1],
					Confidence: 0.8,
				})
			}
		}
	}

	return conditions, nil
}
