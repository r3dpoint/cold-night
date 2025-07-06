package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"securities-marketplace/domains/users"
	"securities-marketplace/domains/shared/web"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	jwtManager   *JWTManager
	sessionStore SessionStore
	userService  UserService
	rbac         *RBAC
}

// UserService interface for user operations
type UserService interface {
	GetByEmail(email string) (*users.UserAggregate, error)
	GetByID(userID string) (*users.UserAggregate, error)
	Save(user *users.UserAggregate) error
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(jwtManager *JWTManager, sessionStore SessionStore, userService UserService, rbac *RBAC) *AuthHandler {
	return &AuthHandler{
		jwtManager:   jwtManager,
		sessionStore: sessionStore,
		userService:  userService,
		rbac:         rbac,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.ShowLoginForm).Methods("GET")
	router.HandleFunc("/login", h.HandleLogin).Methods("POST")
	router.HandleFunc("/logout", h.HandleLogout).Methods("POST")
	
	// API endpoints
	router.HandleFunc("/api/auth/login", h.HandleAPILogin).Methods("POST")
	router.HandleFunc("/api/auth/logout", h.HandleAPILogout).Methods("POST")
	router.HandleFunc("/api/auth/refresh", h.HandleAPIRefreshToken).Methods("POST")
	router.HandleFunc("/api/auth/profile", h.HandleAPIGetProfile).Methods("GET")
}

// ShowLoginForm displays the login form
func (h *AuthHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	// Check if user is already logged in
	if token, err := h.extractToken(r); err == nil {
		if _, err := h.jwtManager.VerifyToken(token); err == nil {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	data := map[string]interface{}{
		"Title": "Login",
		"Error": r.URL.Query().Get("error"),
		"Message": r.URL.Query().Get("message"),
	}

	web.RenderTemplate(w, "login.html", data)
}

// HandleLogin handles form-based login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error=Invalid form data", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	rememberMe := r.FormValue("rememberMe") == "on"

	loginResponse, err := h.processLogin(email, password, rememberMe, getClientIP(r), r.UserAgent())
	if err != nil {
		http.Redirect(w, r, "/login?error="+err.Error(), http.StatusSeeOther)
		return
	}

	// Set HTTP-only cookie for web sessions
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    loginResponse.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	// Redirect to dashboard or intended page
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// HandleAPILogin handles JSON API login
func (h *AuthHandler) HandleAPILogin(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := loginReq.Validate(); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	loginResponse, err := h.processLogin(
		loginReq.Email, 
		loginReq.Password, 
		loginReq.RememberMe, 
		getClientIP(r), 
		r.UserAgent(),
	)
	if err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	web.WriteJSON(w, loginResponse, http.StatusOK)
}

// HandleLogout handles logout requests
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Extract token from request
	token, err := h.extractToken(r)
	if err == nil {
		// Verify and extract claims
		claims, err := h.jwtManager.VerifyToken(token)
		if err == nil {
			// Remove session
			h.sessionStore.DeleteSession(claims.SessionID)
		}
	}

	// Clear auth cookie
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/login?message=Logged out successfully", http.StatusSeeOther)
}

// HandleAPILogout handles API logout
func (h *AuthHandler) HandleAPILogout(w http.ResponseWriter, r *http.Request) {
	// Extract token from request
	token, err := h.extractToken(r)
	if err == nil {
		// Verify and extract claims
		claims, err := h.jwtManager.VerifyToken(token)
		if err == nil {
			// Remove session
			h.sessionStore.DeleteSession(claims.SessionID)
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	}
	web.WriteJSON(w, response, http.StatusOK)
}

// HandleAPIRefreshToken handles token refresh
func (h *AuthHandler) HandleAPIRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := h.extractToken(r)
	if err != nil {
		web.WriteJSONError(w, "Token required", http.StatusUnauthorized)
		return
	}

	newToken, err := h.jwtManager.RefreshToken(token)
	if err != nil {
		web.WriteJSONError(w, "Token refresh failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"token": newToken,
	}
	web.WriteJSON(w, response, http.StatusOK)
}

// HandleAPIGetProfile returns the current user's profile
func (h *AuthHandler) HandleAPIGetProfile(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r.Context())
	if userCtx == nil {
		web.WriteJSONError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetByID(userCtx.UserID)
	if err != nil {
		web.WriteJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	profile := map[string]interface{}{
		"userId":    user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"roles":     userCtx.Roles,
		"status":    user.Status,
		"accreditation": map[string]interface{}{
			"status":    user.Accreditation.Status,
			"type":      user.Accreditation.Type,
			"validUntil": user.Accreditation.ValidUntil,
		},
		"compliance": map[string]interface{}{
			"overallStatus": user.Compliance.OverallStatus,
			"riskScore":     user.Compliance.RiskScore,
		},
		"capabilities": map[string]interface{}{
			"canTrade":     user.CanTrade(),
			"isAccredited": user.IsAccredited(),
			"isCompliant":  user.IsCompliant(),
		},
	}

	web.WriteJSON(w, profile, http.StatusOK)
}

// processLogin handles the core login logic
func (h *AuthHandler) processLogin(email, password string, rememberMe bool, ipAddress, userAgent string) (*LoginResponse, error) {
	// Get user by email
	user, err := h.userService.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is suspended
	if user.Status == users.UserSuspended {
		return nil, errors.New("account is suspended")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Create session
	session, err := h.sessionStore.CreateSession(user.ID)
	if err != nil {
		return nil, errors.New("failed to create session")
	}

	// Update session with client info
	h.sessionStore.UpdateSessionActivity(session.ID, ipAddress, userAgent)

	// Determine user roles
	roles := GetUserRoles("client", string(user.Compliance.OverallStatus), string(user.Accreditation.Status))

	// Generate JWT token
	tokenDuration := 1 * time.Hour
	if rememberMe {
		tokenDuration = 30 * 24 * time.Hour // 30 days
	}

	// Temporarily extend JWT manager duration for this token
	originalDuration := h.jwtManager.tokenDuration
	h.jwtManager.tokenDuration = tokenDuration
	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, roles, session.ID)
	h.jwtManager.tokenDuration = originalDuration

	if err != nil {
		h.sessionStore.DeleteSession(session.ID)
		return nil, errors.New("failed to generate token")
	}

	return &LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int(tokenDuration.Seconds()),
		User: UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Roles:     roles,
		},
	}, nil
}

// extractToken extracts JWT token from request
func (h *AuthHandler) extractToken(r *http.Request) (string, error) {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return ExtractTokenFromHeader(authHeader)
	}

	// Try cookie as fallback
	cookie, err := r.Cookie("auth_token")
	if err == nil {
		return cookie.Value, nil
	}

	return "", errors.New("no token found")
}

// Request/Response types

// LoginRequest represents a login request
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string   `json:"token"`
	TokenType string   `json:"tokenType"`
	ExpiresIn int      `json:"expiresIn"`
	User      UserInfo `json:"user"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
}