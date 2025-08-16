package models

import (
	"time"
	"github.com/google/uuid"
)

type StorageNodeResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	MaxStorage  int64      `json:"max_storage"`
	UsedStorage int64      `json:"used_storage"`
	Priority    int        `json:"priority"`
	IsActive    bool       `json:"is_active"`
	IsHealthy   bool       `json:"is_healthy"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastPing    *time.Time `json:"last_ping,omitempty"`
}

type RegisterNodeRequest struct {
	Name       string `json:"name" validate:"required,min=3,max=100"`
	URL        string `json:"url" validate:"required,url"`
	AuthKey    string `json:"auth_key" validate:"required,min=32"`
	MaxStorage int64  `json:"max_storage" validate:"min=0"`
	Priority   int    `json:"priority" validate:"min=0,max=100"`
	IsActive   bool   `json:"is_active"`
}

type UpdateNodeRequest struct {
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	StoragePath *string   `json:"storage_path,omitempty"`
	MaxStorage  *int64    `json:"max_storage,omitempty" validate:"omitempty,min=0"`
	IsActive    *bool     `json:"is_active,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
}

type NodeHealthCheckRequest struct {
	NodeID uuid.UUID `json:"node_id" validate:"required"`
}

type NodeHealthCheckResponse struct {
	NodeID      uuid.UUID `json:"node_id"`
	IsHealthy   bool      `json:"is_healthy"`
	ResponseTime int64     `json:"response_time_ms"`
	Error       string    `json:"error,omitempty"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
}

type NodeInstallationRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	StoragePath string `json:"storage_path" validate:"required"`
	MaxStorage  int64  `json:"max_storage" validate:"min=1"`
	Port        int    `json:"port" validate:"min=1,max=65535"`
	APIKey      string `json:"api_key,omitempty"`
}

type NodeInstallationResponse struct {
	Node         StorageNodeResponse `json:"node"`
	InstallPath  string              `json:"install_path"`
	ConfigPath   string              `json:"config_path"`
	StartCommand string              `json:"start_command"`
	Success      bool                `json:"success"`
	Message      string              `json:"message"`
}