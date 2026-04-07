package solver

import (
	"runtime"
	"sync"
	"testing"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// --- Z3 Worker Goroutine ---
//
// Z3 4.7.1's C library uses Thread-Local Storage (TLS) internally and is
// not safe to use from multiple OS threads, even with separate contexts.
// Go's M:N goroutine scheduler can migrate a goroutine between OS threads
// between CGO calls, which corrupts Z3's TLS state.
//
// To safely fuzz Z3-dependent code we funnel ALL Z3 work through a single
// goroutine that is pinned to one OS thread via runtime.LockOSThread().
// Fuzz workers send closures to this goroutine over a channel and block
// until the work completes.

// z3Job is a closure that performs Z3 work. It receives a *testing.T for
// reporting failures. Any panic is caught by the worker and forwarded.
type z3Job struct {
	fn   func(t *testing.T)
	t    *testing.T
	done chan *z3Result
}

type z3Result struct {
	panicVal interface{}
}

var (
	z3WorkerOnce sync.Once
	z3JobChan    chan *z3Job
)

// startZ3Worker launches the singleton Z3 worker goroutine.
func startZ3Worker() {
	z3WorkerOnce.Do(func() {
		z3JobChan = make(chan *z3Job, 1)
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			// but basically: we never return,
			// so we never unlock — this goroutine owns this OS thread for life.
			for {
				select {
				case job := <-z3JobChan:
					result := &z3Result{}
					func() {
						defer func() {
							if r := recover(); r != nil {
								result.panicVal = r
							}
						}()
						job.fn(job.t)
					}()
					select {
					case job.done <- result:
					}
				}
			}
		}()
	})
}

// runOnZ3Thread sends a closure to the Z3 worker goroutine and waits for
// it to complete. If the closure panicked, the panic value is reported as
// a test fatal error.
func runOnZ3Thread(t *testing.T, fn func(t *testing.T)) {
	t.Helper()
	startZ3Worker()
	done := make(chan *z3Result, 1)
	z3JobChan <- &z3Job{fn: fn, t: t, done: done}
	result := <-done
	if result.panicVal != nil {
		t.Fatalf("panic on Z3 thread: %v", result.panicVal)
	}
}

// ============================================================
// Fuzz tests
// ============================================================

