package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

// APIKeyUserContext represents the user context from an API key
type APIKeyUserContext struct {
	APIKeyID    uuid.UUID
	UserID      uuid.UUID
	Username    string
	Email       string
	Role        string
	IsActive    bool
	Permissions entities.APIKeyPermission
	Source      string // "jwt" or "api_key"
}

// RequireRoleOrAPIKey creates middleware that supports both JWT and API key authentication
func (a *AuthorizationService) RequireRoleOrAPIKey(requiredRole string, dbContext *persistence.AppDbContext) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First try API key authentication
		if apiKeyHeader := c.Get("X-API-Key"); apiKeyHeader != "" {
			userContext, err := a.validateAPIKeyAuth(apiKeyHeader, dbContext)
			if err != nil {
				return c.Status(401).JSON(fiber.Map{
					"error": "Invalid API key: " + err.Error(),
				})
			}

			// Check if API key has required permissions based on role
			if !a.hasAPIKeyPermissionForRole(userContext.Permissions, requiredRole) {
				return c.Status(403).JSON(fiber.Map{
					"error": "API key does not have sufficient permissions",
				})
			}

			// Store API key user context in fiber locals
			c.Locals("user", &UserContext{
				UserID:   userContext.UserID,
				Username: userContext.Username,
				Email:    userContext.Email,
				Role:     userContext.Role,
				IsActive: userContext.IsActive,
			})
			c.Locals("api_key_context", userContext)
			return c.Next()
		}

		// Fall back to JWT authentication
		userContext, err := a.AuthorizeRequest(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "unauthorized: " + err.Error(),
			})
		}

		if !a.HasRole(userContext.Role, requiredRole) {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: insufficient permissions",
			})
		}

		// Store user context in fiber locals
		c.Locals("user", userContext)
		return c.Next()
	}
}

// validateAPIKeyAuth validates an API key and returns user context
func (a *AuthorizationService) validateAPIKeyAuth(apiKey string, dbContext *persistence.AppDbContext) (*APIKeyUserContext, error) {
	// Hash the provided API key
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	// Find API key in database using GoNtext
	dbAPIKey, err := dbContext.APIKeys.Where(&entities.APIKey{KeyHash: keyHash,IsActive: true}).FirstOrDefault()
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if dbAPIKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Check if API key has expired (use time.Now() instead of NowFunc)
	if dbAPIKey.ExpiresAt != nil && dbAPIKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Get the associated user using GoNtext
	user, err := dbContext.Users.Where(&entities.User{Id: dbAPIKey.UserId}).FirstOrDefault()
	if err != nil {
		return nil, fmt.Errorf("database error finding user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("associated user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("associated user account is disabled")
	}

	// Parse permissions
	var permissions entities.APIKeyPermission
	if err := json.Unmarshal(dbAPIKey.Permissions, &permissions); err != nil {
		return nil, fmt.Errorf("failed to parse API key permissions: %w", err)
	}

	// Create API key user context
	userContext := &APIKeyUserContext{
		APIKeyID:    dbAPIKey.Id,
		UserID:      user.Id,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		Permissions: permissions,
		Source:      "api_key",
	}

	return userContext, nil
}

// hasAPIKeyPermissionForRole checks if API key permissions meet role requirements
func (a *AuthorizationService) hasAPIKeyPermissionForRole(permissions entities.APIKeyPermission, requiredRole string) bool {
	switch strings.ToLower(requiredRole) {
	case "viewer":
		return permissions.Read
	case "editor":
		return permissions.Read && permissions.Write
	case "manager":
		return permissions.Read && permissions.Write
	case "admin":
		return permissions.Read && permissions.Write
	default:
		return false
	}
}

// GetAPIKeyContextFromRequest extracts API key context from fiber locals
func GetAPIKeyContextFromRequest(c *fiber.Ctx) (*APIKeyUserContext, bool) {
	apiKeyContext := c.Locals("api_key_context")
	if apiKeyContext == nil {
		return nil, false
	}

	userContext, ok := apiKeyContext.(*APIKeyUserContext)
	return userContext, ok
}

// IsAPIKeyRequest checks if the current request was authenticated with an API key
func IsAPIKeyRequest(c *fiber.Ctx) bool {
	_, isAPIKey := GetAPIKeyContextFromRequest(c)
	return isAPIKey
}