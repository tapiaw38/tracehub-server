package domain

import "time"

// Project represents a monitored application/service
type Project struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Language    string    `json:"language" db:"language"` // javascript, go, python, etc.
	RepoURL     string    `json:"repo_url" db:"repo_url"`
	ApiKeyID    string    `json:"api_key_id" db:"api_key_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ApiKey represents an API key for a project
type ApiKey struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	Key       string    `json:"key" db:"key"`            // Hashed API key
	RawKey    string    `json:"raw_key,omitempty" db:"-"` // Only set when creating
	Name      string    `json:"name" db:"name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	LastUsed  *time.Time `json:"last_used" db:"last_used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
}
