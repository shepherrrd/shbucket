package models

import (
	"time"
)

type SetupStatusResponse struct {
	IsSetup   bool   `json:"is_setup"`
	SetupType string `json:"setup_type,omitempty"`
	NodeName  string `json:"node_name,omitempty"`
	Message   string `json:"message"`
}

type MasterSetupRequest struct {
	AdminUsername    string                 `json:"admin_username" validate:"required,min=3,max=50"`
	AdminEmail       string                 `json:"admin_email" validate:"required,email"`
	AdminPassword    string                 `json:"admin_password" validate:"required,min=6"`
	StoragePath      string                 `json:"storage_path" validate:"required"`
	MaxStorage       int64                  `json:"max_storage" validate:"min=1"`
	DefaultAuthRule  AuthRuleResponse       `json:"default_auth_rule"`
	DefaultSettings  BucketSettingsResponse `json:"default_settings"`
	JWTSecret        string                 `json:"jwt_secret,omitempty"`
	SystemName       string                 `json:"system_name" validate:"required,min=3,max=100"`
}

type NodeSetupRequest struct {
	MasterURL     string `json:"master_url" validate:"required,url"`
	NodeName      string `json:"node_name" validate:"required,min=3,max=100"`
	NodeAPIKey    string `json:"node_api_key" validate:"required,min=10"`
	StoragePath   string `json:"storage_path" validate:"required"`
	MaxStorage    int64  `json:"max_storage" validate:"min=1"`
	MasterAPIKey  string `json:"master_api_key" validate:"required"`
}

type SetupResponse struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	SetupType   string                 `json:"setup_type"`
	AdminUser   *UserResponse          `json:"admin_user,omitempty"`
	Node        *StorageNodeResponse   `json:"node,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type SystemInfoResponse struct {
	SystemName    string    `json:"system_name"`
	Version       string    `json:"version"`
	SetupType     string    `json:"setup_type"`
	StoragePath   string    `json:"storage_path"`
	MaxStorage    int64     `json:"max_storage"`
	UsedStorage   int64     `json:"used_storage"`
	FreeStorage   int64     `json:"free_storage"`
	NodeCount     int       `json:"node_count,omitempty"`
	IsHealthy     bool      `json:"is_healthy"`
	LastChecked   time.Time `json:"last_checked"`
}