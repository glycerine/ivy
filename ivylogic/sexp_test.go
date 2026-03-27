package ivylogic

import (
	"path/filepath"
	"runtime"
	"testing"

	lg "github.com/glycerine/goivy/logic"
	tv "github.com/glycerine/goivy/test_vectors"
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

func setup(t *testing.T) (*lg.UninterpretedSort, *lg.Variable, *lg.Variable, *lg.Eq) {
	t.Helper()
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
	return S, X, Y, eq
}

func TestSexpSome(t *testing.T) {
	vecs := loadVectors(t)
	_, X, _, eq := setup(t)

	some := NewSome([]lg.Expr{X}, eq)
	checkSexp(t, vecs, "some_basic", string(some.Sexp()))
}

func TestSexpSomeWithElse(t *testing.T) {
	vecs := loadVectors(t)
	_, X, Y, eq := setup(t)

	some := NewSomeWithElse([]lg.Expr{X}, eq, X, Y)
	checkSexp(t, vecs, "some_with_else", string(some.Sexp()))
}

func TestSexpLet(t *testing.T) {
	vecs := loadVectors(t)
	_, X, Y, _ := setup(t)

	def := lg.NewDefinition(X, Y)
	let := NewLet([]lg.Expr{def}, X)
	checkSexp(t, vecs, "let", string(let.Sexp()))
}

func TestSexpLiteral(t *testing.T) {
	vecs := loadVectors(t)
	_, _, _, eq := setup(t)

	litPos := NewLiteral(1, eq)
	checkSexp(t, vecs, "literal_pos", string(litPos.Sexp()))

	litNeg := NewLiteral(0, eq)
	checkSexp(t, vecs, "literal_neg", string(litNeg.Sexp()))
}

func TestCanonEqualsSexpIvylogic(t *testing.T) {
	_, X, Y, eq := setup(t)

	some := NewSome([]lg.Expr{X}, eq)
	if string(some.Sexp()) != string(some.Canon()) {
		t.Errorf("Canon != Sexp for Some")
	}

	someE := NewSomeWithElse([]lg.Expr{X}, eq, X, Y)
	if string(someE.Sexp()) != string(someE.Canon()) {
		t.Errorf("Canon != Sexp for SomeWithElse")
	}

	def := lg.NewDefinition(X, Y)
	let := NewLet([]lg.Expr{def}, X)
	if string(let.Sexp()) != string(let.Canon()) {
		t.Errorf("Canon != Sexp for Let")
	}

	lit := NewLiteral(1, eq)
	if string(lit.Sexp()) != string(lit.Canon()) {
		t.Errorf("Canon != Sexp for Literal")
	}
}
