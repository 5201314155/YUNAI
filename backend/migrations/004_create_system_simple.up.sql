-- Create feature flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_key VARCHAR(100) UNIQUE NOT NULL,
    flag_name VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- Flag status
    enabled BOOLEAN DEFAULT FALSE,
    
    -- Permission control
    min_user_type VARCHAR(20) DEFAULT 'basic' CHECK (min_user_type IN ('basic', 'vip', 'creator', 'admin')),
    
    -- Gradual rollout
    rollout_percentage INTEGER DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    rollout_strategy VARCHAR(20) DEFAULT 'user_id' CHECK (rollout_strategy IN ('user_id', 'region', 'device', 'random')),
    
    -- Environment control
    environments JSONB DEFAULT '["dev", "stg", "prod"]',
    
    -- Config parameters
    config_params JSONB DEFAULT '{}',
    
    -- Priority (higher number = higher priority)
    priority INTEGER DEFAULT 0,
    
    -- Creation and update info
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Feature flag indexes
CREATE INDEX IF NOT EXISTS idx_feature_flags_flag_key ON feature_flags(flag_key);
CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled ON feature_flags(enabled);
CREATE INDEX IF NOT EXISTS idx_feature_flags_min_user_type ON feature_flags(min_user_type);
CREATE INDEX IF NOT EXISTS idx_feature_flags_priority ON feature_flags(priority);

-- Feature flag update trigger
CREATE TRIGGER update_feature_flags_updated_at 
    BEFORE UPDATE ON feature_flags 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create model providers table
CREATE TABLE IF NOT EXISTS model_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_key VARCHAR(50) UNIQUE NOT NULL,
    provider_name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    
    -- API config
    api_base_url TEXT NOT NULL,
    auth_type VARCHAR(20) NOT NULL CHECK (auth_type IN ('bearer', 'api_key', 'basic', 'custom')),
    auth_header VARCHAR(50) DEFAULT 'Authorization',
    
    -- Rate limit config
    rate_limit_config JSONB DEFAULT '{}',
    
    -- Health check
    health_check_url TEXT,
    health_check_interval INTEGER DEFAULT 300,
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Model provider indexes
CREATE INDEX IF NOT EXISTS idx_model_providers_provider_key ON model_providers(provider_key);
CREATE INDEX IF NOT EXISTS idx_model_providers_is_active ON model_providers(is_active);
CREATE INDEX IF NOT EXISTS idx_model_providers_health_status ON model_providers(health_status);

-- Model provider update trigger
CREATE TRIGGER update_model_providers_updated_at 
    BEFORE UPDATE ON model_providers 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create AI models table
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_key VARCHAR(100) UNIQUE NOT NULL,
    internal_key VARCHAR(100) NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    
    -- Provider info
    provider_id UUID NOT NULL REFERENCES model_providers(id) ON DELETE CASCADE,
    
    -- Capability config
    capabilities JSONB NOT NULL DEFAULT '[]',
    
    -- Parameter config
    params_schema JSONB DEFAULT '{}',
    default_params JSONB DEFAULT '{}',
    
    -- System prompt
    system_prompt TEXT,
    
    -- Pricing config
    pricing_config JSONB NOT NULL DEFAULT '{}',
    
    -- Permission control
    visibility VARCHAR(20) DEFAULT 'public' CHECK (visibility IN ('public', 'vip_only', 'creator_only', 'admin_only', 'hidden')),
    min_user_type VARCHAR(20) DEFAULT 'basic' CHECK (min_user_type IN ('basic', 'vip', 'creator', 'admin')),
    min_balance DECIMAL(15,2) DEFAULT 0.00,
    daily_limit INTEGER DEFAULT 1000,
    user_limit INTEGER DEFAULT 100,
    
    -- Health and performance
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    last_health_check TIMESTAMP WITH TIME ZONE,
    avg_response_time INTEGER DEFAULT 0,
    
    -- Weight and fallback
    weight INTEGER DEFAULT 100,
    fallback_models JSONB DEFAULT '[]',
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- AI model indexes
CREATE INDEX IF NOT EXISTS idx_ai_models_model_key ON ai_models(model_key);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider_id ON ai_models(provider_id);
CREATE INDEX IF NOT EXISTS idx_ai_models_visibility ON ai_models(visibility);
CREATE INDEX IF NOT EXISTS idx_ai_models_min_user_type ON ai_models(min_user_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_is_active ON ai_models(is_active);
CREATE INDEX IF NOT EXISTS idx_ai_models_health_status ON ai_models(health_status);

-- AI model update trigger
CREATE TRIGGER update_ai_models_updated_at 
    BEFORE UPDATE ON ai_models 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default feature flags
INSERT INTO feature_flags (flag_key, flag_name, description, enabled, min_user_type, rollout_percentage, created_by) VALUES
('enable_txt2img', 'Text to Image', 'Text to image generation feature', TRUE, 'basic', 100, (SELECT id FROM users WHERE username = 'admin')),
('enable_img2img', 'Image to Image', 'Image to image generation feature', TRUE, 'vip', 100, (SELECT id FROM users WHERE username = 'admin')),
('enable_img2video', 'Image to Video', 'Image to video generation feature', TRUE, 'vip', 80, (SELECT id FROM users WHERE username = 'admin')),
('enable_txt2video', 'Text to Video', 'Text to video generation feature', FALSE, 'creator', 50, (SELECT id FROM users WHERE username = 'admin')),
('enable_remove_bg', 'Remove Background', 'Background removal feature', TRUE, 'basic', 100, (SELECT id FROM users WHERE username = 'admin')),
('enable_talking_head', 'Talking Head', 'Virtual human generation feature', FALSE, 'creator', 10, (SELECT id FROM users WHERE username = 'admin')),
('enable_ai_proactive_call', 'AI Proactive Call', 'AI proactive calling feature', TRUE, 'basic', 70, (SELECT id FROM users WHERE username = 'admin')),
('enable_moments_auto', 'Auto Moments', 'Auto moments posting feature', TRUE, 'basic', 100, (SELECT id FROM users WHERE username = 'admin'))
ON CONFLICT (flag_key) DO NOTHING;

-- Insert default model providers
INSERT INTO model_providers (provider_key, provider_name, display_name, api_base_url, auth_type, health_check_url) VALUES
('openai', 'openai', 'OpenAI', 'https://api.openai.com/v1', 'bearer', 'https://api.openai.com/v1/models'),
('stability', 'stability', 'Stability AI', 'https://api.stability.ai/v1', 'bearer', 'https://api.stability.ai/v1/engines/list'),
('elevenlabs', 'elevenlabs', 'ElevenLabs', 'https://api.elevenlabs.io/v1', 'api_key', 'https://api.elevenlabs.io/v1/voices')
ON CONFLICT (provider_key) DO NOTHING;
