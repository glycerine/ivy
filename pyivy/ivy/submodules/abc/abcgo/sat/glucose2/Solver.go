package glucose2

/****************************************************************************************[Solver.h]
 Glucose -- Copyright (c) 2009, Gilles Audemard, Laurent Simon
                CRIL - Univ. Artois, France
                LRI  - Univ. Paris Sud, France

Glucose sources are based on MiniSat (see below MiniSat copyrights). Permissions and copyrights of
Glucose are exactly the same as Minisat on which it is based on. (see below).

---------------
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

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

import (
	"io"
	"sync/atomic"
	"time"
)

// Watcher keeps a satisfied blocking literal alongside a watched clause.
type Watcher struct {
	Clause  *Clause
	Blocker Lit
}

// Solver implements Glucose's first-UIP CDCL search. Clause references are stable
// Go pointers, so collection never changes implication reasons or watch entries.
type Solver struct {
	JustificationMode int
	JustModel         []Lit
	circuit           *circuitState

	// ABC integration callbacks use Go values rather than unmanaged C buffers.
	CnfManager           any
	CnfCallback          func(manager any, message int, model []int) int
	Stop                 func() bool
	Deadline             time.Time
	TerminateSearchEarly bool

	Model                                                               []LBool
	Conflict                                                            []Lit
	K, R                                                                float64
	SizeLBDQueue, SizeTrailQueue                                        int
	FirstReduceDB, IncReduceDB, SpecialIncReduceDB                      int
	LBLBDFrozenClause, LBSizeMinimizingClause, LBLBDMinimizingClause    int
	VarDecay, ClauseDecay, RandomVarFreq, RandomSeed                    float64
	CCMinMode, PhaseSaving                                              int
	RandomPolarity, RandomInitialActivity                               bool
	Verbosity                                                           int
	CertifiedOutput                                                     io.Writer
	CertifiedUNSAT                                                      bool
	Solves, Starts, Decisions, RandomDecisions, Propagations, Conflicts int64
	ConflictsRestarts, RemovedClauses, ReducedClauses, ReduceDBs        int64
	clauses, learnts                                                    []*Clause
	assigns                                                             []LBool
	levels                                                              []int
	reasons                                                             []*Clause
	polarity, decision, seen                                            []bool
	activity                                                            []float64
	watches, watchesBin                                                 [][]Watcher
	trail                                                               []Lit
	trailLim                                                            []int
	qhead                                                               int
	assumptions                                                         []Lit
	orderHeap                                                           *Heap
	allocator                                                           ClauseAllocator
	ok                                                                  bool
	varInc, clauseInc                                                   float64
	incremental                                                         bool
	nbVarsInitialFormula                                                int
	lbdQueue, trailQueue                                                *BoundedQueue
	sumLBD                                                              float64
	nextReduce, curRestart                                              int
	simpAssigns, simpProps                                              int
	removeSatisfied                                                     bool
	conflictBudget, propagationBudget                                   int64
	runtimeLimit                                                        int64
	interrupted                                                         atomic.Bool
	analyzeToClear                                                      []Lit
}

func NewSolver() *Solver {
	s := &Solver{K: .8, R: 1.4, SizeLBDQueue: 50, SizeTrailQueue: 5000, FirstReduceDB: 2000, IncReduceDB: 300, SpecialIncReduceDB: 1000, LBLBDFrozenClause: 30, LBSizeMinimizingClause: 30, LBLBDMinimizingClause: 6, VarDecay: .8, ClauseDecay: .999, RandomSeed: 91648253, CCMinMode: 2, PhaseSaving: 2, ok: true, varInc: 1, clauseInc: 1, nbVarsInitialFormula: 1<<31 - 1, nextReduce: 2000, curRestart: 1, simpAssigns: -1, removeSatisfied: true, conflictBudget: -1, propagationBudget: -1}
	s.orderHeap = NewHeap(func(a, b int) bool { return s.activity[a] > s.activity[b] })
	s.lbdQueue = NewBoundedQueue(s.SizeLBDQueue)
	s.trailQueue = NewBoundedQueue(s.SizeTrailQueue)
	return s
}
func (s *Solver) NVars() int              { return len(s.assigns) }
func (s *Solver) NClauses() int           { return len(s.clauses) }
func (s *Solver) NLearnts() int           { return len(s.learnts) }
func (s *Solver) Okay() bool              { return s.ok }
func (s *Solver) DecisionLevel() int      { return len(s.trailLim) }
func (s *Solver) ValueVar(v Var) LBool    { return s.assigns[v] }
func (s *Solver) ValueLit(p Lit) LBool    { return s.assigns[p.Var()].Xor(p.Sign()) }
func (s *Solver) ModelValue(p Lit) LBool  { return s.Model[p.Var()].Xor(p.Sign()) }
func (s *Solver) SetIncrementalMode()     { s.incremental = true }
func (s *Solver) InitNbInitialVars(n int) { s.nbVarsInitialFormula = n }
func (s *Solver) isSelector(v Var) bool   { return s.incremental && int(v) > s.nbVarsInitialFormula }
func (s *Solver) NewVar(sign, decision bool) Var {
	v := Var(len(s.assigns))
	s.assigns = append(s.assigns, LUndef)
	s.levels = append(s.levels, 0)
	s.reasons = append(s.reasons, nil)
	s.polarity = append(s.polarity, sign)
	s.decision = append(s.decision, decision)
	s.seen = append(s.seen, false)
	a := 0.
	if s.RandomInitialActivity {
		a = s.random() * 1e-5
	}
	s.activity = append(s.activity, a)
	s.watches = append(s.watches, nil, nil)
	s.watchesBin = append(s.watchesBin, nil, nil)
	if decision {
		s.orderHeap.Insert(int(v))
	}
	if s.JustificationMode > 0 {
		s.ensureCircuit()
	}
	return v
}
func (s *Solver) SetDecisionVar(v Var, b bool) {
	s.decision[v] = b
	if b && !s.orderHeap.InHeap(int(v)) {
		s.orderHeap.Insert(int(v))
	}
}
func (s *Solver) SetPolarity(v Var, b bool) { s.polarity[v] = b }
func (s *Solver) SetConfBudget(n int64)     { s.conflictBudget = s.Conflicts + n }
func (s *Solver) SetPropBudget(n int64)     { s.propagationBudget = s.Propagations + n }
func (s *Solver) BudgetOff()                { s.conflictBudget = -1; s.propagationBudget = -1 }
func (s *Solver) Interrupt()                { s.interrupted.Store(true) }
func (s *Solver) ClearInterrupt()           { s.interrupted.Store(false) }
func (s *Solver) WithinBudget() bool {
	return !s.interrupted.Load() && !s.TerminateSearchEarly && (s.Stop == nil || !s.Stop()) && (s.Deadline.IsZero() || time.Now().Before(s.Deadline)) && (s.conflictBudget < 0 || s.Conflicts < s.conflictBudget) && (s.propagationBudget < 0 || s.Propagations < s.propagationBudget)
}
func (s *Solver) random() float64 {
	s.RandomSeed *= 1389796
	q := int64(s.RandomSeed / 2147483647)
	s.RandomSeed -= float64(q) * 2147483647
	return s.RandomSeed / 2147483647
}
func (s *Solver) newDecisionLevel() { s.trailLim = append(s.trailLim, len(s.trail)) }
func (s *Solver) uncheckedEnqueue(p Lit, c *Clause) {
	s.assigns[p.Var()] = LTrue.Xor(p.Sign())
	s.levels[p.Var()] = s.DecisionLevel()
	s.reasons[p.Var()] = c
	s.trail = append(s.trail, p)
	if s.JustificationMode > 0 {
		c := s.ensureCircuit()
		c.data[p.Var()] = JustData{Now: s.isJReason(p.Var())}
		if s.isJReason(p.Var()) {
			f := c.fanins[p.Var()]
			if s.ValueVar(f[0].Var()).IsUndef() && s.ValueVar(f[1].Var()).IsUndef() {
				c.stack = append(c.stack, p.Var())
			}
		}
	}
}
func (s *Solver) enqueue(p Lit, c *Clause) bool {
	v := s.ValueLit(p)
	if !v.IsUndef() {
		return v == LTrue
	}
	s.uncheckedEnqueue(p, c)
	return true
}
func (s *Solver) CancelUntil(level int) {
	if s.DecisionLevel() <= level {
		return
	}
	end := s.trailLim[level]
	last := s.trailLim[len(s.trailLim)-1]
	for i := len(s.trail) - 1; i >= end; i-- {
		v := s.trail[i].Var()
		s.assigns[v] = LUndef
		if s.PhaseSaving > 1 || (s.PhaseSaving == 1 && i > last) {
			s.polarity[v] = s.trail[i].Sign()
		}
		if s.decision[v] && !s.orderHeap.InHeap(int(v)) {
			s.orderHeap.Insert(int(v))
		}
	}
	if s.JustificationMode > 0 && s.circuit != nil {
		for i := len(s.trail) - 1; i >= end; i-- {
			v := s.trail[i].Var()
			s.gateClearJwatch(v, level)
			s.circuit.data[v].Now = false
		}
		s.circuit.stack = s.circuit.stack[:0]
	}
	s.qhead = end
	s.trail = s.trail[:end]
	s.trailLim = s.trailLim[:level]
}
func (s *Solver) bumpVar(v Var) {
	s.activity[v] += s.varInc
	if s.activity[v] > 1e100 {
		for i := range s.activity {
			s.activity[i] *= 1e-100
		}
		if s.JustificationMode > 0 && s.circuit != nil {
			c := s.circuit
			for _, v := range c.heap.nodes {
				f := c.fanins[v]
				a, b := s.activity[f[0].Var()], s.activity[f[1].Var()]
				if b > a {
					a = b
				}
				c.data[v].ActiveFanin = a
			}
		}
		s.varInc *= 1e-100
	}
	if s.orderHeap.InHeap(int(v)) {
		s.orderHeap.Decrease(int(v))
	}
	if s.JustificationMode > 0 {
		s.updateJustActivity(v)
	}
}
func (s *Solver) bumpClause(c *Clause) {
	c.activity += s.clauseInc
	if c.activity > 1e20 {
		for _, a := range s.learnts {
			a.activity *= 1e-20
		}
		s.clauseInc *= 1e-20
	}
}

func (s *Solver) AddVar(v Var) {
	for s.NVars() <= int(v) {
		s.NewVar(true, true)
	}
}
func (s *Solver) AddEmptyClause() bool { return s.AddClause(nil) }
func (s *Solver) NAssigns() int        { return len(s.trail) }
func (s *Solver) NFreeVars() int {
	n := 0
	for v, b := range s.decision {
		if b && s.assigns[v].IsUndef() {
			n++
		}
	}
	return n
}
func (s *Solver) Level(v Var) int           { return s.levels[v] }
func (s *Solver) ModelValueVar(v Var) LBool { return s.Model[v] }
func (s *Solver) GetCex() []int {
	result := make([]int, len(s.Model))
	for i, v := range s.Model {
		if v == LTrue {
			result[i] = 1
		}
	}
	return result
}

// GarbageCollect drops dead watch entries and allocator ownership. Live clauses
// keep stable addresses; Go's collector reclaims storage no longer referenced.
func (s *Solver) GarbageCollect() {
	clean := func(ws []Watcher) []Watcher {
		dst := ws[:0]
		for _, w := range ws {
			if !w.Clause.freed {
				dst = append(dst, w)
			}
		}
		clear(ws[len(dst):])
		return dst
	}
	for i := range s.watches {
		s.watches[i] = clean(s.watches[i])
		s.watchesBin[i] = clean(s.watchesBin[i])
	}
	live := s.allocator.clauses[:0]
	for _, c := range s.allocator.clauses {
		if !c.freed {
			live = append(live, c)
		}
	}
	clear(s.allocator.clauses[len(live):])
	s.allocator.clauses = live
	s.allocator.wasted = 0
}
func (s *Solver) Reset() {
	// Preserve caller configuration while restoring original reset defaults.
	s.K = .8
	s.R = 1.4
	s.FirstReduceDB = 2000
	s.VarDecay = .8
	s.Model = nil
	s.JustModel = nil
	s.circuit = nil
	s.Conflict = nil
	s.clauses = nil
	s.learnts = nil
	s.assigns = nil
	s.levels = nil
	s.reasons = nil
	s.polarity = nil
	s.decision = nil
	s.seen = nil
	s.activity = nil
	s.watches = nil
	s.watchesBin = nil
	s.trail = nil
	s.trailLim = nil
	s.qhead = 0
	s.assumptions = nil
	s.allocator = ClauseAllocator{}
	s.analyzeToClear = nil
	s.ok = true
	s.varInc = 1
	s.clauseInc = 1
	s.nbVarsInitialFormula = 1<<31 - 1
	s.sumLBD = 0
	s.nextReduce = s.FirstReduceDB
	s.curRestart = 1
	s.simpAssigns = -1
	s.simpProps = 0
	s.BudgetOff()
	s.ClearInterrupt()
	s.Solves = 0
	s.Starts = 0
	s.Decisions = 0
	s.RandomDecisions = 0
	s.Propagations = 0
	s.Conflicts = 0
	s.ConflictsRestarts = 0
	s.RemovedClauses = 0
	s.ReducedClauses = 0
	s.ReduceDBs = 0
	s.orderHeap = NewHeap(func(a, b int) bool { return s.activity[a] > s.activity[b] })
	s.lbdQueue = NewBoundedQueue(s.SizeLBDQueue)
	s.trailQueue = NewBoundedQueue(s.SizeTrailQueue)
}
