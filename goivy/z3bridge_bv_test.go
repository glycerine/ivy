package goivy

import (
	"github.com/glycerine/ivy/goivy/smt"
	"testing"
)

// helper: assert SAT
func assertSat(t *testing.T, slv *smt.Z3Solver, msg string) {
	t.Helper()
	r := slv.Check()
	if r != smt.Sat {
		t.Errorf("%s: expected SAT, got %s", msg, r)
	}
}

// helper: assert UNSAT
func assertUnsat(t *testing.T, slv *smt.Z3Solver, msg string) {
	t.Helper()
	r := slv.Check()
	if r != smt.Unsat {
		t.Errorf("%s: expected UNSAT, got %s", msg, r)
	}
}

// --- Bitvector sort and value creation ---

func TestBvSort(t *testing.T) {
	ctx := smt.NewZ3Context()
	bv8 := ctx.BvSort(8)
	if !ctx.IsBvSort(bv8) {
		t.Fatal("BvSort(8) should be a bv sort")
	}
	if ctx.BvSortSize(bv8) != 8 {
		t.Errorf("BvSort(8) size should be 8, got %d", ctx.BvSortSize(bv8))
	}
	bv32 := ctx.BvSort(32)
	if ctx.BvSortSize(bv32) != 32 {
		t.Errorf("BvSort(32) size should be 32, got %d", ctx.BvSortSize(bv32))
	}
	// Int sort should NOT be BV
	if ctx.IsBvSort(ctx.IntSort()) {
		t.Fatal("IntSort should not be a bv sort")
	}
	// Bool sort should NOT be BV
	if ctx.IsBvSort(ctx.BoolSort()) {
		t.Fatal("BoolSort should not be a bv sort")
	}
}

func TestBvVal(t *testing.T) {
	ctx := smt.NewZ3Context()
	v := ctx.BvVal(42, 8)
	if !ctx.IsBvExpr(v) {
		t.Fatal("BvVal should produce a BV expression")
	}
	sort := v.ExprSort()
	if ctx.BvSortSize(sort) != 8 {
		t.Errorf("BvVal(42, 8) sort size should be 8, got %d", ctx.BvSortSize(sort))
	}
}

func TestIsBvExpr(t *testing.T) {
	ctx := smt.NewZ3Context()
	if !ctx.IsBvExpr(ctx.BvVal(5, 8)) {
		t.Error("BvVal should be BV expr")
	}
	if ctx.IsBvExpr(ctx.IntVal(5)) {
		t.Error("IntVal should not be BV expr")
	}
	if ctx.IsBvExpr(ctx.BoolVal(true)) {
		t.Error("BoolVal should not be BV expr")
	}
}

// --- Bitvector arithmetic ---

func TestBvAdd(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	a := ctx.Const("a", ctx.BvSort(8))
	b := ctx.Const("b", ctx.BvSort(8))
	// 200 + 100 = 300 mod 256 = 44
	slv.Assert(ctx.Eq(a, ctx.BvVal(200, 8)))
	slv.Assert(ctx.Eq(b, ctx.BvVal(100, 8)))
	slv.Assert(ctx.Eq(ctx.BvAdd(a, b), ctx.BvVal(44, 8)))
	assertSat(t, slv, "BvAdd 200+100=44 (overflow)")
}

func TestBvSub(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	a := ctx.Const("a", ctx.BvSort(8))
	b := ctx.Const("b", ctx.BvSort(8))
	// 10 - 20 = -10 mod 256 = 246
	slv.Assert(ctx.Eq(a, ctx.BvVal(10, 8)))
	slv.Assert(ctx.Eq(b, ctx.BvVal(20, 8)))
	slv.Assert(ctx.Eq(ctx.BvSub(a, b), ctx.BvVal(246, 8)))
	assertSat(t, slv, "BvSub 10-20=246 (underflow)")
}

func TestBvMul(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvMul(ctx.BvVal(7, 8), ctx.BvVal(6, 8)), ctx.BvVal(42, 8)))
	assertSat(t, slv, "BvMul 7*6=42")
}

func TestBvUdiv(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvUdiv(ctx.BvVal(200, 8), ctx.BvVal(10, 8)), ctx.BvVal(20, 8)))
	assertSat(t, slv, "BvUdiv 200/10=20")
}

