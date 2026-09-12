package model

import (
	"time"
)

// AuditLog records data changes for CNAS compliance.
type AuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AffectedTable string    `gorm:"column:affected_table;size:100;not null;index" json:"affected_table"`
	RecordID      uint      `gorm:"not null;index" json:"record_id"`
	Action        string    `gorm:"size:20;not null" json:"action"` // CREATE / UPDATE / DELETE
	OperatorID    uint      `gorm:"index;constraint:OnDelete:SET NULL;references:users(id)" json:"operator_id"`
	Operator      string    `gorm:"size:50" json:"operator"`
	OldData       string    `gorm:"type:jsonb" json:"old_data"` // previous state
	NewData       string    `gorm:"type:jsonb" json:"new_data"` // new state
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
