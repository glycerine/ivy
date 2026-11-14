package wlc

import (
	"fmt"
	"path"
	"runtime"
)

func assert(x bool, message string) {
	if x {
		return
	}
	_, fileName, fileLine, ok := runtime.Caller(1)
	var where string
	if ok {
		where = fmt.Sprintf("%s:%d: ", path.Base(fileName), fileLine)
	}
	panic(where + "assertion failed, we do not see: " + message)
}
