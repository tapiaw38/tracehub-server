package models

import "time"

// ServiceStatus represents the status of a monitored service
type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusError   ServiceStatus = "error"
	StatusWarning ServiceStatus = "warning"
	StatusStopped ServiceStatus = "stopped"
)

// Service represents a monitored microservice
type Service struct {
	Name         string        `json:"name"`
	LogPath      string        `json:"log_path"`
	Format       string        `json:"format"`
	RepoPath     string        `json:"repo_path"`
	Status       ServiceStatus `json:"status"`
	Metrics      *Metrics      `json:"metrics"`
	LastError    *DetectedError `json:"last_error,omitempty"`
	LastUpdated  time.Time     `json:"last_updated"`
	IsActive     bool          `json:"is_active"`
}

// Metrics represents service metrics
type Metrics struct {
	RequestRate   float64   `json:"request_rate"`   // requests per second
	ErrorRate     float64   `json:"error_rate"`     // error percentage
	TotalRequests int64     `json:"total_requests"`
	TotalErrors   int64     `json:"total_errors"`
	LastHour      *TimeMetrics `json:"last_hour,omitempty"`
}

// TimeMetrics represents metrics over a time window
type TimeMetrics struct {
	Requests      int64     `json:"requests"`
	Errors        int64     `json:"errors"`
	AvgLatency    float64   `json:"avg_latency,omitempty"`
	WindowStart   time.Time `json:"window_start"`
	WindowEnd     time.Time `json:"window_end"`
}

// NewService creates a new Service instance with default values
func NewService(name, logPath, format, repoPath string) *Service {
	return &Service{
		Name:        name,
		LogPath:     logPath,
		Format:      format,
		RepoPath:    repoPath,
		Status:      StatusRunning,
		Metrics:     &Metrics{},
		LastUpdated: time.Now(),
		IsActive:    true,
	}
}
