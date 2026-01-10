package models

import "time"

// Severity represents the severity level of an error
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypePanic          ErrorType = "panic"
	ErrorTypeNilPointer     ErrorType = "nil_pointer"
	ErrorTypeHTTP5xx        ErrorType = "http_5xx"
	ErrorTypeDatabaseConn   ErrorType = "database_connection"
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeUnknown        ErrorType = "unknown"
)

// DetectedError represents an error detected in logs
type DetectedError struct {
	ID            string            `json:"id"`
	ServiceName   string            `json:"service_name"`
	Type          ErrorType         `json:"type"`
	Severity      Severity          `json:"severity"`
	Message       string            `json:"message"`
	StackTrace    string            `json:"stack_trace,omitempty"`
	FilePath      string            `json:"file_path,omitempty"`
	LineNumber    int               `json:"line_number,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	Context       map[string]string `json:"context,omitempty"`
	RelatedLogs   []*LogEntry       `json:"related_logs,omitempty"`
	Count         int               `json:"count"`
	FirstOccurred time.Time         `json:"first_occurred"`
	LastOccurred  time.Time         `json:"last_occurred"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Level       string            `json:"level"`
	Message     string            `json:"message"`
	ServiceName string            `json:"service_name"`
	Fields      map[string]string `json:"fields,omitempty"`
	Raw         string            `json:"raw"`
}

// ErrorGroup groups similar errors together
type ErrorGroup struct {
	Pattern       string           `json:"pattern"`
	Errors        []*DetectedError `json:"errors"`
	TotalCount    int              `json:"total_count"`
	FirstOccurred time.Time        `json:"first_occurred"`
	LastOccurred  time.Time        `json:"last_occurred"`
}
