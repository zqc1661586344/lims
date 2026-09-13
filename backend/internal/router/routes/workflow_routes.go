package routes

import (
	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	"lims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterWorkflowRoutes registers workflow engine routes (Phase 5).
func RegisterWorkflowRoutes(r *gin.RouterGroup, cfg *config.Config, logger *zap.Logger, db *gorm.DB) {
	wfH := handler.NewWorkflowHandler(logger, db)

	group := r.Group("/workflow")
	group.Use(middleware.AuthMiddleware(cfg, db))
	group.Use(middleware.GormContextMiddleware(db))
	{
		group.POST("/instances", wfH.StartInstance)

		viewG := group.Group("")
		viewG.Use(middleware.PermissionMiddleware(db, "workflow:view"))
		{
			viewG.GET("/instances/:id", wfH.GetInstance)
			viewG.GET("/instances/:id/history", wfH.GetProcessHistory)
			viewG.GET("/tasks/pending", wfH.GetPendingTasks)
			viewG.GET("/tasks/pending/user", wfH.GetPendingTasksByUser)
			viewG.GET("/nodes", wfH.GetNodeDefinitions)
			viewG.GET("/progress/:businessType/:businessId", wfH.GetProgress)
		}

		taskG := group.Group("")
		taskG.Use(middleware.PermissionMiddleware(db, "workflow:task"))
		{
			taskG.POST("/tasks/:id/approve", wfH.ApproveTask)
			taskG.POST("/tasks/:id/reject", wfH.RejectTask)
		}
	}
}
