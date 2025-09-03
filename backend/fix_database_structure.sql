-- 修复YUNAI数据库结构问题
-- 作者：小云
-- 时间：2025-08-30

-- 1. 修复关系网表结构 - 添加缺失的 source_character_id 字段
DO $$
BEGIN
    -- 检查 character_relationships_enhanced 表是否存在
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'character_relationships_enhanced') THEN
        -- 创建关系网增强表
        CREATE TABLE character_relationships_enhanced (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            source_character_id UUID NOT NULL,
            target_character_id UUID NOT NULL,
            relationship_type VARCHAR(50) NOT NULL,
            strength DECIMAL(3,2) DEFAULT 0.5,
            intimacy_level INTEGER DEFAULT 1,
            trust_level DECIMAL(3,2) DEFAULT 0.5,
            interaction_frequency INTEGER DEFAULT 0,
            last_interaction_at TIMESTAMP,
            relationship_status VARCHAR(20) DEFAULT 'active',
            tags TEXT[],
            metadata JSONB,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            
            -- 约束
            CONSTRAINT fk_source_character FOREIGN KEY (source_character_id) REFERENCES characters(id) ON DELETE CASCADE,
            CONSTRAINT fk_target_character FOREIGN KEY (target_character_id) REFERENCES characters(id) ON DELETE CASCADE,
            CONSTRAINT unique_relationship UNIQUE (source_character_id, target_character_id, relationship_type)
        );
        
        -- 创建索引
        CREATE INDEX idx_char_rel_source ON character_relationships_enhanced(source_character_id);
        CREATE INDEX idx_char_rel_target ON character_relationships_enhanced(target_character_id);
        CREATE INDEX idx_char_rel_type ON character_relationships_enhanced(relationship_type);
        CREATE INDEX idx_char_rel_strength ON character_relationships_enhanced(strength);
        
        RAISE NOTICE '✅ 创建 character_relationships_enhanced 表成功';
    ELSE
        -- 检查并添加缺失的字段
        IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'character_relationships_enhanced' AND column_name = 'source_character_id') THEN
            ALTER TABLE character_relationships_enhanced ADD COLUMN source_character_id UUID NOT NULL;
            RAISE NOTICE '✅ 添加 source_character_id 字段成功';
        END IF;
    END IF;
END $$;

-- 2. 修复群聊表结构 - 添加缺失的 can_invite 字段
DO $$
BEGIN
    -- 检查 group_chat_members 表是否存在 can_invite 字段
    IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'group_chat_members' AND column_name = 'can_invite') THEN
        ALTER TABLE group_chat_members ADD COLUMN can_invite BOOLEAN DEFAULT true;
        RAISE NOTICE '✅ 添加 can_invite 字段到 group_chat_members 表成功';
    END IF;
    
    -- 检查 group_chats 表是否存在
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'group_chats') THEN
        CREATE TABLE group_chats (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(100) NOT NULL,
            description TEXT,
            creator_id UUID NOT NULL,
            avatar_url VARCHAR(500),
            member_count INTEGER DEFAULT 0,
            max_members INTEGER DEFAULT 500,
            is_public BOOLEAN DEFAULT false,
            join_approval_required BOOLEAN DEFAULT false,
            settings JSONB,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE INDEX idx_group_chats_creator ON group_chats(creator_id);
        CREATE INDEX idx_group_chats_public ON group_chats(is_public);
        
        RAISE NOTICE '✅ 创建 group_chats 表成功';
    END IF;
    
    -- 检查 group_chat_members 表是否存在
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'group_chat_members') THEN
        CREATE TABLE group_chat_members (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            group_id UUID NOT NULL,
            user_id UUID,
            character_id UUID,
            role VARCHAR(20) DEFAULT 'member',
            can_invite BOOLEAN DEFAULT true,
            can_manage BOOLEAN DEFAULT false,
            joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            
            CONSTRAINT fk_group_member_group FOREIGN KEY (group_id) REFERENCES group_chats(id) ON DELETE CASCADE
        );
        
        CREATE INDEX idx_group_members_group ON group_chat_members(group_id);
        CREATE INDEX idx_group_members_user ON group_chat_members(user_id);
        CREATE INDEX idx_group_members_character ON group_chat_members(character_id);
        
        RAISE NOTICE '✅ 创建 group_chat_members 表成功';
    END IF;
END $$;

-- 3. 创建支付相关表（如果不存在）
DO $$
BEGIN
    -- 充值套餐表
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'recharge_packages') THEN
        CREATE TABLE recharge_packages (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(100) NOT NULL,
            description TEXT,
            amount DECIMAL(10,2) NOT NULL,
            bonus_amount DECIMAL(10,2) DEFAULT 0,
            coins INTEGER NOT NULL,
            bonus_coins INTEGER DEFAULT 0,
            is_active BOOLEAN DEFAULT true,
            sort_order INTEGER DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        -- 插入默认套餐
        INSERT INTO recharge_packages (name, description, amount, coins, bonus_coins, sort_order) VALUES
        ('基础套餐', '适合轻度使用', 10.00, 100, 10, 1),
        ('标准套餐', '适合日常使用', 30.00, 300, 50, 2),
        ('高级套餐', '适合重度使用', 68.00, 680, 120, 3),
        ('豪华套餐', '超值优惠', 128.00, 1280, 320, 4);
        
        RAISE NOTICE '✅ 创建 recharge_packages 表并插入默认数据成功';
    END IF;
END $$;

-- 4. 创建测试数据
DO $$
BEGIN
    -- 插入测试角色关系数据
    INSERT INTO character_relationships_enhanced (
        source_character_id, 
        target_character_id, 
        relationship_type, 
        strength, 
        intimacy_level,
        trust_level
    ) VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', 'friend', 0.8, 3, 0.7),
    ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'friend', 0.8, 3, 0.7)
    ON CONFLICT (source_character_id, target_character_id, relationship_type) DO NOTHING;
    
    RAISE NOTICE '✅ 插入测试关系数据成功';
END $$;

RAISE NOTICE '🎉 数据库结构修复完成！';
