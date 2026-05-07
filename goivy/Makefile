.PHONY: all build test test-conform test-js test-js-watch test-browser test-control-browser test-web test-vue build-vue vitest-setup control-browser-setup webvue-setup unimerge-audit unimerge-audit-strict clean z3ivy z3-build tr trt vs golden setup this bootstrap

# Makefile for goivy (Go port of Ivy)
#
# The Go Z3 bridge links against the same Z3 4.7.1 fork that Python Ivy uses.
# Set PYIVY_ROOT to the Python Ivy source tree root if it's not at ~/pyivy/ivy.

PYIVY_ROOT ?= $(HOME)/pyivy/ivy
Z3_SRC     := $(PYIVY_ROOT)/submodules/z3
Z3_BUILD   := $(Z3_SRC)/build
Z3_API     := $(Z3_SRC)/src/api
Z3IVY      := z3ivy

# Headers needed by the Go CGO bridge.
Z3_HEADERS := z3.h z3_api.h z3_macros.h z3_v1.h z3_algebraic.h \
              z3_ast_containers.h z3_fixedpoint.h z3_fpa.h z3_interp.h \
              z3_optimization.h z3_polynomial.h z3_rcf.h z3_spacer.h

all: build
	ivyweb

# for bashisms like &>
SHELL := /bin/bash

golden: # parse with python ivy_check and goivy_check_xtrace. full xtrace output is written now to the files: out.go.xtrace; out.py.xtrace
	if [ -f log.red ]; then mv log.red log.red.prev; fi
	go test -v -timeout=0 -count=1 -run TestOrdLive &> log.red

golden-2hr: # takes 2 hours to check all isolates.
	if [ -f log.golden.2hr ]; then mv log.golden.2hr log.golden.2hr.prev; fi
	go test -v -timeout=0 -count=1 -run Test2hrOrdLive &> log.golden.2hr

rfn: # TestRfnAbsIso start 3 hours in on golden-2hr. start with it. do we go faster?
	if [ -f log.rfn ]; then mv log.rfn log.rfn.prev; fi
	go test -v -timeout=0 -count=1 -run TestRfnAbsIso &> log.rfn

syslive: # parser/golden_test.go:489 TestSysLiveIso
	if [ -f log.syslive ]; then mv log.syslive log.syslive.prev; fi
	go test -v -timeout=0 -count=1 -run TestSysLiveIso &> log.syslive

this: # parser/golden_test.go:496 TestThisIso
	if [ -f log.this ]; then mv log.this log.this.prev; fi
	go test -v -timeout=0 -count=1 -run TestThisIso &> log.this

golden-v: # to see all productions not just last 10
	go test -v -run TestVerboseOrdLive &> log.all.til.div

golden-nonstop: # to see all productions and not stop at first diff.
	go test -v -run TestVerboseNonstopOrdLive &> log.nonstop

tlb:
	if [ -f log.tlb ]; then mv log.tlb log.tlb.prev; fi
	go test -v -timeout=0 -count=1 -run TestIvyTlbModel &> log.tlb

tilelink: # an ivy v1.1 .ivy file: ivy/ivy-lang-examples/examples/tilelink/tilelink1.ivy
	if [ -f log.tilelink ]; then mv log.tilelink log.tilelink.prev; fi
	go test -v -timeout=0 -count=1 -run TestIvy_1dot1_tilelink1_model &> log.tilelink

ord: # xtraced 
	~/ivy/pyivy/goivy-venv/bin/ivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

ordo: # not xtraced
	time python3 -O ~/ivy/pyivy/goivy-venv/bin/ivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy



# Create z3ivy/ with symlinks to the Ivy Z3 fork's headers and library.
#z3ivy: $(Z3IVY)/lib/libz3.dylib

get_goyacc:
	go install golang.org/x/tools/cmd/goyacc@latest

# static linked to z3vendor now, for completeness within this repo.
z3ivy: z3vendor/z3/build/libz3.a
	echo "must manually build static libz3.a using z3vendor/z3/compile.sh"

$(Z3IVY)/lib/libz3.dylib: $(Z3_BUILD)/libz3.dylib
	@mkdir -p $(Z3IVY)/include $(Z3IVY)/lib
	@for h in $(Z3_HEADERS); do \
		ln -sf $(Z3_API)/$$h $(Z3IVY)/include/$$h; \
	done
	@ln -sf $(Z3_BUILD)/libz3.dylib $(Z3IVY)/lib/libz3.dylib
	@echo "z3ivy: linked to Z3 4.7.1 fork at $(Z3_SRC)"

