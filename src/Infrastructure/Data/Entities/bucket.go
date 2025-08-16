package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Bucket represents the bucket entity in the database
type Bucket struct {
	Id          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string       `gorm:"uniqueIndex;not null" json:"name"`
	Description string       `json:"description"`
	OwnerId     uuid.UUID    `gorm:"type:uuid;not null;index" json:"owner_id"`
	Owner       User         `gorm:"foreignKey:OwnerId" json:"owner,omitempty"`
	AuthRule    AuthRule     `gorm:"embedded;embeddedPrefix:auth_" json:"auth_rule"`
	Settings    BucketSettings `gorm:"embedded;embeddedPrefix:settings_" json:"settings"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Navigation properties
	Files       []File       `gorm:"foreignKey:BucketId" json:"files,omitempty"`
}

// AuthRule represents authentication rules embedded in bucket
type AuthRule struct {
	Type    string                 `gorm:"not null;default:'jwt'" json:"type"`
	Enabled bool                   `gorm:"not null;default:true" json:"enabled"`
	Config  datatypes.JSON `gorm:"type:jsonb" json:"config"`
}

// BucketSettings represents bucket configuration embedded in bucket
type BucketSettings struct {
	MaxFileSize         int64    `gorm:"not null;default:0" json:"max_file_size"`
	MaxTotalSize        int64    `gorm:"not null;default:0" json:"max_total_size"`
	AllowedMimeTypes    []string `gorm:"type:text[]" json:"allowed_mime_types"`
	BlockedMimeTypes    []string `gorm:"type:text[]" json:"blocked_mime_types"`
	AllowedExtensions   []string `gorm:"type:text[]" json:"allowed_extensions"`
	BlockedExtensions   []string `gorm:"type:text[]" json:"blocked_extensions"`
	MaxFilesPerBucket   int64    `gorm:"not null;default:0" json:"max_files_per_bucket"`
	PublicRead          bool     `gorm:"not null;default:false" json:"public_read"`
	Versioning          bool     `gorm:"not null;default:false" json:"versioning"`
	Encryption          bool     `gorm:"not null;default:false" json:"encryption"`
	AllowOverwrite      bool     `gorm:"not null;default:true" json:"allow_overwrite"`
	RequireContentType  bool     `gorm:"not null;default:false" json:"require_content_type"`
}

// BeforeCreate is a GORM hook that runs before creating a Bucket record
func (b *Bucket) BeforeCreate(tx *gorm.DB) error {
	// Ensure ID is nil to allow auto-generation by PostgreSQL
	if b.Id == uuid.Nil {
		// For PostgreSQL with gen_random_uuid(), we explicitly exclude the ID from the query
		// This prevents the "00000000-0000-0000-0000-000000000000" from being inserted
		tx.Statement.Omit("id", "Id")
	}
	return nil
}

