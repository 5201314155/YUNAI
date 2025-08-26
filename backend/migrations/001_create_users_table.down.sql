-- 删除用户权限表
DROP TABLE IF EXISTS user_permissions;

-- 删除用户会话表
DROP TABLE IF EXISTS user_sessions;

-- 删除触发器和函数
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除用户表
DROP TABLE IF EXISTS users;
