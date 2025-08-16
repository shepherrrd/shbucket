package user

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type GetUserCommand struct {
	UserID uuid.UUID `json:"user_id"`
}

type GetUserResponse struct {
	User    models.UserResponse `json:"user"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
}

type GetUserRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewGetUserRequestHandler(dbContext *persistence.AppDbContext) *GetUserRequestHandler {
	return &GetUserRequestHandler{
		dbContext: dbContext,
	}
}

func (h *GetUserRequestHandler) Handle(ctx context.Context, command *GetUserCommand) (*GetUserResponse, error) {
	// Use GoNtext LINQ to find by ID (like EF Core: context.Users.Find(id) or FirstOrDefault())
	user, err := h.dbContext.Users.ById(command.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	userResponse := models.UserResponse{
		ID:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return &GetUserResponse{
		User:    userResponse,
		Success: true,
		Message: "User retrieved successfully",
	}, nil
}