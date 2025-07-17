package auth

import (
	"context"
	"time"
)

// The canonical user as seen by your business logic/service layers
type User struct {
	ID             string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	Role           UserRole
	Sub            string
	Avatar         string
	Email          string
	CurrentOrgID   *string            // Current organization ID for dynamic role switching
	ActiveSessions *UserActiveSession // one-to-one
}

// Role enum
type UserRole string

const (
	UserRoleUser  UserRole = "USER"
	UserRoleAdmin UserRole = "ADMIN"
)

type OAuthProvider string

const (
	OAuthProviderGOOGLE   OAuthProvider = "google"
	OAuthProviderFACEBOOK OAuthProvider = "facebook"
	OAuthProviderGITHUB   OAuthProvider = "github"
	OAuthProviderX        OAuthProvider = "x"
	OAuthProviderLINKEDIN OAuthProvider = "linkedin"
)

type OAuthState struct {
	State     string
	Provider  OAuthProvider
	UserID    *string
	Verifier  string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SessionInvalidationReason string

const (
	SessionInvalidationReasonLogout           SessionInvalidationReason = "logout"
	SessionInvalidationReasonUserInvalidated  SessionInvalidationReason = "user_invalidated"
	SessionInvalidationReasonInactivityExpiry SessionInvalidationReason = "inactivity_expiry"
)

type UserActiveSession struct {
	ID           string
	UserID       string
	SessionToken string // Links to the `sessions.token` table managed by scs
	UserAgent    string
	IpAddress    string
	LastActiveAt time.Time
	CreatedAt    time.Time
}

type VerificationCode struct {
	ID            string
	UserID        *string
	ContactMethod string
	Code          string
	Channel       VerificationChannel
	Purpose       VerificationPurpose
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

type VerificationChannel string

const (
	VerificationChannelEmail    VerificationChannel = "EMAIL"
	VerificationChannelPhone    VerificationChannel = "PHONE"
	VerificationChannelWhatsapp VerificationChannel = "WHATSAPP"
	VerificationChannelTelegram VerificationChannel = "TELEGRAM"
)

type VerificationPurpose string

const (
	VerificationPurposeAuth VerificationPurpose = "AUTH"
)

type AuthRepository interface {
	// User CRUD
	CreateUser(ctx context.Context, user *User) error
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserBySub(ctx context.Context, sub string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateUserSub(ctx context.Context, id string, newSub string) error
	UpdateUserCurrentOrgID(ctx context.Context, userID string, orgID *string) error

	// Session/token
	CreateUserActiveSession(ctx context.Context, sess *UserActiveSession) error
	UpdateUserActiveSessionTimestamp(ctx context.Context, sessionToken string) error
	DeleteSessionByToken(ctx context.Context, sessionToken string) error

	// Verification codes (for email login)
	InsertVerificationCode(ctx context.Context, code *VerificationCode) error
	GetVerificationCodeByEmailAndCode(ctx context.Context, email, code string, purpose VerificationPurpose) (*VerificationCode, error)
	GetVerificationCode(ctx context.Context, userID, code string, purpose VerificationPurpose) (*VerificationCode, error)
	DeleteVerificationCode(ctx context.Context, id string) error

	// Oauth states (for social login)
	InsertOAuthState(ctx context.Context, state *OAuthState) error
	GetOAuthStateByState(ctx context.Context, state string) (*OAuthState, error)
	UpdateOAuthStateUserID(ctx context.Context, state string, userID string) (*OAuthState, error)
	DeleteOAuthState(ctx context.Context, state string) error
	DeleteExpiredOAuthStates(ctx context.Context) error
}
