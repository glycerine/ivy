package if_

import (
	abcutil "github.com/glycerine/abcgo/misc/util"
	"sync"
)

// ExactDelayCost connects ABC's exact synthesis cost routine without a package
// import cycle. The truth, arrival and pin buffers retain their source order.
type ExactDelayCost func(*uint64, int32, *int32, []byte, *int32, int32) int32

var ifExactDelay struct {
	sync.RWMutex
	cost ExactDelayCost
}

// RegisterExactDelayCost installs the source exact-synthesis cost routine.
func RegisterExactDelayCost(cost ExactDelayCost) {
	ifExactDelay.Lock()
	ifExactDelay.cost = cost
	ifExactDelay.Unlock()
}

func exactDelayBackend() ExactDelayCost {
	ifExactDelay.RLock()
	cost := ifExactDelay.cost
	ifExactDelay.RUnlock()
	return cost
}

// The two source fUserSesLib branches share this evaluation. Successful area
// Cost is intentionally not assigned to cut.Cost in C; only failure changes it.
func ifCutExactDelay(p *If_Man_t, obj *If_Obj_t, cut *If_Cut_t) float32 {
	evaluate := exactDelayBackend()
	if evaluate == nil {
		panic("if: exact-delay backend is not registered")
	}
	var cost int32
	cut.fUser = 1
	delay := evaluate(If_CutTruthW(p, cut), int32(cut.nLeaves), If_CutArrTimeProfile(p, cut), If_CutPerm(cut), &cost, If_ManCutAigDelay(p, obj, cut))
	if cost == abcutil.ABC_INFINITY {
		for pin := range If_CutPerm(cut)[:cut.nLeaves] {
			cut.ext.permutation[pin] = IF_BIG_CHAR
		}
		cut.Cost, cut.fUseless = IF_COST_MAX, 1
	}
	return float32(delay)
}
