package lalr_logicparser

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// =========================================================================
// Randomized grammar-guided cross-validation: generate random formulas
// from the v1.7 LALR grammar productions and verify that the hand-written
// Pratt parser produces identical ASTs.
// =========================================================================

// formulaGen generates random formulas by walking grammar productions.
type formulaGen struct {
	rng   *rand.Rand
	depth int    // current nesting depth
	max   int    // max nesting depth
	buf   strings.Builder
}

func newFormulaGen(seed int64, maxDepth int) *formulaGen {
	return &formulaGen{
		rng: rand.New(rand.NewSource(seed)),
		max: maxDepth,
	}
}

func (g *formulaGen) String() string {
	return g.buf.String()
}

func (g *formulaGen) reset() {
	g.buf.Reset()
	g.depth = 0
}

// pick returns a random int in [0, n).
func (g *formulaGen) pick(n int) int {
	return g.rng.Intn(n)
}

// symbols used as leaf terms
var symbols = []string{"a", "b", "c", "d", "e", "f", "g", "h", "p", "q", "r", "s"}
var variables = []string{"X", "Y", "Z", "W"}

func (g *formulaGen) symbol() string {
	return symbols[g.pick(len(symbols))]
}

func (g *formulaGen) variable() string {
	return variables[g.pick(len(variables))]
}

// Generate a complete formula string.
func (g *formulaGen) generate() string {
	g.reset()
	g.term()
	return g.buf.String()
}

// term generates a random term according to the v1.7 grammar.
func (g *formulaGen) term() {
	g.depth++
	defer func() { g.depth-- }()

	// At max depth, only produce leaves (atoms, variables, true, false).
	if g.depth >= g.max {
		g.leaf()
		return
	}

	// Weight productions to get a good mix.
	// Leaves get higher weight to keep formulas from exploding.
	switch g.pick(20) {
	case 0, 1, 2, 3, 4:
		// leaf: symbol, variable, true, false, funcall
		g.leaf()
	case 5, 6:
		// binary arithmetic: term OP term
		g.binaryArith()
	case 7, 8:
		// comparison: term RELOP term
		g.comparison()
	case 9, 10:
		// boolean binary: term BOOLOP term
		g.boolBinary()
	case 11:
		// arrow: term -> term
		g.term()
		g.buf.WriteString(" -> ")
		g.term()
	case 12:
		// iff: term <-> term
		g.term()
		g.buf.WriteString(" <-> ")
		g.term()
	case 13:
		// negation: ~term
		g.buf.WriteString("~")
		g.term()
	case 14:
		// parenthesized: (term)
		g.buf.WriteString("(")
		g.term()
		g.buf.WriteString(")")
	case 15:
		// forall: forall V. term
		g.buf.WriteString("forall ")
		g.buf.WriteString(g.variable())
		g.buf.WriteString(". ")
		g.term()
	case 16:
		// exists: exists V. term
		g.buf.WriteString("exists ")
		g.buf.WriteString(g.variable())
		g.buf.WriteString(". ")
		g.term()
	case 17:
		// globally: globally term
		g.buf.WriteString("globally ")
		g.term()
	case 18:
		// eventually: eventually term
		g.buf.WriteString("eventually ")
		g.term()
	case 19:
		// tildaeq: term ~= term
		g.term()
		g.buf.WriteString(" ~= ")
		g.term()
	}
}

func (g *formulaGen) leaf() {
	switch g.pick(6) {
	case 0, 1:
		// bare symbol
		g.buf.WriteString(g.symbol())
	case 2:
		// variable
		g.buf.WriteString(g.variable())
	case 3:
		// true
		g.buf.WriteString("true")
	case 4:
		// false
		g.buf.WriteString("false")
	case 5:
		// function application: sym(args...)
		g.buf.WriteString(g.symbol())
		g.buf.WriteString("(")
		nArgs := 1 + g.pick(3) // 1-3 args
		for i := 0; i < nArgs; i++ {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.term()
		}
		g.buf.WriteString(")")
	}
}

func (g *formulaGen) binaryArith() {
	ops := []string{" + ", " - ", " * ", " / "}
	g.term()
	g.buf.WriteString(ops[g.pick(len(ops))])
	g.term()
}

