package lalr_logicparser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/parser"
)

var (
	ver12 = lexer.Version{1, 2}
	ver16 = lexer.Version{1, 6}
	ver17 = lexer.Version{1, 7}
)

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

// normalizeAtomSymbol: treat bare Symbol(x) and Atom(x) as equivalent.
var normalizeAtomSymbol = true

// astShape returns a string representation of the AST structure.
func astShape(n ast.Node) string {
	if n == nil {
		return "<nil>"
	}
	switch t := n.(type) {
	case *ast.Symbol:
		if normalizeAtomSymbol {
			return fmt.Sprintf("Atom(%s)", t.Rep)
		}
		return fmt.Sprintf("Symbol(%s)", t.Rep)
	case *ast.Variable:
		return fmt.Sprintf("Var(%s)", t.Rep)
	case *ast.Atom:
		if len(t.Terms) == 0 {
			return fmt.Sprintf("Atom(%s)", t.Rep)
		}
		args := shapeList(t.Terms)
		return fmt.Sprintf("Atom(%s,[%s])", t.Rep, args)
	case *ast.App:
		args := shapeList(t.Terms)
		return fmt.Sprintf("App(%v,[%s])", t.Rep, args)
	case *ast.And:
		if len(t.Terms) == 0 {
			return "True"
		}
		// Flatten nested And for comparison (both parsers are correct,
		// they just differ in representation: binary tree vs n-ary).
		flat := flattenAnd(t)
		return fmt.Sprintf("And(%s)", shapeList(flat))
	case *ast.Or:
		if len(t.Terms) == 0 {
			return "False"
		}
		flat := flattenOr(t)
		return fmt.Sprintf("Or(%s)", shapeList(flat))
	case *ast.Not:
		return fmt.Sprintf("Not(%s)", astShape(t.Body))
	case *ast.Implies:
		return fmt.Sprintf("Implies(%s,%s)", astShape(t.T1), astShape(t.T2))
	case *ast.Iff:
		return fmt.Sprintf("Iff(%s,%s)", astShape(t.T1), astShape(t.T2))
	case *ast.Ite:
		return fmt.Sprintf("Ite(%s,%s,%s)", astShape(t.Cond), astShape(t.Then), astShape(t.Else))
	case *ast.Forall:
		return fmt.Sprintf("Forall([%s],%s)", shapeList(t.Bounds), astShape(t.Body))
	case *ast.Exists:
		return fmt.Sprintf("Exists([%s],%s)", shapeList(t.Bounds), astShape(t.Body))
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
		return fmt.Sprintf("Isa(%s)", shapeList(t.Terms))
	case *ast.NamedBinder:
		return fmt.Sprintf("NamedBinder(%s,%s)", t.Name, astShape(t.Body))
	default:
		return fmt.Sprintf("?(%T)", n)
	}
}

// flattenAnd recursively flattens nested And nodes into a single list.
func flattenAnd(n *ast.And) []ast.Node {
	var result []ast.Node
	for _, t := range n.Terms {
		if inner, ok := t.(*ast.And); ok && len(inner.Terms) > 0 {
			result = append(result, flattenAnd(inner)...)
		} else {
			result = append(result, t)
		}
	}
	return result
}

// flattenOr recursively flattens nested Or nodes into a single list.
func flattenOr(n *ast.Or) []ast.Node {
	var result []ast.Node
	for _, t := range n.Terms {
		if inner, ok := t.(*ast.Or); ok && len(inner.Terms) > 0 {
			result = append(result, flattenOr(inner)...)
		} else {
			result = append(result, t)
		}
	}
	return result
}

func shapeList(nodes []ast.Node) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = astShape(n)
	}
	return strings.Join(parts, ",")
}

