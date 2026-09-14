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
	seedUsers(db, logger)
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
		{Name: "流程查看", Code: "workflow:view", Type: "api", Sort: 20},
		{Name: "流程审批", Code: "workflow:task", Type: "api", Sort: 21},
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
			role:  model.Role{Name: "超级管理员", Code: "super_admin", Remark: "拥有全部业务权限（不受部门约束）"},
			codes: []string{"*"},
		},
		{
			role: model.Role{Name: "业务经理", Code: "business_manager", Remark: "业务室：委托单创建、报告发放"},
			codes: []string{
				"business:task-order", "business:report-print",
				"workflow:view", "workflow:task",
			},
		},
		{
			role:  model.Role{Name: "合同评审员", Code: "contract_reviewer", Remark: "技术室：合同评审"},
			codes: []string{"business:contract-review", "workflow:view", "workflow:task"},
		},
		{
			role:  model.Role{Name: "授权签字人", Code: "authorized_signer", Remark: "技术室：报告签发"},
			codes: []string{"business:report-sign", "workflow:view", "workflow:task"},
		},
		{
			role:  model.Role{Name: "质控任务员", Code: "qc_staff", Remark: "质控室：质控任务分派"},
			codes: []string{"business:qc-task", "workflow:view", "workflow:task"},
		},
		{
			role:  model.Role{Name: "报告审核人", Code: "report_auditor", Remark: "质控室：报告审核"},
			codes: []string{"business:report-audit", "workflow:view", "workflow:task"},
		},
		{
			role: model.Role{Name: "采样员", Code: "sampler", Remark: "现场室：采样计划与现场采样"},
			codes: []string{
				"business:sampling-schedule", "business:field-sampling",
				"workflow:view", "workflow:task",
			},
		},
		{
			role:  model.Role{Name: "样品管理员", Code: "sample_manager", Remark: "样品室：样品接收"},
			codes: []string{"business:sample-receiving", "workflow:view"},
		},
		{
			role: model.Role{Name: "实验室技术员", Code: "lab_technician", Remark: "实验室：任务分配、数据录入"},
			codes: []string{
				"business:task-assign", "business:data-entry",
				"workflow:view", "workflow:task",
			},
		},
		{
			role:  model.Role{Name: "数据复核人", Code: "data_reviewer", Remark: "实验室：数据复核（与录入人职责分离）"},
			codes: []string{"business:data-review", "workflow:view", "workflow:task"},
		},
		{
			role:  model.Role{Name: "数据审核人", Code: "data_auditor", Remark: "实验室：数据审核（与复核人职责分离）"},
			codes: []string{"business:data-audit", "workflow:view", "workflow:task"},
		},
		{
			role:  model.Role{Name: "报告复核人", Code: "report_reviewer", Remark: "实验室：报告复核（与编制人职责分离）"},
			codes: []string{"business:report-review", "workflow:view", "workflow:task"},
		},
		{
			role:  model.Role{Name: "报告编制人", Code: "report_preparer", Remark: "报告室：报告编制"},
			codes: []string{"business:report-prepare", "workflow:view", "workflow:task"},
		},
		{
			role:  model.Role{Name: "档案管理员", Code: "archive_manager", Remark: "报告室：项目归档"},
			codes: []string{"business:project-archive", "workflow:view"},
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

// seedUsers creates one user per (department, role) pairing so that
// resolveAssigneeForNode can actually find someone for every workflow node.
// Passwords are "<username>123" — change in production.
func seedUsers(db *gorm.DB, logger *zap.Logger) {
	type userSeed struct {
		username string
		realName string
		deptCode string
		roleCode string
	}

	seeds := []userSeed{
		{"biz_mgr", "王经理", "dept_business", "business_manager"},
		{"tech_reviewer", "李评审", "dept_tech", "contract_reviewer"},
		{"signer", "张授权", "dept_tech", "authorized_signer"},
		{"qc_staff", "赵质控", "dept_qc", "qc_staff"},
		{"qc_auditor", "陈审核", "dept_qc", "report_auditor"},
		{"sampler", "刘采样", "dept_field", "sampler"},
		{"sample_mgr", "周样品", "dept_sample", "sample_manager"},
		{"lab_tech", "孙技术员", "dept_lab", "lab_technician"},
		{"data_reviewer", "吴复核", "dept_lab", "data_reviewer"},
		{"data_auditor", "郑审核", "dept_lab", "data_auditor"},
		{"report_reviewer", "钱复核", "dept_lab", "report_reviewer"},
		{"report_preparer", "冯编制", "dept_report", "report_preparer"},
		{"archivist", "褚档案", "dept_report", "archive_manager"},
	}

	for _, s := range seeds {
		var existing model.User
		if err := db.Where("username = ?", s.username).First(&existing).Error; err == nil {
			continue
		}

		var dept model.Dept
		if err := db.Where("code = ?", s.deptCode).First(&dept).Error; err != nil {
			logger.Warn("seedUsers: dept not found, skipping",
				zap.String("dept_code", s.deptCode))
			continue
		}

		var role model.Role
		if err := db.Where("code = ?", s.roleCode).First(&role).Error; err != nil {
			logger.Warn("seedUsers: role not found, skipping",
				zap.String("role_code", s.roleCode))
			continue
		}

		hashed, err := utils.HashPassword(s.username + "123")
		if err != nil {
			logger.Error("seedUsers: password hash failed",
				zap.String("username", s.username), zap.Error(err))
			continue
		}

		user := model.User{
			Username: s.username,
			Password: hashed,
			RealName: s.realName,
			DeptID:   &dept.ID,
			Status:   1,
			IsAdmin:  false,
		}
		if err := db.Create(&user).Error; err != nil {
			logger.Error("seedUsers: create user failed",
				zap.String("username", s.username), zap.Error(err))
			continue
		}
		db.Create(&model.UserRole{UserID: user.ID, RoleID: role.ID})
		logger.Info("Seeded user",
			zap.String("username", s.username),
			zap.String("dept", s.deptCode),
			zap.String("role", s.roleCode))
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
