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
	"yunai/internal/repository"
)

// MemoryService 记忆服务接口
type MemoryService interface {
	// 记忆片段管理
	CreateMemoryFragment(ctx context.Context, req *domain.CreateMemoryFragmentRequest) (*domain.MemoryFragment, error)
	GetMemoryFragment(ctx context.Context, id uuid.UUID) (*domain.MemoryFragment, error)
	UpdateMemoryFragment(ctx context.Context, id uuid.UUID, content, summary *string, importance *float64) (*domain.MemoryFragment, error)
	DeleteMemoryFragment(ctx context.Context, id uuid.UUID) error
	ListMemoryFragments(ctx context.Context, characterID *uuid.UUID, memoryTypes []string, page, limit int) ([]*domain.MemoryFragment, int, error)

	// 记忆搜索和检索
	SearchMemories(ctx context.Context, req *domain.MemorySearchRequest) (*domain.MemorySearchResponse, error)
	GetRecentMemories(ctx context.Context, characterID uuid.UUID, hours int, limit int) ([]*domain.MemoryFragment, error)
	GetImportantMemories(ctx context.Context, characterID uuid.UUID, minImportance float64, limit int) ([]*domain.MemoryFragment, error)
	GetRelatedMemories(ctx context.Context, characterID uuid.UUID, keywords []string, limit int) ([]*domain.MemoryFragment, error)

	// 自动记忆生成
	CreateMemoryFromMessage(ctx context.Context, message *domain.ChatMessage, importance float64) (*domain.MemoryFragment, error)
	CreateMemoryFromEvent(ctx context.Context, characterID uuid.UUID, eventDescription string, importance float64) (*domain.MemoryFragment, error)

	// 记忆摘要和整理
	SummarizeMemories(ctx context.Context, characterID uuid.UUID, memoryIDs []uuid.UUID) (string, error)
	ConsolidateMemories(ctx context.Context, characterID uuid.UUID, timeRange time.Duration) error
	CleanupExpiredMemories(ctx context.Context) (int, error)

	// 对话上下文管理
	CreateConversationContext(ctx context.Context, groupChatID uuid.UUID) (*domain.ConversationContext, error)
	GetConversationContext(ctx context.Context, groupChatID uuid.UUID) (*domain.ConversationContext, error)
	UpdateConversationContext(ctx context.Context, groupChatID uuid.UUID, updates map[string]interface{}) (*domain.ConversationContext, error)

	// 上下文构建
	BuildContextForCharacter(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, maxTokens int) (string, error)
	GetRelevantContext(ctx context.Context, characterID uuid.UUID, currentMessage string, maxMemories int) ([]*domain.MemoryFragment, error)

	// 记忆统计
	GetMemoryStats(ctx context.Context, characterID uuid.UUID) (map[string]interface{}, error)
}

type memoryService struct {
	memoryRepo repository.MemoryRepository
	logger     *logrus.Logger
}

// NewMemoryService 创建记忆服务
func NewMemoryService(memoryRepo repository.MemoryRepository, logger *logrus.Logger) MemoryService {
	return &memoryService{
		memoryRepo: memoryRepo,
		logger:     logger,
	}
}

// CreateMemoryFragment 创建记忆片段
func (s *memoryService) CreateMemoryFragment(ctx context.Context, req *domain.CreateMemoryFragmentRequest) (*domain.MemoryFragment, error) {
	// 生成摘要
	summary := s.generateSummary(req.Content)

	// 生成向量嵌入（模拟实现）
	embedding := s.generateEmbedding(req.Content)

	memory := &domain.MemoryFragment{
		ID:                 uuid.New(),
		CharacterID:        req.CharacterID,
		Content:            req.Content,
		Summary:            &summary,
		MemoryType:         req.MemoryType,
		ImportanceScore:    req.ImportanceScore,
		EmotionalIntensity: req.EmotionalIntensity,
		RelatedCharacters:  req.RelatedCharacters,
		RelatedTopics:      req.RelatedTopics,
		RelatedEmotions:    req.RelatedEmotions,
		SourceMessageID:    req.SourceMessageID,
		SourceGroupChatID:  req.SourceGroupChatID,
		ContextMetadata:    json.RawMessage("{}"),
		Embedding:          embedding,
		AccessCount:        0,
		ExpiresAt:          req.ExpiresAt,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := s.memoryRepo.CreateMemoryFragment(ctx, memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory fragment: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"memory_id":    memory.ID,
		"character_id": memory.CharacterID,
		"memory_type":  memory.MemoryType,
		"importance":   memory.ImportanceScore,
	}).Info("Memory fragment created")

	return memory, nil
}

