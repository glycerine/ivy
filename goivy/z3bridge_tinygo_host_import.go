//go:build tinygo && (wasm || wasm_unknown)

package goivy

//go:wasmimport goivy_z3 bool_sort
//go:noescape
func z3HostMkBoolSort() uint32
