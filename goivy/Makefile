# Makefile for goivy (Go port of Ivy)
#
# NOTE: keep all targets and paths explicit. Do not abstract 
# into variables. That hurts read-ability and navigation.
#
# The Go Z3 bridge links against the same Z3 4.7.1 fork that Python Ivy uses.
WAZ_MEM_GC_HEAP_LIMIT ?=
WAZ_CLI_MEM_GC_HEAP_LIMIT ?=
# Z3 4.7.1's generator infers _AMD64_ from the host machine, but Emscripten
# targets wasm32 here. Native builds must keep their normal 64-bit settings.
Z3_WASM_CC_FLAGS    := -Wno-deprecated-declarations -fwasm-exceptions -U_AMD64_

all: build

# for bashisms like &>
SHELL := /bin/bash

golden: # parse with python ivy_check and goivy_check_xtrace. full xtrace output is written now to the files: out.go.xtrace; out.py.xtrace
	if [ -f log.red ]; then mv log.red log.red.prev; fi
	go test -v -timeout=0 -count=1 -run TestOrdLive &> log.red

golden-2hr: # takes 2 hours to check all isolates.
	if [ -f log.golden.2hr ]; then mv log.golden.2hr log.golden.2hr.prev; fi
	go test -v -timeout=0 -count=1 -run Test2hrOrdLive &> log.golden.2hr

node-golden-2hr: # takes 2 hours to check all isolates.
	if [ -f log.node.golden.2hr ]; then mv log.node.golden.2hr log.node.golden.2hr.prev; fi
	go test -v -timeout=0 -count=1 -run Test2hrNodeGoldenOrdLive &> log.node.golden.2hr

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
	~/ivy/pyivy/goivy-venv/bin/ivy_check isolate=cf_live $(HOME)/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

ordo: # not xtraced
	time python3 -O ~/ivy/pyivy/goivy-venv/bin/ivy_check isolate=cf_live $(HOME)/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy



# Create z3ivy/ with symlinks to the Ivy Z3 fork's headers and library.
#z3ivy: z3ivy/lib/libz3.dylib

get_goyacc:
	go install golang.org/x/tools/cmd/goyacc@latest

# static linked to z3vendor now, for completeness within this repo.
z3ivy: z3vendor/native_lib/libz3.a #z3vendor/z3/build/libz3.a
	echo "must manually build static libz3.a using z3vendor/z3/compile.sh and z3vendor/z3/compile-wasm.sh"

z3ivy/lib/libz3.dylib: $(HOME)/pyivy/ivy/submodules/z3/build/libz3.dylib
	@mkdir -p z3ivy/include z3ivy/lib
	@for h in z3.h z3_api.h z3_macros.h z3_v1.h z3_algebraic.h \
		z3_ast_containers.h z3_fixedpoint.h z3_fpa.h z3_interp.h \
		z3_optimization.h z3_polynomial.h z3_rcf.h z3_spacer.h; do \
		ln -sf $(HOME)/pyivy/ivy/submodules/z3/src/api/$$h z3ivy/include/$$h; \
	done
	@ln -sf $(HOME)/pyivy/ivy/submodules/z3/build/libz3.dylib z3ivy/lib/libz3.dylib
	@echo "z3ivy: linked to Z3 4.7.1 fork at $(HOME)/pyivy/ivy/submodules/z3"

# Build the Z3 fork from the submodule (only needed if not already built).
z3-build:
	@if [ ! -f $(HOME)/pyivy/ivy/submodules/z3/build/libz3.dylib ]; then \
		echo "Building Z3 fork from $(HOME)/pyivy/ivy/submodules/z3..."; \
		cd $(HOME)/pyivy/ivy/submodules/z3 && \
                python3 scripts/mk_make.py --python --prefix=$(HOME)/pyivy/ivy; \
		cd $(HOME)/pyivy/ivy/submodules/z3/build && $(MAKE) -j4; \
	else \
		echo "Z3 fork already built at $(HOME)/pyivy/ivy/submodules/z3/build/libz3.dylib"; \
	fi

