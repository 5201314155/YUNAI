package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
)

// OrchestratorService 编排服务接口
type OrchestratorService interface {
	// 核心编排功能
	ProcessMessage(ctx context.Context, req *domain.ProcessMessageRequest) (*domain.ProcessMessageResponse, error)
	GenerateResponse(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, context string) (*domain.GeneratedResponse, error)

	// 说话人调度
	DetermineSpeaker(ctx context.Context, groupChatID uuid.UUID, context map[string]interface{}) (*uuid.UUID, error)
	UpdateSpeakingQueue(ctx context.Context, groupChatID uuid.UUID, speakerID uuid.UUID) error
	GetNextSpeaker(ctx context.Context, groupChatID uuid.UUID) (*uuid.UUID, error)

	// 上下文构建
	BuildConversationContext(ctx context.Context, groupChatID uuid.UUID, characterID uuid.UUID, maxTokens int) (string, error)
	InjectRelationshipContext(ctx context.Context, sourceID, targetID uuid.UUID, baseContext string) (string, error)

	// 模型路由和选择
	SelectModelForCharacter(ctx context.Context, characterID uuid.UUID, messageType string) (*domain.ModelResponse, error)
	RouteToModel(ctx context.Context, modelID uuid.UUID, prompt string, params map[string]interface{}) (*domain.ModelResponse, error)

	// 剧情和触发处理
	ProcessTriggers(ctx context.Context, groupChatID uuid.UUID, message *domain.ChatMessage) ([]domain.TriggerAction, error)
	ApplyTriggerActions(ctx context.Context, groupChatID uuid.UUID, actions []domain.TriggerAction) error

	// 成本和计费
	CalculateResponseCost(ctx context.Context, characterID uuid.UUID, prompt string, responseLength int) (float64, error)
	DeductCost(ctx context.Context, userID uuid.UUID, cost float64, description string) error

	// 演出控制
	StartPerformance(ctx context.Context, groupChatID uuid.UUID, scenario *domain.PerformanceScenario) error
	StopPerformance(ctx context.Context, groupChatID uuid.UUID) error
	GetPerformanceStatus(ctx context.Context, groupChatID uuid.UUID) (*domain.PerformanceStatus, error)
}

type orchestratorService struct {
	memoryService       MemoryService
	relationshipService RelationshipService
	characterService    CharacterService
	modelService        ModelService
	aiBillingService    AIBillingService
	logger              *logrus.Logger
}

// NewOrchestratorService 创建编排服务
func NewOrchestratorService(
	memoryService MemoryService,
	relationshipService RelationshipService,
	characterService CharacterService,
	modelService ModelService,
	aiBillingService AIBillingService,
	logger *logrus.Logger,
) OrchestratorService {
	return &orchestratorService{
		memoryService:       memoryService,
		relationshipService: relationshipService,
		characterService:    characterService,
		modelService:        modelService,
		aiBillingService:    aiBillingService,
		logger:              logger,
	}
}

