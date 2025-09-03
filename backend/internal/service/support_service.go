package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"yunai/internal/domain"
	"yunai/internal/repository"
)

// SupportService 客户支持服务接口
type SupportService interface {
	// 工单管理
	CreateTicket(ctx context.Context, ticket *domain.SupportTicket) error
	UpdateTicket(ctx context.Context, ticket *domain.SupportTicket) error
	GetTicket(ctx context.Context, ticketID uuid.UUID) (*domain.SupportTicket, error)
	ListTickets(ctx context.Context, req *domain.TicketListRequest) (*domain.TicketListResponse, error)
	AssignTicket(ctx context.Context, ticketID, agentID uuid.UUID) error
	CloseTicket(ctx context.Context, ticketID uuid.UUID, resolution string) error

	// 工单消息
	AddTicketMessage(ctx context.Context, message *domain.TicketMessage) error
	GetTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]*domain.TicketMessage, error)

	// 智能分配
	AutoAssignTicket(ctx context.Context, ticketID uuid.UUID) error
	GetAvailableAgents(ctx context.Context) ([]*domain.SupportAgent, error)
	UpdateAgentStatus(ctx context.Context, agentID uuid.UUID, status string) error

	// 实时聊天
	StartChatSession(ctx context.Context, userID uuid.UUID) (*domain.ChatSession, error)
	EndChatSession(ctx context.Context, sessionID uuid.UUID) error
	SendChatMessage(ctx context.Context, message *domain.SupportChatMessage) error
	GetChatHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.SupportChatMessage, error)

	// 知识库
	CreateKnowledgeArticle(ctx context.Context, article *domain.KnowledgeArticle) error
	UpdateKnowledgeArticle(ctx context.Context, article *domain.KnowledgeArticle) error
	SearchKnowledge(ctx context.Context, query string) ([]*domain.KnowledgeArticle, error)
	GetPopularArticles(ctx context.Context, limit int) ([]*domain.KnowledgeArticle, error)

	// 满意度评价
	SubmitFeedback(ctx context.Context, feedback *domain.SupportFeedback) error
	GetFeedbackStats(ctx context.Context, agentID *uuid.UUID) (*domain.FeedbackStats, error)

	// 统计分析
	GetSupportStats(ctx context.Context) (*domain.SupportStats, error)
	GetAgentPerformance(ctx context.Context, agentID uuid.UUID) (*domain.AgentPerformance, error)
}

// supportService 客户支持服务实现
type supportService struct {
	supportRepo repository.SupportRepository
	redis       *redis.Client
	logger      *logrus.Logger

	// 智能分配配置
	assignmentConfig *AssignmentConfig
}

// AssignmentConfig 智能分配配置
type AssignmentConfig struct {
	MaxTicketsPerAgent  int     `json:"max_tickets_per_agent"`
	SkillMatchWeight    float64 `json:"skill_match_weight"`
	WorkloadWeight      float64 `json:"workload_weight"`
	ResponseTimeWeight  float64 `json:"response_time_weight"`
	CustomerSatWeight   float64 `json:"customer_sat_weight"`
	AutoAssignEnabled   bool    `json:"auto_assign_enabled"`
	EscalationThreshold int     `json:"escalation_threshold"` // 小时
}

// NewSupportService 创建客户支持服务
func NewSupportService(
	supportRepo repository.SupportRepository,
	redis *redis.Client,
	logger *logrus.Logger,
) SupportService {
	config := &AssignmentConfig{
		MaxTicketsPerAgent:  10,
		SkillMatchWeight:    0.4,
		WorkloadWeight:      0.3,
		ResponseTimeWeight:  0.2,
		CustomerSatWeight:   0.1,
		AutoAssignEnabled:   true,
		EscalationThreshold: 24,
	}

	return &supportService{
		supportRepo:      supportRepo,
		redis:            redis,
		logger:           logger,
		assignmentConfig: config,
	}
}

