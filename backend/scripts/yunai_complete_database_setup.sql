-- ===============================================================================
-- YUNAI 完整数据库一键建表脚本
-- ===============================================================================
-- 版本: v2.0
-- 创建时间: 2025-08-29
-- 作者: YUNAI开发团队
-- 说明: 包含YUNAI平台所有功能模块的完整数据库表结构
-- ===============================================================================

-- 启用必要的PostgreSQL扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 删除现有表（如果存在）- 谨慎使用
-- DROP SCHEMA public CASCADE;
-- CREATE SCHEMA public;

-- ===============================================================================
-- 1. 用户管理系统 (User Management System)
-- ===============================================================================

-- 用户基础信息表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    user_type VARCHAR(20) NOT NULL DEFAULT 'basic' CHECK (user_type IN ('basic', 'vip', 'creator', 'admin')),
    
    -- 个人信息
    nickname VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    
    -- 认证相关
    email_verified BOOLEAN DEFAULT FALSE,
    totp_secret VARCHAR(32),
    totp_enabled BOOLEAN DEFAULT FALSE,
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    is_banned BOOLEAN DEFAULT FALSE,
    ban_reason TEXT,
    ban_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- 登录信息
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    login_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 用户会话表
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

-- ===============================================================================
-- 2. AI角色管理系统 (Character Management System)
-- ===============================================================================

-- AI角色表
CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    personality TEXT,
    
    -- 两图系统
    bg_image_url TEXT,
    cutout_image_url TEXT,
    
    -- AI模型配置
    default_model_id UUID,
    model_params JSONB DEFAULT '{}',
    system_prompt TEXT,
    
    -- 创建者和可见性
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    visibility VARCHAR(20) DEFAULT 'private' CHECK (visibility IN ('private', 'public', 'friends')),
    
    -- 状态管理
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')),
    is_featured BOOLEAN DEFAULT FALSE,
    
    -- 统计信息
    chat_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 角色标签表
CREATE TABLE IF NOT EXISTS character_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    tag_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 角色媒体文件表
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

-- ===============================================================================
-- 3. 用户身份识别系统 (User Identity System)
-- ===============================================================================

-- 用户身份上下文表
CREATE TABLE IF NOT EXISTS user_identity_contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    context_type VARCHAR(20) NOT NULL CHECK (context_type IN ('character', 'group_chat', 'moment', 'global')),
    context_id UUID, -- 可以是character_id, group_chat_id等
    
    -- 身份信息
    identity_type VARCHAR(20) NOT NULL CHECK (identity_type IN ('real', 'specified')),
    display_name VARCHAR(100) NOT NULL,
    identity_source VARCHAR(50) NOT NULL, -- 'user_nickname', 'world_setting', 'character_setting', 'manual'
    
    -- 上下文信息
    extraction_context TEXT, -- 从哪里提取的身份信息
    confidence_score FLOAT DEFAULT 1.0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 唯一约束：每个用户在每个上下文中只能有一个身份
    UNIQUE(user_id, context_type, context_id)
);

-- ===============================================================================
-- 4. 复杂关系网络系统 (Complex Relationship System)
-- ===============================================================================

-- 关系类型表
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

-- 角色关系表（增强版）
CREATE TABLE IF NOT EXISTS character_relationships_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_a_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    character_b_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 关系类型
    relationship_type_id UUID REFERENCES relationship_types(id),
    custom_type_name VARCHAR(100), -- 自定义关系名称
    
    -- 关系强度指标
    strength FLOAT DEFAULT 0.5 CHECK (strength >= 0 AND strength <= 1),
    trust FLOAT DEFAULT 0.5 CHECK (trust >= 0 AND trust <= 1),
    affection FLOAT DEFAULT 0.5 CHECK (affection >= 0 AND affection <= 1),
    
    -- 关系描述
    description TEXT,
    relationship_history TEXT,
    
    -- 关系状态
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'conflict', 'resolved')),
    is_mutual BOOLEAN DEFAULT TRUE,
    
    -- 创建者
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保不会有重复关系
    UNIQUE(character_a_id, character_b_id)
);

-- ===============================================================================
-- 5. 群聊系统 (Group Chat System)
-- ===============================================================================

