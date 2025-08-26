-- Create AI models table
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_key VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    model_type VARCHAR(50) NOT NULL, -- chat, image, video, audio, etc.
    
    -- Capabilities
    capabilities JSONB NOT NULL DEFAULT '[]',
    
    -- Parameters schema
    params_schema JSONB NOT NULL DEFAULT '{}',
    
    -- System configuration
    model_system_prompt TEXT,
    base_url TEXT,
    api_key_encrypted TEXT,
    
    -- Pricing configuration
    pricing JSONB NOT NULL DEFAULT '{}',
    
    -- Visibility and permissions
    visibility VARCHAR(20) NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'vip_only', 'hidden')),
    min_user_type VARCHAR(20) NOT NULL DEFAULT 'basic' CHECK (min_user_type IN ('basic', 'vip', 'creator', 'admin')),
    min_balance DECIMAL(15,2) DEFAULT 0.00,
    daily_limit INTEGER DEFAULT -1, -- -1 means unlimited
    user_limit INTEGER DEFAULT -1,
    
    -- Health and routing
    health_status VARCHAR(20) NOT NULL DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    weight INTEGER DEFAULT 100,
    fallback_chain JSONB DEFAULT '[]',
    connectivity_test_endpoint TEXT,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_health_check TIMESTAMP WITH TIME ZONE
);

-- Create model capabilities table for better querying
CREATE TABLE IF NOT EXISTS model_capabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    capability VARCHAR(50) NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create model health logs table
CREATE TABLE IF NOT EXISTS model_health_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    response_time_ms INTEGER,
    error_message TEXT,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create model usage statistics table
CREATE TABLE IF NOT EXISTS model_usage_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    request_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,
    total_cost DECIMAL(15,4) DEFAULT 0.0000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(model_id, user_id, date)
);

-- Indexes for ai_models
CREATE INDEX IF NOT EXISTS idx_ai_models_provider ON ai_models(provider);
CREATE INDEX IF NOT EXISTS idx_ai_models_model_type ON ai_models(model_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_visibility ON ai_models(visibility);
CREATE INDEX IF NOT EXISTS idx_ai_models_is_active ON ai_models(is_active);
CREATE INDEX IF NOT EXISTS idx_ai_models_health_status ON ai_models(health_status);

-- Indexes for model_capabilities
CREATE INDEX IF NOT EXISTS idx_model_capabilities_model_id ON model_capabilities(model_id);
CREATE INDEX IF NOT EXISTS idx_model_capabilities_capability ON model_capabilities(capability);

-- Indexes for model_health_logs
CREATE INDEX IF NOT EXISTS idx_model_health_logs_model_id ON model_health_logs(model_id);
CREATE INDEX IF NOT EXISTS idx_model_health_logs_checked_at ON model_health_logs(checked_at);

-- Indexes for model_usage_stats
CREATE INDEX IF NOT EXISTS idx_model_usage_stats_model_id ON model_usage_stats(model_id);
CREATE INDEX IF NOT EXISTS idx_model_usage_stats_user_id ON model_usage_stats(user_id);
CREATE INDEX IF NOT EXISTS idx_model_usage_stats_date ON model_usage_stats(date);

-- Update trigger for ai_models
CREATE TRIGGER update_ai_models_updated_at 
    BEFORE UPDATE ON ai_models 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Update trigger for model_usage_stats
CREATE TRIGGER update_model_usage_stats_updated_at
    BEFORE UPDATE ON model_usage_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create model price history table
CREATE TABLE IF NOT EXISTS model_price_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,
    pricing_data JSONB NOT NULL,
    reason VARCHAR(200) NOT NULL,
    operator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for model_price_history
CREATE INDEX IF NOT EXISTS idx_model_price_history_model_id ON model_price_history(model_id);
CREATE INDEX IF NOT EXISTS idx_model_price_history_created_at ON model_price_history(created_at);

-- Insert default models
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, params_schema, pricing, visibility, min_user_type) VALUES
('gpt-3.5-turbo', 'GPT-3.5 Turbo', 'openai', 'chat', 
 '["chat", "text_generation"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 1000, "min": 1, "max": 4000}}',
 '{"input_token": 0.0015, "output_token": 0.002, "currency": "USD", "unit": "1k_tokens"}',
 'public', 'basic'),

('gpt-4', 'GPT-4', 'openai', 'chat',
 '["chat", "text_generation", "code_generation"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 1000, "min": 1, "max": 8000}}',
 '{"input_token": 0.03, "output_token": 0.06, "currency": "USD", "unit": "1k_tokens"}',
 'public', 'vip'),

('dall-e-3', 'DALL-E 3', 'openai', 'image',
 '["txt2img", "image_generation"]',
 '{"size": {"type": "string", "default": "1024x1024", "enum": ["1024x1024", "1792x1024", "1024x1792"]}, "quality": {"type": "string", "default": "standard", "enum": ["standard", "hd"]}}',
 '{"standard_1024": 0.04, "hd_1024": 0.08, "standard_1792": 0.08, "hd_1792": 0.12, "currency": "USD", "unit": "image"}',
 'public', 'vip')

ON CONFLICT (internal_key) DO NOTHING;
