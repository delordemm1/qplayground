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
		Columns("id", "automation_id", "name", "step_order", "config_json").
		Values(step.ID, step.AutomationID, step.Name, step.StepOrder, step.ConfigJSON).
		Suffix("RETURNING id, automation_id, name, step_order, config_json, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var createdAt, updatedAt pgtype.Timestamp
	var configJSON pgtype.Text
	var configJSON pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&step.ID, &step.AutomationID, &step.Name, &step.StepOrder, &configJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create step: %w", err)
	}

	if configJSON.Valid {
		step.ConfigJSON = configJSON.String
	}
	if configJSON.Valid {
		step.ConfigJSON = configJSON.String
	}
	step.CreatedAt = createdAt.Time
	step.UpdatedAt = updatedAt.Time
	return nil
}

func (r *automationRepository) GetStepsByAutomationID(ctx context.Context, automationID string) ([]*AutomationStep, error) {
	query, args, err := r.sq.Select("id", "automation_id", "name", "step_order", "config_json", "created_at", "updated_at").
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
		var configJSON pgtype.Text
		err := rows.Scan(&step.ID, &step.AutomationID, &step.Name, &step.StepOrder, &configJSON, &createdAt, &updatedAt)
		err := rows.Scan(&step.ID, &step.AutomationID, &step.Name, &step.StepOrder, &configJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step: %w", err)
		}
		if configJSON.Valid {
			step.ConfigJSON = configJSON.String
		}
		if configJSON.Valid {
			step.ConfigJSON = configJSON.String
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
		Set("config_json", step.ConfigJSON).
		Set("config_json", step.ConfigJSON).
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

func (r *automationRepository) ShiftActionOrdersAfterDelete(ctx context.Context, stepID string, deletedOrder int) error {
	query, args, err := r.sq.Update("automation_actions").
		Set("action_order", sq.Expr("action_order - 1")).
		Set("updated_at", time.Now()).
		Where(sq.And{
			sq.Eq{"step_id": stepID},
			sq.Gt{"action_order": deletedOrder},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to shift action orders after delete: %w", err)
	}

	return nil
}

// Order management methods
func (r *automationRepository) GetStepByID(ctx context.Context, id string) (*AutomationStep, error) {
				"step_name":         step.Name,
				"action_type":       "step_skip",
				"message":           fmt.Sprintf("Step skipped: %s", skipReason),
				"status":            "skipped",
				"loop_index":        loopIndex,
				"local_loop_index":  0,
				"duration_ms":       0,
			}
			runContext.Logs = append(runContext.Logs, logEntry)
			continue // Skip to next step
		}

		From("automation_steps").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var step AutomationStep
	var createdAt, updatedAt pgtype.Timestamp
	var configJSON pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&step.ID, &step.AutomationID, &step.Name, &step.StepOrder, &configJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("step not found")
		}
		return nil, fmt.Errorf("failed to get step: %w", err)
	}

	if configJSON.Valid {
		step.ConfigJSON = configJSON.String
	}
	step.CreatedAt = createdAt.Time
	step.UpdatedAt = updatedAt.Time
	return &step, nil
}

func (r *automationRepository) GetActionByID(ctx context.Context, id string) (*AutomationAction, error) {
	query, args, err := r.sq.Select("id", "step_id", "action_type", "action_config_json", "action_order", "created_at", "updated_at").
		From("automation_actions").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var action AutomationAction
	var createdAt, updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&action.ID, &action.StepID, &action.ActionType, &action.ActionConfigJSON, &action.ActionOrder, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("action not found")
		}
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	action.CreatedAt = createdAt.Time
	action.UpdatedAt = updatedAt.Time
	return &action, nil
}

