package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"yunai/internal/domain"
)

// CharacterRepository 角色仓库接口
type CharacterRepository interface {
	// 角色管理
	CreateCharacter(ctx context.Context, character *domain.Character) error
	GetCharacterByID(ctx context.Context, id uuid.UUID) (*domain.Character, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Character, error)
	UpdateCharacter(ctx context.Context, character *domain.Character) error
	DeleteCharacter(ctx context.Context, id uuid.UUID) error
	ListCharacters(ctx context.Context, req *domain.CharacterListRequest) ([]*domain.Character, int, error)

	// 角色标签
	CreateCharacterTags(ctx context.Context, characterID uuid.UUID, tags []string) error
	GetCharacterTags(ctx context.Context, characterID uuid.UUID) ([]*domain.CharacterTag, error)
	DeleteCharacterTags(ctx context.Context, characterID uuid.UUID) error

	// 群聊管理
	CreateGroupChat(ctx context.Context, groupChat *domain.GroupChat) error
	GetGroupChatByID(ctx context.Context, id uuid.UUID) (*domain.GroupChat, error)
	UpdateGroupChat(ctx context.Context, groupChat *domain.GroupChat) error
	DeleteGroupChat(ctx context.Context, id uuid.UUID) error
	ListGroupChats(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.GroupChat, int, error)

	// 群聊成员
	AddGroupChatMember(ctx context.Context, member *domain.GroupChatMember) error
	RemoveGroupChatMember(ctx context.Context, groupChatID, memberID uuid.UUID, memberType string) error
	GetGroupChatMembers(ctx context.Context, groupChatID uuid.UUID) ([]*domain.GroupChatMember, error)

	// 聊天消息
	CreateMessage(ctx context.Context, message *domain.ChatMessage) error
	GetMessages(ctx context.Context, groupChatID uuid.UUID, page, limit int) ([]*domain.ChatMessage, error)
	UpdateMessage(ctx context.Context, message *domain.ChatMessage) error
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error

	// 章节管理（添加到群聊相关功能中）
	UpdateCurrentChapter(ctx context.Context, groupChatID, chapterID uuid.UUID) error
}

type characterRepository struct {
	db *sqlx.DB
}

// NewCharacterRepository 创建角色仓库
func NewCharacterRepository(db *sqlx.DB) CharacterRepository {
	return &characterRepository{db: db}
}

// CreateCharacter 创建角色
func (r *characterRepository) CreateCharacter(ctx context.Context, character *domain.Character) error {
	query := `
		INSERT INTO characters (
			id, created_by, name, description, personality,
			bg_image_url, cutout_image_url, bg_image_width, bg_image_height,
			cutout_image_width, cutout_image_height, default_model_id, model_params,
			system_prompt, visibility, is_featured, allow_chat, allow_group_chat,
			allow_calls, chat_count, like_count, view_count
		) VALUES (
			:id, :created_by, :name, :description, :personality,
			:bg_image_url, :cutout_image_url, :bg_image_width, :bg_image_height,
			:cutout_image_width, :cutout_image_height, :default_model_id, :model_params,
			:system_prompt, :visibility, :is_featured, :allow_chat, :allow_group_chat,
			:allow_calls, :chat_count, :like_count, :view_count
		)`

	_, err := r.db.NamedExecContext(ctx, query, character)
	return err
}

// GetCharacterByID 根据ID获取角色
func (r *characterRepository) GetCharacterByID(ctx context.Context, id uuid.UUID) (*domain.Character, error) {
	query := `
		SELECT id, created_by, name, description, personality,
			   bg_image_url, cutout_image_url, bg_image_width, bg_image_height,
			   cutout_image_width, cutout_image_height, default_model_id, model_params,
			   system_prompt, visibility, is_featured, allow_chat, allow_group_chat,
			   allow_calls, chat_count, like_count, view_count, created_at, updated_at
		FROM characters WHERE id = $1`

	var character domain.Character
	err := r.db.GetContext(ctx, &character, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "角色不存在", err)
		}
		return nil, err
	}

	return &character, nil
}

