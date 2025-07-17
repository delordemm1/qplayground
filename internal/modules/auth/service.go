package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/delordemm1/qplayground/internal/modules/notification"
	"github.com/delordemm1/qplayground/internal/platform"

	"github.com/delordemm1/qplayground/internal/modules/organization"
	"github.com/alexedwards/scs/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const AuthUserIDSessionKey = "authenticatedUserID"

func NewAuthService(authRepo AuthRepository, ns notification.NotificationService, sm *scs.SessionManager, orgService organization.OrganizationService) *AuthService {
	return &AuthService{
		authRepo:             authRepo,
		notificationService:  ns,
		sessionManager:       sm,
		organizationService:  orgService,
	}
}

type AuthService struct {
	authRepo             AuthRepository
	notificationService  notification.NotificationService
	sessionManager       *scs.SessionManager
	organizationService  organization.OrganizationService
}

type OAuth interface {
	getOAuthConfig() *oauth2.Config
	getUserInfo(accessToken string) (*UserInfo, error)
}

// RequestEmailOtp generates and sends an OTP to the provided email
func (s *AuthService) RequestEmailOtp(ctx context.Context, email string) error {
	// Generate 6-digit numeric OTP
	otp := generateOTP()

	// Create verification code record
	verificationCode := &VerificationCode{
		ID:            platform.UtilGenerateUUID(),
		UserID:        nil, // No user ID for email-based OTP
		ContactMethod: email,
		Code:          otp,
		Channel:       VerificationChannelEmail,
		Purpose:       VerificationPurposeAuth,
		ExpiresAt:     time.Now().Add(10 * time.Minute), // 10 minutes expiry
		CreatedAt:     time.Now(),
	}

	// Store verification code in database
	err := s.authRepo.InsertVerificationCode(ctx, verificationCode)
	if err != nil {
		slog.Error("Failed to store verification code", "error", err, "email", email)
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	// Send OTP via email
	err = s.notificationService.SendLoginCode(ctx, email, otp)
	if err != nil {
		slog.Error("Failed to send OTP email", "error", err, "email", email)
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	slog.Info("OTP sent successfully", "email", email)
	return nil
}

// UpdateUserCurrentOrgID
func (s *AuthService) UpdateUserCurrentOrgID(ctx context.Context, userID string, orgID *string) error {
	return s.authRepo.UpdateUserCurrentOrgID(ctx, userID, orgID)
}

// VerifyEmailOtp verifies the provided OTP and creates a session
func (s *AuthService) VerifyEmailOtp(ctx context.Context, uaIp *platform.UserAgent, email, otp string) (user *User, err error) {
	// Get verification code from database
	verificationCode, err := s.authRepo.GetVerificationCodeByEmailAndCode(ctx, email, otp, VerificationPurposeAuth)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			slog.Warn("Invalid OTP provided", "email", email, "otp", otp)
			return nil, ErrExpired
		}
		slog.Error("Failed to get verification code", "error", err, "email", email)
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}

	var isNewUser bool
	// Check if OTP has expired
	if time.Now().After(verificationCode.ExpiresAt) {
		slog.Warn("Expired OTP provided", "email", email, "otp", otp, "expired_at", verificationCode.ExpiresAt)
		// Clean up expired code
		s.authRepo.DeleteVerificationCode(ctx, verificationCode.ID)
		return nil, ErrExpired
	}

	// Find or create user
	user, err = s.authRepo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			isNewUser = true
			// Create new user
			user = &User{
				ID:    platform.UtilGenerateUUID(),
				Email: email,
				Sub:   "email:" + email,
				Role:  UserRoleUser,
			}
			err = s.authRepo.CreateUser(ctx, user)
			if err != nil {
				slog.Error("Failed to create new user", "error", err, "email", email)
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			slog.Info("New user created", "user_id", user.ID, "email", email)
		} else {
			slog.Error("Failed to find user by email", "error", err, "email", email)
			return nil, fmt.Errorf("failed to find user: %w", err)
		}
	}

	// Create personal organization for new users
	if isNewUser || user.CurrentOrgID == nil {
		personalOrg, err := s.organizationService.CreatePersonalOrganization(ctx, user.ID, user.Email)
		if err != nil {
			slog.Error("Failed to create personal organization", "error", err, "userID", user.ID)
			// Don't fail the login, but log the error
		} else {
			// Update user with personal organization ID
			err = s.authRepo.UpdateUserCurrentOrgID(ctx, user.ID, &personalOrg.ID)
			if err != nil {
				slog.Warn("Failed to update user current org ID", "error", err, "userID", user.ID)
			}
		}
	}

	// Delete used verification code
	err = s.authRepo.DeleteVerificationCode(ctx, verificationCode.ID)
	if err != nil {
		slog.Warn("Failed to delete used verification code", "error", err, "code_id", verificationCode.ID)
		// Don't fail the request for this
	}

	slog.Info("OTP verified successfully", "user_id", user.ID, "email", email)
	return user, nil
}

