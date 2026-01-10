package collector

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tapiaw38/tracehub/pkg/models"
	"github.com/tidwall/gjson"
)

// Parser parses log entries from different formats
type Parser struct {
	format string
}

// NewParser creates a new Parser
func NewParser(format string) *Parser {
	return &Parser{
		format: format,
	}
}

// Parse parses a log line into a LogEntry
func (p *Parser) Parse(line, serviceName string) *models.LogEntry {
	if strings.TrimSpace(line) == "" {
		return nil
	}

	switch p.format {
	case "json":
		return p.parseJSON(line, serviceName)
	case "text", "plain":
		return p.parsePlainText(line, serviceName)
	default:
		// Try JSON first, fall back to plain text
		entry := p.parseJSON(line, serviceName)
		if entry == nil {
			entry = p.parsePlainText(line, serviceName)
		}
		return entry
	}
}

// parseJSON parses a JSON-formatted log line
func (p *Parser) parseJSON(line, serviceName string) *models.LogEntry {
	if !gjson.Valid(line) {
		return nil
	}

	parsed := gjson.Parse(line)

	entry := &models.LogEntry{
		ServiceName: serviceName,
		Raw:         line,
		Fields:      make(map[string]string),
	}

	// Extract timestamp
	if ts := parsed.Get("timestamp"); ts.Exists() {
		if t, err := time.Parse(time.RFC3339, ts.String()); err == nil {
			entry.Timestamp = t
		}
	} else if ts := parsed.Get("time"); ts.Exists() {
		if t, err := time.Parse(time.RFC3339, ts.String()); err == nil {
			entry.Timestamp = t
		}
	} else if ts := parsed.Get("@timestamp"); ts.Exists() {
		if t, err := time.Parse(time.RFC3339, ts.String()); err == nil {
			entry.Timestamp = t
		}
	}

	// Default to now if no timestamp found
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Extract level
	if level := parsed.Get("level"); level.Exists() {
		entry.Level = level.String()
	} else if level := parsed.Get("severity"); level.Exists() {
		entry.Level = level.String()
	} else {
		entry.Level = "info"
	}

	// Extract message
	if msg := parsed.Get("message"); msg.Exists() {
		entry.Message = msg.String()
	} else if msg := parsed.Get("msg"); msg.Exists() {
		entry.Message = msg.String()
	}

	// Extract additional fields
	parsed.ForEach(func(key, value gjson.Result) bool {
		k := key.String()
		if k != "timestamp" && k != "time" && k != "@timestamp" && k != "level" && k != "message" && k != "msg" {
			entry.Fields[k] = value.String()
		}
		return true
	})

	return entry
}

// parsePlainText parses a plain text log line
func (p *Parser) parsePlainText(line, serviceName string) *models.LogEntry {
	entry := &models.LogEntry{
		ServiceName: serviceName,
		Raw:         line,
		Timestamp:   time.Now(),
		Level:       "info",
		Message:     line,
		Fields:      make(map[string]string),
	}

	// Try to extract timestamp from common formats
	// Format: 2024-01-10 15:34:22
	timestampRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`)
	if matches := timestampRegex.FindStringSubmatch(line); len(matches) > 1 {
		if t, err := time.Parse("2006-01-02 15:04:05", matches[1]); err == nil {
			entry.Timestamp = t
		}
	}

	// Try ISO8601 format
	iso8601Regex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?)`)
	if matches := iso8601Regex.FindStringSubmatch(line); len(matches) > 1 {
		if t, err := time.Parse(time.RFC3339, matches[1]); err == nil {
			entry.Timestamp = t
		}
	}

	// Extract level from common patterns
	levelRegex := regexp.MustCompile(`(?i)\[(ERROR|WARN|WARNING|INFO|DEBUG|FATAL|PANIC)\]`)
	if matches := levelRegex.FindStringSubmatch(line); len(matches) > 1 {
		entry.Level = strings.ToLower(matches[1])
	}

	// Check for error keywords in the line
	lowerLine := strings.ToLower(line)
	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "failed") {
		entry.Level = "error"
	} else if strings.Contains(lowerLine, "warn") {
		entry.Level = "warning"
	} else if strings.Contains(lowerLine, "panic") || strings.Contains(lowerLine, "fatal") {
		entry.Level = "fatal"
	}

	return entry
}

// ExtractStackTrace extracts stack trace from log lines
func ExtractStackTrace(logs []string) string {
	var stackTrace []string
	inStackTrace := false

	for _, line := range logs {
		// Stack trace patterns
		if strings.Contains(line, "goroutine") ||
		   strings.Contains(line, "panic:") ||
		   strings.Contains(line, "runtime error:") {
			inStackTrace = true
			stackTrace = append(stackTrace, line)
			continue
		}

		if inStackTrace {
			// Stack trace lines typically start with whitespace or contain file:line references
			if strings.HasPrefix(line, "\t") ||
			   strings.HasPrefix(line, "  ") ||
			   regexp.MustCompile(`\w+\.go:\d+`).MatchString(line) {
				stackTrace = append(stackTrace, line)
			} else if strings.TrimSpace(line) == "" {
				continue
			} else {
				// End of stack trace
				break
			}
		}
	}

	return strings.Join(stackTrace, "\n")
}

// ParseError attempts to parse error information from a log entry
func ParseError(entry *models.LogEntry) (filePath string, lineNumber int, errorType string) {
	// Try to extract file:line from the message or stack trace
	fileLineRegex := regexp.MustCompile(`([a-zA-Z0-9_/.-]+\.go):(\d+)`)
	if matches := fileLineRegex.FindStringSubmatch(entry.Message); len(matches) > 2 {
		filePath = matches[1]
		// Parse line number
		fmt.Sscanf(matches[2], "%d", &lineNumber)
	}

	// Detect error type
	lowerMsg := strings.ToLower(entry.Message)
	if strings.Contains(lowerMsg, "panic") {
		errorType = "panic"
	} else if strings.Contains(lowerMsg, "nil pointer") {
		errorType = "nil_pointer"
	} else if strings.Contains(lowerMsg, "connection refused") || strings.Contains(lowerMsg, "connection failed") {
		errorType = "database_connection"
	} else if strings.Contains(lowerMsg, "timeout") {
		errorType = "timeout"
	} else if regexp.MustCompile(`50\d`).MatchString(entry.Message) {
		errorType = "http_5xx"
	}

	return
}
