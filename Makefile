#
#  SPDX-License-Identifier: AGPL-3.0-only
#  Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
#
SHELL := /bin/bash
OUTPUT ?= dae-wing
DAE_READY = vendor/github.com/daeuniverse/dae/control/headers

include functions.mk

# Get version from .git.
date=$(shell git log -1 --format="%cd" --date=short | sed s/-//g)
count=$(shell git rev-list --count HEAD)
commit=$(shell git rev-parse --short HEAD)
ifeq ($(wildcard .git/.),)
	VERSION ?= unstable-0.nogit
else
	VERSION ?= unstable-$(date).r$(count).$(commit)
endif

.PHONY: schema-resolver deps dae-wing bundle

all: dae-wing

deps: schema-resolver $(DAE_READY)

vendor: go.mod go.sum
	go mod vendor && touch vendor

schema-resolver: vendor
	unset GOOS && \
	unset GOARCH && \
	unset GOARM && \
	go generate ./...

$(DAE_READY): DAE_VERSION := $(shell grep '\s*github.com/daeuniverse/dae\s*v' go.mod | rev | cut -d' ' -f1 | cut -d- -f1 | rev )
$(DAE_READY): BUILD_DIR := ./build-dae-ebpf
$(DAE_READY): vendor
	git clone --single-branch -- https://github.com/daeuniverse/dae $(BUILD_DIR) && \
	pushd "$(BUILD_DIR)" && \
	git checkout $(DAE_VERSION) && git submodule update --init --recursive && \
	make ebpf && \
	popd && \
	cp "$(BUILD_DIR)"/control/bpf_bpf*.{go,o} vendor/github.com/daeuniverse/dae/control/ && \
	rm -rf "$(BUILD_DIR)" && \
	touch $@

fmt:
	go fmt ./...

dae-wing: deps
	go build -o $(OUTPUT) -trimpath -ldflags "-s -w -X github.com/daeuniverse/dae/cmd.Version=$(VERSION)" .

bundle: deps
	$(call check_defined, WEB_DIST)
	@if [ $$(realpath "$(WEB_DIST)") != $$(realpath "webrender/web") ]; then \
		rm -r webrender/web 2>/dev/null; \
		cp -r $(WEB_DIST) webrender/web; \
	fi && \
	go build -tags=embedallowed -o $(OUTPUT) -trimpath -ldflags "-s -w -X github.com/daeuniverse/dae/cmd.Version=$(VERSION)" .
