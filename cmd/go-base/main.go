// cmd/go-base/main.go
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// ваши контроллеры
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/oulabla/go-base/internal/endpoints/user"

	// сгенерированные пакеты
	pb "github.com/oulabla/go-base/gen/go/user/v1"
)

const (
	grpcPort = ":50051" // gRPC + HTTP/2
	httpPort = ":8080"  // чистый HTTP/1.1 JSON (опционально отдельный порт)
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// ────────────────────────────────────────────────
	// Инициализация зависимостей и контроллеров
	// ────────────────────────────────────────────────
	userController := user.NewController( /* deps... */ )

	// ────────────────────────────────────────────────
	// gRPC сервер
	// ────────────────────────────────────────────────
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	userController.Register(grpcServer) // pb.RegisterUserServiceServer(grpcServer, userController)

	// ────────────────────────────────────────────────
	// HTTP/JSON Gateway (grpc-gateway)
	// ────────────────────────────────────────────────
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Регистрируем gateway handler
	err = pb.RegisterUserServiceHandlerFromEndpoint(ctx, gwmux, grpcPort, opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	// ────────────────────────────────────────────────
	// Запускаем всё
	// ────────────────────────────────────────────────
	go func() {
		log.Printf("gRPC + gRPC-Web + Reflection listening on %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	}()

	// Отдельный HTTP-сервер только для JSON (если хотите другой порт)
	go func() {
		log.Printf("HTTP/JSON gateway listening on %s", httpPort)
		if err := http.ListenAndServe(httpPort, gwmux); err != nil {
			log.Fatalf("HTTP gateway failed: %v", err)
		}
	}()

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	grpcServer.GracefulStop()
	cancel()
	log.Println("Shutdown complete")
}
