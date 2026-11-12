package bsat2

import "sync/atomic"

/****************************************************************************************[Solver.h]
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

type watcher struct {
	ref     CRef
	blocker Lit
}
type variableData struct {
	reason CRef
	level  int
}
type Solver struct {
	VarDecay, ClauseDecay, RandomFrequency, RandomSeed, GarbageFraction, RestartIncrease float64
	LubyRestart, RandomPolarity, RandomInitialActivity                                   bool
	Minimization, PhaseSaving, RestartFirst, Verbosity                                   int
	Model                                                                                []LBool
	Conflict                                                                             []Lit
	Solves, Starts, Decisions, RandomDecisions, Propagations, Conflicts                  uint64
	ClauseLiterals, LearntLiterals, MaxLiterals, TotalLiterals                           uint64
	decisionVariables                                                                    int
	ok                                                                                   bool
	clauseInc, variableInc                                                               float64
	clauses, learnts                                                                     []CRef
	ca                                                                                   *ClauseAllocator
	watches                                                                              *OccLists[watcher]
	activity                                                                             []float64
	assigns                                                                              []LBool
	polarity, decision, seen                                                             []bool
	vardata                                                                              []variableData
	trail                                                                                []Lit
	trailLimits                                                                          []int
	qhead, simplifyAssigns                                                               int
	simplifyProps                                                                        int64
	assumptions                                                                          []Lit
	order                                                                                *Heap
	progress                                                                             float64
	removeSatisfiedOriginal                                                              bool
	analyzeStack, analyzeClear                                                           []Lit
	maxLearnts, learntAdjustConflicts                                                    float64
	learntAdjustCount                                                                    int
	conflictBudget, propagationBudget                                                    int64
	interrupted                                                                          atomic.Bool
	relocateExtra                                                                        func(*ClauseAllocator)
}

func NewSolver() *Solver {
	s := &Solver{VarDecay: 0.95, ClauseDecay: 0.999, RandomSeed: 91648253, GarbageFraction: 0.2, RestartIncrease: 2, LubyRestart: true, Minimization: 2, PhaseSaving: 2, RestartFirst: 100, ok: true, clauseInc: 1, variableInc: 1, simplifyAssigns: -1, removeSatisfiedOriginal: true, conflictBudget: -1, propagationBudget: -1, ca: NewClauseAllocator()}
	s.watches = NewOccLists(s.watcherDeleted)
	s.order = NewHeap(s.variableLess)
	return s
}
func (s *Solver) watcherDeleted(w watcher) bool { return s.ca.At(w.ref).Mark == 1 }
func (s *Solver) variableLess(a, b int32) bool  { return s.activity[a] > s.activity[b] }
func (s *Solver) NumVars() int                  { return len(s.assigns) }
func (s *Solver) NumClauses() int               { return len(s.clauses) }
func (s *Solver) NumLearnts() int               { return len(s.learnts) }
func (s *Solver) Okay() bool                    { return s.ok }
func (s *Solver) NewVar(negative, decision bool) Var {
	v := Var(len(s.assigns))
	s.watches.Init(int(MakeLit(v, true)))
	s.assigns = append(s.assigns, Undef)
	s.vardata = append(s.vardata, variableData{reason: CRefUndef})
	activity := float64(0)
	if s.RandomInitialActivity {
		activity = s.random() * 0.00001
	}
	s.activity = append(s.activity, activity)
	s.seen = append(s.seen, false)
	s.polarity = append(s.polarity, negative)
	s.decision = append(s.decision, false)
	s.SetDecisionVar(v, decision)
	return v
}
func (s *Solver) NewVariable() Var { return s.NewVar(true, true) }
func (s *Solver) SetDecisionVar(v Var, enabled bool) {
	if enabled && !s.decision[v] {
		s.decisionVariables++
	}
	if !enabled && s.decision[v] {
		s.decisionVariables--
	}
	s.decision[v] = enabled
	s.insertOrder(v)
}
func (s *Solver) SetPolarity(v Var, negative bool) { s.polarity[v] = negative }
func (s *Solver) Value(l Lit) LBool                { return s.assigns[l.Var()].Xor(l.Sign()) }
func (s *Solver) ModelValue(l Lit) LBool {
	if int(l.Var()) >= len(s.Model) {
		return Undef
	}
	return s.Model[l.Var()].Xor(l.Sign())
}
func (s *Solver) level() int { return len(s.trailLimits) }
func (s *Solver) insertOrder(v Var) {
	if !s.order.Contains(int32(v)) && s.decision[v] {
		s.order.Insert(int32(v))
	}
}
func (s *Solver) random() float64 {
	s.RandomSeed *= 1389796
	q := int64(s.RandomSeed / 2147483647)
	s.RandomSeed -= float64(q) * 2147483647
	return s.RandomSeed / 2147483647
}
func (s *Solver) bumpVariable(v Var) {
	s.activity[v] += s.variableInc
	if s.activity[v] > 1e100 {
		for i := range s.activity {
			s.activity[i] *= 1e-100
		}
		s.variableInc *= 1e-100
	}
	if s.order.Contains(int32(v)) {
		s.order.Decrease(int32(v))
	}
}
func (s *Solver) bumpClause(c *Clause) {
	c.Activity += float32(s.clauseInc)
	if c.Activity > 1e20 {
		for _, h := range s.learnts {
			s.ca.At(h).Activity *= 1e-20
		}
		s.clauseInc *= 1e-20
	}
}
func (s *Solver) SetConflictBudget(budget int64) { s.conflictBudget = int64(s.Conflicts) + budget }
func (s *Solver) SetPropagationBudget(budget int64) {
	s.propagationBudget = int64(s.Propagations) + budget
}
func (s *Solver) BudgetOff()      { s.conflictBudget, s.propagationBudget = -1, -1 }
func (s *Solver) Interrupt()      { s.interrupted.Store(true) }
func (s *Solver) ClearInterrupt() { s.interrupted.Store(false) }
func (s *Solver) withinBudget() bool {
	return !s.interrupted.Load() && (s.conflictBudget < 0 || s.Conflicts < uint64(s.conflictBudget)) && (s.propagationBudget < 0 || s.Propagations < uint64(s.propagationBudget))
}
