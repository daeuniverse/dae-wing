#
#  SPDX-License-Identifier: AGPL-3.0-only
#  Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
#
SHELL := /bin/bash
OUTPUT ?= dae-wing

# Get version from .git.
date=$(shell git log -1 --format="%cd" --date=short | sed s/-//g)
count=$(shell git rev-list --count HEAD)
commit=$(shell git rev-parse --short HEAD)
ifeq ($(wildcard .git/.),)
	VERSION ?= unstable-0.nogit
else
	VERSION ?= unstable-$(date).r$(count).$(commit)
endif

.PHONY: schema-resolver dae-deps deps dae-wing vendor

all: dae-wing

deps: schema-resolver dae-deps

dae-wing: deps
	go build -o $(OUTPUT) -trimpath -ldflags "-s -w -X github.com/v2rayA/dae/cmd.Version=$(VERSION)" .

vendor:
	go mod vendor

schema-resolver: vendor
	unset GOOS && \
	unset GOARCH && \
	unset GOARM && \
	go generate ./...

dae-deps: DAE_VERSION := $(shell grep '\s*github.com/v2rayA/dae\s*v' go.mod | awk '{print $2}' | cut -d- -f3)
dae-deps: BUILD_DIR := ./build-dae-ebpf
dae-deps: vendor
	git clone https://github.com/v2rayA/dae build-dae-ebpf && \
	pushd "$(BUILD_DIR)" && \
	git checkout $(DAE_VERSION) && \
	git submodule update --init && \
	make ebpf && \
	popd && \
	cp "$(BUILD_DIR)"/control/bpf_bpf*.{go,o} vendor/github.com/v2rayA/dae/control/ && \
	rm -rf "$(BUILD_DIR)"
