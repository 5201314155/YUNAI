# YUNAI 数据库结构设计

## 文档信息
- **文档名称**: YUNAI 数据库结构设计
- **版本**: v1.0
- **创建时间**: 2025-08-26
- **作者**: 小云
- **更新时间**: 2025-08-26

## 数据库概述

YUNAI 系统使用 PostgreSQL 作为主数据库，Redis 作为缓存数据库。数据库设计遵循第三范式，确保数据一致性和完整性。

### 数据库配置
- **主数据库**: PostgreSQL 14+
- **缓存数据库**: Redis 6+
- **字符集**: UTF-8
- **时区**: Asia/Shanghai

## 核心数据表

### 1. 用户管理表

#### users - 用户基础信息表
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**字段说明**:
- `id`: 用户唯一标识符
- `username`: 用户名，唯一
- `email`: 邮箱地址，唯一
- `password_hash`: 密码哈希值
- `is_active`: 是否激活
- `is_verified`: 是否已验证邮箱
- `last_login_at`: 最后登录时间

#### user_profiles - 用户详细资料表
```sql
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    phone VARCHAR(20),
    birth_date DATE,
    gender VARCHAR(10),
    location VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 2. 支付系统表

#### payment_cards - 支付卡片表
```sql
CREATE TABLE payment_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 卡片信息
    card_number VARCHAR(20) NOT NULL,
    card_number_hash VARCHAR(255) NOT NULL,
    card_type VARCHAR(20) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(10) NOT NULL,
    cardholder_name VARCHAR(100) NOT NULL,
    
    -- 邮箱绑定（安全功能）
    bound_email VARCHAR(255),
    bound_at TIMESTAMP WITH TIME ZONE,
    email_verified BOOLEAN DEFAULT FALSE,
    
    -- 卡片余额
    balance DECIMAL(15,2) DEFAULT 0.00 CHECK (balance >= 0),
    currency VARCHAR(10) DEFAULT 'CNY',
    
    -- 状态管理
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    is_frozen BOOLEAN DEFAULT FALSE,
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP WITH TIME ZONE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**字段说明**:
- `balance`: 卡片余额（金额，单位：元）
- `currency`: 货币类型，默认人民币
- `is_frozen`: 是否冻结
- `bound_email`: 绑定的邮箱地址

#### recharge_orders - 充值订单表
```sql
CREATE TABLE recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_no VARCHAR(50) UNIQUE NOT NULL,
    
    -- 金额信息
    amount DECIMAL(10,2) NOT NULL,
    coins_amount BIGINT NOT NULL,
    exchange_rate INTEGER NOT NULL DEFAULT 10,
    
    -- 支付信息
    payment_method VARCHAR(20) NOT NULL,
    payment_card_id UUID REFERENCES payment_cards(id),
    
    -- 优惠信息
    original_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    bonus_coins BIGINT DEFAULT 0,
    card_code VARCHAR(50),
    
    -- 状态管理
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    
    -- 时间管理
    expired_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### proxy_recharge_orders - 代充订单表
```sql
CREATE TABLE proxy_recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no VARCHAR(50) UNIQUE NOT NULL,
    payer_user_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID NOT NULL REFERENCES users(id),
    payment_card_id UUID NOT NULL REFERENCES payment_cards(id),
    
    -- 金额信息
    original_amount DECIMAL(10,2) NOT NULL,
    actual_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    coins_amount BIGINT NOT NULL,
    bonus_coins BIGINT DEFAULT 0,
    total_coins BIGINT NOT NULL,
    exchange_rate INTEGER NOT NULL DEFAULT 10,
    
    -- 卡片信息
    card_type VARCHAR(50),
    message TEXT,
    
    -- 状态管理
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 3. 钱包管理表

#### wallets - 用户钱包表
```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 余额信息
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'coins',
    exchange_rate INTEGER NOT NULL DEFAULT 10,
    
    -- 限额管理
    daily_limit DECIMAL(15,2) DEFAULT 10000.00,
    monthly_limit DECIMAL(15,2) DEFAULT 100000.00,
    daily_spent DECIMAL(15,2) DEFAULT 0.00,
    monthly_spent DECIMAL(15,2) DEFAULT 0.00,
    last_reset_date DATE DEFAULT CURRENT_DATE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**字段说明**:
- `balance`: 钱包余额（金币数量）
- `currency`: 货币类型，默认 'coins'
- `exchange_rate`: 汇率（1元 = N金币）

#### wallet_transactions - 钱包交易记录表
```sql
CREATE TABLE wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 交易信息
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    
    -- 关联信息
    related_order_id UUID,
    description TEXT,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 4. 系统配置表

#### system_configs - 系统配置表
```sql
CREATE TABLE system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    config_type VARCHAR(20) NOT NULL DEFAULT 'string',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### recharge_packages - 充值套餐表
```sql
CREATE TABLE recharge_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    coins_amount BIGINT NOT NULL,
    bonus_coins BIGINT DEFAULT 0,
    discount_rate DECIMAL(5,4) DEFAULT 1.0000,
    is_popular BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## 索引设计

### 主要索引
```sql
-- 用户表索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);

-- 支付卡片索引
CREATE INDEX idx_payment_cards_user_id ON payment_cards(user_id);
CREATE INDEX idx_payment_cards_card_number_hash ON payment_cards(card_number_hash);
CREATE INDEX idx_payment_cards_is_active ON payment_cards(is_active);

-- 充值订单索引
CREATE INDEX idx_recharge_orders_user_id ON recharge_orders(user_id);
CREATE INDEX idx_recharge_orders_order_no ON recharge_orders(order_no);
CREATE INDEX idx_recharge_orders_status ON recharge_orders(status);
CREATE INDEX idx_recharge_orders_created_at ON recharge_orders(created_at);

-- 钱包索引
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_user_id ON wallet_transactions(user_id);
```

## 数据约束

### 外键约束
- 所有关联表都设置了外键约束
- 使用 CASCADE 删除确保数据一致性

### 检查约束
- 余额字段不能为负数
- 汇率必须大于 0
- 状态字段只能是预定义值

### 唯一约束
- 用户名和邮箱全局唯一
- 订单号全局唯一
- 卡号哈希唯一

## 数据备份策略

### 备份频率
- **全量备份**: 每日凌晨 2:00
- **增量备份**: 每 4 小时
- **日志备份**: 实时

### 备份保留
- 全量备份保留 30 天
- 增量备份保留 7 天
- 日志备份保留 3 天

## 性能优化

### 查询优化
- 合理使用索引
- 避免全表扫描
- 使用分页查询

### 连接池配置
- 最大连接数: 100
- 最小连接数: 10
- 连接超时: 30s

---

**文档维护**: 本文档由小云维护，数据库结构变更需要更新此文档。
