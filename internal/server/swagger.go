package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/oulabla/go-base/internal/config"
	httpSwagger "github.com/swaggo/http-swagger"
)

// internal/server/swagger.go

func StartSwaggerServer(ctx context.Context) {
	addr := config.GetString(ctx, config.K.ServerSwaggerPort) // например ":8081"
	if addr == "" {
		log.Println("Swagger port not configured, skipping")
		return
	}

	mux := http.NewServeMux()

	// 1. Отдаём объединённый OpenAPI JSON
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		serveMergedSwaggerJSON(ctx, w)
	})

	// 2. Swagger UI на корневом пути (или /swagger, как вам удобнее)
	mux.Handle("/", httpSwagger.Handler(
		httpSwagger.URL("/openapi.json"),
		httpSwagger.DeepLinking(true),
		// httpSwagger.Prefix("/swagger"),   // если хотите путь /swagger/
	))

	log.Printf("Unified Swagger UI available at http://localhost%s/", addr)

	if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
		log.Printf("Swagger server failed: %v", err)
	}
}

func serveMergedSwaggerJSON(ctx context.Context, w http.ResponseWriter) {
	filePath := "gen/openapi/all-apis.swagger.json"

	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Cannot read merged OpenAPI file", http.StatusInternalServerError)
		log.Printf("Failed to read %s: %v", filePath, err)
		return
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		http.Error(w, "Invalid merged OpenAPI JSON", http.StatusInternalServerError)
		return
	}

	// Динамическая подмена host / schemes (очень полезно)
	doc["host"] = config.GetString(ctx, config.K.ServerSwaggerHost) // например "localhost:8080"
	if doc["host"] == "" {
		doc["host"] = "localhost:8080"
	}
	doc["schemes"] = []string{"http"} // или {"http","https"}
	// doc["basePath"] = "/"           // если нужно

	modified, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		http.Error(w, "Cannot encode OpenAPI", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(modified)
}
