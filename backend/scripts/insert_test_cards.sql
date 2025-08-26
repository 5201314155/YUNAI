-- Insert test gift cards
INSERT INTO gift_cards (
    card_code, card_type, value, discount_rate, bonus_rate, 
    privileges, privilege_duration, description, created_by
) VALUES 
-- Discount card (10% off)
('6688990012345678', 'discount', 100.00, 0.9000, 0.0000, 
 '[]', NULL, '10% discount card (test)', 
 (SELECT id FROM users WHERE username = 'admin')),

-- Bonus card (20% bonus)
('6688990087654321', 'bonus', 100.00, 1.0000, 0.2000, 
 '[]', NULL, '20% bonus card (test)', 
 (SELECT id FROM users WHERE username = 'admin')),

-- Privilege card (video generation) - fixed value to 1.00
('6688990011112222', 'privilege', 1.00, 1.0000, 0.0000, 
 '["txt2video", "hd_upscaler"]', '30 days', 'Video generation privilege card (test)', 
 (SELECT id FROM users WHERE username = 'admin')),

-- Combo card (super privileges)
('6688990099998888', 'combo', 200.00, 0.8000, 0.3000, 
 '["txt2video", "hd_upscaler", "advanced_tts"]', '60 days', 'Super combo card (test)', 
 (SELECT id FROM users WHERE username = 'admin'))

ON CONFLICT (card_code) DO UPDATE SET
    value = EXCLUDED.value,
    discount_rate = EXCLUDED.discount_rate,
    bonus_rate = EXCLUDED.bonus_rate,
    privileges = EXCLUDED.privileges,
    privilege_duration = EXCLUDED.privilege_duration,
    description = EXCLUDED.description;