// CreateTicket 创建工单
func (s *supportService) CreateTicket(ctx context.Context, ticket *domain.SupportTicket) error {
	ticket.ID = uuid.New()
	ticket.Status = "open"
	ticket.Priority = s.calculateTicketPriority(ticket)
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	if err := s.supportRepo.CreateTicket(ctx, ticket); err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	// 自动分配工单
	if s.assignmentConfig.AutoAssignEnabled {
		go func() {
			if err := s.AutoAssignTicket(context.Background(), ticket.ID); err != nil {
				s.logger.WithError(err).WithField("ticket_id", ticket.ID).Error("Failed to auto-assign ticket")
			}
		}()
	}

	s.logger.WithFields(logrus.Fields{
		"ticket_id": ticket.ID,
		"user_id":   ticket.UserID,
		"category":  ticket.Category,
		"priority":  ticket.Priority,
	}).Info("Support ticket created")

	return nil
}

// UpdateTicket 更新工单
func (s *supportService) UpdateTicket(ctx context.Context, ticket *domain.SupportTicket) error {
	ticket.UpdatedAt = time.Now()

	if err := s.supportRepo.UpdateTicket(ctx, ticket); err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	s.logger.WithField("ticket_id", ticket.ID).Info("Support ticket updated")
	return nil
}

// GetTicket 获取工单
func (s *supportService) GetTicket(ctx context.Context, ticketID uuid.UUID) (*domain.SupportTicket, error) {
	ticket, err := s.supportRepo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	return ticket, nil
}

// ListTickets 列出工单
func (s *supportService) ListTickets(ctx context.Context, req *domain.TicketListRequest) (*domain.TicketListResponse, error) {
	tickets, total, err := s.supportRepo.ListTickets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	page := req.Offset/req.Limit + 1
	hasMore := req.Offset+req.Limit < total

	return &domain.TicketListResponse{
		Tickets: tickets,
		Total:   total,
		Page:    page,
		Limit:   req.Limit,
		HasMore: hasMore,
	}, nil
}

// AssignTicket 分配工单
func (s *supportService) AssignTicket(ctx context.Context, ticketID, agentID uuid.UUID) error {
	ticket, err := s.supportRepo.GetTicket(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}

	ticket.AssignedTo = &agentID
	ticket.Status = "assigned"
	ticket.UpdatedAt = time.Now()

	if err := s.supportRepo.UpdateTicket(ctx, ticket); err != nil {
		return fmt.Errorf("failed to assign ticket: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"ticket_id": ticketID,
		"agent_id":  agentID,
	}).Info("Ticket assigned")

	return nil
}

// CloseTicket 关闭工单
func (s *supportService) CloseTicket(ctx context.Context, ticketID uuid.UUID, resolution string) error {
	ticket, err := s.supportRepo.GetTicket(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}

	ticket.Status = "closed"
	ticket.Resolution = &resolution
	now := time.Now()
	ticket.ResolvedAt = &now
	ticket.UpdatedAt = now

	if err := s.supportRepo.UpdateTicket(ctx, ticket); err != nil {
		return fmt.Errorf("failed to close ticket: %w", err)
	}

	s.logger.WithField("ticket_id", ticketID).Info("Ticket closed")
	return nil
}

// AddTicketMessage 添加工单消息
func (s *supportService) AddTicketMessage(ctx context.Context, message *domain.TicketMessage) error {
	message.ID = uuid.New()
	message.CreatedAt = time.Now()

	if err := s.supportRepo.CreateTicketMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to add ticket message: %w", err)
	}

	// 更新工单的最后活动时间
	ticket, err := s.supportRepo.GetTicket(ctx, message.TicketID)
	if err == nil {
		ticket.UpdatedAt = time.Now()
		s.supportRepo.UpdateTicket(ctx, ticket)
	}

	return nil
}

// GetTicketMessages 获取工单消息
func (s *supportService) GetTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]*domain.TicketMessage, error) {
	messages, err := s.supportRepo.GetTicketMessages(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket messages: %w", err)
	}
	return messages, nil
}

