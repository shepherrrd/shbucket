package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
	"shbucket/src/Utils"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type DistributedUploadCommand struct {
	BucketID     uuid.UUID             `json:"bucket_id"`
	File         *multipart.FileHeader `json:"-"`
	FileReader   io.Reader             `json:"-"`
	FileName     string                `json:"file_name"`
	ContentType  string                `json:"content_type"`
	Metadata     map[string]interface{} `json:"metadata"`
	UploadedBy   uuid.UUID             `json:"uploaded_by"`
}

type DistributedUploadResponse struct {
	File       models.FileResponse `json:"file"`
	StorageNode *models.StorageNodeResponse `json:"storage_node,omitempty"`
	Success    bool                `json:"success"`
	Message    string              `json:"message"`
}

type DistributedUploadRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewDistributedUploadRequestHandler(dbContext *persistence.AppDbContext) *DistributedUploadRequestHandler {
	return &DistributedUploadRequestHandler{
		dbContext: dbContext,
	}
}

func (h *DistributedUploadRequestHandler) Handle(ctx context.Context, command *DistributedUploadCommand) (*DistributedUploadResponse, error) {
	// First check the master server storage
	masterConfig, err := h.dbContext.SetupConfigs.Where(&entities.SetupConfig{SetupType: "master"}).FirstOrDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to get master configuration: %w", err)
	}
	fileSize := command.File.Size
	
	// Check if master has enough space
	masterUsedStorage, err := h.dbContext.Files.SumField("Size")
	if err != nil {
		return nil, fmt.Errorf("failed to calculate used storage: %w", err)
	}

	masterFreeSpace := masterConfig.MaxStorage - int64(masterUsedStorage)
	storageNode := (*models.StorageNodeResponse)(nil)
	
	// Generate file ID for storage path
	fileID := uuid.New()
	
	if masterFreeSpace < fileSize {
		var availableNode entities.StorageNode
		availableNodePtr, err := h.dbContext.StorageNodes.Where(&entities.StorageNode{
			IsActive: true,
			IsHealthy: true,
		}).FirstOrDefault()
		if err != nil || availableNodePtr == nil {
			return nil, fmt.Errorf("upload failed: no active storage nodes available")
		}
		
		availableNode = *availableNodePtr
		
		// Check if node has enough space
		if availableNode.MaxStorage - availableNode.UsedStorage < fileSize {
			return nil, fmt.Errorf("upload failed: no storage space available. Master: %d bytes free, File: %d bytes", 
				masterFreeSpace, fileSize)
		}
		
		// Upload to the storage node
		success, err := h.uploadToNode(&availableNode, command, fileID)
		if err != nil {
			return nil, fmt.Errorf("failed to upload to storage node: %w", err)
		}
		
		if !success {
			return nil, fmt.Errorf("storage node rejected the upload")
		}
		
		// Update node storage usage
		availableNode.UsedStorage += fileSize
		h.dbContext.StorageNodes.Update(availableNode)
		h.dbContext.SaveChanges()
		
		storageNodeResponse := &models.StorageNodeResponse{
			ID:          availableNode.Id,
			Name:        availableNode.Name,
			URL:         availableNode.URL,
			MaxStorage:  availableNode.MaxStorage,
			UsedStorage: availableNode.UsedStorage,
			Priority:    availableNode.Priority,
			IsActive:    availableNode.IsActive,
			IsHealthy:   availableNode.IsHealthy,
			CreatedAt:   availableNode.CreatedAt,
			UpdatedAt:   availableNode.UpdatedAt,
			LastPing:    availableNode.LastPing,
		}
		storageNode = storageNodeResponse
	}
	
	// Create file record (stored on master regardless of where file is physically stored)
	bucketPtr, err := h.dbContext.Buckets.Where(&entities.Bucket{Id: command.BucketID}).FirstOrDefault()
	if err != nil || bucketPtr == nil {
		return nil, fmt.Errorf("bucket not found")
	}
	
	bucket := *bucketPtr
	
	// Save file to local storage if not uploaded to node
	var filePath string
	var checksum string
	
	if storageNode == nil {
		// Get master storage path from config
		// configData := utils.ConvertJSONToMap(masterConfig.ConfigData)
		storagePath  := masterConfig.StoragePath
		if storagePath == "" {
			return nil, fmt.Errorf("storage_path not configured in master config")
		}
		
		
		// Create bucket directory if it doesn't exist
		bucketDir := filepath.Join(storagePath, bucket.Name)
		if err := os.MkdirAll(bucketDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create bucket directory: %w", err)
		}
		
		// Set file path: storage_path/bucket_name/file_id
		filePath = filepath.Join(bucketDir, fileID.String())
		
		// Read file content for saving and checksum calculation
		fileContent, err := io.ReadAll(command.FileReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}
		
		// Calculate checksum
		hash := sha256.Sum256(fileContent)
		checksum = fmt.Sprintf("%x", hash)
		
		// Save file to disk
		if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
			return nil, fmt.Errorf("failed to save file to disk: %w", err)
		}
	} else {
		// File is stored on node, use bucket ID in path format: node://{nodeid}/{bucketid}/{fileid}
		filePath = fmt.Sprintf("node://%s/%s/%s", storageNode.ID.String(), command.BucketID.String(), fileID.String())
		checksum = "stored-on-node"
	}
	
	customMetadata := command.Metadata
	if customMetadata == nil {
		customMetadata = make(map[string]interface{})
	}
	
	if storageNode != nil {
		customMetadata["storage_node_id"] = storageNode.ID.String()
		customMetadata["storage_node_url"] = storageNode.URL
	}
	
	customMetadataJSON, err := json.Marshal(customMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal custom metadata: %w", err)
	}
	
	file := &entities.File{
		Id:           fileID, 
		BucketId:     command.BucketID,
		Name:         command.FileName,
		OriginalName: command.FileName,
		Path:         filePath,
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
			CustomMetadata:     datatypes.JSON(customMetadataJSON),
		},
		UploadedBy: command.UploadedBy,
		// CreatedAt and UpdatedAt are automatically set by GORM autoCreateTime/autoUpdateTime tags
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
	
	message := "File uploaded successfully to master"
	if storageNode != nil {
		message = fmt.Sprintf("File uploaded successfully to storage node: %s", storageNode.Name)
	}
	
	return &DistributedUploadResponse{
		File:        fileResponse,
		StorageNode: storageNode,
		Success:     true,
		Message:     message,
	}, nil
}