// --- Bitvector bitwise operations ---

func TestBvAnd(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvAnd(ctx.BvVal(0xF0, 8), ctx.BvVal(0x3C, 8)), ctx.BvVal(0x30, 8)))
	assertSat(t, slv, "BvAnd 0xF0&0x3C=0x30")
}

func TestBvOr(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvOr(ctx.BvVal(0xF0, 8), ctx.BvVal(0x0F, 8)), ctx.BvVal(0xFF, 8)))
	assertSat(t, slv, "BvOr 0xF0|0x0F=0xFF")
}

func TestBvXor(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvXor(ctx.BvVal(0xFF, 8), ctx.BvVal(0x0F, 8)), ctx.BvVal(0xF0, 8)))
	assertSat(t, slv, "BvXor 0xFF^0x0F=0xF0")
}

func TestBvNot(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvNot(ctx.BvVal(0xF0, 8)), ctx.BvVal(0x0F, 8)))
	assertSat(t, slv, "BvNot ~0xF0=0x0F")
}

// --- Bitvector shifts ---

func TestBvShl(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvShl(ctx.BvVal(1, 8), ctx.BvVal(4, 8)), ctx.BvVal(16, 8)))
	assertSat(t, slv, "BvShl 1<<4=16")
}

func TestBvLshr(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	// 128 >> 3 = 16
	slv.Assert(ctx.Eq(ctx.BvLshr(ctx.BvVal(0x80, 8), ctx.BvVal(3, 8)), ctx.BvVal(0x10, 8)))
	assertSat(t, slv, "BvLshr 0x80>>3=0x10")
}

func TestBvAshr(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	// 0x80 = -128 signed, arithmetic right shift by 1 = 0xC0 = -64 signed
	slv.Assert(ctx.Eq(ctx.BvAshr(ctx.BvVal(0x80, 8), ctx.BvVal(1, 8)), ctx.BvVal(0xC0, 8)))
	assertSat(t, slv, "BvAshr 0x80>>a1=0xC0")
}

// --- Bitvector unsigned comparisons ---

func TestBvUlt(t *testing.T) {
	ctx := smt.NewZ3Context()

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.BvUlt(ctx.BvVal(5, 8), ctx.BvVal(10, 8)))
	assertSat(t, slv, "5 < 10 unsigned")

	slv2 := ctx.NewZ3Solver()
	slv2.Assert(ctx.BvUlt(ctx.BvVal(200, 8), ctx.BvVal(10, 8)))
	assertUnsat(t, slv2, "200 < 10 unsigned")
}

func TestBvUle(t *testing.T) {
	ctx := smt.NewZ3Context()

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.BvUle(ctx.BvVal(42, 8), ctx.BvVal(42, 8)))
	assertSat(t, slv, "42 <= 42 unsigned")

	slv2 := ctx.NewZ3Solver()
	slv2.Assert(ctx.BvUle(ctx.BvVal(5, 8), ctx.BvVal(10, 8)))
	assertSat(t, slv2, "5 <= 10 unsigned")

	slv3 := ctx.NewZ3Solver()
	slv3.Assert(ctx.BvUle(ctx.BvVal(200, 8), ctx.BvVal(10, 8)))
	assertUnsat(t, slv3, "200 <= 10 unsigned")
}

func TestBvUgt(t *testing.T) {
	ctx := smt.NewZ3Context()

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.BvUgt(ctx.BvVal(200, 8), ctx.BvVal(10, 8)))
	assertSat(t, slv, "200 > 10 unsigned")

	slv2 := ctx.NewZ3Solver()
	slv2.Assert(ctx.BvUgt(ctx.BvVal(5, 8), ctx.BvVal(10, 8)))
	assertUnsat(t, slv2, "5 > 10 unsigned")
}

func TestBvUge(t *testing.T) {
	ctx := smt.NewZ3Context()

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.BvUge(ctx.BvVal(42, 8), ctx.BvVal(42, 8)))
	assertSat(t, slv, "42 >= 42 unsigned")

	slv2 := ctx.NewZ3Solver()
	slv2.Assert(ctx.BvUge(ctx.BvVal(200, 8), ctx.BvVal(10, 8)))
	assertSat(t, slv2, "200 >= 10 unsigned")

	slv3 := ctx.NewZ3Solver()
	slv3.Assert(ctx.BvUge(ctx.BvVal(5, 8), ctx.BvVal(10, 8)))
	assertUnsat(t, slv3, "5 >= 10 unsigned")
}

