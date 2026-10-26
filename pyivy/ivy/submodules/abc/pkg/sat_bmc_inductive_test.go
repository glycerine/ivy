package abc

// Go-only exhaustive and fuzz tests for src/sat/bmc/bmcICheck.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func mInductiveTruthGraph(t testing.TB, output, next0, next1 uint8) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	input, _ := graph.AppendCI()
	register0, _ := graph.AppendCI()
	register1, _ := graph.AppendCI()
	if err := graph.HashStart(); err != nil {
		t.Fatal(err)
	}
	variables := []GiaLit{input, register0, register1}
	build := func(table uint8) GiaLit {
		root := GiaConst0
		for assignment := 0; assignment < 8; assignment++ {
			if table>>assignment&1 == 0 {
				continue
			}
			cube := GiaConst1
			for variable, literal := range variables {
				var err error
				cube, err = graph.HashAnd(cube, literal.NotIf(assignment>>variable&1 == 0))
				if err != nil {
					t.Fatal(err)
				}
			}
			var err error
			root, err = graph.HashOr(root, cube)
			if err != nil {
				t.Fatal(err)
			}
		}
		return root
	}
	graph.AppendCO(build(output))
	graph.AppendCO(build(next0))
	graph.AppendCO(build(next1))
	graph.HashStop()
	if err := graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return graph
}

func mInductiveReference(output, next0, next1 uint8, selected []bool, frames int) bool {
	evaluate := func(table uint8, input, state int) int {
		assignment := input | (state&1)<<1 | (state>>1&1)<<2
		return int(table >> assignment & 1)
	}
	for leftInitial := 0; leftInitial < 4; leftInitial++ {
		for rightInitial := 0; rightInitial < 4; rightInitial++ {
			for inputs := 0; inputs < 1<<(frames+1); inputs++ {
				left, right, admissible := leftInitial, rightInitial, true
				for frame := 0; frame <= frames; frame++ {
					input := inputs >> frame & 1
					leftOutput, rightOutput := evaluate(output, input, left), evaluate(output, input, right)
					leftNext := evaluate(next0, input, left) | evaluate(next1, input, left)<<1
					rightNext := evaluate(next0, input, right) | evaluate(next1, input, right)<<1
					if frame < frames {
						if leftOutput != rightOutput {
							admissible = false
						}
						for register := range selected {
							if selected[register] && leftNext>>register&1 != rightNext>>register&1 {
								admissible = false
							}
						}
					} else {
						difference := leftOutput != rightOutput
						for register := range selected {
							difference = difference || selected[register] && leftNext>>register&1 != rightNext>>register&1
						}
						if admissible && difference {
							return false
						}
					}
					left, right = leftNext, rightNext
				}
			}
		}
	}
	return true
}

