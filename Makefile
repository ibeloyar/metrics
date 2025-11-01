GO = go

.DEFAULT_GOAL := help

.PHONY: build
build:
	$(GO) build -o cmd/server/server cmd/server/main.go
	$(GO) build -o cmd/agent/agent cmd/agent/main.go

.PHONY: run
run:
	$(GO) run cmd/server/main.go
	$(GO) run cmd/agent/main.gos

.PHONY: test
test:
	$(GO) test -v ./... | { grep -v 'no test files'; true; }

.PHONY: help
help:
	@echo "command           | description"
	@echo "===================================================="
	@echo "run               | run server server and then agent"
	@echo "build             | build agent and server"
	@echo "test              | run tests with 'clean' out"