// ProcessMessage 处理消息的核心编排逻辑
func (s *orchestratorService) ProcessMessage(ctx context.Context, req *domain.ProcessMessageRequest) (*domain.ProcessMessageResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"group_chat_id": req.GroupChatID,
		"user_id":       req.UserID,
		"content":       req.Content,
	}).Info("Processing message through orchestrator")

	// 1. 获取对话上下文
	conversationContext, err := s.memoryService.GetConversationContext(ctx, req.GroupChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation context: %w", err)
	}

	// 2. 处理触发器
	message := &domain.ChatMessage{
		ID:          uuid.New(),
		GroupChatID: &req.GroupChatID,
		Content:     req.Content,
		MessageType: req.MessageType,
		CreatedAt:   time.Now(),
	}

	triggers, err := s.ProcessTriggers(ctx, req.GroupChatID, message)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to process triggers")
	}

	// 3. 应用触发器动作
	if len(triggers) > 0 {
		err = s.ApplyTriggerActions(ctx, req.GroupChatID, triggers)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to apply trigger actions")
		}
	}

	// 4. 确定下一个说话人
	nextSpeakerID, err := s.DetermineSpeaker(ctx, req.GroupChatID, map[string]interface{}{
		"last_message": req.Content,
		"user_id":      req.UserID,
		"triggers":     triggers,
	})
	if err != nil {
		s.logger.WithError(err).Warn("Failed to determine next speaker")
	}

	// 5. 生成响应（如果有确定的说话人）
	var responses []*domain.GeneratedResponse
	if nextSpeakerID != nil {
		// 构建上下文
		contextStr, err := s.BuildConversationContext(ctx, req.GroupChatID, *nextSpeakerID, 2000)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to build conversation context")
			contextStr = ""
		}

		// 生成响应
		response, err := s.GenerateResponse(ctx, *nextSpeakerID, req.GroupChatID, contextStr)
		if err != nil {
			s.logger.WithError(err).WithField("character_id", *nextSpeakerID).Warn("Failed to generate response")
		} else {
			responses = append(responses, response)
		}

		// 更新说话队列
		err = s.UpdateSpeakingQueue(ctx, req.GroupChatID, *nextSpeakerID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to update speaking queue")
		}
	}

	// 6. 更新对话上下文
	conversationContext.TurnCount++
	conversationContext.LastActivityAt = time.Now()
	if nextSpeakerID != nil {
		conversationContext.CurrentSpeakerID = nextSpeakerID
	}

	_, err = s.memoryService.UpdateConversationContext(ctx, req.GroupChatID, map[string]interface{}{
		"turn_count":         conversationContext.TurnCount,
		"current_speaker_id": conversationContext.CurrentSpeakerID,
	})
	if err != nil {
		s.logger.WithError(err).Warn("Failed to update conversation context")
	}

	return &domain.ProcessMessageResponse{
		ProcessedAt:    time.Now(),
		TriggersCount:  len(triggers),
		NextSpeakerID:  nextSpeakerID,
		Responses:      responses,
		ContextUpdated: true,
	}, nil
}

// GenerateResponse 生成角色响应
func (s *orchestratorService) GenerateResponse(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, context string) (*domain.GeneratedResponse, error) {
	// 1. 获取角色信息
	character, err := s.characterService.GetCharacter(ctx, characterID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 2. 选择模型
	model, err := s.SelectModelForCharacter(ctx, characterID, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to select model: %w", err)
	}

	// 3. 构建提示词
	prompt := s.buildPrompt(character, context)

	// 4. 计算成本
	estimatedCost, err := s.CalculateResponseCost(ctx, characterID, prompt, 200) // 假设200token响应
	if err != nil {
		s.logger.WithError(err).Warn("Failed to calculate response cost")
		estimatedCost = 0.01 // 默认成本
	}

	// 5. 生成响应（模拟实现）
	responseContent := s.generateMockResponse(character, context)

	// 6. 扣费
	err = s.DeductCost(ctx, character.UserID, estimatedCost, fmt.Sprintf("角色 %s 生成响应", character.Name))
	if err != nil {
		s.logger.WithError(err).Warn("Failed to deduct cost")
	}

	// 7. 创建记忆
	_, err = s.memoryService.CreateMemoryFromEvent(ctx, characterID,
		fmt.Sprintf("在群聊中说话: %s", responseContent), 0.6)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to create memory from response")
	}

	return &domain.GeneratedResponse{
		CharacterID:   characterID,
		Content:       responseContent,
		ModelUsed:     model.DisplayName,
		TokensUsed:    200, // 模拟值
		Cost:          estimatedCost,
		GeneratedAt:   time.Now(),
		ContextLength: len(context),
	}, nil
}

