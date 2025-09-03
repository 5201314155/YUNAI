-- 修复所有外键约束，将users表引用改为core_users表

-- 删除所有指向users表的外键约束
ALTER TABLE group_chat_members DROP CONSTRAINT IF EXISTS group_chat_members_user_id_fkey;
ALTER TABLE user_sessions DROP CONSTRAINT IF EXISTS user_sessions_user_id_fkey;
ALTER TABLE moment_comments DROP CONSTRAINT IF EXISTS moment_comments_user_id_fkey;
ALTER TABLE group_chats DROP CONSTRAINT IF EXISTS group_chats_creator_user_id_fkey;
ALTER TABLE group_chat_messages DROP CONSTRAINT IF EXISTS group_chat_messages_sender_user_id_fkey;
ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS chat_messages_user_id_fkey;
ALTER TABLE moments DROP CONSTRAINT IF EXISTS moments_user_id_fkey;
ALTER TABLE moment_likes DROP CONSTRAINT IF EXISTS moment_likes_user_id_fkey;
ALTER TABLE moment_shares DROP CONSTRAINT IF EXISTS moment_shares_shared_by_user_id_fkey;
ALTER TABLE stories DROP CONSTRAINT IF EXISTS stories_user_id_fkey;
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_fkey;
ALTER TABLE world_settings DROP CONSTRAINT IF EXISTS world_settings_user_id_fkey;
ALTER TABLE voice_calls DROP CONSTRAINT IF EXISTS voice_calls_user_id_fkey;

-- 添加指向core_users表的外键约束
ALTER TABLE group_chat_members ADD CONSTRAINT group_chat_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE user_sessions ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE moment_comments ADD CONSTRAINT moment_comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE group_chats ADD CONSTRAINT group_chats_creator_user_id_fkey FOREIGN KEY (creator_user_id) REFERENCES core_users(id);
ALTER TABLE group_chat_messages ADD CONSTRAINT group_chat_messages_sender_user_id_fkey FOREIGN KEY (sender_user_id) REFERENCES core_users(id);
ALTER TABLE chat_messages ADD CONSTRAINT chat_messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE moments ADD CONSTRAINT moments_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE moment_likes ADD CONSTRAINT moment_likes_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE moment_shares ADD CONSTRAINT moment_shares_shared_by_user_id_fkey FOREIGN KEY (shared_by_user_id) REFERENCES core_users(id);
ALTER TABLE stories ADD CONSTRAINT stories_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE wallets ADD CONSTRAINT wallets_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE world_settings ADD CONSTRAINT world_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
ALTER TABLE voice_calls ADD CONSTRAINT voice_calls_user_id_fkey FOREIGN KEY (user_id) REFERENCES core_users(id);
