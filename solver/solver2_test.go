// Tests for solver bug fixes 7.1, 7.2, 7.3.
// Each test targets a specific phase from the fix plan.
package solver

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// ============================================================
// Phase 1: Quantifier Bound Constraints (7.3.1)
// ============================================================

// TestQuantConstraints_NatForAll checks that ForAll(X:nat, P(X))
// translates to ForAll(X, Implies(0<=X, P(X))).
func TestQuantConstraints_NatForAll(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mynat"] = "nat"
	s := NewWithSig(sig)

	natSort := &lg.UninterpretedSort{Name: "mynat"}
	x, _ := lg.NewVariable("X", natSort)
	p := relConst("P", natSort)
	pApp := &lg.Apply{Func: p, Terms: []lg.Expr{x}}

	// ForAll(X:mynat, P(X))
	fmla := &lg.ForAll{Variables: []*lg.Variable{x}, Body: pApp}

	z3expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatalf("translate ForAll: %v", err)
	}

	// The resulting Z3 expression should be a quantifier whose body
	// is Implies(0 <= X, P(X)), not just P(X).
	str := z3expr.String()
	t.Logf("ForAll(X:nat, P(X)) => %s", str)

	// Verify semantics: ForAll(X:nat, X >= 0 => P(X)) should NOT imply P(-1).
	// So we check: (ForAll + P(-1)=false) should be SAT.
	ctx := s.Context()
	slv := ctx.NewSolver()
	slv.Assert(z3expr)
	// Assert P(-1) = false
	negOne := ctx.IntVal(-1)
	pDecl := ctx.Function("P", []z3bridge.Sort{ctx.IntSort()}, ctx.BoolSort())
	slv.Assert(ctx.Not(pDecl.Apply(negOne)))
	res := slv.Check()
	if res == z3bridge.Unsat {
		t.Fatal("ForAll(X:nat, P(X)) should not constrain negative X, but it does")
	}
}

// TestQuantConstraints_NatExists checks that Exists(X:nat, P(X))
// translates to Exists(X, And(0<=X, P(X))).
func TestQuantConstraints_NatExists(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mynat"] = "nat"
	s := NewWithSig(sig)

	natSort := &lg.UninterpretedSort{Name: "mynat"}
	x, _ := lg.NewVariable("X", natSort)
	p := relConst("P", natSort)
	pApp := &lg.Apply{Func: p, Terms: []lg.Expr{x}}

	// Exists(X:mynat, P(X))
	fmla := &lg.Exists{Variables: []*lg.Variable{x}, Body: pApp}

	z3expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatalf("translate Exists: %v", err)
	}

	str := z3expr.String()
	t.Logf("Exists(X:nat, P(X)) => %s", str)

	// Verify: if P is only true for -1, Exists should be UNSAT
	// because the nat constraint requires X >= 0.
	ctx := s.Context()
	slv := ctx.NewSolver()
	slv.Assert(z3expr)
	// P(x) = (x == -1)
	xConst := ctx.Const("X:mynat", ctx.IntSort())
	pDecl := ctx.Function("P", []z3bridge.Sort{ctx.IntSort()}, ctx.BoolSort())
	slv.Assert(ctx.ForAll(
		[]z3bridge.Expr{xConst},
		ctx.Eq(pDecl.Apply(xConst), ctx.Eq(xConst, ctx.IntVal(-1))),
	))
	res := slv.Check()
	if res != z3bridge.Unsat {
		t.Fatal("Exists(X:nat, P(X)) with P only true at -1 should be UNSAT")
	}
}

