package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"lims-backend/internal/config"
	"lims-backend/internal/router"
	"lims-backend/internal/seed"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("LIMS migrate starting",
		zap.String("env", cfg.Env),
		zap.String("database", cfg.Database.DBName),
	)

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
	)

	var db *gorm.DB
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		})
		if err == nil {
			break
		}
		logger.Warn("postgres not ready, retrying",
			zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Fatal("Failed to connect database after retries", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get sql.DB", zap.Error(err))
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	logger.Info("Running schema migration ...")
	router.RunAutoMigrate(logger, db)

	logger.Info("Running seed data ...")
	seed.Run(db, logger)

	logger.Info("Migration completed successfully")
	os.Exit(0)
}
