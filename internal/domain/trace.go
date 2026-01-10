package domain

import "time"

// Trace represents a single log/trace entry from a monitored service
type Trace struct {
	ID          string            `json:"id" db:"id"`
	ProjectID   string            `json:"project_id" db:"project_id"`
	Level       string            `json:"level" db:"level"` // info, warn, error, fatal
	Message     string            `json:"message" db:"message"`
	Timestamp   time.Time         `json:"timestamp" db:"timestamp"`
	Source      string            `json:"source" db:"source"`           // file:line or service name
	StackTrace  string            `json:"stack_trace" db:"stack_trace"` // Full stack trace if available
	Context     map[string]string `json:"context" db:"context"`         // Additional context (user_id, request_id, etc.)
	ServiceName string            `json:"service_name" db:"service_name"`
	Environment string            `json:"environment" db:"environment"` // production, staging, development
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
}

// TraceFilter represents filters for querying traces
type TraceFilter struct {
	ProjectID   string
	Level       []string  // Filter by levels
	ServiceName string
	Environment string
	Since       time.Time
	Until       time.Time
	SearchText  string // Full-text search in message
	Limit       int
	Offset      int
}
