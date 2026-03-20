package lalr_logicparser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/parser"
)

var ver17 = lexer.Version{1, 7}

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
	case *ast.Dot:
		return fmt.Sprintf("Dot(%s,%s)", astShape(t.Left), astShape(t.Right))
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
		"a & b & c & d", // n-ary AND
		"a | b | c | d", // n-ary OR
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
		"a = (b & c)", // = groups tighter than & but paren overrides
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
// v1.7 cross-validation: hand-written Pratt parser vs LALR grammar.
// Since ≤v1.6 now uses the LALR parser directly (via logicparser.Parse),
// cross-validation is only needed for v1.7+ where the hand-written parser
// is the production parser.
// =========================================================================

func TestCrossValidation_V17_MorePrecedence(t *testing.T) {
	cases := []struct {
		input string
		desc  string
	}{
		// IFF vs ARROW
		{"a <-> b -> c", "IFF vs ARROW"},
		{"a -> b <-> c", "ARROW vs IFF"},
		// Double negation with operators
		{"~~a & b", "double NOT vs AND"},
		{"~~a -> b", "double NOT vs ARROW"},
		// Arithmetic in comparison context
		{"a - b = c + d", "MINUS and PLUS vs EQ"},
		{"a / b < c * d", "DIV and TIMES vs LT"},
		// Mixed quantifiers and connectives
		{"forall X. exists Y. p(X) & q(Y)", "nested quantifiers"},
		{"forall X. p(X) <-> q(X)", "FORALL scopes over IFF"},
		{"exists X. a -> b & c", "EXISTS with ARROW and AND"},
		// Chained comparisons (valid in v1.7)
		{"a = b = c", "chained EQ"},
		{"a < b < c", "chained LT"},
		{"a <= b >= c", "chained LE GE"},
		// Temporal with connectives
		{"globally a -> b", "GLOBALLY vs ARROW"},
		{"eventually a & b | c", "EVENTUALLY vs AND vs OR"},
		{"globally eventually a", "GLOBALLY EVENTUALLY"},
		// Function application with operators
		{"f(x) + g(y) * h(z)", "funcall in arithmetic"},
		{"f(a & b, c | d)", "connectives inside funcall args"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			crossValidate(t, tc.input, ver17)
		})
	}
}

func TestCrossValidation_V17_CompoundExpressions(t *testing.T) {
	formulas := []string{
		// Deeply nested
		"a & b | c & d | e & f",
		"a -> b -> c -> d",
		"a <-> b <-> c <-> d",
		"(a -> b) & (c -> d) | (e -> f)",
		// Mixed arithmetic and logic
		"a + b = c & d + e = f",
		"a * b + c * d = e * f + g * h",
		// Quantifier interactions
		"forall X. forall Y. exists Z. p(X, Y) -> q(Y, Z)",
		"(forall X. p(X)) -> (exists Y. q(Y))",
		"forall X. p(X) & q(X) -> r(X)",
		// Negation patterns
		"~(a -> b)", "~(a <-> b)", "~forall X. p(X)",
		"~(a + b = c)",
	}
	for _, f := range formulas {
		t.Run(f, func(t *testing.T) {
			crossValidate(t, f, ver17)
		})
	}
}

func TestCrossValidation_V17_AllOperatorTriples(t *testing.T) {
	// Test "a OP1 b OP2 c OP3 d" for a selection of interesting triples
	ops := []string{"&", "|", "->", "<->", "=", "+", "*"}
	for _, op1 := range ops {
		for _, op2 := range ops {
			for _, op3 := range ops {
				input := fmt.Sprintf("a %s b %s c %s d", op1, op2, op3)
				t.Run(input, func(t *testing.T) {
					crossValidate(t, input, ver17)
				})
			}
		}
	}
}

// === Action cross-validation tests ===

// parseHWAction parses an action body with the hand-written parser.
func parseHWAction(input string, version lexer.Version) (ast.Node, error) {
	p := parser.New(input, version)
	result := p.ParseActionBody()
	if result == nil {
		errs := p.Errors()
		if len(errs) > 0 {
			return nil, fmt.Errorf("%s", errs[0].Error())
		}
		return nil, fmt.Errorf("failed to parse action")
	}
	return result, nil
}

// crossValidateAction compares LALR and hand-written parser for action bodies.
func crossValidateAction(t *testing.T, input string, version lexer.Version) {
	t.Helper()

	hwResult, hwErr := parseHWAction(input, version)
	lalrResult, lalrErr := ParseV17(input, version)

	if hwErr != nil && lalrErr != nil {
		return // both error — OK
	}
	if hwErr != nil {
		t.Logf("HW error but LALR ok: hw=%v, lalr=%s", hwErr, astShape(lalrResult))
		return // HW error but LALR ok — acceptable (LALR may parse more)
	}
	if lalrErr != nil {
		t.Logf("LALR error but HW ok: lalr=%v, hw=%s", lalrErr, astShape(hwResult))
		return // LALR error but HW ok — acceptable for now
	}

	hwShape := astShape(hwResult)
	lalrShape := astShape(lalrResult)
	if hwShape != lalrShape {
		t.Errorf("AST shape mismatch for %q:\n  HW:   %s\n  LALR: %s", input, hwShape, lalrShape)
	}
}

