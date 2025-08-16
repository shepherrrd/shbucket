package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents the user session entity in the database
type Session struct {
	Id        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid();column:Id" json:"id"`
	UserId    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserId" json:"user,omitempty"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	LastUsed  time.Time `gorm:"autoUpdateTime" json:"last_used"`
}

// BeforeCreate is a GORM hook that runs before creating a Session record
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	// ALWAYS force auto-generation by omitting the ID field, regardless of its current value
	// This prevents any UUID that might have been set by other mechanisms from being used
	tx.Statement.Omit("id", "Id")
	
	// Reset the ID to nil to ensure auto-generation
	s.Id = uuid.Nil
	
	return nil
}