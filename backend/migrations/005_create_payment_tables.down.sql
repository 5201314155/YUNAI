-- Drop triggers
DROP TRIGGER IF EXISTS ensure_single_default_card_trigger ON payment_cards;

-- Drop functions
DROP FUNCTION IF EXISTS ensure_single_default_card();
DROP FUNCTION IF EXISTS generate_order_no();

-- Drop admin recharge records table
DROP TABLE IF EXISTS admin_recharge_records;

-- Drop recharge packages table
DROP TRIGGER IF EXISTS update_recharge_packages_updated_at ON recharge_packages;
DROP TABLE IF EXISTS recharge_packages;

-- Drop recharge orders table
DROP TRIGGER IF EXISTS update_recharge_orders_updated_at ON recharge_orders;
DROP TABLE IF EXISTS recharge_orders;

-- Drop payment passwords table
DROP TRIGGER IF EXISTS update_payment_passwords_updated_at ON payment_passwords;
DROP TABLE IF EXISTS payment_passwords;

-- Drop payment cards table
DROP TRIGGER IF EXISTS update_payment_cards_updated_at ON payment_cards;
DROP TABLE IF EXISTS payment_cards;