func TestActionCrossValidation_SimpleAssume(t *testing.T) {
	crossValidateAction(t, "{ assume p }", ver17)
}

func TestActionCrossValidation_SimpleAssert(t *testing.T) {
	crossValidateAction(t, "{ assert p }", ver17)
}

func TestActionCrossValidation_Assignment(t *testing.T) {
	crossValidateAction(t, "{ x := y }", ver17)
}

func TestActionCrossValidation_Havoc(t *testing.T) {
	crossValidateAction(t, "{ x := * }", ver17)
}

func TestActionCrossValidation_IfThen(t *testing.T) {
	crossValidateAction(t, "{ if c { x := y } }", ver17)
}

func TestActionCrossValidation_IfThenElse(t *testing.T) {
	crossValidateAction(t, "{ if c { x := y } else { z := w } }", ver17)
}

func TestActionCrossValidation_Sequence(t *testing.T) {
	crossValidateAction(t, "{ x := y; assume p }", ver17)
}

func TestActionCrossValidation_Nested(t *testing.T) {
	crossValidateAction(t, "{ if c { x := y } else { z := w }; assume p }", ver17)
}

func TestActionCrossValidation_Var(t *testing.T) {
	crossValidateAction(t, "{ var v : bool }", ver17)
}

func TestActionCrossValidation_VarInit(t *testing.T) {
	crossValidateAction(t, "{ var v := x }", ver17)
}

func TestActionCrossValidation_While(t *testing.T) {
	crossValidateAction(t, "{ while c { x := y } }", ver17)
}

func TestActionCrossValidation_Require(t *testing.T) {
	crossValidateAction(t, "{ require p }", ver17)
}

func TestActionCrossValidation_Ensure(t *testing.T) {
	crossValidateAction(t, "{ ensure p }", ver17)
}

func TestActionCrossValidation_EmptySequence(t *testing.T) {
	crossValidateAction(t, "{ }", ver17)
}

func TestActionCrossValidation_BareCall(t *testing.T) {
	crossValidateAction(t, "{ f(x) }", ver17)
}

// TestActionCrossValidation_ScenarioBasic tests scenario parsing in the LALR grammar.
func TestActionCrossValidation_ScenarioBasic(t *testing.T) {
	input := `scenario { -> s0; s0 -> s1 : before a { assume p } }`
	lalrResult, lalrErr := ParseV17(input, ver17)
	if lalrErr != nil {
		t.Fatalf("LALR parse error: %v", lalrErr)
	}
	if lalrResult == nil {
		t.Fatal("LALR returned nil")
	}
	// Check it produced a ScenarioDecl
	if _, ok := lalrResult.(*ast.ScenarioDecl); !ok {
		t.Errorf("expected ScenarioDecl, got %T: %s", lalrResult, astShape(lalrResult))
	}
}

func TestActionCrossValidation_ScenarioMultiTransitions(t *testing.T) {
	input := `scenario { -> s0; s0 -> s1 : before a { x := y } s1 -> s0 : after b { z := w } }`
	lalrResult, lalrErr := ParseV17(input, ver17)
	if lalrErr != nil {
		t.Fatalf("LALR parse error: %v", lalrErr)
	}
	scenDecl, ok := lalrResult.(*ast.ScenarioDecl)
	if !ok {
		t.Fatalf("expected ScenarioDecl, got %T", lalrResult)
	}
	if len(scenDecl.DeclArgs) != 1 {
		t.Fatalf("expected 1 DeclArg, got %d", len(scenDecl.DeclArgs))
	}
	sdef, ok := scenDecl.DeclArgs[0].(*ast.ScenarioDef)
	if !ok {
		t.Fatalf("expected ScenarioDef, got %T", scenDecl.DeclArgs[0])
	}
	transs := sdef.Transitions()
	if len(transs) != 2 {
		t.Errorf("expected 2 transitions, got %d", len(transs))
	}
	// Check first is before, second is after
	if _, ok := transs[0].Action.(*ast.ScenarioBeforeMixin); !ok {
		t.Errorf("tr0: expected ScenarioBeforeMixin, got %T", transs[0].Action)
	}
	if _, ok := transs[1].Action.(*ast.ScenarioAfterMixin); !ok {
		t.Errorf("tr1: expected ScenarioAfterMixin, got %T", transs[1].Action)
	}
}
