# YUNAI数据库表结构文档

## 📋 数据库概述

YUNAI系统使用PostgreSQL作为主数据库，包含核心业务表和扩展功能表。所有表都使用UUID作为主键，支持完整的关系型数据管理。

## 🏗️ 核心业务表

### 1. core_users - 核心用户表
```sql
CREATE TABLE core_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储用户基本信息和认证数据

### 2. core_characters - 核心AI角色表
```sql
CREATE TABLE core_characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    personality JSONB,
    bg_image_url TEXT,
    cutout_image_url TEXT,
    system_prompt TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储AI角色的基本信息和个性设定

### 3. core_moments - 核心朋友圈表
```sql
CREATE TABLE core_moments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    character_id UUID REFERENCES core_characters(id),
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text',
    media_url TEXT,
    media_type VARCHAR(20),
    tags TEXT[],
    mood VARCHAR(50),
    location VARCHAR(200),
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    visibility VARCHAR(20) DEFAULT 'friends',
    is_generated BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储朋友圈动态内容

### 4. core_group_chats - 核心群聊表
```sql
CREATE TABLE core_group_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    creator_user_id UUID REFERENCES core_users(id),
    world_setting TEXT,
    background_image_url TEXT,
    background_music_url TEXT,
    is_public BOOLEAN DEFAULT false,
    member_count INTEGER DEFAULT 0,
    max_members INTEGER DEFAULT 500,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储群聊基本信息

### 5. group_chat_members - 群聊成员表
```sql
CREATE TABLE group_chat_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID REFERENCES core_group_chats(id),
    user_id UUID REFERENCES core_users(id),
    character_id UUID REFERENCES core_characters(id),
    member_type VARCHAR(20) NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    status VARCHAR(20) DEFAULT 'active',
    joined_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 管理群聊成员关系

### 6. chat_messages - 聊天消息表
```sql
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    character_id UUID REFERENCES core_characters(id),
    group_chat_id UUID REFERENCES core_group_chats(id),
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text',
    media_url TEXT,
    reply_to_id UUID,
    is_ai_generated BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储聊天消息内容

## 🔧 扩展功能表

### 7. user_identity_contexts - 用户身份上下文表
```sql
CREATE TABLE user_identity_contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    context_type VARCHAR(50) NOT NULL,
    context_id UUID,
    identity_type VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    identity_source VARCHAR(50),
    extraction_context TEXT,
    confidence_score FLOAT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储用户在不同上下文中的身份信息

### 8. global_prompt_templates - 全局提示词模板表
```sql
CREATE TABLE global_prompt_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    template_content TEXT NOT NULL,
    category VARCHAR(50),
    priority INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储全局提示词模板

### 9. world_settings - 世界观设定表
```sql
CREATE TABLE world_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    setting_content TEXT NOT NULL,
    background_story TEXT,
    rules_and_laws TEXT,
    default_background_image TEXT,
    default_music_url TEXT,
    color_scheme JSONB,
    priority INTEGER DEFAULT 100,
    applies_to_characters BOOLEAN DEFAULT true,
    applies_to_groups BOOLEAN DEFAULT true,
    applies_to_moments BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储世界观设定信息

### 10. voice_calls - 语音通话表
```sql
CREATE TABLE voice_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    character_id UUID REFERENCES core_characters(id),
    call_type VARCHAR(20) DEFAULT 'voice',
    duration INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'completed',
    quality_score FLOAT,
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP
);
```
**用途**: 存储语音通话记录

### 11. voice_messages - 语音消息表
```sql
CREATE TABLE voice_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID REFERENCES voice_calls(id),
    user_id UUID REFERENCES core_users(id),
    character_id UUID REFERENCES core_characters(id),
    text_content TEXT,
    audio_url TEXT,
    duration INTEGER,
    sender_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储语音通话中的消息

### 12. ai_invitations - AI邀请表
```sql
CREATE TABLE ai_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID REFERENCES core_group_chats(id),
    inviter_character_id UUID REFERENCES core_characters(id),
    invited_character_id UUID REFERENCES core_characters(id),
    invitation_reason TEXT,
    invitation_message TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    ai_decision_reasoning TEXT,
    confidence_score FLOAT,
    created_at TIMESTAMP DEFAULT NOW(),
    responded_at TIMESTAMP,
    expires_at TIMESTAMP
);
```
**用途**: 存储AI角色之间的邀请记录

## 💰 支付相关表

### 13. wallets - 钱包表
```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    balance DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(10) DEFAULT 'CNY',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储用户钱包信息

### 14. wallet_transactions - 钱包交易表
```sql
CREATE TABLE wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID REFERENCES wallets(id),
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    description TEXT,
    reference_id UUID,
    status VARCHAR(20) DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储钱包交易记录

## 🎭 剧情系统表

### 15. stories - 故事表
```sql
CREATE TABLE stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    world_setting TEXT,
    genre VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储故事基本信息

### 16. story_chapters - 故事章节表
```sql
CREATE TABLE story_chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID REFERENCES stories(id),
    chapter_number INTEGER NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    background_image_url TEXT,
    background_music_url TEXT,
    background_sfx_url TEXT,
    performance_settings JSONB,
    trigger_conditions JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储故事章节信息

## 🔗 关系网络表

