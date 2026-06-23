SHELL := /bin/bash

GO ?= go
NPM ?= npm

# Pick a default C++ compiler per-OS. macOS uses Homebrew LLVM clang++;
# elsewhere (Linux) use the system c++. An explicitly set CXX (env or
# command line) always wins; only make's implicit/unset value is overridden.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
CXX_DEFAULT := /usr/local/opt/llvm/bin/clang++
else
CXX_DEFAULT := c++
endif

ifeq ($(origin CXX),default)
CXX := $(CXX_DEFAULT)
endif
ifeq ($(origin CXX),undefined)
CXX := $(CXX_DEFAULT)
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
