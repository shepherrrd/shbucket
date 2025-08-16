package bucket

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

type DeleteBucketCommand struct {
	BucketID uuid.UUID `json:"bucket_id"`
	UserID   uuid.UUID `json:"user_id"`
}

type DeleteBucketResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteBucketRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewDeleteBucketRequestHandler(dbContext *persistence.AppDbContext) *DeleteBucketRequestHandler {
	return &DeleteBucketRequestHandler{
		dbContext: dbContext,
	}
}

func (h *DeleteBucketRequestHandler) Handle(ctx context.Context, command *DeleteBucketCommand) (*DeleteBucketResponse, error) {
	// Find the bucket using GoNtext static typing
	bucket, err := h.dbContext.Buckets.Where(&entities.Bucket{Id: command.BucketID}).FirstOrDefault()
	if err != nil || bucket == nil {
		return nil, fmt.Errorf("bucket not found")
	}

	// Check authorization 
	if bucket.OwnerId != command.UserID { // Fixed field name
		return nil, fmt.Errorf("unauthorized: only bucket owner can delete bucket")
	}

	// Check if bucket has files using GoNtext static typing
	fileCount, err := h.dbContext.Files.Where(&entities.File{BucketId: command.BucketID}).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket files: %w", err)
	}

	if fileCount > 0 {
		return nil, fmt.Errorf("cannot delete bucket: bucket contains %d files", fileCount)
	}

	// Delete bucket using GoNtext
	h.dbContext.Buckets.Remove(*bucket)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to delete bucket: %w", err)
	}

	return &DeleteBucketResponse{
		Success: true,
		Message: "Bucket deleted successfully",
	}, nil
}