package model

import "time"

// Dept represents a department (7 departments in the workflow).
type Dept struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name     string `gorm:"size:64;not null" json:"name"`
	Code     string `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int    `gorm:"default:1" json:"status"`
	ParentID *uint  `gorm:"index" json:"parent_id"`

	Children []Dept `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (Dept) TableName() string { return "depts" }