package collector

import (
	"container/ring"
	"context"
	"sync"
	"time"

	"github.com/tapiaw38/tracehub-server/pkg/models"
)

const (
	// MaxLogsInMemory limits the number of logs kept in memory per service
	MaxLogsInMemory = 1000
	// MetricsWindowDuration is the time window for calculating metrics
	MetricsWindowDuration = 1 * time.Hour
)

// Aggregator aggregates logs from multiple services
type Aggregator struct {
	services    map[string]*ServiceAggregator
	mu          sync.RWMutex
	logChan     chan *models.LogEntry
	errorChan   chan *models.DetectedError
	ctx         context.Context
	cancel      context.CancelFunc
}

// ServiceAggregator aggregates logs for a single service
type ServiceAggregator struct {
	serviceName string
	logs        *ring.Ring
	recentLogs  []*models.LogEntry
	metrics     *models.Metrics
	mu          sync.RWMutex
}

// NewAggregator creates a new Aggregator
func NewAggregator() *Aggregator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Aggregator{
		services:  make(map[string]*ServiceAggregator),
		logChan:   make(chan *models.LogEntry, 1000),
		errorChan: make(chan *models.DetectedError, 100),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the aggregator
func (a *Aggregator) Start() {
	go a.processLogs()
}

// Stop stops the aggregator
func (a *Aggregator) Stop() {
	a.cancel()
}

// AddService adds a service to track
func (a *Aggregator) AddService(serviceName string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.services[serviceName]; !exists {
		a.services[serviceName] = &ServiceAggregator{
			serviceName: serviceName,
			logs:        ring.New(MaxLogsInMemory),
			recentLogs:  make([]*models.LogEntry, 0),
			metrics: &models.Metrics{
				TotalRequests: 0,
				TotalErrors:   0,
				RequestRate:   0,
				ErrorRate:     0,
			},
		}
	}
}

// AddLog adds a log entry to the aggregator
func (a *Aggregator) AddLog(log *models.LogEntry) {
	select {
	case a.logChan <- log:
	default:
		// Channel full, drop log
	}
}

// processLogs processes incoming log entries
func (a *Aggregator) processLogs() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return

		case log := <-a.logChan:
			a.processLog(log)

		case <-ticker.C:
			a.updateMetrics()
		}
	}
}

// processLog processes a single log entry
func (a *Aggregator) processLog(log *models.LogEntry) {
	a.mu.RLock()
	svcAgg, exists := a.services[log.ServiceName]
	a.mu.RUnlock()

	if !exists {
		return
	}

	svcAgg.mu.Lock()
	defer svcAgg.mu.Unlock()

	// Add to ring buffer
	svcAgg.logs.Value = log
	svcAgg.logs = svcAgg.logs.Next()

	// Add to recent logs (keep last 100)
	svcAgg.recentLogs = append(svcAgg.recentLogs, log)
	if len(svcAgg.recentLogs) > 100 {
		svcAgg.recentLogs = svcAgg.recentLogs[1:]
	}

	// Update metrics
	svcAgg.metrics.TotalRequests++
	if isErrorLevel(log.Level) {
		svcAgg.metrics.TotalErrors++
	}
}

// updateMetrics updates metrics for all services
func (a *Aggregator) updateMetrics() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, svcAgg := range a.services {
		svcAgg.mu.Lock()

		// Calculate rates (simple implementation)
		if svcAgg.metrics.TotalRequests > 0 {
			svcAgg.metrics.ErrorRate = float64(svcAgg.metrics.TotalErrors) / float64(svcAgg.metrics.TotalRequests) * 100
		}

		svcAgg.mu.Unlock()
	}
}

// GetServiceMetrics returns metrics for a service
func (a *Aggregator) GetServiceMetrics(serviceName string) *models.Metrics {
	a.mu.RLock()
	svcAgg, exists := a.services[serviceName]
	a.mu.RUnlock()

	if !exists {
		return nil
	}

	svcAgg.mu.RLock()
	defer svcAgg.mu.RUnlock()

	// Return a copy
	metrics := *svcAgg.metrics
	return &metrics
}

// GetRecentLogs returns recent logs for a service
func (a *Aggregator) GetRecentLogs(serviceName string, limit int) []*models.LogEntry {
	a.mu.RLock()
	svcAgg, exists := a.services[serviceName]
	a.mu.RUnlock()

	if !exists {
		return nil
	}

	svcAgg.mu.RLock()
	defer svcAgg.mu.RUnlock()

	// Get the most recent logs
	start := len(svcAgg.recentLogs) - limit
	if start < 0 {
		start = 0
	}

	logs := make([]*models.LogEntry, len(svcAgg.recentLogs[start:]))
	copy(logs, svcAgg.recentLogs[start:])
	return logs
}

// GetAllLogs returns all logs in the ring buffer for a service
func (a *Aggregator) GetAllLogs(serviceName string) []*models.LogEntry {
	a.mu.RLock()
	svcAgg, exists := a.services[serviceName]
	a.mu.RUnlock()

	if !exists {
		return nil
	}

	svcAgg.mu.RLock()
	defer svcAgg.mu.RUnlock()

	var logs []*models.LogEntry
	svcAgg.logs.Do(func(v interface{}) {
		if v != nil {
			if log, ok := v.(*models.LogEntry); ok {
				logs = append(logs, log)
			}
		}
	})

	return logs
}

// isErrorLevel checks if a log level indicates an error
func isErrorLevel(level string) bool {
	errorLevels := []string{"error", "err", "fatal", "panic", "critical"}
	for _, el := range errorLevels {
		if level == el {
			return true
		}
	}
	return false
}

// ErrorChan returns the error channel
func (a *Aggregator) ErrorChan() <-chan *models.DetectedError {
	return a.errorChan
}
