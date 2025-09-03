-- 🎯 YUNAI充值卡系统数据库表结构
-- 创建支付系统相关的所有表

-- 1. 支付卡片表
CREATE TABLE IF NOT EXISTS payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_number VARCHAR(20) NOT NULL,
    card_number_hash VARCHAR(64) NOT NULL,
    card_type VARCHAR(20) NOT NULL DEFAULT 'debit', -- debit, credit
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(20) NOT NULL DEFAULT 'YUNAI',
    cardholder_name VARCHAR(100) NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    discount_rate DECIMAL(5,4) NOT NULL DEFAULT 1.0000, -- 折扣率，1.0=无折扣，0.9=9折
    bonus_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0000,    -- 返利率，0.2=20%返利
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_frozen BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, card_number),
    UNIQUE(card_number_hash)
);

-- 2. 钱包表
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'COINS',
    is_frozen BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. 钱包交易记录表
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL, -- recharge, consume, transfer, refund
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2),
    balance_after DECIMAL(15,2),
    description TEXT,
    reference_id UUID, -- 关联的订单或操作ID
    reference_type VARCHAR(50), -- 关联类型：ai_service, card_recharge, etc.
    status VARCHAR(20) NOT NULL DEFAULT 'completed', -- pending, completed, failed, cancelled
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. 管理员充值记录表
CREATE TABLE IF NOT EXISTS admin_recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admin_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id UUID REFERENCES payment_cards(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    coins_amount BIGINT NOT NULL, -- 对应的金币数量
    reason TEXT,
    admin_payment_verified BOOLEAN NOT NULL DEFAULT false,
    admin_payment_verified_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. 卡片充值记录表
CREATE TABLE IF NOT EXISTS card_recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES payment_cards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admin_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    recharge_type VARCHAR(20) NOT NULL, -- admin, user, system
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 6. 金币兑换记录表
CREATE TABLE IF NOT EXISTS coin_exchange_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id UUID NOT NULL REFERENCES payment_cards(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    exchange_amount DECIMAL(15,2) NOT NULL, -- 兑换的人民币金额
    actual_cost DECIMAL(15,2) NOT NULL,     -- 实际扣费金额（应用折扣后）
    base_coins DECIMAL(15,2) NOT NULL,      -- 基础获得金币
    bonus_coins DECIMAL(15,2) NOT NULL DEFAULT 0.00, -- 返利金币
    total_coins DECIMAL(15,2) NOT NULL,     -- 总获得金币
    discount_rate DECIMAL(5,4) NOT NULL,    -- 使用的折扣率
    bonus_rate DECIMAL(5,4) NOT NULL,       -- 使用的返利率
    exchange_rate DECIMAL(10,4) NOT NULL DEFAULT 10.0000, -- 汇率：1元=10金币
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 7. AI服务消费记录表
CREATE TABLE IF NOT EXISTS ai_service_consumption (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    service_type VARCHAR(50) NOT NULL, -- chat, image, voice, video, etc.
    service_name VARCHAR(100) NOT NULL,
    model_id VARCHAR(100),
    cost DECIMAL(15,2) NOT NULL,
    tokens_used INTEGER,
    request_data JSONB,
    response_data JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
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

-- 添加触发器更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_payment_cards_updated_at BEFORE UPDATE ON payment_cards FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入一些测试数据
-- 注意：这里只是为了测试，实际生产环境中不应该有这些数据

-- 为现有用户创建钱包
INSERT INTO wallets (user_id, balance) 
SELECT id, 0.00 FROM users 
WHERE id NOT IN (SELECT user_id FROM wallets)
ON CONFLICT (user_id) DO NOTHING;

COMMENT ON TABLE payment_cards IS 'YUNAI虚拟银行卡表';
COMMENT ON TABLE wallets IS 'YUNAI用户钱包表';
COMMENT ON TABLE wallet_transactions IS 'YUNAI钱包交易记录表';
COMMENT ON TABLE admin_recharge_records IS 'YUNAI管理员充值记录表';
COMMENT ON TABLE card_recharge_records IS 'YUNAI卡片充值记录表';
COMMENT ON TABLE coin_exchange_records IS 'YUNAI金币兑换记录表';
COMMENT ON TABLE ai_service_consumption IS 'YUNAI AI服务消费记录表';

-- 显示创建结果
SELECT 'YUNAI充值卡系统数据库表创建完成！' as result;
