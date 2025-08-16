package entities

import (
	"time"
	"github.com/google/uuid"
)

// StorageNode represents the storage node entity in the database
type StorageNode struct {
	Id            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string     `gorm:"not null" json:"name"`
	URL           string     `gorm:"not null;unique" json:"url"`
	AuthKey       string     `gorm:"not null" json:"-"` // Hidden from JSON for security
	IsActive      bool       `gorm:"not null;default:true" json:"is_active"`
	IsHealthy     bool       `gorm:"not null;default:false" json:"is_healthy"` // Start as unhealthy until first ping
	Priority      int        `gorm:"not null;default:0" json:"priority"`
	MaxStorage    int64      `gorm:"not null;default:0" json:"max_storage"`
	UsedStorage   int64      `gorm:"not null;default:0" json:"used_storage"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastPing      *time.Time `json:"last_ping,omitempty"`
}