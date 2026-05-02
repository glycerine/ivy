# Plan: WebUI ShowVerification Entry Point

**Created:** 2026-05-02 15:38 UTC

## Context

`ShowVerification()` in `webui/ui_show.go:15-33` is the entry point that connects "load an .ivy file from disk" to "show its verification state in the web UI." It corresponds to Python's `ivy_show.py:main()` + `check_module()`.

**The bug:** `ShowVerification` calls `sess.LoadFile(filePath)` which only stores the path string as metadata — it never reads, parses, or compiles the file. The returned session has no `CompiledModule`, no `AnalysisGraph`, no concept domain. The web UI gets an empty shell.

**The fix is wiring:** `session.LoadFileContent(filename, content)` already implements the full pipeline (parse → compile → isolate → sorts/symbols → concept sessions → AnalysisGraph). We just need to read the file from disk and feed it through `LoadFileContent`.

Additionally, `CheckModuleAndShow()` creates a `Server` but discards it (`_ = srv`) and never calls `srv.Start()`, so the server never listens.

## Changes

### 1. Fix `ShowVerification` — `webui/ui_show.go:6-33`

Add `"os"` to imports. Replace the function body to:
1. Read the file from disk with `os.ReadFile(filePath)`
2. Set `cfg.ShowCompiled = true` (matches Python `iu.set_parameters({'show_compiled':'true'})` at `ivy_show.py:40`)
3. Call `sess.LoadFileContent(filePath, content)` to run the full compiler pipeline

The Python two-step pattern (compile with `create_isolate=False`, then `create_isolate(isolate)` on a module copy) is used by `ivy_check` which iterates over multiple isolates. The web UI entry point only needs the default isolate, and `LoadFileContent` already calls `IvyCompile(decls, mod, true)` which handles this. If `cfg.Isolate` is set before calling, the compiler respects it.

**Reuses:** `session.LoadFileContent()` at `webui/session.go:82-218` (full pipeline), `module.Config.ShowCompiled` at `module/config.go:211`.

### 2. Fix `CheckModuleAndShow` — `webui/ui_show.go:69-83`

Replace `_ = srv; return nil` with `return srv.Start()`. `srv.Start()` calls `http.ListenAndServe()` which blocks, matching Python's `ui_main_loop()`.

### 3. Comprehensive unit tests — `webui/ui_test.go`

Replace the existing `TestShowVerification` (lines 688-702, which passes a nonexistent "test.ivy" path) with these tests:

| Test | What it verifies |
|------|-----------------|
| `TestShowVerificationEmptyPath` | Empty string returns error |
| `TestShowVerificationNonexistentFile` | Missing file returns "failed to read file" error |
| `TestShowVerificationInvalidContent` | Unparseable content returns "failed to compile file" error |
| `TestShowVerificationSuccess` | Full pipeline: writes `ivySample` to temp file, calls `ShowVerification`, asserts `CompiledModule`, `CompiledSig`, `AG`, `AGUI`, `ConceptSess`, `SimpleSess` are all non-nil; verifies `cfg.ShowCompiled == true` |
| `TestShowVerificationFromTestVectors` | Uses real `../test_vectors/bmc_minimal.ivy` file |
| `TestShowVerificationNoHeader` | File without `#lang ivy` header still compiles |
| `TestCheckModuleAndShowComponents` | ShowVerification + LaunchUI produce working Server |

All tests use `//go:build web` (already present in file header). Tests reference `ivySample` from `backend_conform_test.go:17-26`.

Add `"os"` and `"path/filepath"` to test file imports.

## Verification

Run: `cd ~/ivy/goivy && make dylib_test-web`

Or targeted: `cd ~/ivy/goivy && DYLD_LIBRARY_PATH=z3ivy/lib:$DYLD_LIBRARY_PATH go test -v ./webui -count=1 -tags web -run TestShowVerification`

All new `TestShowVerification*` tests must pass. Existing `TestLaunchUI` must still pass.

## Files to modify

- `webui/ui_show.go` — fix `ShowVerification` and `CheckModuleAndShow`
- `webui/ui_test.go` — replace and expand test suite
