-- Drop triggers
DROP TRIGGER IF EXISTS update_media_assets_updated_at ON media_assets;
DROP TRIGGER IF EXISTS update_media_templates_updated_at ON media_templates;
DROP TRIGGER IF EXISTS update_media_generation_tasks_updated_at ON media_generation_tasks;

-- Drop indexes
DROP INDEX IF EXISTS idx_character_media_media_role;
DROP INDEX IF EXISTS idx_character_media_media_asset_id;
DROP INDEX IF EXISTS idx_character_media_character_id;

DROP INDEX IF EXISTS idx_media_assets_created_at;
DROP INDEX IF EXISTS idx_media_assets_is_public;
DROP INDEX IF EXISTS idx_media_assets_asset_type;
DROP INDEX IF EXISTS idx_media_assets_task_id;
DROP INDEX IF EXISTS idx_media_assets_user_id;

DROP INDEX IF EXISTS idx_media_templates_is_featured;
DROP INDEX IF EXISTS idx_media_templates_is_public;
DROP INDEX IF EXISTS idx_media_templates_template_type;
DROP INDEX IF EXISTS idx_media_templates_user_id;

DROP INDEX IF EXISTS idx_media_generation_tasks_created_at;
DROP INDEX IF EXISTS idx_media_generation_tasks_task_type;
DROP INDEX IF EXISTS idx_media_generation_tasks_status;
DROP INDEX IF EXISTS idx_media_generation_tasks_user_id;

-- Drop tables
DROP TABLE IF EXISTS character_media;
DROP TABLE IF EXISTS media_assets;
DROP TABLE IF EXISTS media_templates;
DROP TABLE IF EXISTS media_generation_tasks;
