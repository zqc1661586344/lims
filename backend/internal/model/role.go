package model

import "time"

// Role represents a role with associated permissions.
type Role struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name   string `gorm:"size:64;not null" json:"name"`
	Code   string `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Status int    `gorm:"default:1" json:"status"`
	Remark string `gorm:"size:256" json:"remark"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// RolePermission is the join table for Role <-> Permission.
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

func (Role) TableName() string { return "roles" }
func (RolePermission) TableName() string { return "role_permissions" }