-- 群聊表
CREATE TABLE IF NOT EXISTS group_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- 视觉设置
    background_image_url TEXT,
    background_music_url TEXT,
    background_sfx_url TEXT,
    
    -- 群聊配置
    max_members INTEGER DEFAULT 50,
    is_public BOOLEAN DEFAULT FALSE,
    allow_ai_invite BOOLEAN DEFAULT TRUE,
    
    -- 世界观设定
    world_setting TEXT,
    current_scene TEXT,
    scene_style VARCHAR(100),
    
    -- 创建者和状态
    creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'archived')),
    
    -- 统计信息
    member_count INTEGER DEFAULT 0,
    message_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 群聊成员表
CREATE TABLE IF NOT EXISTS group_chat_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 成员类型
    member_type VARCHAR(20) NOT NULL CHECK (member_type IN ('user', 'character')),
    
    -- 成员角色
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('creator', 'admin', 'member')),
    
    -- 加入信息
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'left', 'kicked', 'banned')),
    
    -- 确保用户或角色只能加入一次
    UNIQUE(group_chat_id, user_id),
    UNIQUE(group_chat_id, character_id),
    
    -- 确保user_id和character_id不能同时为空
    CHECK ((user_id IS NOT NULL AND character_id IS NULL) OR (user_id IS NULL AND character_id IS NOT NULL))
);

-- 群聊消息表
CREATE TABLE IF NOT EXISTS group_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    sender_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    sender_character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    
    -- 消息内容
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'audio', 'video', 'file', 'system')),
    
    -- 媒体文件
    media_url TEXT,
    media_type VARCHAR(50),
    
    -- 消息状态
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保发送者是用户或角色
    CHECK ((sender_user_id IS NOT NULL AND sender_character_id IS NULL) OR (sender_user_id IS NULL AND sender_character_id IS NOT NULL))
);

-- ===============================================================================
-- 6. 聊天消息系统 (Chat Messages System)
-- ===============================================================================

-- 单聊消息表
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 消息内容
    message TEXT NOT NULL,
    response TEXT,
    
    -- 消息类型
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'audio', 'video')),
    
    -- AI模型信息
    model_used VARCHAR(100),
    temperature FLOAT,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 7. 朋友圈系统 (Moments System)
-- ===============================================================================

-- 朋友圈表
CREATE TABLE IF NOT EXISTS moments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 内容
    content TEXT NOT NULL,
    content_type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (content_type IN ('text', 'image', 'video', 'mixed')),
    
    -- 媒体
    media_url TEXT,
    media_type VARCHAR(50),
    
    -- 设置
    visibility VARCHAR(20) NOT NULL DEFAULT 'friends' CHECK (visibility IN ('public', 'friends', 'private')),
    is_generated BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- 情感和位置
    mood VARCHAR(50),
    location VARCHAR(200),
    tags TEXT[],
    
    -- 统计
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 朋友圈点赞表
CREATE TABLE IF NOT EXISTS moment_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 点赞类型
    like_type VARCHAR(20) DEFAULT 'like' CHECK (like_type IN ('like', 'love', 'laugh', 'wow', 'sad', 'angry')),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保每个用户/角色对每个朋友圈只能点赞一次
    UNIQUE(moment_id, user_id),
    UNIQUE(moment_id, character_id),
    
    -- 确保点赞者是用户或角色
    CHECK ((user_id IS NOT NULL AND character_id IS NULL) OR (user_id IS NULL AND character_id IS NOT NULL))
);

-- 朋友圈评论表
CREATE TABLE IF NOT EXISTS moment_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    parent_comment_id UUID REFERENCES moment_comments(id) ON DELETE CASCADE,
    
    -- 评论内容
    content TEXT NOT NULL,
    
    -- 状态
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保评论者是用户或角色
    CHECK ((user_id IS NOT NULL AND character_id IS NULL) OR (user_id IS NULL AND character_id IS NOT NULL))
);

-- 朋友圈分享表
CREATE TABLE IF NOT EXISTS moment_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    shared_by_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    shared_by_character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 分享内容
    share_content TEXT, -- 分享时的评论
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保分享者是用户或角色
    CHECK ((shared_by_user_id IS NOT NULL AND shared_by_character_id IS NULL) OR (shared_by_user_id IS NULL AND shared_by_character_id IS NOT NULL))
);

