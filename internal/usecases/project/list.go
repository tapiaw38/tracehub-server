package project

import (
	"context"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type ListUsecase struct {
	repo ProjectRepository
}

func NewListUsecase(repo ProjectRepository) *ListUsecase {
	return &ListUsecase{repo: repo}
}

type ListOutput struct {
	Projects []*domain.Project `json:"projects"`
	Total    int               `json:"total"`
}

func (uc *ListUsecase) Execute(ctx context.Context, limit, offset int) (*ListOutput, error) {
	if limit == 0 {
		limit = 50
	}

	projects, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Projects: projects,
		Total:    len(projects),
	}, nil
}
