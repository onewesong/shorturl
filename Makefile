SHELL := /bin/bash

APP_NAME := shorturl
MAIN_PKG := ./cmd/shorturl
BUILD_DIR := bin
GO_FILES := $(shell find . -name '*.go' -type f)
ADMIN_DIR := web/admin

define load_local_env
set -a; \
if [ -f .env ]; then \
	source ./.env; \
fi; \
set +a;
endef

.PHONY: help fmt tidy test build run clean check admin-install admin-dev admin-test admin-build docker-up docker-down docker-logs docker-build

help:
	@printf "Available targets:\n"
	@printf "  make fmt          - format Go files\n"
	@printf "  make tidy         - tidy go modules\n"
	@printf "  make test         - run Go tests\n"
	@printf "  make build        - build binary to $(BUILD_DIR)/$(APP_NAME)\n"
	@printf "  make run          - run service locally and auto-load .env\n"
	@printf "  make admin-install - install admin frontend deps\n"
	@printf "  make admin-dev     - run admin frontend dev server and auto-load .env\n"
	@printf "  make admin-test    - run admin frontend tests\n"
	@printf "  make admin-build   - build admin frontend assets\n"
	@printf "  make docker-build  - build docker image via compose\n"
	@printf "  make docker-up     - start service with docker compose\n"
	@printf "  make docker-down   - stop docker compose services\n"
	@printf "  make docker-logs   - tail docker compose logs\n"
	@printf "  make check        - run fmt, tidy, test\n"
	@printf "  make clean        - remove build artifacts\n"

fmt:
	gofmt -w $(GO_FILES)

tidy:
	go mod tidy

test:
	go test ./... -count=1

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PKG)

run:
	@$(load_local_env) go run $(MAIN_PKG)

admin-install:
	cd $(ADMIN_DIR) && npm install

admin-dev:
	@$(load_local_env) cd $(ADMIN_DIR) && npm run dev

admin-test:
	cd $(ADMIN_DIR) && npm test

admin-build:
	@$(load_local_env) cd $(ADMIN_DIR) && npm run build

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f shorturl

clean:
	rm -rf $(BUILD_DIR)

check: fmt tidy test
