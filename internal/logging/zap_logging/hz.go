package zap_logging

import (
	"avito-shop/internal/config"
	"avito-shop/internal/logging"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	Logger *zap.Logger
}

func (l logger) Debug(msg string) {
	l.Logger.Debug(msg)
}

func (l logger) Info(msg string) {
	l.Logger.Info(msg)
}

func (l logger) Warn(msg string, err error) {
	l.Logger.Warn(
		msg,
		zap.Error(err),
	)
}

func (l logger) Error(msg string, err error) {
	l.Logger.Error(
		msg,
		zap.Error(err),
	)
}

func (l logger) Fatal(msg string, err error) {
	l.Logger.Fatal(
		msg,
		zap.Error(err),
	)
}

func (l logger) Sync() error {
	if err := l.Logger.Sync(); err != nil {
		return err
	}
	return nil
}

func New() (logging.Logger, func() error, error) {
	if err := os.MkdirAll("logs", 755); err != nil {
		return nil, nil, err
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000000")
	logFilePath := filepath.Join("logs", fmt.Sprintf("%s.log", timestamp))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	var encoderCfg zapcore.EncoderConfig
	switch config.App.ServerType {
	case "development":
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	case "production":
		encoderCfg = zap.NewProductionEncoderConfig()
	default:
		err = fmt.Errorf("there is no server type in the configuration")
		return nil, nil, err
	}
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000")
	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	var level zapcore.Level
	switch config.App.Logger.Level {
	case "debug":
		level = -1
	case "info":
		level = 0
	case "warn":
		level = 1
	default:
		err = fmt.Errorf("there is no logging level in the configuration")
		return nil, nil, err
	}

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(encoder, zapcore.AddSync(logFile), level),
	)

	l := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	l.Info("Log file created")

	return logger{l}, logFile.Close, nil
}