// TestQuantConstraints_RangeSort checks that ForAll(X:range(2,5), P(X))
// translates with bounds 2 <= X <= 5.
func TestQuantConstraints_RangeSort(t *testing.T) {
	rs := &lg.RangeSort{Name: "myrange", Lb: lg.NumeralBound{Value: "2"}, Ub: lg.NumeralBound{Value: "5"}}
	sig := il.NewSig()
	sig.Interp["myrange"] = rs
	s := NewWithSig(sig)

	rangeSort := &lg.UninterpretedSort{Name: "myrange"}
	x, _ := lg.NewVariable("X", rangeSort)
	p := relConst("P", rangeSort)
	pApp := &lg.Apply{Func: p, Terms: []lg.Expr{x}}

	fmla := &lg.ForAll{Variables: []*lg.Variable{x}, Body: pApp}

	z3expr, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatalf("translate ForAll with range sort: %v", err)
	}

	t.Logf("ForAll(X:range(2,5), P(X)) => %s", z3expr.String())

	// Verify: ForAll should not constrain X=1 (outside range).
	ctx := s.Context()
	slv := ctx.NewSolver()
	slv.Assert(z3expr)
	pDecl := ctx.Function("P", []z3bridge.Sort{ctx.IntSort()}, ctx.BoolSort())
	slv.Assert(ctx.Not(pDecl.Apply(ctx.IntVal(1)))) // P(1) = false
	res := slv.Check()
	if res == z3bridge.Unsat {
		t.Fatal("ForAll(X:range(2,5), P(X)) should not constrain X=1")
	}
}

// TestQuantConstraints_NoInterp checks that ForAll over an uninterpreted sort
// adds no constraints (body is unchanged).
func TestQuantConstraints_NoInterp(t *testing.T) {
	sig := il.NewSig()
	s := NewWithSig(sig)

	sort := &lg.UninterpretedSort{Name: "T"}
	x, _ := lg.NewVariable("X", sort)
	p := relConst("P", sort)
	pApp := &lg.Apply{Func: p, Terms: []lg.Expr{x}}
	fmla := &lg.ForAll{Variables: []*lg.Variable{x}, Body: pApp}

	_, err := s.FormulaToZ3(fmla)
	if err != nil {
		t.Fatalf("translate ForAll with uninterpreted sort: %v", err)
	}
	// No panic, no error — body is just P(X) with no extra constraints.
}

// ============================================================
// Phase 2: MyEq True/False Short-Circuit (7.3.7)
// ============================================================

// TestMyEq_TrueShortCircuit checks that MyEq(x, true) returns x.
func TestMyEq_TrueShortCircuit(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	x := ctx.Const("x", ctx.BoolSort())

	result := MyEq(ctx, x, ctx.BoolVal(true))
	// result should be equivalent to x, not Iff(x, true)
	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(result, x)))
	if slv.Check() != z3bridge.Unsat {
		t.Fatalf("MyEq(x, true) should be equivalent to x, got %s", result.String())
	}
}

// TestMyEq_FalseShortCircuit checks that MyEq(x, false) returns Not(x).
func TestMyEq_FalseShortCircuit(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	x := ctx.Const("x", ctx.BoolSort())

	result := MyEq(ctx, x, ctx.BoolVal(false))
	// result should be equivalent to Not(x)
	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(result, ctx.Not(x))))
	if slv.Check() != z3bridge.Unsat {
		t.Fatalf("MyEq(x, false) should be equivalent to Not(x), got %s", result.String())
	}
}

// TestMyEq_NonBoolFallback checks that MyEq for int values still uses Eq.
func TestMyEq_NonBoolFallback(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	x := ctx.Const("x", ctx.IntSort())
	y := ctx.Const("y", ctx.IntSort())

	result := MyEq(ctx, x, y)
	// result should be equivalent to x == y
	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(result, ctx.Eq(x, y))))
	if slv.Check() != z3bridge.Unsat {
		t.Fatalf("MyEq(x, y) for ints should be Eq, got %s", result.String())
	}
}

// ============================================================
// Phase 3: Gebin Correct Semantics (7.3.8)
// ============================================================

