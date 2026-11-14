// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

func (s *Solver) AddVariable(negative bool) int {
	v := len(s.values)
	s.values = append(s.values, -1)
	s.levels = append(s.levels, 0)
	s.reasons = append(s.reasons, -1)
	phase := uint8(0)
	if negative {
		phase = 1
	}
	s.phases = append(s.phases, phase)
	s.decision = append(s.decision, true)
	s.marks = append(s.marks, false)
	s.watches = append(s.watches, nil, nil)
	s.activity.Push(0)
	s.order.Insert(uint32(v))
	return v
}
func (s *Solver) SetVariables(n int) {
	if n < 0 {
		panic("negative variable count")
	}
	for len(s.values) < n {
		s.AddVariable(false)
	}
}
func (s *Solver) AddClause(lits []int) bool {
	if !s.ok {
		return false
	}
	maxVariable := -1
	for _, lit := range lits {
		if lit < 0 {
			panic("negative encoded literal")
		}
		if lit>>1 > maxVariable {
			maxVariable = lit >> 1
		}
	}
	s.cancel(0)
	for maxVariable >= len(s.values) {
		s.AddVariable(true)
	}
	copyLits := append([]int(nil), lits...)
	sort.Ints(copyLits)
	out := copyLits[:0]
	last := -2
	for _, lit := range copyLits {
		if s.literalValue(lit) == 1 {
			return true
		}
		if s.literalValue(lit) == 0 {
			continue
		}
		if lit == last {
			continue
		}
		if lit == last^1 {
			return true
		}
		out = append(out, lit)
		last = lit
	}
	s.model = nil
	id := s.addStored(out, false, 0)
	if len(out) == 0 {
		s.ok = false
		return false
	}
	if len(out) == 1 && (!s.enqueue(out[0], id) || s.propagate() >= 0) {
		s.ok = false
		return false
	}
	return true
}
func (s *Solver) PushAssumption(lit int) {
	if lit < 0 || lit>>1 >= len(s.values) {
		panic("assumption variable out of range")
	}
	s.assumptions = append(s.assumptions, lit)
	s.phases[lit>>1] = uint8(lit & 1)
}
func (s *Solver) PopAssumption() {
	if len(s.assumptions) == 0 {
		panic("pop empty assumptions")
	}
	s.assumptions = s.assumptions[:len(s.assumptions)-1]
	s.cancel(0)
}
func (s *Solver) SolveAssumptions(lits []int) int {
	for _, lit := range lits {
		if lit < 0 || lit>>1 >= len(s.values) {
			panic("assumption variable out of range")
		}
	}
	old := s.assumptions
	s.assumptions = append([]int(nil), lits...)
	defer func() { s.assumptions = old }()
	return s.Solve()
}
func (s *Solver) SolveAssumptionsLimit(lits []int, limit int64) int {
	old := s.opts.ConflictLimit
	s.opts.ConflictLimit = limit
	defer func() { s.opts.ConflictLimit = old }()
	return s.SolveAssumptions(lits)
}
func (s *Solver) MinimizeAssumptions(lits []int, limit int64) ([]int, int) {
	out := append([]int(nil), lits...)
	result := s.SolveAssumptionsLimit(out, limit)
	if result != SATOKO_UNSAT {
		return out, result
	}
	for i := 0; i < len(out); {
		trial := append([]int(nil), out[:i]...)
		trial = append(trial, out[i+1:]...)
		if s.SolveAssumptionsLimit(trial, limit) == SATOKO_UNSAT {
			out = trial
		} else {
			i++
		}
	}
	s.final = make([]int, len(out))
	for i, lit := range out {
		s.final[i] = lit ^ 1
	}
	return out, SATOKO_UNSAT
}

