SHELL := /bin/bash

.PHONY: mock-claude
mock-claude:
	@echo "[mock-claude] running scripts/mock-claude.sh"
	@/bin/bash scripts/mock-claude.sh

.PHONY: test-go
test-go:
	@echo "[test-go] running go test ./..."
	@mkdir -p $(CURDIR)/.gocache
	@GOCACHE=$(CURDIR)/.gocache go test ./...

.PHONY: test-frontend
test-frontend:
	@echo "[test-frontend] running npm run test (frontend)"
	@cd frontend && npm run test

.PHONY: test
test: test-go test-frontend

.PHONY: dev
dev:
	@echo "[dev] starting wails dev"
	@wails dev
