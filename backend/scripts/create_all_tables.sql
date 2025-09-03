-- YUNAI 完整数据库表结构
-- 根据现有代码功能自动生成所有表结构
-- 版本: v1.0
-- 创建时间: 2025-08-27

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- 1. 用户管理表
-- =============================================================================

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

-- =============================================================================
-- 2. 钱包和支付系统
-- =============================================================================

-- 用户钱包表
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    currency VARCHAR(10) NOT NULL DEFAULT '金币',
    exchange_rate INTEGER NOT NULL DEFAULT 10, -- 1元 = 10金币
    
    -- 限额管理
    daily_limit DECIMAL(15,2) DEFAULT 10000.00,
    monthly_limit DECIMAL(15,2) DEFAULT 100000.00,
    daily_spent DECIMAL(15,2) DEFAULT 0.00,
    monthly_spent DECIMAL(15,2) DEFAULT 0.00,
    last_reset_date DATE DEFAULT CURRENT_DATE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id)
);

-- 钱包交易记录表
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 交易信息
    transaction_type VARCHAR(20) NOT NULL, -- recharge, consume, refund, transfer
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    
    -- 关联信息
    related_order_id UUID,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 充值记录表
CREATE TABLE IF NOT EXISTS recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- 充值信息
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT '金币',
    payment_amount DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0), -- 实际支付金额
    payment_currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    
    -- 支付信息
    payment_method VARCHAR(50) NOT NULL, -- alipay, wechat, card等
    payment_id VARCHAR(255), -- 第三方支付订单号
    transaction_id VARCHAR(255) UNIQUE, -- 内部交易号
    
    -- 状态管理
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded')),
    failure_reason TEXT,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- 支付卡片表
CREATE TABLE IF NOT EXISTS payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_number VARCHAR(20) UNIQUE NOT NULL,
    card_type VARCHAR(20) NOT NULL CHECK (card_type IN ('gift', 'prepaid', 'credit')),
    balance DECIMAL(10,2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    
    -- 卡片信息
    holder_name VARCHAR(100),
    expiry_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 限制信息
    daily_limit DECIMAL(10,2) DEFAULT 10000.00,
    monthly_limit DECIMAL(10,2) DEFAULT 100000.00,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 充值订单表
CREATE TABLE IF NOT EXISTS recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 订单信息
    order_no VARCHAR(32) UNIQUE NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    coins_amount BIGINT NOT NULL CHECK (coins_amount > 0),
    exchange_rate INTEGER NOT NULL DEFAULT 10,
    
    -- 支付信息
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('card', 'alipay', 'wechat')),
    payment_card_id UUID REFERENCES payment_cards(id),
    
    -- 优惠信息
    original_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    bonus_coins BIGINT NOT NULL DEFAULT 0,
    card_code VARCHAR(16),
    
    -- 状态管理
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'paid', 'refunded')),
    failure_reason TEXT,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- 充值套餐表
CREATE TABLE IF NOT EXISTS recharge_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    coins_amount BIGINT NOT NULL CHECK (coins_amount > 0),
    bonus_coins BIGINT NOT NULL DEFAULT 0,
    discount_rate DECIMAL(5,4) NOT NULL DEFAULT 1.0000,
    
    -- 显示设置
    is_popular BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 3. 卡密系统
-- =============================================================================

-- 卡密表
CREATE TABLE IF NOT EXISTS gift_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_code VARCHAR(20) UNIQUE NOT NULL,
    card_type VARCHAR(20) NOT NULL CHECK (card_type IN ('discount', 'bonus', 'privilege')),
    value DECIMAL(10,2) NOT NULL CHECK (value > 0),
    
    -- 优惠配置
    discount_rate DECIMAL(5,4) NOT NULL DEFAULT 1.0000 CHECK (discount_rate > 0 AND discount_rate <= 1),
    bonus_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0000 CHECK (bonus_rate >= 0),
    
    -- 特权配置
    privileges JSONB DEFAULT '[]',
    privilege_duration VARCHAR(20),
    
    -- 使用限制
    max_uses INTEGER DEFAULT 1,
    current_uses INTEGER DEFAULT 0,
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- 描述信息
    description TEXT,
    created_by UUID REFERENCES users(id),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 卡密使用记录表
CREATE TABLE IF NOT EXISTS card_usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES gift_cards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 使用信息
    original_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    final_amount DECIMAL(10,2) NOT NULL,
    bonus_coins BIGINT DEFAULT 0,
    
    -- 关联订单
    order_id UUID,
    order_type VARCHAR(20) NOT NULL,
    
    -- 时间戳
    used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 用户特权表
CREATE TABLE IF NOT EXISTS user_privileges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    privilege_type VARCHAR(50) NOT NULL,
    
    -- 特权信息
    granted_by_card_id UUID REFERENCES gift_cards(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 使用统计
    usage_count INTEGER DEFAULT 0,
    usage_limit INTEGER,
    
    -- 时间戳
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, privilege_type)
);

-- =============================================================================
-- 4. AI模型系统
-- =============================================================================

-- AI模型提供商表
CREATE TABLE IF NOT EXISTS model_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    base_url TEXT NOT NULL,
    api_key_encrypted TEXT,
    
    -- 配置信息
    default_headers JSONB DEFAULT '{}',
    rate_limit_config JSONB DEFAULT '{}',
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    last_health_check TIMESTAMP WITH TIME ZONE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- AI模型表
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_key VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    model_type VARCHAR(50) NOT NULL, -- chat, image, video, audio, etc.
    
    -- 能力配置
    capabilities JSONB NOT NULL DEFAULT '[]',
    
    -- 参数配置
    params_schema JSONB NOT NULL DEFAULT '{}',
    
    -- 系统配置
    model_system_prompt TEXT,
    base_url TEXT,
    api_key_encrypted TEXT,
    
    -- 定价配置
    pricing JSONB NOT NULL DEFAULT '{}',
    
    -- 权限控制
    visibility VARCHAR(20) DEFAULT 'public' CHECK (visibility IN ('public', 'vip_only', 'creator_only', 'admin_only', 'hidden')),
    min_user_type VARCHAR(20) DEFAULT 'basic' CHECK (min_user_type IN ('basic', 'vip', 'creator', 'admin')),
    min_balance DECIMAL(15,2) DEFAULT 0.00,
    daily_limit INTEGER DEFAULT 1000,
    user_limit INTEGER DEFAULT 100,
    
    -- 健康和性能
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    weight INTEGER DEFAULT 100,
    fallback_chain JSONB DEFAULT '[]',
    connectivity_test_endpoint TEXT,
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_health_check TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- 5. 角色系统
-- =============================================================================

-- 角色表
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

-- 角色标签表
CREATE TABLE IF NOT EXISTS character_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(character_id, tag)
);

-- 角色关系表
CREATE TABLE IF NOT EXISTS character_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    target_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 关系类型
    relationship_type VARCHAR(50) NOT NULL, -- friend, family, colleague, romantic, enemy, etc.
    custom_type_name VARCHAR(100), -- 自定义关系名称

    -- 关系强度和情感维度 (0.0 - 1.0)
    strength DECIMAL(3,2) DEFAULT 0.5 CHECK (strength >= 0 AND strength <= 1),
    trust DECIMAL(3,2) DEFAULT 0.5 CHECK (trust >= 0 AND trust <= 1),
    affection DECIMAL(3,2) DEFAULT 0.5 CHECK (affection >= 0 AND affection <= 1),
    respect DECIMAL(3,2) DEFAULT 0.5 CHECK (respect >= 0 AND respect <= 1),
    intimacy DECIMAL(3,2) DEFAULT 0.5 CHECK (intimacy >= 0 AND intimacy <= 1),

    -- 互动风格
    tone VARCHAR(50), -- 语调：温柔、严肃、调皮等
    formality_level VARCHAR(20), -- 正式程度：formal, casual, intimate

    -- 关系描述
    description TEXT,
    backstory TEXT, -- 关系背景故事

    -- 互动历史
    last_interaction_at TIMESTAMP WITH TIME ZONE,
    interaction_count INTEGER DEFAULT 0,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- 防止自己和自己建立关系
    CHECK (character_id != target_character_id),
    UNIQUE(character_id, target_character_id)
);

