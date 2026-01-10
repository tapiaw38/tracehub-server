package domain

import "time"

// Severity represents error severity levels
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// ErrorType represents different error types
type ErrorType string

const (
	ErrorTypePanic        ErrorType = "panic"
	ErrorTypeNilPointer   ErrorType = "nil_pointer"
	ErrorTypeHTTP5xx      ErrorType = "http_5xx"
	ErrorTypeDatabaseConn ErrorType = "database_connection"
	ErrorTypeTimeout      ErrorType = "timeout"
	ErrorTypeUnknown      ErrorType = "unknown"
)

// DetectedError represents an error detected in traces
type DetectedError struct {
	ID            string            `json:"id" db:"id"`
	ProjectID     string            `json:"project_id" db:"project_id"`
	Type          ErrorType         `json:"type" db:"type"`
	Severity      Severity          `json:"severity" db:"severity"`
	Message       string            `json:"message" db:"message"`
	StackTrace    string            `json:"stack_trace" db:"stack_trace"`
	FilePath      string            `json:"file_path" db:"file_path"`
	LineNumber    int               `json:"line_number" db:"line_number"`
	Context       map[string]string `json:"context" db:"context"`
	Count         int               `json:"count" db:"count"` // Number of occurrences
	FirstOccurred time.Time         `json:"first_occurred" db:"first_occurred"`
	LastOccurred  time.Time         `json:"last_occurred" db:"last_occurred"`
	Resolved      bool              `json:"resolved" db:"resolved"`
	ResolvedAt    *time.Time        `json:"resolved_at" db:"resolved_at"`
	Signature     string            `json:"signature" db:"signature"` // Unique identifier for grouping
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
}

// FixProposal represents an AI-generated fix proposal
type FixProposal struct {
	ID          string    `json:"id" db:"id"`
	ErrorID     string    `json:"error_id" db:"error_id"`
	RootCause   string    `json:"root_cause" db:"root_cause"`
	Explanation string    `json:"explanation" db:"explanation"`
	FixCode     string    `json:"fix_code" db:"fix_code"`
	FilePath    string    `json:"file_path" db:"file_path"`
	Tests       string    `json:"tests" db:"tests"`
	Prevention  string    `json:"prevention" db:"prevention"`
	Status      string    `json:"status" db:"status"` // pending, accepted, rejected, applied
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
