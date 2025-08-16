package user

import (
	"context"
	"fmt"
	
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type ListUsersCommand struct {
	Page            int  `json:"page"`
	Limit           int  `json:"limit"`
	IncludeBuckets  bool `json:"include_buckets"`
	IncludeSessions bool `json:"include_sessions"`
	IncludeAll      bool `json:"include_all"`
}

type ListUsersResponse struct {
	Users   []models.UserResponse `json:"users"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
	Success bool                  `json:"success"`
	Message string                `json:"message"`
}

type ListUsersRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewListUsersRequestHandler(dbContext *persistence.AppDbContext) *ListUsersRequestHandler {
	return &ListUsersRequestHandler{
		dbContext: dbContext,
	}
}

func (h *ListUsersRequestHandler) Handle(ctx context.Context, command *ListUsersCommand) (*ListUsersResponse, error) {
	page := command.Page
	limit := command.Limit
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	var users []entities.User
	var total int64

	// Use GoNtext LINQ for counting (like EF Core: context.Users.Count())
	userCount, err := h.dbContext.Users.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	total = int64(userCount)

	// Use GoNtext LINQ for pagination with Include functionality (like EF Core)
	query := h.dbContext.Users
	
	// Apply Include based on command parameters
	if command.IncludeAll {
		// Load all relationships automatically: context.Users.IncludeAll()
		query = query.IncludeAll()
	} else {
		// Type-safe includes with field validation (cleaner syntax):
		var includes []interface{}
		if command.IncludeBuckets {
			includes = append(includes, "Buckets")
		}
		if command.IncludeSessions {
			includes = append(includes, "Sessions")
		}
		if len(includes) > 0 {
			query = query.Include(includes...)
		}
	}
	
	users, err = query.Skip(offset).Take(limit).ToList()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = models.UserResponse{
			ID:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return &ListUsersResponse{
		Users:   userResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Success: true,
		Message: "Users retrieved successfully",
	}, nil
}