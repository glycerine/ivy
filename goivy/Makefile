.PHONY: all build test test-conform clean z3ivy z3-build tr trt vs golden

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

# for bashisms like &>
SHELL := /bin/bash

golden: # parse with python ivy_check and goivy_check_xtrace. full xtrace output is written now to the files: out.go.xtrace; out.py.xtrace
	@if [ -f log.red ]; then mv log.red log.red.prev; fi
	cd lalr_full && go test -v -run TestOrdLive &> ../log.red

golden-v: # to see all productions not just last 10
	cd lalr_full && go test -v -run TestVerboseOrdLive

golden-nonstop: # to see all productions and not stop at first diff.
	cd lalr_full && go test -v -run TestVerboseNonstopOrdLive


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

build: # be sure .y -> .go up to date
	cd lalr_full && go generate # grammar_v17.y
	cd conceptspace && go generate # cs_grammar.y
	cd lalr_logicparser/v12 && go generate # grammar_v12.y
	cd lalr_logicparser/v16 && go generate # grammar_v16.y
	cd lalr_logicparser && go generate # grammar_v17.y
	cd evparser && go generate # ev_grammar.y
	go install  -tags xtracer_off ./cmd/ivyweb
	go install  -tags xtracer_off ./cmd/goivy_check

GOBIN ?= $(HOME)/go/bin

tr: # xtracer builds:
	cd cmd/goivy_check && go build -o ${GOBIN}/goivy_check_xtrace
	cd cmd/ivyweb && go build -o ${GOBIN}/ivyweb_xtrace

trt: tr  # xtracer test
	goivy_check_xtrace isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

dylib_build: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY_ABS)/lib:$$DYLD_LIBRARY_PATH go install ./cmd/ivyweb

dylib_test: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY_ABS)/lib:$$DYLD_LIBRARY_PATH go test ./... -count=1 # -tags web

test:
	XTRACE_OFF=1 go test ./... -count=1 # -tags web

dylib_test-web: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY_ABS)/lib:$$DYLD_LIBRARY_PATH go test -v ./webui -count=1 -tags web

test-web:
	XTRACE_OFF=1 go test -v ./webui -count=1 -tags xtrace_off,web

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

pprof: # Quick profile — sub-operations breakdown
	go test -tags web -bench BenchmarkRodSub -benchtime 5s -count 1 ./webui/

pprof-full:  # Full profile with CPU/mem profiling
	go test -tags web -bench BenchmarkRod -benchtime 10s -cpuprofile cpu.prof \
        -memprofile mem.prof ./webui/

# Then inspect
# go tool pprof cpu.prof

vs:
	cd cmd/goivy_check && go build -o ${GOBIN}/goivy_check_xtrace
	if [ -f log.red ]; then mv log.red log.red.prev; fi
	-cd lalr_full && go test -v -run TestOrdLive &> ../log.red
	-/Users/jaten/ivy/pyivy/goivy-venv/bin/ivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy 2>&1 | awk 'NR <= 300000 { print } NR == 300000 { exit }' &> py.ord.out
	-goivy_check_xtrace  isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy &> log.gox
	-diff py.ord.out log.gox > diffs

tags:
	find . -name "*.go" > .gtags.files
	gtags -f .gtags.files
