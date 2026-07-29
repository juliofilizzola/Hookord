# Binary name
BINARY_NAME=hookord
CMD_PATH=./cmd/hookord

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Detect OS
ifeq ($(OS),Windows_NT)
    BINARY_EXT=.exe
    RM=del /q
    FIX_PATH=$(subst /,\,$1)
else
    BINARY_EXT=
    RM=rm -f
    FIX_PATH=$1
endif

BINARY=$(BINARY_NAME)$(BINARY_EXT)

.PHONY: all build run test clean tidy dev

all: build

build:
	$(GOBUILD) -o $(BINARY) $(CMD_PATH)

run: build
	./$(BINARY)

dev:
	docker-compose up app-dev

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	$(RM) $(BINARY)

tidy:
	$(GOMOD) tidy

# Help target
help:
	@echo "Available commands:"
	@echo "  make build  - Build the binary"
	@echo "  make run    - Build and run the binary"
	@echo "  make dev    - Run with hot-reload via Docker"
	@echo "  make test   - Run tests"
	@echo "  make clean  - Remove binary and clean cache"
	@echo "  make tidy   - Run go mod tidy"
