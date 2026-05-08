// Package smt is the Ivy/Z3 boundary.
//
// Rule for browser/local-worker Ivy-Go builds: wasip1 only.
//
// Do not add raw "wasm" or "wasm_unknown" build-tag variants here. They are a
// different execution environment from wasip1, and mixing the two makes the
// host-import contract ambiguous. Go still emits a .wasm binary for GOOS=wasip1
// GOARCH=wasm, but the implementation is selected by GOOS=wasip1.
//
// The manual command "go test -tags web ./smt" is reserved for a host-side
// test harness that runs the browser/worker Playwright round trip.
package smt
