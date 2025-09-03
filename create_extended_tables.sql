-- 创建扩展功能表

-- AI模型配置表
CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_key VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model_type VARCHAR(50) NOT NULL,
    capabilities JSONB,
    params_schema JSONB,
    model_system_prompt TEXT,
    base_url TEXT,
    api_key_encrypted TEXT,
    pricing JSONB,
    visibility VARCHAR(20) DEFAULT 'public',
    min_user_type VARCHAR(20) DEFAULT 'free',
    min_balance DECIMAL(10,2),
    daily_limit INTEGER,
    user_limit INTEGER,
    health_status VARCHAR(20) DEFAULT 'unknown',
    weight INTEGER DEFAULT 100,
    fallback_chain JSONB,
    connectivity_test_endpoint TEXT,
    is_active BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_health_check TIMESTAMP
);

-- AI调用计费记录表
CREATE TABLE IF NOT EXISTS ai_call_billing_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    model_id UUID REFERENCES ai_models(id),
    session_id VARCHAR(100),
    call_type VARCHAR(50) NOT NULL,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    coins_charged DECIMAL(10,2) NOT NULL,
    original_balance DECIMAL(10,2) NOT NULL,
    new_balance DECIMAL(10,2) NOT NULL,
    pricing_details JSONB,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    refund_amount DECIMAL(10,2),
    refund_reason TEXT,
    refunded_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 模型使用记录表
CREATE TABLE IF NOT EXISTS model_usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID REFERENCES ai_models(id),
    user_id UUID REFERENCES core_users(id),
    session_id VARCHAR(100),
    request_type VARCHAR(50),
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    coins_charged DECIMAL(10,2),
    response_time INTEGER, -- 毫秒
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 支付卡片表
CREATE TABLE IF NOT EXISTS payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    card_number VARCHAR(50) UNIQUE NOT NULL,
    card_name VARCHAR(100) NOT NULL,
    card_type VARCHAR(20) NOT NULL,
    balance DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'CNY',
    expiry_date TIMESTAMP,
    is_frozen BOOLEAN DEFAULT false,
    is_default BOOLEAN DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 钱包交易表（如果不存在）
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID REFERENCES wallets(id),
    transaction_type VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    description TEXT,
    reference_id UUID,
    status VARCHAR(20) DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT NOW()
);

-- 钱包表（如果不存在）
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES core_users(id),
    balance DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(10) DEFAULT 'CNY',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ai_models_provider ON ai_models(provider);
CREATE INDEX IF NOT EXISTS idx_ai_models_type ON ai_models(model_type);
CREATE INDEX IF NOT EXISTS idx_ai_models_active ON ai_models(is_active);

CREATE INDEX IF NOT EXISTS idx_billing_records_user_id ON ai_call_billing_records(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_records_model_id ON ai_call_billing_records(model_id);
CREATE INDEX IF NOT EXISTS idx_billing_records_created_at ON ai_call_billing_records(created_at);

CREATE INDEX IF NOT EXISTS idx_payment_cards_user_id ON payment_cards(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_cards_card_number ON payment_cards(card_number);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions(transaction_type);

-- 插入一些默认的AI模型配置
INSERT INTO ai_models (internal_key, display_name, provider, model_type, capabilities, pricing, weight) 
VALUES 
    ('deepseek-chat', 'DeepSeek Chat', 'deepseek', 'chat', 
     '{"chat": true, "image": false, "voice": false}',
     '{"input_token_price": 0.001, "output_token_price": 0.002, "unit": "1k_tokens", "currency": "CNY"}',
     100),
    ('deepseek-coder', 'DeepSeek Coder', 'deepseek', 'code', 
     '{"chat": true, "code": true, "image": false}',
     '{"input_token_price": 0.001, "output_token_price": 0.002, "unit": "1k_tokens", "currency": "CNY"}',
     90),
    ('gpt-4o', 'GPT-4o', 'openai', 'chat', 
     '{"chat": true, "image": true, "voice": false}',
     '{"input_token_price": 0.005, "output_token_price": 0.015, "unit": "1k_tokens", "currency": "CNY"}',
     80)
ON CONFLICT (internal_key) DO NOTHING;
