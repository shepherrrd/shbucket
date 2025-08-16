package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type CreateAPIKeyCommand struct {
	Name        string                      `json:"name" validate:"required,min=3,max=100"`
	UserID      uuid.UUID                   `json:"user_id" validate:"required"`
	Permissions entities.APIKeyPermission  `json:"permissions"`
	ExpiresAt   *time.Time                  `json:"expires_at,omitempty"`
}

type CreateAPIKeyResponse struct {
	APIKey    models.APIKeyResponse `json:"api_key"`
	Key       string                `json:"key"` // Only returned on creation
	Success   bool                  `json:"success"`
	Message   string                `json:"message"`
}

type CreateAPIKeyRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewCreateAPIKeyRequestHandler(dbContext *persistence.AppDbContext) *CreateAPIKeyRequestHandler {
	return &CreateAPIKeyRequestHandler{
		dbContext: dbContext,
	}
}

func (h *CreateAPIKeyRequestHandler) Handle(ctx context.Context, command *CreateAPIKeyCommand) (*CreateAPIKeyResponse, error) {
	// Generate API key
	plainKey, keyHash, keyPrefix, err := h.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	
	// Marshal permissions to JSON
	permissionsJSON, err := json.Marshal(command.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}
	
	// Create API key record
	apiKey := &entities.APIKey{
		Name:        command.Name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		UserId:      command.UserID, // Map to UserId field
		IsActive:    true,
		Permissions: datatypes.JSON(permissionsJSON),
		ExpiresAt:   command.ExpiresAt,
	}
	
	// Add API key using GoNtext
	h.dbContext.APIKeys.Add(*apiKey)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}
	
	// Load user for response using GoNtext
	user, err := h.dbContext.Users.Where(&entities.User{Id: apiKey.UserId}).FirstOrDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	
	// Convert permissions back to struct for response
	var permissions entities.APIKeyPermission
	json.Unmarshal(apiKey.Permissions, &permissions)
	
	response := models.APIKeyResponse{
		ID:          apiKey.Id,
		Name:        apiKey.Name,
		KeyPrefix:   apiKey.KeyPrefix,
		UserID:      apiKey.UserId,
		Username:    user.Username,
		IsActive:    apiKey.IsActive,
		Permissions: permissions,
		ExpiresAt:   apiKey.ExpiresAt,
		LastUsed:    apiKey.LastUsed,
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
	}
	
	return &CreateAPIKeyResponse{
		APIKey:  response,
		Key:     plainKey,
		Success: true,
		Message: "API key created successfully",
	}, nil
}

func (h *CreateAPIKeyRequestHandler) generateAPIKey() (plainKey, keyHash, keyPrefix string, err error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}
	
	// Create plain key with prefix
	plainKey = "shb_" + hex.EncodeToString(bytes)
	
	// Create hash
	hash := sha256.Sum256([]byte(plainKey))
	keyHash = hex.EncodeToString(hash[:])
	
	// Create prefix (first 12 chars including shb_)
	keyPrefix = plainKey[:12]
	
	return plainKey, keyHash, keyPrefix, nil
}