package repository

import (
	"context"

	"github.com/google/uuid"

	"yunai/internal/domain"
)

// StoryRepository 故事仓储接口
type StoryRepository interface {
	// 章节管理
	CreateChapter(ctx context.Context, chapter *domain.StoryChapter) error
	GetChapterByID(ctx context.Context, id uuid.UUID) (*domain.StoryChapter, error)
	GetChaptersByGroupChatID(ctx context.Context, groupChatID uuid.UUID) ([]*domain.StoryChapter, error)
	GetCurrentChapterByGroupChatID(ctx context.Context, groupChatID uuid.UUID) (*domain.StoryChapter, error)
	UpdateChapter(ctx context.Context, chapter *domain.StoryChapter) error
	DeleteChapter(ctx context.Context, id uuid.UUID) error

	// 章节排序
	ReorderChapters(ctx context.Context, groupChatID uuid.UUID, chapterOrders []domain.ChapterOrder) error

	// 章节激活状态
	ActivateChapter(ctx context.Context, id uuid.UUID) error
	DeactivateChapter(ctx context.Context, id uuid.UUID) error

	// 触发条件管理
	UpdateChapterTriggerConditions(ctx context.Context, chapterID uuid.UUID, conditions string) error
	GetChapterTriggerConditions(ctx context.Context, chapterID uuid.UUID) (string, error)
}

// ChapterOrder 章节排序
type ChapterOrder struct {
	ChapterID uuid.UUID `json:"chapter_id"`
	Order     int       `json:"order"`
}
