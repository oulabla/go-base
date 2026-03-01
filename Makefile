# =============================================================================
# Makefile
# =============================================================================
-include .project.mk

# =============================================================================
# Основные переменные
# =============================================================================
PROJECT_NAME ?= go-base
MODULE       ?= github.com/oulabla/$(PROJECT_NAME)
BIN_DIR      := $(CURDIR)/bin
PROTO_DIR    := proto
GEN_DIR      := gen
CURRENT_DATE := $(shell date +%Y-%m-%d)

# =============================================================================
# Инструменты и версии
# =============================================================================
BUF_VERSION          := v1.45.0
YQ_VERSION           := v4.52.4
PROTOC_GEN_GO        := v1.36.0
PROTOC_GEN_GO_GRPC   := v1.5.1
GRPC_GATEWAY_VERSION := v2.23.0
OPENAPI_VERSION      := v2.20.0
MOCKERY_VERSION      := v3.6.4
GOLANGCI_LINT_VERSION := v2.10.1
GOLANGCI_LINT         := $(BIN_DIR)/golangci-lint

# =============================================================================
# Пути к бинарникам (все внешние инструменты лежат в ./bin)
# =============================================================================
BUF      := $(BIN_DIR)/buf
YQ       := $(BIN_DIR)/yq
MOCKERY  := $(BIN_DIR)/mockery

# =============================================================================
.DEFAULT_GOAL := help
.PHONY: help all build run generate generate-config proto-generate scaffold tools clean test lint tidy add-service

help:
	@echo ""
	@echo "Доступные команды:"
	@echo ""
	@echo "  make tools          Установить все инструменты в ./bin"
	@echo "  make generate       Генерация proto + stubs + config keys"
	@echo "  make generate-config Только config keys"
	@echo "  make build          Сборка → ./bin/$(PROJECT_NAME)"
	@echo "  make run            Запуск"
	@echo "  make test           Тесты + покрытие (без config/metric/server/script/gen)"
	@echo "  make lint           golangci-lint (если установлен)"
	@echo "  make tidy           go mod tidy"
	@echo "  make clean          Очистка"
	@echo ""
	@echo "  make add-service users      → proto/users/v1/..."
	@echo "  make add-service users v2   → proto/users/v2/..."
	@echo ""

# =============================================================================
# Tools — все бинарники ставятся в ./bin
# =============================================================================
tools: $(BIN_DIR) buf yq protoc-plugins mockery golangci-lint

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

buf: $(BIN_DIR)
	curl -sSL \
	  "https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/buf-Linux-x86_64" \
	  -o $(BUF)
	chmod +x $(BUF)

yq: $(BIN_DIR)
	curl -sSL \
	  "https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_linux_amd64" \
	  -o $(YQ)
	chmod +x $(YQ)
	@echo "yq $(YQ_VERSION) → $(YQ)"

protoc-plugins: $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO)
	GOBIN=$(BIN_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC)
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(GRPC_GATEWAY_VERSION)
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(OPENAPI_VERSION)

mockery: $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)


golangci-lint: $(BIN_DIR)
	@echo "→ Installing golangci-lint $(GOLANGCI_LINT_VERSION)"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
	  | sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)
	@$(GOLANGCI_LINT) --version

# =============================================================================
# Генерация
# =============================================================================
generate: proto-generate scaffold generate-config

generate-config:
	@echo "→ Generating config keys..."
	@bash script/generate_config.sh

proto-generate:
	PATH="$(CURDIR)/bin:$$PATH" $(BUF) generate $(PROTO_DIR)

scaffold:
	@PROJECT_NAME="$(PROJECT_NAME)" MODULE="$(MODULE)" \
		bash script/gen_stub.sh

# =============================================================================
# Тесты — с исключением папок из покрытия
# =============================================================================
EXCLUDE_REGEX := internal/config|internal/metric|internal/server|script|gen

# Пакеты, которые попадают в покрытие
COVER_PKGS := $(shell go list ./... | grep -vE '$(EXCLUDE_REGEX)' | tr '\n' ',' | sed 's/,$$//')

test: tidy
	@echo "→ Запуск тестов с покрытием (исключаем: config, metric, server, script, gen/...)"
	@if [ -z "$(COVER_PKGS)" ]; then \
		echo "→ Нет пакетов для покрытия после исключения — запускаем обычные тесты"; \
		go test -v -race ./...; \
		exit 0; \
	fi
	go test -v -race \
		-coverpkg=$(COVER_PKGS) \
		-coverprofile=coverage.out \
		-covermode=atomic \
		./...
	@echo ""
	@echo "→ Покрытие (только по выбранным пакетам):"
	@go tool cover -func=coverage.out | grep total || echo "→ Покрытие 0.0% (возможно, тестов пока нет)"
	@echo ""
	@echo "→ HTML-отчёт создан: coverage.html"
	@go tool cover -html=coverage.out -o coverage.html 2>/dev/null || true

# =============================================================================
# Линтинг (опционально — требует установки golangci-lint)
# =============================================================================
lint:
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "→ golangci-lint не найден. Запусти: make tools"; \
		exit 1; \
	fi
	PATH="$(BIN_DIR):$$PATH" $(GOLANGCI_LINT) run --timeout=5m --color=always

tidy:
	go mod tidy

# =============================================================================
# Сборка и запуск
# =============================================================================
build: tidy generate
	CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags "-s -w -X main.version=$(CURRENT_DATE)" \
		-o $(BIN_DIR)/$(PROJECT_NAME) \
		./cmd/$(PROJECT_NAME)/main.go
	@echo "→ Собрано: $(BIN_DIR)/$(PROJECT_NAME)"

run: build
	@$(BIN_DIR)/$(PROJECT_NAME)

# =============================================================================
# Добавление сервиса
# =============================================================================
.PHONY: add-service
add-service: ; @bash script/add_service.sh $(filter-out $@,$(MAKECMDGOALS))
%:
	@:

# =============================================================================
# Очистка
# =============================================================================
clean:
	rm -rf $(GEN_DIR) $(BIN_DIR) coverage.out coverage.html