func (g *formulaGen) comparison() {
	ops := []string{" = ", " < ", " <= ", " > ", " >= "}
	g.term()
	g.buf.WriteString(ops[g.pick(len(ops))])
	g.term()
}

func (g *formulaGen) boolBinary() {
	ops := []string{" & ", " | "}
	g.term()
	g.buf.WriteString(ops[g.pick(len(ops))])
	g.term()
}

// TestRandomCrossValidation_V17 generates random formulas from the grammar
// and cross-validates the hand-written parser against the LALR parser.
func TestRandomCrossValidation_V17(t *testing.T) {
	const (
		numFormulas = 5000
		maxDepth    = 6
		seed        = 42
	)

	gen := newFormulaGen(seed, maxDepth)
	mismatches := 0
	lalrFails := 0
	hwFails := 0
	bothFail := 0
	ok := 0

	for i := 0; i < numFormulas; i++ {
		formula := gen.generate()

		hw, hwErr := parseHW(formula, ver17)
		lalr, lalrErr := Parse(formula, ver17)

		if hwErr != nil && lalrErr != nil {
			bothFail++
			continue
		}
		if hwErr != nil {
			hwFails++
			t.Errorf("seed %d formula #%d %q: hand-written failed (%v) but LALR succeeded: %s",
				seed, i, formula, hwErr, astShape(lalr))
			continue
		}
		if lalrErr != nil {
			lalrFails++
			t.Errorf("seed %d formula #%d %q: LALR failed (%v) but hand-written succeeded: %s",
				seed, i, formula, lalrErr, astShape(hw))
			continue
		}

		hwShape := astShape(hw)
		lalrShape := astShape(lalr)
		if hwShape != lalrShape {
			mismatches++
			if mismatches <= 20 { // cap error output
				t.Errorf("seed %d formula #%d %q: MISMATCH\n  hand-written: %s\n  LALR:         %s",
					seed, i, formula, hwShape, lalrShape)
			}
			continue
		}
		ok++
	}

	t.Logf("Results: %d ok, %d mismatches, %d hw-only-fail, %d lalr-only-fail, %d both-fail (of %d total)",
		ok, mismatches, hwFails, lalrFails, bothFail, numFormulas)
}

// TestRandomCrossValidation_V17_MultiSeed runs multiple seeds to increase
// coverage beyond a single deterministic run.
func TestRandomCrossValidation_V17_MultiSeed(t *testing.T) {
	const (
		numSeeds    = 20
		perSeed     = 1000
		maxDepth    = 5
	)

	totalMismatches := 0
	for seed := int64(0); seed < numSeeds; seed++ {
		gen := newFormulaGen(seed, maxDepth)
		for i := 0; i < perSeed; i++ {
			formula := gen.generate()

			hw, hwErr := parseHW(formula, ver17)
			lalr, lalrErr := Parse(formula, ver17)

			if hwErr != nil || lalrErr != nil {
				if hwErr != nil && lalrErr != nil {
					continue // both fail, OK
				}
				if hwErr != nil {
					t.Errorf("seed %d #%d %q: HW failed (%v), LALR ok: %s",
						seed, i, formula, hwErr, astShape(lalr))
				} else {
					t.Errorf("seed %d #%d %q: LALR failed (%v), HW ok: %s",
						seed, i, formula, lalrErr, astShape(hw))
				}
				totalMismatches++
				continue
			}

			hwShape := astShape(hw)
			lalrShape := astShape(lalr)
			if hwShape != lalrShape {
				totalMismatches++
				if totalMismatches <= 20 {
					t.Errorf("seed %d #%d %q: MISMATCH\n  HW:   %s\n  LALR: %s",
						seed, i, formula, hwShape, lalrShape)
				}
			}
		}
	}
	if totalMismatches > 20 {
		t.Errorf("... and %d more mismatches (showing first 20)", totalMismatches-20)
	}
	t.Logf("Tested %d random formulas across %d seeds", numSeeds*perSeed, numSeeds)
}

