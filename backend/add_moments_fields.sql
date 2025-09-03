-- 添加朋友圈表缺失的字段

-- 检查并添加缺失的字段到moments表
DO $$ 
BEGIN
    -- 添加emotion字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='moments' AND column_name='emotion') THEN
        ALTER TABLE moments ADD COLUMN emotion VARCHAR(50);
    END IF;
    
    -- 添加share_count字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='moments' AND column_name='share_count') THEN
        ALTER TABLE moments ADD COLUMN share_count INTEGER DEFAULT 0;
    END IF;
    
    -- 添加location字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='moments' AND column_name='location') THEN
        ALTER TABLE moments ADD COLUMN location VARCHAR(255);
    END IF;
    
    -- 添加mood_status字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='moments' AND column_name='mood_status') THEN
        ALTER TABLE moments ADD COLUMN mood_status VARCHAR(100);
    END IF;
END $$;

-- 创建评论表（如果不存在）
CREATE TABLE IF NOT EXISTS moment_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL,
    user_id UUID,
    character_id UUID,
    content TEXT NOT NULL,
    reply_to_id UUID,
    like_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建点赞表（如果不存在）
CREATE TABLE IF NOT EXISTS moment_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL,
    user_id UUID,
    character_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插入一些测试数据
INSERT INTO moments (user_id, character_id, content, tags, emotion, visibility, location, mood_status) 
SELECT 
    u.id,
    c.id,
    '测试朋友圈 - 今天天气真好，心情也很棒！☀️',
    ARRAY['测试', '心情', '天气'],
    'happy',
    'public',
    '北京市朝阳区',
    '开心'
FROM users u, characters c 
WHERE u.id IS NOT NULL AND c.id IS NOT NULL
LIMIT 1;

INSERT INTO moments (user_id, character_id, content, tags, emotion, visibility, mood_status) 
SELECT 
    u.id,
    c.id,
    '测试朋友圈 - 刚刚完成了一个重要项目！💪',
    ARRAY['工作', '成就'],
    'excited',
    'public',
    '兴奋'
FROM users u, characters c 
WHERE u.id IS NOT NULL AND c.id IS NOT NULL
LIMIT 1;

INSERT INTO moments (user_id, character_id, content, tags, emotion, visibility, mood_status) 
SELECT 
    u.id,
    c.id,
    '测试朋友圈 - 深夜思考人生，有些感悟...',
    ARRAY['思考', '人生'],
    'thoughtful',
    'friends',
    '沉思'
FROM users u, characters c 
WHERE u.id IS NOT NULL AND c.id IS NOT NULL
LIMIT 1;
