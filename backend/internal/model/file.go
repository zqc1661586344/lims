package model

import "time"

type File struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ObjectName   string    `gorm:"size:255;uniqueIndex;not null" json:"object_name"`
	OriginalName string    `gorm:"size:255;not null" json:"original_name"`
	ContentType  string    `gorm:"size:100" json:"content_type"`
	Size         int64     `json:"size"`
	Category     string    `gorm:"size:50;index" json:"category"`
	BusinessType string    `gorm:"size:50;index" json:"business_type"`
	BusinessID   *uint     `gorm:"index" json:"business_id"`
	CreatedBy    *uint     `gorm:"index;constraint:OnDelete:SET NULL;references:users(id)" json:"created_by"`
}

func (File) TableName() string { return "files" }
