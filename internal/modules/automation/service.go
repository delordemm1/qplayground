package automation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/delordemm1/qplayground/internal/platform"
)

type automationService struct {
	automationRepo AutomationRepository
}

func NewAutomationService(automationRepo AutomationRepository) AutomationService {
	return &automationService{
		automationRepo: automationRepo,
	}
}

// Automation management
func (s *automationService) CreateAutomation(ctx context.Context, projectID, name, description, configJSON string) (*Automation, error) {
	automation := &Automation{
		ID:          platform.UtilGenerateUUID(),
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		ConfigJSON:  configJSON,
	}

	err := s.automationRepo.CreateAutomation(ctx, automation)
	if err != nil {
		slog.Error("Failed to create automation", "error", err, "projectID", projectID, "name", name)
		return nil, fmt.Errorf("failed to create automation: %w", err)
	}

	slog.Info("Automation created", "automationID", automation.ID, "projectID", projectID, "name", name)
	return automation, nil
}

func (s *automationService) GetAutomationsByProject(ctx context.Context, projectID string) ([]*Automation, error) {
	automations, err := s.automationRepo.GetAutomationsByProjectID(ctx, projectID)
	if err != nil {
		slog.Error("Failed to get automations by project", "error", err, "projectID", projectID)
		return nil, fmt.Errorf("failed to get automations: %w", err)
	}

	return automations, nil
}

func (s *automationService) GetAutomationByID(ctx context.Context, id string) (*Automation, error) {
	automation, err := s.automationRepo.GetAutomationByID(ctx, id)
	if err != nil {
		slog.Error("Failed to get automation by ID", "error", err, "automationID", id)
		return nil, fmt.Errorf("failed to get automation: %w", err)
	}

	return automation, nil
}

func (s *automationService) UpdateAutomation(ctx context.Context, automation *Automation) error {
	err := s.automationRepo.UpdateAutomation(ctx, automation)
	if err != nil {
		slog.Error("Failed to update automation", "error", err, "automationID", automation.ID)
		return fmt.Errorf("failed to update automation: %w", err)
	}

	slog.Info("Automation updated", "automationID", automation.ID, "name", automation.Name)
	return nil
}

func (s *automationService) DeleteAutomation(ctx context.Context, id string) error {
	err := s.automationRepo.DeleteAutomation(ctx, id)
	if err != nil {
		slog.Error("Failed to delete automation", "error", err, "automationID", id)
		return fmt.Errorf("failed to delete automation: %w", err)
	}

	slog.Info("Automation deleted", "automationID", id)
	return nil
}

// Step management
func (s *automationService) CreateStep(ctx context.Context, automationID, name string, stepOrder int) (*AutomationStep, error) {
	step := &AutomationStep{
		ID:           platform.UtilGenerateUUID(),
		AutomationID: automationID,
		Name:         name,
		StepOrder:    stepOrder,
	}

	err := s.automationRepo.CreateStep(ctx, step)
	if err != nil {
		slog.Error("Failed to create step", "error", err, "automationID", automationID, "name", name)
		return nil, fmt.Errorf("failed to create step: %w", err)
	}

	slog.Info("Step created", "stepID", step.ID, "automationID", automationID, "name", name)
	return step, nil
}

func (s *automationService) GetStepsByAutomation(ctx context.Context, automationID string) ([]*AutomationStep, error) {
	steps, err := s.automationRepo.GetStepsByAutomationID(ctx, automationID)
	if err != nil {
		slog.Error("Failed to get steps by automation", "error", err, "automationID", automationID)
		return nil, fmt.Errorf("failed to get steps: %w", err)
	}

	return steps, nil
}

func (s *automationService) UpdateStep(ctx context.Context, step *AutomationStep) error {
	err := s.automationRepo.UpdateStep(ctx, step)
	if err != nil {
		slog.Error("Failed to update step", "error", err, "stepID", step.ID)
		return fmt.Errorf("failed to update step: %w", err)
	}

	slog.Info("Step updated", "stepID", step.ID, "name", step.Name)
	return nil
}

