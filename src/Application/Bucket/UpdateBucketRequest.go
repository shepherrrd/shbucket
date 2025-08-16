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

type UpdateBucketCommand struct {
	BucketID    uuid.UUID                        `json:"bucket_id"`
	UserID      uuid.UUID                        `json:"user_id"`
	Description *string                          `json:"description,omitempty" validate:"omitempty,max=500"`
	AuthRule    *models.AuthRuleResponse         `json:"auth_rule,omitempty"`
	Settings    *models.BucketSettingsResponse   `json:"settings,omitempty"`
}

type UpdateBucketResponse struct {
	Bucket  models.BucketResponse `json:"bucket"`
	Success bool                  `json:"success"`
	Message string                `json:"message"`
}

type UpdateBucketRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewUpdateBucketRequestHandler(dbContext *persistence.AppDbContext) *UpdateBucketRequestHandler {
	return &UpdateBucketRequestHandler{
		dbContext: dbContext,
	}
}

func (h *UpdateBucketRequestHandler) Handle(ctx context.Context, command *UpdateBucketCommand) (*UpdateBucketResponse, error) {
	// Get existing bucket
	bucketPtr, err := h.dbContext.Buckets.Where(&entities.Bucket{
		Id: command.BucketID,
		OwnerId: command.UserID,
	}).FirstOrDefault()
	if err != nil || bucketPtr == nil {
		return nil, fmt.Errorf("bucket not found or access denied")
	}

	bucket := *bucketPtr

	// Update description if provided
	if command.Description != nil {
		bucket.Description = *command.Description
	}

	// Update auth rule if provided
	if command.AuthRule != nil {
		bucket.AuthRule.Type = command.AuthRule.Type
		bucket.AuthRule.Enabled = command.AuthRule.Enabled
		if command.AuthRule.Config != nil {
			bucket.AuthRule.Config = utils.ConvertMapToJSON(command.AuthRule.Config)
		}
	}

	// Update settings if provided
	if command.Settings != nil {
		bucket.Settings.MaxFileSize = command.Settings.MaxFileSize
		bucket.Settings.MaxTotalSize = command.Settings.MaxTotalSize
		bucket.Settings.AllowedMimeTypes = command.Settings.AllowedMimeTypes
		bucket.Settings.BlockedMimeTypes = command.Settings.BlockedMimeTypes
		bucket.Settings.AllowedExtensions = command.Settings.AllowedExtensions
		bucket.Settings.BlockedExtensions = command.Settings.BlockedExtensions
		bucket.Settings.MaxFilesPerBucket = command.Settings.MaxFilesPerBucket
		bucket.Settings.PublicRead = command.Settings.PublicRead
		bucket.Settings.Versioning = command.Settings.Versioning
		bucket.Settings.Encryption = command.Settings.Encryption
		bucket.Settings.AllowOverwrite = command.Settings.AllowOverwrite
		bucket.Settings.RequireContentType = command.Settings.RequireContentType
	}

	// Save changes
	h.dbContext.Buckets.Update(bucket)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to update bucket: %w", err)
	}

	// Return response
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
		CreatedAt: bucket.CreatedAt,
		UpdatedAt: bucket.UpdatedAt,
	}

	return &UpdateBucketResponse{
		Bucket:  bucketResponse,
		Success: true,
		Message: "Bucket updated successfully",
	}, nil
}