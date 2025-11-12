package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/config"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/consumer"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/repository"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/routes"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/internal/services"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/pkg/logger"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/pkg/metrics"
	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/pkg/retry"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logr := logger.New(cfg.LogLevel)
	logr.Info("starting push service", slog.String("app", cfg.AppName))

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		logr.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}

	var redisRepo *repository.RedisRepository
	if cfg.RedisURL != "" {
		rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
		redisRepo = repository.NewRedisRepository(rdb, 24*time.Hour)
		defer rdb.Close()
	}

	statusStore := repository.NewStatusStore(db, cfg.StatusTable)
	statusUpdater := services.NewStatusUpdater(statusStore, logr)

	templateClient := services.NewTemplateClient(cfg.TemplateServiceURL, cfg.ProviderTimeout)
	fcmProvider := services.NewFCMProvider(cfg.FCMServerKey, cfg.FCMEndpoint, cfg.ProviderTimeout, logr)
	metricsCollector := metrics.New()

	retryCfg := retry.Config{
		MaxAttempts:    cfg.RetryMaxAttempts,
		InitialBackoff: cfg.RetryInitialBackoff,
		MaxBackoff:     cfg.RetryMaxBackoff,
	}

	processor := services.NewPushProcessor(
		templateClient,
		fcmProvider,
		statusUpdater,
		redisRepo,
		metricsCollector,
		logr,
		retryCfg,
	)

	conn, err := amqp.Dial(cfg.RabbitURL)
	if err != nil {
		logr.Error("failed to connect rabbitmq", slog.Any("error", err))
		os.Exit(1)
	}
	defer conn.Close()

	base := consumer.NewBaseConsumer(
		conn,
		cfg.PushQueue,
		cfg.DeadLetterQueue,
		cfg.PrefetchCount,
		cfg.WorkerCount,
		logr,
	)
	pushConsumer := consumer.NewPushConsumer(base, processor, logr, cfg.RetryMaxAttempts)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	started := time.Now()
	httpSrv := startHTTPServer(cfg.HTTPPort, metricsCollector, logr, started)

	if err := pushConsumer.Start(ctx); err != nil {
		logr.Error("push consumer exited", slog.Any("error", err))
	}

	shutdownHTTP(httpSrv, logr)
	logr.Info("push service stopped")
}

func startHTTPServer(port string, metricsCollector *metrics.Metrics, logr *slog.Logger, started time.Time) *http.Server {
	if port == "" {
		port = "8082"
	}
	handler := routes.NewRouter(metricsCollector, started)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Error("http server error", slog.Any("error", err))
		}
	}()
	return srv
}

func shutdownHTTP(srv *http.Server, logr *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logr.Error("failed to shutdown http server", slog.Any("error", err))
	}
}
