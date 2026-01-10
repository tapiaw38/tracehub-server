package detector

import (
	"regexp"
	"strings"

	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// ErrorPattern represents a pattern for detecting errors
type ErrorPattern struct {
	Pattern     *regexp.Regexp
	ErrorType   models.ErrorType
	Severity    models.Severity
	Description string
}

// DefaultPatterns returns the default error patterns
func DefaultPatterns() []*ErrorPattern {
	return []*ErrorPattern{
		{
			Pattern:     regexp.MustCompile(`panic:`),
			ErrorType:   models.ErrorTypePanic,
			Severity:    models.SeverityCritical,
			Description: "Go panic detected",
		},
		{
			Pattern:     regexp.MustCompile(`runtime error: invalid memory address or nil pointer dereference`),
			ErrorType:   models.ErrorTypeNilPointer,
			Severity:    models.SeverityCritical,
			Description: "Nil pointer dereference",
		},
		{
			Pattern:     regexp.MustCompile(`nil pointer`),
			ErrorType:   models.ErrorTypeNilPointer,
			Severity:    models.SeverityHigh,
			Description: "Nil pointer error",
		},
		{
			Pattern:     regexp.MustCompile(`HTTP/\d\.\d" 5\d\d`),
			ErrorType:   models.ErrorTypeHTTP5xx,
			Severity:    models.SeverityHigh,
			Description: "HTTP 5xx server error",
		},
		{
			Pattern:     regexp.MustCompile(`connection refused|connection failed|dial tcp.*refused`),
			ErrorType:   models.ErrorTypeDatabaseConn,
			Severity:    models.SeverityHigh,
			Description: "Database connection error",
		},
		{
			Pattern:     regexp.MustCompile(`context deadline exceeded|timeout|timed out`),
			ErrorType:   models.ErrorTypeTimeout,
			Severity:    models.SeverityMedium,
			Description: "Timeout error",
		},
		{
			Pattern:     regexp.MustCompile(`fatal error:`),
			ErrorType:   models.ErrorTypePanic,
			Severity:    models.SeverityCritical,
			Description: "Fatal error",
		},
		{
			Pattern:     regexp.MustCompile(`out of memory|OOM|cannot allocate memory`),
			ErrorType:   models.ErrorTypeUnknown,
			Severity:    models.SeverityCritical,
			Description: "Out of memory error",
		},
	}
}

// PatternMatcher matches log entries against error patterns
type PatternMatcher struct {
	patterns         []*ErrorPattern
	customPatterns   []string
	severityKeywords map[models.Severity][]string
}

// NewPatternMatcher creates a new PatternMatcher
func NewPatternMatcher(customPatterns []string, severityKeywords map[models.Severity][]string) *PatternMatcher {
	return &PatternMatcher{
		patterns:         DefaultPatterns(),
		customPatterns:   customPatterns,
		severityKeywords: severityKeywords,
	}
}

// Match matches a log entry against patterns
func (pm *PatternMatcher) Match(entry *models.LogEntry) (bool, models.ErrorType, models.Severity) {
	text := entry.Message
	if text == "" {
		text = entry.Raw
	}

	// Check default patterns
	for _, pattern := range pm.patterns {
		if pattern.Pattern.MatchString(text) {
			return true, pattern.ErrorType, pattern.Severity
		}
	}

	// Check custom patterns
	for _, customPattern := range pm.customPatterns {
		if strings.Contains(strings.ToLower(text), strings.ToLower(customPattern)) {
			severity := pm.classifySeverity(text)
			return true, models.ErrorTypeUnknown, severity
		}
	}

	// Check if it's an error level log
	if isErrorLog(entry) {
		severity := pm.classifySeverity(text)
		return true, models.ErrorTypeUnknown, severity
	}

	return false, "", ""
}

// classifySeverity classifies the severity based on keywords
func (pm *PatternMatcher) classifySeverity(text string) models.Severity {
	lowerText := strings.ToLower(text)

	// Check severity keywords
	if keywords, ok := pm.severityKeywords[models.SeverityCritical]; ok {
		for _, keyword := range keywords {
			if strings.Contains(lowerText, strings.ToLower(keyword)) {
				return models.SeverityCritical
			}
		}
	}

	if keywords, ok := pm.severityKeywords[models.SeverityHigh]; ok {
		for _, keyword := range keywords {
			if strings.Contains(lowerText, strings.ToLower(keyword)) {
				return models.SeverityHigh
			}
		}
	}

	if keywords, ok := pm.severityKeywords[models.SeverityMedium]; ok {
		for _, keyword := range keywords {
			if strings.Contains(lowerText, strings.ToLower(keyword)) {
				return models.SeverityMedium
			}
		}
	}

	// Default severity based on common keywords
	criticalKeywords := []string{"panic", "fatal", "critical"}
	for _, keyword := range criticalKeywords {
		if strings.Contains(lowerText, keyword) {
			return models.SeverityCritical
		}
	}

	highKeywords := []string{"error", "failed", "failure"}
	for _, keyword := range highKeywords {
		if strings.Contains(lowerText, keyword) {
			return models.SeverityHigh
		}
	}

	mediumKeywords := []string{"warning", "warn", "timeout"}
	for _, keyword := range mediumKeywords {
		if strings.Contains(lowerText, keyword) {
			return models.SeverityMedium
		}
	}

	return models.SeverityLow
}

// isErrorLog checks if a log entry is at error level
func isErrorLog(entry *models.LogEntry) bool {
	errorLevels := []string{"error", "err", "fatal", "panic", "critical"}
	level := strings.ToLower(entry.Level)
	for _, el := range errorLevels {
		if level == el {
			return true
		}
	}
	return false
}

// ExtractFileAndLine extracts file path and line number from text
func ExtractFileAndLine(text string) (string, int) {
	// Pattern: /path/to/file.go:123
	fileLineRegex := regexp.MustCompile(`([a-zA-Z0-9_/.-]+\.go):(\d+)`)
	matches := fileLineRegex.FindStringSubmatch(text)
	if len(matches) >= 3 {
		line := 0
		// Parse line number from matches[2]
		for _, c := range matches[2] {
			if c >= '0' && c <= '9' {
				line = line*10 + int(c-'0')
			}
		}
		return matches[1], line
	}
	return "", 0
}
