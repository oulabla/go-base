package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/oulabla/go-base/gen/go/user/v1"
	"github.com/oulabla/go-base/internal/endpoints/user"
)

const (
	grpcPort    = ":50051"
	httpPort    = ":8080"
	swaggerPort = ":8081"
	apiHost     = "localhost:8080"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ────────────────────────────────────────────────
	// Controller
	// ────────────────────────────────────────────────
	userController := user.NewController()

	// ────────────────────────────────────────────────
	// gRPC server
	// ────────────────────────────────────────────────
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userController)

	go func() {
		log.Printf("gRPC server listening on %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	}()

	// ────────────────────────────────────────────────
	// HTTP Gateway (JSON API)
	// ────────────────────────────────────────────────
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, gwmux, grpcPort, opts); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	}).Handler(gwmux)

	go func() {
		log.Printf("HTTP/JSON gateway listening on %s", httpPort)
		if err := http.ListenAndServe(httpPort, corsHandler); err != nil {
			log.Fatalf("HTTP gateway failed: %v", err)
		}
	}()

	// ────────────────────────────────────────────────
	// Swagger UI (порт 8081)
	// ────────────────────────────────────────────────
	go func() {
		mux := http.NewServeMux()

		// Отдаём swagger.json с подменой host
		mux.HandleFunc("/swagger-files/", func(w http.ResponseWriter, r *http.Request) {
			targetFile := filepath.Base(r.URL.Path)
			var foundPath string

			_ = filepath.Walk("./gen/openapi", func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && info.Name() == targetFile {
					foundPath = path
				}
				return nil
			})

			if foundPath == "" {
				http.Error(w, "Swagger file not found", http.StatusNotFound)
				return
			}

			data, err := os.ReadFile(foundPath)
			if err != nil {
				http.Error(w, "Failed to read swagger file", http.StatusInternalServerError)
				return
			}

			var swagger map[string]interface{}
			if err := json.Unmarshal(data, &swagger); err != nil {
				http.Error(w, "Invalid swagger file", http.StatusInternalServerError)
				return
			}

			// Критично: подменяем host и schemes
			swagger["host"] = apiHost
			swagger["schemes"] = []string{"http"}

			modified, err := json.Marshal(swagger)
			if err != nil {
				http.Error(w, "Failed to encode swagger", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(modified)
		})

		// Swagger UI
		mux.Handle("/swagger/", httpSwagger.Handler(
			httpSwagger.URL("/swagger-files/user.swagger.json"),
		))

		log.Printf("Swagger UI available at http://localhost%s/swagger/index.html", swaggerPort)

		if err := http.ListenAndServe(swaggerPort, mux); err != nil {
			log.Fatalf("Swagger server failed: %v", err)
		}
	}()

	// ────────────────────────────────────────────────
	// Graceful shutdown
	// ────────────────────────────────────────────────
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	grpcServer.GracefulStop()
	cancel()
}
