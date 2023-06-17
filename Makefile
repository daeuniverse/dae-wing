#
#  SPDX-License-Identifier: AGPL-3.0-only
#  Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
#
SHELL := /bin/bash
OUTPUT ?= ./dae-wing
APPNAME ?= dae-wing
DESCRIPTION ?= $(APPNAME) is a integration solution of dae, API and UI.
VERSION ?= 0.0.0.unknown
LDFLAGS = '-s -w -X github.com/daeuniverse/dae-wing/cmd.Version=$(VERSION) -X github.com/daeuniverse/dae-wing/cmd.AppName=$(APPNAME) -X "github.com/daeuniverse/dae-wing/cmd.Description=$(DESCRIPTION)"'

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

SCHEMA_INPUT = graphql/service/config/global/generated_input.go
SCHEMA_RESOLVER = graphql/service/config/global/generated_resolver.go
SCHEMA_READY = $(SCHEMA_INPUT) $(SCHEMA_RESOLVER)
DAE_BPFEB = dae-core/control/bpf_bpfeb.o
DAE_BPFEL = dae-core/control/bpf_bpfel.o
DAE_READY = $(DAE_BPFEB) $(DAE_BPFEL)
DEPS = $(SCHEMA_READY) $(DAE_READY)

all: dae-wing
.PHONY: all

clean:
	rm -f $(DEPS)

deps: $(DEPS)
.PHONY: deps

$(SCHEMA_INPUT): $(SCHEMA_RESOLVER)
$(SCHEMA_RESOLVER):
	@unset GOOS && \
	unset GOARCH && \
	unset GOARM && \
	unset CC && \
	go generate ./...

schema-resolver: $(SCHEMA_RESOLVER)
.PHONY: schema-resolver

dae-core: .gitmodules
	@git submodule update --init --recursive dae-core

$(DAE_BPFEL): $(DAE_BPFEB)
$(DAE_BPFEB): dae-core
	cd dae-core && \
	$(MAKE) ebpf

dae-wing: $(DEPS)
	go build -o $(OUTPUT) -trimpath -ldflags $(LDFLAGS) .
.PHONY: dae-wing

bundle: $(DEPS)
	$(call check_defined, WEB_DIST)
	@if [ $$(realpath -m "$(WEB_DIST)") != $$(realpath -m "webrender/web") ]; then \
		rm -r webrender/web 2>/dev/null; \
		cp -r $(WEB_DIST) webrender/web; \
		find webrender/web -type f -size +16k ! -name "*.gz" ! -name "*.woff"  ! -name "*.woff2" -exec sh -c '\
			gzip -9 -k '{}'; \
			if [ "$$(stat -c %s {})" \< "$$(stat -c %s {}.gz)" ]; then \
				rm {}.gz; \
			else \
				rm {}; \
			fi' ';' ; \
	fi && \
	go build -tags=embedallowed -o $(OUTPUT) -trimpath -ldflags $(LDFLAGS) .
.PHONY: bundle

fmt:
	go fmt ./...
.PHONY: fmt
