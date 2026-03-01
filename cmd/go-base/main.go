package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/oulabla/go-base/internal/config"
	"github.com/oulabla/go-base/internal/config/secret"
	"github.com/oulabla/go-base/internal/metric"
	"github.com/oulabla/go-base/internal/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type appNameKeyType string

const appNameKey appNameKeyType = "application_name"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ────────────────────────────────────────────────
	// Флаги
	// ────────────────────────────────────────────────
	var (
		useLocalConfig bool
		debug          bool
	)

	flag.BoolVar(&useLocalConfig, "local", false, "use config/local.yaml instead of config/prod.yaml")
	flag.BoolVar(&debug, "debug", false, "enable debug logging and colored console output")
	flag.Parse()

	// ────────────────────────────────────────────────
	// Инициализация логгера (JSON в prod, цветной в debug)
	// ────────────────────────────────────────────────
	server.Init(debug)

	// Устанавливаем уровень лога глобально
	if debug {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}

	// Логируем запуск приложения
	log.Info().
		Bool("debug", debug).
		Str("config", func() string {
			if useLocalConfig {
				return "local.yaml"
			}
			return "prod.yaml"
		}()).
		Msg("starting application")

	// ────────────────────────────────────────────────
	// Конфигурация
	// ────────────────────────────────────────────────
	configFile := "config/prod.yaml"
	if useLocalConfig {
		configFile = "config/local.yaml"
	}

	configProvider, err := config.NewYAMLProvider(configFile)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("config_file", configFile).
			Msg("failed to load config")
	}

	config.SetProvider(configProvider)

	appName := config.GetString(ctx, config.K.ApplicationName)
	if appName == "" {
		log.Fatal().Msg("application name is not set")
	}
	ctx = context.WithValue(ctx, appNameKey, appName)

	// ────────────────────────────────────────────────
	// Секреты
	// ────────────────────────────────────────────────
	secretProvider, err := secret.NewYAMLSecretProvider("config/secret.yaml")
	if err != nil {
		log.Fatal().
			Err(err).
			Str("config_file", configFile).
			Msg("failed to load secrets")
	}
	secret.SetProvider(secretProvider)

	// ────────────────────────────────────────────────
	// Prometheus metrics endpoint
	// ────────────────────────────────────────────────
	http.Handle("/metrics", promhttp.Handler())

	metricAddr := config.GetString(ctx, config.K.ServerMetricPort)

	go func() {
		log.Info().
			Str("addr", metricAddr).
			Msg("starting HTTP metrics server")

		if err := http.ListenAndServe(metricAddr, nil); err != nil {
			log.Error().
				Err(err).
				Str("addr", metricAddr).
				Msg("HTTP metrics server failed")
		}
	}()

	// ────────────────────────────────────────────────
	// gRPC сервер
	// ────────────────────────────────────────────────
	grpcAddr := config.GetString(ctx, config.K.ServerGrpcPort)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("addr", grpcAddr).
			Msg("failed to listen on gRPC address")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(metric.UnaryServerInterceptor()),
	)

	server.RegisterAllGRPC(grpcServer)

	go func() {
		log.Info().
			Str("addr", grpcAddr).
			Msg("starting gRPC server")

		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Error().
				Err(err).
				Str("addr", grpcAddr).
				Msg("gRPC server failed")
		}
	}()

	// ────────────────────────────────────────────────
	// HTTP/JSON gateway (REST API)
	// ────────────────────────────────────────────────
	gwmux := runtime.NewServeMux()

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := server.RegisterAllGateway(ctx, gwmux, dialOpts); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to register gateway endpoints")
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	}).Handler(gwmux)

	httpAddr := config.GetString(ctx, config.K.ServerHttpPort)

	go func() {
		log.Info().
			Str("addr", httpAddr).
			Msg("starting HTTP/JSON gateway")

		if err := http.ListenAndServe(httpAddr, corsHandler); err != nil {
			log.Error().
				Err(err).
				Str("addr", httpAddr).
				Msg("HTTP/JSON gateway failed")
		}
	}()

	// ────────────────────────────────────────────────
	// Swagger UI (если реализован)
	// ────────────────────────────────────────────────
	go server.StartSwaggerServer(ctx)

	// ────────────────────────────────────────────────
	// Graceful shutdown
	// ────────────────────────────────────────────────
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan

	log.Info().Msg("received shutdown signal, gracefully stopping servers...")

	grpcServer.GracefulStop()
	cancel()

	log.Info().Msg("shutdown complete")
}
