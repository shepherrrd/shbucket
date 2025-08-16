package entities

import (
	"time"
	"github.com/google/uuid"
)

type NodeFileMetadata struct {
	Id         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	BucketId   uuid.UUID `gorm:"type:uuid;not null" json:"bucket_id"`
	BucketName string    `gorm:"type:text;not null" json:"bucket_name"`
	Filename   string    `gorm:"type:text;not null" json:"filename"`
	Path       string    `gorm:"type:text;not null" json:"path"`
	Size       int64     `gorm:"type:bigint;not null" json:"size"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

