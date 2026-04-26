package logger_factory

import (
	"avito-shop/internal/logging"
	"avito-shop/internal/logging/zap_logging"
	"fmt"
)

func New(realisation string, serverType string, logLvl string) (logging.Logger, func() error, error) {
	switch realisation {
	case "zap_logging":
		return zap_logging.New(serverType, logLvl)
	default:
		err := fmt.Errorf(
			"unknown logger type in the configuration: %v",
			realisation,
		)
		return nil, nil, err
	}
}
