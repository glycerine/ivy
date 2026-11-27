package reo

import (
	"fmt"
	"path"
	"runtime"
)

// reoDebug enables the checks that C runs only in debug builds (the
// native ABC is built with -DNDEBUG, which compiles its assertions out).
const reoDebug = false

// assert is small enough to inline into hot code; the failure path is not.
func assert(x bool, message string) {
	if !x {
		assertFailed(message)
	}
}

//go:noinline
func assertFailed(message string) {
	_, fileName, fileLine, ok := runtime.Caller(2)
	var where string
	if ok {
		where = fmt.Sprintf("%s:%d: ", path.Base(fileName), fileLine)
	}
	panic(where + "assertion failed, we do not see: " + message)
}
