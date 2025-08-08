SHELL := /bin/bash

APP_NAME ?= feitian
FTINIT_NAME ?= ftinit
BIN_DIR ?= bin
GO ?= go

.PHONY: help deps tidy build build-app build-ftinit run test clean

help:
	@echo "Available targets:"
	@echo "  tidy          - go mod tidy"
	@echo "  build         - build app and ftinit binaries into $(BIN_DIR)/"
	@echo "  build-app     - build main app binary into $(BIN_DIR)/$(APP_NAME)"
	@echo "  build-ftinit  - build scaffold tool into $(BIN_DIR)/$(FTINIT_NAME)"
	@echo "  run           - run the server with default config path"
	@echo "  test          - run unit tests"
	@echo "  clean         - remove $(BIN_DIR)/"

# Aliases
 deps: tidy

tidy:
	$(GO) mod tidy

build: build-app build-ftinit

build-app:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) ./cmd

build-ftinit:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(FTINIT_NAME) ./cmd/ftinit

run:
	$(GO) run ./cmd -c config -cPath "./,./configs/"

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR)
