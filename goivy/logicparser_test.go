package goivy

import (
	"fmt"
	"strings"
	"testing"
)

// nodeShape returns a compact string representation of AST structure for test assertions.
func nodeShape(n Node) string {
	if n == nil {
		return "<nil>"
	}
	switch t := n.(type) {
	case *Symbol:
		return t.Rep
	case *Atom:
		if len(t.Terms) == 0 {
			return t.Rep
		}
		return fmt.Sprintf("Atom(%s,[%s])", t.Rep, shapeList(t.Terms))
	case *App:
		return fmt.Sprintf("App(%v,[%s])", t.Rep, shapeList(t.Terms))
	case *And:
		if len(t.Terms) == 0 {
			return "True"
		}
		return fmt.Sprintf("And(%s)", shapeList(t.Terms))
	case *Or:
		if len(t.Terms) == 0 {
			return "False"
		}
		return fmt.Sprintf("Or(%s)", shapeList(t.Terms))
	case *Not:
		return fmt.Sprintf("Not(%s)", nodeShape(t.Body))
	case *Implies:
		return fmt.Sprintf("Implies(%s,%s)", nodeShape(t.T1), nodeShape(t.T2))
	case *Iff:
		return fmt.Sprintf("Iff(%s,%s)", nodeShape(t.T1), nodeShape(t.T2))
	default:
		return fmt.Sprintf("?(%T)", n)
	}
}

func shapeList(nodes []Node) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = nodeShape(n)
	}
	return strings.Join(parts, ",")
}

var (
	ver12 = Version{1, 2}
	ver16 = Version{1, 6}
	ver17 = Version{1, 7}
)

// TestLogicParser_V16_ChainedComparisonRejected verifies that the public
// ParseFormula API correctly rejects chained comparisons in v1.6,
// now that it dispatches to the LALR parser for ≤v1.6.
func TestLogicParser_V16_ChainedComparisonRejected(t *testing.T) {
	_, err := ParseFormula("a = b = c", ver16)
	if err == nil {
		t.Error("should reject chained comparisons in v1.6")
	}
}

// TestLogicParser_V16_ArrowRightAssoc verifies that -> is right-associative
// in v1.6 (dispatched to LALR parser).
func TestLogicParser_V16_ArrowRightAssoc(t *testing.T) {
	// In v1.6, "a -> b -> c" should parse as "a -> (b -> c)"
	node, err := ParseFormula("a -> b -> c", ver16)
	if err != nil {
		t.Fatal(err)
	}
	// The result should be Implies(a, Implies(b, c)), not Implies(Implies(a, b), c)
	s := nodeShape(node)
	expected := "Implies(a,Implies(b,c))"
	if s != expected {
		t.Errorf("v1.6 'a -> b -> c' = %s, want %s", s, expected)
	}
}

// TestLogicParser_V16_ArrowVsAnd verifies that "a & b -> c" parses as
// "a & (b -> c)" in v1.6 (ARROW has no precedence entry in PLY grammar).
func TestLogicParser_V16_ArrowVsAnd(t *testing.T) {
	node, err := ParseFormula("a & b -> c", ver16)
	if err != nil {
		t.Fatal(err)
	}
	s := nodeShape(node)
	expected := "And(a,Implies(b,c))"
	if s != expected {
		t.Errorf("v1.6 'a & b -> c' = %s, want %s", s, expected)
	}
}

// TestLogicParser_V16_IffVsAnd verifies that "a & b <-> c" parses as
// "a & (b <-> c)" in v1.6.
func TestLogicParser_V16_IffVsAnd(t *testing.T) {
	node, err := ParseFormula("a & b <-> c", ver16)
	if err != nil {
		t.Fatal(err)
	}
	s := nodeShape(node)
	expected := "And(a,Iff(b,c))"
	if s != expected {
		t.Errorf("v1.6 'a & b <-> c' = %s, want %s", s, expected)
	}
}

