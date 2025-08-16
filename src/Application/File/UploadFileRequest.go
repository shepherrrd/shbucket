package file

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
	"shbucket/src/Utils"
)

type UploadFileCommand struct {
	BucketID     uuid.UUID             `json:"bucket_id"`
	File         *multipart.FileHeader `json:"-"`
	FileReader   io.Reader             `json:"-"`
	FileName     string                `json:"file_name"`
	ContentType  string                `json:"content_type"`
	Metadata     map[string]interface{} `json:"metadata"`
	UploadedBy   uuid.UUID             `json:"uploaded_by"`
}

type UploadFileResponse struct {
	File    models.FileResponse `json:"file"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
}

type UploadFileRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewUploadFileRequestHandler(dbContext *persistence.AppDbContext) *UploadFileRequestHandler {
	return &UploadFileRequestHandler{
		dbContext: dbContext,
	}
}

func (h *UploadFileRequestHandler) Handle(ctx context.Context, command *UploadFileCommand) (*UploadFileResponse, error) {
	bucketPtr, err := h.dbContext.Buckets.Where(&entities.Bucket{Id: command.BucketID}).FirstOrDefault()
	if err != nil || bucketPtr == nil {
		return nil, fmt.Errorf("bucket not found")
	}
	
	bucket := *bucketPtr

	fileSize := command.File.Size
	if bucket.Settings.MaxFileSize > 0 && fileSize > bucket.Settings.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	fileExtension := filepath.Ext(command.FileName)
	if len(bucket.Settings.AllowedExtensions) > 0 {
		allowed := false
		for _, ext := range bucket.Settings.AllowedExtensions {
			if ext == fileExtension {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("file extension not allowed")
		}
	}

	for _, ext := range bucket.Settings.BlockedExtensions {
		if ext == fileExtension {
			return nil, fmt.Errorf("file extension is blocked")
		}
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, command.FileReader); err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	file := &entities.File{
		Id:           uuid.Nil, // Explicitly set to nil to ensure auto-generation
		BucketId:     command.BucketID,
		Name:         command.FileName,
		OriginalName: command.FileName,
		Path:         fmt.Sprintf("/%s/%s", bucket.Name, command.FileName),
		Size:         fileSize,
		MimeType:     command.ContentType,
		Checksum:     checksum,
		Version:      1,
		AuthRule: entities.AuthRule{
			Type:    bucket.AuthRule.Type,
			Enabled: bucket.AuthRule.Enabled,
			Config:  bucket.AuthRule.Config,
		},
		Metadata: entities.FileMetadata{
			ContentType:        command.ContentType,
			ContentEncoding:    "",
			ContentDisposition: "",
			CacheControl:       "",
			CustomMetadata:     createJSONFromMetadata(command.Metadata),
		},
		UploadedBy: command.UploadedBy,
	}

	h.dbContext.Files.Add(*file)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	fileResponse := models.FileResponse{
		ID:           file.Id,
		BucketID:     file.BucketId,
		Name:         file.Name,
		OriginalName: file.OriginalName,
		Path:         file.Path,
		Size:         file.Size,
		MimeType:     file.MimeType,
		Checksum:     file.Checksum,
		Version:      file.Version,
		AuthRule: &models.AuthRuleResponse{
			Type:    file.AuthRule.Type,
			Enabled: file.AuthRule.Enabled,
			Config:  utils.ConvertJSONToMap(file.AuthRule.Config),
		},
		Metadata: models.FileMetadataResponse{
			ContentType:        file.Metadata.ContentType,
			ContentEncoding:    file.Metadata.ContentEncoding,
			ContentDisposition: file.Metadata.ContentDisposition,
			CacheControl:       file.Metadata.CacheControl,
			CustomMetadata:     utils.ConvertJSONToMap(file.Metadata.CustomMetadata),
		},
		SecuredUrl:  file.SecuredUrl,
		CreatedAt:  file.CreatedAt,
		UpdatedAt:  file.UpdatedAt,
		AccessedAt: file.AccessedAt,
	}

	return &UploadFileResponse{
		File:    fileResponse,
		Success: true,
		Message: "File uploaded successfully",
	}, nil
}

func createJSONFromMetadata(metadata map[string]interface{}) datatypes.JSON {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadataJSON, _ := json.Marshal(metadata)
	return datatypes.JSON(metadataJSON)
}