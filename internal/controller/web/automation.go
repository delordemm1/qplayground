package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/delordemm1/qplayground/internal/modules/auth"
	"github.com/delordemm1/qplayground/internal/modules/automation"
	"github.com/delordemm1/qplayground/internal/modules/project"
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/go-playground/validator/v10"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	inertia "github.com/romsar/gonertia/v2"
)

func NewAutomationRouter(automationHandler *AutomationHandler) chi.Router {
	r := chi.NewRouter()

	r.Get("/", automationHandler.ListAutomations)
	r.Post("/", automationHandler.CreateAutomation)
	r.Get("/{id}", automationHandler.GetAutomation)
	r.Put("/{id}", automationHandler.UpdateAutomation)
	r.Delete("/{id}", automationHandler.DeleteAutomation)

	// Step management
	r.Post("/{id}/steps", automationHandler.CreateStep)
	r.Put("/{id}/steps/{stepId}", automationHandler.UpdateStep)
	r.Delete("/{id}/steps/{stepId}", automationHandler.DeleteStep)

	// Run management
	r.Post("/{id}/runs", automationHandler.TriggerRun)
	r.Get("/{id}/runs", automationHandler.ListRuns)
	r.Get("/{id}/runs/{runId}", automationHandler.GetRun)
	r.Post("/{id}/runs/{runId}/cancel", automationHandler.CancelRun)

	// Export automation config
	r.Get("/{id}/export", automationHandler.ExportAutomationConfig)

	// SSE endpoint for run progress
	r.Get("/{id}/runs/{runId}/events", automationHandler.GetRunEvents)

	return r
}

func NewAutomationHandler(inertia *inertia.Inertia, sessionManager *scs.SessionManager, automationService automation.AutomationService, projectService project.ProjectService, scheduler *automation.Scheduler, sseManager *automation.SSEManager) *AutomationHandler {
	return &AutomationHandler{
		inertia:           inertia,
		sessionManager:    sessionManager,
		automationService: automationService,
		projectService:    projectService,
		scheduler:         scheduler,
		sseManager:        sseManager,
		runContexts:       make(map[string]context.CancelFunc),
		// stepService:       automationService, // AutomationService also handles steps
		// actionService:     automationService, // AutomationService also handles actions
		// runService:        automationService, // AutomationService also handles runs
	}
}

type AutomationHandler struct {
	inertia           *inertia.Inertia
	sessionManager    *scs.SessionManager
	automationService automation.AutomationService
	// stepService       automation.AutomationService
	// actionService     automation.AutomationService
	projectService project.ProjectService
	scheduler      *automation.Scheduler
	sseManager     *automation.SSEManager
	runContexts    map[string]context.CancelFunc
	mu             sync.Mutex
}

type CreateAutomationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	ConfigJSON  string `json:"config_json"`
}

