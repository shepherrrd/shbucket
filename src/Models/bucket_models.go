package models

import (
	"time"
	"github.com/google/uuid"
)

// AuthRule model for API responses
type AuthRuleResponse struct {
	Type    string                 `json:"type"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

// BucketSettings model for API responses
type BucketSettingsResponse struct {
	MaxFileSize         int64    `json:"max_file_size"`
	MaxTotalSize        int64    `json:"max_total_size"`
	AllowedMimeTypes    []string `json:"allowed_mime_types"`
	BlockedMimeTypes    []string `json:"blocked_mime_types"`
	AllowedExtensions   []string `json:"allowed_extensions"`
	BlockedExtensions   []string `json:"blocked_extensions"`
	MaxFilesPerBucket   int64    `json:"max_files_per_bucket"`
	PublicRead          bool     `json:"public_read"`
	Versioning          bool     `json:"versioning"`
	Encryption          bool     `json:"encryption"`
	AllowOverwrite      bool     `json:"allow_overwrite"`
	RequireContentType  bool     `json:"require_content_type"`
}

// BucketStats model for API responses
type BucketStatsResponse struct {
	TotalFiles int64      `json:"total_files"`
	TotalSize  int64      `json:"total_size"`
	LastAccess *time.Time `json:"last_access,omitempty"`
}

// Bucket response model
type BucketResponse struct {
	ID          uuid.UUID               `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	OwnerID     uuid.UUID               `json:"owner_id"`
	AuthRule    AuthRuleResponse        `json:"auth_rule"`
	Settings    BucketSettingsResponse  `json:"settings"`
	Stats       BucketStatsResponse     `json:"stats"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// Create bucket request schema
type CreateBucketRequest struct {
	Name        string                  `json:"name" validate:"required,min=3,max=63,alphanum"`
	Description string                  `json:"description" validate:"max=500"`
	AuthRule    AuthRuleResponse        `json:"auth_rule"`
	Settings    BucketSettingsResponse  `json:"settings"`
}

// Create bucket response schema
type CreateBucketResponse struct {
	BucketID uuid.UUID `json:"bucket_id"`
	Name     string    `json:"name"`
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
}

// Update bucket request schema
type UpdateBucketRequest struct {
	Description *string                  `json:"description,omitempty" validate:"omitempty,max=500"`
	AuthRule    *AuthRuleResponse        `json:"auth_rule,omitempty"`
	Settings    *BucketSettingsResponse  `json:"settings,omitempty"`
}

// Update bucket response schema
type UpdateBucketResponse struct {
	Bucket  BucketResponse `json:"bucket"`
	Success bool           `json:"success"`
	Message string         `json:"message"`
}

// Delete bucket response schema
type DeleteBucketResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// List buckets response schema
type ListBucketsResponse struct {
	Buckets []BucketResponse `json:"buckets"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}