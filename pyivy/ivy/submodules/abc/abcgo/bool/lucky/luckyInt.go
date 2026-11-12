package lucky

import (
	"fmt"
	"sync"
	"time"
)

type word = uint64
type Abc_TtStore_t struct {
	NVars, NWords, NFuncs int32
	Funcs                 [][]uint64
}
type varInfo struct{ direction, position int32 }
type swapInfo struct {
	posArray                               []varInfo
	realArray                              []int32
	varN, positionToSwap1, positionToSwap2 int32
}

var timePrintState = struct {
	sync.Mutex
	begin time.Time
}{begin: time.Now()}

func TimePrint(message string) {
	timePrintState.Lock()
	defer timePrintState.Unlock()
	now := time.Now()
	if message != "" {
		fmt.Printf("%s = %f sec.\n", message, now.Sub(timePrintState.begin).Seconds())
	}
	timePrintState.begin = now
}
func CompareWords(x, y uint64) int32 {
	if x > y {
		return 1
	}
	if x < y {
		return -1
	}
	return 0
}
func luckyMin(x, y int32) int32 { return min(x, y) }
func luckyMax(x, y int32) int32 { return max(x, y) }
