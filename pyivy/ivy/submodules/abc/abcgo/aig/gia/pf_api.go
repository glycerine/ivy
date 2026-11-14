package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

const (
	PF_LEAF_MAX = 6
	PF_CUT_MAX  = 32
	PF_NO_LEAF  = 31
	PF_NO_FUNC  = 0x3ffffff
)

type Pf_Mat_t struct{ fCompl, Phase, Perm uint32 }

func Pf_Mat2Int(m Pf_Mat_t) int32 { return pfConfigInt(pfMatchConfig{m.fCompl, m.Phase, m.Perm}) }
func Pf_Int2Mat(v int32) Pf_Mat_t {
	c := pfIntConfig(v)
	return Pf_Mat_t{c.complement, c.phase, c.permutation}
}

type Pf_Man_t struct {
	graph   *Gia_Man_t
	pars    *Jf_Par_t
	library *pfLibrary
	cuts    *nfCutManager
}

func Pf_ManSetDefaultPars(p *Jf_Par_t) {
	*p = Jf_Par_t{NLutSize: 6, NCutNum: 16, NRounds: 3, NCoarseLimit: 3, NAreaTuner: 1, NVerbLimit: 5, DelayTarget: -1, FOptEdge: 1, FCutMin: 1, NLutSizeMax: 6, NCutNumMax: 32, MapDelayTarget: -1, Epsilon: 0.01}
}
func Pf_StoCreate(g *Gia_Man_t, pars *Jf_Par_t) *Pf_Man_t {
	if g == nil || pars == nil {
		return nil
	}
	cells := mio.Mio_CollectRootsNewDefault(6, nil, 0)
	if len(cells) < 4 {
		return nil
	}
	g.Refs = nil
	g.VCellMapping = nil
	library := pfPrepareLibrary(cells)
	return &Pf_Man_t{graph: g, pars: pars, library: library, cuts: library.cutManager(g, pars)}
}
func Pf_StoDelete(p *Pf_Man_t) {
	if p == nil || p.library == nil {
		return
	}
	vec.Vec_MemHashFree(p.library.truth.memory)
	vec.Vec_MemFree(p.library.truth.memory)
	p.library = nil
	p.cuts = nil
}
func Pf_ManComputeCuts(p *Pf_Man_t) { p.cuts.computeCuts() }

// The original C Pf mapper has an empty mapping pass.
func Pf_ManComputeMapping(p *Pf_Man_t) {}
func Pf_StoPrint(p *Pf_Man_t, verbose int32) {
	count := 0
	for id := 2; id < len(p.library.matches); id++ {
		for i, m := range p.library.matches[id] {
			count++
			if verbose == 0 || id >= 10 {
				continue
			}
			cell := &p.library.cells[m.gate]
			truth := vec.Vec_MemReadEntry(p.library.truth.memory, id)
			fmt.Printf("%6d : %6d : %6d : Gate %16s  Area =%8.2f  In = %d   ", count, id, i, cell.Name(), cell.Area, cell.Fanins())
			if m.config.complement != 0 {
				fmt.Print(" compl ")
			} else {
				fmt.Print("       ")
			}
			for pin := int32(0); pin < cell.Fanins(); pin++ {
				v := (m.config.permutation >> uint(3*pin)) & 7
				ch := rune('a' + v)
				if m.config.phase&(1<<uint(pin)) != 0 {
					ch -= 32
				}
				fmt.Printf("%c", ch)
			}
			fmt.Print("  ")
			dau.Dau_DsdPrintFromTruth(truth, util.Abc_TtSupportSize(truth, 6))
		}
	}
	fmt.Printf("Gates = %d.  Truths = %d.  Matches = %d.\n", len(p.library.cells), vec.Vec_MemEntryNum(p.library.truth.memory), count)
}

// Pf_ManPerformMapping mirrors the incomplete C implementation: it computes
// and reports cuts, then returns an unmapped duplicate of the input graph.
func Pf_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if g == nil || pars == nil {
		return nil
	}
	if Gia_ManHasChoices(g) != 0 {
		pars.FCoarsen = 0
	}
	work := g
	if pars.FCoarsen != 0 {
		work = Gia_ManDupMuxes(g, pars.NCoarseLimit)
	}
	p := Pf_StoCreate(work, pars)
	if p == nil {
		return nil
	}
	defer Pf_StoDelete(p)
	Pf_StoPrint(p, 1)
	Pf_ManComputeCuts(p)
	return Gia_ManDup(g)
}
