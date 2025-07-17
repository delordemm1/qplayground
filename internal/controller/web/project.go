package web

import (
	"encoding/json"
	"net/http"

	"github.com/delordemm1/qplayground/internal/modules/project"
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/go-playground/validator/v10"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	inertia "github.com/romsar/gonertia/v2"
)

func NewProjectRouter(projectHandler *ProjectHandler) chi.Router {
	r := chi.NewRouter()

	r.Get("/", projectHandler.ListProjects)
	r.Post("/", projectHandler.CreateProject)
	r.Get("/{id}", projectHandler.GetProject)
	r.Put("/{id}", projectHandler.UpdateProject)
	r.Delete("/{id}", projectHandler.DeleteProject)

	return r
}

func NewProjectHandler(inertia *inertia.Inertia, sessionManager *scs.SessionManager, projectService project.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		inertia:        inertia,
		sessionManager: sessionManager,
		projectService: projectService,
	}
}

type ProjectHandler struct {
	inertia        *inertia.Inertia
	sessionManager *scs.SessionManager
	projectService project.ProjectService
}

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	// Get user's current organization ID
	if user.CurrentOrgID == nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "No organization found")
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	projects, err := h.projectService.GetProjectsByOrganization(r.Context(), *user.CurrentOrgID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	err = h.inertia.Render(w, r, "projects/index", inertia.Props{
		"projects": projects,
		"user":     user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	if user.CurrentOrgID == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "No organization found"})
		return
	}

	var req CreateProjectRequest
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

	project, err := h.projectService.CreateProject(r.Context(), *user.CurrentOrgID, req.Name, req.Description)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to create project")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create project"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Project created successfully")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Project created successfully",
		"project": project,
	})
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	projectID := chi.URLParam(r, "id")
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Check if project belongs to user's organization
	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	err = h.inertia.Render(w, r, "projects/show", inertia.Props{
		"project": project,
		"user":    user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "id")
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	// Check if project belongs to user's organization
	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	var req CreateProjectRequest
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

	project.Name = req.Name
	project.Description = req.Description

	err = h.projectService.UpdateProject(r.Context(), project)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to update project")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update project"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Project updated successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Project updated successfully",
		"project": project,
	})
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	projectID := chi.URLParam(r, "id")
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found"})
		return
	}

	// Check if project belongs to user's organization
	if user.CurrentOrgID == nil || project.OrganizationID != *user.CurrentOrgID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	err = h.projectService.DeleteProject(r.Context(), projectID)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to delete project")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete project"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Project deleted successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Project deleted successfully"})
}