z3vendor/z3/src/api/js/exported-functions.json: z3vendor/z3/src/api/js/scripts/get-exports.ts z3vendor/z3/src/api/js/scripts/parse-api.ts z3vendor/z3/src/api/z3_api.h z3vendor/z3/src/api/z3_algebraic.h z3vendor/z3/src/api/z3_ast_containers.h z3vendor/z3/src/api/z3_fixedpoint.h z3vendor/z3/src/api/z3_fpa.h z3vendor/z3/src/api/z3_optimization.h z3vendor/z3/src/api/z3_polynomial.h z3vendor/z3/src/api/z3_rcf.h z3vendor/z3/src/api/z3_spacer.h z3vendor/z3/src/api/z3_interp.h
	cd z3vendor/z3/src/api/js && ./node_modules/.bin/ts-node --transpileOnly scripts/get-exports.ts > exported-functions.json

webui/static/wasm/z3-wasm-exported-functions.json: z3vendor/z3/src/api/js/exported-functions.json
	cp -p z3vendor/z3/src/api/js/exported-functions.json webui/static/wasm/z3-wasm-exported-functions.json

z3-wasm-exports: webui/static/wasm/z3-wasm-exported-functions.json

# builds: libz3.a(wasm despite the .a re-use); and in turn,
# the build-wasm-api.sh script turns that into: z3-api.wasm and z3-api.js
z3vendor/z3/wasm.build/libz3.a:
	cd z3vendor/z3 && \
		EMSDK_QUIET=1 source "$(HOME)/go/src/github.com/emscripten-core/emsdk/emsdk_env.sh" && \
		EM_CACHE="/tmp/ivy-emscripten-cache" AR=emar CXX=em++ CC=emcc \
		CXXFLAGS="$(Z3_WASM_CC_FLAGS)" \
		CFLAGS="$(Z3_WASM_CC_FLAGS)" \
		python3 scripts/mk_make.py --staticlib --build=wasm.build
	cd z3vendor/z3/wasm.build && \
		EMSDK_QUIET=1 source "$(HOME)/go/src/github.com/emscripten-core/emsdk/emsdk_env.sh" && \
		EM_CACHE="/tmp/ivy-emscripten-cache" emmake $(MAKE) -j8 libz3.a

z3-wasm-lib-wasm-eh: z3vendor/z3/wasm.build/libz3.a

z3vendor/z3/wasm.build/.z3-api-wasm-eh.stamp: z3vendor/z3/wasm.build/libz3.a webui/static/wasm/z3-wasm-exported-functions.json z3vendor/z3/build-wasm-api.sh z3vendor/z3/src/api/js/ivy_z3_wasm_syscalls.cc
	cd z3vendor/z3 && \
		EMSDK_QUIET=1 source "$(HOME)/go/src/github.com/emscripten-core/emsdk/emsdk_env.sh" && \
		EM_CACHE="/tmp/ivy-emscripten-cache" EMCC="$(HOME)/go/src/github.com/emscripten-core/emsdk/upstream/emscripten/emcc" ./build-wasm-api.sh
	touch z3vendor/z3/wasm.build/.z3-api-wasm-eh.stamp

z3-wasm-api-link: z3vendor/z3/wasm.build/.z3-api-wasm-eh.stamp

z3vendor/z3/wasm.build/z3-api.exnref.wasm: z3vendor/z3/wasm.build/.z3-api-wasm-eh.stamp
	cp -p z3vendor/z3/wasm.build/z3-api.wasm z3vendor/z3/wasm.build/z3-api.pre-exnref.wasm
	/usr/local/bin/wasm-opt --translate-to-exnref \
		--enable-exception-handling \
		--enable-reference-types \
		--enable-bulk-memory \
		--enable-bulk-memory-opt \
		--enable-multivalue \
		--enable-mutable-globals \
		--enable-sign-ext \
		--enable-nontrapping-float-to-int \
		-o z3vendor/z3/wasm.build/z3-api.exnref.wasm z3vendor/z3/wasm.build/z3-api.pre-exnref.wasm

webui/static/wasm/z3-471-api.wasm: z3vendor/z3/wasm.build/z3-api.exnref.wasm
	cp -p z3vendor/z3/wasm.build/z3-api.exnref.wasm webui/static/wasm/z3-471-api.wasm
	chmod 0644 webui/static/wasm/z3-471-api.wasm

webui/static/wasm/z3-471-api.js: z3vendor/z3/wasm.build/.z3-api-wasm-eh.stamp
	cp -p z3vendor/z3/wasm.build/z3-api.js webui/static/wasm/z3-471-api.js
	chmod 0644 webui/static/wasm/z3-471-api.js

z3-wasm-api: webui/static/wasm/z3-471-api.wasm webui/static/wasm/z3-471-api.js

z3-wasm-api-wazero: z3-wasm-api

