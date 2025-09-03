-- 创建角色审核相关表

-- 角色审核记录表
CREATE TABLE IF NOT EXISTS character_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    review_type VARCHAR(20) NOT NULL,
    score INTEGER DEFAULT 0,
    reason TEXT,
    auto_review BOOLEAN DEFAULT true,
    review_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP WITH TIME ZONE
);

-- 审核规则表
CREATE TABLE IF NOT EXISTS review_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL,
    pattern TEXT,
    action VARCHAR(50) NOT NULL,
    severity INTEGER NOT NULL DEFAULT 1,
    enabled BOOLEAN DEFAULT true,
    description TEXT,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 敏感词表
CREATE TABLE IF NOT EXISTS sensitive_words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(255) NOT NULL UNIQUE,
    category VARCHAR(100) NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    action VARCHAR(50) NOT NULL DEFAULT 'block',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 审核配置表
CREATE TABLE IF NOT EXISTS review_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auto_review_enabled BOOLEAN DEFAULT true,
    pass_threshold INTEGER DEFAULT 70,
    reject_threshold INTEGER DEFAULT 30,
    require_manual BOOLEAN DEFAULT false,
    review_timeout INTEGER DEFAULT 24,
    notify_reviewer BOOLEAN DEFAULT true,
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 审核队列表
CREATE TABLE IF NOT EXISTS review_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES character_reviews(id) ON DELETE CASCADE,
    priority INTEGER DEFAULT 3,
    queued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'queued',
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 审核员分配表
CREATE TABLE IF NOT EXISTS reviewer_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    review_id UUID NOT NULL REFERENCES character_reviews(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'assigned',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 审核日志表
CREATE TABLE IF NOT EXISTS review_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES character_reviews(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_type VARCHAR(20) NOT NULL DEFAULT 'user',
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 审核通知表
CREATE TABLE IF NOT EXISTS review_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES character_reviews(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    data JSONB,
    read BOOLEAN DEFAULT false,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_character_reviews_character_id ON character_reviews(character_id);
CREATE INDEX IF NOT EXISTS idx_character_reviews_status ON character_reviews(status);
CREATE INDEX IF NOT EXISTS idx_character_reviews_reviewer_id ON character_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_character_reviews_review_type ON character_reviews(review_type);
CREATE INDEX IF NOT EXISTS idx_character_reviews_created_at ON character_reviews(created_at);
CREATE INDEX IF NOT EXISTS idx_character_reviews_auto_review ON character_reviews(auto_review);

CREATE INDEX IF NOT EXISTS idx_review_rules_type ON review_rules(type);
CREATE INDEX IF NOT EXISTS idx_review_rules_category ON review_rules(category);
CREATE INDEX IF NOT EXISTS idx_review_rules_enabled ON review_rules(enabled);

CREATE INDEX IF NOT EXISTS idx_sensitive_words_word ON sensitive_words(word);
CREATE INDEX IF NOT EXISTS idx_sensitive_words_category ON sensitive_words(category);
CREATE INDEX IF NOT EXISTS idx_sensitive_words_enabled ON sensitive_words(enabled);

CREATE INDEX IF NOT EXISTS idx_review_queue_status ON review_queue(status);
CREATE INDEX IF NOT EXISTS idx_review_queue_priority ON review_queue(priority);
CREATE INDEX IF NOT EXISTS idx_review_queue_queued_at ON review_queue(queued_at);

CREATE INDEX IF NOT EXISTS idx_reviewer_assignments_reviewer_id ON reviewer_assignments(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_reviewer_assignments_review_id ON reviewer_assignments(review_id);
CREATE INDEX IF NOT EXISTS idx_reviewer_assignments_status ON reviewer_assignments(status);

CREATE INDEX IF NOT EXISTS idx_review_logs_review_id ON review_logs(review_id);
CREATE INDEX IF NOT EXISTS idx_review_logs_action ON review_logs(action);
CREATE INDEX IF NOT EXISTS idx_review_logs_created_at ON review_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_review_notifications_user_id ON review_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_review_notifications_review_id ON review_notifications(review_id);
CREATE INDEX IF NOT EXISTS idx_review_notifications_read ON review_notifications(read);

-- 插入默认审核规则
INSERT INTO review_rules (name, type, category, pattern, action, severity, enabled, description) VALUES
('Sensitive Word Check', 'content', 'sensitive_words', '', 'reject', 5, true, 'Check for sensitive words in content'),
('Content Length Check', 'content', 'length_check', '', 'warn', 2, true, 'Check if content length is appropriate'),
('Special Characters Check', 'content', 'special_chars', '', 'warn', 1, true, 'Check for excessive special characters'),
('Image Format Check', 'avatar', 'format_check', '', 'reject', 3, true, 'Check if image format is valid'),
('Name Length Check', 'name', 'length_check', '', 'warn', 2, true, 'Check if name length is appropriate')
ON CONFLICT DO NOTHING;

-- 插入默认敏感词
INSERT INTO sensitive_words (word, category, level, action, enabled) VALUES
('政治', 'political', 5, 'block', true),
('暴力', 'violence', 4, 'block', true),
('色情', 'adult', 5, 'block', true),
('赌博', 'gambling', 3, 'block', true),
('毒品', 'drugs', 5, 'block', true),
('恐怖', 'terrorism', 5, 'block', true),
('仇恨', 'hate', 4, 'block', true),
('歧视', 'discrimination', 4, 'block', true)
ON CONFLICT (word) DO NOTHING;

-- 插入默认审核配置
INSERT INTO review_configs (auto_review_enabled, pass_threshold, reject_threshold, require_manual, review_timeout, notify_reviewer) VALUES
(true, 70, 30, false, 24, true)
ON CONFLICT DO NOTHING;

-- 创建触发器更新updated_at字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为相关表创建触发器
CREATE TRIGGER update_character_reviews_updated_at BEFORE UPDATE ON character_reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_review_rules_updated_at BEFORE UPDATE ON review_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sensitive_words_updated_at BEFORE UPDATE ON sensitive_words
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_review_configs_updated_at BEFORE UPDATE ON review_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_review_queue_updated_at BEFORE UPDATE ON review_queue
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reviewer_assignments_updated_at BEFORE UPDATE ON reviewer_assignments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_review_notifications_updated_at BEFORE UPDATE ON review_notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加约束
ALTER TABLE character_reviews ADD CONSTRAINT check_status 
    CHECK (status IN ('pending', 'approved', 'rejected', 'revision'));

ALTER TABLE character_reviews ADD CONSTRAINT check_review_type 
    CHECK (review_type IN ('content', 'avatar', 'name', 'all'));

ALTER TABLE character_reviews ADD CONSTRAINT check_score_range 
    CHECK (score >= 0 AND score <= 100);

ALTER TABLE review_rules ADD CONSTRAINT check_severity_range 
    CHECK (severity >= 1 AND severity <= 10);

ALTER TABLE sensitive_words ADD CONSTRAINT check_level_range 
    CHECK (level >= 1 AND level <= 5);

ALTER TABLE review_queue ADD CONSTRAINT check_priority_range 
    CHECK (priority >= 1 AND priority <= 5);

-- 添加注释
COMMENT ON TABLE character_reviews IS 'Character review records';
COMMENT ON TABLE review_rules IS 'Review rules configuration';
COMMENT ON TABLE sensitive_words IS 'Sensitive words dictionary';
COMMENT ON TABLE review_configs IS 'Review system configuration';
COMMENT ON TABLE review_queue IS 'Review processing queue';
COMMENT ON TABLE reviewer_assignments IS 'Reviewer task assignments';
COMMENT ON TABLE review_logs IS 'Review operation logs';
COMMENT ON TABLE review_notifications IS 'Review notifications';

COMMENT ON COLUMN character_reviews.status IS 'Review status: pending, approved, rejected, revision';
COMMENT ON COLUMN character_reviews.review_type IS 'Review type: content, avatar, name, all';
COMMENT ON COLUMN character_reviews.score IS 'Review score (0-100)';
COMMENT ON COLUMN character_reviews.auto_review IS 'Whether this is an automatic review';

COMMENT ON COLUMN review_rules.type IS 'Rule type: content, avatar, name';
COMMENT ON COLUMN review_rules.category IS 'Rule category: sensitive_words, length_check, etc.';
COMMENT ON COLUMN review_rules.action IS 'Action to take: reject, warn, flag';
COMMENT ON COLUMN review_rules.severity IS 'Rule severity level (1-10)';

COMMENT ON COLUMN sensitive_words.level IS 'Sensitivity level (1-5)';
COMMENT ON COLUMN sensitive_words.action IS 'Action: block, replace, warn';

COMMENT ON COLUMN review_queue.priority IS 'Queue priority (1-5, 1=highest)';
COMMENT ON COLUMN review_queue.status IS 'Queue status: queued, processing, completed';
