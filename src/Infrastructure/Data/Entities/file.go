package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// File represents the file entity in the database
type File struct {
	Id             uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid();column:Id" json:"id"`
	BucketId       uuid.UUID    `gorm:"type:uuid;not null;index" json:"bucket_id"`
	Bucket         Bucket       `gorm:"foreignKey:BucketId" json:"bucket,omitempty"`
	Name           string       `gorm:"not null" json:"name"`
	OriginalName   string       `gorm:"not null" json:"original_name"`
	Path           string       `gorm:"not null" json:"path"`
	Size           int64        `gorm:"not null" json:"size"`
	MimeType       string       `gorm:"not null" json:"mime_type"`
	Checksum       string       `gorm:"not null" json:"checksum"`
	Version        int          `gorm:"not null;default:1" json:"version"`
	AuthRule       AuthRule     `gorm:"embedded;embeddedPrefix:auth_" json:"auth_rule"`
	Metadata       FileMetadata `gorm:"embedded;embeddedPrefix:metadata_" json:"metadata"`
	UploadedBy     uuid.UUID    `gorm:"type:uuid;not null;index" json:"uploaded_by"`
	CreatedAt      time.Time    `gorm:"autoCreateTime" json:"created_at"`
	SecuredUrl     string 		`gorm:"not null" json:"secured_url"`
	UpdatedAt      time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	AccessedAt     *time.Time   `json:"accessed_at,omitempty"`
}

// FileMetadata represents file metadata embedded in file
type FileMetadata struct {
	ContentType        string                 `json:"content_type"`
	ContentEncoding    string                 `json:"content_encoding"`
	ContentDisposition string                 `json:"content_disposition"`
	CacheControl       string                 `json:"cache_control"`
	CustomMetadata     datatypes.JSON `gorm:"type:jsonb" json:"custom_metadata"`
}

// BeforeCreate is a GORM hook that runs before creating a File record
func (f *File) BeforeCreate(tx *gorm.DB) error {
	// Ensure ID is nil to allow auto-generation by PostgreSQL
	if f.Id == uuid.Nil {
		// For PostgreSQL with gen_random_uuid(), we explicitly exclude the ID from the query
		// This prevents the "00000000-0000-0000-0000-000000000000" from being inserted
		tx.Statement.Omit("id", "Id")
	}
	return nil
}