webui/static/wasm/wasm_exec-go1.25.6.js: /usr/local/go/lib/wasm/wasm_exec.js
	cp -p /usr/local/go/lib/wasm/wasm_exec.js webui/static/wasm/wasm_exec-go1.25.6.js

vendor-go-wasm-exec-js: webui/static/wasm/wasm_exec-go1.25.6.js

# webui/static/wasm/z3-471-api.wasm: 
# 	mkdir -p webui/static/wasm
# 	cp -p webui/static/wasm/z3-471-api.wasm webui/static/wasm/z3-471-api.wasm
# 	chmod 0644 webui/static/wasm/z3-471-api.wasm

# webui/static/wasm/z3-471-api.js: webui/static/wasm/z3-471-api.js
# 	mkdir -p webui/static/wasm
# 	cp -p webui/static/wasm/z3-471-api.js webui/static/wasm/z3-471-api.js
# 	chmod 0644 webui/static/wasm/z3-471-api.js

# webui/static/wasm/wasm_exec-go1.25.6.js: webui/static/wasm/wasm_exec-go1.25.6.js
# 	mkdir -p webui/static/wasm
# 	cp -p webui/static/wasm/wasm_exec-go1.25.6.js webui/static/wasm/wasm_exec-go1.25.6.js
# 	chmod 0644 webui/static/wasm/wasm_exec-go1.25.6.js

webui/static/wasm/include-tree.json: webui/scripts/build-include-tree.mjs $(shell find ../ivy-lang-examples/ivy/include -type f -name '*.ivy' 2>/dev/null)
	node webui/scripts/build-include-tree.mjs ../ivy-lang-examples/ivy/include webui/static/wasm/include-tree.json
	chmod 0644 webui/static/wasm/include-tree.json

webui/static/wasm/goivy-webengine.wasm: FORCE
	mkdir -p webui/static/wasm
	GOOS=js GOARCH=wasm go build -tags xtracer_off -o webui/static/wasm/goivy-webengine.wasm ./cmd/goivy-webengine-wasm
	chmod 0644 webui/static/wasm/goivy-webengine.wasm

webui-wasm-engine: webui/static/wasm/goivy-webengine.wasm webui/static/wasm/z3-471-api.wasm webui/static/wasm/z3-471-api.js webui/static/wasm/wasm_exec-go1.25.6.js webui/static/wasm/include-tree.json

test-z3-wasm-wazero: z3-wasm-api-wazero
	go test ./z3wasm -run TestZ3WasmArtifactCompilesWithWazeroExceptionHandling -count=1


FORCE:

webui/static/wasm/goivy-check-js.wasm: FORCE
	GOOS=js GOARCH=wasm go build -a -tags xtracer_off -o webui/static/wasm/goivy-check-js.wasm ./cmd/goivy_check_jswasm

goivy-check-js-wasm: webui/static/wasm/goivy-check-js.wasm

build: # grammer_gen
	go install  -tags xtracer_off ./cmd/ivyweb
	go install  -tags xtracer_off ./cmd/goivy_check
	go install  -tags xtracer_off ./cmd/whatwasm

grammar_gen: # be sure .y -> .go up to date
	goyacc -o parser_grammar_v17.go -p parser17 parser_grammar_v17.y
	go run parser_patch_parser.go
	cd conceptspace && go generate # cs_grammar.y
	goyacc -o v12_grammar_v12.go -p lalr12 v12_grammar_v12.y
	goyacc -o v16_grammar_v16.go -p lalr16 v16_grammar_v16.y
	goyacc -o lalr_logicparser_grammar_v17.go -p lalr17 lalr_logicparser_grammar_v17.y
	cd evparser && go generate # ev_grammar.y


PSQL ?= psql
PG_BOOTSTRAP_DSN ?= postgres://jaten:jaten@127.0.0.1:5432/postgres?sslmode=disable
IVYVUE_BOOTSTRAP_DSN ?= postgres://jaten:jaten@127.0.0.1:5432/ivyvue?sslmode=disable
IVYVUE_MIGRATION_DSN ?= postgres://ivyvue_migrator:ivyvue_migrator_dev@127.0.0.1:5432/ivyvue?sslmode=disable

