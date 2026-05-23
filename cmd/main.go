package main

import (
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
	"github.com/prometheus/client_golang/prometheus"

	"avito-shop/cmd/handler"
	"avito-shop/internal/api_middleware"
	"avito-shop/internal/config"
	"avito-shop/internal/hasher"
	"avito-shop/internal/jsoncodec"
	"avito-shop/internal/jwtmanager"
	"avito-shop/internal/logging"
	"avito-shop/internal/prometheus_metrics"
	"avito-shop/internal/service"
	"avito-shop/internal/storage/postgres"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	cfg, err := config.Load("./cmd/config.yaml")
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
			logger.Errorf("failed to sync logger: %v", err)
		}
		if err = closeLogger(); err != nil {
			logger.Fatalf("failed to close log file: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := CreatePool(ctx, cfg.Storage)
	if err != nil {
		logger.Fatalf(
			"can't connect db: %v",
			err,
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

	reg := prometheus.DefaultRegisterer
	m := prometheus_metrics.New(reg)

	storageAPI := postgres.NewStorageAPI(conn, logger)
	serviceAPI := service.NewApi(storageAPI, logger)

	storageAuth := postgres.NewStorageAuth(conn, logger)
	serviceAuth := service.NewAuth(storageAuth, logger, tokenMaker, Hasher, m)

	MainHandler := handler.NewMain(serviceAPI, logger, jsonCodec, cfg.Storage.QueryTimeout, tokenMaker, m)
	AuthHandler := handler.NewAuth(serviceAuth, logger, jsonCodec, cfg.Storage.QueryTimeout, m)
	MetricsHandler := handler.NewMetrics(m, reg)

	serviceRouter := chi.NewRouter()
	serviceRouter.Use(api_middleware.Stopwatch(logger))
	serviceRouter.Use(prometheus_metrics.Middleware(m, logger))

	techRouter := chi.NewRouter()

	serviceRouter.Route("/api", func(r chi.Router) {
		MainHandler.RegisterRoutes(r)
		AuthHandler.RegisterRoutes(r)
	})
	techRouter.Route("/", func(r chi.Router) {
		MetricsHandler.RegisterRoutes(r)
	})

	serviceServer := http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.ServicePort),
		Handler: serviceRouter,
	}
	techServer := http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.TechPort),
		Handler: techRouter,
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

		if shutdownErr := serviceServer.Shutdown(shutDownCtx); shutdownErr != nil {
			logger.Errorf(
				"failed to shutdown service server gracefully: %v",
				shutdownErr,
			)
		}
		if shutdownErr := techServer.Shutdown(shutDownCtx); shutdownErr != nil {
			logger.Errorf(
				"failed to shutdown technical server gracefully: %v",
				shutdownErr,
			)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof("Start service HTTP-server")
		if listenErr := serviceServer.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			logger.Errorf(
				"service http server failed: %v",
				listenErr,
			)
		} else {
			logger.Infof("Stop service HTTP-server")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof("Start technical HTTP-server")
		if listenErr := techServer.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			logger.Errorf(
				"technical http server failed: %v",
				listenErr,
			)
		} else {
			logger.Infof("Stop technical HTTP-server")
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

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 300 // TODO добавить в конфиг
	cfg.MinConns = 75
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetimeJitter = 5 * time.Second
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
