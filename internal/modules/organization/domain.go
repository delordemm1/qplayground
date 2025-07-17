package organization

import (
	"context"
	"time"
)

// Organization represents a workspace that contains projects
type Organization struct {
	ID          string
	Name        string
	OwnerUserID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// OrganizationRepository defines the interface for organization data operations
type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id string) (*Organization, error)
	GetByOwnerUserID(ctx context.Context, ownerUserID string) ([]*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id string) error
}

// OrganizationService defines the interface for organization business logic
type OrganizationService interface {
	CreatePersonalOrganization(ctx context.Context, userID, userEmail string) (*Organization, error)
	GetUserOrganizations(ctx context.Context, userID string) ([]*Organization, error)
	GetOrganizationByID(ctx context.Context, id string) (*Organization, error)
}