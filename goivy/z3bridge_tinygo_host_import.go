//go:build wasip1

package goivy

//go:wasmimport goivy_z3 bool_sort
//go:noescape
func z3HostMkBoolSort() uint32
