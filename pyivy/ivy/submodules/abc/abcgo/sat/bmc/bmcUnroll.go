package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
)

// Unr_Man_t_ retains the source graph for successive frame requests. Native
// frame construction owns its result, so the manager itself needs no arena.
type Unr_Man_t_ struct {
	source  *aig.Graph
	verbose int32
}

func Unr_ManUnrollStart(graph *gia.Gia_Man_t, verbose int32) *Unr_Man_t {
	if graph == nil {
		return nil
	}
	return &Unr_Man_t_{source: gia.Gia_ManToAigSimple(graph), verbose: verbose}
}

func Unr_ManUnrollFrame(manager *Unr_Man_t, frame int32) *gia.Gia_Man_t {
	if manager == nil || frame < 0 {
		return nil
	}
	return gia.Gia_ManFromAig(UnrollAIGFrames(manager.source, frame+1))
}

func Unr_ManUnroll(graph *gia.Gia_Man_t, frames int32) *gia.Gia_Man_t {
	if graph == nil {
		return nil
	}
	return gia.Gia_ManFromAig(UnrollAIGFrames(gia.Gia_ManToAigSimple(graph), frames))
}

func Unr_ManUnrollSimple(graph *gia.Gia_Man_t, frames int32) *gia.Gia_Man_t {
	return Unr_ManUnroll(graph, frames)
}

func Unr_ManFree(manager *Unr_Man_t) {
	if manager != nil {
		manager.source = nil
	}
}
