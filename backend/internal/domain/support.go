package domain

import (
	"time"

	"github.com/google/uuid"
)

// SupportTicket 工单
type SupportTicket struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	AssignedTo  *uuid.UUID `json:"assigned_to" gorm:"type:uuid"`
	Subject     string     `json:"subject" gorm:"not null"`
	Description string     `json:"description" gorm:"not null"`
	Category    string     `json:"category" gorm:"not null"`      // payment, technical, general, bug, feature
	Priority    string     `json:"priority" gorm:"default:'low'"` // low, medium, high, urgent
	Status      string     `json:"status" gorm:"default:'open'"`  // open, assigned, in_progress, resolved, closed
	Resolution  *string    `json:"resolution"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TicketMessage 工单消息
type TicketMessage struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TicketID   uuid.UUID `json:"ticket_id" gorm:"type:uuid;not null"`
	SenderID   uuid.UUID `json:"sender_id" gorm:"type:uuid;not null"`
	SenderType string    `json:"sender_type" gorm:"not null"` // user, agent, system
	Content    string    `json:"content" gorm:"not null"`
	IsInternal bool      `json:"is_internal" gorm:"default:false"` // 内部备注
	CreatedAt  time.Time `json:"created_at"`
}

// SupportAgent 客服
type SupportAgent struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID               uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Name                 string    `json:"name" gorm:"not null"`
	Email                string    `json:"email" gorm:"not null"`
	Status               string    `json:"status" gorm:"default:'offline'"` // online, offline, busy, away
	Skills               []string  `json:"skills" gorm:"type:jsonb"`
	MaxConcurrentTickets int       `json:"max_concurrent_tickets" gorm:"default:10"`
	IsActive             bool      `json:"is_active" gorm:"default:true"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ChatSession 聊天会话
type ChatSession struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	AgentID   *uuid.UUID `json:"agent_id" gorm:"type:uuid"`
	Status    string     `json:"status" gorm:"default:'waiting'"` // waiting, active, ended
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// SupportChatMessage 客服聊天消息
type SupportChatMessage struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID  uuid.UUID `json:"session_id" gorm:"type:uuid;not null"`
	SenderID   uuid.UUID `json:"sender_id" gorm:"type:uuid;not null"`
	SenderType string    `json:"sender_type" gorm:"not null"` // user, agent
	Content    string    `json:"content" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
}

// KnowledgeArticle 知识库文章
type KnowledgeArticle struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title        string    `json:"title" gorm:"not null"`
	Content      string    `json:"content" gorm:"not null"`
	Category     string    `json:"category" gorm:"not null"`
	Tags         []string  `json:"tags" gorm:"type:jsonb"`
	ViewCount    int       `json:"view_count" gorm:"default:0"`
	HelpfulCount int       `json:"helpful_count" gorm:"default:0"`
	IsPublished  bool      `json:"is_published" gorm:"default:false"`
	AuthorID     uuid.UUID `json:"author_id" gorm:"type:uuid;not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SupportFeedback 支持反馈
type SupportFeedback struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TicketID  *uuid.UUID `json:"ticket_id" gorm:"type:uuid"`
	SessionID *uuid.UUID `json:"session_id" gorm:"type:uuid"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	AgentID   *uuid.UUID `json:"agent_id" gorm:"type:uuid"`
	Type      string     `json:"type" gorm:"not null"`   // ticket, chat, general
	Rating    int        `json:"rating" gorm:"not null"` // 1-5
	Comment   string     `json:"comment"`
	CreatedAt time.Time  `json:"created_at"`
}

// TicketListRequest 工单列表请求
type TicketListRequest struct {
	Status     string     `json:"status"`
	Category   string     `json:"category"`
	Priority   string     `json:"priority"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
	UserID     *uuid.UUID `json:"user_id"`
	Offset     int        `json:"offset"`
	Limit      int        `json:"limit"`
}

// TicketListResponse 工单列表响应
type TicketListResponse struct {
	Tickets []*SupportTicket `json:"tickets"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
	HasMore bool             `json:"has_more"`
}

// FeedbackStats 反馈统计
type FeedbackStats struct {
	AgentID            *uuid.UUID  `json:"agent_id"`
	TotalFeedbacks     int         `json:"total_feedbacks"`
	AverageRating      float64     `json:"average_rating"`
	RatingDistribution map[int]int `json:"rating_distribution"`
	PositiveFeedbacks  int         `json:"positive_feedbacks"` // 4-5星
	NegativeFeedbacks  int         `json:"negative_feedbacks"` // 1-2星
}

// SupportStats 支持统计
type SupportStats struct {
	TotalTickets         int     `json:"total_tickets"`
	OpenTickets          int     `json:"open_tickets"`
	ResolvedTickets      int     `json:"resolved_tickets"`
	AvgResolutionTime    float64 `json:"avg_resolution_time"`   // 小时
	AvgResponseTime      float64 `json:"avg_response_time"`     // 小时
	CustomerSatisfaction float64 `json:"customer_satisfaction"` // 平均评分
	ActiveAgents         int     `json:"active_agents"`
	ActiveChatSessions   int     `json:"active_chat_sessions"`
}

// AgentPerformance 客服表现
type AgentPerformance struct {
	AgentID           uuid.UUID `json:"agent_id"`
	TicketsHandled    int       `json:"tickets_handled"`
	TicketsResolved   int       `json:"tickets_resolved"`
	AvgResolutionTime float64   `json:"avg_resolution_time"`
	AvgResponseTime   float64   `json:"avg_response_time"`
	CustomerRating    float64   `json:"customer_rating"`
	ChatSessions      int       `json:"chat_sessions"`
	OnlineHours       float64   `json:"online_hours"`
}
