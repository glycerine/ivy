package bsat

import (
	"fmt"
	"io"
	"time"
)

type proofAnnotation interface {
	root(*Sto_Cls_t)
	begin(*Sto_Cls_t, *Sto_Cls_t)
	resolve(*Sto_Cls_t, *Sto_Cls_t, lit)
}
type resolutionTrace struct {
	cnf          *Sto_Man_t
	assignments  satAssignments
	memory       *Sat_Mem_t
	watches      []veci
	stored       map[cla]*Sto_Cls_t
	labels       []int32
	dependencies [][]int32
	counter      int32
	observer     proofAnnotation
	writer       io.Writer
	deadline     int64
	clock        func() int64
	stats        stats_t
}

func (r *resolutionTrace) expired() bool {
	if r.deadline == 0 {
		return false
	}
	if r.clock != nil {
		return r.clock() > r.deadline
	}
	return time.Now().UnixNano() > r.deadline
}
func (r *resolutionTrace) writeRoot(c *Sto_Cls_t) error {
	r.counter++
	r.labels[c.Id] = r.counter
	if r.observer != nil {
		r.observer.root(c)
	}
	if r.writer == nil {
		return nil
	}
	if _, err := fmt.Fprint(r.writer, r.counter); err != nil {
		return err
	}
	for _, l := range c.pLits {
		if _, err := fmt.Fprintf(r.writer, " %d", lit_print(l)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(r.writer, " 0 0")
	return err
}
func (r *resolutionTrace) add(c *Sto_Cls_t) cla {
	// Reconstruction uses even opaque handles, not packed memory pages.
	// Stored proof clauses therefore do not inherit a page-size limit.
	if len(r.stored) >= 0x3fffffff {
		panic("proof clause handle overflow")
	}
	h := int32(len(r.stored)+1) * 2
	r.memory.clauses[h] = &clause{size: c.nLits, lits: append([]lit(nil), c.pLits...)}
	r.stored[h] = c
	if c.nLits >= 2 {
		veci_push(&r.watches[lit_neg(c.pLits[0])], h)
		veci_push(&r.watches[lit_neg(c.pLits[1])], h)
	}
	return h
}
func (r *resolutionTrace) propagate() cla {
	return satPropagate(&r.assignments, r.memory, r.watches, 0, nil, nil, &r.stats)
}
func (r *resolutionTrace) cancel(bound int32) {
	a := &r.assignments
	for i := a.qtail - 1; i >= bound; i-- {
		v := lit_var(a.trail[i])
		a.assigns[v] = satUnassigned
		a.reasons[v] = 0
	}
	a.qhead = bound
	a.qtail = bound
}
func resolutionStep(left, right []lit, pivot lit) ([]lit, error) {
	foundLeft, foundRight := false, false
	seen := map[lit]bool{}
	out := make([]lit, 0, len(left)+len(right)-2)
	for _, l := range left {
		if l == lit_neg(pivot) {
			foundLeft = true
			continue
		}
		seen[l] = true
		out = append(out, l)
	}
	for _, l := range right {
		if l == pivot {
			foundRight = true
			continue
		}
		if seen[lit_neg(l)] {
			return nil, fmt.Errorf("resolution has extra complementary literal %d", l)
		}
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	if !foundLeft || !foundRight {
		return nil, fmt.Errorf("resolution pivot %d missing or wrong polarity", pivot)
	}
	return out, nil
}
func (r *resolutionTrace) trace(conflict cla, final *Sto_Cls_t) error {
	source := r.stored[conflict]
	if source == nil {
		return fmt.Errorf("conflict lacks stored provenance")
	}
	resolvent := append([]lit(nil), clause_begin(Sat_MemClauseHand(r.memory, conflict))...)
	seen := make([]bool, r.cnf.nVars)
	for _, l := range resolvent {
		seen[lit_var(l)] = true
	}
	if r.observer != nil {
		r.observer.begin(final, source)
	}
	parents := []int32{source.Id}
	previous := r.labels[source.Id]
	for i := r.assignments.qtail - 1; i >= 0; i-- {
		pivot := r.assignments.trail[i]
		v := lit_var(pivot)
		if !seen[v] {
			continue
		}
		seen[v] = false
		h := r.assignments.reasons[v]
		if h == 0 {
			continue
		}
		reason := r.stored[h]
		if reason == nil {
			return fmt.Errorf("reason lacks stored provenance")
		}
		literals := clause_begin(Sat_MemClauseHand(r.memory, h))
		if len(literals) == 0 || literals[0] != pivot {
			return fmt.Errorf("reason's first literal is not its implication")
		}
		var err error
		resolvent, err = resolutionStep(resolvent, literals, pivot)
		if err != nil {
			return err
		}
		for _, l := range literals[1:] {
			seen[lit_var(l)] = true
		}
		if r.observer != nil {
			r.observer.resolve(final, reason, pivot)
		}
		parents = append(parents, reason.Id)
		r.counter++
		if r.writer != nil {
			if _, err := fmt.Fprintf(r.writer, "%d * %d %d 0\n", r.counter, previous, r.labels[reason.Id]); err != nil {
				return err
			}
		}
		previous = r.counter
	}
	allowed := map[lit]bool{}
	for _, l := range final.pLits {
		allowed[l] = true
	}
	for _, l := range resolvent {
		if !allowed[l] {
			return fmt.Errorf("derived resolvent is not a subset of clause %d", final.Id)
		}
	}
	// The original verifier strengthens a weaker logged clause to its resolvent.
	final.pLits = append(final.pLits[:0], resolvent...)
	final.nLits = int32(len(resolvent))
	r.labels[final.Id] = previous
	r.dependencies[final.Id] = parents
	return nil
}
func reconstructResolution(cnf *Sto_Man_t, observer proofAnnotation, writer io.Writer, deadline int64) (*resolutionTrace, error) {
	return reconstructResolutionWithClock(cnf, observer, writer, deadline, nil)
}

func reconstructResolutionWithClock(cnf *Sto_Man_t, observer proofAnnotation, writer io.Writer, deadline int64, clock func() int64) (*resolutionTrace, error) {
	if cnf == nil || cnf.pHead == nil || cnf.pTail == nil || cnf.pEmpty == nil || cnf.pTail != cnf.pEmpty || cnf.nRoots > cnf.nClauses || cnf.nClausesA > cnf.nRoots {
		return nil, fmt.Errorf("proof store lacks valid roots/final empty clause")
	}
	r := &resolutionTrace{cnf: cnf, assignments: newSatAssignments(cnf.nVars), memory: &Sat_Mem_t{clauses: make(map[cla]*clause)}, watches: make([]veci, 2*cnf.nVars), stored: make(map[cla]*Sto_Cls_t), labels: make([]int32, cnf.nClauses), dependencies: make([][]int32, cnf.nClauses), observer: observer, writer: writer, deadline: deadline, clock: clock}
	index := int32(0)
	for c := cnf.pHead; c != nil; c = c.pNext {
		if c.Id != index || c.nLits != int32(len(c.pLits)) || (c.fRoot != 0) != (index < cnf.nRoots) || (c.fA != 0) != (index < cnf.nClausesA) {
			return nil, fmt.Errorf("inconsistent clause metadata at %d", index)
		}
		index++
	}
	if index != cnf.nClauses {
		return nil, fmt.Errorf("clause count mismatch")
	}
	for c := cnf.pHead; c != nil && c.fRoot != 0; c = c.pNext {
		if err := r.writeRoot(c); err != nil {
			return nil, err
		}
	}
	for c := cnf.pHead; c != nil && c.fRoot != 0; c = c.pNext {
		if r.expired() {
			return nil, fmt.Errorf("proof deadline exceeded")
		}
		if c.nLits == 0 {
			if observer != nil {
				observer.begin(cnf.pEmpty, c)
			}
			r.dependencies[cnf.pEmpty.Id] = []int32{c.Id}
			return r, nil
		}
		h := r.add(c)
		if c.nLits == 1 && !r.assignments.enqueue(c.pLits[0], h) {
			return r, r.trace(h, cnf.pEmpty)
		}
	}
	if h := r.propagate(); h != 0 {
		return r, r.trace(h, cnf.pEmpty)
	}
	for c := cnf.pHead; c != nil; c = c.pNext {
		if c.fRoot != 0 {
			continue
		}
		if r.expired() {
			return nil, fmt.Errorf("proof deadline exceeded")
		}
		bound := r.assignments.qtail
		satisfied := false
		for _, l := range c.pLits {
			if r.assignments.assigns[lit_var(l)] == lit_sign(l) {
				satisfied = true
				break
			}
		}
		if satisfied {
			continue
		}
		for _, l := range c.pLits {
			if !r.assignments.enqueue(lit_neg(l), 0) {
				return nil, fmt.Errorf("inconsistent RUP assumption")
			}
		}
		conflict := r.propagate()
		if conflict == 0 {
			return nil, fmt.Errorf("clause %d is not reverse-unit-propagation derivable", c.Id)
		}
		if err := r.trace(conflict, c); err != nil {
			return nil, err
		}
		r.cancel(bound)
		if c.nLits == 0 {
			if c != cnf.pEmpty {
				if observer != nil {
					observer.begin(cnf.pEmpty, c)
				}
				r.dependencies[cnf.pEmpty.Id] = []int32{c.Id}
				r.labels[cnf.pEmpty.Id] = r.labels[c.Id]
			}
			return r, nil
		}
		h := r.add(c)
		if c.nLits == 1 {
			if !r.assignments.enqueue(c.pLits[0], h) {
				return r, r.trace(h, cnf.pEmpty)
			}
			if conflict := r.propagate(); conflict != 0 {
				return r, r.trace(conflict, cnf.pEmpty)
			}
		}
	}
	return nil, fmt.Errorf("proof did not derive its empty clause")
}
