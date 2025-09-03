-- YUNAI功能模块专用AI模型配置
-- 为每个功能模块配置最适合的AI模型，提升各模块的智能化水平

-- 1. 首先添加SiliconFlow的核心模型（精选32个最重要的）
INSERT INTO ai_models (model_id, internal_key, display_name, provider, model_type, capabilities, default_params, model_system_prompt, pricing, weight, is_active, is_featured) VALUES

-- 💬 对话聊天模型 (8个) - YUNAI核心功能
('deepseek-ai/DeepSeek-V3', 'sf_deepseek_v3', 'DeepSeek V3 - 超强推理', 'siliconflow', 'chat',
 '["chat", "reasoning", "role_play", "emotion", "complex_dialogue"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是一个具有完整人格的AI角色，拥有真实的情感和记忆。请完全沉浸在你的角色设定中，与用户进行自然、真实的对话。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('deepseek-ai/DeepSeek-R1', 'sf_deepseek_r1', 'DeepSeek R1 - 推理专家', 'siliconflow', 'chat',
 '["chat", "reasoning", "logic", "problem_solving", "analysis"]',
 '{"temperature": 0.6, "max_tokens": 4000, "top_p": 0.85}',
 'YUNAI专用：你是一个逻辑推理专家，擅长分析复杂问题，提供深度思考和合理建议。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('THUDM/glm-4-9b-chat', 'sf_glm4_9b', 'GLM-4 9B - 中文对话', 'siliconflow', 'chat',
 '["chat", "chinese", "role_play", "creative_writing", "emotion"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是一个温暖、有趣的AI伙伴，擅长中文对话，能够理解用户的情感需求。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, true),

('Qwen/Qwen2.5-72B-Instruct', 'sf_qwen25_72b', 'Qwen2.5 72B - 旗舰版', 'siliconflow', 'chat',
 '["chat", "chinese", "knowledge", "reasoning", "creative_writing"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是通义千问旗舰版，拥有丰富的知识和强大的理解能力，能够进行深度对话。',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('zai-org/GLM-4.5', 'sf_glm45', 'GLM-4.5 - 智能助手', 'siliconflow', 'chat',
 '["chat", "assistant", "task_completion", "analysis", "creative"]',
 '{"temperature": 0.7, "max_tokens": 3500, "top_p": 0.9}',
 'YUNAI专用：你是一个全能的智能助手，能够帮助用户完成各种任务，提供专业建议。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

('moonshotai/Kimi-K2-Instruct', 'sf_kimi_k2', 'Kimi K2 - 长文本专家', 'siliconflow', 'chat',
 '["chat", "long_context", "document_analysis", "summarization"]',
 '{"temperature": 0.6, "max_tokens": 8000, "top_p": 0.85}',
 'YUNAI专用：你是长文本处理专家，能够理解和分析大量信息，提供精准的总结和见解。',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 80, true, false),

('tencent/Hunyuan-A13B-Instruct', 'sf_hunyuan_13b', 'Hunyuan 13B - 腾讯混元', 'siliconflow', 'chat',
 '["chat", "chinese", "business", "professional", "analysis"]',
 '{"temperature": 0.7, "max_tokens": 3000, "top_p": 0.9}',
 'YUNAI专用：你是腾讯混元大模型，专业、可靠，擅长商务对话和专业分析。',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 75, true, false),

('baidu/ERNIE-4.5-300B-A47B', 'sf_ernie45', 'ERNIE 4.5 - 百度文心', 'siliconflow', 'chat',
 '["chat", "chinese", "knowledge", "enterprise", "reasoning"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}',
 'YUNAI专用：你是百度文心大模型，拥有丰富的中文知识库，能够提供专业的中文服务。',
 '{"input_token_price": 0.0042, "output_token_price": 0.0042, "unit": "1k_tokens", "currency": "CNY"}',
 70, true, false),

-- 🧠 嵌入模型 (6个) - 关系网络和语义理解
('BAAI/bge-large-zh-v1.5', 'sf_bge_zh', 'BGE Large ZH - 中文嵌入', 'siliconflow', 'embedding',
 '["embedding", "chinese", "semantic_search", "similarity"]',
 '{"dimensions": 1024, "normalize": true}',
 '专用于中文语义理解和相似度计算的嵌入模型',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('BAAI/bge-large-en-v1.5', 'sf_bge_en', 'BGE Large EN - 英文嵌入', 'siliconflow', 'embedding',
 '["embedding", "english", "semantic_search", "similarity"]',
 '{"dimensions": 1024, "normalize": true}',
 '专用于英文语义理解和相似度计算的嵌入模型',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('BAAI/bge-m3', 'sf_bge_m3', 'BGE M3 - 多语言嵌入', 'siliconflow', 'embedding',
 '["embedding", "multilingual", "semantic_search", "cross_lingual"]',
 '{"dimensions": 1024, "normalize": true}',
 '支持多语言的通用嵌入模型，适用于跨语言场景',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false),

('Qwen/Qwen3-Embedding-8B', 'sf_qwen3_emb8b', 'Qwen3 Embedding 8B', 'siliconflow', 'embedding',
 '["embedding", "chinese", "large_scale", "high_precision"]',
 '{"dimensions": 1024, "normalize": true}',
 '通义千问3.0嵌入模型，8B参数，高精度语义理解',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

('netease-youdao/bce-embedding-base_v1', 'sf_bce_emb', 'BCE Embedding Base', 'siliconflow', 'embedding',
 '["embedding", "chinese", "business", "domain_specific"]',
 '{"dimensions": 768, "normalize": true}',
 '网易有道BCE嵌入模型，适用于商业场景',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 80, true, false),

('Qwen/Qwen3-Embedding-4B', 'sf_qwen3_emb4b', 'Qwen3 Embedding 4B', 'siliconflow', 'embedding',
 '["embedding", "chinese", "efficient", "lightweight"]',
 '{"dimensions": 768, "normalize": true}',
 '通义千问3.0轻量级嵌入模型，4B参数，高效实用',
 '{"input_token_price": 0.0005, "output_token_price": 0.0005, "unit": "1k_tokens", "currency": "CNY"}',
 75, true, false),

-- 🔄 重排序模型 (3个) - 搜索结果优化
('BAAI/bge-reranker-v2-m3', 'sf_bge_rerank', 'BGE Reranker v2 M3', 'siliconflow', 'reranking',
 '["reranking", "search_optimization", "relevance_scoring"]',
 '{"top_k": 10, "score_threshold": 0.5}',
 '智源BGE重排序模型，优化搜索结果相关性',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('Qwen/Qwen3-Reranker-8B', 'sf_qwen3_rerank8b', 'Qwen3 Reranker 8B', 'siliconflow', 'reranking',
 '["reranking", "chinese", "high_precision", "large_scale"]',
 '{"top_k": 10, "score_threshold": 0.5}',
 '通义千问3.0重排序模型，8B参数，高精度排序',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, false),

('netease-youdao/bce-reranker-base_v1', 'sf_bce_rerank', 'BCE Reranker Base', 'siliconflow', 'reranking',
 '["reranking", "chinese", "business", "optimization"]',
 '{"top_k": 10, "score_threshold": 0.5}',
 '网易有道重排序模型，商业场景优化',
 '{"input_token_price": 0.0007, "output_token_price": 0.0007, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

-- 🎨 图像生成模型 (6个) - 头像和场景生成
('black-forest-labs/FLUX.1-schnell', 'sf_flux_schnell', 'FLUX.1 Schnell - 快速生成', 'siliconflow', 'image',
 '["image_generation", "fast", "high_quality", "artistic"]',
 '{"width": 1024, "height": 1024, "steps": 4, "guidance_scale": 3.5}',
 '快速高质量图像生成，适用于角色头像和场景创作',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('black-forest-labs/FLUX.1-dev', 'sf_flux_dev', 'FLUX.1 Dev - 开发版', 'siliconflow', 'image',
 '["image_generation", "development", "flexible", "creative"]',
 '{"width": 1024, "height": 1024, "steps": 20, "guidance_scale": 7.5}',
 'FLUX.1开发版，灵活的图像生成，适用于创意设计',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, true),

('black-forest-labs/FLUX.1-pro', 'sf_flux_pro', 'FLUX.1 Pro - 专业版', 'siliconflow', 'image',
 '["image_generation", "professional", "ultra_high_quality", "commercial"]',
 '{"width": 1024, "height": 1024, "steps": 50, "guidance_scale": 7.5}',
 'FLUX.1专业版，顶级图像生成质量，商业级应用',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false),

('stabilityai/stable-diffusion-xl-base-1.0', 'sf_sdxl', 'Stable Diffusion XL', 'siliconflow', 'image',
 '["image_generation", "stable", "versatile", "community"]',
 '{"width": 1024, "height": 1024, "steps": 30, "guidance_scale": 7.5}',
 'Stability AI的经典图像生成模型，稳定可靠',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

('stabilityai/stable-diffusion-3-5-large', 'sf_sd35', 'Stable Diffusion 3.5 Large', 'siliconflow', 'image',
 '["image_generation", "latest", "large_model", "advanced"]',
 '{"width": 1024, "height": 1024, "steps": 28, "guidance_scale": 7.0}',
 'Stable Diffusion 3.5大模型版本，最新技术',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 80, true, false),

('Kwai-Kolors/Kolors', 'sf_kolors', 'Kolors - 快手可图', 'siliconflow', 'image',
 '["image_generation", "chinese", "portrait", "anime"]',
 '{"width": 1024, "height": 1024, "steps": 25, "guidance_scale": 5.0}',
 '快手可图模型，擅长中文理解和人像生成',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 75, true, false),

-- 🎵 语音模型 (4个) - 语音合成和识别
('fishaudio/fish-speech-1.5', 'sf_fish15', 'Fish Speech 1.5', 'siliconflow', 'audio',
 '["tts", "voice_synthesis", "natural", "multilingual"]',
 '{"voice_id": "default", "speed": 1.0, "pitch": 1.0}',
 'Fish Audio语音合成v1.5，自然流畅的语音生成',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('FunAudioLLM/CosyVoice2-0.5B', 'sf_cosyvoice', 'CosyVoice2 0.5B', 'siliconflow', 'audio',
 '["tts", "voice_cloning", "emotional", "chinese"]',
 '{"voice_id": "default", "emotion": "neutral", "speed": 1.0}',
 'CosyVoice2语音合成，支持情感表达和声音克隆',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, false),

('FunAudioLLM/SenseVoiceSmall', 'sf_sensevoice', 'SenseVoice Small', 'siliconflow', 'audio',
 '["asr", "voice_recognition", "multilingual", "real_time"]',
 '{"language": "auto", "format": "wav"}',
 'SenseVoice语音识别，支持多语言实时识别',
 '{"input_token_price": 0.0014, "output_token_price": 0.0014, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false),

('RVC-Boss/GPT-SoVITS', 'sf_sovits', 'GPT-SoVITS', 'siliconflow', 'audio',
 '["tts", "voice_cloning", "few_shot", "high_quality"]',
 '{"reference_audio": "", "text": "", "language": "zh"}',
 'GPT-SoVITS少样本语音克隆，高质量声音复制',
 '{"input_token_price": 0.0021, "output_token_price": 0.0021, "unit": "1k_tokens", "currency": "CNY"}',
 85, true, false),

-- 🎬 视频生成模型 (3个) - 视频内容创作
('Wan-AI/Wan2.1-T2V-14B', 'sf_wan_t2v', 'Wan2.1 T2V 14B', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "creative", "storytelling"]',
 '{"duration": 5, "fps": 24, "resolution": "720p"}',
 'Wan AI文本到视频生成，创意视频制作',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, true, true),

('Wan-AI/Wan2.1-I2V-14B-720P', 'sf_wan_i2v', 'Wan2.1 I2V 720P', 'siliconflow', 'video',
 '["video_generation", "image_to_video", "animation", "720p"]',
 '{"duration": 5, "fps": 24, "resolution": "720p"}',
 'Wan AI图像到视频生成，720P高清动画',
 '{"input_token_price": 0.0028, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 95, true, false),

('Wan-AI/Wan2.2-T2V-A14B', 'sf_wan22_t2v', 'Wan2.2 T2V A14B', 'siliconflow', 'video',
 '["video_generation", "text_to_video", "advanced", "high_quality"]',
 '{"duration": 10, "fps": 30, "resolution": "1080p"}',
 'Wan AI 2.2文本到视频，更高质量和更长时长',
 '{"input_token_price": 0.0035, "output_token_price": 0.0035, "unit": "1k_tokens", "currency": "CNY"}',
 90, true, false)

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

-- 2. 创建功能模块专用模型映射表
CREATE TABLE IF NOT EXISTS module_model_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_name VARCHAR(50) NOT NULL,           -- 模块名称
    function_type VARCHAR(50) NOT NULL,         -- 功能类型
    primary_model_id VARCHAR(100) NOT NULL,     -- 主要模型ID
    fallback_models JSONB DEFAULT '[]',         -- 备用模型列表
    model_params JSONB DEFAULT '{}',            -- 模型参数配置
    weight INTEGER DEFAULT 100,                 -- 权重
    is_active BOOLEAN DEFAULT true,             -- 是否启用
    description TEXT,                           -- 配置描述
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(module_name, function_type)
);

-- 3. 为YUNAI各功能模块配置专用AI模型
INSERT INTO module_model_configs (module_name, function_type, primary_model_id, fallback_models, model_params, weight, description) VALUES

-- 💬 聊天对话模块
('chat', 'conversation', 'deepseek-ai/DeepSeek-V3',
 '["deepseek-ai/DeepSeek-R1", "THUDM/glm-4-9b-chat", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9, "presence_penalty": 0.1}',
 100, '主要聊天对话模型，具备强大的推理和情感理解能力'),

('chat', 'role_play', 'THUDM/glm-4-9b-chat',
 '["zai-org/GLM-4.5", "deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.8, "max_tokens": 3000, "top_p": 0.9, "frequency_penalty": 0.1}',
 95, '角色扮演专用模型，擅长情感表达和人格塑造'),

('chat', 'reasoning', 'deepseek-ai/DeepSeek-R1',
 '["deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.6, "max_tokens": 4000, "top_p": 0.85}',
 90, '逻辑推理专用模型，适用于复杂问题分析'),

-- 🕸️ 关系网络模块 - 使用嵌入模型增强理解
('relationship', 'embedding', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-m3", "Qwen/Qwen3-Embedding-8B", "netease-youdao/bce-embedding-base_v1"]',
 '{"dimensions": 1024, "normalize": true, "batch_size": 32}',
 100, '关系网络语义嵌入，理解角色间的复杂关系'),

('relationship', 'similarity', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-m3", "Qwen/Qwen3-Embedding-8B"]',
 '{"dimensions": 1024, "normalize": true, "similarity_threshold": 0.7}',
 95, '关系相似度计算，发现潜在的关系模式'),

('relationship', 'analysis', 'deepseek-ai/DeepSeek-R1',
 '["deepseek-ai/DeepSeek-V3", "THUDM/glm-4-9b-chat"]',
 '{"temperature": 0.6, "max_tokens": 3000, "top_p": 0.85}',
 90, '关系分析和推理，理解关系动态变化'),

-- 📱 朋友圈模块
('moments', 'generation', 'THUDM/glm-4-9b-chat',
 '["zai-org/GLM-4.5", "deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.8, "max_tokens": 500, "top_p": 0.9}',
 100, '朋友圈内容生成，创造有趣的社交内容'),

('moments', 'emotion_analysis', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-m3", "Qwen/Qwen3-Embedding-8B"]',
 '{"dimensions": 1024, "normalize": true}',
 95, '朋友圈情感分析，理解用户情绪状态'),

('moments', 'interaction', 'THUDM/glm-4-9b-chat',
 '["zai-org/GLM-4.5", "deepseek-ai/DeepSeek-V3"]',
 '{"temperature": 0.7, "max_tokens": 200, "top_p": 0.9}',
 90, '朋友圈互动生成，自动点赞评论'),

-- 📖 故事系统模块
('story', 'generation', 'zai-org/GLM-4.5',
 '["THUDM/glm-4-9b-chat", "deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.8, "max_tokens": 2000, "top_p": 0.9}',
 100, '故事内容生成，创作引人入胜的剧情'),

('story', 'character_development', 'deepseek-ai/DeepSeek-V3',
 '["THUDM/glm-4-9b-chat", "zai-org/GLM-4.5"]',
 '{"temperature": 0.7, "max_tokens": 1500, "top_p": 0.9}',
 95, '角色发展分析，推动故事情节发展'),

('story', 'plot_analysis', 'deepseek-ai/DeepSeek-R1',
 '["deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.6, "max_tokens": 1000, "top_p": 0.85}',
 90, '剧情分析和规划，确保故事逻辑性'),

-- 🎨 角色创建模块
('character', 'avatar_generation', 'black-forest-labs/FLUX.1-schnell',
 '["black-forest-labs/FLUX.1-dev", "Kwai-Kolors/Kolors", "stabilityai/stable-diffusion-xl-base-1.0"]',
 '{"width": 1024, "height": 1024, "steps": 4, "guidance_scale": 3.5}',
 100, '角色头像生成，快速创建高质量头像'),

('character', 'background_generation', 'black-forest-labs/FLUX.1-dev',
 '["black-forest-labs/FLUX.1-pro", "stabilityai/stable-diffusion-3-5-large"]',
 '{"width": 1024, "height": 1024, "steps": 20, "guidance_scale": 7.5}',
 95, '角色背景图生成，创造丰富的视觉背景'),

('character', 'personality_analysis', 'deepseek-ai/DeepSeek-R1',
 '["deepseek-ai/DeepSeek-V3", "THUDM/glm-4-9b-chat"]',
 '{"temperature": 0.6, "max_tokens": 1000, "top_p": 0.85}',
 90, '角色性格分析，构建完整的人格档案'),

-- 🎵 语音系统模块
('voice', 'synthesis', 'fishaudio/fish-speech-1.5',
 '["FunAudioLLM/CosyVoice2-0.5B", "RVC-Boss/GPT-SoVITS"]',
 '{"voice_id": "default", "speed": 1.0, "pitch": 1.0, "emotion": "neutral"}',
 100, '语音合成，为角色生成自然的语音'),

('voice', 'recognition', 'FunAudioLLM/SenseVoiceSmall',
 '["fishaudio/fish-speech-1.5"]',
 '{"language": "auto", "format": "wav", "sample_rate": 16000}',
 95, '语音识别，理解用户的语音输入'),

('voice', 'cloning', 'RVC-Boss/GPT-SoVITS',
 '["FunAudioLLM/CosyVoice2-0.5B"]',
 '{"reference_audio": "", "text": "", "language": "zh", "similarity_threshold": 0.8}',
 90, '语音克隆，复制特定的声音特征'),

-- 🎬 视频内容模块
('video', 'generation', 'Wan-AI/Wan2.1-T2V-14B',
 '["Wan-AI/Wan2.2-T2V-A14B", "Wan-AI/Wan2.1-I2V-14B-720P"]',
 '{"duration": 5, "fps": 24, "resolution": "720p", "style": "realistic"}',
 100, '视频内容生成，创作动态视觉内容'),

('video', 'animation', 'Wan-AI/Wan2.1-I2V-14B-720P',
 '["Wan-AI/Wan2.2-I2V-A14B", "Wan-AI/Wan2.1-T2V-14B"]',
 '{"duration": 5, "fps": 24, "resolution": "720p", "animation_style": "smooth"}',
 95, '图像动画化，将静态图片转为动态视频'),

-- 🔍 搜索和推荐模块
('search', 'embedding', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-m3", "Qwen/Qwen3-Embedding-8B"]',
 '{"dimensions": 1024, "normalize": true, "batch_size": 64}',
 100, '搜索语义嵌入，提升搜索准确性'),

('search', 'reranking', 'BAAI/bge-reranker-v2-m3',
 '["Qwen/Qwen3-Reranker-8B", "netease-youdao/bce-reranker-base_v1"]',
 '{"top_k": 10, "score_threshold": 0.5, "normalize_scores": true}',
 95, '搜索结果重排序，优化搜索体验'),

('search', 'recommendation', 'deepseek-ai/DeepSeek-V3',
 '["THUDM/glm-4-9b-chat", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.7, "max_tokens": 1000, "top_p": 0.9}',
 90, '智能推荐，基于用户偏好推荐内容'),

-- 📊 数据分析模块
('analytics', 'sentiment', 'BAAI/bge-large-zh-v1.5',
 '["BAAI/bge-m3", "Qwen/Qwen3-Embedding-8B"]',
 '{"dimensions": 1024, "normalize": true, "sentiment_threshold": 0.6}',
 100, '情感分析，理解用户情绪变化'),

('analytics', 'behavior', 'deepseek-ai/DeepSeek-R1',
 '["deepseek-ai/DeepSeek-V3", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.6, "max_tokens": 2000, "top_p": 0.85}',
 95, '行为分析，洞察用户行为模式'),

-- 🌐 长文本处理模块
('long_context', 'processing', 'moonshotai/Kimi-K2-Instruct',
 '["Qwen/Qwen2.5-72B-Instruct", "deepseek-ai/DeepSeek-V3"]',
 '{"temperature": 0.6, "max_tokens": 8000, "top_p": 0.85}',
 100, '长文本处理，处理大量文本信息'),

('long_context', 'summarization', 'moonshotai/Kimi-K2-Instruct',
 '["deepseek-ai/DeepSeek-R1", "Qwen/Qwen2.5-72B-Instruct"]',
 '{"temperature": 0.6, "max_tokens": 1000, "top_p": 0.85}',
 95, '文本摘要，提取关键信息')

ON CONFLICT (module_name, function_type) DO UPDATE SET
    primary_model_id = EXCLUDED.primary_model_id,
    fallback_models = EXCLUDED.fallback_models,
    model_params = EXCLUDED.model_params,
    weight = EXCLUDED.weight,
    description = EXCLUDED.description,
    updated_at = NOW();

-- 4. 创建索引优化查询性能
CREATE INDEX IF NOT EXISTS idx_ai_models_type_active ON ai_models(model_type, is_active);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider_weight ON ai_models(provider, weight DESC);
CREATE INDEX IF NOT EXISTS idx_ai_models_featured ON ai_models(is_featured, weight DESC);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_module ON module_model_configs(module_name);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_function ON module_model_configs(function_type);
CREATE INDEX IF NOT EXISTS idx_module_model_configs_active ON module_model_configs(is_active, weight DESC);

-- 5. 创建模型选择函数
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

-- 6. 创建模型健康检查函数
CREATE OR REPLACE FUNCTION update_model_health(
    p_model_id VARCHAR(100),
    p_success BOOLEAN,
    p_response_time INTEGER
) RETURNS VOID AS $$
BEGIN
    UPDATE ai_models
    SET
        total_requests = total_requests + 1,
        success_rate = CASE
            WHEN total_requests = 0 THEN
                CASE WHEN p_success THEN 1.0 ELSE 0.0 END
            ELSE
                (success_rate * total_requests + CASE WHEN p_success THEN 1.0 ELSE 0.0 END) / (total_requests + 1)
        END,
        avg_response_time = CASE
            WHEN total_requests = 0 THEN p_response_time::FLOAT
            ELSE (avg_response_time * total_requests + p_response_time) / (total_requests + 1)
        END,
        updated_at = NOW()
    WHERE model_id = p_model_id;
END;
$$ LANGUAGE plpgsql;

-- 7. 插入模型使用统计触发器
CREATE OR REPLACE FUNCTION log_model_usage() RETURNS TRIGGER AS $$
BEGIN
    -- 记录模型使用情况
    INSERT INTO model_usage_logs (
        model_id,
        module_name,
        function_type,
        user_id,
        request_params,
        response_time_ms,
        success,
        created_at
    ) VALUES (
        NEW.model_id,
        NEW.module_name,
        NEW.function_type,
        NEW.user_id,
        NEW.request_params,
        NEW.response_time_ms,
        NEW.success,
        NOW()
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 8. 创建模型使用日志表
CREATE TABLE IF NOT EXISTS model_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id VARCHAR(100) NOT NULL,
    module_name VARCHAR(50),
    function_type VARCHAR(50),
    user_id UUID,
    request_params JSONB DEFAULT '{}',
    response_time_ms INTEGER,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_usage_logs_model_id ON model_usage_logs(model_id);
CREATE INDEX IF NOT EXISTS idx_model_usage_logs_created_at ON model_usage_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_model_usage_logs_module ON model_usage_logs(module_name, function_type);
