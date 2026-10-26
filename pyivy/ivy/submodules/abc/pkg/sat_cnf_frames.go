package abc

// Timeframe clauses from sat/cnf/cnfMan.c:Cnf_DataWriteIntoSolverInt.

import "slices"

type CNFFrameOptions struct {
	Base                                                         SatVar
	MaxFrames, MaxObjects, MaxVariables, MaxClauses, MaxLiterals int // zero: 1024, 1M, 1M, 4M, 16M
	Stop                                                         func() bool
}
type AigCNFFrames struct {
	variables, frames, stride int
	base                      SatVar
	clauses                   [][]SatLit
	inputs, outputs           []SatLit
}

func (f *AigCNFFrames) VariableCount() int { return f.variables }
func (f *AigCNFFrames) FrameCount() int    { return f.frames }
func (f *AigCNFFrames) Clauses() [][]SatLit {
	r := make([][]SatLit, len(f.clauses))
	for i, c := range f.clauses {
		r[i] = slices.Clone(c)
	}
	return r
}
func (f *AigCNFFrames) frameLiterals(frame int, source []SatLit) ([]SatLit, error) {
	if f == nil || frame < 0 || frame >= f.frames {
		return nil, ErrDimension
	}
	offset := SatLit(uint64(f.base)+uint64(frame)*uint64(f.stride)) << 1
	r := slices.Clone(source)
	for i := range r {
		r[i] += offset
	}
	return r, nil
}
func (f *AigCNFFrames) InputLiterals(frame int) ([]SatLit, error) {
	if f == nil {
		return nil, ErrDimension
	}
	return f.frameLiterals(frame, f.inputs)
}
func (f *AigCNFFrames) OutputLiterals(frame int) ([]SatLit, error) {
	if f == nil {
		return nil, ErrDimension
	}
	return f.frameLiterals(frame, f.outputs)
}
func (f *AigCNFFrames) LoadInto(s *SatSolver) (bool, error) {
	if f == nil {
		return false, ErrDimension
	}
	return (&AigCNF{variables: f.variables, clauses: f.clauses}).LoadInto(s, 0)
}

// NewSolver creates and simplifies a legacy BSAT instance containing this
// frame relation. A false result reports a trivially inconsistent relation.
func (f *AigCNFFrames) NewSolver(maxVariables int) (*SatSolver, bool, error) {
	if f == nil || maxVariables < 0 {
		return nil, false, ErrDimension
	}
	if maxVariables == 0 {
		maxVariables = 1 << 20
	}
	solver, err := NewSatSolverWithLimit(maxVariables)
	if err != nil {
		return nil, false, err
	}
	ok, err := f.LoadInto(solver)
	if err != nil || !ok {
		return nil, ok, err
	}
	ok, err = solver.Simplify()
	if err != nil || !ok {
		return nil, ok, err
	}
	return solver, true, nil
}

// LoadIntoSolver2 atomically loads this relation into the proof-capable BSAT2
// facade. Stable positive IDs are assigned in native clause order.
func (f *AigCNFFrames) LoadIntoSolver2(solver *SatSolver2, firstID int) (bool, error) {
	if f == nil || solver == nil || firstID <= 0 {
		return false, ErrDimension
	}
	if !solver.usable() || solver.core.proof != nil || solver.core.resolutionProof != nil {
		return false, ErrInvariant
	}
	if len(f.clauses) > 0 && firstID > int(^uint(0)>>1)-len(f.clauses)+1 {
		return false, ErrLimit
	}
	trial := *solver
	core := solver.core.cloneState()
	core.bookmark = solver.core.bookmark
	trial.core = &core
	trial.clauses = cloneSatSolver2Clauses(solver.clauses)
	trial.variablePart = slices.Clone(solver.variablePart)
	if err := trial.SetVariableCount(f.variables); err != nil {
		return false, err
	}
	specifications := make([]SatSolver2ClauseSpec, len(f.clauses))
	for index, clause := range f.clauses {
		if len(clause) == 0 {
			return false, nil
		}
		specifications[index] = SatSolver2ClauseSpec{Literals: slices.Clone(clause), ID: firstID + index}
	}
	if _, err := trial.AddClauseSpecs(specifications); err != nil {
		if err == ErrInvariant {
			return false, nil
		}
		return false, err
	}
	ok, err := trial.Simplify()
	if err != nil || !ok {
		return ok, err
	}
	*solver = trial
	return true, nil
}

func (f *AigCNFFrames) NewSolver2(options SatSolver2Options) (*SatSolver2, bool, error) {
	if f == nil {
		return nil, false, ErrDimension
	}
	solver, err := NewSatSolver2(options)
	if err != nil {
		return nil, false, err
	}
	ok, err := f.LoadIntoSolver2(solver, 1)
	if err != nil || !ok {
		solver.Stop()
		return nil, ok, err
	}
	return solver, true, nil
}

