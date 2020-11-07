SHELL := /bin/bash

GOHOSTOS := $(shell go env GOHOSTOS)
GOHOSTARCH := $(shell go env GOHOSTARCH)
HOST_PLATFORM := $(GOHOSTOS)_$(GOHOSTARCH)

# include the common make file
COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell cd $(COMMON_SELF_DIR)/.. && pwd -P))
endif

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/bin
endif

ifeq ($(origin WORK_DIR), undefined)
WORK_DIR := $(ROOT_DIR)
endif

ifeq ($(origin CACHE_DIR), undefined)
CACHE_DIR := $(ROOT_DIR)/.cache
endif

TOOLS_DIR := $(CACHE_DIR)/tools
TOOLS_HOST_DIR := $(TOOLS_DIR)/$(HOST_PLATFORM)

ifeq ($(origin HOSTNAME), undefined)
HOSTNAME := $(shell hostname)
endif

ROOKCLIENT_IMG := rookclient
ROOKCLIENT_VERSION := v1.0
BUILD_REGISTRY := muralibashyam

define \n


endef
