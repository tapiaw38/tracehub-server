package tui

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tapiaw38/tracehub-server/internal/ai"
	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 10 // Reserve space for header and footer
		return m, nil

	case tickMsg:
		// Update metrics
		m.updateMetrics()
		// Schedule next tick
		return m, tickCmd()

	case errorDetectedMsg:
		// New error detected
		err := (*models.DetectedError)(msg)
		m.errors = append(m.errors, err)

		// Update service status
		for _, svc := range m.services {
			if svc.Name == err.ServiceName {
				svc.Status = "error"
				svc.LastError = err
				svc.Metrics.TotalErrors++
			}
		}

		m.statusMessage = "New error detected: " + err.Message
		return m, waitForErrors(m.analyzer)

	case analysisCompleteMsg:
		// Analysis complete
		proposal := (*models.FixProposal)(msg)
		m.currentProposal = proposal
		m.isAnalyzing = false
		m.viewMode = ViewFixProposal
		m.statusMessage = "Analysis complete"
		return m, nil

	case analysisErrorMsg:
		// Analysis error
		m.isAnalyzing = false
		m.statusMessage = "Analysis failed: " + msg.Error()
		log.Printf("Analysis error: %v", msg)
		return m, nil
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case ViewDashboard:
		return m.handleDashboardKeys(msg)
	case ViewServiceDetail:
		return m.handleServiceDetailKeys(msg)
	case ViewErrorDetail:
		return m.handleErrorDetailKeys(msg)
	case ViewAnalyzing:
		return m.handleAnalyzingKeys(msg)
	case ViewFixProposal:
		return m.handleFixProposalKeys(msg)
	}
	return m, nil
}

// handleDashboardKeys handles keys in dashboard view
func (m *Model) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.cancel()
		return m, tea.Quit

	case "up", "k":
		if m.selectedService > 0 {
			m.selectedService--
		}

	case "down", "j":
		if m.selectedService < len(m.services)-1 {
			m.selectedService++
		}

	case "enter":
		// View service details
		m.viewMode = ViewServiceDetail
		return m, nil

	case "e":
		// View errors list
		if len(m.errors) > 0 {
			m.viewMode = ViewErrorDetail
		}

	case "a":
		// Analyze selected error
		if len(m.errors) > 0 {
			return m.startAnalysis()
		}
	}

	return m, nil
}

// handleServiceDetailKeys handles keys in service detail view
func (m *Model) handleServiceDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.viewMode = ViewDashboard
		return m, nil

	case "a":
		// Analyze latest error for this service
		serviceName := m.services[m.selectedService].Name
		for i := len(m.errors) - 1; i >= 0; i-- {
			if m.errors[i].ServiceName == serviceName {
				m.selectedError = i
				return m.startAnalysis()
			}
		}
	}

	return m, nil
}

// handleErrorDetailKeys handles keys in error detail view
func (m *Model) handleErrorDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.viewMode = ViewDashboard
		return m, nil

	case "up", "k":
		if m.selectedError > 0 {
			m.selectedError--
		}

	case "down", "j":
		if m.selectedError < len(m.errors)-1 {
			m.selectedError++
		}

	case "a":
		// Analyze selected error
		return m.startAnalysis()
	}

	return m, nil
}

// handleAnalyzingKeys handles keys in analyzing view
func (m *Model) handleAnalyzingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.viewMode = ViewDashboard
		m.isAnalyzing = false
		return m, nil
	}

	return m, nil
}

// handleFixProposalKeys handles keys in fix proposal view
func (m *Model) handleFixProposalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.viewMode = ViewDashboard
		return m, nil

	case "c":
		// Create PR (placeholder)
		m.statusMessage = "PR creation not yet implemented"
		m.viewMode = ViewDashboard
		return m, nil

	case "r":
		// Reject fix
		m.currentProposal.Status = "rejected"
		m.statusMessage = "Fix rejected"
		m.viewMode = ViewDashboard
		return m, nil
	}

	return m, nil
}

// startAnalysis starts analyzing an error
func (m *Model) startAnalysis() (tea.Model, tea.Cmd) {
	if m.selectedError >= len(m.errors) {
		return m, nil
	}

	selectedErr := m.errors[m.selectedError]

	// Get service config for repo path
	svcConfig := m.config.GetService(selectedErr.ServiceName)
	if svcConfig == nil {
		m.statusMessage = "Service configuration not found"
		return m, nil
	}

	// Create AI client for this service
	aiClient := ai.NewClient(&m.config.Claude, svcConfig.RepoPath)

	m.isAnalyzing = true
	m.viewMode = ViewAnalyzing
	m.statusMessage = "Analyzing error with Claude AI..."

	return m, analyzeError(aiClient, selectedErr)
}

// updateMetrics updates service metrics
func (m *Model) updateMetrics() {
	for _, svc := range m.services {
		metrics := m.aggregator.GetServiceMetrics(svc.Name)
		if metrics != nil {
			svc.Metrics = metrics
		}

		// Update service status based on error rate
		if svc.Metrics.ErrorRate > 5.0 {
			svc.Status = "error"
		} else if svc.Metrics.ErrorRate > 1.0 {
			svc.Status = "warning"
		} else {
			svc.Status = "running"
		}

		svc.LastUpdated = time.Now()
	}
}
