-- Create relationship types table for reusable relationship definitions
CREATE TABLE IF NOT EXISTS relationship_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for system types
    
    -- Type information
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL DEFAULT 'custom' CHECK (category IN ('system', 'public', 'custom')),
    
    -- Relationship properties
    is_mutual BOOLEAN DEFAULT FALSE, -- Whether this relationship is bidirectional
    default_strength DECIMAL(3,2) DEFAULT 0.5 CHECK (default_strength >= 0 AND default_strength <= 1),
    
    -- Emotional dimensions defaults
    default_trust DECIMAL(3,2) DEFAULT 0.5 CHECK (default_trust >= 0 AND default_trust <= 1),
    default_affection DECIMAL(3,2) DEFAULT 0.5 CHECK (default_affection >= 0 AND default_affection <= 1),
    default_respect DECIMAL(3,2) DEFAULT 0.5 CHECK (default_respect >= 0 AND default_respect <= 1),
    default_intimacy DECIMAL(3,2) DEFAULT 0.5 CHECK (default_intimacy >= 0 AND default_intimacy <= 1),
    
    -- Behavioral settings
    default_tone VARCHAR(50) DEFAULT 'neutral', -- formal, casual, intimate, hostile, playful
    default_address_style VARCHAR(50) DEFAULT 'name', -- name, title, nickname, pet_name
    
    -- Usage tracking
    usage_count INTEGER DEFAULT 0,
    is_featured BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, name)
);

-- Create enhanced character relationships table
CREATE TABLE IF NOT EXISTS character_relationships_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    target_character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- Relationship type
    relationship_type_id UUID REFERENCES relationship_types(id) ON DELETE SET NULL,
    custom_type_name VARCHAR(100), -- For ad-hoc relationships
    
    -- Relationship strength and status
    strength DECIMAL(3,2) DEFAULT 0.5 CHECK (strength >= 0 AND strength <= 1),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'dormant', 'broken', 'hidden')),
    
    -- Multi-dimensional emotions (0-1 scale)
    trust DECIMAL(3,2) DEFAULT 0.5 CHECK (trust >= 0 AND trust <= 1),
    affection DECIMAL(3,2) DEFAULT 0.5 CHECK (affection >= 0 AND affection <= 1),
    respect DECIMAL(3,2) DEFAULT 0.5 CHECK (respect >= 0 AND respect <= 1),
    intimacy DECIMAL(3,2) DEFAULT 0.5 CHECK (intimacy >= 0 AND intimacy <= 1),
    jealousy DECIMAL(3,2) DEFAULT 0.0 CHECK (jealousy >= 0 AND jealousy <= 1),
    dependency DECIMAL(3,2) DEFAULT 0.0 CHECK (dependency >= 0 AND dependency <= 1),
    
    -- Communication style
    address_name VARCHAR(100), -- How source addresses target
    tone VARCHAR(50) DEFAULT 'neutral', -- Communication tone
    formality_level DECIMAL(3,2) DEFAULT 0.5 CHECK (formality_level >= 0 AND formality_level <= 1),
    
    -- Behavioral modifiers
    speaking_frequency DECIMAL(3,2) DEFAULT 0.5 CHECK (speaking_frequency >= 0 AND speaking_frequency <= 1),
    initiative_level DECIMAL(3,2) DEFAULT 0.5 CHECK (initiative_level >= 0 AND initiative_level <= 1),
    conflict_tendency DECIMAL(3,2) DEFAULT 0.2 CHECK (conflict_tendency >= 0 AND conflict_tendency <= 1),
    
    -- Trigger conditions
    trigger_keywords TEXT[], -- Keywords that activate this relationship context
    trigger_emotions TEXT[], -- Emotional states that strengthen this relationship
    trigger_scenarios TEXT[], -- Scenarios where this relationship is prominent
    
    -- Memory and history
    relationship_history JSONB DEFAULT '[]', -- Array of relationship events
    shared_memories TEXT[], -- Important shared experiences
    private_notes TEXT, -- Private relationship notes
    
    -- Constraints and boundaries
    forbidden_topics TEXT[], -- Topics to avoid in this relationship
    preferred_topics TEXT[], -- Topics that strengthen the relationship
    interaction_limits JSONB DEFAULT '{}', -- Limits on interaction types
    
    -- Timestamps and tracking
    last_interaction_at TIMESTAMP WITH TIME ZONE,
    relationship_established_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(source_character_id, target_character_id)
);

