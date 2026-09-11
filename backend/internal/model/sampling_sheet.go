package model

import (
	"encoding/json"
	"time"
)

type SamplingSheetTemplate struct {
	ID             uint            `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Code           string          `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name           string          `gorm:"size:200;not null" json:"name"`
	SampleType     string          `gorm:"size:100" json:"sample_type"`
	NodeCode       string          `gorm:"size:64" json:"node_code"`
	Structure      json.RawMessage `gorm:"type:jsonb;not null" json:"structure"`
	EditableRanges json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"editable_ranges"`
	ReadOnlyRanges json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"readonly_ranges"`
	Version        int             `gorm:"default:1" json:"version"`
	Status         int             `gorm:"default:1" json:"status"`
}

func (SamplingSheetTemplate) TableName() string { return "sampling_sheet_templates" }

type SamplingSheet struct {
	ID             uint                   `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	TaskOrderID    uint                   `gorm:"index;not null" json:"task_order_id"`
	SamplingPoint  string                 `gorm:"size:200" json:"sampling_point"`
	TemplateID     *uint                  `gorm:"index" json:"template_id"`
	NodeCode       string                 `gorm:"size:64" json:"node_code"`
	SheetData      json.RawMessage        `gorm:"type:jsonb;not null" json:"sheet_data"`
	FormulaResults json.RawMessage        `gorm:"type:jsonb;default:'{}'" json:"formula_results"`
	Status         int                    `gorm:"default:0;index" json:"status"`
	CreatedBy      *uint                  `json:"created_by"`
	UpdatedBy      *uint                  `json:"updated_by"`
	Template       *SamplingSheetTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

func (SamplingSheet) TableName() string { return "sampling_sheets" }
