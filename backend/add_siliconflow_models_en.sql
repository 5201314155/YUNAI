-- Quick add SiliconFlow models to database (English version)

-- Add missing fields
ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS max_tokens INTEGER DEFAULT 8192;

-- Clean old SiliconFlow models
DELETE FROM ai_models WHERE provider = 'siliconflow';

-- Insert complete SiliconFlow model list
INSERT INTO ai_models (
    internal_key, display_name, provider, model_type, category, description,
    capabilities, params_schema, model_system_prompt, pricing, max_tokens,
    support_streaming, is_active, is_featured, weight, created_at, updated_at
) VALUES

-- Chat Models
('sf_deepseek_v3', 'DeepSeek V3', 'siliconflow', 'chat', 'DeepSeek', 'DeepSeek V3 super reasoning model, excellent at logic and code generation',
 '["chat", "reasoning", "code_generation", "analysis"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 'You are DeepSeek V3 AI assistant, excellent at logic reasoning and code generation. Please provide accurate and helpful answers.',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 100, NOW(), NOW()),

('sf_deepseek_r1', 'DeepSeek R1', 'siliconflow', 'chat', 'DeepSeek', 'DeepSeek R1 reasoning model, focused on complex reasoning tasks',
 '["chat", "reasoning", "complex_analysis", "problem_solving"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 'You are DeepSeek R1 reasoning model, focused on complex reasoning and problem solving.',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 95, NOW(), NOW()),

('sf_qwen25_72b', 'Qwen2.5 72B', 'siliconflow', 'chat', 'Alibaba Qwen', 'Alibaba Qwen2.5 72B flagship model, excellent Chinese understanding',
 '["chat", "chinese", "knowledge", "reasoning", "creative_writing"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 'You are Qwen2.5, developed by Alibaba. Please answer in Chinese and provide accurate, helpful information.',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 90, NOW(), NOW()),

('sf_glm4_9b', 'GLM-4 9B', 'siliconflow', 'chat', 'Tsinghua GLM', 'Tsinghua GLM-4 9B model, Chinese-English bilingual conversation expert',
 '["chat", "chinese", "role_play", "creative_writing", "emotion"]',
 '{"temperature": {"type": "number", "default": 0.8, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 3000, "min": 1, "max": 6000}, "stream": {"type": "boolean", "default": false}}',
 'You are GLM-4 AI assistant, developed by Tsinghua University. Please provide professional and accurate answers.',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 6000, true, true, true, 85, NOW(), NOW()),

('sf_qwen3_235b', 'Qwen3 235B', 'siliconflow', 'chat', 'Alibaba Qwen', 'Qwen3 235B ultra-large model, strongest Chinese understanding',
 '["chat", "chinese", "knowledge", "reasoning", "complex_analysis"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 'You are Qwen3, with the strongest Chinese understanding and reasoning capabilities.',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 98, NOW(), NOW()),

('sf_kimi_k2', 'Kimi K2', 'siliconflow', 'chat', 'Moonshot AI', 'Kimi K2 long-context processing expert, supports ultra-long context',
 '["chat", "long_context", "document_analysis", "summarization"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 200000}, "stream": {"type": "boolean", "default": false}}',
 'You are Kimi K2, developed by Moonshot AI, expert in long-text processing, document analysis and summarization.',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 200000, true, true, true, 88, NOW(), NOW()),

-- Embedding Models
('sf_bge_large_zh', 'BGE Large ZH', 'siliconflow', 'embedding', 'Embedding Models', 'BGE Chinese large model, professional Chinese semantic understanding',
 '["embedding", "chinese", "semantic_search", "similarity"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'BGE Chinese semantic understanding embedding model for Chinese text vectorization',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_bge_large_en', 'BGE Large EN', 'siliconflow', 'embedding', 'Embedding Models', 'BGE English large model, professional English semantic understanding',
 '["embedding", "english", "semantic_search", "similarity"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'BGE English semantic understanding embedding model for English text vectorization',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 95, NOW(), NOW()),

('sf_bge_m3', 'BGE M3', 'siliconflow', 'embedding', 'Embedding Models', 'BGE M3 multilingual embedding model, supports cross-lingual retrieval',
 '["embedding", "multilingual", "cross_lingual", "semantic_search"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'BGE M3 multilingual embedding model, supports Chinese, English and other languages',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 90, NOW(), NOW()),

('sf_qwen3_embedding', 'Qwen3 Embedding', 'siliconflow', 'embedding', 'Alibaba Qwen', 'Qwen3 embedding model, Alibaba latest embedding technology',
 '["embedding", "chinese", "semantic_search", "text_analysis"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'Qwen3 embedding model, provides high-quality text vectorization services',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 85, NOW(), NOW()),

-- Reranking Models
('sf_bge_reranker_v2', 'BGE Reranker V2', 'siliconflow', 'reranking', 'Reranking Models', 'BGE reranking model V2, optimizes search result relevance',
 '["reranking", "search_optimization", "relevance_scoring"]',
 '{"top_k": {"type": "integer", "default": 10}, "threshold": {"type": "number", "default": 0.5}}',
 'BGE reranking model for optimizing search result relevance ranking',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 80, NOW(), NOW()),

('sf_qwen3_reranker', 'Qwen3 Reranker', 'siliconflow', 'reranking', 'Alibaba Qwen', 'Qwen3 reranking model, intelligent relevance scoring',
 '["reranking", "search_optimization", "relevance_scoring", "chinese"]',
 '{"top_k": {"type": "integer", "default": 10}, "threshold": {"type": "number", "default": 0.5}}',
 'Qwen3 reranking model, specialized for Chinese search result optimization',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 75, NOW(), NOW()),

-- Image Generation Models
('sf_flux_schnell', 'FLUX.1 Schnell', 'siliconflow', 'image', 'Image Generation', 'FLUX.1 Schnell fast image generation, 4-step high-quality images',
 '["image_generation", "text_to_image", "fast", "high_quality"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 4}}',
 'FLUX.1 Schnell fast image generation model for high-quality image creation',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_flux_dev', 'FLUX.1 Dev', 'siliconflow', 'image', 'Image Generation', 'FLUX.1 Dev version, more flexible image generation control',
 '["image_generation", "text_to_image", "flexible", "creative"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}',
 'FLUX.1 Dev version image generation model with more creative control options',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 95, NOW(), NOW()),

('sf_flux_pro', 'FLUX.1 Pro', 'siliconflow', 'image', 'Image Generation', 'FLUX.1 Pro professional version, highest quality image generation',
 '["image_generation", "text_to_image", "professional", "ultra_quality"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 25}}',
 'FLUX.1 Pro professional version, provides highest quality image generation services',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 98, NOW(), NOW()),

('sf_stable_diffusion_xl', 'Stable Diffusion XL', 'siliconflow', 'image', 'Image Generation', 'Stable Diffusion XL classic image generation model',
 '["image_generation", "text_to_image", "artistic", "versatile"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}',
 'Stable Diffusion XL classic image generation model, supports various artistic styles',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 85, NOW(), NOW()),

('sf_kolors', 'Kolors', 'siliconflow', 'image', 'Image Generation', 'Kolors Kuaishou image generation model, excellent Chinese understanding',
 '["image_generation", "text_to_image", "chinese", "creative"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}',
 'Kolors Kuaishou image generation model, better understanding of Chinese prompts',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 80, NOW(), NOW()),

-- TTS Models (Text-to-Speech)
('sf_fish_speech_15', 'Fish Speech 1.5', 'siliconflow', 'audio', 'Voice Synthesis', 'Fish Speech 1.5 voice synthesis, natural and fluent speech generation',
 '["tts", "voice_synthesis", "natural", "multilingual"]',
 '{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}, "pitch": {"type": "number", "default": 1.0}}',
 'Fish Speech 1.5 voice synthesis model, generates natural and fluent speech',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_fish_speech_14', 'Fish Speech 1.4', 'siliconflow', 'audio', 'Voice Synthesis', 'Fish Speech 1.4 voice synthesis, stable speech generation service',
 '["tts", "voice_synthesis", "stable", "multilingual"]',
 '{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}, "pitch": {"type": "number", "default": 1.0}}',
 'Fish Speech 1.4 voice synthesis model, provides stable speech generation services',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 95, NOW(), NOW()),

('sf_cosyvoice', 'CosyVoice', 'siliconflow', 'audio', 'Voice Synthesis', 'CosyVoice Alibaba voice synthesis, supports multiple voice tones',
 '["tts", "voice_synthesis", "multi_voice", "chinese"]',
 '{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}, "emotion": {"type": "string", "default": "neutral"}}',
 'CosyVoice Alibaba voice synthesis model, supports multiple voice tones and emotional expressions',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 90, NOW(), NOW()),

-- ASR Models (Automatic Speech Recognition)
('sf_sensevoice', 'SenseVoice', 'siliconflow', 'audio', 'Speech Recognition', 'SenseVoice speech recognition, supports multilingual real-time recognition',
 '["asr", "speech_to_text", "multilingual", "real_time"]',
 '{"language": {"type": "string", "default": "auto"}, "format": {"type": "string", "default": "wav"}}',
 'SenseVoice speech recognition model, supports multilingual speech-to-text conversion',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 85, NOW(), NOW()),

-- Video Generation Models
('sf_wan_t2v_14b', 'Wan2.1 T2V 14B', 'siliconflow', 'video', 'Video Generation', 'Wan AI text-to-video 14B model, creative video generation',
 '["video_generation", "text_to_video", "creative", "storytelling"]',
 '{"duration": {"type": "integer", "default": 5}, "fps": {"type": "integer", "default": 24}, "resolution": {"type": "string", "default": "720p"}}',
 'Wan AI text-to-video model for creative video content generation',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_wan_i2v_14b', 'Wan2.1 I2V 14B', 'siliconflow', 'video', 'Video Generation', 'Wan AI image-to-video 14B model, image animation',
 '["video_generation", "image_to_video", "animation", "motion"]',
 '{"duration": {"type": "integer", "default": 5}, "fps": {"type": "integer", "default": 24}, "resolution": {"type": "string", "default": "720p"}}',
 'Wan AI image-to-video model, converts static images to dynamic videos',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 95, NOW(), NOW())

ON CONFLICT (internal_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    category = EXCLUDED.category,
    description = EXCLUDED.description,
    capabilities = EXCLUDED.capabilities,
    params_schema = EXCLUDED.params_schema,
    model_system_prompt = EXCLUDED.model_system_prompt,
    pricing = EXCLUDED.pricing,
    max_tokens = EXCLUDED.max_tokens,
    support_streaming = EXCLUDED.support_streaming,
    is_active = EXCLUDED.is_active,
    is_featured = EXCLUDED.is_featured,
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- Show results
SELECT 
    category,
    model_type,
    COUNT(*) as model_count,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_count
FROM ai_models 
WHERE provider = 'siliconflow'
GROUP BY category, model_type
ORDER BY category, model_count DESC;