// GetMemoryFragment 获取记忆片段
func (s *memoryService) GetMemoryFragment(ctx context.Context, id uuid.UUID) (*domain.MemoryFragment, error) {
	memory, err := s.memoryRepo.GetMemoryFragmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 增加访问次数
	err = s.memoryRepo.IncrementMemoryAccess(ctx, id)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to increment memory access count")
	}

	return memory, nil
}

// UpdateMemoryFragment 更新记忆片段
func (s *memoryService) UpdateMemoryFragment(ctx context.Context, id uuid.UUID, content, summary *string, importance *float64) (*domain.MemoryFragment, error) {
	memory, err := s.memoryRepo.GetMemoryFragmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if content != nil {
		memory.Content = *content
		// 重新生成摘要和嵌入
		newSummary := s.generateSummary(*content)
		memory.Summary = &newSummary
		memory.Embedding = s.generateEmbedding(*content)
	}

	if summary != nil {
		memory.Summary = summary
	}

	if importance != nil {
		memory.ImportanceScore = *importance
	}

	err = s.memoryRepo.UpdateMemoryFragment(ctx, memory)
	if err != nil {
		return nil, fmt.Errorf("failed to update memory fragment: %w", err)
	}

	return memory, nil
}

// DeleteMemoryFragment 删除记忆片段
func (s *memoryService) DeleteMemoryFragment(ctx context.Context, id uuid.UUID) error {
	return s.memoryRepo.DeleteMemoryFragment(ctx, id)
}

// ListMemoryFragments 获取记忆片段列表
func (s *memoryService) ListMemoryFragments(ctx context.Context, characterID *uuid.UUID, memoryTypes []string, page, limit int) ([]*domain.MemoryFragment, int, error) {
	return s.memoryRepo.ListMemoryFragments(ctx, characterID, memoryTypes, page, limit)
}

