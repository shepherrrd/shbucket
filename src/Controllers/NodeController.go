package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	
	"shbucket/src/Application/Node"
	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Mediator"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type NodeController struct {
	mediator    *mediator.Mediator
	validator   *validator.Validate
	authService *auth.AuthorizationService
	dbContext   *persistence.AppDbContext
}

func NewNodeController(mediator *mediator.Mediator, validator *validator.Validate, authService *auth.AuthorizationService, dbContext *persistence.AppDbContext) *NodeController {
	return &NodeController{
		mediator:    mediator,
		validator:   validator,
		authService: authService,
		dbContext:   dbContext,
	}
}

//	@Summary		Register storage node
//	@Description	Register a new storage node in the distributed system
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			request	body		models.RegisterNodeRequest		true	"Node registration details"
//	@Success		201		{object}	node.RegisterNodeResponse		"Node registered successfully"
//	@Failure		400		{object}	map[string]string				"Bad request"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Router			/nodes [post]
func (ctrl *NodeController) RegisterNode(c *fiber.Ctx) error {
	var req models.RegisterNodeRequest
	
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
	
	command := &node.RegisterNodeCommand{
		Name:       req.Name,
		URL:        req.URL,
		AuthKey:    req.AuthKey,
		MaxStorage: req.MaxStorage,
		Priority:   req.Priority,
		IsActive:   req.IsActive,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	registerResponse := response.(*node.RegisterNodeResponse)
	return c.Status(http.StatusCreated).JSON(registerResponse)
}

//	@Summary		List storage nodes
//	@Description	Get a list of all storage nodes in the distributed system
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			page	query		int		false	"Page number"		default(1)
//	@Param			limit	query		int		false	"Items per page"	default(10)
//	@Param			active	query		bool	false	"Show only active nodes"	default(false)
//	@Success		200		{object}	node.ListNodesResponse			"Nodes retrieved successfully"
//	@Failure		400		{object}	map[string]string				"Bad request"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Router			/nodes [get]
func (ctrl *NodeController) ListNodes(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	onlyActive := c.QueryBool("active", false)
	
	command := &node.ListNodesCommand{
		Page:       page,
		Limit:      limit,
		OnlyActive: onlyActive,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	listResponse := response.(*node.ListNodesResponse)
	return c.JSON(listResponse)
}

//	@Summary		Install storage node
//	@Description	Install and configure a new storage node
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			request	body		models.NodeInstallationRequest	true	"Node installation details"
//	@Success		201		{object}	models.NodeInstallationResponse	"Node installed successfully"
//	@Failure		400		{object}	map[string]string				"Bad request"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Router			/nodes/install [post]
func (ctrl *NodeController) InstallNode(c *fiber.Ctx) error {
	var req models.NodeInstallationRequest
	
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
	
	// Generate API key if not provided
	if req.APIKey == "" {
		req.APIKey = generateAuthKey()
	}
	
	// Create installation response
	baseURL := fmt.Sprintf("http://localhost:%d", req.Port)
	
	// Register the node
	registerCommand := &node.RegisterNodeCommand{
		Name:       req.Name,
		URL:        baseURL,
		AuthKey:    req.APIKey,
		MaxStorage: req.MaxStorage,
		Priority:   1,
		IsActive:   true,
	}
	
	response, err := ctrl.mediator.Send(context.Background(), registerCommand)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	registerResponse := response.(*node.RegisterNodeResponse)
	
	installResponse := &models.NodeInstallationResponse{
		Node:         registerResponse.Node,
		InstallPath:  "./shbucket-node",
		ConfigPath:   "./shbucket-node.env",
		StartCommand: fmt.Sprintf("cd shbucket-node && ./shbucket --port=%d --storage=%s --max-storage=%d", req.Port, req.StoragePath, req.MaxStorage),
		Success:      true,
		Message:      "Node installed and registered successfully",
	}
	
	return c.Status(http.StatusCreated).JSON(installResponse)
}

//	@Summary		Check node health
//	@Description	Check the health status of a specific storage node
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path		string	true	"Node ID"
//	@Success		200	{object}	models.NodeHealthCheckResponse	"Node health status"
//	@Failure		400	{object}	map[string]string				"Bad request"
//	@Failure		401	{object}	map[string]string				"Unauthorized"
//	@Failure		404	{object}	map[string]string				"Node not found"
//	@Router			/nodes/{id}/health [get]
func (ctrl *NodeController) HealthCheck(c *fiber.Ctx) error {
	nodeIDParam := c.Params("id")
	nodeID, err := uuid.Parse(nodeIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}
	
	// Get the node from database
	storageNode, err := ctrl.dbContext.StorageNodes.Where(entities.StorageNode{Id: nodeID}).FirstOrDefault()
	if err != nil || storageNode == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}
	
	// Perform actual health check
	isHealthy, responseTime, errorMsg := ctrl.pingNode(storageNode)
	
	// Update node health status in database
	now := time.Now()
	storageNode.IsHealthy = isHealthy
	storageNode.LastPing = &now
	storageNode.IsActive = true
	
	if err := ctrl.dbContext.SaveChanges(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update node health status",
		})
	}else{
		fmt.Sprintf("Healthy node %s",storageNode.IsHealthy)
	}

	
	response := &models.NodeHealthCheckResponse{
		NodeID:       nodeID,
		IsHealthy:    isHealthy,
		ResponseTime: responseTime,
		Success:      isHealthy,
		Message:      func() string {
			if isHealthy {
				return "Node is healthy"
			}
			return fmt.Sprintf("Node is unhealthy: %s", errorMsg)
		}(),
		Error: errorMsg,
	}
	
	return c.JSON(response)
}

