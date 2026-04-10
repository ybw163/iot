package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"iot/pkg/config"
	"iot/pkg/logger"
)

var DB *gorm.DB

func Init(cfg config.DatabaseConfig) error {
	var logLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	DB = db
	logger.Log.Info("database connected", zap.String("host", cfg.Host), zap.Int("port", cfg.Port))

	return nil
}

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}
