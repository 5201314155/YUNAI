-- 创建功能开关表
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_key VARCHAR(100) UNIQUE NOT NULL, -- 功能开关键名
    flag_name VARCHAR(200) NOT NULL, -- 功能名称
    description TEXT, -- 功能描述
    
    -- 开关状态
    enabled BOOLEAN DEFAULT FALSE,
    
    -- 权限控制
    min_user_type VARCHAR(20) DEFAULT 'basic' CHECK (min_user_type IN ('basic', 'vip', 'creator', 'admin')),
    
    -- 灰度发布
    rollout_percentage INTEGER DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    rollout_strategy VARCHAR(20) DEFAULT 'user_id' CHECK (rollout_strategy IN ('user_id', 'region', 'device', 'random')),
    
    -- 环境控制
    environments JSONB DEFAULT '["dev", "stg", "prod"]', -- 启用的环境
    
    -- 配置参数
    config_params JSONB DEFAULT '{}', -- 额外的配置参数
    
    -- 优先级（数字越大优先级越高）
    priority INTEGER DEFAULT 0,
    
    -- 创建和更新信息
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 功能开关表索引
CREATE INDEX idx_feature_flags_flag_key ON feature_flags(flag_key);
CREATE INDEX idx_feature_flags_enabled ON feature_flags(enabled);
CREATE INDEX idx_feature_flags_min_user_type ON feature_flags(min_user_type);
CREATE INDEX idx_feature_flags_priority ON feature_flags(priority);

-- 功能开关更新时间触发器
CREATE TRIGGER update_feature_flags_updated_at 
    BEFORE UPDATE ON feature_flags 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 创建模型提供商表
CREATE TABLE IF NOT EXISTS model_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_key VARCHAR(50) UNIQUE NOT NULL, -- 提供商标识
    provider_name VARCHAR(100) NOT NULL, -- 提供商名称
    display_name VARCHAR(100) NOT NULL, -- 显示名称
    
    -- API配置
    api_base_url TEXT NOT NULL,
    auth_type VARCHAR(20) NOT NULL CHECK (auth_type IN ('bearer', 'api_key', 'basic', 'custom')),
    auth_header VARCHAR(50) DEFAULT 'Authorization',
    
    -- 限流配置
    rate_limit_config JSONB DEFAULT '{}', -- 限流配置
    
    -- 健康检查
    health_check_url TEXT,
    health_check_interval INTEGER DEFAULT 300, -- 秒
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 模型提供商表索引
CREATE INDEX idx_model_providers_provider_key ON model_providers(provider_key);
CREATE INDEX idx_model_providers_is_active ON model_providers(is_active);
CREATE INDEX idx_model_providers_health_status ON model_providers(health_status);

-- 模型提供商更新时间触发器
CREATE TRIGGER update_model_providers_updated_at 
    BEFORE UPDATE ON model_providers 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 创建AI模型表
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_key VARCHAR(100) UNIQUE NOT NULL, -- 模型标识
    internal_key VARCHAR(100) NOT NULL, -- 内部标识
    display_name VARCHAR(200) NOT NULL, -- 显示名称
    
    -- 提供商信息
    provider_id UUID NOT NULL REFERENCES model_providers(id) ON DELETE CASCADE,
    
    -- 能力配置
    capabilities JSONB NOT NULL DEFAULT '[]', -- 能力列表：chat, txt2img, img2img等
    
    -- 参数配置
    params_schema JSONB DEFAULT '{}', -- JSON Schema格式的参数定义
    default_params JSONB DEFAULT '{}', -- 默认参数
    
    -- 系统提示词
    system_prompt TEXT,
    
    -- 定价配置
    pricing_config JSONB NOT NULL DEFAULT '{}', -- 定价信息
    
    -- 权限控制
    visibility VARCHAR(20) DEFAULT 'public' CHECK (visibility IN ('public', 'vip_only', 'creator_only', 'admin_only', 'hidden')),
    min_user_type VARCHAR(20) DEFAULT 'basic' CHECK (min_user_type IN ('basic', 'vip', 'creator', 'admin')),
    min_balance DECIMAL(15,2) DEFAULT 0.00,
    daily_limit INTEGER DEFAULT 1000,
    user_limit INTEGER DEFAULT 100,
    
    -- 健康和性能
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    last_health_check TIMESTAMP WITH TIME ZONE,
    avg_response_time INTEGER DEFAULT 0, -- 毫秒
    
    -- 权重和降级
    weight INTEGER DEFAULT 100, -- 权重，用于负载均衡
    fallback_models JSONB DEFAULT '[]', -- 降级模型列表
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- AI模型表索引
CREATE INDEX idx_ai_models_model_key ON ai_models(model_key);
CREATE INDEX idx_ai_models_provider_id ON ai_models(provider_id);
CREATE INDEX idx_ai_models_visibility ON ai_models(visibility);
CREATE INDEX idx_ai_models_min_user_type ON ai_models(min_user_type);
CREATE INDEX idx_ai_models_is_active ON ai_models(is_active);
CREATE INDEX idx_ai_models_health_status ON ai_models(health_status);

-- AI模型更新时间触发器
CREATE TRIGGER update_ai_models_updated_at 
    BEFORE UPDATE ON ai_models 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 插入默认功能开关
INSERT INTO feature_flags (flag_key, flag_name, description, enabled, min_user_type, rollout_percentage, created_by) VALUES
('enable_txt2img', '文生图功能', '文本生成图片功能开关', TRUE, 'basic', 100, (SELECT id FROM users WHERE username = 'admin')),
('enable_img2img', '图生图功能', '图片生成图片功能开关', TRUE, 'vip', 100, (SELECT id FROM users WHERE username = 'admin')),
('enable_img2video', '图生视频功能', '图片生成视频功能开关', TRUE, 'vip', 80, (SELECT id FROM users WHERE username = 'admin')),
('enable_txt2video', '文生视频功能', '文本生成视频功能开关', FALSE, 'creator', 50, (SELECT id FROM users WHERE username = 'admin')),
('enable_remove_bg', '背景移除功能', '图片背景移除功能开关', TRUE, 'basic', 100, (SELECT id FROM users WHERE username = 'admin')),
('enable_talking_head', '虚拟人功能', '虚拟人生成功能开关', FALSE, 'creator', 10, (SELECT id FROM users WHERE username = 'admin')),
('enable_ai_proactive_call', 'AI主动来电', 'AI主动发起通话功能开关', TRUE, 'basic', 70, (SELECT id FROM users WHERE username = 'admin')),
('enable_moments_auto', '朋友圈自动发布', '朋友圈自动发布功能开关', TRUE, 'basic', 100, (SELECT id FROM users WHERE username = 'admin'))
ON CONFLICT (flag_key) DO NOTHING;

-- 插入默认模型提供商
INSERT INTO model_providers (provider_key, provider_name, display_name, api_base_url, auth_type, health_check_url) VALUES
('openai', 'openai', 'OpenAI', 'https://api.openai.com/v1', 'bearer', 'https://api.openai.com/v1/models'),
('stability', 'stability', 'Stability AI', 'https://api.stability.ai/v1', 'bearer', 'https://api.stability.ai/v1/engines/list'),
('elevenlabs', 'elevenlabs', 'ElevenLabs', 'https://api.elevenlabs.io/v1', 'api_key', 'https://api.elevenlabs.io/v1/voices')
ON CONFLICT (provider_key) DO NOTHING;
