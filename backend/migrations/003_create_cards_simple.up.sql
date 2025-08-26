-- Create gift cards table
CREATE TABLE IF NOT EXISTS gift_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_code VARCHAR(50) UNIQUE NOT NULL, -- 16-digit format like: 6688990012345678
    
    -- Card type and value
    card_type VARCHAR(20) NOT NULL CHECK (card_type IN ('discount', 'bonus', 'privilege', 'combo')),
    value DECIMAL(15,2) NOT NULL CHECK (value > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'coins',
    
    -- Discount settings (for discount and combo types)
    discount_rate DECIMAL(5,4) DEFAULT 1.0000,
    
    -- Bonus settings (for bonus and combo types)
    bonus_rate DECIMAL(5,4) DEFAULT 0.0000,
    bonus_amount DECIMAL(15,2) DEFAULT 0.00,
    
    -- Privilege settings (for privilege and combo types)
    privileges JSONB DEFAULT '[]',
    privilege_duration INTERVAL,
    
    -- Usage limits
    max_uses INTEGER DEFAULT 1,
    used_count INTEGER DEFAULT 0,
    min_user_type VARCHAR(20) DEFAULT 'basic',
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_used BOOLEAN DEFAULT FALSE,
    
    -- Validity
    valid_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    
    -- Creation info
    created_by UUID REFERENCES users(id),
    batch_id UUID,
    
    -- Notes
    description TEXT,
    internal_notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Gift card indexes
CREATE INDEX IF NOT EXISTS idx_gift_cards_card_code ON gift_cards(card_code);
CREATE INDEX IF NOT EXISTS idx_gift_cards_card_type ON gift_cards(card_type);
CREATE INDEX IF NOT EXISTS idx_gift_cards_is_active ON gift_cards(is_active);
CREATE INDEX IF NOT EXISTS idx_gift_cards_is_used ON gift_cards(is_used);
CREATE INDEX IF NOT EXISTS idx_gift_cards_valid_until ON gift_cards(valid_until);
CREATE INDEX IF NOT EXISTS idx_gift_cards_batch_id ON gift_cards(batch_id);
CREATE INDEX IF NOT EXISTS idx_gift_cards_created_at ON gift_cards(created_at);

-- Gift card update trigger
CREATE TRIGGER update_gift_cards_updated_at 
    BEFORE UPDATE ON gift_cards 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create card usage records table
CREATE TABLE IF NOT EXISTS card_usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES gift_cards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Usage details
    original_amount DECIMAL(15,2) NOT NULL,
    discount_rate DECIMAL(5,4) NOT NULL,
    final_amount DECIMAL(15,2) NOT NULL,
    bonus_amount DECIMAL(15,2) DEFAULT 0.00,
    total_received DECIMAL(15,2) NOT NULL,
    
    -- Privilege info
    privileges_granted JSONB DEFAULT '[]',
    privilege_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Related info
    recharge_record_id UUID REFERENCES recharge_records(id),
    
    -- Timestamps
    used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Card usage record indexes
CREATE INDEX IF NOT EXISTS idx_card_usage_records_card_id ON card_usage_records(card_id);
CREATE INDEX IF NOT EXISTS idx_card_usage_records_user_id ON card_usage_records(user_id);
CREATE INDEX IF NOT EXISTS idx_card_usage_records_used_at ON card_usage_records(used_at);
CREATE INDEX IF NOT EXISTS idx_card_usage_records_recharge_record_id ON card_usage_records(recharge_record_id);

-- Create user privileges table
CREATE TABLE IF NOT EXISTS user_privileges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    privilege VARCHAR(100) NOT NULL,
    
    -- Source info
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('card', 'purchase', 'gift', 'promotion')),
    source_id UUID,
    
    -- Validity
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    
    UNIQUE(user_id, privilege, source_id)
);

-- User privilege indexes
CREATE INDEX IF NOT EXISTS idx_user_privileges_user_id ON user_privileges(user_id);
CREATE INDEX IF NOT EXISTS idx_user_privileges_privilege ON user_privileges(privilege);
CREATE INDEX IF NOT EXISTS idx_user_privileges_expires_at ON user_privileges(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_privileges_is_active ON user_privileges(is_active);

-- Insert sample gift cards (16-digit pure numbers)
INSERT INTO gift_cards (
    card_code, card_type, value, discount_rate, bonus_rate, 
    privileges, privilege_duration, description, created_by
) VALUES 
-- Discount card (10% off)
('6688990012345678', 'discount', 100.00, 0.9000, 0.0000, 
 '[]', NULL, '10% discount card (test)', 
 (SELECT id FROM users WHERE username = 'admin')),

-- Bonus card (20% bonus)
('6688990087654321', 'bonus', 100.00, 1.0000, 0.2000, 
 '[]', NULL, '20% bonus card (test)', 
 (SELECT id FROM users WHERE username = 'admin')),

-- Privilege card (video generation)
('6688990011112222', 'privilege', 0.00, 1.0000, 0.0000, 
 '["txt2video", "hd_upscaler"]', '30 days', 'Video generation privilege card (test)', 
 (SELECT id FROM users WHERE username = 'admin')),

-- Combo card (super privileges)
('6688990099998888', 'combo', 200.00, 0.8000, 0.3000, 
 '["txt2video", "hd_upscaler", "advanced_tts"]', '60 days', 'Super combo card (test)', 
 (SELECT id FROM users WHERE username = 'admin'))

ON CONFLICT (card_code) DO NOTHING;
