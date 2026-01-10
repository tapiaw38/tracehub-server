package project

import (
	"context"
	"fmt"

	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type CreateUsecase struct {
	repo ProjectRepository
}

func NewCreateUsecase(repo ProjectRepository) *CreateUsecase {
	return &CreateUsecase{repo: repo}
}

type CreateInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Language    string `json:"language" binding:"required"`
	RepoURL     string `json:"repo_url"`
}

type CreateOutput struct {
	Project *domain.Project `json:"project"`
}

func (uc *CreateUsecase) Execute(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	project := &domain.Project{
		Name:        input.Name,
		Description: input.Description,
		Language:    input.Language,
		RepoURL:     input.RepoURL,
	}

	if err := uc.repo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &CreateOutput{Project: project}, nil
}