func (h *DistributedUploadRequestHandler) uploadToNode(node *entities.StorageNode, command *DistributedUploadCommand, fileID uuid.UUID) (bool, error) {
	// Create a buffer to store the file content for uploading to node
	fileContent, err := io.ReadAll(command.FileReader)
	if err != nil {
		return false, fmt.Errorf("failed to read file content: %w", err)
	}
	
	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add file field
	fileWriter, err := writer.CreateFormFile("file", command.FileName)
	if err != nil {
		return false, err
	}
	_, err = fileWriter.Write(fileContent)
	if err != nil {
		return false, err
	}
	
	// Get bucket name for the node
	bucket, err := h.dbContext.Buckets.First(&entities.Bucket{Id: command.BucketID})
	if err != nil {
		return false, fmt.Errorf("bucket not found: %w", err)
	}
	
	// Add metadata and required fields
	metadataJSON, _ := json.Marshal(command.Metadata)
	writer.WriteField("metadata", string(metadataJSON))
	writer.WriteField("content_type", command.ContentType)
	writer.WriteField("bucket_id", command.BucketID.String())
	writer.WriteField("bucket_name", bucket.Name)
	writer.WriteField("file_id", fileID.String())
	writer.WriteField("filename", command.FileName)
	
	writer.Close()
	
	// Create HTTP request to storage node
	req, err := http.NewRequest("POST", 
		fmt.Sprintf("%s/api/v1/internal/upload", node.URL), 
		body)
	if err != nil {
		return false, err
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+node.AuthKey)
	
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated, nil
}