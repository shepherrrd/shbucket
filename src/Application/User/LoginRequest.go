package user

import (
	"context"
	"fmt"

	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginCommand struct {
	EmailOrUsername string `json:"email" validate:"required,min=3"`
	Password        string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	User         models.UserResponse `json:"user"`
	Token        string              `json:"token"`
	RefreshToken string              `json:"refresh_token"`
	ExpiresIn    int                 `json:"expires_in"`
	Success      bool                `json:"success"`
	Message      string              `json:"message"`
}

type LoginRequestHandler struct {
	dbContext  *persistence.AppDbContext
	jwtHandler *auth.JWTHandler
}

func NewLoginRequestHandler(dbContext *persistence.AppDbContext, jwtHandler *auth.JWTHandler) *LoginRequestHandler {
	return &LoginRequestHandler{
		dbContext:  dbContext,
		jwtHandler: jwtHandler,
	}
}

func (h *LoginRequestHandler) Handle(ctx context.Context, command *LoginCommand) (*LoginResponse, error) {
	// Use GoNtext LINQ with OR to find user by email or username 
	// (like EF Core: context.Users.Where(u => u.Email == emailOrUsername || u.Username == emailOrUsername).FirstOrDefault())
	user, err := h.dbContext.Users.Where(&entities.User{Email: command.EmailOrUsername}).OrField("Username", command.EmailOrUsername).FirstOrDefault()
	if err != nil || user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(command.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, sessionInfo, err := h.jwtHandler.GenerateToken(user.Id, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	session := entities.Session{
		Id:        uuid.Nil, // Explicitly set to nil to ensure auto-generation
		UserId:    sessionInfo.UserID,
		TokenHash: sessionInfo.TokenHash,
		ExpiresAt: sessionInfo.ExpiresAt,
		IsActive:  true,
	}

	// Use GoNtext to add session (like EF Core: context.Sessions.Add(session))
	_, err = h.dbContext.Sessions.Add(session)
	if err != nil {
		return nil, fmt.Errorf("failed to add session: %w", err)
	}
	
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
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

	return &LoginResponse{
		User:         userResponse,
		Token:        token,
		RefreshToken: token,
		ExpiresIn:    24 * 3600,
		Success:      true,
		Message:      "Login successful",
	}, nil
}