-- 插入SiliconFlow模型到数据库
-- 共102个模型，涵盖对话、代码、中文、嵌入、多模态等多个分类

-- 1. 图像生成模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'stabilityai/stable-diffusion-xl-base-1.0', 'Stable Diffusion XL Base 1.0', 'SiliconFlow', '图像生成', 'Stability AI的高质量图像生成模型，支持1024x1024分辨率', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 1, NOW(), NOW());

-- 2. 中文对话模型 - GLM系列
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'THUDM/glm-4-9b-chat', 'GLM-4 9B Chat', 'SiliconFlow', '中文模型', '清华大学GLM-4系列，9B参数的中英双语对话模型', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 2, NOW(), NOW());

-- 3. 通义千问系列
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen2-7B-Instruct', 'Qwen2 7B Instruct', 'SiliconFlow', '中文模型', '阿里巴巴通义千问2.0，7B参数指令微调模型', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 3, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen2.5-72B-Instruct', 'Qwen2.5 72B Instruct', 'SiliconFlow', '中文模型', '通义千问2.5旗舰版，72B参数，性能强劲', 131072, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 4, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen2.5-7B-Instruct', 'Qwen2.5 7B Instruct', 'SiliconFlow', '中文模型', '通义千问2.5轻量版，7B参数，高效实用', 32768, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 5, NOW(), NOW());

-- 4. 代码模型 - DeepSeek系列
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'deepseek-ai/DeepSeek-V2.5', 'DeepSeek V2.5', 'SiliconFlow', '代码模型', '深度求索V2.5，专业编程助手，支持多种编程语言', 32768, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 6, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'deepseek-ai/DeepSeek-V3', 'DeepSeek V3', 'SiliconFlow', '代码模型', '深度求索V3最新版本，顶级代码生成能力', 65536, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 7, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'deepseek-ai/DeepSeek-R1', 'DeepSeek R1', 'SiliconFlow', '代码模型', '深度求索R1推理模型，具备强大的逻辑推理能力', 32768, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 8, NOW(), NOW());

-- 5. 嵌入模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'BAAI/bge-large-en-v1.5', 'BGE Large EN v1.5', 'SiliconFlow', '嵌入模型', '智源BGE英文嵌入模型，适用于语义搜索和相似度计算', 512, '¥0.007/1K tokens', '¥0.007/1K tokens', true, 9, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'BAAI/bge-large-zh-v1.5', 'BGE Large ZH v1.5', 'SiliconFlow', '嵌入模型', '智源BGE中文嵌入模型，中文语义理解优秀', 512, '¥0.007/1K tokens', '¥0.007/1K tokens', true, 10, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'netease-youdao/bce-embedding-base_v1', 'BCE Embedding Base v1', 'SiliconFlow', '嵌入模型', '网易有道BCE嵌入模型，中英双语支持', 512, '¥0.007/1K tokens', '¥0.007/1K tokens', true, 11, NOW(), NOW());

-- 6. 多模态模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen2-VL-72B-Instruct', 'Qwen2 VL 72B', 'SiliconFlow', '多模态模型', '通义千问视觉语言模型，支持图像理解和对话', 32768, '¥0.021/1K tokens', '¥0.021/1K tokens', true, 12, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen2.5-VL-72B-Instruct', 'Qwen2.5 VL 72B', 'SiliconFlow', '多模态模型', '通义千问2.5视觉模型，图文理解能力更强', 32768, '¥0.021/1K tokens', '¥0.021/1K tokens', true, 13, NOW(), NOW());

-- 7. 图像生成模型 - FLUX系列
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'black-forest-labs/FLUX.1-schnell', 'FLUX.1 Schnell', 'SiliconFlow', '图像生成', 'Black Forest Labs快速图像生成模型', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 14, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'black-forest-labs/FLUX.1-dev', 'FLUX.1 Dev', 'SiliconFlow', '图像生成', 'FLUX.1开发版，高质量图像生成', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 15, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'black-forest-labs/FLUX.1-pro', 'FLUX.1 Pro', 'SiliconFlow', '图像生成', 'FLUX.1专业版，顶级图像生成质量', 8192, '¥0.028/1K tokens', '¥0.028/1K tokens', true, 16, NOW(), NOW());

