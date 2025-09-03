package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// MomentsService 朋友圈服务接口（深度沉浸式版本）
type MomentsService interface {
	// 朋友圈动态管理
	CreateMoment(ctx context.Context, moment *domain.Moment) error
	GetMoment(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.MomentResponse, error)
	ListMoments(ctx context.Context, req *domain.MomentListRequest) ([]*domain.MomentResponse, int, error)
	UpdateMoment(ctx context.Context, moment *domain.Moment) error
	DeleteMoment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// 朋友圈草稿管理
	GenerateMomentDrafts(ctx context.Context, req *domain.MomentGenerationRequest) (*domain.MomentGenerationResult, error)
	GetMomentDrafts(ctx context.Context, characterID, userID uuid.UUID) ([]*domain.MomentDraft, error)
	PublishMomentDraft(ctx context.Context, req *domain.MomentPublishRequest) (*domain.MomentPublishResult, error)
	DeleteMomentDraft(ctx context.Context, draftID, userID uuid.UUID) error

	// 朋友圈互动
	LikeMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID) error
	UnlikeMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID) error
	CommentMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID, content string) error
	ShareMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID) error

	// 自动生成配置
	GetAutoGenerationConfig(ctx context.Context, characterID uuid.UUID) (*domain.MomentAutoGenerationConfig, error)
	UpdateAutoGenerationConfig(ctx context.Context, config *domain.MomentAutoGenerationConfig) error

	// 自动生成任务
	AutoGenerateMoments(ctx context.Context, characterID uuid.UUID) (*domain.MomentGenerationResult, error)

	// 统计分析
	GetMomentAnalytics(ctx context.Context, characterID, userID uuid.UUID) (*domain.MomentAnalytics, error)

	// 通知管理
	GetUnreadNotifications(ctx context.Context, userID uuid.UUID) ([]*domain.MomentNotification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID, userID uuid.UUID) error

	// 🎭 深度沉浸式朋友圈系统
	GenerateAuthenticMoments(ctx context.Context, characterID uuid.UUID) (*domain.MomentGenerationResult, error)
	SimulateCharacterLifeMoments(ctx context.Context, characterID uuid.UUID, lifeEvents []string) error
	UpdateCharacterSocialDynamics(ctx context.Context, characterID uuid.UUID, socialContext map[string]interface{}) error
}

type momentsService struct {
	momentsRepo      repository.MomentsRepository
	characterRepo    repository.CharacterRepository
	memoryRepo       repository.MemoryRepository
	relationshipRepo repository.RelationshipRepository
	modelService     ModelService
	deepSeekClient   *DeepSeekClient
	logger           *logrus.Logger

	// 🧠 深度沉浸式系统组件
	promptService    DynamicPromptService
	memoryService    MemoryService
	embeddingService EmbeddingService
	characterService CharacterService
}

// NewMomentsService 创建朋友圈服务（集成深度沉浸式系统）
func NewMomentsService(
	momentsRepo repository.MomentsRepository,
	characterRepo repository.CharacterRepository,
	memoryRepo repository.MemoryRepository,
	relationshipRepo repository.RelationshipRepository,
	modelService ModelService,
	deepSeekAPIKey string,
	logger *logrus.Logger,
	// 新增：深度沉浸式系统组件
	promptService DynamicPromptService,
	memoryService MemoryService,
	embeddingService EmbeddingService,
	characterService CharacterService,
) MomentsService {
	return &momentsService{
		momentsRepo:      momentsRepo,
		characterRepo:    characterRepo,
		memoryRepo:       memoryRepo,
		relationshipRepo: relationshipRepo,
		modelService:     modelService,
		deepSeekClient:   NewDeepSeekClient(deepSeekAPIKey, logger),
		logger:           logger,
		promptService:    promptService,
		memoryService:    memoryService,
		embeddingService: embeddingService,
		characterService: characterService,
	}
}

// CreateMoment 创建朋友圈动态
func (s *momentsService) CreateMoment(ctx context.Context, moment *domain.Moment) error {
	s.logger.WithFields(logrus.Fields{
		"character_id": moment.CharacterID,
		"user_id":      moment.UserID,
		"content_type": moment.ContentType,
	}).Info("Creating moment")

	// 设置默认值
	if moment.ID == uuid.Nil {
		moment.ID = uuid.New()
	}
	if moment.CreatedAt.IsZero() {
		moment.CreatedAt = time.Now()
	}
	if moment.UpdatedAt.IsZero() {
		moment.UpdatedAt = time.Now()
	}

	return s.momentsRepo.CreateMoment(ctx, moment)
}