// calculateTicketPriority 计算工单优先级
func (s *supportService) calculateTicketPriority(ticket *domain.SupportTicket) string {
	// 根据类别和关键词确定优先级
	urgentKeywords := []string{"紧急", "无法使用", "支付失败", "账户被盗"}

	for _, keyword := range urgentKeywords {
		if strings.Contains(ticket.Subject, keyword) || strings.Contains(ticket.Description, keyword) {
			return "high"
		}
	}

	switch ticket.Category {
	case "payment", "security":
		return "high"
	case "bug", "feature":
		return "medium"
	default:
		return "low"
	}
}

// AutoAssignTicket 自动分配工单
func (s *supportService) AutoAssignTicket(ctx context.Context, ticketID uuid.UUID) error {
	ticket, err := s.supportRepo.GetTicket(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}

	if ticket.AssignedTo != nil {
		return nil // 已经分配
	}

	// 获取可用客服
	agents, err := s.GetAvailableAgents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get available agents: %w", err)
	}

	if len(agents) == 0 {
		s.logger.WithField("ticket_id", ticketID).Warn("No available agents for auto-assignment")
		return nil
	}

	// 计算最佳分配
	bestAgent := s.findBestAgent(ticket, agents)
	if bestAgent == nil {
		return fmt.Errorf("no suitable agent found")
	}

	return s.AssignTicket(ctx, ticketID, bestAgent.ID)
}

// GetAvailableAgents 获取可用客服
func (s *supportService) GetAvailableAgents(ctx context.Context) ([]*domain.SupportAgent, error) {
	agents, err := s.supportRepo.GetActiveAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active agents: %w", err)
	}

	var availableAgents []*domain.SupportAgent
	for _, agent := range agents {
		if agent.Status == "online" || agent.Status == "available" {
			// 检查工作负载
			workload, err := s.getAgentWorkload(ctx, agent.ID)
			if err != nil {
				s.logger.WithError(err).WithField("agent_id", agent.ID).Warn("Failed to get agent workload")
				continue
			}

			if workload < s.assignmentConfig.MaxTicketsPerAgent {
				availableAgents = append(availableAgents, agent)
			}
		}
	}

	return availableAgents, nil
}

// UpdateAgentStatus 更新客服状态
func (s *supportService) UpdateAgentStatus(ctx context.Context, agentID uuid.UUID, status string) error {
	agent, err := s.supportRepo.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	agent.Status = status
	agent.UpdatedAt = time.Now()

	if err := s.supportRepo.UpdateAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"agent_id": agentID,
		"status":   status,
	}).Info("Agent status updated")

	return nil
}

// StartChatSession 开始聊天会话
func (s *supportService) StartChatSession(ctx context.Context, userID uuid.UUID) (*domain.ChatSession, error) {
	// 检查是否有活跃会话
	existingSession, err := s.supportRepo.GetActiveChatSession(ctx, userID)
	if err == nil && existingSession != nil {
		return existingSession, nil
	}

	session := &domain.ChatSession{
		ID:        uuid.New(),
		UserID:    userID,
		Status:    "waiting",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.supportRepo.CreateChatSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	// 尝试分配客服
	go func() {
		if err := s.assignChatAgent(context.Background(), session.ID); err != nil {
			s.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to assign chat agent")
		}
	}()

	s.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"user_id":    userID,
	}).Info("Chat session started")

	return session, nil
}

// EndChatSession 结束聊天会话
func (s *supportService) EndChatSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.supportRepo.GetChatSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get chat session: %w", err)
	}

	session.Status = "ended"
	now := time.Now()
	session.EndedAt = &now
	session.UpdatedAt = now

	if err := s.supportRepo.UpdateChatSession(ctx, session); err != nil {
		return fmt.Errorf("failed to end chat session: %w", err)
	}

	s.logger.WithField("session_id", sessionID).Info("Chat session ended")
	return nil
}

