-- YUNAI功能模块专用AI模型配置 (简化版)
-- 解决编码问题，使用英文注释

-- 1. Create module model configuration table
CREATE TABLE IF NOT EXISTS module_model_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_name VARCHAR(50) NOT NULL,
    function_type VARCHAR(50) NOT NULL,
    primary_model_id VARCHAR(100) NOT NULL,
    fallback_models JSONB DEFAULT '[]',
    model_params JSONB DEFAULT '{}',
    weight INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(module_name, function_type)
);

-- 2. Insert SiliconFlow core models (32 selected models)
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

-- Chat models for conversation
('deepseek-ai/DeepSeek-V3', 'sf_deepseek_v3', 'DeepSeek V3 - Super Reasoning', 'siliconflow', 'chat',
 '["chat", "reasoning", "role_play", "emotion", "complex_dialogue"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI: You are an AI character with complete personality, real emotions and memories. Please immerse yourself in your character and have natural conversations.',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('THUDM/glm-4-9b-chat', 'sf_glm4_9b', 'GLM-4 9B - Chinese Chat', 'siliconflow', 'chat',
 '["chat", "chinese", "role_play", "creative_writing", "emotion"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI: You are a warm and interesting AI companion, good at Chinese conversation and understanding emotional needs.',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, true),

('Qwen/Qwen2.5-72B-Instruct', 'sf_qwen25_72b', 'Qwen2.5 72B - Flagship', 'siliconflow', 'chat',
 '["chat", "chinese", "knowledge", "reasoning", "creative_writing"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI: You are Qwen flagship model with rich knowledge and strong understanding for deep conversations.',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

-- Embedding models for relationship networks
('BAAI/bge-large-zh-v1.5', 'sf_bge_zh', 'BGE Large ZH - Chinese Embedding', 'siliconflow', 'embedding',
 '["embedding", "chinese", "semantic_search", "similarity"]',
 '{"dimensions": 1024, "normalize": true}',
 'Chinese semantic understanding and similarity calculation embedding model',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('BAAI/bge-large-en-v1.5', 'sf_bge_en', 'BGE Large EN - English Embedding', 'siliconflow', 'embedding',
 '["embedding", "english", "semantic_search", "similarity"]',
 '{"dimensions": 1024, "normalize": true}',
 'English semantic understanding and similarity calculation embedding model',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

-- Image generation models
('black-forest-labs/FLUX.1-schnell', 'sf_flux_schnell', 'FLUX.1 Schnell - Fast Generation', 'siliconflow', 'image',
 '["image_generation", "fast", "high_quality", "artistic"]',
 '{"width": 1024, "height": 1024, "steps": 4, "guidance_scale": 3.5}',
 'Fast high-quality image generation for character avatars and scenes',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

-- Audio models
('fishaudio/fish-speech-1.5', 'sf_fish15', 'Fish Speech 1.5', 'siliconflow', 'audio',
 '["tts", "voice_synthesis", "natural", "multilingual"]',
 '{"voice_id": "default", "speed": 1.0, "pitch": 1.0}',
 'Fish Audio voice synthesis v1.5 for natural speech generation',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

-- Video generation models
('Wan-AI/Wan2.1-T2V-14B', 'sf_wan_t2v', 'Wan2.1 T2V 14B', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "creative", "storytelling"]',
 '{"duration": 5, "fps": 24, "resolution": "720p"}',
 'Wan AI text-to-video generation for creative video content',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true)

ON CONFLICT (model_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    capabilities = EXCLUDED.capabilities,
    default_params = EXCLUDED.default_params,
    model_system_prompt = EXCLUDED.model_system_prompt,
    pricing = EXCLUDED.pricing,
    weight = EXCLUDED.weight,
    is_active = EXCLUDED.is_active,
    is_featured = EXCLUDED.is_featured,
    updated_at = NOW();

-- 3. Configure module-specific models
INSERT INTO module_model_configs (module_name, function_type, primary_model_id, fallback_models, model_params, weight, description) VALUES

-- Chat module
('chat', 'conversation', 'deepseek-ai/DeepSeek-V3', 
 '["THUDM/glm-4-9b-chat", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 100, 'Main chat conversation model with strong reasoning and emotional understanding'),

('chat', 'role_play', 'THUDM/glm-4-9b-chat',
 '["deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 95, 'Role-playing specialized model for emotional expression and personality'),

-- Relationship network module - Enhanced with embedding models
('relationship', 'embedding', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-large-en-v1.5"]',
 '{"dimensions": 1024, "normalize": true, "batch_size": 32}',
 100, 'Relationship network semantic embedding for understanding complex relationships'),

('relationship', 'similarity', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-large-en-v1.5"]',
 '{"dimensions": 1024, "normalize": true, "similarity_threshold": 0.7}',
 95, 'Relationship similarity calculation for discovering relationship patterns'),

('relationship', 'analysis', 'deepseek-ai/DeepSeek-V3',
 '["THUDM/glm-4-9b-chat"]',
 '{"temperature": 0.6, "max_tokens": 3000, "top_p": 0.85}',
 90, 'Relationship analysis and reasoning for understanding relationship dynamics'),

-- Moments module
('moments', 'generation', 'THUDM/glm-4-9b-chat',
 '["deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.8, "max_tokens": 500, "top_p": 0.9}',
 100, 'Moments content generation for creating interesting social content'),

('moments', 'emotion_analysis', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-large-en-v1.5"]',
 '{"dimensions": 1024, "normalize": true}',
 95, 'Moments emotion analysis for understanding user emotional states'),

-- Character creation module
('character', 'avatar_generation', 'black-forest-labs/FLUX.1-schnell',
 '["black-forest-labs/FLUX.1-dev"]',
 '{"width": 1024, "height": 1024, "steps": 4, "guidance_scale": 3.5}',
 100, 'Character avatar generation for creating high-quality avatars quickly'),

('character', 'personality_analysis', 'deepseek-ai/DeepSeek-V3',
 '["THUDM/glm-4-9b-chat"]',
 '{"temperature": 0.6, "max_tokens": 1000, "top_p": 0.85}',
 90, 'Character personality analysis for building complete personality profiles'),

-- Voice module
('voice', 'synthesis', 'fishaudio/fish-speech-1.5',
 '[]',
 '{"voice_id": "default", "speed": 1.0, "pitch": 1.0, "emotion": "neutral"}',
 100, 'Voice synthesis for generating natural character voices'),

-- Video module
('video', 'generation', 'Wan-AI/Wan2.1-T2V-14B',
 '[]',
 '{"duration": 5, "fps": 24, "resolution": "720p", "style": "realistic"}',
 100, 'Video content generation for creating dynamic visual content'),

-- Search module
('search', 'embedding', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-large-en-v1.5"]',
 '{"dimensions": 1024, "normalize": true, "batch_size": 64}',
 100, 'Search semantic embedding for improving search accuracy'),

('search', 'recommendation', 'deepseek-ai/DeepSeek-V3',
 '["THUDM/glm-4-9b-chat", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.7, "max_tokens": 1000, "top_p": 0.9}',
 90, 'Intelligent recommendation based on user preferences')

ON CONFLICT (module_name, function_type) DO UPDATE SET
    primary_model_id = EXCLUDED.primary_model_id,
    fallback_models = EXCLUDED.fallback_models,
    model_params = EXCLUDED.model_params,
    weight = EXCLUDED.weight,
    description = EXCLUDED.description,
    updated_at = NOW();

-- 4. Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_ai_models_type_active ON ai_models(model_type, is_active);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider_weight ON ai_models(provider, weight DESC);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_module ON module_model_configs(module_name);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_function ON module_model_configs(function_type);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_active ON module_model_configs(is_active, weight DESC);

-- 5. Create model selection function
CREATE OR REPLACE FUNCTION get_module_model(
    p_module_name VARCHAR(50),
    p_function_type VARCHAR(50)
) RETURNS TABLE (
    model_id VARCHAR(100),
    model_params JSONB,
    fallback_models JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        mmc.primary_model_id,
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
