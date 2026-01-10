package trace

import (
	"context"
	"time"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type TraceRepository interface {
	Create(ctx context.Context, trace *domain.Trace) error
	CreateBatch(ctx context.Context, traces []*domain.Trace) error
	GetByID(ctx context.Context, id string) (*domain.Trace, error)
	Query(ctx context.Context, filter domain.TraceFilter) ([]*domain.Trace, error)
	Count(ctx context.Context, filter domain.TraceFilter) (int, error)
	DeleteOlderThan(ctx context.Context, projectID string, before time.Time) error
}

type ProjectRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Project, error)
}
