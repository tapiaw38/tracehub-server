package trace

import (
	"context"
	"time"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type QueryUsecase struct {
	traceRepo TraceRepository
}

func NewQueryUsecase(traceRepo TraceRepository) *QueryUsecase {
	return &QueryUsecase{traceRepo: traceRepo}
}

type QueryInput struct {
	ProjectID   string   `form:"project_id"`
	Level       []string `form:"level"`
	ServiceName string   `form:"service_name"`
	Environment string   `form:"environment"`
	Since       string   `form:"since"`
	Until       string   `form:"until"`
	SearchText  string   `form:"search"`
	Limit       int      `form:"limit"`
	Offset      int      `form:"offset"`
}

type QueryOutput struct {
	Traces []*domain.Trace `json:"traces"`
	Total  int             `json:"total"`
}

func (uc *QueryUsecase) Execute(ctx context.Context, input QueryInput) (*QueryOutput, error) {
	filter := domain.TraceFilter{
		ProjectID:   input.ProjectID,
		Level:       input.Level,
		ServiceName: input.ServiceName,
		Environment: input.Environment,
		SearchText:  input.SearchText,
		Limit:       input.Limit,
		Offset:      input.Offset,
	}

	// Parse time filters
	if input.Since != "" {
		if t, err := time.Parse(time.RFC3339, input.Since); err == nil {
			filter.Since = t
		}
	}
	if input.Until != "" {
		if t, err := time.Parse(time.RFC3339, input.Until); err == nil {
			filter.Until = t
		}
	}

	// Default limit
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	traces, err := uc.traceRepo.Query(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &QueryOutput{
		Traces: traces,
		Total:  len(traces),
	}, nil
}
