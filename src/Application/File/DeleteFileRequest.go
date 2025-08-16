package file

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

type DeleteFileCommand struct {
	FileID   uuid.UUID `json:"file_id"`
	BucketID uuid.UUID `json:"bucket_id"`
	UserID   uuid.UUID `json:"user_id"`
}

type DeleteFileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteFileRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewDeleteFileRequestHandler(dbContext *persistence.AppDbContext) *DeleteFileRequestHandler {
	return &DeleteFileRequestHandler{
		dbContext: dbContext,
	}
}

func (h *DeleteFileRequestHandler) Handle(ctx context.Context, command *DeleteFileCommand) (*DeleteFileResponse, error) {
	// Find file using GoNtext static typing
	file, err := h.dbContext.Files.Where(&entities.File{
		Id:       command.FileID,
		BucketId: command.BucketID,
	}).FirstOrDefault()
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found")
	}

	// Find bucket using GoNtext static typing
	bucket, err := h.dbContext.Buckets.Where(&entities.Bucket{Id: command.BucketID}).FirstOrDefault()
	if err != nil || bucket == nil {
		return nil, fmt.Errorf("bucket not found")
	}

	if bucket.OwnerId != command.UserID && file.UploadedBy != command.UserID {
		return nil, fmt.Errorf("unauthorized: insufficient permissions to delete file")
	}

	// Delete physical file from storage
	if err := h.deletePhysicalFile(file.Path); err != nil {
		return nil, fmt.Errorf("failed to delete physical file: %w", err)
	}

	// Delete from database using GoNtext
	h.dbContext.Files.Remove(*file)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to delete file record: %w", err)
	}

	return &DeleteFileResponse{
		Success: true,
		Message: "File deleted successfully",
	}, nil
}

func (h *DeleteFileRequestHandler) deletePhysicalFile(filePath string) error {
	// Check if file is stored on a remote node
	if strings.HasPrefix(filePath, "node://") {
		return h.deleteFromNode(filePath)
	}
	
	// Delete local file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, which is fine for our purposes
			return nil
		}
		return fmt.Errorf("failed to remove file: %w", err)
	}
	
	return nil
}

func (h *DeleteFileRequestHandler) deleteFromNode(filePath string) error {
	// Extract node ID and file path from node:// URL
	// Format: node://nodeID/bucketID/fileID
	pathParts := strings.Split(strings.TrimPrefix(filePath, "node://"), "/")
	if len(pathParts) < 3 {
		return fmt.Errorf("invalid node file path format: %s", filePath)
	}
	
	nodeID := pathParts[0]
	bucketIDStr := pathParts[1] 
	nodeFileID := pathParts[2]
	
	// Get bucket name from bucket ID
	bucketID, err := uuid.Parse(bucketIDStr)
	if err != nil {
		return fmt.Errorf("invalid bucket ID in path: %w", err)
	}
	
	bucket, err := h.dbContext.Buckets.First(&entities.Bucket{Id: bucketID})
	if err != nil {
		return fmt.Errorf("bucket not found: %w", err)
	}
	
	bucketName := bucket.Name
	
	// Get storage node info
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return fmt.Errorf("invalid node ID: %w", err)
	}
	
	storageNode, err := h.dbContext.StorageNodes.First(&entities.StorageNode{Id: nodeUUID})
	if err != nil {
		return fmt.Errorf("storage node not found: %w", err)
	}
	
	// Create DELETE request to the node's internal deletion endpoint
	req, err := http.NewRequest("DELETE", 
		fmt.Sprintf("%s/api/v1/internal/delete", storageNode.URL), 
		nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	
	// Add query parameters for the file to delete
	q := req.URL.Query()
	q.Add("bucket_name", bucketName)
	q.Add("file_name", nodeFileID)
	req.URL.RawQuery = q.Encode()
	
	// Add authentication header using the node's auth key
	req.Header.Set("Authorization", "Bearer "+storageNode.AuthKey)
	
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("node deletion failed with status: %d", resp.StatusCode)
	}
	
	return nil
}