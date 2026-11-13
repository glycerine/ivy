package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Gia_ManFaultAnalyze disables fault parameters proven impossible by SAT.
// Pending candidates are returned in scratch, as in the original routine.
func Gia_ManFaultAnalyze(s *bsat.Solver, params, enabled, scratch *vec.Vec_Int_t, iteration int32) int32 {
	p, m := vec.Vec_IntArray(params), vec.Vec_IntArray(enabled)
	if len(p) != len(m) {
		panic("bmc: fault parameter map length mismatch")
	}
	vec.Vec_IntClear(scratch)
	for _, v := range m {
		if v != 0 && v != 1 {
			panic("bmc: non-binary fault map")
		}
		vec.Vec_IntPush(scratch, v)
	}
	pending := vec.Vec_IntArray(scratch)
	start := time.Now()
	runs, unsat := 0, int32(0)
	s.SetLimits(100, 0)
	defer s.SetLimits(0, 0)
	for i, variable := range p {
		if pending[i] == 0 {
			continue
		}
		runs++
		switch s.Solve(bsat.Positive(variable)) {
		case bsat.Unsatisfiable:
			m[i] = 0
			unsat++
		case bsat.Satisfiable:
			for j, v := range p {
				value, ok := s.Value(v)
				if !ok {
					panic("bmc: missing solver model")
				}
				if pending[j] != 0 && value {
					pending[j] = 0
				}
			}
		}
	}
	disabled := 0
	for _, v := range m {
		if v == 0 {
			disabled++
		}
	}
	fmt.Printf("Iteration %3d has determined %5d (out of %5d) parameters after %6d SAT calls.  Time = %s\n", iteration, disabled, len(p), runs, time.Since(start))
	return unsat
}