// crossValidate parses the same input with both parsers and compares.
func crossValidate(t *testing.T, input string, version lexer.Version) {
	t.Helper()
	hw, hwErr := parseHW(input, version)
	lalr, lalrErr := Parse(input, version)

	if hwErr != nil && lalrErr != nil {
		return // both fail, OK
	}
	if hwErr != nil {
		t.Errorf("v%d.%d %q: hand-written failed (%v) but LALR succeeded: %s",
			version[0], version[1], input, hwErr, astShape(lalr))
		return
	}
	if lalrErr != nil {
		t.Errorf("v%d.%d %q: LALR failed (%v) but hand-written succeeded: %s",
			version[0], version[1], input, lalrErr, astShape(hw))
		return
	}

	hwShape := astShape(hw)
	lalrShape := astShape(lalr)
	if hwShape != lalrShape {
		t.Errorf("v%d.%d %q: MISMATCH\n  hand-written: %s\n  LALR:         %s",
			version[0], version[1], input, hwShape, lalrShape)
	}
}

// === Basic LALR parsing tests ===

func TestLALR_Symbol(t *testing.T) {
	n, err := ParseV17("foo", ver17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "Atom(foo)" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_Variable(t *testing.T) {
	n, err := ParseV17("X", ver17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "Var(X)" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_True(t *testing.T) {
	n, err := ParseV17("true", ver17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "True" {
		t.Errorf("got %s", astShape(n))
	}
}

func TestLALR_False(t *testing.T) {
	n, err := ParseV17("false", ver17)
	if err != nil {
		t.Fatal(err)
	}
	if astShape(n) != "False" {
		t.Errorf("got %s", astShape(n))
	}
}

// === Cross-validation: v1.7 basic formulas ===

func TestCrossValidation_V17_BasicFormulas(t *testing.T) {
	formulas := []string{
		"foo", "X", "true", "false",
		"~p", "a & b", "a | b", "a -> b", "a <-> b",
		"f(x)", "f(x, y)", "g(x, y, z)",
		"x = y", "x ~= y", "x < y", "x <= y", "x > y", "x >= y",
		"x + y", "x - y", "x * y", "x / y",
		"forall X. p(X)", "exists X. p(X)",
		"forall X, Y. r(X, Y)",
		"globally p", "eventually p",
		"(a & b) | c", "a & (b | c)",
		"~(a & b)", "~~p",
	}
	for _, f := range formulas {
		crossValidate(t, f, ver17)
	}
}

// === Cross-validation: v1.7 precedence edge cases ===

func TestCrossValidation_V17_Precedence(t *testing.T) {
	cases := []struct {
		input string
		desc  string
	}{
		// AND vs OR
		{"a & b | c", "AND binds tighter than OR"},
		{"a | b & c", "AND binds tighter than OR (right)"},
		// NOT vs comparison
		{"~a & b", "NOT vs AND"},
		{"~a = b", "NOT vs EQ"},
		// ARROW
		{"a -> b -> c", "ARROW left-assoc"},
		{"a & b -> c", "AND vs ARROW"},
		{"a -> b & c", "ARROW vs AND (right)"},
		// IFF
		{"a <-> b <-> c", "IFF left-assoc"},
		{"a -> b <-> c", "ARROW vs IFF (same level)"},
		// Arithmetic vs comparison
		{"a + b = c", "PLUS vs EQ"},
		{"a * b + c", "TIMES vs PLUS"},
		{"a + b * c", "PLUS vs TIMES (right)"},
		// Compound
		{"a = b & c = d", "EQ vs AND"},
		{"a < b & c > d", "LT/GT vs AND"},
		{"globally a & b", "GLOBALLY vs AND"},
		{"eventually a | b", "EVENTUALLY vs OR"},
		// Quantifiers
		{"forall X. a & b", "FORALL scopes over AND"},
		{"forall X. a -> b", "FORALL scopes over ARROW"},
		// Nested arithmetic
		{"a + b + c", "PLUS left-assoc"},
		{"a * b * c", "TIMES left-assoc"},
		{"a - b - c", "MINUS left-assoc"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			crossValidate(t, tc.input, ver17)
		})
	}
}

// === Cross-validation: v1.7 complex formulas ===

func TestCrossValidation_V17_Complex(t *testing.T) {
	formulas := []string{
		"forall X:t. exists Y:t. r(X, Y)",
		"forall X. p(X) -> q(X)",
		"exists X. p(X) & q(X)",
		"(forall X. p(X)) & (exists Y. q(Y))",
		"a = b -> c = d",
		"f(g(x)) = h(y, z)",
		"~(a & b) | (c -> d)",
		"a & b & c & d",      // n-ary AND
		"a | b | c | d",      // n-ary OR
		"forall X. forall Y. r(X, Y) -> r(Y, X)",
		"f(x) + g(y) = h(z)",
		"a * (b + c) = a * b + a * c",
	}
	for _, f := range formulas {
		crossValidate(t, f, ver17)
	}
}

// === Generated formulas: systematic operator combinations ===

func TestGenerated_BinaryOperatorPairs(t *testing.T) {
	// For each pair of binary operators, test "a OP1 b OP2 c"
	// to verify precedence and associativity match.
	ops := []struct {
		sym     string
		version lexer.Version
	}{
		{"&", ver17}, {"|", ver17}, {"->", ver17}, {"<->", ver17},
		{"=", ver17}, {"<", ver17}, {"<=", ver17}, {">", ver17}, {">=", ver17},
		{"+", ver17}, {"-", ver17}, {"*", ver17}, {"/", ver17},
	}

	for _, op1 := range ops {
		for _, op2 := range ops {
			input := fmt.Sprintf("a %s b %s c", op1.sym, op2.sym)
			t.Run(input, func(t *testing.T) {
				crossValidate(t, input, ver17)
			})
		}
	}
}

func TestGenerated_UnaryPrefixCombinations(t *testing.T) {
	// Test prefix operators combined with binary operators
	binOps := []string{"&", "|", "->", "=", "<", "+", "*"}
	for _, op := range binOps {
		// ~a OP b
		crossValidate(t, fmt.Sprintf("~a %s b", op), ver17)
		// a OP ~b
		crossValidate(t, fmt.Sprintf("a %s ~b", op), ver17)
		// ~~a OP b
		crossValidate(t, fmt.Sprintf("~~a %s b", op), ver17)
	}
}

func TestGenerated_QuantifierScoping(t *testing.T) {
	// Test quantifier body scoping with various operators
	ops := []string{"&", "|", "->", "<->", "="}
	for _, op := range ops {
		input := fmt.Sprintf("forall X. p(X) %s q(X)", op)
		t.Run("forall+"+op, func(t *testing.T) {
			crossValidate(t, input, ver17)
		})
		input2 := fmt.Sprintf("exists X. p(X) %s q(X)", op)
		t.Run("exists+"+op, func(t *testing.T) {
			crossValidate(t, input2, ver17)
		})
	}
}

func TestGenerated_TripleOperatorChains(t *testing.T) {
	// Test three-operator chains: a OP b OP c OP d
	ops := []string{"&", "|", "->", "+", "*"}
	for _, op := range ops {
		input := fmt.Sprintf("a %s b %s c %s d", op, op, op)
		t.Run(op+"+"+op+"+"+op, func(t *testing.T) {
			crossValidate(t, input, ver17)
		})
	}
}

func TestGenerated_ParenthesizedVariants(t *testing.T) {
	// Test that explicit parenthesization overrides precedence
	cases := []string{
		"(a & b) | c",
		"a & (b | c)",
		"(a | b) & c",
		"a | (b & c)",
		"(a -> b) & c",
		"a -> (b & c)",
		"(a + b) * c",
		"a + (b * c)",
		"(a = b) & (c = d)",
		"a = (b & c)",  // = groups tighter than & but paren overrides
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			crossValidate(t, c, ver17)
		})
	}
}

func TestGenerated_FunctionApplications(t *testing.T) {
	cases := []string{
		"f(a & b)",
		"f(a, b) = g(c, d)",
		"f(a + b, c * d)",
		"f(g(h(x)))",
		"~f(x) & g(y)",
		"f(x) -> g(y)",
		"forall X. f(X) = g(X)",
		"f(x, y, z) & g(a, b)",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			crossValidate(t, c, ver17)
		})
	}
}

