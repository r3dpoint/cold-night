package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// ContextKey represents a key for storing values in request context
type ContextKey string

const (
	// UserContextKey is the key for storing user information in context
	UserContextKey ContextKey = "user"
	// SessionContextKey is the key for storing session information in context
	SessionContextKey ContextKey = "session"
)

// UserContext represents user information stored in request context
type UserContext struct {
	UserID    string   `json:"userId"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	SessionID string   `json:"sessionId"`
}

// AuthMiddleware provides authentication and authorization middleware
type AuthMiddleware struct {
	jwtManager    *JWTManager
	rbac          *RBAC
	blacklist     TokenBlacklist
	excludePaths  map[string]bool
	sessionStore  SessionStore
}

// SessionStore interface for managing user sessions
type SessionStore interface {
	GetSession(sessionID string) (*Session, error)
	CreateSession(userID string) (*Session, error)
	DeleteSession(sessionID string) error
	IsSessionValid(sessionID string) bool
	UpdateSessionActivity(sessionID, ipAddress, userAgent string) error
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	Active    bool      `json:"active"`
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *JWTManager, rbac *RBAC, blacklist TokenBlacklist, sessionStore SessionStore) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:   jwtManager,
		rbac:         rbac,
		blacklist:    blacklist,
		sessionStore: sessionStore,
		excludePaths: map[string]bool{
			"/health":        true,
			"/login":         true,
			"/register":      true,
			"/api/health":    true,
			"/api/login":     true,
			"/api/register":  true,
			"/static/":       true,
			"/favicon.ico":   true,
		},
	}
}

// AddExcludePath adds a path to be excluded from authentication
func (m *AuthMiddleware) AddExcludePath(path string) {
	m.excludePaths[path] = true
}

// AuthenticateMiddleware verifies JWT tokens and sets user context
func (m *AuthMiddleware) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should be excluded from authentication
		if m.shouldExcludePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from header or cookie
		token, err := m.extractToken(r)
		if err != nil {
			m.writeUnauthorizedError(w, "Missing or invalid authorization token")
			return
		}

		// Verify JWT token
		claims, err := m.jwtManager.VerifyToken(token)
		if err != nil {
			m.writeUnauthorizedError(w, "Invalid token: "+err.Error())
			return
		}

		// Check if token is blacklisted
		if m.blacklist != nil && m.blacklist.IsBlacklisted(claims.ID) {
			m.writeUnauthorizedError(w, "Token has been revoked")
			return
		}

		// Verify session is still valid
		if m.sessionStore != nil && !m.sessionStore.IsSessionValid(claims.SessionID) {
			m.writeUnauthorizedError(w, "Session is no longer valid")
			return
		}

		// Create user context
		userCtx := &UserContext{
			UserID:    claims.UserID,
			Email:     claims.Email,
			Roles:     claims.Roles,
			SessionID: claims.SessionID,
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission creates middleware that requires specific permissions
func (m *AuthMiddleware) RequirePermission(permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserFromContext(r.Context())
			if userCtx == nil {
				m.writeForbiddenError(w, "User not authenticated")
				return
			}

			if !m.rbac.HasPermission(userCtx.Roles, permission) {
				m.writeForbiddenError(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func (m *AuthMiddleware) RequireAnyPermission(permissions []Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserFromContext(r.Context())
			if userCtx == nil {
				m.writeForbiddenError(w, "User not authenticated")
				return
			}

			if !m.rbac.HasAnyPermission(userCtx.Roles, permissions) {
				m.writeForbiddenError(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates middleware that requires specific roles
func (m *AuthMiddleware) RequireRole(roles []Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserFromContext(r.Context())
			if userCtx == nil {
				m.writeForbiddenError(w, "User not authenticated")
				return
			}

			hasRole := false
			for _, role := range roles {
				for _, userRole := range userCtx.Roles {
					if userRole == string(role) {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				m.writeForbiddenError(w, "Insufficient role permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireResourceAccess creates middleware for resource-based access control
func (m *AuthMiddleware) RequireResourceAccess(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserFromContext(r.Context())
			if userCtx == nil {
				m.writeForbiddenError(w, "User not authenticated")
				return
			}

			err := m.rbac.CanAccessResource(userCtx.UserID, userCtx.Roles, resource, action)
			if err != nil {
				m.writeForbiddenError(w, err.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func (m *AuthMiddleware) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides basic rate limiting
func (m *AuthMiddleware) RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter (use Redis in production)
	clients := make(map[string][]time.Time)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()
			
			// Clean old requests
			var validRequests []time.Time
			if requests, exists := clients[clientIP]; exists {
				for _, requestTime := range requests {
					if now.Sub(requestTime) < time.Minute {
						validRequests = append(validRequests, requestTime)
					}
				}
			}
			
			// Check rate limit
			if len(validRequests) >= requestsPerMinute {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded",
				})
				return
			}
			
			// Add current request
			validRequests = append(validRequests, now)
			clients[clientIP] = validRequests
			
			next.ServeHTTP(w, r)
		})
	}
}

// Helper methods

// shouldExcludePath checks if a path should be excluded from authentication
func (m *AuthMiddleware) shouldExcludePath(path string) bool {
	// Exact match
	if m.excludePaths[path] {
		return true
	}
	
	// Prefix match for static files
	for excludePath := range m.excludePaths {
		if strings.HasSuffix(excludePath, "/") && strings.HasPrefix(path, excludePath) {
			return true
		}
	}
	
	return false
}

// extractToken extracts JWT token from request
func (m *AuthMiddleware) extractToken(r *http.Request) (string, error) {
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

// writeUnauthorizedError writes a 401 error response
func (m *AuthMiddleware) writeUnauthorizedError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// writeForbiddenError writes a 403 error response
func (m *AuthMiddleware) writeForbiddenError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// getClientIP extracts client IP address from request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check for X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Use remote address
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	
	return ip
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) *UserContext {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok {
		return nil
	}
	return user
}

// GetSessionFromContext extracts session information from request context
func GetSessionFromContext(ctx context.Context) *Session {
	session, ok := ctx.Value(SessionContextKey).(*Session)
	if !ok {
		return nil
	}
	return session
}