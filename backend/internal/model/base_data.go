package model

import "time"

// TestItem represents a test/detection item (检测项目).
type TestItem struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Name       string  `gorm:"uniqueIndex:idx_item_name;size:200;not null" json:"name"`
	Code       string  `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Category   string  `gorm:"size:100" json:"category"`
	Unit       string  `gorm:"size:50" json:"unit"`
	Method     string  `gorm:"size:200" json:"method"`
	StandardID *uint   `gorm:"index" json:"standard_id"`
	Price      float64 `gorm:"type:decimal(10,2);default:0" json:"price"`
	Status     int     `gorm:"default:1" json:"status"` // 1=active, 0=disabled

	// Associations
	Standard *TestStandard `gorm:"foreignKey:StandardID" json:"standard,omitempty"`
}

// TestStandard represents a test standard/method (检测标准).
type TestStandard struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string    `gorm:"uniqueIndex:idx_std_name;size:200;not null" json:"name"`
	Code        string    `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Issuer      string    `gorm:"size:200" json:"issuer"`       // 发布机构
	Version     string    `gorm:"size:50" json:"version"`       // 版本号
	PublishDate time.Time `json:"publish_date"`
	FilePath    string    `gorm:"size:500" json:"file_path"`    // 标准文件路径（MinIO）
	Status      int       `gorm:"default:1" json:"status"`
}

// Equipment represents a piece of lab equipment (仪器设备).
type Equipment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name             string    `gorm:"uniqueIndex:idx_equip_name;size:200;not null" json:"name"`
	Code             string    `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Model            string    `gorm:"size:100" json:"model"`
	Factory          string    `gorm:"size:200" json:"factory"`          // 生产厂家
	CalibrationDate  time.Time `json:"calibration_date"`                 // 校准日期
	NextCalDate      time.Time `json:"next_cal_date"`                    // 下次校准日期
	Status           int       `gorm:"default:1" json:"status"`          // 1=normal, 0=maintenance, 2=retired
}

// Reagent represents a reagent/consumable (试剂耗材).
type Reagent struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name         string    `gorm:"uniqueIndex:idx_reagent_name;size:200;not null" json:"name"`
	Code         string    `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Spec         string    `gorm:"size:100" json:"spec"`           // 规格
	Manufacturer string    `gorm:"size:200" json:"manufacturer"`   // 生产厂商
	BatchNo      string    `gorm:"size:100" json:"batch_no"`       // 批号
	StockQty     float64   `gorm:"type:decimal(12,2);default:0" json:"stock_qty"`
	Unit         string    `gorm:"size:20" json:"unit"`            // 单位
	ExpireDate   time.Time `json:"expire_date"`
	Status       int       `gorm:"default:1" json:"status"`
}

func (TestItem) TableName() string     { return "test_items" }
func (TestStandard) TableName() string { return "test_standards" }
func (Equipment) TableName() string    { return "equipment" }
func (Reagent) TableName() string      { return "reagents" }