-- 删除消费记录表
DROP TABLE IF EXISTS consumption_records;

-- 删除充值记录表
DROP TRIGGER IF EXISTS update_recharge_records_updated_at ON recharge_records;
DROP TABLE IF EXISTS recharge_records;

-- 删除钱包表
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TABLE IF EXISTS wallets;
