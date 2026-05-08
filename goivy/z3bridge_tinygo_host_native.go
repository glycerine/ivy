//go:build tinygo && !(wasm || wasm_unknown)

package goivy

func z3HostMkBoolSort() uint32 { return 1 }
