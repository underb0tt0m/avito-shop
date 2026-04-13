package main

import (
	"avito-shop/internal/core/tools"
	"avito-shop/internal/features/api/repository"
	"avito-shop/internal/features/api/service"
	"avito-shop/internal/features/api/transport"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewDevelopment())
	defer func() { _ = logger.Sync() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := tools.Create(ctx, logger)
	var repo repository.Storage = repository.StorageImpl{Conn: conn, Logger: logger}
	var serv service.Service = service.ServiceImpl{Repo: repo, Logger: logger}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signalChan
		logger.Info("Got exit signal, exit context")
		cancel()
	}()

	rout := transport.Register(serv, logger)
	server := http.Server{
		Addr:    ":8080",
		Handler: rout,
	}

	go func() {
		logger.Info("Start HTTP-server")
		if err := server.ListenAndServe(); err != nil {
			logger.Info("Stop HTTP-server")
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error(
			"failed to shutdown server gracefully",
			zap.Error(err),
		)

	}

}
