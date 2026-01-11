package ai

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// ContextBuilder builds context for Claude analysis
type ContextBuilder struct {
	repoPath string
}

// NewContextBuilder creates a new ContextBuilder
func NewContextBuilder(repoPath string) *ContextBuilder {
	return &ContextBuilder{
		repoPath: repoPath,
	}
}

// BuildContext builds context for analyzing an error
func (cb *ContextBuilder) BuildContext(err *models.DetectedError) (string, error) {
	if err.FilePath == "" {
		return "", fmt.Errorf("no file path in error")
	}

	// Read the file content
	fullPath := filepath.Join(cb.repoPath, err.FilePath)
	content, readErr := cb.readFile(fullPath)
	if readErr != nil {
		// If exact path doesn't work, try to find the file
		foundPath, findErr := cb.findFile(err.FilePath)
		if findErr != nil {
			return "", fmt.Errorf("failed to read file %s: %w (also tried to find: %w)", fullPath, readErr, findErr)
		}
		content, readErr = cb.readFile(foundPath)
		if readErr != nil {
			return "", fmt.Errorf("failed to read found file %s: %w", foundPath, readErr)
		}
	}

	// Extract relevant snippet around the error line
	snippet := cb.extractSnippet(content, err.LineNumber, err.FilePath)
	return snippet, nil
}

// readFile reads a file and returns its content
func (cb *ContextBuilder) readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return content.String(), nil
}

// extractSnippet extracts relevant code snippet around the error line
func (cb *ContextBuilder) extractSnippet(content string, lineNumber int, filePath string) string {
	lines := strings.Split(content, "\n")

	// If line number is 0 or invalid, return first 50 lines
	if lineNumber <= 0 || lineNumber > len(lines) {
		end := 50
		if len(lines) < end {
			end = len(lines)
		}
		return cb.formatSnippet(lines[0:end], 1, filePath)
	}

	// Extract 20 lines before and after the error line
	contextLines := 20
	start := lineNumber - contextLines - 1
	end := lineNumber + contextLines

	if start < 0 {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}

	snippet := lines[start:end]
	return cb.formatSnippet(snippet, start+1, filePath)
}

// detectLanguage detects the programming language from file extension
func detectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	switch strings.ToLower(ext) {
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py", ".pyw", ".pyi":
		return "python"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".rs":
		return "rust"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "cpp"
	case ".c":
		return "c"
	case ".cs":
		return "csharp"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".sh", ".bash":
		return "bash"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".sql":
		return "sql"
	case ".md":
		return "markdown"
	default:
		return "text"
	}
}

// formatSnippet formats a code snippet with line numbers
func (cb *ContextBuilder) formatSnippet(lines []string, startLine int, filePath string) string {
	language := detectLanguage(filePath)
	var formatted strings.Builder
	formatted.WriteString(fmt.Sprintf("```%s\n", language))
	for i, line := range lines {
		formatted.WriteString(fmt.Sprintf("%4d | %s\n", startLine+i, line))
	}
	formatted.WriteString("```\n")
	return formatted.String()
}

// findFile tries to find a file in the repository
func (cb *ContextBuilder) findFile(fileName string) (string, error) {
	// Get just the file name
	baseName := filepath.Base(fileName)

	var foundPath string
	err := filepath.Walk(cb.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		if info.IsDir() {
			// Skip common directories to speed up search
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Base(path) == baseName {
			foundPath = path
			return filepath.SkipAll // Stop walking
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", err
	}
	if foundPath == "" {
		return "", fmt.Errorf("file %s not found in %s", fileName, cb.repoPath)
	}

	return foundPath, nil
}