# Build the Z3 fork from the submodule (only needed if not already built).
z3-build:
	@if [ ! -f $(Z3_BUILD)/libz3.dylib ]; then \
		echo "Building Z3 fork from $(Z3_SRC)..."; \
		cd $(Z3_SRC) && python3 scripts/mk_make.py --python --prefix=$(PYIVY_ROOT); \
		cd $(Z3_BUILD) && $(MAKE) -j4; \
	else \
		echo "Z3 fork already built at $(Z3_BUILD)/libz3.dylib"; \
	fi

Z3IVY_ABS := $(CURDIR)/$(Z3IVY)

build: # grammer_gen
	go install  -tags xtracer_off ./cmd/ivyweb
	go install  -tags xtracer_off ./cmd/goivy_check
	go install  -tags xtracer_off ./cmd/ivy-control
	go install  -tags xtracer_off ./cmd/ivy-control-admin

grammar_gen: # be sure .y -> .go up to date
	goyacc -o parser_grammar_v17.go -p parser17 parser_grammar_v17.y
	go run parser_patch_parser.go
	cd conceptspace && go generate # cs_grammar.y
	goyacc -o v12_grammar_v12.go -p lalr12 v12_grammar_v12.y
	goyacc -o v16_grammar_v16.go -p lalr16 v16_grammar_v16.y
	goyacc -o lalr_logicparser_grammar_v17.go -p lalr17 lalr_logicparser_grammar_v17.y
	cd evparser && go generate # ev_grammar.y


GOBIN ?= $(HOME)/go/bin
NODE_MODULES_DIR := webui/node_modules
VITEST_BIN := $(NODE_MODULES_DIR)/.bin/vitest
PLAYWRIGHT_BIN := $(NODE_MODULES_DIR)/.bin/playwright
WEBUI_NODE_MODULES_STAMP := $(NODE_MODULES_DIR)/.package-lock.stamp
CONTROL_NODE_MODULES_DIR := control/node_modules
CONTROL_PLAYWRIGHT_BIN := $(CONTROL_NODE_MODULES_DIR)/.bin/playwright
WEBVUE_NODE_MODULES_DIR := webvue/node_modules
WEBVUE_VITEST_BIN := $(WEBVUE_NODE_MODULES_DIR)/.bin/vitest
WEBVUE_VITE_BIN := $(WEBVUE_NODE_MODULES_DIR)/.bin/vite
WEBVUE_NODE_MODULES_STAMP := $(WEBVUE_NODE_MODULES_DIR)/.package-lock.stamp

PSQL ?= psql
CONTROL_PG_DIR := control/admin/postgres
PG_BOOTSTRAP_DSN ?= postgres://jaten:jaten@127.0.0.1:5432/postgres?sslmode=disable
IVYVUE_BOOTSTRAP_DSN ?= postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable
IVYVUE_MIGRATION_DSN ?= postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable

bootstrap:
	@command -v $(PSQL) >/dev/null || { echo "psql not found; install PostgreSQL client tools first"; exit 1; }
	$(PSQL) '$(PG_BOOTSTRAP_DSN)' -v ON_ERROR_STOP=1 -f $(CONTROL_PG_DIR)/00_bootstrap_databases.sql
	$(PSQL) '$(IVYVUE_BOOTSTRAP_DSN)' -v ON_ERROR_STOP=1 -f $(CONTROL_PG_DIR)/01_ivyvue_roles_and_schemas.sql
	$(PSQL) '$(IVYVUE_MIGRATION_DSN)' -v ON_ERROR_STOP=1 -f $(CONTROL_PG_DIR)/02_control_schema.sql
	$(PSQL) '$(IVYVUE_MIGRATION_DSN)' -v ON_ERROR_STOP=1 -f $(CONTROL_PG_DIR)/03_rls_template.sql
	$(PSQL) '$(IVYVUE_BOOTSTRAP_DSN)' -v ON_ERROR_STOP=1 -f $(CONTROL_PG_DIR)/04_verify_bootstrap.sql
	@echo "bootstrap complete: ivyvue PostgreSQL database is ready for local servers/tests"

tr: # xtracer builds:
	cd cmd/goivy_check && go build -o ${GOBIN}/goivy_check_xtrace
	cd cmd/ivyweb && go build -o ${GOBIN}/ivyweb_xtrace

