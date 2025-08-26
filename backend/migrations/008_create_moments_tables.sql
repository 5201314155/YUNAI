-- 朋友圈相关表结构

-- 朋友圈动态表
CREATE TABLE moments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('text', 'image', 'video', 'talking_head', '3d')),
    media_url TEXT,
    media_type VARCHAR(50),
    visibility VARCHAR(10) NOT NULL CHECK (visibility IN ('public', 'friends', 'private')) DEFAULT 'friends',
    is_generated BOOLEAN NOT NULL DEFAULT true,
    mood VARCHAR(50),
    location VARCHAR(200),
    tags TEXT[], -- PostgreSQL数组类型
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    share_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 朋友圈草稿表
CREATE TABLE moment_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('text', 'image', 'video', 'talking_head', '3d')),
    media_prompt TEXT,
    visibility VARCHAR(10) NOT NULL CHECK (visibility IN ('public', 'friends', 'private')) DEFAULT 'friends',
    mood VARCHAR(50),
    tags TEXT[],
    priority INTEGER NOT NULL DEFAULT 0,
    similarity_score DECIMAL(3,2) NOT NULL DEFAULT 0.0,
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    is_published BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMP WITH TIME ZONE
);

-- 朋友圈互动表
CREATE TABLE moment_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('like', 'comment', 'share')),
    content TEXT, -- 评论内容，点赞和分享为空
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- 确保用户或角色至少有一个
    CONSTRAINT check_interaction_actor CHECK (
        (user_id IS NOT NULL AND character_id IS NULL) OR 
        (user_id IS NULL AND character_id IS NOT NULL)
    ),
    
    -- 防止重复点赞
    CONSTRAINT unique_like_per_actor UNIQUE (moment_id, user_id, character_id, type)
);

-- 朋友圈自动生成配置表
CREATE TABLE moment_auto_generation_configs (
    character_id UUID PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    frequency VARCHAR(10) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'custom')) DEFAULT 'daily',
    max_drafts_per_day INTEGER NOT NULL DEFAULT 5,
    auto_publish BOOLEAN NOT NULL DEFAULT false,
    content_types TEXT[] NOT NULL DEFAULT ARRAY['text'],
    visibility VARCHAR(10) NOT NULL CHECK (visibility IN ('public', 'friends', 'private')) DEFAULT 'friends',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 朋友圈通知表
CREATE TABLE moment_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('new_moment', 'like', 'comment', 'share')),
    content TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_moments_character_id ON moments(character_id);
CREATE INDEX idx_moments_user_id ON moments(user_id);
CREATE INDEX idx_moments_created_at ON moments(created_at DESC);
CREATE INDEX idx_moments_visibility ON moments(visibility);
CREATE INDEX idx_moments_content_type ON moments(content_type);

CREATE INDEX idx_moment_drafts_character_id ON moment_drafts(character_id);
CREATE INDEX idx_moment_drafts_user_id ON moment_drafts(user_id);
CREATE INDEX idx_moment_drafts_generated_at ON moment_drafts(generated_at DESC);
CREATE INDEX idx_moment_drafts_is_published ON moment_drafts(is_published);

CREATE INDEX idx_moment_interactions_moment_id ON moment_interactions(moment_id);
CREATE INDEX idx_moment_interactions_user_id ON moment_interactions(user_id);
CREATE INDEX idx_moment_interactions_character_id ON moment_interactions(character_id);
CREATE INDEX idx_moment_interactions_type ON moment_interactions(type);
CREATE INDEX idx_moment_interactions_created_at ON moment_interactions(created_at DESC);

CREATE INDEX idx_moment_notifications_user_id ON moment_notifications(user_id);
CREATE INDEX idx_moment_notifications_is_read ON moment_notifications(is_read);
CREATE INDEX idx_moment_notifications_created_at ON moment_notifications(created_at DESC);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_moments_updated_at 
    BEFORE UPDATE ON moments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_moment_auto_generation_configs_updated_at 
    BEFORE UPDATE ON moment_auto_generation_configs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建统计视图
CREATE VIEW moment_analytics AS
SELECT 
    m.character_id,
    m.user_id,
    COUNT(*) as total_moments,
    SUM(m.like_count) as total_likes,
    SUM(m.comment_count) as total_comments,
    SUM(m.share_count) as total_shares,
    AVG(m.like_count) as avg_likes_per_moment,
    MODE() WITHIN GROUP (ORDER BY m.mood) as most_popular_mood,
    EXTRACT(HOUR FROM MODE() WITHIN GROUP (ORDER BY m.created_at)) as most_active_hour,
    CASE 
        WHEN COUNT(*) > 0 THEN 
            (SUM(m.like_count) + SUM(m.comment_count) + SUM(m.share_count))::DECIMAL / COUNT(*) 
        ELSE 0 
    END as engagement_rate,
    MAX(m.created_at) as last_moment_at
FROM moments m
GROUP BY m.character_id, m.user_id;

-- 添加注释
COMMENT ON TABLE moments IS '朋友圈动态表';
COMMENT ON TABLE moment_drafts IS '朋友圈草稿表';
COMMENT ON TABLE moment_interactions IS '朋友圈互动表';
COMMENT ON TABLE moment_auto_generation_configs IS '朋友圈自动生成配置表';
COMMENT ON TABLE moment_notifications IS '朋友圈通知表';
COMMENT ON VIEW moment_analytics IS '朋友圈分析统计视图';

COMMENT ON COLUMN moments.content_type IS '内容类型: text, image, video, talking_head, 3d';
COMMENT ON COLUMN moments.visibility IS '可见性: public, friends, private';
COMMENT ON COLUMN moments.is_generated IS '是否为AI生成内容';
COMMENT ON COLUMN moments.tags IS '标签数组';

COMMENT ON COLUMN moment_drafts.similarity_score IS '与历史内容的相似度分数(0-1)';
COMMENT ON COLUMN moment_drafts.priority IS '优先级，数字越大优先级越高';

COMMENT ON COLUMN moment_interactions.type IS '互动类型: like, comment, share';

COMMENT ON COLUMN moment_auto_generation_configs.frequency IS '生成频率: daily, weekly, custom';
COMMENT ON COLUMN moment_auto_generation_configs.content_types IS '允许的内容类型数组';
