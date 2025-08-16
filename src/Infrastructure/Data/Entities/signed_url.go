package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SignedURL represents the signed URL entity in the database
type SignedURL struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Signature  string    `gorm:"uniqueIndex;not null" json:"signature"`
	BucketName string    `gorm:"not null" json:"bucket_name"`
	FileName   string    `gorm:"not null" json:"file_name"`
	Method     string    `gorm:"not null" json:"method"`
	ExpiresAt  time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	SingleUse  bool      `gorm:"not null;default:false" json:"single_use"`
	Used       bool      `gorm:"not null;default:false" json:"used"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
}

// BeforeCreate is a GORM hook that runs before creating a SignedURL record
func (s *SignedURL) BeforeCreate(tx *gorm.DB) error {
	// ALWAYS force auto-generation by omitting the ID field
	tx.Statement.Omit("id", "ID")
	
	// Reset the ID to nil to ensure auto-generation
	s.ID = uuid.Nil
	
	return nil
}