// GetMoment 获取朋友圈动态
func (s *momentsService) GetMoment(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.MomentResponse, error) {
	moment, err := s.momentsRepo.GetMoment(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, moment.CharacterID)
	if err != nil {
		return nil, err
	}

	// 检查用户是否点赞
	interactions, err := s.momentsRepo.GetInteractions(ctx, id, domain.InteractionTypeLike)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get interactions")
	}

	isLiked := false
	for _, interaction := range interactions {
		if interaction.UserID != nil && *interaction.UserID == userID {
			isLiked = true
			break
		}
	}

	// 检查权限
	canEdit := moment.UserID == userID
	canDelete := moment.UserID == userID

	return &domain.MomentResponse{
		Moment:    moment,
		Character: s.convertToCharacterResponse(character),
		IsLiked:   isLiked,
		CanEdit:   canEdit,
		CanDelete: canDelete,
	}, nil
}

// ListMoments 获取朋友圈动态列表
func (s *momentsService) ListMoments(ctx context.Context, req *domain.MomentListRequest) ([]*domain.MomentResponse, int, error) {
	moments, total, err := s.momentsRepo.ListMoments(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	var responses []*domain.MomentResponse
	for _, moment := range moments {
		response, err := s.GetMoment(ctx, moment.ID, req.UserID)
		if err != nil {
			s.logger.WithError(err).WithField("moment_id", moment.ID).Warn("Failed to get moment response")
			continue
		}
		responses = append(responses, response)
	}

	return responses, total, nil
}

// UpdateMoment 更新朋友圈动态
func (s *momentsService) UpdateMoment(ctx context.Context, moment *domain.Moment) error {
	// 检查权限
	existingMoment, err := s.momentsRepo.GetMoment(ctx, moment.ID)
	if err != nil {
		return err
	}

	if existingMoment.UserID != moment.UserID {
		return fmt.Errorf("permission denied: user cannot edit this moment")
	}

	moment.UpdatedAt = time.Now()
	return s.momentsRepo.UpdateMoment(ctx, moment)
}

// DeleteMoment 删除朋友圈动态
func (s *momentsService) DeleteMoment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// 检查权限
	moment, err := s.momentsRepo.GetMoment(ctx, id)
	if err != nil {
		return err
	}

	if moment.UserID != userID {
		return fmt.Errorf("permission denied: user cannot delete this moment")
	}

	return s.momentsRepo.DeleteMoment(ctx, id)
}

// GenerateMomentDrafts 生成朋友圈草稿
func (s *momentsService) GenerateMomentDrafts(ctx context.Context, req *domain.MomentGenerationRequest) (*domain.MomentGenerationResult, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id": req.CharacterID,
		"user_id":      req.UserID,
		"count":        req.Count,
	}).Info("Generating moment drafts")

	result := &domain.MomentGenerationResult{
		CharacterID: req.CharacterID,
		UserID:      req.UserID,
		GeneratedAt: time.Now(),
		Success:     false,
		Drafts:      []*domain.MomentDraft{},
	}

	// 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, req.CharacterID)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to get character: %v", err)
		return result, err
	}

	// 构建生成上下文
	generationContext, err := s.buildGenerationContext(ctx, character, req.UserID)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to build generation context: %v", err)
		return result, err
	}

	// 生成草稿内容
	drafts, err := s.generateDraftContents(ctx, generationContext, req)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to generate draft contents: %v", err)
		return result, err
	}

	// 相似度检查和过滤
	filteredDrafts, filteredCount := s.filterSimilarDrafts(ctx, drafts, req.CharacterID)

	// 保存草稿到数据库
	var savedDrafts []*domain.MomentDraft
	for _, draft := range filteredDrafts {
		err := s.momentsRepo.CreateMomentDraft(ctx, draft)
		if err != nil {
			s.logger.WithError(err).WithField("draft_id", draft.ID).Warn("Failed to save draft")
			continue
		}
		savedDrafts = append(savedDrafts, draft)
	}

	result.Success = len(savedDrafts) > 0
	result.Drafts = savedDrafts
	result.TotalCount = len(savedDrafts)
	result.FilteredCount = filteredCount

	if result.Success {
		result.Message = fmt.Sprintf("Successfully generated %d drafts", len(savedDrafts))
	} else {
		result.Message = "No drafts were generated"
	}

	s.logger.WithFields(logrus.Fields{
		"character_id":    req.CharacterID,
		"generated_count": len(savedDrafts),
		"filtered_count":  filteredCount,
	}).Info("Moment drafts generation completed")

	return result, nil
}

// GetMomentDrafts 获取朋友圈草稿
func (s *momentsService) GetMomentDrafts(ctx context.Context, characterID, userID uuid.UUID) ([]*domain.MomentDraft, error) {
	return s.momentsRepo.ListMomentDrafts(ctx, characterID, userID, 20)
}

