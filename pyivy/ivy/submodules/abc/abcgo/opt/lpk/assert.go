package lpk

import (
	"fmt"
	"path"
	"runtime"
)

// assert accepts C-derived integer flags as well as Go booleans while the
// port preserves the original assertion sites.
func assert(condition any, message string) {
	ok := false
	switch value := condition.(type) {
	case bool:
		ok = value
	case int:
		ok = value != 0
	case int32:
		ok = value != 0
	case int64:
		ok = value != 0
	case uint:
		ok = value != 0
	case uint32:
		ok = value != 0
	case uint64:
		ok = value != 0
	case float32:
		ok = value != 0
	case float64:
		ok = value != 0
	}
	if ok {
		return
	}
	_, fileName, fileLine, callerOK := runtime.Caller(1)
	where := ""
	if callerOK {
		where = fmt.Sprintf("%s:%d: ", path.Base(fileName), fileLine)
	}
	panic(where + "assertion failed, we do not see: " + message)
}
