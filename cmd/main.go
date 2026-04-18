package main

import (
	"avito-shop/cmd/handler"
	"avito-shop/internal/api_middleware"
	"avito-shop/internal/config"
	"avito-shop/internal/logging/logger_factory"
	"avito-shop/internal/service"
	"avito-shop/internal/storage/postgres"
	"avito-shop/internal/tools"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	if err := config.Init("cmd/config.yaml"); err != nil {
		panic(err)
	}

	logger, closeLogger, err := logger_factory.New()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
		if err = closeLogger(); err != nil {
			panic(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := tools.Create(ctx)
	if err != nil {
		logger.Fatal(
			"can't connect db",
			err,
		)
	}

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
		r.Use(api_middleware.Stopwatch(logger))
		r.Group(func(r chi.Router) {
			r.Use(api_middleware.Auth(logger))
			handler.Main(serviceAPI, r, logger)
		})
		handler.Auth(serviceAuth, r, logger)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", config.App.Port),
		Handler: router,
	}

	go func() {
		logger.Info("Start HTTP-server")
		if err = server.ListenAndServe(); err != nil {
			logger.Info("Stop HTTP-server")
		}
	}()

	<-ctx.Done()
	if err = server.Shutdown(ctx); err != nil {
		logger.Error(
			"failed to shutdown server gracefully",
			err,
		)

	}

}