// GetByUserID 根据用户ID获取角色列表
func (r *characterRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Character, error) {
	query := `
		SELECT id, created_by, name, description, personality, bg_image_url, cutout_image_url,
			   bg_image_width, bg_image_height, cutout_image_width, cutout_image_height,
			   default_model_id, model_params, system_prompt, visibility, is_featured,
			   allow_chat, allow_group_chat, allow_calls, created_at, updated_at
		FROM characters
		WHERE created_by = $1
		ORDER BY created_at DESC`

	var characters []*domain.Character
	err := r.db.SelectContext(ctx, &characters, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get characters by user ID: %w", err)
	}

	return characters, nil
}

// UpdateCharacter 更新角色
func (r *characterRepository) UpdateCharacter(ctx context.Context, character *domain.Character) error {
	query := `
		UPDATE characters SET
			name = :name, description = :description, personality = :personality,
			bg_image_url = :bg_image_url, cutout_image_url = :cutout_image_url,
			bg_image_width = :bg_image_width, bg_image_height = :bg_image_height,
			cutout_image_width = :cutout_image_width, cutout_image_height = :cutout_image_height,
			default_model_id = :default_model_id, model_params = :model_params,
			system_prompt = :system_prompt, visibility = :visibility, is_featured = :is_featured,
			allow_chat = :allow_chat, allow_group_chat = :allow_group_chat, allow_calls = :allow_calls,
			updated_at = :updated_at
		WHERE id = :id`

	character.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, character)
	return err
}

// DeleteCharacter 删除角色
func (r *characterRepository) DeleteCharacter(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM characters WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListCharacters 获取角色列表
func (r *characterRepository) ListCharacters(ctx context.Context, req *domain.CharacterListRequest) ([]*domain.Character, int, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, *req.Visibility)
		argIndex++
	}

	if req.IsFeatured != nil {
		conditions = append(conditions, fmt.Sprintf("is_featured = $%d", argIndex))
		args = append(args, *req.IsFeatured)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		searchTerm := "%" + *req.Search + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	// 标签过滤
	if req.Tag != nil && *req.Tag != "" {
		conditions = append(conditions, fmt.Sprintf(`id IN (
			SELECT character_id FROM character_tags WHERE tag = $%d
		)`, argIndex))
		args = append(args, *req.Tag)
		argIndex++
	}

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM characters %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (req.Page - 1) * req.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, created_by, name, description, personality,
			   bg_image_url, cutout_image_url, bg_image_width, bg_image_height,
			   cutout_image_width, cutout_image_height, default_model_id, model_params,
			   system_prompt, visibility, is_featured, allow_chat, allow_group_chat,
			   allow_calls, chat_count, like_count, view_count, created_at, updated_at
		FROM characters %s
		ORDER BY is_featured DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	var characters []*domain.Character
	err = r.db.SelectContext(ctx, &characters, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return characters, total, nil
}

// CreateCharacterTags 创建角色标签
func (r *characterRepository) CreateCharacterTags(ctx context.Context, characterID uuid.UUID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	// 先删除现有标签
	err := r.DeleteCharacterTags(ctx, characterID)
	if err != nil {
		return err
	}

	// 批量插入新标签
	query := `INSERT INTO character_tags (id, character_id, tag) VALUES `
	var values []string
	var args []interface{}

	for i, tag := range tags {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, uuid.New(), characterID, tag)
	}

	query += strings.Join(values, ", ")
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// GetCharacterTags 获取角色标签
func (r *characterRepository) GetCharacterTags(ctx context.Context, characterID uuid.UUID) ([]*domain.CharacterTag, error) {
	query := `
		SELECT id, character_id, tag, created_at
		FROM character_tags
		WHERE character_id = $1
		ORDER BY created_at`

	var tags []*domain.CharacterTag
	err := r.db.SelectContext(ctx, &tags, query, characterID)
	return tags, err
}

// DeleteCharacterTags 删除角色标签
func (r *characterRepository) DeleteCharacterTags(ctx context.Context, characterID uuid.UUID) error {
	query := `DELETE FROM character_tags WHERE character_id = $1`
	_, err := r.db.ExecContext(ctx, query, characterID)
	return err
}

// CreateGroupChat 创建群聊
func (r *characterRepository) CreateGroupChat(ctx context.Context, groupChat *domain.GroupChat) error {
	query := `
		INSERT INTO group_chats (
			id, creator_user_id, name, description, background_image_url,
			background_music_url, background_sfx_url, max_members, is_public,
			allow_ai_invite, world_setting, current_scene, scene_style,
			member_count, message_count
		) VALUES (
			:id, :creator_user_id, :name, :description, :background_image_url,
			:background_music_url, :background_sfx_url, :max_members, :is_public,
			:allow_ai_invite, :world_setting, :current_scene, :scene_style,
			:member_count, :message_count
		)`

	_, err := r.db.NamedExecContext(ctx, query, groupChat)
	return err
}

// GetGroupChatByID 根据ID获取群聊
func (r *characterRepository) GetGroupChatByID(ctx context.Context, id uuid.UUID) (*domain.GroupChat, error) {
	query := `
		SELECT id, creator_user_id, name, description, background_image_url,
			   background_music_url, background_sfx_url, max_members, is_public,
			   allow_ai_invite, world_setting, current_scene, scene_style,
			   member_count, message_count, created_at, updated_at
		FROM group_chats WHERE id = $1`

	var groupChat domain.GroupChat
	err := r.db.GetContext(ctx, &groupChat, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(domain.CodeNotFound, "群聊不存在", err)
		}
		return nil, err
	}

	return &groupChat, nil
}

