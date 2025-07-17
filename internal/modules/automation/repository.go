package automation

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

type automationRepository struct {
	db DBTX
	sq sq.StatementBuilderType
}

func NewAutomationRepository(conn DBTX) AutomationRepository {
	return &automationRepository{
		db: conn,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// Automation CRUD
func (r *automationRepository) CreateAutomation(ctx context.Context, automation *Automation) error {
	query, args, err := r.sq.Insert("automations").
		Columns("id", "project_id", "name", "description", "config_json").
		Values(automation.ID, automation.ProjectID, automation.Name, automation.Description, automation.ConfigJSON).
		Suffix("RETURNING id, project_id, name, description, config_json, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var createdAt, updatedAt pgtype.Timestamp
	var description, configJSON pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&automation.ID, &automation.ProjectID, &automation.Name, &description, &configJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create automation: %w", err)
	}

	if description.Valid {
		automation.Description = description.String
	}
	if configJSON.Valid {
		automation.ConfigJSON = configJSON.String
	}
	automation.CreatedAt = createdAt.Time
	automation.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) GetAutomationByID(ctx context.Context, id string) (*Automation, error) {
	query, args, err := r.sq.Select("id", "project_id", "name", "description", "config_json", "created_at", "updated_at").
		From("automations").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var automation Automation
	var createdAt, updatedAt pgtype.Timestamp
	var description, configJSON pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&automation.ID, &automation.ProjectID, &automation.Name, &description, &configJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("automation not found")
		}
		return nil, fmt.Errorf("failed to get automation: %w", err)
	}

	if description.Valid {
		automation.Description = description.String
	}
	if configJSON.Valid {
		automation.ConfigJSON = configJSON.String
	}
	automation.CreatedAt = createdAt.Time
	automation.UpdatedAt = updatedAt.Time
	return &automation, nil
}

func (r *automationRepository) GetAutomationsByProjectID(ctx context.Context, projectID string) ([]*Automation, error) {
	query, args, err := r.sq.Select("id", "project_id", "name", "description", "config_json", "created_at", "updated_at").
		From("automations").
		Where(sq.Eq{"project_id": projectID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query automations: %w", err)
	}
	defer rows.Close()

	var automations []*Automation
	for rows.Next() {
		var automation Automation
		var createdAt, updatedAt pgtype.Timestamp
		var description, configJSON pgtype.Text
		err := rows.Scan(&automation.ID, &automation.ProjectID, &automation.Name, &description, &configJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan automation: %w", err)
		}
		if description.Valid {
			automation.Description = description.String
		}
		if configJSON.Valid {
			automation.ConfigJSON = configJSON.String
		}
		automation.CreatedAt = createdAt.Time
		automation.UpdatedAt = updatedAt.Time
		automations = append(automations, &automation)
	}

	return automations, nil
}

func (r *automationRepository) UpdateAutomation(ctx context.Context, automation *Automation) error {
	query, args, err := r.sq.Update("automations").
		Set("name", automation.Name).
		Set("description", automation.Description).
		Set("config_json", automation.ConfigJSON).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": automation.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update automation: %w", err)
	}

	automation.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) DeleteAutomation(ctx context.Context, id string) error {
	query, args, err := r.sq.Delete("automations").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete automation: %w", err)
	}

	return nil
}

// Step CRUD
func (r *automationRepository) CreateStep(ctx context.Context, step *AutomationStep) error {
	query, args, err := r.sq.Insert("automation_steps").
		Columns("id", "automation_id", "name", "step_order").
		Values(step.ID, step.AutomationID, step.Name, step.StepOrder).
		Suffix("RETURNING id, automation_id, name, step_order, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var createdAt, updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&step.ID, &step.AutomationID, &step.Name, &step.StepOrder, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create step: %w", err)
	}

	step.CreatedAt = createdAt.Time
	step.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) GetStepsByAutomationID(ctx context.Context, automationID string) ([]*AutomationStep, error) {
	query, args, err := r.sq.Select("id", "automation_id", "name", "step_order", "created_at", "updated_at").
		From("automation_steps").
		Where(sq.Eq{"automation_id": automationID}).
		OrderBy("step_order ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query steps: %w", err)
	}
	defer rows.Close()

	var steps []*AutomationStep
	for rows.Next() {
		var step AutomationStep
		var createdAt, updatedAt pgtype.Timestamp
		err := rows.Scan(&step.ID, &step.AutomationID, &step.Name, &step.StepOrder, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step: %w", err)
		}
		step.CreatedAt = createdAt.Time
		step.UpdatedAt = updatedAt.Time
		steps = append(steps, &step)
	}

	return steps, nil
}

func (r *automationRepository) UpdateStep(ctx context.Context, step *AutomationStep) error {
	query, args, err := r.sq.Update("automation_steps").
		Set("name", step.Name).
		Set("step_order", step.StepOrder).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": step.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update step: %w", err)
	}

	step.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) DeleteStep(ctx context.Context, id string) error {
	query, args, err := r.sq.Delete("automation_steps").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete step: %w", err)
	}

	return nil
}

