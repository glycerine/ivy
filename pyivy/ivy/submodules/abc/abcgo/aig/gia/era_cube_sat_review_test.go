package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"slices"
	"strings"
	"testing"
)

func TestCubeReachabilityCounterexampleSATBindings(t *testing.T) {
	for _, typed := range []bool{false, true} {
		t.Run(fmt.Sprint(typed), func(t *testing.T) {
			g := Gia_ManStart(24)
			a, b, r := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendXor(g, a, b)
			if typed {
				x = Gia_ManAppendXorReal(g, a, b)
			}
			Gia_ManAppendCo(g, x)
			Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, r))
			Gia_ManSetRegNum(g, 1)
			p := Gia_ManAreCreate(g)
			defer Gia_ManAreFree(p)
			Gia_ManAreDeriveCexSatStart(p)
			// C exposes every CI and CO in one persistent combinational CNF. Bind all
			// inputs, query both output polarities, and ensure contradictory assumptions
			// do not poison later calls on the same solver.
			if p.vSatNumCis == nil || p.vSatNumCos == nil || Vec_IntSize(p.vSatNumCis) != 3 || Vec_IntSize(p.vSatNumCos) != 2 {
				t.Fatalf("missing SAT interface: CI%v CO%v", p.vSatNumCis, p.vSatNumCos)
			}
			solver, ok := any(p.pSat).(*bsat.Solver)
			if !ok || solver == nil {
				t.Fatalf("counterexample SAT backend has type %T", p.pSat)
			}
			for bits := 0; bits < 8; bits++ {
				in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
				want, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				assumptions := make([]bsat.Literal, 3)
				for i, v := range in {
					variable := Vec_IntArray(p.vSatNumCis)[i]
					if variable < 0 {
						t.Fatal("CI missing from CNF")
					}
					assumptions[i] = bsat.Negative(variable)
					if v {
						assumptions[i] = bsat.Positive(variable)
					}
				}
				for output, value := range want {
					variable := Vec_IntArray(p.vSatNumCos)[output]
					lit := bsat.Negative(variable)
					if value {
						lit = bsat.Positive(variable)
					}
					if got := solver.Solve(append(assumptions, lit)...); got != bsat.Satisfiable {
						t.Fatalf("input%d CO%d expected SAT, got%v", bits, output, got)
					}
					if got := solver.Solve(append(assumptions, lit^1)...); got != bsat.Unsatisfiable {
						t.Fatalf("input%d CO%d wrong polarity expected UNSAT, got%v", bits, output, got)
					}
				}
			}
			p.pTarget = Gia_ManAreCreateSta(p)
			Gia_ManAreDeriveCexSatStop(p)
			if p.pSat != nil || p.pTarget != nil || p.vSatNumCis != nil || p.vSatNumCos != nil {
				t.Fatal("SAT state retained after stop")
			}
		})
	}
}

func TestCubeReachabilityCounterexampleSATTransitions(t *testing.T) {
	g := Gia_ManStart(32)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a, r0)
	m := Gia_ManAppendMuxReal(g, b, r1, r0)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, m)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, m)
	Gia_ManSetRegNum(g, 2)
	p := Gia_ManAreCreate(g)
	defer Gia_ManAreFree(p)
	Gia_ManAreDeriveCexSatStart(p)
	original := p.pSat
	// Trit zero is don't-care, one is fixed zero, two is fixed one.
	matches := func(code int, values []bool) bool {
		for _, v := range values {
			if code%3 == 1 && v || code%3 == 2 && !v {
				return false
			}
			code /= 3
		}
		return true
	}
	state := func(code int) *Gia_StaAre_t {
		s := Gia_ManAreCreateSta(p)
		for i := int32(0); i < 2; i++ {
			if code%3 == 1 {
				Gia_StaSetValue0(s, i)
			}
			if code%3 == 2 {
				Gia_StaSetValue1(s, i)
			}
			code /= 3
		}
		return s
	}
	for cur := 0; cur < 9; cur++ {
		for next := -1; next < 9; next++ {
			for output := int32(-1); output < 2; output++ {
				var witnesses []int
				for mask := 0; mask < 16; mask++ {
					in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					out, err := g.Evaluate(in)
					if err != nil {
						t.Fatal(err)
					}
					if matches(cur, in[2:]) && (next < 0 || matches(next, out[2:])) && (output < 0 || out[output]) {
						witnesses = append(witnesses, mask)
					}
				}
				current := state(cur)
				var target *Gia_StaAre_t
				if next >= 0 {
					target = state(next)
				}
				var failure any
				func() { defer func() { failure = recover() }(); Gia_ManAreDeriveCexSat(p, current, target, output) }()
				if len(witnesses) == 0 {
					if failure == nil {
						t.Fatalf("cur%d next%d PO%d accepted impossible transition", cur, next, output)
					}
					if !strings.Contains(fmt.Sprint(failure), "could not reconstruct reachability transition") {
						t.Fatalf("unexpected panic: %v", failure)
					}
					continue
				}
				if failure != nil {
					t.Fatalf("cur%d next%d PO%d rejected valid transition: %v", cur, next, output, failure)
				}
				mask := 0
				for _, pi := range Vec_IntArray(p.vCofVars) {
					if pi < 0 || pi >= 2 {
						t.Fatal("invalid PI index")
					}
					mask |= 1 << pi
				}
				for i := int32(0); i < 2; i++ {
					zero, one := Gia_StaHasValue0(current, i), Gia_StaHasValue1(current, i)
					if zero+one != 1 {
						t.Fatal("model did not complete current cube")
					}
					if one != 0 {
						mask |= 1 << (i + 2)
					}
				}
				if !slices.Contains(witnesses, mask) {
					t.Fatalf("cur%d next%d PO%d invalid model%d, want one of%v", cur, next, output, mask, witnesses)
				}
				if p.pSat != original {
					t.Fatal("query replaced persistent solver")
				}
			}
		}
	}
	p.pTarget = Gia_ManAreCreateSta(p)
	Gia_ManAreDeriveCexSatStop(p)
}
