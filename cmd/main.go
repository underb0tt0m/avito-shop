package main

import (
	"avito-shop/cmd/handler"
	"avito-shop/internal/service"
	"avito-shop/internal/storage/postgres"
	"avito-shop/internal/tools"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewDevelopment())
	defer func() { _ = logger.Sync() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := tools.Create(ctx, logger)

	storageAPI := postgres.NewStorageAPI(conn, logger)
	serviceAPI := service.NewApi(storageAPI, logger)

	storageAuth := postgres.NewStorageAuth(conn, logger)
	serviceAuth := service.NewAuth(storageAuth, logger)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signalChan
		logger.Info("Got exit signal, exit context")
		cancel()
	}()

	router := chi.NewRouter()

	router.Route("/api", func(r chi.Router) {
		handler.Main(serviceAPI, r, logger)
		handler.Auth(serviceAuth, r, logger)
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
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
