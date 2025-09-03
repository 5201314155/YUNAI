-- YUNAI功能模块专用AI模型配置 (最终版)

-- 删除旧函数
DROP FUNCTION IF EXISTS get_module_model(character varying, character varying);

-- 插入模块配置
INSERT INTO module_model_configs (module_name, function_type, primary_model_id, fallback_models, model_params, weight, description) VALUES
('chat', 'conversation', 'sf_deepseek_v3', '["sf_glm4_9b", "sf_qwen25_72b"]', '{"temperature": 0.7, "max_tokens": 4000, "top_p": 0.9}', 100, 'Main chat conversation model'),
('relationship', 'embedding', 'sf_bge_zh', '["sf_bge_en", "sf_bge_m3"]', '{"dimensions": 1024, "normalize": true}', 100, 'Relationship network semantic embedding'),
('relationship', 'analysis', 'sf_deepseek_v3', '["sf_glm4_9b"]', '{"temperature": 0.6, "max_tokens": 3000}', 90, 'Relationship analysis and reasoning'),
('moments', 'generation', 'sf_glm4_9b', '["sf_deepseek_v3"]', '{"temperature": 0.8, "max_tokens": 500}', 100, 'Moments content generation'),
('character', 'avatar_generation', 'sf_flux_schnell', '["sf_flux_dev"]', '{"width": 1024, "height": 1024}', 100, 'Character avatar generation'),
('voice', 'synthesis', 'sf_fish15', '[]', '{"voice_id": "default", "speed": 1.0}', 100, 'Voice synthesis'),
('search', 'embedding', 'sf_bge_zh', '["sf_bge_en"]', '{"dimensions": 1024, "batch_size": 64}', 100, 'Search semantic embedding')
ON CONFLICT (module_name, function_type) DO UPDATE SET
    primary_model_id = EXCLUDED.primary_model_id,
    fallback_models = EXCLUDED.fallback_models,
    model_params = EXCLUDED.model_params,
    weight = EXCLUDED.weight,
    description = EXCLUDED.description,
    updated_at = NOW();

-- 创建新的模型选择函数
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