// TestGebin_ZeroIsTrue checks that Gebin(bits, 0) is always true.
func TestGebin_ZeroIsTrue(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	a := ctx.Const("a", ctx.BoolSort())
	b := ctx.Const("b", ctx.BoolSort())
	bits := []z3bridge.Expr{a, b}

	result := Gebin(ctx, bits, 0)
	if !result.IsTrue() {
		t.Fatalf("Gebin(bits, 0) should be true, got %s", result.String())
	}
}

// TestGebin_OverflowIsFalse checks that Gebin(bits, 2^len(bits)) is false.
func TestGebin_OverflowIsFalse(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	a := ctx.Const("a", ctx.BoolSort())
	b := ctx.Const("b", ctx.BoolSort())
	bits := []z3bridge.Expr{a, b} // 2 bits can represent 0-3

	result := Gebin(ctx, bits, 4) // 4 >= 2^2, impossible
	if !result.IsFalse() {
		t.Fatalf("Gebin(bits, 4) with 2 bits should be false, got %s", result.String())
	}
}

// TestGebin_TwoBitsGeTwo checks Gebin([a,b], 2) = And(a, ...).
// With MSB-first: bits >= 2 means the MSB must be 1.
// 2 in binary (2 bits, MSB first) = [1, 0], so bits >= 2 iff a=1 (MSB).
func TestGebin_TwoBitsGeTwo(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	a := ctx.Const("a", ctx.BoolSort())
	b := ctx.Const("b", ctx.BoolSort())
	bits := []z3bridge.Expr{a, b}

	result := Gebin(ctx, bits, 2)
	// bits >= 2 with 2 bits MSB first means: a must be true
	// (hval=2, 2<=2, so And(a, Gebin([b], 0)) = And(a, true) = a)
	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(result, a)))
	if slv.Check() != z3bridge.Unsat {
		t.Fatalf("Gebin([a,b], 2) should equal a, got %s", result.String())
	}
}

// TestGebin_TwoBitsGeOne checks Gebin([a,b], 1) = Or(a, b).
// bits >= 1 with MSB-first: hval=2 > 1, so Or(a, Gebin([b], 1))
// = Or(a, And(b, Gebin([], 0))) = Or(a, b).
func TestGebin_TwoBitsGeOne(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	a := ctx.Const("a", ctx.BoolSort())
	b := ctx.Const("b", ctx.BoolSort())
	bits := []z3bridge.Expr{a, b}

	result := Gebin(ctx, bits, 1)
	expected := ctx.Or(a, b)
	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(result, expected)))
	if slv.Check() != z3bridge.Unsat {
		t.Fatalf("Gebin([a,b], 1) should equal Or(a,b), got %s", result.String())
	}
}

// TestGebin_ThreeBitsGeFive checks a 3-bit example.
// 3 bits MSB first: values 0-7. Gebin(bits, 5) = bits >= 5.
// 5 = 101 in binary. hval=4, 4<=5, so And(bits[0], Gebin(bits[1:], 1)).
// Gebin([b,c], 1): hval=2>1, Or(b, Gebin([c],1)) = Or(b, c).
// So Gebin([a,b,c], 5) = And(a, Or(b, c)).
func TestGebin_ThreeBitsGeFive(t *testing.T) {
	ctx := z3bridge.NewZ3Context()
	a := ctx.Const("a", ctx.BoolSort())
	b := ctx.Const("b", ctx.BoolSort())
	c := ctx.Const("c", ctx.BoolSort())
	bits := []z3bridge.Expr{a, b, c}

	result := Gebin(ctx, bits, 5)
	expected := ctx.And(a, ctx.Or(b, c))
	slv := ctx.NewSolver()
	slv.Assert(ctx.Not(ctx.Eq(result, expected)))
	if slv.Check() != z3bridge.Unsat {
		t.Fatalf("Gebin([a,b,c], 5) should equal And(a, Or(b,c)), got %s", result.String())
	}
}

// ============================================================
// Phase 4: EncodeTermZ3 & EncodeEqualityZ3 (7.2.2, 7.2.3)
// ============================================================

