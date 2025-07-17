package web

import (
	"encoding/json"
	"net/http"

	"github.com/delordemm1/qplayground/internal/controller/web/dto"
	"github.com/delordemm1/qplayground/internal/modules/auth"
	"github.com/delordemm1/qplayground/internal/platform"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	inertia "github.com/romsar/gonertia/v2"
)

func NewAuthRouter(authHandler *AuthHandler) chi.Router {
	r := chi.NewRouter()

	// Authentication routes
	r.Get("/", authHandler.AuthPage)
	r.Post("/request-otp", authHandler.RequestOTP)
	r.Post("/verify-otp", authHandler.VerifyOTP)
	r.Post("/logout", authHandler.Logout)

	// OAuth routes (for future use)
	r.Get("/oauth/{provider}", authHandler.OAuthLogin)
	r.Get("/oauth/{provider}/callback", authHandler.OAuthCallback)

	return r
}

func NewAuthHandler(inertia *inertia.Inertia, sessionManager *scs.SessionManager, authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		inertia:        inertia,
		sessionManager: sessionManager,
		authService:    authService,
	}
}

type AuthHandler struct {
	inertia        *inertia.Inertia
	sessionManager *scs.SessionManager
	authService    *auth.AuthService
}

func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req dto.SendOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Invalid request format")
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
		platform.SetFlashError(r.Context(), h.sessionManager, "Validation failed")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	// Send OTP
	err := h.authService.RequestEmailOtp(r.Context(), req.Email)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Failed to send OTP. Please try again.")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send OTP"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "OTP sent to your email address")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent successfully"})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginWithOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Invalid request format")
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
		platform.SetFlashError(r.Context(), h.sessionManager, "Validation failed")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	// Get user agent info
	userAgent := getUserAgentProps(r)

	// Verify OTP
	user, err := h.authService.VerifyEmailOtp(r.Context(), userAgent, req.Email, req.OTP)
	if err != nil {
		if err == auth.ErrExpired {
			platform.SetFlashError(r.Context(), h.sessionManager, "Invalid or expired OTP")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired OTP"})
			return
		}
		platform.SetFlashError(r.Context(), h.sessionManager, "Authentication failed")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication failed"})
		return
	}

	// Create session
	h.sessionManager.Put(r.Context(), auth.AuthUserIDSessionKey, user.ID)
	sessionToken := h.sessionManager.Token(r.Context())

	// Create active session record
	err = h.authService.CreateActiveSession(r.Context(), user.ID, sessionToken, userAgent.UserAgent, userAgent.IP)
	if err != nil {
		// Log error but don't fail the login
		platform.SetFlashWarning(r.Context(), h.sessionManager, "Session tracking failed, but login successful")
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Login successful!")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
		"redirect": "/dashboard",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := h.sessionManager.Token(r.Context())

	err := h.authService.Logout(r.Context(), sessionToken)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Logout failed")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Logout failed"})
		return
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Logged out successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	url, err := h.authService.OAuthLogin(r.Context(), provider)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "OAuth login failed")
		http.Redirect(w, r, "/auth?error=oauth_failed", http.StatusFound)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		platform.SetFlashError(r.Context(), h.sessionManager, "Invalid OAuth callback")
		http.Redirect(w, r, "/auth?error=invalid_callback", http.StatusFound)
		return
	}

	userAgent := getUserAgentProps(r)

	user, err := h.authService.OAuthCallback(r.Context(), userAgent, auth.OAuthProvider(provider), state, code)
	if err != nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "OAuth authentication failed")
		http.Redirect(w, r, "/auth?error=oauth_failed", http.StatusFound)
		return
	}

	if user == nil {
		platform.SetFlashError(r.Context(), h.sessionManager, "Authentication failed")
		http.Redirect(w, r, "/auth?error=auth_failed", http.StatusFound)
		return
	}

	// Create session
	h.sessionManager.Put(r.Context(), auth.AuthUserIDSessionKey, user.ID)
	sessionToken := h.sessionManager.Token(r.Context())

	// Create active session record
	err = h.authService.CreateActiveSession(r.Context(), user.ID, sessionToken, userAgent.UserAgent, userAgent.IP)
	if err != nil {
		// Log error but don't fail the login
		platform.SetFlashWarning(r.Context(), h.sessionManager, "Session tracking failed, but login successful")
	}

	platform.SetFlashSuccess(r.Context(), h.sessionManager, "Login successful!")
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (h *AuthHandler) AuthPage(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	err := h.inertia.Render(w, r, "auth", inertia.Props{})

	if err != nil {
		platform.UtilHandleServerErr(w, err)
		return
	}
}
