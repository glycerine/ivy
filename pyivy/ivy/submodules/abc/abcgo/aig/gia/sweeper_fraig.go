package gia

import abcmain "github.com/glycerine/abcgo/base/main"

// Gia_SweeperFraig sweeps selected probes under the current conditions and
// grafts the result back, keeping the original probe IDs.
func Gia_SweeperFraig(p *Gia_Man_t, probes *Vec_Int_t, command string, words, conflicts, verify, verbose int32) int32 {
	graph := Gia_SweeperSweep(p, probes, words, conflicts, verify, verbose)
	if graph == nil {
		return 0
	}
	if command != "" {
		frame := abcmain.Abc_FrameGetGlobalFrame()
		abcmain.Abc_FrameUpdateGia(frame, graph)
		status := abcmain.Cmd_CommandExecute(frame, command)
		result, ok := abcmain.Abc_FrameGetGia(frame).(*Gia_Man_t)
		if status != 0 || !ok || result == nil {
			return 0
		}
		graph = result
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
func Gia_SweeperFraigTest(initial *Gia_Man_t, words, conflicts, verbose int32) *Gia_Man_t {
	p := Gia_ManDupOneHot(initial)
	Gia_SweeperStart(p)
	outputs := Vec_IntAlloc(int(Gia_ManPoNum(p)))
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		probe := Gia_SweeperProbeCreate(p, Gia_ObjFaninLit0p(p, Gia_ManPo(p, i)))
		if i < Gia_ManPoNum(p)-p.nConstrs {
			Vec_IntPush(outputs, probe)
		} else {
			Gia_SweeperCondPush(p, probe)
		}
	}
	result := Gia_SweeperSweep(p, outputs, words, conflicts, verbose, 0)
	Gia_SweeperStop(p)
	Gia_ManStop(p)
	return result
}
