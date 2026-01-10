-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    language VARCHAR(50) NOT NULL,
    repo_url VARCHAR(500),
    api_key_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP
);

-- Add foreign key to projects.api_key_id
ALTER TABLE projects ADD CONSTRAINT fk_project_api_key
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL;

-- Traces table
CREATE TABLE IF NOT EXISTS traces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    source VARCHAR(500),
    stack_trace TEXT,
    context JSONB,
    service_name VARCHAR(255),
    environment VARCHAR(50) NOT NULL DEFAULT 'production',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Detected Errors table
CREATE TABLE IF NOT EXISTS detected_errors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    stack_trace TEXT,
    file_path VARCHAR(500),
    line_number INTEGER,
    context JSONB,
    count INTEGER NOT NULL DEFAULT 1,
    first_occurred TIMESTAMP NOT NULL,
    last_occurred TIMESTAMP NOT NULL,
    resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at TIMESTAMP,
    signature VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, signature)
);

-- Fix Proposals table
CREATE TABLE IF NOT EXISTS fix_proposals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    error_id UUID NOT NULL REFERENCES detected_errors(id) ON DELETE CASCADE,
    root_cause TEXT NOT NULL,
    explanation TEXT NOT NULL,
    fix_code TEXT NOT NULL,
    file_path VARCHAR(500),
    tests TEXT,
    prevention TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_traces_project_id ON traces(project_id);
CREATE INDEX idx_traces_timestamp ON traces(timestamp DESC);
CREATE INDEX idx_traces_level ON traces(level);
CREATE INDEX idx_traces_project_timestamp ON traces(project_id, timestamp DESC);
CREATE INDEX idx_traces_service_name ON traces(service_name);

CREATE INDEX idx_detected_errors_project_id ON detected_errors(project_id);
CREATE INDEX idx_detected_errors_signature ON detected_errors(signature);
CREATE INDEX idx_detected_errors_last_occurred ON detected_errors(last_occurred DESC);
CREATE INDEX idx_detected_errors_resolved ON detected_errors(resolved);

CREATE INDEX idx_api_keys_project_id ON api_keys(project_id);
CREATE INDEX idx_api_keys_key ON api_keys(key);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_detected_errors_updated_at BEFORE UPDATE ON detected_errors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
