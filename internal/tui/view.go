package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tapiaw38/tracehub/pkg/models"
)

// View renders the UI
func (m *Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	switch m.viewMode {
	case ViewDashboard:
		return m.renderDashboard()
	case ViewServiceDetail:
		return m.renderServiceDetail()
	case ViewErrorDetail:
		return m.renderErrorDetail()
	case ViewAnalyzing:
		return m.renderAnalyzing()
	case ViewFixProposal:
		return m.renderFixProposal()
	default:
		return "Unknown view"
	}
}

// renderDashboard renders the main dashboard
func (m *Model) renderDashboard() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Width(m.width).Render("TraceHub Dashboard"))
	b.WriteString("\n\n")

	// Services list
	b.WriteString(titleStyle.Render("Services"))
	b.WriteString("\n\n")

	for i, svc := range m.services {
		line := m.renderServiceLine(svc, i == m.selectedService)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Recent errors
	b.WriteString(titleStyle.Render("Recent Errors"))
	b.WriteString("\n\n")

	if len(m.errors) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("  No errors detected"))
		b.WriteString("\n")
	} else {
		// Show last 5 errors
		start := len(m.errors) - 5
		if start < 0 {
			start = 0
		}

		for i := len(m.errors) - 1; i >= start && i >= 0; i-- {
			err := m.errors[i]
			b.WriteString(m.renderErrorLine(err))
			b.WriteString("\n")
		}
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter("↑/↓: Navigate | Enter: Details | E: Errors | A: Analyze | Q: Quit"))

	// Status message
	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(baseStyle.Foreground(warningColor).Render(m.statusMessage))
	}

	return b.String()
}

// renderServiceLine renders a single service line
func (m *Model) renderServiceLine(svc *models.Service, selected bool) string {
	indicator := getStatusIndicator(string(svc.Status))

	name := svc.Name
	if selected {
		name = selectedErrorStyle.Render(name)
	}

	metrics := fmt.Sprintf("%s %.0f req/s %s %.1f%% errors",
		metricLabelStyle.Render("│"),
		svc.Metrics.RequestRate,
		metricLabelStyle.Render("│"),
		svc.Metrics.ErrorRate,
	)

	line := fmt.Sprintf("  %s %-30s %s", indicator, name, metrics)

	if selected {
		return lipgloss.NewStyle().Foreground(highlightColor).Render(line)
	}
	return line
}

// renderErrorLine renders a single error line
func (m *Model) renderErrorLine(err *models.DetectedError) string {
	timestamp := err.Timestamp.Format("15:04:05")
	severity := getSeverityStyle(string(err.Severity)).Render(string(err.Severity))

	msg := err.Message
	if len(msg) > 60 {
		msg = msg[:57] + "..."
	}

	return fmt.Sprintf("  [%s] %s %s: %s",
		lipgloss.NewStyle().Foreground(mutedColor).Render(timestamp),
		severity,
		err.ServiceName,
		msg,
	)
}

// renderServiceDetail renders the service detail view
func (m *Model) renderServiceDetail() string {
	if m.selectedService >= len(m.services) {
		return "Service not found"
	}

	svc := m.services[m.selectedService]
	var b strings.Builder

	// Header
	title := fmt.Sprintf("Service: %s", svc.Name)
	b.WriteString(headerStyle.Width(m.width).Render(title))
	b.WriteString("\n\n")

	// Status
	status := fmt.Sprintf("Status: %s", getStatusStyle(string(svc.Status)).Render(string(svc.Status)))
	b.WriteString(baseStyle.Render(status))
	b.WriteString("\n\n")

	// Metrics
	b.WriteString(titleStyle.Render("Metrics"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Request Rate:  %s\n", metricValueStyle.Render(fmt.Sprintf("%.2f req/s", svc.Metrics.RequestRate))))
	b.WriteString(fmt.Sprintf("  Error Rate:    %s\n", metricValueStyle.Render(fmt.Sprintf("%.2f%%", svc.Metrics.ErrorRate))))
	b.WriteString(fmt.Sprintf("  Total Requests: %s\n", metricValueStyle.Render(fmt.Sprintf("%d", svc.Metrics.TotalRequests))))
	b.WriteString(fmt.Sprintf("  Total Errors:   %s\n", metricValueStyle.Render(fmt.Sprintf("%d", svc.Metrics.TotalErrors))))

	b.WriteString("\n")

	// Recent logs
	b.WriteString(titleStyle.Render("Recent Logs"))
	b.WriteString("\n\n")

	logs := m.aggregator.GetRecentLogs(svc.Name, 10)
	if len(logs) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("  No logs available"))
	} else {
		for _, log := range logs {
			levelStyle := logStyle
			if log.Level == "error" || log.Level == "fatal" {
				levelStyle = logErrorStyle
			} else if log.Level == "warning" || log.Level == "warn" {
				levelStyle = logWarningStyle
			}

			msg := log.Message
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}

			b.WriteString(fmt.Sprintf("  [%s] %s %s\n",
				log.Timestamp.Format("15:04:05"),
				levelStyle.Render(fmt.Sprintf("%-7s", log.Level)),
				msg,
			))
		}
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter("A: Analyze Error | Q/ESC: Back"))

	return b.String()
}

