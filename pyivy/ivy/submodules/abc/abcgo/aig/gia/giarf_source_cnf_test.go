package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// Compare the actual encoded clause count, not just the backend's type name.
// The three-input supergate needs three implications and one long clause;
// the constant unit is propagated rather than stored as a normal clause.
func TestGiarfCECUsesSourceSupergateEncoding(t *testing.T) {
	g := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, Gia_ManAppendAnd(g, a, b), c)
	Gia_ManCreateRefs(g)
	Gia_ManFillValue(g)
	s := newGiarfSolver(g, 1000, true)
	defer s.close()
	if status, _, err := s.check(root, nil); err != nil || status != 0 {
		t.Fatalf("source query status %d: %v", status, err)
	}
	if got := giarfEncodedClauseCount(s); got != 4 {
		t.Fatalf("encoded %d clauses, want the source's 4-clause supergate", got)
	}
	ids := s.cec.IDs()
	if ids[root>>1] != 2 || ids[c>>1] != 3 || ids[a>>1] != 4 || ids[b>>1] != 5 || ids[root>>1-1] != 0 {
		t.Fatalf("source frontier allocation changed: %v", ids)
	}
}

func TestGiarfSourceQueriesMatchCompiledCEC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "giarf_source_sat.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("source CEC queries: %v: %s", err, want)
	}
	g := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	mid := Gia_ManAppendAnd(g, a, b)
	root := Gia_ManAppendAnd(g, mid, c)
	yes, no := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, a^1, c)
	mux := Gia_ManAppendAnd(g, yes^1, no^1)
	Gia_ManCreateRefs(g)
	Gia_ManFillValue(g)
	s := newGiarfSolver(g, 1000, true)
	defer s.close()
	var got strings.Builder
	queries := [][2]int32{{root, -1}, {a, -1}, {root ^ 1, -1}, {root ^ 1, a}, {root, a ^ 1}, {root, root ^ 1}, {mux, -1}, {mux ^ 1, root ^ 1}, {mux, root}, {root, -1}}
	for i, query := range queries {
		var second *int32
		if query[1] >= 0 {
			second = &query[1]
		}
		status, _, err := s.check(query[0], second)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&got, "query %d %d vars %d clauses %d since %d ids", i, status, s.cec.Solver().NumVars(), giarfEncodedClauseCount(s), s.cecSince)
		for _, variable := range s.cec.IDs() {
			fmt.Fprintf(&got, " %d", variable)
		}
		got.WriteByte('\n')
	}
	if got.String() != string(want) {
		t.Fatalf("Go source queries:\n%sC queries:\n%s", got.String(), want)
	}
}

func TestGiarfSourceRecycleBoundaries(t *testing.T) {
	g := Gia_ManStart(1)
	root := Gia_ManAppendCi(g)
	for i := 1; i < 2000; i++ {
		root = Gia_ManAppendAnd(g, root, Gia_ManAppendCi(g))
	}
	Gia_ManCreateRefs(g)
	Gia_ManFillValue(g)
	s := newGiarfSolver(g, 1000, true)
	defer s.close()
	for i := 0; i <= 201; i++ {
		if status, model, err := s.check(root, nil); err != nil || status != 0 || Vec_IntSize(model) != 2000 {
			t.Fatalf("query %d: status %d, model size %d, %v", i, status, Vec_IntSize(model), err)
		}
		want := int32(i)
		if i == 201 {
			want = 0
		}
		if s.cecSince != want {
			t.Fatalf("query %d: source recycle counter %d, want %d", i, s.cecSince, want)
		}
	}
}

func giarfEncodedClauseCount(s *giarfSolver) int {
	return int(s.cec.Solver().Statistics().Clauses)
}
