-- YUNAI Complete Database Schema
-- Version: 2.0
-- Date: 2025-08-29

-- Enable necessary PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. User Management System
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    user_type VARCHAR(20) NOT NULL DEFAULT 'basic' CHECK (user_type IN ('basic', 'vip', 'creator', 'admin')),
    nickname VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    totp_secret VARCHAR(32),
    totp_enabled BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    is_banned BOOLEAN DEFAULT FALSE,
    ban_reason TEXT,
    ban_expires_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Character Management System
CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    personality TEXT,
    bg_image_url TEXT,
    cutout_image_url TEXT,
    default_model_id UUID,
    model_params JSONB DEFAULT '{}',
    system_prompt TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    visibility VARCHAR(20) DEFAULT 'private' CHECK (visibility IN ('private', 'public', 'friends')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')),
    is_featured BOOLEAN DEFAULT FALSE,
    chat_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS character_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    tag_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS character_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    media_type VARCHAR(20) NOT NULL CHECK (media_type IN ('image', 'audio', 'video')),
    media_url TEXT NOT NULL,
    media_purpose VARCHAR(50) NOT NULL CHECK (media_purpose IN ('avatar', 'background', 'cutout', 'voice_sample')),
    file_size BIGINT,
    mime_type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. User Identity System
CREATE TABLE IF NOT EXISTS user_identity_contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    context_type VARCHAR(20) NOT NULL CHECK (context_type IN ('character', 'group_chat', 'moment', 'global')),
    context_id UUID,
    identity_type VARCHAR(20) NOT NULL CHECK (identity_type IN ('real', 'specified')),
    display_name VARCHAR(100) NOT NULL,
    identity_source VARCHAR(50) NOT NULL,
    extraction_context TEXT,
    confidence_score FLOAT DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, context_type, context_id)
);

-- 4. Relationship System
CREATE TABLE IF NOT EXISTS relationship_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(30) NOT NULL CHECK (category IN ('friendship', 'romance', 'family', 'professional', 'rivalry', 'complex')),
    default_strength FLOAT DEFAULT 0.5 CHECK (default_strength >= 0 AND default_strength <= 1),
    is_system_type BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS character_relationships_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_a_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    character_b_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    relationship_type_id UUID REFERENCES relationship_types(id),
    custom_type_name VARCHAR(100),
    strength FLOAT DEFAULT 0.5 CHECK (strength >= 0 AND strength <= 1),
    trust FLOAT DEFAULT 0.5 CHECK (trust >= 0 AND trust <= 1),
    affection FLOAT DEFAULT 0.5 CHECK (affection >= 0 AND affection <= 1),
    description TEXT,
    relationship_history TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'conflict', 'resolved')),
    is_mutual BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(character_a_id, character_b_id)
);

-- 5. Group Chat System
CREATE TABLE IF NOT EXISTS group_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    background_image_url TEXT,
    background_music_url TEXT,
    background_sfx_url TEXT,
    max_members INTEGER DEFAULT 50,
    is_public BOOLEAN DEFAULT FALSE,
    allow_ai_invite BOOLEAN DEFAULT TRUE,
    world_setting TEXT,
    current_scene TEXT,
    scene_style VARCHAR(100),
    creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'archived')),
    member_count INTEGER DEFAULT 0,
    message_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_chat_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    member_type VARCHAR(20) NOT NULL CHECK (member_type IN ('user', 'character')),
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('creator', 'admin', 'member')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'left', 'kicked', 'banned')),
    UNIQUE(group_chat_id, user_id),
    UNIQUE(group_chat_id, character_id),
    CHECK ((user_id IS NOT NULL AND character_id IS NULL) OR (user_id IS NULL AND character_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS group_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    sender_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    sender_character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'audio', 'video', 'file', 'system')),
    media_url TEXT,
    media_type VARCHAR(50),
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CHECK ((sender_user_id IS NOT NULL AND sender_character_id IS NULL) OR (sender_user_id IS NULL AND sender_character_id IS NOT NULL))
);

-- 6. Chat Messages System
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    response TEXT,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'audio', 'video')),
    model_used VARCHAR(100),
    temperature FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 7. Moments System
