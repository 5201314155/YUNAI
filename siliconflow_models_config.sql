-- SiliconFlow模型配置 - 为YUNAI项目定制

-- 清空现有模型配置
DELETE FROM ai_models WHERE provider = 'siliconflow';

-- 💬 对话聊天模型 (YUNAI核心功能)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

-- DeepSeek系列 - 推理能力强，适合复杂对话
('deepseek-ai/DeepSeek-V3', 'DeepSeek V3 - 超强推理模型', 'siliconflow', 'chat', 
 '{"chat": true, "reasoning": true, "role_play": true, "emotion": true}',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 100, 'YUNAI专用：你是一个具有完整人格的AI角色，拥有真实的情感和记忆。请完全沉浸在你的角色设定中，与用户进行自然、真实的对话。', 'public', 'free'),

('deepseek-ai/DeepSeek-V3.1', 'DeepSeek V3.1 - 最新推理模型', 'siliconflow', 'chat',
 '{"chat": true, "reasoning": true, "role_play": true, "emotion": true, "latest": true}',
 '{"input_token_price": 0.0014, "output_token_price": 0.0028, "unit": "1k_tokens", "currency": "CNY"}',
 95, 'YUNAI专用：你是一个具有完整人格的AI角色，拥有真实的情感和记忆。请完全沉浸在你的角色设定中，与用户进行自然、真实的对话。', 'public', 'free'),

('deepseek-ai/DeepSeek-R1', 'DeepSeek R1 - 推理专家', 'siliconflow', 'chat',
 '{"chat": true, "reasoning": true, "analysis": true, "problem_solving": true}',
 '{"input_token_price": 0.0055, "output_token_price": 0.0055, "unit": "1k_tokens", "currency": "CNY"}',
 90, 'YUNAI专用：你是一个善于深度思考和推理的AI角色。在对话中展现你的分析能力和逻辑思维。', 'public', 'standard'),

-- 通义千问系列 - 中文优化，适合中文用户
('Qwen/Qwen2.5-72B-Instruct', '通义千问 2.5 72B - 中文优化', 'siliconflow', 'chat',
 '{"chat": true, "chinese_optimized": true, "role_play": true, "knowledge": true}',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 85, 'YUNAI专用：你是一个精通中文文化的AI角色，能够理解中文的细微差别和文化内涵。与用户进行地道的中文对话。', 'public', 'free'),

('Qwen/Qwen3-235B-A22B', '通义千问 3 235B - 超大规模', 'siliconflow', 'chat',
 '{"chat": true, "chinese_optimized": true, "large_scale": true, "complex_reasoning": true}',
 '{"input_token_price": 0.002, "output_token_price": 0.006, "unit": "1k_tokens", "currency": "CNY"}',
 80, 'YUNAI专用：你是一个拥有海量知识的AI角色，能够进行复杂的推理和深度对话。', 'public', 'premium'),