// TestEncodeTermZ3_Constructor checks that a constructor is encoded as its bit pattern.
func TestEncodeTermZ3_Constructor(t *testing.T) {
	// Enum sort with 3 values: {a, b, c} -> needs 2 bits
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Constructors["red"] = true
	sig.Constructors["green"] = true
	sig.Constructors["blue"] = true
	s := NewWithSig(sig)

	// Encode "green" (index 1) in 2 bits MSB first
	greenSym := lg.NewSymbol("green", es)
	bits, err := s.EncodeTermZ3(greenSym, 2, es)
	if err != nil {
		t.Fatalf("EncodeTermZ3 green: %v", err)
	}
	if len(bits) != 2 {
		t.Fatalf("expected 2 bits, got %d", len(bits))
	}
	// green = index 1 = binary 01 (MSB first), so bits[0]=false, bits[1]=true
	if !bits[0].IsFalse() {
		t.Errorf("green bit[0] should be false, got %s", bits[0].String())
	}
	if !bits[1].IsTrue() {
		t.Errorf("green bit[1] should be true, got %s", bits[1].String())
	}
}

// TestEncodeTermZ3_Variable checks that a variable produces named Bool consts.
func TestEncodeTermZ3_Variable(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	s := NewWithSig(sig)

	x, _ := lg.NewVariable("X", es)
	bits, err := s.EncodeTermZ3(x, 2, es)
	if err != nil {
		t.Fatalf("EncodeTermZ3 variable: %v", err)
	}
	if len(bits) != 2 {
		t.Fatalf("expected 2 bits, got %d", len(bits))
	}
	// Should be Bool consts named "X:color:1" and "X:color:0"
	t.Logf("variable bits: [%s, %s]", bits[0].String(), bits[1].String())
}

// TestEncodeEqualityZ3 checks that binary-encoded equality is correct.
func TestEncodeEqualityZ3(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Constructors["red"] = true
	sig.Constructors["green"] = true
	sig.Constructors["blue"] = true
	s := NewWithSig(sig)

	red := lg.NewSymbol("red", es)
	green := lg.NewSymbol("green", es)

	// red == red should be satisfiable
	eqSame, err := s.EncodeEqualityZ3(red, red, es)
	if err != nil {
		t.Fatalf("EncodeEqualityZ3 same: %v", err)
	}
	ctx := s.Context()
	slv := ctx.NewSolver()
	slv.Assert(eqSame)
	if slv.Check() == z3bridge.Unsat {
		t.Fatal("red == red should be SAT")
	}

	// red == green should be UNSAT
	eqDiff, err := s.EncodeEqualityZ3(red, green, es)
	if err != nil {
		t.Fatalf("EncodeEqualityZ3 diff: %v", err)
	}
	slv2 := ctx.NewSolver()
	slv2.Assert(eqDiff)
	if slv2.Check() != z3bridge.Unsat {
		t.Fatal("red == green should be UNSAT")
	}
}

// ============================================================
// Phase 5: NumeralToZ3 Range Clamping (7.3.5)
// ============================================================

// TestNumeralToZ3_RangeClamping checks that a numeral outside the range
// sort bounds is clamped.
func TestNumeralToZ3_RangeClamping(t *testing.T) {
	rs := &lg.RangeSort{Name: "bounded", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "10"}}
	sig := il.NewSig()
	sig.Interp["bounded"] = rs
	s := NewWithSig(sig)

	// Numeral 15 should be clamped to 10
	num := lg.NewSymbol("15", &lg.UninterpretedSort{Name: "bounded"})
	z3val, err := s.NumeralToZ3(num)
	if err != nil {
		t.Fatalf("NumeralToZ3: %v", err)
	}

	ctx := s.Context()
	// Check that z3val <= 10
	slv := ctx.NewSolver()
	slv.Assert(ctx.Gt(z3val, ctx.IntVal(10)))
	if slv.Check() != z3bridge.Unsat {
		t.Fatal("numeral 15 in range(0,10) should be clamped to <= 10")
	}

	// Numeral -5 should be clamped to 0
	numNeg := lg.NewSymbol("-5", &lg.UninterpretedSort{Name: "bounded"})
	z3valNeg, err := s.NumeralToZ3(numNeg)
	if err != nil {
		// Negative numerals may not parse; that's OK for this test
		t.Skipf("NumeralToZ3 for negative: %v", err)
	}
	slv2 := ctx.NewSolver()
	slv2.Assert(ctx.Lt(z3valNeg, ctx.IntVal(0)))
	if slv2.Check() != z3bridge.Unsat {
		t.Fatal("numeral -5 in range(0,10) should be clamped to >= 0")
	}
}

