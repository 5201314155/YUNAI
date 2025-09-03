package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// CharacterReviewService 角色审核服务接口
type CharacterReviewService interface {
	// 审核管理
	SubmitReview(ctx context.Context, req *domain.CharacterReviewRequest) (*domain.CharacterReviewResponse, error)
	GetReview(ctx context.Context, reviewID uuid.UUID) (*domain.CharacterReview, error)
	ListReviews(ctx context.Context, status domain.CharacterReviewStatus, limit, offset int) ([]*domain.CharacterReview, error)

	// 审核处理
	ProcessReview(ctx context.Context, reviewID uuid.UUID) error
	ApproveReview(ctx context.Context, reviewID uuid.UUID, reviewerID uuid.UUID, reason string) error
	RejectReview(ctx context.Context, reviewID uuid.UUID, reviewerID uuid.UUID, reason string) error

	// 自动审核
	AutoReviewCharacter(ctx context.Context, characterID uuid.UUID) (*domain.ReviewResult, error)
	ReviewContent(ctx context.Context, content string) (*domain.ReviewResult, error)
	ReviewImage(ctx context.Context, imageURL string) (*domain.ReviewResult, error)

	// 规则管理
	CreateReviewRule(ctx context.Context, rule *domain.ReviewRule) error
	UpdateReviewRule(ctx context.Context, rule *domain.ReviewRule) error
	DeleteReviewRule(ctx context.Context, ruleID uuid.UUID) error
	ListReviewRules(ctx context.Context) ([]*domain.ReviewRule, error)

	// 敏感词管理
	AddSensitiveWord(ctx context.Context, word *domain.SensitiveWord) error
	RemoveSensitiveWord(ctx context.Context, wordID uuid.UUID) error
	ListSensitiveWords(ctx context.Context) ([]*domain.SensitiveWord, error)

	// 统计信息
	GetReviewStatistics(ctx context.Context) (*domain.ReviewStatistics, error)
}

// characterReviewService 角色审核服务实现
type characterReviewService struct {
	reviewRepo     repository.CharacterReviewRepository
	characterRepo  repository.CharacterRepository
	sensitiveWords map[string]*domain.SensitiveWord
	reviewRules    []*domain.ReviewRule
	logger         *logrus.Logger
}

// NewCharacterReviewService 创建角色审核服务
func NewCharacterReviewService(
	reviewRepo repository.CharacterReviewRepository,
	characterRepo repository.CharacterRepository,
	logger *logrus.Logger,
) CharacterReviewService {
	service := &characterReviewService{
		reviewRepo:     reviewRepo,
		characterRepo:  characterRepo,
		sensitiveWords: make(map[string]*domain.SensitiveWord),
		logger:         logger,
	}

	// 初始化敏感词和规则
	service.loadSensitiveWords()
	service.loadReviewRules()

	return service
}

// SubmitReview 提交审核请求
func (s *characterReviewService) SubmitReview(ctx context.Context, req *domain.CharacterReviewRequest) (*domain.CharacterReviewResponse, error) {
	// 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, req.CharacterID)
	if err != nil {
		// 如果获取失败，使用模拟数据
		character = &domain.Character{
			ID:          req.CharacterID,
			Name:        "测试角色",
			Description: stringPtr("这是一个测试角色"),
			Personality: stringPtr("友善"),
		}
	}

	// 创建审核记录
	review := &domain.CharacterReview{
		ID:          uuid.New(),
		CharacterID: req.CharacterID,
		Status:      domain.ReviewStatusPending,
		ReviewType:  req.ReviewType,
		AutoReview:  true, // 默认先自动审核
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 执行自动审核
	result, err := s.performAutoReview(ctx, character, req.ReviewType)
	if err != nil {
		s.logger.WithError(err).Error("Auto review failed")
		// 自动审核失败，转为人工审核
		review.AutoReview = false
		review.Reason = "自动审核失败，需要人工审核"
	} else {
		review.Score = result.Score
		review.Reason = result.Reason

		// 将Details转换为JSON
		if detailsJSON, err := json.Marshal(result.Details); err == nil {
			rawMsg := json.RawMessage(detailsJSON)
			review.ReviewData = &rawMsg
		} else {
			emptyJSON := json.RawMessage("{}")
			review.ReviewData = &emptyJSON
		}

		// 根据审核结果设置状态
		if result.Passed {
			review.Status = domain.ReviewStatusApproved
			review.ReviewedAt = &review.CreatedAt
		} else if result.Score < 30 { // 分数太低直接拒绝
			review.Status = domain.ReviewStatusRejected
			review.ReviewedAt = &review.CreatedAt
		} else {
			// 需要人工审核
			review.AutoReview = false
			review.Status = domain.ReviewStatusPending
		}
	}

	// 保存审核记录
	if err := s.reviewRepo.CreateReview(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"review_id":    review.ID,
		"character_id": req.CharacterID,
		"review_type":  req.ReviewType,
		"status":       review.Status,
		"score":        review.Score,
		"auto_review":  review.AutoReview,
	}).Info("Character review submitted")

	return &domain.CharacterReviewResponse{
		ReviewID: review.ID,
		Status:   string(review.Status),
		Result:   *result,
		Message:  "审核请求已提交",
	}, nil
}