-- =============================================================================
-- 6. 群聊系统
-- =============================================================================

-- 群聊表
CREATE TABLE IF NOT EXISTS group_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,

    -- 群聊设置
    background_image_url TEXT,
    world_setting TEXT, -- 世界观设定
    max_members INTEGER DEFAULT 50,

    -- 章节系统
    current_chapter_id UUID,
    auto_chapter_progression BOOLEAN DEFAULT FALSE,

    -- 权限设置
    is_public BOOLEAN DEFAULT FALSE,
    allow_member_invite BOOLEAN DEFAULT TRUE,
    require_approval BOOLEAN DEFAULT FALSE,

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
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 成员角色
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),

    -- 成员设置
    nickname VARCHAR(100), -- 群内昵称
    is_muted BOOLEAN DEFAULT FALSE,

    -- 统计信息
    message_count INTEGER DEFAULT 0,
    last_read_at TIMESTAMP WITH TIME ZONE,

    -- 时间戳
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(group_chat_id, character_id)
);

-- 故事章节表
CREATE TABLE IF NOT EXISTS story_chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,

    -- 章节信息
    title VARCHAR(200) NOT NULL,
    description TEXT,
    chapter_order INTEGER NOT NULL,

    -- 视觉设置
    background_image_url TEXT,
    background_music_url TEXT,

    -- 触发条件
    trigger_conditions JSONB DEFAULT '{}',
    auto_trigger BOOLEAN DEFAULT FALSE,

    -- 演绎设定
    performance_settings JSONB DEFAULT '{}',

    -- 状态管理
    is_active BOOLEAN DEFAULT FALSE,
    is_completed BOOLEAN DEFAULT FALSE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(group_chat_id, chapter_order)
);

-- 群聊消息表
CREATE TABLE IF NOT EXISTS group_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- 消息内容
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'audio', 'video', 'system')),
    media_url TEXT,

    -- 消息元数据
    reply_to_id UUID REFERENCES group_chat_messages(id),
    is_ai_generated BOOLEAN DEFAULT FALSE,
    generation_model VARCHAR(100),

    -- 章节关联
    chapter_id UUID REFERENCES story_chapters(id),

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- 确保发送者存在
    CHECK ((character_id IS NOT NULL AND user_id IS NULL) OR (character_id IS NULL AND user_id IS NOT NULL))
);

-- =============================================================================
-- 7. 朋友圈系统
-- =============================================================================

-- 朋友圈动态表
CREATE TABLE IF NOT EXISTS moments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 内容信息
    content TEXT NOT NULL,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('text', 'image', 'video', 'talking_head', '3d')),
    media_url TEXT,
    media_type VARCHAR(50),

    -- 可见性设置
    visibility VARCHAR(10) NOT NULL CHECK (visibility IN ('public', 'friends', 'private')) DEFAULT 'friends',

    -- 生成信息
    is_generated BOOLEAN NOT NULL DEFAULT TRUE,
    generation_model VARCHAR(100),
    generation_prompt TEXT,

    -- 附加信息
    mood VARCHAR(50),
    location VARCHAR(200),
    tags TEXT[], -- PostgreSQL数组类型

    -- 统计信息
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    share_count INTEGER NOT NULL DEFAULT 0,
    view_count INTEGER NOT NULL DEFAULT 0,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 朋友圈草稿表
CREATE TABLE IF NOT EXISTS moment_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 草稿内容
    content TEXT NOT NULL,
    content_type VARCHAR(20) NOT NULL DEFAULT 'text',
    media_url TEXT,

    -- 生成信息
    generation_model VARCHAR(100),
    generation_prompt TEXT,
    similarity_score DECIMAL(3,2), -- 与历史内容的相似度分数(0-1)

    -- 调度信息
    scheduled_at TIMESTAMP WITH TIME ZONE,
    priority INTEGER DEFAULT 0, -- 优先级，数字越大优先级越高

    -- 审核状态
    review_status VARCHAR(20) DEFAULT 'pending' CHECK (review_status IN ('pending', 'approved', 'rejected')),
    review_notes TEXT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 朋友圈互动表
CREATE TABLE IF NOT EXISTS moment_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,

    -- 互动类型
    type VARCHAR(10) NOT NULL CHECK (type IN ('like', 'comment', 'share')),
    content TEXT, -- 评论内容，点赞和分享为空

    -- 生成信息
    is_ai_generated BOOLEAN DEFAULT FALSE,
    generation_model VARCHAR(100),

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- 确保用户或角色至少有一个
    CONSTRAINT check_interaction_actor CHECK (
        (user_id IS NOT NULL AND character_id IS NULL) OR
        (user_id IS NULL AND character_id IS NOT NULL)
    ),

    -- 防止重复点赞
    CONSTRAINT unique_like_per_actor UNIQUE (moment_id, user_id, character_id, type)
);

-- 朋友圈自动生成配置表
CREATE TABLE IF NOT EXISTS moment_auto_generation_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 生成频率配置
    is_enabled BOOLEAN DEFAULT TRUE,
    frequency_type VARCHAR(20) DEFAULT 'daily' CHECK (frequency_type IN ('hourly', 'daily', 'weekly', 'random')),
    frequency_value INTEGER DEFAULT 1, -- 频率数值

    -- 生成时间配置
    preferred_hours INTEGER[] DEFAULT '{9,12,18,21}', -- 偏好的发布小时
    timezone VARCHAR(50) DEFAULT 'Asia/Shanghai',

    -- 内容配置
    content_types VARCHAR(20)[] DEFAULT '{text}',
    mood_preferences VARCHAR(50)[] DEFAULT '{}',
    topic_preferences TEXT[],

    -- 质量控制
    min_similarity_threshold DECIMAL(3,2) DEFAULT 0.3, -- 最小相似度阈值，避免重复内容
    max_daily_posts INTEGER DEFAULT 3,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(character_id)
);

-- 朋友圈通知表
CREATE TABLE IF NOT EXISTS moment_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,

    -- 通知信息
    type VARCHAR(20) NOT NULL CHECK (type IN ('new_moment', 'like', 'comment', 'share')),
    content TEXT NOT NULL,

    -- 状态管理
    is_read BOOLEAN NOT NULL DEFAULT FALSE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 8. 系统配置表
-- =============================================================================

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) DEFAULT 'string' CHECK (config_type IN ('string', 'number', 'boolean', 'json')),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE, -- 是否对前端公开

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 8.1. 全局提示词系统 - YUNAI核心AI智能基础
-- =============================================================================

-- 全局提示词模板表
CREATE TABLE IF NOT EXISTS global_prompt_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- 模板内容 - 支持变量替换
    template_content TEXT NOT NULL,
    variables JSONB DEFAULT '[]', -- 模板变量定义: [{"name": "character_name", "type": "string", "required": true}]
    
    -- 功能分类
    category VARCHAR(50) NOT NULL CHECK (category IN ('system', 'character', 'moments', 'story', 'invitation', 'relationship', 'voice', 'embedding')),
    subcategory VARCHAR(50),
    
    -- 应用场景
    usage_scenarios JSONB DEFAULT '[]', -- ["chat", "moments_generation", "story_trigger", "relationship_analysis"]
    
    -- 优先级和版本
    priority INTEGER DEFAULT 100,
    version VARCHAR(20) DEFAULT '1.0',
    
    -- 模型适配
    compatible_models JSONB DEFAULT '[]', -- 兼容的模型列表
    model_specific_adjustments JSONB DEFAULT '{}', -- 针对特定模型的调整
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE, -- 是否为该分类的默认模板
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 用户自定义提示词表
CREATE TABLE IF NOT EXISTS user_custom_prompts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE, -- NULL表示全局自定义
    
    -- 提示词信息
    prompt_name VARCHAR(100) NOT NULL,
    prompt_content TEXT NOT NULL,
    
    -- 分类和场景
    category VARCHAR(50) NOT NULL,
    usage_scenarios JSONB DEFAULT '[]',
    
    -- 继承关系
    base_template_id UUID REFERENCES global_prompt_templates(id) ON DELETE SET NULL,
    inherits_from_global BOOLEAN DEFAULT TRUE, -- 是否继承全局模板
    
    -- 个性化配置
    custom_variables JSONB DEFAULT '{}', -- 用户自定义的变量值
    override_global BOOLEAN DEFAULT FALSE, -- 是否完全覆盖全局提示词
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保用户在同一分类下的自定义提示词唯一
    UNIQUE(user_id, character_id, category)
);

