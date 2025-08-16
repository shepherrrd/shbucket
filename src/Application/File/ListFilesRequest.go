package file

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
	"shbucket/src/Utils"
)

type ListFilesCommand struct {
	BucketID uuid.UUID `json:"bucket_id"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type ListFilesResponse struct {
	Files   []models.FileResponse `json:"files"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
	Success bool                  `json:"success"`
	Message string                `json:"message"`
}

type ListFilesRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewListFilesRequestHandler(dbContext *persistence.AppDbContext) *ListFilesRequestHandler {
	return &ListFilesRequestHandler{
		dbContext: dbContext,
	}
}

func (h *ListFilesRequestHandler) Handle(ctx context.Context, command *ListFilesCommand) (*ListFilesResponse, error) {
	page := command.Page
	limit := command.Limit
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	total, err := h.dbContext.Files.Where(&entities.File{BucketId: command.BucketID}).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count files: %w", err)
	}

	files, err := h.dbContext.Files.Where(&entities.File{BucketId: command.BucketID}).
		Skip(offset).Take(limit).ToList()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch files: %w", err)
	}

	fileResponses := make([]models.FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = models.FileResponse{
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
	}

	return &ListFilesResponse{
		Files:   fileResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Success: true,
		Message: "Files retrieved successfully",
	}, nil
}