package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	primaryColor   = lipgloss.Color("86")  // Cyan
	successColor   = lipgloss.Color("46")  // Green
	warningColor   = lipgloss.Color("226") // Yellow
	errorColor     = lipgloss.Color("196") // Red
	criticalColor  = lipgloss.Color("201") // Magenta
	mutedColor     = lipgloss.Color("241") // Gray
	highlightColor = lipgloss.Color("212") // Pink

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(mutedColor).
			Padding(0, 1)

	// Service status styles
	serviceRunningStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	serviceErrorStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	serviceWarningStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	serviceStoppedStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	// Metrics styles
	metricLabelStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	metricValueStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	// Error styles
	errorListStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1).
			Margin(1, 0)

	errorItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedErrorStyle = lipgloss.NewStyle().
				Background(highlightColor).
				Foreground(lipgloss.Color("0")).
				Bold(true).
				Padding(0, 1)

	// Severity styles
	criticalStyle = lipgloss.NewStyle().
			Foreground(criticalColor).
			Bold(true)

	highStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	mediumStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	lowStyle = lipgloss.NewStyle().
		Foreground(mutedColor)

	// Log styles
	logStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1)

	logErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	logWarningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(mutedColor).
			Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Margin(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(highlightColor).
				Padding(1, 2).
				Margin(0, 1)

	// Status indicator styles
	statusIndicatorRunning = lipgloss.NewStyle().
				Foreground(successColor).
				SetString("●")

	statusIndicatorError = lipgloss.NewStyle().
				Foreground(errorColor).
				SetString("⚠")

	statusIndicatorWarning = lipgloss.NewStyle().
				Foreground(warningColor).
				SetString("⚠")

	statusIndicatorStopped = lipgloss.NewStyle().
				Foreground(mutedColor).
				SetString("○")

	// Code styles
	codeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("234")).
			Foreground(lipgloss.Color("252")).
			Padding(1, 2).
			Margin(1, 0)
)

// getSeverityStyle returns the style for a severity level
func getSeverityStyle(severity string) lipgloss.Style {
	switch severity {
	case "critical":
		return criticalStyle
	case "high":
		return highStyle
	case "medium":
		return mediumStyle
	case "low":
		return lowStyle
	default:
		return baseStyle
	}
}

// getStatusStyle returns the style for a service status
func getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "running":
		return serviceRunningStyle
	case "error":
		return serviceErrorStyle
	case "warning":
		return serviceWarningStyle
	case "stopped":
		return serviceStoppedStyle
	default:
		return baseStyle
	}
}

// getStatusIndicator returns the status indicator for a service status
func getStatusIndicator(status string) string {
	switch status {
	case "running":
		return statusIndicatorRunning.Render()
	case "error":
		return statusIndicatorError.Render()
	case "warning":
		return statusIndicatorWarning.Render()
	case "stopped":
		return statusIndicatorStopped.Render()
	default:
		return "○"
	}
}