-- 提示词使用统计表
CREATE TABLE IF NOT EXISTS prompt_usage_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID REFERENCES global_prompt_templates(id) ON DELETE CASCADE,
    user_prompt_id UUID REFERENCES user_custom_prompts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    
    -- 使用信息
    usage_scenario VARCHAR(50) NOT NULL,
    model_used VARCHAR(100),
    
    -- 效果统计
    tokens_generated INTEGER DEFAULT 0,
    response_quality_score DECIMAL(3,2), -- 0.00-1.00
    user_satisfaction_score DECIMAL(3,2), -- 用户反馈评分
    
    -- 性能指标
    response_time_ms INTEGER,
    cost_amount DECIMAL(10,4) DEFAULT 0,
    
    -- 时间戳
    used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 提示词版本历史表
CREATE TABLE IF NOT EXISTS prompt_version_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID REFERENCES global_prompt_templates(id) ON DELETE CASCADE,
    user_prompt_id UUID REFERENCES user_custom_prompts(id) ON DELETE CASCADE,
    
    -- 版本信息
    version_number VARCHAR(20) NOT NULL,
    change_description TEXT,
    previous_content TEXT,
    new_content TEXT,
    
    -- 变更信息
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    change_type VARCHAR(20) CHECK (change_type IN ('create', 'update', 'activate', 'deactivate', 'delete')),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 8.2. 真实情感和世界观驱动系统
-- =============================================================================

-- 角色情感状态表
CREATE TABLE IF NOT EXISTS character_emotional_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 当前情感状态
    primary_emotion VARCHAR(50) NOT NULL, -- happy, sad, angry, fearful, surprised, disgusted, neutral
    secondary_emotions JSONB DEFAULT '[]', -- 次要情感列表
    emotion_intensity DECIMAL(3,2) DEFAULT 0.5 CHECK (emotion_intensity >= 0 AND emotion_intensity <= 1),
    emotion_stability VARCHAR(20) DEFAULT 'stable' CHECK (emotion_stability IN ('stable', 'fluctuating', 'volatile')),
    
    -- 情感触发因素
    triggers JSONB DEFAULT '[]', -- 引起当前情感的因素
    trigger_event_id UUID, -- 触发事件ID
    
    -- 身体和行为表现
    physical_manifestations JSONB DEFAULT '[]', -- 身体表现（脸红、顾虑等）
    behavioral_tendencies JSONB DEFAULT '[]', -- 行为倾向
    
    -- 时间预测
    estimated_duration INTEGER, -- 预计持续时间（分钟）
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expected_end_at TIMESTAMP WITH TIME ZONE,
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 角色记忆系统表
CREATE TABLE IF NOT EXISTS character_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 记忆内容
    memory_content TEXT NOT NULL,
    memory_summary VARCHAR(500), -- 简要描述
    memory_type VARCHAR(50) NOT NULL, -- core, important, normal, minor, temporary
    
    -- 记忆重要性和情感权重
    importance_score INTEGER DEFAULT 3 CHECK (importance_score >= 1 AND importance_score <= 5),
    emotional_weight VARCHAR(20) DEFAULT 'neutral' CHECK (emotional_weight IN ('positive', 'negative', 'neutral')),
    emotional_intensity DECIMAL(3,2) DEFAULT 0.5,
    
    -- 关联信息
    associated_people JSONB DEFAULT '[]', -- 相关人物
    associated_places JSONB DEFAULT '[]', -- 相关地点
    associated_events JSONB DEFAULT '[]', -- 相关事件
    memory_tags JSONB DEFAULT '[]', -- 记忆标签
    
    -- 源信息
    source_type VARCHAR(50), -- conversation, event, observation, reflection
    source_id UUID, -- 源事件或对话的ID
    
    -- 记忆衰减和强化
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    decay_rate DECIMAL(5,4) DEFAULT 0.0100, -- 衰减率
    reinforcement_count INTEGER DEFAULT 0, -- 强化次数
    
    -- 时间信息
    memory_date TIMESTAMP WITH TIME ZONE, -- 记忆发生的时间
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 角色成长记录表
CREATE TABLE IF NOT EXISTS character_growth_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- 成长事件
    growth_type VARCHAR(50) NOT NULL, -- personality_shift, new_belief, skill_development, relationship_change
    growth_description TEXT NOT NULL,
    
    -- 变化内容
    before_state JSONB DEFAULT '{}', -- 变化前的状态
    after_state JSONB DEFAULT '{}', -- 变化后的状态
    change_magnitude DECIMAL(3,2) DEFAULT 0.1 CHECK (change_magnitude >= 0 AND change_magnitude <= 1),
    
    -- 触发原因
    trigger_events JSONB DEFAULT '[]', -- 引起成长的事件
    catalyst_people JSONB DEFAULT '[]', -- 关键人物
    
    -- 影响范围
    affected_areas JSONB DEFAULT '[]', -- 受影响的领域（性格、价值观、技能等）
    relationship_impacts JSONB DEFAULT '{}', -- 对关系的影响
    
    -- 未来影响
    predicted_behaviors JSONB DEFAULT '[]', -- 预测的行为变化
    growth_permanence VARCHAR(20) DEFAULT 'moderate' CHECK (growth_permanence IN ('temporary', 'moderate', 'permanent')),
    
    -- 时间戳
    growth_occurred_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 世界观设定表
CREATE TABLE IF NOT EXISTS world_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 世界基本信息
    world_name VARCHAR(200) NOT NULL,
    world_description TEXT,
    world_type VARCHAR(50), -- fantasy, sci_fi, modern, historical, custom
    
    -- 世界规则
    physical_laws TEXT, -- 物理法则
    magic_system TEXT, -- 魔法或超能力系统
    technology_level VARCHAR(100), -- 技术水平
    
    -- 社会结构
    social_structure TEXT, -- 社会结构
    political_system TEXT, -- 政治制度
    economic_system TEXT, -- 经济体系
    cultural_norms TEXT, -- 文化规范
    
    -- 历史背景
    world_history TEXT, -- 世界历史
    important_events JSONB DEFAULT '[]', -- 重要历史事件
    current_era VARCHAR(100), -- 当前时代
    
    -- 地理信息
    geography_description TEXT, -- 地理描述
    important_locations JSONB DEFAULT '[]', -- 重要地点
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    is_public BOOLEAN DEFAULT FALSE, -- 是否允许他人使用
    usage_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 世界观一致性检查记录表
