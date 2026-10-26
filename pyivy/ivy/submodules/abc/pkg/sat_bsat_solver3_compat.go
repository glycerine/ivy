package abc

// Utility, trace and store compatibility for declarations in satSolver3.h.

func (s *SatSolver3) DIMACS(assumptions []SatLit, options SatSolverDIMACSOptions) ([]byte, error) {
	if !s.usable() {
		return nil, ErrInvariant
	}
	all := make([]SatLit, 0, len(s.pushed)+len(assumptions))
	all = append(all, s.pushed...)
	all = append(all, assumptions...)
	return SatSolverDIMACS(s.core, all, options)
}

func (s *SatSolver3) WriteDIMACS(name string, assumptions []SatLit, options SatSolverDIMACSOptions) error {
	if !s.usable() {
		return ErrInvariant
	}
	all := make([]SatLit, 0, len(s.pushed)+len(assumptions))
	all = append(all, s.pushed...)
	all = append(all, assumptions...)
	return WriteSatSolverDIMACSFile(s.core, name, all, options)
}

func (s *SatSolver3) FormatStatistics() string {
	if !s.usable() {
		return ""
	}
	return FormatSatSolverStatistics(s.core.Statistics(), true)
}

func (s *SatSolver3) Model(variables []SatVar) ([]int, error) {
	if !s.usable() || s.core.LastStatus() != SatSatisfiable {
		return nil, ErrInvariant
	}
	return SatSolverModel(SatResult{Status: SatSatisfiable, Model: s.core.model}, variables)
}

func (s *SatSolver3) DoubleClauses(variable SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.DoubleClauses(variable)
}

func (s *SatSolver3) ClearVariableActivity() error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.ClearVariableActivity()
}

func (s *SatSolver3) ReduceLearnedDatabase() (SatLearnedReductionStats, error) {
	if !s.usable() {
		return SatLearnedReductionStats{}, ErrInvariant
	}
	return s.core.ReduceLearnedDatabase()
}

func (s *SatSolver3) Check() error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.Check()
}

func (s *SatSolver3) VerifyModel() error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.VerifyModel()
}

func (s *SatSolver3) StartTrace(name string, limits SatTraceLimits) error {
	if !s.usable() || s.trace != nil {
		return ErrInvariant
	}
	trace, err := StartSatSolverTrace(name, limits)
	if err != nil {
		return err
	}
	s.trace = trace
	return nil
}

func (s *SatSolver3) TraceWrite(literals []SatLit, root bool) error {
	if !s.usable() || s.trace == nil {
		return ErrInvariant
	}
	return s.trace.Write(literals, root)
}

func (s *SatSolver3) StopTrace() error {
	if !s.usable() || s.trace == nil {
		return ErrInvariant
	}
	trace := s.trace
	s.trace = nil
	return trace.Stop(s.VariableCount())
}

func (s *SatSolver3) AllocateProofStore(limits SatProofStoreLimits) error {
	if !s.usable() || s.store != nil {
		return ErrInvariant
	}
	store, err := NewSatProofStore(limits)
	if err != nil {
		return err
	}
	s.store = store
	return nil
}

// StoreProofClause is the checked insertion path absent from satSolver3.h.
// The source declares store lifecycle calls but provides no solver3 definitions
// or automatic clause hook, so callers must make recording explicit in Go.
func (s *SatSolver3) StoreProofClause(literals []SatLit) (int, error) {
	if !s.usable() || s.store == nil {
		return -1, ErrInvariant
	}
	return s.store.AddClause(literals)
}

func (s *SatSolver3) WriteProofStore(name string) error {
	if !s.usable() || s.store == nil {
		return ErrInvariant
	}
	return s.store.WriteFile(name)
}

func (s *SatSolver3) MarkProofRoots() error {
	if !s.usable() || s.store == nil {
		return ErrInvariant
	}
	return s.store.MarkRoots()
}

func (s *SatSolver3) MarkProofClausesA() error {
	if !s.usable() || s.store == nil {
		return ErrInvariant
	}
	return s.store.MarkClausesA()
}

func (s *SatSolver3) ReleaseProofStore() (*SatProofStore, error) {
	if !s.usable() || s.store == nil {
		return nil, ErrInvariant
	}
	store := s.store
	s.store = nil
	return store, nil
}

func (s *SatSolver3) FreeProofStore() error {
	if !s.usable() || s.store == nil {
		return ErrInvariant
	}
	s.store.Stop()
	s.store = nil
	return nil
}
