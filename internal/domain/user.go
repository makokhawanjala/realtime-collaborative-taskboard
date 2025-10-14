// internal/domain/user.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the payload for creating a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
}

// UpdateUserRequest represents the payload for updating a user
type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// Validate validates the CreateUserRequest
func (r *CreateUserRequest) Validate() error {
	if r.Username == "" {
		return ErrInvalidInput
	}
	if r.Email == "" {
		return ErrInvalidInput
	}
	return nil
}

// Validate validates the UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	if r.Username == "" && r.Email == "" {
		return ErrInvalidInput
	}
	return nil
}
