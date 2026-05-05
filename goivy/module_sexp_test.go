package goivy

import (
	"path/filepath"
	"runtime"
	"testing"

	tv "github.com/glycerine/ivy/goivy/test_vectors"
)

func moduleVectorsPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "test_vectors", "sexp_vectors.sexp")
}

func moduleLoadVectors(t *testing.T) map[string]string {
	t.Helper()
	vecs, err := tv.LoadVectors(moduleVectorsPath())
	if err != nil {
		t.Fatalf("Failed to load vectors: %v", err)
	}
	m := make(map[string]string, len(vecs))
	for _, v := range vecs {
		m[v.ID] = v.Expected
	}
	return m
}

func moduleCheckSexp(t *testing.T, vecs map[string]string, id string, got string) {
	t.Helper()
	expected, ok := vecs[id]
	if !ok {
		t.Fatalf("No vector for id %q", id)
	}
	if got != expected {
		t.Errorf("Sexp mismatch for %s\n  got:  %s\n  want: %s", id, got, expected)
	}
}

func TestSexpClausesBasic(t *testing.T) {
	vecs := moduleLoadVectors(t)

	S := &UninterpretedSort{Name: "S"}
	X, err := NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	Y, err := NewVariable("Y", S)
	if err != nil {
		t.Fatal(err)
	}
	eq, err := NewEq(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	def := NewDefinition(X, Y)

	cl := &Clauses{
		Fmlas: []Expr{eq},
		Defs:  []*LogicDefinition{def},
	}
	moduleCheckSexp(t, vecs, "clauses_basic", string(cl.Canon()))
}

func TestSexpClausesEmpty(t *testing.T) {
	vecs := moduleLoadVectors(t)

	cl := &Clauses{
		Fmlas: []Expr{},
		Defs:  []*LogicDefinition{},
	}
	moduleCheckSexp(t, vecs, "clauses_empty", string(cl.Canon()))
}
