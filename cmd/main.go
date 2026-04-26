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
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	cfg, err := config.Load("cmd/config.yaml")
	if err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	logger, closeLogger, err := logger_factory.New(cfg.Logger.Type, cfg.ServerType, cfg.Logger.Level)
	if err != nil {
		log.Fatalf("failed create logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
		if err = closeLogger(); err != nil {
			logger.Fatal("failed to close log file: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := CreatePool(ctx, cfg.Storage)
	if err != nil {
		logger.Fatal(
			"can't connect db",
			err,
		)
	}
	jsonCodec := tools.NewJSONCodec(cfg.Tools.JSON)
	tokenMaker := tools.NewToken(
		logger,
		jsonCodec,
		cfg.Security.JWTToken.Prefix,
		cfg.Security.JWTToken.Lifetime,
		cfg.Security.JWTToken.SecretKey,
	)
	hasher := tools.NewHasher(cfg.Security.Hash.Cost)

	storageAPI := postgres.NewStorageAPI(conn, logger)
	serviceAPI := service.NewApi(storageAPI, logger)

	storageAuth := postgres.NewStorageAuth(conn, logger)
	serviceAuth := service.NewAuth(storageAuth, logger, tokenMaker, hasher)

	MainHandler := handler.NewMain(serviceAPI, logger, jsonCodec, cfg.Storage.QueryTimeout, tokenMaker)
	AuthHandler := handler.NewAuth(serviceAuth, logger, jsonCodec, cfg.Storage.QueryTimeout)

	router := chi.NewRouter()
	router.Use(api_middleware.Stopwatch(logger))

	router.Route("/api", func(r chi.Router) {
		MainHandler.RegisterRoutes(r)
		AuthHandler.RegisterRoutes(r)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: router,
	}

	wg := sync.WaitGroup{}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-signalChan
		logger.Info("Got exit signal, exit context")
		cancel()

		shutDownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		) //TODO добавить в конфиг
		defer shutdownCancel()

		if shutdownErr := server.Shutdown(shutDownCtx); err != nil {
			logger.Error(
				"failed to shutdown server gracefully",
				shutdownErr,
			)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Start HTTP-server")
		if listenErr := server.ListenAndServe(); err != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			logger.Fatal("http server failed", listenErr)
		} else {
			logger.Info("Stop HTTP-server")
		}
	}()

	wg.Wait()
}

func CreatePool(ctx context.Context, connParam config.Storage) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"%v://%v:%v@%v:%v/%v",
		connParam.Connection.Driver,
		connParam.Connection.User,
		connParam.Connection.Password,
		connParam.Connection.Host,
		connParam.Connection.Port,
		connParam.Connection.Database,
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
