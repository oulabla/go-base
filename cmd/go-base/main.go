package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/oulabla/go-base/internal/config"
	_ "github.com/oulabla/go-base/internal/endpoints"
	"github.com/oulabla/go-base/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ────────────────────────────────────────────────
	// Парсинг флага --local
	// ────────────────────────────────────────────────
	var useLocalConfig bool
	flag.BoolVar(&useLocalConfig, "local", false, "use local.yaml instead of prod.yaml")
	flag.Parse() // важно вызвать после всех определений флагов

	// ────────────────────────────────────────────────
	// Config
	// ────────────────────────────────────────────────
	configFile := "config/prod.yaml"
	if useLocalConfig {
		configFile = "config/local.yaml"
	}

	configProvider, err := config.NewYAMLProvider(configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	config.SetProvider(configProvider)

	// ────────────────────────────────────────────────
	// gRPC сервер (один на все сервисы)
	// ────────────────────────────────────────────────
	grpcAddr := config.GetString(ctx, config.K.ServerGrpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen gRPC on %s: %v", grpcAddr, err)
	}

	grpcServer := grpc.NewServer(
	// grpc.ChainUnaryInterceptor(...),
	// grpc.ChainStreamInterceptor(...),
	)

	server.RegisterAllGRPC(grpcServer)

	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	}()

	// ────────────────────────────────────────────────
	// HTTP/JSON gateway (один mux на все сервисы)
	// ────────────────────────────────────────────────
	gwmux := runtime.NewServeMux(
	// runtime.WithErrorHandler(...),
	// runtime.WithForwardResponseOption(...),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := server.RegisterAllGateway(ctx, gwmux, dialOpts); err != nil {
		log.Fatalf("failed to register gateway endpoints: %v", err)
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	}).Handler(gwmux)

	httpAddr := config.GetString(ctx, config.K.ServerHttpPort)
	go func() {
		log.Printf("HTTP/JSON gateway listening on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, corsHandler); err != nil {
			log.Fatalf("HTTP gateway failed: %v", err)
		}
	}()

	// ────────────────────────────────────────────────
	// Swagger UI — по одному серверу на каждый сервис
	// ────────────────────────────────────────────────
	go server.StartSwaggerServer(ctx)

	// ────────────────────────────────────────────────
	// Graceful shutdown
	// ────────────────────────────────────────────────
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan

	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()
	cancel()

	// Даём время на завершение http-серверов swagger (опционально)
	// time.Sleep(2 * time.Second)
}