// SearchMemories 搜索记忆
func (s *memoryService) SearchMemories(ctx context.Context, req *domain.MemorySearchRequest) (*domain.MemorySearchResponse, error) {
	// 首先尝试内容搜索
	memories, err := s.memoryRepo.SearchMemoriesByContent(ctx, req.CharacterID, req.Query, req.MemoryTypes, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	// 如果需要更精确的搜索，可以使用向量搜索
	if len(memories) < req.Limit {
		// 暂时跳过向量搜索，因为需要更复杂的实现
		// embedding := s.generateEmbedding(req.Query)
		// vectorMemories, err := s.memoryRepo.SearchMemoriesByEmbedding(ctx, req.CharacterID, embedding, req.MemoryTypes, req.Limit-len(memories), req.MinScore)
	}

	// 转换为正确的类型
	var memoryList []domain.MemoryFragment
	for _, m := range memories {
		memoryList = append(memoryList, *m)
	}

	return &domain.MemorySearchResponse{
		Memories: memoryList,
		Query:    req.Query,
		Total:    len(memories),
	}, nil
}

// GetRecentMemories 获取最近记忆
func (s *memoryService) GetRecentMemories(ctx context.Context, characterID uuid.UUID, hours int, limit int) ([]*domain.MemoryFragment, error) {
	return s.memoryRepo.GetRecentMemories(ctx, characterID, hours, limit)
}

// GetImportantMemories 获取重要记忆
func (s *memoryService) GetImportantMemories(ctx context.Context, characterID uuid.UUID, minImportance float64, limit int) ([]*domain.MemoryFragment, error) {
	return s.memoryRepo.GetImportantMemories(ctx, characterID, minImportance, limit)
}

// GetRelatedMemories 获取相关记忆
func (s *memoryService) GetRelatedMemories(ctx context.Context, characterID uuid.UUID, keywords []string, limit int) ([]*domain.MemoryFragment, error) {
	// 使用关键词搜索
	query := strings.Join(keywords, " ")
	searchReq := &domain.MemorySearchRequest{
		CharacterID: &characterID,
		Query:       query,
		Limit:       limit,
		MinScore:    0.3,
	}

	result, err := s.SearchMemories(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	// 转换为指针类型
	var memoryPtrs []*domain.MemoryFragment
	for i := range result.Memories {
		memoryPtrs = append(memoryPtrs, &result.Memories[i])
	}
	return memoryPtrs, nil
}

// CreateMemoryFromMessage 从消息创建记忆
func (s *memoryService) CreateMemoryFromMessage(ctx context.Context, message *domain.ChatMessage, importance float64) (*domain.MemoryFragment, error) {
	// 分析消息内容，提取关键信息
	relatedTopics := s.extractTopics(message.Content)
	relatedEmotions := s.extractEmotions(message.Content)

	var relatedCharacters []string
	if message.SenderCharacterID != nil {
		relatedCharacters = append(relatedCharacters, message.SenderCharacterID.String())
	}

	req := &domain.CreateMemoryFragmentRequest{
		CharacterID:        *message.SenderCharacterID, // 假设是角色发送的消息
		Content:            message.Content,
		MemoryType:         domain.MemoryTypeConversation,
		ImportanceScore:    importance,
		EmotionalIntensity: s.calculateEmotionalIntensity(message.Content),
		RelatedCharacters:  relatedCharacters,
		RelatedTopics:      relatedTopics,
		RelatedEmotions:    relatedEmotions,
		SourceMessageID:    &message.ID,
		SourceGroupChatID:  message.GroupChatID,
	}

	return s.CreateMemoryFragment(ctx, req)
}

// CreateMemoryFromEvent 从事件创建记忆
func (s *memoryService) CreateMemoryFromEvent(ctx context.Context, characterID uuid.UUID, eventDescription string, importance float64) (*domain.MemoryFragment, error) {
	req := &domain.CreateMemoryFragmentRequest{
		CharacterID:        characterID,
		Content:            eventDescription,
		MemoryType:         domain.MemoryTypeEvent,
		ImportanceScore:    importance,
		EmotionalIntensity: s.calculateEmotionalIntensity(eventDescription),
		RelatedTopics:      s.extractTopics(eventDescription),
		RelatedEmotions:    s.extractEmotions(eventDescription),
	}

	return s.CreateMemoryFragment(ctx, req)
}

// SummarizeMemories 摘要记忆
func (s *memoryService) SummarizeMemories(ctx context.Context, characterID uuid.UUID, memoryIDs []uuid.UUID) (string, error) {
	var contents []string

	for _, memoryID := range memoryIDs {
		memory, err := s.memoryRepo.GetMemoryFragmentByID(ctx, memoryID)
		if err != nil {
			continue
		}
		if memory.CharacterID == characterID {
			contents = append(contents, memory.Content)
		}
	}

	if len(contents) == 0 {
		return "", fmt.Errorf("no valid memories found")
	}

	// 简化的摘要生成
	summary := s.generateSummaryFromMultiple(contents)
	return summary, nil
}

// ConsolidateMemories 整理记忆
func (s *memoryService) ConsolidateMemories(ctx context.Context, characterID uuid.UUID, timeRange time.Duration) error {
	// 获取指定时间范围内的记忆
	hours := int(timeRange.Hours())
	memories, err := s.memoryRepo.GetRecentMemories(ctx, characterID, hours, 100)
	if err != nil {
		return err
	}

	// 按类型分组
	memoryGroups := make(map[string][]*domain.MemoryFragment)
	for _, memory := range memories {
		memoryGroups[memory.MemoryType] = append(memoryGroups[memory.MemoryType], memory)
	}

	// 为每个类型生成摘要
	for memoryType, groupMemories := range memoryGroups {
		if len(groupMemories) > 5 { // 只有超过5个记忆才进行整理
			var memoryIDs []uuid.UUID
			for _, m := range groupMemories {
				memoryIDs = append(memoryIDs, m.ID)
			}

			summary, err := s.SummarizeMemories(ctx, characterID, memoryIDs)
			if err != nil {
				continue
			}

			// 创建整理后的记忆
			consolidatedReq := &domain.CreateMemoryFragmentRequest{
				CharacterID:     characterID,
				Content:         summary,
				MemoryType:      memoryType,
				ImportanceScore: 0.8, // 整理后的记忆通常比较重要
			}

			_, err = s.CreateMemoryFragment(ctx, consolidatedReq)
			if err != nil {
				s.logger.WithError(err).Warn("Failed to create consolidated memory")
			}
		}
	}

	return nil
}

// CleanupExpiredMemories 清理过期记忆
func (s *memoryService) CleanupExpiredMemories(ctx context.Context) (int, error) {
	return s.memoryRepo.CleanupExpiredMemories(ctx)
}

// CreateConversationContext 创建对话上下文
func (s *memoryService) CreateConversationContext(ctx context.Context, groupChatID uuid.UUID) (*domain.ConversationContext, error) {
	context := &domain.ConversationContext{
		ID:                uuid.New(),
		GroupChatID:       groupChatID,
		TurnCount:         0,
		LastActivityAt:    time.Now(),
		OrchestrationMode: domain.OrchestrationModeAuto,
		InterventionLevel: 0.3,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := s.memoryRepo.CreateConversationContext(ctx, context)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation context: %w", err)
	}

	return context, nil
}

// GetConversationContext 获取对话上下文
func (s *memoryService) GetConversationContext(ctx context.Context, groupChatID uuid.UUID) (*domain.ConversationContext, error) {
	context, err := s.memoryRepo.GetConversationContextByGroupID(ctx, groupChatID)
	if err != nil {
		// 如果不存在，自动创建
		if err.Error() == "对话上下文不存在" {
			return s.CreateConversationContext(ctx, groupChatID)
		}
		return nil, err
	}

	return context, nil
}

// UpdateConversationContext 更新对话上下文
func (s *memoryService) UpdateConversationContext(ctx context.Context, groupChatID uuid.UUID, updates map[string]interface{}) (*domain.ConversationContext, error) {
	context, err := s.GetConversationContext(ctx, groupChatID)
	if err != nil {
		return nil, err
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "current_scene":
			if v, ok := value.(string); ok {
				context.CurrentScene = &v
			}
		case "mood":
			if v, ok := value.(string); ok {
				context.Mood = &v
			}
		case "current_speaker_id":
			if v, ok := value.(uuid.UUID); ok {
				context.CurrentSpeakerID = &v
			}
		case "next_speaker_id":
			if v, ok := value.(uuid.UUID); ok {
				context.NextSpeakerID = &v
			}
		case "turn_count":
			if v, ok := value.(int); ok {
				context.TurnCount = v
			}
		}
	}

	context.LastActivityAt = time.Now()
	err = s.memoryRepo.UpdateConversationContext(ctx, context)
	if err != nil {
		return nil, fmt.Errorf("failed to update conversation context: %w", err)
	}

	return context, nil
}

// BuildContextForCharacter 为角色构建上下文
func (s *memoryService) BuildContextForCharacter(ctx context.Context, characterID uuid.UUID, groupChatID uuid.UUID, maxTokens int) (string, error) {
	var contextParts []string

	// 获取重要记忆
	importantMemories, err := s.GetImportantMemories(ctx, characterID, 0.7, 5)
	if err == nil && len(importantMemories) > 0 {
		contextParts = append(contextParts, "重要记忆:")
		for _, memory := range importantMemories {
			if memory.Summary != nil {
				contextParts = append(contextParts, "- "+*memory.Summary)
			} else {
				contextParts = append(contextParts, "- "+s.truncateText(memory.Content, 100))
			}
		}
	}

	// 获取最近记忆
	recentMemories, err := s.GetRecentMemories(ctx, characterID, 24, 3)
	if err == nil && len(recentMemories) > 0 {
		contextParts = append(contextParts, "\n最近记忆:")
		for _, memory := range recentMemories {
			if memory.Summary != nil {
				contextParts = append(contextParts, "- "+*memory.Summary)
			} else {
				contextParts = append(contextParts, "- "+s.truncateText(memory.Content, 100))
			}
		}
	}

	context := strings.Join(contextParts, "\n")

	// 根据maxTokens截断（简化实现，实际应该按token计算）
	if len(context) > maxTokens*4 { // 假设平均4字符=1token
		context = context[:maxTokens*4] + "..."
	}

	return context, nil
}

// GetRelevantContext 获取相关上下文
func (s *memoryService) GetRelevantContext(ctx context.Context, characterID uuid.UUID, currentMessage string, maxMemories int) ([]*domain.MemoryFragment, error) {
	// 提取当前消息的关键词
	keywords := s.extractTopics(currentMessage)

	// 搜索相关记忆
	return s.GetRelatedMemories(ctx, characterID, keywords, maxMemories)
}

// GetMemoryStats 获取记忆统计
func (s *memoryService) GetMemoryStats(ctx context.Context, characterID uuid.UUID) (map[string]interface{}, error) {
	return s.memoryRepo.GetMemoryStats(ctx, characterID)
}

// 辅助方法

// generateSummary 生成摘要（简化实现）
func (s *memoryService) generateSummary(content string) string {
	if len(content) <= 100 {
		return content
	}

	// 简化的摘要生成：取前100个字符
	return content[:100] + "..."
}

// generateEmbedding 生成向量嵌入（模拟实现）
func (s *memoryService) generateEmbedding(content string) json.RawMessage {
	// 模拟生成1536维的向量嵌入
	embedding := make([]float32, 1536)

	// 简化实现：基于内容长度和字符生成伪随机向量
	hash := 0
	for _, char := range content {
		hash = hash*31 + int(char)
	}

	for i := range embedding {
		hash = hash*1103515245 + 12345
		embedding[i] = float32(hash%1000) / 1000.0
	}

	// 转换为JSON
	embeddingJSON, _ := json.Marshal(embedding)
	return json.RawMessage(embeddingJSON)
}

// extractTopics 提取主题（简化实现）
func (s *memoryService) extractTopics(content string) []string {
	// 简化的主题提取：基于关键词
	keywords := []string{"爱情", "友情", "工作", "学习", "家庭", "梦想", "困难", "快乐", "悲伤", "愤怒"}
	var topics []string

	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, keyword) {
			topics = append(topics, keyword)
		}
	}

	return topics
}

