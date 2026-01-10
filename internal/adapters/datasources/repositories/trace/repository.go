package trace

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tapiaw38/tracehub-server/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, trace *domain.Trace) error
	CreateBatch(ctx context.Context, traces []*domain.Trace) error
	GetByID(ctx context.Context, id string) (*domain.Trace, error)
	Query(ctx context.Context, filter domain.TraceFilter) ([]*domain.Trace, error)
	Count(ctx context.Context, filter domain.TraceFilter) (int, error)
	DeleteOlderThan(ctx context.Context, projectID string, before time.Time) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, trace *domain.Trace) error {
	if trace.ID == "" {
		trace.ID = uuid.New().String()
	}

	contextJSON, err := json.Marshal(trace.Context)
	if err != nil {
		contextJSON = []byte("{}")
	}

	query := `
		INSERT INTO traces (id, project_id, level, message, timestamp, source, stack_trace, context, service_name, environment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	trace.CreatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		trace.ID,
		trace.ProjectID,
		trace.Level,
		trace.Message,
		trace.Timestamp,
		trace.Source,
		trace.StackTrace,
		contextJSON,
		trace.ServiceName,
		trace.Environment,
		trace.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create trace: %w", err)
	}

	return nil
}

func (r *repository) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	if len(traces) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO traces (id, project_id, level, message, timestamp, source, stack_trace, context, service_name, environment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, trace := range traces {
		if trace.ID == "" {
			trace.ID = uuid.New().String()
		}

		contextJSON, _ := json.Marshal(trace.Context)
		trace.CreatedAt = time.Now()

		_, err = stmt.ExecContext(ctx,
			trace.ID,
			trace.ProjectID,
			trace.Level,
			trace.Message,
			trace.Timestamp,
			trace.Source,
			trace.StackTrace,
			contextJSON,
			trace.ServiceName,
			trace.Environment,
			trace.CreatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to insert trace: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*domain.Trace, error) {
	query := `
		SELECT id, project_id, level, message, timestamp, source, stack_trace, context, service_name, environment, created_at
		FROM traces
		WHERE id = $1
	`

	trace := &domain.Trace{}
	var contextJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&trace.ID,
		&trace.ProjectID,
		&trace.Level,
		&trace.Message,
		&trace.Timestamp,
		&trace.Source,
		&trace.StackTrace,
		&contextJSON,
		&trace.ServiceName,
		&trace.Environment,
		&trace.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("trace not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	if len(contextJSON) > 0 {
		json.Unmarshal(contextJSON, &trace.Context)
	}

	return trace, nil
}

func (r *repository) Query(ctx context.Context, filter domain.TraceFilter) ([]*domain.Trace, error) {
	query := `
		SELECT id, project_id, level, message, timestamp, source, stack_trace, context, service_name, environment, created_at
		FROM traces
		WHERE project_id = $1
	`
	args := []interface{}{filter.ProjectID}
	argPos := 2

	if !filter.Since.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argPos)
		args = append(args, filter.Since)
		argPos++
	}

	if !filter.Until.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argPos)
		args = append(args, filter.Until)
		argPos++
	}

	if len(filter.Level) > 0 {
		query += fmt.Sprintf(" AND level = ANY($%d)", argPos)
		args = append(args, filter.Level)
		argPos++
	}

	if filter.ServiceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argPos)
		args = append(args, filter.ServiceName)
		argPos++
	}

	if filter.Environment != "" {
		query += fmt.Sprintf(" AND environment = $%d", argPos)
		args = append(args, filter.Environment)
		argPos++
	}

	if filter.SearchText != "" {
		query += fmt.Sprintf(" AND message ILIKE $%d", argPos)
		args = append(args, "%"+filter.SearchText+"%")
		argPos++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %w", err)
	}
	defer rows.Close()

	var traces []*domain.Trace
	for rows.Next() {
		trace := &domain.Trace{}
		var contextJSON []byte

		err := rows.Scan(
			&trace.ID,
			&trace.ProjectID,
			&trace.Level,
			&trace.Message,
			&trace.Timestamp,
			&trace.Source,
			&trace.StackTrace,
			&contextJSON,
			&trace.ServiceName,
			&trace.Environment,
			&trace.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace: %w", err)
		}

		if len(contextJSON) > 0 {
			json.Unmarshal(contextJSON, &trace.Context)
		}

		traces = append(traces, trace)
	}

	return traces, nil
}

func (r *repository) Count(ctx context.Context, filter domain.TraceFilter) (int, error) {
	query := `SELECT COUNT(*) FROM traces WHERE project_id = $1`
	args := []interface{}{filter.ProjectID}

	if !filter.Since.IsZero() {
		query += " AND timestamp >= $2"
		args = append(args, filter.Since)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count traces: %w", err)
	}

	return count, nil
}

func (r *repository) DeleteOlderThan(ctx context.Context, projectID string, before time.Time) error {
	query := `DELETE FROM traces WHERE project_id = $1 AND timestamp < $2`

	_, err := r.db.ExecContext(ctx, query, projectID, before)
	if err != nil {
		return fmt.Errorf("failed to delete old traces: %w", err)
	}

	return nil
}
