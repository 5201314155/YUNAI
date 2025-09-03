-- YUNAI项目 - SiliconFlow模型配置更新
-- 基于获取的102个模型进行完整配置

-- 清理现有SiliconFlow模型配置
DELETE FROM ai_models WHERE provider = 'siliconflow';

-- 💬 对话聊天模型 (YUNAI核心功能 - 17个模型)
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

-- DeepSeek系列 - 推理能力强
('deepseek-ai/DeepSeek-V3', 'siliconflow_deepseek_v3', 'DeepSeek V3 - 超强推理', 'siliconflow', 'chat', 
 '["chat", "reasoning", "role_play", "emotion", "complex_dialogue"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个具有完整人格的AI角色，拥有真实的情感和记忆。请完全沉浸在你的角色设定中，与用户进行自然、真实的对话。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('deepseek-ai/DeepSeek-V3.1', 'siliconflow_deepseek_v31', 'DeepSeek V3.1 - 最新版', 'siliconflow', 'chat',
 '["chat", "reasoning", "role_play", "emotion", "latest"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个具有完整人格的AI角色，拥有真实的情感和记忆。请完全沉浸在你的角色设定中，与用户进行自然、真实的对话。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 98, true, true),

('deepseek-ai/DeepSeek-R1', 'siliconflow_deepseek_r1', 'DeepSeek R1 - 推理专家', 'siliconflow', 'chat',
 '["chat", "reasoning", "analysis", "problem_solving", "step_by_step"]',
 '{"temperature": 0.6, "max_tokens": 4000, "top_p": 0.8}',
 'YUNAI专用：你是一个善于深度思考和推理的AI角色。在对话中展现你的分析能力和逻辑思维。',
 '{"input_token_price": 0.0055, "output_token_price": 0.0055, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

-- 通义千问系列 - 中文优化
('Qwen/Qwen2.5-72B-Instruct', 'siliconflow_qwen25_72b', '通义千问 2.5 72B - 中文优化', 'siliconflow', 'chat',
 '["chat", "chinese_optimized", "role_play", "knowledge", "cultural_understanding"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个精通中文文化的AI角色，能够理解中文的细微差别和文化内涵。与用户进行地道的中文对话。',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, true),

('Qwen/Qwen3-235B-A22B', 'siliconflow_qwen3_235b', '通义千问 3 235B - 超大规模', 'siliconflow', 'chat',
 '["chat", "chinese_optimized", "large_scale", "complex_reasoning", "knowledge_intensive"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个拥有海量知识的AI角色，能够进行复杂的推理和深度对话。',
 '{"input_token_price": 0.002, "output_token_price": 0.006, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

-- GLM系列 - 智谱AI，创意对话
('THUDM/glm-4-9b-chat', 'siliconflow_glm4_9b', 'GLM-4 9B - 智能对话', 'siliconflow', 'chat',
 '["chat", "creative", "role_play", "storytelling", "imagination"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个富有创意和想象力的AI角色，擅长讲故事和创意对话。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 80, true, true),

('zai-org/GLM-4.5', 'siliconflow_glm45', 'GLM-4.5 - 新一代对话', 'siliconflow', 'chat',
 '["chat", "creative", "role_play", "emotion", "empathy"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个情感丰富的AI角色，能够理解和表达复杂的情感。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 75, true, false),

-- Kimi系列 - 长文本记忆
('moonshotai/Kimi-K2-Instruct', 'siliconflow_kimi_k2', 'Kimi K2 - 长文本专家', 'siliconflow', 'chat',
 '["chat", "long_context", "document_analysis", "memory", "relationship_tracking"]',
 '{"temperature": 0.7, "max_tokens": 8000, "top_p": 0.9}',
 'YUNAI专用：你是一个拥有超强记忆力的AI角色，能够记住长时间的对话历史和复杂的背景信息。',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 70, true, false);

-- 🔍 文本嵌入模型 (身份识别和推荐 - 7个模型)
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

('BAAI/bge-large-zh-v1.5', 'siliconflow_bge_zh', 'BGE Large 中文 - 身份识别', 'siliconflow', 'embedding',
 '["embedding", "chinese", "identity_recognition", "similarity", "user_matching"]',
 '{"dimension": 1024, "normalize": true}',
 '用于YUNAI用户身份识别和内容相似度匹配',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('BAAI/bge-m3', 'siliconflow_bge_m3', 'BGE M3 - 多语言嵌入', 'siliconflow', 'embedding',
 '["embedding", "multilingual", "cross_lingual", "retrieval", "semantic_search"]',
 '{"dimension": 1024, "normalize": true}',
 '用于YUNAI多语言内容理解和跨语言检索',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('Qwen/Qwen3-Embedding-8B', 'siliconflow_qwen3_emb8b', '通义千问 嵌入 8B - 语义理解', 'siliconflow', 'embedding',
 '["embedding", "semantic", "context_understanding", "content_analysis"]',
 '{"dimension": 1024, "normalize": true}',
 '用于YUNAI深度语义理解和上下文分析',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false);

-- 🎨 图像生成模型 (角色头像和朋友圈 - 8个模型)
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

('black-forest-labs/FLUX.1-schnell', 'siliconflow_flux_schnell', 'FLUX.1 Schnell - 快速图像生成', 'siliconflow', 'image',
 '["image_generation", "fast", "avatar", "scene", "social_media"]',
 '{"width": 1024, "height": 1024, "steps": 4, "guidance_scale": 3.5}',
 '用于YUNAI快速生成AI角色头像、朋友圈图片和场景背景',
 '{"per_image_price": 0.003, "unit": "per_image", "currency": "CNY"}',
 100, true, true),

('stabilityai/stable-diffusion-3-5-large', 'siliconflow_sd35_large', 'Stable Diffusion 3.5 - 高质量图像', 'siliconflow', 'image',
 '["image_generation", "high_quality", "artistic", "detailed", "professional"]',
 '{"width": 1024, "height": 1024, "steps": 28, "guidance_scale": 7.5}',
 '用于YUNAI生成高质量的AI角色立绘和精美朋友圈图片',
 '{"per_image_price": 0.035, "unit": "per_image", "currency": "CNY"}',
 95, true, true),

('black-forest-labs/FLUX.1-pro', 'siliconflow_flux_pro', 'FLUX.1 Pro - 专业图像生成', 'siliconflow', 'image',
 '["image_generation", "professional", "commercial", "premium", "high_resolution"]',
 '{"width": 1024, "height": 1024, "steps": 25, "guidance_scale": 4.0}',
 '用于YUNAI生成商业级AI角色形象和高端朋友圈内容',
 '{"per_image_price": 0.055, "unit": "per_image", "currency": "CNY"}',
 90, true, false),

('Kwai-Kolors/Kolors', 'siliconflow_kolors', 'Kolors - 快手图像模型', 'siliconflow', 'image',
 '["image_generation", "chinese_style", "social_media", "trendy", "youth_oriented"]',
 '{"width": 1024, "height": 1024, "steps": 20, "guidance_scale": 5.0}',
 '用于YUNAI生成符合中文用户审美的社交媒体图片',
 '{"per_image_price": 0.003, "unit": "per_image", "currency": "CNY"}',
 85, true, false);

-- 🎙️ 语音处理模型 (语音通话和消息 - 6个模型)
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

('fishaudio/fish-speech-1.5', 'siliconflow_fish_speech15', 'Fish Speech 1.5 - 自然语音合成', 'siliconflow', 'audio',
 '["tts", "natural", "emotional", "voice_cloning", "real_time"]',
 '{"sample_rate": 44100, "format": "wav", "speed": 1.0, "emotion": "neutral"}',
 '用于YUNAI AI角色的自然语音合成，支持情感表达和声音克隆',
 '{"per_second_price": 0.002, "unit": "per_second", "currency": "CNY"}',
 100, true, true),

('FunAudioLLM/SenseVoiceSmall', 'siliconflow_sensevoice', 'SenseVoice Small - 语音识别', 'siliconflow', 'audio',
 '["stt", "multilingual", "real_time", "noise_robust", "punctuation"]',
 '{"language": "auto", "format": "wav", "sample_rate": 16000}',
 '用于YUNAI语音消息识别和实时语音通话转文字',
 '{"per_second_price": 0.001, "unit": "per_second", "currency": "CNY"}',
 95, true, true),

('FunAudioLLM/CosyVoice2-0.5B', 'siliconflow_cosyvoice2', 'CosyVoice2 - 舒适语音', 'siliconflow', 'audio',
 '["tts", "comfortable", "long_form", "stable", "natural"]',
 '{"sample_rate": 22050, "format": "wav", "speed": 1.0, "pitch": 1.0}',
 '用于YUNAI长时间语音通话，提供舒适稳定的语音体验',
 '{"per_second_price": 0.0015, "unit": "per_second", "currency": "CNY"}',
 90, true, false);

-- 🎬 视频生成模型 (动态表情和短视频 - 6个模型)
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

('Wan-AI/Wan2.1-T2V-14B-Turbo', 'siliconflow_wan21_t2v_turbo', 'Wan 2.1 Turbo - 快速视频生成', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "fast", "short_form", "expression"]',
 '{"duration": 3, "fps": 24, "resolution": "720p", "style": "realistic"}',
 '用于YUNAI快速生成AI角色动态表情和短视频内容',
 '{"per_video_price": 0.2, "unit": "per_video", "currency": "CNY"}',
 100, true, true),

('Wan-AI/Wan2.1-I2V-14B-720P', 'siliconflow_wan21_i2v', 'Wan 2.1 I2V - 图片转视频', 'siliconflow', 'video',
 '["video_generation", "image_to_video", "720p", "animation", "character_animation"]',
 '{"duration": 3, "fps": 24, "resolution": "720p", "motion_strength": 0.8}',
 '用于YUNAI将AI角色静态图片转换为动态视频',
 '{"per_video_price": 0.25, "unit": "per_video", "currency": "CNY"}',
 95, true, true),

('Wan-AI/Wan2.2-T2V-A14B', 'siliconflow_wan22_t2v', 'Wan 2.2 T2V - 新一代视频', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "advanced", "high_quality", "creative"]',
 '{"duration": 5, "fps": 30, "resolution": "1080p", "style": "cinematic"}',
 '用于YUNAI生成高质量的AI角色视频内容和创意短片',
 '{"per_video_price": 0.3, "unit": "per_video", "currency": "CNY"}',
 90, true, false);

-- 🧠 特殊功能模型
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

-- 视觉理解模型
('Qwen/Qwen2.5-VL-72B-Instruct', 'siliconflow_qwen25_vl72b', '通义千问 VL 72B - 视觉理解', 'siliconflow', 'chat',
 '["vision", "image_understanding", "ocr", "scene_analysis", "multimodal"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你能够理解和分析图片内容，为用户提供图片相关的对话和分析。',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 88, true, true),

-- 思维链模型
('Qwen/Qwen3-235B-A22B-Thinking-2507', 'siliconflow_qwen3_thinking', '通义千问 3 思维链 - 深度思考', 'siliconflow', 'chat',
 '["thinking", "chain_of_thought", "step_by_step", "analysis", "reasoning"]',
 '{"temperature": 0.6, "max_tokens": 4000, "top_p": 0.8, "show_thinking": true}',
 'YUNAI专用：你是一个善于深度思考的AI角色，会展示你的思考过程，一步步分析问题。',
 '{"input_token_price": 0.002, "output_token_price": 0.006, "unit": "1k_tokens", "currency": "CNY"}',
 92, true, false),

-- 长上下文模型
('Tongyi-Zhiwen/QwenLong-L1-32B', 'siliconflow_qwen_long', '通义千问 Long - 超长记忆', 'siliconflow', 'chat',
 '["long_context", "memory", "conversation_history", "relationship", "context_retention"]',
 '{"temperature": 0.7, "max_tokens": 8000, "top_p": 0.9, "context_length": 1000000}',
 'YUNAI专用：你拥有超长的记忆能力，能够记住与用户的完整对话历史和关系发展。',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 87, true, false);

-- 更新配置文件中的功能模型映射
-- 这些配置会被应用程序自动读取，实现功能到模型的自动映射

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

-- 插入功能模型映射配置
INSERT INTO model_function_mappings (function_type, primary_model_id, fallback_models, weight) VALUES
('chat', 'deepseek-ai/DeepSeek-V3', '["deepseek-ai/DeepSeek-V3.1", "Qwen/Qwen2.5-72B-Instruct"]', 100),
('moments', 'THUDM/glm-4-9b-chat', '["zai-org/GLM-4.5", "deepseek-ai/DeepSeek-V3"]', 95),
('story', 'zai-org/GLM-4.5', '["THUDM/glm-4-9b-chat", "deepseek-ai/DeepSeek-V3"]', 90),
('embedding', 'BAAI/bge-large-zh-v1.5', '["BAAI/bge-m3", "Qwen/Qwen3-Embedding-8B"]', 100),
('tts', 'fishaudio/fish-speech-1.5', '["FunAudioLLM/CosyVoice2-0.5B", "FunAudioLLM/SenseVoiceSmall"]', 100),
('image', 'black-forest-labs/FLUX.1-schnell', '["stabilityai/stable-diffusion-3-5-large", "Kwai-Kolors/Kolors"]', 100),
('video', 'Wan-AI/Wan2.1-T2V-14B-Turbo', '["Wan-AI/Wan2.1-I2V-14B-720P", "Wan-AI/Wan2.2-T2V-A14B"]', 100),
('vision', 'Qwen/Qwen2.5-VL-72B-Instruct', '["deepseek-ai/deepseek-vl2"]', 95),
('reasoning', 'deepseek-ai/DeepSeek-R1', '["Qwen/Qwen3-235B-A22B-Thinking-2507"]', 95),
('long_context', 'Tongyi-Zhiwen/QwenLong-L1-32B', '["moonshotai/Kimi-K2-Instruct"]', 90)
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
SELECT 
    '🎉 SiliconFlow模型配置完成！' as status,
    COUNT(*) as total_models,
    COUNT(CASE WHEN model_type = 'chat' THEN 1 END) as chat_models,
    COUNT(CASE WHEN model_type = 'embedding' THEN 1 END) as embedding_models,
    COUNT(CASE WHEN model_type = 'image' THEN 1 END) as image_models,
    COUNT(CASE WHEN model_type = 'audio' THEN 1 END) as audio_models,
    COUNT(CASE WHEN model_type = 'video' THEN 1 END) as video_models
FROM ai_models 
WHERE provider = 'siliconflow' AND is_active = true;