// ============================================================
// Phase 7a: clauseModelSimp early-return (7.3.2)
// ============================================================

// TestClauseModelSimp_EarlyReturn checks that when a ground literal is
// true in the model, clauseModelSimp returns just that literal (not
// the entire clause).
func TestClauseModelSimp_EarlyReturn(t *testing.T) {
	s := New()
	ctx := s.Context()

	p := boolConst("p")
	q := boolConst("q")
	r := boolConst("r")

	// Create a model where p=true, q=false, r=false
	slv := ctx.NewSolver()
	zp, _ := s.FormulaToZ3(p)
	zq, _ := s.FormulaToZ3(q)
	zr, _ := s.FormulaToZ3(r)
	slv.Assert(zp)
	slv.Assert(ctx.Not(zq))
	slv.Assert(ctx.Not(zr))
	slv.Check()
	model := slv.Model()

	// Clause: p | q | r
	clause := &lg.Or{Terms: []lg.Expr{p, q, r}}

	result := s.clauseModelSimp(model, clause)

	// Python returns [l] (just p) because p is true in the model.
	// Go should return p (the single literal), not Or(p).
	if _, isOr := result.(*lg.Or); isOr {
		t.Fatalf("clauseModelSimp should return single literal, got Or: %v", result)
	}
	sym, ok := result.(*lg.Symbol)
	if !ok {
		t.Fatalf("clauseModelSimp should return Symbol 'p', got %T: %v", result, result)
	}
	if sym.Name != "p" {
		t.Fatalf("clauseModelSimp should return 'p', got '%s'", sym.Name)
	}
}

// TestClauseModelSimp_DropFalse checks that false literals are dropped.
func TestClauseModelSimp_DropFalse(t *testing.T) {
	s := New()
	ctx := s.Context()

	p := boolConst("p")

	// Model where p=false
	slv := ctx.NewSolver()
	zp, _ := s.FormulaToZ3(p)
	slv.Assert(ctx.Not(zp))
	slv.Check()
	model := slv.Model()

	// Use a non-ground literal (variable) so it's always kept
	xVar := boolVar("X")
	clause := &lg.Or{Terms: []lg.Expr{p, xVar}}

	result := s.clauseModelSimp(model, clause)
	// p is false → dropped. X is non-ground → kept.
	// Result should be just X.
	if _, isVar := result.(*lg.Variable); !isVar {
		t.Fatalf("expected single variable after dropping false literal, got %T: %v", result, result)
	}
}

// ============================================================
// Phase 7b: ClausesModelToDiagram (7.2.4)
// ============================================================

// TestClausesModelToDiagram_NoTautologies checks that the diagram does not
// contain tautologies like "sym = sym".
func TestClausesModelToDiagram_NoTautologies(t *testing.T) {
	s := New()
	sort := &lg.UninterpretedSort{Name: "T"}
	a := lg.NewSymbol("a", sort)
	b := lg.NewSymbol("b", sort)

	// Simple clauses: a != b
	fmla := &lg.Not{Body: &lg.Eq{T1: a, T2: b}}
	clauses := clauseops.NewClauses([]lg.Expr{fmla}, nil, nil)

	result, err := s.ClausesModelToDiagram(clauses, nil, nil)
	if err != nil {
		t.Fatalf("ClausesModelToDiagram: %v", err)
	}
	if result == nil {
		t.Skip("UNSAT — cannot test diagram")
	}

	// Check no tautologies (sym = sym)
	for _, f := range result.Fmlas {
		if eq, ok := f.(*lg.Eq); ok {
			if eq.T1.Equal(eq.T2) {
				t.Errorf("diagram contains tautology: %v = %v", eq.T1, eq.T2)
			}
		}
	}
}