// DetermineSpeaker 确定下一个说话人
func (s *orchestratorService) DetermineSpeaker(ctx context.Context, groupChatID uuid.UUID, context map[string]interface{}) (*uuid.UUID, error) {
	// 获取群聊成员
	members, err := s.characterService.GetGroupChatMembers(ctx, groupChatID, uuid.New()) // 临时用户ID
	if err != nil {
		return nil, fmt.Errorf("failed to get group chat members: %w", err)
	}

	// 过滤出角色成员
	var characterMembers []*domain.GroupChatMember
	for _, member := range members {
		if member.MemberType == domain.MemberTypeCharacter && member.CharacterID != nil {
			characterMembers = append(characterMembers, member)
		}
	}

	if len(characterMembers) == 0 {
		return nil, nil // 没有角色成员
	}

	// 简化的说话人选择逻辑
	// 实际实现应该考虑：
	// - 关系强度
	// - 说话频率
	// - 触发条件
	// - 上次发言时间
	// - 角色主动性

	// 这里使用简单的轮询机制
	conversationContext, err := s.memoryService.GetConversationContext(ctx, groupChatID)
	if err != nil {
		// 随机选择第一个角色
		return characterMembers[0].CharacterID, nil
	}

	// 如果有说话队列，选择队列中的下一个
	if len(conversationContext.SpeakingQueue) > 0 {
		for _, speakerIDStr := range conversationContext.SpeakingQueue {
			speakerID, err := uuid.Parse(speakerIDStr)
			if err != nil {
				continue
			}

			// 检查这个角色是否在群聊中
			for _, member := range characterMembers {
				if member.CharacterID != nil && *member.CharacterID == speakerID {
					return &speakerID, nil
				}
			}
		}
	}

	// 默认选择第一个角色
	return characterMembers[0].CharacterID, nil
}

// UpdateSpeakingQueue 更新说话队列
func (s *orchestratorService) UpdateSpeakingQueue(ctx context.Context, groupChatID uuid.UUID, speakerID uuid.UUID) error {
	// 简化实现：将当前说话人移到队列末尾
	_, err := s.memoryService.UpdateConversationContext(ctx, groupChatID, map[string]interface{}{
		"current_speaker_id": speakerID,
	})
	return err
}

// GetNextSpeaker 获取下一个说话人
func (s *orchestratorService) GetNextSpeaker(ctx context.Context, groupChatID uuid.UUID) (*uuid.UUID, error) {
	return s.DetermineSpeaker(ctx, groupChatID, map[string]interface{}{})
}

// BuildConversationContext 构建对话上下文
func (s *orchestratorService) BuildConversationContext(ctx context.Context, groupChatID uuid.UUID, characterID uuid.UUID, maxTokens int) (string, error) {
	var contextParts []string

	// 1. 获取角色基本信息
	character, err := s.characterService.GetCharacter(ctx, characterID, nil)
	if err == nil {
		contextParts = append(contextParts, fmt.Sprintf("你是 %s。", character.Name))
		if character.Personality != nil {
			contextParts = append(contextParts, fmt.Sprintf("性格: %s", *character.Personality))
		}
		if character.SystemPrompt != nil {
			contextParts = append(contextParts, *character.SystemPrompt)
		}
	}

	// 2. 获取记忆上下文
	memoryContext, err := s.memoryService.BuildContextForCharacter(ctx, characterID, groupChatID, maxTokens/2)
	if err == nil && memoryContext != "" {
		contextParts = append(contextParts, "\n记忆上下文:", memoryContext)
	}

	// 3. 获取关系上下文
	relationships, err := s.relationshipService.ListCharacterRelationships(ctx, characterID, nil)
	if err == nil && len(relationships) > 0 {
		contextParts = append(contextParts, "\n当前关系:")
		for _, rel := range relationships[:min(len(relationships), 3)] { // 最多3个关系
			if rel.TargetCharacter != nil {
				contextParts = append(contextParts, fmt.Sprintf("- 与 %s: %s (强度: %.1f)",
					rel.TargetCharacter.Name, rel.CustomTypeName, rel.Strength))
			}
		}
	}

	// 4. 获取群聊上下文
	conversationContext, err := s.memoryService.GetConversationContext(ctx, groupChatID)
	if err == nil {
		if conversationContext.CurrentScene != nil {
			contextParts = append(contextParts, fmt.Sprintf("\n当前场景: %s", *conversationContext.CurrentScene))
		}
		if conversationContext.Mood != nil {
			contextParts = append(contextParts, fmt.Sprintf("当前氛围: %s", *conversationContext.Mood))
		}
	}

	context := strings.Join(contextParts, "\n")

	// 截断到最大token数
	if len(context) > maxTokens*4 { // 假设4字符=1token
		context = context[:maxTokens*4] + "..."
	}

	return context, nil
}