func (h *AutomationHandler) ListAutomations(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	projectID := chi.URLParam(r, "projectId")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	automations, err := h.automationService.GetAutomationsByProject(r.Context(), projectID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	err = h.inertia.Render(w, r, "automations/index", inertia.Props{
		"automations": automations,
		"project":     project,
		"user":        user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *AutomationHandler) CreateAutomation(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	var req CreateAutomationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": ConvertValidationErrorsToInertia(validationErrors),
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	// Default config if empty
	configJSON := req.ConfigJSON
	if configJSON == "" {
		configJSON = "{}"
	}

	automation, err := h.automationService.CreateAutomation(r.Context(), projectID, req.Name, req.Description, configJSON)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to create automation")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create automation"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Automation created successfully")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Automation created successfully",
		"automation": automation,
	})
}

func (h *AutomationHandler) GetAutomation(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	// Get steps and actions
	steps, err := h.automationService.GetStepsByAutomation(r.Context(), automationID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Get actions for each step
	stepsWithActions := make([]map[string]interface{}, len(steps))
	for i, step := range steps {
		actions, err := h.automationService.GetActionsByStep(r.Context(), step.ID)
		if err != nil {
			platform.UtilHandleServerErr(w, err)
			return
		}

		// Get max action order for this step
		maxActionOrder, err := h.automationService.GetMaxActionOrder(r.Context(), step.ID)
		if err != nil {
			platform.UtilHandleServerErr(w, err)
			return
		}

		stepsWithActions[i] = map[string]interface{}{
			"step":           step,
			"actions":        actions,
			"maxActionOrder": maxActionOrder,
		}
	}

	// Get max step order for the automation
	maxStepOrder, err := h.automationService.GetMaxStepOrder(r.Context(), automationID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Get recent runs for this automation (limit to 5)
	allRuns, err := h.automationService.GetRunsByAutomation(r.Context(), automationID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Limit to 5 most recent runs
	recentRuns := allRuns
	if len(allRuns) > 5 {
		recentRuns = allRuns[:5]
	}

	err = h.inertia.Render(w, r, "projects/[projectId]/automations/[automationId]", inertia.Props{
		"params":       map[string]string{"automationId": automationID, "projectId": projectID},
		"automation":   automation,
		"project":      project,
		"steps":        stepsWithActions,
		"maxStepOrder": maxStepOrder,
		"recentRuns":   recentRuns,
		"user":         user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *AutomationHandler) UpdateAutomation(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Automation not found"})
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	var req CreateAutomationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": ConvertValidationErrorsToInertia(validationErrors),
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	automation.Name = req.Name
	automation.Description = req.Description
	automation.ConfigJSON = req.ConfigJSON

	err = h.automationService.UpdateAutomation(r.Context(), automation)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to update automation")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update automation"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Automation updated successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Automation updated successfully",
		"automation": automation,
	})
}

func (h *AutomationHandler) DeleteAutomation(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Automation not found"})
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	err = h.automationService.DeleteAutomation(r.Context(), automationID)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to delete automation")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete automation"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Automation deleted successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Automation deleted successfully"})
}

type CreateStepRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=255"`
	StepOrder int    `json:"step_order" validate:"min=0"`
	ConfigJSON string `json:"config_json"`
}

func (h *AutomationHandler) CreateStep(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")

	// Verify project and automation ownership
	if err := h.verifyAutomationAccess(r.Context(), user, projectID, automationID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var req CreateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": ConvertValidationErrorsToInertia(validationErrors),
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	// Default config if empty
	configJSON := req.ConfigJSON
	if configJSON == "" {
		configJSON = "{}"
	}

	step, err := h.automationService.CreateStep(r.Context(), automationID, req.Name, req.StepOrder, configJSON)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to create step")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create step"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Step created successfully")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Step created successfully",
		"step":    step,
	})
}

func (h *AutomationHandler) UpdateStep(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")

	if err := h.verifyAutomationAccess(r.Context(), user, projectID, automationID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var req CreateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": ConvertValidationErrorsToInertia(validationErrors),
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	step := &automation.AutomationStep{
		ID:           stepID,
		AutomationID: automationID,
		Name:         req.Name,
		StepOrder:    req.StepOrder,
		ConfigJSON:   req.ConfigJSON,
	}

	err := h.automationService.UpdateStep(r.Context(), step)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to update step")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update step"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Step updated successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Step updated successfully",
		"step":    step,
	})
}

func (h *AutomationHandler) DeleteStep(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")

	if err := h.verifyAutomationAccess(r.Context(), user, projectID, automationID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	err := h.automationService.DeleteStep(r.Context(), stepID)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to delete step")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete step"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Step deleted successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Step deleted successfully"})
}

func (h *AutomationHandler) TriggerRun(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Automation not found"})
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	run, err := h.automationService.TriggerRun(r.Context(), automationID)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to trigger automation run")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to trigger run"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Automation run triggered successfully")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Run triggered successfully",
		"run":     run,
	})
}

