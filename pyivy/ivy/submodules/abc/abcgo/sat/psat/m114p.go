// Package psat implements the proof-producing solver contract in m114p.h.
// The external MiniSat-1.14p implementation is absent from this checkout; this
// Go implementation produces independently replayable resolution proofs.
package psat

import (
	"fmt"
	"os"
)

type Solver struct {
	roots                   [][]int
	variables               int
	recordProof             bool
	deleted                 bool
	proof                   *Proof
	model                   []int
	conflicts               int
	rootCursor, chainCursor int
}

func New(recordProof bool) *Solver { return &Solver{recordProof: recordProof} }
func (s *Solver) checkOpen() {
	if s == nil || s.deleted {
		panic("invalid or deleted PSAT solver")
	}
}
func (s *Solver) SetVariableCount(n int) {
	s.checkOpen()
	if n < 0 {
		panic("negative variable count")
	}
	if n > s.variables {
		s.variables = n
	}
	s.model = nil
}
func (s *Solver) VariableCount() int { s.checkOpen(); return s.variables }
func (s *Solver) AddClause(lits []int) bool {
	s.checkOpen()
	for _, lit := range lits {
		if lit < 0 {
			panic("negative encoded literal")
		}
	}
	normalized, _ := normalize(lits)
	for _, lit := range normalized {
		if n := (lit >> 1) + 1; n > s.variables {
			s.variables = n
		}
	}
	s.roots = append(s.roots, normalized)
	s.proof = nil
	s.model = nil
	s.rootCursor = 0
	s.chainCursor = 0
	return s.Simplify()
}
func (s *Solver) Simplify() bool {
	s.checkOpen()
	values := make([]int8, s.variables)
	for i := range values {
		values[i] = -1
	}
	changed := true
	for changed {
		changed = false
		for _, clause := range s.roots {
			if tautology(clause) {
				continue
			}
			open := 0
			last := 0
			satisfied := false
			for _, lit := range clause {
				value := values[lit>>1]
				if value < 0 {
					open++
					last = lit
				} else if int(value)^(lit&1) == 1 {
					satisfied = true
					break
				}
			}
			if satisfied {
				continue
			}
			if open == 0 {
				return false
			}
			if open == 1 {
				v := last >> 1
				want := int8(1 - (last & 1))
				if values[v] >= 0 && values[v] != want {
					return false
				}
				if values[v] < 0 {
					values[v] = want
					changed = true
				}
			}
		}
	}
	return true
}
func (s *Solver) ModelValue(v int) int {
	if v < 0 || v >= len(s.model) {
		return -1
	}
	return s.model[v]
}
func (s *Solver) ConflictCount() int { s.checkOpen(); return s.conflicts }
func (s *Solver) ProofReady() bool   { return s != nil && !s.deleted && s.recordProof && s.proof != nil }
func (s *Solver) Proof() *Proof {
	if !s.ProofReady() {
		return nil
	}
	return s.proof.Clone()
}
func (s *Solver) SaveProof(path string) error {
	if !s.ProofReady() {
		return fmt.Errorf("no UNSAT proof available")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err = s.proof.Write(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
func (s *Solver) PrintStats(seconds float64) {
	fmt.Printf("PSAT: %d variables, %d roots, %d resolution conflicts, %.3f seconds\n", s.variables, len(s.roots), s.conflicts, seconds)
}
