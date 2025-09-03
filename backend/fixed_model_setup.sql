-- YUNAI功能模块专用AI模型配置 (修复版)
-- 使用正确的字段名

-- 1. Create module model configuration table
CREATE TABLE IF NOT EXISTS module_model_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_name VARCHAR(50) NOT NULL,
    function_type VARCHAR(50) NOT NULL,
    primary_model_key VARCHAR(100) NOT NULL,  -- 使用internal_key
    fallback_models JSONB DEFAULT '[]',
    model_params JSONB DEFAULT '{}',
    weight INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(module_name, function_type)
);

-- 2. Insert SiliconFlow core models (使用正确的字段名)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, params_schema, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

-- Chat models for conversation
('sf_deepseek_v3', 'DeepSeek V3 - Super Reasoning', 'siliconflow', 'chat',
 '["chat", "reasoning", "role_play", "emotion", "complex_dialogue"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}}',
 'YUNAI: You are an AI character with complete personality, real emotions and memories. Please immerse yourself in your character and have natural conversations.',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('sf_glm4_9b', 'GLM-4 9B - Chinese Chat', 'siliconflow', 'chat',
 '["chat", "chinese", "role_play", "creative_writing", "emotion"]',
 '{"temperature": {"type": "number", "default": 0.8, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 3000, "min": 1, "max": 6000}}',
 'YUNAI: You are a warm and interesting AI companion, good at Chinese conversation and understanding emotional needs.',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, true),

('sf_qwen25_72b', 'Qwen2.5 72B - Flagship', 'siliconflow', 'chat',
 '["chat", "chinese", "knowledge", "reasoning", "creative_writing"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}}',
 'YUNAI: You are Qwen flagship model with rich knowledge and strong understanding for deep conversations.',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

-- Embedding models for relationship networks
('sf_bge_zh', 'BGE Large ZH - Chinese Embedding', 'siliconflow', 'embedding',
 '["embedding", "chinese", "semantic_search", "similarity"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'Chinese semantic understanding and similarity calculation embedding model',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('sf_bge_en', 'BGE Large EN - English Embedding', 'siliconflow', 'embedding',
 '["embedding", "english", "semantic_search", "similarity"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'English semantic understanding and similarity calculation embedding model',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('sf_bge_m3', 'BGE M3 - Multilingual Embedding', 'siliconflow', 'embedding',
 '["embedding", "multilingual", "semantic_search", "cross_lingual"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'Multilingual embedding model for cross-language scenarios',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false),

-- Image generation models
('sf_flux_schnell', 'FLUX.1 Schnell - Fast Generation', 'siliconflow', 'image',
 '["image_generation", "fast", "high_quality", "artistic"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 4}}',
 'Fast high-quality image generation for character avatars and scenes',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('sf_flux_dev', 'FLUX.1 Dev - Development Version', 'siliconflow', 'image',
 '["image_generation", "development", "flexible", "creative"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}',
 'FLUX.1 development version for flexible image generation',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

-- Audio models
('sf_fish15', 'Fish Speech 1.5', 'siliconflow', 'audio',
 '["tts", "voice_synthesis", "natural", "multilingual"]',
 '{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}, "pitch": {"type": "number", "default": 1.0}}',
 'Fish Audio voice synthesis v1.5 for natural speech generation',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

-- Video generation models
('sf_wan_t2v', 'Wan2.1 T2V 14B', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "creative", "storytelling"]',
 '{"duration": {"type": "integer", "default": 5}, "fps": {"type": "integer", "default": 24}, "resolution": {"type": "string", "default": "720p"}}',
 'Wan AI text-to-video generation for creative video content',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true)

ON CONFLICT (internal_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    capabilities = EXCLUDED.capabilities,
    params_schema = EXCLUDED.params_schema,
    model_system_prompt = EXCLUDED.model_system_prompt,
    pricing = EXCLUDED.pricing,
    weight = EXCLUDED.weight,
    is_active = EXCLUDED.is_active,
    is_featured = EXCLUDED.is_featured,
    updated_at = NOW();

-- 3. Configure module-specific models
INSERT INTO module_model_configs (module_name, function_type, primary_model_key, fallback_models, model_params, weight, description) VALUES

-- Chat module
('chat', 'conversation', 'sf_deepseek_v3', 
 '["sf_glm4_9b", "sf_qwen25_72b"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 100, 'Main chat conversation model with strong reasoning and emotional understanding'),

('chat', 'role_play', 'sf_glm4_9b',
 '["sf_deepseek_v3", "sf_qwen25_72b"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 95, 'Role-playing specialized model for emotional expression and personality'),

-- Relationship network module - Enhanced with embedding models
('relationship', 'embedding', 'sf_bge_zh',
 '["sf_bge_en", "sf_bge_m3"]',
 '{"dimensions": 1024, "normalize": true, "batch_size": 32}',
 100, 'Relationship network semantic embedding for understanding complex relationships'),

('relationship', 'similarity', 'sf_bge_zh',
 '["sf_bge_en", "sf_bge_m3"]',
 '{"dimensions": 1024, "normalize": true, "similarity_threshold": 0.7}',
 95, 'Relationship similarity calculation for discovering relationship patterns'),

('relationship', 'analysis', 'sf_deepseek_v3',
 '["sf_glm4_9b"]',
 '{"temperature": 0.6, "max_tokens": 3000, "top_p": 0.85}',
 90, 'Relationship analysis and reasoning for understanding relationship dynamics'),

-- Moments module
('moments', 'generation', 'sf_glm4_9b',
 '["sf_deepseek_v3", "sf_qwen25_72b"]',
 '{"temperature": 0.8, "max_tokens": 500, "top_p": 0.9}',
 100, 'Moments content generation for creating interesting social content'),

('moments', 'emotion_analysis', 'sf_bge_zh',
 '["sf_bge_en", "sf_bge_m3"]',
 '{"dimensions": 1024, "normalize": true}',
 95, 'Moments emotion analysis for understanding user emotional states'),

-- Character creation module
('character', 'avatar_generation', 'sf_flux_schnell',
 '["sf_flux_dev"]',
 '{"width": 1024, "height": 1024, "steps": 4, "guidance_scale": 3.5}',
 100, 'Character avatar generation for creating high-quality avatars quickly'),

('character', 'personality_analysis', 'sf_deepseek_v3',
 '["sf_glm4_9b"]',
 '{"temperature": 0.6, "max_tokens": 1000, "top_p": 0.85}',
 90, 'Character personality analysis for building complete personality profiles'),

-- Voice module
('voice', 'synthesis', 'sf_fish15',
 '[]',
 '{"voice_id": "default", "speed": 1.0, "pitch": 1.0, "emotion": "neutral"}',
 100, 'Voice synthesis for generating natural character voices'),

-- Video module
('video', 'generation', 'sf_wan_t2v',
 '[]',
 '{"duration": 5, "fps": 24, "resolution": "720p", "style": "realistic"}',
 100, 'Video content generation for creating dynamic visual content'),

-- Search module
('search', 'embedding', 'sf_bge_zh',
 '["sf_bge_en", "sf_bge_m3"]',
 '{"dimensions": 1024, "normalize": true, "batch_size": 64}',
 100, 'Search semantic embedding for improving search accuracy'),

('search', 'recommendation', 'sf_deepseek_v3',
 '["sf_glm4_9b", "sf_qwen25_72b"]',
 '{"temperature": 0.7, "max_tokens": 1000, "top_p": 0.9}',
 90, 'Intelligent recommendation based on user preferences')

ON CONFLICT (module_name, function_type) DO UPDATE SET
    primary_model_key = EXCLUDED.primary_model_key,
    fallback_models = EXCLUDED.fallback_models,
    model_params = EXCLUDED.model_params,
    weight = EXCLUDED.weight,
    description = EXCLUDED.description,
    updated_at = NOW();

-- 4. Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_module_model_configs_module ON module_model_configs(module_name);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_function ON module_model_configs(function_type);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_active ON module_model_configs(is_active, weight DESC);

-- 5. Create model selection function
CREATE OR REPLACE FUNCTION get_module_model(
    p_module_name VARCHAR(50),
    p_function_type VARCHAR(50)
) RETURNS TABLE (
    model_key VARCHAR(100),
    model_params JSONB,
    fallback_models JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        mmc.primary_model_key,
        mmc.model_params,
        mmc.fallback_models
    FROM module_model_configs mmc
    WHERE mmc.module_name = p_module_name 
      AND mmc.function_type = p_function_type
      AND mmc.is_active = true
    ORDER BY mmc.weight DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;
