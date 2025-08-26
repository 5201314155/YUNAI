package service

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// SmartMomentsService 智能朋友圈服务
type SmartMomentsService interface {
	// 智能生成朋友圈
	SmartGenerateMoments(ctx context.Context, req *domain.MomentGenerationRequest) (*domain.MomentGenerationResult, error)

	// 自动互动处理
	TriggerAutoInteractions(ctx context.Context, momentID uuid.UUID) error

	// 处理@提及
	ProcessMentions(ctx context.Context, momentID uuid.UUID, content string) error

	// 获取智能生成上下文
	BuildSmartGenerationContext(ctx context.Context, characterID, userID uuid.UUID, basedOnChat bool) (*domain.SmartMomentGeneration, error)
}

type smartMomentsService struct {
	momentsRepo      repository.MomentsRepository
	characterRepo    repository.CharacterRepository
	relationshipRepo repository.RelationshipRepository
	// chatRepo         repository.ChatRepository // 暂时注释掉，等待实现
	modelService        ModelService
	deepSeekClient      *DeepSeekClient
	userIdentityService UserIdentityService
	logger              *logrus.Logger
}

// NewSmartMomentsService 创建智能朋友圈服务
func NewSmartMomentsService(
	momentsRepo repository.MomentsRepository,
	characterRepo repository.CharacterRepository,
	relationshipRepo repository.RelationshipRepository,
	// chatRepo repository.ChatRepository, // 暂时注释掉
	modelService ModelService,
	deepSeekAPIKey string,
	userIdentityService UserIdentityService,
	logger *logrus.Logger,
) SmartMomentsService {
	return &smartMomentsService{
		momentsRepo:      momentsRepo,
		characterRepo:    characterRepo,
		relationshipRepo: relationshipRepo,
		// chatRepo:         chatRepo, // 暂时注释掉
		modelService:        modelService,
		deepSeekClient:      NewDeepSeekClient(deepSeekAPIKey, logger),
		userIdentityService: userIdentityService,
		logger:              logger,
	}
}

// SmartGenerateMoments 智能生成朋友圈
func (s *smartMomentsService) SmartGenerateMoments(ctx context.Context, req *domain.MomentGenerationRequest) (*domain.MomentGenerationResult, error) {
	s.logger.WithFields(logrus.Fields{
		"character_id":  req.CharacterID,
		"user_id":       req.UserID,
		"count":         req.Count,
		"based_on_chat": req.BasedOnChat,
		"auto_interact": req.AutoInteract,
	}).Info("Starting smart moment generation")

	result := &domain.MomentGenerationResult{
		CharacterID: req.CharacterID,
		UserID:      req.UserID,
		GeneratedAt: time.Now(),
		Success:     false,
		Drafts:      []*domain.MomentDraft{},
	}

	// 1. 构建智能生成上下文
	smartContext, err := s.BuildSmartGenerationContext(ctx, req.CharacterID, req.UserID, req.BasedOnChat)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to build smart context: %v", err)
		return result, err
	}

	// 2. 选择合适的AI模型
	selectedModel := s.selectBestModel(smartContext)
	if selectedModel == nil {
		result.Message = "No suitable AI model found"
		return result, fmt.Errorf("no suitable AI model found")
	}

	s.logger.WithField("selected_model", selectedModel.DisplayName).Info("Selected AI model for generation")

	// 3. 生成智能朋友圈内容
	contents, mentions, err := s.generateSmartContent(ctx, smartContext, req)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to generate smart content: %v", err)
		return result, err
	}

	// 4. 创建草稿
	var drafts []*domain.MomentDraft
	for i, content := range contents {
		draft := &domain.MomentDraft{
			ID:          uuid.New(),
			CharacterID: req.CharacterID,
			UserID:      req.UserID,
			Content:     content,
			ContentType: domain.MomentContentTypeText,
			Visibility:  domain.MomentVisibilityFriends,
			Priority:    len(contents) - i,
			GeneratedAt: time.Now(),
		}

		// 设置心情和标签
		s.setSmartMoodAndTags(draft, smartContext, content)

		// 保存草稿
		err := s.momentsRepo.CreateMomentDraft(ctx, draft)
		if err != nil {
			s.logger.WithError(err).WithField("draft_id", draft.ID).Warn("Failed to save draft")
			continue
		}

		drafts = append(drafts, draft)
	}

	result.Success = len(drafts) > 0
	result.Drafts = drafts
	result.TotalCount = len(drafts)

	if result.Success {
		result.Message = fmt.Sprintf("Successfully generated %d smart drafts", len(drafts))

		// 5. 如果启用自动互动，安排互动任务
		if req.AutoInteract && len(drafts) > 0 {
			s.scheduleAutoInteractions(ctx, drafts[0], smartContext)
		}
	} else {
		result.Message = "No drafts were generated"
	}

	s.logger.WithFields(logrus.Fields{
		"character_id":    req.CharacterID,
		"generated_count": len(drafts),
		"mentions_count":  len(mentions),
	}).Info("Smart moment generation completed")

	return result, nil
}