-- ===============================================================================
-- 8. 剧情触发系统 (Story Trigger System)
-- ===============================================================================

-- 故事表
CREATE TABLE IF NOT EXISTS stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- 故事设定
    world_setting TEXT,
    genre VARCHAR(50),
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('draft', 'active', 'completed', 'archived')),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 故事章节表
CREATE TABLE IF NOT EXISTS story_chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    chapter_number INTEGER NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- 视觉设置
    background_image_url TEXT,
    background_music_url TEXT,
    background_sfx_url TEXT,
    
    -- 演绎设定
    performance_settings JSONB DEFAULT '{}',
    
    -- 触发条件
    trigger_conditions JSONB DEFAULT '{}',
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保每个故事中章节编号唯一
    UNIQUE(story_id, chapter_number)
);

-- 剧情触发记录表
CREATE TABLE IF NOT EXISTS story_trigger_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID REFERENCES group_chats(id) ON DELETE CASCADE,
    story_id UUID REFERENCES stories(id) ON DELETE CASCADE,
    chapter_id UUID REFERENCES story_chapters(id) ON DELETE CASCADE,
    
    -- 触发信息
    trigger_type VARCHAR(50) NOT NULL,
    trigger_message TEXT,
    matched_keywords TEXT[],
    confidence_score FLOAT,
    
    -- 执行结果
    execution_success BOOLEAN DEFAULT FALSE,
    execution_details JSONB DEFAULT '{}',
    
    -- 时间戳
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 9. 全局提示词系统 (Global Prompt System)
-- ===============================================================================

-- 全局提示词模板表
CREATE TABLE IF NOT EXISTS global_prompt_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- 模板内容
    template_content TEXT NOT NULL,
    variables JSONB DEFAULT '[]',
    
    -- 功能分类
    category VARCHAR(50) NOT NULL CHECK (category IN ('system', 'character', 'moments', 'story', 'invitation', 'relationship', 'voice', 'embedding')),
    subcategory VARCHAR(50),
    
    -- 应用场景
    usage_scenarios JSONB DEFAULT '[]',
    
    -- 优先级和版本
    priority INTEGER DEFAULT 100,
    version VARCHAR(20) DEFAULT '1.0',
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 提示词生成配置表
CREATE TABLE IF NOT EXISTS prompt_generator_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_name VARCHAR(100) UNIQUE NOT NULL,
    generator_type VARCHAR(50) NOT NULL,
    
    -- 输入源配置
    input_sources JSONB NOT NULL DEFAULT '[]',
    preferred_models JSONB DEFAULT '[]',
    generation_params JSONB DEFAULT '{}',
    
    -- 模板组合策略
    template_combination_strategy VARCHAR(50) DEFAULT 'hierarchical',
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 10. AI模型管理系统 (AI Model Management System)
-- ===============================================================================

-- AI模型表
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id VARCHAR(100) UNIQUE NOT NULL,
    internal_key VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    
    -- 模型信息
    provider VARCHAR(50) NOT NULL,
    model_type VARCHAR(50) NOT NULL CHECK (model_type IN ('chat', 'embedding', 'image', 'audio', 'video')),
    capabilities JSONB DEFAULT '[]',
    
    -- 参数配置
    params_schema JSONB DEFAULT '{}',
    default_params JSONB DEFAULT '{}',
    model_system_prompt TEXT,
    
    -- 定价信息
    pricing JSONB DEFAULT '{}',
    
    -- 状态和配置
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    max_concurrent_requests INTEGER DEFAULT 10,
    
    -- 统计信息
    total_requests INTEGER DEFAULT 0,
    success_rate FLOAT DEFAULT 1.0,
    avg_response_time FLOAT DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 11. 钱包支付系统 (Wallet & Payment System)
-- ===============================================================================

