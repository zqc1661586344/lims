package model

import "time"

// User represents a system user.
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username   string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password   string `gorm:"size:128;not null" json:"-"`
	RealName   string `gorm:"size:64" json:"real_name"`
	Email      string `gorm:"size:128" json:"email"`
	Phone      string `gorm:"size:20" json:"phone"`
	DeptID     *uint  `gorm:"index" json:"dept_id"`
	Status     int    `gorm:"default:1" json:"status"` // 1=active, 0=disabled
	IsAdmin    bool   `gorm:"default:false" json:"is_admin"`
	LastLogin  *time.Time `json:"last_login"`

	// Associations
	Dept  *Dept  `gorm:"foreignKey:DeptID" json:"dept,omitempty"`
	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// UserRole is the join table for User <-> Role.
type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (User) TableName() string { return "users" }
func (UserRole) TableName() string { return "user_roles" }