package collector

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/nxadm/tail"
	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// LogReader reads logs from a file in real-time
type LogReader struct {
	serviceName string
	logPath     string
	format      string
	parser      *Parser
	tail        *tail.Tail
	logChan     chan *models.LogEntry
	mu          sync.Mutex
	running     bool
}

// NewLogReader creates a new LogReader instance
func NewLogReader(serviceName, logPath, format string) *LogReader {
	return &LogReader{
		serviceName: serviceName,
		logPath:     logPath,
		format:      format,
		parser:      NewParser(format),
		logChan:     make(chan *models.LogEntry, 100),
		running:     false,
	}
}

// Start starts reading logs
func (lr *LogReader) Start(ctx context.Context) error {
	lr.mu.Lock()
	if lr.running {
		lr.mu.Unlock()
		return fmt.Errorf("log reader already running")
	}
	lr.running = true
	lr.mu.Unlock()

	// Configure tail to follow the file
	config := tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true, // Use polling for better compatibility
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: 2, // Start from end of file
		},
	}

	t, err := tail.TailFile(lr.logPath, config)
	if err != nil {
		lr.mu.Lock()
		lr.running = false
		lr.mu.Unlock()
		return fmt.Errorf("failed to tail file %s: %w", lr.logPath, err)
	}

	lr.tail = t

	// Start reading in a goroutine
	go lr.readLoop(ctx)

	log.Printf("Started reading logs for service %s from %s", lr.serviceName, lr.logPath)
	return nil
}

// readLoop reads lines from the tail and parses them
func (lr *LogReader) readLoop(ctx context.Context) {
	defer func() {
		lr.mu.Lock()
		lr.running = false
		lr.mu.Unlock()
		close(lr.logChan)
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping log reader for service %s", lr.serviceName)
			if lr.tail != nil {
				lr.tail.Stop()
			}
			return

		case line, ok := <-lr.tail.Lines:
			if !ok {
				log.Printf("Log tail closed for service %s", lr.serviceName)
				return
			}

			if line.Err != nil {
				log.Printf("Error reading log line for %s: %v", lr.serviceName, line.Err)
				continue
			}

			// Parse the log line
			logEntry := lr.parser.Parse(line.Text, lr.serviceName)
			if logEntry != nil {
				select {
				case lr.logChan <- logEntry:
				case <-ctx.Done():
					return
				default:
					// Channel full, drop oldest
					log.Printf("Warning: log channel full for %s, dropping log", lr.serviceName)
				}
			}
		}
	}
}

// Logs returns the channel for receiving log entries
func (lr *LogReader) Logs() <-chan *models.LogEntry {
	return lr.logChan
}

// Stop stops the log reader
func (lr *LogReader) Stop() {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if lr.tail != nil {
		lr.tail.Stop()
		lr.tail.Cleanup()
	}
	lr.running = false
}

// IsRunning returns whether the reader is currently running
func (lr *LogReader) IsRunning() bool {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	return lr.running
}
