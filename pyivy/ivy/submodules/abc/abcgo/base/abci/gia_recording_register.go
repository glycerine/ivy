package abci

import (
	"github.com/glycerine/abcgo/aig/gia"
	ifpkg "github.com/glycerine/abcgo/map/if"
)

func init() {
	gia.RegisterRecordingState(func() bool { return Abc_NtkRecIsRunning3() != 0 })
	ifpkg.RegisterRecordedLibrary(ifpkg.RecordedLibrary{DelayCost: If_CutDelayRecCost3, BuildGia: Abc_RecToGia3})
	ifpkg.RegisterExactDelayCost(ifExactDelayCost)
}

// IF stores pin bytes while the source exact-synthesis API exposes signed chars.
// Copy through a bounded stack buffer, preserving bits outside the active pins.
func ifExactDelayCost(truth *uint64, vars int32, arrivals *int32, pins []byte, cost *int32, aigLevel int32) int32 {
	var signed [8]int8
	size := min(len(pins), len(signed))
	for i := 0; i < size; i++ {
		signed[i] = int8(pins[i])
	}
	delay := Abc_ExactDelayCost(truth, vars, arrivals, signed[:size], cost, aigLevel)
	for i := 0; i < size; i++ {
		pins[i] = byte(signed[i])
	}
	return delay
}
