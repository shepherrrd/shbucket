package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shbucket/src/Application/File"
	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Mediator"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Infrastructure/Services"
)

type FileController struct {
	mediator            *mediator.Mediator
	validator           *validator.Validate
	authService         *auth.AuthorizationService
	dbContext           *persistence.AppDbContext
	signatureService    *services.SignatureValidationService
}

func NewFileController(mediator *mediator.Mediator, validator *validator.Validate, authService *auth.AuthorizationService, dbContext *persistence.AppDbContext) *FileController {
	return &FileController{
		mediator:         mediator,
		validator:        validator,
		authService:      authService,
		dbContext:        dbContext,
		signatureService: services.NewSignatureValidationService(dbContext),
	}
}

//	@Summary		Upload file to bucket
//	@Description	Upload a file to the specified bucket with authentication
//	@Tags			files
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			bucketId	path		string							true	"Bucket ID"
//	@Param			file		formData	file							true	"File to upload"
//	@Success		201			{object}	file.DistributedUploadResponse	"File uploaded successfully"
//	@Failure		400			{object}	map[string]string				"Bad request"
//	@Failure		401			{object}	map[string]string				"Unauthorized"
//	@Router			/buckets/{bucketId}/files [post]
func (ctrl *FileController) UploadFile(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	bucketIDParam := c.Params("bucketId")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}
	
	fileReader, err := fileHeader.Open()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer fileReader.Close()
	
	// Use distributed upload by default
	command := &file.DistributedUploadCommand{
		BucketID:    bucketID,
		File:        fileHeader,
		FileReader:  fileReader,
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		UploadedBy:  userContext.UserID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	uploadFileResponse := response.(*file.DistributedUploadResponse)
	return c.Status(http.StatusCreated).JSON(uploadFileResponse)
}

//	@Summary		Delete file from bucket
//	@Description	Delete a specific file from a bucket
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			bucketId	path		string	true	"Bucket ID"
//	@Param			fileId		path		string	true	"File ID"
//	@Success		200			{object}	file.DeleteFileResponse	"File deleted successfully"
//	@Failure		400			{object}	map[string]string		"Bad request"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		404			{object}	map[string]string		"File not found"
//	@Router			/buckets/{bucketId}/files/{fileId} [delete]
func (ctrl *FileController) DeleteFile(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	bucketIDParam := c.Params("bucketId")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}
	
	command := &file.DeleteFileCommand{
		FileID:   fileID,
		BucketID: bucketID,
		UserID:   userContext.UserID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	deleteFileResponse := response.(*file.DeleteFileResponse)
	return c.JSON(deleteFileResponse)
}

//	@Summary		Get file metadata
//	@Description	Get metadata and information about a specific file
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			bucketId	path		string	true	"Bucket ID"
//	@Param			fileId		path		string	true	"File ID"
//	@Success		200			{object}	file.GetFileResponse	"File metadata retrieved successfully"
//	@Failure		400			{object}	map[string]string		"Bad request"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		404			{object}	map[string]string		"File not found"
//	@Router			/buckets/{bucketId}/files/{fileId}/info [get]
func (ctrl *FileController) GetFile(c *fiber.Ctx) error {
	bucketIDParam := c.Params("bucketId")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}
	
	command := &file.GetFileCommand{
		FileID:   fileID,
		BucketID: bucketID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	getFileResponse := response.(*file.GetFileResponse)
	return c.JSON(getFileResponse)
}

