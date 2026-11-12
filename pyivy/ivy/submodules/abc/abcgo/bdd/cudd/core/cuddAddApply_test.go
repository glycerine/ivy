package core

import (
	"math"
	"reflect"
	"testing"
)

func TestADDApplyPointwiseArithmetic(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := Cudd_addIthVar(dd, 0), Cudd_addIthVar(dd, 1)
	f := Cudd_addIte(dd, x, Cudd_addConst(dd, 4), Cudd_addConst(dd, -2))
	g := Cudd_addIte(dd, y, Cudd_addConst(dd, 3), dd.Zero)
	for _, tc := range []struct {
		name string
		op   DD_AOP
		want func(float64, float64) float64
	}{
		{"plus", Cudd_addPlus, func(a, b float64) float64 { return a + b }},
		{"times", Cudd_addTimes, func(a, b float64) float64 { return a * b }},
		{"minus", Cudd_addMinus, func(a, b float64) float64 { return a - b }},
		{"divide", Cudd_addDivide, func(a, b float64) float64 {
			if a == 0 {
				return 0
			}
			return a / b
		}},
		{"min", Cudd_addMinimum, math.Min}, {"max", Cudd_addMaximum, math.Max},
		{"threshold", Cudd_addThreshold, func(a, b float64) float64 {
			if a >= b {
				return a
			}
			return 0
		}},
		{"setnz", Cudd_addSetNZ, func(a, b float64) float64 {
			if b != 0 {
				return b
			}
			return a
		}},
		{"onezero", Cudd_addOneZeroMaximum, func(a, b float64) float64 {
			if a > b {
				return 1
			}
			return 0
		}},
		{"diff", Cudd_addDiff, func(a, b float64) float64 {
			if a == b {
				return math.Inf(1)
			}
			return math.Min(a, b)
		}},
		{"agreement", Cudd_addAgreement, func(a, b float64) float64 {
			if a == b {
				return a
			}
			return 0
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := Cudd_addApply(dd, tc.op, f, g)
			if n == nil {
				t.Fatal("nil result")
			}
			for bits := 0; bits < 4; bits++ {
				assignment := []bool{bits&1 != 0, bits&2 != 0}
				a, _ := dd.EvaluateADD(f, assignment)
				b, _ := dd.EvaluateADD(g, assignment)
				got, err := dd.EvaluateADD(n, assignment)
				if err != nil || got != tc.want(a, b) {
					t.Fatal(bits, got, tc.want(a, b), err)
				}
			}
		})
	}
	if !reflect.DeepEqual(dd.Support(f), []int{0}) || !reflect.DeepEqual(dd.Support(Cudd_addApply(dd, Cudd_addPlus, f, g)), []int{0, 1}) {
		t.Fatal("ADD support missing")
	}
	if Cudd_Support(dd, f) != dd.Vars[0] {
		t.Fatal("ADD support cube not canonical BDD")
	}
}
func TestADDApplyBooleanAndQueryContracts(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := Cudd_addIthVar(dd, 0), Cudd_addIthVar(dd, 1)
	for _, tc := range []struct {
		op   DD_AOP
		want func(bool, bool) bool
	}{
		{Cudd_addOr, func(a, b bool) bool { return a || b }}, {Cudd_addNand, func(a, b bool) bool { return !(a && b) }},
		{Cudd_addNor, func(a, b bool) bool { return !(a || b) }}, {Cudd_addXor, func(a, b bool) bool { return a != b }}, {Cudd_addXnor, func(a, b bool) bool { return a == b }},
	} {
		n := Cudd_addApply(dd, tc.op, x, y)
		for bits := 0; bits < 4; bits++ {
			a, b := bits&1 != 0, bits&2 != 0
			v, err := dd.EvaluateADD(n, []bool{a, b})
			if err != nil || (v == 1) != tc.want(a, b) {
				t.Fatal(bits, v, err)
			}
		}
	}
	a := Cudd_addConst(dd, 7)
	f := Cudd_addIte(dd, x, a, dd.Zero)
	before := Cudd_ReadNodeCount(dd)
	if Cudd_addEvalConst(dd, x, f) != a || Cudd_addEvalConst(dd, x, y) != DD_NON_CONSTANT {
		t.Fatal("evaluation constancy query")
	}
	if Cudd_addIteConstant(dd, x, a, a) != a || Cudd_addIteConstant(dd, x, a, dd.Zero) != DD_NON_CONSTANT {
		t.Fatal("ITE constancy query")
	}
	if !Cudd_addLeq(dd, dd.Zero, f) || Cudd_addLeq(dd, f, dd.Zero) {
		t.Fatal("numeric implication")
	}
	if Cudd_ReadNodeCount(dd) != before {
		t.Fatal("query allocated nodes")
	}
	complement := Cudd_addCmpl(dd, f)
	for _, assignment := range [][]bool{{false, false}, {true, false}} {
		v, _ := dd.EvaluateADD(complement, assignment)
		want := 1.0
		if assignment[0] {
			want = 0
		}
		if v != want {
			t.Fatal(v, want)
		}
	}
}
func TestADDApplyCallbackIsolationAndFailure(t *testing.T) {
	dd := NewBDDManager(1)
	x := Cudd_addIthVar(dd, 0)
	f := Cudd_addIte(dd, x, Cudd_addConst(dd, 4), dd.One)
	makeOffset := func(offset float64) DD_MAOP {
		return func(dd *DdManager, n *DdNode) *DdNode {
			if Cudd_IsConstant(n) {
				return Cudd_addConst(dd, n.Value+offset)
			}
			return nil
		}
	}
	a := Cudd_addMonadicApply(dd, makeOffset(2), f)
	b := Cudd_addMonadicApply(dd, makeOffset(3), f)
	av, _ := dd.EvaluateADD(a, []bool{true})
	bv, _ := dd.EvaluateADD(b, []bool{true})
	if av != 6 || bv != 7 {
		t.Fatal("callback cache collision", av, bv)
	}
	neg := Cudd_addNegate(dd, f)
	v, _ := dd.EvaluateADD(neg, []bool{true})
	if v != -4 {
		t.Fatal(v)
	}
	logged := Cudd_addMonadicApply(dd, Cudd_addLog, f)
	v, _ = dd.EvaluateADD(logged, []bool{true})
	if v != math.Log(4) {
		t.Fatal(v)
	}
	if n := Cudd_addMonadicApply(dd, Cudd_addLog, Cudd_addConst(dd, -1)); n == nil || !math.IsNaN(n.Value) {
		t.Fatal("negative logarithm lost IEEE NaN")
	}
	if Cudd_addApply(dd, func(*DdManager, **DdNode, **DdNode) *DdNode { return nil }, dd.One, dd.Zero) != nil {
		t.Fatal("unsupported terminal recursed")
	}
	zero, inf := dd.Zero, dd.PlusInfinity
	if Cudd_addTimes(dd, &zero, &inf) != dd.Zero {
		t.Fatal("zero-times-infinity compatibility")
	}
}
