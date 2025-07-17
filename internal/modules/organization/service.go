package organization

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/delordemm1/qplayground/internal/platform"
)

type organizationService struct {
	orgRepo OrganizationRepository
}

func NewOrganizationService(orgRepo OrganizationRepository) OrganizationService {
	return &organizationService{
		orgRepo: orgRepo,
	}
}

func (s *organizationService) CreatePersonalOrganization(ctx context.Context, userID, userEmail string) (*Organization, error) {
	org := &Organization{
		ID:          platform.UtilGenerateUUID(),
		Name:        fmt.Sprintf("Personal Organization - %s", userEmail),
		OwnerUserID: userID,
	}

	err := s.orgRepo.Create(ctx, org)
	if err != nil {
		slog.Error("Failed to create personal organization", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to create personal organization: %w", err)
	}

	slog.Info("Personal organization created", "orgID", org.ID, "userID", userID)
	return org, nil
}

func (s *organizationService) GetUserOrganizations(ctx context.Context, userID string) ([]*Organization, error) {
	organizations, err := s.orgRepo.GetByOwnerUserID(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user organizations", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	return organizations, nil
}

func (s *organizationService) GetOrganizationByID(ctx context.Context, id string) (*Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		slog.Error("Failed to get organization by ID", "error", err, "orgID", id)
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}