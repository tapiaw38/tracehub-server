package detector

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// Classifier classifies and groups errors
type Classifier struct {
	errorGroups map[string]*models.ErrorGroup
	errors      map[string]*models.DetectedError
	mu          sync.RWMutex
}

// NewClassifier creates a new Classifier
func NewClassifier() *Classifier {
	return &Classifier{
		errorGroups: make(map[string]*models.ErrorGroup),
		errors:      make(map[string]*models.DetectedError),
	}
}

// Classify classifies an error and groups similar errors
func (c *Classifier) Classify(entry *models.LogEntry, errorType models.ErrorType, severity models.Severity) *models.DetectedError {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate error signature for grouping
	signature := c.generateSignature(entry, errorType)

	// Check if we've seen this error before
	if existingError, exists := c.errors[signature]; exists {
		// Update existing error
		existingError.Count++
		existingError.LastOccurred = entry.Timestamp
		return existingError
	}

	// Create new error
	filePath, lineNumber := ExtractFileAndLine(entry.Message)

	detectedError := &models.DetectedError{
		ID:            generateID(entry),
		ServiceName:   entry.ServiceName,
		Type:          errorType,
		Severity:      severity,
		Message:       entry.Message,
		FilePath:      filePath,
		LineNumber:    lineNumber,
		Timestamp:     entry.Timestamp,
		Context:       entry.Fields,
		RelatedLogs:   []*models.LogEntry{entry},
		Count:         1,
		FirstOccurred: entry.Timestamp,
		LastOccurred:  entry.Timestamp,
	}

	// Store error
	c.errors[signature] = detectedError

	// Add to error group
	c.addToGroup(signature, detectedError)

	return detectedError
}

// generateSignature generates a unique signature for an error
func (c *Classifier) generateSignature(entry *models.LogEntry, errorType models.ErrorType) string {
	// Normalize the error message to group similar errors
	normalized := normalizeErrorMessage(entry.Message)

	// Include service name and error type in signature
	data := fmt.Sprintf("%s:%s:%s", entry.ServiceName, errorType, normalized)

	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// normalizeErrorMessage normalizes error messages for grouping
func normalizeErrorMessage(message string) string {
	// Remove timestamps
	message = removeTimestamps(message)

	// Remove line numbers (file.go:123 -> file.go:XXX)
	message = regexp.MustCompile(`\.go:\d+`).ReplaceAllString(message, ".go:XXX")

	// Remove memory addresses (0xc00012e000 -> 0xXXX)
	message = regexp.MustCompile(`0x[0-9a-fA-F]+`).ReplaceAllString(message, "0xXXX")

	// Remove numeric IDs
	message = regexp.MustCompile(`\bid=\d+`).ReplaceAllString(message, "id=XXX")
	message = regexp.MustCompile(`\buuid=[a-f0-9-]+`).ReplaceAllString(message, "uuid=XXX")

	return message
}

// removeTimestamps removes timestamps from text
func removeTimestamps(text string) string {
	// Remove ISO8601 timestamps
	text = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?`).ReplaceAllString(text, "")

	// Remove simple timestamps
	text = regexp.MustCompile(`\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}`).ReplaceAllString(text, "")

	return text
}

// addToGroup adds an error to its group
func (c *Classifier) addToGroup(signature string, err *models.DetectedError) {
	// Use error type as group pattern
	pattern := string(err.Type)

	if group, exists := c.errorGroups[pattern]; exists {
		group.Errors = append(group.Errors, err)
		group.TotalCount++
		group.LastOccurred = err.LastOccurred
	} else {
		c.errorGroups[pattern] = &models.ErrorGroup{
			Pattern:       pattern,
			Errors:        []*models.DetectedError{err},
			TotalCount:    1,
			FirstOccurred: err.FirstOccurred,
			LastOccurred:  err.LastOccurred,
		}
	}
}

// GetError returns an error by ID
func (c *Classifier) GetError(id string) *models.DetectedError {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, err := range c.errors {
		if err.ID == id {
			return err
		}
	}
	return nil
}

// GetErrors returns all detected errors
func (c *Classifier) GetErrors() []*models.DetectedError {
	c.mu.RLock()
	defer c.mu.RUnlock()

	errors := make([]*models.DetectedError, 0, len(c.errors))
	for _, err := range c.errors {
		errors = append(errors, err)
	}
	return errors
}

// GetErrorGroups returns all error groups
func (c *Classifier) GetErrorGroups() []*models.ErrorGroup {
	c.mu.RLock()
	defer c.mu.RUnlock()

	groups := make([]*models.ErrorGroup, 0, len(c.errorGroups))
	for _, group := range c.errorGroups {
		groups = append(groups, group)
	}
	return groups
}

// GetRecentErrors returns recent errors for a service
func (c *Classifier) GetRecentErrors(serviceName string, limit int) []*models.DetectedError {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var errors []*models.DetectedError
	for _, err := range c.errors {
		if err.ServiceName == serviceName {
			errors = append(errors, err)
		}
	}

	// Sort by timestamp (most recent first)
	// Simple implementation - in production, use sort.Slice
	if len(errors) > limit {
		errors = errors[:limit]
	}

	return errors
}

// generateID generates a unique ID for an error
func generateID(entry *models.LogEntry) string {
	timestamp := entry.Timestamp.Format(time.RFC3339Nano)
	data := fmt.Sprintf("%s:%s:%s", entry.ServiceName, timestamp, entry.Message)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}
