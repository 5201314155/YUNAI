-- 快速添加SiliconFlow所有模型到数据库
-- 基于官方文档的完整模型列表

-- 添加缺失字段
ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS max_tokens INTEGER DEFAULT 8192;

-- 清理旧的SiliconFlow模型
DELETE FROM ai_models WHERE provider = 'siliconflow';

-- 插入完整的SiliconFlow模型列表
INSERT INTO ai_models (
    internal_key, display_name, provider, model_type, category, description,
    capabilities, params_schema, model_system_prompt, pricing, max_tokens,
    support_streaming, is_active, is_featured, weight, created_at, updated_at
) VALUES

-- 🤖 对话模型 (Chat Models)
('sf_deepseek_v3', 'DeepSeek V3', 'siliconflow', 'chat', '深度求索', 'DeepSeek V3超强推理模型，擅长逻辑推理和代码生成',
 '["chat", "reasoning", "code_generation", "analysis"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 '你是DeepSeek V3 AI助手，擅长逻辑推理和代码生成。请提供准确、有用的回答。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 100, NOW(), NOW()),

('sf_deepseek_r1', 'DeepSeek R1', 'siliconflow', 'chat', '深度求索', 'DeepSeek R1推理模型，专注于复杂推理任务',
 '["chat", "reasoning", "complex_analysis", "problem_solving"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 '你是DeepSeek R1推理模型，专注于复杂推理和问题解决。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 95, NOW(), NOW()),

