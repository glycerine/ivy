// Package cgt implements SAT-validated sequential clock gating.
package cgt

import "github.com/glycerine/abcgo/aig/aig"

type Cgt_Par_t struct {
	LevelMax, CandidateMax, ODCMax, ConflictMax, VarsMin, FlopsMin int
	AreaOnly, Verbose, VeryVerbose                                 bool
}

func Cgt_SetDefaultParams(p *Cgt_Par_t) {
	*p = Cgt_Par_t{LevelMax: 25, CandidateMax: 1000, ConflictMax: 10, VarsMin: 1000, FlopsMin: 10}
}
func DefaultParams() Cgt_Par_t { var p Cgt_Par_t; Cgt_SetDefaultParams(&p); return p }

type Gates [][]aig.Edge
type Statistics struct{ Recycles, Calls, SAT, UNSAT, Unknown, Filtered, Patterns int }
