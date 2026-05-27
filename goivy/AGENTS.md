# AGENTS.md instructions for /Users/jaten/go/src/github.com/glycerine/ivy/goivy

## Universal behavior

- The parent repository instructions still apply here. In particular, do not use
  git commands; the user owns git activity for this repository.

## Ivy language-version parser invariants

- Treat the full-file Ivy parsers as thin syntax adapters. A parser for Ivy
  1.6, 1.7, or 1.8 should recognize version-specific syntax and then hand off
  to shared Go helpers for AST construction, declarations, include/import
  handling, module/object processing, action construction, proof construction,
  and post-parse normalization.
- Keep post-parsing behavior shared unless Python Ivy has a real semantic
  difference for that language version. Do not duplicate accumulator logic,
  include logic, module instantiation, object expansion, action lowering,
  compiler passes, isolate passes, AST node types, or code-generation contracts
  just because a second grammar needs to call them.
- Prefer ordinary Go files for reusable helpers. Do not leave reusable behavior
  trapped in `.y` grammar prologues or inline semantic actions if a second
  parser will need it. Generated parser `.go` files are build artifacts, not the
  source of truth for shared behavior.
- Version-specific grammar actions should be small: trace the same event Python
  would trace, adapt syntax-only shape differences, and call shared builders.
  If a version-specific wrapper is needed for XTrace fidelity, keep the wrapper
  narrow and keep the underlying AST/declaration construction shared.
- All supported language versions must be compiled into the Go binary and wasm
  build. Do not rely on Python-style runtime grammar mutation, parser-module
  reloads, or process-global grammar hacks.
- Language version state must flow from explicit parse/compiler configuration.
  Avoid hidden globals. Keep parser config, Ivy utility config, include
  selection, isolate config, and downstream compiler behavior synchronized for
  each compilation.
- Preserve the hardened Ivy >=1.7 behavior while extracting helpers. Move one
  reusable behavior at a time, add or preserve tests first, and verify that the
  existing >=1.7 parser emits the same AST, diagnostics, and XTrace behavior
  before adding new Ivy 1.6 grammar coverage.
- A future Ivy 1.6 parser should share the existing post-parse framework from
  day one. If a production cannot use a shared builder, write the smallest
  syntax adapter needed and document the Python source behavior that requires
  the exception.
- Add cross-version tests that compile Ivy 1.6, 1.7, and 1.8 inputs in the same
  process. This guards the browser/wasm requirement that language versions work
  without restarting or recompiling.
