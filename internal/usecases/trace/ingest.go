package trace

import (
	"context"
	"fmt"
	"time"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type IngestUsecase struct {
	traceRepo   TraceRepository
	projectRepo ProjectRepository
}

func NewIngestUsecase(traceRepo TraceRepository, projectRepo ProjectRepository) *IngestUsecase {
	return &IngestUsecase{
		traceRepo:   traceRepo,
		projectRepo: projectRepo,
	}
}

type IngestInput struct {
	Level       string            `json:"level" binding:"required"`
	Message     string            `json:"message" binding:"required"`
	Timestamp   *time.Time        `json:"timestamp"`
	Source      string            `json:"source"`
	StackTrace  string            `json:"stack_trace"`
	Context     map[string]string `json:"context"`
	ServiceName string            `json:"service_name"`
	Environment string            `json:"environment"`
}

type IngestBatchInput struct {
	Traces []IngestInput `json:"traces" binding:"required"`
}

func (uc *IngestUsecase) Execute(ctx context.Context, projectID string, input IngestInput) error {
	// Validate project exists
	_, err := uc.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Create trace
	trace := &domain.Trace{
		ProjectID:   projectID,
		Level:       input.Level,
		Message:     input.Message,
		Source:      input.Source,
		StackTrace:  input.StackTrace,
		Context:     input.Context,
		ServiceName: input.ServiceName,
		Environment: input.Environment,
	}

	// Use provided timestamp or current time
	if input.Timestamp != nil {
		trace.Timestamp = *input.Timestamp
	} else {
		trace.Timestamp = time.Now()
	}

	// Default environment
	if trace.Environment == "" {
		trace.Environment = "production"
	}

	// Create trace in repository
	if err := uc.traceRepo.Create(ctx, trace); err != nil {
		return fmt.Errorf("failed to create trace: %w", err)
	}

	return nil
}

func (uc *IngestUsecase) ExecuteBatch(ctx context.Context, projectID string, input IngestBatchInput) error {
	// Validate project exists
	_, err := uc.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Convert inputs to domain traces
	traces := make([]*domain.Trace, len(input.Traces))
	for i, t := range input.Traces {
		timestamp := time.Now()
		if t.Timestamp != nil {
			timestamp = *t.Timestamp
		}

		environment := t.Environment
		if environment == "" {
			environment = "production"
		}

		traces[i] = &domain.Trace{
			ProjectID:   projectID,
			Level:       t.Level,
			Message:     t.Message,
			Timestamp:   timestamp,
			Source:      t.Source,
			StackTrace:  t.StackTrace,
			Context:     t.Context,
			ServiceName: t.ServiceName,
			Environment: environment,
		}
	}

	// Batch insert
	if err := uc.traceRepo.CreateBatch(ctx, traces); err != nil {
		return fmt.Errorf("failed to create traces: %w", err)
	}

	return nil
}
