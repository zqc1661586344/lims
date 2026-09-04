package model

import (
	"encoding/json"
	"time"
)

// LabSheetTemplate represents a reusable Univer Sheet template for a specific test item.
// Each template defines the workbook structure (rows, columns, formulas, merged cells)
// and which ranges are editable vs read-only.
type LabSheetTemplate struct {
	ID             uint            `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Code           string          `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name           string          `gorm:"size:200;not null" json:"name"`
	TestItemID     uint            `gorm:"index" json:"test_item_id"`
	NodeCode       string          `gorm:"size:64" json:"node_code"`
	Structure      json.RawMessage `gorm:"type:jsonb;not null" json:"structure"`
	EditableRanges json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"editable_ranges"`
	ReadOnlyRanges json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"readonly_ranges"`
	Version        int             `gorm:"default:1" json:"version"`
	Status         int             `gorm:"default:1" json:"status"` // 1=active 0=disabled
}

func (LabSheetTemplate) TableName() string { return "lab_sheet_templates" }

// LabSheet represents an instance of a test sheet filled in for a specific task order.
// Created by the workflow engine when a process advances to node_data_entry,
// or manually created when a user picks a template and starts filling data.
type LabSheet struct {
	ID             uint            `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	TaskOrderID    uint            `gorm:"index" json:"task_order_id"`
	TestItemID     uint            `gorm:"index" json:"test_item_id"`
	TemplateID     *uint           `gorm:"index" json:"template_id"`
	NodeCode       string          `gorm:"size:64" json:"node_code"`
	SheetData      json.RawMessage `gorm:"type:jsonb;not null" json:"sheet_data"`
	FormulaResults json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"formula_results"`
	Status         int             `gorm:"default:0;index" json:"status"` // 0=draft 1=submitted 2=reviewed 3=audited
	CreatedBy      *uint           `json:"created_by"`
	UpdatedBy      *uint           `json:"updated_by"`

	Template *LabSheetTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

func (LabSheet) TableName() string { return "lab_sheets" }
