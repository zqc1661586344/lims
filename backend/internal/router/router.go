package router

import (
	"lims-backend/internal/config"
	"lims-backend/internal/middleware"
	"lims-backend/internal/router/routes"
	"strings"

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
		RunAutoMigrate(logger, db)
	} else {
		logger.Warn("skipping AutoMigrate in production environment")
	}

	// Register audit plugin (GORM hooks for data change tracking)
	auditPlugin := middleware.NewAuditPlugin(logger)
	if err := db.Use(auditPlugin); err != nil {
		logger.Fatal("Failed to register audit plugin", zap.Error(err))
	}

	r := gin.New()
	r.SetTrustedProxies(cfg.Server.TrustedProxies)

	// Global middleware
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.LoggerMiddleware(logger))

	// CORS: 使用 AllowOriginFunc 精确控制 Origin 校验，绕过库内部 normalize 的大小写问题
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if cfg.Env == "development" {
				return strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:") ||
					strings.HasPrefix(origin, "http://0.0.0.0:")
			}
			for _, allowed := range cfg.CORS.AllowOrigins {
				if allowed == origin {
					return true
				}
			}
			return false
		},
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
		// Register lab sheet routes (Univer Sheet 检验单)
		routes.RegisterLabSheetRoutes(api, cfg, logger, db)
		// Register sampling sheet routes (Univer Sheet 采样单)
		routes.RegisterSamplingSheetRoutes(api, cfg, logger, db)
	}

	return r
}

func preMigrateCleanup(db *gorm.DB, logger *zap.Logger, delete bool) {
	type orphanRel struct {
		name    string
		countQ  string
		deleteQ string
	}
	rels := []orphanRel{
		{"process_tasks → process_instances",
			`SELECT COUNT(*) FROM process_tasks WHERE process_instance_id NOT IN (SELECT id FROM process_instances)`,
			`DELETE FROM process_tasks WHERE process_instance_id NOT IN (SELECT id FROM process_instances)`},
		{"contract_reviews → task_orders",
			`SELECT COUNT(*) FROM contract_reviews WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM contract_reviews WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"qc_tasks → task_orders",
			`SELECT COUNT(*) FROM qc_tasks WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM qc_tasks WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"sampling_schedules → task_orders",
			`SELECT COUNT(*) FROM sampling_schedules WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM sampling_schedules WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"field_sampling_records → task_orders",
			`SELECT COUNT(*) FROM field_sampling_records WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM field_sampling_records WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"sample_receivings → task_orders",
			`SELECT COUNT(*) FROM sample_receivings WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM sample_receivings WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"task_assigns → task_orders",
			`SELECT COUNT(*) FROM task_assigns WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM task_assigns WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"data_entries → task_orders",
			`SELECT COUNT(*) FROM data_entries WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM data_entries WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"data_reviews → task_orders",
			`SELECT COUNT(*) FROM data_reviews WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM data_reviews WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"data_audits → task_orders",
			`SELECT COUNT(*) FROM data_audits WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM data_audits WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"report_prepares → task_orders",
			`SELECT COUNT(*) FROM report_prepares WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM report_prepares WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"report_reviews → task_orders",
			`SELECT COUNT(*) FROM report_reviews WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM report_reviews WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"report_audits → task_orders",
			`SELECT COUNT(*) FROM report_audits WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM report_audits WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"report_signs → task_orders",
			`SELECT COUNT(*) FROM report_signs WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM report_signs WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"report_prints → task_orders",
			`SELECT COUNT(*) FROM report_prints WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM report_prints WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
		{"project_archives → task_orders",
			`SELECT COUNT(*) FROM project_archives WHERE task_order_id NOT IN (SELECT id FROM task_orders)`,
			`DELETE FROM project_archives WHERE task_order_id NOT IN (SELECT id FROM task_orders)`},
	}

	anyOrphan := false
	for _, r := range rels {
		var cnt int64
		if err := db.Raw(r.countQ).Scan(&cnt).Error; err != nil {
			logger.Debug("orphan count query skipped (table may not exist)",
				zap.String("relation", r.name), zap.Error(err))
			continue
		}
		if cnt == 0 {
			continue
		}
		anyOrphan = true
		if delete {
			result := db.Exec(r.deleteQ)
			logger.Warn("Cleaned orphan rows before FK constraint",
				zap.String("relation", r.name),
				zap.Int64("rows_removed", result.RowsAffected))
		} else {
			logger.Error("Orphan rows detected but NOT deleted (LIMS_STRICT_CLEANUP=false)",
				zap.String("relation", r.name),
				zap.Int64("orphan_count", cnt),
				zap.String("hint", "run LIMS_STRICT_CLEANUP=true once to clean, or repair data manually"))
		}
	}

	if anyOrphan && !delete {
		logger.Warn("AutoMigrate will add FK constraints that may FAIL if orphan rows still exist — resolve manually")
	}
}