// PublishMomentDraft 发布朋友圈草稿
func (s *momentsService) PublishMomentDraft(ctx context.Context, req *domain.MomentPublishRequest) (*domain.MomentPublishResult, error) {
	s.logger.WithFields(logrus.Fields{
		"draft_id":       req.DraftID,
		"user_id":        req.UserID,
		"generate_media": req.GenerateMedia,
	}).Info("Publishing moment draft")

	result := &domain.MomentPublishResult{
		Success:     false,
		PublishedAt: time.Now(),
	}

	// 获取草稿
	draft, err := s.momentsRepo.GetMomentDraft(ctx, req.DraftID)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to get draft: %v", err)
		return result, err
	}

	// 检查权限
	if draft.UserID != req.UserID {
		result.Message = "Permission denied: user cannot publish this draft"
		return result, fmt.Errorf("permission denied")
	}

	// 如果需要生成媒体内容
	if req.GenerateMedia && draft.MediaPrompt != nil {
		mediaURL, err := s.generateMediaContent(ctx, draft)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to generate media content, publishing without media")
		} else {
			result.MediaURL = &mediaURL
		}
	}

	// 发布草稿
	moment, err := s.momentsRepo.PublishDraft(ctx, req.DraftID)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to publish draft: %v", err)
		return result, err
	}

	// 如果生成了媒体，更新朋友圈
	if result.MediaURL != nil {
		moment.MediaURL = result.MediaURL
		moment.MediaType = &draft.ContentType
		err = s.momentsRepo.UpdateMoment(ctx, moment)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to update moment with media URL")
		}
	}

	result.Success = true
	result.MomentID = moment.ID
	result.Message = "Draft published successfully"

	// 创建通知
	s.createMomentNotification(ctx, moment, "new_moment")

	return result, nil
}

// DeleteMomentDraft 删除朋友圈草稿
func (s *momentsService) DeleteMomentDraft(ctx context.Context, draftID, userID uuid.UUID) error {
	// 检查权限
	draft, err := s.momentsRepo.GetMomentDraft(ctx, draftID)
	if err != nil {
		return err
	}

	if draft.UserID != userID {
		return fmt.Errorf("permission denied: user cannot delete this draft")
	}

	return s.momentsRepo.DeleteMomentDraft(ctx, draftID)
}

// LikeMoment 点赞朋友圈
func (s *momentsService) LikeMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID) error {
	interaction := &domain.MomentInteraction{
		ID:          uuid.New(),
		MomentID:    momentID,
		UserID:      &userID,
		CharacterID: characterID,
		Type:        domain.InteractionTypeLike,
		CreatedAt:   time.Now(),
	}

	err := s.momentsRepo.CreateInteraction(ctx, interaction)
	if err != nil {
		return err
	}

	// 更新点赞数
	moment, err := s.momentsRepo.GetMoment(ctx, momentID)
	if err != nil {
		return err
	}

	moment.LikeCount++
	err = s.momentsRepo.UpdateMoment(ctx, moment)
	if err != nil {
		return err
	}

	// 创建通知
	s.createMomentNotification(ctx, moment, "like")

	return nil
}

// UnlikeMoment 取消点赞朋友圈
func (s *momentsService) UnlikeMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID) error {
	// 查找点赞记录
	interactions, err := s.momentsRepo.GetInteractions(ctx, momentID, domain.InteractionTypeLike)
	if err != nil {
		return err
	}

	var targetInteraction *domain.MomentInteraction
	for _, interaction := range interactions {
		if interaction.UserID != nil && *interaction.UserID == userID {
			if characterID == nil && interaction.CharacterID == nil {
				targetInteraction = interaction
				break
			} else if characterID != nil && interaction.CharacterID != nil && *characterID == *interaction.CharacterID {
				targetInteraction = interaction
				break
			}
		}
	}

	if targetInteraction == nil {
		return fmt.Errorf("like not found")
	}

	err = s.momentsRepo.DeleteInteraction(ctx, targetInteraction.ID)
	if err != nil {
		return err
	}

	// 更新点赞数
	moment, err := s.momentsRepo.GetMoment(ctx, momentID)
	if err != nil {
		return err
	}

	if moment.LikeCount > 0 {
		moment.LikeCount--
		err = s.momentsRepo.UpdateMoment(ctx, moment)
		if err != nil {
			return err
		}
	}

	return nil
}

// CommentMoment 评论朋友圈
func (s *momentsService) CommentMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID, content string) error {
	interaction := &domain.MomentInteraction{
		ID:          uuid.New(),
		MomentID:    momentID,
		UserID:      &userID,
		CharacterID: characterID,
		Type:        domain.InteractionTypeComment,
		Content:     &content,
		CreatedAt:   time.Now(),
	}

	err := s.momentsRepo.CreateInteraction(ctx, interaction)
	if err != nil {
		return err
	}

	// 更新评论数
	moment, err := s.momentsRepo.GetMoment(ctx, momentID)
	if err != nil {
		return err
	}

	moment.CommentCount++
	err = s.momentsRepo.UpdateMoment(ctx, moment)
	if err != nil {
		return err
	}

	// 创建通知
	s.createMomentNotification(ctx, moment, "comment")

	return nil
}

