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

.PHONY: help
help:
	@echo "command           | description"
	@echo "===================================================="
	@echo "run-agent         | run metric agent"
	@echo "run-server        | run metric server"
	@echo "build             | build agent and server"
	@echo "test              | run tests with 'clean' out"