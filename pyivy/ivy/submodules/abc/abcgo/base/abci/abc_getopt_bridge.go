package abci

import (
	"sync"

	"github.com/glycerine/abcgo/misc/extra"
)

// globalUtilOptind mirrors ABC's C global option index. Command handlers often
// advance it themselves after consuming positional arguments, so every call
// into extra's legacy parser first synchronizes its index from this variable.
var globalUtilOptind int32

var abcUtilGetoptMu sync.Mutex

func Extra_UtilGetoptReset() {
	abcUtilGetoptMu.Lock()
	defer abcUtilGetoptMu.Unlock()
	globalUtilOptind = 0
	extra.Extra_UtilGetoptReset()
}

func Extra_UtilGetopt(argc int32, argv []string, spec string) int32 {
	abcUtilGetoptMu.Lock()
	defer abcUtilGetoptMu.Unlock()
	extra.Extra_UtilGetoptSetIndex(globalUtilOptind)
	result := extra.Extra_UtilGetopt(argc, argv, spec)
	globalUtilOptind = extra.Extra_UtilGetoptIndex()
	return result
}

func Extra_UtilGetoptArg() string {
	abcUtilGetoptMu.Lock()
	defer abcUtilGetoptMu.Unlock()
	return extra.Extra_UtilGetoptArg()
}

func Extra_UtilGetoptIndex() int32 {
	abcUtilGetoptMu.Lock()
	defer abcUtilGetoptMu.Unlock()
	return globalUtilOptind
}

func Extra_UtilGetoptSetIndex(index int32) {
	abcUtilGetoptMu.Lock()
	defer abcUtilGetoptMu.Unlock()
	globalUtilOptind = index
	extra.Extra_UtilGetoptSetIndex(index)
}