// ShareMoment 分享朋友圈
func (s *momentsService) ShareMoment(ctx context.Context, momentID, userID uuid.UUID, characterID *uuid.UUID) error {
	interaction := &domain.MomentInteraction{
		ID:          uuid.New(),
		MomentID:    momentID,
		UserID:      &userID,
		CharacterID: characterID,
		Type:        domain.InteractionTypeShare,
		CreatedAt:   time.Now(),
	}

	err := s.momentsRepo.CreateInteraction(ctx, interaction)
	if err != nil {
		return err
	}

	// 更新分享数
	moment, err := s.momentsRepo.GetMoment(ctx, momentID)
	if err != nil {
		return err
	}

	moment.ShareCount++
	err = s.momentsRepo.UpdateMoment(ctx, moment)
	if err != nil {
		return err
	}

	// 创建通知
	s.createMomentNotification(ctx, moment, "share")

	return nil
}

// GetAutoGenerationConfig 获取自动生成配置
func (s *momentsService) GetAutoGenerationConfig(ctx context.Context, characterID uuid.UUID) (*domain.MomentAutoGenerationConfig, error) {
	config, err := s.momentsRepo.GetAutoGenerationConfig(ctx, characterID)
	if err != nil {
		// 如果配置不存在，返回默认配置
		if strings.Contains(err.Error(), "no rows") {
			return &domain.MomentAutoGenerationConfig{
				CharacterID:     characterID,
				Enabled:         true,
				Frequency:       "daily",
				MaxDraftsPerDay: 5,
				AutoPublish:     false,
				ContentTypes:    []string{domain.MomentContentTypeText},
				Visibility:      domain.MomentVisibilityFriends,
			}, nil
		}
		return nil, err
	}
	return config, nil
}

// UpdateAutoGenerationConfig 更新自动生成配置
func (s *momentsService) UpdateAutoGenerationConfig(ctx context.Context, config *domain.MomentAutoGenerationConfig) error {
	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	return s.momentsRepo.UpsertAutoGenerationConfig(ctx, config)
}

// AutoGenerateMoments 自动生成朋友圈
func (s *momentsService) AutoGenerateMoments(ctx context.Context, characterID uuid.UUID) (*domain.MomentGenerationResult, error) {
	// 获取配置
	config, err := s.GetAutoGenerationConfig(ctx, characterID)
	if err != nil {
		return nil, err
	}

	if !config.Enabled {
		return &domain.MomentGenerationResult{
			CharacterID: characterID,
			UserID:      config.UserID,
			GeneratedAt: time.Now(),
			Success:     false,
			Message:     "Auto generation is disabled",
		}, nil
	}

	// 生成草稿
	req := &domain.MomentGenerationRequest{
		CharacterID:  characterID,
		UserID:       config.UserID,
		Count:        config.MaxDraftsPerDay,
		ContentTypes: config.ContentTypes,
	}

	result, err := s.GenerateMomentDrafts(ctx, req)
	if err != nil {
		return result, err
	}

	// 如果配置了自动发布，发布草稿
	if config.AutoPublish && result.Success {
		for _, draft := range result.Drafts {
			publishReq := &domain.MomentPublishRequest{
				DraftID:       draft.ID,
				UserID:        config.UserID,
				GenerateMedia: false, // 自动发布时不生成媒体
			}
			_, err := s.PublishMomentDraft(ctx, publishReq)
			if err != nil {
				s.logger.WithError(err).WithField("draft_id", draft.ID).Warn("Failed to auto-publish draft")
			}
		}
	}

	return result, nil
}

// GetMomentAnalytics 获取朋友圈分析数据
func (s *momentsService) GetMomentAnalytics(ctx context.Context, characterID, userID uuid.UUID) (*domain.MomentAnalytics, error) {
	return s.momentsRepo.GetMomentAnalytics(ctx, characterID, userID)
}

// GetUnreadNotifications 获取未读通知
func (s *momentsService) GetUnreadNotifications(ctx context.Context, userID uuid.UUID) ([]*domain.MomentNotification, error) {
	return s.momentsRepo.GetUnreadNotifications(ctx, userID)
}

// MarkNotificationAsRead 标记通知为已读
func (s *momentsService) MarkNotificationAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	// 这里可以添加权限检查
	return s.momentsRepo.MarkNotificationAsRead(ctx, notificationID)
}

// 辅助方法

// convertToCharacterResponse 转换角色响应
func (s *momentsService) convertToCharacterResponse(character *domain.Character) *domain.CharacterResponse {
	return &domain.CharacterResponse{
		Character: character,
		Tags:      []string{}, // 可以从其他地方获取标签
		IsOwner:   false,      // 需要根据上下文判断
		CanEdit:   false,      // 需要根据上下文判断
	}
}

// buildGenerationContext 构建生成上下文
func (s *momentsService) buildGenerationContext(ctx context.Context, character *domain.Character, userID uuid.UUID) (*domain.MomentGenerationContext, error) {
	context := &domain.MomentGenerationContext{
		Character: s.convertToCharacterResponse(character),
	}

	// 获取最近聊天记录
	recentChats, err := s.getRecentChats(ctx, character.ID, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get recent chats")
	} else {
		context.RecentChats = recentChats
	}

	// 获取长期记忆
	longTermMemory, err := s.getLongTermMemory(ctx, character.ID, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get long term memory")
	} else {
		context.LongTermMemory = longTermMemory
	}

	// 获取最近的朋友圈，用于去重
	recentMoments, err := s.getRecentMoments(ctx, character.ID, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get recent moments")
	} else {
		context.RecentMoments = recentMoments
	}

	return context, nil
}