// UnrollFrames copies the relation at frame strides, inserts register-input /
// next-register-output equivalences before each later frame, and optionally
// initializes every first-frame register to zero. Ordering matches the native
// loader; graph/source clauses remain unchanged even on error or cancellation.
// Returned signal literals preserve phase-transformed relation meanings. Multi-
// frame use requires assigned CO variables for the latch-input suffix. No bad-
// state or output assertions are added beyond those already in the source CNF.
func (c *AigMappedCNF) UnrollFrames(g *AigManager, frames int, initialize bool, o CNFFrameOptions) (*AigCNFFrames, error) {
	if g == nil || frames < 1 || o.MaxFrames < 0 || o.MaxObjects < 0 || o.MaxVariables < 0 || o.MaxClauses < 0 || o.MaxLiterals < 0 {
		return nil, ErrDimension
	}
	if o.MaxFrames == 0 {
		o.MaxFrames = 1024
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVariables == 0 {
		o.MaxVariables = 1 << 20
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 4 << 20
	}
	if o.MaxLiterals == 0 {
		o.MaxLiterals = 16 << 20
	}
	stop := func() bool { return o.Stop != nil && o.Stop() }
	if stop() {
		return nil, ErrCancelled
	}
	if frames > o.MaxFrames || len(g.objects) > o.MaxObjects {
		return nil, ErrLimit
	}
	if e := checkCNFTransform(c, CNFTransformOptions{MaxVariables: o.MaxVariables, MaxClauses: o.MaxClauses, MaxLiterals: o.MaxLiterals, Stop: o.Stop}); e != nil {
		return nil, e
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	if len(c.objectVariables) != len(g.objects) || len(c.inputs) != len(g.cis) || len(c.outputs) != len(g.cos) {
		return nil, ErrDimension
	}
	if frames > 1 {
		for _, id := range g.cos[len(g.cos)-g.registers:] {
			if c.objectVariables[id] < 0 {
				return nil, ErrDimension
			}
		}
	}
	variables, e := checkedSize(int(o.Base), frames, c.variables, min(o.MaxVariables, int(MaxLiteralID)+1))
	if e != nil {
		return nil, e
	}
	clauses, e := checkedSize(0, frames, len(c.clauses), o.MaxClauses)
	if e != nil {
		return nil, e
	}
	clauses, e = checkedSize(clauses, frames-1, 2*g.registers, o.MaxClauses)
	if e != nil {
		return nil, e
	}
	literals := 0
	for _, clause := range c.clauses {
		literals += len(clause)
	}
	literals, e = checkedSize(0, frames, literals, o.MaxLiterals)
	if e != nil {
		return nil, e
	}
	literals, e = checkedSize(literals, frames-1, 4*g.registers, o.MaxLiterals)
	if e != nil {
		return nil, e
	}
	if initialize {
		clauses, e = checkedSize(clauses, 1, g.registers, o.MaxClauses)
		if e != nil {
			return nil, e
		}
		literals, e = checkedSize(literals, 1, g.registers, o.MaxLiterals)
		if e != nil {
			return nil, e
		}
	}
	r := &AigCNFFrames{variables: variables, frames: frames, stride: c.variables, base: o.Base, inputs: slices.Clone(c.inputs), outputs: slices.Clone(c.outputs), clauses: make([][]SatLit, 0, clauses)}
	add := func(clause []SatLit) error {
		if stop() {
			return ErrCancelled
		}
		r.clauses = append(r.clauses, clause)
		return nil
	}
	for frame := 0; frame < frames; frame++ {
		if stop() {
			return nil, ErrCancelled
		}
		offset := SatLit(uint64(o.Base)+uint64(frame)*uint64(c.variables)) << 1
		if frame > 0 {
			previous := offset - (SatLit(c.variables) << 1)
			for register := 0; register < g.registers; register++ {
				li := c.outputs[len(c.outputs)-g.registers+register] + previous
				lo := c.inputs[len(c.inputs)-g.registers+register] + offset
				if e := add([]SatLit{li, lo.Not()}); e != nil {
					return nil, e
				}
				if e := add([]SatLit{li.Not(), lo}); e != nil {
					return nil, e
				}
			}
		}
		for _, clause := range c.clauses {
			copy := make([]SatLit, len(clause))
			for i, l := range clause {
				copy[i] = l + offset
			}
			if e := add(copy); e != nil {
				return nil, e
			}
		}
	}
	if initialize {
		offset := SatLit(o.Base) << 1
		for _, lo := range c.inputs[len(c.inputs)-g.registers:] {
			if e := add([]SatLit{(lo + offset).Not()}); e != nil {
				return nil, e
			}
		}
	}
	if stop() {
		return nil, ErrCancelled
	}
	return r, nil
}