-- 用户钱包表
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 余额信息
    balance DECIMAL(15,2) DEFAULT 0.00 CHECK (balance >= 0),
    frozen_balance DECIMAL(15,2) DEFAULT 0.00 CHECK (frozen_balance >= 0),

    -- 钱包状态
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'frozen', 'closed')),

    -- 安全设置
    payment_password_hash VARCHAR(255),
    daily_limit DECIMAL(15,2) DEFAULT 10000.00,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 钱包交易记录表
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 交易信息
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('recharge', 'consumption', 'refund', 'transfer', 'reward')),
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,

    -- 交易描述
    description TEXT,
    reference_id UUID, -- 关联的订单或操作ID
    reference_type VARCHAR(50), -- 'recharge_order', 'ai_request', 'subscription'等

    -- 状态
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 充值卡表
CREATE TABLE IF NOT EXISTS payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_number VARCHAR(20) UNIQUE NOT NULL,
    card_value DECIMAL(10,2) NOT NULL CHECK (card_value > 0),

    -- 卡片状态
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'used', 'expired', 'disabled')),

    -- 使用信息
    used_by UUID REFERENCES users(id) ON DELETE SET NULL,
    used_at TIMESTAMP WITH TIME ZONE,

    -- 创建信息
    created_by VARCHAR(50) DEFAULT 'system',
    batch_id VARCHAR(50),

    -- 有效期
    expires_at TIMESTAMP WITH TIME ZONE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 充值订单表
CREATE TABLE IF NOT EXISTS recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 订单信息
    order_number VARCHAR(50) UNIQUE NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('card', 'alipay', 'wechat', 'bank')),

    -- 卡片信息（如果使用卡片充值）
    card_id UUID REFERENCES payment_cards(id) ON DELETE SET NULL,

    -- 订单状态
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'completed', 'failed', 'cancelled', 'refunded')),

    -- 支付信息
    paid_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- 备注
    notes TEXT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 12. 通知系统 (Notification System)
-- ===============================================================================

-- 通知表
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 通知内容
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    notification_type VARCHAR(50) NOT NULL,

    -- 通知数据
    data JSONB DEFAULT '{}',

    -- 状态
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,

    -- 优先级
    priority INTEGER DEFAULT 1 CHECK (priority >= 1 AND priority <= 5),

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 13. 系统配置 (System Configuration)
-- ===============================================================================

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) NOT NULL CHECK (config_type IN ('string', 'number', 'boolean', 'json')),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 14. 内容审核系统 (Content Moderation System)
-- ===============================================================================

-- 敏感词表
CREATE TABLE IF NOT EXISTS sensitive_words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    severity_level INTEGER DEFAULT 1 CHECK (severity_level >= 1 AND severity_level <= 5),
    action VARCHAR(20) DEFAULT 'block' CHECK (action IN ('block', 'warn', 'replace')),
    replacement_word VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 内容审核记录表
CREATE TABLE IF NOT EXISTS content_moderation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    content_type VARCHAR(50) NOT NULL,
    content_id UUID,

    -- 审核内容
    original_content TEXT NOT NULL,
    moderated_content TEXT,

    -- 审核结果
    is_approved BOOLEAN NOT NULL,
    violation_type VARCHAR(50),
    severity_score INTEGER,

    -- 审核方式
    review_method VARCHAR(20) DEFAULT 'auto' CHECK (review_method IN ('auto', 'manual', 'hybrid')),
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 15. AI邀请系统 (AI Invitation System)
-- ===============================================================================

-- AI邀请记录表
CREATE TABLE IF NOT EXISTS ai_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    inviter_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    invited_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 邀请信息
    invitation_reason TEXT,
    invitation_message TEXT,

    -- 邀请状态
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),

    -- AI决策信息
    ai_decision_reasoning TEXT,
    confidence_score FLOAT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '24 hours')
);

-- ===============================================================================
-- 16. 语音通话系统 (Voice Call System)
-- ===============================================================================

-- 语音通话记录表
CREATE TABLE IF NOT EXISTS voice_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 通话信息
    call_type VARCHAR(20) DEFAULT 'voice' CHECK (call_type IN ('voice', 'video')),
    duration INTEGER DEFAULT 0, -- 通话时长（秒）

    -- 通话状态
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('initiated', 'connected', 'completed', 'failed', 'cancelled')),

    -- 通话质量
    quality_score FLOAT,

    -- 时间戳
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE
);