// SendChatMessage 发送聊天消息
func (s *supportService) SendChatMessage(ctx context.Context, message *domain.SupportChatMessage) error {
	message.ID = uuid.New()
	message.CreatedAt = time.Now()

	if err := s.supportRepo.CreateChatMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to send chat message: %w", err)
	}

	// 更新会话活动时间
	session, err := s.supportRepo.GetChatSession(ctx, message.SessionID)
	if err == nil {
		session.UpdatedAt = time.Now()
		s.supportRepo.UpdateChatSession(ctx, session)
	}

	return nil
}

// GetChatHistory 获取聊天历史
func (s *supportService) GetChatHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.SupportChatMessage, error) {
	messages, err := s.supportRepo.GetChatMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}
	return messages, nil
}

// CreateKnowledgeArticle 创建知识库文章
func (s *supportService) CreateKnowledgeArticle(ctx context.Context, article *domain.KnowledgeArticle) error {
	article.ID = uuid.New()
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()

	if err := s.supportRepo.CreateKnowledgeArticle(ctx, article); err != nil {
		return fmt.Errorf("failed to create knowledge article: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"article_id": article.ID,
		"title":      article.Title,
		"category":   article.Category,
	}).Info("Knowledge article created")

	return nil
}

// UpdateKnowledgeArticle 更新知识库文章
func (s *supportService) UpdateKnowledgeArticle(ctx context.Context, article *domain.KnowledgeArticle) error {
	article.UpdatedAt = time.Now()

	if err := s.supportRepo.UpdateKnowledgeArticle(ctx, article); err != nil {
		return fmt.Errorf("failed to update knowledge article: %w", err)
	}

	s.logger.WithField("article_id", article.ID).Info("Knowledge article updated")
	return nil
}

// SearchKnowledge 搜索知识库
func (s *supportService) SearchKnowledge(ctx context.Context, query string) ([]*domain.KnowledgeArticle, error) {
	articles, err := s.supportRepo.SearchKnowledgeArticles(ctx, query, 50) // 默认限制50条
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge articles: %w", err)
	}
	return articles, nil
}

// GetPopularArticles 获取热门文章
func (s *supportService) GetPopularArticles(ctx context.Context, limit int) ([]*domain.KnowledgeArticle, error) {
	articles, err := s.supportRepo.GetPopularKnowledgeArticles(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular articles: %w", err)
	}
	return articles, nil
}

// SubmitFeedback 提交反馈
func (s *supportService) SubmitFeedback(ctx context.Context, feedback *domain.SupportFeedback) error {
	feedback.ID = uuid.New()
	feedback.CreatedAt = time.Now()

	if err := s.supportRepo.CreateFeedback(ctx, feedback); err != nil {
		return fmt.Errorf("failed to submit feedback: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"feedback_id": feedback.ID,
		"rating":      feedback.Rating,
		"type":        feedback.Type,
	}).Info("Support feedback submitted")

	return nil
}

// GetFeedbackStats 获取反馈统计
func (s *supportService) GetFeedbackStats(ctx context.Context, agentID *uuid.UUID) (*domain.FeedbackStats, error) {
	stats, err := s.supportRepo.GetFeedbackStats(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback stats: %w", err)
	}
	return stats, nil
}

// GetSupportStats 获取支持统计
func (s *supportService) GetSupportStats(ctx context.Context) (*domain.SupportStats, error) {
	stats, err := s.supportRepo.GetSupportStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get support stats: %w", err)
	}
	return stats, nil
}

// GetAgentPerformance 获取客服表现
func (s *supportService) GetAgentPerformance(ctx context.Context, agentID uuid.UUID) (*domain.AgentPerformance, error) {
	performance, err := s.supportRepo.GetAgentPerformance(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent performance: %w", err)
	}
	return performance, nil
}

// 辅助方法

