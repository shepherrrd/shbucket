package user

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

type LogoutCommand struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"token_hash"`
}

type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LogoutRequestHandler struct {
	dbContext  *persistence.AppDbContext
	jwtHandler *auth.JWTHandler
}

func NewLogoutRequestHandler(dbContext *persistence.AppDbContext, jwtHandler *auth.JWTHandler) *LogoutRequestHandler {
	return &LogoutRequestHandler{
		dbContext:  dbContext,
		jwtHandler: jwtHandler,
	}
}

func (h *LogoutRequestHandler) Handle(ctx context.Context, command *LogoutCommand) (*LogoutResponse, error) {
	// Find and delete session using GoNtext
	session, err := h.dbContext.Sessions.Where(&entities.Session{
		UserId:    command.UserID,
		TokenHash: command.TokenHash,
	}).FirstOrDefault()
	
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	
	if session == nil {
		return &LogoutResponse{
			Success: false,
			Message: "Session not found",
		}, nil
	}
	
	// Delete the session
	h.dbContext.Sessions.Remove(*session)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to logout: %w", err)
	}

	return &LogoutResponse{
		Success: true,
		Message: "Logout successful",
	}, nil
}