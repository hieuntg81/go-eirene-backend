-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255),
    oauth_provider VARCHAR(50),
    oauth_id VARCHAR(255),
    display_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'both',
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    is_available BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    total_cases_reported INTEGER NOT NULL DEFAULT 0,
    total_cases_resolved INTEGER NOT NULL DEFAULT 0,
    location_updated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id) WHERE oauth_provider IS NOT NULL;
CREATE INDEX idx_users_location ON users(latitude, longitude) WHERE latitude IS NOT NULL AND longitude IS NOT NULL;
CREATE INDEX idx_users_available ON users(is_available) WHERE is_available = true;

-- User preferences table
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    push_enabled BOOLEAN NOT NULL DEFAULT true,
    case_types TEXT[] NOT NULL DEFAULT ARRAY['animal', 'flood', 'accident'],
    notification_radius_km INTEGER NOT NULL DEFAULT 10,
    center_latitude DECIMAL(10, 8),
    center_longitude DECIMAL(11, 8),
    use_current_location BOOLEAN NOT NULL DEFAULT true,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Push tokens table
CREATE TABLE push_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(20) NOT NULL,
    device_id VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(token)
);

CREATE INDEX idx_push_tokens_user ON push_tokens(user_id);
CREATE INDEX idx_push_tokens_active ON push_tokens(user_id, is_active) WHERE is_active = true;

-- Cases table
CREATE TABLE cases (
    id UUID PRIMARY KEY,
    case_type VARCHAR(20) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    urgency VARCHAR(20) NOT NULL DEFAULT 'medium',
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    address TEXT,
    location_note TEXT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    reporter_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reporter_name VARCHAR(100),
    reporter_phone VARCHAR(20) NOT NULL,
    is_anonymous BOOLEAN NOT NULL DEFAULT false,
    volunteer_count INTEGER NOT NULL DEFAULT 0,
    max_volunteers INTEGER NOT NULL DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_cases_status ON cases(status);
CREATE INDEX idx_cases_type ON cases(case_type);
CREATE INDEX idx_cases_location ON cases(latitude, longitude);
CREATE INDEX idx_cases_reporter ON cases(reporter_id) WHERE reporter_id IS NOT NULL;
CREATE INDEX idx_cases_created ON cases(created_at DESC);
CREATE INDEX idx_cases_active ON cases(status, created_at DESC) WHERE status IN ('pending', 'accepted', 'in_progress');

-- Animal details table
CREATE TABLE case_animal_details (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    animal_type VARCHAR(20) NOT NULL,
    animal_type_other VARCHAR(100),
    condition VARCHAR(20) NOT NULL,
    condition_description TEXT,
    estimated_count INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(case_id)
);

-- Flood details table
CREATE TABLE case_flood_details (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    people_count INTEGER,
    has_children BOOLEAN NOT NULL DEFAULT false,
    has_elderly BOOLEAN NOT NULL DEFAULT false,
    has_disabled BOOLEAN NOT NULL DEFAULT false,
    water_level_cm INTEGER,
    floor_level INTEGER,
    has_power BOOLEAN,
    has_food_water BOOLEAN,
    medical_needs TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(case_id)
);

-- Accident details table
CREATE TABLE case_accident_details (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    accident_type VARCHAR(20) NOT NULL,
    victim_count INTEGER NOT NULL DEFAULT 1,
    has_unconscious BOOLEAN NOT NULL DEFAULT false,
    has_bleeding BOOLEAN NOT NULL DEFAULT false,
    has_fracture BOOLEAN NOT NULL DEFAULT false,
    is_trapped BOOLEAN NOT NULL DEFAULT false,
    hazard_present BOOLEAN NOT NULL DEFAULT false,
    hazard_description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(case_id)
);

-- Case media table
CREATE TABLE case_media (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    media_type VARCHAR(20) NOT NULL,
    url TEXT NOT NULL,
    thumbnail_url TEXT,
    file_name VARCHAR(255),
    file_size BIGINT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_case_media_case ON case_media(case_id);

-- Case volunteers table
CREATE TABLE case_volunteers (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'accepted',
    accepted_latitude DECIMAL(10, 8),
    accepted_longitude DECIMAL(11, 8),
    distance_km DECIMAL(10, 2),
    accepted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    arrived_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    note TEXT,
    UNIQUE(case_id, volunteer_id)
);

CREATE INDEX idx_case_volunteers_case ON case_volunteers(case_id);
CREATE INDEX idx_case_volunteers_volunteer ON case_volunteers(volunteer_id);
CREATE INDEX idx_case_volunteers_status ON case_volunteers(status);

-- Case updates table
CREATE TABLE case_updates (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    update_type VARCHAR(30) NOT NULL,
    content TEXT,
    old_status VARCHAR(30),
    new_status VARCHAR(30),
    media_urls TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_case_updates_case ON case_updates(case_id, created_at DESC);

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    case_id UUID REFERENCES cases(id) ON DELETE SET NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON user_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_push_tokens_updated_at BEFORE UPDATE ON push_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cases_updated_at BEFORE UPDATE ON cases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update volunteer count
CREATE OR REPLACE FUNCTION update_volunteer_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE cases SET volunteer_count = volunteer_count + 1 WHERE id = NEW.case_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE cases SET volunteer_count = volunteer_count - 1 WHERE id = OLD.case_id;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_case_volunteer_count
AFTER INSERT OR DELETE ON case_volunteers
    FOR EACH ROW EXECUTE FUNCTION update_volunteer_count();