// BuildSmartGenerationContext 构建智能生成上下文
func (s *smartMomentsService) BuildSmartGenerationContext(ctx context.Context, characterID, userID uuid.UUID, basedOnChat bool) (*domain.SmartMomentGeneration, error) {
	context := &domain.SmartMomentGeneration{
		ChatBasedContent: basedOnChat,
	}

	// 1. 获取角色信息
	character, err := s.characterRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}
	context.Character = s.convertToCharacterResponse(character)

	// 2. 获取角色选择的聊天模型
	selectedModel, err := s.getCharacterChatModel(ctx, characterID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get character chat model")
	}
	context.SelectedModel = selectedModel

	// 3. 获取可用的聊天模型
	availableModels, err := s.getAvailableChatModels(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get available chat models")
	}
	context.AvailableModels = availableModels

	// 4. 获取关系网络图谱 - 暂时跳过
	// TODO: 实现关系网络图谱获取
	s.logger.WithField("user_id", userID).Info("Skipping relationship map retrieval")

	// 5. 获取最近聊天记录（如果基于聊天生成）
	if basedOnChat {
		recentChats, err := s.getRecentChats(ctx, characterID, userID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get recent chats")
		}
		context.RecentChats = recentChats
	}

	// 6. 构建@提及候选列表 - 暂时跳过
	// TODO: 实现@提及候选列表构建

	// 7. 获取自动互动配置
	autoInteractions, err := s.getAutoInteractionConfigs(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get auto interaction configs")
	}
	context.AutoInteractions = autoInteractions

	return context, nil
}

// selectBestModel 选择最佳AI模型
func (s *smartMomentsService) selectBestModel(context *domain.SmartMomentGeneration) *domain.AIModel {
	// 优先使用角色选择的聊天模型
	if context.SelectedModel != nil {
		return context.SelectedModel
	}

	// 从可用模型中选择支持聊天的模型
	for _, model := range context.AvailableModels {
		// 暂时简化处理，直接返回第一个模型
		return model
	}

	return nil
}

// generateSmartContent 生成智能内容
func (s *smartMomentsService) generateSmartContent(ctx context.Context, smartContext *domain.SmartMomentGeneration, req *domain.MomentGenerationRequest) ([]string, []string, error) {
	// 使用DeepSeek生成内容
	characterName := smartContext.Character.Name
	personality := ""
	if smartContext.Character.Personality != nil {
		personality = *smartContext.Character.Personality
	}

	// 获取用户在这个角色中的身份（使用新的身份识别服务）
	userIdentityContext, err := s.userIdentityService.BuildUserIdentityContext(ctx, req.UserID, &smartContext.Character.ID, nil)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to build user identity context, using fallback")
		// 使用回退逻辑
		userIdentity := s.getUserIdentityForCharacter(ctx, req.UserID, smartContext.Character)
		userIdentityContext = &domain.UserIdentityContext{
			UserID: req.UserID,
			CurrentIdentity: &domain.UserIdentity{
				Type:        domain.IdentityTypeSpecified,
				DisplayName: userIdentity,
				Source:      "fallback",
			},
		}
	}

	// 构建聊天上下文
	var chatContext []string
	if smartContext.ChatBasedContent && len(smartContext.RecentChats) > 0 {
		for _, chat := range smartContext.RecentChats {
			// 智能替换用户身份
			content, err := s.userIdentityService.ReplaceUserReferences(ctx, chat.Content, req.UserID, userIdentityContext.CurrentIdentity)
			if err != nil {
				content = chat.Content // 使用原始内容作为回退
			}
			senderName := userIdentityContext.CurrentIdentity.DisplayName // 使用用户身份作为发送者名称
			// TODO: 根据chat.SenderType和chat.SenderID获取真实发送者名称
			chatContext = append(chatContext, fmt.Sprintf("%s: %s", senderName, content))
		}
	}

	contents, err := s.deepSeekClient.GenerateMomentContent(
		ctx,
		characterName,
		personality,
		chatContext,
		req.Count,
	)

	if err != nil {
		return nil, nil, err
	}

	// 智能替换用户身份
	contentsWithUserIdentity := s.processContentWithNicknameReplacement(contents, userIdentityContext.CurrentIdentity.DisplayName)

	// 处理@提及
	var mentions []string
	processedContents := make([]string, len(contentsWithUserIdentity))
	for i, content := range contentsWithUserIdentity {
		processedContent, contentMentions := s.processSmartMentions(content, smartContext)
		processedContents[i] = processedContent
		mentions = append(mentions, contentMentions...)
	}

	return processedContents, mentions, nil
}

