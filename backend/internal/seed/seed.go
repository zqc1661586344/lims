package seed

import (
	"lims-backend/internal/model"
	"lims-backend/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Run initializes the database with required default data.
func Run(db *gorm.DB, logger *zap.Logger) {
	seedDepts(db, logger)
	seedPermissions(db, logger)
	seedRoles(db, logger)
	seedAdminUser(db, logger)
	seedAdminRoleBindings(db, logger)
}

func seedDepts(db *gorm.DB, logger *zap.Logger) {
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
}

func seedPermissions(db *gorm.DB, logger *zap.Logger) {
	perms := []model.Permission{
		{Name: "委托管理", Code: "business:task-order", Type: "api", Sort: 1},
		{Name: "合同评审", Code: "business:contract-review", Type: "api", Sort: 2},
		{Name: "质控任务", Code: "business:qc-task", Type: "api", Sort: 3},
		{Name: "采样计划", Code: "business:sampling-schedule", Type: "api", Sort: 4},
		{Name: "现场采样", Code: "business:field-sampling", Type: "api", Sort: 5},
		{Name: "样品接收", Code: "business:sample-receiving", Type: "api", Sort: 6},
		{Name: "任务分配", Code: "business:task-assign", Type: "api", Sort: 7},
		{Name: "数据录入", Code: "business:data-entry", Type: "api", Sort: 8},
		{Name: "数据复核", Code: "business:data-review", Type: "api", Sort: 9},
		{Name: "数据审核", Code: "business:data-audit", Type: "api", Sort: 10},
		{Name: "报告编制", Code: "business:report-prepare", Type: "api", Sort: 11},
		{Name: "报告复核", Code: "business:report-review", Type: "api", Sort: 12},
		{Name: "报告审核", Code: "business:report-audit", Type: "api", Sort: 13},
		{Name: "报告签发", Code: "business:report-sign", Type: "api", Sort: 14},
		{Name: "报告打印发放", Code: "business:report-print", Type: "api", Sort: 15},
		{Name: "项目归档", Code: "business:project-archive", Type: "api", Sort: 16},
	}

	for _, p := range perms {
		var count int64
		db.Model(&model.Permission{}).Where("code = ?", p.Code).Count(&count)
		if count == 0 {
			db.Create(&p)
			logger.Info("Seeded permission", zap.String("code", p.Code))
		}
	}
}

func seedRoles(db *gorm.DB, logger *zap.Logger) {
	roles := []struct {
		role  model.Role
		codes []string
	}{
		{
			role:  model.Role{Name: "超级管理员", Code: "super_admin", Remark: "拥有全部业务权限"},
			codes: []string{"*"},
		},
		{
			role: model.Role{Name: "实验室技术员", Code: "lab_technician", Remark: "负责数据录入与质控执行"},
			codes: []string{
				"business:qc-task", "business:task-assign", "business:data-entry",
			},
		},
		{
			role: model.Role{Name: "采样员", Code: "sampler", Remark: "负责采样计划与现场采样"},
			codes: []string{
				"business:sampling-schedule", "business:field-sampling",
			},
		},
		{
			role:  model.Role{Name: "样品管理员", Code: "sample_manager", Remark: "负责样品接收"},
			codes: []string{"business:sample-receiving"},
		},
		{
			role:  model.Role{Name: "数据复核人", Code: "data_reviewer", Remark: "负责数据复核（与录入人职责分离）"},
			codes: []string{"business:data-review"},
		},
		{
			role:  model.Role{Name: "数据审核人", Code: "data_auditor", Remark: "负责数据审核（与复核人职责分离）"},
			codes: []string{"business:data-audit"},
		},
		{
			role:  model.Role{Name: "报告编制人", Code: "report_preparer", Remark: "负责报告编制"},
			codes: []string{"business:report-prepare"},
		},
		{
			role:  model.Role{Name: "报告复核人", Code: "report_reviewer", Remark: "负责报告复核（与编制人职责分离）"},
			codes: []string{"business:report-review"},
		},
		{
			role:  model.Role{Name: "报告审核人", Code: "report_auditor", Remark: "负责报告审核（与复核人职责分离）"},
			codes: []string{"business:report-audit"},
		},
		{
			role:  model.Role{Name: "授权签字人", Code: "authorized_signer", Remark: "负责报告签发"},
			codes: []string{"business:report-sign"},
		},
		{
			role:  model.Role{Name: "档案管理员", Code: "archive_manager", Remark: "负责报告打印发放与项目归档"},
			codes: []string{"business:report-print", "business:project-archive"},
		},
		{
			role:  model.Role{Name: "业务经理", Code: "business_manager", Remark: "负责委托单与合同评审"},
			codes: []string{"business:task-order", "business:contract-review"},
		},
	}

	for _, entry := range roles {
		var role model.Role
		result := db.Where("code = ?", entry.role.Code).First(&role)
		if result.Error == gorm.ErrRecordNotFound {
			db.Create(&entry.role)
			db.Where("code = ?", entry.role.Code).First(&role)
			logger.Info("Seeded role", zap.String("code", entry.role.Code))
		} else if result.Error != nil {
			logger.Error("Failed to query role", zap.Error(result.Error))
			continue
		}

		if len(entry.codes) == 1 && entry.codes[0] == "*" {
			var allPerms []model.Permission
			db.Find(&allPerms)
			if len(allPerms) > 0 {
				db.Model(&role).Association("Permissions").Replace(allPerms)
			}
			continue
		}

		var perms []model.Permission
		db.Where("code IN ?", entry.codes).Find(&perms)
		if len(perms) > 0 {
			db.Model(&role).Association("Permissions").Replace(perms)
		}
	}
}

func seedAdminUser(db *gorm.DB, logger *zap.Logger) {
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

func seedAdminRoleBindings(db *gorm.DB, logger *zap.Logger) {
	var admin model.User
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		return
	}
	var superRole model.Role
	if err := db.Where("code = ?", "super_admin").First(&superRole).Error; err != nil {
		return
	}

	var existing int64
	db.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ?", admin.ID, superRole.ID).Count(&existing)
	if existing == 0 {
		db.Create(&model.UserRole{UserID: admin.ID, RoleID: superRole.ID})
		logger.Info("Bound admin to super_admin role")
	}
}