// generateDraftContents 生成草稿内容
func (s *momentsService) generateDraftContents(ctx context.Context, generationContext *domain.MomentGenerationContext, req *domain.MomentGenerationRequest) ([]*domain.MomentDraft, error) {
	var drafts []*domain.MomentDraft

	// 使用AI模型生成内容
	contents, err := s.generateMomentContentsWithAI(ctx, generationContext, req)
	if err != nil {
		return nil, err
	}

	for i, content := range contents {
		draft := &domain.MomentDraft{
			ID:          uuid.New(),
			CharacterID: req.CharacterID,
			UserID:      req.UserID,
			Content:     content,
			ContentType: domain.MomentContentTypeText,
			Visibility:  domain.MomentVisibilityFriends,
			Priority:    len(contents) - i, // 优先级递减
			GeneratedAt: time.Now(),
		}

		// 设置心情和标签
		s.setMoodAndTags(draft, generationContext)

		drafts = append(drafts, draft)
	}

	return drafts, nil
}

// filterSimilarDrafts 过滤相似草稿
func (s *momentsService) filterSimilarDrafts(ctx context.Context, drafts []*domain.MomentDraft, characterID uuid.UUID) ([]*domain.MomentDraft, int) {
	var filteredDrafts []*domain.MomentDraft
	filteredCount := 0

	for _, draft := range drafts {
		// 检查相似度
		similarityChecks, err := s.momentsRepo.CheckSimilarity(ctx, characterID, draft.Content, 0.7)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to check similarity")
			filteredDrafts = append(filteredDrafts, draft)
			continue
		}

		// 如果相似度过高，跳过
		if len(similarityChecks) > 0 {
			maxSimilarity := 0.0
			for _, check := range similarityChecks {
				if check.SimilarityScore > maxSimilarity {
					maxSimilarity = check.SimilarityScore
				}
			}
			draft.SimilarityScore = maxSimilarity

			if maxSimilarity > 0.8 {
				filteredCount++
				continue
			}
		}

		filteredDrafts = append(filteredDrafts, draft)
	}

	return filteredDrafts, filteredCount
}

// 更多辅助方法

// getRecentChats 获取最近聊天记录
func (s *momentsService) getRecentChats(ctx context.Context, characterID, userID uuid.UUID) ([]string, error) {
	// 这里应该调用聊天服务获取最近的聊天记录
	// 暂时返回模拟数据
	return []string{
		"今天天气真不错",
		"最近工作有点忙",
		"想念家乡的美食",
	}, nil
}

// getLongTermMemory 获取长期记忆
func (s *momentsService) getLongTermMemory(ctx context.Context, characterID, userID uuid.UUID) ([]string, error) {
	// 这里应该调用记忆服务获取长期记忆
	// 暂时返回模拟数据
	return []string{
		"用户喜欢音乐和电影",
		"经常讨论技术话题",
		"对美食很有兴趣",
	}, nil
}

// getRecentMoments 获取最近的朋友圈
func (s *momentsService) getRecentMoments(ctx context.Context, characterID, userID uuid.UUID) ([]*domain.Moment, error) {
	req := &domain.MomentListRequest{
		UserID:      userID,
		CharacterID: &characterID,
		Page:        1,
		Limit:       10,
	}

	moments, _, err := s.momentsRepo.ListMoments(ctx, req)
	return moments, err
}

// generateMomentContentsWithAI 使用AI生成朋友圈内容
func (s *momentsService) generateMomentContentsWithAI(ctx context.Context, generationContext *domain.MomentGenerationContext, req *domain.MomentGenerationRequest) ([]string, error) {
	// 获取角色信息
	characterName := generationContext.Character.Name
	personality := ""
	if generationContext.Character.Personality != nil {
		personality = *generationContext.Character.Personality
	}

	// 使用DeepSeek API生成内容
	contents, err := s.deepSeekClient.GenerateMomentContent(
		ctx,
		characterName,
		personality,
		generationContext.RecentChats,
		req.Count,
	)

	if err != nil {
		s.logger.WithError(err).Error("Failed to generate moment contents with DeepSeek")
		// 如果API调用失败，返回备用内容
		return s.getFallbackContents(req.Count), nil
	}

	s.logger.WithFields(logrus.Fields{
		"character":       characterName,
		"generated_count": len(contents),
		"requested_count": req.Count,
	}).Info("Successfully generated moment contents with DeepSeek")

	return contents, nil
}

// getFallbackContents 获取备用内容
func (s *momentsService) getFallbackContents(count int) []string {
	fallbackContents := []string{
		"今天阳光明媚，心情也跟着好起来了 ☀️",
		"刚刚看了一部很棒的电影，推荐给大家！",
		"最近在学习新技能，感觉很充实",
		"想念家乡的味道，什么时候能回去看看",
		"和朋友们聊天总是很开心，友谊万岁！",
		"今天的咖啡特别香，配上这本书，完美的下午 ☕📖",
		"突然想起小时候的那些美好时光",
		"努力工作，也要记得享受生活 💪",
	}

	if count < len(fallbackContents) {
		return fallbackContents[:count]
	}

	return fallbackContents
}

