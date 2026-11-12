package bsat

import (
	"math/rand"
	"reflect"
	"testing"
)

func partitionedProof(t *testing.T, a, b [][]lit) *Sto_Man_t {
	t.Helper()
	s := sat_solver_new()
	sat_solver_store_alloc(s)
	consistent := true
	for _, c := range a {
		if sat_solver_addclause(s, c) == 0 {
			consistent = false
		}
	}
	sat_solver_store_mark_clauses_a(s)
	for _, c := range b {
		if sat_solver_addclause(s, c) == 0 {
			consistent = false
		}
	}
	sat_solver_store_mark_roots(s)
	if !consistent {
		s.fSolved = 1
	}
	if sat_solver_solve(s, nil, 0, 0, 0, 0) != l_False {
		return nil
	}
	return solverClauseStore(s)
}
func checkTruthCraig(t *testing.T, p *Sto_Man_t, table []uint32, globals []int32) {
	t.Helper()
	for model := 0; model < 1<<uint(p.nVars); model++ {
		a, b := true, true
		for c := p.pHead; c != nil && c.fRoot != 0; c = c.pNext {
			if !modelSatisfiesClause(model, c.pLits) {
				if c.fA != 0 {
					a = false
				} else {
					b = false
				}
			}
		}
		point := 0
		for bit, v := range globals {
			if model&(1<<uint(v)) != 0 {
				point |= 1 << uint(bit)
			}
		}
		value := table[point/32]&(uint32(1)<<uint(point%32)) != 0
		if a && !value || b && value {
			t.Fatalf("Craig violation at model %d: A=%v B=%v I=%v globals %v", model, a, b, value, globals)
		}
	}
}
func TestTruthInterpolationLocalAndSharedVariables(t *testing.T) {
	p := partitionedProof(t, [][]lit{{0, 2}, {1, 2}}, [][]lit{{3}})
	if p == nil {
		t.Fatal("expected UNSAT")
	}
	table, globals, err := TruthInterpolant(p, nil)
	if err != nil || !reflect.DeepEqual(globals, []int32{1}) || len(table) != 1 || table[0] != 0xaaaaaaaa {
		t.Fatal(table, globals, err)
	}
	checkTruthCraig(t, p, table, globals)
}
func TestTruthInterpolationConstantPartitions(t *testing.T) {
	for _, aOnly := range []bool{false, true} {
		a, b := [][]lit(nil), [][]lit{{0}, {1}}
		if aOnly {
			a, b = b, a
		}
		p := partitionedProof(t, a, b)
		table, globals, err := TruthInterpolant(p, nil)
		if err != nil || len(globals) != 0 || len(table) != 1 {
			t.Fatal(table, globals, err)
		}
		want := ^uint32(0)
		if aOnly {
			want = 0
		}
		if table[0] != want {
			t.Fatal("constant interpolant", table, want)
		}
		checkTruthCraig(t, p, table, globals)
	}
}
func TestTruthInterpolationRandomCraigConditions(t *testing.T) {
	rng := rand.New(rand.NewSource(625))
	checked := 0
	for trial := 0; trial < 100; trial++ {
		var a, b [][]lit
		for k := 0; k < 14; k++ {
			perm := rng.Perm(4)
			c := make([]lit, 1+rng.Intn(3))
			for i := range c {
				c[i] = toLitCond(int32(perm[i]), int32(rng.Intn(2)))
			}
			if k < 7 {
				a = append(a, c)
			} else {
				b = append(b, c)
			}
		}
		p := partitionedProof(t, a, b)
		if p == nil {
			continue
		}
		table, globals, err := TruthInterpolant(p, nil)
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		checkTruthCraig(t, p, table, globals)
		checked++
	}
	if checked < 10 {
		t.Fatal("insufficient UNSAT cases", checked)
	}
}