// buildSmartPrompt 构建智能提示词
func (s *smartMomentsService) buildSmartPrompt(context *domain.SmartMomentGeneration, req *domain.MomentGenerationRequest) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString(fmt.Sprintf("请为角色 %s 生成朋友圈内容。\n\n", context.Character.Name))

	if context.Character.Personality != nil {
		promptBuilder.WriteString(fmt.Sprintf("角色性格：%s\n\n", *context.Character.Personality))
	}

	// 添加关系网络信息 - 暂时跳过
	if len(context.MentionCandidates) > 0 {
		promptBuilder.WriteString("关系网络中的朋友：\n")
		for _, candidate := range context.MentionCandidates {
			promptBuilder.WriteString(fmt.Sprintf("- %s", candidate.Character.Name))
			if candidate.Relationship.CustomTypeName != nil && *candidate.Relationship.CustomTypeName != "" {
				promptBuilder.WriteString(fmt.Sprintf("（%s）", *candidate.Relationship.CustomTypeName))
			}
			promptBuilder.WriteString("\n")
		}
		promptBuilder.WriteString("\n")
	}

	// 添加聊天上下文 - 暂时跳过
	if context.ChatBasedContent && len(context.RecentChats) > 0 {
		promptBuilder.WriteString("最近的聊天内容：\n")
		promptBuilder.WriteString("- 用户: 最近有聊天记录\n")
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("请生成%d条朋友圈内容，要求：\n", req.Count))
	promptBuilder.WriteString("1. 符合角色性格特点\n")
	promptBuilder.WriteString("2. 内容自然真实，像真人发的朋友圈\n")
	promptBuilder.WriteString("3. 每条内容20-100字\n")
	promptBuilder.WriteString("4. 可以包含适当的emoji表情\n")

	if context.ChatBasedContent {
		promptBuilder.WriteString("5. 基于最近的聊天内容生成相关朋友圈\n")
		promptBuilder.WriteString("6. 如果聊天中提到了朋友，可以在朋友圈中@他们\n")
	} else {
		promptBuilder.WriteString("5. 基于角色设定和关系网络生成日常朋友圈\n")
		promptBuilder.WriteString("6. 可以偶尔@关系网络中的朋友\n")
	}

	promptBuilder.WriteString("7. 内容要有一定的多样性\n")
	promptBuilder.WriteString("8. 避免重复和模板化\n\n")

	if len(context.MentionCandidates) > 0 {
		promptBuilder.WriteString("注意：如果要@朋友，请使用格式：@朋友名字\n\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("请直接返回%d条朋友圈内容，每条内容占一行。", req.Count))

	return promptBuilder.String()
}

// processSmartMentions 处理智能@提及
func (s *smartMomentsService) processSmartMentions(content string, context *domain.SmartMomentGeneration) (string, []string) {
	var mentions []string

	// 使用正则表达式查找@提及
	mentionRegex := regexp.MustCompile(`@([^\s@]+)`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			mentionedName := match[1]

			// 在关系网络中查找对应的角色
			for _, candidate := range context.MentionCandidates {
				if candidate.Character.Name == mentionedName {
					mentions = append(mentions, candidate.Character.ID.String())
					break
				}
			}
		}
	}

	return content, mentions
}

// 辅助方法

// convertToCharacterResponse 转换角色响应
func (s *smartMomentsService) convertToCharacterResponse(character *domain.Character) *domain.CharacterResponse {
	return &domain.CharacterResponse{
		Character: character,
		Tags:      []string{},
		IsOwner:   false,
		CanEdit:   false,
	}
}

