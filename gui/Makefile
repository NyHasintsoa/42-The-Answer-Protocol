.SILENT:

IS_DOCKER := $(shell docker info >/dev/null 2>&1 && echo 1)

#---------------------------------------------------------------
# Docker Configuration
#---------------------------------------------------------------
GO_SERVICE_NAME := app
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
#---------------------------------------------------------------

DOCKER := docker
COMPOSE := USER_ID=$(USER_ID) GROUP_ID=$(GROUP_ID) $(DOCKER) compose
EXEC := $(COMPOSE) exec $(GO_SERVICE_NAME)

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

.PHONY: build
build: ## Build the Go binary executable
	$(call GREEN,"Building $(BINARY_NAME)...")
	$(GO_BUILD) -o $(BINARY_NAME) .

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

.PHONY: docker-bash
docker-bash: ## Terminal inside the Go container
	$(COMPOSE) exec -it $(GO_SERVICE_NAME) sh

.PHONY: docker-build
docker-build: ## Build docker image
	$(call GREEN,"Build docker image")
	$(COMPOSE) build

.PHONY: docker-up
docker-up: ## Start docker containers
	$(call GREEN,"Start docker containers")
	$(COMPOSE) up --build -d

.PHONY: docker-down
docker-down: ## Stop docker containers
	$(call GREEN,"Stop docker containers")
	$(COMPOSE) down

.PHONY: docker-restart
docker-restart: ## Restart docker containers
	$(call GREEN,"Restart docker containers")
	make docker-down
	make docker-up

.PHONY: docker-logs
docker-logs: ## Show docker logs
	$(call GREEN,"Show docker logs")
	$(COMPOSE) logs -f

##
##-----------------------------------
## Help
##-----------------------------------

.PHONY: help
help: ## List available commands
	grep -E '(^[a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'