// findBestAgent 找到最佳客服
func (s *supportService) findBestAgent(ticket *domain.SupportTicket, agents []*domain.SupportAgent) *domain.SupportAgent {
	if len(agents) == 0 {
		return nil
	}

	type agentScore struct {
		agent *domain.SupportAgent
		score float64
	}

	var scores []agentScore

	for _, agent := range agents {
		score := s.calculateAgentScore(ticket, agent)
		scores = append(scores, agentScore{agent: agent, score: score})
	}

	// 找到最高分的客服
	bestScore := scores[0]
	for _, score := range scores[1:] {
		if score.score > bestScore.score {
			bestScore = score
		}
	}

	return bestScore.agent
}

// calculateAgentScore 计算客服分数
func (s *supportService) calculateAgentScore(ticket *domain.SupportTicket, agent *domain.SupportAgent) float64 {
	score := 0.0

	// 技能匹配度
	skillMatch := s.calculateSkillMatch(ticket.Category, agent.Skills)
	score += skillMatch * s.assignmentConfig.SkillMatchWeight

	// 工作负载（负载越低分数越高）
	workload, _ := s.getAgentWorkload(context.Background(), agent.ID)
	workloadScore := 1.0 - float64(workload)/float64(s.assignmentConfig.MaxTicketsPerAgent)
	score += workloadScore * s.assignmentConfig.WorkloadWeight

	// 响应时间（越快分数越高）
	responseTimeScore := s.calculateResponseTimeScore(agent.ID)
	score += responseTimeScore * s.assignmentConfig.ResponseTimeWeight

	// 客户满意度
	satisfactionScore := s.calculateSatisfactionScore(agent.ID)
	score += satisfactionScore * s.assignmentConfig.CustomerSatWeight

	return score
}

// calculateSkillMatch 计算技能匹配度
func (s *supportService) calculateSkillMatch(category string, skills []string) float64 {
	if len(skills) == 0 {
		return 0.5 // 默认分数
	}

	for _, skill := range skills {
		if strings.EqualFold(skill, category) {
			return 1.0 // 完全匹配
		}
	}

	// 部分匹配逻辑
	switch category {
	case "payment":
		for _, skill := range skills {
			if strings.Contains(strings.ToLower(skill), "payment") ||
				strings.Contains(strings.ToLower(skill), "billing") {
				return 0.8
			}
		}
	case "technical":
		for _, skill := range skills {
			if strings.Contains(strings.ToLower(skill), "tech") ||
				strings.Contains(strings.ToLower(skill), "bug") {
				return 0.8
			}
		}
	}

	return 0.3 // 基础分数
}

// getAgentWorkload 获取客服工作负载
func (s *supportService) getAgentWorkload(ctx context.Context, agentID uuid.UUID) (int, error) {
	count, err := s.supportRepo.GetAgentActiveTicketCount(ctx, agentID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// calculateResponseTimeScore 计算响应时间分数
func (s *supportService) calculateResponseTimeScore(agentID uuid.UUID) float64 {
	// 简化实现，实际应该基于历史数据计算
	return 0.8
}

// calculateSatisfactionScore 计算满意度分数
func (s *supportService) calculateSatisfactionScore(agentID uuid.UUID) float64 {
	// 简化实现，实际应该基于反馈数据计算
	return 0.85
}

// assignChatAgent 分配聊天客服
func (s *supportService) assignChatAgent(ctx context.Context, sessionID uuid.UUID) error {
	agents, err := s.GetAvailableAgents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get available agents: %w", err)
	}

	if len(agents) == 0 {
		s.logger.WithField("session_id", sessionID).Warn("No available agents for chat assignment")
		return nil
	}

	// 简单分配：选择第一个可用客服
	agent := agents[0]

	session, err := s.supportRepo.GetChatSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get chat session: %w", err)
	}

	session.AgentID = &agent.ID
	session.Status = "active"
	session.UpdatedAt = time.Now()

	if err := s.supportRepo.UpdateChatSession(ctx, session); err != nil {
		return fmt.Errorf("failed to assign chat agent: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"agent_id":   agent.ID,
	}).Info("Chat agent assigned")

	return nil
}
