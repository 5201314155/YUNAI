-- YUNAI Payment System Tables

-- 1. Payment Cards Table
CREATE TABLE IF NOT EXISTS payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_number VARCHAR(20) NOT NULL,
    card_number_hash VARCHAR(64) NOT NULL,
    card_type VARCHAR(20) NOT NULL DEFAULT 'debit',
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(20) NOT NULL DEFAULT 'YUNAI',
    cardholder_name VARCHAR(100) NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    discount_rate DECIMAL(5,4) NOT NULL DEFAULT 1.0000,
    bonus_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_frozen BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, card_number),
    UNIQUE(card_number_hash)
);

-- 2. Wallets Table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'COINS',
    is_frozen BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Wallet Transactions Table
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2),
    balance_after DECIMAL(15,2),
    description TEXT,
    reference_id UUID,
    reference_type VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Admin Recharge Records Table
CREATE TABLE IF NOT EXISTS admin_recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admin_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id UUID REFERENCES payment_cards(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    coins_amount BIGINT NOT NULL,
    reason TEXT,
    admin_payment_verified BOOLEAN NOT NULL DEFAULT false,
    admin_payment_verified_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Card Recharge Records Table
CREATE TABLE IF NOT EXISTS card_recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES payment_cards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admin_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    recharge_type VARCHAR(20) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 6. Coin Exchange Records Table
CREATE TABLE IF NOT EXISTS coin_exchange_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id UUID NOT NULL REFERENCES payment_cards(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    exchange_amount DECIMAL(15,2) NOT NULL,
    actual_cost DECIMAL(15,2) NOT NULL,
    base_coins DECIMAL(15,2) NOT NULL,
    bonus_coins DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_coins DECIMAL(15,2) NOT NULL,
    discount_rate DECIMAL(5,4) NOT NULL,
    bonus_rate DECIMAL(5,4) NOT NULL,
    exchange_rate DECIMAL(10,4) NOT NULL DEFAULT 10.0000,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 7. AI Service Consumption Table
CREATE TABLE IF NOT EXISTS ai_service_consumption (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    service_type VARCHAR(50) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    model_id VARCHAR(100),
    cost DECIMAL(15,2) NOT NULL,
    tokens_used INTEGER,
    request_data JSONB,
    response_data JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create Indexes
CREATE INDEX IF NOT EXISTS idx_payment_cards_user_id ON payment_cards(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_cards_card_number ON payment_cards(card_number);
CREATE INDEX IF NOT EXISTS idx_payment_cards_is_active ON payment_cards(is_active);
CREATE INDEX IF NOT EXISTS idx_payment_cards_is_default ON payment_cards(is_default);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_id ON wallet_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at);

CREATE INDEX IF NOT EXISTS idx_admin_recharge_target_user ON admin_recharge_records(target_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_recharge_admin_user ON admin_recharge_records(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_recharge_created_at ON admin_recharge_records(created_at);

CREATE INDEX IF NOT EXISTS idx_card_recharge_card_id ON card_recharge_records(card_id);
CREATE INDEX IF NOT EXISTS idx_card_recharge_user_id ON card_recharge_records(user_id);

CREATE INDEX IF NOT EXISTS idx_coin_exchange_user_id ON coin_exchange_records(user_id);
CREATE INDEX IF NOT EXISTS idx_coin_exchange_card_id ON coin_exchange_records(card_id);
CREATE INDEX IF NOT EXISTS idx_coin_exchange_created_at ON coin_exchange_records(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_service_user_id ON ai_service_consumption(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_service_type ON ai_service_consumption(service_type);
CREATE INDEX IF NOT EXISTS idx_ai_service_created_at ON ai_service_consumption(created_at);

-- Add triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_payment_cards_updated_at BEFORE UPDATE ON payment_cards FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create wallets for existing users
INSERT INTO wallets (user_id, balance) 
SELECT id, 0.00 FROM users 
WHERE id NOT IN (SELECT user_id FROM wallets WHERE user_id IS NOT NULL)
ON CONFLICT (user_id) DO NOTHING;

SELECT 'YUNAI Payment System Tables Created Successfully!' as result;
