package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/delordemm1/qplayground/internal/platform"

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

type authRepository struct {
	db DBTX
	sq sq.StatementBuilderType
}

func NewAuthRepository(conn DBTX) *authRepository {
	return &authRepository{
		db: conn,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *authRepository) CreateUser(ctx context.Context, user *User) error {
	query, args, err := r.sq.Insert("users").
		Columns("id", "role", "sub", "avatar", "email", "current_org_id").
		Values(user.ID, "USER", user.Sub, user.Avatar, user.Email, user.CurrentOrgID).
		Suffix("RETURNING id, email, avatar, role, sub, current_org_id, deleted_at, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var deletedAt pgtype.Timestamp
	var createdAt, updatedAt pgtype.Timestamp
	var role string
	var currentOrgID pgtype.Text

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.Avatar, &role, &user.Sub, &currentOrgID,
		&deletedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.Role = UserRole(role)
	if currentOrgID.Valid {
		user.CurrentOrgID = &currentOrgID.String
	}
	user.CreatedAt = createdAt.Time
	user.UpdatedAt = updatedAt.Time
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return nil
}

func (r *authRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := r.sq.Select("id", "email", "avatar", "role", "sub", "current_org_id", "deleted_at", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user User
	var deletedAt pgtype.Timestamp
	var createdAt, updatedAt pgtype.Timestamp
	var role string
	var currentOrgID pgtype.Text

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.Avatar, &role, &user.Sub, &currentOrgID,
		&deletedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	user.Role = UserRole(role)
	if currentOrgID.Valid {
		user.CurrentOrgID = &currentOrgID.String
	}
	user.CreatedAt = createdAt.Time
	user.UpdatedAt = updatedAt.Time
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

func (r *authRepository) FindUserBySub(ctx context.Context, sub string) (*User, error) {
	query, args, err := r.sq.Select("id", "email", "avatar", "role", "sub", "current_org_id", "deleted_at", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"sub": sub}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user User
	var deletedAt pgtype.Timestamp
	var createdAt, updatedAt pgtype.Timestamp
	var role string
	var currentOrgID pgtype.Text

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.Avatar, &role, &user.Sub, &currentOrgID,
		&deletedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by sub: %w", err)
	}

	user.Role = UserRole(role)
	if currentOrgID.Valid {
		user.CurrentOrgID = &currentOrgID.String
	}
	user.CreatedAt = createdAt.Time
	user.UpdatedAt = updatedAt.Time
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	query, args, err := r.sq.Select("id", "email", "avatar", "role", "sub", "current_org_id", "deleted_at", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user User
	var deletedAt pgtype.Timestamp
	var createdAt, updatedAt pgtype.Timestamp
	var role string
	var currentOrgID pgtype.Text

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.Avatar, &role, &user.Sub, &currentOrgID,
		&deletedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	user.Role = UserRole(role)
	if currentOrgID.Valid {
		user.CurrentOrgID = &currentOrgID.String
	}
	user.CreatedAt = createdAt.Time
	user.UpdatedAt = updatedAt.Time
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

func (r *authRepository) UpdateUserSub(ctx context.Context, id string, newSub string) error {
	query, args, err := r.sq.Update("users").
		Set("sub", newSub).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user sub: %w", err)
	}

	return nil
}

func (r *authRepository) UpdateUserCurrentOrgID(ctx context.Context, userID string, orgID *string) error {
	query, args, err := r.sq.Update("users").
		Set("current_org_id", orgID).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user current org ID: %w", err)
	}

	return nil
}

func (r *authRepository) CreateUserActiveSession(ctx context.Context, sess *UserActiveSession) error {
	query, args, err := r.sq.Insert("user_active_sessions").
		Columns("id", "user_id", "session_token", "user_agent", "ip_address").
		Values(sess.ID, sess.UserID, sess.SessionToken, sess.UserAgent, sess.IpAddress).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create user active session: %w", err)
	}
	return nil
}
func (r *authRepository) UpdateUserActiveSessionTimestamp(ctx context.Context, sessionToken string) error {
	query, args, err := r.sq.Update("user_active_sessions").
		Set("last_active_at", time.Now()).
		Where(sq.Eq{"session_token": sessionToken}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = r.db.Exec(ctx, query, args...)
	// It's okay if this fails occasionally (e.g., race condition on logout), so we don't return the error.
	if err != nil {
		slog.Warn("Failed to update session activity timestamp", "error", err)
	}
	return nil
}
func (r *authRepository) DeleteSessionByToken(ctx context.Context, sessionToken string) error {
	// Note: We are deleting from `sessions`, the table used by `scs`.
	query, args, err := r.sq.Delete("sessions").
		Where(sq.Eq{"token": sessionToken}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete session by token: %w", err)
	}
	return nil
}

// Verification code methods
func (r *authRepository) InsertVerificationCode(ctx context.Context, v *VerificationCode) error {
	query, args, err := r.sq.Insert("verification_codes").
		Columns("id", "user_id", "contact_method", "code", "channel", "purpose", "expires_at").
		Values(v.ID, platform.UtilStrPtrToPGText(v.UserID), v.ContactMethod, v.Code, string(v.Channel), string(v.Purpose), v.ExpiresAt).
		Suffix("RETURNING id, user_id, contact_method, code, channel, purpose, expires_at, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var userID pgtype.Text
	var expiresAt, createdAt pgtype.Timestamp
	var channel, purpose string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&v.ID, &userID, &v.ContactMethod, &v.Code, &channel, &purpose,
		&expiresAt, &createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert verification code: %w", err)
	}

	if userID.Valid {
		v.UserID = &userID.String
	}
	v.Channel = VerificationChannel(channel)
	v.Purpose = VerificationPurpose(purpose)
	v.ExpiresAt = expiresAt.Time
	v.CreatedAt = createdAt.Time

	return nil
}

func (r *authRepository) GetVerificationCodeByEmailAndCode(ctx context.Context, email, code string, purpose VerificationPurpose) (*VerificationCode, error) {
	query, args, err := r.sq.Select("id", "user_id", "contact_method", "code", "channel", "purpose", "expires_at", "created_at").
		From("verification_codes").
		Where(sq.Eq{
			"contact_method": email,
			"code":           code,
			"purpose":        string(purpose),
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var v VerificationCode
	var userID pgtype.Text
	var expiresAt, createdAt pgtype.Timestamp
	var channel, purposeStr string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&v.ID, &userID, &v.ContactMethod, &v.Code, &channel, &purposeStr,
		&expiresAt, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}

	if userID.Valid {
		v.UserID = &userID.String
	}
	v.Channel = VerificationChannel(channel)
	v.Purpose = VerificationPurpose(purposeStr)
	v.ExpiresAt = expiresAt.Time
	v.CreatedAt = createdAt.Time

	return &v, nil
}

func (r *authRepository) GetVerificationCode(ctx context.Context, userID, code string, purpose VerificationPurpose) (*VerificationCode, error) {
	query, args, err := r.sq.Select("id", "user_id", "contact_method", "code", "channel", "purpose", "expires_at", "created_at").
		From("verification_codes").
		Where(sq.Eq{
			"user_id": userID,
			"code":    code,
			"purpose": string(purpose),
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var v VerificationCode
	var userIDPg pgtype.Text
	var expiresAt, createdAt pgtype.Timestamp
	var channel, purposeStr string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&v.ID, &userIDPg, &v.ContactMethod, &v.Code, &channel, &purposeStr,
		&expiresAt, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}

	if userIDPg.Valid {
		v.UserID = &userIDPg.String
	}
	v.Channel = VerificationChannel(channel)
	v.Purpose = VerificationPurpose(purposeStr)
	v.ExpiresAt = expiresAt.Time
	v.CreatedAt = createdAt.Time

	return &v, nil
}

func (r *authRepository) DeleteVerificationCode(ctx context.Context, id string) error {
	query, args, err := r.sq.Delete("verification_codes").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete verification code: %w", err)
	}

	return nil
}

// OAuth state methods
func (r *authRepository) InsertOAuthState(ctx context.Context, s *OAuthState) error {
	query, args, err := r.sq.Insert("oauth_states").
		Columns("state", "provider", "verifier", "expires_at").
		Values(s.State, string(s.Provider), s.Verifier, s.ExpiresAt).
		Suffix("RETURNING state, provider, user_id, verifier, expires_at, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	var userID pgtype.Text
	var expiresAt, createdAt, updatedAt pgtype.Timestamp
	var provider string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&s.State, &provider, &userID, &s.Verifier,
		&expiresAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert OAuth state: %w", err)
	}

	s.Provider = OAuthProvider(provider)
	if userID.Valid {
		s.UserID = &userID.String
	}
	s.ExpiresAt = expiresAt.Time
	s.CreatedAt = createdAt.Time
	s.UpdatedAt = updatedAt.Time

	return nil
}

func (r *authRepository) GetOAuthStateByState(ctx context.Context, state string) (*OAuthState, error) {
	query, args, err := r.sq.Select("state", "provider", "user_id", "verifier", "expires_at", "created_at", "updated_at").
		From("oauth_states").
		Where(sq.Eq{"state": state}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var s OAuthState
	var userID pgtype.Text
	var expiresAt, createdAt, updatedAt pgtype.Timestamp
	var provider string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&s.State, &provider, &userID, &s.Verifier,
		&expiresAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get OAuth state: %w", err)
	}

	s.Provider = OAuthProvider(provider)
	if userID.Valid {
		s.UserID = &userID.String
	}
	s.ExpiresAt = expiresAt.Time
	s.CreatedAt = createdAt.Time
	s.UpdatedAt = updatedAt.Time

	return &s, nil
}

func (r *authRepository) UpdateOAuthStateUserID(ctx context.Context, state string, userID string) (*OAuthState, error) {
	query, args, err := r.sq.Update("oauth_states").
		Set("user_id", userID).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"state": state}).
		Suffix("RETURNING state, provider, user_id, verifier, expires_at, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var s OAuthState
	var userIDPg pgtype.Text
	var expiresAt, createdAt, updatedAt pgtype.Timestamp
	var provider string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&s.State, &provider, &userIDPg, &s.Verifier,
		&expiresAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update OAuth state: %w", err)
	}

	s.Provider = OAuthProvider(provider)
	if userIDPg.Valid {
		s.UserID = &userIDPg.String
	}
	s.ExpiresAt = expiresAt.Time
	s.CreatedAt = createdAt.Time
	s.UpdatedAt = updatedAt.Time

	return &s, nil
}

func (r *authRepository) DeleteOAuthState(ctx context.Context, state string) error {
	query, args, err := r.sq.Delete("oauth_states").
		Where(sq.Eq{"state": state}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth state: %w", err)
	}

	return nil
}

func (r *authRepository) DeleteExpiredOAuthStates(ctx context.Context) error {
	query, args, err := r.sq.Delete("oauth_states").
		Where(sq.Lt{"expires_at": time.Now()}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete expired OAuth states: %w", err)
	}

	return nil
}
