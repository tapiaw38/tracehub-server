package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tapiaw38/tracehub-server/internal/ai"
	"github.com/tapiaw38/tracehub-server/internal/collector"
	"github.com/tapiaw38/tracehub-server/internal/config"
	"github.com/tapiaw38/tracehub-server/internal/detector"
	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewServiceDetail
	ViewErrorDetail
	ViewAnalyzing
	ViewFixProposal
)

// Model represents the application state
type Model struct {
	// Configuration
	config *config.Config

	// Components
	aggregator *collector.Aggregator
	analyzer   *detector.Analyzer
	aiClient   *ai.Client

	// State
	services        []*models.Service
	selectedService int
	errors          []*models.DetectedError
	selectedError   int
	currentProposal *models.FixProposal
	viewMode        ViewMode

	// UI components
	viewport viewport.Model
	width    int
	height   int

	// Logs
	recentLogs []*models.LogEntry

	// Status messages
	statusMessage string
	isAnalyzing   bool

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, aggregator *collector.Aggregator, analyzer *detector.Analyzer) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize services
	services := make([]*models.Service, len(cfg.Services))
	for i, svc := range cfg.Services {
		services[i] = models.NewService(svc.Name, svc.LogPath, svc.Format, svc.RepoPath)
	}

	// Create viewport for scrolling
	vp := viewport.New(80, 20)

	return &Model{
		config:          cfg,
		aggregator:      aggregator,
		analyzer:        analyzer,
		services:        services,
		selectedService: 0,
		errors:          make([]*models.DetectedError, 0),
		selectedError:   0,
		viewMode:        ViewDashboard,
		viewport:        vp,
		recentLogs:      make([]*models.LogEntry, 0),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		waitForErrors(m.analyzer),
	)
}

// Msg types
type tickMsg time.Time
type errorDetectedMsg *models.DetectedError
type analysisCompleteMsg *models.FixProposal
type analysisErrorMsg error

// tickCmd returns a command that sends a tick message
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// waitForErrors waits for new errors from the analyzer
func waitForErrors(analyzer *detector.Analyzer) tea.Cmd {
	return func() tea.Msg {
		for err := range analyzer.Errors() {
			return errorDetectedMsg(err)
		}
		return nil
	}
}

// analyzeError analyzes an error with AI
func analyzeError(client *ai.Client, err *models.DetectedError) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		proposal, analyzeErr := client.AnalyzeError(ctx, err)
		if analyzeErr != nil {
			return analysisErrorMsg(analyzeErr)
		}
		return analysisCompleteMsg(proposal)
	}
}
