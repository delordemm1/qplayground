package config

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionConfig holds session configuration
type SessionConfig struct {
	Lifetime   time.Duration
	CookieName string
	HttpOnly   bool
	Secure     bool
	SameSite   http.SameSite
	Path       string
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		Lifetime:   120 * time.Hour, // 5 days
		CookieName: "deltechverse_session_token",
		HttpOnly:   true,
		Secure:     true, // Set to false for development if not using HTTPS
		SameSite:   http.SameSiteLaxMode,
		Path:       "/",
	}
}

// InitSession initializes and returns a session manager
func InitSession(pool *pgxpool.Pool, config SessionConfig) *scs.SessionManager {
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(pool)
	sessionManager.Lifetime = config.Lifetime
	sessionManager.Cookie.Name = config.CookieName
	sessionManager.Cookie.HttpOnly = config.HttpOnly
	sessionManager.Cookie.Secure = config.Secure
	sessionManager.Cookie.SameSite = config.SameSite
	sessionManager.Cookie.Path = config.Path

	return sessionManager
}