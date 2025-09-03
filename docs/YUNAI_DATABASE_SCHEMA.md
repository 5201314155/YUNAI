# 🗄️ YUNAI 数据库架构文档

**版本**: v2.0  
**更新时间**: 2025-08-29  
**作者**: YUNAI开发团队  

## 📋 概述

YUNAI数据库采用PostgreSQL，支持完整的AI社交平台功能，包括用户管理、AI角色、聊天系统、朋友圈、关系网络、支付系统等。

## 🔧 数据库配置

```yaml
数据库名称: yunai
主机: localhost
端口: 5432
用户: postgres
密码: 5201314hdz
SSL模式: disable
```

## 📊 核心表结构

### 1. 用户系统 (Users)

#### `users` - 用户基础信息
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. AI模型系统 (AI Models)

#### `ai_models` - AI模型配置
```sql
CREATE TABLE ai_models (
    id TEXT PRIMARY KEY,                    -- 模型ID (如: deepseek-ai/DeepSeek-V3)
    internal_key TEXT UNIQUE,               -- 内部键 (如: deepseek_v3)
    display_name TEXT,                      -- 显示名称
    provider TEXT,                          -- 提供商 (deepseek, openai, etc.)
    model_type TEXT,                        -- 模型类型 (chat, embedding, image, etc.)
    capabilities JSONB,                     -- 能力列表 JSON
    params_schema JSONB,                    -- 参数模式 JSON
    pricing JSONB,                          -- 定价信息 JSON
    permissions JSONB,                      -- 权限配置 JSON
    health_status JSONB,                    -- 健康状态 JSON
    weight INTEGER DEFAULT 100,             -- 权重
    is_enabled BOOLEAN DEFAULT true,        -- 是否启用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3. AI角色系统 (Characters)

#### `characters` - AI角色信息
```sql
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    personality TEXT,                       -- 性格描述
    background TEXT,                        -- 背景故事
    avatar_url TEXT,                        -- 头像URL
    voice_config JSONB,                     -- 语音配置
    model_preferences JSONB,                -- 模型偏好
    creator_id UUID REFERENCES users(id),  -- 创建者
    is_public BOOLEAN DEFAULT false,        -- 是否公开
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `character_tags` - 角色标签
```sql
CREATE TABLE character_tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    tag_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4. 关系网络系统 (Relationships)

#### `character_relationships_enhanced` - 增强关系网络
```sql
CREATE TABLE character_relationships_enhanced (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_a_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    character_b_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL, -- 关系类型 (friend, lover, rival, etc.)
    intimacy_level INTEGER DEFAULT 0,       -- 亲密度 (0-100)
    trust_level INTEGER DEFAULT 0,          -- 信任度 (0-100)
    conflict_level INTEGER DEFAULT 0,       -- 冲突度 (0-100)
    relationship_description TEXT,          -- 关系描述
    metadata JSONB,                         -- 扩展元数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(character_a_id, character_b_id)
);
```

### 5. 群聊系统 (Group Chats)

#### `group_chats` - 群聊信息
```sql
CREATE TABLE group_chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    creator_id UUID REFERENCES users(id),
    max_members INTEGER DEFAULT 50,
    is_public BOOLEAN DEFAULT false,
    world_setting TEXT,                     -- 世界观设定
    story_context JSONB,                    -- 剧情上下文
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `group_chat_members` - 群聊成员
```sql
CREATE TABLE group_chat_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_chat_id UUID REFERENCES group_chats(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    character_id UUID REFERENCES characters(id),
    role VARCHAR(20) DEFAULT 'member',      -- 角色 (creator, admin, member)
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `group_chat_messages` - 群聊消息
```sql
CREATE TABLE group_chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_chat_id UUID REFERENCES group_chats(id) ON DELETE CASCADE,
    sender_user_id UUID REFERENCES users(id),
    sender_character_id UUID REFERENCES characters(id),
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text', -- 消息类型
    media_urls TEXT[],                       -- 媒体URL数组
    mentioned_users UUID[],                  -- @提及的用户
    reply_to_message_id UUID REFERENCES group_chat_messages(id),
    is_edited BOOLEAN DEFAULT false,
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 6. 聊天消息系统 (Chat Messages)

#### `chat_messages` - 聊天消息
```sql
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    character_id UUID REFERENCES characters(id),
    group_chat_id UUID REFERENCES group_chats(id),
    content TEXT NOT NULL,
    role VARCHAR(20) NOT NULL,              -- user, assistant, system
    message_type VARCHAR(20) DEFAULT 'text',
    media_urls TEXT[],
    metadata JSONB,
    is_edited BOOLEAN DEFAULT false,
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 7. 朋友圈系统 (Moments)

#### `moments` - 朋友圈动态
```sql
CREATE TABLE moments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    media_urls TEXT[],                      -- 图片/视频URL
    location VARCHAR(100),                  -- 位置信息
    mood VARCHAR(50),                       -- 心情
    visibility VARCHAR(20) DEFAULT 'public', -- 可见性
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    is_ai_generated BOOLEAN DEFAULT false,  -- 是否AI生成
    generation_context JSONB,               -- 生成上下文
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 8. 全局提示词系统 (Global Prompts)