bootstrap:
	@command -v $(PSQL) >/dev/null || { echo "psql not found; install PostgreSQL client tools first"; exit 1; }
	$(PSQL) '$(PG_BOOTSTRAP_DSN)' -v ON_ERROR_STOP=1 -f control/admin/postgres/00_bootstrap_databases.sql
	$(PSQL) '$(IVYVUE_BOOTSTRAP_DSN)' -v ON_ERROR_STOP=1 -f control/admin/postgres/01_ivyvue_roles_and_schemas.sql
	$(PSQL) '$(IVYVUE_MIGRATION_DSN)' -v ON_ERROR_STOP=1 -f control/admin/postgres/02_control_schema.sql
	$(PSQL) '$(IVYVUE_MIGRATION_DSN)' -v ON_ERROR_STOP=1 -f control/admin/postgres/03_rls_template.sql
	$(PSQL) '$(IVYVUE_BOOTSTRAP_DSN)' -v ON_ERROR_STOP=1 -f control/admin/postgres/04_verify_bootstrap.sql
	@echo "bootstrap complete: ivyvue PostgreSQL database is ready for local servers/tests"

tr: # xtracer builds:
	cd cmd/goivy_check && go build -o $(HOME)/go/bin/goivy_check_xtrace
	cd cmd/ivyweb && go build -o $(HOME)/go/bin/ivyweb_xtrace

trt: build tr  # xtracer test
	XTRACE_OFF=1 time goivy_check isolate=cf_live $(HOME)/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy &> log.go.trt.ord_live.standalone

dylib_build: z3ivy
	DYLD_LIBRARY_PATH=$(CURDIR)/z3ivy/lib:$$DYLD_LIBRARY_PATH go install ./cmd/ivyweb

dylib_test: z3ivy
	DYLD_LIBRARY_PATH=$(CURDIR)/z3ivy/lib:$$DYLD_LIBRARY_PATH go test ./... -count=1 # -tags web

test:
	XTRACE_OFF=1 go test ./... -count=1 -tags xtracer_off

WEBUI_NODE_PLATFORM := $(shell node -e 'const h = process.report && process.report.getReport ? process.report.getReport().header : {}; let libc = ""; if (process.platform === "linux") libc = h.glibcVersionRuntime ? "gnu" : "musl"; console.log([process.platform, process.arch, libc].filter(Boolean).join("-"));' 2>/dev/null || printf '%s-%s\n' "$$(uname -s)" "$$(uname -m)")
WEBUI_NODE_DEPS_STAMP := webui/node_modules/.platform-deps-$(WEBUI_NODE_PLATFORM).stamp
WEBUI_NODE_DEPS_CHECK := node --input-type=module -e "await import('rollup'); const graphviz = await import('@hpcc-js/wasm-graphviz'); await graphviz.Graphviz.load();"

webui/node_modules/.package-lock.stamp: webui/package.json webui/package-lock.json
	cd webui && npm ci
	touch webui/node_modules/.package-lock.stamp

# @hpcc-js/wasm-graphviz ships prebuilt wasm through npm. Rollup also ships
# platform-specific optional native packages, so a reused node_modules tree can
# satisfy package-lock but still miss the current platform's runtime package.
$(WEBUI_NODE_DEPS_STAMP): webui/package.json webui/package-lock.json webui/node_modules/.package-lock.stamp
	cd webui && ($(WEBUI_NODE_DEPS_CHECK) || (npm ci && $(WEBUI_NODE_DEPS_CHECK)))
	touch webui/node_modules/.package-lock.stamp
	touch $(WEBUI_NODE_DEPS_STAMP)

webui-node-deps webui-wasm-graphviz: $(WEBUI_NODE_DEPS_STAMP)

webui/node_modules/.bin/vitest webui/node_modules/.bin/playwright: webui/node_modules/.package-lock.stamp $(WEBUI_NODE_DEPS_STAMP)

vitest-setup: webui/node_modules/.bin/vitest webui/node_modules/.bin/playwright $(WEBUI_NODE_DEPS_STAMP)

control/node_modules/.bin/playwright: control/package.json control/package-lock.json
	cd control && npm ci

control-browser-setup: control/node_modules/.bin/playwright

test-js: webui/node_modules/.bin/vitest $(WEBUI_NODE_DEPS_STAMP)
	cd webui && npm run test:webui:js

test-js-watch: webui/node_modules/.bin/vitest $(WEBUI_NODE_DEPS_STAMP)
	cd webui && npm run test:webui:js:watch

test-browser: webui/node_modules/.bin/playwright $(WEBUI_NODE_DEPS_STAMP)
	cd webui && npm run test:webui:browser

