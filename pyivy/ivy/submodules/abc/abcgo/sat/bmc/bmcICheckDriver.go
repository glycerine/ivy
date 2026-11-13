package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func inductionSolver(p *gia.Gia_Man_t, frames, timeout int32) *bsat.Solver {
	m := gia.Gia_ManMiter(p, gia.Gia_ManDup(p), 0, 1, 1, 0, 0)
	enc, err := gia.Mf_ManGenerateCnfResult(m, 8, true, false, false)
	if err != nil {
		panic(err)
	}
	data := &cnf.Cnf_Dat_t{Variables: enc.Variables, Clauses: enc.Clauses, VarNums: enc.VarNums}
	return Bmc_DeriveSolver(p, m, data, frames, timeout, 0)
}
func Bmc_PerformISearchOne(p *gia.Gia_Man_t, frames, timeout, reverse, backTopo, verbose int32, lits *vec.Vec_Int_t) int32 {
	s := inductionSolver(p, frames, timeout)
	assumptions := vec.Vec_IntArray(lits)
	if s.Solve(assumptions...) != bsat.Unsatisfiable {
		return 1
	}
	order := vec.Vec_IntStartNatural(int(gia.Gia_ManRegNum(p)))
	if backTopo != 0 {
		Bmc_PerformFindFlopOrder(p, order)
	}
	if reverse != 0 {
		vec.Vec_IntReverseOrder(order)
	}
	for _, i := range vec.Vec_IntArray(order) {
		if assumptions[i]&1 != 0 {
			continue
		}
		assumptions[i] ^= 1
		switch s.Solve(assumptions...) {
		case bsat.Satisfiable:
			assumptions[i] ^= 1
		case bsat.Unknown:
			return 1
		}
		if verbose != 0 {
			fmt.Printf("Tested flop %d at depth %d\n", i, frames)
		}
	}
	return 0
}
func Bmc_PerformISearch(p *gia.Gia_Man_t, frames, timeout, reverse, backTopo, dump, verbose int32) *vec.Vec_Int_t {
	lits := vec.Vec_IntAlloc(int(gia.Gia_ManRegNum(p)))
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		vec.Vec_IntPush(lits, bsat.Positive(i))
	}
	for f := int32(1); f <= frames; f++ {
		if Bmc_PerformISearchOne(p, f, timeout, reverse, backTopo, verbose, lits) != 0 {
			return nil
		}
	}
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(lits))
	for _, l := range vec.Vec_IntArray(lits) {
		value := int32(1)
		if l&1 != 0 {
			value = 0
		}
		vec.Vec_IntPush(result, value)
	}
	if dump != 0 {
		fmt.Printf("Inductive flop selection: %v\n", vec.Vec_IntArray(result))
	}
	return result
}
func Bmc_PerformICheck(p *gia.Gia_Man_t, frames, timeout, empty, verbose int32) {
	lits := make([]bsat.Literal, gia.Gia_ManRegNum(p))
	for i := range lits {
		lits[i] = bsat.Positive(int32(i))
		if empty != 0 {
			lits[i] ^= 1
		}
	}
	for {
		s := inductionSolver(p, frames, timeout)
		status := s.Solve(lits...)
		if status != bsat.Unsatisfiable {
			fmt.Printf("Inductiveness status: %v\n", status)
			return
		}
		used := make(map[int32]bool)
		for _, lit := range s.UnsatCore() {
			used[lit>>1] = true
		}
		changed := false
		for i := range lits {
			if lits[i]&1 == 0 && !used[int32(i)] {
				lits[i] ^= 1
				changed = true
			}
		}
		if verbose != 0 {
			fmt.Printf("Inductive assumptions: %v\n", lits)
		}
		if !changed || empty != 0 {
			return
		}
	}
}
