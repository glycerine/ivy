package lalr_logicparser

import (
	"fmt"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/parser"
)

var v17 = lexer.Version{1, 7}

// parseHW parses with the hand-written parser.
func parseHW(input string, version lexer.Version) (ast.Node, error) {
	p := parser.New(input, version)
	result := p.ParseExpr(0)
	if result == nil {
		errs := p.Errors()
		if len(errs) > 0 {
			return nil, fmt.Errorf("%s", errs[0].Error())
		}
		return nil, fmt.Errorf("failed to parse")
	}
	return result, nil
}

// parseLALR parses with the LALR parser.
func parseLALR(input string, version lexer.Version) (ast.Node, error) {
	return ParseV17(input, version)
}

// astTypeTag returns a short string identifying the AST node type.
func astTypeTag(n ast.Node) string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T", n)
}

// normalizeAtomSymbol controls whether bare Symbol/Atom with no args
// should be treated as equivalent (they are different AST types but
// represent the same thing — the hand-written parser produces Symbol
// while the LALR grammar produces Atom for bare identifiers).
var normalizeAtomSymbol = true

// astShape returns a string representation of the AST structure,
// ignoring source locations. Used for cross-validation.
func astShape(n ast.Node) string {
	if n == nil {
		return "<nil>"
	}
	switch t := n.(type) {
	case *ast.Symbol:
		if normalizeAtomSymbol {
			return fmt.Sprintf("Atom(%s)", t.Rep) // normalize to Atom
		}
		return fmt.Sprintf("Symbol(%s)", t.Rep)
	case *ast.Variable:
		return fmt.Sprintf("Var(%s)", t.Rep)
	case *ast.Atom:
		if len(t.Terms) == 0 {
			return fmt.Sprintf("Atom(%s)", t.Rep)
		}
		args := ""
		for i, a := range t.Terms {
			if i > 0 {
				args += ","
			}
			args += astShape(a)
		}
		return fmt.Sprintf("Atom(%s,[%s])", t.Rep, args)
	case *ast.App:
		args := ""
		for i, a := range t.Terms {
			if i > 0 {
				args += ","
			}
			args += astShape(a)
		}
		return fmt.Sprintf("App(%v,[%s])", t.Rep, args)
	case *ast.And:
		if len(t.Terms) == 0 {
			return "True"
		}
		args := ""
		for i, a := range t.Terms {
			if i > 0 {
				args += ","
			}
			args += astShape(a)
		}
		return fmt.Sprintf("And(%s)", args)
	case *ast.Or:
		if len(t.Terms) == 0 {
			return "False"
		}
		args := ""
		for i, a := range t.Terms {
			if i > 0 {
				args += ","
			}
			args += astShape(a)
		}
		return fmt.Sprintf("Or(%s)", args)
	case *ast.Not:
		return fmt.Sprintf("Not(%s)", astShape(t.Body))
	case *ast.Implies:
		return fmt.Sprintf("Implies(%s,%s)", astShape(t.T1), astShape(t.T2))
	case *ast.Iff:
		return fmt.Sprintf("Iff(%s,%s)", astShape(t.T1), astShape(t.T2))
	case *ast.Ite:
		return fmt.Sprintf("Ite(%s,%s,%s)", astShape(t.Cond), astShape(t.Then), astShape(t.Else))
	case *ast.Forall:
		bounds := ""
		for i, b := range t.Bounds {
			if i > 0 {
				bounds += ","
			}
			bounds += astShape(b)
		}
		return fmt.Sprintf("Forall([%s],%s)", bounds, astShape(t.Body))
	case *ast.Exists:
		bounds := ""
		for i, b := range t.Bounds {
			if i > 0 {
				bounds += ","
			}
			bounds += astShape(b)
		}
		return fmt.Sprintf("Exists([%s],%s)", bounds, astShape(t.Body))
	case *ast.Globally:
		return fmt.Sprintf("Globally(%s)", astShape(t.Body))
	case *ast.Eventually:
		return fmt.Sprintf("Eventually(%s)", astShape(t.Body))
	case *ast.WhenOperator:
		return fmt.Sprintf("When(%s,%s,%s)", t.Name, astShape(t.T1), astShape(t.T2))
	case *ast.Old:
		return fmt.Sprintf("Old(%s)", astShape(t.Term))
	case *ast.This:
		return "This"
	case *ast.MethodCall:
		return fmt.Sprintf("MethodCall(%s,%s)", astShape(t.Obj), astShape(t.Method))
	case *ast.Isa:
		args := ""
		for i, a := range t.Terms {
			if i > 0 {
				args += ","
			}
			args += astShape(a)
		}
		return fmt.Sprintf("Isa(%s)", args)
	case *ast.NamedBinder:
		return fmt.Sprintf("NamedBinder(%s,%s)", t.Name, astShape(t.Body))
	default:
		return fmt.Sprintf("?(%T)", n)
	}
}

