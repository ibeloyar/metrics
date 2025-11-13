GO = go
MAKE = make

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
	metricstest_v2 -test.v -test.run=^TestIteration$(ITER) -binary-path=./cmd/server/server -agent-binary-path=cmd/agent/agent -source-path=. -server-port=8080
else
	@echo "Require variable ITER not found"
endif

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