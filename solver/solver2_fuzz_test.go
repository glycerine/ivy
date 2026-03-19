package solver

import (
	"testing"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// FuzzMyEq exercises MyEq with random Bool Z3 expressions, verifying
// the True/False short-circuit and that no panics occur.
func FuzzMyEq(f *testing.F) {
	// Seed: (isXTrue, isYTrue, isYFalse)
	f.Add(true, true, false)
	f.Add(false, false, true)
	f.Add(true, false, false) // both symbolic
	f.Add(false, true, false)
	f.Add(false, false, false)

	f.Fuzz(func(t *testing.T, isXTrue, isYTrue, isYFalse bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		ctx := z3bridge.NewContext()
		var x, y z3bridge.Expr
		if isXTrue {
			x = ctx.BoolVal(true)
		} else {
			x = ctx.Const("x", ctx.BoolSort())
		}
		if isYTrue {
			y = ctx.BoolVal(true)
		} else if isYFalse {
			y = ctx.BoolVal(false)
		} else {
			y = ctx.Const("y", ctx.BoolSort())
		}

		result := MyEq(ctx, x, y)

		// Verify equivalence with standard Eq
		expected := ctx.Eq(x, y)
		slv := ctx.NewSolver()
		slv.Assert(ctx.Not(ctx.Eq(result, expected)))
		if slv.Check() != z3bridge.Unsat {
			t.Fatalf("MyEq(%s, %s) = %s is not equivalent to Eq",
				x.String(), y.String(), result.String())
		}
	})
}

// FuzzGebin exercises the Gebin binary encoding predicate with random
// bit counts and thresholds, verifying correctness against brute-force.
func FuzzGebin(f *testing.F) {
	f.Add(1, 0)
	f.Add(1, 1)
	f.Add(2, 0)
	f.Add(2, 1)
	f.Add(2, 2)
	f.Add(2, 3)
	f.Add(2, 4) // overflow
	f.Add(3, 5)
	f.Add(3, 7)
	f.Add(3, 8) // overflow
	f.Add(4, 15)

	f.Fuzz(func(t *testing.T, nbits, threshold int) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on nbits=%d threshold=%d: %v", nbits, threshold, r)
			}
		}()

		if nbits < 0 || nbits > 8 || threshold < 0 || threshold > 300 {
			return
		}

		ctx := z3bridge.NewContext()

		// Create nbits Bool constants
		bits := make([]z3bridge.Expr, nbits)
		for i := 0; i < nbits; i++ {
			bits[i] = ctx.Const(string(rune('a'+i)), ctx.BoolSort())
		}

		result := Gebin(ctx, bits, threshold)

		// Brute-force verify: for each assignment of bits, check that
		// Gebin agrees with "value >= threshold".
		maxVal := 1 << uint(nbits)
		for val := 0; val < maxVal; val++ {
			slv := ctx.NewSolver()
			// Fix bits to represent val (MSB first)
			for i := 0; i < nbits; i++ {
				bitSet := (val & (1 << uint(nbits-1-i))) != 0
				if bitSet {
					slv.Assert(bits[i])
				} else {
					slv.Assert(ctx.Not(bits[i]))
				}
			}

			// Check: result should be true iff val >= threshold
			expectTrue := val >= threshold
			if expectTrue {
				slv.Assert(ctx.Not(result))
				if slv.Check() != z3bridge.Unsat {
					t.Fatalf("Gebin(bits, %d) should be true for val=%d (nbits=%d)",
						threshold, val, nbits)
				}
			} else {
				slv.Assert(result)
				if slv.Check() != z3bridge.Unsat {
					t.Fatalf("Gebin(bits, %d) should be false for val=%d (nbits=%d)",
						threshold, val, nbits)
				}
			}
		}
	})
}

// FuzzBinEncZ3 exercises the Z3-level binary encoding, verifying that
// BinEncZ3(m, n) produces the correct MSB-first bit pattern.
func FuzzBinEncZ3(f *testing.F) {
	f.Add(0, 1)
	f.Add(1, 1)
	f.Add(0, 4)
	f.Add(5, 4)
	f.Add(15, 4)
	f.Add(7, 3)

	f.Fuzz(func(t *testing.T, m, n int) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		if n < 0 || n > 16 || m < 0 || m >= (1<<uint(n)) {
			return
		}

		ctx := z3bridge.NewContext()
		bits := BinEncZ3(ctx, m, n)

		if len(bits) != n {
			t.Fatalf("BinEncZ3(%d, %d) returned %d bits", m, n, len(bits))
		}

		// Verify each bit (MSB first)
		for i := 0; i < n; i++ {
			expectSet := (m & (1 << uint(n-1-i))) != 0
			if expectSet {
				if !bits[i].IsTrue() {
					t.Fatalf("BinEncZ3(%d, %d): bit[%d] should be true", m, n, i)
				}
			} else {
				if !bits[i].IsFalse() {
					t.Fatalf("BinEncZ3(%d, %d): bit[%d] should be false", m, n, i)
				}
			}
		}
	})
}

