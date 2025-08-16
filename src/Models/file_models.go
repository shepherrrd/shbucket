package models

import (
	"time"
	"github.com/google/uuid"
)

// FileMetadata model for API responses
type FileMetadataResponse struct {
	ContentType        string                 `json:"content_type"`
	ContentEncoding    string                 `json:"content_encoding,omitempty"`
	ContentDisposition string                 `json:"content_disposition,omitempty"`
	CacheControl       string                 `json:"cache_control,omitempty"`
	CustomMetadata     map[string]interface{} `json:"custom_metadata,omitempty"`
}

// File response model
type FileResponse struct {
	ID           uuid.UUID             `json:"id"`
	BucketID     uuid.UUID             `json:"bucket_id"`
	Name         string                `json:"name"`
	OriginalName string                `json:"original_name"`
	Path         string                `json:"path"`
	Size         int64                 `json:"size"`
	MimeType     string                `json:"mime_type"`
	Checksum     string                `json:"checksum"`
	Version      int                   `json:"version"`
	AuthRule     *AuthRuleResponse     `json:"auth_rule,omitempty"`
	Metadata     FileMetadataResponse  `json:"metadata"`
	SecuredUrl   string                `json:"secured_url,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	AccessedAt   *time.Time            `json:"accessed_at,omitempty"`
}

// Upload file response schema
type UploadFileResponse struct {
	File    FileResponse `json:"file"`
	Success bool         `json:"success"`
	Message string       `json:"message"`
}

// Delete file response schema
type DeleteFileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// List files response schema
type ListFilesResponse struct {
	Files []FileResponse `json:"files"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// Update file auth request schema
type UpdateFileAuthRequest struct {
	AuthRule AuthRuleResponse `json:"auth_rule"`
}

// Update file auth response schema
type UpdateFileAuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Multipart upload models
type InitiateMultipartUploadRequest struct {
	FileName    string                 `json:"file_name" validate:"required"`
	ContentType string                 `json:"content_type" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type InitiateMultipartUploadResponse struct {
	UploadID   string `json:"upload_id"`
	BucketName string `json:"bucket_name"`
	FileName   string `json:"file_name"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

type UploadPartResponse struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

type PartInfo struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

type CompleteMultipartUploadRequest struct {
	Parts []PartInfo `json:"parts" validate:"required,dive"`
}

type CompleteMultipartUploadResponse struct {
	File    FileResponse `json:"file"`
	Success bool         `json:"success"`
	Message string       `json:"message"`
}

type AbortMultipartUploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListPartsResponse struct {
	UploadID   string     `json:"upload_id"`
	BucketName string     `json:"bucket_name"`
	FileName   string     `json:"file_name"`
	Parts      []PartInfo `json:"parts"`
}

// Signed URL models
type GenerateSignedURLRequest struct {
	BucketName string `json:"bucket_name" validate:"required"`
	FileName   string `json:"file_name" validate:"required"`
	Method     string `json:"method" validate:"required,oneof=GET POST PUT DELETE"`
	ExpiresIn  int64  `json:"expires_in" validate:"required,min=1,max=604800"` // max 7 days
}

type GenerateSignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
}