// buildMomentGenerationPrompt 构建朋友圈生成提示词
func (s *momentsService) buildMomentGenerationPrompt(generationContext *domain.MomentGenerationContext, req *domain.MomentGenerationRequest) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("请为以下角色生成朋友圈内容：\n")
	promptBuilder.WriteString(fmt.Sprintf("角色名称：%s\n", generationContext.Character.Name))

	if generationContext.Character.Personality != nil {
		promptBuilder.WriteString(fmt.Sprintf("性格特点：%s\n", *generationContext.Character.Personality))
	}

	if len(generationContext.RecentChats) > 0 {
		promptBuilder.WriteString("最近聊天内容：\n")
		for _, chat := range generationContext.RecentChats {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", chat))
		}
	}

	if len(generationContext.LongTermMemory) > 0 {
		promptBuilder.WriteString("长期记忆：\n")
		for _, memory := range generationContext.LongTermMemory {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", memory))
		}
	}

	promptBuilder.WriteString(fmt.Sprintf("请生成%d条朋友圈内容，要求：\n", req.Count))
	promptBuilder.WriteString("1. 符合角色性格特点\n")
	promptBuilder.WriteString("2. 内容自然真实\n")
	promptBuilder.WriteString("3. 避免重复\n")
	promptBuilder.WriteString("4. 长度适中（20-100字）\n")

	return promptBuilder.String()
}

// setMoodAndTags 设置心情和标签
func (s *momentsService) setMoodAndTags(draft *domain.MomentDraft, generationContext *domain.MomentGenerationContext) {
	// 根据内容分析心情
	content := strings.ToLower(draft.Content)

	if strings.Contains(content, "开心") || strings.Contains(content, "高兴") || strings.Contains(content, "😊") {
		mood := "happy"
		draft.Mood = &mood
		draft.Tags = []string{"开心", "心情好"}
	} else if strings.Contains(content, "难过") || strings.Contains(content, "伤心") || strings.Contains(content, "😢") {
		mood := "sad"
		draft.Mood = &mood
		draft.Tags = []string{"难过", "心情低落"}
	} else if strings.Contains(content, "工作") || strings.Contains(content, "学习") {
		mood := "focused"
		draft.Mood = &mood
		draft.Tags = []string{"工作", "学习"}
	} else if strings.Contains(content, "美食") || strings.Contains(content, "吃") {
		mood := "satisfied"
		draft.Mood = &mood
		draft.Tags = []string{"美食", "生活"}
	} else {
		mood := "neutral"
		draft.Mood = &mood
		draft.Tags = []string{"日常", "生活"}
	}
}

// generateMediaContent 生成媒体内容
func (s *momentsService) generateMediaContent(ctx context.Context, draft *domain.MomentDraft) (string, error) {
	// 这里应该调用媒体生成服务
	// 暂时返回模拟URL
	return "https://example.com/generated-media.jpg", nil
}

// createMomentNotification 创建朋友圈通知
func (s *momentsService) createMomentNotification(ctx context.Context, moment *domain.Moment, notificationType string) {
	notification := &domain.MomentNotification{
		ID:          uuid.New(),
		UserID:      moment.UserID,
		CharacterID: moment.CharacterID,
		MomentID:    moment.ID,
		Type:        notificationType,
		Content:     s.buildNotificationContent(moment, notificationType),
		CreatedAt:   time.Now(),
	}

	err := s.momentsRepo.CreateNotification(ctx, notification)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to create moment notification")
	}
}

// buildNotificationContent 构建通知内容
func (s *momentsService) buildNotificationContent(moment *domain.Moment, notificationType string) string {
	switch notificationType {
	case "new_moment":
		return fmt.Sprintf("发布了新的朋友圈：%s", s.truncateContent(moment.Content, 30))
	case "like":
		return fmt.Sprintf("有人点赞了你的朋友圈：%s", s.truncateContent(moment.Content, 30))
	case "comment":
		return fmt.Sprintf("有人评论了你的朋友圈：%s", s.truncateContent(moment.Content, 30))
	case "share":
		return fmt.Sprintf("有人分享了你的朋友圈：%s", s.truncateContent(moment.Content, 30))
	default:
		return "朋友圈有新动态"
	}
}

// truncateContent 截断内容
func (s *momentsService) truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

// =============================================================================
// 🎭 深度沉浸式朋友圈系统核心实现
// =============================================================================

