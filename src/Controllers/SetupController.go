package controllers

import (
	"context"
	"net/http"
	"time"
	
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
	"shbucket/src/Application/Setup"
	"shbucket/src/Infrastructure/Mediator"
	"shbucket/src/Models"
)

type SetupController struct {
	mediator  *mediator.Mediator
	validator *validator.Validate
}

func NewSetupController(mediator *mediator.Mediator, validator *validator.Validate) *SetupController {
	return &SetupController{
		mediator:  mediator,
		validator: validator,
	}
}

//	@Summary		Check setup status
//	@Description	Check if the system has been set up and configured
//	@Tags			setup
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	setup.CheckSetupResponse	"Setup status information"
//	@Failure		500	{object}	map[string]string			"Internal server error"
//	@Router			/setup/status [get]
func (ctrl *SetupController) CheckSetup(c *fiber.Ctx) error {
	command := &setup.CheckSetupCommand{}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	checkResponse := response.(*setup.CheckSetupResponse)
	return c.JSON(checkResponse)
}

//	@Summary		Setup master node
//	@Description	Initialize the system as a master node with database and admin user
//	@Tags			setup
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.MasterSetupRequest	true	"Master setup configuration"
//	@Success		200		{object}	setup.MasterSetupResponse	"Master setup successful"
//	@Failure		400		{object}	map[string]string			"Invalid request or setup failed"
//	@Router			/setup/master [post]
func (ctrl *SetupController) SetupMaster(c *fiber.Ctx) error {
	var req models.MasterSetupRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	if err := ctrl.validator.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}
	
	command := &setup.MasterSetupCommand{
		AdminUsername:   req.AdminUsername,
		AdminEmail:      req.AdminEmail,
		AdminPassword:   req.AdminPassword,
		StoragePath:     req.StoragePath,
		MaxStorage:      req.MaxStorage,
		DefaultAuthRule: req.DefaultAuthRule,
		DefaultSettings: req.DefaultSettings,
		JWTSecret:       req.JWTSecret,
		SystemName:      req.SystemName,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	masterResponse := response.(*setup.MasterSetupResponse)
	return c.Status(http.StatusCreated).JSON(masterResponse)
}

//	@Summary		Setup node
//	@Description	Initialize the system as a storage node connected to a master
//	@Tags			setup
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.NodeSetupRequest	true	"Node setup configuration"
//	@Success		201	{object}	setup.NodeSetupResponse	"Node setup successful"
//	@Failure		400	{object}	map[string]string		"Invalid request or setup failed"
//	@Router			/setup/node [post]
func (ctrl *SetupController) SetupNode(c *fiber.Ctx) error {
	var req models.NodeSetupRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	if err := ctrl.validator.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}
	
	command := &setup.NodeSetupCommand{
		MasterURL:    req.MasterURL,
		NodeName:     req.NodeName,
		NodeAPIKey:   req.NodeAPIKey,
		StoragePath:  req.StoragePath,
		MaxStorage:   req.MaxStorage,
		MasterAPIKey: req.MasterAPIKey,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	nodeResponse := response.(*setup.NodeSetupResponse)
	return c.Status(http.StatusCreated).JSON(nodeResponse)
}

//	@Summary		Get system information
//	@Description	Retrieve system status and information after setup
//	@Tags			setup
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.SystemInfoResponse	"System information"
//	@Failure		500	{object}	map[string]string			"Internal server error"
//	@Router			/setup/info [get]
func (ctrl *SetupController) GetSystemInfo(c *fiber.Ctx) error {
	// This would return system information after setup
	return c.JSON(models.SystemInfoResponse{
		SystemName:  "SHBucket",
		Version:     "2.0.0",
		IsHealthy:   true,
		LastChecked: time.Now(),
	})
}