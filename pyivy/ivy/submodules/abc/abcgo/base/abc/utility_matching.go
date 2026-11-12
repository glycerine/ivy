package abc

import "reflect"

// Abc_NtkDetectMatching is retained for compatibility; ABC's implementation
// contains only disabled diagnostic code.
func Abc_NtkDetectMatching(_ *Network) {}

// Abc_ObjPointerCompare orders object pointers for the original pointer sort.
func Abc_ObjPointerCompare(a, b *Object) int32 {
	x, y := reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer()
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}
