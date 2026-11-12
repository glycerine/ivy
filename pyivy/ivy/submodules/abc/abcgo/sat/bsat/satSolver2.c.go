package bsat

import "reflect"

/**************************************************************************************************
MiniSat -- Copyright (c) 2005, Niklas Sorensson
http://www.cs.chalmers.se/Cs/Research/FormalMethods/MiniSat/

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

func sat_solver2_delete(s *sat_solver2) {
	if s != nil {
		if s.pFile != nil {
			s.pFile.Close()
		}
		*s = sat_solver2{}
	}
}
func luby2(y float64, x int32) float64        { return luby(y, x) }
func solver2_progress(s *sat_solver2) float64 { return sat_solver_progress(&s.sat_solver_t) }
func sat_solver2_memory(s *sat_solver2, all int32) float64 {
	memory := sat_solver_memory(&s.sat_solver_t)
	if all == 0 {
		for _, w := range s.wlists {
			memory -= float64(cap(w.ptr)) * 4
		}
	}
	memory += float64(reflect.TypeOf(*s).Size() - reflect.TypeOf(s.sat_solver_t).Size())
	memory += float64(4*(cap(s.units)+cap(s.variableA))) + float64(cap(s.markedLevels))
	for _, v := range []*veci{&s.claProofs, &s.temp_proof, &s.mark_levels, &s.min_lit_order, &s.min_step_order, &s.learnt_live} {
		memory += float64(cap(v.ptr)) * 4
	}
	return memory
}
func sat_solver2_memory_proof(s *sat_solver2) float64 {
	memory := s.dPrfMemory
	if s.pPrf1 != nil {
		memory += float64(reflect.TypeOf(*s.pPrf1).Size()) + float64(cap(s.pPrf1.order))*4
		for _, node := range s.pPrf1.nodes {
			memory += float64(reflect.TypeOf(*node).Size()) + float64(cap(node.pEnts))*4
		}
	}
	return memory
}
func sat_solver2_check_watched(s *sat_solver2) int32 {
	for i := range s.wlists {
		w := &s.wlists[i]
		kept := int32(0)
		for _, h := range veci_begin(w) {
			c := clause2_read(s, h)
			satisfied := false
			for _, l := range c.lits[:c.size] {
				if s.assigns[lit_var(l)] == lit_sign(l) {
					satisfied = true
					break
				}
			}
			if !satisfied {
				w.ptr[kept] = h
				kept++
			}
		}
		veci_resize(w, kept)
	}
	return 0
}
func Sat_ProofCore(s *sat_solver2) []int32 {
	if s.pPrf1 != nil {
		return Proof_DeriveCore(s.pPrf1, s.hProofLast)
	}
	if s.pPrf2 != nil {
		s.dPrfMemory = max(s.dPrfMemory, Prf_ManMemory(s.pPrf2))
		return Prf_ManUnsatCore(s.pPrf2)
	}
	return nil
}