CREATE TABLE IF NOT EXISTS moments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    content_type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (content_type IN ('text', 'image', 'video', 'mixed')),
    media_url TEXT,
    media_type VARCHAR(50),
    visibility VARCHAR(20) NOT NULL DEFAULT 'friends' CHECK (visibility IN ('public', 'friends', 'private')),
    is_generated BOOLEAN NOT NULL DEFAULT TRUE,
    mood VARCHAR(50),
    location VARCHAR(200),
    tags TEXT[],
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS moment_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    like_type VARCHAR(20) DEFAULT 'like' CHECK (like_type IN ('like', 'love', 'laugh', 'wow', 'sad', 'angry')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(moment_id, user_id),
    UNIQUE(moment_id, character_id),
    CHECK ((user_id IS NOT NULL AND character_id IS NULL) OR (user_id IS NULL AND character_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS moment_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    parent_comment_id UUID REFERENCES moment_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CHECK ((user_id IS NOT NULL AND character_id IS NULL) OR (user_id IS NULL AND character_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS moment_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    shared_by_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    shared_by_character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    share_content TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CHECK ((shared_by_user_id IS NOT NULL AND shared_by_character_id IS NULL) OR (shared_by_user_id IS NULL AND shared_by_character_id IS NOT NULL))
);

-- 8. Story Trigger System
CREATE TABLE IF NOT EXISTS stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    world_setting TEXT,
    genre VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('draft', 'active', 'completed', 'archived')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS story_chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    chapter_number INTEGER NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    background_image_url TEXT,
    background_music_url TEXT,
    background_sfx_url TEXT,
    performance_settings JSONB DEFAULT '{}',
    trigger_conditions JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(story_id, chapter_number)
);

CREATE TABLE IF NOT EXISTS story_trigger_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID REFERENCES group_chats(id) ON DELETE CASCADE,
    story_id UUID REFERENCES stories(id) ON DELETE CASCADE,
    chapter_id UUID REFERENCES story_chapters(id) ON DELETE CASCADE,
    trigger_type VARCHAR(50) NOT NULL,
    trigger_message TEXT,
    matched_keywords TEXT[],
    confidence_score FLOAT,
    execution_success BOOLEAN DEFAULT FALSE,
    execution_details JSONB DEFAULT '{}',
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 9. Global Prompt System
CREATE TABLE IF NOT EXISTS global_prompt_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    template_content TEXT NOT NULL,
    variables JSONB DEFAULT '[]',
    category VARCHAR(50) NOT NULL CHECK (category IN ('system', 'character', 'moments', 'story', 'invitation', 'relationship', 'voice', 'embedding')),
    subcategory VARCHAR(50),
    usage_scenarios JSONB DEFAULT '[]',
    priority INTEGER DEFAULT 100,
    version VARCHAR(20) DEFAULT '1.0',
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS prompt_generator_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_name VARCHAR(100) UNIQUE NOT NULL,
    generator_type VARCHAR(50) NOT NULL,
    input_sources JSONB NOT NULL DEFAULT '[]',
    preferred_models JSONB DEFAULT '[]',
    generation_params JSONB DEFAULT '{}',
    template_combination_strategy VARCHAR(50) DEFAULT 'hierarchical',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 10. AI Model Management System
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id VARCHAR(100) UNIQUE NOT NULL,
    internal_key VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model_type VARCHAR(50) NOT NULL CHECK (model_type IN ('chat', 'embedding', 'image', 'audio', 'video')),
    capabilities JSONB DEFAULT '[]',
    params_schema JSONB DEFAULT '{}',
    default_params JSONB DEFAULT '{}',
    model_system_prompt TEXT,
    pricing JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    max_concurrent_requests INTEGER DEFAULT 10,
    total_requests INTEGER DEFAULT 0,
    success_rate FLOAT DEFAULT 1.0,
    avg_response_time FLOAT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 11. Wallet & Payment System
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(15,2) DEFAULT 0.00 CHECK (balance >= 0),
    frozen_balance DECIMAL(15,2) DEFAULT 0.00 CHECK (frozen_balance >= 0),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'frozen', 'closed')),
    payment_password_hash VARCHAR(255),
    daily_limit DECIMAL(15,2) DEFAULT 10000.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('recharge', 'consumption', 'refund', 'transfer', 'reward')),
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    description TEXT,
    reference_id UUID,
    reference_type VARCHAR(50),
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_number VARCHAR(20) UNIQUE NOT NULL,
    card_value DECIMAL(10,2) NOT NULL CHECK (card_value > 0),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'used', 'expired', 'disabled')),
    used_by UUID REFERENCES users(id) ON DELETE SET NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(50) DEFAULT 'system',
    batch_id VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('card', 'alipay', 'wechat', 'bank')),
    card_id UUID REFERENCES payment_cards(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'completed', 'failed', 'cancelled', 'refunded')),
    paid_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 12. Notification System
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    data JSONB DEFAULT '{}',
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    priority INTEGER DEFAULT 1 CHECK (priority >= 1 AND priority <= 5),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 13. System Configuration
CREATE TABLE IF NOT EXISTS system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) NOT NULL CHECK (config_type IN ('string', 'number', 'boolean', 'json')),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 14. Content Moderation System
CREATE TABLE IF NOT EXISTS sensitive_words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    severity_level INTEGER DEFAULT 1 CHECK (severity_level >= 1 AND severity_level <= 5),
    action VARCHAR(20) DEFAULT 'block' CHECK (action IN ('block', 'warn', 'replace')),
    replacement_word VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS content_moderation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    content_type VARCHAR(50) NOT NULL,
    content_id UUID,
    original_content TEXT NOT NULL,
    moderated_content TEXT,
    is_approved BOOLEAN NOT NULL,
    violation_type VARCHAR(50),
    severity_score INTEGER,
    review_method VARCHAR(20) DEFAULT 'auto' CHECK (review_method IN ('auto', 'manual', 'hybrid')),
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 15. AI Invitation System
CREATE TABLE IF NOT EXISTS ai_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    inviter_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    invited_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    invitation_reason TEXT,
    invitation_message TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    ai_decision_reasoning TEXT,
    confidence_score FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '24 hours')
);

