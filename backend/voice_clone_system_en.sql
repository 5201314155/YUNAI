-- YUNAI Voice Clone Management System Database Structure

-- 1. Voice Clones Table (Store all voice information)
CREATE TABLE IF NOT EXISTS voice_clones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voice_id VARCHAR(100) NOT NULL UNIQUE,           -- SiliconFlow voice ID
    voice_name VARCHAR(100) NOT NULL,                -- User-set voice name
    voice_image_url TEXT,                            -- Voice display image URL
    voice_description TEXT,                          -- Voice description
    audio_sample_url TEXT,                           -- Audio sample URL for preview
    original_audio_url TEXT,                         -- Original uploaded audio URL
    
    -- Permission control
    is_public BOOLEAN DEFAULT false,                 -- Public or private
    creator_user_id UUID NOT NULL,                   -- Creator user ID
    
    -- Voice characteristics
    gender VARCHAR(10),                              -- Gender (male/female/neutral)
    age_range VARCHAR(20),                           -- Age range (child/young/adult/elder)
    language VARCHAR(10) DEFAULT 'zh',               -- Primary language
    accent VARCHAR(50),                              -- Accent characteristics
    emotion_tags JSONB DEFAULT '[]',                 -- Emotion tags ["gentle", "lively", "serious"]
    
    -- Technical parameters
    model_provider VARCHAR(50) DEFAULT 'siliconflow', -- Model provider
    clone_model VARCHAR(100),                        -- Clone model used
    quality_score DECIMAL(3,2) DEFAULT 0.0,         -- Voice quality score (0-10)
    clone_status VARCHAR(20) DEFAULT 'processing',   -- Clone status (processing/completed/failed)
    
    -- Usage statistics
    usage_count INTEGER DEFAULT 0,                  -- Usage count
    like_count INTEGER DEFAULT 0,                   -- Like count
    download_count INTEGER DEFAULT 0,               -- Download count
    
    -- Review status
    review_status VARCHAR(20) DEFAULT 'pending',    -- Review status (pending/approved/rejected)
    review_reason TEXT,                             -- Review reason
    reviewed_by UUID,                               -- Reviewer ID
    reviewed_at TIMESTAMP WITH TIME ZONE,          -- Review time
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Voice Usage Logs Table (Record who used which voice)
CREATE TABLE IF NOT EXISTS voice_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voice_id UUID NOT NULL,                         -- Voice ID
    user_id UUID NOT NULL,                          -- User ID
    character_id UUID,                              -- Character ID (if setting voice for character)
    usage_type VARCHAR(20) NOT NULL,                -- Usage type (character_creation/voice_test/audio_generation)
    usage_duration INTEGER DEFAULT 0,               -- Usage duration (seconds)
    generated_audio_url TEXT,                       -- Generated audio URL
    cost_amount DECIMAL(10,4) DEFAULT 0.0,         -- Cost amount
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE
);

-- 3. Voice Favorites Table (User favorite voices)
CREATE TABLE IF NOT EXISTS voice_favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,                          -- User ID
    voice_id UUID NOT NULL,                         -- Voice ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, voice_id),
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE
);

-- 4. Voice Reviews Table (User reviews for voices)
CREATE TABLE IF NOT EXISTS voice_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voice_id UUID NOT NULL,                         -- Voice ID
    user_id UUID NOT NULL,                          -- Reviewer ID
    rating INTEGER CHECK (rating >= 1 AND rating <= 5), -- Rating 1-5 stars
    review_text TEXT,                               -- Review content
    is_anonymous BOOLEAN DEFAULT false,             -- Anonymous review
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, voice_id),
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE
);

-- 5. Character Voices Table (Voices used by characters)
CREATE TABLE IF NOT EXISTS character_voices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL,                     -- Character ID
    voice_id UUID NOT NULL,                         -- Voice ID
    is_primary BOOLEAN DEFAULT true,                -- Is primary voice
    voice_settings JSONB DEFAULT '{}',              -- Voice settings {"speed": 1.0, "pitch": 1.0, "emotion": "neutral"}
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (voice_id) REFERENCES voice_clones(id) ON DELETE CASCADE
);

-- Create indexes for query optimization
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

-- Create voice management functions

