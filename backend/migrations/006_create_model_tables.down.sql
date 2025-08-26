-- Drop triggers
DROP TRIGGER IF EXISTS update_model_usage_stats_updated_at ON model_usage_stats;
DROP TRIGGER IF EXISTS update_ai_models_updated_at ON ai_models;

-- Drop indexes
DROP INDEX IF EXISTS idx_model_usage_stats_date;
DROP INDEX IF EXISTS idx_model_usage_stats_user_id;
DROP INDEX IF EXISTS idx_model_usage_stats_model_id;
DROP INDEX IF EXISTS idx_model_health_logs_checked_at;
DROP INDEX IF EXISTS idx_model_health_logs_model_id;
DROP INDEX IF EXISTS idx_model_capabilities_capability;
DROP INDEX IF EXISTS idx_model_capabilities_model_id;
DROP INDEX IF EXISTS idx_ai_models_health_status;
DROP INDEX IF EXISTS idx_ai_models_is_active;
DROP INDEX IF EXISTS idx_ai_models_visibility;
DROP INDEX IF EXISTS idx_ai_models_model_type;
DROP INDEX IF EXISTS idx_ai_models_provider;

-- Drop tables
DROP TABLE IF EXISTS model_usage_stats;
DROP TABLE IF EXISTS model_health_logs;
DROP TABLE IF EXISTS model_capabilities;
DROP TABLE IF EXISTS ai_models;
