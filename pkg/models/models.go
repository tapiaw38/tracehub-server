package models

import (
	"time"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type ErrorType = domain.ErrorType
type Severity = domain.Severity

const (
	ErrorTypePanic        = domain.ErrorTypePanic
	ErrorTypeNilPointer   = domain.ErrorTypeNilPointer
	ErrorTypeHTTP5xx      = domain.ErrorTypeHTTP5xx
	ErrorTypeDatabaseConn = domain.ErrorTypeDatabaseConn
	ErrorTypeTimeout      = domain.ErrorTypeTimeout
	ErrorTypeUnknown      = domain.ErrorTypeUnknown
)

const (
	SeverityCritical = domain.SeverityCritical
	SeverityHigh     = domain.SeverityHigh
	SeverityMedium   = domain.SeverityMedium
	SeverityLow      = domain.SeverityLow
)

type LogEntry struct {
	ServiceName string
	Timestamp   time.Time
	Level       string
	Message     string
	Raw         string
	Fields      map[string]string
}

type DetectedError struct {
	ID            string
	ProjectID     string
	Type          ErrorType
	Severity      Severity
	Message       string
	StackTrace    string
	FilePath      string
	LineNumber    int
	Context       map[string]string
	Count         int
	FirstOccurred time.Time
	LastOccurred  time.Time
	Resolved      bool
	ResolvedAt    *time.Time
	Signature     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ServiceName   string
	Timestamp     time.Time
	RelatedLogs   []*LogEntry
}

type ErrorGroup struct {
	Pattern       string
	Errors        []*DetectedError
	TotalCount    int
	FirstOccurred time.Time
	LastOccurred  time.Time
}

type Metrics struct {
	TotalRequests int
	TotalErrors   int
	RequestRate   float64
	ErrorRate     float64
	LastUpdate    string
}

type ServiceStatus string

const (
	ServiceStatusOK      ServiceStatus = "ok"
	ServiceStatusError   ServiceStatus = "error"
	ServiceStatusWarn    ServiceStatus = "warn"
	ServiceStatusRunning ServiceStatus = "running"
)

type Service struct {
	Name        string
	LogPath     string
	Format      string
	RepoPath    string
	Status      ServiceStatus
	Metrics     *Metrics
	LastError   *DetectedError
	LastUpdated time.Time
}

func NewService(name, logPath, format, repoPath string) *Service {
	return &Service{
		Name:     name,
		LogPath:  logPath,
		Format:   format,
		RepoPath: repoPath,
		Status:   ServiceStatusOK,
		Metrics: &Metrics{
			TotalRequests: 0,
			TotalErrors:   0,
			RequestRate:   0.0,
			ErrorRate:     0.0,
			LastUpdate:    time.Now().Format("15:04:05"),
		},
	}
}

type Fix struct {
	FilePath string
	Changes  string
	Code     string
}

type ClaudeResponse struct {
	RootCause   string
	Explanation string
	Fix         Fix
	Tests       string
	Prevention  string
}

type ProposalStatus string

const (
	ProposalPending  ProposalStatus = "pending"
	ProposalAccepted ProposalStatus = "accepted"
	ProposalRejected ProposalStatus = "rejected"
	ProposalApplied  ProposalStatus = "applied"
)

type FixProposal struct {
	ID          string
	ErrorID     string
	ServiceName string
	RootCause   string
	Explanation string
	Fix         Fix
	Tests       string
	Prevention  string
	Status      ProposalStatus
	CreatedAt   time.Time
}

type PullRequest struct {
	URL         string
	Number      int
	Title       string
	Body        string
	Branch      string
	ServiceName string
	CreatedAt   time.Time
}
