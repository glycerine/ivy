package abc

// Tests for the Go port of src/proof/int2.

import (
	"errors"
	"slices"
	"testing"
)

func int2TruthGraph(t testing.TB, inputs, registers int, bad uint64, next []uint64) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	variables := make([]GiaLit, inputs+registers)
	for i := range variables {
		var err error
		variables[i], err = g.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := g.HashStart(); err != nil {
		t.Fatal(err)
	}
	build := func(mask uint64) GiaLit {
		result := GiaConst0
		for assignment := 0; assignment < 1<<len(variables); assignment++ {
			if mask&(uint64(1)<<assignment) == 0 {
				continue
			}
			term := GiaConst1
			for i, variable := range variables {
				var err error
				term, err = g.HashAnd(term, variable.NotIf(assignment&(1<<i) == 0))
				if err != nil {
					t.Fatal(err)
				}
			}
			var err error
			result, err = g.HashOr(result, term)
			if err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	if _, err := g.AppendCO(build(bad)); err != nil {
		t.Fatal(err)
	}
	for _, mask := range next {
		if _, err := g.AppendCO(build(mask)); err != nil {
			t.Fatal(err)
		}
	}
	g.HashStop()
	if err := g.SetRegisterCount(registers); err != nil {
		t.Fatal(err)
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestInt2DefaultsAndValidation(t *testing.T) {
	p := DefaultInt2Parameters()
	if p.StartFrames != 1 || p.InductionFrames != 1 || p.FrameReached != -1 || p.MaxFrames != 0 || p.ConflictLimit != 0 {
		t.Fatal(p)
	}
	if _, err := PerformInt2Interpolation(nil, p); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	comb := NewGiaManager()
	pi, _ := comb.AppendCI()
	comb.AppendCO(pi)
	if _, err := PerformInt2Interpolation(comb, p); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestInt2UnsafeWitnessAndSafeInvariant(t *testing.T) {
	// q' = input, bad = q: the shortest failure is frame one.
	g := int2TruthGraph(t, 1, 1, 0b1100, []uint64{0b1010})
	p := DefaultInt2Parameters()
	p.ReturnInvariant = true
	r, err := PerformInt2Interpolation(g, p)
	if err != nil || r.Status != Int2Unsafe || r.FrameReached != 1 || r.Counterexample == nil {
		t.Fatalf("%+v %v", r, err)
	}
	if err := r.ValidateWitness(g); err != nil {
		t.Fatal(err)
	}
	// q' = q and bad = q cannot leave the zero initial state.
	safe := int2TruthGraph(t, 1, 1, 0b1100, []uint64{0b1100})
	r, err = PerformInt2Interpolation(safe, p)
	if err != nil || r.Status != Int2Safe || len(r.InvariantStates) != 1 || r.InvariantStates[0][0] {
		t.Fatalf("%+v %v", r, err)
	}
	for _, state := range r.InvariantStates {
		_, next, e := safe.Step([]bool{false}, state)
		if e != nil || !slices.ContainsFunc(r.InvariantStates, func(s []bool) bool { return slices.Equal(s, next) }) {
			t.Fatal(state, next, e)
		}
	}
}

func TestInt2ConstraintsLimitsCancellationAndOwnership(t *testing.T) {
	// bad=q, constraint=input, q'=1.  ABC constraints are violation signals,
	// so only input=false is an admissible step.
	g := NewGiaManager()
	pi, _ := g.AppendCI()
	ro, _ := g.AppendCI()
	g.AppendCO(ro)
	g.AppendCO(pi)
	g.AppendCO(GiaConst1)
	g.SetRegisterCount(1)
	g.SetConstraintCount(1)
	p := DefaultInt2Parameters()
	r, err := PerformInt2Interpolation(g, p)
	if err != nil || r.Status != Int2Unsafe || r.FrameReached != 1 || r.ValidateWitness(g) != nil {
		t.Fatal(r, err)
	}
	// With q'=input, the same constraint excludes the only path to bad=q.
	polarity := NewGiaManager()
	polarityPI, _ := polarity.AppendCI()
	polarityRO, _ := polarity.AppendCI()
	polarity.AppendCO(polarityRO)
	polarity.AppendCO(polarityPI)
	polarity.AppendCO(polarityPI)
	polarity.SetRegisterCount(1)
	polarity.SetConstraintCount(1)
	polarityResult, polarityErr := PerformInt2Interpolation(polarity, DefaultInt2Parameters())
	if polarityErr != nil || polarityResult.Status != Int2Safe {
		t.Fatal(polarityResult, polarityErr)
	}
	p.MaxTransitions = 1
	limited, err := PerformInt2Interpolation(g, p)
	if err != nil || !limited.LimitHit || limited.Status != Int2Unknown {
		t.Fatal(limited, err)
	}
	p = DefaultInt2Parameters()
	p.Stop = func() bool { return true }
	cancelled, err := PerformInt2Interpolation(g, p)
	if err != nil || !cancelled.Cancelled {
		t.Fatal(cancelled, err)
	}
	m, err := NewInt2Manager(g, DefaultInt2Parameters())
	if err != nil {
		t.Fatal(err)
	}
	m.Stop()
	if _, err = m.Perform(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	// Construction snapshots the source.
	m, err = NewInt2Manager(g, DefaultInt2Parameters())
	if err != nil {
		t.Fatal(err)
	}
	g.FlipCODriver(0)
	owned, err := m.Perform()
	if err != nil || owned.Status != Int2Unsafe {
		t.Fatal(owned, err)
	}
}

func TestInt2DuplicateInitAndInitialFramesExhaustive(t *testing.T) {
	g := int2TruthGraph(t, 1, 1, 0b0110, []uint64{0b0110}) // xor(input,q)
	d, err := DuplicateInt2InitialTransition(g)
	if err != nil {
		t.Fatal(err)
	}
	for q := 0; q < 2; q++ {
		for input := 0; input < 2; input++ {
			_, want, _ := g.Step([]bool{input != 0}, []bool{q != 0})
			for control := 0; control < 2; control++ {
				out, got, e := d.Step([]bool{input != 0, control != 0}, []bool{q != 0})
				if e != nil || out[0] != (input != q) || (control != 0 && got[0] != (q != 0)) || (control == 0 && got[0] != want[0]) {
					t.Fatal(q, input, control, out, got, want, e)
				}
			}
		}
	}
	u, err := BuildInt2InitialFrames(g, 3)
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		state := []bool{false}
		for _, v := range inputs {
			_, state, err = g.Step([]bool{v}, state)
			if err != nil {
				t.Fatal(err)
			}
		}
		out, err := u.Evaluate(inputs)
		if err != nil || out[0] != state[0] {
			t.Fatal(bits, out, state, err)
		}
	}
}

func TestInt2TransitionToZeroFrameMapAndRefinement(t *testing.T) {
	g := int2TruthGraph(t, 1, 1, 0b1000, []uint64{0b1000}) // q'=input&q has a zero transition
	yes, err := Int2HasTransitionToZero(g, 0, nil)
	if err != nil || !yes {
		t.Fatal(yes, err)
	}
	noZero := int2TruthGraph(t, 1, 1, 0, []uint64{0b1111})
	_, err = Int2HasTransitionToZero(noZero, 1, nil)
	if !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	m, _ := NewInt2FrameMap(2, 3)
	if _, err = m.Read(0, 0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = m.Write(1, 2, GiaConst1); err != nil {
		t.Fatal(err)
	}
	if l, err := m.Read(1, 2); err != nil || l != GiaConst1 {
		t.Fatal(l, err)
	}
	selected, err := RefineInt2Cube(g, []bool{true, true}, []int{9, 2})
	if err != nil || !slices.Equal(selected, []int{0, 1}) {
		t.Fatal(selected, err)
	}
	// Priority propagation is defined for structural ANDs.  A native mux has
	// a third semantic fanin and must not be silently treated as a two-input AND.
	muxGraph := NewGiaManager()
	control, _ := muxGraph.AppendCI()
	data, _ := muxGraph.AppendCI()
	muxRO, _ := muxGraph.AppendCI()
	mux, _ := muxGraph.AppendMuxReal(control, data, muxRO)
	muxGraph.AppendCO(mux)
	muxGraph.AppendCO(GiaConst0)
	muxGraph.SetRegisterCount(1)
	if _, err = RefineInt2Cube(muxGraph, []bool{true, true, false}, []int{1, 2, 3}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}

func TestInt2ProblemUtilities(t *testing.T) {
	g := int2TruthGraph(t, 1, 2, 0xf0, []uint64{0xaa, 0xf0})
	cover := []Int2Cube{{{Register: 1, Value: true}}}
	covered, err := Int2CoveredRegisterInputs(cover, 2)
	if err != nil || !slices.Equal(covered, []int{2}) {
		t.Fatal(covered, err)
	}
	nodes, err := Int2CollectInternal(g, covered)
	if err != nil || len(nodes) == 0 {
		t.Fatal(nodes, err)
	}
	problem, err := Int2ProblemToGia(g, cover)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		in := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		out, err := problem.Evaluate(in)
		if err != nil || out[0] != (assignment&4 != 0) {
			t.Fatal(assignment, out, err)
		}
	}
	if _, err = Int2CoveredRegisterInputs([]Int2Cube{{{Register: 2}}}, 2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestInt2CollectInternalIncludesMuxControlCone(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	g.HashStart()
	control, _ := g.HashAnd(a, b)
	mux, _ := g.AppendMuxReal(control, b, c)
	g.AppendCO(mux)
	g.HashStop()
	nodes, err := Int2CollectInternal(g, []int{0})
	if err != nil || !slices.Contains(nodes, control.ID()) || !slices.Contains(nodes, mux.ID()) {
		t.Fatal(nodes, err)
	}
}

func TestInt2Deterministic5000Models(t *testing.T) {
	seed := uint64(0x9e3779b97f4a7c15)
	nextRand := func() uint64 { seed ^= seed << 7; seed ^= seed >> 9; seed ^= seed << 8; return seed }
	for model := 0; model < 5000; model++ {
		bad := nextRand() & 0xff
		n0 := nextRand() & 0xff
		n1 := nextRand() & 0xff
		g := int2TruthGraph(t, 1, 2, bad, []uint64{n0, n1})
		// Independent table reachability over four states.
		seen := [4]bool{true, false, false, false}
		frontier := []int{0}
		unsafe := false
		for len(frontier) > 0 && !unsafe {
			nextFront := []int{}
			for _, state := range frontier {
				for input := 0; input < 2; input++ {
					a := input | (state << 1)
					if bad&(1<<a) != 0 {
						unsafe = true
						break
					}
					ns := 0
					if n0&(1<<a) != 0 {
						ns |= 1
					}
					if n1&(1<<a) != 0 {
						ns |= 2
					}
					if !seen[ns] {
						seen[ns] = true
						nextFront = append(nextFront, ns)
					}
				}
			}
			frontier = nextFront
		}
		r, err := PerformInt2Interpolation(g, DefaultInt2Parameters())
		if err != nil || (r.Status == Int2Unsafe) != unsafe {
			t.Fatalf("model %d bad=%02x next=%02x/%02x result=%+v err=%v", model, bad, n0, n1, r, err)
		}
		if unsafe && r.ValidateWitness(g) != nil {
			t.Fatalf("model %d invalid witness", model)
		}
	}
}

func FuzzInt2Reachability(f *testing.F) {
	for _, v := range []uint64{0, 1, 0x55aa, 0xffff, 0xdeadbeef} {
		f.Add(v)
	}
	f.Fuzz(func(t *testing.T, v uint64) {
		bad := v & 0xf
		n := v >> 4 & 0xf
		g := int2TruthGraph(t, 1, 1, bad, []uint64{n})
		r, err := PerformInt2Interpolation(g, DefaultInt2Parameters())
		if err != nil {
			t.Fatal(err)
		}
		if r.Status == Int2Unsafe {
			if err = r.ValidateWitness(g); err != nil {
				t.Fatal(err)
			}
		} else if r.Status != Int2Safe {
			t.Fatal(r)
		}
	})
}