-- Create relationship events table for tracking relationship changes
CREATE TABLE IF NOT EXISTS relationship_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relationship_id UUID NOT NULL REFERENCES character_relationships_enhanced(id) ON DELETE CASCADE,
    
    -- Event information
    event_type VARCHAR(50) NOT NULL, -- strength_change, emotion_change, conflict, bonding, milestone
    event_description TEXT NOT NULL,
    
    -- Changes made
    strength_change DECIMAL(4,3), -- Can be negative
    emotion_changes JSONB, -- Changes to emotional dimensions
    
    -- Context
    triggered_by_message_id UUID REFERENCES chat_messages(id) ON DELETE SET NULL,
    triggered_by_scenario VARCHAR(100),
    trigger_keywords TEXT[],
    
    -- Metadata
    automatic BOOLEAN DEFAULT TRUE, -- Whether this was automatically triggered
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create memory fragments table for RAG system
CREATE TABLE IF NOT EXISTS memory_fragments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- Memory content
    content TEXT NOT NULL,
    summary TEXT, -- Auto-generated summary
    memory_type VARCHAR(50) NOT NULL DEFAULT 'conversation', -- conversation, event, fact, emotion, relationship
    
    -- Importance and relevance
    importance_score DECIMAL(3,2) DEFAULT 0.5 CHECK (importance_score >= 0 AND importance_score <= 1),
    emotional_intensity DECIMAL(3,2) DEFAULT 0.0 CHECK (emotional_intensity >= 0 AND emotional_intensity <= 1),
    
    -- Associations
    related_characters UUID[], -- Characters involved in this memory
    related_topics TEXT[], -- Topics/keywords associated with this memory
    related_emotions TEXT[], -- Emotions associated with this memory
    
    -- Context
    source_message_id UUID REFERENCES chat_messages(id) ON DELETE SET NULL,
    source_group_chat_id UUID REFERENCES group_chats(id) ON DELETE SET NULL,
    context_metadata JSONB DEFAULT '{}',
    
    -- Vector embedding for similarity search (stored as JSONB for now)
    embedding JSONB, -- OpenAI embedding dimension stored as JSON array
    
    -- Lifecycle
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE, -- For temporary memories
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create conversation contexts table for managing chat sessions
CREATE TABLE IF NOT EXISTS conversation_contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_chat_id UUID NOT NULL REFERENCES group_chats(id) ON DELETE CASCADE,
    
    -- Context information
    context_name VARCHAR(200),
    current_scene VARCHAR(200),
    mood VARCHAR(100),
    active_themes TEXT[],
    
    -- Participants and their states
    active_characters UUID[], -- Currently active characters
    speaking_queue UUID[], -- Queue of characters to speak
    current_speaker_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    next_speaker_id UUID REFERENCES characters(id) ON DELETE SET NULL,
    
    -- Conversation flow
    turn_count INTEGER DEFAULT 0,
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Memory context
    recent_memory_ids UUID[], -- Recently accessed memories
    context_summary TEXT, -- Summary of recent conversation
    
    -- Orchestration settings
    orchestration_mode VARCHAR(50) DEFAULT 'auto', -- auto, manual, guided
    intervention_level DECIMAL(3,2) DEFAULT 0.3 CHECK (intervention_level >= 0 AND intervention_level <= 1),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger rules table for scenario-based triggers
CREATE TABLE IF NOT EXISTS trigger_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    group_chat_id UUID REFERENCES group_chats(id) ON DELETE CASCADE,
    
    -- Rule information
    name VARCHAR(200) NOT NULL,
    description TEXT,
    rule_type VARCHAR(50) NOT NULL, -- keyword, emotion, relationship, time, event
    
    -- Trigger conditions
    trigger_conditions JSONB NOT NULL, -- Flexible condition definition
    
    -- Actions to take
    actions JSONB NOT NULL, -- Actions to execute when triggered
    
    -- Constraints
    cooldown_minutes INTEGER DEFAULT 0,
    max_triggers_per_day INTEGER,
    priority INTEGER DEFAULT 50,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_relationship_types_category ON relationship_types(category);
CREATE INDEX IF NOT EXISTS idx_relationship_types_user_id ON relationship_types(user_id);

