package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type projectRepository struct {
	db DBTX
	sq sq.StatementBuilderType
}

func NewProjectRepository(conn DBTX) ProjectRepository {
	return &projectRepository{
		db: conn,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *projectRepository) Create(ctx context.Context, project *Project) error {
	query, args, err := r.sq.Insert("projects").
		Columns("id", "organization_id", "name", "description").
		Values(project.ID, project.OrganizationID, project.Name, project.Description).
		Suffix("RETURNING id, organization_id, name, description, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var createdAt, updatedAt pgtype.Timestamp
	var description pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&project.ID, &project.OrganizationID, &project.Name, &description, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	if description.Valid {
		project.Description = description.String
	}
	project.CreatedAt = createdAt.Time
	project.UpdatedAt = updatedAt.Time
	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, id string) (*Project, error) {
	query, args, err := r.sq.Select("id", "organization_id", "name", "description", "created_at", "updated_at").
		From("projects").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var project Project
	var createdAt, updatedAt pgtype.Timestamp
	var description pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&project.ID, &project.OrganizationID, &project.Name, &description, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if description.Valid {
		project.Description = description.String
	}
	project.CreatedAt = createdAt.Time
	project.UpdatedAt = updatedAt.Time
	return &project, nil
}

func (r *projectRepository) GetByOrganizationID(ctx context.Context, organizationID string) ([]*Project, error) {
	query, args, err := r.sq.Select("id", "organization_id", "name", "description", "created_at", "updated_at").
		From("projects").
		Where(sq.Eq{"organization_id": organizationID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var project Project
		var createdAt, updatedAt pgtype.Timestamp
		var description pgtype.Text
		err := rows.Scan(&project.ID, &project.OrganizationID, &project.Name, &description, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		if description.Valid {
			project.Description = description.String
		}
		project.CreatedAt = createdAt.Time
		project.UpdatedAt = updatedAt.Time
		projects = append(projects, &project)
	}

	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, project *Project) error {
	query, args, err := r.sq.Update("projects").
		Set("name", project.Name).
		Set("description", project.Description).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": project.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	project.UpdatedAt = updatedAt.Time
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	query, args, err := r.sq.Delete("projects").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}