// UpdateGroupChat 更新群聊
func (r *characterRepository) UpdateGroupChat(ctx context.Context, groupChat *domain.GroupChat) error {
	query := `
		UPDATE group_chats SET
			name = :name, description = :description, background_image_url = :background_image_url,
			background_music_url = :background_music_url, background_sfx_url = :background_sfx_url,
			max_members = :max_members, is_public = :is_public, allow_ai_invite = :allow_ai_invite,
			world_setting = :world_setting, current_scene = :current_scene, scene_style = :scene_style,
			updated_at = :updated_at
		WHERE id = :id`

	groupChat.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, groupChat)
	return err
}

// DeleteGroupChat 删除群聊
func (r *characterRepository) DeleteGroupChat(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM group_chats WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListGroupChats 获取群聊列表
func (r *characterRepository) ListGroupChats(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.GroupChat, int, error) {
	// 获取用户参与的群聊
	countQuery := `
		SELECT COUNT(DISTINCT gc.id)
		FROM group_chats gc
		JOIN group_chat_members gcm ON gc.id = gcm.group_chat_id
		WHERE gcm.user_id = $1 OR gc.creator_user_id = $1`

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	dataQuery := `
		SELECT DISTINCT gc.id, gc.creator_user_id, gc.name, gc.description,
			   gc.background_image_url, gc.background_music_url, gc.background_sfx_url,
			   gc.max_members, gc.is_public, gc.allow_ai_invite, gc.world_setting,
			   gc.current_scene, gc.scene_style, gc.member_count, gc.message_count,
			   gc.created_at, gc.updated_at
		FROM group_chats gc
		JOIN group_chat_members gcm ON gc.id = gcm.group_chat_id
		WHERE gcm.user_id = $1 OR gc.creator_user_id = $1
		ORDER BY gc.updated_at DESC
		LIMIT $2 OFFSET $3`

	var groupChats []*domain.GroupChat
	err = r.db.SelectContext(ctx, &groupChats, dataQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return groupChats, total, nil
}

// AddGroupChatMember 添加群聊成员
func (r *characterRepository) AddGroupChatMember(ctx context.Context, member *domain.GroupChatMember) error {
	query := `
		INSERT INTO group_chat_members (
			id, group_chat_id, character_id, user_id, member_type,
			role, can_invite, can_kick, invited_by
		) VALUES (
			:id, :group_chat_id, :character_id, :user_id, :member_type,
			:role, :can_invite, :can_kick, :invited_by
		)`

	_, err := r.db.NamedExecContext(ctx, query, member)
	return err
}

// RemoveGroupChatMember 移除群聊成员
func (r *characterRepository) RemoveGroupChatMember(ctx context.Context, groupChatID, memberID uuid.UUID, memberType string) error {
	var query string
	if memberType == domain.MemberTypeUser {
		query = `DELETE FROM group_chat_members WHERE group_chat_id = $1 AND user_id = $2`
	} else {
		query = `DELETE FROM group_chat_members WHERE group_chat_id = $1 AND character_id = $2`
	}

	_, err := r.db.ExecContext(ctx, query, groupChatID, memberID)
	return err
}

// GetGroupChatMembers 获取群聊成员
func (r *characterRepository) GetGroupChatMembers(ctx context.Context, groupChatID uuid.UUID) ([]*domain.GroupChatMember, error) {
	query := `
		SELECT id, group_chat_id, character_id, user_id, member_type,
			   role, can_invite, can_kick, joined_at, invited_by
		FROM group_chat_members
		WHERE group_chat_id = $1
		ORDER BY joined_at`

	var members []*domain.GroupChatMember
	err := r.db.SelectContext(ctx, &members, query, groupChatID)
	return members, err
}

// CreateMessage 创建消息
func (r *characterRepository) CreateMessage(ctx context.Context, message *domain.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (
			id, group_chat_id, sender_character_id, sender_user_id, sender_type,
			content, message_type, media_urls, media_metadata, model_used,
			generation_cost, tokens_used, is_edited, is_deleted
		) VALUES (
			:id, :group_chat_id, :sender_character_id, :sender_user_id, :sender_type,
			:content, :message_type, :media_urls, :media_metadata, :model_used,
			:generation_cost, :tokens_used, :is_edited, :is_deleted
		)`

	_, err := r.db.NamedExecContext(ctx, query, message)
	return err
}

// GetMessages 获取消息列表
func (r *characterRepository) GetMessages(ctx context.Context, groupChatID uuid.UUID, page, limit int) ([]*domain.ChatMessage, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, group_chat_id, sender_character_id, sender_user_id, sender_type,
			   content, message_type, media_urls, media_metadata, model_used,
			   generation_cost, tokens_used, is_edited, is_deleted, created_at, updated_at
		FROM chat_messages
		WHERE group_chat_id = $1 AND is_deleted = FALSE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var messages []*domain.ChatMessage
	err := r.db.SelectContext(ctx, &messages, query, groupChatID, limit, offset)
	return messages, err
}

// UpdateMessage 更新消息
func (r *characterRepository) UpdateMessage(ctx context.Context, message *domain.ChatMessage) error {
	query := `
		UPDATE chat_messages SET
			content = :content, is_edited = :is_edited, updated_at = :updated_at
		WHERE id = :id`

	message.UpdatedAt = time.Now()
	message.IsEdited = true
	_, err := r.db.NamedExecContext(ctx, query, message)
	return err
}

// DeleteMessage 删除消息
func (r *characterRepository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	query := `UPDATE chat_messages SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, messageID)
	return err
}

// UpdateCurrentChapter 更新群聊当前章节
func (r *characterRepository) UpdateCurrentChapter(ctx context.Context, groupChatID, chapterID uuid.UUID) error {
	query := `UPDATE group_chats SET current_chapter_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, chapterID, groupChatID)
	return err
}