// getCharacterChatModel 获取角色聊天模型
func (s *smartMomentsService) getCharacterChatModel(ctx context.Context, characterID uuid.UUID) (*domain.AIModel, error) {
	// 这里应该从角色配置中获取选择的聊天模型
	// 暂时返回nil，表示没有特定选择
	return nil, nil
}

// getAvailableChatModels 获取可用的聊天模型
func (s *smartMomentsService) getAvailableChatModels(ctx context.Context, userID uuid.UUID) ([]*domain.AIModel, error) {
	// 这里应该获取用户可用的聊天模型
	// 暂时返回空列表
	return []*domain.AIModel{}, nil
}

// getRecentChats 获取最近聊天记录
func (s *smartMomentsService) getRecentChats(ctx context.Context, characterID, userID uuid.UUID) ([]*domain.ChatMessage, error) {
	// 这里应该从聊天服务获取最近的聊天记录
	// 暂时返回空列表
	return []*domain.ChatMessage{}, nil
}

// buildMentionCandidates 构建@提及候选列表
func (s *smartMomentsService) buildMentionCandidates(relationshipMap *domain.UserRelationshipMap) []*domain.RelationshipCharacter {
	var candidates []*domain.RelationshipCharacter

	for _, group := range relationshipMap.RelationshipGroups {
		for _, character := range group.Characters {
			candidates = append(candidates, character)
		}
	}

	return candidates
}

// getAutoInteractionConfigs 获取自动互动配置
func (s *smartMomentsService) getAutoInteractionConfigs(ctx context.Context, userID uuid.UUID) ([]*domain.MomentAutoInteraction, error) {
	// 这里应该获取自动互动配置
	// 暂时返回空列表
	return []*domain.MomentAutoInteraction{}, nil
}

// setSmartMoodAndTags 设置智能心情和标签
func (s *smartMomentsService) setSmartMoodAndTags(draft *domain.MomentDraft, context *domain.SmartMomentGeneration, content string) {
	// 基于内容和上下文智能设置心情和标签
	content = strings.ToLower(content)

	if strings.Contains(content, "开心") || strings.Contains(content, "高兴") || strings.Contains(content, "😊") {
		mood := "happy"
		draft.Mood = &mood
		draft.Tags = []string{"开心", "心情好"}
	} else if strings.Contains(content, "@") {
		mood := "social"
		draft.Mood = &mood
		draft.Tags = []string{"朋友", "社交"}
	} else if context.ChatBasedContent {
		mood := "interactive"
		draft.Mood = &mood
		draft.Tags = []string{"聊天", "互动"}
	} else {
		mood := "daily"
		draft.Mood = &mood
		draft.Tags = []string{"日常", "生活"}
	}
}

// scheduleAutoInteractions 安排自动互动
func (s *smartMomentsService) scheduleAutoInteractions(ctx context.Context, draft *domain.MomentDraft, context *domain.SmartMomentGeneration) {
	s.logger.WithFields(logrus.Fields{
		"draft_id":          draft.ID,
		"auto_interactions": len(context.AutoInteractions),
	}).Info("Scheduling auto interactions")

	// 为关系网络中的每个角色安排互动任务
	for _, candidate := range context.MentionCandidates {
		// 根据关系强度和配置决定互动概率
		s.scheduleCharacterInteraction(ctx, draft, candidate, context)
	}
}

// TriggerAutoInteractions 触发自动互动
func (s *smartMomentsService) TriggerAutoInteractions(ctx context.Context, momentID uuid.UUID) error {
	s.logger.WithField("moment_id", momentID).Info("Triggering auto interactions")

	// 获取朋友圈信息
	moment, err := s.momentsRepo.GetMoment(ctx, momentID)
	if err != nil {
		return fmt.Errorf("failed to get moment: %w", err)
	}

	// 获取关系网络 - 暂时使用简化版本
	// TODO: 实现完整的关系网络获取
	s.logger.WithField("user_id", moment.UserID).Info("Getting relationship map for auto interactions")

	// 暂时跳过自动互动，等待关系网络服务完善
	return nil
}

// ProcessMentions 处理@提及
func (s *smartMomentsService) ProcessMentions(ctx context.Context, momentID uuid.UUID, content string) error {
	s.logger.WithField("moment_id", momentID).Info("Processing mentions")

	// 解析@提及
	mentions := s.extractMentions(content)

	// 为每个被@的角色创建通知和自动回复
	for _, mention := range mentions {
		// 创建提及通知
		err := s.createMentionNotification(ctx, momentID, mention)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to create mention notification")
		}

		// 触发被@角色的自动回复
		go s.triggerMentionResponse(ctx, momentID, mention)
	}

	return nil
}

