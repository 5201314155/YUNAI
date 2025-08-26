# YUNAI 智能邀请系统接口文档 v2.0

## 文档信息
- **版本**: v2.0
- **作者**: 小云
- **创建日期**: 2025-08-26
- **更新日期**: 2025-08-26
- **状态**: 已实现

## 摘要
本文档详细描述了YUNAI智能邀请系统的所有API接口，包括关系网络分析、聊天内容分析、邀请建议生成、邀请响应处理等核心功能的接口规范。

## 目录
1. [接口概览](#接口概览)
2. [认证方式](#认证方式)
3. [数据格式](#数据格式)
4. [错误码](#错误码)
5. [核心接口](#核心接口)
6. [数据结构](#数据结构)
7. [示例代码](#示例代码)
8. [更新记录](#更新记录)

## 接口概览

### 基础信息
- **Base URL**: `https://api.yunai.com/v2`
- **协议**: HTTPS
- **数据格式**: JSON
- **字符编码**: UTF-8

### 接口列表
| 接口名称 | 方法 | 路径 | 功能描述 |
|---------|------|------|----------|
| 获取关系网络图谱 | GET | `/invitation/relationship-map` | 获取用户的完整关系网络 |
| 分析聊天内容 | POST | `/invitation/analyze-chat` | 分析聊天内容并生成邀请建议 |
| 生成邀请弹窗 | POST | `/invitation/generate-popup` | 生成邀请确认弹窗 |
| 处理邀请响应 | POST | `/invitation/respond` | 处理用户的邀请响应 |
| 获取角色反应 | GET | `/invitation/character-reaction/{id}` | 获取角色的反应内容 |
| 发送私信 | POST | `/invitation/send-private-message` | 发送私信消息 |

## 认证方式

### JWT Token认证
```http
Authorization: Bearer <jwt_token>
```

### 请求头要求
```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <jwt_token>
```

## 数据格式

### 统一响应格式
```json
{
    "code": 200,
    "message": "success",
    "data": {},
    "timestamp": "2025-08-26T17:00:00Z",
    "request_id": "uuid"
}
```

### 分页格式
```json
{
    "items": [],
    "total": 100,
    "page": 1,
    "limit": 20,
    "has_more": true
}
```

## 错误码

| 错误码 | 描述 | 解决方案 |
|--------|------|----------|
| 200 | 成功 | - |
| 400 | 请求参数错误 | 检查请求参数格式 |
| 401 | 未授权 | 检查认证token |
| 403 | 权限不足 | 检查用户权限 |
| 404 | 资源不存在 | 检查资源ID |
| 429 | 请求频率限制 | 降低请求频率 |
| 500 | 服务器内部错误 | 联系技术支持 |

## 核心接口

### 1. 获取关系网络图谱

#### 接口信息
- **方法**: GET
- **路径**: `/invitation/relationship-map`
- **功能**: 获取用户的完整关系网络图谱

#### 请求参数
```http
GET /invitation/relationship-map?user_id=uuid
```

#### 响应示例
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
        "relationship_groups": {
            "friend": {
                "relationship_type": {
                    "id": "61d41a76-1653-48f9-9c63-dbe8139c707a",
                    "name": "friend",
                    "description": "朋友关系"
                },
                "characters": [
                    {
                        "character": {
                            "id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
                            "name": "小雪",
                            "personality": "害羞、内向、温柔、喜欢读书"
                        },
                        "relationship": {
                            "custom_type_name": "bestfriend",
                            "strength": 0.87,
                            "trust": 0.9,
                            "affection": 0.85
                        },
                        "match_score": 0.87
                    }
                ]
            }
        },
        "total_characters": 11
    }
}
```

### 2. 分析聊天内容

#### 接口信息
- **方法**: POST
- **路径**: `/invitation/analyze-chat`
- **功能**: 分析聊天内容并生成邀请建议

#### 请求参数
```json
{
    "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
    "group_chat_id": "d37fe579-d8e6-4949-85dd-ee4c7fc5ccb2",
    "chat_message": "今天工作压力好大，希望有朋友能安慰我一下",
    "sender_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
    "message_type": "text"
}
```

#### 响应示例
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "analyzed_at": "2025-08-26T17:00:00Z",
        "chat_message": "今天工作压力好大，希望有朋友能安慰我一下",
        "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
        "group_chat_id": "d37fe579-d8e6-4949-85dd-ee4c7fc5ccb2",
        "success": true,
        "message": "基于关系网络找到了 1 个邀请建议",
        "suggestions": [
            {
                "character": {
                    "id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
                    "name": "小雪",
                    "personality": "害羞、内向、温柔、喜欢读书"
                },
                "relationship": {
                    "custom_type_name": "bestfriend",
                    "strength": 0.87
                },
                "match_score": 0.94,
                "invite_reason": "基于你们的bestfriend关系，小雪很温柔，擅长安慰人",
                "expected_reaction": "应该会很高兴地同意",
                "relationship_desc": "你们是bestfriend关系，关系强度: 0.9",
                "priority": 1,
                "created_at": "2025-08-26T17:00:00Z"
            }
        ]
    }
}
```

### 3. 生成邀请弹窗

#### 接口信息
- **方法**: POST
- **路径**: `/invitation/generate-popup`
- **功能**: 生成邀请确认弹窗

#### 请求参数
```json
{
    "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
    "group_chat_id": "d37fe579-d8e6-4949-85dd-ee4c7fc5ccb2",
    "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
    "invite_reason": "基于你们的bestfriend关系，小雪很温柔，擅长安慰人",
    "expected_reaction": "应该会很高兴地同意",
    "match_score": 0.94
}
```

#### 响应示例
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
        "character_name": "小雪",
        "character_avatar": null,
        "group_chat_id": "d37fe579-d8e6-4949-85dd-ee4c7fc5ccb2",
        "group_chat_name": "朋友聚会群",
        "invite_reason": "基于你们的bestfriend关系，小雪很温柔，擅长安慰人",
        "relationship_desc": "你们认识",
        "expected_reaction": "应该会很高兴地同意",
        "match_score": 0.94,
        "priority": 1,
        "popup_title": "邀请 小雪 加入群聊",
        "popup_message": "你们认识，基于你们的bestfriend关系，小雪很温柔，擅长安慰人。\n\n应该会很高兴地同意",
        "accept_button_text": "邀请加入",
        "reject_button_text": "暂时不邀请",
        "created_at": "2025-08-26T17:00:00Z"
    }
}
```

### 4. 处理邀请响应

#### 接口信息
- **方法**: POST
- **路径**: `/invitation/respond`
- **功能**: 处理用户的邀请响应（接受或拒绝）

#### 请求参数
```json
{
    "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
    "group_chat_id": "d37fe579-d8e6-4949-85dd-ee4c7fc5ccb2",
    "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
    "action": "accept",
    "reason": "现在不太方便"
}
```

#### 响应示例（接受邀请）
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "processed_at": "2025-08-26T17:00:00Z",
        "success": true,
        "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
        "action": "accept",
        "message": "角色已成功加入群聊",
        "character_reaction": {
            "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
            "character_name": "小雪",
            "action": "accept_invitation",
            "content": "(温柔地笑) 谢谢你的邀请，我很高兴能加入大家。希望能和大家好好相处呢。",
            "emotion": "happy",
            "intensity": 0.8,
            "generated_at": "2025-08-26T17:00:00Z"
        }
    }
}
```

#### 响应示例（拒绝邀请）
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "processed_at": "2025-08-26T17:00:00Z",
        "success": true,
        "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
        "action": "reject",
        "message": "角色拒绝了邀请",
        "character_reaction": {
            "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
            "character_name": "小雪",
            "action": "reject_invitation",
            "content": "(低头) 对不起...我现在还不太适应群聊，可能会给大家添麻烦...下次有机会再说吧。",
            "emotion": "apologetic",
            "intensity": 0.6,
            "generated_at": "2025-08-26T17:00:00Z"
        },
        "private_message": {
            "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
            "character_name": "小雪",
            "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
            "content": "(私信) 刚才真的很抱歉...其实我很想和大家一起聊天的，只是我在群里会很紧张。如果你不介意的话，我们可以私下聊聊天吗？",
            "message_type": "text",
            "sent_at": "2025-08-26T17:00:00Z",
            "success": true
        }
    }
}
```

### 5. 获取角色反应

#### 接口信息
- **方法**: GET
- **路径**: `/invitation/character-reaction/{reaction_id}`
- **功能**: 获取特定的角色反应内容

#### 请求参数
```http
GET /invitation/character-reaction/uuid
```

#### 响应示例
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "reaction-uuid",
        "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
        "character_name": "小雪",
        "action": "accept_invitation",
        "content": "(温柔地笑) 谢谢你的邀请，我很高兴能加入大家。希望能和大家好好相处呢。",
        "emotion": "happy",
        "intensity": 0.8,
        "generated_at": "2025-08-26T17:00:00Z"
    }
}
```

### 6. 发送私信

#### 接口信息
- **方法**: POST
- **路径**: `/invitation/send-private-message`
- **功能**: 发送私信消息

#### 请求参数
```json
{
    "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
    "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
    "content": "(私信) 刚才真的很抱歉...其实我很想和大家一起聊天的，只是我在群里会很紧张。如果你不介意的话，我们可以私下聊聊天吗？",
    "message_type": "text",
    "context": "rejection_followup"
}
```

#### 响应示例
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "message_id": "message-uuid",
        "character_id": "22ba2638-4fbd-4bab-96fe-240ead7d60ff",
        "character_name": "小雪",
        "user_id": "daa19f83-c430-41d1-9aba-cc73176b5582",
        "content": "(私信) 刚才真的很抱歉...其实我很想和大家一起聊天的，只是我在群里会很紧张。如果你不介意的话，我们可以私下聊聊天吗？",
        "message_type": "text",
        "sent_at": "2025-08-26T17:00:00Z",
        "success": true
    }
}
```

## 数据结构

### 关系网络图谱
```typescript
interface UserRelationshipMap {
    user_id: string;
    relationship_groups: Record<string, RelationshipGroup>;
    total_characters: number;
}

