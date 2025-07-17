package project

import (
	"context"
	"time"
)

// Project represents a project within an organization
type Project struct {
	ID             string
	OrganizationID string
	Name           string
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ProjectRepository defines the interface for project data operations
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	GetByOrganizationID(ctx context.Context, organizationID string) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
}

// ProjectService defines the interface for project business logic
type ProjectService interface {
	CreateProject(ctx context.Context, organizationID, name, description string) (*Project, error)
	GetProjectsByOrganization(ctx context.Context, organizationID string) ([]*Project, error)
	GetProjectByID(ctx context.Context, id string) (*Project, error)
	UpdateProject(ctx context.Context, project *Project) error
	DeleteProject(ctx context.Context, id string) error
}