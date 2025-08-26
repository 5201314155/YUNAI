-- Create payment cards table
CREATE TABLE IF NOT EXISTS payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Card information
    card_number VARCHAR(20) NOT NULL,        -- Masked card number for display
    card_number_hash VARCHAR(255) NOT NULL,  -- Hashed card number for storage
    card_type VARCHAR(10) NOT NULL CHECK (card_type IN ('debit', 'credit')),
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(10) NOT NULL,
    
    -- Cardholder information
    cardholder_name VARCHAR(100) NOT NULL,

    -- Email binding (security feature)
    bound_email VARCHAR(255),
    bound_at TIMESTAMP WITH TIME ZONE,
    email_verified BOOLEAN DEFAULT FALSE,

    -- Card balance
    balance DECIMAL(15,2) DEFAULT 0.00 CHECK (balance >= 0),
    currency VARCHAR(10) DEFAULT 'coins',

    -- Status management
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Verification
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(user_id, card_number_hash)
);

-- Payment card indexes
CREATE INDEX IF NOT EXISTS idx_payment_cards_user_id ON payment_cards(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_cards_card_number_hash ON payment_cards(card_number_hash);
CREATE INDEX IF NOT EXISTS idx_payment_cards_is_default ON payment_cards(is_default);
CREATE INDEX IF NOT EXISTS idx_payment_cards_is_active ON payment_cards(is_active);

-- Payment card update trigger
CREATE TRIGGER update_payment_cards_updated_at 
    BEFORE UPDATE ON payment_cards 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create payment passwords table
CREATE TABLE IF NOT EXISTS payment_passwords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Password information
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(32) NOT NULL,
    
    -- Security settings
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    last_failed_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(user_id)
);

-- Payment password indexes
CREATE INDEX IF NOT EXISTS idx_payment_passwords_user_id ON payment_passwords(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_passwords_is_active ON payment_passwords(is_active);

-- Payment password update trigger
CREATE TRIGGER update_payment_passwords_updated_at 
    BEFORE UPDATE ON payment_passwords 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create recharge orders table
CREATE TABLE IF NOT EXISTS recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Order information
    order_no VARCHAR(32) UNIQUE NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    coins_amount BIGINT NOT NULL CHECK (coins_amount > 0),
    exchange_rate INTEGER NOT NULL DEFAULT 10,
    
    -- Payment information
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('card', 'alipay', 'wechat')),
    payment_card_id UUID REFERENCES payment_cards(id),
    
    -- Discount information
    original_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    bonus_coins BIGINT NOT NULL DEFAULT 0,
    card_code VARCHAR(16),
    
    -- Order status
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'paid', 'refunded')),
    
    -- Third-party payment information
    third_party_order_no VARCHAR(64),
    payment_url TEXT,
    
    -- Failure information
    failure_reason TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '30 minutes')
);

-- Recharge order indexes
CREATE INDEX IF NOT EXISTS idx_recharge_orders_user_id ON recharge_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_order_no ON recharge_orders(order_no);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_status ON recharge_orders(status);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_payment_status ON recharge_orders(payment_status);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_payment_method ON recharge_orders(payment_method);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_created_at ON recharge_orders(created_at);
CREATE INDEX IF NOT EXISTS idx_recharge_orders_expired_at ON recharge_orders(expired_at);

-- Recharge order update trigger
CREATE TRIGGER update_recharge_orders_updated_at 
    BEFORE UPDATE ON recharge_orders 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create recharge packages table
CREATE TABLE IF NOT EXISTS recharge_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    coins_amount BIGINT NOT NULL CHECK (coins_amount > 0),
    bonus_coins BIGINT NOT NULL DEFAULT 0,
    discount_rate DECIMAL(5,4) NOT NULL DEFAULT 1.0000,
    
    -- Display settings
    is_popular BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Recharge package indexes
CREATE INDEX IF NOT EXISTS idx_recharge_packages_is_active ON recharge_packages(is_active);
CREATE INDEX IF NOT EXISTS idx_recharge_packages_sort_order ON recharge_packages(sort_order);

-- Recharge package update trigger
CREATE TRIGGER update_recharge_packages_updated_at 
    BEFORE UPDATE ON recharge_packages 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create admin recharge records table (for admin top-up operations)
