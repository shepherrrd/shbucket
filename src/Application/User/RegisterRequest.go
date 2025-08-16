package user

import (
	"context"
	"fmt"
	
	"golang.org/x/crypto/bcrypt"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type RegisterCommand struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"omitempty,oneof=viewer editor manager admin"`
}

type RegisterResponse struct {
	User    models.UserResponse `json:"user"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
}

type RegisterRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewRegisterRequestHandler(dbContext *persistence.AppDbContext) *RegisterRequestHandler {
	return &RegisterRequestHandler{
		dbContext: dbContext,
	}
}

func (h *RegisterRequestHandler) Handle(ctx context.Context, command *RegisterCommand) (*RegisterResponse, error) {
	// Check if user exists by email or username (can't use OR in GoNtext yet, so check separately)
	existingUserByEmail, _ := h.dbContext.Users.Where(&entities.User{Email: command.Email}).FirstOrDefault()
	existingUserByUsername, _ := h.dbContext.Users.Where(&entities.User{Username: command.Username}).FirstOrDefault()
	if existingUserByEmail != nil || existingUserByUsername != nil {
		return nil, fmt.Errorf("user with this email or username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := command.Role
	if role == "" {
		role = "viewer"
	}

	user := &entities.User{
		Username:     command.Username,
		Email:        command.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		IsActive:     true,
	}

	// Create user using GoNtext
	h.dbContext.Users.Add(*user)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
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

	return &RegisterResponse{
		User:    userResponse,
		Success: true,
		Message: "User registered successfully",
	}, nil
}