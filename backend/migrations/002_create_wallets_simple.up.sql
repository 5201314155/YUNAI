-- Create wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'coins',
    exchange_rate INTEGER NOT NULL DEFAULT 10,
    
    -- Limits
    daily_limit DECIMAL(15,2) DEFAULT 10000.00,
    monthly_limit DECIMAL(15,2) DEFAULT 100000.00,
    daily_spent DECIMAL(15,2) DEFAULT 0.00,
    monthly_spent DECIMAL(15,2) DEFAULT 0.00,
    last_reset_date DATE DEFAULT CURRENT_DATE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id)
);

-- Wallet indexes
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_balance ON wallets(balance);

-- Wallet update trigger
CREATE TRIGGER update_wallets_updated_at 
    BEFORE UPDATE ON wallets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create recharge records table
CREATE TABLE IF NOT EXISTS recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- Recharge info
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'coins',
    payment_amount DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0),
    payment_currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    
    -- Payment info
    payment_method VARCHAR(50) NOT NULL,
    payment_id VARCHAR(255),
    transaction_id VARCHAR(255) UNIQUE,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded')),
    
    -- Discount and bonus
    discount_rate DECIMAL(5,4) DEFAULT 1.0000,
    bonus_amount DECIMAL(15,2) DEFAULT 0.00,
    card_code VARCHAR(50),
    
    -- Notes
    remark TEXT,
    failure_reason TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Recharge record indexes
CREATE INDEX IF NOT EXISTS idx_recharge_records_user_id ON recharge_records(user_id);
CREATE INDEX IF NOT EXISTS idx_recharge_records_wallet_id ON recharge_records(wallet_id);
CREATE INDEX IF NOT EXISTS idx_recharge_records_status ON recharge_records(status);
CREATE INDEX IF NOT EXISTS idx_recharge_records_payment_method ON recharge_records(payment_method);
CREATE INDEX IF NOT EXISTS idx_recharge_records_created_at ON recharge_records(created_at);
CREATE INDEX IF NOT EXISTS idx_recharge_records_transaction_id ON recharge_records(transaction_id);

-- Recharge record update trigger
CREATE TRIGGER update_recharge_records_updated_at 
    BEFORE UPDATE ON recharge_records 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create consumption records table
CREATE TABLE IF NOT EXISTS consumption_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- Consumption info
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'coins',
    
    -- Service info
    service_type VARCHAR(50) NOT NULL,
    service_id UUID,
    model_name VARCHAR(100),
    
    -- Billing details
    billing_unit VARCHAR(20),
    billing_quantity DECIMAL(15,4),
    unit_price DECIMAL(15,6),
    
    -- Description
    description TEXT,
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Consumption record indexes
CREATE INDEX IF NOT EXISTS idx_consumption_records_user_id ON consumption_records(user_id);
CREATE INDEX IF NOT EXISTS idx_consumption_records_wallet_id ON consumption_records(wallet_id);
CREATE INDEX IF NOT EXISTS idx_consumption_records_service_type ON consumption_records(service_type);
CREATE INDEX IF NOT EXISTS idx_consumption_records_model_name ON consumption_records(model_name);
CREATE INDEX IF NOT EXISTS idx_consumption_records_created_at ON consumption_records(created_at);

-- Create wallet for admin user
INSERT INTO wallets (user_id, balance) 
SELECT id, 100000.00 
FROM users 
WHERE username = 'admin' 
ON CONFLICT (user_id) DO NOTHING;