// --- Critical: unsigned vs signed interpretation ---
// 0x80 = 128 unsigned, -128 signed.
// Unsigned: 0x80 > 0x01 (128 > 1) — TRUE
// Signed:   0x80 < 0x01 (-128 < 1) — TRUE
// Python uses unsigned (ULT/ULE/UGT/UGE) for bitvectors.

func TestBvUnsignedSemantics(t *testing.T) {
	ctx := smt.NewZ3Context()
	highBit := ctx.BvVal(0x80, 8)
	one := ctx.BvVal(1, 8)

	// Unsigned: 0x80 > 1 → true
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.BvUgt(highBit, one))
	assertSat(t, slv, "0x80 > 1 unsigned (critical BV semantics)")

	// Unsigned: 1 < 0x80 → true
	slv2 := ctx.NewZ3Solver()
	slv2.Assert(ctx.BvUlt(one, highBit))
	assertSat(t, slv2, "1 < 0x80 unsigned")

	// Unsigned: 0x80 < 1 → false (would be true under signed)
	slv3 := ctx.NewZ3Solver()
	slv3.Assert(ctx.BvUlt(highBit, one))
	assertUnsat(t, slv3, "0x80 < 1 unsigned (signed would give opposite)")
}

// --- Concat ---

func TestBvConcat(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	result := ctx.Concat(ctx.BvVal(0xAB, 8), ctx.BvVal(0xCD, 8))
	sort := result.ExprSort()
	if ctx.BvSortSize(sort) != 16 {
		t.Errorf("Concat of two 8-bit should be 16-bit, got %d", ctx.BvSortSize(sort))
	}
	slv.Assert(ctx.Eq(result, ctx.BvVal(0xABCD, 16)))
	assertSat(t, slv, "Concat 0xAB++0xCD=0xABCD")
}

// --- Bv2Int ---

func TestBv2IntUnsigned(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.Bv2Int(ctx.BvVal(42, 8), false), ctx.IntVal(42)))
	assertSat(t, slv, "Bv2Int(42, unsigned)=42")
}

func TestBv2IntSigned(t *testing.T) {
	ctx := smt.NewZ3Context()
	slv := ctx.NewZ3Solver()
	// 0xFF as signed 8-bit = -1
	slv.Assert(ctx.Eq(ctx.Bv2Int(ctx.BvVal(0xFF, 8), true), ctx.IntVal(-1)))
	assertSat(t, slv, "Bv2Int(0xFF, signed)=-1")
}

// --- Symbolic BV test ---

func TestBvSymbolicSolve(t *testing.T) {
	ctx := smt.NewZ3Context()
	bv8 := ctx.BvSort(8)
	x := ctx.Const("x", bv8)
	y := ctx.Const("y", bv8)

	if !ctx.IsBvExpr(x) {
		t.Fatal("BV constant should be detected as BV expr")
	}

	// Find x, y such that x & y == 0x30 and x | y == 0xFC
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.BvAnd(x, y), ctx.BvVal(0x30, 8)))
	slv.Assert(ctx.Eq(ctx.BvOr(x, y), ctx.BvVal(0xFC, 8)))
	assertSat(t, slv, "BV symbolic AND/OR constraint")
}

func TestBvSymbolicCompare(t *testing.T) {
	ctx := smt.NewZ3Context()
	bv8 := ctx.BvSort(8)
	x := ctx.Const("x", bv8)

	// Find x such that x > 200 and x < 250 (unsigned)
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.BvUgt(x, ctx.BvVal(200, 8)))
	slv.Assert(ctx.BvUlt(x, ctx.BvVal(250, 8)))
	assertSat(t, slv, "200 < x < 250 unsigned")

	// No x such that x > 255 (8-bit max) — UNSAT
	slv2 := ctx.NewZ3Solver()
	slv2.Assert(ctx.BvUgt(x, ctx.BvVal(255, 8)))
	assertUnsat(t, slv2, "x > 255 in 8-bit")
}
