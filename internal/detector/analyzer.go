package detector

import (
	"context"
	"log"
	"sync"

	"github.com/tapiaw38/tracehub-server/internal/config"
	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// Analyzer analyzes log entries and detects errors
type Analyzer struct {
	matcher    *PatternMatcher
	classifier *Classifier
	errorChan  chan *models.DetectedError
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
}

// NewAnalyzer creates a new Analyzer
func NewAnalyzer(cfg *config.DetectionConfig) *Analyzer {
	ctx, cancel := context.WithCancel(context.Background())

	// Convert severity keywords to map
	severityKeywords := make(map[models.Severity][]string)
	severityKeywords[models.SeverityCritical] = cfg.SeverityKeywords.Critical
	severityKeywords[models.SeverityHigh] = cfg.SeverityKeywords.High
	severityKeywords[models.SeverityMedium] = cfg.SeverityKeywords.Medium

	return &Analyzer{
		matcher:    NewPatternMatcher(cfg.Patterns, severityKeywords),
		classifier: NewClassifier(),
		errorChan:  make(chan *models.DetectedError, 100),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the analyzer
func (a *Analyzer) Start() {
	log.Println("Error analyzer started")
}

// Stop stops the analyzer
func (a *Analyzer) Stop() {
	a.cancel()
	close(a.errorChan)
	log.Println("Error analyzer stopped")
}

// Analyze analyzes a log entry for errors
func (a *Analyzer) Analyze(entry *models.LogEntry) {
	if entry == nil {
		return
	}

	// Match against patterns
	isError, errorType, severity := a.matcher.Match(entry)
	if !isError {
		return
	}

	// Classify the error
	detectedError := a.classifier.Classify(entry, errorType, severity)

	// Send to error channel (only for new errors or significant updates)
	if detectedError.Count == 1 {
		select {
		case a.errorChan <- detectedError:
		case <-a.ctx.Done():
			return
		default:
			log.Printf("Warning: error channel full, dropping error notification")
		}
	}
}

// Errors returns the error channel
func (a *Analyzer) Errors() <-chan *models.DetectedError {
	return a.errorChan
}

// GetError returns a specific error by ID
func (a *Analyzer) GetError(id string) *models.DetectedError {
	return a.classifier.GetError(id)
}

// GetAllErrors returns all detected errors
func (a *Analyzer) GetAllErrors() []*models.DetectedError {
	return a.classifier.GetErrors()
}

// GetErrorGroups returns all error groups
func (a *Analyzer) GetErrorGroups() []*models.ErrorGroup {
	return a.classifier.GetErrorGroups()
}

// GetRecentErrors returns recent errors for a service
func (a *Analyzer) GetRecentErrors(serviceName string, limit int) []*models.DetectedError {
	return a.classifier.GetRecentErrors(serviceName, limit)
}

// GetErrorStats returns error statistics
func (a *Analyzer) GetErrorStats() map[string]int {
	errors := a.classifier.GetErrors()

	stats := map[string]int{
		"total":    len(errors),
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	for _, err := range errors {
		switch err.Severity {
		case models.SeverityCritical:
			stats["critical"]++
		case models.SeverityHigh:
			stats["high"]++
		case models.SeverityMedium:
			stats["medium"]++
		case models.SeverityLow:
			stats["low"]++
		}
	}

	return stats
}