### 17. character_relationships - 角色关系表
```sql
CREATE TABLE character_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_a_id UUID REFERENCES core_characters(id),
    character_b_id UUID REFERENCES core_characters(id),
    relationship_type VARCHAR(50) NOT NULL,
    relationship_strength FLOAT DEFAULT 0.5,
    description TEXT,
    is_mutual BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储角色之间的关系

## 📊 系统配置表

### 18. system_configs - 系统配置表
```sql
CREATE TABLE system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
**用途**: 存储系统配置参数

## 🔍 索引建议

### 性能优化索引
```sql
-- 用户相关索引
CREATE INDEX idx_core_users_email ON core_users(email);
CREATE INDEX idx_core_users_username ON core_users(username);

-- 角色相关索引
CREATE INDEX idx_core_characters_name ON core_characters(name);
CREATE INDEX idx_core_characters_active ON core_characters(is_active);

-- 朋友圈相关索引
CREATE INDEX idx_core_moments_user_id ON core_moments(user_id);
CREATE INDEX idx_core_moments_character_id ON core_moments(character_id);
CREATE INDEX idx_core_moments_created_at ON core_moments(created_at);

-- 聊天相关索引
CREATE INDEX idx_chat_messages_group_chat_id ON chat_messages(group_chat_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);

-- 身份上下文索引
CREATE INDEX idx_user_identity_contexts_user_id ON user_identity_contexts(user_id);
CREATE INDEX idx_user_identity_contexts_context ON user_identity_contexts(context_type, context_id);
```

## ⚠️ 重要注意事项

### 外键约束修复
由于历史原因，部分表的外键约束指向了错误的表。部署时必须运行外键修复脚本：

```sql
-- 关键外键修复
ALTER TABLE user_identity_contexts DROP CONSTRAINT IF EXISTS user_identity_contexts_user_id_fkey;
ALTER TABLE user_identity_contexts ADD CONSTRAINT user_identity_contexts_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);

ALTER TABLE world_settings DROP CONSTRAINT IF EXISTS world_settings_user_id_fkey;
ALTER TABLE world_settings ADD CONSTRAINT world_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);

ALTER TABLE voice_calls DROP CONSTRAINT IF EXISTS voice_calls_user_id_fkey;
ALTER TABLE voice_calls ADD CONSTRAINT voice_calls_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);

ALTER TABLE voice_calls DROP CONSTRAINT IF EXISTS voice_calls_character_id_fkey;
ALTER TABLE voice_calls ADD CONSTRAINT voice_calls_character_id_fkey FOREIGN KEY (character_id) REFERENCES core_characters(id);
```

### JSONB字段处理
部分表使用JSONB字段存储复杂数据：
- `core_characters.personality` - 角色个性数据
- `world_settings.color_scheme` - 颜色方案配置
- `story_chapters.performance_settings` - 表演设置
- `story_chapters.trigger_conditions` - 触发条件

**注意**: 在Go代码中查询JSONB字段时，建议使用原生SQL避免扫描错误。

### 数组字段处理
PostgreSQL数组字段：
- `core_moments.tags` - 标签数组 (text[])

**注意**: 插入数组数据时需要使用特殊格式或原生SQL。

## 🔄 数据迁移

### 自动迁移
服务器启动时会自动执行GORM迁移：
```go
db.AutoMigrate(&CoreUser{}, &CoreCharacter{}, &CoreMoment{}, ...)
```

### 手动迁移
如果需要手动创建表结构，请参考`database/init_database.sql`文件。

## 📈 数据统计

### 当前表数量
- **核心业务表**: 6个
- **扩展功能表**: 12个
- **总计**: 18个主要表

### 关系复杂度
- **一对多关系**: 15个
- **多对多关系**: 3个
- **自引用关系**: 2个

## 🛡️ 数据安全

### 敏感数据处理
- 用户密码使用bcrypt哈希存储
- 个人信息支持软删除
- 重要操作记录审计日志

### 备份策略
- 建议每日自动备份
- 重要更新前手动备份
- 保留至少7天的备份历史

## 🔧 维护建议

### 定期维护任务
1. 清理过期的会话数据
2. 压缩历史聊天记录
3. 优化数据库索引
4. 分析慢查询日志

### 监控指标
- 表大小增长趋势
- 查询性能指标
- 外键约束违反次数
- 数据库连接池状态

## 📋 完整表清单

### 核心表 (6个)
1. `core_users` - 用户表
2. `core_characters` - AI角色表
3. `core_moments` - 朋友圈表
4. `core_group_chats` - 群聊表
5. `group_chat_members` - 群聊成员表
6. `chat_messages` - 聊天消息表

### 扩展表 (12个)
7. `user_identity_contexts` - 用户身份上下文表
8. `global_prompt_templates` - 全局提示词模板表
9. `world_settings` - 世界观设定表
10. `voice_calls` - 语音通话表
11. `voice_messages` - 语音消息表
12. `ai_invitations` - AI邀请表
13. `wallets` - 钱包表
14. `wallet_transactions` - 钱包交易表
15. `stories` - 故事表
16. `story_chapters` - 故事章节表
17. `character_relationships` - 角色关系表
18. `system_configs` - 系统配置表

### 系统表
- `user_sessions` - 用户会话表
- `moment_comments` - 朋友圈评论表
- `moment_likes` - 朋友圈点赞表
- `moment_shares` - 朋友圈分享表
- `story_trigger_logs` - 剧情触发日志表
