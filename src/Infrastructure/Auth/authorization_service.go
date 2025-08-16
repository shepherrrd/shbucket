package auth

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Utils"
)

// AuthorizationService handles authorization logic
type AuthorizationService struct {
	jwtHandler *JWTHandler
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID   uuid.UUID
	Username string
	Email    string
	Role     string
	IsActive bool
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(jwtHandler *JWTHandler) *AuthorizationService {
	return &AuthorizationService{
		jwtHandler: jwtHandler,
	}
}

// AuthorizeRequest validates the request and extracts user context
func (a *AuthorizationService) AuthorizeRequest(c *fiber.Ctx) (*UserContext, error) {
	// Get Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is required")
	}

	// Extract token
	token, err := a.jwtHandler.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return nil, err
	}

	// Validate token
	claims, err := a.jwtHandler.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Create user context
	userContext := &UserContext{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     claims.Role,
		IsActive: true, // Would be checked against database in production
	}

	return userContext, nil
}

// RequireRole creates middleware that requires specific role
func (a *AuthorizationService) RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
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

// RequirePermission creates middleware that requires specific permission
func (a *AuthorizationService) RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userContext, err := a.AuthorizeRequest(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "unauthorized: " + err.Error(),
			})
		}

		if !a.HasPermission(userContext.Role, permission) {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: missing permission: " + permission,
			})
		}

		// Store user context in fiber locals
		c.Locals("user", userContext)
		return c.Next()
	}
}

// ValidateAccess validates access based on auth rules
func (a *AuthorizationService) ValidateAccess(authRule entities.AuthRule) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If auth is disabled, allow access
		if !authRule.Enabled || authRule.Type == "none" {
			return c.Next()
		}

		switch authRule.Type {
		case "jwt":
			// Validate JWT token
			userContext, err := a.AuthorizeRequest(c)
			if err != nil {
				return c.Status(401).JSON(fiber.Map{
					"error": "unauthorized: " + err.Error(),
				})
			}
			c.Locals("user", userContext)
			return c.Next()

		case "api_key":
			return a.validateAPIKey(c, utils.ConvertJSONToMap(authRule.Config))

		case "signed_url":
			return a.validateSignedURL(c, utils.ConvertJSONToMap(authRule.Config))

		case "session":
			return a.validateSession(c, utils.ConvertJSONToMap(authRule.Config))

		default:
			return c.Status(401).JSON(fiber.Map{
				"error": "unsupported authentication type: " + authRule.Type,
			})
		}
	}
}

// HasRole checks if user has the required role
func (a *AuthorizationService) HasRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"viewer":  1,
		"editor":  2,
		"manager": 3,
		"admin":   4,
	}

	userLevel, exists := roleHierarchy[strings.ToLower(userRole)]
	if !exists {
		return false
	}

	requiredLevel, exists := roleHierarchy[strings.ToLower(requiredRole)]
	if !exists {
		return false
	}

	return userLevel >= requiredLevel
}

// HasPermission checks if user has the required permission
func (a *AuthorizationService) HasPermission(userRole, permission string) bool {
	permissions := map[string][]string{
		"viewer": {"read", "list"},
		"editor": {"read", "list", "write", "upload", "download"},
		"manager": {"read", "list", "write", "upload", "download", "delete", "manage_buckets"},
		"admin": {"read", "list", "write", "upload", "download", "delete", "manage_buckets", "manage_users", "manage_system"},
	}

	userPermissions, exists := permissions[strings.ToLower(userRole)]
	if !exists {
		return false
	}

	for _, perm := range userPermissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// GetUserFromContext extracts user context from fiber locals
func (a *AuthorizationService) GetUserFromContext(c *fiber.Ctx) (*UserContext, error) {
	user := c.Locals("user")
	if user == nil {
		return nil, fmt.Errorf("user context not found")
	}

	userContext, ok := user.(*UserContext)
	if !ok {
		return nil, fmt.Errorf("invalid user context type")
	}

	return userContext, nil
}

// validateAPIKey validates API key authentication
func (a *AuthorizationService) validateAPIKey(c *fiber.Ctx, config map[string]interface{}) error {
	// Get API key from header
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		// Try Authorization header
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if apiKey == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "API key is required",
		})
	}

	// Extract allowed keys from config
	allowedKeysInterface, ok := config["allowed_api_keys"]
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "no API keys configured",
		})
	}

	allowedKeys, ok := allowedKeysInterface.([]interface{})
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "invalid API keys configuration",
		})
	}

	// Check if provided key is allowed
	for _, allowedKeyInterface := range allowedKeys {
		allowedKey, ok := allowedKeyInterface.(string)
		if ok && allowedKey == apiKey {
			return c.Next()
		}
	}

	return c.Status(401).JSON(fiber.Map{
		"error": "invalid API key",
	})
}

// validateSignedURL validates signed URL authentication
func (a *AuthorizationService) validateSignedURL(c *fiber.Ctx, config map[string]interface{}) error {
	signature := c.Query("signature")
	if signature == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "signature parameter is required",
		})
	}

	// TODO: Implement signed URL validation logic
	// This would involve checking the signature against the signing secret
	// and validating expiration time

	return c.Status(501).JSON(fiber.Map{
		"error": "signed URL authentication not fully implemented",
	})
}

// validateSession validates session-based authentication
func (a *AuthorizationService) validateSession(c *fiber.Ctx, config map[string]interface{}) error {
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		sessionToken = c.Get("X-Session-Token")
	}

	if sessionToken == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "session token is required",
		})
	}

	// TODO: Implement session validation logic
	// This would involve checking the session against the database

	return c.Status(501).JSON(fiber.Map{
		"error": "session authentication not fully implemented",
	})
}