//	@Summary		Serve file content
//	@Description	Serve file content directly with support for signed URLs, API keys, and image processing
//	@Tags			files
//	@Accept			json
//	@Produce		application/octet-stream
//	@Produce		image/jpeg
//	@Produce		image/png
//	@Param			bucketId	path		string	true	"Bucket ID"
//	@Param			fileId		path		string	true	"File ID"
//	@Param			signature	query		string	false	"Signed URL signature for temporary access"
//	@Param			width		query		int		false	"Image width for scaling (images only)"
//	@Param			height		query		int		false	"Image height for scaling (images only)"
//	@Param			quality		query		int		false	"Image quality for JPEG compression"	default(85)
//	@Param			resolution	query		string	false	"Predefined resolution (144p, 240p, 360p, 480p, 720p, 1080p, 1440p, 2160p, 4k)"
//	@Success		200			"File content served successfully"
//	@Failure		400			{object}	map[string]string		"Bad request"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		404			{object}	map[string]string		"File not found"
//	@Router			/file/{bucketId}/{fileId} [get]
func (ctrl *FileController) ServeFile(c *fiber.Ctx) error {
	
	bucketIDParam := c.Params("bucketId")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}
	
	// First get file metadata to check access rules
	command := &file.GetFileCommand{
		FileID:   fileID,
		BucketID: bucketID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	getFileResponse := response.(*file.GetFileResponse)
	fileInfo := getFileResponse.File
	
	// Get bucket information to check public_read setting using static typing
	bucket, err := ctrl.dbContext.Buckets.First(&entities.Bucket{Id: bucketID})
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Bucket not found",
		})
	}
	
	// Check if authentication is required
	// public_read: true means files can be read without authentication
	// public_read: false means authentication is required for reading
	requiresAuth := !bucket.Settings.PublicRead
	
	if requiresAuth {
		// Check for API key or signed URL
		apiKey := c.Get("X-API-Key")
		signedToken := c.Query("signature")
		
		if signedToken != "" {
			// Validate signature and mark as used if single-use (simple approach)
			signedURL, err := ctrl.signatureService.ValidateSignatureOnly(signedToken)
			if err != nil {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired signed URL",
				})
			}
			
			// If it's single-use, mark as used on first access
			if signedURL.SingleUse && !signedURL.Used {
				if err := ctrl.signatureService.MarkSignatureAsUsed(signedToken); err != nil {
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to mark signature as used",
					})
				}
			}
		} else if apiKey != "" {
			// Validate API key
			if !ctrl.validateAPIKey(apiKey, bucketID) {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired API key",
				})
			}
		} else {
			// Check JWT auth as fallback
			_, err := ctrl.authService.AuthorizeRequest(c)
			if err != nil {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Authentication required. Use API key, signed URL, or JWT token.",
				})
			}
		}
	}
	
	// Check for image scaling parameters
	width, _ := strconv.Atoi(c.Query("width", "0"))
	height, _ := strconv.Atoi(c.Query("height", "0"))
	quality, _ := strconv.Atoi(c.Query("quality", "85"))
	
	// Check for resolution presets
	resolution := c.Query("resolution")
	if resolution != "" {
		switch strings.ToLower(resolution) {
		case "144p":
			height = 144
			width = 0 // Maintain aspect ratio
		case "240p":
			height = 240
			width = 0
		case "360p":
			height = 360
			width = 0
		case "480p":
			height = 480
			width = 0
		case "720p":
			height = 720
			width = 0
		case "1080p":
			height = 1080
			width = 0
		case "1440p":
			height = 1440
			width = 0
		case "2160p", "4k":
			height = 2160
			width = 0
		}
	}
	
	// Check if this is an image and scaling is requested
	isImage := strings.HasPrefix(fileInfo.MimeType, "image/")
	needsProcessing := isImage && (width > 0 || height > 0 || resolution != "" || quality != 85)
	
	if needsProcessing {
		// Process the image
		processedImage, outputMimeType, err := ctrl.processImage(fileInfo.Path, fileInfo.MimeType, width, height, quality)
		if err != nil {
			// Fallback to serving original file
			needsProcessing = false
		} else {
			// Set headers for processed image
			c.Set("Content-Type", outputMimeType)
			c.Set("Content-Length", fmt.Sprintf("%d", len(processedImage)))
			c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileInfo.Name))
			
			// Set cache headers based on access level
			if requiresAuth {
				c.Set("Cache-Control", "private, no-cache")
			} else {
				c.Set("Cache-Control", "public, max-age=3600") // Cache processed images for 1 hour
			}
			
			// Send processed image
			return c.Send(processedImage)
		}
	}
	
	// Send original file (either not an image, no scaling requested, or processing failed)
	c.Set("Content-Type", fileInfo.MimeType)
	c.Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileInfo.Name))
	
	if requiresAuth {
		c.Set("Cache-Control", "private, no-cache")
	} else {
		c.Set("Cache-Control", "public, max-age=31536000")
	}
	
	// Check if file is stored on a node (path starts with "node://")
	if strings.HasPrefix(fileInfo.Path, "node://") {
		// Extract node ID from path: node://nodeID/bucketID/fileID
		pathParts := strings.Split(strings.TrimPrefix(fileInfo.Path, "node://"), "/")
		if len(pathParts) >= 3 {
			nodeID := pathParts[0]
			// pathParts[1] is bucketID, pathParts[2] is fileID
			
			// Fetch file from storage node
			fileData, err := ctrl.fetchFileFromNode(nodeID, bucketID, fileID, fileInfo.Name)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to fetch file from storage node: %v", err),
				})
			}
			
			return c.Send(fileData)
		}
	}
	
	return c.SendFile(fileInfo.Path)
}


