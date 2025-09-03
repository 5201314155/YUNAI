-- YUNAI音色克隆管理系统数据库结构

-- 1. 音色库表 (存储所有音色信息)
CREATE TABLE IF NOT EXISTS voice_clones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voice_id VARCHAR(100) NOT NULL UNIQUE,           -- SiliconFlow返回的音色ID
    voice_name VARCHAR(100) NOT NULL,                -- 用户设置的音色名称
    voice_image_url TEXT,                            -- 音色展示图片URL
    voice_description TEXT,                          -- 音色描述
    audio_sample_url TEXT,                           -- 试听音频URL
    original_audio_url TEXT,                         -- 原始上传音频URL
    
    -- 权限控制
    is_public BOOLEAN DEFAULT false,                 -- 是否公开 (true=公开, false=私密)
    creator_user_id UUID NOT NULL,                   -- 创建者用户ID
    
    -- 音色特征
    gender VARCHAR(10),                              -- 性别 (male/female/neutral)
    age_range VARCHAR(20),                           -- 年龄段 (child/young/adult/elder)
    language VARCHAR(10) DEFAULT 'zh',               -- 主要语言
    accent VARCHAR(50),                              -- 口音特征
    emotion_tags JSONB DEFAULT '[]',                 -- 情感标签 ["温柔", "活泼", "严肃"]
    
    -- 技术参数
    model_provider VARCHAR(50) DEFAULT 'siliconflow', -- 模型提供商
    clone_model VARCHAR(100),                        -- 使用的克隆模型
    quality_score DECIMAL(3,2) DEFAULT 0.0,         -- 音色质量评分 (0-10)
    clone_status VARCHAR(20) DEFAULT 'processing',   -- 克隆状态 (processing/completed/failed)
    
    -- 使用统计
    usage_count INTEGER DEFAULT 0,                  -- 使用次数
    like_count INTEGER DEFAULT 0,                   -- 点赞数
    download_count INTEGER DEFAULT 0,               -- 下载次数
    
    -- 审核状态
    review_status VARCHAR(20) DEFAULT 'pending',    -- 审核状态 (pending/approved/rejected)
    review_reason TEXT,                             -- 审核原因
    reviewed_by UUID,                               -- 审核员ID
    reviewed_at TIMESTAMP WITH TIME ZONE,          -- 审核时间
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 外键约束
    FOREIGN KEY (creator_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 2. 音色使用记录表 (记录谁使用了哪个音色)
CREATE TABLE IF NOT EXISTS voice_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voice_id UUID NOT NULL,                         -- 音色ID
    user_id UUID NOT NULL,                          -- 使用者ID
    character_id UUID,                              -- 角色ID (如果是为角色设置音色)
    usage_type VARCHAR(20) NOT NULL,                -- 使用类型 (character_creation/voice_test/audio_generation)
    usage_duration INTEGER DEFAULT 0,               -- 使用时长(秒)
    generated_audio_url TEXT,                       -- 生成的音频URL
    cost_amount DECIMAL(10,4) DEFAULT 0.0,         -- 消费金额
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 3. 音色收藏表 (用户收藏的音色)
CREATE TABLE IF NOT EXISTS voice_favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,                          -- 用户ID
    voice_id UUID NOT NULL,                         -- 音色ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, voice_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE
);

-- 4. 音色评价表 (用户对音色的评价)
CREATE TABLE IF NOT EXISTS voice_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voice_id UUID NOT NULL,                         -- 音色ID
    user_id UUID NOT NULL,                          -- 评价者ID
    rating INTEGER CHECK (rating >= 1 AND rating <= 5), -- 评分 1-5星
    review_text TEXT,                               -- 评价内容
    is_anonymous BOOLEAN DEFAULT false,             -- 是否匿名评价
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, voice_id),
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 5. 角色音色关联表 (角色使用的音色)
CREATE TABLE IF NOT EXISTS character_voices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL,                     -- 角色ID
    voice_id UUID NOT NULL,                         -- 音色ID
    is_primary BOOLEAN DEFAULT true,                -- 是否为主要音色
    voice_settings JSONB DEFAULT '{}',              -- 音色设置 {"speed": 1.0, "pitch": 1.0, "emotion": "neutral"}
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE
);

