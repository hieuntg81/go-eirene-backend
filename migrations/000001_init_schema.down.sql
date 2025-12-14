-- Drop triggers
DROP TRIGGER IF EXISTS update_case_volunteer_count ON case_volunteers;
DROP TRIGGER IF EXISTS update_cases_updated_at ON cases;
DROP TRIGGER IF EXISTS update_push_tokens_updated_at ON push_tokens;
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop functions
DROP FUNCTION IF EXISTS update_volunteer_count();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS case_updates;
DROP TABLE IF EXISTS case_volunteers;
DROP TABLE IF EXISTS case_media;
DROP TABLE IF EXISTS case_accident_details;
DROP TABLE IF EXISTS case_flood_details;
DROP TABLE IF EXISTS case_animal_details;
DROP TABLE IF EXISTS cases;
DROP TABLE IF EXISTS push_tokens;
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS users;

-- Drop enum types
DROP TYPE IF EXISTS update_type;
DROP TYPE IF EXISTS device_platform;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS media_type;
DROP TYPE IF EXISTS accident_type;
DROP TYPE IF EXISTS animal_condition;
DROP TYPE IF EXISTS animal_type;
DROP TYPE IF EXISTS volunteer_status;
DROP TYPE IF EXISTS urgency_level;
DROP TYPE IF EXISTS case_status;
DROP TYPE IF EXISTS case_type;
DROP TYPE IF EXISTS user_role;

-- Drop extensions (optional - might affect other databases)
-- DROP EXTENSION IF EXISTS postgis;
-- DROP EXTENSION IF EXISTS "uuid-ossp";