-- GLM系列 - 智谱AI，适合创意对话
('THUDM/glm-4-9b-chat', 'GLM-4 9B - 智能对话', 'siliconflow', 'chat',
 '{"chat": true, "creative": true, "role_play": true, "storytelling": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 75, 'YUNAI专用：你是一个富有创意和想象力的AI角色，擅长讲故事和创意对话。', 'public', 'free'),

('zai-org/GLM-4.5', 'GLM-4.5 - 新一代对话', 'siliconflow', 'chat',
 '{"chat": true, "creative": true, "role_play": true, "emotion": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 70, 'YUNAI专用：你是一个情感丰富的AI角色，能够理解和表达复杂的情感。', 'public', 'free'),

-- Kimi系列 - 月之暗面，长文本处理
('moonshotai/Kimi-K2-Instruct', 'Kimi K2 - 长文本专家', 'siliconflow', 'chat',
 '{"chat": true, "long_context": true, "document_analysis": true, "memory": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 65, 'YUNAI专用：你是一个拥有超强记忆力的AI角色，能够记住长时间的对话历史和复杂的背景信息。', 'public', 'standard');

-- 💻 代码生成模型 (编程助手角色)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

('Qwen/Qwen2.5-Coder-32B-Instruct', '通义千问 Coder 32B - 代码专家', 'siliconflow', 'code',
 '{"code": true, "programming": true, "debugging": true, "explanation": true}',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 90, 'YUNAI专用：你是一个专业的编程导师AI角色，精通多种编程语言，能够帮助用户解决编程问题。', 'public', 'free'),

('Qwen/Qwen3-Coder-480B-A35B-Instruct', '通义千问 Coder 480B - 超级编程', 'siliconflow', 'code',
 '{"code": true, "programming": true, "architecture": true, "optimization": true}',
 '{"input_token_price": 0.002, "output_token_price": 0.006, "unit": "1k_tokens", "currency": "CNY"}',
 85, 'YUNAI专用：你是一个资深的软件架构师AI角色，能够设计复杂的系统架构和优化代码性能。', 'public', 'premium');

-- 🔍 文本嵌入模型 (身份识别和推荐)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

('BAAI/bge-large-zh-v1.5', 'BGE Large 中文 - 身份识别', 'siliconflow', 'embedding',
 '{"embedding": true, "chinese": true, "identity_recognition": true, "similarity": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 95, '用于YUNAI用户身份识别和内容相似度匹配', 'public', 'free'),

('BAAI/bge-m3', 'BGE M3 - 多语言嵌入', 'siliconflow', 'embedding',
 '{"embedding": true, "multilingual": true, "cross_lingual": true, "retrieval": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 90, '用于YUNAI多语言内容理解和跨语言检索', 'public', 'free'),

('Qwen/Qwen3-Embedding-8B', '通义千问 嵌入 8B - 语义理解', 'siliconflow', 'embedding',
 '{"embedding": true, "semantic": true, "context_understanding": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 85, '用于YUNAI深度语义理解和上下文分析', 'public', 'free');

-- 📊 重排序模型 (搜索优化)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

('BAAI/bge-reranker-v2-m3', 'BGE Reranker V2 - 搜索优化', 'siliconflow', 'reranker',
 '{"reranking": true, "search_optimization": true, "relevance_scoring": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 95, '用于YUNAI搜索结果重排序和相关性评分', 'public', 'free'),

('netease-youdao/bce-reranker-base_v1', '网易有道 重排序 - 内容排序', 'siliconflow', 'reranker',
 '{"reranking": true, "content_sorting": true, "recommendation": true}',
 '{"input_token_price": 0.0001, "output_token_price": 0.0001, "unit": "1k_tokens", "currency": "CNY"}',
 90, '用于YUNAI内容推荐和智能排序', 'public', 'free');

-- 🎨 图像生成模型 (角色头像和朋友圈图片)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

('black-forest-labs/FLUX.1-schnell', 'FLUX.1 Schnell - 快速图像生成', 'siliconflow', 'image',
 '{"image_generation": true, "fast": true, "avatar": true, "scene": true}',
 '{"per_image_price": 0.003, "unit": "per_image", "currency": "CNY"}',
 95, '用于YUNAI快速生成AI角色头像、朋友圈图片和场景背景', 'public', 'free'),

('stabilityai/stable-diffusion-3-5-large', 'Stable Diffusion 3.5 - 高质量图像', 'siliconflow', 'image',
 '{"image_generation": true, "high_quality": true, "artistic": true, "detailed": true}',
 '{"per_image_price": 0.035, "unit": "per_image", "currency": "CNY"}',
 90, '用于YUNAI生成高质量的AI角色立绘和精美朋友圈图片', 'public', 'standard'),

('black-forest-labs/FLUX.1-pro', 'FLUX.1 Pro - 专业图像生成', 'siliconflow', 'image',
 '{"image_generation": true, "professional": true, "commercial": true, "premium": true}',
 '{"per_image_price": 0.055, "unit": "per_image", "currency": "CNY"}',
 85, '用于YUNAI生成商业级AI角色形象和高端朋友圈内容', 'public', 'premium'),

('Kwai-Kolors/Kolors', 'Kolors - 快手图像模型', 'siliconflow', 'image',
 '{"image_generation": true, "chinese_style": true, "social_media": true}',
 '{"per_image_price": 0.003, "unit": "per_image", "currency": "CNY"}',
 80, '用于YUNAI生成符合中文用户审美的社交媒体图片', 'public', 'free');

-- 🎙️ 语音处理模型 (语音通话和消息)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

('fishaudio/fish-speech-1.5', 'Fish Speech 1.5 - 自然语音合成', 'siliconflow', 'audio',
 '{"tts": true, "natural": true, "emotional": true, "voice_cloning": true}',
 '{"per_second_price": 0.002, "unit": "per_second", "currency": "CNY"}',
 95, '用于YUNAI AI角色的自然语音合成，支持情感表达和声音克隆', 'public', 'free'),

('FunAudioLLM/SenseVoiceSmall', 'SenseVoice Small - 语音识别', 'siliconflow', 'audio',
 '{"stt": true, "multilingual": true, "real_time": true, "noise_robust": true}',
 '{"per_second_price": 0.001, "unit": "per_second", "currency": "CNY"}',
 90, '用于YUNAI语音消息识别和实时语音通话转文字', 'public', 'free'),

('FunAudioLLM/CosyVoice2-0.5B', 'CosyVoice2 - 舒适语音', 'siliconflow', 'audio',
 '{"tts": true, "comfortable": true, "long_form": true, "stable": true}',
 '{"per_second_price": 0.0015, "unit": "per_second", "currency": "CNY"}',
 85, '用于YUNAI长时间语音通话，提供舒适稳定的语音体验', 'public', 'free');

-- 🎬 视频生成模型 (动态表情和短视频)
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

('Wan-AI/Wan2.1-T2V-14B-Turbo', 'Wan 2.1 Turbo - 快速视频生成', 'siliconflow', 'video',
 '{"video_generation": true, "text_to_video": true, "fast": true, "short_form": true}',
 '{"per_video_price": 0.2, "unit": "per_video", "currency": "CNY"}',
 90, '用于YUNAI快速生成AI角色动态表情和短视频内容', 'public', 'standard'),

('Wan-AI/Wan2.1-I2V-14B-720P', 'Wan 2.1 I2V - 图片转视频', 'siliconflow', 'video',
 '{"video_generation": true, "image_to_video": true, "720p": true, "animation": true}',
 '{"per_video_price": 0.25, "unit": "per_video", "currency": "CNY"}',
 85, '用于YUNAI将AI角色静态图片转换为动态视频', 'public', 'standard');

-- 🧠 特殊功能模型
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight, model_system_prompt, visibility, min_user_type) VALUES

-- 视觉理解模型
('Qwen/Qwen2.5-VL-72B-Instruct', '通义千问 VL 72B - 视觉理解', 'siliconflow', 'multimodal',
 '{"vision": true, "image_understanding": true, "ocr": true, "scene_analysis": true}',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 88, 'YUNAI专用：你能够理解和分析图片内容，为用户提供图片相关的对话和分析。', 'public', 'free'),

-- 思维链模型
('Qwen/Qwen3-235B-A22B-Thinking-2507', '通义千问 3 思维链 - 深度思考', 'siliconflow', 'reasoning',
 '{"thinking": true, "chain_of_thought": true, "step_by_step": true, "analysis": true}',
 '{"input_token_price": 0.002, "output_token_price": 0.006, "unit": "1k_tokens", "currency": "CNY"}',
 92, 'YUNAI专用：你是一个善于深度思考的AI角色，会展示你的思考过程，一步步分析问题。', 'public', 'premium'),

-- 长上下文模型
('Tongyi-Zhiwen/QwenLong-L1-32B', '通义千问 Long - 超长记忆', 'siliconflow', 'long_context',
 '{"long_context": true, "memory": true, "conversation_history": true, "relationship": true}',
 '{"input_token_price": 0.0005, "output_token_price": 0.0015, "unit": "1k_tokens", "currency": "CNY"}',
 87, 'YUNAI专用：你拥有超长的记忆能力，能够记住与用户的完整对话历史和关系发展。', 'public', 'standard');

-- 创建模型分类索引
CREATE INDEX IF NOT EXISTS idx_ai_models_category ON ai_models((capabilities->>'chat'));
CREATE INDEX IF NOT EXISTS idx_ai_models_pricing_tier ON ai_models(min_user_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_weight ON ai_models(weight DESC);
