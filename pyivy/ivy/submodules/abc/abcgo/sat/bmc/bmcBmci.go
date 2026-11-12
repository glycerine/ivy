package bmc

import (
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManBmciTest compares the supplied initial state against the all-zero
// state using a shared-input sequential miter. The legacy simulation knobs are
// retained for command compatibility; the exact SAT check supersedes sampling.
func Gia_ManBmciTest(graph *gia.Gia_Man_t, init *vec.Vec_Int_t, frames, words, timeout, sim, verbose int32) int32 {
	if graph == nil || init == nil {
		return -1
	}
	native := gia.Gia_ManToAigSimple(graph)
	regs := int(gia.Gia_ManRegNum(graph))
	if vec.Vec_IntSize(init) != regs {
		return -1
	}
	initial := make([]int32, regs)
	for i := range initial {
		initial[i] = vec.Vec_IntEntry(init, i)
		if initial[i] < 0 {
			return -1
		}
	}
	_ = words
	_ = sim
	_ = verbose
	var duration time.Duration
	if timeout > 0 {
		duration = time.Duration(timeout) * time.Second
	}
	return compareInitialStateLiteralsAIG(native, initial, make([]int32, regs), frames, duration)
}