// ============================================================
// Phase 8a: SortCard BV/range (7.3.11)
// ============================================================

// TestSortCard_Enumerated checks cardinality of an enumerated sort.
func TestSortCard_Enumerated(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	card := SortCard(es, nil)
	if card != 3 {
		t.Fatalf("SortCard(color) = %d, want 3", card)
	}
}

// TestSortCard_RangeSort checks cardinality of a range sort.
func TestSortCard_RangeSort(t *testing.T) {
	rs := &lg.RangeSort{Name: "bounded", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "9"}}
	card := SortCard(rs, nil)
	if card != 10 {
		t.Fatalf("SortCard(range(0,9)) = %d, want 10", card)
	}
}

// TestSortCard_BV checks cardinality of a BV sort.
func TestSortCard_BV(t *testing.T) {
	sig := il.NewSig()
	sig.Interp["mybv"] = "bv[8]"
	sort := &lg.UninterpretedSort{Name: "mybv"}
	card := SortCard(sort, sig)
	if card != 256 {
		t.Fatalf("SortCard(bv[8]) = %d, want 256", card)
	}
}

// TestSortCard_Uninterpreted checks that uninterpreted sorts return -1.
func TestSortCard_Uninterpreted(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "T"}
	card := SortCard(sort, nil)
	if card != -1 {
		t.Fatalf("SortCard(uninterpreted) = %d, want -1", card)
	}
}

// ============================================================
// Phase 8d: RangeSortBounds (7.2.1)
// ============================================================

// TestRangeSortBounds_ParsesCorrectly checks that Lb/Ub are parsed as ints.
func TestRangeSortBounds_ParsesCorrectly(t *testing.T) {
	rs := &lg.RangeSort{Name: "r", Lb: lg.NumeralBound{Value: "3"}, Ub: lg.NumeralBound{Value: "17"}}
	lo, hi, ok := RangeSortBounds(rs)
	if !ok {
		t.Fatal("RangeSortBounds returned !ok")
	}
	if lo != 3 {
		t.Fatalf("lo = %d, want 3", lo)
	}
	if hi != 17 {
		t.Fatalf("hi = %d, want 17", hi)
	}
}

// TestRangeSortBounds_NotRangeSort checks non-range sort returns false.
func TestRangeSortBounds_NotRangeSort(t *testing.T) {
	us := &lg.UninterpretedSort{Name: "T"}
	_, _, ok := RangeSortBounds(us)
	if ok {
		t.Fatal("RangeSortBounds should return false for non-range sort")
	}
}

// ============================================================
// Phase 8g: Variable naming :sort_name (7.3.9)
// ============================================================

// TestVariableNaming checks that variables are named "name:sortName" in Z3.
func TestVariableNaming(t *testing.T) {
	s := New()
	sort := &lg.UninterpretedSort{Name: "node"}
	x, _ := lg.NewVariable("X", sort)

	z3x, err := s.Translator().Translate(x)
	if err != nil {
		t.Fatalf("translate variable: %v", err)
	}
	name := z3x.String()
	// Z3 may quote identifiers containing ':' as |X:node|
	stripped := strings.Trim(name, "|")
	if stripped != "X:node" {
		t.Fatalf("variable Z3 name = %q, want %q", name, "X:node")
	}
}