-- 8. 语音模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'FunAudioLLM/SenseVoiceSmall', 'SenseVoice Small', 'SiliconFlow', '语音模型', '阿里巴巴语音识别模型，支持多语言', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 17, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'fishaudio/fish-speech-1.4', 'Fish Speech 1.4', 'SiliconFlow', '语音模型', 'Fish Audio语音合成模型v1.4', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 18, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'fishaudio/fish-speech-1.5', 'Fish Speech 1.5', 'SiliconFlow', '语音模型', 'Fish Audio语音合成模型v1.5，音质更佳', 8192, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 19, NOW(), NOW());

-- 9. 代码专用模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen2.5-Coder-7B-Instruct', 'Qwen2.5 Coder 7B', 'SiliconFlow', '代码模型', '通义千问代码专用模型，7B参数', 32768, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 20, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen2.5-Coder-32B-Instruct', 'Qwen2.5 Coder 32B', 'SiliconFlow', '代码模型', '通义千问代码专用模型，32B参数，性能更强', 32768, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 21, NOW(), NOW());

-- 10. 重排序模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'BAAI/bge-reranker-v2-m3', 'BGE Reranker v2 M3', 'SiliconFlow', '重排序模型', '智源BGE重排序模型，提升搜索结果质量', 512, '¥0.007/1K tokens', '¥0.007/1K tokens', true, 22, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'netease-youdao/bce-reranker-base_v1', 'BCE Reranker Base v1', 'SiliconFlow', '重排序模型', '网易有道重排序模型，优化检索结果', 512, '¥0.007/1K tokens', '¥0.007/1K tokens', true, 23, NOW(), NOW());

-- 11. 最新Qwen3系列
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen3-8B', 'Qwen3 8B', 'SiliconFlow', '中文模型', '通义千问3.0，8B参数基础模型', 32768, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 24, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen3-32B', 'Qwen3 32B', 'SiliconFlow', '中文模型', '通义千问3.0，32B参数高性能模型', 32768, '¥0.021/1K tokens', '¥0.021/1K tokens', true, 25, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Qwen/Qwen3-235B-A22B', 'Qwen3 235B A22B', 'SiliconFlow', '中文模型', '通义千问3.0旗舰版，235B参数顶级性能', 32768, '¥0.042/1K tokens', '¥0.042/1K tokens', true, 26, NOW(), NOW());

-- 12. 月之暗面Kimi系列
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'moonshotai/Kimi-Dev-72B', 'Kimi Dev 72B', 'SiliconFlow', '对话模型', '月之暗面Kimi开发版，72B参数', 200000, '¥0.021/1K tokens', '¥0.021/1K tokens', true, 27, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'moonshotai/Kimi-K2-Instruct', 'Kimi K2 Instruct', 'SiliconFlow', '对话模型', '月之暗面Kimi K2指令模型，长文本处理能力强', 200000, '¥0.021/1K tokens', '¥0.021/1K tokens', true, 28, NOW(), NOW());

-- 13. 腾讯混元模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'tencent/Hunyuan-A13B-Instruct', 'Hunyuan A13B Instruct', 'SiliconFlow', '中文模型', '腾讯混元大模型，13B参数中文优化', 32768, '¥0.014/1K tokens', '¥0.014/1K tokens', true, 29, NOW(), NOW());

-- 14. 百度文心模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'baidu/ERNIE-4.5-300B-A47B', 'ERNIE 4.5 300B A47B', 'SiliconFlow', '中文模型', '百度文心大模型4.5，300B参数企业级', 32768, '¥0.042/1K tokens', '¥0.042/1K tokens', true, 30, NOW(), NOW());

-- 15. 视频生成模型
INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Wan-AI/Wan2.1-T2V-14B', 'Wan2.1 T2V 14B', 'SiliconFlow', '视频生成', 'Wan AI文本到视频生成模型', 8192, '¥0.028/1K tokens', '¥0.028/1K tokens', true, 31, NOW(), NOW());

INSERT INTO ai_models (id, name, display_name, provider, category, description, max_tokens, input_price, output_price, is_active, sort_order, created_at, updated_at) VALUES (gen_random_uuid(), 'Wan-AI/Wan2.1-I2V-14B-720P', 'Wan2.1 I2V 14B 720P', 'SiliconFlow', '视频生成', 'Wan AI图像到视频生成模型，720P输出', 8192, '¥0.028/1K tokens', '¥0.028/1K tokens', true, 32, NOW(), NOW());