//	@Summary		List files in bucket
//	@Description	Get a list of all files in a specific bucket
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			bucketId	path		string	true	"Bucket ID"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			limit		query		int		false	"Items per page"	default(10)
//	@Success		200			{object}	file.ListFilesResponse	"Files retrieved successfully"
//	@Failure		400			{object}	map[string]string		"Bad request"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		404			{object}	map[string]string		"Bucket not found"
//	@Router			/buckets/{bucketId}/files [get]
func (ctrl *FileController) ListFiles(c *fiber.Ctx) error {
	bucketIDParam := c.Params("bucketId")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	command := &file.ListFilesCommand{
		BucketID: bucketID,
		Page:     page,
		Limit:    limit,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	listFilesResponse := response.(*file.ListFilesResponse)
	return c.JSON(listFilesResponse)
}

//	@Summary		Generate signed URL for file
//	@Description	Generate a temporary signed URL for secure file access with optional single-use functionality
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			bucketId	path		string	true	"Bucket ID"
//	@Param			fileId		path		string	true	"File ID"
//	@Param			request		body		object	true	"Signed URL generation parameters"	example({"expires_in":3600,"single_use":false})
//	@Success		200			{object}	file.GenerateSignedURLResponse	"Signed URL generated successfully"
//	@Failure		400			{object}	map[string]string				"Bad request"
//	@Failure		401			{object}	map[string]string				"Unauthorized"
//	@Failure		404			{object}	map[string]string				"File not found"
//	@Router			/buckets/{bucketId}/files/{fileId}/signed-url [post]
func (ctrl *FileController) GenerateSignedURL(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	bucketIDParam := c.Params("bucketId")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}
	
	var request struct {
		ExpiresIn int  `json:"expires_in" validate:"required,min=60,max=604800"` // 1 minute to 7 days
		SingleUse bool `json:"single_use"`                                        // Optional single-use checkbox
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	if err := ctrl.validator.Struct(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	command := &file.GenerateSignedURLCommand{
		BucketID:  bucketID,
		FileID:    fileID,
		ExpiresIn: request.ExpiresIn,
		UserID:    userContext.UserID,
		SingleUse: request.SingleUse,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	signedURLResponse := response.(*file.GenerateSignedURLResponse)
	return c.JSON(signedURLResponse)
}

// validateAPIKey validates an API key and checks permissions
func (ctrl *FileController) validateAPIKey(apiKey string, bucketID uuid.UUID) bool {
	// Hash the provided API key
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])
	
	// Find API key in database using GoNtext
	dbAPIKey, err := ctrl.dbContext.APIKeys.Where(&entities.APIKey{KeyHash: keyHash, IsActive: true}).FirstOrDefault()
	if err != nil || dbAPIKey == nil {
		return false
	}
	
	// Check if API key has expired
	if dbAPIKey.ExpiresAt != nil && dbAPIKey.ExpiresAt.Before(time.Now()) {
		return false
	}
	
	// Check bucket permissions (if specific buckets are specified)
	var permissions entities.APIKeyPermission
	if err := json.Unmarshal(dbAPIKey.Permissions, &permissions); err != nil {
		return false
	}
	
	// If buckets array is specified, check if this bucket is allowed
	if len(permissions.Buckets) > 0 {
		bucketAllowed := false
		for _, allowedBucket := range permissions.Buckets {
			if allowedBucket == bucketID.String() {
				bucketAllowed = true
				break
			}
		}
		if !bucketAllowed {
			return false
		}
	}
	
	// Check if API key has read permission
	return permissions.Read
}


// processImage processes an image file with scaling parameters
func (ctrl *FileController) processImage(filePath, mimeType string, width, height, quality int) ([]byte, string, error) {
	// Open the image file
	src, err := imaging.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open image: %w", err)
	}

	// Get original dimensions
	bounds := src.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate scaling dimensions
	if width == 0 && height == 0 {
		// No scaling requested, return original
		width = originalWidth
		height = originalHeight
	} else if width == 0 && height > 0 {
		// Scale by height, maintain aspect ratio
		aspectRatio := float64(originalWidth) / float64(originalHeight)
		width = int(float64(height) * aspectRatio)
	} else if height == 0 && width > 0 {
		// Scale by width, maintain aspect ratio
		aspectRatio := float64(originalHeight) / float64(originalWidth)
		height = int(float64(width) * aspectRatio)
	}
	// If both width and height are specified, use them as-is (may distort image)
	
	// Ensure minimum dimensions
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	// Only scale if dimensions are different
	var processed image.Image = src
	if width != originalWidth || height != originalHeight {
		processed = imaging.Resize(src, width, height, imaging.Lanczos)
	}

	// Encode to bytes
	var buf []byte
	var outputMimeType string

	// Determine output format and quality
	if quality == 0 {
		quality = 85 // Default quality
	}

	// Convert to JPEG for better compression if quality is specified or if scaling was applied
	if strings.Contains(strings.ToLower(mimeType), "png") && (width == originalWidth && height == originalHeight) {
		// Keep as PNG if no scaling and original is PNG
		buf, err = encodePNG(processed)
		outputMimeType = "image/png"
	} else {
		// Convert to JPEG for scaling or if quality parameter is used
		buf, err = encodeJPEG(processed, quality)
		outputMimeType = "image/jpeg"
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to encode image: %w", err)
	}

	return buf, outputMimeType, nil
}

