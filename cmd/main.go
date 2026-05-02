package main

import (
	"avito-shop/cmd/handler"
	"avito-shop/internal/api_middleware"
	"avito-shop/internal/config"
	"avito-shop/internal/hasher"
	"avito-shop/internal/jsoncodec"
	"avito-shop/internal/jwtmanager"
	"avito-shop/internal/logging"

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

	"avito-shop/internal/service"
	"avito-shop/internal/storage/postgres"

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

	logger, closeLogger, err := logging.New(
		cfg.Logger.Type,
		cfg.ServerType,
		cfg.Logger.Level,
	)
	if err != nil {
		log.Fatalf("failed create logger: %v", err)
	}
	defer func() {
		if err = logger.Sync(); err != nil {
			logger.Errorf(err, "failed to sync logger: %v", err)
		}
		if err = closeLogger(); err != nil {
			logger.Fatalf(err, "failed to close log file: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := CreatePool(ctx, cfg.Storage)
	if err != nil {
		logger.Fatalf(
			err,
			"can't connect db",
		)
	}
	jsonCodec := jsoncodec.NewJSONCodec(cfg.Tools.JSON)
	tokenMaker := jwtmanager.NewToken(
		logger,
		jsonCodec,
		cfg.Security.JWTToken.Prefix,
		cfg.Security.JWTToken.Lifetime,
		cfg.Security.JWTToken.SecretKey,
	)
	Hasher := hasher.NewHasher(cfg.Security.Hash.Cost)

	storageAPI := postgres.NewStorageAPI(conn, logger)
	serviceAPI := service.NewApi(storageAPI, logger)

	storageAuth := postgres.NewStorageAuth(conn, logger)
	serviceAuth := service.NewAuth(storageAuth, logger, tokenMaker, Hasher)

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
		logger.Infof("Got exit signal, exit context")
		cancel()

		shutDownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		) //TODO добавить в конфиг
		defer shutdownCancel()

		if shutdownErr := server.Shutdown(shutDownCtx); err != nil {
			logger.Errorf(
				shutdownErr,
				"failed to shutdown server gracefully",
			)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof("Start HTTP-server")
		if listenErr := server.ListenAndServe(); err != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			logger.Fatalf(listenErr, "http server failed")
		} else {
			logger.Infof("Stop HTTP-server")
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
