package routes

import (
	"time"

	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	system "lims-backend/internal/handler/system"
	"lims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterSystemRoutes registers RBAC-related routes.
func RegisterSystemRoutes(r *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {
	loginLimiter := middleware.NewRateLimiter(5*time.Minute, 20)
	auth := handler.NewAuthHandler(cfg, db, loginLimiter)

	// Public auth routes (no token required)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", loginLimiter.Middleware(), auth.Login)
	}

	// Authenticated routes
	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware(cfg, db))
	{
		protected.GET("/auth/profile", auth.Profile)
		protected.POST("/auth/logout", auth.Logout)
		protected.POST("/auth/refresh", auth.Refresh)
	}

	// System management routes (authenticated + admin by default)
	userH := system.NewUserHandler(db)
	deptH := system.NewDeptHandler(db)
	roleH := system.NewRoleHandler(db)
	permH := system.NewPermissionHandler(db)

	systemGroup := r.Group("/system")
	systemGroup.Use(middleware.AuthMiddleware(cfg, db))
	systemGroup.Use(middleware.RequireAdmin())
	systemGroup.Use(middleware.GormContextMiddleware(db))
	{
		// Users
		systemGroup.GET("/users", userH.List)
		systemGroup.POST("/users", userH.Create)
		systemGroup.GET("/users/:id", userH.Get)
		systemGroup.PUT("/users/:id", userH.Update)
		systemGroup.DELETE("/users/:id", userH.Delete)
		systemGroup.PUT("/users/:id/roles", userH.UpdateRoles)

		// Departments
		systemGroup.GET("/depts", deptH.List)
		systemGroup.POST("/depts", deptH.Create)
		systemGroup.GET("/depts/:id", deptH.Get)
		systemGroup.PUT("/depts/:id", deptH.Update)
		systemGroup.DELETE("/depts/:id", deptH.Delete)

		// Roles
		systemGroup.GET("/roles", roleH.List)
		systemGroup.POST("/roles", roleH.Create)
		systemGroup.GET("/roles/:id", roleH.Get)
		systemGroup.PUT("/roles/:id", roleH.Update)
		systemGroup.DELETE("/roles/:id", roleH.Delete)
		systemGroup.PUT("/roles/:id/permissions", roleH.UpdatePermissions)

		// Permissions
		systemGroup.GET("/permissions", permH.List)
		systemGroup.POST("/permissions", permH.Create)
		systemGroup.PUT("/permissions/:id", permH.Update)
		systemGroup.DELETE("/permissions/:id", permH.Delete)
	}
}
