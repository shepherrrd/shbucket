package apikey

import (
	"context"
	"encoding/json"
	"fmt"

	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"

	"github.com/google/uuid"
)

type ListAPIKeysCommand struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Page   int        `json:"page" validate:"min=1"`
	Limit  int        `json:"limit" validate:"min=1,max=100"`
}

type ListAPIKeysResponse struct {
	APIKeys []models.APIKeyResponse `json:"api_keys"`
	Total   int64                   `json:"total"`
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
}

type ListAPIKeysRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewListAPIKeysRequestHandler(dbContext *persistence.AppDbContext) *ListAPIKeysRequestHandler {
	return &ListAPIKeysRequestHandler{
		dbContext: dbContext,
	}
}

func (h *ListAPIKeysRequestHandler) Handle(ctx context.Context, command *ListAPIKeysCommand) (*ListAPIKeysResponse, error) {
	
	
	// Set defaults
	if command.Page == 0 {
		command.Page = 1
	}
	if command.Limit == 0 {
		command.Limit = 20
	}
	
	offset := (command.Page - 1) * command.Limit
	
	// Get total count
	total,err := h.dbContext.APIKeys.Where(&entities.APIKey{UserId: command.UserID}).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count API keys: %w", err)
	}

	// Get API keys with user info using GoNtext's type-safe Include
	apiKeys, err := h.dbContext.APIKeys.
		Where(&entities.APIKey{UserId: command.UserID}).
		Include("User").
		OrderByDescending("CreatedAt").
		Take(command.Limit).
		Skip(offset).
		ToList(); 
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve API keys: %w", err)
	}
	// Convert to response format
	var apiKeyResponses []models.APIKeyResponse
	for _, apiKey := range apiKeys {
		var permissions entities.APIKeyPermission
		json.Unmarshal(apiKey.Permissions, &permissions)
		
		response := models.APIKeyResponse{
			ID:          apiKey.Id,
			Name:        apiKey.Name,
			KeyPrefix:   apiKey.KeyPrefix,
			UserID:      apiKey.UserId,
			Username:    apiKey.User.Username,
			IsActive:    apiKey.IsActive,
			Permissions: permissions,
			ExpiresAt:   apiKey.ExpiresAt,
			LastUsed:    apiKey.LastUsed,
			CreatedAt:   apiKey.CreatedAt,
			UpdatedAt:   apiKey.UpdatedAt,
		}
		apiKeyResponses = append(apiKeyResponses, response)
	}
	
	return &ListAPIKeysResponse{
		APIKeys: apiKeyResponses,
		Total:   total,
		Success: true,
		Message: "API keys retrieved successfully",
	}, nil
}