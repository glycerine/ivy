package gia

import (
	"math/bits"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type nfGateMatch struct {
	gate   int32
	cfg    Nf_Cfg_t
	area   float32
	delays []int32
}
type nfLibrary struct {
	truth                        *nfTruthStore
	cells                        []mio.Mio_Cell2_t
	matches                      [][]nfGateMatch
	pinFilter, pinPerm, pinQuick bool
}

func (p *nfLibrary) add(cell *mio.Mio_Cell2_t, truth uint64, fans []int32) {
	phase := uint32(truth & 1)
	if phase != 0 {
		truth = ^truth
	}
	id := vec.Vec_MemHashInsert(p.truth.memory, []uint64{truth})
	for len(p.matches) <= id {
		p.matches = append(p.matches, nil)
	}
	cfg := Nf_Cfg_t{fCompl: phase}
	delays := make([]int32, len(fans))
	for i, lit := range fans {
		variable := uint32(lit >> 1)
		cfg.Perm |= uint32(i) << (4 * variable)
		cfg.Phase |= uint32(lit&1) << variable
		delays[i] = cell.Delay(int(variable))
	}
	for _, old := range p.matches[id] {
		if p.pinFilter && old.cfg.Phase == cfg.Phase && old.cfg.fCompl == cfg.fCompl && cell.AreaF+NF_EPSILON >= old.area {
			dominated := true
			for i, d := range delays {
				if d < old.delays[i] {
					dominated = false
					break
				}
			}
			if dominated {
				return
			}
		}
		if !p.pinPerm && old.gate == int32(cell.Id) {
			if p.pinQuick && bits.OnesCount32(old.cfg.Phase) == bits.OnesCount32(cfg.Phase) || !p.pinQuick && old.cfg.Phase == cfg.Phase {
				return
			}
		}
	}
	p.matches[id] = append(p.matches[id], nfGateMatch{gate: int32(cell.Id), cfg: cfg, area: cell.AreaF, delays: delays})
}

// nfPrepareLibrary enumerates input polarities and pin permutations in the
// same Gray-code and adjacent-swap order as the source mapper.
func nfPrepareLibrary(cells []mio.Mio_Cell2_t, pinFilter, pinPerm, pinQuick bool) *nfLibrary {
	p := &nfLibrary{truth: &nfTruthStore{memory: vec.Vec_MemAllocForTT(6, 0)}, cells: cells, matches: make([][]nfGateMatch, 2), pinFilter: pinFilter, pinPerm: pinPerm, pinQuick: pinQuick}
	for i := 2; i < len(cells); i++ {
		cell := &cells[i]
		n := cell.Fanins()
		if n < 1 || n > 6 {
			panic("gia: invalid Nf library cell")
		}
		fans := make([]int32, n)
		for j := range fans {
			fans[j] = int32(2 * j)
		}
		truth := cell.Truth()
		initial := truth
		gray := extra.Extra_GreyCodeSchedule(n)
		perm := extra.Extra_PermSchedule(n)
		for _, swap := range perm {
			before := truth
			for _, flip := range gray {
				p.add(cell, truth, fans)
				truth = util.Abc_Tt6Flip(truth, flip)
				fans[flip] ^= 1
			}
			if truth != before {
				panic("gia: invalid Nf phase schedule")
			}
			if len(perm) > 1 {
				truth = util.Abc_Tt6SwapAdjacent(truth, swap)
				fans[swap], fans[swap+1] = fans[swap+1], fans[swap]
			}
		}
		if truth != initial {
			panic("gia: invalid Nf permutation schedule")
		}
	}
	p.truth.matchCount = len(p.matches)
	return p
}