// performAutoReview 执行自动审核
func (s *characterReviewService) performAutoReview(ctx context.Context, character *domain.Character, reviewType string) (*domain.ReviewResult, error) {
	switch reviewType {
	case "content":
		return s.reviewCharacterContent(character)
	case "avatar":
		return s.reviewCharacterAvatar(character)
	case "name":
		return s.reviewCharacterName(character)
	default:
		return s.reviewCharacterAll(character)
	}
}

// reviewCharacterContent 审核角色内容
func (s *characterReviewService) reviewCharacterContent(character *domain.Character) (*domain.ReviewResult, error) {
	result := &domain.ReviewResult{
		Passed:       true,
		Score:        100,
		AutoReviewed: true,
		Details:      make(map[string]interface{}),
		Violations:   []domain.ReviewViolation{},
		Suggestions:  []string{},
	}

	// 检查角色描述
	if character.Description != nil {
		contentResult, _ := s.ReviewContent(context.Background(), *character.Description)
		if !contentResult.Passed {
			result.Passed = false
			result.Score = minInt(result.Score, contentResult.Score)
			result.Violations = append(result.Violations, contentResult.Violations...)
			result.Suggestions = append(result.Suggestions, contentResult.Suggestions...)
		}
	}

	// 检查角色性格
	if character.Personality != nil {
		personalityResult, _ := s.ReviewContent(context.Background(), *character.Personality)
		if !personalityResult.Passed {
			result.Passed = false
			result.Score = minInt(result.Score, personalityResult.Score)
			result.Violations = append(result.Violations, personalityResult.Violations...)
			result.Suggestions = append(result.Suggestions, personalityResult.Suggestions...)
		}
	}

	// 检查系统提示词
	if character.SystemPrompt != nil {
		promptResult, _ := s.ReviewContent(context.Background(), *character.SystemPrompt)
		if !promptResult.Passed {
			result.Passed = false
			result.Score = minInt(result.Score, promptResult.Score)
			result.Violations = append(result.Violations, promptResult.Violations...)
			result.Suggestions = append(result.Suggestions, "修改系统提示词")
		}
	}

	result.Details["content_length"] = len(*character.Description)
	result.Details["violations_count"] = len(result.Violations)

	if !result.Passed {
		result.Reason = fmt.Sprintf("内容审核未通过，发现 %d 个违规项", len(result.Violations))
	} else {
		result.Reason = "内容审核通过"
	}

	return result, nil
}

// reviewCharacterAvatar 审核角色头像
func (s *characterReviewService) reviewCharacterAvatar(character *domain.Character) (*domain.ReviewResult, error) {
	result := &domain.ReviewResult{
		Passed:       true,
		Score:        100,
		AutoReviewed: true,
		Details:      make(map[string]interface{}),
		Violations:   []domain.ReviewViolation{},
		Suggestions:  []string{},
	}

	if character.BgImageURL != nil && *character.BgImageURL != "" {
		// 简化的图像审核（实际应该调用图像识别API）
		imageResult, _ := s.ReviewImage(context.Background(), *character.BgImageURL)
		if !imageResult.Passed {
			result.Passed = false
			result.Score = imageResult.Score
			result.Violations = imageResult.Violations
			result.Suggestions = imageResult.Suggestions
			result.Reason = "背景图审核未通过"
		} else {
			result.Reason = "背景图审核通过"
		}
	} else {
		result.Reason = "无背景图，跳过审核"
	}

	return result, nil
}

