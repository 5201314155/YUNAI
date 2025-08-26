-- 删除触发器
DROP TRIGGER IF EXISTS update_chat_messages_updated_at ON chat_messages;
DROP TRIGGER IF EXISTS update_group_chats_updated_at ON group_chats;
DROP TRIGGER IF EXISTS update_character_relationships_updated_at ON character_relationships;
DROP TRIGGER IF EXISTS update_characters_updated_at ON characters;

-- 删除索引
DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_chat_messages_sender_character_id;
DROP INDEX IF EXISTS idx_chat_messages_group_chat_id;

DROP INDEX IF EXISTS idx_group_chat_members_user_id;
DROP INDEX IF EXISTS idx_group_chat_members_character_id;
DROP INDEX IF EXISTS idx_group_chat_members_group_chat_id;

DROP INDEX IF EXISTS idx_group_chats_is_public;
DROP INDEX IF EXISTS idx_group_chats_creator_user_id;

DROP INDEX IF EXISTS idx_character_relationships_target_character_id;
DROP INDEX IF EXISTS idx_character_relationships_character_id;

DROP INDEX IF EXISTS idx_character_tags_tag;
DROP INDEX IF EXISTS idx_character_tags_character_id;

DROP INDEX IF EXISTS idx_characters_created_at;
DROP INDEX IF EXISTS idx_characters_is_featured;
DROP INDEX IF EXISTS idx_characters_visibility;
DROP INDEX IF EXISTS idx_characters_user_id;

-- 删除表
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS group_chat_members;
DROP TABLE IF EXISTS group_chats;
DROP TABLE IF EXISTS character_relationships;
DROP TABLE IF EXISTS character_tags;
DROP TABLE IF EXISTS characters;