// =========================================================================
// v1.6 cross-validation tests
// =========================================================================

func TestCrossValidation_V16_BasicFormulas(t *testing.T) {
	formulas := []string{
		"foo", "X", "true", "false",
		"~p", "a & b", "a | b", "a -> b", "a <-> b",
		"f(x)", "f(x, y)",
		"x = y", "x ~= y", "x < y", "x <= y",
		"x + y", "x * y",
		"forall X. p(X)", "exists X. p(X)",
		"globally p", "eventually p",
	}
	for _, f := range formulas {
		crossValidate(t, f, ver16)
	}
}

func TestCrossValidation_V16_PrecedenceDifferences(t *testing.T) {
	cases := []struct {
		input string
		desc  string
	}{
		{"a & b", "basic AND"},
		{"a | b", "basic OR"},
		{"a & b | c", "AND vs OR"},
		{"~a & b", "NOT vs AND"},
		{"~a = b", "NOT vs EQ in v1.6"},
		{"a -> b -> c", "ARROW associativity in v1.6"},
		{"a = b & c = d", "EQ vs AND in v1.6"},
		{"a + b = c", "PLUS vs EQ in v1.6"},
		{"a * b + c", "TIMES vs PLUS in v1.6"},
		{"forall X. p(X) & q(X)", "FORALL scoping in v1.6"},
		{"globally a & b", "GLOBALLY vs AND in v1.6"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			crossValidate(t, tc.input, ver16)
		})
	}
}

