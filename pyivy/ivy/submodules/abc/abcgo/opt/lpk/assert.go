package lpk

import (
	"fmt"
	"path"
	"runtime"
)

// assert accepts C-derived integer flags as well as Go booleans while the
// port preserves the original assertion sites: it fails when condition is
// false or zero. Being generic, it neither boxes its argument nor blocks
// inlining.
func assert[T comparable](condition T, message string) {
	var zero T
	if condition == zero {
		assertFailed(message)
	}
}

//go:noinline
func assertFailed(message string) {
	_, fileName, fileLine, ok := runtime.Caller(2)
	where := ""
	if ok {
		where = fmt.Sprintf("%s:%d: ", path.Base(fileName), fileLine)
	}
	panic(where + "assertion failed, we do not see: " + message)
}