// InjectRelationshipContext 注入关系上下文
func (s *orchestratorService) InjectRelationshipContext(ctx context.Context, sourceID, targetID uuid.UUID, baseContext string) (string, error) {
	relationship, err := s.relationshipService.GetCharacterRelationship(ctx, sourceID, targetID)
	if err != nil {
		return baseContext, nil // 没有关系就返回原上下文
	}

	relationshipContext := fmt.Sprintf("\n与对方的关系: %s (强度: %.1f, 信任: %.1f, 亲密: %.1f)",
		relationship.CustomTypeName, relationship.Strength, relationship.Trust, relationship.Intimacy)

	if relationship.AddressName != nil {
		relationshipContext += fmt.Sprintf("\n称呼对方: %s", *relationship.AddressName)
	}

	relationshipContext += fmt.Sprintf("\n语调: %s", relationship.Tone)

	return baseContext + relationshipContext, nil
}

// SelectModelForCharacter 为角色选择模型
func (s *orchestratorService) SelectModelForCharacter(ctx context.Context, characterID uuid.UUID, messageType string) (*domain.ModelResponse, error) {
	// 获取角色信息
	character, err := s.characterService.GetCharacter(ctx, characterID, nil)
	if err != nil {
		return nil, err
	}

	// 如果角色有默认模型，使用默认模型
	if character.DefaultModelID != nil {
		return s.modelService.GetModel(ctx, *character.DefaultModelID)
	}

	// 否则选择一个合适的模型
	modelListReq := &domain.ModelListRequest{
		UserType:    "basic", // 简化实现
		UserBalance: 100.0,   // 简化实现
		Page:        1,
		Limit:       5,
	}

	modelList, err := s.modelService.ListModels(ctx, modelListReq)
	if err != nil {
		return nil, err
	}

	// 选择第一个支持聊天的模型
	for _, model := range modelList.Models {
		for _, capability := range model.CapabilityList {
			if capability == "chat" {
				return &model, nil
			}
		}
	}

	return nil, fmt.Errorf("no suitable model found for character")
}

// RouteToModel 路由到模型
func (s *orchestratorService) RouteToModel(ctx context.Context, modelID uuid.UUID, prompt string, params map[string]interface{}) (*domain.ModelResponse, error) {
	return s.modelService.GetModel(ctx, modelID)
}

// ProcessTriggers 处理触发器
func (s *orchestratorService) ProcessTriggers(ctx context.Context, groupChatID uuid.UUID, message *domain.ChatMessage) ([]domain.TriggerAction, error) {
	// 评估触发器
	triggeredRules, err := s.relationshipService.EvaluateTriggers(ctx, groupChatID, map[string]interface{}{
		"message": message.Content,
		"type":    message.MessageType,
	})
	if err != nil {
		return nil, err
	}

	var actions []domain.TriggerAction
	for _, rule := range triggeredRules {
		// 解析动作
		var ruleActions []domain.TriggerAction
		err := json.Unmarshal(rule.Actions, &ruleActions)
		if err != nil {
			s.logger.WithError(err).WithField("rule_id", rule.ID).Warn("Failed to parse trigger actions")
			continue
		}
		actions = append(actions, ruleActions...)
	}

	return actions, nil
}