//	@Summary		Check all nodes health
//	@Description	Check the health status of all storage nodes
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	map[string]interface{}	"Health check results for all nodes"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Router			/nodes/health [get]
func (ctrl *NodeController) CheckAllNodesHealth(c *fiber.Ctx) error {
	// Get all nodes from database
	allNodes, err := ctrl.dbContext.StorageNodes.ToList()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch nodes",
		})
	}
	
	// Perform health checks on all nodes
	healthResults := make([]models.NodeHealthCheckResponse, 0, len(allNodes))
	healthyCount := 0
	
	for i := range allNodes {
		isHealthy, responseTime, errorMsg := ctrl.pingNode(&allNodes[i])
		
		// Update node health status directly in the original slice
		now := time.Now()
		allNodes[i].IsHealthy = isHealthy
		allNodes[i].LastPing = &now
		allNodes[i].IsActive = true
		
		if isHealthy {
			healthyCount++
		}
		
		result := models.NodeHealthCheckResponse{
			NodeID:       allNodes[i].Id,
			IsHealthy:    isHealthy,
			ResponseTime: responseTime,
			Success:      isHealthy,
			Message: func() string {
				if isHealthy {
					return "Node is healthy"
				}
				return fmt.Sprintf("Node is unhealthy: %s", errorMsg)
			}(),
			Error: errorMsg,
		}
		healthResults = append(healthResults, result)
	}
	
	// Bulk update all nodes at once using UpdateRange
	ctrl.dbContext.StorageNodes.UpdateRange(allNodes)
	
	if err := ctrl.dbContext.SaveChanges(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save health check changes",
		})
	}
	
	return c.JSON(fiber.Map{
		"success":        true,
		"total_nodes":    len(allNodes),
		"healthy_nodes":  healthyCount,
		"unhealthy_nodes": len(allNodes) - healthyCount,
		"health_results": healthResults,
	})
}

//	@Summary		Delete storage node
//	@Description	Remove a storage node from the distributed system
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			id	path		string	true	"Node ID"
//	@Success		200	{object}	map[string]interface{}	"Node deleted successfully"
//	@Failure		400	{object}	map[string]string		"Bad request"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		404	{object}	map[string]string		"Node not found"
//	@Router			/nodes/{id} [delete]
func (ctrl *NodeController) DeleteNode(c *fiber.Ctx) error {
	nodeIDStr := c.Params("id")
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid node ID",
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Node deletion not yet implemented - placeholder response",
		"node_id": nodeID,
	})
}

