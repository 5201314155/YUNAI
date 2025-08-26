-- Create media generation tasks table
CREATE TABLE IF NOT EXISTS media_generation_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Task information
    task_type VARCHAR(50) NOT NULL CHECK (task_type IN ('txt2img', 'img2img', 'img2video', 'txt2video', 'remove_bg', 'animation')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    
    -- Model and parameters
    model_id UUID REFERENCES ai_models(id) ON DELETE SET NULL,
    model_params JSONB NOT NULL DEFAULT '{}',
    
    -- Input data
    text_prompt TEXT,
    negative_prompt TEXT,
    reference_image_url TEXT,
    reference_video_url TEXT,
    
    -- Output data
    result_urls TEXT[], -- Array of generated media URLs
    result_metadata JSONB, -- Metadata about generated media
    
    -- Processing information
    workflow_id VARCHAR(100), -- ComfyUI/WebUI workflow ID
    external_task_id VARCHAR(100), -- Third-party service task ID
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    error_message TEXT,
    
    -- Cost and billing
    estimated_cost DECIMAL(10,6),
    actual_cost DECIMAL(10,6),
    tokens_used INTEGER,
    processing_time_seconds INTEGER,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create media templates table for reusable configurations
CREATE TABLE IF NOT EXISTS media_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for system templates
    
    -- Template information
    name VARCHAR(100) NOT NULL,
    description TEXT,
    template_type VARCHAR(50) NOT NULL CHECK (template_type IN ('txt2img', 'img2img', 'img2video', 'txt2video')),
    
    -- Template configuration
    model_id UUID REFERENCES ai_models(id) ON DELETE SET NULL,
    default_params JSONB NOT NULL DEFAULT '{}',
    prompt_template TEXT, -- Template with placeholders like {character_name}, {scene}
    
    -- Visibility and usage
    is_public BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    usage_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create media assets table for storing generated media
CREATE TABLE IF NOT EXISTS media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id UUID REFERENCES media_generation_tasks(id) ON DELETE SET NULL,
    
    -- Asset information
    asset_type VARCHAR(20) NOT NULL CHECK (asset_type IN ('image', 'video', 'audio')),
    file_url TEXT NOT NULL,
    thumbnail_url TEXT,
    
    -- File metadata
    file_size BIGINT, -- Size in bytes
    width INTEGER,
    height INTEGER,
    duration_seconds DECIMAL(10,3), -- For video/audio
    format VARCHAR(20), -- jpg, png, mp4, etc.
    
    -- Generation metadata
    generation_params JSONB,
    model_used VARCHAR(100),
    prompt_used TEXT,
    
    -- Usage and organization
    title VARCHAR(200),
    description TEXT,
    tags TEXT[],
    is_favorite BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    
    -- Usage tracking
    download_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create character media table for linking media to characters
CREATE TABLE IF NOT EXISTS character_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    media_asset_id UUID NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
    
    -- Media role for character
    media_role VARCHAR(50) NOT NULL CHECK (media_role IN ('bg_image', 'cutout_image', 'avatar', 'gallery')),
    
    -- Auto-generation information
    is_auto_generated BOOLEAN DEFAULT FALSE,
    generation_prompt TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(character_id, media_asset_id, media_role)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_user_id ON media_generation_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_status ON media_generation_tasks(status);
CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_task_type ON media_generation_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_created_at ON media_generation_tasks(created_at);

CREATE INDEX IF NOT EXISTS idx_media_templates_user_id ON media_templates(user_id);
CREATE INDEX IF NOT EXISTS idx_media_templates_template_type ON media_templates(template_type);
CREATE INDEX IF NOT EXISTS idx_media_templates_is_public ON media_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_media_templates_is_featured ON media_templates(is_featured);

CREATE INDEX IF NOT EXISTS idx_media_assets_user_id ON media_assets(user_id);
CREATE INDEX IF NOT EXISTS idx_media_assets_task_id ON media_assets(task_id);
CREATE INDEX IF NOT EXISTS idx_media_assets_asset_type ON media_assets(asset_type);
CREATE INDEX IF NOT EXISTS idx_media_assets_is_public ON media_assets(is_public);
CREATE INDEX IF NOT EXISTS idx_media_assets_created_at ON media_assets(created_at);

CREATE INDEX IF NOT EXISTS idx_character_media_character_id ON character_media(character_id);
CREATE INDEX IF NOT EXISTS idx_character_media_media_asset_id ON character_media(media_asset_id);
CREATE INDEX IF NOT EXISTS idx_character_media_media_role ON character_media(media_role);

-- Create update triggers
CREATE TRIGGER update_media_generation_tasks_updated_at 
    BEFORE UPDATE ON media_generation_tasks 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_media_templates_updated_at 
    BEFORE UPDATE ON media_templates 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_media_assets_updated_at 
    BEFORE UPDATE ON media_assets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert some default media templates
INSERT INTO media_templates (name, description, template_type, default_params, prompt_template, is_public, is_featured) VALUES
('Character Avatar', 'Generate high quality character avatars', 'txt2img',
 '{"width": 512, "height": 512, "steps": 20, "cfg_scale": 7}',
 'portrait of {character_name}, {personality_traits}, high quality, detailed, anime style',
 true, true),

('Character Full Body', 'Generate full body character illustrations', 'txt2img',
 '{"width": 512, "height": 768, "steps": 25, "cfg_scale": 7}',
 'full body portrait of {character_name}, {personality_traits}, standing pose, high quality, detailed',
 true, true),

('Scene Background', 'Generate scene backgrounds for group chats', 'txt2img',
 '{"width": 1024, "height": 576, "steps": 20, "cfg_scale": 7}',
 '{scene_description}, {scene_style}, beautiful scenery, high quality, detailed background',
 true, true),

('Character Animation', 'Convert character images to animations', 'img2video',
 '{"duration": 3, "fps": 24, "motion_strength": 0.7}',
 'animate {character_name}, subtle movement, breathing animation',
 true, false)

ON CONFLICT DO NOTHING;
