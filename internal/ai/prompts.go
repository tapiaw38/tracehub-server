package ai

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// detectLanguageFromFile detects the programming language from file path
func detectLanguageFromFile(filePath string) string {
	ext := filepath.Ext(filePath)
	switch strings.ToLower(ext) {
	case ".go":
		return "Go"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "JavaScript"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".py", ".pyw", ".pyi":
		return "Python"
	case ".java":
		return "Java"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".rs":
		return "Rust"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "C++"
	case ".c":
		return "C"
	case ".cs":
		return "C#"
	case ".swift":
		return "Swift"
	case ".kt":
		return "Kotlin"
	case ".scala":
		return "Scala"
	default:
		return "unknown"
	}
}

// AnalysisPrompt generates a prompt for analyzing an error
func AnalysisPrompt(err *models.DetectedError, codeSnippet string) string {
	language := detectLanguageFromFile(err.FilePath)
	timestampStr := ""
	if !err.FirstOccurred.IsZero() {
		timestampStr = err.FirstOccurred.Format("2006-01-02 15:04:05")
	}

	var recentLogs strings.Builder
	if err.RelatedLogs != nil {
		for i, log := range err.RelatedLogs {
			if i >= 10 {
				break
			}
			timestamp := ""
			if !log.Timestamp.IsZero() {
				timestamp = log.Timestamp.Format("15:04:05")
			}
			recentLogs.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				timestamp,
				log.Level,
				log.Message))
		}
	}

	languageContext := fmt.Sprintf("This error comes from a %s application", language)
	if language == "unknown" {
		languageContext = "This error comes from an application. Detect the language from the provided code."
	}

	prompt := fmt.Sprintf(`You are an expert in software systems debugging. Analyze the following error:

Language: %s
Error Type: %s
Service: %s
File: %s:%d
Timestamp: %s
Severity: %s

Error Message:
%s

Stack Trace:
%s

Related logs (last 10):
%s

Relevant code:
%s

%s

Please:
1. Identify the root cause of the error
2. Propose a specific fix in %s (adjust the language according to the code)
3. Explain why it occurred
4. Suggest how to prevent similar errors
5. Generate tests to validate the fix

IMPORTANT: Respond ONLY with a valid JSON object in the following format, without any additional text before or after:
{
  "root_cause": "description of the root cause",
  "explanation": "detailed explanation of why it occurred",
  "fix": {
    "file": "path/to/file.%s",
    "changes": "description of the changes",
    "code": "complete fix code"
  },
  "tests": "test code to validate the fix",
  "prevention": "how to prevent similar errors in the future"
}`,
		language,
		err.Type,
		err.ServiceName,
		err.FilePath,
		err.LineNumber,
		timestampStr,
		err.Severity,
		err.Message,
		err.StackTrace,
		recentLogs.String(),
		codeSnippet,
		languageContext,
		language,
		getFileExtension(err.FilePath),
	)

	return prompt
}

// getFileExtension returns the file extension without the dot
func getFileExtension(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext != "" {
		return strings.TrimPrefix(ext, ".")
	}
	return "txt"
}

// ParseClaudeResponse parses the JSON response from Claude
func ParseClaudeResponse(response string) (*models.ClaudeResponse, error) {
	// Clean the response - sometimes Claude adds markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var claudeResp models.ClaudeResponse
	if err := json.Unmarshal([]byte(response), &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w\nResponse: %s", err, response)
	}

	return &claudeResp, nil
}

// SystemPrompt returns the system prompt for Claude
func SystemPrompt() string {
	return `You are an expert in debugging and developing software systems. Your task is to analyze errors in applications and services, identify their root causes, and propose specific, tested fixes. You work with multiple programming languages (Go, JavaScript, TypeScript, Python, Java, Ruby, PHP, Rust, C++, C, C#, Swift, Kotlin, Scala, etc.) and always adapt your response to the language of the code you are analyzing. You always respond with production-quality code, following best practices for the corresponding language.`
}
