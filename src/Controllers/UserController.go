package controllers

import (
	"context"
	"net/http"
	
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	
	"shbucket/src/Application/User"
	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Mediator"
)

type UserController struct {
	mediator       *mediator.Mediator
	validator      *validator.Validate
	authService    *auth.AuthorizationService
}

func NewUserController(mediator *mediator.Mediator, validator *validator.Validate, authService *auth.AuthorizationService) *UserController {
	return &UserController{
		mediator:    mediator,
		validator:   validator,
		authService: authService,
	}
}

//	@Summary		User login
//	@Description	Authenticate user with email and password, returns JWT token for subsequent requests
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		user.LoginCommand							true	"Login credentials"
//	@Success		200			{object}	user.LoginResponse							"Login successful"
//	@Failure		400			{object}	map[string]string							"Invalid credentials"
//	@Router			/auth/login [post]
func (ctrl *UserController) Login(c *fiber.Ctx) error {
	var command user.LoginCommand
	
	if err := c.BodyParser(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	if err := ctrl.validator.Struct(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}
	
	response, err := ctrl.mediator.Send(context.Background(), &command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	loginResponse := response.(*user.LoginResponse)
	return c.JSON(loginResponse)
}

//	@Summary		User registration
//	@Description	Register a new user account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		user.RegisterCommand	true	"User registration data"
//	@Success		201		{object}	user.RegisterResponse	"User created successfully"
//	@Failure		400		{object}	map[string]string		"Validation error"
//	@Router			/auth/register [post]
func (ctrl *UserController) Register(c *fiber.Ctx) error {
	var command user.RegisterCommand
	
	if err := c.BodyParser(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	if err := ctrl.validator.Struct(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}
	
	response, err := ctrl.mediator.Send(context.Background(), &command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	registerResponse := response.(*user.RegisterResponse)
	return c.Status(http.StatusCreated).JSON(registerResponse)
}

//	@Summary		User logout
//	@Description	Logout user and invalidate session token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	user.LogoutResponse	"Logout successful"
//	@Failure		400	{object}	map[string]string	"Bad request"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Router			/auth/logout [post]
func (ctrl *UserController) Logout(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	command := &user.LogoutCommand{
		UserID:    userContext.UserID,
		TokenHash: "",
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	logoutResponse := response.(*user.LogoutResponse)
	return c.JSON(logoutResponse)
}

//	@Summary		Change password
//	@Description	Change the password for the authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		user.ChangePasswordCommand	true	"Password change details"
//	@Success		200	{object}	user.ChangePasswordResponse	"Password changed successfully"
//	@Failure		400	{object}	map[string]string			"Bad request"
//	@Failure		401	{object}	map[string]string			"Unauthorized"
//	@Router			/auth/change-password [post]
func (ctrl *UserController) ChangePassword(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	var command user.ChangePasswordCommand
	
	if err := c.BodyParser(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	command.UserID = userContext.UserID
	
	if err := ctrl.validator.Struct(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}
	
	response, err := ctrl.mediator.Send(context.Background(), &command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	changePasswordResponse := response.(*user.ChangePasswordResponse)
	return c.JSON(changePasswordResponse)
}

//	@Summary		Get user by ID
//	@Description	Get information about a specific user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			id	path		string					true	"User ID"
//	@Success		200	{object}	user.GetUserResponse	"User information"
//	@Failure		400	{object}	map[string]string		"Invalid user ID"
//	@Failure		404	{object}	map[string]string		"User not found"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Router			/users/{id} [get]
func (ctrl *UserController) GetUser(c *fiber.Ctx) error {
	userIDParam := c.Params("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}
	
	command := &user.GetUserCommand{
		UserID: userID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	getUserResponse := response.(*user.GetUserResponse)
	return c.JSON(getUserResponse)
}

//	@Summary		List users
//	@Description	Retrieve a paginated list of all users (admin only)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			page			query		int			false	"Page number (default: 1)"
//	@Param			limit			query		int			false	"Items per page (default: 10)"
//	@Param			include_buckets	query		bool		false	"Include user buckets"
//	@Param			include_sessions	query		bool		false	"Include user sessions"
//	@Param			include_all		query		bool		false	"Include all related data"
//	@Success		200	{object}	user.ListUsersResponse	"List of users"
//	@Failure		400	{object}	map[string]string		"Bad request"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Router			/users [get]
func (ctrl *UserController) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	// Support for Include functionality (like EF Core)
	includeBuckets := c.QueryBool("include_buckets", false)
	includeSessions := c.QueryBool("include_sessions", false)
	includeAll := c.QueryBool("include_all", false)
	
	command := &user.ListUsersCommand{
		Page:            page,
		Limit:           limit,
		IncludeBuckets:  includeBuckets,
		IncludeSessions: includeSessions,
		IncludeAll:      includeAll,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	listUsersResponse := response.(*user.ListUsersResponse)
	return c.JSON(listUsersResponse)
}