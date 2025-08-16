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

type ListBucketsCommand struct {
	UserID uuid.UUID `json:"user_id"`
	Page   int       `json:"page"`
	Limit  int       `json:"limit"`
}

type ListBucketsResponse struct {
	Buckets []models.BucketResponse `json:"buckets"`
	Total   int64                   `json:"total"`
	Page    int                     `json:"page"`
	Limit   int                     `json:"limit"`
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
}

type ListBucketsRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewListBucketsRequestHandler(dbContext *persistence.AppDbContext) *ListBucketsRequestHandler {
	return &ListBucketsRequestHandler{
		dbContext: dbContext,
	}
}

func (h *ListBucketsRequestHandler) Handle(ctx context.Context, command *ListBucketsCommand) (*ListBucketsResponse, error) {
	page := command.Page
	limit := command.Limit
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get total count using GoNtext
	total, err := h.dbContext.Buckets.Where(&entities.Bucket{OwnerId: command.UserID}).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count buckets: %w", err)
	}

	// Get buckets using GoNtext
	buckets, err := h.dbContext.Buckets.Where(&entities.Bucket{OwnerId: command.UserID}).
		Skip(offset).Take(limit).ToList()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch buckets: %w", err)
	}

	bucketResponses := make([]models.BucketResponse, len(buckets))
	for i, bucket := range buckets {
		// Get file count using GoNtext
		totalFiles, _ := h.dbContext.Files.Where(&entities.File{BucketId: bucket.Id}).Count()
		
		// Get total size using raw query for now (SUM not implemented in GoNtext yet)
		totalSize,err  := h.dbContext.Files.SumField("Size")
		if err != nil {
			return nil, fmt.Errorf("failed to get total size: %w", err)
		}
		bucketResponses[i] = models.BucketResponse{
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
	}

	return &ListBucketsResponse{
		Buckets: bucketResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Success: true,
		Message: "Buckets retrieved successfully",
	}, nil
}