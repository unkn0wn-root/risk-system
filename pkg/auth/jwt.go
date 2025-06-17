package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
	issuer        string
}

type Claims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IssuedAt int64    `json:"iat"`
	jwt.RegisteredClaims
}

// UserRole represents user roles
type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleAdmin     UserRole = "admin"
	RoleService   UserRole = "service"
	RoleModerator UserRole = "moderator"
)

func NewJWTManager(secretKey string, tokenDuration time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
		issuer:        issuer,
	}
}

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

func (c *Claims) HasRole(role UserRole) bool {
	for _, r := range c.Roles {
		if r == string(role) {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (c *Claims) HasAnyRole(roles ...UserRole) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

func GenerateSecretKey() string {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
