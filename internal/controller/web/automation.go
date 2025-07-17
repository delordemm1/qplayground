package web

import (
	"encoding/json"
	"net/http"

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

	// Run management
	r.Post("/{id}/runs", automationHandler.TriggerRun)
	r.Get("/{id}/runs", automationHandler.ListRuns)
	r.Get("/{id}/runs/{runId}", automationHandler.GetRun)

	return r
}

func NewAutomationHandler(inertia *inertia.Inertia, sessionManager *scs.SessionManager, automationService automation.AutomationService, projectService project.ProjectService) *AutomationHandler {
	return &AutomationHandler{
		inertia:           inertia,
		sessionManager:    sessionManager,
		automationService: automationService,
		projectService:    projectService,
	}
}

type AutomationHandler struct {
	inertia           *inertia.Inertia
	sessionManager    *scs.SessionManager
	automationService automation.AutomationService
	projectService    project.ProjectService
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
		stepsWithActions[i] = map[string]interface{}{
			"step":    step,
			"actions": actions,
		}
	}

	err = h.inertia.Render(w, r, "automations/show", inertia.Props{
		"automation": automation,
		"project":    project,
		"steps":      stepsWithActions,
		"user":       user,
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

	err = h.inertia.Render(w, r, "automations/runs/index", inertia.Props{
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

	err = h.inertia.Render(w, r, "automations/runs/show", inertia.Props{
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
