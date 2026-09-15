.SILENT:

IS_DOCKER := $(shell docker info >/dev/null 2>&1 && echo 1)

#---------------------------------------------------------------
# Docker Configuration
#---------------------------------------------------------------
GO_SERVICE_NAME := tap-gui
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
#---------------------------------------------------------------

DOCKER := docker

BINARY_NAME = tap-gui

GO := go
GO_BUILD = $(GO) build
GO_RUN = $(GO) run
GO_CLEAN = $(GO) clean

RM = rm -f

GREEN = /bin/echo -e "\x1b[32m\#\# $1\x1b[0m"
RED = /bin/echo -e "\x1b[31m\#\# $1\x1b[0m"

.DEFAULT_GOAL := help

##-----------------------------------
## Go Commands
##-----------------------------------
.PHONY: install
install:
	$(call GREEN,"Installing Go dependencies...")
	$(GO) mod download
	$(GO) mod tidy

.PHONY: build
build: ## Build the Go binary executable
	$(call GREEN,"Building $(BINARY_NAME)...")
	@if [ "$(IS_DOCKER)" = "1" ]; then \
		$(MAKE) docker-build; \
		$(MAKE) docker-compile; \
	else \
		$(GO_BUILD) -o $(BINARY_NAME) .; \
	fi

.PHONY: run
run: ## Run the Go project
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

##
##-----------------------------------
## Help
##-----------------------------------

.PHONY: help
help: ## List available commands
	grep -E '(^[a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'