// Package smt is the Ivy/Z3 boundary.
//
// Rule for browser/local-worker Ivy-Go builds: wasip1 only.
//
// Do not add raw "wasm" or "wasm_unknown" build-tag variants here. They are a
// different execution environment from wasip1, and mixing the two makes the
// host-import contract ambiguous. Go still emits a .wasm binary for GOOS=wasip1
// GOARCH=wasm, but code selection in this package is by the wasip1 build tag.
package smt