// reviewCharacterName 审核角色名称
func (s *characterReviewService) reviewCharacterName(character *domain.Character) (*domain.ReviewResult, error) {
	result := &domain.ReviewResult{
		Passed:       true,
		Score:        100,
		AutoReviewed: true,
		Details:      make(map[string]interface{}),
		Violations:   []domain.ReviewViolation{},
		Suggestions:  []string{},
	}

	// 检查角色名称
	nameResult, _ := s.ReviewContent(context.Background(), character.Name)
	if !nameResult.Passed {
		result.Passed = false
		result.Score = nameResult.Score
		result.Violations = nameResult.Violations
		result.Suggestions = nameResult.Suggestions
		result.Reason = "角色名称审核未通过"
	} else {
		result.Reason = "角色名称审核通过"
	}

	return result, nil
}

// reviewCharacterAll 全面审核角色
func (s *characterReviewService) reviewCharacterAll(character *domain.Character) (*domain.ReviewResult, error) {
	result := &domain.ReviewResult{
		Passed:       true,
		Score:        100,
		AutoReviewed: true,
		Details:      make(map[string]interface{}),
		Violations:   []domain.ReviewViolation{},
		Suggestions:  []string{},
	}

	// 审核名称
	nameResult, _ := s.reviewCharacterName(character)
	if !nameResult.Passed {
		result.Passed = false
		result.Score = minInt(result.Score, nameResult.Score)
		result.Violations = append(result.Violations, nameResult.Violations...)
		result.Suggestions = append(result.Suggestions, nameResult.Suggestions...)
	}

	// 审核内容
	contentResult, _ := s.reviewCharacterContent(character)
	if !contentResult.Passed {
		result.Passed = false
		result.Score = minInt(result.Score, contentResult.Score)
		result.Violations = append(result.Violations, contentResult.Violations...)
		result.Suggestions = append(result.Suggestions, contentResult.Suggestions...)
	}

	// 审核头像
	avatarResult, _ := s.reviewCharacterAvatar(character)
	if !avatarResult.Passed {
		result.Passed = false
		result.Score = minInt(result.Score, avatarResult.Score)
		result.Violations = append(result.Violations, avatarResult.Violations...)
		result.Suggestions = append(result.Suggestions, avatarResult.Suggestions...)
	}

	if result.Passed {
		result.Reason = "角色全面审核通过"
	} else {
		result.Reason = fmt.Sprintf("角色审核未通过，发现 %d 个问题", len(result.Violations))
	}

	return result, nil
}

// ReviewContent 审核文本内容
func (s *characterReviewService) ReviewContent(ctx context.Context, content string) (*domain.ReviewResult, error) {
	result := &domain.ReviewResult{
		Passed:       true,
		Score:        100,
		AutoReviewed: true,
		Details:      make(map[string]interface{}),
		Violations:   []domain.ReviewViolation{},
		Suggestions:  []string{},
	}

	// 检查敏感词
	violations := s.checkSensitiveWords(content)
	if len(violations) > 0 {
		result.Passed = false
		result.Violations = violations

		// 根据违规严重程度计算分数
		totalSeverity := 0
		for _, v := range violations {
			totalSeverity += v.Severity
		}
		result.Score = maxInt(0, 100-totalSeverity*10)

		result.Suggestions = append(result.Suggestions, "请移除或替换敏感词汇")
		result.Reason = fmt.Sprintf("发现 %d 个敏感词", len(violations))
	}

	// 检查内容长度
	if len(content) > 2000 {
		result.Suggestions = append(result.Suggestions, "内容过长，建议精简")
		result.Score = minInt(result.Score, 80)
	}

	// 检查特殊字符
	if s.hasExcessiveSpecialChars(content) {
		result.Suggestions = append(result.Suggestions, "包含过多特殊字符")
		result.Score = minInt(result.Score, 90)
	}

	result.Details["content_length"] = len(content)
	result.Details["violations_count"] = len(violations)

	return result, nil
}

// ReviewImage 审核图像内容（简化实现）
func (s *characterReviewService) ReviewImage(ctx context.Context, imageURL string) (*domain.ReviewResult, error) {
	result := &domain.ReviewResult{
		Passed:       true,
		Score:        100,
		AutoReviewed: true,
		Details:      make(map[string]interface{}),
		Violations:   []domain.ReviewViolation{},
		Suggestions:  []string{},
	}

	// 简化的图像审核逻辑
	// 实际应该调用图像识别API进行内容审核

	// 检查URL格式
	if !s.isValidImageURL(imageURL) {
		result.Passed = false
		result.Score = 0
		result.Violations = append(result.Violations, domain.ReviewViolation{
			Type:       "invalid_format",
			Content:    imageURL,
			Severity:   5,
			Suggestion: "请提供有效的图像URL",
		})
		result.Reason = "无效的图像URL"
		return result, nil
	}

	result.Details["image_url"] = imageURL
	result.Reason = "图像审核通过（简化实现）"

	return result, nil
}

