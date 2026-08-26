package middleware

import (
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionMiddleware checks if the current user has the required permission code.
// Pass one or more permission codes — user needs at least one to pass.
func PermissionMiddleware(db *gorm.DB, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) {
			c.Next()
			return
		}

		userID := GetUserID(c)
		if userID == 0 {
			utils.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		hasPermission, err := checkUserPermission(db, userID, codes)
		if err != nil || !hasPermission {
			utils.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkUserPermission queries whether the user has any of the given permission codes.
func checkUserPermission(db *gorm.DB, userID uint, codes []string) (bool, error) {
	if len(codes) == 0 {
		return true, nil
	}

	var count int64
	err := db.Table("users").
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN role_permissions ON role_permissions.role_id = user_roles.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("users.id = ? AND permissions.code IN ?", userID, codes).
		Count(&count).Error

	return count > 0, err
}