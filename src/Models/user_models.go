package models

import (
	"time"
	"github.com/google/uuid"
)

// User response model
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// Login request schema
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// Login response schema
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// Change password request schema
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// Change password response schema
type ChangePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Create user request schema
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin manager editor viewer"`
}

// Create user response schema
type CreateUserResponse struct {
	User    UserResponse `json:"user"`
	Success bool         `json:"success"`
	Message string       `json:"message"`
}

// List users response schema
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// Logout response schema
type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}