package bucket

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
	"shbucket/src/Utils"
)

type CreateBucketCommand struct {
	OwnerID     uuid.UUID               `json:"owner_id"`
	Name        string                  `json:"name" validate:"required,min=3,max=63,alphanum"`
	Description string                  `json:"description" validate:"max=500"`
	AuthRule    models.AuthRuleResponse `json:"auth_rule"`
	Settings    models.BucketSettingsResponse `json:"settings"`
}

type CreateBucketResponse struct {
	Bucket  models.BucketResponse `json:"bucket"`
	Success bool                  `json:"success"`
	Message string                `json:"message"`
}

type CreateBucketRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewCreateBucketRequestHandler(dbContext *persistence.AppDbContext) *CreateBucketRequestHandler {
	return &CreateBucketRequestHandler{
		dbContext: dbContext,
	}
}

func (h *CreateBucketRequestHandler) Handle(ctx context.Context, command *CreateBucketCommand) (*CreateBucketResponse, error) {
	// Check if bucket with this name already exists using static typing
	existingBucket, err := h.dbContext.Buckets.Where(&entities.Bucket{Name: command.Name}).FirstOrDefault()
	if err == nil && existingBucket != nil {
		return nil, fmt.Errorf("bucket with name '%s' already exists", command.Name)
	}

	// Set default auth rule if not provided
	defaultConfig := make(map[string]interface{})
	configJSON, err := json.Marshal(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}
	
	authRule := entities.AuthRule{
		Type:    "none",
		Enabled: false,
		Config:  datatypes.JSON(configJSON),
	}
	
	if command.AuthRule.Type != "" {
		authRule.Type = command.AuthRule.Type
		authRule.Enabled = command.AuthRule.Enabled
		if command.AuthRule.Config != nil {
			authRule.Config = utils.ConvertMapToJSON(command.AuthRule.Config)
		}
	}

	// Set default settings if not provided
	settings := entities.BucketSettings{
		MaxFileSize:         100 * 1024 * 1024, // 100MB default
		MaxTotalSize:        10 * 1024 * 1024 * 1024, // 10GB default
		AllowedMimeTypes:    []string{},
		BlockedMimeTypes:    []string{},
		AllowedExtensions:   []string{},
		BlockedExtensions:   []string{},
		MaxFilesPerBucket:   10000,
		PublicRead:          false,
		Versioning:          false,
		Encryption:          false,
		AllowOverwrite:      true,
		RequireContentType:  false,
	}

	// Override with provided settings
	if command.Settings.MaxFileSize > 0 {
		settings.MaxFileSize = command.Settings.MaxFileSize
	}
	if command.Settings.MaxTotalSize > 0 {
		settings.MaxTotalSize = command.Settings.MaxTotalSize
	}
	if command.Settings.AllowedMimeTypes != nil {
		settings.AllowedMimeTypes = command.Settings.AllowedMimeTypes
	}
	if command.Settings.BlockedMimeTypes != nil {
		settings.BlockedMimeTypes = command.Settings.BlockedMimeTypes
	}
	if command.Settings.AllowedExtensions != nil {
		settings.AllowedExtensions = command.Settings.AllowedExtensions
	}
	if command.Settings.BlockedExtensions != nil {
		settings.BlockedExtensions = command.Settings.BlockedExtensions
	}
	if command.Settings.MaxFilesPerBucket > 0 {
		settings.MaxFilesPerBucket = command.Settings.MaxFilesPerBucket
	}
	settings.PublicRead = command.Settings.PublicRead
	settings.Versioning = command.Settings.Versioning
	settings.Encryption = command.Settings.Encryption
	settings.AllowOverwrite = command.Settings.AllowOverwrite
	settings.RequireContentType = command.Settings.RequireContentType

	bucket := &entities.Bucket{
		Name:        command.Name,
		Description: command.Description,
		OwnerId:     command.OwnerID, // Fixed field name
		AuthRule:    authRule,
		Settings:    settings,
	}

	// Add bucket using GoNtext
	h.dbContext.Buckets.Add(*bucket)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	bucketResponse := models.BucketResponse{
		ID:          bucket.Id,      // Fixed field name
		Name:        bucket.Name,
		Description: bucket.Description,
		OwnerID:     bucket.OwnerId, // Fixed field name
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
			TotalFiles: 0,
			TotalSize:  0,
		},
		CreatedAt: bucket.CreatedAt,
		UpdatedAt: bucket.UpdatedAt,
	}

	return &CreateBucketResponse{
		Bucket:  bucketResponse,
		Success: true,
		Message: "Bucket created successfully",
	}, nil
}