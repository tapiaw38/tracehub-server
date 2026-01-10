package project

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByName(ctx context.Context, name string) (*domain.Project, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, project *domain.Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}

	query := `
		INSERT INTO projects (id, name, description, language, repo_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		project.ID,
		project.Name,
		project.Description,
		project.Language,
		project.RepoURL,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	query := `
		SELECT id, name, description, language, repo_url, api_key_id, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	project := &domain.Project{}
	var apiKeyID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Language,
		&project.RepoURL,
		&apiKeyID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if apiKeyID.Valid {
		project.ApiKeyID = apiKeyID.String
	}

	return project, nil
}

func (r *repository) GetByName(ctx context.Context, name string) (*domain.Project, error) {
	query := `
		SELECT id, name, description, language, repo_url, api_key_id, created_at, updated_at
		FROM projects
		WHERE name = $1
	`

	project := &domain.Project{}
	var apiKeyID sql.NullString

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Language,
		&project.RepoURL,
		&apiKeyID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if apiKeyID.Valid {
		project.ApiKeyID = apiKeyID.String
	}

	return project, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]*domain.Project, error) {
	query := `
		SELECT id, name, description, language, repo_url, api_key_id, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		project := &domain.Project{}
		var apiKeyID sql.NullString

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Language,
			&project.RepoURL,
			&apiKeyID,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		if apiKeyID.Valid {
			project.ApiKeyID = apiKeyID.String
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (r *repository) Update(ctx context.Context, project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, language = $3, repo_url = $4, updated_at = $5
		WHERE id = $6
	`

	project.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		project.Name,
		project.Description,
		project.Language,
		project.RepoURL,
		project.UpdatedAt,
		project.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}
