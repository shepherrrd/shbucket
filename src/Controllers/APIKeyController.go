package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"
	
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	
	apikey "shbucket/src/Application/APIKey"
	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Mediator"
)

type APIKeyController struct {
	mediator    *mediator.Mediator
	validator   *validator.Validate
	authService *auth.AuthorizationService
}

func NewAPIKeyController(mediator *mediator.Mediator, validator *validator.Validate, authService *auth.AuthorizationService) *APIKeyController {
	return &APIKeyController{
		mediator:    mediator,
		validator:   validator,
		authService: authService,
	}
}

//	@Summary		Create API key
//	@Description	Create a new API key for programmatic access
//	@Tags			api-keys
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		object							true	"API key creation request"
//	@Success		201		{object}	apikey.CreateAPIKeyResponse		"API key created successfully"
//	@Failure		400		{object}	map[string]string				"Bad request"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Router			/api-keys [post]
func (ctrl *APIKeyController) CreateAPIKey(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	var request struct {
		Name        string                      `json:"name" validate:"required,min=3,max=100"`
		Permissions entities.APIKeyPermission  `json:"permissions"`
		ExpiresIn   *int                        `json:"expires_in,omitempty"` // Seconds from now
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	if err := ctrl.validator.Struct(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	// Calculate expiration time
	var expiresAt *time.Time
	if request.ExpiresIn != nil {
		expiry := time.Now().Add(time.Duration(*request.ExpiresIn) * time.Second)
		expiresAt = &expiry
	}
	
	command := &apikey.CreateAPIKeyCommand{
		Name:        request.Name,
		UserID:      userContext.UserID,
		Permissions: request.Permissions,
		ExpiresAt:   expiresAt,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	createResponse := response.(*apikey.CreateAPIKeyResponse)
	return c.Status(http.StatusCreated).JSON(createResponse)
}

//	@Summary		List API keys
//	@Description	Retrieve a paginated list of API keys for the authenticated user
//	@Tags			api-keys
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			page	query		int						false	"Page number (default: 1)"
//	@Param			limit	query		int						false	"Items per page (default: 20)"
//	@Success		200	{object}	apikey.ListAPIKeysResponse	"List of API keys"
//	@Failure		400	{object}	map[string]string			"Bad request"
//	@Failure		401	{object}	map[string]string			"Unauthorized"
//	@Router			/api-keys [get]
func (ctrl *APIKeyController) ListAPIKeys(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	command := &apikey.ListAPIKeysCommand{
		UserID: userContext.UserID,
		Page:   page,
		Limit:  limit,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	listResponse := response.(*apikey.ListAPIKeysResponse)
	return c.JSON(listResponse)
}

//	@Summary		Delete API key
//	@Description	Delete an API key by ID
//	@Tags			api-keys
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			id	path		string						true	"API Key ID"
//	@Success		200	{object}	apikey.DeleteAPIKeyResponse	"API key deleted successfully"
//	@Failure		400	{object}	map[string]string			"Bad request"
//	@Failure		401	{object}	map[string]string			"Unauthorized"
//	@Router			/api-keys/{id} [delete]
func (ctrl *APIKeyController) DeleteAPIKey(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	keyIDParam := c.Params("id")
	keyID, err := uuid.Parse(keyIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid API key ID",
		})
	}
	
	command := &apikey.DeleteAPIKeyCommand{
		ID:     keyID,
		UserID: userContext.UserID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	deleteResponse := response.(*apikey.DeleteAPIKeyResponse)
	return c.JSON(deleteResponse)
}