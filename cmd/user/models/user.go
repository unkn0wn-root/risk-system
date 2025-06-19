package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a system user with authentication and profile information.
// includes security features like password hashing, roles, and verification status.
type User struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"not null"` // Never include in JSON
	FirstName    string     `json:"first_name" gorm:"not null"`
	LastName     string     `json:"last_name" gorm:"not null"`
	Phone        string     `json:"phone"`
	Roles        []string   `json:"roles" gorm:"serializer:json"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	IsVerified   bool       `json:"is_verified" gorm:"default:false"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SetPassword securely hashes and stores a user's password using bcrypt.
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies a plaintext password against the stored hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// HasRole checks if the user has a specific role assigned.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// AddRole assigns a new role to the user if not already present.
// prevents duplicate roles by checking existence before adding.
func (u *User) AddRole(role string) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, role)
	}
}

// RemoveRole removes a specific role from the user's role list.
// performs an in-place removal and stops after finding the first match.
func (u *User) RemoveRole(role string) {
	for i, r := range u.Roles {
		if r == role {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)
			break
		}
	}
}

// GetFullName returns the user's complete name by combining first and last names.
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}
