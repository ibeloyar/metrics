GO = go
MAKE = make
DB_HOST=192.168.0.102
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
	-file-storage-path=data/metrics.json
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

.PHONY: profile-base
profile-base:
	@mkdir -p ./profiles
	@wrk -t4 -c100 -d30s http://localhost:8080 > /dev/null 2>&1 &
	@echo "Wait... 10s"
	@sleep 10
	@curl http://localhost:8080/debug/pprof/heap > profiles/base.pprof

.PHONY: profile-result
profile-result:
	@mkdir -p ./profiles
	@wrk -t4 -c100 -d30s http://localhost:8080 > /dev/null 2>&1 &
	@echo "Wait... 10s"
	@sleep 10
	@curl http://localhost:8080/debug/pprof/heap > profiles/result.pprof

.PHONY: profile-diff
profile-diff:
	@go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

.PHONY: gofmt
gofmt:
	@gofmt -w ./..

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
	@echo "profile-base      | base pprof heap check"
	@echo "profile-result    | result pprof heap check"
	@echo "profile-diff      | show pprof result difference"
	@echo "gofmt             | format code"