// GenerateAuthenticMoments 生成真实的朋友圈动态（深度沉浸式）
func (s *momentsService) GenerateAuthenticMoments(ctx context.Context, characterID uuid.UUID) (*domain.MomentGenerationResult, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id":    characterID,
		"system":          "deep_immersion_moments",
		"generation_type": "authentic_life_sharing",
	}).Info("开始生成深度沉浸式朋友圈动态")

	// 1. 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("获取角色信息失败: %w", err)
	}

	// 2. 激活角色的深度沉浸式身份状态
	err = s.characterService.ActivateDeepImmersion(ctx, characterID, character.UserID)
	if err != nil {
		s.logger.WithError(err).Warn("激活深度沉浸状态失败")
	}

	// 3. 获取角色的生活记忆和情感状态
	lifeMemories, err := s.getCharacterLifeMemories(ctx, characterID)
	if err != nil {
		s.logger.WithError(err).Warn("获取生活记忆失败")
		lifeMemories = []map[string]interface{}{}
	}

	emotionalState, err := s.getCharacterEmotionalState(ctx, characterID)
	if err != nil {
		s.logger.WithError(err).Warn("获取情感状态失败")
		emotionalState = map[string]interface{}{
			"primary_emotion":    "content",
			"authenticity_level": 0.95,
		}
	}

	// 4. 获取关系网络信息
	relationships, err := s.getCharacterRelationships(ctx, characterID)
	if err != nil {
		s.logger.WithError(err).Warn("获取关系网络信息失败")
		relationships = []map[string]interface{}{}
	}

	// 5. 生成真实的朋友圈内容
	momentsContent, err := s.generateAuthenticMomentsContent(ctx, character, lifeMemories, emotionalState, relationships)
	if err != nil {
		return nil, fmt.Errorf("生成朋友圈内容失败: %w", err)
	}

	// 6. 创建朋友圈草稿
	drafts := make([]*domain.MomentDraft, len(momentsContent))
	for i, content := range momentsContent {
		draft := &domain.MomentDraft{
			ID:          uuid.New(),
			CharacterID: characterID,
			UserID:      character.UserID,
			Content:     content["content"].(string),
			ContentType: "text",
			Visibility:  "public",
			GeneratedAt: time.Now(),
			IsPublished: false,
		}

		// 设置情感和标签
		if mood, ok := content["mood"].(string); ok {
			draft.Mood = &mood
		}
		if tags, ok := content["tags"].([]string); ok {
			draft.Tags = tags
		}

		// 保存草稿
		err = s.momentsRepo.CreateMomentDraft(ctx, draft)
		if err != nil {
			s.logger.WithError(err).WithField("draft_id", draft.ID).Warn("保存朋友圈草稿失败")
			continue
		}

		drafts[i] = draft
	}

	s.logger.WithFields(logrus.Fields{
		"character_id":       characterID,
		"character_name":     character.Name,
		"generated_count":    len(drafts),
		"authenticity_level": "deep_immersion",
	}).Info("深度沉浸式朋友圈动态生成完成")

	return &domain.MomentGenerationResult{
		Drafts:      drafts,
		TotalCount:  len(drafts),
		GeneratedAt: time.Now(),
	}, nil
}

// generateAuthenticMomentsContent 生成真实的朋友圈内容
func (s *momentsService) generateAuthenticMomentsContent(ctx context.Context, character *domain.Character, lifeMemories []map[string]interface{}, emotionalState map[string]interface{}, relationships []map[string]interface{}) ([]map[string]interface{}, error) {
	// 构建深度沉浸式提示词参数
	promptParams := map[string]interface{}{
		"character_name":        character.Name,
		"character_personality": character.Personality,
		"character_description": character.Description,
		"life_memories":         lifeMemories,
		"emotional_state":       emotionalState,
		"relationships":         relationships,
		// 深度沉浸式参数
		"identity_strength":      1.0,
		"reality_conviction":     0.99,
		"authentic_sharing":      true,
		"life_immersion":         true,
		"emotional_authenticity": 0.98,
		"social_awareness":       true,
	}

	// 生成朋友圈内容
	_, err := s.promptService.GeneratePrompt(ctx, "authentic_moments_generation", promptParams)
	if err != nil {
		return nil, fmt.Errorf("生成朋友圈内容失败: %w", err)
	}

	// 解析生成的内容
	// 这里应该解析JSON格式的回复，但为了简化，我们返回示例数据
	return []map[string]interface{}{
		{
			"content":            fmt.Sprintf("今天的阳光特别温暖，让我想起了很多美好的回忆。作为%s，我珍惜这样的时刻。", character.Name),
			"mood":               "content",
			"tags":               []string{"日常", "感恩", "生活"},
			"authenticity_score": 0.95,
		},
		{
			"content":            fmt.Sprintf("刚刚和好朋友聊天，又学到了新东西。人生的每一天都有新的可能性，我是%s，我相信这一点。", character.Name),
			"mood":               "inspired",
			"tags":               []string{"学习", "成长", "朋友"},
			"authenticity_score": 0.92,
		},
	}, nil
}

