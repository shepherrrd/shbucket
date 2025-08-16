package user

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

type ChangePasswordCommand struct {
	UserID      uuid.UUID `json:"user_id"`
	OldPassword string    `json:"old_password" validate:"required"`
	NewPassword string    `json:"new_password" validate:"required,min=6"`
}

type ChangePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ChangePasswordRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewChangePasswordRequestHandler(dbContext *persistence.AppDbContext) *ChangePasswordRequestHandler {
	return &ChangePasswordRequestHandler{
		dbContext: dbContext,
	}
}

func (h *ChangePasswordRequestHandler) Handle(ctx context.Context, command *ChangePasswordCommand) (*ChangePasswordResponse, error) {
	// Find user using GoNtext
	user, err := h.dbContext.Users.Where(&entities.User{Id: command.UserID}).FirstOrDefault()
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(command.OldPassword)); err != nil {
		return nil, fmt.Errorf("invalid old password")
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(command.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password using GoNtext
	user.PasswordHash = string(hashedNewPassword)
	if err := h.dbContext.Users.Update(*user); err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	return &ChangePasswordResponse{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}