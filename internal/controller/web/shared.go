package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/delordemm1/qplayground/internal/modules/auth"
	"github.com/delordemm1/qplayground/internal/platform"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/validator/v10"
	inertia "github.com/romsar/gonertia/v2"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func getUserFromContext(ctx context.Context) *auth.User {
	if user, ok := ctx.Value(auth.AuthUserIDSessionKey).(*auth.User); ok {
		return user
	}
	return nil
}

type SiteMiddleware struct {
	inertia        *inertia.Inertia
	sessionManager *scs.SessionManager
}
type AuthMiddleware struct {
	inertia        *inertia.Inertia
	sessionManager *scs.SessionManager
	authService    *auth.AuthService
}

func NewSiteMiddleware(inertia *inertia.Inertia, sessionManager *scs.SessionManager) *SiteMiddleware {
	return &SiteMiddleware{
		inertia:        inertia,
		sessionManager: sessionManager,
	}
}
func NewAuthMiddleware(inertia *inertia.Inertia, sessionManager *scs.SessionManager, authService *auth.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		inertia:        inertia,
		sessionManager: sessionManager,
		authService:    authService,
	}
}

// FlashMessageSharingMiddleware shares flash messages as Inertia props
func (m *SiteMiddleware) FlashMessageSharingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ROUGH PATCH: Clear flash message from previous request
		m.inertia.ShareProp("flash", nil)
		// Get flash message from session
		flashMessage := platform.GetFlashMessage(r.Context(), m.sessionManager)
		if flashMessage != nil {
			m.inertia.ShareProp("flash", map[string]any{
				"type":    flashMessage.Type,
				"message": flashMessage.Message,
			})
		}

		next.ServeHTTP(w, r)
	})
}

// OnlyUser middleware ensures only authenticated users can access the route
// func (m *SiteMiddleware) OnlyUser(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		userID := m.sessionManager.GetString(r.Context(), auth.AuthUserIDSessionKey)
// 		if userID == "" {
// 			http.Redirect(w, r, "/auth", http.StatusFound)
// 			return
// 		}

// 		// For now, we'll just pass the userID in context
// 		// In a full implementation, you'd fetch the full user from the database
// 		ctx := context.WithValue(r.Context(), "userID", userID)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// OnlyGuest middleware ensures only non-authenticated users can access the route
// func (m *SiteMiddleware) OnlyGuest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		userID := m.sessionManager.GetString(r.Context(), auth.AuthUserIDSessionKey)
// 		if userID != "" {
// 			http.Redirect(w, r, "/dashboard", http.StatusFound)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

func (m *AuthMiddleware) OnlyGuest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := m.authService.Auth(r.Context())
		if err == nil {
			redirectURL := r.URL.Query().Get("redirectTo")
			if redirectURL == "" {
				redirectURL = "/user" // Default redirect destination
			}

			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) OnlyUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUser, err := m.authService.Auth(r.Context())
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), auth.AuthUserIDSessionKey, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserAgentProps(r *http.Request) *platform.UserAgent {
	return &platform.UserAgent{
		UserAgent: r.UserAgent(),
		IP:        r.RemoteAddr,
	}
}

// FormatValidationErrorsToMap converts validation errors to a user-friendly map.
func formatValidationErrorsToMap(errs validator.ValidationErrors) map[string]string {
	// The map that will hold our field-to-error-message mapping
	errorMessages := make(map[string]string)

	// Iterate over each validation error
	for _, err := range errs {
		// Get the field name from the struct
		field := err.Field()

		// Get the validation tag that failed (e.g., "required", "email")
		tag := err.Tag()

		// Get the validation parameter if it exists (e.g., "6" for "len=6")
		param := err.Param()

		// Create a user-friendly message based on the tag
		var message string
		switch tag {
		case "required":
			message = fmt.Sprintf("%s is a required field.", field)
		case "email":
			message = fmt.Sprintf("Please provide a valid email address for the %s field.", field)
		case "len":
			message = fmt.Sprintf("%s must be exactly %s characters long.", field, param)
		case "numeric":
			message = fmt.Sprintf("%s must contain only numbers.", field)
		default:
			message = fmt.Sprintf("%s is not valid.", field)
		}

		errorMessages[field] = message
	}

	return errorMessages
}

// ConvertValidationErrorsToInertia converts validator.ValidationErrors to inertia.ValidationErrors
func ConvertValidationErrorsToInertia(errs validator.ValidationErrors) inertia.ValidationErrors {
	validationErrors := make(inertia.ValidationErrors)
	errorMap := formatValidationErrorsToMap(errs)

	for field, message := range errorMap {
		validationErrors[field] = message
	}

	return validationErrors
}
