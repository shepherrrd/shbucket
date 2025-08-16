package node

import (
	"context"
	"fmt"
	"net/url"
	
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type RegisterNodeCommand struct {
	Name       string `json:"name" validate:"required,min=3,max=100"`
	URL        string `json:"url" validate:"required,url"`
	AuthKey    string `json:"auth_key" validate:"required,min=32"` // 32+ chars for security
	MaxStorage int64  `json:"max_storage" validate:"min=0"`
	Priority   int    `json:"priority" validate:"min=0,max=100"`
	IsActive   bool   `json:"is_active"`
}

type RegisterNodeResponse struct {
	Node    models.StorageNodeResponse `json:"node"`
	Success bool                       `json:"success"`
	Message string                     `json:"message"`
}

type RegisterNodeRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewRegisterNodeRequestHandler(dbContext *persistence.AppDbContext) *RegisterNodeRequestHandler {
	return &RegisterNodeRequestHandler{
		dbContext: dbContext,
	}
}

func (h *RegisterNodeRequestHandler) Handle(ctx context.Context, command *RegisterNodeCommand) (*RegisterNodeResponse, error) {
	// Validate URL
	_, err := url.Parse(command.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check if storage node with this URL already exists
	existingNode, err := h.dbContext.StorageNodes.Where(&entities.StorageNode{URL: command.URL}).FirstOrDefault()
	if err == nil && existingNode != nil {
		return nil, fmt.Errorf("storage node with this URL already exists")
	}

	node := &entities.StorageNode{
		Name:        command.Name,
		URL:         command.URL,
		AuthKey:     command.AuthKey,
		MaxStorage:  command.MaxStorage,
		UsedStorage: 0,
		Priority:    command.Priority,
		IsActive:    command.IsActive,
		IsHealthy:   false, // Will be set to true on first successful ping
	}

	// Add the node using GoNtext
	h.dbContext.StorageNodes.Add(*node)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to register storage node: %w", err)
	}

	nodeResponse := models.StorageNodeResponse{
		ID:          node.Id,
		Name:        node.Name,
		URL:         node.URL,
		MaxStorage:  node.MaxStorage,
		UsedStorage: node.UsedStorage,
		Priority:    node.Priority,
		IsActive:    node.IsActive,
		IsHealthy:   node.IsHealthy,
		CreatedAt:   node.CreatedAt,
		UpdatedAt:   node.UpdatedAt,
		LastPing:    node.LastPing,
	}

	
	return &RegisterNodeResponse{
		Node:    nodeResponse,
		Success: true,
		Message: "Storage node registered successfully",
	}, nil
}