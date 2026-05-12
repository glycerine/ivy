//go:build !js

package static_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	vendoredWasmExecVersion = "go1.25.6" // stayed the same for go1.26
	vendoredWasmExecName    = "wasm_exec-" + vendoredWasmExecVersion + ".js"
)

// note: this is a "Big" Go (standard toolchain) test.
func TestVendoredWasmExecMatchesGoToolchain(t *testing.T) {
	got := runtime.Version()
	if strings.HasPrefix(got, "go1.25") || strings.HasPrefix(got, "go1.26") {
		// okay
	} else {
		t.Fatalf("vendored %s is for %s (go1.25.x and go1.26.x support), but the current Go toolchain is %s; refresh /usr/local/go/lib/wasm/wasm_exec.js and rename the vendored file", vendoredWasmExecName, vendoredWasmExecVersion, got)
	}

	sourcePath := filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read Go toolchain wasm_exec.js at %s: %v", sourcePath, err)
	}

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed while locating vendored wasm_exec.js")
	}
	vendoredPath := filepath.Join(filepath.Dir(testFile), vendoredWasmExecName)
	vendored, err := os.ReadFile(vendoredPath)
	if err != nil {
		t.Fatalf("read vendored wasm_exec.js at %s: %v", vendoredPath, err)
	}

	if len(vendored) != len(source) {
		t.Fatalf("vendored %s differs from %s: length %d, want %d; refresh with `make vendor-go-wasm-exec-js`", vendoredPath, sourcePath, len(vendored), len(source))
	}
	for i := range source {
		if vendored[i] != source[i] {
			t.Fatalf("vendored %s differs from %s at byte %d: vendored=0x%02x toolchain=0x%02x; refresh with `make vendor-go-wasm-exec-js`", vendoredPath, sourcePath, i, vendored[i], source[i])
		}
	}
}
