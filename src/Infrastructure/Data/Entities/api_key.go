package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type APIKey struct {
	Id          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	KeyHash     string         `gorm:"not null;unique" json:"key_hash"`
	KeyPrefix   string         `gorm:"not null" json:"key_prefix"` 
	UserId      uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	Permissions datatypes.JSON `gorm:"type:jsonb" json:"permissions"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	LastUsed    *time.Time     `json:"last_used"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserId" json:"user,omitempty"`
}

type APIKeyPermission struct {
	Read     bool     `json:"read"`
	Write    bool     `json:"write"`
	SignURLs bool     `json:"sign_urls"`  
	Buckets  []string `json:"buckets,omitempty"` 
}