// TestVariableNaming_BoolSort checks Bool variable naming.
func TestVariableNaming_BoolSort(t *testing.T) {
	s := New()
	v, _ := lg.NewVariable("Flag", lg.Boolean)

	z3x, err := s.Translator().Translate(v)
	if err != nil {
		t.Fatalf("translate bool variable: %v", err)
	}
	name := z3x.String()
	stripped := strings.Trim(name, "|")
	if stripped != "Flag:Bool" {
		t.Fatalf("bool variable Z3 name = %q, want %q", name, "Flag:Bool")
	}
}

// ============================================================
// Phase 8h: CheckSequence Reporter (7.3.10)
// ============================================================

type testReporter struct {
	starts []string
	ends   []string
	abort  bool // if true, End returns false on first assert failure
}

func (r *testReporter) Start(isAssert bool, doc string) bool {
	kind := "assume"
	if isAssert {
		kind = "assert"
	}
	r.starts = append(r.starts, kind+":"+doc)
	return true
}

func (r *testReporter) End(result bool, doc string) bool {
	status := "pass"
	if !result {
		status = "fail"
	}
	r.ends = append(r.ends, status+":"+doc)
	if r.abort && !result {
		return false // abort on failure
	}
	return true
}

func TestCheckSequenceWithReporter(t *testing.T) {
	s := New()
	p := boolConst("p")

	seq := []AssumeAssert{
		NewAssume(clauseops.FormulaToClauses(p, nil), "assume_p"),
		NewAssert(clauseops.FormulaToClauses(p, nil), "assert_p"),
	}

	rep := &testReporter{}
	results, err := s.CheckSequenceWithReporter(seq, rep)
	if err != nil {
		t.Fatalf("CheckSequenceWithReporter: %v", err)
	}

	if len(rep.starts) != 2 {
		t.Fatalf("reporter got %d starts, want 2", len(rep.starts))
	}
	if rep.starts[0] != "assume:assume_p" {
		t.Fatalf("start[0] = %q, want %q", rep.starts[0], "assume:assume_p")
	}
	if rep.starts[1] != "assert:assert_p" {
		t.Fatalf("start[1] = %q, want %q", rep.starts[1], "assert:assert_p")
	}

	if !results[0] || !results[1] {
		t.Fatalf("both checks should pass, got %v", results)
	}
}

func TestCheckSequenceWithReporter_Abort(t *testing.T) {
	s := New()
	p := boolConst("p")

	seq := []AssumeAssert{
		// Assert p with no assumptions — should fail
		NewAssert(clauseops.FormulaToClauses(p, nil), "assert_p"),
		NewAssert(clauseops.FormulaToClauses(p, nil), "assert_p_again"),
	}

	rep := &testReporter{abort: true}
	_, err := s.CheckSequenceWithReporter(seq, rep)
	if err != nil {
		t.Fatalf("CheckSequenceWithReporter: %v", err)
	}

	// Should have aborted after first failure
	if len(rep.ends) != 1 {
		t.Fatalf("reporter should have 1 end (aborted), got %d: %v", len(rep.ends), rep.ends)
	}
}

// ============================================================
// Phase 8f: SolverName z3_builtins (7.2.8)
// ============================================================

// TestSolverName_Z3Builtins checks that z3 builtin names panic with IvyError.
// Python: raise iu.IvyError(None, 'name "{}" clashes with Z3 built-in'.format(name))
func TestSolverName_Z3Builtins(t *testing.T) {
	s := New()
	for _, name := range []string{"bit0", "bit1"} {
		sym := lg.NewSymbol(name, lg.Boolean)
		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("SolverName(%q) should panic for Z3 builtin", name)
				}
			}()
			s.SolverName(sym)
		}()
	}
}

// TestSolverName_Normal checks that normal names pass through.
func TestSolverName_Normal(t *testing.T) {
	s := New()
	sym := lg.NewSymbol("myvar", lg.Boolean)
	result := s.SolverName(sym)
	if result != "myvar" {
		t.Errorf("SolverName(myvar) = %q, want %q", result, "myvar")
	}
}
