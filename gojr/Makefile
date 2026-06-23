SHELL := /bin/bash

GO ?= go
NPM ?= npm

ifeq ($(origin CXX),default)
CXX := /usr/local/opt/llvm/bin/clang++
endif
ifeq ($(origin CXX),undefined)
CXX := /usr/local/opt/llvm/bin/clang++
endif

GOBIN_VALUE := $(shell $(GO) env GOBIN)
ifeq ($(GOBIN_VALUE),)
GOJR_BINDIR := $(shell $(GO) env GOPATH)/bin
else
GOJR_BINDIR := $(GOBIN_VALUE)
endif

GOJR_EXE := $(GOJR_BINDIR)/gojr
NPM_STAMP := node_modules/.gojr-npm-ci.stamp

.PHONY: all npm-deps js install typecheck test repl print-bin clean

all: install

$(NPM_STAMP): package.json package-lock.json
	$(NPM) ci
	@touch $@

npm-deps: $(NPM_STAMP)

js: npm-deps
	$(NPM) run build

install: js
	CXX="$(CXX)" $(GO) install ./cmd/gojr
	@printf 'installed %s\n' "$(GOJR_EXE)"

typecheck: npm-deps
	$(NPM) run typecheck

test: npm-deps
	$(NPM) run typecheck
	$(NPM) test
	CXX="$(CXX)" $(GO) test ./cmd/gojr

repl: install
	"$(GOJR_EXE)"

print-bin:
	@printf '%s\n' "$(GOJR_EXE)"

clean:
	rm -rf dist "$(NPM_STAMP)"