CREATE TABLE IF NOT EXISTS admin_recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admin_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Recharge information
    coins_amount BIGINT NOT NULL CHECK (coins_amount > 0),
    reason TEXT NOT NULL,
    
    -- Admin verification
    admin_payment_verified BOOLEAN DEFAULT FALSE,
    admin_payment_verified_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Admin recharge record indexes
CREATE INDEX IF NOT EXISTS idx_admin_recharge_records_target_user_id ON admin_recharge_records(target_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_recharge_records_admin_user_id ON admin_recharge_records(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_recharge_records_created_at ON admin_recharge_records(created_at);

-- Create system config table
CREATE TABLE IF NOT EXISTS system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    config_type VARCHAR(20) NOT NULL DEFAULT 'string' CHECK (config_type IN ('int', 'float', 'string', 'bool')),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- System config indexes
CREATE INDEX IF NOT EXISTS idx_system_configs_config_key ON system_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_system_configs_is_active ON system_configs(is_active);

-- System config update trigger
CREATE TRIGGER update_system_configs_updated_at
    BEFORE UPDATE ON system_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default system configs
INSERT INTO system_configs (config_key, config_value, description, config_type) VALUES
('coin_exchange_rate', '10', '金币汇率：1元 = N金币', 'int'),
('min_recharge_amount', '1.00', '最小充值金额（元）', 'float'),
('max_recharge_amount', '10000.00', '最大充值金额（元）', 'float'),
('enable_email_binding', 'true', '是否启用邮箱绑定功能', 'bool'),
('card_bind_email_required', 'true', '绑卡是否必须验证邮箱', 'bool')
ON CONFLICT (config_key) DO NOTHING;

-- Create card transactions table
CREATE TABLE IF NOT EXISTS card_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES payment_cards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Transaction details
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('recharge', 'transfer_in', 'transfer_out', 'consume')),
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,

    -- Related information
    related_card_id UUID REFERENCES payment_cards(id),
    order_id UUID REFERENCES recharge_orders(id),
    description TEXT,

    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Card transaction indexes
CREATE INDEX IF NOT EXISTS idx_card_transactions_card_id ON card_transactions(card_id);
CREATE INDEX IF NOT EXISTS idx_card_transactions_user_id ON card_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_card_transactions_type ON card_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_card_transactions_created_at ON card_transactions(created_at);

-- Insert default recharge packages
INSERT INTO recharge_packages (name, amount, coins_amount, bonus_coins, discount_rate, is_popular, sort_order) VALUES
('体验包', 6.00, 60, 0, 1.0000, FALSE, 1),
('基础包', 30.00, 300, 30, 1.0000, FALSE, 2),
('超值包', 68.00, 680, 100, 1.0000, TRUE, 3),
('豪华包', 128.00, 1280, 280, 1.0000, TRUE, 4),
('至尊包', 328.00, 3280, 1000, 1.0000, FALSE, 5),
('终极包', 648.00, 6480, 2500, 1.0000, FALSE, 6)
ON CONFLICT DO NOTHING;

-- Function to generate order number
CREATE OR REPLACE FUNCTION generate_order_no() RETURNS TEXT AS $$
DECLARE
    order_no TEXT;
BEGIN
    -- Generate order number: YN + YYYYMMDD + 8-digit random number
    order_no := 'YN' || TO_CHAR(NOW(), 'YYYYMMDD') || LPAD(FLOOR(RANDOM() * 100000000)::TEXT, 8, '0');
    
    -- Check if order number already exists
    WHILE EXISTS (SELECT 1 FROM recharge_orders WHERE order_no = order_no) LOOP
        order_no := 'YN' || TO_CHAR(NOW(), 'YYYYMMDD') || LPAD(FLOOR(RANDOM() * 100000000)::TEXT, 8, '0');
    END LOOP;
    
    RETURN order_no;
END;
$$ LANGUAGE plpgsql;

-- Function to ensure only one default card per user
CREATE OR REPLACE FUNCTION ensure_single_default_card() RETURNS TRIGGER AS $$
BEGIN
    -- If setting this card as default, unset all other default cards for this user
    IF NEW.is_default = TRUE THEN
        UPDATE payment_cards 
        SET is_default = FALSE 
        WHERE user_id = NEW.user_id AND id != NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to ensure only one default card per user
CREATE TRIGGER ensure_single_default_card_trigger
    BEFORE INSERT OR UPDATE ON payment_cards
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_default_card();
