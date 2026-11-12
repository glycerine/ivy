package bsat

import (
	"math/rand"
	"reflect"
	"testing"
)

type activityRecorder struct {
	vars    []int32
	clauses int
}

func (r *activityRecorder) bumpVariable(v int32) { r.vars = append(r.vars, v) }
func (r *activityRecorder) bumpClause(c *clause) { r.clauses++ }
func TestCDCLFirstUIPLearning(t *testing.T) {
	a := newSatAssignments(3)
	m := Sat_MemAlloc(9)
	reason := Sat_MemAppend(m, []int32{toLit(1), lit_neg(toLit(0))}, 2, 0, 0)
	conflict := Sat_MemAppend(m, []int32{lit_neg(toLit(1)), lit_neg(toLit(0))}, 2, 1, 0)
	a.decision(toLit(0))
	a.enqueue(toLit(1), reason)
	a.qhead = a.qtail
	var stats stats_t
	hooks := new(activityRecorder)
	learnt := satAnalyze(&a, m, conflict, hooks, &stats)
	if !reflect.DeepEqual(learnt, []lit{lit_neg(toLit(0))}) || a.tagged.size != 0 || stats.tot_literals != 1 || hooks.clauses != 1 {
		t.Fatal(learnt, a.tags, stats, hooks)
	}
}
func TestCDCLLearnedBacktrackLiteral(t *testing.T) {
	a := newSatAssignments(3)
	m := Sat_MemAlloc(9)
	a.decision(toLit(0))
	a.qhead = a.qtail
	a.decision(toLit(1))
	a.qhead = a.qtail
	conflict := Sat_MemAppend(m, []int32{lit_neg(toLit(0)), lit_neg(toLit(1))}, 2, 0, 0)
	var stats stats_t
	learnt := satAnalyze(&a, m, conflict, nil, &stats)
	if !reflect.DeepEqual(learnt, []lit{lit_neg(toLit(1)), lit_neg(toLit(0))}) {
		t.Fatal(learnt)
	}
	a.cancelUntil(a.levels[lit_var(learnt[1])])
	if !a.enqueue(learnt[0], conflict) {
		t.Fatal("learned clause not asserting")
	}
}
func TestCDCLMinimizationRollbackTags(t *testing.T) {
	a := newSatAssignments(3)
	m := Sat_MemAlloc(9)
	a.decision(toLit(0))
	a.enqueue(toLit(1), clause_from_lit(toLit(0)))
	a.qhead = a.qtail
	a.tagVariable(2, 1)
	if satLiteralRemovable(&a, m, 1, 1<<1) {
		t.Fatal("decision dependency cannot be removed")
	}
	if a.tagged.size != 1 || a.tags[0] != 0 || a.tags[2] != 1 {
		t.Fatal(a.tags)
	}
}

func TestCDCLLearnedClausesEntailedByFormula(t *testing.T) {
	rng := rand.New(rand.NewSource(7821))
	conflicts := 0
	for trial := 0; trial < 300; trial++ {
		a := newSatAssignments(5)
		m := Sat_MemAlloc(9)
		w := make([]veci, 10)
		binaryHandle := Sat_MemAppend(m, nil, 2, 0, 0)
		binary := Sat_MemClauseHand(m, binaryHandle)
		var formula [][]lit
		for k := 0; k < 12; k++ {
			permutation := rng.Perm(5)
			c := make([]lit, 2+rng.Intn(2))
			for i := range c {
				c[i] = toLitCond(int32(permutation[i]), int32(rng.Intn(2)))
			}
			formula = append(formula, c)
			if len(c) == 2 {
				veci_push(&w[lit_neg(c[0])], clause_from_lit(c[1]))
				veci_push(&w[lit_neg(c[1])], clause_from_lit(c[0]))
			} else {
				h := Sat_MemAppend(m, c, int32(len(c)), 0, 0)
				veci_push(&w[lit_neg(c[0])], h)
				veci_push(&w[lit_neg(c[1])], h)
			}
		}
		var stats stats_t
		for {
			next := int32(-1)
			for v, value := range a.assigns {
				if value == satUnassigned {
					next = int32(v)
					break
				}
			}
			if next < 0 {
				break
			}
			a.decision(toLitCond(next, int32(rng.Intn(2))))
			h := satPropagate(&a, m, w, binaryHandle, binary, nil, &stats)
			if h == 0 {
				continue
			}
			conflicts++
			learnt := satAnalyze(&a, m, h, nil, &stats)
			for model := 0; model < 32; model++ {
				satisfies := true
				for _, c := range formula {
					if !modelSatisfiesClause(model, c) {
						satisfies = false
						break
					}
				}
				if satisfies && !modelSatisfiesClause(model, learnt) {
					t.Fatalf("trial %d: learnt %v not entailed by %v (model %d)", trial, learnt, formula, model)
				}
			}
			if len(learnt) > 1 {
				level := a.levels[lit_var(learnt[1])]
				a.cancelUntil(level)
				if !a.enqueue(learnt[0], h) {
					t.Fatal("learned clause not asserting")
				}
			}
			break
		}
	}
	if conflicts < 10 {
		t.Fatal("insufficient conflict coverage", conflicts)
	}
}
func modelSatisfiesClause(model int, c []lit) bool {
	for _, l := range c {
		value := (model >> uint(lit_var(l))) & 1
		if value == 1-int(lit_sign(l)) {
			return true
		}
	}
	return false
}
