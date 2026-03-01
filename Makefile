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

# =============================================================================
# Пути к бинарникам
# =============================================================================
BUF                  := $(BIN_DIR)/buf
YQ                   := $(BIN_DIR)/yq
MOCKERY              := $(BIN_DIR)/mockery

# =============================================================================
.DEFAULT_GOAL := help
.PHONY: help all build run generate generate-config proto-generate scaffold tools clean test lint add-service

help:
	@echo ""
	@echo "Доступные команды:"
	@echo ""
	@echo "  make tools                                 Установить инструменты (buf, yq, mockery, protoc plugins)"
	@echo "  make generate                              Генерация proto + stubs + config keys"
	@echo "  make generate-config                       Генерация только config keys"
	@echo "  make build                                 Сборка → ./bin/$(PROJECT_NAME)"
	@echo "  make run                                   Запуск"
	@echo "  make test                                  Тесты"
	@echo "  make lint                                  Buf lint"
	@echo "  make tidy                                  go mod tidy"
	@echo "  make clean                                 Очистка"
	@echo ""
	@echo "  make add-service users                     → proto/users/v1/..."
	@echo "  make add-service users v2                  → proto/users/v2/..."
	@echo ""

# =============================================================================
# Tools
# =============================================================================
tools: $(BIN_DIR) buf yq protoc-plugins mockery

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

buf:
	curl -sSL \
	  "https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/buf-Linux-x86_64" \
	  -o $(BUF)
	chmod +x $(BUF)

yq:
	curl -sSL \
	  "https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_linux_amd64" \
	  -o $(YQ)
	chmod +x $(YQ)
	@echo "yq $(YQ_VERSION) installed → $(YQ)"

protoc-plugins:
	GOBIN=$(BIN_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO)
	GOBIN=$(BIN_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC)
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(GRPC_GATEWAY_VERSION)
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(OPENAPI_VERSION)

mockery:
	GOBIN=$(BIN_DIR) go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)

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
# Сборка и запуск
# =============================================================================
build: tidy generate
	CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags "-s -w -X main.version=$(CURRENT_DATE)" \
		-o $(BIN_DIR)/$(PROJECT_NAME) \
		./cmd/$(PROJECT_NAME)/main.go
	@echo "→ $(BIN_DIR)/$(PROJECT_NAME)"

run: build
	@$(BIN_DIR)/$(PROJECT_NAME)

test: tidy
	go test -v -race -cover ./...

tidy:
	go mod tidy

# =============================================================================
# Добавление сервиса
# =============================================================================
.PHONY: add-service

add-service: ; @bash script/add_service.sh $(filter-out $@,$(MAKECMDGOALS))
%:
	@:

# =============================================================================
# Очистка и линтинг
# =============================================================================
clean:
	rm -rf $(GEN_DIR) $(BIN_DIR)

lint:
	@$(BUF) lint $(PROTO_DIR)