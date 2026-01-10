-- Drop triggers
DROP TRIGGER IF EXISTS update_detected_errors_updated_at ON detected_errors;
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_is_active;
DROP INDEX IF EXISTS idx_api_keys_key;
DROP INDEX IF EXISTS idx_api_keys_project_id;

DROP INDEX IF EXISTS idx_detected_errors_resolved;
DROP INDEX IF EXISTS idx_detected_errors_last_occurred;
DROP INDEX IF EXISTS idx_detected_errors_signature;
DROP INDEX IF EXISTS idx_detected_errors_project_id;

DROP INDEX IF EXISTS idx_traces_service_name;
DROP INDEX IF EXISTS idx_traces_project_timestamp;
DROP INDEX IF EXISTS idx_traces_level;
DROP INDEX IF EXISTS idx_traces_timestamp;
DROP INDEX IF EXISTS idx_traces_project_id;

-- Drop tables in correct order (respecting foreign keys)
DROP TABLE IF EXISTS fix_proposals;
DROP TABLE IF EXISTS detected_errors;
DROP TABLE IF EXISTS traces;

-- Remove foreign key before dropping api_keys
ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_project_api_key;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS projects;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
