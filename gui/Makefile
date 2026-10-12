.SILENT:

#---------------------------------------------------------------
# GUI Project Variables
#---------------------------------------------------------------
GO_PACKAGE_NAME := go1.26.5.linux-amd64.tar.gz
GO_DOWNLOAD_URL := https://go.dev/dl/$(GO_PACKAGE_NAME)

FT_CACHE_DIR := $(shell dirname $(PWD))/.cache
GO_DIR := $(FT_CACHE_DIR)/go
GO_BIN_PATH := $(GO_DIR)/go/bin
GO_PKG_PATH := $(GO_DIR)/pkg

export GOPATH := $(GO_PKG_PATH)

#---------------------------------------------------------------
# Docker Configuration Variables
#---------------------------------------------------------------
GO_SERVICE_NAME := tap-gui
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
#---------------------------------------------------------------

DOCKER := docker

BINARY_NAME = tap-gui

GO := $(GO_BIN_PATH)/go
GO_FMT = $(GO_BIN_PATH)/gofmt
GO_BUILD = $(GO) build
GO_RUN = $(GO) run
GO_CLEAN = $(GO) clean

RM = rm -f

GREEN = /bin/echo -e "\x1b[32m\#\# $1\x1b[0m"

.DEFAULT_GOAL := help

##-----------------------------------
## Main Commands
##-----------------------------------
.PHONY: install
install: $(GO_DIR) ## Install all Go dependencies
	$(call GREEN,"Installing Go dependencies...")
	$(GO) mod download
	$(GO) mod tidy

.PHONY: build
build: $(GO_DIR) ## Build the Go binary executable
	$(call GREEN,"Building $(BINARY_NAME)...")
	$(GO_BUILD) -o $(BINARY_NAME)

.PHONY: run
run: $(GO_DIR) ## Run the Go project
	$(GO_RUN) .

.PHONY: clean
clean: ## Clean Go build cache
	$(call GREEN,"Cleaning Go cache...")
	$(GO_CLEAN)

.PHONY: fclean
fclean: clean ## Remove build cache and binary executable
	$(call GREEN,"Removing executable $(BINARY_NAME)...")
	$(RM) $(BINARY_NAME)

.PHONY: re
re: fclean build ## Rebuild project from scratch

.PHONY: go-download
go-download: ## Download go packages executable
	$(call GREEN,"Downloading Go packages...")
	mkdir -p $(GO_DIR)
	curl -L -o $(FT_CACHE_DIR)/$(GO_PACKAGE_NAME) $(GO_DOWNLOAD_URL)
	tar -C $(GO_DIR) -xzf $(FT_CACHE_DIR)/$(GO_PACKAGE_NAME)
	rm $(FT_CACHE_DIR)/$(GO_PACKAGE_NAME)

.PHONY: cache-clear
cache-clear: ## Clear the cache directory
	$(call GREEN,"Clearing cache directory...")
	chmod -R u+w $(FT_CACHE_DIR)
	rm -rf $(FT_CACHE_DIR)

##
##-----------------------------------
## Docker Commands
##-----------------------------------
docker-build: ## Build the Docker image
	$(call GREEN,"Building Docker image...")
	$(DOCKER) build -t $(GO_SERVICE_NAME) .

docker-compile: ## Compile the Go binary inside a Docker container
	$(call GREEN,"Running Docker container...")
	$(DOCKER) run --rm -v $(PWD):/output $(GO_SERVICE_NAME) cp /app/$(BINARY_NAME) /output/

docker-remove: ## Remove the Docker image
	$(call GREEN,"Removing Docker image...")
	$(DOCKER) rmi $(GO_SERVICE_NAME)
	$(DOCKER) volume prune

##
##-----------------------------------
## Other Commands
##-----------------------------------
.PHONY: help
help: ## List available commands
	grep -E '(^[a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

##

#
#-----------------------------------
# Dependencies
#-----------------------------------
$(GO_DIR): ## Ensure Go is downloaded and extracted
	$(MAKE) go-download