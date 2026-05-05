package goivy

import (
	"path/filepath"
	"runtime"
	"testing"

	tv "github.com/glycerine/ivy/goivy/test_vectors"
)

func ivyLogicVectorsPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "test_vectors", "sexp_vectors.sexp")
}

func ivyLogicLoadVectors(t *testing.T) map[string]string {
	t.Helper()
	vecs, err := tv.LoadVectors(ivyLogicVectorsPath())
	if err != nil {
		t.Fatalf("Failed to load vectors: %v", err)
	}
	m := make(map[string]string, len(vecs))
	for _, v := range vecs {
		m[v.ID] = v.Expected
	}
	return m
}

func ivyLogicCheckSexp(t *testing.T, vecs map[string]string, id string, got string) {
	t.Helper()
	expected, ok := vecs[id]
	if !ok {
		t.Fatalf("No vector for id %q", id)
	}
	if got != expected {
		t.Errorf("Sexp mismatch for %s\n  got:  %s\n  want: %s", id, got, expected)
	}
}

func setup(t *testing.T) (*UninterpretedSort, *Variable, *Variable, *Eq) {
	t.Helper()
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
	return S, X, Y, eq
}

func TestSexpSome(t *testing.T) {
	vecs := ivyLogicLoadVectors(t)
	_, X, _, eq := setup(t)

	some := NewSome([]Expr{X}, eq)
	ivyLogicCheckSexp(t, vecs, "some_basic", string(some.Sexp()))
}

func TestSexpSomeWithElse(t *testing.T) {
	vecs := ivyLogicLoadVectors(t)
	_, X, Y, eq := setup(t)

	some := NewSomeWithElse([]Expr{X}, eq, X, Y)
	ivyLogicCheckSexp(t, vecs, "some_with_else", string(some.Sexp()))
}

func TestSexpLet(t *testing.T) {
	vecs := ivyLogicLoadVectors(t)
	_, X, Y, _ := setup(t)

	def := NewDefinition(X, Y)
	let := NewLet([]Expr{def}, X)
	ivyLogicCheckSexp(t, vecs, "let", string(let.Sexp()))
}

func TestSexpLiteral(t *testing.T) {
	vecs := ivyLogicLoadVectors(t)
	_, _, _, eq := setup(t)

	litPos := NewLiteral(1, eq)
	ivyLogicCheckSexp(t, vecs, "literal_pos", string(litPos.Sexp()))

	litNeg := NewLiteral(0, eq)
	ivyLogicCheckSexp(t, vecs, "literal_neg", string(litNeg.Sexp()))
}

func TestCanonEqualsSexpIvylogic(t *testing.T) {
	_, X, Y, eq := setup(t)

	some := NewSome([]Expr{X}, eq)
	if string(some.Sexp()) != string(some.Canon()) {
		t.Errorf("Canon != Sexp for Some")
	}

	someE := NewSomeWithElse([]Expr{X}, eq, X, Y)
	if string(someE.Sexp()) != string(someE.Canon()) {
		t.Errorf("Canon != Sexp for SomeWithElse")
	}

	def := NewDefinition(X, Y)
	let := NewLet([]Expr{def}, X)
	if string(let.Sexp()) != string(let.Canon()) {
		t.Errorf("Canon != Sexp for Let")
	}

	lit := NewLiteral(1, eq)
	if string(lit.Sexp()) != string(lit.Canon()) {
		t.Errorf("Canon != Sexp for Literal")
	}
}