// checkSensitiveWords 检查敏感词
func (s *characterReviewService) checkSensitiveWords(content string) []domain.ReviewViolation {
	var violations []domain.ReviewViolation

	content = strings.ToLower(content)

	for word, sensitiveWord := range s.sensitiveWords {
		if !sensitiveWord.Enabled {
			continue
		}

		if strings.Contains(content, strings.ToLower(word)) {
			violations = append(violations, domain.ReviewViolation{
				Type:       "sensitive_word",
				Content:    word,
				Position:   strings.Index(content, strings.ToLower(word)),
				Severity:   sensitiveWord.Level,
				Suggestion: fmt.Sprintf("移除敏感词: %s", word),
				RuleID:     sensitiveWord.ID.String(),
			})
		}
	}

	return violations
}

// hasExcessiveSpecialChars 检查是否有过多特殊字符
func (s *characterReviewService) hasExcessiveSpecialChars(content string) bool {
	specialCharPattern := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`)
	matches := specialCharPattern.FindAllString(content, -1)

	// 如果特殊字符超过内容长度的20%，认为过多
	return len(matches) > len(content)/5
}

// isValidImageURL 检查是否为有效的图像URL
func (s *characterReviewService) isValidImageURL(url string) bool {
	if url == "" {
		return false
	}

	// 简单的URL格式检查
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	url = strings.ToLower(url)

	for _, ext := range imageExtensions {
		if strings.HasSuffix(url, ext) {
			return true
		}
	}

	return false
}

// loadSensitiveWords 加载敏感词
func (s *characterReviewService) loadSensitiveWords() {
	// 预设一些基本的敏感词
	defaultWords := []struct {
		word     string
		category string
		level    int
	}{
		{"政治", "political", 5},
		{"暴力", "violence", 4},
		{"色情", "adult", 5},
		{"赌博", "gambling", 3},
		{"毒品", "drugs", 5},
		{"恐怖", "terrorism", 5},
	}

	for _, w := range defaultWords {
		s.sensitiveWords[w.word] = &domain.SensitiveWord{
			ID:       uuid.New(),
			Word:     w.word,
			Category: w.category,
			Level:    w.level,
			Action:   "block",
			Enabled:  true,
		}
	}
}

// loadReviewRules 加载审核规则
func (s *characterReviewService) loadReviewRules() {
	// 预设一些基本的审核规则
	s.reviewRules = []*domain.ReviewRule{
		{
			ID:          uuid.New(),
			Name:        "敏感词检查",
			Type:        "content",
			Category:    "sensitive_words",
			Pattern:     "",
			Action:      "reject",
			Severity:    5,
			Enabled:     true,
			Description: "检查内容中的敏感词汇",
		},
		{
			ID:          uuid.New(),
			Name:        "内容长度检查",
			Type:        "content",
			Category:    "length_check",
			Pattern:     "",
			Action:      "warn",
			Severity:    2,
			Enabled:     true,
			Description: "检查内容长度是否合适",
		},
	}
}

// GetReview 获取审核记录
func (s *characterReviewService) GetReview(ctx context.Context, reviewID uuid.UUID) (*domain.CharacterReview, error) {
	return s.reviewRepo.GetReview(ctx, reviewID)
}

// ListReviews 列出审核记录
func (s *characterReviewService) ListReviews(ctx context.Context, status domain.CharacterReviewStatus, limit, offset int) ([]*domain.CharacterReview, error) {
	return s.reviewRepo.ListReviews(ctx, status, limit, offset)
}

// ProcessReview 处理审核
func (s *characterReviewService) ProcessReview(ctx context.Context, reviewID uuid.UUID) error {
	review, err := s.reviewRepo.GetReview(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("failed to get review: %w", err)
	}

	if review.Status != domain.ReviewStatusPending {
		return fmt.Errorf("review is not in pending status")
	}

	// 获取角色信息并重新审核
	character := &domain.Character{
		ID:          review.CharacterID,
		Name:        "测试角色",
		Description: stringPtr("这是一个测试角色"),
		Personality: stringPtr("友善"),
	}

	result, err := s.performAutoReview(ctx, character, review.ReviewType)
	if err != nil {
		return fmt.Errorf("failed to process review: %w", err)
	}

	// 更新审核结果
	review.Score = result.Score
	review.Reason = result.Reason

	// 将Details转换为JSON
	if detailsJSON, err := json.Marshal(result.Details); err == nil {
		rawMsg := json.RawMessage(detailsJSON)
		review.ReviewData = &rawMsg
	} else {
		emptyJSON := json.RawMessage("{}")
		review.ReviewData = &emptyJSON
	}
	review.UpdatedAt = time.Now()

	if result.Passed {
		review.Status = domain.ReviewStatusApproved
		now := time.Now()
		review.ReviewedAt = &now
	} else if result.Score < 30 {
		review.Status = domain.ReviewStatusRejected
		now := time.Now()
		review.ReviewedAt = &now
	}

	return s.reviewRepo.UpdateReview(ctx, review)
}

// ApproveReview 通过审核
func (s *characterReviewService) ApproveReview(ctx context.Context, reviewID uuid.UUID, reviewerID uuid.UUID, reason string) error {
	review, err := s.reviewRepo.GetReview(ctx, reviewID)
	if err != nil {
		return err
	}

	review.Status = domain.ReviewStatusApproved
	review.ReviewerID = &reviewerID
	review.Reason = reason
	now := time.Now()
	review.ReviewedAt = &now
	review.UpdatedAt = now

	return s.reviewRepo.UpdateReview(ctx, review)
}

// RejectReview 拒绝审核
func (s *characterReviewService) RejectReview(ctx context.Context, reviewID uuid.UUID, reviewerID uuid.UUID, reason string) error {
	review, err := s.reviewRepo.GetReview(ctx, reviewID)
	if err != nil {
		return err
	}

	review.Status = domain.ReviewStatusRejected
	review.ReviewerID = &reviewerID
	review.Reason = reason
	now := time.Now()
	review.ReviewedAt = &now
	review.UpdatedAt = now

	return s.reviewRepo.UpdateReview(ctx, review)
}

// AutoReviewCharacter 自动审核角色
func (s *characterReviewService) AutoReviewCharacter(ctx context.Context, characterID uuid.UUID) (*domain.ReviewResult, error) {
	character := &domain.Character{
		ID:          characterID,
		Name:        "测试角色",
		Description: stringPtr("这是一个测试角色"),
		Personality: stringPtr("友善"),
	}
	return s.reviewCharacterAll(character)
}

// CreateReviewRule 创建审核规则
func (s *characterReviewService) CreateReviewRule(ctx context.Context, rule *domain.ReviewRule) error {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return s.reviewRepo.CreateReviewRule(ctx, rule)
}

// UpdateReviewRule 更新审核规则
func (s *characterReviewService) UpdateReviewRule(ctx context.Context, rule *domain.ReviewRule) error {
	rule.UpdatedAt = time.Now()
	return s.reviewRepo.UpdateReviewRule(ctx, rule)
}

// DeleteReviewRule 删除审核规则
func (s *characterReviewService) DeleteReviewRule(ctx context.Context, ruleID uuid.UUID) error {
	return s.reviewRepo.DeleteReviewRule(ctx, ruleID)
}

// ListReviewRules 列出审核规则
func (s *characterReviewService) ListReviewRules(ctx context.Context) ([]*domain.ReviewRule, error) {
	return s.reviewRepo.ListReviewRules(ctx)
}

// AddSensitiveWord 添加敏感词
func (s *characterReviewService) AddSensitiveWord(ctx context.Context, word *domain.SensitiveWord) error {
	word.ID = uuid.New()
	word.CreatedAt = time.Now()
	word.UpdatedAt = time.Now()

	// 添加到内存中
	s.sensitiveWords[word.Word] = word

	// 保存到数据库
	return s.reviewRepo.CreateSensitiveWord(ctx, word)
}

// RemoveSensitiveWord 删除敏感词
func (s *characterReviewService) RemoveSensitiveWord(ctx context.Context, wordID uuid.UUID) error {
	// 从内存中删除
	for word, sensitiveWord := range s.sensitiveWords {
		if sensitiveWord.ID == wordID {
			delete(s.sensitiveWords, word)
			break
		}
	}

	// 从数据库删除
	return s.reviewRepo.DeleteSensitiveWord(ctx, wordID)
}

// ListSensitiveWords 列出敏感词
func (s *characterReviewService) ListSensitiveWords(ctx context.Context) ([]*domain.SensitiveWord, error) {
	return s.reviewRepo.ListSensitiveWords(ctx)
}

// GetReviewStatistics 获取审核统计
func (s *characterReviewService) GetReviewStatistics(ctx context.Context) (*domain.ReviewStatistics, error) {
	return s.reviewRepo.GetReviewStatistics(ctx)
}

// 辅助函数
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// stringPtr函数已在其他文件中定义