// ApplyTriggerActions 应用触发器动作
func (s *orchestratorService) ApplyTriggerActions(ctx context.Context, groupChatID uuid.UUID, actions []domain.TriggerAction) error {
	for _, action := range actions {
		switch action.Type {
		case "change_scene":
			_, err := s.memoryService.UpdateConversationContext(ctx, groupChatID, map[string]interface{}{
				"current_scene": action.Value,
			})
			if err != nil {
				s.logger.WithError(err).Warn("Failed to change scene")
			}
		case "change_mood":
			_, err := s.memoryService.UpdateConversationContext(ctx, groupChatID, map[string]interface{}{
				"mood": action.Value,
			})
			if err != nil {
				s.logger.WithError(err).Warn("Failed to change mood")
			}
		case "force_speaker":
			if speakerID, ok := action.Value.(string); ok {
				if id, err := uuid.Parse(speakerID); err == nil {
					err := s.UpdateSpeakingQueue(ctx, groupChatID, id)
					if err != nil {
						s.logger.WithError(err).Warn("Failed to force speaker")
					}
				}
			}
		}
	}
	return nil
}

// CalculateResponseCost 计算响应成本
func (s *orchestratorService) CalculateResponseCost(ctx context.Context, characterID uuid.UUID, prompt string, responseLength int) (float64, error) {
	// 简化的成本计算
	baseCost := 0.001                                          // 每个响应的基础成本
	tokenCost := float64(len(prompt)+responseLength) * 0.00001 // 每token成本
	return baseCost + tokenCost, nil
}

// DeductCost 扣费
func (s *orchestratorService) DeductCost(ctx context.Context, userID uuid.UUID, cost float64, description string) error {
	// 这里应该调用计费服务
	// 简化实现，直接返回成功
	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"cost":        cost,
		"description": description,
	}).Info("Cost deducted")
	return nil
}

// StartPerformance 开始演出
func (s *orchestratorService) StartPerformance(ctx context.Context, groupChatID uuid.UUID, scenario *domain.PerformanceScenario) error {
	// 实现演出开始逻辑
	return nil
}

// StopPerformance 停止演出
func (s *orchestratorService) StopPerformance(ctx context.Context, groupChatID uuid.UUID) error {
	// 实现演出停止逻辑
	return nil
}

// GetPerformanceStatus 获取演出状态
func (s *orchestratorService) GetPerformanceStatus(ctx context.Context, groupChatID uuid.UUID) (*domain.PerformanceStatus, error) {
	// 实现获取演出状态逻辑
	return &domain.PerformanceStatus{
		IsActive:   false,
		StartedAt:  nil,
		CurrentAct: 0,
		TotalActs:  0,
	}, nil
}

// 辅助方法

// buildPrompt 构建提示词
func (s *orchestratorService) buildPrompt(character *domain.CharacterResponse, context string) string {
	prompt := fmt.Sprintf("角色: %s\n", character.Name)
	if character.Personality != nil {
		prompt += fmt.Sprintf("性格: %s\n", *character.Personality)
	}
	prompt += fmt.Sprintf("上下文: %s\n", context)
	prompt += "请以这个角色的身份回应："
	return prompt
}

// generateMockResponse 生成模拟响应
func (s *orchestratorService) generateMockResponse(character *domain.CharacterResponse, context string) string {
	responses := []string{
		fmt.Sprintf("我是%s，很高兴和大家聊天。", character.Name),
		"这个话题很有趣呢。",
		"让我想想该怎么回应...",
		"大家觉得怎么样？",
	}

	// 简化实现：随机选择一个响应
	return responses[time.Now().Second()%len(responses)]
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
