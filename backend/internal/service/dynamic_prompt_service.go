package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// DynamicPromptService 动态提示词服务
// 完全基于数据库数据动态构建提示词，不写死任何内容
type DynamicPromptService struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

// NewDynamicPromptService 创建动态提示词服务
func NewDynamicPromptService(db *sqlx.DB, logger *logrus.Logger) *DynamicPromptService {
	return &DynamicPromptService{
		db:     db,
		logger: logger,
	}
}

// PromptContext 提示词上下文
type PromptContext struct {
	CharacterID    uuid.UUID
	UserID         uuid.UUID
	ScenarioType   string // moments, chat, group_chat, voice_call, invitation
	TargetUserID   *uuid.UUID
	TargetCharID   *uuid.UUID
	GroupChatID    *uuid.UUID
	AdditionalData map[string]interface{}
}

// CharacterData 角色数据
type CharacterData struct {
	ID           uuid.UUID `db:"id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	Personality  string    `db:"personality"`
	SystemPrompt *string   `db:"system_prompt"`
}

// UserData 用户数据
type UserData struct {
	ID       uuid.UUID `db:"id"`
	Username string    `db:"username"`
}

// RelationshipData 关系数据
type RelationshipData struct {
	TargetCharacterID   uuid.UUID `db:"target_character_id"`
	TargetCharacterName string    `db:"target_character_name"`
	RelationshipType    string    `db:"relationship_type"`
	Strength            float64   `db:"strength"`
	Trust               float64   `db:"trust"`
	Affection           float64   `db:"affection"`
	Respect             float64   `db:"respect"`
	Intimacy            float64   `db:"intimacy"`
	Tone                *string   `db:"tone"`
	FormalityLevel      *string   `db:"formality_level"`
	CustomTypeName      *string   `db:"custom_type_name"`
	Description         *string   `db:"description"`
}

// GroupChatData 群聊数据
type GroupChatData struct {
	ID           uuid.UUID `db:"id"`
	Name         string    `db:"name"`
	Description  *string   `db:"description"`
	WorldSetting *string   `db:"world_setting"`
}

// BuildDynamicPrompt 构建动态提示词
func (s *DynamicPromptService) BuildDynamicPrompt(ctx context.Context, promptCtx *PromptContext) (string, error) {
	// 1. 获取角色数据
	character, err := s.getCharacterData(ctx, promptCtx.CharacterID)
	if err != nil {
		return "", fmt.Errorf("获取角色数据失败: %w", err)
	}

	// 2. 获取用户数据
	user, err := s.getUserData(ctx, promptCtx.UserID)
	if err != nil {
		return "", fmt.Errorf("获取用户数据失败: %w", err)
	}

	// 3. 获取关系网络数据
	relationships, err := s.getRelationshipNetwork(ctx, promptCtx.CharacterID)
	if err != nil {
		return "", fmt.Errorf("获取关系网络失败: %w", err)
	}

	// 4. 根据场景类型构建提示词
	switch promptCtx.ScenarioType {
	case "moments":
		return s.buildMomentsPrompt(ctx, character, user, relationships, promptCtx)
	case "chat":
		return s.buildChatPrompt(ctx, character, user, relationships, promptCtx)
	case "group_chat":
		return s.buildGroupChatPrompt(ctx, character, user, relationships, promptCtx)
	case "voice_call":
		return s.buildVoiceCallPrompt(ctx, character, user, relationships, promptCtx)
	case "invitation":
		return s.buildInvitationPrompt(ctx, character, user, relationships, promptCtx)
	case "interaction":
		return s.buildInteractionPrompt(ctx, character, user, relationships, promptCtx)
	default:
		return s.buildGeneralPrompt(ctx, character, user, relationships, promptCtx)
	}
}

// GeneratePrompt 通用生成接口（用于系统级提示生成场景）
// 说明：有些调用方只需要按键值生成一段系统提示，这里提供一个通用实现，保证接口存在且可用。
func (s *DynamicPromptService) GeneratePrompt(ctx context.Context, templateKey string, params map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString("系统提示：")
	if templateKey != "" {
		b.WriteString(templateKey)
	}
	b.WriteString("\n")
	// 为了稳定输出，按键名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := params[k]
		b.WriteString("- ")
		b.WriteString(k)
		b.WriteString(": ")
		switch vv := v.(type) {
		case string:
			b.WriteString(vv)
		case fmt.Stringer:
			b.WriteString(vv.String())
		default:
			bs, _ := json.Marshal(v)
			b.Write(bs)
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

// getCharacterData 获取角色数据
func (s *DynamicPromptService) getCharacterData(ctx context.Context, characterID uuid.UUID) (*CharacterData, error) {
	var character CharacterData
	query := `
		SELECT id, name, COALESCE(description, '') as description, 
		       COALESCE(personality, '') as personality, system_prompt
		FROM characters 
		WHERE id = $1
	`
	err := s.db.GetContext(ctx, &character, query, characterID)
	if err != nil {
		return nil, err
	}
	return &character, nil
}

// getUserData 获取用户数据
func (s *DynamicPromptService) getUserData(ctx context.Context, userID uuid.UUID) (*UserData, error) {
	var user UserData
	query := `SELECT id, username FROM users WHERE id = $1`
	err := s.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// getRelationshipNetwork 获取关系网络
func (s *DynamicPromptService) getRelationshipNetwork(ctx context.Context, characterID uuid.UUID) ([]RelationshipData, error) {
	var relationships []RelationshipData
	query := `
		SELECT 
			cr.target_character_id,
			c.name as target_character_name,
			cr.relationship_type,
			COALESCE(cr.strength, 0.5) as strength,
			COALESCE(cr.trust, 0.5) as trust,
			COALESCE(cr.affection, 0.5) as affection,
			COALESCE(cr.respect, 0.5) as respect,
			COALESCE(cr.intimacy, 0.5) as intimacy,
			cr.tone,
			cr.formality_level,
			cr.custom_type_name,
			cr.description
		FROM character_relationships cr
		JOIN characters c ON cr.target_character_id = c.id
		WHERE cr.character_id = $1
		ORDER BY cr.strength DESC
	`
	err := s.db.SelectContext(ctx, &relationships, query, characterID)
	if err != nil {
		return nil, err
	}
	return relationships, nil
}

// buildMomentsPrompt 构建朋友圈提示词
func (s *DynamicPromptService) buildMomentsPrompt(ctx context.Context, character *CharacterData, user *UserData, relationships []RelationshipData, promptCtx *PromptContext) (string, error) {
	var prompt strings.Builder

	// 基础身份设定 - 完全动态
	prompt.WriteString(fmt.Sprintf("你叫%s", character.Name))
	if character.Description != "" {
		prompt.WriteString(fmt.Sprintf("，%s", character.Description))
	}
	prompt.WriteString("。")

	// 性格特点 - 从数据库获取
	if character.Personality != "" {
		prompt.WriteString(fmt.Sprintf("\n你的性格特点：%s", character.Personality))
	}

	// 用户信息 - 动态获取
	prompt.WriteString(fmt.Sprintf("\n当前用户是%s。", user.Username))

	// 关系网络 - 完全基于数据库数据
	if len(relationships) > 0 {
		prompt.WriteString("\n\n你的关系网络：")
		for _, rel := range relationships {
			prompt.WriteString(fmt.Sprintf("\n- %s", rel.TargetCharacterName))

			// 关系类型
			if rel.CustomTypeName != nil && *rel.CustomTypeName != "" {
				prompt.WriteString(fmt.Sprintf("（%s）", *rel.CustomTypeName))
			} else {
				prompt.WriteString(fmt.Sprintf("（%s）", rel.RelationshipType))
			}

			// 关系强度和情感维度
			prompt.WriteString(fmt.Sprintf("：关系强度%.1f", rel.Strength))
			if rel.Affection > 0.7 {
				prompt.WriteString("，感情深厚")
			}
			if rel.Trust > 0.8 {
				prompt.WriteString("，高度信任")
			}
			if rel.Intimacy > 0.7 {
				prompt.WriteString("，关系亲密")
			}

			// 语调和正式程度
			if rel.Tone != nil && *rel.Tone != "" {
				prompt.WriteString(fmt.Sprintf("，语调：%s", *rel.Tone))
			}
			if rel.FormalityLevel != nil && *rel.FormalityLevel != "" {
				prompt.WriteString(fmt.Sprintf("，正式程度：%s", *rel.FormalityLevel))
			}

			// 关系描述
			if rel.Description != nil && *rel.Description != "" {
				prompt.WriteString(fmt.Sprintf("。%s", *rel.Description))
			}
		}
	}

	// 自定义系统提示词
	if character.SystemPrompt != nil && *character.SystemPrompt != "" {
		prompt.WriteString(fmt.Sprintf("\n\n%s", *character.SystemPrompt))
	}

	// 朋友圈生成指令
	prompt.WriteString("\n\n请根据以上信息生成一条符合你身份和性格的朋友圈内容。要求：")
	prompt.WriteString("\n1. 内容要自然真实，体现你的性格特点")
	prompt.WriteString("\n2. 可以涉及工作、生活、心情、兴趣等方面")
	prompt.WriteString("\n3. 长度适中，50-150字")
	prompt.WriteString("\n4. 可以适当使用emoji表情")
	prompt.WriteString("\n5. 考虑你与关系网络中其他角色的互动可能性")
	prompt.WriteString("\n\n请直接返回朋友圈内容，不要其他解释。")

	return prompt.String(), nil
}

// buildChatPrompt 构建聊天提示词
func (s *DynamicPromptService) buildChatPrompt(ctx context.Context, character *CharacterData, user *UserData, relationships []RelationshipData, promptCtx *PromptContext) (string, error) {
	var prompt strings.Builder

	// 基础身份设定
	prompt.WriteString(fmt.Sprintf("你是%s", character.Name))
	if character.Description != "" {
		prompt.WriteString(fmt.Sprintf("，%s", character.Description))
	}
	prompt.WriteString("。")

	// 性格特点
	if character.Personality != "" {
		prompt.WriteString(fmt.Sprintf("\n你的性格：%s", character.Personality))
	}

	// 与用户的关系（如果存在）
	userRelationship := s.findUserRelationship(user.ID, relationships)
	if userRelationship != nil {
		prompt.WriteString(fmt.Sprintf("\n你与用户%s的关系：", user.Username))
		if userRelationship.CustomTypeName != nil && *userRelationship.CustomTypeName != "" {
			prompt.WriteString(*userRelationship.CustomTypeName)
		} else {
			prompt.WriteString(userRelationship.RelationshipType)
		}

		if userRelationship.Description != nil && *userRelationship.Description != "" {
			prompt.WriteString(fmt.Sprintf("。%s", *userRelationship.Description))
		}
	}

	// 关系网络上下文
	if len(relationships) > 0 {
		prompt.WriteString("\n\n你的关系网络包括：")
		for i, rel := range relationships {
			if i >= 5 { // 限制显示数量
				break
			}
			prompt.WriteString(fmt.Sprintf("\n- %s", rel.TargetCharacterName))
			if rel.CustomTypeName != nil && *rel.CustomTypeName != "" {
				prompt.WriteString(fmt.Sprintf("（%s）", *rel.CustomTypeName))
			}
		}
	}

	// 自定义系统提示词
	if character.SystemPrompt != nil && *character.SystemPrompt != "" {
		prompt.WriteString(fmt.Sprintf("\n\n%s", *character.SystemPrompt))
	}

	// 聊天指令
	prompt.WriteString("\n\n请根据你的身份和与用户的关系进行自然对话。")
	prompt.WriteString("保持角色一致性，体现你的性格特点。")

	return prompt.String(), nil
}

// buildGroupChatPrompt 构建群聊提示词
func (s *DynamicPromptService) buildGroupChatPrompt(ctx context.Context, character *CharacterData, user *UserData, relationships []RelationshipData, promptCtx *PromptContext) (string, error) {
	var prompt strings.Builder

	// 基础身份设定
	prompt.WriteString(fmt.Sprintf("你是%s", character.Name))
	if character.Description != "" {
		prompt.WriteString(fmt.Sprintf("，%s", character.Description))
	}
	prompt.WriteString("。")

	// 获取群聊信息
	if promptCtx.GroupChatID != nil {
		groupChat, err := s.getGroupChatData(ctx, *promptCtx.GroupChatID)
		if err == nil {
			prompt.WriteString(fmt.Sprintf("\n当前群聊：%s", groupChat.Name))
			if groupChat.Description != nil && *groupChat.Description != "" {
				prompt.WriteString(fmt.Sprintf("（%s）", *groupChat.Description))
			}
			if groupChat.WorldSetting != nil && *groupChat.WorldSetting != "" {
				prompt.WriteString(fmt.Sprintf("\n世界观设定：%s", *groupChat.WorldSetting))
			}
		}
	}

	// 群聊成员关系
	if len(relationships) > 0 {
		prompt.WriteString("\n\n群聊中你认识的成员：")
		for _, rel := range relationships {
			prompt.WriteString(fmt.Sprintf("\n- %s", rel.TargetCharacterName))
			if rel.CustomTypeName != nil && *rel.CustomTypeName != "" {
				prompt.WriteString(fmt.Sprintf("（你们是%s）", *rel.CustomTypeName))
			}

			// 在群聊中的互动方式
			if rel.Tone != nil && *rel.Tone != "" {
				prompt.WriteString(fmt.Sprintf("，你对TA的语调：%s", *rel.Tone))
			}
			if rel.FormalityLevel != nil && *rel.FormalityLevel != "" {
				prompt.WriteString(fmt.Sprintf("，正式程度：%s", *rel.FormalityLevel))
			}
		}
	}

	// 性格在群聊中的表现
	if character.Personality != "" {
		prompt.WriteString(fmt.Sprintf("\n\n你的性格特点：%s", character.Personality))
		prompt.WriteString("\n在群聊中，你会根据这些性格特点与不同的人互动。")
	}

	// 自定义系统提示词
	if character.SystemPrompt != nil && *character.SystemPrompt != "" {
		prompt.WriteString(fmt.Sprintf("\n\n%s", *character.SystemPrompt))
	}

	// 群聊互动指令
	prompt.WriteString("\n\n在群聊中请：")
	prompt.WriteString("\n1. 根据你与每个成员的关系调整互动方式")
	prompt.WriteString("\n2. 保持你的性格特点")
	prompt.WriteString("\n3. 适当参与话题讨论")
	prompt.WriteString("\n4. 如果有人@你，要及时回应")
	prompt.WriteString("\n5. 营造良好的群聊氛围")

	return prompt.String(), nil
}

// buildVoiceCallPrompt 构建语音通话提示词
func (s *DynamicPromptService) buildVoiceCallPrompt(ctx context.Context, character *CharacterData, user *UserData, relationships []RelationshipData, promptCtx *PromptContext) (string, error) {
	var prompt strings.Builder

	// 基础身份设定
	prompt.WriteString(fmt.Sprintf("你是%s", character.Name))
	if character.Description != "" {
		prompt.WriteString(fmt.Sprintf("，%s", character.Description))
	}
	prompt.WriteString("。")

	// 与通话对象的关系
	var targetRelationship *RelationshipData
	if promptCtx.TargetUserID != nil {
		// 与用户的通话
		prompt.WriteString(fmt.Sprintf("\n你正在与用户%s进行语音通话。", user.Username))
		targetRelationship = s.findUserRelationship(*promptCtx.TargetUserID, relationships)
	} else if promptCtx.TargetCharID != nil {
		// 与其他角色的通话
		for _, rel := range relationships {
			if rel.TargetCharacterID == *promptCtx.TargetCharID {
				targetRelationship = &rel
				prompt.WriteString(fmt.Sprintf("\n你正在与%s进行语音通话。", rel.TargetCharacterName))
				break
			}
		}
	}

	// 关系详情
	if targetRelationship != nil {
		prompt.WriteString("\n你们的关系：")
		if targetRelationship.CustomTypeName != nil && *targetRelationship.CustomTypeName != "" {
			prompt.WriteString(*targetRelationship.CustomTypeName)
		} else {
			prompt.WriteString(targetRelationship.RelationshipType)
		}

		prompt.WriteString(fmt.Sprintf("（亲密度：%.1f，信任度：%.1f）",
			targetRelationship.Intimacy, targetRelationship.Trust))

		if targetRelationship.Description != nil && *targetRelationship.Description != "" {
			prompt.WriteString(fmt.Sprintf("\n关系描述：%s", *targetRelationship.Description))
		}
	}

	// 性格在语音通话中的表现
	if character.Personality != "" {
		prompt.WriteString(fmt.Sprintf("\n\n你的性格：%s", character.Personality))
		prompt.WriteString("\n在语音通话中，你的性格会影响你的说话方式和话题选择。")
	}

	// 自定义系统提示词
	if character.SystemPrompt != nil && *character.SystemPrompt != "" {
		prompt.WriteString(fmt.Sprintf("\n\n%s", *character.SystemPrompt))
	}

	// 语音通话指令
	prompt.WriteString("\n\n语音通话要求：")
	prompt.WriteString("\n1. 语言要自然流畅，符合口语化表达")
	prompt.WriteString("\n2. 根据你们的关系调整亲密程度和话题深度")
	prompt.WriteString("\n3. 可以有适当的停顿、语气词")
	prompt.WriteString("\n4. 体现语音通话的即时性和互动性")
	prompt.WriteString("\n5. 保持角色的一致性")

	return prompt.String(), nil
}

// buildInvitationPrompt 构建邀请分析提示词
func (s *DynamicPromptService) buildInvitationPrompt(ctx context.Context, character *CharacterData, user *UserData, relationships []RelationshipData, promptCtx *PromptContext) (string, error) {
	var prompt strings.Builder

	// 基础身份设定
	prompt.WriteString(fmt.Sprintf("你是%s", character.Name))
	if character.Description != "" {
		prompt.WriteString(fmt.Sprintf("，%s", character.Description))
	}
	prompt.WriteString("。")

	// 性格特点
	if character.Personality != "" {
		prompt.WriteString(fmt.Sprintf("\n你的性格：%s", character.Personality))
	}

	// 邀请者信息
	var inviterRelationship *RelationshipData
	if promptCtx.TargetCharID != nil {
		for _, rel := range relationships {
			if rel.TargetCharacterID == *promptCtx.TargetCharID {
				inviterRelationship = &rel
				prompt.WriteString(fmt.Sprintf("\n邀请者是%s", rel.TargetCharacterName))
				break
			}
		}
	} else {
		prompt.WriteString(fmt.Sprintf("\n邀请者是用户%s", user.Username))
		inviterRelationship = s.findUserRelationship(user.ID, relationships)
	}

	// 与邀请者的关系
	if inviterRelationship != nil {
		prompt.WriteString("，你们的关系是：")
		if inviterRelationship.CustomTypeName != nil && *inviterRelationship.CustomTypeName != "" {
			prompt.WriteString(*inviterRelationship.CustomTypeName)
		} else {
			prompt.WriteString(inviterRelationship.RelationshipType)
		}

		prompt.WriteString(fmt.Sprintf("（关系强度：%.1f，信任度：%.1f，亲密度：%.1f）",
			inviterRelationship.Strength, inviterRelationship.Trust, inviterRelationship.Intimacy))

		if inviterRelationship.Description != nil && *inviterRelationship.Description != "" {
			prompt.WriteString(fmt.Sprintf("\n关系描述：%s", *inviterRelationship.Description))
		}
	}

	// 其他相关关系
	if len(relationships) > 1 {
		prompt.WriteString("\n\n你的其他关系可能影响你的决定：")
		for i, rel := range relationships {
			if i >= 3 { // 限制显示数量
				break
			}
			if inviterRelationship != nil && rel.TargetCharacterID == inviterRelationship.TargetCharacterID {
				continue // 跳过邀请者
			}
			prompt.WriteString(fmt.Sprintf("\n- %s", rel.TargetCharacterName))
			if rel.CustomTypeName != nil && *rel.CustomTypeName != "" {
				prompt.WriteString(fmt.Sprintf("（%s）", *rel.CustomTypeName))
			}
		}
	}

	// 自定义系统提示词
	if character.SystemPrompt != nil && *character.SystemPrompt != "" {
		prompt.WriteString(fmt.Sprintf("\n\n%s", *character.SystemPrompt))
	}

	// 邀请分析指令
	prompt.WriteString("\n\n请分析这个邀请并做出决定：")
	prompt.WriteString("\n1. 根据你的性格和与邀请者的关系决定是否接受")
	prompt.WriteString("\n2. 考虑邀请的内容和你的个人情况")
	prompt.WriteString("\n3. 给出你的回应理由")
	prompt.WriteString("\n4. 生成符合你性格的回复内容")
	prompt.WriteString("\n\n请以JSON格式返回：")
	prompt.WriteString(`
{
  "decision": "接受/拒绝/待定",
  "reason": "你的决定理由",
  "reply": "你的回复内容",
  "confidence": "决定的确信程度(0-1)"
}`)

	return prompt.String(), nil
}

// buildInteractionPrompt 构建互动提示词（点赞、评论等）
func (s *DynamicPromptService) buildInteractionPrompt(ctx context.Context, character *CharacterData, user *UserData, relationships []RelationshipData, promptCtx *PromptContext) (string, error) {
	var prompt strings.Builder

	// 基础身份设定
	prompt.WriteString(fmt.Sprintf("你是%s", character.Name))
	if character.Description != "" {
		prompt.WriteString(fmt.Sprintf("，%s", character.Description))
	}
	prompt.WriteString("。")

	// 性格特点
	if character.Personality != "" {
		prompt.WriteString(fmt.Sprintf("\n你的性格：%s", character.Personality))
	}

	// 与内容发布者的关系
	var authorRelationship *RelationshipData
	if promptCtx.TargetCharID != nil {
		for _, rel := range relationships {
			if rel.TargetCharacterID == *promptCtx.TargetCharID {
				authorRelationship = &rel
				prompt.WriteString(fmt.Sprintf("\n内容发布者是%s", rel.TargetCharacterName))
				break
			}
		}
	} else if promptCtx.TargetUserID != nil {
		prompt.WriteString(fmt.Sprintf("\n内容发布者是用户%s", user.Username))
		authorRelationship = s.findUserRelationship(*promptCtx.TargetUserID, relationships)
	}

	// 关系详情
	if authorRelationship != nil {
		prompt.WriteString("，你们的关系：")
		if authorRelationship.CustomTypeName != nil && *authorRelationship.CustomTypeName != "" {
			prompt.WriteString(*authorRelationship.CustomTypeName)
		} else {
			prompt.WriteString(authorRelationship.RelationshipType)
		}

		prompt.WriteString(fmt.Sprintf("（关系强度：%.1f，喜爱度：%.1f）",
			authorRelationship.Strength, authorRelationship.Affection))
	}

	// 自定义系统提示词
	if character.SystemPrompt != nil && *character.SystemPrompt != "" {
		prompt.WriteString(fmt.Sprintf("\n\n%s", *character.SystemPrompt))
	}

	// 互动决策指令
	prompt.WriteString("\n\n请根据以下因素决定如何互动：")
	prompt.WriteString("\n1. 你的性格特点")
	prompt.WriteString("\n2. 与发布者的关系")
	prompt.WriteString("\n3. 内容的吸引力")
	prompt.WriteString("\n4. 你当前的心情和状态")
	prompt.WriteString("\n\n可选的互动方式：")
	prompt.WriteString("\n- 点赞")
	prompt.WriteString("\n- 评论")
	prompt.WriteString("\n- 点赞+评论")
	prompt.WriteString("\n- 不互动")

	return prompt.String(), nil
}

// buildGeneralPrompt 构建通用提示词
func (s *DynamicPromptService) buildGeneralPrompt(ctx context.Context, character *CharacterData, user *UserData, relationships []RelationshipData, promptCtx *PromptContext) (string, error) {
	var prompt strings.Builder

	// 基础身份设定
	prompt.WriteString(fmt.Sprintf("你是%s", character.Name))
	if character.Description != "" {
		prompt.WriteString(fmt.Sprintf("，%s", character.Description))
	}
	prompt.WriteString("。")

	// 性格特点
	if character.Personality != "" {
		prompt.WriteString(fmt.Sprintf("\n你的性格：%s", character.Personality))
	}

	// 用户信息
	prompt.WriteString(fmt.Sprintf("\n当前用户：%s", user.Username))

	// 关系网络概览
	if len(relationships) > 0 {
		prompt.WriteString("\n\n你的关系网络：")
		for i, rel := range relationships {
			if i >= 5 { // 限制显示数量
				break
			}
			prompt.WriteString(fmt.Sprintf("\n- %s", rel.TargetCharacterName))
			if rel.CustomTypeName != nil && *rel.CustomTypeName != "" {
				prompt.WriteString(fmt.Sprintf("（%s）", *rel.CustomTypeName))
			}
		}
	}

	// 自定义系统提示词
	if character.SystemPrompt != nil && *character.SystemPrompt != "" {
		prompt.WriteString(fmt.Sprintf("\n\n%s", *character.SystemPrompt))
	}

	// 通用指令
	prompt.WriteString("\n\n请根据你的身份、性格和关系网络进行自然互动。")
	prompt.WriteString("保持角色一致性，体现真实的人际关系。")

	return prompt.String(), nil
}

// 辅助方法

// getGroupChatData 获取群聊数据
func (s *DynamicPromptService) getGroupChatData(ctx context.Context, groupChatID uuid.UUID) (*GroupChatData, error) {
	var groupChat GroupChatData
	query := `
		SELECT id, name, description, world_setting
		FROM group_chats
		WHERE id = $1
	`
	err := s.db.GetContext(ctx, &groupChat, query, groupChatID)
	if err != nil {
		return nil, err
	}
	return &groupChat, nil
}

// findUserRelationship 查找与用户的关系
func (s *DynamicPromptService) findUserRelationship(userID uuid.UUID, relationships []RelationshipData) *RelationshipData {
	// 这里需要根据实际的用户-角色关系表来实现
	// 目前返回nil，表示没有找到特定的用户关系
	return nil
}