// encodeJPEG encodes an image to JPEG with specified quality
func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	buf := make([]byte, 0)
	w := &bytesWriter{buf: &buf}
	
	options := &jpeg.Options{Quality: quality}
	err := jpeg.Encode(w, img, options)
	return buf, err
}

//	@Summary		Internal upload for distributed storage
//	@Description	Receives files from master node for storage on this node
//	@Tags			files
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		Bearer
//	@Param			file		formData	file	true	"File to upload"
//	@Param			bucket_id	formData	string	true	"Bucket ID"
//	@Param			file_id		formData	string	true	"File ID"
//	@Param			filename	formData	string	true	"Original filename"
//	@Success		200			{object}	map[string]interface{}	"Upload successful"
//	@Failure		400			{object}	map[string]string		"Bad request"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Router			/internal/upload [post]
func (ctrl *FileController) InternalUpload(c *fiber.Ctx) error {
	// Validate node auth key from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing Authorization header",
		})
	}
	
	// Extract Bearer token (auth key)
	var authKey string
	if strings.HasPrefix(authHeader, "Bearer ") {
		authKey = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Authorization header format",
		})
	}
	
	// Validate auth key against node setup config
	nodeConfig, err := ctrl.dbContext.SetupConfigs.Where(&entities.SetupConfig{SetupType: "node"}).FirstOrDefault()
	if err != nil || nodeConfig == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Node configuration not found",
		})
	}
	
	// Parse ConfigData JSON to get node_auth_key
	var configData map[string]interface{}
	if err := json.Unmarshal(nodeConfig.ConfigData, &configData); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse node configuration",
		})
	}
	
	nodeAuthKey, ok := configData["node_auth_key"].(string)
	if !ok || nodeAuthKey == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Node auth key not found in configuration",
		})
	}
	
	if nodeAuthKey != authKey {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid auth key",
		})
	}
	
	// Get file from multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Get metadata from form
	bucketID := c.FormValue("bucket_id")
	bucketName := c.FormValue("bucket_name")
	fileID := c.FormValue("file_id")
	filename := c.FormValue("filename")

	if bucketID == "" || bucketName == "" || fileID == "" || filename == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required metadata (bucket_id, bucket_name, file_id, filename)",
		})
	}

	// Use the same nodeConfig for storage path
	storagePath := nodeConfig.StoragePath
	if storagePath == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Storage path not configured in node config",
		})
	}

	// Create bucket directory using bucket name from form
	storageDir := fmt.Sprintf("%s/%s", storagePath, bucketName)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create storage directory",
		})
	}

	// Save file to local storage using node's configured path - just use fileID
	filePath := fmt.Sprintf("%s/%s", storageDir, fileID)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Parse UUIDs for database record
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid file ID format",
		})
	}

	bucketUUID, err := uuid.Parse(bucketID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid bucket ID format",
		})
	}

	// Create a file metadata record using GoNtext entity
	nodeMetadata := entities.NodeFileMetadata{
		Id:         fileUUID,
		BucketId:   bucketUUID,
		BucketName: bucketName,
		Filename:   filename,
		Path:       filePath,
		Size:       file.Size,
		CreatedAt:  time.Now(),
	}
	
	ctrl.dbContext.NodeFileMetadata.Add(nodeMetadata)
	if err := ctrl.dbContext.SaveChanges(); err != nil {
		// Log error but don't fail the upload since file is already saved
		log.Printf("Warning: Failed to create file metadata record: %v", err)
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "File uploaded successfully to storage node",
		"file_path": filePath,
		"file_size": file.Size,
	})
}

