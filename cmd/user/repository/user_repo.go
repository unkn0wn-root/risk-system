package repository

import (
	"user-risk-system/cmd/user/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository provides database operations for user entities.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository with the provided database connection.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database with auto-generated UUID.
// assigns a unique identifier before persisting the user record.
func (r *UserRepository) Create(user *models.User) error {
	user.ID = uuid.New().String()
	return r.db.Create(user).Error
}

// GetByID retrieves a user by their unique identifier.
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update modifies an existing user record in the database.
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete permanently removes a user from the database by ID.
func (r *UserRepository) Delete(id string) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// List retrieves multiple users with pagination support.
func (r *UserRepository) List(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}
