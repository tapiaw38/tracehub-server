package project

import (
	"context"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type GetUsecase struct {
	repo ProjectRepository
}

func NewGetUsecase(repo ProjectRepository) *GetUsecase {
	return &GetUsecase{repo: repo}
}

func (uc *GetUsecase) Execute(ctx context.Context, id string) (*domain.Project, error) {
	return uc.repo.GetByID(ctx, id)
}
