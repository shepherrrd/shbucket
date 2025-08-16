package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SetupConfig struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IsSetup      bool           `gorm:"not null;default:false" json:"is_setup"`
	SetupType    string         `gorm:"not null" json:"setup_type"` // "master" or "node"
	MasterURL    string         `gorm:"size:500" json:"master_url,omitempty"`
	NodeName     string         `gorm:"size:100" json:"node_name,omitempty"`
	StoragePath  string         `gorm:"size:500" json:"storage_path"`
	MaxStorage   int64          `gorm:"default:0" json:"max_storage"`
	ConfigData   datatypes.JSON `gorm:"type:jsonb" json:"config_data"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}