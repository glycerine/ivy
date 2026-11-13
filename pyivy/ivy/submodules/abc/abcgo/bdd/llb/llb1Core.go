package llb

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func Llb_ManSetDefaultParams(p *Gia_ParLlb_t) {
	if p == nil {
		return
	}
	*p = Gia_ParLlb_t{nBddMax: 10000000, nIterMax: 10000000, nClusterMax: 20, nVolumeMax: 100, nVolumeMin: 30, nPartValue: 5, fReorder: 1, iFrame: -1}
}
func Llb_ManPrintAig(p *Llb_Man_t) {
	if p == nil || p.pAig == nil {
		return
	}
	fmt.Printf("pi=%d po=%d ff=%d vars=%d and=%d\n", saig.Saig_ManPiNum(p.pAig), saig.Saig_ManPoNum(p.pAig), saig.Saig_ManRegNum(p.pAig), vec.Vec_IntSize(p.vVar2Obj), aig.Aig_ManNodeNum(p.pAig))
}

// Llb_ManModelCheckAig creates an independent working graph and runs exact
// symbolic reachability. The source graph remains caller-owned.
func Llb_ManModelCheckAig(original *Aig_Man_t, params *Gia_ParLlb_t, hints *Vec_Int_t, global **DdManager) int32 {
	if original == nil {
		return -1
	}
	if params == nil {
		params = &Gia_ParLlb_t{}
		Llb_ManSetDefaultParams(params)
	}
	working := aig.Aig_ManDupSimple(original)
	if hints != nil {
		working = aig.Aig_ManDupSimpleWithHints(original, vec.Vec_IntArray(hints))
	}
	manager := &Llb_Man_t_{pAigGlo: original, pAig: working, pPars: params}
	status := int32(-1)
	if params.fSkipReach == 0 {
		status = Llb_ManReachability(manager, hints, global)
	}
	if manager.reached != nil && global != nil && *global == nil {
		*global = manager.dd
		manager.dd = nil
	}
	if manager.dd != nil {
		core.Cudd_Quit(manager.dd)
	}
	aig.Aig_ManStop(working)
	return status
}

func Llb_ManModelCheckGia(g *Gia_Man_t, params *Gia_ParLlb_t) int32 {
	if g == nil {
		return -1
	}
	intermediate := gia.Gia_ManDupDfs(g)
	graph := gia.Gia_ManToAigSimple(intermediate)
	gia.Gia_ManStop(intermediate)
	result := Llb_ManModelCheckAig(graph, params, nil, nil)
	aig.Aig_ManStop(graph)
	return result
}
