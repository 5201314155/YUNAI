-- 创建钱包表
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    currency VARCHAR(10) NOT NULL DEFAULT '金币',
    exchange_rate INTEGER NOT NULL DEFAULT 10, -- 1元 = 10金币
    
    -- 限额管理
    daily_limit DECIMAL(15,2) DEFAULT 10000.00,
    monthly_limit DECIMAL(15,2) DEFAULT 100000.00,
    daily_spent DECIMAL(15,2) DEFAULT 0.00,
    monthly_spent DECIMAL(15,2) DEFAULT 0.00,
    last_reset_date DATE DEFAULT CURRENT_DATE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id)
);

-- 钱包表索引
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_balance ON wallets(balance);

-- 钱包更新时间触发器
CREATE TRIGGER update_wallets_updated_at 
    BEFORE UPDATE ON wallets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 创建充值记录表
CREATE TABLE IF NOT EXISTS recharge_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- 充值信息
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT '金币',
    payment_amount DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0), -- 实际支付金额
    payment_currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    
    -- 支付信息
    payment_method VARCHAR(50) NOT NULL, -- alipay, wechat, card等
    payment_id VARCHAR(255), -- 第三方支付订单号
    transaction_id VARCHAR(255) UNIQUE, -- 内部交易号
    
    -- 状态管理
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded')),
    
    -- 折扣和返利
    discount_rate DECIMAL(5,4) DEFAULT 1.0000, -- 折扣率，1.0为无折扣
    bonus_amount DECIMAL(15,2) DEFAULT 0.00, -- 返利金额
    card_code VARCHAR(50), -- 使用的卡密
    
    -- 备注信息
    remark TEXT,
    failure_reason TEXT,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- 充值记录索引
CREATE INDEX idx_recharge_records_user_id ON recharge_records(user_id);
CREATE INDEX idx_recharge_records_wallet_id ON recharge_records(wallet_id);
CREATE INDEX idx_recharge_records_status ON recharge_records(status);
CREATE INDEX idx_recharge_records_payment_method ON recharge_records(payment_method);
CREATE INDEX idx_recharge_records_created_at ON recharge_records(created_at);
CREATE INDEX idx_recharge_records_transaction_id ON recharge_records(transaction_id);

-- 充值记录更新时间触发器
CREATE TRIGGER update_recharge_records_updated_at 
    BEFORE UPDATE ON recharge_records 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 创建消费记录表
CREATE TABLE IF NOT EXISTS consumption_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- 消费信息
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT '金币',
    
    -- 服务信息
    service_type VARCHAR(50) NOT NULL, -- chat, media, call等
    service_id UUID, -- 关联的服务记录ID
    model_name VARCHAR(100), -- 使用的模型名称
    
    -- 计费详情
    billing_unit VARCHAR(20), -- token, image, second等
    billing_quantity DECIMAL(15,4), -- 计费数量
    unit_price DECIMAL(15,6), -- 单价
    
    -- 描述信息
    description TEXT,
    metadata JSONB, -- 额外的元数据
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 消费记录索引
CREATE INDEX idx_consumption_records_user_id ON consumption_records(user_id);
CREATE INDEX idx_consumption_records_wallet_id ON consumption_records(wallet_id);
CREATE INDEX idx_consumption_records_service_type ON consumption_records(service_type);
CREATE INDEX idx_consumption_records_model_name ON consumption_records(model_name);
CREATE INDEX idx_consumption_records_created_at ON consumption_records(created_at);

-- 创建用户时自动创建钱包的触发器函数
CREATE OR REPLACE FUNCTION create_user_wallet()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO wallets (user_id, balance, currency, exchange_rate)
    VALUES (NEW.id, 0.00, 'coins', 10);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器：用户创建时自动创建钱包
CREATE TRIGGER trigger_create_user_wallet
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION create_user_wallet();

-- 为默认管理员创建钱包
INSERT INTO wallets (user_id, balance)
SELECT id, 100000.00
FROM users
WHERE username = 'admin'
ON CONFLICT (user_id) DO NOTHING;
