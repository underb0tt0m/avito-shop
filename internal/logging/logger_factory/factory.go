package logger_factory

import (
	"avito-shop/internal/config"
	"avito-shop/internal/logging"
	"avito-shop/internal/logging/zap_logging"
	"fmt"
)

func New() (logging.Logger, func() error, error) {
	switch config.App.Logger.Type {
	case "zap_logging":
		return zap_logging.New()
	default:
		err := fmt.Errorf(
			"unknown logger type in the configuration: %v",
			config.App.Logger.Type,
		)
		return nil, nil, err
	}
}
