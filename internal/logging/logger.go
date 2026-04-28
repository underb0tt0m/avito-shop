package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(msg string, v ...any)
	Infof(msg string, v ...any)
	Warnf(err error, msg string, v ...any)
	Errorf(err error, msg string, v ...any)
	Fatalf(err error, msg string, v ...any)
	Sync() error
}

func New(realisation string, serverType string, logLvl string) (Logger, func() error, error) {
	switch realisation {
	case "zap_logging":
		return newZap(serverType, logLvl)
	default:
		err := fmt.Errorf(
			"unknown logger type in the configuration: %v",
			realisation,
		)
		return nil, nil, err
	}
}

type loggerZap struct {
	Logger *zap.Logger
}

func (l loggerZap) Debugf(msg string, v ...any) {
	fmtMsg := fmt.Sprintf(msg, v...)
	l.Logger.Debug(fmtMsg)
}

func (l loggerZap) Infof(msg string, v ...any) {
	fmtMsg := fmt.Sprintf(msg, v...)
	l.Logger.Info(fmtMsg)
}

func (l loggerZap) Warnf(err error, msg string, v ...any) {
	fmtMsg := fmt.Sprintf(msg, v...)
	l.Logger.Warn(
		fmtMsg,
		zap.Error(err),
	)
}

func (l loggerZap) Errorf(err error, msg string, v ...any) {
	fmtMsg := fmt.Sprintf(msg, v...)
	l.Logger.Error(
		fmtMsg,
		zap.Error(err),
	)
}

func (l loggerZap) Fatalf(err error, msg string, v ...any) {
	fmtMsg := fmt.Sprintf(msg, v...)
	l.Logger.Fatal(
		fmtMsg,
		zap.Error(err),
	)
}

func (l loggerZap) Sync() error {
	if err := l.Logger.Sync(); err != nil {
		return err
	}
	return nil
}

func newZap(serverType string, logLvl string) (Logger, func() error, error) {
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
	switch serverType {
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
	switch logLvl {
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

	return loggerZap{Logger: l}, logFile.Close, nil
}

type LoggerNoop struct{}

func (l LoggerNoop) Debugf(msg string, v ...any) {}

func (l LoggerNoop) Infof(msg string, v ...any) {}

func (l LoggerNoop) Warnf(err error, msg string, v ...any) {}

func (l LoggerNoop) Errorf(err error, msg string, v ...any) {}

func (l LoggerNoop) Fatalf(err error, msg string, v ...any) {}

func (l LoggerNoop) Sync() error {
	return nil
}