func (r *automationRepository) GetMaxStepOrder(ctx context.Context, automationID string) (int, error) {
	query, args, err := r.sq.Select("COALESCE(MAX(step_order), 0)").
		From("automation_steps").
		Where(sq.Eq{"automation_id": automationID}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var maxOrder int
	err = r.db.QueryRow(ctx, query, args...).Scan(&maxOrder)
	if err != nil {
		return 0, fmt.Errorf("failed to get max step order: %w", err)
	}

	return maxOrder, nil
}

func (r *automationRepository) GetMaxActionOrder(ctx context.Context, stepID string) (int, error) {
	query, args, err := r.sq.Select("COALESCE(MAX(action_order), 0)").
		From("automation_actions").
		Where(sq.Eq{"step_id": stepID}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var maxOrder int
	err = r.db.QueryRow(ctx, query, args...).Scan(&maxOrder)
	if err != nil {
		return 0, fmt.Errorf("failed to get max action order: %w", err)
	}

	return maxOrder, nil
}

// GetStepByAutomationIDAndOrder retrieves a step by automation ID and order
func (r *automationRepository) GetStepByAutomationIDAndOrder(ctx context.Context, automationID string, order int) (*AutomationStep, error) {
	query, args, err := r.sq.Select("id", "automation_id", "name", "step_order", "config_json", "created_at", "updated_at").
		From("automation_steps").
		Where(sq.And{
			sq.Eq{"automation_id": automationID},
			sq.Eq{"step_order": order},
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var step AutomationStep
	var createdAt, updatedAt pgtype.Timestamp
	var configJSON pgtype.Text
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&step.ID, &step.AutomationID, &step.Name, &step.StepOrder, &configJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("step not found")
		}
		return nil, fmt.Errorf("failed to get step: %w", err)
	}

	if configJSON.Valid {
		step.ConfigJSON = configJSON.String
	}
	step.CreatedAt = createdAt.Time
	step.UpdatedAt = updatedAt.Time
	return &step, nil
}

// GetActionByStepIDAndOrder retrieves an action by step ID and order
func (r *automationRepository) GetActionByStepIDAndOrder(ctx context.Context, stepID string, order int) (*AutomationAction, error) {
	query, args, err := r.sq.Select("id", "step_id", "action_type", "action_config_json", "action_order", "created_at", "updated_at").
		From("automation_actions").
		Where(sq.And{
			sq.Eq{"step_id": stepID},
			sq.Eq{"action_order": order},
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var action AutomationAction
	var createdAt, updatedAt pgtype.Timestamp
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&action.ID, &action.StepID, &action.ActionType, &action.ActionConfigJSON, &action.ActionOrder, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("action not found")
		}
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	action.CreatedAt = createdAt.Time
	action.UpdatedAt = updatedAt.Time
	return &action, nil
}

func (r *automationRepository) ShiftStepOrders(ctx context.Context, automationID string, startOrder, endOrder int, increment bool) error {
	var query string
	var args []interface{}
	var err error

	if increment {
		// Shift orders up (increase by 1)
		query, args, err = r.sq.Update("automation_steps").
			Set("step_order", sq.Expr("step_order + 1")).
			Set("updated_at", time.Now()).
			Where(sq.And{
				sq.Eq{"automation_id": automationID},
				sq.GtOrEq{"step_order": startOrder},
				sq.LtOrEq{"step_order": endOrder},
			}).
			ToSql()
	} else {
		// Shift orders down (decrease by 1)
		query, args, err = r.sq.Update("automation_steps").
			Set("step_order", sq.Expr("step_order - 1")).
			Set("updated_at", time.Now()).
			Where(sq.And{
				sq.Eq{"automation_id": automationID},
				sq.GtOrEq{"step_order": startOrder},
				sq.LtOrEq{"step_order": endOrder},
			}).
			ToSql()
	}

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to shift step orders: %w", err)
	}

	return nil
}

func (r *automationRepository) ShiftActionOrders(ctx context.Context, stepID string, startOrder, endOrder int, increment bool) error {
	var query string
	var args []interface{}
	var err error

	if increment {
		// Shift orders up (increase by 1)
		query, args, err = r.sq.Update("automation_actions").
			Set("action_order", sq.Expr("action_order + 1")).
			Set("updated_at", time.Now()).
			Where(sq.And{
				sq.Eq{"step_id": stepID},
				sq.GtOrEq{"action_order": startOrder},
				sq.LtOrEq{"action_order": endOrder},
			}).
			ToSql()
	} else {
		// Shift orders down (decrease by 1)
		query, args, err = r.sq.Update("automation_actions").
			Set("action_order", sq.Expr("action_order - 1")).
			Set("updated_at", time.Now()).
			Where(sq.And{
				sq.Eq{"step_id": stepID},
				sq.GtOrEq{"action_order": startOrder},
				sq.LtOrEq{"action_order": endOrder},
			}).
			ToSql()
	}

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to shift action orders: %w", err)
	}

	return nil
}