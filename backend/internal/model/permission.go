package model

import "time"

// Permission represents a button-level or route-level permission.
type Permission struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name     string `gorm:"size:64;not null" json:"name"`
	Code     string `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Type     string `gorm:"size:20;not null" json:"type"` // menu, button, api
	ParentID *uint  `gorm:"index" json:"parent_id"`
	Path     string `gorm:"size:128" json:"path"`
	Icon     string `gorm:"size:64" json:"icon"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int    `gorm:"default:1" json:"status"`

	Children []Permission `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (Permission) TableName() string { return "permissions" }