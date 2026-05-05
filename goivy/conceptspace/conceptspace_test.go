package conceptspace

import (
	goivy "github.com/glycerine/ivy/goivy"
	"testing"
)

func TestParseNamedSpace(t *testing.T) {
	s, err := ToConceptSpace("rel(X,Y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns, ok := s.(*NamedSpace)
	if !ok {
		t.Fatalf("expected NamedSpace, got %T", s)
	}
	if ns.Lit.Polarity != 1 {
		t.Errorf("expected polarity 1, got %d", ns.Lit.Polarity)
	}
}

func TestParseNegatedLit(t *testing.T) {
	s, err := ToConceptSpace("~rel(X)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns, ok := s.(*NamedSpace)
	if !ok {
		t.Fatalf("expected NamedSpace, got %T", s)
	}
	if ns.Lit.Polarity != 0 {
		t.Errorf("expected polarity 0, got %d", ns.Lit.Polarity)
	}
}

func TestParseProductSpace(t *testing.T) {
	s, err := ToConceptSpace("(rel(X) * rel(Y))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ps, ok := s.(*ProductSpace)
	if !ok {
		t.Fatalf("expected ProductSpace, got %T", s)
	}
	if len(ps.Spaces) != 2 {
		t.Errorf("expected 2 spaces, got %d", len(ps.Spaces))
	}
}

func TestParseSumSpace(t *testing.T) {
	s, err := ToConceptSpace("(rel(X) + rel(Y))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ss, ok := s.(*SumSpace)
	if !ok {
		t.Fatalf("expected SumSpace, got %T", s)
	}
	if len(ss.Spaces) != 2 {
		t.Errorf("expected 2 spaces, got %d", len(ss.Spaces))
	}
}

func TestParseTripleProduct(t *testing.T) {
	s, err := ToConceptSpace("(rel(X) * rel(Y) * rel(Z))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ps, ok := s.(*ProductSpace)
	if !ok {
		t.Fatalf("expected ProductSpace, got %T", s)
	}
	if len(ps.Spaces) != 3 {
		t.Errorf("expected 3 spaces, got %d", len(ps.Spaces))
	}
}

func TestParseTripleSum(t *testing.T) {
	s, err := ToConceptSpace("(rel(X) + rel(Y) + rel(Z))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ss, ok := s.(*SumSpace)
	if !ok {
		t.Fatalf("expected SumSpace, got %T", s)
	}
	if len(ss.Spaces) != 3 {
		t.Errorf("expected 3 spaces, got %d", len(ss.Spaces))
	}
}

func TestParseEmptyTerms(t *testing.T) {
	s, err := ToConceptSpace("rel()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns, ok := s.(*NamedSpace)
	if !ok {
		t.Fatalf("expected NamedSpace, got %T", s)
	}
	_ = ns
}

func TestEnumerateNamedSpace(t *testing.T) {
	s, err := ToConceptSpace("rel(X)")
	if err != nil {
		t.Fatal(err)
	}
	result := s.Enumerate(nil, func(lits []*goivy.Literal) bool { return true })
	if len(result) != 1 {
		t.Errorf("expected 1 clause, got %d", len(result))
	}
}

func TestEnumerateNamedSpaceFails(t *testing.T) {
	s, err := ToConceptSpace("rel(X)")
	if err != nil {
		t.Fatal(err)
	}
	result := s.Enumerate(nil, func(lits []*goivy.Literal) bool { return false })
	if len(result) != 0 {
		t.Errorf("expected 0 clauses, got %d", len(result))
	}
}

// Use il import so the test file compiles
var _ = goivy.NewLiteral