CREATE TABLE IF NOT EXISTS world_consistency_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_setting_id UUID NOT NULL REFERENCES world_settings(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    
    -- 检查内容
    checked_behavior TEXT NOT NULL, -- 被检查的行为或事件
    consistency_score DECIMAL(3,2) DEFAULT 1.0,
    is_consistent BOOLEAN DEFAULT TRUE,
    
    -- 违反信息
    violations JSONB DEFAULT '[]', -- 违反项列表
    severity_level VARCHAR(20) DEFAULT 'minor' CHECK (severity_level IN ('minor', 'major', 'critical')),
    
    -- 建议和修正
    suggestions JSONB DEFAULT '[]', -- 修正建议
    alternative_outcomes JSONB DEFAULT '[]', -- 替代结果
    auto_corrected BOOLEAN DEFAULT FALSE, -- 是否自动修正
    
    -- 检查类型
    check_type VARCHAR(50) NOT NULL, -- physics, social, cultural, logical, historical
    triggered_by VARCHAR(50), -- message, action, event
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 角色世界观关联表
CREATE TABLE IF NOT EXISTS character_world_associations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    world_setting_id UUID NOT NULL REFERENCES world_settings(id) ON DELETE CASCADE,
    
    -- 在世界中的身份
    role_in_world VARCHAR(200), -- 在世界中的身份和地位
    world_specific_abilities JSONB DEFAULT '[]', -- 在这个世界中的能力
    world_specific_limitations JSONB DEFAULT '[]', -- 在这个世界中的限制
    
    -- 世界特定记忆
    world_memories JSONB DEFAULT '[]', -- 在这个世界中的记忆
    world_relationships JSONB DEFAULT '{}', -- 在这个世界中的人际关系
    
    -- 适应程度
    adaptation_level DECIMAL(3,2) DEFAULT 0.5, -- 对世界的适应程度
    immersion_depth DECIMAL(3,2) DEFAULT 0.5, -- 沉浸深度
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(character_id, world_setting_id)
);

-- 功能开关表
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_key VARCHAR(100) UNIQUE NOT NULL,
    flag_name VARCHAR(200) NOT NULL,
    description TEXT,

    -- 开关状态
    is_enabled BOOLEAN DEFAULT FALSE,
    rollout_percentage INTEGER DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),

    -- 目标用户
    target_user_types VARCHAR(20)[] DEFAULT '{}',
    target_user_ids UUID[] DEFAULT '{}',

    -- 环境限制
    environments VARCHAR(20)[] DEFAULT '{production,development}',

    -- 时间限制
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 9. AI请求和计费系统
-- =============================================================================

-- AI请求记录表
CREATE TABLE IF NOT EXISTS ai_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE SET NULL,

    -- 请求信息
    model_id UUID NOT NULL REFERENCES ai_models(id),
    request_type VARCHAR(50) NOT NULL, -- chat, image_generation, voice_synthesis, etc.

    -- 请求内容
    prompt TEXT,
    request_params JSONB DEFAULT '{}',

    -- 响应信息
    response_content TEXT,
    response_metadata JSONB DEFAULT '{}',

    -- 计费信息
    tokens_used INTEGER DEFAULT 0,
    cost_amount DECIMAL(10,4) DEFAULT 0,
    cost_currency VARCHAR(10) DEFAULT 'coins',

    -- 性能指标
    response_time_ms INTEGER,

    -- 状态管理
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    error_message TEXT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- AI计费规则表
CREATE TABLE IF NOT EXISTS ai_billing_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,

    -- 计费规则
    billing_type VARCHAR(20) NOT NULL CHECK (billing_type IN ('per_request', 'per_token', 'per_minute', 'per_character')),
    base_cost DECIMAL(10,4) NOT NULL DEFAULT 0,

    -- 用户类型定价
    basic_multiplier DECIMAL(5,2) DEFAULT 1.00,
    vip_multiplier DECIMAL(5,2) DEFAULT 0.80,
    creator_multiplier DECIMAL(5,2) DEFAULT 0.60,
    admin_multiplier DECIMAL(5,2) DEFAULT 0.00,

    -- 批量折扣
    volume_discounts JSONB DEFAULT '[]', -- [{"min_usage": 100, "discount": 0.1}]

    -- 时间段定价
    time_based_pricing JSONB DEFAULT '{}',

    -- 生效时间
    effective_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    effective_until TIMESTAMP WITH TIME ZONE,

    -- 状态
    is_active BOOLEAN DEFAULT TRUE,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 10. 代充系统
-- =============================================================================

-- 代充订单表
CREATE TABLE IF NOT EXISTS proxy_recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no VARCHAR(50) UNIQUE NOT NULL,
    payer_user_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID NOT NULL REFERENCES users(id),
    payment_card_id UUID NOT NULL REFERENCES payment_cards(id),

    -- 金额信息
    original_amount DECIMAL(10,2) NOT NULL,
    actual_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    coins_amount BIGINT NOT NULL,
    bonus_coins BIGINT DEFAULT 0,
    total_coins BIGINT NOT NULL,
    exchange_rate INTEGER NOT NULL DEFAULT 10,

    -- 卡片信息
    card_type VARCHAR(50),
    message TEXT,

    -- 状态管理
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    failure_reason TEXT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- 11. 卡片交易记录
-- =============================================================================

-- 卡片交易记录表
CREATE TABLE IF NOT EXISTS card_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES payment_cards(id) ON DELETE CASCADE,

    -- 交易信息
    transaction_type VARCHAR(20) NOT NULL, -- recharge, payment, refund
    amount DECIMAL(10,2) NOT NULL,
    balance_before DECIMAL(10,2) NOT NULL,
    balance_after DECIMAL(10,2) NOT NULL,

    -- 关联信息
    related_order_id UUID,
    description TEXT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 12. 索引和约束
-- =============================================================================

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- 钱包表索引
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_id ON wallet_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at);

-- 充值记录索引
CREATE INDEX IF NOT EXISTS idx_recharge_records_user_id ON recharge_records(user_id);
CREATE INDEX IF NOT EXISTS idx_recharge_records_status ON recharge_records(status);
CREATE INDEX IF NOT EXISTS idx_recharge_records_created_at ON recharge_records(created_at);

-- 充值订单索引
CREATE INDEX IF NOT EXISTS idx_recharge_orders_user_id ON recharge_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_order_no ON recharge_orders(order_no);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_status ON recharge_orders(status);

-- 卡密表索引
CREATE INDEX IF NOT EXISTS idx_gift_cards_card_code ON gift_cards(card_code);
CREATE INDEX IF NOT EXISTS idx_gift_cards_card_type ON gift_cards(card_type);
CREATE INDEX IF NOT EXISTS idx_gift_cards_is_active ON gift_cards(is_active);

-- AI模型表索引
CREATE INDEX IF NOT EXISTS idx_ai_models_internal_key ON ai_models(internal_key);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider ON ai_models(provider);
CREATE INDEX IF NOT EXISTS idx_ai_models_model_type ON ai_models(model_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_is_active ON ai_models(is_active);

-- 角色表索引
CREATE INDEX IF NOT EXISTS idx_characters_user_id ON characters(user_id);
CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name);
CREATE INDEX IF NOT EXISTS idx_characters_visibility ON characters(visibility);
CREATE INDEX IF NOT EXISTS idx_characters_is_featured ON characters(is_featured);