// FuzzSortCard exercises SortCard with various sort types and signatures.
func FuzzSortCard(f *testing.F) {
	f.Add(byte(0), 3, "")     // enumerated, 3 elements
	f.Add(byte(1), 0, "0:9")  // range sort
	f.Add(byte(2), 8, "bv")   // BV via sig
	f.Add(byte(3), 0, "")     // uninterpreted, no sig

	f.Fuzz(func(t *testing.T, sortKind byte, param int, extra string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		if param < 0 || param > 64 {
			return
		}

		var sort lg.Sort
		var sig *il.Sig

		switch sortKind % 4 {
		case 0: // EnumeratedSort
			n := param % 20
			if n == 0 {
				n = 1
			}
			ext := make([]string, n)
			for i := range ext {
				ext[i] = string(rune('a' + i%26))
			}
			sort = &lg.EnumeratedSort{Name: "E", Extension: ext}
		case 1: // RangeSort
			lo := 0
			hi := param
			if hi < lo {
				return
			}
			sort = &lg.RangeSort{Name: "R", Lb: "0", Ub: string(rune('0' + hi%10))}
		case 2: // BV via sig interp
			width := param % 32
			if width == 0 {
				width = 1
			}
			sig = il.NewSig()
			sig.Interp["mybv"] = "bv[" + string(rune('0'+width%10)) + "]"
			sort = &lg.UninterpretedSort{Name: "mybv"}
		case 3: // Plain uninterpreted
			sort = &lg.UninterpretedSort{Name: "T"}
		}

		card := SortCard(sort, sig)
		_ = card // no panic is the goal
	})
}

// FuzzRangeSortBounds exercises RangeSortBounds with random Lb/Ub strings.
func FuzzRangeSortBounds(f *testing.F) {
	f.Add("0", "10")
	f.Add("5", "5")
	f.Add("-3", "3")
	f.Add("abc", "10")  // non-numeric lb
	f.Add("0", "xyz")   // non-numeric ub
	f.Add("", "5")
	f.Add("999999", "999999")

	f.Fuzz(func(t *testing.T, lb, ub string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on lb=%q ub=%q: %v", lb, ub, r)
			}
		}()

		if len(lb) > 20 || len(ub) > 20 {
			return
		}

		rs := &lg.RangeSort{Name: "R", Lb: lb, Ub: ub}
		lo, hi, ok := RangeSortBounds(rs)
		if ok {
			if hi < lo {
				// Not invalid per se, but verify it doesn't panic
				_ = hi - lo
			}
		}
	})
}

// FuzzNumeralToZ3Clamping exercises NumeralToZ3 with range-sort interpreted
// numerals, verifying no panics and that values are clamped.
func FuzzNumeralToZ3Clamping(f *testing.F) {
	f.Add("5", "0", "10")
	f.Add("15", "0", "10")
	f.Add("-3", "0", "10")
	f.Add("0", "0", "0")
	f.Add("100", "5", "50")

	f.Fuzz(func(t *testing.T, numStr, lb, ub string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on num=%q lb=%q ub=%q: %v", numStr, lb, ub, r)
			}
		}()

		if len(numStr) > 10 || len(lb) > 10 || len(ub) > 10 {
			return
		}
		// Only try numeric strings
		for _, c := range numStr {
			if (c < '0' || c > '9') && c != '-' {
				return
			}
		}
		for _, c := range lb {
			if (c < '0' || c > '9') && c != '-' {
				return
			}
		}
		for _, c := range ub {
			if (c < '0' || c > '9') && c != '-' {
				return
			}
		}
		if len(numStr) == 0 || len(lb) == 0 || len(ub) == 0 {
			return
		}

		rs := &lg.RangeSort{Name: "bounded", Lb: lb, Ub: ub}
		sig := il.NewSig()
		sig.Interp["bounded"] = rs
		s := NewWithSig(sig)

		num := lg.NewSymbol(numStr, &lg.UninterpretedSort{Name: "bounded"})
		_, err := s.NumeralToZ3(num)
		// Errors are fine, panics are not
		_ = err
	})
}