// TestLogicParser_V17_Unchanged verifies that v1.7+ still uses the
// hand-written parser and produces the same results as before.
func TestLogicParser_V17_Unchanged(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"a -> b -> c", "Implies(Implies(App(a,[]),App(b,[])),App(c,[]))"}, // left-assoc in v1.7
		{"a & b -> c", "Implies(And(App(a,[]),App(b,[])),App(c,[]))"},      // AND tighter than ARROW in v1.7
		{"a = b = c", "Atom(=,[Atom(=,[App(a,[]),App(b,[])]),App(c,[])])"}, // chained OK in v1.7
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			node, err := ParseFormula(tc.input, ver17)
			if err != nil {
				t.Fatal(err)
			}
			s := nodeShape(node)
			if s != tc.expected {
				t.Errorf("v1.7 %q = %s, want %s", tc.input, s, tc.expected)
			}
		})
	}
}

// TestLogicParser_V12_ChainedComparisonRejected verifies v1.2 also
// dispatches to LALR and rejects chained comparisons.
func TestLogicParser_V12_ChainedComparisonRejected(t *testing.T) {
	_, err := ParseFormula("a = b = c", ver12)
	if err == nil {
		t.Error("should reject chained comparisons in v1.2")
	}
}

// TestLogicParser_V16_BasicFormulas verifies that basic formulas
// still parse correctly when dispatched to LALR for v1.6.
func TestLogicParser_V16_BasicFormulas(t *testing.T) {
	cases := []string{
		"foo", "X", "true", "false",
		"~p", "a & b", "a | b", "a -> b", "a <-> b",
		"f(x)", "f(x, y)",
		"x = y", "x + y", "x * y",
		"forall X. p(X)", "exists X. p(X)",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := ParseFormula(input, ver16)
			if err != nil {
				t.Errorf("ParseFormula(%q, v1.6) failed: %v", input, err)
			}
		})
	}
}

func TestLogicParser_V16_SubscriptedSymbols(t *testing.T) {
	node, err := ParseFormula("lbl[this.part]", ver16)
	if err != nil {
		t.Fatalf("ParseFormula v1.6 subscripted symbol: %v", err)
	}
	if got := NodeRep(node); got != "lbl[this.part]" {
		t.Fatalf("subscripted symbol parsed as %q, want lbl[this.part]", got)
	}
}

func TestLogicParser_V16_NamedBinderTerms(t *testing.T) {
	node, err := ParseFormula("($snap X. p(X))(a)", ver16)
	if err != nil {
		t.Fatalf("ParseFormula v1.6 named binder application: %v", err)
	}
	app, ok := node.(*App)
	if !ok {
		t.Fatalf("named binder application = %T, want *App", node)
	}
	binder, ok := app.Rep.(*NamedBinder)
	if !ok {
		t.Fatalf("named binder app rep = %T, want *NamedBinder", app.Rep)
	}
	if binder.Name != "snap" || len(binder.Bounds) != 1 {
		t.Fatalf("named binder = %s with %d bounds, want snap with 1", binder.Name, len(binder.Bounds))
	}

	for _, input := range []string{"$saved. p(a)", "$saved $ p(a)"} {
		t.Run(input, func(t *testing.T) {
			node, err := ParseFormula(input, ver16)
			if err != nil {
				t.Fatalf("ParseFormula(%q, v1.6): %v", input, err)
			}
			binder, ok := node.(*NamedBinder)
			if !ok {
				t.Fatalf("ParseFormula(%q) = %T, want *NamedBinder", input, node)
			}
			if binder.Name != "saved" {
				t.Fatalf("binder name = %q, want saved", binder.Name)
			}
		})
	}
}

func TestLogicParser_V16_ParseTermUsesTermStart(t *testing.T) {
	if _, err := ParseFormula("x = y", ver16); err != nil {
		t.Fatalf("ParseFormula should accept v1.6 equality formula: %v", err)
	}
	if _, err := ParseTerm("x = y", ver16); err == nil {
		t.Fatal("ParseTerm accepted v1.6 equality formula; want term-start parse error")
	}
	if _, err := ParseTerm("f(x)", ver16); err != nil {
		t.Fatalf("ParseTerm should accept v1.6 term application: %v", err)
	}
}