// getCharacterLifeMemories 获取角色生活记忆
func (s *momentsService) getCharacterLifeMemories(ctx context.Context, characterID uuid.UUID) ([]map[string]interface{}, error) {
	// 获取生活相关的记忆
	lifeList, _, err := s.memoryService.ListMemoryFragments(ctx, &characterID, []string{"life_experience"}, 1, 50)
	var lifeMemories []map[string]interface{}
	if err == nil {
		for _, m := range lifeList {
			lifeMemories = append(lifeMemories, map[string]interface{}{"content": m.Content, "type": m.MemoryType})
		}
	} else {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	// 获取日常活动记忆
	dailyList, _, err := s.memoryService.ListMemoryFragments(ctx, &characterID, []string{"daily_activity"}, 1, 50)
	if err == nil {
		for _, m := range dailyList {
			lifeMemories = append(lifeMemories, map[string]interface{}{"content": m.Content, "type": m.MemoryType})
		}
	}

	// 获取情感经历记忆
	emoList, _, err := s.memoryService.ListMemoryFragments(ctx, &characterID, []string{"emotional_experience"}, 1, 50)
	if err == nil {
		for _, m := range emoList {
			lifeMemories = append(lifeMemories, map[string]interface{}{"content": m.Content, "type": m.MemoryType})
		}
	}

	return lifeMemories, nil
}

// getCharacterEmotionalState 获取角色情感状态
func (s *momentsService) getCharacterEmotionalState(ctx context.Context, characterID uuid.UUID) (map[string]interface{}, error) {
	// 从记忆系统获取最新情感状态
	emoList, _, err := s.memoryService.ListMemoryFragments(ctx, &characterID, []string{"emotional_state"}, 1, 1)
	if err != nil || len(emoList) == 0 {
		return map[string]interface{}{
			"primary_emotion":    "content",
			"authenticity_level": 0.9,
			"reality_conviction": 0.95,
		}, nil
	}

	// 返回最新的情感状态
	m := emoList[0]
	return map[string]interface{}{
		"raw":      m.Content,
		"summary":  m.Summary,
		"metadata": m.ContextMetadata,
	}, nil
}

// getCharacterRelationships 获取角色关系网络
func (s *momentsService) getCharacterRelationships(ctx context.Context, characterID uuid.UUID) ([]map[string]interface{}, error) {
	// 这里应该调用关系服务获取关系网络
	// 暂时返回模拟数据
	return []map[string]interface{}{
		{
			"relationship_type":     "friend",
			"relationship_strength": 0.8,
			"interaction_frequency": "daily",
			"shared_interests":      []string{"音乐", "电影", "美食"},
		},
		{
			"relationship_type":     "close_friend",
			"relationship_strength": 0.95,
			"interaction_frequency": "frequent",
			"shared_memories":       []string{"一起旅行", "共同爱好"},
		},
	}, nil
}

// SimulateCharacterLifeMoments 模拟角色生活朋友圈
func (s *momentsService) SimulateCharacterLifeMoments(ctx context.Context, characterID uuid.UUID, lifeEvents []string) error {
	s.logger.WithFields(logrus.Fields{
		"character_id":      characterID,
		"life_events_count": len(lifeEvents),
		"action":            "simulate_life_moments",
	}).Info("开始模拟角色生活朋友圈")

	// 为每个生活事件创建记忆
	for _, event := range lifeEvents {
		memory := map[string]interface{}{
			"memory_type":      "life_experience",
			"content":          event,
			"importance":       4,
			"emotional_weight": "significant",
			"timestamp":        time.Now(),
			"is_moment_worthy": true,
		}

		err := s.memoryService.StoreMemory(ctx, characterID, memory)
		if err != nil {
			s.logger.WithError(err).WithField("event", event).Warn("存储生活事件记忆失败")
		}
	}

	// 自动生成相关朋友圈
	_, err := s.GenerateAuthenticMoments(ctx, characterID)
	if err != nil {
		s.logger.WithError(err).Warn("自动生成朋友圈失败")
	}

	return nil
}

// UpdateCharacterSocialDynamics 更新角色社交动态
func (s *momentsService) UpdateCharacterSocialDynamics(ctx context.Context, characterID uuid.UUID, socialContext map[string]interface{}) error {
	s.logger.WithFields(logrus.Fields{
		"character_id":   characterID,
		"social_updates": len(socialContext),
		"action":         "update_social_dynamics",
	}).Debug("更新角色社交动态")

	// 存储社交动态信息
	socialMemory := map[string]interface{}{
		"memory_type":     "social_dynamics",
		"content":         socialContext,
		"importance":      3,
		"timestamp":       time.Now(),
		"affects_moments": true,
	}

	err := s.memoryService.StoreMemory(ctx, characterID, socialMemory)
	if err != nil {
		return fmt.Errorf("存储社交动态信息失败: %w", err)
	}

	// 更新角色情感状态
	if emotionalImpact, exists := socialContext["emotional_impact"]; exists {
		if emotionalMap, ok := emotionalImpact.(map[string]interface{}); ok {
			err = s.characterService.UpdateCharacterEmotionalState(ctx, characterID, emotionalMap)
			if err != nil {
				s.logger.WithError(err).Warn("更新情感状态失败")
			}
		}
	}

	return nil
}