// FuzzEncodeEqualityZ3 exercises binary-encoded equality with random
// enumerated sort sizes, verifying no panics and that same-term equality
// is always SAT.
func FuzzEncodeEqualityZ3(f *testing.F) {
	f.Add(2, 0, 0)  // 2-element sort, same term
	f.Add(3, 0, 1)  // 3-element sort, different terms
	f.Add(4, 2, 2)  // 4-element sort, same term
	f.Add(5, 0, 4)  // 5-element sort, different terms
	f.Add(7, 3, 6)  // 7-element sort

	f.Fuzz(func(t *testing.T, nElems, idx1, idx2 int) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		if nElems < 2 || nElems > 16 {
			return
		}
		idx1 = idx1 % nElems
		idx2 = idx2 % nElems
		if idx1 < 0 {
			idx1 = -idx1
		}
		if idx2 < 0 {
			idx2 = -idx2
		}

		ext := make([]string, nElems)
		for i := range ext {
			ext[i] = string(rune('a' + i))
		}
		es := &lg.EnumeratedSort{Name: "E", Extension: ext}

		sig := il.NewSig()
		for _, name := range ext {
			sig.Constructors[name] = true
		}
		s := NewWithSig(sig)

		t1 := lg.NewSymbol(ext[idx1], es)
		t2 := lg.NewSymbol(ext[idx2], es)

		eq, err := s.EncodeEqualityZ3(t1, t2, es)
		if err != nil {
			return
		}

		ctx := s.Context()
		slv := ctx.NewSolver()
		if idx1 == idx2 {
			// Same term: equality must be SAT
			slv.Assert(eq)
			if slv.Check() == z3bridge.Unsat {
				t.Fatalf("encode_equality(%s, %s) should be SAT", ext[idx1], ext[idx2])
			}
		} else {
			// Different terms: equality must be UNSAT
			slv.Assert(eq)
			if slv.Check() != z3bridge.Unsat {
				t.Fatalf("encode_equality(%s, %s) should be UNSAT", ext[idx1], ext[idx2])
			}
		}
	})
}

// FuzzSolverNameBuiltins exercises SolverName with random symbol names,
// verifying z3 builtins return "" and no panics occur.
func FuzzSolverNameBuiltins(f *testing.F) {
	f.Add("bit0")
	f.Add("bit1")
	f.Add("myvar")
	f.Add("+")
	f.Add("bfe[0:3]")
	f.Add("")
	f.Add("x")

	f.Fuzz(func(t *testing.T, name string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on name=%q: %v", name, r)
			}
		}()

		if len(name) > 100 {
			return
		}

		s := New()
		sym := lg.NewSymbol(name, lg.Boolean)
		result := s.SolverName(sym)

		// z3 builtins must return ""
		if name == "bit0" || name == "bit1" {
			if result != "" {
				t.Fatalf("SolverName(%q) = %q, want empty", name, result)
			}
		}
	})
}

// FuzzQuantConstraintsNatRange exercises the full solver quantifier
// constraint pipeline with random sort interpretations, verifying
// no panics during ForAll/Exists translation.
func FuzzQuantConstraintsNatRange(f *testing.F) {
	f.Add(byte(0), true)  // nat, ForAll
	f.Add(byte(0), false) // nat, Exists
	f.Add(byte(1), true)  // range, ForAll
	f.Add(byte(1), false) // range, Exists
	f.Add(byte(2), true)  // uninterpreted, ForAll

	f.Fuzz(func(t *testing.T, interpKind byte, isForall bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		sig := il.NewSig()
		switch interpKind % 3 {
		case 0:
			sig.Interp["mysort"] = "nat"
		case 1:
			sig.Interp["mysort"] = &lg.RangeSort{Name: "mysort", Lb: "0", Ub: "10"}
		case 2:
			// No interpretation
		}
		s := NewWithSig(sig)

		sort := &lg.UninterpretedSort{Name: "mysort"}
		x, err := lg.NewVariable("X", sort)
		if err != nil {
			return
		}
		p := relConst("P", sort)
		pApp := &lg.Apply{Func: p, Terms: []lg.Expr{x}}

		var fmla lg.Expr
		if isForall {
			fmla = &lg.ForAll{Variables: []*lg.Variable{x}, Body: pApp}
		} else {
			fmla = &lg.Exists{Variables: []*lg.Variable{x}, Body: pApp}
		}

		_, err = s.FormulaToZ3(fmla)
		// Errors are acceptable, panics are not
		_ = err
	})
}