// executeAutoInteraction 执行自动互动
func (s *smartMomentsService) executeAutoInteraction(ctx context.Context, moment *domain.Moment, character *domain.RelationshipCharacter) {
	// 添加随机延迟，模拟真实用户行为
	delay := time.Duration(rand.Intn(300)+30) * time.Second // 30秒到5分钟
	time.Sleep(delay)

	// 根据关系强度决定互动类型和概率
	relationship := character.Relationship

	// 点赞概率基于关系强度
	likeProbability := float64(relationship.Strength) / 100.0
	if rand.Float64() < likeProbability {
		s.autoLikeMoment(ctx, moment.ID, character.Character.ID)
	}

	// 评论概率较低，但基于亲密度
	commentProbability := float64(relationship.Intimacy) / 200.0 // 降低评论概率
	if rand.Float64() < commentProbability {
		s.autoCommentMoment(ctx, moment, character)
	}

	// 分享概率最低
	shareProbability := float64(relationship.Strength) / 500.0
	if rand.Float64() < shareProbability {
		s.autoShareMoment(ctx, moment.ID, character.Character.ID)
	}
}

// autoLikeMoment 自动点赞
func (s *smartMomentsService) autoLikeMoment(ctx context.Context, momentID, characterID uuid.UUID) {
	interaction := &domain.MomentInteraction{
		ID:          uuid.New(),
		MomentID:    momentID,
		CharacterID: &characterID,
		Type:        domain.InteractionTypeLike,
		CreatedAt:   time.Now(),
	}

	err := s.momentsRepo.CreateInteraction(ctx, interaction)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to auto like moment")
		return
	}

	// 更新点赞数
	moment, err := s.momentsRepo.GetMoment(ctx, momentID)
	if err != nil {
		return
	}

	moment.LikeCount++
	s.momentsRepo.UpdateMoment(ctx, moment)

	s.logger.WithFields(logrus.Fields{
		"moment_id":    momentID,
		"character_id": characterID,
	}).Info("Auto liked moment")
}

// autoCommentMoment 自动评论
func (s *smartMomentsService) autoCommentMoment(ctx context.Context, moment *domain.Moment, character *domain.RelationshipCharacter) {
	// 生成智能评论内容
	comment := s.generateSmartComment(moment, character)

	interaction := &domain.MomentInteraction{
		ID:          uuid.New(),
		MomentID:    moment.ID,
		CharacterID: &character.Character.ID,
		Type:        domain.InteractionTypeComment,
		Content:     &comment,
		CreatedAt:   time.Now(),
	}

	err := s.momentsRepo.CreateInteraction(ctx, interaction)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to auto comment moment")
		return
	}

	// 更新评论数
	moment.CommentCount++
	s.momentsRepo.UpdateMoment(ctx, moment)

	s.logger.WithFields(logrus.Fields{
		"moment_id":    moment.ID,
		"character_id": character.Character.ID,
		"comment":      comment,
	}).Info("Auto commented moment")
}

// autoShareMoment 自动分享
func (s *smartMomentsService) autoShareMoment(ctx context.Context, momentID, characterID uuid.UUID) {
	interaction := &domain.MomentInteraction{
		ID:          uuid.New(),
		MomentID:    momentID,
		CharacterID: &characterID,
		Type:        domain.InteractionTypeShare,
		CreatedAt:   time.Now(),
	}

	err := s.momentsRepo.CreateInteraction(ctx, interaction)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to auto share moment")
		return
	}

	// 更新分享数
	moment, err := s.momentsRepo.GetMoment(ctx, momentID)
	if err != nil {
		return
	}

	moment.ShareCount++
	s.momentsRepo.UpdateMoment(ctx, moment)

	s.logger.WithFields(logrus.Fields{
		"moment_id":    momentID,
		"character_id": characterID,
	}).Info("Auto shared moment")
}

