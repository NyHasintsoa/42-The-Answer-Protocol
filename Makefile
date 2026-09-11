IS_DOCKER := $(shell docker info >/dev/null 2>&1 && echo 1)

test:
ifeq ($(IS_DOCKER),1)
	@echo "Docker is running!"
else
	@echo "Docker is not running."
endif