test-control-browser: control/node_modules/.bin/playwright
	cd control && npm run test:browser

dylib_test-web: z3ivy
	DYLD_LIBRARY_PATH=$(CURDIR)/z3ivy/lib:$$DYLD_LIBRARY_PATH go test -v ./webui -count=1 -tags web

test-web: $(WEBUI_NODE_DEPS_STAMP)
	XTRACE_OFF=1 go test -v ./webui -count=1 -tags xtracer_off,web
	cd webui && npm --prefix frontend run test

unimerge-audit:
	go run ./cmd/unimerge_audit -details -max 120

unimerge-audit-strict:
	go run ./cmd/unimerge_audit -strict

# Run conformance tests (requires Python Ivy + Z3 sidecar).
test-conform: z3ivy
	DYLD_LIBRARY_PATH=$(CURDIR)/z3ivy/lib:$$DYLD_LIBRARY_PATH go test ./webui/ -run TestConform -v -count=1 -timeout 120s

clean:
	rm -rf z3ivy
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
	cd cmd/goivy_check && go build -o $(HOME)/go/bin/goivy_check_xtrace
	if [ -f log.red ]; then mv log.red log.red.prev; fi
	-go test -v -run TestOrdLive &> log.red
	-$(HOME)/ivy/pyivy/goivy-venv/bin/ivy_check isolate=cf_live $(HOME)/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy 2>&1 | awk 'NR <= 300000 { print } NR == 300000 { exit }' &> py.ord.out
	-goivy_check_xtrace  isolate=cf_live $(HOME)/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy &> log.gox
	-diff py.ord.out log.gox > diffs

setup: # install abc and aigtoaig binaries that we shell out to.
	cd ../pyivy && ./onetime.sh

cleardb:
	psql -c "drop database IF EXISTS ivyvue"
	make bootstrap

goldweb: goivy-check-js-wasm vendor-go-wasm-exec-js
	if [ -f py.xtrace.log ]; then mv py.xtrace.log py.xtrace.log.prev; fi
	if [ -f browser.xtrace.log ]; then mv browser.xtrace.log browser.xtrace.log.prev; fi
	go run ./goldweb -goivy-runtime js -goivy-wasm $(CURDIR)/webui/static/wasm/goivy-check-js.wasm ~/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy &> log.goldweb

nodegold: goivy-check-js-wasm vendor-go-wasm-exec-js
	time go run ./nodegold -goivy-wasm $(CURDIR)/webui/static/wasm/goivy-check-js.wasm ~/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy # 2>&1 | zstd -c &> log.nodegold.zst

nodegold2: goivy-check-js-wasm vendor-go-wasm-exec-js
	go install ./nodegold
	time ~/go/bin/nodegold -goivy-wasm $(CURDIR)/webui/static/wasm/goivy-check-js.wasm ~/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy # 2>&1 | zstd -c &> log.nodegold.zst

# wasm-opt is from the binaryen project.
# we need a recent build for the flag we use,
# wasm-opt --translate-to-exnref
# and ubuntu 24.04's binaryen package is an old
# version 108. It lacks the flag. Hence a source build:
# (but homebrew on macOS did have version 129 which
# souce is just a little after anyway).
install-wasm-opt:
	cd ~/go/src/github.com/ && git clone https://github.com/WebAssembly/binaryen && cd binaryen && git submodule init && git submodule update && cmake . && make -j8 && sudo make install

prod: webui-wasm-engine
	cd webui && npm run typecheck:webui # run tsc compiler
	cd webui && npm --prefix frontend run typecheck # type check the frontend/ dir
	cd webui/frontend && npm run build
	cd webui && npm run build:webui     # Vite bundle out: webui/static/dist/ivyweb.js
	go install -tags xtracer_off ./cmd/ivyweb
	ivyweb

.PHONY: all build test test-conform test-js test-js-watch test-browser test-control-browser test-web test-vue build-vue vitest-setup control-browser-setup webui-node-deps webui-wasm-graphviz webui-wasm-engine unimerge-audit unimerge-audit-strict clean z3ivy z3-build z3-wasm-exports z3-wasm-lib-wasm-eh z3-wasm-api-link z3-wasm-api z3-wasm-api-wazero vendor-go-wasm-exec-js goivy-check-js-wasm test-z3-wasm-wazero test-z3-wasm-js test-z3-wasm-browser tr trt vs golden setup this bootstrap goldweb nodegold FORCE prod