-- 创建索引优化查询性能
CREATE INDEX IF NOT EXISTS idx_voice_clones_creator ON voice_clones(creator_user_id);
CREATE INDEX IF NOT EXISTS idx_voice_clones_public ON voice_clones(is_public, review_status);
CREATE INDEX IF NOT EXISTS idx_voice_clones_status ON voice_clones(clone_status);
CREATE INDEX IF NOT EXISTS idx_voice_clones_created ON voice_clones(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_voice_clones_usage ON voice_clones(usage_count DESC);
CREATE INDEX IF NOT EXISTS idx_voice_clones_quality ON voice_clones(quality_score DESC);

CREATE INDEX IF NOT EXISTS idx_voice_usage_user ON voice_usage_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_voice_usage_voice ON voice_usage_logs(voice_id);
CREATE INDEX IF NOT EXISTS idx_voice_usage_created ON voice_usage_logs(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_voice_favorites_user ON voice_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_voice_reviews_voice ON voice_reviews(voice_id);
CREATE INDEX IF NOT EXISTS idx_character_voices_character ON character_voices(character_id);

-- 创建音色管理相关函数

-- 1. 获取用户可用音色列表 (包括公开音色和自己的私密音色)
CREATE OR REPLACE FUNCTION get_available_voices(p_user_id UUID)
RETURNS TABLE (
    voice_id UUID,
    voice_name VARCHAR(100),
    voice_image_url TEXT,
    voice_description TEXT,
    audio_sample_url TEXT,
    is_public BOOLEAN,
    is_own BOOLEAN,
    creator_name VARCHAR(100),
    quality_score DECIMAL(3,2),
    usage_count INTEGER,
    like_count INTEGER,
    gender VARCHAR(10),
    age_range VARCHAR(20),
    emotion_tags JSONB,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        vc.id,
        vc.voice_name,
        vc.voice_image_url,
        vc.voice_description,
        vc.audio_sample_url,
        vc.is_public,
        (vc.creator_user_id = p_user_id) as is_own,
        u.username as creator_name,
        vc.quality_score,
        vc.usage_count,
        vc.like_count,
        vc.gender,
        vc.age_range,
        vc.emotion_tags,
        vc.created_at
    FROM voice_clones vc
    LEFT JOIN users u ON vc.creator_user_id = u.id
    WHERE (vc.is_public = true AND vc.review_status = 'approved') 
       OR vc.creator_user_id = p_user_id
    AND vc.clone_status = 'completed'
    ORDER BY vc.is_public DESC, vc.quality_score DESC, vc.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 2. 记录音色使用
CREATE OR REPLACE FUNCTION log_voice_usage(
    p_voice_id UUID,
    p_user_id UUID,
    p_character_id UUID DEFAULT NULL,
    p_usage_type VARCHAR(20) DEFAULT 'voice_test',
    p_usage_duration INTEGER DEFAULT 0,
    p_cost_amount DECIMAL(10,4) DEFAULT 0.0
) RETURNS UUID AS $$
DECLARE
    log_id UUID;
BEGIN
    -- 插入使用记录
    INSERT INTO voice_usage_logs (voice_id, user_id, character_id, usage_type, usage_duration, cost_amount)
    VALUES (p_voice_id, p_user_id, p_character_id, p_usage_type, p_usage_duration, p_cost_amount)
    RETURNING id INTO log_id;
    
    -- 更新音色使用次数
    UPDATE voice_clones 
    SET usage_count = usage_count + 1, updated_at = NOW()
    WHERE id = p_voice_id;
    
    RETURN log_id;
END;
$$ LANGUAGE plpgsql;

-- 3. 切换音色公开/私密状态
CREATE OR REPLACE FUNCTION toggle_voice_privacy(
    p_voice_id UUID,
    p_user_id UUID,
    p_is_public BOOLEAN
) RETURNS BOOLEAN AS $$
DECLARE
    voice_exists BOOLEAN;
BEGIN
    -- 检查音色是否存在且属于该用户
    SELECT EXISTS(
        SELECT 1 FROM voice_clones 
        WHERE id = p_voice_id AND creator_user_id = p_user_id
    ) INTO voice_exists;
    
    IF NOT voice_exists THEN
        RETURN FALSE;
    END IF;
    
    -- 更新公开状态
    UPDATE voice_clones 
    SET is_public = p_is_public, 
        review_status = CASE WHEN p_is_public THEN 'pending' ELSE 'approved' END,
        updated_at = NOW()
    WHERE id = p_voice_id AND creator_user_id = p_user_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 插入一些示例音色数据 (用于测试)
INSERT INTO voice_clones (
    voice_id, voice_name, voice_image_url, voice_description, audio_sample_url,
    is_public, creator_user_id, gender, age_range, language, emotion_tags,
    quality_score, clone_status, review_status
) VALUES 
-- 公开音色示例
('sf_voice_001', '温柔女声', '/images/voices/gentle_female.jpg', '温柔甜美的女性声音，适合温馨角色', '/audio/samples/gentle_female.mp3',
 true, (SELECT id FROM users LIMIT 1), 'female', 'young', 'zh', '["温柔", "甜美", "治愈"]',
 8.5, 'completed', 'approved'),

('sf_voice_002', '磁性男声', '/images/voices/magnetic_male.jpg', '低沉磁性的男性声音，适合成熟角色', '/audio/samples/magnetic_male.mp3',
 true, (SELECT id FROM users LIMIT 1), 'male', 'adult', 'zh', '["磁性", "成熟", "稳重"]',
 9.0, 'completed', 'approved'),

('sf_voice_003', '活泼少女', '/images/voices/lively_girl.jpg', '活泼可爱的少女声音，充满青春活力', '/audio/samples/lively_girl.mp3',
 true, (SELECT id FROM users LIMIT 1), 'female', 'young', 'zh', '["活泼", "可爱", "青春"]',
 8.8, 'completed', 'approved'),

-- 私密音色示例
('sf_voice_004', '我的专属音色', '/images/voices/custom_voice.jpg', '个人定制音色，仅自己可用', '/audio/samples/custom_voice.mp3',
 false, (SELECT id FROM users LIMIT 1), 'female', 'adult', 'zh', '["个性", "独特"]',
 7.5, 'completed', 'approved')

ON CONFLICT (voice_id) DO NOTHING;

-- 显示创建结果
SELECT 
    'voice_clones' as table_name,
    COUNT(*) as record_count
FROM voice_clones
UNION ALL
SELECT 
    'voice_usage_logs' as table_name,
    COUNT(*) as record_count
FROM voice_usage_logs
UNION ALL
SELECT 
    'voice_favorites' as table_name,
    COUNT(*) as record_count
FROM voice_favorites
UNION ALL
SELECT 
    'voice_reviews' as table_name,
    COUNT(*) as record_count
FROM voice_reviews
UNION ALL
SELECT 
    'character_voices' as table_name,
    COUNT(*) as record_count
FROM character_voices;
