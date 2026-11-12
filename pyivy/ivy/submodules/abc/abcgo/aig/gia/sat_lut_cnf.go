package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

// Sbl_ManCreateCnf encodes that a selected window node chooses a cut, and
// that each selected cut requires its root and internal fanin nodes.
func Sbl_ManCreateCnf(p *Sbl_Man_t) int32 {
	if p.pSat.NumVars() != p.FirstVar {
		panic("gia: SAT window solver not reset")
	}
	total := p.FirstVar + int32(vec.Vec_WrdSize(p.vCutsI1))
	for p.pSat.NumVars() < total {
		p.pSat.NewVariable()
	}
	add := func(lits ...int32) {
		if !p.pSat.AddClause(lits...) {
			panic("gia: inconsistent SAT window clauses")
		}
	}
	for i := 0; i < vec.Vec_IntSize(p.vAnds); i++ {
		index := vec.Vec_IntSize(p.vLeaves) + i
		start := vec.Vec_IntEntry(p.vCutsStart, index)
		end := start + vec.Vec_IntEntry(p.vCutsNum, index) - 1
		clause := []int32{2*int32(i) + 1}
		for k := start; k < end; k++ {
			clause = append(clause, 2*(p.FirstVar+k))
		}
		add(clause...)
		for k := start; k < end; k++ {
			cut := p.FirstVar + k
			add(2*cut+1, 2*int32(i))
			for part, mask := range []word{vec.Vec_WrdEntry(p.vCutsN1, int(k)), vec.Vec_WrdEntry(p.vCutsN2, int(k))} {
				for mask != 0 {
					bit := bits.TrailingZeros64(mask)
					add(2*cut+1, 2*int32(part*64+bit))
					mask &= mask - 1
				}
			}
		}
	}
	return 1
}
