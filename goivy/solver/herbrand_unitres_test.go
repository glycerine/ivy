package solver

import (
	"testing"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/resolution"
	"github.com/glycerine/goivy/unitres"
)

// ---------------------------------------------------------------
// Conversion round-trip tests
// ---------------------------------------------------------------

func TestIvyLitToUnitResLitApply(t *testing.T) {
	// p(a) — positive literal with Apply atom
	p := relConst("p", unintSort("S"))
	a := lg.NewConst("a", unintSort("S"))
	applyAtom := &lg.Apply{Func: p, Terms: []lg.Expr{a}}
	lit := il.NewLiteral(1, applyAtom)

	symMap := make(map[string]*lg.Const)
	urLit := ivyLitToUnitResLit(lit, symMap)

	if urLit.Polarity != 1 {
		t.Fatalf("expected polarity 1, got %d", urLit.Polarity)
	}
	if urLit.Atom.RelName != "p" {
		t.Fatalf("expected relname 'p', got %q", urLit.Atom.RelName)
	}
	if len(urLit.Atom.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(urLit.Atom.Args))
	}
	if symMap["p"] != p {
		t.Fatal("symMap should record the original symbol")
	}
}

func TestIvyLitToUnitResLitEquality(t *testing.T) {
	// a = b — equality literal
	s := unintSort("S")
	a := lg.NewConst("a", s)
	b := lg.NewConst("b", s)
	eqAtom := &lg.Eq{T1: a, T2: b}
	lit := il.NewLiteral(1, eqAtom)

	symMap := make(map[string]*lg.Const)
	urLit := ivyLitToUnitResLit(lit, symMap)

	if urLit.Atom.RelName != "=" {
		t.Fatalf("expected relname '=', got %q", urLit.Atom.RelName)
	}
	if len(urLit.Atom.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(urLit.Atom.Args))
	}
}

func TestIvyLitToUnitResLitNullary(t *testing.T) {
	// p — nullary boolean symbol
	p := boolConst("p")
	lit := il.NewLiteral(0, p) // ~p

	symMap := make(map[string]*lg.Const)
	urLit := ivyLitToUnitResLit(lit, symMap)

	if urLit.Polarity != 0 {
		t.Fatalf("expected polarity 0, got %d", urLit.Polarity)
	}
	if urLit.Atom.RelName != "p" {
		t.Fatalf("expected relname 'p', got %q", urLit.Atom.RelName)
	}
	if len(urLit.Atom.Args) != 0 {
		t.Fatalf("expected 0 args, got %d", len(urLit.Atom.Args))
	}
}

func TestRoundTripLitConversion(t *testing.T) {
	// Test: il.Literal → unitres.Literal → il.Literal preserves semantics
	cases := []struct {
		name     string
		makeLit  func() *il.Literal
		checkStr string
	}{
		{
			name: "positive nullary",
			makeLit: func() *il.Literal {
				return il.NewLiteral(1, boolConst("q"))
			},
			checkStr: "q",
		},
		{
			name: "negative nullary",
			makeLit: func() *il.Literal {
				return il.NewLiteral(0, boolConst("q"))
			},
			checkStr: "~q",
		},
		{
			name: "positive apply",
			makeLit: func() *il.Literal {
				p := relConst("r", unintSort("T"))
				x := lg.NewConst("x", unintSort("T"))
				return il.NewLiteral(1, &lg.Apply{Func: p, Terms: []lg.Expr{x}})
			},
			checkStr: "r(x)",
		},
		{
			name: "equality",
			makeLit: func() *il.Literal {
				s := unintSort("U")
				return il.NewLiteral(1, &lg.Eq{T1: lg.NewConst("a", s), T2: lg.NewConst("b", s)})
			},
			checkStr: "(a == b)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := tc.makeLit()
			symMap := make(map[string]*lg.Const)
			urLit := ivyLitToUnitResLit(orig, symMap)
			roundTripped := unitResLitToIvyLit(urLit, symMap)

			if roundTripped.Polarity != orig.Polarity {
				t.Fatalf("polarity mismatch: got %d, want %d",
					roundTripped.Polarity, orig.Polarity)
			}
			gotStr := roundTripped.String()
			if gotStr != tc.checkStr {
				t.Fatalf("round-trip mismatch: got %q, want %q", gotStr, tc.checkStr)
			}
		})
	}
}

// ---------------------------------------------------------------
// Batch conversion tests
// ---------------------------------------------------------------