interface RelationshipGroup {
    relationship_type: RelationshipType;
    characters: RelationshipCharacter[];
    description?: string;
    custom_prompt?: string;
}

interface RelationshipCharacter {
    character: Character;
    relationship: CharacterRelationship;
    match_score: number;
}
```

### 邀请建议
```typescript
interface InvitationSuggestion {
    character: Character;
    relationship: CharacterRelationship;
    relationship_type: RelationshipType;
    match_score: number;
    invite_reason: string;
    expected_reaction: string;
    relationship_desc: string;
    priority: number;
    created_at: string;
}
```

### 角色反应
```typescript
interface CharacterReaction {
    character_id: string;
    character_name: string;
    action: string;
    content: string;
    emotion: string;
    intensity: number;
    generated_at: string;
}
```

## 示例代码

### JavaScript/TypeScript
```typescript
// 分析聊天内容
async function analyzeChatForInvitation(chatMessage: string, groupChatId: string) {
    const response = await fetch('/api/v2/invitation/analyze-chat', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
            user_id: userId,
            group_chat_id: groupChatId,
            chat_message: chatMessage,
            message_type: 'text'
        })
    });
    
    const result = await response.json();
    return result.data;
}

// 处理邀请响应
async function respondToInvitation(characterId: string, action: 'accept' | 'reject', reason?: string) {
    const response = await fetch('/api/v2/invitation/respond', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
            character_id: characterId,
            group_chat_id: groupChatId,
            user_id: userId,
            action: action,
            reason: reason
        })
    });
    
    const result = await response.json();
    return result.data;
}
```

### Go
```go
// 分析聊天内容
func AnalyzeChatForInvitation(ctx context.Context, req *domain.ChatAnalysisRequest) (*domain.RelationshipBasedInvitationResult, error) {
    return invitationService.AnalyzeChatAndSuggestInvitation(ctx, req)
}

// 处理邀请响应
func ProcessInvitationResponse(ctx context.Context, req *domain.InvitationResponseRequest) (*domain.InvitationResponseResult, error) {
    return invitationService.ProcessInvitationResponse(ctx, req)
}
```

## 更新记录

### v2.0 (2025-08-26)
- **新增**: 关系网络图谱接口
- **新增**: 聊天内容分析接口
- **新增**: 邀请弹窗生成接口
- **新增**: 邀请响应处理接口
- **新增**: 角色反应获取接口
- **新增**: 私信发送接口
- **优化**: 统一响应格式
- **优化**: 错误码规范

### v1.0 (2025-08-25)
- **初版**: 基础邀请接口

---

**文档维护**: 本文档由小云维护，如有问题请联系开发团队。