CREATE INDEX IF NOT EXISTS idx_character_relationships_enhanced_source ON character_relationships_enhanced(source_character_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_enhanced_target ON character_relationships_enhanced(target_character_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_enhanced_type ON character_relationships_enhanced(relationship_type_id);
CREATE INDEX IF NOT EXISTS idx_character_relationships_enhanced_strength ON character_relationships_enhanced(strength);
CREATE INDEX IF NOT EXISTS idx_character_relationships_enhanced_status ON character_relationships_enhanced(status);

CREATE INDEX IF NOT EXISTS idx_relationship_events_relationship_id ON relationship_events(relationship_id);
CREATE INDEX IF NOT EXISTS idx_relationship_events_event_type ON relationship_events(event_type);
CREATE INDEX IF NOT EXISTS idx_relationship_events_created_at ON relationship_events(created_at);

CREATE INDEX IF NOT EXISTS idx_memory_fragments_character_id ON memory_fragments(character_id);
CREATE INDEX IF NOT EXISTS idx_memory_fragments_memory_type ON memory_fragments(memory_type);
CREATE INDEX IF NOT EXISTS idx_memory_fragments_importance_score ON memory_fragments(importance_score);
CREATE INDEX IF NOT EXISTS idx_memory_fragments_created_at ON memory_fragments(created_at);

CREATE INDEX IF NOT EXISTS idx_conversation_contexts_group_chat_id ON conversation_contexts(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_conversation_contexts_current_speaker ON conversation_contexts(current_speaker_id);
CREATE INDEX IF NOT EXISTS idx_conversation_contexts_last_activity ON conversation_contexts(last_activity_at);

CREATE INDEX IF NOT EXISTS idx_trigger_rules_group_chat_id ON trigger_rules(group_chat_id);
CREATE INDEX IF NOT EXISTS idx_trigger_rules_rule_type ON trigger_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_trigger_rules_is_active ON trigger_rules(is_active);

-- Create update triggers
CREATE TRIGGER update_relationship_types_updated_at 
    BEFORE UPDATE ON relationship_types 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_character_relationships_enhanced_updated_at 
    BEFORE UPDATE ON character_relationships_enhanced 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_memory_fragments_updated_at 
    BEFORE UPDATE ON memory_fragments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversation_contexts_updated_at 
    BEFORE UPDATE ON conversation_contexts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_trigger_rules_updated_at 
    BEFORE UPDATE ON trigger_rules 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default relationship types
INSERT INTO relationship_types (name, display_name, description, category, is_mutual, default_strength, default_trust, default_affection, default_respect, default_intimacy, default_tone, is_featured) VALUES
('friend', 'Friend', 'A friendly relationship with mutual trust and affection', 'system', true, 0.7, 0.8, 0.7, 0.7, 0.4, 'casual', true),
('rival', 'Rival', 'A competitive relationship with tension and challenge', 'system', true, 0.6, 0.3, 0.2, 0.6, 0.1, 'formal', true),
('lover', 'Lover', 'A romantic relationship with high intimacy and affection', 'system', true, 0.9, 0.9, 0.9, 0.8, 0.9, 'intimate', true),
('family', 'Family', 'A family relationship with deep bonds and care', 'system', true, 0.8, 0.9, 0.8, 0.9, 0.7, 'casual', true),
('mentor', 'Mentor', 'A teaching relationship with guidance and respect', 'system', false, 0.7, 0.8, 0.6, 0.9, 0.3, 'formal', true),
('student', 'Student', 'A learning relationship with respect and dependency', 'system', false, 0.6, 0.7, 0.6, 0.8, 0.2, 'formal', true),
('enemy', 'Enemy', 'A hostile relationship with conflict and distrust', 'system', true, 0.8, 0.1, 0.1, 0.2, 0.0, 'hostile', true),
('stranger', 'Stranger', 'An unknown relationship with neutral feelings', 'system', true, 0.1, 0.5, 0.5, 0.5, 0.0, 'formal', false),
('colleague', 'Colleague', 'A professional relationship with mutual respect', 'system', true, 0.5, 0.6, 0.5, 0.7, 0.2, 'formal', true),
('protector', 'Protector', 'A protective relationship with care and responsibility', 'system', false, 0.8, 0.8, 0.7, 0.7, 0.4, 'caring', true)

ON CONFLICT DO NOTHING;
