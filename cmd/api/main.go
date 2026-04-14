package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"avito-shop/internal/features/api/mainRoutService"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"avito-shop/cmd/api/handler"
	"avito-shop/internal/config"
	"avito-shop/internal/db"
	"avito-shop/internal/features/api/mainRoutRepository"
	"avito-shop/internal/features/auth/authRepository"
	"avito-shop/internal/features/auth/authService"
	"avito-shop/internal/features/auth/authTransport"
)

func main() {
	logger := zap.Must(zap.NewDevelopment())
	defer func() { _ = logger.Sync() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cfg config.API
	parseConfig(&cfg, "filename")

	conn := db.Create(ctx, logger)

	var mainRepo mainRoutRepository.Storage = mainRoutRepository.StorageImpl{Conn: conn, Logger: logger}
	var mainServ mainRoutService.Service = mainRoutService.ServiceImpl{Repo: mainRepo, Logger: logger}

	var authRepo authRepository.Storage = authRepository.StorageImpl{Conn: conn, Logger: logger}
	var authServ authService.Service = authService.ServiceImpl{Repo: authRepo, Logger: logger}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signalChan
		logger.Info("Got exit signal, exit context")
		cancel()
	}()

	router := chi.NewRouter()

	router.Route("/api", func(r chi.Router) {
		handler.Register(mainServ, r, logger)
		authTransport.Register(authServ, r, logger)
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
