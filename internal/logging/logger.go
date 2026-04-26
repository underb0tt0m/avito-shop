package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"avito-shop/internal/config"
)

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string, err error)
	Error(msg string, err error)
	Fatal(msg string, err error)
	Sync() error
}

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

func New() (Logger, func() error, error) {
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
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger{l}, logFile.Close, nil
}

/*
1. Сделал интерфейс - хорош, пропихни его внутрь всех пакетов, где он нужен
2. Рядышком написал реализацию на zap, создал в main экземпляр и пропихиваешь его под интерфейс

Фабрика, конфиг это все не надо
*/

type loggerNoop struct {
}

func NewLoggerNoop() Logger {
	return loggerNoop{}
}

func (l loggerNoop) Debug(msg string) {}

func (l loggerNoop) Info(msg string) {}

func (l loggerNoop) Warn(msg string, err error) {}

func (l loggerNoop) Error(msg string, err error) {}

func (l loggerNoop) Fatal(msg string, err error) {}

func (l loggerNoop) Sync() error {
	return nil
}
