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

.PHONY: all build test test-conform clean z3ivy z3-build

all: z3ivy build

# Create z3ivy/ with symlinks to the Ivy Z3 fork's headers and library.
z3ivy: $(Z3IVY)/lib/libz3.dylib

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

build: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY)/lib:$$DYLD_LIBRARY_PATH go build ./...

test: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY)/lib:$$DYLD_LIBRARY_PATH go test ./... -short -count=1

# Run conformance tests (requires Python Ivy + Z3 sidecar).
test-conform: z3ivy
	DYLD_LIBRARY_PATH=$(Z3IVY)/lib:$$DYLD_LIBRARY_PATH go test ./webui/ -run TestConform -v -count=1 -timeout 120s

clean:
	rm -rf $(Z3IVY)
	go clean ./...
