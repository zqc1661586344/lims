package seed

import (
	"lims-backend/internal/model"
	"lims-backend/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Run initializes the database with required default data.
func Run(db *gorm.DB, logger *zap.Logger) {
	// Create 7 departments
	depts := []model.Dept{
		{Name: "业务室", Code: "dept_business", Sort: 1},
		{Name: "技术室", Code: "dept_tech", Sort: 2},
		{Name: "报告室", Code: "dept_report", Sort: 3},
		{Name: "现场室", Code: "dept_field", Sort: 4},
		{Name: "样品室", Code: "dept_sample", Sort: 5},
		{Name: "实验室", Code: "dept_lab", Sort: 6},
		{Name: "质控室", Code: "dept_qc", Sort: 7},
	}

	for _, d := range depts {
		var count int64
		db.Model(&model.Dept{}).Where("code = ?", d.Code).Count(&count)
		if count == 0 {
			db.Create(&d)
			logger.Info("Seeded department", zap.String("name", d.Name))
		}
	}

	// Create default admin user
	var count int64
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		hashed, err := utils.HashPassword("admin123")
		if err != nil {
			logger.Error("Failed to hash admin password", zap.Error(err))
			return
		}

		var dept model.Dept
		db.Where("code = ?", "dept_business").First(&dept)

		admin := model.User{
			Username: "admin",
			Password: hashed,
			RealName: "系统管理员",
			DeptID:   &dept.ID,
			Status:   1,
			IsAdmin:  true,
		}
		db.Create(&admin)
		logger.Info("Seeded admin user: admin/admin123")
	}
}