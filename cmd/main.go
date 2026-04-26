package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"avito-shop/internal/logging/logger_factory"
	"github.com/jackc/pgx/v5/pgxpool"

	"avito-shop/cmd/handler"
	"avito-shop/internal/api_middleware"
	"avito-shop/internal/config"
	"avito-shop/internal/service"
	"avito-shop/internal/storage/postgres"
	"avito-shop/internal/tools"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	if err := config.Init("cmd/config.yaml"); err != nil {
		panic(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

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

	conn, err := tools.CreatePool(ctx)
	if err != nil {
		logger.Fatal(
			"can't connect db",
			err,
		)
	}
	jsonCodec := tools.NewJSONCodec()
	tokenMaker := tools.NewToken(logger, jsonCodec)
	hasher := tools.NewHasher()

	storageAPI := postgres.NewStorageAPI(conn, logger)
	serviceAPI := service.NewApi(storageAPI, logger)

	storageAuth := postgres.NewStorageAuth(conn, logger)
	serviceAuth := service.NewAuth(storageAuth, logger, tokenMaker, hasher)

	router := chi.NewRouter()

	router.Route("/api", func(r chi.Router) {
		r.Use(api_middleware.Stopwatch(logger))
		r.Group(func(r chi.Router) {
			r.Use(api_middleware.Auth(logger, tokenMaker))
			handler.Main(serviceAPI, r, logger, jsonCodec)
		})
		handler.Auth(serviceAuth, r, logger, jsonCodec)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", config.App.Port),
		Handler: creaRoutes(),
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

func CreatePool(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"%v://%v:%v@%v:%v/%v",
		config.App.Storage.Connection.Driver,
		config.App.Storage.Connection.User,
		config.App.Storage.Connection.Password,
		config.App.Storage.Connection.Host,
		config.App.Storage.Connection.Port,
		config.App.Storage.Connection.Database,
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func creaRoutes() http.Handler {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Get("/info", h.Info)
		r.Delete()
	})

	router.Route("/api", func(r chi.Router) {
		r.Use(api_middleware.Stopwatch(logger))
		r.Group(func(r chi.Router) {
			r.Use(api_middleware.Auth(logger, tokenMaker))
			handler.Main(serviceAPI, r, logger, jsonCodec)
		})
		handler.Auth(serviceAuth, r, logger, jsonCodec)
	})

	return router
}
