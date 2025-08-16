package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	
	"gorm.io/datatypes"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type NodeSetupCommand struct {
	MasterURL     string `json:"master_url" validate:"required,url"`
	NodeName      string `json:"node_name" validate:"required,min=3,max=100"`
	NodeAPIKey    string `json:"node_api_key" validate:"required,min=10"`
	StoragePath   string `json:"storage_path" validate:"required"`
	MaxStorage    int64  `json:"max_storage" validate:"min=1"`
	MasterAPIKey  string `json:"master_api_key" validate:"required"`
}

type NodeSetupResponse struct {
	Success bool                        `json:"success"`
	Message string                      `json:"message"`
	Node    models.StorageNodeResponse  `json:"node"`
	Config  map[string]interface{}      `json:"config"`
}

type NodeSetupRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewNodeSetupRequestHandler(dbContext *persistence.AppDbContext) *NodeSetupRequestHandler {
	return &NodeSetupRequestHandler{
		dbContext: dbContext,
	}
}

func (h *NodeSetupRequestHandler) Handle(ctx context.Context, command *NodeSetupCommand) (*NodeSetupResponse, error) {
	// Check if already setup using GoNtext
	existingConfig, _ := h.dbContext.SetupConfigs.Where(&entities.SetupConfig{IsSetup: true}).FirstOrDefault()
	if existingConfig != nil {
		return nil, fmt.Errorf("system is already configured")
	}

	// Create storage directory
	if err := os.MkdirAll(command.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Register with master server using self-registration endpoint (no auth required)
	nodeRegistration := map[string]interface{}{
		"name":        command.NodeName,
		"url":         getCurrentServerURL(),
		"max_storage": command.MaxStorage,
		"priority":    1,
	}

	registrationJSON, _ := json.Marshal(nodeRegistration)
	
	masterURL := strings.TrimSuffix(command.MasterURL, "/")
	req, err := http.NewRequest("POST", masterURL+"/api/v1/node/register", strings.NewReader(string(registrationJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register with master server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("master server rejected node registration: status %d", resp.StatusCode)
	}

	var masterResponse struct {
		Node    models.StorageNodeResponse `json:"node"`
		AuthKey string                     `json:"auth_key"`
		Success bool                       `json:"success"`
		Message string                     `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&masterResponse); err != nil {
		return nil, fmt.Errorf("failed to decode master response: %w", err)
	}

	if !masterResponse.Success {
		return nil, fmt.Errorf("master server registration failed: %s", masterResponse.Message)
	}

	// Save local node configuration
	configData := map[string]interface{}{
		"node_api_key":    command.NodeAPIKey,
		"master_api_key":  command.MasterAPIKey,
		"node_auth_key":   masterResponse.AuthKey, // Auth key from master for node identification
		"node_id":         masterResponse.Node.ID.String(),
		"registered_at":   time.Now().Unix(),
	}
	
	configDataJSON, err := json.Marshal(configData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config data: %w", err)
	}

	config := &entities.SetupConfig{
		IsSetup:     true,
		SetupType:   "node",
		MasterURL:   command.MasterURL,
		NodeName:    command.NodeName,
		StoragePath: command.StoragePath,
		MaxStorage:  command.MaxStorage,
		ConfigData:  datatypes.JSON(configDataJSON),
	}

	// Save node configuration using GoNtext
	h.dbContext.SetupConfigs.Add(*config)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to save node configuration: %w", err)
	}

	return &NodeSetupResponse{
		Success: true,
		Message: "Node setup completed and registered with master",
		Node:    masterResponse.Node,
		Config: map[string]interface{}{
			"setup_type":   "node",
			"master_url":   command.MasterURL,
			"node_name":    command.NodeName,
			"storage_path": command.StoragePath,
			"max_storage":  command.MaxStorage,
		},
	}, nil
}

func getCurrentServerURL() string {
	host := os.Getenv("HOST")
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	return fmt.Sprintf("http://%s:%s", host, port)
}