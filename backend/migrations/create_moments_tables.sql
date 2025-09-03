-- 🌟 YUNAI朋友圈功能数据库表结构
-- 创建朋友圈、评论、点赞相关的表

-- 朋友圈主表
CREATE TABLE IF NOT EXISTS moments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    images TEXT[], -- 图片URL数组
    tags TEXT[], -- 标签数组
    emotion VARCHAR(50), -- 情感状态
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    visibility VARCHAR(20) DEFAULT 'public', -- public, friends, private
    location VARCHAR(255), -- 位置信息
    mood_status VARCHAR(100), -- 心情状态
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 朋友圈评论表
CREATE TABLE IF NOT EXISTS moment_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    reply_to_id UUID REFERENCES moment_comments(id) ON DELETE CASCADE, -- 回复的评论ID
    like_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 朋友圈点赞表
CREATE TABLE IF NOT EXISTS moment_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(moment_id, user_id, character_id) -- 防止重复点赞
);

-- 朋友圈分享表
CREATE TABLE IF NOT EXISTS moment_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_moment_id UUID REFERENCES moments(id) ON DELETE CASCADE,
    shared_moment_id UUID REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    share_text TEXT, -- 分享时的文字
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_moments_user_id ON moments(user_id);
CREATE INDEX IF NOT EXISTS idx_moments_character_id ON moments(character_id);
CREATE INDEX IF NOT EXISTS idx_moments_created_at ON moments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_moments_visibility ON moments(visibility);

CREATE INDEX IF NOT EXISTS idx_moment_comments_moment_id ON moment_comments(moment_id);
CREATE INDEX IF NOT EXISTS idx_moment_comments_user_id ON moment_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_moment_comments_character_id ON moment_comments(character_id);

CREATE INDEX IF NOT EXISTS idx_moment_likes_moment_id ON moment_likes(moment_id);
CREATE INDEX IF NOT EXISTS idx_moment_likes_user_id ON moment_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_moment_likes_character_id ON moment_likes(character_id);

-- 添加触发器自动更新updated_at字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_moments_updated_at BEFORE UPDATE ON moments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_moment_comments_updated_at BEFORE UPDATE ON moment_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加一些示例数据
INSERT INTO moments (user_id, character_id, content, tags, emotion, visibility, location, mood_status) VALUES
(
    (SELECT id FROM users LIMIT 1),
    (SELECT id FROM characters LIMIT 1),
    '今天天气真好，心情也很棒！☀️',
    ARRAY['日常', '心情', '天气'],
    'happy',
    'public',
    '北京市朝阳区',
    '开心'
),
(
    (SELECT id FROM users LIMIT 1),
    (SELECT id FROM characters LIMIT 1),
    '刚刚完成了一个重要项目，感觉很有成就感！💪',
    ARRAY['工作', '成就', '项目'],
    'excited',
    'public',
    '',
    '兴奋'
),
(
    (SELECT id FROM users LIMIT 1),
    (SELECT id FROM characters LIMIT 1),
    '深夜思考人生，有些感悟想要分享...',
    ARRAY['思考', '人生', '感悟'],
    'thoughtful',
    'friends',
    '',
    '沉思'
);

-- 添加一些评论示例
INSERT INTO moment_comments (moment_id, user_id, character_id, content) VALUES
(
    (SELECT id FROM moments LIMIT 1),
    (SELECT id FROM users LIMIT 1),
    (SELECT id FROM characters LIMIT 1),
    '这个朋友圈很棒！👍'
),
(
    (SELECT id FROM moments LIMIT 1),
    (SELECT id FROM users LIMIT 1),
    (SELECT id FROM characters LIMIT 1),
    '我也有同感！'
);

-- 添加一些点赞示例
INSERT INTO moment_likes (moment_id, user_id, character_id) VALUES
(
    (SELECT id FROM moments LIMIT 1),
    (SELECT id FROM users LIMIT 1),
    (SELECT id FROM characters LIMIT 1)
);

COMMENT ON TABLE moments IS '朋友圈主表';
COMMENT ON TABLE moment_comments IS '朋友圈评论表';
COMMENT ON TABLE moment_likes IS '朋友圈点赞表';
COMMENT ON TABLE moment_shares IS '朋友圈分享表';