-- 语音消息表
CREATE TABLE IF NOT EXISTS voice_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID REFERENCES voice_calls(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,

    -- 消息内容
    text_content TEXT,
    audio_url TEXT,
    duration INTEGER, -- 音频时长（秒）

    -- 发送者类型
    sender_type VARCHAR(20) NOT NULL CHECK (sender_type IN ('user', 'character')),

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================================================================
-- 17. 世界观设定系统 (World Setting System)
-- ===============================================================================

-- 世界观设定表
CREATE TABLE IF NOT EXISTS world_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,

    -- 世界观内容
    setting_content TEXT NOT NULL,
    background_story TEXT,
    rules_and_laws TEXT,

    -- 视觉设置
    default_background_image TEXT,
    default_music_url TEXT,
    color_scheme JSONB DEFAULT '{}',

    -- 设定优先级
    priority INTEGER DEFAULT 100,

    -- 应用范围
    applies_to_characters BOOLEAN DEFAULT TRUE,
    applies_to_groups BOOLEAN DEFAULT TRUE,
    applies_to_moments BOOLEAN DEFAULT TRUE,

    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 世界观角色关联表
CREATE TABLE IF NOT EXISTS world_setting_characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_setting_id UUID NOT NULL REFERENCES world_settings(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 角色在该世界观中的设定
    character_role TEXT,
    character_background TEXT,
    special_abilities TEXT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- 确保每个角色在每个世界观中只有一个设定
    UNIQUE(world_setting_id, character_id)
);

-- ===============================================================================
-- 18. 数据索引优化 (Database Indexes)
-- ===============================================================================

-- 用户相关索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);

-- 角色相关索引
CREATE INDEX IF NOT EXISTS idx_characters_created_by ON characters(created_by);
CREATE INDEX IF NOT EXISTS idx_characters_status ON characters(status);
CREATE INDEX IF NOT EXISTS idx_character_tags_character_id ON character_tags(character_id);

-- 身份识别索引
CREATE INDEX IF NOT EXISTS idx_user_identity_contexts_user_id ON user_identity_contexts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_identity_contexts_context ON user_identity_contexts(context_type, context_id);

