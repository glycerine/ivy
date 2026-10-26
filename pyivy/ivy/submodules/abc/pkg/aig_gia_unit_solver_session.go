package abc

// Reusable cardinality-prefix lifecycle from giaSatLut.c:Sbl_ManClean.

type giaUnitWindowSolverSession struct {
	log            int
	solver         *GiaWindowSolver
	builds, resets int
}

func (s *giaUnitWindowSolverSession) prepare(nodes int, cuts []GiaWindowCut, ranges []GiaWindowCutRange, roots []int, o GiaWindowCNFOptions) (*GiaWindowSolver, error) {
	if s.solver == nil {
		solver, e := NewGiaWindowSolver(s.log, nodes, cuts, ranges, roots, o)
		if e != nil {
			return nil, e
		}
		s.solver = solver
		s.builds++
	} else {
		if e := s.solver.ResetWindow(nodes, cuts, ranges, roots, o); e != nil {
			return nil, e
		}
		s.resets++
	}
	return s.solver, nil
}