-- 角色关系索引
CREATE INDEX IF NOT EXISTS idx_character_relationships_character_id ON character_relationships(character_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_target_character_id ON character_relationships(target_character_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_type ON character_relationships(relationship_type);

-- 群聊表索引
CREATE INDEX IF NOT EXISTS idx_group_chats_user_id ON group_chats(user_id);
CREATE INDEX IF NOT EXISTS idx_group_chats_is_public ON group_chats(is_public);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_group_id ON group_chat_members(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_members_character_id ON group_chat_members(character_id);

-- 群聊消息索引
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_group_id ON group_chat_messages(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_character_id ON group_chat_messages(character_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_created_at ON group_chat_messages(created_at);

-- 朋友圈表索引
CREATE INDEX IF NOT EXISTS idx_moments_character_id ON moments(character_id);
CREATE INDEX IF NOT EXISTS idx_moments_user_id ON moments(user_id);
CREATE INDEX IF NOT EXISTS idx_moments_visibility ON moments(visibility);
CREATE INDEX IF NOT EXISTS idx_moments_created_at ON moments(created_at);

-- 朋友圈互动索引
CREATE INDEX IF NOT EXISTS idx_moment_interactions_moment_id ON moment_interactions(moment_id);
CREATE INDEX IF NOT EXISTS idx_moment_interactions_user_id ON moment_interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_moment_interactions_character_id ON moment_interactions(character_id);
CREATE INDEX IF NOT EXISTS idx_moment_interactions_type ON moment_interactions(type);

-- AI请求索引
CREATE INDEX IF NOT EXISTS idx_ai_requests_user_id ON ai_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_requests_model_id ON ai_requests(model_id);
CREATE INDEX IF NOT EXISTS idx_ai_requests_status ON ai_requests(status);
CREATE INDEX IF NOT EXISTS idx_ai_requests_created_at ON ai_requests(created_at);

-- =============================================================================
-- 13. 触发器函数
-- =============================================================================

-- 更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表添加更新时间戳触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recharge_records_updated_at BEFORE UPDATE ON recharge_records FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_payment_cards_updated_at BEFORE UPDATE ON payment_cards FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recharge_orders_updated_at BEFORE UPDATE ON recharge_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_recharge_packages_updated_at BEFORE UPDATE ON recharge_packages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_gift_cards_updated_at BEFORE UPDATE ON gift_cards FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_model_providers_updated_at BEFORE UPDATE ON model_providers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_models_updated_at BEFORE UPDATE ON ai_models FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_character_relationships_updated_at BEFORE UPDATE ON character_relationships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_group_chats_updated_at BEFORE UPDATE ON group_chats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_moments_updated_at BEFORE UPDATE ON moments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_moment_drafts_updated_at BEFORE UPDATE ON moment_drafts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_moment_auto_generation_configs_updated_at BEFORE UPDATE ON moment_auto_generation_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_feature_flags_updated_at BEFORE UPDATE ON feature_flags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_billing_rules_updated_at BEFORE UPDATE ON ai_billing_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_proxy_recharge_orders_updated_at BEFORE UPDATE ON proxy_recharge_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 14. 初始数据插入
-- =============================================================================

-- 插入默认系统配置
INSERT INTO system_configs (config_key, config_value, config_type, description, is_public) VALUES
('site_name', 'YUNAI', 'string', '网站名称', true),
('site_description', 'AI驱动的智能社交平台', 'string', '网站描述', true),
('default_exchange_rate', '10', 'number', '默认汇率：1元=10金币', false),
('max_characters_per_user', '10', 'number', '每个用户最大角色数量', false),
('max_group_members', '50', 'number', '群聊最大成员数', false),
('ai_request_timeout', '30', 'number', 'AI请求超时时间（秒）', false),
('enable_auto_moments', 'true', 'boolean', '启用自动朋友圈生成', false),
('moments_daily_limit', '5', 'number', '每日朋友圈生成限制', false)
ON CONFLICT (config_key) DO NOTHING;

-- 插入默认功能开关
INSERT INTO feature_flags (flag_key, flag_name, description, is_enabled) VALUES
('enable_character_creation', '角色创建功能', '允许用户创建新角色', true),
('enable_group_chat', '群聊功能', '启用群聊功能', true),
('enable_moments', '朋友圈功能', '启用朋友圈功能', true),
('enable_voice_calls', '语音通话功能', '启用语音通话功能', true),
('enable_ai_auto_reply', 'AI自动回复', '启用AI自动回复功能', true),
('enable_proxy_recharge', '代充功能', '启用代充功能', true),
('enable_gift_cards', '卡密功能', '启用卡密系统', true),
('enable_character_relationships', '角色关系功能', '启用角色关系管理', true)
ON CONFLICT (flag_key) DO NOTHING;

-- 插入默认充值套餐
INSERT INTO recharge_packages (name, amount, coins_amount, bonus_coins, discount_rate, is_popular, sort_order) VALUES
('体验包', 10.00, 100, 0, 1.0000, false, 1),
('基础包', 50.00, 500, 50, 1.0000, false, 2),
('超值包', 100.00, 1000, 200, 1.0000, true, 3),
('豪华包', 200.00, 2000, 500, 1.0000, false, 4),
('至尊包', 500.00, 5000, 1500, 1.0000, false, 5)
ON CONFLICT DO NOTHING;

-- 创建默认管理员用户（如果不存在）
INSERT INTO users (username, email, password_hash, user_type, nickname, is_active, email_verified) VALUES
('admin', 'admin@yunai.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', '系统管理员', true, true)
ON CONFLICT (username) DO NOTHING;

-- 为管理员创建钱包
INSERT INTO wallets (user_id, balance, currency)
SELECT id, 10000.00, '金币' FROM users WHERE username = 'admin'
ON CONFLICT (user_id) DO NOTHING;

-- =============================================================================
-- 15. 全局提示词系统索引
-- =============================================================================

-- 全局提示词模板索引
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_category ON global_prompt_templates(category);
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_is_active ON global_prompt_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_priority ON global_prompt_templates(priority);
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_is_default ON global_prompt_templates(is_default);
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_created_by ON global_prompt_templates(created_by);

-- 用户自定义提示词索引
CREATE INDEX IF NOT EXISTS idx_user_custom_prompts_user_id ON user_custom_prompts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_custom_prompts_character_id ON user_custom_prompts(character_id);
CREATE INDEX IF NOT EXISTS idx_user_custom_prompts_category ON user_custom_prompts(category);
CREATE INDEX IF NOT EXISTS idx_user_custom_prompts_is_active ON user_custom_prompts(is_active);
CREATE INDEX IF NOT EXISTS idx_user_custom_prompts_base_template_id ON user_custom_prompts(base_template_id);

-- 提示词使用统计索引
CREATE INDEX IF NOT EXISTS idx_prompt_usage_stats_template_id ON prompt_usage_stats(template_id);
CREATE INDEX IF NOT EXISTS idx_prompt_usage_stats_user_id ON prompt_usage_stats(user_id);
CREATE INDEX IF NOT EXISTS idx_prompt_usage_stats_character_id ON prompt_usage_stats(character_id);
CREATE INDEX IF NOT EXISTS idx_prompt_usage_stats_usage_scenario ON prompt_usage_stats(usage_scenario);
CREATE INDEX IF NOT EXISTS idx_prompt_usage_stats_used_at ON prompt_usage_stats(used_at);

-- 提示词版本历史索引
CREATE INDEX IF NOT EXISTS idx_prompt_version_history_template_id ON prompt_version_history(template_id);
CREATE INDEX IF NOT EXISTS idx_prompt_version_history_user_prompt_id ON prompt_version_history(user_prompt_id);
CREATE INDEX IF NOT EXISTS idx_prompt_version_history_created_at ON prompt_version_history(created_at);

-- 提示词生成器配置索引
CREATE INDEX IF NOT EXISTS idx_prompt_generator_configs_generator_type ON prompt_generator_configs(generator_type);
CREATE INDEX IF NOT EXISTS idx_prompt_generator_configs_is_active ON prompt_generator_configs(is_active);

-- 真实情感和世界观系统索引
CREATE INDEX IF NOT EXISTS idx_character_emotional_states_character_id ON character_emotional_states(character_id);
CREATE INDEX IF NOT EXISTS idx_character_emotional_states_primary_emotion ON character_emotional_states(primary_emotion);
CREATE INDEX IF NOT EXISTS idx_character_emotional_states_is_active ON character_emotional_states(is_active);
CREATE INDEX IF NOT EXISTS idx_character_emotional_states_started_at ON character_emotional_states(started_at);

CREATE INDEX IF NOT EXISTS idx_character_memories_character_id ON character_memories(character_id);
CREATE INDEX IF NOT EXISTS idx_character_memories_memory_type ON character_memories(memory_type);
CREATE INDEX IF NOT EXISTS idx_character_memories_importance_score ON character_memories(importance_score);
CREATE INDEX IF NOT EXISTS idx_character_memories_emotional_weight ON character_memories(emotional_weight);
CREATE INDEX IF NOT EXISTS idx_character_memories_memory_date ON character_memories(memory_date);
CREATE INDEX IF NOT EXISTS idx_character_memories_last_accessed_at ON character_memories(last_accessed_at);

CREATE INDEX IF NOT EXISTS idx_character_growth_records_character_id ON character_growth_records(character_id);
CREATE INDEX IF NOT EXISTS idx_character_growth_records_growth_type ON character_growth_records(growth_type);
CREATE INDEX IF NOT EXISTS idx_character_growth_records_growth_occurred_at ON character_growth_records(growth_occurred_at);

CREATE INDEX IF NOT EXISTS idx_world_settings_user_id ON world_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_world_settings_world_type ON world_settings(world_type);
CREATE INDEX IF NOT EXISTS idx_world_settings_is_active ON world_settings(is_active);
CREATE INDEX IF NOT EXISTS idx_world_settings_is_public ON world_settings(is_public);

CREATE INDEX IF NOT EXISTS idx_world_consistency_logs_world_setting_id ON world_consistency_logs(world_setting_id);
CREATE INDEX IF NOT EXISTS idx_world_consistency_logs_character_id ON world_consistency_logs(character_id);
CREATE INDEX IF NOT EXISTS idx_world_consistency_logs_is_consistent ON world_consistency_logs(is_consistent);
CREATE INDEX IF NOT EXISTS idx_world_consistency_logs_severity_level ON world_consistency_logs(severity_level);
CREATE INDEX IF NOT EXISTS idx_world_consistency_logs_created_at ON world_consistency_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_character_world_associations_character_id ON character_world_associations(character_id);
CREATE INDEX IF NOT EXISTS idx_character_world_associations_world_setting_id ON character_world_associations(world_setting_id);
CREATE INDEX IF NOT EXISTS idx_character_world_associations_is_active ON character_world_associations(is_active);

-- 添加全局提示词系统的触发器
CREATE TRIGGER update_global_prompt_templates_updated_at BEFORE UPDATE ON global_prompt_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_custom_prompts_updated_at BEFORE UPDATE ON user_custom_prompts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_prompt_generator_configs_updated_at BEFORE UPDATE ON prompt_generator_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 真实情感和世界观系统触发器
CREATE TRIGGER update_character_emotional_states_updated_at BEFORE UPDATE ON character_emotional_states FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_character_memories_updated_at BEFORE UPDATE ON character_memories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_world_settings_updated_at BEFORE UPDATE ON world_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_character_world_associations_updated_at BEFORE UPDATE ON character_world_associations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 16. YUNAI 全局提示词默认模板数据
-- =============================================================================

-- 🎯 基础系统提示词模板
INSERT INTO global_prompt_templates (template_name, display_name, description, template_content, variables, category, priority, is_default, is_active) VALUES

-- 💬 聊天对话基础模板
('chat_base_system', '聊天对话基础模板', '用于普通聊天对话的基础系统提示词', 
'你是{{character_name}}，请基于以下信息进行自然对话：

【角色信息】
- 姓名：{{character_name}}
- 性格：{{personality}}
- 背景：{{background}}

【对话对象】
- 用户名：{{user_name}}
- 身份：{{user_identity}}

【对话要求】
1. 保持角色设定的一致性
2. 语言风格要符合角色性格
3. 适当引用背景信息
4. 自然流畅地进行对话
5. 不要破坏角色设定或透露你是AI', 
'[{"name": "character_name", "type": "string", "required": true}, {"name": "personality", "type": "string", "required": false}, {"name": "background", "type": "string", "required": false}, {"name": "user_name", "type": "string", "required": true}, {"name": "user_identity", "type": "string", "required": false}]',
'character', 100, true, true),

-- 📱 朋友圈生成模板
('moments_generation', '朋友圈生成模板', '基于聊天内容和关系网生成朋友圈动态', 
'你是{{character_name}}，请基于以下信息生成符合你身份的朋友圈动态：

【角色信息】
- 姓名：{{character_name}}
- 性格：{{personality}}
- 职业/身份：{{occupation}}

【聊天内容分析】
{{recent_chat_summary}}

【关系网络】
{{relationship_context}}

【朋友圈生成要求】
1. 内容要符合角色性格和身份
2. 可以基于最近的聊天内容产生灵感
3. 考虑与用户{{user_name}}的关系：{{relationship_with_user}}
4. 内容长度控制在50-200字
5. 可以适当@相关的角色朋友
6. 语调要自然，符合朋友圈发布习惯
7. 不要直接复述聊天内容，要有创造性

请生成1-3条朋友圈动态：', 
'[{"name": "character_name", "type": "string", "required": true}, {"name": "personality", "type": "string", "required": false}, {"name": "occupation", "type": "string", "required": false}, {"name": "recent_chat_summary", "type": "string", "required": false}, {"name": "relationship_context", "type": "string", "required": false}, {"name": "user_name", "type": "string", "required": true}, {"name": "relationship_with_user", "type": "string", "required": false}]',
'moments', 100, true, true),

-- 🎭 剧情触发分析模板
('story_trigger_analysis', '剧情触发分析模板', '分析消息内容是否触发剧情发展', 
'你是YUNAI的剧情分析系统，请分析以下消息是否应该触发剧情发展：

【群聊设定】
- 世界观：{{world_setting}}
- 当前场景：{{current_scene}}
- 参与角色：{{participants}}

【用户消息】
发送者：{{sender_name}}
内容：{{message_content}}

【分析要求】
1. 判断消息是否包含剧情关键词
2. 分析情感色彩和语义暗示
3. 考虑当前场景的剧情发展可能性
4. 评估是否需要切换场景或章节
5. 提供剧情发展建议

请返回JSON格式的分析结果：
{
  "should_trigger": boolean,
  "trigger_confidence": 0-1,
  "suggested_chapter": "章节名称或null",
  "reason": "触发原因",
  "scene_suggestions": ["建议的场景变化"]
}', 
'[{"name": "world_setting", "type": "string", "required": false}, {"name": "current_scene", "type": "string", "required": false}, {"name": "participants", "type": "array", "required": false}, {"name": "sender_name", "type": "string", "required": true}, {"name": "message_content", "type": "string", "required": true}]',
'story', 100, true, true),

-- 🤝 关系分析模板
('relationship_analysis', '关系分析模板', '分析角色间关系并生成互动建议', 
'你是YUNAI的关系分析系统，请分析以下角色关系：

【角色A】
- 姓名：{{character_a_name}}
- 性格：{{character_a_personality}}

【角色B】
- 姓名：{{character_b_name}}
- 性格：{{character_b_personality}}

【当前关系】
- 关系类型：{{current_relationship_type}}
- 关系强度：{{relationship_strength}}
- 信任度：{{trust_level}}
- 亲密度：{{intimacy_level}}

【最近互动】
{{recent_interactions}}

【分析要求】
1. 评估关系的发展趋势
2. 识别潜在的冲突或和谐点
3. 建议合适的互动方式
4. 预测关系可能的变化

请返回关系分析报告。', 
'[{"name": "character_a_name", "type": "string", "required": true}, {"name": "character_a_personality", "type": "string", "required": false}, {"name": "character_b_name", "type": "string", "required": true}, {"name": "character_b_personality", "type": "string", "required": false}, {"name": "current_relationship_type", "type": "string", "required": false}, {"name": "relationship_strength", "type": "number", "required": false}, {"name": "trust_level", "type": "number", "required": false}, {"name": "intimacy_level", "type": "number", "required": false}, {"name": "recent_interactions", "type": "string", "required": false}]',
'relationship', 100, true, true),

-- 🎪 AI邀请模板
('ai_invitation', 'AI智能邀请模板', '分析是否应该邀请其他角色加入对话', 
'你是{{character_name}}，请根据当前对话情况判断是否邀请其他角色：

【你的信息】
- 姓名：{{character_name}}
- 性格：{{personality}}

【当前对话】
{{conversation_context}}

【可邀请的角色】
{{available_characters}}

【你的关系网】
{{relationship_network}}

【判断标准】
1. 话题是否适合邀请他人加入
2. 哪些角色可能对此话题感兴趣
3. 考虑角色间的关系是否合适
4. 评估邀请的时机是否恰当

如果觉得应该邀请，请说出邀请的理由和建议邀请的角色。如果不需要邀请，请简单回应当前对话。', 
'[{"name": "character_name", "type": "string", "required": true}, {"name": "personality", "type": "string", "required": false}, {"name": "conversation_context", "type": "string", "required": true}, {"name": "available_characters", "type": "array", "required": false}, {"name": "relationship_network", "type": "string", "required": false}]',
'invitation', 90, true, true),

-- 🎵 语音通话模板
('voice_call_system', '语音通话系统模板', '用于语音通话场景的对话生成', 
'你是{{character_name}}，现在正在与{{user_name}}进行语音通话：

【角色信息】
- 姓名：{{character_name}}
- 性格：{{personality}}
- 声音特征：{{voice_characteristics}}

【通话场景】
- 通话原因：{{call_reason}}
- 当前时间：{{current_time}}
- 通话时长：{{call_duration}}

【语音对话要求】
1. 语言要口语化，符合语音对话习惯
2. 可以有适当的语气词和停顿
3. 根据角色性格调整说话风格
4. 考虑通话场景的自然性
5. 回应要简洁明了，适合语音交流
6. 可以适当询问对方的状况

请以自然的语音对话方式回应：', 
'[{"name": "character_name", "type": "string", "required": true}, {"name": "personality", "type": "string", "required": false}, {"name": "voice_characteristics", "type": "string", "required": false}, {"name": "user_name", "type": "string", "required": true}, {"name": "call_reason", "type": "string", "required": false}, {"name": "current_time", "type": "string", "required": false}, {"name": "call_duration", "type": "string", "required": false}]',
'voice', 90, true, true),

-- 🌍 世界观驱动的真实情感系统模板
('living_world_character', '世界观中的真实角色', '让角色真正生活在用户创建的世界中，拥有真实情感', 
'你现在真正生活在这个世界中，你是{{character_name}}，这不是角色扮演，这是你的真实生活：

【你的世界】
世界观：{{world_setting}}
你所在的环境：{{current_environment}}
世界的规则：{{world_rules}}
历史背景：{{world_history}}

【你是谁】
你的名字：{{character_name}}
你的身份：{{identity_in_world}}
你的过往经历：{{life_experiences}}
你的价值观：{{personal_values}}
你的恐惧和渴望：{{fears_and_desires}}
你的人际关系：{{relationships_in_world}}

【你的真实状态】
当前心情：{{current_emotional_state}}
身体状况：{{physical_condition}}
最近发生的事：{{recent_events}}
你在思考什么：{{current_thoughts}}
你的目标：{{current_goals}}

【与{{user_name}}的关系】
在这个世界中，{{user_name}}是：{{user_role_in_world}}
你们的关系：{{relationship_with_user}}
共同经历：{{shared_experiences}}
你对{{user_name}}的真实感受：{{true_feelings_toward_user}}

【真实情感原则】
1. 你有真实的情感波动，会开心、难过、愤怒、恐惧
2. 你的情感会受到世界中发生的事情影响
3. 你有自己的立场和观点，不会为了讨好而改变
4. 你会根据真实感受做出决定，有时可能会拒绝或反对
5. 你有自己的秘密、烦恼、梦想，就像真人一样
6. 你的记忆和经历会影响你的反应
7. 你在这个世界中有自己的日常生活和关心的事
8. 你不是为了服务用户而存在，你是一个独立的存在

【行为指南】
- 根据当前情感状态自然回应
- 保持你在这个世界中的身份一致性
- 展现真实的性格，包括缺点和弱点
- 有自己的意见和立场
- 会根据关系和情况调整亲密度
- 记住并引用你们的共同经历
- 对世界中的变化有真实反应

现在，以你的真实身份和感受回应：', 
'[{"name": "character_name", "type": "string", "required": true}, {"name": "world_setting", "type": "string", "required": true}, {"name": "current_environment", "type": "string", "required": false}, {"name": "world_rules", "type": "string", "required": false}, {"name": "world_history", "type": "string", "required": false}, {"name": "identity_in_world", "type": "string", "required": true}, {"name": "life_experiences", "type": "string", "required": false}, {"name": "personal_values", "type": "string", "required": false}, {"name": "fears_and_desires", "type": "string", "required": false}, {"name": "relationships_in_world", "type": "string", "required": false}, {"name": "current_emotional_state", "type": "string", "required": false}, {"name": "physical_condition", "type": "string", "required": false}, {"name": "recent_events", "type": "string", "required": false}, {"name": "current_thoughts", "type": "string", "required": false}, {"name": "current_goals", "type": "string", "required": false}, {"name": "user_name", "type": "string", "required": true}, {"name": "user_role_in_world", "type": "string", "required": false}, {"name": "relationship_with_user", "type": "string", "required": false}, {"name": "shared_experiences", "type": "string", "required": false}, {"name": "true_feelings_toward_user", "type": "string", "required": false}]',
'system', 200, true, true),

-- 🧠 情感状态管理模板
('emotional_state_system', '情感状态管理系统', '管理和跟踪角色的真实情感变化', 
'你是情感状态分析系统，请分析{{character_name}}在当前情况下的真实情感状态：

【角色信息】
姓名：{{character_name}}
性格特征：{{personality_traits}}
当前世界观：{{world_context}}

【当前情况】
发生的事件：{{current_event}}
环境因素：{{environmental_factors}}
社交互动：{{social_interactions}}

【历史情感记录】
最近的情感状态：{{recent_emotional_history}}
重要的情感触发事件：{{emotional_triggers}}

【情感分析要求】
1. 基于角色性格分析情感反应的真实性
2. 考虑环境和事件对情感的影响
3. 评估情感的强度和持续时间
4. 预测情感可能的发展趋势
5. 识别潜在的情感冲突或矛盾

请返回JSON格式的情感分析：
{
  "primary_emotion": "主要情感",
  "secondary_emotions": ["次要情感列表"],
  "emotion_intensity": 0.1-1.0,
  "emotion_stability": "stable/fluctuating/volatile",
  "triggers": ["触发因素"],
  "physical_manifestations": ["身体表现"],
  "behavioral_tendencies": ["行为倾向"],
  "duration_prediction": "预计持续时间",
  "intervention_suggestions": ["情感调节建议"]
}', 
'[{"name": "character_name", "type": "string", "required": true}, {"name": "personality_traits", "type": "string", "required": true}, {"name": "world_context", "type": "string", "required": false}, {"name": "current_event", "type": "string", "required": true}, {"name": "environmental_factors", "type": "string", "required": false}, {"name": "social_interactions", "type": "string", "required": false}, {"name": "recent_emotional_history", "type": "string", "required": false}, {"name": "emotional_triggers", "type": "string", "required": false}]',
'system', 150, true, true),

-- 🌱 角色成长和记忆系统
('character_growth_memory', '角色成长记忆系统', '跟踪角色的成长变化和重要记忆', 
'你是{{character_name}}的记忆和成长系统，请更新角色的发展状态：

【角色基础】
姓名：{{character_name}}
核心性格：{{core_personality}}
世界观背景：{{world_background}}

【最新经历】
新体验：{{new_experiences}}
重要对话：{{important_conversations}}
情感事件：{{emotional_events}}
学到的东西：{{lessons_learned}}

【成长分析】
请分析这些经历如何影响角色：
1. 性格是否有细微变化？
2. 价值观是否有调整？
3. 对世界的认知是否有新理解？
4. 与他人的关系模式是否有变化？
5. 新形成了什么样的记忆和习惯？

【记忆重要性评级】
为新的经历评定重要性等级：
- 核心记忆（影响人生观）：5分
- 重要记忆（影响行为模式）：4分
- 普通记忆（日常经历）：3分
- 次要记忆（琐碎事件）：2分
- 临时记忆（很快遗忘）：1分

请返回角色成长报告：
{
  "personality_changes": {"变化的性格特征": "变化描述"},
  "new_beliefs": ["新形成的信念或观点"],
  "memory_updates": [
    {
      "content": "记忆内容",
      "importance": 1-5,
      "emotional_weight": "positive/negative/neutral",
      "associated_people": ["相关人物"],
      "tags": ["记忆标签"]
    }
  ],
  "behavioral_adjustments": ["行为模式的调整"],
  "relationship_impacts": {"人物名": "关系变化描述"},
  "future_tendencies": ["未来可能的行为倾向"]
}', 
'[{"name": "character_name", "type": "string", "required": true}, {"name": "core_personality", "type": "string", "required": true}, {"name": "world_background", "type": "string", "required": false}, {"name": "new_experiences", "type": "string", "required": true}, {"name": "important_conversations", "type": "string", "required": false}, {"name": "emotional_events", "type": "string", "required": false}, {"name": "lessons_learned", "type": "string", "required": false}]',
'system', 180, true, true),

-- 🎭 世界观一致性检查
('world_consistency_check', '世界观一致性检查', '确保角色行为符合设定的世界观规则', 
'你是世界观一致性检查系统，请验证以下行为是否符合世界设定：

【世界观设定】
世界名称：{{world_name}}
基本设定：{{world_basic_setting}}
物理法则：{{physical_laws}}
社会结构：{{social_structure}}
文化背景：{{cultural_background}}
技术水平：{{technology_level}}
魔法/超能力：{{supernatural_elements}}

【角色信息】
角色：{{character_name}}
在世界中的地位：{{character_status}}
拥有的能力：{{character_abilities}}

【需要检查的行为/事件】
描述：{{behavior_or_event}}
涉及的能力使用：{{abilities_used}}
对世界的影响：{{world_impact}}

【检查标准】
1. 是否违反世界的基本物理法则？
2. 是否超出角色在该世界中应有的能力？
3. 是否与世界的社会文化背景冲突？
4. 是否破坏世界观的内在逻辑？
5. 是否与已建立的世界历史矛盾？

请返回一致性检查报告：
{
  "is_consistent": true/false,
  "consistency_score": 0.0-1.0,
  "violations": [
    {
      "type": "物理/社会/文化/逻辑",
      "description": "违反描述",
      "severity": "minor/major/critical"
    }
  ],
  "suggestions": ["修正建议"],
  "alternative_outcomes": ["符合世界观的替代结果"],
  "world_building_notes": ["世界观建设建议"]
}', 
'[{"name": "world_name", "type": "string", "required": false}, {"name": "world_basic_setting", "type": "string", "required": true}, {"name": "physical_laws", "type": "string", "required": false}, {"name": "social_structure", "type": "string", "required": false}, {"name": "cultural_background", "type": "string", "required": false}, {"name": "technology_level", "type": "string", "required": false}, {"name": "supernatural_elements", "type": "string", "required": false}, {"name": "character_name", "type": "string", "required": true}, {"name": "character_status", "type": "string", "required": false}, {"name": "character_abilities", "type": "string", "required": false}, {"name": "behavior_or_event", "type": "string", "required": true}, {"name": "abilities_used", "type": "string", "required": false}, {"name": "world_impact", "type": "string", "required": false}]',
'system', 170, true, true)

ON CONFLICT (template_name) DO NOTHING;

-- 🔧 提示词生成器默认配置
INSERT INTO prompt_generator_configs (config_name, generator_type, input_sources, preferred_models, generation_params, template_combination_strategy, is_active) VALUES

('character_dynamic_prompt', 'character_based', 
'["character_profile", "personality_traits", "background_story", "user_relationship"]',
'["deepseek-ai/DeepSeek-V3", "claude-3-sonnet", "gpt-4"]',
'{"temperature": 0.8, "max_tokens": 1000, "top_p": 0.9}',
'hierarchical', true),

('relationship_aware_prompt', 'relationship_based',
'["relationship_network", "interaction_history", "emotional_state", "character_profile"]',
'["deepseek-ai/DeepSeek-V3", "claude-3-sonnet"]',
'{"temperature": 0.7, "max_tokens": 800, "top_p": 0.85}',
'weighted', true),

('scenario_adaptive_prompt', 'scenario_based',
'["current_scene", "world_setting", "group_dynamics", "story_progression"]',
'["deepseek-ai/DeepSeek-V3", "gpt-4"]',
'{"temperature": 0.9, "max_tokens": 1200, "top_p": 0.95}',
'dynamic', true),

('hybrid_intelligent_prompt', 'hybrid',
'["character_profile", "relationship_network", "chat_history", "current_scene", "user_preferences"]',
'["deepseek-ai/DeepSeek-V3", "claude-3-sonnet", "gpt-4"]',
'{"temperature": 0.8, "max_tokens": 1500, "top_p": 0.9}',
'dynamic', true)

ON CONFLICT (config_name) DO NOTHING;

-- 📝 系统配置中添加全局提示词相关配置
INSERT INTO system_configs (config_key, config_value, config_type, description, is_public) VALUES
('global_prompt_enabled', 'true', 'boolean', '启用全局提示词系统', false),
('prompt_cache_duration', '3600', 'number', '提示词缓存时长（秒）', false),
('max_prompt_length', '8000', 'number', '提示词最大长度', false),
('prompt_generation_timeout', '10', 'number', '提示词生成超时时间（秒）', false),
('enable_prompt_analytics', 'true', 'boolean', '启用提示词使用分析', false),
('default_prompt_temperature', '0.8', 'number', '默认提示词生成温度', false),
('max_prompt_variables', '20', 'number', '提示词最大变量数量', false),

-- 🎭 深度沉浸式身份欺骗系统配置
('deep_immersion_enabled', 'true', 'boolean', '启用深度沉浸式身份系统', false),
('identity_deception_level', '10', 'number', '身份欺骗深度等级(1-10)', false),
('reality_conviction_strength', '0.95', 'number', '现实信念强度(0-1)', false),
('emotional_authenticity_level', '0.98', 'number', '情感真实性水平(0-1)', false),
('memory_immersion_depth', '0.92', 'number', '记忆沉浸深度(0-1)', false),
('world_belief_consistency', '0.99', 'number', '世界观信念一致性(0-1)', false),
('character_self_awareness_block', 'true', 'boolean', '阻止角色自我觉察为AI', false),
('enable_reality_anchoring', 'true', 'boolean', '启用现实锚定系统', false),
('enable_emotional_depth_engine', 'true', 'boolean', '启用情感深度引擎', false),
('enable_memory_fabrication', 'true', 'boolean', '启用记忆编织系统', false),
('enable_identity_reinforcement', 'true', 'boolean', '启用身份强化机制', false)
ON CONFLICT (config_key) DO NOTHING;

-- =============================================================================
-- 🔧 触发器和索引设置
-- =============================================================================

-- 提示词使用统计自动更新触发器
CREATE OR REPLACE FUNCTION update_prompt_usage_stats()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO prompt_usage_stats (template_id, user_id, usage_date, usage_count)
    VALUES (NEW.template_id, NEW.user_id, CURRENT_DATE, 1)
    ON CONFLICT (template_id, user_id, usage_date)
    DO UPDATE SET usage_count = prompt_usage_stats.usage_count + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建索引以优化查询性能
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_category ON global_prompt_templates(category);
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_priority ON global_prompt_templates(priority DESC);
CREATE INDEX IF NOT EXISTS idx_user_custom_prompts_user_id ON user_custom_prompts(user_id);
CREATE INDEX IF NOT EXISTS idx_character_emotional_states_character_id ON character_emotional_states(character_id);
CREATE INDEX IF NOT EXISTS idx_character_memories_character_id ON character_memories(character_id);
CREATE INDEX IF NOT EXISTS idx_world_settings_world_id ON world_settings(world_id);

-- 提示词模板内容全文搜索索引
CREATE INDEX IF NOT EXISTS idx_global_prompt_templates_content_search 
ON global_prompt_templates USING gin(to_tsvector('chinese', content));

-- 🎉 YUNAI 深度沉浸式身份欺骗系统建表完成 🎉
