//go:build (tinygo && (wasm || wasm_unknown)) || (wasip1 && wasm)

package goivy

//go:wasmimport goivy_z3 bool_sort
//go:noescape
func z3HostMkBoolSort() uint32