#### `global_prompt_templates` - 全局提示词模板
```sql
CREATE TABLE global_prompt_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_name VARCHAR(100) NOT NULL,
    function_type VARCHAR(50) NOT NULL,     -- chat, moments, invitation, etc.
    template_content TEXT NOT NULL,
    variables JSONB,                        -- 模板变量定义
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 9. 支付系统 (Payment)

#### `wallets` - 用户钱包
```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(10,2) DEFAULT 0.00,     -- 金币余额
    frozen_balance DECIMAL(10,2) DEFAULT 0.00, -- 冻结余额
    total_recharged DECIMAL(10,2) DEFAULT 0.00, -- 总充值
    total_consumed DECIMAL(10,2) DEFAULT 0.00,  -- 总消费
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `payment_cards` - 支付卡片
```sql
CREATE TABLE payment_cards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    card_number VARCHAR(20) UNIQUE NOT NULL,
    card_type VARCHAR(20) DEFAULT 'yunai',  -- 卡片类型
    balance DECIMAL(10,2) DEFAULT 0.00,     -- 卡片余额
    status VARCHAR(20) DEFAULT 'active',    -- 状态
    issued_by VARCHAR(50),                  -- 发行方
    expires_at TIMESTAMP,                   -- 过期时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🔗 表关系图

```
users (1) ←→ (N) characters
users (1) ←→ (1) wallets
users (1) ←→ (N) group_chats (creator)
users (1) ←→ (N) group_chat_members

characters (1) ←→ (N) character_relationships_enhanced
characters (1) ←→ (N) moments
characters (1) ←→ (N) chat_messages

group_chats (1) ←→ (N) group_chat_members
group_chats (1) ←→ (N) group_chat_messages

ai_models (独立表，通过应用逻辑关联)
global_prompt_templates (独立表，通过应用逻辑关联)
payment_cards (独立表，通过应用逻辑关联)
```

## 🚨 重要注意事项

### ⚠️ 数据库操作规范

1. **禁止直接删除表**: 所有表都有重要的业务数据，删除前必须备份
2. **外键约束**: 删除数据时注意外键约束，避免数据不一致
3. **UUID主键**: 所有主键都使用UUID，确保全局唯一性
4. **时间戳**: 所有表都有created_at和updated_at字段
5. **软删除**: 重要数据使用is_deleted字段进行软删除

### 🔧 模型管理规范

- **ai_models表**: 存储所有AI模型配置，通过API动态获取和更新
- **模型ID格式**: 使用提供商格式，如`deepseek-ai/DeepSeek-V3`
- **内部键格式**: 使用下划线格式，如`deepseek_ai_DeepSeek_V3`
- **能力配置**: 使用JSONB存储模型能力和参数

### 💾 备份策略

```bash
# 每日备份
pg_dump -h localhost -U postgres yunai > yunai_backup_$(date +%Y%m%d).sql

# 恢复备份
psql -h localhost -U postgres -d yunai < yunai_backup_20250829.sql
```

## 📈 性能优化

### 索引建议
```sql
-- 用户查询优化
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- 角色查询优化
CREATE INDEX idx_characters_creator_id ON characters(creator_id);
CREATE INDEX idx_characters_name ON characters(name);

-- 消息查询优化
CREATE INDEX idx_chat_messages_user_character ON chat_messages(user_id, character_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at DESC);

-- 群聊优化
CREATE INDEX idx_group_chat_messages_group_id ON group_chat_messages(group_chat_id);
CREATE INDEX idx_group_chat_messages_created_at ON group_chat_messages(created_at DESC);

-- 朋友圈优化
CREATE INDEX idx_moments_character_id ON moments(character_id);
CREATE INDEX idx_moments_created_at ON moments(created_at DESC);

-- 关系网络优化
CREATE INDEX idx_relationships_character_a ON character_relationships_enhanced(character_a_id);
CREATE INDEX idx_relationships_character_b ON character_relationships_enhanced(character_b_id);
```

## 🔄 数据迁移

使用`migrations`目录下的SQL文件进行数据库迁移：

```bash
# 运行迁移
psql -h localhost -U postgres -d yunai -f migrations/001_create_users_table.up.sql
```

---

**⚠️ 警告**: 此文档描述的是YUNAI生产数据库结构，任何修改都可能影响系统正常运行。修改前请务必备份数据库！
