package web

import (
	"net/http"

	"github.com/delordemm1/qplayground/internal/platform"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	inertia "github.com/romsar/gonertia/v2"
)

func NewPublicRouter(publicHandler *PublicHandler) chi.Router {
	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.Handle("/build/*", http.StripPrefix("/build/", http.FileServer(http.Dir("./public/build"))))

	// Public
	r.Get("/", publicHandler.Home)
	r.Get("/about", publicHandler.About)
	r.Get("/contact", publicHandler.Contact)
	r.Get("/blog", publicHandler.Blog)
	r.Get("/blog/{slug}", publicHandler.BlogPost)
	r.Get("/terms", publicHandler.Terms)

	return r
}

func NewPublicHandler(inertia *inertia.Inertia, sessionManager *scs.SessionManager) *PublicHandler {
	return &PublicHandler{
		inertia:        inertia,
		sessionManager: sessionManager,
	}
}

type PublicHandler struct {
	inertia        *inertia.Inertia
	sessionManager *scs.SessionManager
}

func (h *PublicHandler) Home(w http.ResponseWriter, r *http.Request) {

	err := h.inertia.Render(w, r, "/", inertia.Props{})

	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}
func (h *PublicHandler) About(w http.ResponseWriter, r *http.Request) {
	err := h.inertia.Render(w, r, "about", inertia.Props{})

	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}
func (h *PublicHandler) Contact(w http.ResponseWriter, r *http.Request) {
	err := h.inertia.Render(w, r, "contact", inertia.Props{})

	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *PublicHandler) Blog(w http.ResponseWriter, r *http.Request) {
	err := h.inertia.Render(w, r, "blog", inertia.Props{})

	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *PublicHandler) BlogPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	err := h.inertia.Render(w, r, "blogPost", inertia.Props{
		"slug": slug,
	})

	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}

func (h *PublicHandler) Terms(w http.ResponseWriter, r *http.Request) {
	err := h.inertia.Render(w, r, "terms", inertia.Props{})

	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}
