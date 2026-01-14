GO = go
MAKE = make
DB_HOST=192.168.0.101
DB_USER=metrics
DB_NAME=metrics
DB_PASS=metrics
DB_PORT=5432
DB_STRING="postgres://$(DB_NAME):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"
DB_MIGRATIONS_PATH="./migrations"

.DEFAULT_GOAL := help

.PHONY: build
build:
	$(GO) build -o cmd/server/server cmd/server/main.go
	$(GO) build -o cmd/agent/agent cmd/agent/main.go

.PHONY: run-agent
run-agent:
	$(GO) run cmd/agent/main.go

.PHONY: run-server
run-server:
	$(GO) run cmd/server/main.go


.PHONY: test
test:
	$(GO) test -v ./... | { grep -v 'no test files'; true; }

.PHONY: test_cover
test_cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	rm coverage.out

.PHONY: test_iter
test_iter:
ifdef ITER
	rm -rf ./data
	metricstest_v2 -test.v \
	-test.run=^TestIteration$(ITER) \
	-binary-path=./cmd/server/server \
	-agent-binary-path=cmd/agent/agent \
	-source-path=. \
	-server-port=8080 \
	-database-dsn="host=$(DB_HOST) user=$(DB_USER) password=$(DB_PASS) dbname=$(DB_NAME) sslmode=disable" \
	-file-storage-path=data/metrics.json \
	-key=test
else
	@echo "Require variable ITER not found"
endif

.PHONY: migrate-up
migrate-up:
	migrate \
	-path $(DB_MIGRATIONS_PATH) \
	-database $(DB_STRING) up

.PHONY: migrate-down
migrate-down:
	migrate \
	-path $(DB_MIGRATIONS_PATH) \
	-database $(DB_STRING) down


.PHONY: migrate-create
migrate-create:
ifdef NAME
	migrate create \
    	-ext sql \
    	-dir $(DB_MIGRATIONS_PATH) \
    	-seq $(NAME)
else
	@echo "Require variable NAME not found"
endif


.PHONY: install-tools
install-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest # golang-migrate CLI

.PHONY: help
help:
	@echo "command           | description"
	@echo "===================================================="
	@echo "run-agent         | run metric agent"
	@echo "run-server        | run metric server"
	@echo "build             | build agent and server"
	@echo "test              | run tests with 'clean' out"
	@echo "test_cover        | run tests with coverage info"
	@echo "test_iter         | run tests for iteration X; EXAMPLE: make ITER=5 test_iter"
	@echo "migrate-up        | run UP migrations"
	@echo "migrate-down      | run DOWN migrations"
	@echo "migrate-create    | run create migration with NAME; EXAMPLE: make NAME=add_users migrate-create"
	@echo "install-tools     | install libs for work with project"