// encodePNG encodes an image to PNG
func encodePNG(img image.Image) ([]byte, error) {
	buf := make([]byte, 0)
	w := &bytesWriter{buf: &buf}
	
	err := png.Encode(w, img)
	return buf, err
}

// bytesWriter implements io.Writer for writing to a byte slice
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// fetchFileFromNode retrieves a file from a storage node
func (ctrl *FileController) fetchFileFromNode(nodeID string, bucketID uuid.UUID, fileID uuid.UUID, filename string) ([]byte, error) {
	// Get storage node info
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, fmt.Errorf("invalid node ID: %w", err)
	}
	
	storageNode, err := ctrl.dbContext.StorageNodes.First(&entities.StorageNode{Id: nodeUUID})
	if err != nil {
		return nil, fmt.Errorf("storage node not found: %w", err)
	}
	
	// Use the storage node's auth key (the master uses this to authenticate with the node)
	nodeAuthKey := storageNode.AuthKey
	
	// Create request to fetch file from node using internal endpoint
	req, err := http.NewRequest("GET", 
		fmt.Sprintf("%s/api/v1/internal/file", storageNode.URL), 
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add query parameters for the file to fetch
	q := req.URL.Query()
	q.Add("bucket_id", bucketID.String())
	q.Add("file_id", fileID.String())
	q.Add("filename", filename)
	req.URL.RawQuery = q.Encode()
	
	// Add authentication header using the node's auth key
	req.Header.Set("Authorization", "Bearer "+nodeAuthKey)
	
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node returned status: %d", resp.StatusCode)
	}
	
	// Read file data
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}
	
	return fileData, nil
}