//	@Summary		Self-register storage node
//	@Description	Allow a node to register itself without authentication (for node setup)
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object	true	"Node self-registration details"	example({"name":"node-1","url":"http://localhost:8081","max_storage":1073741824,"priority":1})
//	@Success		201		{object}	map[string]interface{}	"Node registered successfully with auth key"
//	@Failure		400		{object}	map[string]string		"Bad request"
//	@Router			/node/register [post]
func (ctrl *NodeController) SelfRegister(c *fiber.Ctx) error {
	var req struct {
		Name       string `json:"name" validate:"required,min=3,max=100"`
		URL        string `json:"url" validate:"required,url"`
		MaxStorage int64  `json:"max_storage" validate:"min=0"`
		Priority   int    `json:"priority" validate:"min=0,max=100"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if err := ctrl.validator.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Generate auth key for this node
	authKey := generateAuthKey()

	command := &node.RegisterNodeCommand{
		Name:       req.Name,
		URL:        req.URL,
		AuthKey:    authKey,
		MaxStorage: req.MaxStorage,
		Priority:   req.Priority,
		IsActive:   true,
	}

	response, err := ctrl.mediator.Send(context.Background(), command)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	registerResponse := response.(*node.RegisterNodeResponse)
	
	// Return the auth key so the node can store it
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"node":     registerResponse.Node,
		"auth_key": authKey,
		"success":  true,
		"message":  "Node registered successfully. Save the auth_key - you'll need it for master registration.",
	})
}

//	@Summary		Get node auth key
//	@Description	Retrieve the authentication key for a specific node by URL
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Param			url	query		string	true	"Node URL"
//	@Success		200	{object}	map[string]interface{}	"Node authentication key"
//	@Failure		400	{object}	map[string]string		"Bad request"
//	@Failure		404	{object}	map[string]string		"Node not found"
//	@Router			/node/auth-key [get]
func (ctrl *NodeController) GetAuthKey(c *fiber.Ctx) error {
	nodeURL := c.Query("url")
	if nodeURL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Node URL parameter required",
		})
	}

	// Find the node by URL using static typing
	storageNode, err := ctrl.dbContext.StorageNodes.First(&entities.StorageNode{URL: nodeURL})
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	return c.JSON(fiber.Map{
		"auth_key": storageNode.AuthKey,
		"node_id":  storageNode.Id,
		"name":     storageNode.Name,
		"url":      storageNode.URL,
		"success":  true,
	})
}

//	@Summary		Node ping
//	@Description	Allow a node to ping the master node to update health status
//	@Tags			nodes
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			url	query		string	true	"Node URL"
//	@Success		200	{object}	map[string]interface{}	"Ping successful"
//	@Failure		400	{object}	map[string]string		"Bad request"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Router			/node/ping [post]
func (ctrl *NodeController) Ping(c *fiber.Ctx) error {
	nodeURL := c.Query("url")
	authKey := c.Get("Authorization")
	
	// Also check X-API-Key header for API key authentication
	if authKey == "" {
		authKey = c.Get("X-API-Key")
	}
	
	if nodeURL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Node URL parameter required",
		})
	}

	if authKey == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header or X-API-Key header required",
		})
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(authKey, "Bearer ") {
		authKey = strings.TrimPrefix(authKey, "Bearer ")
	}

	// Find and validate the node using static typing (can't use AND with struct yet, so check separately)
	storageNode, err := ctrl.dbContext.StorageNodes.Where(&entities.StorageNode{URL: nodeURL, AuthKey: authKey}).FirstOrDefault()
	if err != nil || storageNode == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid node credentials",
		})
	}

	// Update last ping and health status
	now := time.Now()
	storageNode.LastPing = &now
	storageNode.IsHealthy = true

	if err := ctrl.dbContext.StorageNodes.Save(storageNode); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update node status",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Ping successful",
		"node_id": storageNode.Id,
	})
}

func generateAuthKey() string {
	// Generate a secure 32+ character authentication key
	return "shbucket_node_auth_" + uuid.New().String()
}

// pingNode performs an actual health check by calling the node's health endpoint
func (ctrl *NodeController) pingNode(node *entities.StorageNode) (bool, int64, string) {
	start := time.Now()
	
	// Create health check request to the node
	healthURL := strings.TrimSuffix(node.URL, "/") + "/api/v1/health"
	
	client := &http.Client{
		Timeout: 10 * time.Second, // 10 second timeout
	}
	
	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		responseTime := time.Since(start).Milliseconds()
		return false, responseTime, fmt.Sprintf("Failed to create request: %v", err)
	}
	
	// Add authentication if node has auth key
	if node.AuthKey != "" {
		req.Header.Set("X-API-Key", node.AuthKey)
	}
	
	resp, err := client.Do(req)
	responseTime := time.Since(start).Milliseconds()
	
	if err != nil {
		return false, responseTime, fmt.Sprintf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// Check if response is successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		
		return true, responseTime, ""
	}
	
	return false, responseTime, fmt.Sprintf("Node returned status %d", resp.StatusCode)
}