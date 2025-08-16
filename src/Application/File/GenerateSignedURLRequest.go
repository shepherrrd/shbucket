package file

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Config"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

type GenerateSignedURLCommand struct {
	BucketID  uuid.UUID `json:"bucket_id" validate:"required"`
	FileID    uuid.UUID `json:"file_id" validate:"required"`
	ExpiresIn int       `json:"expires_in" validate:"required,min=60,max=604800"` // 1 minute to 7 days
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	SingleUse bool      `json:"single_use" validate:""` // Frontend checkbox for single-use URLs
}

type GenerateSignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
}

type GenerateSignedURLRequestHandler struct {
	dbContext *persistence.AppDbContext
	settings  *config.Settings
}

func NewGenerateSignedURLRequestHandler(dbContext *persistence.AppDbContext) *GenerateSignedURLRequestHandler {
	return &GenerateSignedURLRequestHandler{
		dbContext: dbContext,
		settings:  config.GetSettings(),
	}
}

func (h *GenerateSignedURLRequestHandler) Handle(ctx context.Context, command *GenerateSignedURLCommand) (*GenerateSignedURLResponse, error) {
	// Verify file exists and user has access using GoNtext
	file, err := h.dbContext.Files.Where(&entities.File{
		Id:       command.FileID,
		BucketId: command.BucketID,
	}).FirstOrDefault()
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found")
	}
	// Get bucket information for the URL
	bucket, err := h.dbContext.Buckets.Where(&entities.Bucket{Id: command.BucketID}).FirstOrDefault()
	if err != nil || bucket == nil {
		return nil, fmt.Errorf("bucket not found")
	}
	
	// Get signing secret from settings
	signingSecret := h.settings.SignatureSecret
	
	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(command.ExpiresIn) * time.Second)
	
	// Create signature payload (only bucketID:fileID - no expires, no user field)
	payload := fmt.Sprintf("%s:%s", 
		command.BucketID.String(), 
		command.FileID.String())
	
	// Generate HMAC signature
	signature := h.generateHMAC(payload, signingSecret)
	
	// Check if signature already exists in database
	existingSignedURL, err := h.dbContext.SignedURLs.Where(&entities.SignedURL{
		Signature: signature,
	}).FirstOrDefault()
	
	if err == nil && existingSignedURL != nil {
		// Signature exists, validate it's not expired
		if existingSignedURL.ExpiresAt.After(time.Now()) {
			// Return existing valid signed URL using file endpoint with signature parameter
			signedURL := fmt.Sprintf("%s/api/v1/file/%s/%s?signature=%s", 
				h.settings.BaseURL, 
				command.BucketID.String(), 
				command.FileID.String(), 
				signature)
			
			return &GenerateSignedURLResponse{
				URL:       signedURL,
				ExpiresAt: existingSignedURL.ExpiresAt,
				Success:   true,
				Message:   "Existing signed URL returned",
			}, nil
		} else {
			h.dbContext.SignedURLs.Remove(*existingSignedURL)
			if err := h.dbContext.SaveChanges(); err != nil {
				return nil, fmt.Errorf("failed to remove expired signature: %w", err)
			}
		}
	}
	
	// Store signature in database
	signedURLEntity := entities.SignedURL{
		ID:         uuid.Nil, // Auto-generated
		Signature:  signature,
		BucketName: bucket.Name,
		FileName:   file.Name,
		Method:     "GET",
		ExpiresAt:  expiresAt,
		Used:       false,
		SingleUse: command.SingleUse,
	}
	
	// Add to database using GoNtext
	h.dbContext.SignedURLs.Add(signedURLEntity)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to store signature: %w", err)
	}
	
	// Generate signed URL using file endpoint with signature parameter
	signedURL := fmt.Sprintf("%s/api/v1/file/%s/%s?signature=%s", 
		h.settings.BaseURL, 
		command.BucketID.String(), 
		command.FileID.String(), 
		signature)
	
	return &GenerateSignedURLResponse{
		URL:       signedURL,
		ExpiresAt: expiresAt,
		Success:   true,
		Message:   "Signed URL generated successfully",
	}, nil
}


func (h *GenerateSignedURLRequestHandler) generateHMAC(payload, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(payload))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

// ValidateSignedURL validates a signed URL signature against the database
// Now only needs the signature - gets bucketID, fileID, and expires from database
func (h *GenerateSignedURLRequestHandler) ValidateSignedURL(signature string) (*entities.SignedURL, error) {
	// First, check if signature exists in database
	signedURL, err := h.dbContext.SignedURLs.Where(&entities.SignedURL{
		Signature: signature,
	}).FirstOrDefault()
	
	if err != nil || signedURL == nil {
		return nil, fmt.Errorf("signature not found in database")
	}
	
	// Check if signature has expired (get expires from database)
	if signedURL.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("signature has expired")
	}
	
	// Check if signature has already been used (only if single-use is enabled)
	if signedURL.SingleUse && signedURL.Used {
		return nil, fmt.Errorf("single-use signature has already been used")
	}
	
	// Get signing secret from settings
	signingSecret := h.settings.SignatureSecret
	
	bucket, err := h.dbContext.Buckets.Where(&entities.Bucket{Name: signedURL.BucketName}).FirstOrDefault()
	if err != nil || bucket == nil {
		return nil, fmt.Errorf("bucket not found for signature")
	}
	
	file, err := h.dbContext.Files.Where(&entities.File{
		Name:     signedURL.FileName,
		BucketId: bucket.Id,
	}).FirstOrDefault()
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found for signature")
	}
	
	payload := fmt.Sprintf("%s:%s", bucket.Id.String(), file.Id.String())
	
	// Generate expected signature
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(payload))
	expectedSignature := base64.URLEncoding.EncodeToString(hash.Sum(nil))
	
	// Compare signatures for integrity check
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("signature integrity check failed")
	}
	
	return signedURL, nil
}

func (h *GenerateSignedURLRequestHandler) MarkSignatureAsUsed(signature string) error {
	signedURL, err := h.dbContext.SignedURLs.Where(&entities.SignedURL{
		Signature: signature,
	}).FirstOrDefault()
	
	if err != nil || signedURL == nil {
		return fmt.Errorf("signature not found")
	}
	
	// Only mark as used if it's a single-use URL
	if signedURL.SingleUse {
		now := time.Now()
		signedURL.Used = true
		signedURL.UsedAt = &now
		
		// Update in database
		if err := h.dbContext.SignedURLs.Update(*signedURL); err != nil {
			return fmt.Errorf("failed to mark signature as used: %w", err)
		}
		
		return h.dbContext.SaveChanges()
	}
	
	// For non-single-use URLs, no need to mark as used
	return nil
}

// GetFileInfoFromSignature returns file and bucket information from a signature
func (h *GenerateSignedURLRequestHandler) GetFileInfoFromSignature(signature string) (*entities.File, *entities.Bucket, error) {
	signedURL, err := h.ValidateSignedURL(signature)
	if err != nil {
		return nil, nil, err
	}
	
	// Get bucket
	bucket, err := h.dbContext.Buckets.Where(&entities.Bucket{Name: signedURL.BucketName}).FirstOrDefault()
	if err != nil || bucket == nil {
		return nil, nil, fmt.Errorf("bucket not found")
	}
	
	// Get file
	file, err := h.dbContext.Files.Where(&entities.File{
		Name:     signedURL.FileName,
		BucketId: bucket.Id,
	}).FirstOrDefault()
	if err != nil || file == nil {
		return nil, nil, fmt.Errorf("file not found")
	}
	
	return file, bucket, nil
}

