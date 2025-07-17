package web

import (
	"net/http"

	"github.com/delordemm1/qplayground/internal/modules/organization"
	"github.com/delordemm1/qplayground/internal/platform"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	inertia "github.com/romsar/gonertia/v2"
)

func NewOrganizationRouter(orgHandler *OrganizationHandler) chi.Router {
	r := chi.NewRouter()
	
	r.Get("/", orgHandler.ListOrganizations)
	r.Get("/{id}", orgHandler.GetOrganization)
	
	return r
}

func NewOrganizationHandler(inertia *inertia.Inertia, sessionManager *scs.SessionManager, orgService organization.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		inertia:        inertia,
		sessionManager: sessionManager,
		orgService:     orgService,
	}
}

type OrganizationHandler struct {
	inertia        *inertia.Inertia
	sessionManager *scs.SessionManager
	orgService     organization.OrganizationService
}

func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	organizations, err := h.orgService.GetUserOrganizations(r.Context(), user.ID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	err = h.inertia.Render(w, r, "organizations/index", inertia.Props{
		"organizations": organizations,
		"user":          user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	orgID := chi.URLParam(r, "id")
	org, err := h.orgService.GetOrganizationByID(r.Context(), orgID)
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}

	// Check if user owns this organization
	if org.OwnerUserID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return
	}

	err = h.inertia.Render(w, r, "organizations/show", inertia.Props{
		"organization": org,
		"user":         user,
	})
	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}