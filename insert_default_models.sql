-- 插入默认AI模型

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
