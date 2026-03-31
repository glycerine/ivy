package clauseops

import (
	"path/filepath"
	"runtime"
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
	tv "github.com/glycerine/ivy/goivy/test_vectors"
)

func vectorsPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "test_vectors", "sexp_vectors.sexp")
}

func loadVectors(t *testing.T) map[string]string {
	t.Helper()
	vecs, err := tv.LoadVectors(vectorsPath())
	if err != nil {
		t.Fatalf("Failed to load vectors: %v", err)
	}
	m := make(map[string]string, len(vecs))
	for _, v := range vecs {
		m[v.ID] = v.Expected
	}
	return m
}

func checkSexp(t *testing.T, vecs map[string]string, id string, got string) {
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
	vecs := loadVectors(t)

	S := &lg.UninterpretedSort{Name: "S"}
	X, err := lg.NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	Y, err := lg.NewVariable("Y", S)
	if err != nil {
		t.Fatal(err)
	}
	eq, err := lg.NewEq(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	def := lg.NewDefinition(X, Y)

	cl := &Clauses{
		Fmlas: []lg.Expr{eq},
		Defs:  []*lg.Definition{def},
	}
	checkSexp(t, vecs, "clauses_basic", string(cl.Canon()))
}

func TestSexpClausesEmpty(t *testing.T) {
	vecs := loadVectors(t)

	cl := &Clauses{
		Fmlas: []lg.Expr{},
		Defs:  []*lg.Definition{},
	}
	checkSexp(t, vecs, "clauses_empty", string(cl.Canon()))
}