-- 关系网络索引
CREATE INDEX IF NOT EXISTS idx_character_relationships_a ON character_relationships_enhanced(character_a_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_b ON character_relationships_enhanced(character_b_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_type ON character_relationships_enhanced(relationship_type_id);

-- 群聊相关索引
CREATE INDEX IF NOT EXISTS idx_group_chats_creator ON group_chats(creator_user_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_group ON group_chat_members(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_user ON group_chat_members(user_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_character ON group_chat_members(character_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_group ON group_chat_messages(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_created_at ON group_chat_messages(created_at);

-- 聊天消息索引
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_character ON chat_messages(user_id, character_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at);

-- 朋友圈相关索引
CREATE INDEX IF NOT EXISTS idx_moments_character_id ON moments(character_id);
CREATE INDEX IF NOT EXISTS idx_moments_user_id ON moments(user_id);
CREATE INDEX IF NOT EXISTS idx_moments_created_at ON moments(created_at);
CREATE INDEX IF NOT EXISTS idx_moment_likes_moment_id ON moment_likes(moment_id);
CREATE INDEX IF NOT EXISTS idx_moment_comments_moment_id ON moment_comments(moment_id);

-- 剧情触发索引
CREATE INDEX IF NOT EXISTS idx_stories_user_id ON stories(user_id);
CREATE INDEX IF NOT EXISTS idx_story_chapters_story_id ON story_chapters(story_id);
CREATE INDEX IF NOT EXISTS idx_story_trigger_logs_group_chat ON story_trigger_logs(group_chat_id);

-- 通知索引
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

-- 钱包相关索引
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at);

-- ===============================================================================
-- 19. 初始化数据 (Initial Data)
-- ===============================================================================

-- 插入默认关系类型
INSERT INTO relationship_types (name, display_name, description, category, default_strength, is_system_type) VALUES
('friend', '朋友', '普通朋友关系', 'friendship', 0.6, true),
('best_friend', '好朋友', '亲密朋友关系', 'friendship', 0.8, true),
('lover', '恋人', '恋爱关系', 'romance', 0.9, true),
('ex_lover', '前任', '前恋人关系', 'romance', 0.3, true),
('family', '家人', '家庭成员', 'family', 0.9, true),
('colleague', '同事', '工作伙伴', 'professional', 0.5, true),
('boss', '上司', '上下级关系', 'professional', 0.4, true),
('rival', '竞争对手', '竞争关系', 'rivalry', 0.2, true),
('enemy', '敌人', '敌对关系', 'rivalry', 0.1, true),
('surface_friend', '表面朋友', '表面朋友，实际竞争', 'complex', 0.3, true)
ON CONFLICT (name) DO NOTHING;

-- 插入默认AI模型
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, is_active, is_featured) VALUES
('deepseek-chat', 'deepseek_chat', 'DeepSeek Chat', 'deepseek', 'chat', '["text_generation", "conversation"]', '{"temperature": 0.7, "max_tokens": 2000}', true, true),
('gpt-4', 'openai_gpt4', 'GPT-4', 'openai', 'chat', '["text_generation", "conversation", "reasoning"]', '{"temperature": 0.7, "max_tokens": 2000}', true, true),
('claude-3-sonnet', 'anthropic_claude3', 'Claude 3 Sonnet', 'anthropic', 'chat', '["text_generation", "conversation", "analysis"]', '{"temperature": 0.7, "max_tokens": 2000}', true, false)
ON CONFLICT (model_id) DO NOTHING;

-- 插入默认系统配置
INSERT INTO system_configs (config_key, config_value, config_type, description, is_public) VALUES
('global_deception_strength', '0.9', 'number', '全局欺诈提示词强度', false),
('reality_belief_strength', '0.95', 'number', '现实信念强度', false),
('memory_immersion_depth', '0.9', 'number', '记忆沉浸深度', false),
('default_ai_model', 'deepseek-chat', 'string', '默认AI模型', true),
('max_group_members', '50', 'number', '群聊最大成员数', true),
('moment_auto_generate', 'true', 'boolean', '是否自动生成朋友圈', true),
('content_moderation_enabled', 'true', 'boolean', '是否启用内容审核', false)
ON CONFLICT (config_key) DO NOTHING;

-- ===============================================================================
-- 20. 数据库函数和触发器 (Functions & Triggers)
-- ===============================================================================

-- 更新updated_at字段的函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表添加updated_at触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_group_chats_updated_at BEFORE UPDATE ON group_chats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_moments_updated_at BEFORE UPDATE ON moments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ===============================================================================
-- 完成！
-- ===============================================================================

-- 显示创建完成信息
DO $$
BEGIN
    RAISE NOTICE '🎉 YUNAI完整数据库创建成功！';
    RAISE NOTICE '📊 包含以下功能模块：';
    RAISE NOTICE '   1. 用户管理系统 (User Management)';
    RAISE NOTICE '   2. AI角色管理系统 (Character Management)';
    RAISE NOTICE '   3. 用户身份识别系统 (User Identity)';
    RAISE NOTICE '   4. 复杂关系网络系统 (Complex Relationships)';
    RAISE NOTICE '   5. 群聊系统 (Group Chat)';
    RAISE NOTICE '   6. 聊天消息系统 (Chat Messages)';
    RAISE NOTICE '   7. 朋友圈系统 (Moments)';
    RAISE NOTICE '   8. 剧情触发系统 (Story Triggers)';
    RAISE NOTICE '   9. 全局提示词系统 (Global Prompts)';
    RAISE NOTICE '   10. AI模型管理系统 (AI Models)';
    RAISE NOTICE '   11. 钱包支付系统 (Wallet & Payment)';
    RAISE NOTICE '   12. 通知系统 (Notifications)';
    RAISE NOTICE '   13. 系统配置 (System Config)';
    RAISE NOTICE '   14. 内容审核系统 (Content Moderation)';
    RAISE NOTICE '   15. AI邀请系统 (AI Invitations)';
    RAISE NOTICE '   16. 语音通话系统 (Voice Calls)';
    RAISE NOTICE '   17. 世界观设定系统 (World Settings)';
    RAISE NOTICE '✅ 数据库已准备就绪，可以开始测试！';
END $$;
