package project

import (
	"context"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByName(ctx context.Context, name string) (*domain.Project, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Project, error)
	Count(ctx context.Context) (int, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id string) error
}

type ApiKeyRepository interface {
	Create(ctx context.Context, apiKey *domain.ApiKey) error
	GetByKey(ctx context.Context, key string) (*domain.ApiKey, error)
	UpdateLastUsed(ctx context.Context, id string) error
}