func (h *AutomationHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	runs, err := h.automationService.GetRunsByAutomation(r.Context(), automationID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	err = h.inertia.Render(w, r, "projects/[projectId]/automations/[automationId]/runs", inertia.Props{
		"params":     map[string]string{"automationId": automationID, "projectId": projectID},
		"runs":       runs,
		"automation": automation,
		"project":    project,
		"user":       user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *AutomationHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	runID := chi.URLParam(r, "runId")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	run, err := h.automationService.GetRunByID(r.Context(), runID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Verify run belongs to the automation
	if run.AutomationID != automationID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	err = h.inertia.Render(w, r, "projects/[projectId]/automations/[automationId]/runs/[runId]", inertia.Props{
		"params":     map[string]string{"automationId": automationID, "projectId": projectID, "runId": runID},
		"run":        run,
		"automation": automation,
		"project":    project,
		"user":       user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

type CreateActionRequest struct {
	ActionType       string `json:"action_type" validate:"required"`
	ActionConfigJSON string `json:"action_config_json"`
	ActionOrder      int    `json:"action_order" validate:"min=0"`
}

func (h *AutomationHandler) CreateAction(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")

	if err := h.verifyAutomationAccess(r.Context(), user, projectID, automationID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var req CreateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": ConvertValidationErrorsToInertia(validationErrors),
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	action, err := h.automationService.CreateAction(r.Context(), stepID, req.ActionType, req.ActionConfigJSON, req.ActionOrder)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to create action")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create action"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Action created successfully")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Action created successfully",
		"action":  action,
	})
}

func (h *AutomationHandler) UpdateAction(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")
	actionID := chi.URLParam(r, "actionId")

	if err := h.verifyAutomationAccess(r.Context(), user, projectID, automationID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var req CreateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": ConvertValidationErrorsToInertia(validationErrors),
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	action := &automation.AutomationAction{
		ID:               actionID,
		StepID:           stepID,
		ActionType:       req.ActionType,
		ActionConfigJSON: req.ActionConfigJSON,
		ActionOrder:      req.ActionOrder,
	}

	err := h.automationService.UpdateAction(r.Context(), action)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to update action")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update action"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Action updated successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Action updated successfully",
		"action":  action,
	})
}

func (h *AutomationHandler) DeleteAction(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")
	_ = stepID
	actionID := chi.URLParam(r, "actionId")

	if err := h.verifyAutomationAccess(r.Context(), user, projectID, automationID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	err := h.automationService.DeleteAction(r.Context(), actionID)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to delete action")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete action"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Action deleted successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Action deleted successfully"})
}

// Helper to verify access to automation based on project and organization ownership
func (h *AutomationHandler) verifyAutomationAccess(ctx context.Context, user *auth.User, projectID, automationID string) error {
	project, err := h.projectService.GetProjectByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found")
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		return fmt.Errorf("access denied to project")
	}

	automation, err := h.automationService.GetAutomationByID(ctx, automationID)
	if err != nil {
		return fmt.Errorf("automation not found")
	}

	if automation.ProjectID != projectID {
		return fmt.Errorf("access denied to automation")
	}
	return nil
}

func (h *AutomationHandler) CancelRun(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	runID := chi.URLParam(r, "runId")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Automation not found"})
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	// Cancel the run using the scheduler
	err = h.scheduler.CancelRun(r.Context(), projectID, runID)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to cancel automation run")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Automation run cancelled successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Run cancelled successfully"})
}

func (h *AutomationHandler) GetRunEvents(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")
	runID := chi.URLParam(r, "runId")

	// Verify access (same as other methods)
	if err := h.verifyAutomationAccess(r.Context(), user, projectID, automationID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Verify run belongs to automation
	run, err := h.automationService.GetRunByID(r.Context(), runID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if run.AutomationID != automationID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Create SSE channel for this specific run
	channel := fmt.Sprintf("/events/run/%s", runID)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Handle the SSE connection
	h.sseManager.GetServer().ServeHTTP(w, r.WithContext(
		context.WithValue(r.Context(), "channel", channel),
	))
}

func (h *AutomationHandler) ExportAutomationConfig(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "projectId")
	automationID := chi.URLParam(r, "id")

	// Verify project belongs to user's organization
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	automation, err := h.automationService.GetAutomationByID(r.Context(), automationID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Automation not found"})
		return
	}

	// Verify automation belongs to the project
	if automation.ProjectID != projectID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	// Get full automation config
	exportedConfig, err := h.automationService.GetFullAutomationConfig(r.Context(), automationID)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to export automation config")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to export automation config"})
		return
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(exportedConfig, "", "  ")
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to serialize automation config")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to serialize automation config"})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("automation_config_%s.json", automationID)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(jsonData)))

	// Write JSON data
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Automation config exported successfully")
}