func TestIvyLitsToUnitResClauses(t *testing.T) {
	p := boolConst("p")
	q := boolConst("q")

	cnf := [][]*il.Literal{
		{il.NewLiteral(1, p), il.NewLiteral(0, q)}, // p ∨ ¬q
		{il.NewLiteral(0, p)},                        // ¬p
	}

	urClauses, symMap := ivyLitsToUnitResClauses(cnf)

	if len(urClauses) != 2 {
		t.Fatalf("expected 2 clauses, got %d", len(urClauses))
	}
	if len(urClauses[0]) != 2 {
		t.Fatalf("expected 2 literals in first clause, got %d", len(urClauses[0]))
	}
	if len(urClauses[1]) != 1 {
		t.Fatalf("expected 1 literal in second clause, got %d", len(urClauses[1]))
	}
	if _, ok := symMap["p"]; !ok {
		t.Fatal("symMap should contain 'p'")
	}
	if _, ok := symMap["q"]; !ok {
		t.Fatal("symMap should contain 'q'")
	}
}

func TestExtractUnitResResults(t *testing.T) {
	// Build a simple UnitRes with known unit propagation
	// Clauses: {~a}, {a ∨ b} → propagate ~a → derive b
	aAtom := resolution.NewAtom("a")
	bAtom := resolution.NewAtom("b")

	clauses := [][]*unitres.Literal{
		{unitres.NewLiteral(0, aAtom)},                                  // ~a
		{unitres.NewLiteral(1, aAtom), unitres.NewLiteral(1, bAtom)},   // a ∨ b
	}

	r := unitres.NewUnitRes(clauses)
	r.Propagate(nil)

	symMap := map[string]*lg.Const{
		"a": boolConst("a"),
		"b": boolConst("b"),
	}

	result := extractUnitResResults(r, symMap)
	// Should have unit queue (including derived units) + remaining clauses
	if len(result) == 0 {
		t.Fatal("expected non-empty results from UnitRes")
	}

	// Check that we got both ~a and b as units
	foundA := false
	foundB := false
	for _, clause := range result {
		if len(clause) == 1 {
			s := clause[0].String()
			if s == "~a" {
				foundA = true
			}
			if s == "b" {
				foundB = true
			}
		}
	}
	if !foundA {
		t.Fatal("expected to find unit ~a")
	}
	if !foundB {
		t.Fatal("expected to find unit b (derived by unit resolution)")
	}
}

// ---------------------------------------------------------------
// ClausesCase integration tests
// ---------------------------------------------------------------

func TestClausesCaseUNSAT(t *testing.T) {
	s := New()
	p := boolConst("p")
	// p AND NOT p — UNSAT
	fmlas := []lg.Expr{p, &lg.Not{Body: p}}
	clauses := clauseops.NewClauses(fmlas, nil, nil)

	result, err := s.ClausesCase(clauses)
	if err != nil {
		t.Fatal(err)
	}
	// Python returns [[]] (false clauses) on UNSAT, not None.
	// Our fix returns FalseClauses() which is non-nil and contains lg.False.
	if result == nil {
		t.Fatal("expected non-nil FalseClauses for UNSAT input")
	}
	if len(result.Fmlas) != 1 || !lg.IsFalse(result.Fmlas[0]) {
		t.Fatalf("expected FalseClauses for UNSAT input, got %v", result.Fmlas)
	}
}

func TestClausesCaseAllUnits(t *testing.T) {
	s := New()
	p := boolConst("p")
	q := boolConst("q")
	// p, q — already unit clauses
	fmlas := []lg.Expr{p, q}
	clauses := clauseops.NewClauses(fmlas, nil, nil)

	result, err := s.ClausesCase(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Fmlas) == 0 {
		t.Fatal("expected non-empty formulas")
	}
}

func TestClausesCaseWithUnitPropagation(t *testing.T) {
	// Clauses where UnitRes should derive additional unit clauses:
	// {~a}, {a ∨ b}, {~b ∨ c}
	// Model: a=false, b=true, c=true
	// After model simp: {~a}, {b}, {c}
	// UnitRes on {~a, b, c}: all units already, no change
	// But this verifies the pipeline works end-to-end.
	s := New()
	a := boolConst("a")
	b := boolConst("b")
	c := boolConst("c")
	notA := &lg.Not{Body: a}
	aOrB := &lg.Or{Terms: []lg.Expr{a, b}}
	notBOrC := &lg.Or{Terms: []lg.Expr{&lg.Not{Body: b}, c}}

	fmlas := []lg.Expr{notA, aOrB, notBOrC}
	clauses := clauseops.NewClauses(fmlas, nil, nil)

	result, err := s.ClausesCase(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil result (clauses are SAT)")
	}
	// Should produce simplified clauses
	if len(result.Fmlas) == 0 {
		t.Fatal("expected non-empty result formulas")
	}
}

func TestClausesCaseSingleClause(t *testing.T) {
	// Single disjunction: a ∨ b
	s := New()
	a := boolConst("a")
	b := boolConst("b")
	aOrB := &lg.Or{Terms: []lg.Expr{a, b}}

	clauses := clauseops.NewClauses([]lg.Expr{aOrB}, nil, nil)
	result, err := s.ClausesCase(clauses)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Model should pick one side; result should be a single unit
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
}