// renderErrorDetail renders the error detail view
func (m *Model) renderErrorDetail() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Width(m.width).Render("Error Details"))
	b.WriteString("\n\n")

	if len(m.errors) == 0 {
		b.WriteString("No errors to display\n")
		b.WriteString("\n")
		b.WriteString(m.renderFooter("Q/ESC: Back"))
		return b.String()
	}

	// Errors list
	for i, err := range m.errors {
		if i == m.selectedError {
			b.WriteString(selectedErrorStyle.Render(m.renderFullError(err)))
		} else {
			b.WriteString(m.renderErrorLine(err))
		}
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter("↑/↓: Navigate | A: Analyze | Q/ESC: Back"))

	return b.String()
}

// renderFullError renders a full error with details
func (m *Model) renderFullError(err *models.DetectedError) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("ID: %s\n", err.ID))
	b.WriteString(fmt.Sprintf("Service: %s\n", err.ServiceName))
	b.WriteString(fmt.Sprintf("Type: %s\n", err.Type))
	b.WriteString(fmt.Sprintf("Severity: %s\n", getSeverityStyle(string(err.Severity)).Render(string(err.Severity))))
	b.WriteString(fmt.Sprintf("Count: %d\n", err.Count))
	b.WriteString(fmt.Sprintf("Message: %s\n", err.Message))

	if err.FilePath != "" {
		b.WriteString(fmt.Sprintf("File: %s:%d\n", err.FilePath, err.LineNumber))
	}

	return b.String()
}

// renderAnalyzing renders the analyzing view
func (m *Model) renderAnalyzing() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Width(m.width).Render("Analyzing with Claude AI"))
	b.WriteString("\n\n")

	b.WriteString(panelStyle.Render("🤖 Analyzing error and generating fix proposal...\n\nThis may take a few moments."))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderFooter("Please wait... | Q/ESC: Cancel"))

	return b.String()
}

// renderFixProposal renders the fix proposal view
func (m *Model) renderFixProposal() string {
	if m.currentProposal == nil {
		return "No proposal available"
	}

	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Width(m.width).Render("Fix Proposal"))
	b.WriteString("\n\n")

	// Root cause
	b.WriteString(titleStyle.Render("Root Cause"))
	b.WriteString("\n")
	b.WriteString(panelStyle.Render(m.currentProposal.RootCause))
	b.WriteString("\n\n")

	// Explanation
	b.WriteString(titleStyle.Render("Explanation"))
	b.WriteString("\n")
	b.WriteString(panelStyle.Render(m.currentProposal.Explanation))
	b.WriteString("\n\n")

	// Fix
	b.WriteString(titleStyle.Render("Proposed Fix"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("File: %s\n", m.currentProposal.Fix.FilePath))
	b.WriteString(fmt.Sprintf("Changes: %s\n\n", m.currentProposal.Fix.Changes))

	// Code preview (first 15 lines)
	codeLines := strings.Split(m.currentProposal.Fix.Code, "\n")
	previewLines := 15
	if len(codeLines) < previewLines {
		previewLines = len(codeLines)
	}
	codePreview := strings.Join(codeLines[:previewLines], "\n")
	if len(codeLines) > previewLines {
		codePreview += "\n... (truncated)"
	}
	b.WriteString(codeStyle.Render(codePreview))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderFooter("C: Create PR | R: Reject | Q/ESC: Back"))

	return b.String()
}

// renderFooter renders the footer with key bindings
func (m *Model) renderFooter(keys string) string {
	return footerStyle.Width(m.width).Render(keys)
}