func TestCrossValidation_V16_OperatorPairs(t *testing.T) {
	ops := []string{"&", "|", "->", "<->", "=", "<", "<=", ">", ">=", "+", "-", "*", "/"}
	for _, op1 := range ops {
		for _, op2 := range ops {
			input := fmt.Sprintf("a %s b %s c", op1, op2)
			t.Run(input, func(t *testing.T) {
				crossValidate(t, input, ver16)
			})
		}
	}
}

// =========================================================================
// v1.2 cross-validation tests
// =========================================================================

func TestCrossValidation_V12_BasicFormulas(t *testing.T) {
	formulas := []string{
		"foo", "X", "true", "false",
		"~p", "a & b", "a | b", "a <-> b",
		"f(x)", "f(x, y)",
		"x = y", "x ~= y", "x < y", "x <= y",
		"forall X. p(X)", "exists X. p(X)",
	}
	for _, f := range formulas {
		crossValidate(t, f, ver12)
	}
}

func TestCrossValidation_V12_PrecedenceDifferences(t *testing.T) {
	// CRITICAL: In v1.2, TILDA binds LOOSER than comparison operators!
	cases := []struct {
		input string
		desc  string
	}{
		{"~a = b", "NOT vs EQ (v1.2: NOT looser!)"},
		{"~a < b", "NOT vs LT (v1.2: NOT looser!)"},
		{"~a <= b", "NOT vs LE (v1.2: NOT looser!)"},
		{"~a > b", "NOT vs GT (v1.2: NOT looser!)"},
		{"~a >= b", "NOT vs GE (v1.2: NOT looser!)"},
		{"~a & b", "NOT vs AND"},
		{"~a | b", "NOT vs OR"},
		{"a & b | c", "AND vs OR"},
		{"forall X. p(X) & q(X)", "FORALL scoping"},
		{"a = b & c = d", "EQ vs AND"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			crossValidate(t, tc.input, ver12)
		})
	}
}

func TestCrossValidation_V12_OperatorPairs(t *testing.T) {
	ops := []string{"&", "|", "<->", "=", "<", "<=", ">", ">="}
	for _, op1 := range ops {
		for _, op2 := range ops {
			input := fmt.Sprintf("a %s b %s c", op1, op2)
			t.Run(input, func(t *testing.T) {
				crossValidate(t, input, ver12)
			})
		}
	}
}

// =========================================================================
// Targeted regression tests for known hand-written parser bugs.
// Each test documents a specific bug found by LALR cross-validation.
// These should FAIL until the hand-written parser is fixed.
// =========================================================================

