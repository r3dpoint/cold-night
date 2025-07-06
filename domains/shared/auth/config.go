package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Config holds authentication configuration
type Config struct {
	JWTSecret          string
	JWTIssuer          string
	JWTTokenDuration   time.Duration
	SessionDuration    time.Duration
	BCryptCost         int
	RateLimitRequests  int
	EnableRateLimit    bool
	EnableCORS         bool
	CookieSecure       bool
	CookieSameSite     string
}

// NewDefaultConfig creates a default authentication configuration
func NewDefaultConfig() *Config {
	return &Config{
		JWTSecret:          getEnvOrDefault("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTIssuer:          getEnvOrDefault("JWT_ISSUER", "securities-marketplace"),
		JWTTokenDuration:   parseDurationOrDefault(getEnvOrDefault("JWT_TOKEN_DURATION", "1h")),
		SessionDuration:    parseDurationOrDefault(getEnvOrDefault("SESSION_DURATION", "24h")),
		BCryptCost:         parseIntOrDefault(getEnvOrDefault("BCRYPT_COST", "12")),
		RateLimitRequests:  parseIntOrDefault(getEnvOrDefault("RATE_LIMIT_REQUESTS", "100")),
		EnableRateLimit:    parseBoolOrDefault(getEnvOrDefault("ENABLE_RATE_LIMIT", "true")),
		EnableCORS:         parseBoolOrDefault(getEnvOrDefault("ENABLE_CORS", "true")),
		CookieSecure:       parseBoolOrDefault(getEnvOrDefault("COOKIE_SECURE", "false")),
		CookieSameSite:     getEnvOrDefault("COOKIE_SAME_SITE", "Strict"),
	}
}

// AuthManager provides a unified interface for authentication operations
type AuthManager struct {
	Config       *Config
	JWTManager   *JWTManager
	RBAC         *RBAC
	SessionStore SessionStore
	Blacklist    TokenBlacklist
	Middleware   *AuthMiddleware
}

// NewAuthManager creates a new authentication manager with all components
func NewAuthManager(config *Config) *AuthManager {
	// Create JWT manager
	jwtManager := NewJWTManager(
		config.JWTSecret,
		config.JWTIssuer,
		config.JWTTokenDuration,
	)

	// Create RBAC
	rbac := NewRBAC()

	// Create session store
	sessionStore := NewInMemorySessionStore()

	// Create token blacklist
	blacklist := NewInMemoryBlacklist()

	// Create middleware
	middleware := NewAuthMiddleware(jwtManager, rbac, blacklist, sessionStore)

	return &AuthManager{
		Config:       config,
		JWTManager:   jwtManager,
		RBAC:         rbac,
		SessionStore: sessionStore,
		Blacklist:    blacklist,
		Middleware:   middleware,
	}
}

// CreateAuthHandler creates an authentication handler
func (am *AuthManager) CreateAuthHandler(userService UserService) *AuthHandler {
	return NewAuthHandler(am.JWTManager, am.SessionStore, userService, am.RBAC)
}

// Utility functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDurationOrDefault(value string) time.Duration {
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	return 1 * time.Hour // default
}

func parseIntOrDefault(value string) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return 0
}

func parseBoolOrDefault(value string) bool {
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return false
}

// Helper functions for password hashing

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string, cost int) (string, error) {
	if cost == 0 {
		cost = 12 // default cost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a password with its hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePasswordStrength checks if password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*':
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character (!@#$%^&*)")
	}
	
	return nil
}