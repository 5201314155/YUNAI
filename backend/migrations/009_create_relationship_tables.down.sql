-- Drop triggers
DROP TRIGGER IF EXISTS update_trigger_rules_updated_at ON trigger_rules;
DROP TRIGGER IF EXISTS update_conversation_contexts_updated_at ON conversation_contexts;
DROP TRIGGER IF EXISTS update_memory_fragments_updated_at ON memory_fragments;
DROP TRIGGER IF EXISTS update_character_relationships_enhanced_updated_at ON character_relationships_enhanced;
DROP TRIGGER IF EXISTS update_relationship_types_updated_at ON relationship_types;

-- Drop indexes
DROP INDEX IF EXISTS idx_trigger_rules_is_active;
DROP INDEX IF EXISTS idx_trigger_rules_rule_type;
DROP INDEX IF EXISTS idx_trigger_rules_group_chat_id;

DROP INDEX IF EXISTS idx_conversation_contexts_last_activity;
DROP INDEX IF EXISTS idx_conversation_contexts_current_speaker;
DROP INDEX IF EXISTS idx_conversation_contexts_group_chat_id;

DROP INDEX IF EXISTS idx_memory_fragments_created_at;
DROP INDEX IF EXISTS idx_memory_fragments_importance_score;
DROP INDEX IF EXISTS idx_memory_fragments_memory_type;
DROP INDEX IF EXISTS idx_memory_fragments_character_id;

DROP INDEX IF EXISTS idx_relationship_events_created_at;
DROP INDEX IF EXISTS idx_relationship_events_event_type;
DROP INDEX IF EXISTS idx_relationship_events_relationship_id;

DROP INDEX IF EXISTS idx_character_relationships_enhanced_status;
DROP INDEX IF EXISTS idx_character_relationships_enhanced_strength;
DROP INDEX IF EXISTS idx_character_relationships_enhanced_type;
DROP INDEX IF EXISTS idx_character_relationships_enhanced_target;
DROP INDEX IF EXISTS idx_character_relationships_enhanced_source;

DROP INDEX IF EXISTS idx_relationship_types_user_id;
DROP INDEX IF EXISTS idx_relationship_types_category;

-- Drop tables
DROP TABLE IF EXISTS trigger_rules;
DROP TABLE IF EXISTS conversation_contexts;
DROP TABLE IF EXISTS memory_fragments;
DROP TABLE IF EXISTS relationship_events;
DROP TABLE IF EXISTS character_relationships_enhanced;
DROP TABLE IF EXISTS relationship_types;
