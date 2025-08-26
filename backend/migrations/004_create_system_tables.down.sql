-- 删除AI模型表
DROP TRIGGER IF EXISTS update_ai_models_updated_at ON ai_models;
DROP TABLE IF EXISTS ai_models;

-- 删除模型提供商表
DROP TRIGGER IF EXISTS update_model_providers_updated_at ON model_providers;
DROP TABLE IF EXISTS model_providers;

-- 删除功能开关表
DROP TRIGGER IF EXISTS update_feature_flags_updated_at ON feature_flags;
DROP TABLE IF EXISTS feature_flags;
