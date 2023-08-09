#
#  SPDX-License-Identifier: AGPL-3.0-only
#  Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
#
SHELL := /bin/bash
OUTPUT ?= ./dae-wing
APPNAME ?= dae-wing
DESCRIPTION ?= $(APPNAME) is a integration solution of dae, API and UI.
VERSION ?= 0.0.0.unknown
LDFLAGS = '-s -w -X github.com/daeuniverse/dae-wing/db.AppVersion=$(VERSION) -X github.com/daeuniverse/dae-wing/db.AppName=$(APPNAME) -X "github.com/daeuniverse/dae-wing/db.AppDescription=$(DESCRIPTION)"'

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

# Do NOT remove the line below. This line is for CI.
#export GOMODCACHE=$(PWD)/go-mod

all: dae-wing
.PHONY: all

deps: schema-resolver $(DAE_READY)
.PHONY: deps

DAE_READY = dae-core/control/bpf_bpfeb.o
DAE_EBPF_SRC = dae-core/control/kern/tproxy.c

schema-resolver: $(DAE_READY)
	@unset GOOS && \
	unset GOARCH && \
	unset GOARM && \
	unset CC && \
	go generate ./...
.PHONY: schema-resolver

$(DAE_EBPF_SRC):
	@git submodule update --init --recursive dae-core

$(DAE_READY): .gitmodules $(DAE_EBPF_SRC)
	cd dae-core && \
	make ebpf && \
	cd ../ && \
	touch $@

dae-wing: deps
	go build -o $(OUTPUT) -trimpath -buildmode=pie -ldflags $(LDFLAGS) .
.PHONY: dae-wing

bundle: deps
	$(call check_defined, WEB_DIST)
	@if [ $$(realpath -m "$(WEB_DIST)") != $$(realpath -m "webrender/web") ]; then \
		rm -r webrender/web 2>/dev/null; \
		cp -r $(WEB_DIST) webrender/web; \
		find webrender/web -type f -size +4k ! -name "*.gz" ! -name "*.woff"  ! -name "*.woff2" -exec sh -c '\
			echo "{}"; \
			gzip -9 -k "{}"; \
			if [ $$(stat -c %s "{}") \< $$(stat -c %s "{}".gz) ]; then \
				rm "{}".gz; \
			else \
				rm "{}"; \
			fi' ';' ; \
	fi && \
	go build -tags=embedallowed -o $(OUTPUT) -trimpath -buildmode=pie -ldflags $(LDFLAGS) .
.PHONY: bundle

fmt:
	go fmt ./...
.PHONY: fmt
