package apikey

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

type DeleteAPIKeyCommand struct {
	ID     uuid.UUID `json:"id" validate:"required"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

type DeleteAPIKeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteAPIKeyRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewDeleteAPIKeyRequestHandler(dbContext *persistence.AppDbContext) *DeleteAPIKeyRequestHandler {
	return &DeleteAPIKeyRequestHandler{
		dbContext: dbContext,
	}
}

func (h *DeleteAPIKeyRequestHandler) Handle(ctx context.Context, command *DeleteAPIKeyCommand) (*DeleteAPIKeyResponse, error) {
	// Find the API key that belongs to the user using GoNtext static typing
	apiKey, err := h.dbContext.APIKeys.Where(&entities.APIKey{
		Id:     command.ID,
		UserId: command.UserID,
	}).FirstOrDefault()
	
	if err != nil || apiKey == nil {
		return nil, fmt.Errorf("API key not found")
	}
	
	// Delete the API key using GoNtext
	h.dbContext.APIKeys.Remove(*apiKey)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to delete API key: %w", err)
	}
	
	return &DeleteAPIKeyResponse{
		Success: true,
		Message: "API key deleted successfully",
	}, nil
}