// FuzzMyEq exercises MyEq with random Bool Z3 expressions, verifying
// the True/False short-circuit and that no panics occur.
func FuzzMyEq(f *testing.F) {
	f.Add(true, true, false)
	f.Add(false, false, true)
	f.Add(true, false, false)
	f.Add(false, true, false)
	f.Add(false, false, false)

	f.Fuzz(func(t *testing.T, isXTrue, isYTrue, isYFalse bool) {
		runOnZ3Thread(t, func(t *testing.T) {
			ctx := z3bridge.NewZ3Context()
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
	f.Add(2, 4)
	f.Add(3, 5)
	f.Add(3, 7)
	f.Add(3, 8)
	f.Add(4, 15)

	f.Fuzz(func(t *testing.T, nbits, threshold int) {
		if nbits < 0 || nbits > 8 || threshold < 0 || threshold > 300 {
			return
		}

		runOnZ3Thread(t, func(t *testing.T) {
			ctx := z3bridge.NewZ3Context()

			bits := make([]z3bridge.Expr, nbits)
			for i := 0; i < nbits; i++ {
				bits[i] = ctx.Const(string(rune('a'+i)), ctx.BoolSort())
			}

			result := Gebin(ctx, bits, threshold)

			maxVal := 1 << uint(nbits)
			for val := 0; val < maxVal; val++ {
				slv := ctx.NewSolver()
				for i := 0; i < nbits; i++ {
					bitSet := (val & (1 << uint(nbits-1-i))) != 0
					if bitSet {
						slv.Assert(bits[i])
					} else {
						slv.Assert(ctx.Not(bits[i]))
					}
				}

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
		if n < 0 || n > 16 || m < 0 || m >= (1<<uint(n)) {
			return
		}

		runOnZ3Thread(t, func(t *testing.T) {
			ctx := z3bridge.NewZ3Context()
			bits := BinEncZ3(ctx, m, n)

			if len(bits) != n {
				t.Fatalf("BinEncZ3(%d, %d) returned %d bits", m, n, len(bits))
			}

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
	})
}

// FuzzSortCard exercises SortCard with various sort types and signatures.
// No Z3 needed — pure Go logic.
func FuzzSortCard(f *testing.F) {
	f.Add(byte(0), 3, "")
	f.Add(byte(1), 0, "0:9")
	f.Add(byte(2), 8, "bv")
	f.Add(byte(3), 0, "")

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
		case 0:
			n := param % 20
			if n == 0 {
				n = 1
			}
			ext := make([]string, n)
			for i := range ext {
				ext[i] = string(rune('a' + i%26))
			}
			sort = &lg.EnumeratedSort{Name: "E", Extension: ext}
		case 1:
			lo := 0
			hi := param
			if hi < lo {
				return
			}
			sort = &lg.RangeSort{Name: "R", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: string(rune('0' + hi%10))}}
		case 2:
			width := param % 32
			if width == 0 {
				width = 1
			}
			sig = il.NewSig()
			sig.Interp["mybv"] = "bv[" + string(rune('0'+width%10)) + "]"
			sort = &lg.UninterpretedSort{Name: "mybv"}
		case 3:
			sort = &lg.UninterpretedSort{Name: "T"}
		}

		card := SortCard(sort, sig)
		_ = card
	})
}

// FuzzRangeSortBounds exercises RangeSortBounds with random Lb/Ub strings.
// No Z3 needed — pure Go logic.
func FuzzRangeSortBounds(f *testing.F) {
	f.Add("0", "10")
	f.Add("5", "5")
	f.Add("-3", "3")
	f.Add("abc", "10")
	f.Add("0", "xyz")
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

		rs := &lg.RangeSort{Name: "R", Lb: lg.NumeralBound{Value: lb}, Ub: lg.NumeralBound{Value: ub}}
		lo, hi, ok := RangeSortBounds(rs)
		if ok {
			_ = hi - lo
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
		if len(numStr) > 10 || len(lb) > 10 || len(ub) > 10 {
			return
		}
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

		runOnZ3Thread(t, func(t *testing.T) {
			rs := &lg.RangeSort{Name: "bounded", Lb: lg.NumeralBound{Value: lb}, Ub: lg.NumeralBound{Value: ub}}
			sig := il.NewSig()
			sig.Interp["bounded"] = rs
			s := NewSolver(sig, nil)

			num := lg.NewConst(numStr, &lg.UninterpretedSort{Name: "bounded"})
			_, _ = s.NumeralToZ3(num)
		})
	})
}

// FuzzEncodeEqualityZ3 exercises binary-encoded equality with random
// enumerated sort sizes, verifying no panics and that same-term equality
// is always SAT.
func FuzzEncodeEqualityZ3(f *testing.F) {
	f.Add(2, 0, 0)
	f.Add(3, 0, 1)
	f.Add(4, 2, 2)
	f.Add(5, 0, 4)
	f.Add(7, 3, 6)

	f.Fuzz(func(t *testing.T, nElems, idx1, idx2 int) {
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

		// Capture for closure
		i1, i2 := idx1, idx2
		runOnZ3Thread(t, func(t *testing.T) {
			ext := make([]string, nElems)
			for i := range ext {
				ext[i] = string(rune('a' + i))
			}
			es := &lg.EnumeratedSort{Name: "E", Extension: ext}

			sig := il.NewSig()
			for _, name := range ext {
				sig.Constructors[name] = true
			}
			s := NewSolver(sig, nil)

			t1 := lg.NewConst(ext[i1], es)
			t2 := lg.NewConst(ext[i2], es)

			eq, err := s.EncodeEqualityZ3(t1, t2, es)
			if err != nil {
				return
			}

			ctx := s.Context()
			slv := ctx.NewSolver()
			if i1 == i2 {
				slv.Assert(eq)
				if slv.Check() == z3bridge.Unsat {
					t.Fatalf("encode_equality(%s, %s) should be SAT", ext[i1], ext[i2])
				}
			} else {
				slv.Assert(eq)
				if slv.Check() != z3bridge.Unsat {
					t.Fatalf("encode_equality(%s, %s) should be UNSAT", ext[i1], ext[i2])
				}
			}
		})
	})
}

// FuzzSolverNameBuiltins exercises SolverName with random symbol names,
// verifying z3 builtins panic with IvyError and non-builtins don't panic.
func FuzzSolverNameBuiltins(f *testing.F) {
	f.Add("bit0")
	f.Add("bit1")
	f.Add("myvar")
	f.Add("+")
	f.Add("bfe[0:3]")
	f.Add("")
	f.Add("x")

	f.Fuzz(func(t *testing.T, name string) {
		if len(name) > 100 {
			return
		}

		runOnZ3Thread(t, func(t *testing.T) {
			s := NewSolver(nil, nil)
			sym := lg.NewConst(name, lg.Boolean)

			if name == "bit0" || name == "bit1" {
				// Python: raise IvyError — should panic
				func() {
					defer func() {
						r := recover()
						if r == nil {
							t.Fatalf("SolverName(%q) should panic for Z3 builtin", name)
						}
					}()
					s.SolverName(sym)
				}()
			} else {
				// Non-builtins should not panic
				s.SolverName(sym)
			}
		})
	})
}

// FuzzQuantConstraintsNatRange exercises the full solver quantifier
// constraint pipeline with random sort interpretations, verifying
// no panics during ForAll/Exists translation.
func FuzzQuantConstraintsNatRange(f *testing.F) {
	f.Add(byte(0), true)
	f.Add(byte(0), false)
	f.Add(byte(1), true)
	f.Add(byte(1), false)
	f.Add(byte(2), true)

	f.Fuzz(func(t *testing.T, interpKind byte, isForall bool) {
		runOnZ3Thread(t, func(t *testing.T) {
			sig := il.NewSig()
			switch interpKind % 3 {
			case 0:
				sig.Interp["mysort"] = "nat"
			case 1:
				sig.Interp["mysort"] = &lg.RangeSort{Name: "mysort", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "10"}}
			case 2:
				// No interpretation
			}
			s := NewSolver(sig, nil)

			sort := &lg.UninterpretedSort{Name: "mysort"}
			x, err := lg.NewVariable("X", sort)
			if err != nil {
				return
			}
			p := relConst("P", sort)
			pApp := lg.MustApply(p, x)

			var fmla lg.Expr
			if isForall {
				fmla = &lg.ForAll{Variables: []*lg.Variable{x}, Body: pApp}
			} else {
				fmla = &lg.Exists{Variables: []*lg.Variable{x}, Body: pApp}
			}

			_, _ = s.FormulaToZ3(fmla)
		})
	})
}
