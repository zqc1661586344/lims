package routes

import (
	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	"lims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterBaseDataRoutes registers base data management routes (Phase 4).
func RegisterBaseDataRoutes(r *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {
	itemH := handler.NewTestItemHandler(db)
	stdH := handler.NewTestStandardHandler(db)
	equipH := handler.NewEquipmentHandler(db)
	reagentH := handler.NewReagentHandler(db)

	group := r.Group("/base-data")
	group.Use(middleware.AuthMiddleware(cfg))
	group.Use(middleware.GormContextMiddleware(db))
	{
		// Test Items
		group.GET("/items", itemH.List)
		group.POST("/items", itemH.Create)
		group.GET("/items/:id", itemH.Get)
		group.PUT("/items/:id", itemH.Update)
		group.DELETE("/items/:id", itemH.Delete)

		// Test Standards
		group.GET("/standards", stdH.List)
		group.POST("/standards", stdH.Create)
		group.GET("/standards/:id", stdH.Get)
		group.PUT("/standards/:id", stdH.Update)
		group.DELETE("/standards/:id", stdH.Delete)

		// Equipment
		group.GET("/equipment", equipH.List)
		group.POST("/equipment", equipH.Create)
		group.GET("/equipment/:id", equipH.Get)
		group.PUT("/equipment/:id", equipH.Update)
		group.DELETE("/equipment/:id", equipH.Delete)

		// Reagents
		group.GET("/reagents", reagentH.List)
		group.POST("/reagents", reagentH.Create)
		group.GET("/reagents/:id", reagentH.Get)
		group.PUT("/reagents/:id", reagentH.Update)
		group.DELETE("/reagents/:id", reagentH.Delete)
	}
}