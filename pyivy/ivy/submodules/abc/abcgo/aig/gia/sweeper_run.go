package gia

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// Gia_SweeperRun applies an ABC command to extracted probe cones and grafts
// the resulting outputs back into the running sweeper.
func Gia_SweeperRun(p *Gia_Man_t, probes *Vec_Int_t, command string, verbose int32) int32 {
	sweeper(p)
	graph := Gia_SweeperExtractUserLogic(p, probes, nil, nil)
	if command != "" {
		if verbose != 0 {
			fmt.Printf("GIA manager before %q: %d AND nodes\n", command, Gia_ManAndNum(graph))
		}
		frame := abcmain.Abc_FrameGetGlobalFrame()
		abcmain.Abc_FrameUpdateGia(frame, graph)
		status := abcmain.Cmd_CommandExecute(frame, command)
		result, ok := abcmain.Abc_FrameGetGia(frame).(*Gia_Man_t)
		if status != 0 || !ok || result == nil {
			return 0
		}
		graph = result
		if verbose != 0 {
			fmt.Printf("GIA manager after %q: %d AND nodes\n", command, Gia_ManAndNum(graph))
		}
	}
	if Gia_ManPoNum(graph) != int32(Vec_IntSize(probes)) || Gia_ManPiNum(graph) != Gia_ManPiNum(p) {
		Gia_ManStop(graph)
		return 0
	}
	lits := Gia_SweeperGraft(p, nil, graph)
	Gia_ManStop(graph)
	for i, id := range Vec_IntArray(probes) {
		Gia_SweeperProbeUpdate(p, id, Vec_IntEntry(lits, i))
	}
	return 1
}
