package router

import (
	"lims-backend/internal/config"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/router/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup configures the Gin router with middleware and routes.
func Setup(cfg *config.Config, logger *zap.Logger, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Auto-migrate database tables (skip in production to avoid unintended schema changes)
	if cfg.Env != "production" {
		autoMigrate(logger, db)
	} else {
		logger.Warn("skipping AutoMigrate in production environment")
	}

	// Register audit plugin (GORM hooks for data change tracking)
	auditPlugin := middleware.NewAuditPlugin(logger)
	if err := db.Use(auditPlugin); err != nil {
		logger.Fatal("Failed to register audit plugin", zap.Error(err))
	}

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
	}))

	// API routes
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "LIMS system is running",
			})
		})

		// Register RBAC routes
		routes.RegisterSystemRoutes(api, cfg, db)
		// Register base data routes
		routes.RegisterBaseDataRoutes(api, cfg, db)
		// Register workflow routes
		routes.RegisterWorkflowRoutes(api, cfg, logger, db)
		// Register business routes (Phase 6)
		routes.RegisterBusinessRoutes(api, cfg, logger, db)
	}

	return r
}

func autoMigrate(logger *zap.Logger, db *gorm.DB) {
	logger.Info("running AutoMigrate to create/update tables")
	db.AutoMigrate(
		&model.User{},
		&model.Dept{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.AuditLog{},
		&model.TestItem{},
		&model.TestStandard{},
		&model.Equipment{},
		&model.Reagent{},
		&model.ProcessInstance{},
		&model.ProcessTask{},
		&model.TaskOrder{},
		&model.ContractReview{},
		&model.QCTask{},
		&model.SamplingSchedule{},
		&model.FieldSamplingRecord{},
		&model.SampleReceiving{},
		&model.TaskAssign{},
		&model.DataEntry{},
		&model.DataReview{},
		&model.DataAudit{},
		&model.ReportPrepare{},
		&model.ReportReview{},
		&model.ReportAudit{},
		&model.ReportSign{},
		&model.ReportPrint{},
		&model.ProjectArchive{},
	)
}
