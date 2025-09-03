-- Fix missing tables for YUNAI monolith service

-- Character relationships enhanced table
CREATE TABLE IF NOT EXISTS character_relationships_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_character_id UUID NOT NULL,
    target_character_id UUID NOT NULL,
    relationship_type VARCHAR(50) NOT NULL DEFAULT 'friend',
    custom_type_name VARCHAR(100),
    strength DECIMAL(3,2) DEFAULT 0.5 CHECK (strength >= 0 AND strength <= 1),
    intimacy_level DECIMAL(3,2) DEFAULT 0.5 CHECK (intimacy_level >= 0 AND intimacy_level <= 1),
    formality_level VARCHAR(20) DEFAULT 'casual',
    description TEXT,
    is_mutual BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_character_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (target_character_id) REFERENCES characters(id) ON DELETE CASCADE,
    UNIQUE(source_character_id, target_character_id)
);

-- Global prompt templates table
CREATE TABLE IF NOT EXISTS global_prompt_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    template_content TEXT NOT NULL,
    variables JSONB DEFAULT '[]',
    category VARCHAR(50) NOT NULL CHECK (category IN ('system', 'character', 'moments', 'story', 'invitation', 'relationship', 'voice', 'embedding')),
    subcategory VARCHAR(50),
    usage_scenarios JSONB DEFAULT '[]',
    priority INTEGER DEFAULT 100,
    version VARCHAR(20) DEFAULT '1.0',
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Group chats table
CREATE TABLE IF NOT EXISTS group_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL,
    avatar_url VARCHAR(500),
    max_members INTEGER DEFAULT 50,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Group chat members table
CREATE TABLE IF NOT EXISTS group_chat_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL,
    character_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('admin', 'moderator', 'member')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    FOREIGN KEY (group_id) REFERENCES group_chats(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(group_id, character_id)
);

-- Group chat messages table
CREATE TABLE IF NOT EXISTS group_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL,
    sender_character_id UUID NOT NULL,
    sender_user_id UUID NOT NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'voice', 'system')),
    reply_to_id UUID,
    mentioned_characters JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES group_chats(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_character_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reply_to_id) REFERENCES group_chat_messages(id) ON DELETE SET NULL
);

-- World settings table
CREATE TABLE IF NOT EXISTS world_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    background_story TEXT,
    rules JSONB DEFAULT '{}',
    timeline JSONB DEFAULT '{}',
    locations JSONB DEFAULT '{}',
    is_public BOOLEAN DEFAULT false,
    creator_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Story chapters table
CREATE TABLE IF NOT EXISTS story_chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    chapter_number INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    trigger_conditions JSONB DEFAULT '{}',
    character_roles JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (world_id) REFERENCES world_settings(id) ON DELETE CASCADE,
    UNIQUE(world_id, chapter_number)
);

-- Story triggers table
CREATE TABLE IF NOT EXISTS story_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(50) NOT NULL CHECK (trigger_type IN ('time', 'relationship', 'event', 'interaction', 'condition')),
    conditions JSONB NOT NULL DEFAULT '{}',
    actions JSONB NOT NULL DEFAULT '{}',
    priority INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    world_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (world_id) REFERENCES world_settings(id) ON DELETE CASCADE
);

-- Insert some default data
INSERT INTO global_prompt_templates (template_name, display_name, description, template_content, category, is_default, is_active) VALUES
('chat_base_system', 'Chat Base Template', 'Basic chat system prompt', 'You are {{character_name}}. Please chat naturally based on your personality: {{personality}}', 'character', true, true),
('moments_generation', 'Moments Generation Template', 'Template for generating moments', 'Generate a social media post as {{character_name}} with personality: {{personality}}', 'moments', true, true),
('group_chat_system', 'Group Chat Template', 'Template for group chat interactions', 'You are {{character_name}} in a group chat. Interact naturally with other members.', 'character', true, true)
ON CONFLICT (template_name) DO NOTHING;

-- Insert default world setting
INSERT INTO world_settings (name, description, background_story, creator_id) VALUES
('Magic Academy', 'A magical school setting', 'A prestigious academy where students learn magic and form lasting friendships.', (SELECT id FROM users LIMIT 1))
ON CONFLICT DO NOTHING;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_character_relationships_enhanced_source ON character_relationships_enhanced(source_character_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_enhanced_target ON character_relationships_enhanced(target_character_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_group ON group_chat_messages(group_id);
CREATE INDEX IF NOT EXISTS idx_group_chat_messages_created ON group_chat_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_story_chapters_world ON story_chapters(world_id);
CREATE INDEX IF NOT EXISTS idx_story_triggers_world ON story_triggers(world_id);
