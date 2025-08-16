package node

import (
	"context"
	"fmt"
	entities "shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type ListNodesCommand struct {
	Page     int  `json:"page"`
	Limit    int  `json:"limit"`
	OnlyActive bool `json:"only_active"`
}

type ListNodesResponse struct {
	Nodes   []models.StorageNodeResponse `json:"nodes"`
	Total   int64                        `json:"total"`
	Page    int                          `json:"page"`
	Limit   int                          `json:"limit"`
	Success bool                         `json:"success"`
	Message string                       `json:"message"`
}

type ListNodesRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewListNodesRequestHandler(dbContext *persistence.AppDbContext) *ListNodesRequestHandler {
	return &ListNodesRequestHandler{
		dbContext: dbContext,
	}
}

func (h *ListNodesRequestHandler) Handle(ctx context.Context, command *ListNodesCommand) (*ListNodesResponse, error) {
	page := command.Page
	limit := command.Limit
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Use GoNtext LINQ-style querying
	nodeQuery := h.dbContext.StorageNodes
	if command.OnlyActive {
		nodeQuery = nodeQuery.Where(&entities.StorageNode{IsActive: true})
	}
	
	// Get total count
	totalCount, err := nodeQuery.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count storage nodes: %w", err)
	}
	total := int64(totalCount)

	// Get paginated results
	nodes, err := nodeQuery.Skip(offset).Take(limit).ToList()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage nodes: %w", err)
	}

	nodeResponses := make([]models.StorageNodeResponse, len(nodes))
	for i, node := range nodes {
		nodeResponses[i] = models.StorageNodeResponse{
			ID:          node.Id, // Updated to use Id (Go naming convention)
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
	}

	return &ListNodesResponse{
		Nodes:   nodeResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Success: true,
		Message: "Storage nodes retrieved successfully",
	}, nil
}