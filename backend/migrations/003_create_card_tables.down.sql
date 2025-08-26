-- 删除用户特权表
DROP TABLE IF EXISTS user_privileges;

-- 删除卡密使用记录表
DROP TABLE IF EXISTS card_usage_records;

-- 删除卡密表
DROP TRIGGER IF EXISTS update_gift_cards_updated_at ON gift_cards;
DROP TABLE IF EXISTS gift_cards;