func (s *automationService) DeleteStep(ctx context.Context, id string) error {
	err := s.automationRepo.DeleteStep(ctx, id)
	if err != nil {
		slog.Error("Failed to delete step", "error", err, "stepID", id)
		return fmt.Errorf("failed to delete step: %w", err)
	}

	slog.Info("Step deleted", "stepID", id)
	return nil
}

// Action management
func (s *automationService) CreateAction(ctx context.Context, stepID, actionType, actionConfigJSON string, actionOrder int) (*AutomationAction, error) {
	action := &AutomationAction{
		ID:               platform.UtilGenerateUUID(),
		StepID:           stepID,
		ActionType:       actionType,
		ActionConfigJSON: actionConfigJSON,
		ActionOrder:      actionOrder,
	}

	err := s.automationRepo.CreateAction(ctx, action)
	if err != nil {
		slog.Error("Failed to create action", "error", err, "stepID", stepID, "actionType", actionType)
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	slog.Info("Action created", "actionID", action.ID, "stepID", stepID, "actionType", actionType)
	return action, nil
}

func (s *automationService) GetActionsByStep(ctx context.Context, stepID string) ([]*AutomationAction, error) {
	actions, err := s.automationRepo.GetActionsByStepID(ctx, stepID)
	if err != nil {
		slog.Error("Failed to get actions by step", "error", err, "stepID", stepID)
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	return actions, nil
}

func (s *automationService) UpdateAction(ctx context.Context, action *AutomationAction) error {
	err := s.automationRepo.UpdateAction(ctx, action)
	if err != nil {
		slog.Error("Failed to update action", "error", err, "actionID", action.ID)
		return fmt.Errorf("failed to update action: %w", err)
	}

	slog.Info("Action updated", "actionID", action.ID, "actionType", action.ActionType)
	return nil
}

func (s *automationService) DeleteAction(ctx context.Context, id string) error {
	err := s.automationRepo.DeleteAction(ctx, id)
	if err != nil {
		slog.Error("Failed to delete action", "error", err, "actionID", id)
		return fmt.Errorf("failed to delete action: %w", err)
	}

	slog.Info("Action deleted", "actionID", id)
	return nil
}

// Helper methods for order management
func (s *automationService) GetMaxStepOrder(ctx context.Context, automationID string) (int, error) {
	return s.automationRepo.GetMaxStepOrder(ctx, automationID)
}

func (s *automationService) GetMaxActionOrder(ctx context.Context, stepID string) (int, error) {
	return s.automationRepo.GetMaxActionOrder(ctx, stepID)
}

// Run management
func (s *automationService) TriggerRun(ctx context.Context, automationID string) (*AutomationRun, error) {
	run := &AutomationRun{
		ID:              platform.UtilGenerateUUID(),
		AutomationID:    automationID,
		Status:          "pending",
		LogsJSON:        "[]",
		OutputFilesJSON: "[]",
	}

	err := s.automationRepo.CreateRun(ctx, run)
	if err != nil {
		slog.Error("Failed to create run", "error", err, "automationID", automationID)
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	// TODO: Trigger actual automation execution in background
	// For now, just create the run record
	slog.Info("Run triggered", "runID", run.ID, "automationID", automationID)
	return run, nil
}

func (s *automationService) GetRunsByAutomation(ctx context.Context, automationID string) ([]*AutomationRun, error) {
	runs, err := s.automationRepo.GetRunsByAutomationID(ctx, automationID)
	if err != nil {
		slog.Error("Failed to get runs by automation", "error", err, "automationID", automationID)
		return nil, fmt.Errorf("failed to get runs: %w", err)
	}

	return runs, nil
}

func (s *automationService) GetRunByID(ctx context.Context, id string) (*AutomationRun, error) {
	run, err := s.automationRepo.GetRunByID(ctx, id)
	if err != nil {
		slog.Error("Failed to get run by ID", "error", err, "runID", id)
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	return run, nil
}
