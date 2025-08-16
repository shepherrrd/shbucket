package file

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
	"shbucket/src/Utils"
)

type GetFileCommand struct {
	FileID   uuid.UUID `json:"file_id"`
	BucketID uuid.UUID `json:"bucket_id"`
}

type GetFileResponse struct {
	File    models.FileResponse `json:"file"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
}

type GetFileRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewGetFileRequestHandler(dbContext *persistence.AppDbContext) *GetFileRequestHandler {
	return &GetFileRequestHandler{
		dbContext: dbContext,
	}
}

func (h *GetFileRequestHandler) Handle(ctx context.Context, command *GetFileCommand) (*GetFileResponse, error) {
	// Find file using GoNtext static typing
	file, err := h.dbContext.Files.Where(&entities.File{
		Id:       command.FileID,
		BucketId: command.BucketID,
	}).FirstOrDefault()
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if(file == nil) {
		return nil, fmt.Errorf("file not found")
	}
	now := time.Now()
	file.AccessedAt = &now
	h.dbContext.SaveChanges()

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

	return &GetFileResponse{
		File:    fileResponse,
		Success: true,
		Message: "File retrieved successfully",
	}, nil
}