// Action CRUD
func (r *automationRepository) CreateAction(ctx context.Context, action *AutomationAction) error {
	query, args, err := r.sq.Insert("automation_actions").
		Columns("id", "step_id", "action_type", "action_config_json", "action_order").
		Values(action.ID, action.StepID, action.ActionType, action.ActionConfigJSON, action.ActionOrder).
		Suffix("RETURNING id, step_id, action_type, action_config_json, action_order, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var createdAt, updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&action.ID, &action.StepID, &action.ActionType, &action.ActionConfigJSON, &action.ActionOrder, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	action.CreatedAt = createdAt.Time
	action.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) GetActionsByStepID(ctx context.Context, stepID string) ([]*AutomationAction, error) {
	query, args, err := r.sq.Select("id", "step_id", "action_type", "action_config_json", "action_order", "created_at", "updated_at").
		From("automation_actions").
		Where(sq.Eq{"step_id": stepID}).
		OrderBy("action_order ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query actions: %w", err)
	}
	defer rows.Close()

	var actions []*AutomationAction
	for rows.Next() {
		var action AutomationAction
		var createdAt, updatedAt pgtype.Timestamp
		err := rows.Scan(&action.ID, &action.StepID, &action.ActionType, &action.ActionConfigJSON, &action.ActionOrder, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		action.CreatedAt = createdAt.Time
		action.UpdatedAt = updatedAt.Time
		actions = append(actions, &action)
	}

	return actions, nil
}

func (r *automationRepository) UpdateAction(ctx context.Context, action *AutomationAction) error {
	query, args, err := r.sq.Update("automation_actions").
		Set("action_type", action.ActionType).
		Set("action_config_json", action.ActionConfigJSON).
		Set("action_order", action.ActionOrder).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": action.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	action.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) DeleteAction(ctx context.Context, id string) error {
	query, args, err := r.sq.Delete("automation_actions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete action: %w", err)
	}

	return nil
}

// Run CRUD
func (r *automationRepository) CreateRun(ctx context.Context, run *AutomationRun) error {
	query, args, err := r.sq.Insert("automation_runs").
		Columns("id", "automation_id", "status", "logs_json", "output_files_json", "error_message").
		Values(run.ID, run.AutomationID, run.Status, run.LogsJSON, run.OutputFilesJSON, run.ErrorMessage).
		Suffix("RETURNING id, automation_id, status, start_time, end_time, logs_json, output_files_json, error_message, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var createdAt, updatedAt, startTime, endTime pgtype.Timestamp
	var logsJSON, outputFilesJSON, errorMessage pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&run.ID, &run.AutomationID, &run.Status, &startTime, &endTime, &logsJSON, &outputFilesJSON, &errorMessage, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	if startTime.Valid {
		run.StartTime = &startTime.Time
	}
	if endTime.Valid {
		run.EndTime = &endTime.Time
	}
	if logsJSON.Valid {
		run.LogsJSON = logsJSON.String
	}
	if outputFilesJSON.Valid {
		run.OutputFilesJSON = outputFilesJSON.String
	}
	if errorMessage.Valid {
		run.ErrorMessage = errorMessage.String
	}
	run.CreatedAt = createdAt.Time
	run.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) GetRunByID(ctx context.Context, id string) (*AutomationRun, error) {
	query, args, err := r.sq.Select("id", "automation_id", "status", "start_time", "end_time", "logs_json", "output_files_json", "error_message", "created_at", "updated_at").
		From("automation_runs").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var run AutomationRun
	var createdAt, updatedAt, startTime, endTime pgtype.Timestamp
	var logsJSON, outputFilesJSON, errorMessage pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&run.ID, &run.AutomationID, &run.Status, &startTime, &endTime, &logsJSON, &outputFilesJSON, &errorMessage, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("run not found")
		}
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	if startTime.Valid {
		run.StartTime = &startTime.Time
	}
	if endTime.Valid {
		run.EndTime = &endTime.Time
	}
	if logsJSON.Valid {
		run.LogsJSON = logsJSON.String
	}
	if outputFilesJSON.Valid {
		run.OutputFilesJSON = outputFilesJSON.String
	}
	if errorMessage.Valid {
		run.ErrorMessage = errorMessage.String
	}
	run.CreatedAt = createdAt.Time
	run.UpdatedAt = updatedAt.Time
	return &run, nil
}

func (r *automationRepository) GetRunsByAutomationID(ctx context.Context, automationID string) ([]*AutomationRun, error) {
	query, args, err := r.sq.Select("id", "automation_id", "status", "start_time", "end_time", "logs_json", "output_files_json", "error_message", "created_at", "updated_at").
		From("automation_runs").
		Where(sq.Eq{"automation_id": automationID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	var runs []*AutomationRun
	for rows.Next() {
		var run AutomationRun
		var createdAt, updatedAt, startTime, endTime pgtype.Timestamp
		var logsJSON, outputFilesJSON, errorMessage pgtype.Text
		err := rows.Scan(&run.ID, &run.AutomationID, &run.Status, &startTime, &endTime, &logsJSON, &outputFilesJSON, &errorMessage, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}
		if startTime.Valid {
			run.StartTime = &startTime.Time
		}
		if endTime.Valid {
			run.EndTime = &endTime.Time
		}
		if logsJSON.Valid {
			run.LogsJSON = logsJSON.String
		}
		if outputFilesJSON.Valid {
			run.OutputFilesJSON = outputFilesJSON.String
		}
		if errorMessage.Valid {
			run.ErrorMessage = errorMessage.String
		}
		run.CreatedAt = createdAt.Time
		run.UpdatedAt = updatedAt.Time
		runs = append(runs, &run)
	}

	return runs, nil
}

func (r *automationRepository) UpdateRun(ctx context.Context, run *AutomationRun) error {
	query, args, err := r.sq.Update("automation_runs").
		Set("status", run.Status).
		Set("start_time", run.StartTime).
		Set("end_time", run.EndTime).
		Set("logs_json", run.LogsJSON).
		Set("output_files_json", run.OutputFilesJSON).
		Set("error_message", run.ErrorMessage).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": run.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update run: %w", err)
	}

	run.UpdatedAt = updatedAt.Time
	return nil
}