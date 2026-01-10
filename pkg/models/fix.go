package models

import "time"

// FixProposal represents a fix proposed by Claude AI
type FixProposal struct {
	ID          string        `json:"id"`
	ErrorID     string        `json:"error_id"`
	ServiceName string        `json:"service_name"`
	RootCause   string        `json:"root_cause"`
	Explanation string        `json:"explanation"`
	Fix         *Fix          `json:"fix"`
	Tests       string        `json:"tests,omitempty"`
	Prevention  string        `json:"prevention,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	Status      ProposalStatus `json:"status"`
}

// Fix represents the actual code fix
type Fix struct {
	FilePath    string   `json:"file"`
	Changes     string   `json:"changes"`
	Code        string   `json:"code"`
	LineNumbers []int    `json:"line_numbers,omitempty"`
	OriginalCode string  `json:"original_code,omitempty"`
}

// ProposalStatus represents the status of a fix proposal
type ProposalStatus string

const (
	ProposalPending  ProposalStatus = "pending"
	ProposalAccepted ProposalStatus = "accepted"
	ProposalRejected ProposalStatus = "rejected"
	ProposalApplied  ProposalStatus = "applied"
)

// PullRequest represents a created PR
type PullRequest struct {
	Number      int       `json:"number"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Branch      string    `json:"branch"`
	ServiceName string    `json:"service_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// ClaudeResponse represents the response from Claude AI
type ClaudeResponse struct {
	RootCause   string `json:"root_cause"`
	Explanation string `json:"explanation"`
	Fix         *Fix   `json:"fix"`
	Tests       string `json:"tests"`
	Prevention  string `json:"prevention"`
}
