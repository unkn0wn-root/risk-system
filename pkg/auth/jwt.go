// Package auth provides JWT token management and user authentication capabilities.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token generation, validation, and refresh operations.
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
	issuer        string
}

// Claims represents the custom JWT claims structure containing user information.
// extends the standard JWT registered claims with user-specific data.
type Claims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IssuedAt int64    `json:"iat"`
	jwt.RegisteredClaims
}

// UserRole represents the different types of user roles in the system.
type UserRole string

// Predefined user roles for authorization purposes.
const (
	RoleUser      UserRole = "user"      // Standard user with basic permissions
	RoleAdmin     UserRole = "admin"     // Administrator with full system access
	RoleService   UserRole = "service"   // Service account for inter-service communication
	RoleModerator UserRole = "moderator" // Moderator with elevated permissions
)

// NewJWTManager creates a new JWT manager instance with the specified configuration.
func NewJWTManager(secretKey string, tokenDuration time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
		issuer:        issuer,
	}
}

// GenerateToken creates a new JWT token for the specified user with the given roles.
// The token includes standard claims (issuer, audience, expiration) and custom user data.
func (manager *JWTManager) GenerateToken(userID, email string, roles []string) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Roles:    roles,
		IssuedAt: now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    manager.issuer,
			Subject:   userID,
			Audience:  []string{"user-risk-system"},
			ExpiresAt: jwt.NewNumericDate(now.Add(manager.tokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// ValidateToken parses and validates a JWT token string, returning the claims if valid.
func (manager *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// RefreshToken generates a new token from an existing valid token if it's close to expiry.
func (manager *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiry (within 10 minutes)
	if time.Until(claims.ExpiresAt.Time) > 10*time.Minute {
		return "", fmt.Errorf("token is still valid, refresh not needed")
	}

	return manager.GenerateToken(claims.UserID, claims.Email, claims.Roles)
}

// HasRole checks if the user has the specified role in their claims.
func (c *Claims) HasRole(role UserRole) bool {
	for _, r := range c.Roles {
		if r == string(role) {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles.
func (c *Claims) HasAnyRole(roles ...UserRole) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// GenerateSecretKey creates a cryptographically secure random 256-bit secret key.
func GenerateSecretKey() string {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
