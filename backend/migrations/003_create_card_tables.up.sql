-- 创建卡密表
CREATE TABLE IF NOT EXISTS gift_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_code VARCHAR(50) UNIQUE NOT NULL, -- 16位纯数字格式，如：6688990012345678
    
    -- 卡密类型和价值
    card_type VARCHAR(20) NOT NULL CHECK (card_type IN ('discount', 'bonus', 'privilege', 'combo')),
    value DECIMAL(15,2) NOT NULL CHECK (value > 0), -- 面值
    currency VARCHAR(10) NOT NULL DEFAULT '金币',
    
    -- 折扣设置（仅discount和combo类型）
    discount_rate DECIMAL(5,4) DEFAULT 1.0000, -- 折扣率，0.9表示9折
    
    -- 返利设置（仅bonus和combo类型）
    bonus_rate DECIMAL(5,4) DEFAULT 0.0000, -- 返利率，0.2表示20%返利
    bonus_amount DECIMAL(15,2) DEFAULT 0.00, -- 固定返利金额
    
    -- 特权设置（仅privilege和combo类型）
    privileges JSONB DEFAULT '[]', -- 特权列表，如["txt2video", "hd_upscaler"]
    privilege_duration INTERVAL, -- 特权有效期，如'30 days'
    
    -- 使用限制
    max_uses INTEGER DEFAULT 1, -- 最大使用次数
    used_count INTEGER DEFAULT 0, -- 已使用次数
    min_user_type VARCHAR(20) DEFAULT 'basic', -- 最低用户类型要求
    
    -- 状态管理
    is_active BOOLEAN DEFAULT TRUE,
    is_used BOOLEAN DEFAULT FALSE,
    
    -- 有效期
    valid_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    
    -- 创建信息
    created_by UUID REFERENCES users(id),
    batch_id UUID, -- 批次ID，用于批量生成的卡密
    
    -- 备注
    description TEXT,
    internal_notes TEXT, -- 内部备注，用户不可见
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 卡密表索引
CREATE INDEX idx_gift_cards_card_code ON gift_cards(card_code);
CREATE INDEX idx_gift_cards_card_type ON gift_cards(card_type);
CREATE INDEX idx_gift_cards_is_active ON gift_cards(is_active);
CREATE INDEX idx_gift_cards_is_used ON gift_cards(is_used);
CREATE INDEX idx_gift_cards_valid_until ON gift_cards(valid_until);
CREATE INDEX idx_gift_cards_batch_id ON gift_cards(batch_id);
CREATE INDEX idx_gift_cards_created_at ON gift_cards(created_at);

-- 卡密更新时间触发器
CREATE TRIGGER update_gift_cards_updated_at 
    BEFORE UPDATE ON gift_cards 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 创建卡密使用记录表
CREATE TABLE IF NOT EXISTS card_usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES gift_cards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 使用详情
    original_amount DECIMAL(15,2) NOT NULL, -- 原始金额
    discount_rate DECIMAL(5,4) NOT NULL, -- 实际折扣率
    final_amount DECIMAL(15,2) NOT NULL, -- 最终金额
    bonus_amount DECIMAL(15,2) DEFAULT 0.00, -- 返利金额
    total_received DECIMAL(15,2) NOT NULL, -- 实际到账金额
    
    -- 特权信息
    privileges_granted JSONB DEFAULT '[]', -- 获得的特权
    privilege_expires_at TIMESTAMP WITH TIME ZONE, -- 特权过期时间
    
    -- 关联信息
    recharge_record_id UUID REFERENCES recharge_records(id), -- 关联的充值记录
    
    -- 时间戳
    used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 卡密使用记录索引
CREATE INDEX idx_card_usage_records_card_id ON card_usage_records(card_id);
CREATE INDEX idx_card_usage_records_user_id ON card_usage_records(user_id);
CREATE INDEX idx_card_usage_records_used_at ON card_usage_records(used_at);
CREATE INDEX idx_card_usage_records_recharge_record_id ON card_usage_records(recharge_record_id);

-- 创建用户特权表
CREATE TABLE IF NOT EXISTS user_privileges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    privilege VARCHAR(100) NOT NULL, -- 特权名称，如txt2video, hd_upscaler
    
    -- 来源信息
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('card', 'purchase', 'gift', 'promotion')),
    source_id UUID, -- 来源ID，如卡密ID
    
    -- 有效期
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    UNIQUE(user_id, privilege, source_id)
);

-- 用户特权表索引
CREATE INDEX idx_user_privileges_user_id ON user_privileges(user_id);
CREATE INDEX idx_user_privileges_privilege ON user_privileges(privilege);
CREATE INDEX idx_user_privileges_expires_at ON user_privileges(expires_at);
CREATE INDEX idx_user_privileges_is_active ON user_privileges(is_active);

-- 插入示例卡密（纯数字格式，类似银行卡）
INSERT INTO gift_cards (
    card_code, card_type, value, discount_rate, bonus_rate,
    privileges, privilege_duration, description, created_by
) VALUES
-- 折扣卡（9折优惠）
('6688990012345678', 'discount', 100.00, 0.9000, 0.0000,
 '[]', NULL, '9折优惠卡（测试）',
 (SELECT id FROM users WHERE username = 'admin')),

-- 返利卡（20%返利）
('6688990087654321', 'bonus', 100.00, 1.0000, 0.2000,
 '[]', NULL, '20%返利卡（测试）',
 (SELECT id FROM users WHERE username = 'admin')),

-- 特权卡（视频生成权限）
('6688990011112222', 'privilege', 0.00, 1.0000, 0.0000,
 '["txt2video", "hd_upscaler"]', '30 days', '视频生成特权卡（测试）',
 (SELECT id FROM users WHERE username = 'admin')),

-- 组合卡（超级权限）
('6688990099998888', 'combo', 200.00, 0.8000, 0.3000,
 '["txt2video", "hd_upscaler", "advanced_tts"]', '60 days', '超级组合卡（测试）',
 (SELECT id FROM users WHERE username = 'admin'))

ON CONFLICT (card_code) DO NOTHING;
