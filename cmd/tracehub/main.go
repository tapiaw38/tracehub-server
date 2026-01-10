package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/tapiaw38/tracehub/internal/collector"
	"github.com/tapiaw38/tracehub/internal/config"
	"github.com/tapiaw38/tracehub/internal/detector"
	"github.com/tapiaw38/tracehub/internal/tui"
)

var (
	cfgFile  string
	services []string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tracehub",
	Short: "TraceHub - AI-powered monitoring and auto-fix tool",
	Long: `TraceHub is a CLI tool that monitors microservice logs in real-time,
detects errors automatically, and uses Claude AI to generate fixes and create PRs.`,
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start monitoring services",
	Long:  `Start the interactive monitoring dashboard to track services in real-time.`,
	RunE:  runMonitor,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate example configuration file",
	Long:  `Generate an example tracehub.yaml configuration file in the current directory.`,
	RunE:  runInit,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a specific error manually",
	Long:  `Analyze a specific error from a log file using Claude AI.`,
	RunE:  runAnalyze,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./tracehub.yaml)")

	monitorCmd.Flags().StringSliceVar(&services, "services", []string{}, "specific services to monitor (comma-separated)")

	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(analyzeCmd)
}

func runMonitor(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Filter services if specified
	if len(services) > 0 {
		cfg.Services = filterServices(cfg.Services, services)
	}

	if len(cfg.Services) == 0 {
		return fmt.Errorf("no services configured")
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	// Create aggregator
	aggregator := collector.NewAggregator()
	aggregator.Start()
	defer aggregator.Stop()

	// Create analyzer
	analyzer := detector.NewAnalyzer(&cfg.Detection)
	analyzer.Start()
	defer analyzer.Stop()

	// Start log readers for each service
	readers := make([]*collector.LogReader, 0, len(cfg.Services))
	for _, svc := range cfg.Services {
		aggregator.AddService(svc.Name)

		reader := collector.NewLogReader(svc.Name, svc.LogPath, svc.Format)
		if err := reader.Start(ctx); err != nil {
			log.Printf("Warning: failed to start log reader for %s: %v", svc.Name, err)
			continue
		}
		readers = append(readers, reader)

		// Pipe logs from reader to aggregator and analyzer
		go func(r *collector.LogReader) {
			for logEntry := range r.Logs() {
				aggregator.AddLog(logEntry)
				analyzer.Analyze(logEntry)
			}
		}(reader)
	}

	// Create and run TUI
	model := tui.NewModel(cfg, aggregator, analyzer)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if _, err := p.Run(); err != nil {
			errChan <- err
		}
		cancel() // Cancel context when TUI exits
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		log.Println("Context cancelled, stopping...")
	case err := <-errChan:
		return fmt.Errorf("TUI error: %w", err)
	}

	// Cleanup
	for _, reader := range readers {
		reader.Stop()
	}

	log.Println("Monitoring stopped")
	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "tracehub.yaml"
	if cfgFile != "" {
		configPath = cfgFile
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists: %s", configPath)
	}

	// Create example configuration
	exampleConfig := `# TraceHub Configuration

services:
  - name: payment-service
    log_path: /var/log/payment/app.log
    format: json
    repo_path: /home/user/projects/payment

  - name: transfer-service
    log_path: /var/log/transfer/app.log
    format: json
    repo_path: /home/user/projects/transfer

detection:
  patterns:
    - "panic:"
    - "nil pointer"
    - "connection refused"
  severity_keywords:
    critical:
      - "panic"
      - "fatal"
    high:
      - "error"
      - "failed"
    medium:
      - "warning"
      - "timeout"

git:
  branch_prefix: "autofix/"
  commit_prefix: "[AutoFix]"

claude:
  model: "claude-sonnet-4-20250514"
  max_tokens: 4000
  # API key should be set via ANTHROPIC_API_KEY environment variable
`

	err := os.WriteFile(configPath, []byte(exampleConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Configuration file created: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the configuration file to match your setup")
	fmt.Println("2. Set ANTHROPIC_API_KEY environment variable")
	fmt.Println("3. Run 'tracehub monitor' to start monitoring")

	return nil
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("analyze command not yet implemented")
}

func loadConfig() (*config.Config, error) {
	configPath := cfgFile
	if configPath == "" {
		configPath = "tracehub.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func filterServices(services []config.ServiceConfig, filter []string) []config.ServiceConfig {
	if len(filter) == 0 {
		return services
	}

	filterMap := make(map[string]bool)
	for _, name := range filter {
		filterMap[name] = true
	}

	var filtered []config.ServiceConfig
	for _, svc := range services {
		if filterMap[svc.Name] {
			filtered = append(filtered, svc)
		}
	}

	return filtered
}
