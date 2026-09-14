package middleware

import (
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DeptScopeMiddleware(db *gorm.DB, allowedDeptCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) {
			c.Next()
			return
		}

		userDeptID := GetDeptIDVal(c)
		if userDeptID == 0 {
			utils.Forbidden(c, "无法获取用户部门信息")
			c.Abort()
			return
		}

		var deptCode string
		if err := db.Raw(`SELECT code FROM depts WHERE id = ?`, userDeptID).Scan(&deptCode).Error; err != nil {
			utils.InternalError(c, "部门校验失败")
			c.Abort()
			return
		}

		for _, allowed := range allowedDeptCodes {
			if deptCode == allowed {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "当前用户部门无权执行此操作")
		c.Abort()
	}
}