// Logout invalidates the user's session
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	err := s.sessionManager.Destroy(ctx)
	if err != nil {
		slog.Error("Failed to destroy session", "error", err)
		return fmt.Errorf("failed to destroy session: %w", err)
	}
	slog.Info("Session destroyed successfully")
	return nil
}

// generateOTP generates a 6-digit numeric OTP
func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *AuthService) OAuthLogin(ctx context.Context, provider string) (string, error) {
	var OAuth OAuth
	var oauthProvider OAuthProvider
	if provider == "google" {
		OAuth = newOAuthGoogle()
		oauthProvider = OAuthProviderGOOGLE
		// } else if provider == "github" {
		// 	OAuth = newOAuthGithub()
	} else {
		slog.Error("Invalid provider", "provider", provider)
		return platform.ENV_APP_URL + "/auth?error=unauthorized", nil
	}
	// generate random state and verifier
	state, err := platform.UtilGenerateRandomState(32)
	if err != nil {
		slog.Error("Error generating random state", "GenerateRandomState", err)
		return platform.ENV_APP_URL + "/auth?error=unauthorized", nil
	}
	verifier := oauth2.GenerateVerifier()
	slog.Debug("verifier generated", "verifier", verifier)
	// store state and verifier in database
	err = s.authRepo.InsertOAuthState(ctx, &OAuthState{
		// ID:     platform.UtilGenerateUUID(),
		State:     state,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Verifier:  verifier,
		Provider:  oauthProvider,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		slog.Error("Error inserting token", "insertToken", err)
		return "/auth?error=unauthorized", errors.New("error inserting token")
	}
	// TODO: redirect to oauth provider
	url := OAuth.getOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	return url, nil
}

func (s *AuthService) OAuthCallback(ctx context.Context, uaIp *platform.UserAgent, provider OAuthProvider, state string, code string) (yser *User, err error) {
	var OAuth OAuth
	if provider == OAuthProviderGOOGLE {
		OAuth = newOAuthGoogle()
		// } else if provider == "github" {
		// 	OAuth = newOAuthGithub()
	} else {
		slog.Error("Invalid provider", "provider", provider)
		return nil, nil
	}
	token, err := s.authRepo.GetOAuthStateByState(ctx, state)
	if err != nil {
		slog.Error("Error getting token", "getToken", err)
		return nil, errors.New("error getting token")
	}
	if time.Now().After(token.ExpiresAt) {
		slog.Error("Token expired", "token", token)
		return nil, errors.New("token expired")
	}
	defer s.authRepo.DeleteOAuthState(ctx, state)
	// get oauth config
	config := OAuth.getOAuthConfig()
	slog.Info("OAuth config", "config", config, "token", token, "provider", provider)
	// get oauth token
	oauthToken, err := config.Exchange(ctx, code, oauth2.VerifierOption(token.Verifier))
	if err != nil {
		slog.Error("Error exchanging code for token", "config.Exchange", err)
		return nil, errors.New("error during token exchange")
	}
	// fetch user info from google
	userInfo, err := OAuth.getUserInfo(oauthToken.AccessToken)
	if err != nil {
		slog.Error("Error getting user info", "getUserInfo", err)
		return nil, errors.New("error getting user info")
	}

	user, err := s.authRepo.FindUserByEmail(ctx, userInfo.email)
	if err != nil {
		if err == ErrUserNotFound {
			// create user
			newUser := &User{
				ID:    platform.UtilGenerateUUID(),
				Email: userInfo.email,
				// Name:      userInfo.,
				Avatar: userInfo.avatar,
				Sub:    string(provider) + ":" + userInfo.sub,
			}
			err = s.authRepo.CreateUser(ctx, newUser)
			if err != nil {
				slog.Error("Failed to create new user from OAuth", "error", err, "email", userInfo.email)
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			slog.Info("New user created via OAuth", "user_id", newUser.ID, "email", newUser.Email)
			return newUser, nil
		} else {
			slog.Error("Failed to find user by email", "error", err, "email", userInfo.email)
			return nil, fmt.Errorf("failed to find user: %w", err)
		}
	}

	// Create personal organization for new users (OAuth)
	if user.CurrentOrgID == nil {
		personalOrg, err := s.organizationService.CreatePersonalOrganization(ctx, user.ID, user.Email)
		if err != nil {
			slog.Error("Failed to create personal organization for OAuth user", "error", err, "userID", user.ID)
		} else {
			// Update user with personal organization ID
			err = s.authRepo.UpdateUserCurrentOrgID(ctx, user.ID, &personalOrg.ID)
			if err != nil {
				slog.Warn("Failed to update OAuth user current org ID", "error", err, "userID", user.ID)
			}
		}
	}

	// validate provider
	err = s.validateProvider(ctx, user.ID, user.Sub, string(provider), userInfo.sub)
	if err != nil {
		slog.Error("Error validating provider", "validateProvider", err)
		return nil, errors.New("error validating provider")
	}
	// redirect to home page
	return user, nil
}

// User can have multiple providers, in form of "provider:sub"
// User cannot have multiple providers of the same type
// If user don't have provider, add it
func (s *AuthService) validateProvider(ctx context.Context, userId string, userSubs string, provider string, sub string) error {
	var err error
	// if doesn't have provider, add it
	if !strings.Contains(userSubs, provider) {
		err = s.authRepo.UpdateUserSub(ctx, userId, userSubs+","+provider+":"+sub)
		if err != nil {
			return fmt.Errorf("updateUserSub: %w", err)
		}
		return nil
	}

	// if have provider, check if it's the same
	subs := strings.Split(userSubs, ",")
	for _, s := range subs {
		if strings.Contains(s, provider) {
			if s != provider+":"+sub {
				return fmt.Errorf("provider already exists")
			}
		}
	}
	return nil
}

// var githubOAuthConfig = oauth2.Config{
// 	ClientID:     platform.GITHUB_CLIENT_ID,
// 	ClientSecret: platform.GITHUB_CLIENT_SECRET,
// 	Endpoint: oauth2.Endpoint{
// 		AuthURL:  "https://github.com/login/oauth/authorize",
// 		TokenURL: "https://github.com/login/oauth/access_token",
// 	},
// 	RedirectURL: platform.APP_URL + "/auth/oauth:github/callback",
// 	Scopes:      []string{"user:email"},
// }

var googleOAuthConfig = oauth2.Config{
	ClientID:     platform.ENV_GOOGLE_OAUTH_CLIENT_ID,
	ClientSecret: platform.ENV_GOOGLE_OAUTH_CLIENT_SECRET,
	Endpoint:     google.Endpoint,
	RedirectURL:  platform.ENV_APP_URL + "/auth/oauth:google/callback",
	Scopes:       []string{"profile", "email", "openid"},
}

type UserInfo struct {
	email  string
	sub    string
	avatar string
}

type OAuthGoogle struct {
	googleOAuthConfig oauth2.Config
}

func newOAuthGoogle() *OAuthGoogle {
	return &OAuthGoogle{
		googleOAuthConfig: googleOAuthConfig,
	}
}

type OAuthGithub struct {
	githubOAuthConfig oauth2.Config
}

// func newOAuthGithub() *OAuthGithub {
// 	return &OAuthGithub{
// 		githubOAuthConfig: githubOAuthConfig,
// 	}
// }

func (o *OAuthGoogle) getOAuthConfig() *oauth2.Config {
	return &o.googleOAuthConfig
}

func (o *OAuthGoogle) getUserInfo(accessToken string) (*UserInfo, error) {
	url := "https://www.googleapis.com/oauth2/v2/userinfo"
	userInfo, err := httpCall("GET", url, accessToken)
	if err != nil {
		return nil, fmt.Errorf("httpCall: %w", err)
	}
	slog.Debug("Received user info", slog.String("user_info", fmt.Sprintf("%+v", userInfo)))
	sub, ok := userInfo["id"].(string)
	if !ok {
		return nil, fmt.Errorf("Invalid user id")
	}
	email, ok := userInfo["email"].(string)
	if !ok {
		email = ""
	}
	avatar, ok := userInfo["picture"].(string)
	if !ok {
		avatar = ""
	}
	return &UserInfo{
		email:  email,
		sub:    sub,
		avatar: avatar,
	}, nil
}

func (o *OAuthGithub) getOAuthConfig() *oauth2.Config {
	return &o.githubOAuthConfig
}

func (o *OAuthGithub) getUserInfo(accessToken string) (*UserInfo, error) {
	url := "https://api.github.com/user"
	userInfo, err := httpCall("GET", url, accessToken)
	if err != nil {
		return nil, fmt.Errorf("httpCall: %w", err)
	}
	userId, ok := userInfo["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("Invalid user id")
	}
	sub := fmt.Sprintf("%.0f", userId)
	email, ok := userInfo["email"].(string)
	if !ok {
		email = ""
	}
	avatar, ok := userInfo["avatar_url"].(string)
	if !ok {
		avatar = ""
	}
	return &UserInfo{
		email:  email,
		sub:    sub,
		avatar: avatar,
	}, nil
}

func httpCall(method, url, accessToken string) (map[string]any, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do: %w", err)
	}
	defer resp.Body.Close()
	var userInfo map[string]any
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, fmt.Errorf("json.NewDecoder: %w", err)
	}
	return userInfo, nil
}

func (s *AuthService) Auth(ctx context.Context) (*User, error) {
	userID := s.sessionManager.GetString(ctx, AuthUserIDSessionKey)
	if userID == "" {
		return nil, ErrUnauthorized // No user ID in session.
	}
	authUser, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	sessionToken := s.sessionManager.Token(ctx)
	go s.authRepo.UpdateUserActiveSessionTimestamp(context.Background(), sessionToken)
	return authUser, nil
}

func (s *AuthService) CreateActiveSession(ctx context.Context, userID, sessionToken, userAgent, ipAddress string) error {
	activeSession := &UserActiveSession{
		ID:           platform.UtilGenerateUUID(),
		UserID:       userID,
		SessionToken: sessionToken,
		UserAgent:    userAgent,
		IpAddress:    ipAddress,
	}
	return s.authRepo.CreateUserActiveSession(ctx, activeSession)
}
