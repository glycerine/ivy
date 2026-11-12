package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// pfMatchConfig preserves the original 8/6/18-bit packed representation.
type pfMatchConfig struct{ complement, phase, permutation uint32 }

func pfConfigInt(c pfMatchConfig) int32 {
	return int32(c.complement&255 | (c.phase&63)<<8 | (c.permutation&0x3ffff)<<14)
}
func pfIntConfig(v int32) pfMatchConfig {
	u := uint32(v)
	return pfMatchConfig{u & 255, (u >> 8) & 63, u >> 14}
}

type pfGateMatch struct {
	gate   int32
	config pfMatchConfig
}
type pfLibrary struct {
	truth   *nfTruthStore
	cells   []mio.Mio_Cell_t
	matches [][]pfGateMatch
}

func (p *pfLibrary) add(truth uint64, fans []int32, gate int32) {
	complement := uint32(truth & 1)
	if complement != 0 {
		truth = ^truth
	}
	id := vec.Vec_MemHashInsert(p.truth.memory, []uint64{truth})
	for len(p.matches) <= id {
		p.matches = append(p.matches, nil)
	}
	cfg := pfMatchConfig{complement: complement}
	for i, lit := range fans {
		cfg.permutation |= uint32(lit>>1) << uint(3*i)
		cfg.phase |= uint32(lit&1) << uint(i)
	}
	for _, old := range p.matches[id] {
		if old.gate == gate && old.config.phase == cfg.phase {
			return
		}
	}
	p.matches[id] = append(p.matches[id], pfGateMatch{gate, cfg})
}
func pfPrepareLibrary(cells []mio.Mio_Cell_t) *pfLibrary {
	p := &pfLibrary{truth: &nfTruthStore{memory: vec.Vec_MemAllocForTT(6, 0)}, cells: cells, matches: make([][]pfGateMatch, 2)}
	for i := 4; i < len(cells); i++ {
		cell := &cells[i]
		n := cell.Fanins()
		if n < 2 || n > 6 {
			panic("gia: invalid Pf cell")
		}
		fans := make([]int32, n)
		for j := range fans {
			fans[j] = int32(2 * j)
		}
		truth := cell.Truth()
		initial := truth
		for _, swap := range extra.Extra_PermSchedule(n) {
			before := truth
			for _, flip := range extra.Extra_GreyCodeSchedule(n) {
				p.add(truth, fans, int32(cell.Id))
				truth = util.Abc_Tt6Flip(truth, flip)
				fans[flip] ^= 1
			}
			if truth != before {
				panic("gia: invalid Pf phase schedule")
			}
			truth = util.Abc_Tt6SwapAdjacent(truth, swap)
			fans[swap], fans[swap+1] = fans[swap+1], fans[swap]
		}
		if truth != initial {
			panic("gia: invalid Pf permutation schedule")
		}
	}
	p.truth.matchCount = len(p.matches)
	return p
}

// cutManager shares the Nf cut representation and truth operations, while
// selecting Pf's exact area-flow ordering.
func (p *pfLibrary) cutManager(g *Gia_Man_t, pars *Jf_Par_t) *nfCutManager {
	lib := &nfLibrary{truth: p.truth, matches: make([][]nfGateMatch, len(p.matches))}
	for i, matches := range p.matches {
		lib.matches[i] = make([]nfGateMatch, len(matches))
	}
	manager := newNfCutManager(g, pars, lib)
	manager.exactFlow = true
	return manager
}