// MinimizeAssumptionsSource ports satoko_minimize_assumptions from
// solver_api.c. It uses the source's recursive divide-and-conquer reduction,
// mutates lits in place, and preserves any assumptions already pushed on s.
func (s *Solver) MinimizeAssumptionsSource(lits []int, limit int64) int {
	if len(lits) == 0 {
		return 0
	}
	oldLimit := s.opts.ConflictLimit
	defer func() { s.opts.ConflictLimit = oldLimit }()
	s.opts.ConflictLimit = limit
	var minimize func(offset, count int) int
	minimize = func(offset, count int) int {
		if count == 1 {
			if s.Solve() != SATOKO_UNSAT {
				return 1
			}
			return 0
		}
		left := count / 2
		right := count - left
		for i := 0; i < left; i++ {
			s.PushAssumption(lits[offset+i])
		}
		if s.Solve() == SATOKO_UNSAT {
			for i := 0; i < left; i++ {
				s.PopAssumption()
			}
			return minimize(offset, left)
		}
		keptRight := 1
		if right != 1 {
			keptRight = minimize(offset+left, right)
		}
		for i := 0; i < left; i++ {
			s.PopAssumption()
		}

		leftValues := append([]int(nil), lits[offset:offset+left]...)
		copy(lits[offset:offset+keptRight], lits[offset+left:offset+left+keptRight])
		copy(lits[offset+keptRight:offset+keptRight+left], leftValues)
		for i := 0; i < keptRight; i++ {
			s.PushAssumption(lits[offset+i])
		}
		if s.Solve() == SATOKO_UNSAT {
			for i := 0; i < keptRight; i++ {
				s.PopAssumption()
			}
			return keptRight
		}
		keptLeft := 1
		if left != 1 {
			keptLeft = minimize(offset+keptRight, left)
		}
		for i := 0; i < keptRight; i++ {
			s.PopAssumption()
		}
		return keptRight + keptLeft
	}
	return minimize(0, len(lits))
}
func (s *Solver) FinalConflict() []int { return append([]int(nil), s.final...) }
func (s *Solver) ModelValue(v int) int {
	if v < 0 || v >= len(s.model) {
		return -1
	}
	return s.model[v]
}
func (s *Solver) VariableCount() int { return len(s.values) }
func (s *Solver) ClauseCount() int {
	n := 0
	for _, c := range s.clauses {
		if !c.deleted && !c.learnt {
			n++
		}
	}
	return n
}
func (s *Solver) LearntCount() int {
	n := 0
	for _, c := range s.clauses {
		if !c.deleted && c.learnt {
			n++
		}
	}
	return n
}
func (s *Solver) Bookmark() {
	s.cancel(0)
	s.bookmark = &solverBookmark{len(s.clauses), len(s.values), s.ok}
}
func (s *Solver) Unbookmark() { s.bookmark = nil }
func (s *Solver) Rollback() {
	if s.bookmark == nil {
		panic("rollback without bookmark")
	}
	b := s.bookmark
	s.clauses = s.clauses[:b.clauses]
	s.values = s.values[:b.variables]
	s.phases = s.phases[:b.variables]
	s.decision = s.decision[:b.variables]
	s.levels = s.levels[:b.variables]
	s.reasons = s.reasons[:b.variables]
	s.marks = s.marks[:b.variables]
	s.activity.Resize(uint32(b.variables))
	s.ok = b.ok
	s.assumptions = nil
	s.model = nil
	s.final = nil
	s.rebuildWatches()
	s.stats.OriginalLiterals = 0
	s.stats.LearntLiterals = 0
	for _, c := range s.clauses {
		if c.deleted {
			continue
		}
		if c.learnt {
			s.stats.LearntLiterals += int64(len(c.lits))
		} else {
			s.stats.OriginalLiterals += int64(len(c.lits))
		}
	}
	s.prepare()
}
func (s *Solver) Reset() {
	opts := s.opts
	stop, stopFunc, runID, deadline := s.stop, s.stopFunc, s.runID, s.deadline
	*s = *New()
	s.opts = opts
	s.stop, s.stopFunc, s.runID, s.deadline = stop, stopFunc, runID, deadline
}
func (s *Solver) MarkCone(vars []int, marked bool) {
	for _, v := range vars {
		if v < 0 || v >= len(s.marks) {
			panic("cone variable out of range")
		}
	}
	s.coneActive = true
	for _, v := range vars {
		s.marks[v] = marked
		if marked {
			s.activity.Set(uint32(v), 0)
			s.order.Update(uint32(v))
		}
	}
}
func (s *Solver) WriteDIMACS(w io.Writer, learnts, zeroBased bool) error {
	clauses := 0
	for _, c := range s.clauses {
		if !c.deleted && (!c.learnt || learnts) {
			clauses++
		}
	}
	if _, err := fmt.Fprintf(w, "p cnf %d %d\n", len(s.values), clauses); err != nil {
		return err
	}
	for _, c := range s.clauses {
		if c.deleted || c.learnt && !learnts {
			continue
		}
		for _, lit := range c.lits {
			v := lit >> 1
			if !zeroBased {
				v++
			}
			if lit&1 != 0 {
				v = -v
			}
			if _, err := fmt.Fprintf(w, "%d ", v); err != nil {
				return err
			}
		}
		terminator := "0"
		if zeroBased {
			terminator = ""
		}
		if _, err := fmt.Fprintln(w, terminator); err != nil {
			return err
		}
	}
	return nil
}
func satoko_setnvars(s *Solver, n int)             { s.SetVariables(n) }
func satoko_add_variable(s *Solver, sign byte) int { return s.AddVariable(sign != 0) }
func satoko_add_clause(s *Solver, lits []int, n int) int {
	if s.AddClause(lits[:n]) {
		return SATOKO_OK
	}
	return SATOKO_ERR
}
func satoko_assump_push(s *Solver, lit int)                     { s.PushAssumption(lit) }
func satoko_assump_pop(s *Solver)                               { s.PopAssumption() }
func satoko_solve(s *Solver) int                                { return s.Solve() }
func satoko_solve_assumptions(s *Solver, lits []int, n int) int { return s.SolveAssumptions(lits[:n]) }
func satoko_solve_assumptions_limit(s *Solver, lits []int, n, limit int) int {
	return s.SolveAssumptionsLimit(lits[:n], int64(limit))
}
func satoko_minimize_assumptions(s *Solver, lits []int, n, limit int) int {
	out, result := s.MinimizeAssumptions(lits[:n], int64(limit))
	if result != SATOKO_UNSAT {
		return -1
	}
	copy(lits, out)
	return len(out)
}
func satoko_final_conflict(s *Solver, out *[]int) int {
	*out = s.FinalConflict()
	if len(*out) == 0 {
		return -1
	}
	return len(*out)
}
func satoko_stats(s *Solver) *Stats                    { return &s.stats }
func satoko_options(s *Solver) *Options                { return &s.opts }
func satoko_bookmark(s *Solver)                        { s.Bookmark() }
func satoko_unbookmark(s *Solver)                      { s.Unbookmark() }
func satoko_rollback(s *Solver)                        { s.Rollback() }
func satoko_reset(s *Solver)                           { s.Reset() }
func satoko_mark_cone(s *Solver, vars []int, n int)    { s.MarkCone(vars[:n], true) }
func satoko_unmark_cone(s *Solver, vars []int, n int)  { s.MarkCone(vars[:n], false) }
func satoko_varnum(s *Solver) int                      { return s.VariableCount() }
func satoko_clausenum(s *Solver) int                   { return s.ClauseCount() }
func satoko_learntnum(s *Solver) int                   { return s.LearntCount() }
func satoko_conflictnum(s *Solver) int                 { return int(s.stats.AllConflicts) }
func satoko_read_cex_varvalue(s *Solver, v int) int    { return s.ModelValue(v) }
func satoko_set_stop(s *Solver, stop *int)             { s.stop = stop }
func satoko_set_stop_func(s *Solver, fn func(int) int) { s.stopFunc = fn }
func satoko_set_runid(s *Solver, id int)               { s.runID = id }
func satoko_var_polarity(s *Solver, v uint32) byte     { return s.phases[v] }
func satoko_set_runtime_limit(s *Solver, limit int64) int64 {
	old := int64(0)
	if !s.deadline.IsZero() {
		old = s.deadline.UnixNano()
	}
	if limit == 0 {
		s.deadline = time.Time{}
	} else {
		s.deadline = time.Unix(0, limit)
	}
	return old
}
func satoko_simplify(s *Solver) int {
	if s.prepare() {
		return SATOKO_OK
	}
	s.ok = false
	return SATOKO_ERR
}
func satoko_write_dimacs(s *Solver, path string, learnt, zero int) {
	if path == "" {
		if err := s.WriteDIMACS(os.Stdout, learnt != 0, zero != 0); err != nil {
			panic(err)
		}
		return
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = s.WriteDIMACS(f, learnt != 0, zero != 0); err != nil {
		panic(err)
	}
}

// SetStopPointer supports a synchronous caller-owned cancellation flag. Use
// SetStopFunc with an atomic load for cross-goroutine cancellation.
func (s *Solver) SetStopPointer(flag *int)       { s.stop = flag }
func (s *Solver) SetStopFunc(fn func(int) int)   { s.stopFunc = fn }
func (s *Solver) SetRunID(id int)                { s.runID = id }
func (s *Solver) SetDeadline(deadline time.Time) { s.deadline = deadline }
func (s *Solver) Simplify() bool {
	if !s.ok {
		return false
	}
	s.cancel(0)
	if s.propagate() >= 0 {
		s.ok = false
		return false
	}
	s.simplifyOrder()
	return true
}

// SetPolarity chooses the saved sign for the next decision on v.
func (s *Solver) SetPolarity(v int, negative bool) {
	if v < 0 || v >= len(s.values) {
		panic("variable out of range")
	}
	if negative {
		s.phases[v] = 1
	} else {
		s.phases[v] = 0
	}
}

// SetDecisionVariable controls whether v can be chosen as a branch variable;
// disabling decisions does not suppress implications involving v.
func (s *Solver) SetDecisionVariable(v int, enabled bool) {
	if v < 0 || v >= len(s.values) {
		panic("variable out of range")
	}
	s.decision[v] = enabled
	if enabled {
		s.order.Update(uint32(v))
	}
}
func (s *Solver) CurrentValue(v int) int {
	if v < 0 || v >= len(s.values) {
		return -1
	}
	return int(s.values[v])
}
func (s *Solver) Okay() bool         { return s.ok }
func (s *Solver) AssignedCount() int { return len(s.trail) }
func (s *Solver) FreeVariableCount() int {
	n := 0
	for v, value := range s.values {
		if value < 0 && s.decision[v] {
			n++
		}
	}
	return n
}