func TestGiaMInductiveExactAndMinimizers(t *testing.T) {
	// output=r0, next0=r1, next1=!r1. With no selected register two
	// trajectories can agree now and differ next. Either selected register is
	// sufficient to exclude that pair.
	graph := mInductiveTruthGraph(t, 0xcc, 0xf0, 0x0f)
	before := graph.Clone()
	for mask := 0; mask < 4; mask++ {
		selected := []bool{mask&1 != 0, mask&2 != 0}
		result, err := graph.CheckMInductiveSet(selected, GiaMInductiveOptions{Frames: 1})
		if err != nil {
			t.Fatal(mask, err)
		}
		want := mInductiveReference(0xcc, 0xf0, 0x0f, selected, 1)
		if (result.Status == GiaMInductive) != want || result.Frame != 1 || result.SATChecks != 1 || !slices.Equal(result.Selected, selected) {
			t.Fatal(mask, result, want)
		}
		result.Selected[0] = !result.Selected[0]
		if selected[0] != (mask&1 != 0) {
			t.Fatal("result aliases caller")
		}
	}

	core, err := graph.MinimizeMInductiveCore(GiaMInductiveOptions{Frames: 1})
	if err != nil || core.Status != GiaMInductive || core.CheckedFrames != 1 || core.SATChecks == 0 {
		t.Fatal("core", core, err)
	}
	verified, err := graph.CheckMInductiveSet(core.Selected, GiaMInductiveOptions{Frames: 1})
	if err != nil || verified.Status != GiaMInductive {
		t.Fatal("unverified core result", core.Selected, verified, err)
	}
	empty, err := graph.MinimizeMInductiveCore(GiaMInductiveOptions{Frames: 1, Empty: true})
	if err != nil || empty.Status != GiaMNotInductive || !slices.Equal(empty.Selected, []bool{false, false}) {
		t.Fatal("empty", empty, err)
	}

	natural, err := graph.SearchMInductiveRegisters(GiaMInductiveOptions{Frames: 1})
	if err != nil || natural.Status != GiaMInductive || !slices.Equal(natural.Selected, []bool{false, true}) {
		t.Fatal("natural", natural, err)
	}
	reverse, err := graph.SearchMInductiveRegisters(GiaMInductiveOptions{Frames: 1, Reverse: true})
	if err != nil || reverse.Status != GiaMInductive || !slices.Equal(reverse.Selected, []bool{true, false}) {
		t.Fatal("reverse", reverse, err)
	}
	backwardGraph := mInductiveTruthGraph(t, 0xf0, 0x33, 0xcc)
	backward, err := backwardGraph.SearchMInductiveRegisters(GiaMInductiveOptions{Frames: 1, BackwardTopological: true})
	if err != nil || backward.Status != GiaMInductive || !slices.Equal(backward.Selected, []bool{true, false}) {
		t.Fatal("backward", backward, err)
	}
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("source mutated")
	}
}

func TestGiaMInductiveBackwardRegisterOrder(t *testing.T) {
	graph := NewGiaManager()
	register0, _ := graph.AppendCI()
	register1, _ := graph.AppendCI()
	register2, _ := graph.AppendCI()
	graph.AppendCO(register2)
	graph.AppendCO(register1)
	graph.AppendCO(GiaConst0)
	graph.AppendCO(register0)
	graph.SetRegisterCount(3)
	order, err := graph.BackwardRegisterOrder(0)
	if err != nil || !slices.Equal(order, []int{2, 0, 1}) {
		t.Fatal(order, err)
	}
	order[0] = 99
	again, _ := graph.BackwardRegisterOrder(0)
	if again[0] != 2 {
		t.Fatal("order aliases manager")
	}
	if _, err = graph.BackwardRegisterOrder(1); !errors.Is(err, ErrLimit) {
		t.Fatal("order limit", err)
	}
}