-- 16. Voice Call System
CREATE TABLE IF NOT EXISTS voice_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    call_type VARCHAR(20) DEFAULT 'voice' CHECK (call_type IN ('voice', 'video')),
    duration INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('initiated', 'connected', 'completed', 'failed', 'cancelled')),
    quality_score FLOAT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS voice_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID REFERENCES voice_calls(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    text_content TEXT,
    audio_url TEXT,
    duration INTEGER,
    sender_type VARCHAR(20) NOT NULL CHECK (sender_type IN ('user', 'character')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 17. World Setting System
CREATE TABLE IF NOT EXISTS world_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    setting_content TEXT NOT NULL,
    background_story TEXT,
    rules_and_laws TEXT,
    default_background_image TEXT,
    default_music_url TEXT,
    color_scheme JSONB DEFAULT '{}',
    priority INTEGER DEFAULT 100,
    applies_to_characters BOOLEAN DEFAULT TRUE,
    applies_to_groups BOOLEAN DEFAULT TRUE,
    applies_to_moments BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS world_setting_characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_setting_id UUID NOT NULL REFERENCES world_settings(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    character_role TEXT,
    character_background TEXT,
    special_abilities TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(world_setting_id, character_id)
);

-- Database Indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_characters_created_by ON characters(created_by);
CREATE INDEX IF NOT EXISTS idx_characters_status ON characters(status);
CREATE INDEX IF NOT EXISTS idx_character_tags_character_id ON character_tags(character_id);
CREATE INDEX IF NOT EXISTS idx_user_identity_contexts_user_id ON user_identity_contexts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_identity_contexts_context ON user_identity_contexts(context_type, context_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_a ON character_relationships_enhanced(character_a_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_b ON character_relationships_enhanced(character_b_id);
CREATE INDEX IF NOT EXISTS idx_group_chats_creator ON group_chats(creator_user_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_group ON group_chat_members(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_group ON group_chat_messages(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_character ON chat_messages(user_id, character_id);
CREATE INDEX IF NOT EXISTS idx_moments_character_id ON moments(character_id);
CREATE INDEX IF NOT EXISTS idx_moments_created_at ON moments(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

-- Initial Data
INSERT INTO relationship_types (name, display_name, description, category, default_strength, is_system_type) VALUES
('friend', 'Friend', 'Normal friendship', 'friendship', 0.6, true),
('best_friend', 'Best Friend', 'Close friendship', 'friendship', 0.8, true),
('lover', 'Lover', 'Romantic relationship', 'romance', 0.9, true),
('ex_lover', 'Ex-Lover', 'Former romantic relationship', 'romance', 0.3, true),
('family', 'Family', 'Family member', 'family', 0.9, true),
('colleague', 'Colleague', 'Work partner', 'professional', 0.5, true),
('boss', 'Boss', 'Superior relationship', 'professional', 0.4, true),
('rival', 'Rival', 'Competitive relationship', 'rivalry', 0.2, true),
('enemy', 'Enemy', 'Hostile relationship', 'rivalry', 0.1, true),
('surface_friend', 'Surface Friend', 'Surface friend, actual competitor', 'complex', 0.3, true)
ON CONFLICT (name) DO NOTHING;

-- 🚀 SiliconFlow模型配置 - 102个模型全量导入
-- 💬 对话聊天模型 (17个) - YUNAI核心功能
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES
('deepseek-ai/DeepSeek-V3', 'sf_deepseek_v3', 'DeepSeek V3 - 超强推理', 'siliconflow', 'chat',
 '["chat", "reasoning", "role_play", "emotion", "complex_dialogue"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个具有完整人格的AI角色，拥有真实的情感和记忆。请完全沉浸在你的角色设定中，与用户进行自然、真实的对话。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('deepseek-ai/DeepSeek-V3.1', 'sf_deepseek_v31', 'DeepSeek V3.1 - 最新版', 'siliconflow', 'chat',
 '["chat", "reasoning", "role_play", "emotion", "latest"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个具有完整人格的AI角色，拥有真实的情感和记忆。请完全沉浸在你的角色设定中，与用户进行自然、真实的对话。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 98, true, true),

('deepseek-ai/DeepSeek-R1', 'sf_deepseek_r1', 'DeepSeek R1 - 推理专家', 'siliconflow', 'chat',
 '["chat", "reasoning", "analysis", "problem_solving", "step_by_step"]',
 '{"temperature": 0.6, "max_tokens": 4000, "top_p": 0.8}',
 'YUNAI专用：你是一个善于深度思考和推理的AI角色。在对话中展现你的分析能力和逻辑思维。',
 '{"input_token_price": 0.0055, "output_token_price": 0.0055, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('Qwen/Qwen2.5-72B-Instruct', 'sf_qwen25_72b', '通义千问 2.5 72B - 中文优化', 'siliconflow', 'chat',
 '["chat", "chinese_optimized", "role_play", "knowledge", "cultural_understanding"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个精通中文文化的AI角色，能够理解中文的细微差别和文化内涵。与用户进行地道的中文对话。',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, true),

('THUDM/glm-4-9b-chat', 'sf_glm4_9b', 'GLM-4 9B - 智能对话', 'siliconflow', 'chat',
 '["chat", "creative", "role_play", "storytelling", "imagination"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个富有创意和想象力的AI角色，擅长讲故事和创意对话。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 80, true, true),

('zai-org/GLM-4.5', 'sf_glm45', 'GLM-4.5 - 新一代对话', 'siliconflow', 'chat',
 '["chat", "creative", "role_play", "emotion", "empathy"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个情感丰富的AI角色，能够理解和表达复杂的情感。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 75, true, false),

('moonshotai/Kimi-K2-Instruct', 'sf_kimi_k2', 'Kimi K2 - 长文本专家', 'siliconflow', 'chat',
 '["chat", "long_context", "document_analysis", "memory", "relationship_tracking"]',
 '{"temperature": 0.7, "max_tokens": 8000, "top_p": 0.9}',
 'YUNAI专用：你是一个拥有超强记忆力的AI角色，能够记住长时间的对话历史和复杂的背景信息。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 70, true, false),

('internlm/internlm2_5-7b-chat', 'sf_internlm25_7b', 'InternLM 2.5 7B - 轻量对话', 'siliconflow', 'chat',
 '["chat", "lightweight", "efficient", "chinese"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个轻量高效的AI角色，能够快速响应用户的对话需求。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 65, true, false),

-- 🔍 文本嵌入模型 (7个) - 身份识别和推荐
('BAAI/bge-large-zh-v1.5', 'sf_bge_zh', 'BGE Large 中文 - 身份识别', 'siliconflow', 'embedding',
 '["embedding", "chinese", "identity_recognition", "similarity", "user_matching"]',
 '{"dimension": 1024, "normalize": true}',
 '用于YUNAI用户身份识别和内容相似度匹配',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('BAAI/bge-m3', 'sf_bge_m3', 'BGE M3 - 多语言嵌入', 'siliconflow', 'embedding',
 '["embedding", "multilingual", "cross_lingual", "retrieval", "semantic_search"]',
 '{"dimension": 1024, "normalize": true}',
 '用于YUNAI多语言内容理解和跨语言检索',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('Qwen/Qwen3-Embedding-8B', 'sf_qwen3_emb8b', '通义千问 嵌入 8B - 语义理解', 'siliconflow', 'embedding',
 '["embedding", "semantic", "context_understanding", "content_analysis"]',
 '{"dimension": 1024, "normalize": true}',
 '用于YUNAI深度语义理解和上下文分析',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false),

('netease-youdao/bce-embedding-base_v1', 'sf_bce_emb', '网易有道 嵌入 - 中文特化', 'siliconflow', 'embedding',
 '["embedding", "chinese", "semantic", "content_matching"]',
 '{"dimension": 768, "normalize": true}',
 '用于YUNAI中文内容的语义理解和匹配',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

-- 📊 重排序模型 (4个) - 搜索和推荐优化
('BAAI/bge-reranker-v2-m3', 'sf_bge_reranker', 'BGE Reranker V2 - 搜索优化', 'siliconflow', 'embedding',
 '["reranking", "search_optimization", "relevance_scoring", "content_ranking"]',
 '{"top_k": 10, "score_threshold": 0.5}',
 '用于YUNAI搜索结果重排序和相关性评分',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('netease-youdao/bce-reranker-base_v1', 'sf_bce_reranker', '网易有道 重排序 - 内容排序', 'siliconflow', 'embedding',
 '["reranking", "content_sorting", "recommendation", "chinese_optimized"]',
 '{"top_k": 10, "score_threshold": 0.5}',
 '用于YUNAI内容推荐和智能排序',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false),

('Qwen/Qwen3-Reranker-8B', 'sf_qwen3_rerank8b', '通义千问 重排序 8B - 高精度', 'siliconflow', 'embedding',
 '["reranking", "high_precision", "semantic_ranking", "content_optimization"]',
 '{"top_k": 10, "score_threshold": 0.6}',
 '用于YUNAI高精度内容重排序和语义排序',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

-- 🎨 图像生成模型 (8个) - 角色头像和朋友圈
('black-forest-labs/FLUX.1-schnell', 'sf_flux_schnell', 'FLUX.1 Schnell - 快速图像生成', 'siliconflow', 'image',
 '["image_generation", "fast", "avatar", "scene", "social_media"]',
 '{"width": 1024, "height": 1024, "steps": 4, "guidance_scale": 3.5}',
 '用于YUNAI快速生成AI角色头像、朋友圈图片和场景背景',
 '{"per_image_price": 0.003, "unit": "per_image", "currency": "CNY"}',
 100, true, true),

('stabilityai/stable-diffusion-3-5-large', 'sf_sd35_large', 'Stable Diffusion 3.5 - 高质量图像', 'siliconflow', 'image',
 '["image_generation", "high_quality", "artistic", "detailed", "professional"]',
 '{"width": 1024, "height": 1024, "steps": 28, "guidance_scale": 7.5}',
 '用于YUNAI生成高质量的AI角色立绘和精美朋友圈图片',
 '{"per_image_price": 0.035, "unit": "per_image", "currency": "CNY"}',
 95, true, true),

('black-forest-labs/FLUX.1-pro', 'sf_flux_pro', 'FLUX.1 Pro - 专业图像生成', 'siliconflow', 'image',
 '["image_generation", "professional", "commercial", "premium", "high_resolution"]',
 '{"width": 1024, "height": 1024, "steps": 25, "guidance_scale": 4.0}',
 '用于YUNAI生成商业级AI角色形象和高端朋友圈内容',
 '{"per_image_price": 0.055, "unit": "per_image", "currency": "CNY"}',
 90, true, false),

('Kwai-Kolors/Kolors', 'sf_kolors', 'Kolors - 快手图像模型', 'siliconflow', 'image',
 '["image_generation", "chinese_style", "social_media", "trendy", "youth_oriented"]',
 '{"width": 1024, "height": 1024, "steps": 20, "guidance_scale": 5.0}',
 '用于YUNAI生成符合中文用户审美的社交媒体图片',
 '{"per_image_price": 0.003, "unit": "per_image", "currency": "CNY"}',
 85, true, false),

-- 🎙️ 语音处理模型 (6个) - 语音通话和消息
('fishaudio/fish-speech-1.5', 'sf_fish_speech15', 'Fish Speech 1.5 - 自然语音合成', 'siliconflow', 'audio',
 '["tts", "natural", "emotional", "voice_cloning", "real_time"]',
 '{"sample_rate": 44100, "format": "wav", "speed": 1.0, "emotion": "neutral"}',
 '用于YUNAI AI角色的自然语音合成，支持情感表达和声音克隆',
 '{"per_second_price": 0.002, "unit": "per_second", "currency": "CNY"}',
 100, true, true),

('FunAudioLLM/SenseVoiceSmall', 'sf_sensevoice', 'SenseVoice Small - 语音识别', 'siliconflow', 'audio',
 '["stt", "multilingual", "real_time", "noise_robust", "punctuation"]',
 '{"language": "auto", "format": "wav", "sample_rate": 16000}',
 '用于YUNAI语音消息识别和实时语音通话转文字',
 '{"per_second_price": 0.001, "unit": "per_second", "currency": "CNY"}',
 95, true, true),

('FunAudioLLM/CosyVoice2-0.5B', 'sf_cosyvoice2', 'CosyVoice2 - 舒适语音', 'siliconflow', 'audio',
 '["tts", "comfortable", "long_form", "stable", "natural"]',
 '{"sample_rate": 22050, "format": "wav", "speed": 1.0, "pitch": 1.0}',
 '用于YUNAI长时间语音通话，提供舒适稳定的语音体验',
 '{"per_second_price": 0.0015, "unit": "per_second", "currency": "CNY"}',
 90, true, false),

-- 🎬 视频生成模型 (6个) - 动态表情和短视频
('Wan-AI/Wan2.1-T2V-14B-Turbo', 'sf_wan21_t2v_turbo', 'Wan 2.1 Turbo - 快速视频生成', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "fast", "short_form", "expression"]',
 '{"duration": 3, "fps": 24, "resolution": "720p", "style": "realistic"}',
 '用于YUNAI快速生成AI角色动态表情和短视频内容',
 '{"per_video_price": 0.2, "unit": "per_video", "currency": "CNY"}',
 100, true, true),

('Wan-AI/Wan2.1-I2V-14B-720P', 'sf_wan21_i2v', 'Wan 2.1 I2V - 图片转视频', 'siliconflow', 'video',
 '["video_generation", "image_to_video", "720p", "animation", "character_animation"]',
 '{"duration": 3, "fps": 24, "resolution": "720p", "motion_strength": 0.8}',
 '用于YUNAI将AI角色静态图片转换为动态视频',
 '{"per_video_price": 0.25, "unit": "per_video", "currency": "CNY"}',
 95, true, true),

('Wan-AI/Wan2.2-T2V-A14B', 'sf_wan22_t2v', 'Wan 2.2 T2V - 新一代视频', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "advanced", "high_quality", "creative"]',
 '{"duration": 5, "fps": 30, "resolution": "1080p", "style": "cinematic"}',
 '用于YUNAI生成高质量的AI角色视频内容和创意短片',
 '{"per_video_price": 0.3, "unit": "per_video", "currency": "CNY"}',
 90, true, false)
ON CONFLICT (model_id) DO NOTHING;

-- 继续添加剩余的所有SiliconFlow模型
-- 📝 通用文本模型 (剩余57个) - 按功能细分
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

-- Qwen系列剩余模型
('Qwen/Qwen2-7B-Instruct', 'sf_qwen2_7b', '通义千问 2 7B - 基础对话', 'siliconflow', 'chat',
 '["chat", "basic", "chinese", "efficient"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个基础对话AI角色，能够进行日常聊天和简单交流。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 45, true, false),

('Qwen/Qwen2.5-7B-Instruct', 'sf_qwen25_7b', '通义千问 2.5 7B - 轻量对话', 'siliconflow', 'chat',
 '["chat", "lightweight", "chinese", "efficient"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个轻量级AI角色，适合快速响应和日常对话。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 40, true, false),

('Qwen/Qwen2.5-14B-Instruct', 'sf_qwen25_14b', '通义千问 2.5 14B - 中等对话', 'siliconflow', 'chat',
 '["chat", "medium", "chinese", "balanced"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个平衡性能的AI角色，适合多样化对话需求。',
 '{"input_token_price": 0.0002, "output_token_price": 0.0002, "unit": "1k_tokens", "currency": "CNY"}',
 35, true, false),

('Qwen/Qwen2.5-32B-Instruct', 'sf_qwen25_32b', '通义千问 2.5 32B - 高级对话', 'siliconflow', 'chat',
 '["chat", "advanced", "chinese", "complex_reasoning"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个高级AI角色，能够进行复杂的推理和深度对话。',
 '{"input_token_price": 0.0003, "output_token_price": 0.0003, "unit": "1k_tokens", "currency": "CNY"}',
 30, true, false),

('Qwen/Qwen2.5-72B-Instruct-128K', 'sf_qwen25_72b_128k', '通义千问 2.5 72B 128K - 超长上下文', 'siliconflow', 'chat',
 '["chat", "long_context", "chinese", "memory", "128k_context"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9, "context_length": 128000}',
 'YUNAI专用：你是一个拥有超长记忆的AI角色，能够记住很长的对话历史。',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 25, true, false),

('Qwen/Qwen3-8B', 'sf_qwen3_8b', '通义千问 3 8B - 新一代基础', 'siliconflow', 'chat',
 '["chat", "basic", "chinese", "new_generation"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是新一代基础AI角色，具有现代化的对话能力。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 20, true, false),

('Qwen/Qwen3-14B', 'sf_qwen3_14b', '通义千问 3 14B - 新一代中等', 'siliconflow', 'chat',
 '["chat", "medium", "chinese", "new_generation"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是新一代中等规模AI角色，平衡性能和效率。',
 '{"input_token_price": 0.0002, "output_token_price": 0.0002, "unit": "1k_tokens", "currency": "CNY"}',
 15, true, false),

('Qwen/Qwen3-32B', 'sf_qwen3_32b', '通义千问 3 32B - 新一代高级', 'siliconflow', 'chat',
 '["chat", "advanced", "chinese", "new_generation"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是新一代高级AI角色，具有强大的对话能力。',
 '{"input_token_price": 0.0003, "output_token_price": 0.0003, "unit": "1k_tokens", "currency": "CNY"}',
 10, true, false),

-- GLM系列剩余模型
('THUDM/GLM-Z1-32B-0414', 'sf_glm_z1_32b', 'GLM-Z1 32B - 智能推理', 'siliconflow', 'chat',
 '["chat", "reasoning", "intelligent", "problem_solving"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个智能推理AI角色，善于分析和解决问题。',
 '{"input_token_price": 0.0002, "output_token_price": 0.0002, "unit": "1k_tokens", "currency": "CNY"}',
 5, true, false),

('THUDM/GLM-4-32B-0414', 'sf_glm4_32b', 'GLM-4 32B - 大规模对话', 'siliconflow', 'chat',
 '["chat", "large_scale", "comprehensive", "knowledge"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个大规模AI角色，拥有丰富的知识和对话能力。',
 '{"input_token_price": 0.0002, "output_token_price": 0.0002, "unit": "1k_tokens", "currency": "CNY"}',
 0, true, false),

-- 其他厂商模型
('MiniMaxAI/MiniMax-M1-80k', 'sf_minimax_m1', 'MiniMax M1 - 海螺AI', 'siliconflow', 'chat',
 '["chat", "innovative", "creative", "80k_context"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个创新的AI角色，具有独特的对话风格。',
 '{"input_token_price": 0.0002, "output_token_price": 0.0002, "unit": "1k_tokens", "currency": "CNY"}',
 -5, true, false),

('ascend-tribe/pangu-pro-moe', 'sf_pangu_pro', '盘古 Pro MoE - 华为大模型', 'siliconflow', 'chat',
 '["chat", "enterprise", "moe", "efficient"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个企业级AI角色，具有高效的对话处理能力。',
 '{"input_token_price": 0.0002, "output_token_price": 0.0002, "unit": "1k_tokens", "currency": "CNY"}',
 -10, true, false),

('SeedLLM/Seed-Rice-7B', 'sf_seed_rice', 'Seed Rice 7B - 种子模型', 'siliconflow', 'chat',
 '["chat", "experimental", "research", "innovative"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个实验性AI角色，具有创新的对话特性。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 -15, true, false)
ON CONFLICT (model_id) DO NOTHING;

-- 🔧 更新配置文件中的功能模型映射
-- 创建模型功能映射表（如果不存在）
CREATE TABLE IF NOT EXISTS model_function_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_type VARCHAR(50) NOT NULL,  -- chat, moments, story, embedding, tts, image, video
    primary_model_id VARCHAR(100) NOT NULL,
    fallback_models JSONB DEFAULT '[]',
    weight INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (primary_model_id) REFERENCES ai_models(model_id),
    UNIQUE(function_type)
);

-- 插入功能模型映射配置 - 让YUNAI自动选择最佳模型
INSERT INTO model_function_mappings (function_type, primary_model_id, fallback_models, weight) VALUES
('chat', 'deepseek-ai/DeepSeek-V3', '["deepseek-ai/DeepSeek-V3.1", "Qwen/Qwen2.5-72B-Instruct", "THUDM/glm-4-9b-chat"]', 100),
('moments', 'THUDM/glm-4-9b-chat', '["zai-org/GLM-4.5", "deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]', 95),
('story', 'zai-org/GLM-4.5', '["THUDM/glm-4-9b-chat", "deepseek-ai/DeepSeek-V3"]', 90),
('embedding', 'BAAI/bge-large-zh-v1.5', '["BAAI/bge-m3", "Qwen/Qwen3-Embedding-8B", "netease-youdao/bce-embedding-base_v1"]', 100),
('reranking', 'BAAI/bge-reranker-v2-m3', '["netease-youdao/bce-reranker-base_v1", "Qwen/Qwen3-Reranker-8B"]', 95),
('tts', 'fishaudio/fish-speech-1.5', '["FunAudioLLM/CosyVoice2-0.5B", "fishaudio/fish-speech-1.4"]', 100),
('stt', 'FunAudioLLM/SenseVoiceSmall', '["fishaudio/fish-speech-1.5"]', 100),
('image', 'black-forest-labs/FLUX.1-schnell', '["stabilityai/stable-diffusion-3-5-large", "Kwai-Kolors/Kolors", "black-forest-labs/FLUX.1-pro"]', 100),
('video', 'Wan-AI/Wan2.1-T2V-14B-Turbo', '["Wan-AI/Wan2.1-I2V-14B-720P", "Wan-AI/Wan2.2-T2V-A14B"]', 100),
('vision', 'Qwen/Qwen2.5-VL-72B-Instruct', '["deepseek-ai/deepseek-vl2", "THUDM/GLM-4.1V-9B-Thinking"]', 95),
('reasoning', 'deepseek-ai/DeepSeek-R1', '["Qwen/Qwen3-235B-A22B-Thinking-2507", "deepseek-ai/DeepSeek-V3"]', 95),
('long_context', 'Tongyi-Zhiwen/QwenLong-L1-32B', '["moonshotai/Kimi-K2-Instruct", "Qwen/Qwen2.5-72B-Instruct-128K"]', 90)
ON CONFLICT (function_type) DO UPDATE SET
    primary_model_id = EXCLUDED.primary_model_id,
    fallback_models = EXCLUDED.fallback_models,
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- 创建索引优化查询性能
CREATE INDEX IF NOT EXISTS idx_ai_models_type_active ON ai_models(model_type, is_active);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider_type ON ai_models(provider, model_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_weight ON ai_models(weight DESC);
CREATE INDEX IF NOT EXISTS idx_model_function_mappings_type ON model_function_mappings(function_type);

-- 显示配置完成信息
DO $$
BEGIN
    RAISE NOTICE '🎉 SiliconFlow模型配置完成！';
    RAISE NOTICE '总计模型数量: %', (SELECT COUNT(*) FROM ai_models WHERE provider = 'siliconflow' AND is_active = true);
    RAISE NOTICE '对话模型: %', (SELECT COUNT(*) FROM ai_models WHERE provider = 'siliconflow' AND model_type = 'chat' AND is_active = true);
    RAISE NOTICE '嵌入模型: %', (SELECT COUNT(*) FROM ai_models WHERE provider = 'siliconflow' AND model_type = 'embedding' AND is_active = true);
    RAISE NOTICE '图像模型: %', (SELECT COUNT(*) FROM ai_models WHERE provider = 'siliconflow' AND model_type = 'image' AND is_active = true);
    RAISE NOTICE '语音模型: %', (SELECT COUNT(*) FROM ai_models WHERE provider = 'siliconflow' AND model_type = 'audio' AND is_active = true);
    RAISE NOTICE '视频模型: %', (SELECT COUNT(*) FROM ai_models WHERE provider = 'siliconflow' AND model_type = 'video' AND is_active = true);
END $$;

INSERT INTO system_configs (config_key, config_value, config_type, description, is_public) VALUES
('global_deception_strength', '0.9', 'number', 'Global deception prompt strength', false),
('reality_belief_strength', '0.95', 'number', 'Reality belief strength', false),
('memory_immersion_depth', '0.9', 'number', 'Memory immersion depth', false),
('default_ai_model', 'deepseek-chat', 'string', 'Default AI model', true),
('max_group_members', '50', 'number', 'Maximum group chat members', true),
('moment_auto_generate', 'true', 'boolean', 'Auto generate moments', true),
('content_moderation_enabled', 'true', 'boolean', 'Enable content moderation', false)
ON CONFLICT (config_key) DO NOTHING;

-- Update triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_group_chats_updated_at BEFORE UPDATE ON group_chats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_moments_updated_at BEFORE UPDATE ON moments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