//	@Summary		Internal delete for distributed storage
//	@Description	Deletes files from this storage node
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			bucket_name	query	string	true	"Bucket name"
//	@Param			file_name	query	string	true	"File name to delete"
//	@Success		200			{object}	map[string]interface{}	"Delete successful"
//	@Failure		400			{object}	map[string]string		"Bad request"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Router			/internal/delete [delete]
func (ctrl *FileController) InternalDelete(c *fiber.Ctx) error {
	// Validate node auth key from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing Authorization header",
		})
	}
	
	// Extract Bearer token (auth key)
	var authKey string
	if strings.HasPrefix(authHeader, "Bearer ") {
		authKey = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Authorization header format",
		})
	}
	
	// Validate auth key against node setup config
	nodeConfig, err := ctrl.dbContext.SetupConfigs.Where(&entities.SetupConfig{SetupType: "node"}).FirstOrDefault()
	if err != nil || nodeConfig == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Node configuration not found",
		})
	}
	
	// Parse ConfigData JSON to get node_auth_key
	var configData map[string]interface{}
	if err := json.Unmarshal(nodeConfig.ConfigData, &configData); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse node configuration",
		})
	}
	
	nodeAuthKey, ok := configData["node_auth_key"].(string)
	if !ok || nodeAuthKey == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Node auth key not found in configuration",
		})
	}
	
	if nodeAuthKey != authKey {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid auth key",
		})
	}

	// Get query parameters
	bucketName := c.Query("bucket_name")
	fileName := c.Query("file_name")

	if bucketName == "" || fileName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required parameters (bucket_name, file_name)",
		})
	}

	// Use the same nodeConfig for storage path
	storagePath := nodeConfig.StoragePath
	if storagePath == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Storage path not configured in node config",
		})
	}

	// Construct file path: storage_path/bucket_name/file_name
	filePath := fmt.Sprintf("%s/%s/%s", storagePath, bucketName, fileName)
	
	// Delete the file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, which is fine
			return c.JSON(fiber.Map{
				"success": true,
				"message": "File already deleted or does not exist",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete file",
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "File deleted successfully from storage node",
		"file_path": filePath,
	})
}

//	@Summary		Internal file serving for distributed storage
//	@Description	Serves files directly from this storage node
//	@Tags			files
//	@Accept			json
//	@Produce		application/octet-stream
//	@Security		Bearer
//	@Param			bucket_id	query	string	true	"Bucket ID"
//	@Param			file_id		query	string	true	"File ID"
//	@Param			filename	query	string	true	"Filename"
//	@Success		200			"File content"
//	@Failure		400			{object}	map[string]string	"Bad request"
//	@Failure		401			{object}	map[string]string	"Unauthorized"
//	@Failure		404			{object}	map[string]string	"File not found"
//	@Router			/internal/file [get]
func (ctrl *FileController) InternalFile(c *fiber.Ctx) error {
	// Validate node auth key from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing Authorization header",
		})
	}
	
	// Extract Bearer token (auth key)
	var authKey string
	if strings.HasPrefix(authHeader, "Bearer ") {
		authKey = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Authorization header format",
		})
	}
	
	// Validate auth key against node setup config
	nodeConfig, err := ctrl.dbContext.SetupConfigs.Where(&entities.SetupConfig{SetupType: "node"}).FirstOrDefault()
	if err != nil || nodeConfig == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Node configuration not found",
		})
	}
	
	// Parse ConfigData JSON to get node_auth_key
	var configData map[string]interface{}
	if err := json.Unmarshal(nodeConfig.ConfigData, &configData); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse node configuration",
		})
	}
	
	nodeAuthKey, ok := configData["node_auth_key"].(string)
	if !ok || nodeAuthKey == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Node auth key not found in configuration",
		})
	}
	
	if nodeAuthKey != authKey {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid auth key",
		})
	}

	// Get query parameters
	bucketID := c.Query("bucket_id")
	fileID := c.Query("file_id")
	filename := c.Query("filename")

	if bucketID == "" || fileID == "" || filename == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required parameters (bucket_id, file_id, filename)",
		})
	}

	// Parse file and bucket IDs
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID format",
		})
	}

	bucketUUID, err := uuid.Parse(bucketID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID format",
		})
	}

	// Look up file in node's metadata using GoNtext
	nodeMetadata, err := ctrl.dbContext.NodeFileMetadata.Where(&entities.NodeFileMetadata{
		Id:       fileUUID,
		BucketId: bucketUUID,
	}).FirstOrDefault()
	if err != nil || nodeMetadata == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "File not found in node metadata",
		})
	}

	// Check if file exists on disk
	if _, err := os.Stat(nodeMetadata.Path); os.IsNotExist(err) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "File not found on disk",
		})
	}

	// Serve the file directly using the path from metadata
	return c.SendFile(nodeMetadata.Path)
}