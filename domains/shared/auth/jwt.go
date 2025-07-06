package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     []byte
	issuer        string
	tokenDuration time.Duration
}

// Claims represents the JWT claims
type Claims struct {
	UserID    string   `json:"userId"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	SessionID string   `json:"sessionId"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, issuer string, tokenDuration time.Duration) *JWTManager {
	// If no secret key provided, generate a random one (for development)
	var key []byte
	if secretKey == "" {
		key = make([]byte, 32)
		rand.Read(key)
	} else {
		key = []byte(secretKey)
	}

	return &JWTManager{
		secretKey:     key,
		issuer:        issuer,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken creates a new JWT token for a user
func (manager *JWTManager) GenerateToken(userID, email string, roles []string, sessionID string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    manager.issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(manager.secretKey)
}

// VerifyToken validates and parses a JWT token
func (manager *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return manager.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (manager *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := manager.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiry (within 15 minutes)
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > 15*time.Minute {
		return "", errors.New("token is not eligible for refresh yet")
	}

	// Generate new token with same claims but updated expiration
	return manager.GenerateToken(claims.UserID, claims.Email, claims.Roles, claims.SessionID)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// TokenBlacklist interface for token revocation
type TokenBlacklist interface {
	IsBlacklisted(tokenID string) bool
	BlacklistToken(tokenID string, expiry time.Time) error
}

// InMemoryBlacklist is a simple in-memory token blacklist implementation
type InMemoryBlacklist struct {
	tokens map[string]time.Time
}

// NewInMemoryBlacklist creates a new in-memory blacklist
func NewInMemoryBlacklist() *InMemoryBlacklist {
	return &InMemoryBlacklist{
		tokens: make(map[string]time.Time),
	}
}

// IsBlacklisted checks if a token is blacklisted
func (b *InMemoryBlacklist) IsBlacklisted(tokenID string) bool {
	expiry, exists := b.tokens[tokenID]
	if !exists {
		return false
	}

	// Remove expired entries
	if time.Now().After(expiry) {
		delete(b.tokens, tokenID)
		return false
	}

	return true
}

// BlacklistToken adds a token to the blacklist
func (b *InMemoryBlacklist) BlacklistToken(tokenID string, expiry time.Time) error {
	b.tokens[tokenID] = expiry
	return nil
}

// CleanupExpiredTokens removes expired tokens from the blacklist
func (b *InMemoryBlacklist) CleanupExpiredTokens() {
	now := time.Now()
	for tokenID, expiry := range b.tokens {
		if now.After(expiry) {
			delete(b.tokens, tokenID)
		}
	}
}