// extractEmotions 提取情感（简化实现）
func (s *memoryService) extractEmotions(content string) []string {
	emotions := []string{"开心", "悲伤", "愤怒", "恐惧", "惊讶", "厌恶", "期待"}
	var detectedEmotions []string

	contentLower := strings.ToLower(content)
	for _, emotion := range emotions {
		if strings.Contains(contentLower, emotion) {
			detectedEmotions = append(detectedEmotions, emotion)
		}
	}

	return detectedEmotions
}

// calculateEmotionalIntensity 计算情感强度（简化实现）
func (s *memoryService) calculateEmotionalIntensity(content string) float64 {
	// 基于感叹号、问号等标点符号计算情感强度
	intensity := 0.0

	exclamationCount := strings.Count(content, "!")
	questionCount := strings.Count(content, "?")

	intensity += float64(exclamationCount) * 0.2
	intensity += float64(questionCount) * 0.1

	if intensity > 1.0 {
		intensity = 1.0
	}

	return intensity
}

// generateSummaryFromMultiple 从多个内容生成摘要
func (s *memoryService) generateSummaryFromMultiple(contents []string) string {
	if len(contents) == 0 {
		return ""
	}

	if len(contents) == 1 {
		return s.generateSummary(contents[0])
	}

	// 简化实现：合并内容并生成摘要
	combined := strings.Join(contents, " ")
	return s.generateSummary(combined)
}

// truncateText 截断文本
func (s *memoryService) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