('sf_qwen25_72b', 'Qwen2.5 72B', 'siliconflow', 'chat', '阿里通义', '阿里通义千问2.5 72B旗舰模型，中文理解能力突出',
 '["chat", "chinese", "knowledge", "reasoning", "creative_writing"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 '你是通义千问2.5，阿里巴巴开发的AI助手。请用中文回答，提供准确、有帮助的信息。',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 90, NOW(), NOW()),

('sf_glm4_9b', 'GLM-4 9B', 'siliconflow', 'chat', '清华GLM', '清华GLM-4 9B模型，中英双语对话专家',
 '["chat", "chinese", "role_play", "creative_writing", "emotion"]',
 '{"temperature": {"type": "number", "default": 0.8, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 3000, "min": 1, "max": 6000}, "stream": {"type": "boolean", "default": false}}',
 '你是GLM-4 AI助手，由清华大学开发。请提供专业、准确的回答。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 6000, true, true, true, 85, NOW(), NOW()),

('sf_qwen3_235b', 'Qwen3 235B', 'siliconflow', 'chat', '阿里通义', '通义千问3代235B超大模型，最强中文理解',
 '["chat", "chinese", "knowledge", "reasoning", "complex_analysis"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 8000}, "stream": {"type": "boolean", "default": false}}',
 '你是通义千问3代，具有最强的中文理解和推理能力。',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, true, true, true, 98, NOW(), NOW()),

('sf_kimi_k2', 'Kimi K2', 'siliconflow', 'chat', '月之暗面', 'Kimi K2长文本处理专家，支持超长上下文',
 '["chat", "long_context", "document_analysis", "summarization"]',
 '{"temperature": {"type": "number", "default": 0.7, "min": 0, "max": 2}, "max_tokens": {"type": "integer", "default": 4000, "min": 1, "max": 200000}, "stream": {"type": "boolean", "default": false}}',
 '你是Kimi K2，月之暗面开发的长文本处理专家，擅长文档分析和总结。',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 200000, true, true, true, 88, NOW(), NOW()),

-- 🧠 嵌入模型 (Embedding Models)
('sf_bge_large_zh', 'BGE Large ZH', 'siliconflow', 'embedding', '嵌入模型', 'BGE中文大模型，专业的中文语义理解',
 '["embedding", "chinese", "semantic_search", "similarity"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'BGE中文语义理解嵌入模型，用于中文文本的向量化表示',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_bge_large_en', 'BGE Large EN', 'siliconflow', 'embedding', '嵌入模型', 'BGE英文大模型，专业的英文语义理解',
 '["embedding", "english", "semantic_search", "similarity"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'BGE英文语义理解嵌入模型，用于英文文本的向量化表示',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 95, NOW(), NOW()),

('sf_bge_m3', 'BGE M3', 'siliconflow', 'embedding', '嵌入模型', 'BGE M3多语言嵌入模型，支持跨语言检索',
 '["embedding", "multilingual", "cross_lingual", "semantic_search"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 'BGE M3多语言嵌入模型，支持中英文等多语言语义理解',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 90, NOW(), NOW()),

('sf_qwen3_embedding', 'Qwen3 Embedding', 'siliconflow', 'embedding', '阿里通义', '通义千问3代嵌入模型，阿里最新嵌入技术',
 '["embedding", "chinese", "semantic_search", "text_analysis"]',
 '{"dimensions": {"type": "integer", "default": 1024}, "normalize": {"type": "boolean", "default": true}}',
 '通义千问3代嵌入模型，提供高质量的文本向量化服务',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 85, NOW(), NOW()),

-- 🔄 重排序模型 (Reranking Models)
('sf_bge_reranker_v2', 'BGE Reranker V2', 'siliconflow', 'reranking', '重排序', 'BGE重排序模型V2，优化搜索结果相关性',
 '["reranking", "search_optimization", "relevance_scoring"]',
 '{"top_k": {"type": "integer", "default": 10}, "threshold": {"type": "number", "default": 0.5}}',
 'BGE重排序模型，用于优化搜索结果的相关性排序',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 80, NOW(), NOW()),

('sf_qwen3_reranker', 'Qwen3 Reranker', 'siliconflow', 'reranking', '阿里通义', '通义千问3代重排序模型，智能相关性评分',
 '["reranking", "search_optimization", "relevance_scoring", "chinese"]',
 '{"top_k": {"type": "integer", "default": 10}, "threshold": {"type": "number", "default": 0.5}}',
 '通义千问3代重排序模型，专门用于中文搜索结果优化',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 75, NOW(), NOW()),

-- 🎨 图像生成模型 (Image Generation Models)
('sf_flux_schnell', 'FLUX.1 Schnell', 'siliconflow', 'image', '图像生成', 'FLUX.1 Schnell快速图像生成，4步生成高质量图像',
 '["image_generation", "text_to_image", "fast", "high_quality"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 4}}',
 'FLUX.1 Schnell快速图像生成模型，专门用于高质量图像创作',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_flux_dev', 'FLUX.1 Dev', 'siliconflow', 'image', '图像生成', 'FLUX.1 Dev开发版，更灵活的图像生成控制',
 '["image_generation", "text_to_image", "flexible", "creative"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}',
 'FLUX.1 Dev开发版图像生成模型，提供更多创作控制选项',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 95, NOW(), NOW()),

('sf_flux_pro', 'FLUX.1 Pro', 'siliconflow', 'image', '图像生成', 'FLUX.1 Pro专业版，最高质量图像生成',
 '["image_generation", "text_to_image", "professional", "ultra_quality"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 25}}',
 'FLUX.1 Pro专业版，提供最高质量的图像生成服务',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 98, NOW(), NOW()),

('sf_stable_diffusion_xl', 'Stable Diffusion XL', 'siliconflow', 'image', '图像生成', 'Stable Diffusion XL经典图像生成模型',
 '["image_generation", "text_to_image", "artistic", "versatile"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}',
 'Stable Diffusion XL经典图像生成模型，支持多种艺术风格',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 85, NOW(), NOW()),

('sf_kolors', 'Kolors', 'siliconflow', 'image', '图像生成', 'Kolors快手图像生成模型，中文理解优秀',
 '["image_generation", "text_to_image", "chinese", "creative"]',
 '{"width": {"type": "integer", "default": 1024}, "height": {"type": "integer", "default": 1024}, "steps": {"type": "integer", "default": 20}}',
 'Kolors快手图像生成模型，对中文提示词理解更准确',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 80, NOW(), NOW()),

-- 🎵 语音合成模型 (TTS Models)
('sf_fish_speech_15', 'Fish Speech 1.5', 'siliconflow', 'audio', '语音合成', 'Fish Speech 1.5语音合成，自然流畅的语音生成',
 '["tts", "voice_synthesis", "natural", "multilingual"]',
 '{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}, "pitch": {"type": "number", "default": 1.0}}',
 'Fish Speech 1.5语音合成模型，生成自然流畅的语音',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_fish_speech_14', 'Fish Speech 1.4', 'siliconflow', 'audio', '语音合成', 'Fish Speech 1.4语音合成，稳定的语音生成服务',
 '["tts", "voice_synthesis", "stable", "multilingual"]',
 '{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}, "pitch": {"type": "number", "default": 1.0}}',
 'Fish Speech 1.4语音合成模型，提供稳定的语音生成服务',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 95, NOW(), NOW()),

('sf_cosyvoice', 'CosyVoice', 'siliconflow', 'audio', '语音合成', 'CosyVoice阿里语音合成，支持多种音色',
 '["tts", "voice_synthesis", "multi_voice", "chinese"]',
 '{"voice_id": {"type": "string", "default": "default"}, "speed": {"type": "number", "default": 1.0}, "emotion": {"type": "string", "default": "neutral"}}',
 'CosyVoice阿里语音合成模型，支持多种音色和情感表达',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 90, NOW(), NOW()),

-- 🎤 语音识别模型 (ASR Models)
('sf_sensevoice', 'SenseVoice', 'siliconflow', 'audio', '语音识别', 'SenseVoice语音识别，支持多语言实时识别',
 '["asr", "speech_to_text", "multilingual", "real_time"]',
 '{"language": {"type": "string", "default": "auto"}, "format": {"type": "string", "default": "wav"}}',
 'SenseVoice语音识别模型，支持多语言语音转文字',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 85, NOW(), NOW()),

-- 🎬 视频生成模型 (Video Generation Models)
('sf_wan_t2v_14b', 'Wan2.1 T2V 14B', 'siliconflow', 'video', '视频生成', 'Wan AI文本转视频14B模型，创意视频生成',
 '["video_generation", "text_to_video", "creative", "storytelling"]',
 '{"duration": {"type": "integer", "default": 5}, "fps": {"type": "integer", "default": 24}, "resolution": {"type": "string", "default": "720p"}}',
 'Wan AI文本转视频模型，用于创意视频内容生成',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 8192, false, true, true, 100, NOW(), NOW()),

('sf_wan_i2v_14b', 'Wan2.1 I2V 14B', 'siliconflow', 'video', '视频生成', 'Wan AI图像转视频14B模型，图像动画化',
 '["video_generation", "image_to_video", "animation", "motion"]',
 '{"duration": {"type": "integer", "default": 5}, "fps": {"type": "integer", "default": 24}, "resolution": {"type": "string", "default": "720p"}}',
 'Wan AI图像转视频模型，将静态图像转换为动态视频',
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

-- 显示添加结果
SELECT 
    category,
    model_type,
    COUNT(*) as model_count,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_count
FROM ai_models 
WHERE provider = 'siliconflow'
GROUP BY category, model_type
ORDER BY category, model_count DESC;
