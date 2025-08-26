-- 创建角色表
CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 基本信息
    name VARCHAR(100) NOT NULL,
    description TEXT,
    personality TEXT,
    
    -- 角色两图
    bg_image_url TEXT, -- 背景图URL
    cutout_image_url TEXT, -- 抠图URL
    
    -- 图片元数据
    bg_image_width INTEGER,
    bg_image_height INTEGER,
    cutout_image_width INTEGER,
    cutout_image_height INTEGER,
    
    -- 模型配置
    default_model_id UUID REFERENCES ai_models(id) ON DELETE SET NULL,
    model_params JSONB DEFAULT '{}',
    system_prompt TEXT,
    
    -- 可见性设置
    visibility VARCHAR(20) NOT NULL DEFAULT 'private' CHECK (visibility IN ('private', 'public', 'friends')),
    is_featured BOOLEAN DEFAULT FALSE,
    
    -- 社交设置
    allow_chat BOOLEAN DEFAULT TRUE,
    allow_group_chat BOOLEAN DEFAULT TRUE,
    allow_calls BOOLEAN DEFAULT TRUE,
    
    -- 统计信息
    chat_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建角色标签表
CREATE TABLE IF NOT EXISTS character_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(character_id, tag)
);

-- 创建角色关系表
CREATE TABLE IF NOT EXISTS character_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    target_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- relationship type and strength
    relationship_type VARCHAR(50) NOT NULL, -- friend, rival, lover, family, etc.
    strength DECIMAL(3,2) DEFAULT 0.5 CHECK (strength >= 0 AND strength <= 1),
    
    -- emotional dimensions
    trust DECIMAL(3,2) DEFAULT 0.5 CHECK (trust >= 0 AND trust <= 1),
    affection DECIMAL(3,2) DEFAULT 0.5 CHECK (affection >= 0 AND affection <= 1),
    respect DECIMAL(3,2) DEFAULT 0.5 CHECK (respect >= 0 AND respect <= 1),

    -- addressing and tone
    address_name VARCHAR(50), -- how to address the target
    tone VARCHAR(50), -- tone: formal, casual, intimate, hostile

    -- trigger rules
    trigger_keywords TEXT[], -- trigger keywords
    trigger_emotions TEXT[], -- trigger emotions
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(character_id, target_character_id)
);

-- 创建群聊表
CREATE TABLE IF NOT EXISTS group_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 群聊信息
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- background settings
    background_image_url TEXT, -- group chat global background
    background_music_url TEXT, -- background music
    background_sfx_url TEXT, -- background sound effects

    -- group chat settings
    max_members INTEGER DEFAULT 10,
    is_public BOOLEAN DEFAULT FALSE,
    allow_ai_invite BOOLEAN DEFAULT TRUE, -- allow AI to invite new characters

    -- story settings
    world_setting TEXT, -- world setting
    current_scene TEXT, -- current scene
    scene_style VARCHAR(50), -- scene style
    
    -- 统计信息
    member_count INTEGER DEFAULT 0,
    message_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建群聊成员表
CREATE TABLE IF NOT EXISTS group_chat_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- if real user

    -- member type
    member_type VARCHAR(20) NOT NULL CHECK (member_type IN ('user', 'character')),

    -- permission settings
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    can_invite BOOLEAN DEFAULT FALSE,
    can_kick BOOLEAN DEFAULT FALSE,

    -- join information
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    invited_by UUID, -- inviter ID
    
    UNIQUE(group_chat_id, character_id),
    UNIQUE(group_chat_id, user_id)
);

-- 创建聊天消息表
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID REFERENCES group_chats(id) ON DELETE CASCADE,
    
    -- sender information
    sender_character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    sender_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    sender_type VARCHAR(20) NOT NULL CHECK (sender_type IN ('user', 'character')),

    -- message content
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'audio', 'video', 'system')),

    -- media attachments
    media_urls TEXT[], -- media file URL array
    media_metadata JSONB, -- media metadata

    -- AI generation information
    model_used VARCHAR(100), -- AI model used
    generation_cost DECIMAL(10,6), -- generation cost
    tokens_used INTEGER, -- tokens used

    -- message status
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_characters_user_id ON characters(user_id);
CREATE INDEX IF NOT EXISTS idx_characters_visibility ON characters(visibility);
CREATE INDEX IF NOT EXISTS idx_characters_is_featured ON characters(is_featured);
CREATE INDEX IF NOT EXISTS idx_characters_created_at ON characters(created_at);

CREATE INDEX IF NOT EXISTS idx_character_tags_character_id ON character_tags(character_id);
CREATE INDEX IF NOT EXISTS idx_character_tags_tag ON character_tags(tag);

CREATE INDEX IF NOT EXISTS idx_character_relationships_character_id ON character_relationships(character_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_target_character_id ON character_relationships(target_character_id);

CREATE INDEX IF NOT EXISTS idx_group_chats_creator_user_id ON group_chats(creator_user_id);
CREATE INDEX IF NOT EXISTS idx_group_chats_is_public ON group_chats(is_public);

CREATE INDEX IF NOT EXISTS idx_group_chat_members_group_chat_id ON group_chat_members(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_character_id ON group_chat_members(character_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_user_id ON group_chat_members(user_id);

CREATE INDEX IF NOT EXISTS idx_chat_messages_group_chat_id ON chat_messages(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_sender_character_id ON chat_messages(sender_character_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at);

-- 更新时间触发器
CREATE TRIGGER update_characters_updated_at 
    BEFORE UPDATE ON characters 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_character_relationships_updated_at 
    BEFORE UPDATE ON character_relationships 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_group_chats_updated_at 
    BEFORE UPDATE ON group_chats 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chat_messages_updated_at 
    BEFORE UPDATE ON chat_messages 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