-- 1. Get available voices for user (including public voices and own private voices)
CREATE OR REPLACE FUNCTION get_available_voices(p_user_id UUID)
RETURNS TABLE (
    voice_id UUID,
    voice_name VARCHAR(100),
    voice_image_url TEXT,
    voice_description TEXT,
    audio_sample_url TEXT,
    is_public BOOLEAN,
    is_own BOOLEAN,
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
        vc.quality_score,
        vc.usage_count,
        vc.like_count,
        vc.gender,
        vc.age_range,
        vc.emotion_tags,
        vc.created_at
    FROM voice_clones vc
    WHERE (vc.is_public = true AND vc.review_status = 'approved') 
       OR vc.creator_user_id = p_user_id
    AND vc.clone_status = 'completed'
    ORDER BY vc.is_public DESC, vc.quality_score DESC, vc.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 2. Log voice usage
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
    -- Insert usage record
    INSERT INTO voice_usage_logs (voice_id, user_id, character_id, usage_type, usage_duration, cost_amount)
    VALUES (p_voice_id, p_user_id, p_character_id, p_usage_type, p_usage_duration, p_cost_amount)
    RETURNING id INTO log_id;
    
    -- Update voice usage count
    UPDATE voice_clones 
    SET usage_count = usage_count + 1, updated_at = NOW()
    WHERE id = p_voice_id;
    
    RETURN log_id;
END;
$$ LANGUAGE plpgsql;

-- 3. Toggle voice privacy status
CREATE OR REPLACE FUNCTION toggle_voice_privacy(
    p_voice_id UUID,
    p_user_id UUID,
    p_is_public BOOLEAN
) RETURNS BOOLEAN AS $$
DECLARE
    voice_exists BOOLEAN;
BEGIN
    -- Check if voice exists and belongs to user
    SELECT EXISTS(
        SELECT 1 FROM voice_clones 
        WHERE id = p_voice_id AND creator_user_id = p_user_id
    ) INTO voice_exists;
    
    IF NOT voice_exists THEN
        RETURN FALSE;
    END IF;
    
    -- Update public status
    UPDATE voice_clones 
    SET is_public = p_is_public, 
        review_status = CASE WHEN p_is_public THEN 'pending' ELSE 'approved' END,
        updated_at = NOW()
    WHERE id = p_voice_id AND creator_user_id = p_user_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Insert sample voice data for testing
INSERT INTO voice_clones (
    voice_id, voice_name, voice_image_url, voice_description, audio_sample_url,
    is_public, creator_user_id, gender, age_range, language, emotion_tags,
    quality_score, clone_status, review_status
) 
SELECT 
    'sf_voice_001', 'Gentle Female Voice', '/images/voices/gentle_female.jpg', 
    'Gentle and sweet female voice, suitable for warm characters', '/audio/samples/gentle_female.mp3',
    true, u.id, 'female', 'young', 'zh', '["gentle", "sweet", "healing"]',
    8.5, 'completed', 'approved'
FROM users u LIMIT 1
WHERE NOT EXISTS (SELECT 1 FROM voice_clones WHERE voice_id = 'sf_voice_001');

INSERT INTO voice_clones (
    voice_id, voice_name, voice_image_url, voice_description, audio_sample_url,
    is_public, creator_user_id, gender, age_range, language, emotion_tags,
    quality_score, clone_status, review_status
) 
SELECT 
    'sf_voice_002', 'Magnetic Male Voice', '/images/voices/magnetic_male.jpg', 
    'Deep magnetic male voice, suitable for mature characters', '/audio/samples/magnetic_male.mp3',
    true, u.id, 'male', 'adult', 'zh', '["magnetic", "mature", "stable"]',
    9.0, 'completed', 'approved'
FROM users u LIMIT 1
WHERE NOT EXISTS (SELECT 1 FROM voice_clones WHERE voice_id = 'sf_voice_002');

INSERT INTO voice_clones (
    voice_id, voice_name, voice_image_url, voice_description, audio_sample_url,
    is_public, creator_user_id, gender, age_range, language, emotion_tags,
    quality_score, clone_status, review_status
) 
SELECT 
    'sf_voice_003', 'Lively Girl Voice', '/images/voices/lively_girl.jpg', 
    'Lively and cute girl voice, full of youthful energy', '/audio/samples/lively_girl.mp3',
    true, u.id, 'female', 'young', 'zh', '["lively", "cute", "youthful"]',
    8.8, 'completed', 'approved'
FROM users u LIMIT 1
WHERE NOT EXISTS (SELECT 1 FROM voice_clones WHERE voice_id = 'sf_voice_003');

INSERT INTO voice_clones (
    voice_id, voice_name, voice_image_url, voice_description, audio_sample_url,
    is_public, creator_user_id, gender, age_range, language, emotion_tags,
    quality_score, clone_status, review_status
) 
SELECT 
    'sf_voice_004', 'My Custom Voice', '/images/voices/custom_voice.jpg', 
    'Personal custom voice, only for own use', '/audio/samples/custom_voice.mp3',
    false, u.id, 'female', 'adult', 'zh', '["personal", "unique"]',
    7.5, 'completed', 'approved'
FROM users u LIMIT 1
WHERE NOT EXISTS (SELECT 1 FROM voice_clones WHERE voice_id = 'sf_voice_004');

-- Show creation results
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
