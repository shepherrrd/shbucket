package models

import (
	"time"
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Data/Entities"
)

type APIKeyResponse struct {
	ID          uuid.UUID                   `json:"id"`
	Name        string                      `json:"name"`
	KeyPrefix   string                      `json:"key_prefix"`
	UserID      uuid.UUID                   `json:"user_id"`
	Username    string                      `json:"username"`
	IsActive    bool                        `json:"is_active"`
	Permissions entities.APIKeyPermission  `json:"permissions"`
	ExpiresAt   *time.Time                  `json:"expires_at"`
	LastUsed    *time.Time                  `json:"last_used"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}