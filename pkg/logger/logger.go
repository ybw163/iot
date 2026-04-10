package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"iot/pkg/config"
)

var Log *zap.Logger

func Init(cfg config.LogConfig) error {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("parse log level failed: %w", err)
	}

	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	// stdout
	if cfg.Output == "stdout" || cfg.Output == "" {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	// file
	if cfg.FilePath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return fmt.Errorf("create log dir failed: %w", err)
		}
		writer := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(writer), level))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	core := zapcore.NewTee(cores...)
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return nil
}
