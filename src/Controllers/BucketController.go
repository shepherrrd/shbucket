package controllers

import (
	"context"
	"net/http"
	
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	
	"shbucket/src/Application/Bucket"
	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Mediator"
)

type BucketController struct {
	mediator    *mediator.Mediator
	validator   *validator.Validate
	authService *auth.AuthorizationService
}

func NewBucketController(mediator *mediator.Mediator, validator *validator.Validate, authService *auth.AuthorizationService) *BucketController {
	return &BucketController{
		mediator:    mediator,
		validator:   validator,
		authService: authService,
	}
}

//	@Summary		Create new bucket
//	@Description	Create a new storage bucket with specified settings
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			request	body		bucket.CreateBucketCommand	true	"Bucket creation details"
//	@Success		201		{object}	bucket.CreateBucketResponse	"Bucket created successfully"
//	@Failure		400		{object}	map[string]string			"Bad request"
//	@Failure		401		{object}	map[string]string			"Unauthorized"
//	@Router			/buckets [post]
func (ctrl *BucketController) CreateBucket(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	var command bucket.CreateBucketCommand
	
	if err := c.BodyParser(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	command.OwnerID = userContext.UserID
	
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
	
	createBucketResponse := response.(*bucket.CreateBucketResponse)
	return c.Status(http.StatusCreated).JSON(createBucketResponse)
}

//	@Summary		Delete bucket
//	@Description	Delete a storage bucket by ID
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path		string						true	"Bucket ID"
//	@Success		200	{object}	bucket.DeleteBucketResponse	"Bucket deleted successfully"
//	@Failure		400	{object}	map[string]string			"Bad request"
//	@Failure		401	{object}	map[string]string			"Unauthorized"
//	@Router			/buckets/{id} [delete]
func (ctrl *BucketController) DeleteBucket(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	bucketIDParam := c.Params("id")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	command := &bucket.DeleteBucketCommand{
		BucketID: bucketID,
		UserID:   userContext.UserID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	deleteBucketResponse := response.(*bucket.DeleteBucketResponse)
	return c.JSON(deleteBucketResponse)
}

//	@Summary		Get bucket by ID
//	@Description	Retrieve detailed information about a specific bucket
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path		string						true	"Bucket ID"
//	@Success		200	{object}	bucket.GetBucketResponse	"Bucket information"
//	@Failure		400	{object}	map[string]string			"Invalid bucket ID"
//	@Failure		404	{object}	map[string]string			"Bucket not found"
//	@Router			/buckets/{id} [get]
func (ctrl *BucketController) GetBucket(c *fiber.Ctx) error {
	bucketIDParam := c.Params("id")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	command := &bucket.GetBucketCommand{
		BucketID: bucketID,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	getBucketResponse := response.(*bucket.GetBucketResponse)
	return c.JSON(getBucketResponse)
}

//	@Summary		List buckets
//	@Description	Retrieve a paginated list of buckets for the authenticated user
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			page	query		int						false	"Page number (default: 1)"
//	@Param			limit	query		int						false	"Items per page (default: 10)"
//	@Success		200	{object}	bucket.ListBucketsResponse	"List of buckets"
//	@Failure		400	{object}	map[string]string			"Bad request"
//	@Failure		401	{object}	map[string]string			"Unauthorized"
//	@Router			/buckets [get]
func (ctrl *BucketController) ListBuckets(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	command := &bucket.ListBucketsCommand{
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
	
	listBucketsResponse := response.(*bucket.ListBucketsResponse)
	return c.JSON(listBucketsResponse)
}

//	@Summary		Update bucket
//	@Description	Update bucket settings and metadata
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id		path		string						true	"Bucket ID"
//	@Param			request	body		bucket.UpdateBucketCommand	true	"Bucket update details"
//	@Success		200	{object}	bucket.UpdateBucketResponse	"Bucket updated successfully"
//	@Failure		400	{object}	map[string]string			"Bad request"
//	@Failure		401	{object}	map[string]string			"Unauthorized"
//	@Router			/buckets/{id} [put]
func (ctrl *BucketController) UpdateBucket(c *fiber.Ctx) error {
	userContext, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	bucketIDParam := c.Params("id")
	bucketID, err := uuid.Parse(bucketIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bucket ID",
		})
	}
	
	var command bucket.UpdateBucketCommand
	
	if err := c.BodyParser(&command); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	command.BucketID = bucketID
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
	
	updateBucketResponse := response.(*bucket.UpdateBucketResponse)
	return c.JSON(updateBucketResponse)
}