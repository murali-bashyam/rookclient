
include build/common.mk

.PHONY: all
all: build

# ====================================================================================
# Build Options

# set the shell to bash in case some environments use sh
SHELL := /bin/bash

# ====================================================================================
# Setup projects
# setup go projects

GO_PROJECT := github.com/murali-bashyam/rookclient

include build/golang.mk

# ====================================================================================
# Build Targets

build.version:
	@mkdir -p $(OUTPUT_DIR)

build.common: build.version mod.check
	@$(MAKE) go.init
	@$(MAKE) go.validate

build: build.common ## Build source code for host platform.
	@$(MAKE) go.build

build.image: build
	@cp $(OUTPUT_DIR)/cmd image
	@sudo docker build -t $(BUILD_REGISTRY)/$(ROOKCLIENT_IMG):$(ROOKCLIENT_VERSION) image
	@rm -f image/cmd

vet: ## Runs lint checks on go sources.
	@$(MAKE) go.init
	@$(MAKE) go.vet

fmt: ## Check formatting of go sources.
	@$(MAKE) go.init
	@$(MAKE) go.fmt

mod.check: go.mod.check ## Check if any go modules changed.
mod.update: go.mod.update ## Update all go modules.

clean: ## Remove all files that are created by building.
	@$(MAKE) go.mod.clean
	@rm -fr $(OUTPUT_DIR) 

distclean: clean ## Remove all files that are created by building or configuring.
	@rm -rf $(CACHE_DIR)

.PHONY: all build.common 
.PHONY: build vet fmt codegen mod.check clean distclean

