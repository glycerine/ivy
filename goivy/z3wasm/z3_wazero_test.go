package z3wasm

// See also: goivy/cmd/whatwasm full a more built out and thorough description
// of wasm binaries. It is based on this kind of check, but goes further.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

func TestZ3WasmArtifactCompilesWithWazeroExceptionHandling(t *testing.T) {
	wasmPath := filepath.Join("..", "webui", "static", "wasm", "z3-471-api.wasm")
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read Z3 wasm artifact: %v", err)
	}

	ctx := context.Background()
	runtimeConfig := wazero.NewRuntimeConfigInterpreter().
		WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesExceptionHandling)
	runtime := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)
	defer runtime.Close(ctx)

	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		t.Fatalf("compile Z3 wasm with wazero exception handling: %v", err)
	}
	defer compiled.Close(ctx)

	var jsExceptionImports []string
	for _, fn := range compiled.ImportedFunctions() {
		moduleName, importName, ok := fn.Import()
		if !ok || moduleName != "env" {
			continue
		}
		switch {
		case importName == "__cxa_throw":
			jsExceptionImports = append(jsExceptionImports, importName)
		case importName == "__resumeException":
			jsExceptionImports = append(jsExceptionImports, importName)
		case strings.HasPrefix(importName, "__cxa_find_matching_catch"):
			jsExceptionImports = append(jsExceptionImports, importName)
		}
	}
	if len(jsExceptionImports) > 0 {
		t.Fatalf("Z3 wasm still imports JS-mediated C++ exception helpers: %v", jsExceptionImports)
	}
}
