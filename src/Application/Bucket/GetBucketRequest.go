package bucket

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
	"shbucket/src/Utils"
)

type GetBucketCommand struct {
	BucketID uuid.UUID `json:"bucket_id"`
}

type GetBucketResponse struct {
	Bucket  models.BucketResponse `json:"bucket"`
	Success bool                  `json:"success"`
	Message string                `json:"message"`
}

type GetBucketRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewGetBucketRequestHandler(dbContext *persistence.AppDbContext) *GetBucketRequestHandler {
	return &GetBucketRequestHandler{
		dbContext: dbContext,
	}
}

func (h *GetBucketRequestHandler) Handle(ctx context.Context, command *GetBucketCommand) (*GetBucketResponse, error) {
	// Find bucket using GoNtext static typing - like GORM: Where(&Bucket{Id: command.BucketID})
	bucket, err := h.dbContext.Buckets.Where(&entities.Bucket{Id: command.BucketID}).FirstOrDefault()
	if err != nil || bucket == nil {
		return nil, fmt.Errorf("bucket not found")
	}

	// Get total files count using GoNtext static typing
	totalFiles, err := h.dbContext.Files.Where(&entities.File{BucketId: command.BucketID}).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to get file count: %w", err)
	}
	
	// Get total size - we'll need to use raw query for SUM operation for now

	totalSize, err := h.dbContext.Files.Where(&entities.File{BucketId: command.BucketID}).Sum(&entities.File{Size: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}
	bucketResponse := models.BucketResponse{
		ID:          bucket.Id,
		Name:        bucket.Name,
		Description: bucket.Description,
		OwnerID:     bucket.OwnerId,
		AuthRule: models.AuthRuleResponse{
			Type:    bucket.AuthRule.Type,
			Enabled: bucket.AuthRule.Enabled,
			Config:  utils.ConvertJSONToMap(bucket.AuthRule.Config),
		},
		Settings: models.BucketSettingsResponse{
			MaxFileSize:         bucket.Settings.MaxFileSize,
			MaxTotalSize:        bucket.Settings.MaxTotalSize,
			AllowedMimeTypes:    bucket.Settings.AllowedMimeTypes,
			BlockedMimeTypes:    bucket.Settings.BlockedMimeTypes,
			AllowedExtensions:   bucket.Settings.AllowedExtensions,
			BlockedExtensions:   bucket.Settings.BlockedExtensions,
			MaxFilesPerBucket:   bucket.Settings.MaxFilesPerBucket,
			PublicRead:          bucket.Settings.PublicRead,
			Versioning:          bucket.Settings.Versioning,
			Encryption:          bucket.Settings.Encryption,
			AllowOverwrite:      bucket.Settings.AllowOverwrite,
			RequireContentType:  bucket.Settings.RequireContentType,
		},
		Stats: models.BucketStatsResponse{
			TotalFiles: totalFiles,
			TotalSize:  int64(totalSize),
		},
		CreatedAt: bucket.CreatedAt,
		UpdatedAt: bucket.UpdatedAt,
	}

	return &GetBucketResponse{
		Bucket:  bucketResponse,
		Success: true,
		Message: "Bucket retrieved successfully",
	}, nil
}