// --- Basic LALR parsing tests ---

func TestLALR_Symbol(t *testing.T) {
	n, err := parseLALR("foo", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "Atom(foo)" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_Variable(t *testing.T) {
	n, err := parseLALR("X", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "Var(X)" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_True(t *testing.T) {
	n, err := parseLALR("true", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "True" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_False(t *testing.T) {
	n, err := parseLALR("false", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "False" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_Not(t *testing.T) {
	n, err := parseLALR("~p", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "Not(Atom(p))" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_And(t *testing.T) {
	n, err := parseLALR("a & b", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "And(Atom(a),Atom(b))" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_Or(t *testing.T) {
	n, err := parseLALR("a | b", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "Or(Atom(a),Atom(b))" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_Implies(t *testing.T) {
	n, err := parseLALR("a -> b", v17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "Implies(Atom(a),Atom(b))" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_Eq(t *testing.T) {
	n, err := parseLALR("x = y", v17)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Atom(=,[Atom(x),Atom(y)])"
	if astShape(n) != expected {
		t.Errorf("got %s, want %s", astShape(n), expected)
	}
}

func TestLALR_Forall(t *testing.T) {
	n, err := parseLALR("forall X. p(X)", v17)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Forall([Var(X)],Atom(p,[Var(X)]))"
	if astShape(n) != expected {
		t.Errorf("got %s, want %s", astShape(n), expected)
	}
}

// --- Cross-validation tests ---

func TestCrossValidation_BasicFormulas(t *testing.T) {
	formulas := []string{
		"foo",
		"X",
		"true",
		"false",
		"~p",
		"a & b",
		"a | b",
		"a -> b",
		"a <-> b",
		"f(x)",
		"f(x, y)",
		"x = y",
		"x ~= y",
		"x < y",
		"x <= y",
		"x + y",
		"x * y",
		"forall X. p(X)",
		"exists X. p(X)",
		"globally p",
		"eventually p",
	}

	for _, f := range formulas {
		hw, hwErr := parseHW(f, v17)
		lalr, lalrErr := parseLALR(f, v17)

		if hwErr != nil && lalrErr != nil {
			continue // both fail, OK
		}
		if hwErr != nil {
			t.Errorf("%q: hand-written failed (%v) but LALR succeeded: %s", f, hwErr, astShape(lalr))
			continue
		}
		if lalrErr != nil {
			t.Errorf("%q: LALR failed (%v) but hand-written succeeded: %s", f, lalrErr, astShape(hw))
			continue
		}

		hwShape := astShape(hw)
		lalrShape := astShape(lalr)
		if hwShape != lalrShape {
			t.Errorf("%q: MISMATCH\n  hand-written: %s\n  LALR:         %s", f, hwShape, lalrShape)
		}
	}
}

// TestPrecedenceEdgeCases tests operator pairs where precedence matters.
func TestPrecedenceEdgeCases(t *testing.T) {
	cases := []struct {
		input    string
		desc     string
	}{
		{"a & b | c", "AND vs OR precedence"},
		{"a | b & c", "OR vs AND precedence"},
		{"~a & b", "NOT vs AND precedence"},
		{"~a = b", "NOT vs EQ precedence (v1.7+: NOT tighter)"},
		{"a -> b -> c", "ARROW associativity"},
		{"a & b -> c", "AND vs ARROW precedence"},
		{"a + b = c", "PLUS vs EQ precedence"},
		{"a * b + c", "TIMES vs PLUS precedence"},
		{"a = b & c = d", "EQ vs AND"},
		{"globally a & b", "GLOBALLY vs AND"},
	}

	for _, tc := range cases {
		hw, hwErr := parseHW(tc.input, v17)
		lalr, lalrErr := parseLALR(tc.input, v17)

		if hwErr != nil || lalrErr != nil {
			if hwErr != nil && lalrErr != nil {
				continue
			}
			t.Errorf("%s (%q): one parser failed: hw=%v, lalr=%v", tc.desc, tc.input, hwErr, lalrErr)
			continue
		}

		hwShape := astShape(hw)
		lalrShape := astShape(lalr)
		if hwShape != lalrShape {
			t.Errorf("%s (%q): MISMATCH\n  hand-written: %s\n  LALR:         %s", tc.desc, tc.input, hwShape, lalrShape)
		}
	}
}