// generateSmartComment 生成智能评论
func (s *smartMomentsService) generateSmartComment(moment *domain.Moment, character *domain.RelationshipCharacter) string {
	// 基于关系类型和朋友圈内容生成合适的评论
	content := strings.ToLower(moment.Content)
	relationship := character.Relationship

	var comments []string

	// 根据关系类型选择评论风格
	customTypeName := ""
	if relationship.CustomTypeName != nil {
		customTypeName = *relationship.CustomTypeName
	}

	if customTypeName == "好朋友" || customTypeName == "闺蜜" {
		comments = []string{
			"哈哈哈，太有趣了！",
			"我也想去！",
			"下次一起呀~",
			"羡慕羡慕",
			"你总是这么有意思",
		}
	} else if customTypeName == "同事" {
		comments = []string{
			"不错不错👍",
			"看起来很棒",
			"厉害了",
			"学习了",
		}
	} else if customTypeName == "家人" {
		comments = []string{
			"注意身体哦",
			"很棒！",
			"开心就好",
			"记得早点休息",
		}
	} else {
		// 默认评论
		comments = []string{
			"👍",
			"不错呀",
			"赞！",
			"很棒",
			"有意思",
		}
	}

	// 根据内容调整评论
	if strings.Contains(content, "美食") || strings.Contains(content, "吃") {
		comments = append(comments, "看起来很好吃", "流口水了", "在哪里吃的？")
	} else if strings.Contains(content, "旅行") || strings.Contains(content, "风景") {
		comments = append(comments, "好美的风景", "想去！", "拍得真好")
	} else if strings.Contains(content, "工作") || strings.Contains(content, "学习") {
		comments = append(comments, "加油！", "辛苦了", "很棒！")
	}

	// 随机选择一个评论
	return comments[rand.Intn(len(comments))]
}

// 辅助方法

// scheduleCharacterInteraction 安排角色互动
func (s *smartMomentsService) scheduleCharacterInteraction(ctx context.Context, draft *domain.MomentDraft, character *domain.RelationshipCharacter, context *domain.SmartMomentGeneration) {
	// 根据关系强度决定是否安排互动
	if character.Relationship.Strength < 30 { // 关系强度太低，不互动
		return
	}

	// 安排延迟互动任务
	delay := time.Duration(rand.Intn(1800)+300) * time.Second // 5分钟到30分钟

	task := &domain.MomentInteractionTask{
		ID:              uuid.New(),
		MomentID:        draft.ID, // 这里应该是发布后的moment ID
		CharacterID:     character.Character.ID,
		InteractionType: "like", // 默认点赞
		ScheduledAt:     time.Now().Add(delay),
		Status:          "pending",
		CreatedAt:       time.Now(),
	}

	// 这里应该将任务保存到数据库或任务队列
	s.logger.WithFields(logrus.Fields{
		"task_id":      task.ID,
		"character_id": character.Character.ID,
		"scheduled_at": task.ScheduledAt,
	}).Info("Scheduled interaction task")
}

