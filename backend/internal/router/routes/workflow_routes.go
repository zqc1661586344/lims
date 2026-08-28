package routes

import (
	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	"lims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterWorkflowRoutes registers workflow engine routes (Phase 5).
func RegisterWorkflowRoutes(r *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {
	wfH := handler.NewWorkflowHandler(db)

	group := r.Group("/workflow")
	group.Use(middleware.AuthMiddleware(cfg))
	group.Use(middleware.GormContextMiddleware(db))
	{
		// Process instance management
		group.POST("/instances", wfH.StartInstance)
		group.GET("/instances/:id", wfH.GetInstance)
		group.GET("/instances/:id/history", wfH.GetProcessHistory)

		// Task operations
		group.GET("/tasks/pending", wfH.GetPendingTasks)
		group.GET("/tasks/pending/user", wfH.GetPendingTasksByUser)
		group.POST("/tasks/:id/approve", wfH.ApproveTask)
		group.POST("/tasks/:id/reject", wfH.RejectTask)

		// Node definitions
		group.GET("/nodes", wfH.GetNodeDefinitions)
	}
}