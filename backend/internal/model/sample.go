package model

import "time"

type Sample struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	TaskOrderID     uint       `gorm:"index;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	SampleCode      string     `gorm:"size:100;uniqueIndex;not null" json:"sample_code"`
	SampleName      string     `gorm:"size:200" json:"sample_name"`
	SampleType      string     `gorm:"size:100" json:"sample_type"`
	Quantity        float64    `gorm:"type:decimal(12,2);default:0" json:"quantity"`
	Unit            string     `gorm:"size:20" json:"unit"`
	Status          string     `gorm:"size:20;default:'pending'" json:"status"`
	StorageLocation string     `gorm:"size:200" json:"storage_location"`
	RetainUntil     *time.Time `json:"retain_until"`
	DisposedAt      *time.Time `json:"disposed_at"`
	DisposedBy      *uint      `gorm:"constraint:OnDelete:SET NULL;references:users(id)" json:"disposed_by"`
	Remark          string     `gorm:"type:text" json:"remark"`
}

func (Sample) TableName() string { return "samples" }

type TaskOrderTestItem struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TaskOrderID uint      `gorm:"index;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	TestItemID  *uint     `gorm:"index;constraint:OnDelete:SET NULL;references:test_items(id)" json:"test_item_id"`
	ItemName    string    `gorm:"size:200;not null" json:"item_name"`
	ItemCode    string    `gorm:"size:64" json:"item_code"`
	Standard    string    `gorm:"size:200" json:"standard"`
	Unit        string    `gorm:"size:50" json:"unit"`
	Method      string    `gorm:"size:200" json:"method"`
	Result      string    `gorm:"size:200" json:"result"`
	Remark      string    `gorm:"type:text" json:"remark"`
}

func (TaskOrderTestItem) TableName() string { return "task_order_test_items" }

type SamplingPoint struct {
	ID                 uint      `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	SamplingScheduleID uint      `gorm:"index;not null;constraint:OnDelete:CASCADE;references:sampling_schedules(id)" json:"sampling_schedule_id"`
	PointName          string    `gorm:"size:200;not null" json:"point_name"`
	PointCode          string    `gorm:"size:64" json:"point_code"`
	Location           string    `gorm:"size:300" json:"location"`
	Longitude          float64   `gorm:"type:decimal(10,6)" json:"longitude"`
	Latitude           float64   `gorm:"type:decimal(10,6)" json:"latitude"`
	SamplingMethod     string    `gorm:"size:100" json:"sampling_method"`
	SampleCount        int       `gorm:"default:1" json:"sample_count"`
	Remark             string    `gorm:"type:text" json:"remark"`
}

func (SamplingPoint) TableName() string { return "sampling_points" }