// BUG 1: In v1.6, ARROW is not in the precedence table so PLY defaults
// to shift (right-associative). The hand-written parser always uses
// left-associativity (correct for v1.7+, wrong for v1.6).
func TestBug_V16_ArrowAssociativity(t *testing.T) {
	hw, _ := parseHW("a -> b -> c", ver16)
	lalr, _ := Parse("a -> b -> c", ver16)
	hwS := astShape(hw)
	lalrS := astShape(lalr)
	// LALR (correct for v1.6): a -> (b -> c) = Implies(a, Implies(b, c))
	// Hand-written (wrong for v1.6): (a -> b) -> c = Implies(Implies(a, b), c)
	if hwS == lalrS {
		t.Log("BUG FIXED: v1.6 ARROW associativity now matches LALR")
	} else {
		t.Errorf("BUG PRESENT: v1.6 ARROW associativity\n  hand-written: %s\n  LALR:         %s\n  Expected LALR result (right-assoc for v1.6)", hwS, lalrS)
	}
}

// BUG 2: In v1.6, ARROW/IFF are not in the precedence table, so they
// have no defined precedence relative to AND/OR. In the LALR grammar
// they only appear as fmla rules, meaning AND/OR (which are also fmla
// rules at the same level) can appear inside ARROW/IFF arguments.
// The hand-written parser gives ARROW explicit lower precedence than
// AND, so "a & b -> c" groups as "(a & b) -> c". The LALR grammar
// produces "a & (b -> c)" because ARROW has no precedence entry.
func TestBug_V16_ArrowVsAndPrecedence(t *testing.T) {
	hw, _ := parseHW("a & b -> c", ver16)
	lalr, _ := Parse("a & b -> c", ver16)
	hwS := astShape(hw)
	lalrS := astShape(lalr)
	if hwS == lalrS {
		t.Log("BUG FIXED: v1.6 ARROW vs AND precedence now matches LALR")
	} else {
		t.Errorf("BUG PRESENT: v1.6 ARROW vs AND precedence\n  hand-written: %s\n  LALR:         %s", hwS, lalrS)
	}
}

// BUG 3: Same as BUG 2 but for IFF (<->).
func TestBug_V16_IffVsAndPrecedence(t *testing.T) {
	hw, _ := parseHW("a & b <-> c", ver16)
	lalr, _ := Parse("a & b <-> c", ver16)
	hwS := astShape(hw)
	lalrS := astShape(lalr)
	if hwS == lalrS {
		t.Log("BUG FIXED: v1.6 IFF vs AND precedence now matches LALR")
	} else {
		t.Errorf("BUG PRESENT: v1.6 IFF vs AND precedence\n  hand-written: %s\n  LALR:         %s", hwS, lalrS)
	}
}

// BUG 4: In v1.6, comparison operators (=, <, <=, >, >=) only appear
// in the rule "fmla : term relop term". This means chained comparisons
// like "a = b = c" are SYNTAX ERRORS — the result of "a = b" is a fmla,
// and "fmla = fmla" has no rule. The hand-written parser accepts them
// because it treats = as a regular binary operator with no restriction.
func TestBug_V16_ChainedComparisonsRejected(t *testing.T) {
	_, lalrErr := Parse("a = b = c", ver16)
	_, hwErr := parseHW("a = b = c", ver16)

	if lalrErr == nil {
		t.Fatal("LALR should reject chained comparisons in v1.6")
	}
	if hwErr != nil {
		t.Log("BUG FIXED: hand-written parser now rejects chained comparisons in v1.6")
	} else {
		t.Errorf("BUG PRESENT: hand-written parser accepts 'a = b = c' in v1.6 but it should be a syntax error\n  LALR correctly rejects with: %v", lalrErr)
	}
}

// Additional chained-comparison variants for v1.6
func TestBug_V16_ChainedComparisonsVariants(t *testing.T) {
	chains := []string{
		"a = b < c",
		"a < b = c",
		"a < b < c",
		"a <= b >= c",
		"a > b > c",
	}
	for _, input := range chains {
		t.Run(input, func(t *testing.T) {
			_, lalrErr := Parse(input, ver16)
			_, hwErr := parseHW(input, ver16)
			if lalrErr == nil {
				t.Fatalf("LALR should reject %q in v1.6", input)
			}
			if hwErr != nil {
				t.Logf("BUG FIXED: hand-written rejects %q in v1.6", input)
			} else {
				t.Errorf("BUG PRESENT: hand-written accepts %q in v1.6 but should be syntax error", input)
			}
		})
	}
}