trt: build tr  # xtracer test
	XTRACE_OFF=1 time goivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy &> log.go.trt.ord_live.standalone

dylib_build: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY_ABS)/lib:$$DYLD_LIBRARY_PATH go install ./cmd/ivyweb

dylib_test: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY_ABS)/lib:$$DYLD_LIBRARY_PATH go test ./... -count=1 # -tags web

test:
	XTRACE_OFF=1 go test ./... -count=1 # -tags web

$(WEBUI_NODE_MODULES_STAMP): webui/package.json webui/package-lock.json
	cd webui && npm ci
	touch $(WEBUI_NODE_MODULES_STAMP)

$(VITEST_BIN) $(PLAYWRIGHT_BIN): $(WEBUI_NODE_MODULES_STAMP)

vitest-setup: $(VITEST_BIN) $(PLAYWRIGHT_BIN)

$(CONTROL_PLAYWRIGHT_BIN): control/package.json control/package-lock.json
	cd control && npm ci

control-browser-setup: $(CONTROL_PLAYWRIGHT_BIN)

$(WEBVUE_NODE_MODULES_STAMP): webvue/package.json webvue/package-lock.json
	cd webvue && npm ci
	touch $(WEBVUE_NODE_MODULES_STAMP)

$(WEBVUE_VITEST_BIN) $(WEBVUE_VITE_BIN): $(WEBVUE_NODE_MODULES_STAMP)

webvue-setup: $(WEBVUE_VITEST_BIN) $(WEBVUE_VITE_BIN)

test-vue: $(WEBVUE_VITEST_BIN)
	cd webvue && ./node_modules/.bin/vitest run --config vitest.config.mjs

build-vue: $(WEBVUE_VITE_BIN)
	cd webvue && ./node_modules/.bin/vite build --config vite.config.mjs

test-js: $(VITEST_BIN)
	cd webui && npm run test:webui:js

test-js-watch: $(VITEST_BIN)
	cd webui && npm run test:webui:js:watch

test-browser: $(PLAYWRIGHT_BIN)
	cd webui && npm run test:webui:browser

test-control-browser: $(CONTROL_PLAYWRIGHT_BIN)
	cd control && npm run test:browser

dylib_test-web: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY_ABS)/lib:$$DYLD_LIBRARY_PATH go test -v ./webui -count=1 -tags web

test-web:
	XTRACE_OFF=1 go test -v ./webui -count=1 -tags xtrace_off,web

unimerge-audit:
	go run ./cmd/unimerge_audit -details -max 120

unimerge-audit-strict:
	go run ./cmd/unimerge_audit -strict

# Run conformance tests (requires Python Ivy + Z3 sidecar).
test-conform: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY_ABS)/lib:$$DYLD_LIBRARY_PATH go test ./webui/ -run TestConform -v -count=1 -timeout 120s

clean:
	rm -rf $(Z3IVY)
	go clean ./...

fuzz:
	cd z3bridge && for fuzz in FuzzQuantConstraintsForAll \
        FuzzVariableNaming FuzzSortLookup FuzzTranslator; do \
		echo "\n=== $$fuzz ==="; \
		/usr/local/bin/go test -run=xxx -fuzz=$$fuzz -fuzztime=30s; \
	done
	cd solver && for fuzz in FuzzMyEq FuzzGebin FuzzBinEncZ3 FuzzSortCard \
	FuzzRangeSortBounds FuzzNumeralToZ3Clamping FuzzEncodeEqualityZ3 \
	FuzzSolverNameBuiltins FuzzQuantConstraintsNatRange; do \
		echo "\n=== $$fuzz ==="; \
		/usr/local/bin/go test -run=xxx -fuzz=$$fuzz -fuzztime=30s; \
	done

vs:
	cd cmd/goivy_check && go build -o ${GOBIN}/goivy_check_xtrace
	if [ -f log.red ]; then mv log.red log.red.prev; fi
	-go test -v -run TestOrdLive &> log.red
	-/Users/jaten/ivy/pyivy/goivy-venv/bin/ivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy 2>&1 | awk 'NR <= 300000 { print } NR == 300000 { exit }' &> py.ord.out
	-goivy_check_xtrace  isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy &> log.gox
	-diff py.ord.out log.gox > diffs

setup: # install abc and aigtoaig binaries that we shell out to.
	cd ../pyivy && ./onetime.sh

cleardb:
	psql -c "drop database ivyvue"
	make bootstrap
