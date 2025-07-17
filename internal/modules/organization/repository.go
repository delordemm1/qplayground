package organization

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

type organizationRepository struct {
	db DBTX
	sq sq.StatementBuilderType
}

func NewOrganizationRepository(conn DBTX) OrganizationRepository {
	return &organizationRepository{
		db: conn,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *organizationRepository) Create(ctx context.Context, org *Organization) error {
	query, args, err := r.sq.Insert("organizations").
		Columns("id", "name", "owner_user_id").
		Values(org.ID, org.Name, org.OwnerUserID).
		Suffix("RETURNING id, name, owner_user_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var createdAt, updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&org.ID, &org.Name, &org.OwnerUserID, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	org.CreatedAt = createdAt.Time
	org.UpdatedAt = updatedAt.Time
	return nil
}

func (r *organizationRepository) GetByID(ctx context.Context, id string) (*Organization, error) {
	query, args, err := r.sq.Select("id", "name", "owner_user_id", "created_at", "updated_at").
		From("organizations").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var org Organization
	var createdAt, updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&org.ID, &org.Name, &org.OwnerUserID, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	org.CreatedAt = createdAt.Time
	org.UpdatedAt = updatedAt.Time
	return &org, nil
}

func (r *organizationRepository) GetByOwnerUserID(ctx context.Context, ownerUserID string) ([]*Organization, error) {
	query, args, err := r.sq.Select("id", "name", "owner_user_id", "created_at", "updated_at").
		From("organizations").
		Where(sq.Eq{"owner_user_id": ownerUserID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*Organization
	for rows.Next() {
		var org Organization
		var createdAt, updatedAt pgtype.Timestamp
		err := rows.Scan(&org.ID, &org.Name, &org.OwnerUserID, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		org.CreatedAt = createdAt.Time
		org.UpdatedAt = updatedAt.Time
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

func (r *organizationRepository) Update(ctx context.Context, org *Organization) error {
	query, args, err := r.sq.Update("organizations").
		Set("name", org.Name).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": org.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	org.UpdatedAt = updatedAt.Time
	return nil
}

func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	query, args, err := r.sq.Delete("organizations").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}