// FuzzCrossValidation_V17 is a Go fuzz target that can be run with:
//   go test -fuzz=FuzzCrossValidation_V17 -fuzztime=30s ./lalr_logicparser/
// It generates random formula strings and cross-validates both parsers.
func FuzzCrossValidation_V17(f *testing.F) {
	// Seed corpus with representative formulas
	seeds := []string{
		"a", "X", "true", "false",
		"~p", "a & b", "a | b", "a -> b", "a <-> b",
		"f(x)", "f(x, y)", "x = y", "x + y",
		"forall X. p(X)", "exists X. p(X)",
		"globally p", "eventually p",
		"a & b | c", "a -> b -> c",
		"a + b * c = d", "~(a & b) | c -> d",
		"forall X. exists Y. p(X) & q(Y) -> r(X, Y)",
		"a ~= b", "a * b + c * d = e",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Skip inputs with characters that aren't valid Ivy tokens —
		// the lexer may produce surprising results with random bytes.
		for _, c := range input {
			if c > 127 {
				t.Skip("non-ASCII")
			}
		}
		if len(input) > 200 {
			t.Skip("too long")
		}

		hw, hwErr := parseHW(input, ver17)
		lalr, lalrErr := Parse(input, ver17)

		if hwErr != nil && lalrErr != nil {
			return // both fail, fine
		}
		if hwErr != nil {
			t.Errorf("%q: HW failed (%v) but LALR ok: %s", input, hwErr, astShape(lalr))
			return
		}
		if lalrErr != nil {
			// LALR grammar is more restrictive for some constructs
			// (e.g. curried application f(x)(y)), so LALR-only failures
			// are expected for inputs outside the grammar.
			return
		}

		hwShape := astShape(hw)
		lalrShape := astShape(lalr)
		if hwShape != lalrShape {
			t.Errorf("%q: MISMATCH\n  HW:   %s\n  LALR: %s", input, hwShape, lalrShape)
		}
	})
}

// TestRandomCrossValidation_V17_DeepNesting specifically tests deeply
// nested formulas to stress-test both parsers' recursion handling.
func TestRandomCrossValidation_V17_DeepNesting(t *testing.T) {
	// Generate deeply nested but simple structures
	cases := []struct {
		name string
		gen  func(depth int) string
	}{
		{"nested_and", func(d int) string {
			s := "a"
			for i := 0; i < d; i++ {
				s = fmt.Sprintf("(%s & %s)", s, symbols[i%len(symbols)])
			}
			return s
		}},
		{"nested_or", func(d int) string {
			s := "a"
			for i := 0; i < d; i++ {
				s = fmt.Sprintf("(%s | %s)", s, symbols[i%len(symbols)])
			}
			return s
		}},
		{"nested_arrow", func(d int) string {
			s := "a"
			for i := 0; i < d; i++ {
				s = fmt.Sprintf("(%s -> %s)", s, symbols[i%len(symbols)])
			}
			return s
		}},
		{"nested_neg", func(d int) string {
			s := "a"
			for i := 0; i < d; i++ {
				s = "~" + s
			}
			return s
		}},
		{"nested_forall", func(d int) string {
			s := "p(X)"
			for i := 0; i < d; i++ {
				v := variables[i%len(variables)]
				s = fmt.Sprintf("forall %s. %s", v, s)
			}
			return s
		}},
		{"nested_funcall", func(d int) string {
			s := "x"
			for i := 0; i < d; i++ {
				fn := symbols[i%len(symbols)]
				s = fmt.Sprintf("%s(%s)", fn, s)
			}
			return s
		}},
		{"nested_arith", func(d int) string {
			s := "a"
			ops := []string{" + ", " * ", " - "}
			for i := 0; i < d; i++ {
				s = fmt.Sprintf("(%s%s%s)", s, ops[i%len(ops)], symbols[i%len(symbols)])
			}
			return s
		}},
	}

	depths := []int{5, 10, 20, 50}
	for _, tc := range cases {
		for _, d := range depths {
			name := fmt.Sprintf("%s_depth%d", tc.name, d)
			t.Run(name, func(t *testing.T) {
				formula := tc.gen(d)
				crossValidate(t, formula, ver17)
			})
		}
	}
}
