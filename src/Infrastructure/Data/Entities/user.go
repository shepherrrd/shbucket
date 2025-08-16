package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the user entity in the database
type User struct {
	Id           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid();column:Id" json:"id"`
	Username     string     `gorm:"uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Role         string     `gorm:"not null;default:'viewer'" json:"role"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	PhoneNumber  *string    `gorm:"size:20" json:"phone_number,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;old_name:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastLoginTime    *time.Time `gorm:"old_name:last_login" json:"last_login"`
	
	// Navigation properties
	Buckets  []Bucket  `gorm:"foreignKey:OwnerId" json:"buckets,omitempty"`
	Sessions []Session `gorm:"foreignKey:UserId" json:"sessions,omitempty"`
}

// BeforeCreate is a GORM hook that runs before creating a User record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Ensure ID is nil to allow auto-generation by PostgreSQL
	if u.Id == uuid.Nil {
		// For PostgreSQL with gen_random_uuid(), we explicitly exclude the ID from the query
		// This prevents the "00000000-0000-0000-0000-000000000000" from being inserted
		tx.Statement.Omit("id", "Id")
	}
	return nil
}
