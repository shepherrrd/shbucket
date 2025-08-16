package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
	"shbucket/src/Models"
)

type MasterSetupCommand struct {
	AdminUsername    string                        `json:"admin_username" validate:"required,min=3,max=50"`
	AdminEmail       string                        `json:"admin_email" validate:"required,email"`
	AdminPassword    string                        `json:"admin_password" validate:"required,min=6"`
	StoragePath      string                        `json:"storage_path" validate:"required"`
	MaxStorage       int64                         `json:"max_storage" validate:"min=1"`
	DefaultAuthRule  models.AuthRuleResponse       `json:"default_auth_rule"`
	DefaultSettings  models.BucketSettingsResponse `json:"default_settings"`
	JWTSecret        string                        `json:"jwt_secret,omitempty"`
	SystemName       string                        `json:"system_name" validate:"required,min=3,max=100"`
}

type MasterSetupResponse struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	AdminUser   models.UserResponse    `json:"admin_user"`
	Config      map[string]interface{} `json:"config"`
}

type MasterSetupRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewMasterSetupRequestHandler(dbContext *persistence.AppDbContext) *MasterSetupRequestHandler {
	return &MasterSetupRequestHandler{
		dbContext: dbContext,
	}
}

func (h *MasterSetupRequestHandler) Handle(ctx context.Context, command *MasterSetupCommand) (*MasterSetupResponse, error) {
	// Check if already setup using GoNtext
	existingConfig, _ := h.dbContext.SetupConfigs.Where(&entities.SetupConfig{IsSetup: true}).FirstOrDefault()
	if existingConfig != nil {
		return nil, fmt.Errorf("system is already configured")
	}

	// Check if admin user already exists 
	existingUser, _ := h.dbContext.Users.Where(&entities.User{Email: command.AdminEmail}).
		OrField("Username", command.AdminUsername).FirstOrDefault()
	if existingUser != nil {
		return nil, fmt.Errorf("admin user already exists") 
	}

	// Create storage directory
	if err := os.MkdirAll(command.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Hash admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(command.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user
	adminUser := &entities.User{
		Username:     command.AdminUsername,
		Email:        command.AdminEmail,
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		IsActive:     true,
	}

	// Add user using GoNtext
	h.dbContext.Users.Add(*adminUser)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Set JWT secret in environment
	jwtSecret := command.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "shb-" + generateRandomString(32)
	}
	os.Setenv("JWT_SECRET", jwtSecret)

	// Create setup configuration
	configData := map[string]interface{}{
		"system_name":        command.SystemName,
		"jwt_secret":         jwtSecret,
		"default_auth_rule":  command.DefaultAuthRule,
		"default_settings":   command.DefaultSettings,
		"admin_user_id":      adminUser.Id.String(),
	}
	
	configDataJSON, err := json.Marshal(configData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config data: %w", err)
	}

	config := &entities.SetupConfig{
		IsSetup:     true,
		SetupType:   "master",
		StoragePath: command.StoragePath,
		MaxStorage:  command.MaxStorage,
		ConfigData:  datatypes.JSON(configDataJSON),
	}

	// Save setup configuration using GoNtext
	h.dbContext.SetupConfigs.Add(*config)
	if err := h.dbContext.SaveChanges(); err != nil {
		return nil, fmt.Errorf("failed to save setup configuration: %w", err)
	}

	adminResponse := models.UserResponse{
		ID:        adminUser.Id,
		Username:  adminUser.Username,
		Email:     adminUser.Email,
		Role:      adminUser.Role,
		IsActive:  adminUser.IsActive,
		CreatedAt: adminUser.CreatedAt,
		UpdatedAt: adminUser.UpdatedAt,
	}

	return &MasterSetupResponse{
		Success:   true,
		Message:   "Master setup completed successfully",
		AdminUser: adminResponse,
		Config: map[string]interface{}{
			"system_name":   command.SystemName,
			"setup_type":    "master",
			"storage_path":  command.StoragePath,
			"max_storage":   command.MaxStorage,
		},
	}, nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(charset)/2] // Simple implementation
	}
	return string(b)
}