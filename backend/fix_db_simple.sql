-- Fix YUNAI Database Structure Issues
-- Author: XiaoYun
-- Date: 2025-08-30

-- 1. Fix character_relationships_enhanced table
DO $$
BEGIN
    -- Check if table exists
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'character_relationships_enhanced') THEN
        -- Create enhanced relationships table
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
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        -- Create indexes
        CREATE INDEX idx_char_rel_source ON character_relationships_enhanced(source_character_id);
        CREATE INDEX idx_char_rel_target ON character_relationships_enhanced(target_character_id);
        CREATE INDEX idx_char_rel_type ON character_relationships_enhanced(relationship_type);
        CREATE INDEX idx_char_rel_strength ON character_relationships_enhanced(strength);
        
        RAISE NOTICE 'Created character_relationships_enhanced table successfully';
    ELSE
        -- Add missing columns if they don't exist
        IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'character_relationships_enhanced' AND column_name = 'source_character_id') THEN
            ALTER TABLE character_relationships_enhanced ADD COLUMN source_character_id UUID;
            RAISE NOTICE 'Added source_character_id column successfully';
        END IF;
    END IF;
END $$;

-- 2. Fix group_chat_members table
DO $$
BEGIN
    -- Add can_invite column if it doesn't exist
    IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'group_chat_members' AND column_name = 'can_invite') THEN
        ALTER TABLE group_chat_members ADD COLUMN can_invite BOOLEAN DEFAULT true;
        RAISE NOTICE 'Added can_invite column to group_chat_members table successfully';
    END IF;
    
    -- Create group_chats table if it doesn't exist
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
        
        RAISE NOTICE 'Created group_chats table successfully';
    END IF;
    
    -- Create group_chat_members table if it doesn't exist
    IF NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'group_chat_members') THEN
        CREATE TABLE group_chat_members (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            group_id UUID NOT NULL,
            user_id UUID,
            character_id UUID,
            role VARCHAR(20) DEFAULT 'member',
            can_invite BOOLEAN DEFAULT true,
            can_manage BOOLEAN DEFAULT false,
            joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE INDEX idx_group_members_group ON group_chat_members(group_id);
        CREATE INDEX idx_group_members_user ON group_chat_members(user_id);
        CREATE INDEX idx_group_members_character ON group_chat_members(character_id);
        
        RAISE NOTICE 'Created group_chat_members table successfully';
    END IF;
END $$;

-- 3. Create recharge_packages table for payment system
DO $$
BEGIN
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
        
        -- Insert default packages
        INSERT INTO recharge_packages (name, description, amount, coins, bonus_coins, sort_order) VALUES
        ('Basic Package', 'For light usage', 10.00, 100, 10, 1),
        ('Standard Package', 'For daily usage', 30.00, 300, 50, 2),
        ('Premium Package', 'For heavy usage', 68.00, 680, 120, 3),
        ('Deluxe Package', 'Best value', 128.00, 1280, 320, 4);
        
        RAISE NOTICE 'Created recharge_packages table and inserted default data successfully';
    END IF;
END $$;

-- 4. Insert test relationship data
DO $$
BEGIN
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
    ON CONFLICT DO NOTHING;
    
    RAISE NOTICE 'Inserted test relationship data successfully';
END $$;

RAISE NOTICE 'Database structure fix completed successfully!';