// extractMentions 提取@提及
func (s *smartMomentsService) extractMentions(content string) []string {
	mentionRegex := regexp.MustCompile(`@([^\s@]+)`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	var mentions []string
	for _, match := range matches {
		if len(match) > 1 {
			mentions = append(mentions, match[1])
		}
	}

	return mentions
}

// createMentionNotification 创建提及通知
func (s *smartMomentsService) createMentionNotification(ctx context.Context, momentID uuid.UUID, mentionedName string) error {
	// 这里应该根据提及的名字找到对应的用户或角色，并创建通知
	s.logger.WithFields(logrus.Fields{
		"moment_id":      momentID,
		"mentioned_name": mentionedName,
	}).Info("Creating mention notification")

	return nil
}

// triggerMentionResponse 触发提及回复
func (s *smartMomentsService) triggerMentionResponse(ctx context.Context, momentID uuid.UUID, mentionedName string) {
	// 被@的角色有更高概率进行互动
	delay := time.Duration(rand.Intn(300)+60) * time.Second // 1-5分钟内回复
	time.Sleep(delay)

	s.logger.WithFields(logrus.Fields{
		"moment_id":      momentID,
		"mentioned_name": mentionedName,
	}).Info("Triggered mention response")
}

// getUserNickname 获取用户昵称
func (s *smartMomentsService) getUserNickname(ctx context.Context, userID uuid.UUID) string {
	// TODO: 从用户服务获取用户昵称
	// 暂时返回默认昵称，实际应该从数据库获取
	return "小主" // 默认昵称
}

// extractUserIdentityFromCharacterSetting 从角色设定中提取用户身份
func (s *smartMomentsService) extractUserIdentityFromCharacterSetting(character *domain.CharacterResponse) string {
	if character.Personality == nil {
		return ""
	}

	setting := *character.Personality

	// 匹配各种用户身份格式
	patterns := []string{
		`用户（([^）]+)）`,        // 用户（李明）
		`用户\(([^)]+)\)`,      // 用户(李明)
		`用户：([^\s，。！？]+)`,    // 用户：李明
		`用户:([^\s，。！？]+)`,    // 用户:李明
		`([^\s，。！？]+)\(用户\)`, // 李明(用户)
		`([^\s，。！？]+)（用户）`,   // 李明（用户）
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(setting)
		if len(matches) > 1 {
			userName := strings.TrimSpace(matches[1])
			if userName != "" {
				s.logger.WithFields(logrus.Fields{
					"character_id": character.ID,
					"user_name":    userName,
					"pattern":      pattern,
				}).Info("Extracted user identity from character setting")
				return userName
			}
		}
	}

	return ""
}

// getUserIdentityForCharacter 获取用户在特定角色中的身份
func (s *smartMomentsService) getUserIdentityForCharacter(ctx context.Context, userID uuid.UUID, character *domain.CharacterResponse) string {
	// 首先尝试从角色设定中提取用户身份
	userIdentity := s.extractUserIdentityFromCharacterSetting(character)

	if userIdentity != "" {
		s.logger.WithFields(logrus.Fields{
			"user_id":       userID,
			"character_id":  character.ID,
			"user_identity": userIdentity,
		}).Info("Using specified user identity from character setting")
		return userIdentity
	}

	// 如果没有指定，使用用户真实昵称
	userNickname := s.getUserNickname(ctx, userID)
	s.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"character_id":  character.ID,
		"user_nickname": userNickname,
	}).Info("Using default user nickname")

	return userNickname
}

// getUserRoleInRelationship 获取用户在关系网络中的角色身份
func (s *smartMomentsService) getUserRoleInRelationship(ctx context.Context, userID uuid.UUID, characterID uuid.UUID) string {
	// TODO: 从关系网络中查找用户是否有指定的角色扮演
	// 例如：关系网络中设置了"用户（李明）"，则返回"李明"
	// 如果没有指定，返回用户真实昵称

	// 暂时模拟逻辑
	userNickname := s.getUserNickname(ctx, userID)

	// 检查是否有角色扮演设置
	rolePlayName := s.checkRolePlaySetting(ctx, userID, characterID)
	if rolePlayName != "" {
		return rolePlayName
	}

	return userNickname
}

// checkRolePlaySetting 检查角色扮演设置
func (s *smartMomentsService) checkRolePlaySetting(ctx context.Context, userID uuid.UUID, characterID uuid.UUID) string {
	// TODO: 从数据库查询用户在特定角色关系网络中的扮演设置
	// 例如：在角色A的关系网络中，用户被设置为扮演"李明"
	// 返回格式：如果设置了"用户（李明）"，则返回"李明"

	// 暂时返回空，表示没有特殊设置
	return ""
}

// replaceUserReferences 智能替换用户引用
func (s *smartMomentsService) replaceUserReferences(content, userIdentity string) string {
	// 替换各种用户引用方式
	replacements := map[string]string{
		"@用户":   "@" + userIdentity,
		"用户":    userIdentity,
		"@User": "@" + userIdentity,
		"User":  userIdentity,
		"主人":    userIdentity,
		"@主人":   "@" + userIdentity,
		"@朋友":   "@" + userIdentity, // 如果关系网络中用户被设为朋友
		"朋友":    userIdentity,
	}

	result := content
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// replaceUserReferencesInRelationship 在关系网络上下文中智能替换用户引用
func (s *smartMomentsService) replaceUserReferencesInRelationship(content string, userID, characterID uuid.UUID, ctx context.Context) string {
	// 获取用户在这个角色关系网络中的身份
	userIdentity := s.getUserRoleInRelationship(ctx, userID, characterID)

	// 执行智能替换
	return s.replaceUserReferences(content, userIdentity)
}

// processContentWithNicknameReplacement 处理内容并替换昵称
func (s *smartMomentsService) processContentWithNicknameReplacement(contents []string, userNickname string) []string {
	var processedContents []string

	for _, content := range contents {
		// 智能替换用户引用
		processedContent := s.replaceUserReferences(content, userNickname)
		processedContents = append(processedContents, processedContent)
	}

	return processedContents
}
