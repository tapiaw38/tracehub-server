package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tapiaw38/tracehub/pkg/models"
)

// AnalysisPrompt generates a prompt for analyzing an error
func AnalysisPrompt(err *models.DetectedError, codeSnippet string) string {
	// Format related logs
	var recentLogs strings.Builder
	for i, log := range err.RelatedLogs {
		if i >= 10 {
			break
		}
		recentLogs.WriteString(fmt.Sprintf("[%s] %s: %s\n",
			log.Timestamp.Format("15:04:05"),
			log.Level,
			log.Message))
	}

	prompt := fmt.Sprintf(`Eres un experto en debugging de sistemas Go. Analiza el siguiente error:

Error Type: %s
Service: %s
File: %s:%d
Timestamp: %s
Severity: %s

Error Message:
%s

Stack Trace:
%s

Logs relacionados (últimos 10):
%s

Código relevante:
%s

Por favor:
1. Identifica la causa raíz del error
2. Propón un fix específico en Go
3. Explica por qué ocurrió
4. Sugiere cómo prevenir errores similares
5. Genera tests para validar el fix

IMPORTANTE: Responde ÚNICAMENTE con un objeto JSON válido en el siguiente formato, sin texto adicional antes o después:
{
  "root_cause": "descripción de la causa raíz",
  "explanation": "explicación detallada de por qué ocurrió",
  "fix": {
    "file": "ruta/al/archivo.go",
    "changes": "descripción de los cambios",
    "code": "código completo del fix"
  },
  "tests": "código de tests para validar el fix",
  "prevention": "cómo prevenir errores similares en el futuro"
}`,
		err.Type,
		err.ServiceName,
		err.FilePath,
		err.LineNumber,
		err.Timestamp.Format("2006-01-02 15:04:05"),
		err.Severity,
		err.Message,
		err.StackTrace,
		recentLogs.String(),
		codeSnippet,
	)

	return prompt
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
	return `Eres un experto en debugging y desarrollo de sistemas en Go. Tu tarea es analizar errores en microservicios, identificar sus causas raíz y proponer fixes específicos y probados. Siempre respondes con código de calidad de producción, siguiendo las mejores prácticas de Go.`
}