func TestGiaMInductiveErrorsLimitsUnknownAndCancellation(t *testing.T) {
	graph := mInductiveTruthGraph(t, 0xcc, 0xf0, 0x0f)
	for _, test := range []struct {
		selected []bool
		options  GiaMInductiveOptions
		want     error
	}{
		{[]bool{true, true}, GiaMInductiveOptions{}, ErrDimension},
		{[]bool{true}, GiaMInductiveOptions{Frames: 1}, ErrDimension},
		{[]bool{true, true}, GiaMInductiveOptions{Frames: 1, MaxObjects: 1}, ErrLimit},
		{[]bool{true, true}, GiaMInductiveOptions{Frames: 1, MaxMappingEntries: 1}, ErrLimit},
		{[]bool{true, true}, GiaMInductiveOptions{Frames: 1, MaxVariables: 1}, ErrLimit},
		{[]bool{true, true}, GiaMInductiveOptions{Frames: 1, MaxClauses: 1}, ErrLimit},
		{[]bool{true, true}, GiaMInductiveOptions{Frames: 1, MaxLiterals: 1}, ErrLimit},
		{[]bool{true, true}, GiaMInductiveOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { return true }}}, ErrCancelled},
	} {
		if _, err := graph.CheckMInductiveSet(test.selected, test.options); !errors.Is(err, test.want) {
			t.Fatal(test.options, err, test.want)
		}
	}
	if _, err := (*GiaManager)(nil).CheckMInductiveSet(nil, GiaMInductiveOptions{Frames: 1}); !errors.Is(err, ErrDimension) {
		t.Fatal("nil manager", err)
	}
	comb := NewGiaManager()
	input, _ := comb.AppendCI()
	comb.AppendCO(input)
	if _, err := comb.CheckMInductiveSet(nil, GiaMInductiveOptions{Frames: 1}); !errors.Is(err, ErrDimension) {
		t.Fatal("combinational", err)
	}
	extended := NewGiaManager()
	input, _ = extended.AppendCI()
	register, _ := extended.AppendCI()
	xor, _ := extended.AppendXorReal(input, register)
	extended.AppendCO(xor)
	extended.AppendCO(register)
	extended.SetRegisterCount(1)
	if _, err := extended.CheckMInductiveSet([]bool{true}, GiaMInductiveOptions{Frames: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended gate", err)
	}
	unknown, err := graph.CheckMInductiveSet([]bool{false, false}, GiaMInductiveOptions{Frames: 1, SAT: SatLimits{Propagations: 1}})
	if err != nil || unknown.Status != GiaMInductiveUnknown || unknown.Reason != SatPropagationLimit {
		t.Fatal("unknown", unknown, err)
	}
	searchUnknown, err := graph.SearchMInductiveRegisters(GiaMInductiveOptions{Frames: 1, SAT: SatLimits{Propagations: 1}})
	if err != nil || searchUnknown.Status != GiaMInductiveUnknown {
		t.Fatal("search unknown", searchUnknown, err)
	}
	if _, err = graph.SearchMInductiveRegisters(GiaMInductiveOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { return true }}}); !errors.Is(err, ErrCancelled) {
		t.Fatal("search cancellation", err)
	}
}

func TestGiaMInductiveDeepIterativeCone(t *testing.T) {
	graph := NewGiaManager()
	input, _ := graph.AppendCI()
	register, _ := graph.AppendCI()
	root := register
	for index := 0; index < 10_000; index++ {
		var err error
		if index&1 == 0 {
			root, err = graph.AppendAnd(root, input.Not())
		} else {
			root, err = graph.AppendAnd(root, input)
		}
		if err != nil {
			t.Fatal(index, err)
		}
	}
	graph.AppendCO(root)
	graph.AppendCO(register)
	graph.SetRegisterCount(1)
	result, err := graph.CheckMInductiveSet([]bool{true}, GiaMInductiveOptions{Frames: 1})
	if err != nil || result.Status != GiaMInductive {
		t.Fatal(result, err)
	}
}

func FuzzGiaMInductive(f *testing.F) {
	for _, seed := range [][6]byte{
		{0, 0, 0, 0, 1, 0},
		{0xcc, 0xf0, 0x0f, 0, 1, 0},
		{0xaa, 0x55, 0x33, 3, 2, 0},
		{0xff, 0xcc, 0xf0, 1, 2, 0},
	} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4], seed[5])
	}
	f.Fuzz(func(t *testing.T, output, next0, next1, mask, depth, _ byte) {
		selected := []bool{mask&1 != 0, mask&2 != 0}
		frames := 1 + int(depth&1)
		graph := mInductiveTruthGraph(t, output, next0, next1)
		result, err := graph.CheckMInductiveSet(selected, GiaMInductiveOptions{Frames: frames})
		if err != nil {
			t.Fatal(err)
		}
		want := mInductiveReference(output, next0, next1, selected, frames)
		if result.Status == GiaMInductiveUnknown || (result.Status == GiaMInductive) != want {
			t.Fatal(output, next0, next1, selected, frames, result.Status, want)
		}
	})
}
