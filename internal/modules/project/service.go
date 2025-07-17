package project

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/delordemm1/qplayground/internal/platform"
)

type projectService struct {
	projectRepo ProjectRepository
}

func NewProjectService(projectRepo ProjectRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
	}
}

func (s *projectService) CreateProject(ctx context.Context, organizationID, name, description string) (*Project, error) {
	project := &Project{
		ID:             platform.UtilGenerateUUID(),
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
	}

	err := s.projectRepo.Create(ctx, project)
	if err != nil {
		slog.Error("Failed to create project", "error", err, "organizationID", organizationID, "name", name)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	slog.Info("Project created", "projectID", project.ID, "organizationID", organizationID, "name", name)
	return project, nil
}

func (s *projectService) GetProjectsByOrganization(ctx context.Context, organizationID string) ([]*Project, error) {
	projects, err := s.projectRepo.GetByOrganizationID(ctx, organizationID)
	if err != nil {
		slog.Error("Failed to get projects by organization", "error", err, "organizationID", organizationID)
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

func (s *projectService) GetProjectByID(ctx context.Context, id string) (*Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		slog.Error("Failed to get project by ID", "error", err, "projectID", id)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

func (s *projectService) UpdateProject(ctx context.Context, project *Project) error {
	err := s.projectRepo.Update(ctx, project)
	if err != nil {
		slog.Error("Failed to update project", "error", err, "projectID", project.ID)
		return fmt.Errorf("failed to update project: %w", err)
	}

	slog.Info("Project updated", "projectID", project.ID, "name", project.Name)
	return nil
}

func (s *projectService) DeleteProject(ctx context.Context, id string) error {
	err := s.projectRepo.Delete(ctx, id)
	if err != nil {
		slog.Error("Failed to delete project", "error", err, "projectID", id)
		return fmt.Errorf("failed to delete project: %w", err)
	}

	slog.Info("Project deleted", "projectID", id)
	return nil
}