package abc

// Go-only exhaustive tests for src/sat/bmc/bmcEco.c.

import (
	"errors"
	"slices"
	"testing"
)

func ecoTruthGraph(t testing.TB, truth uint8) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	giaTestHash(t, graph)
	inputs := []GiaLit{giaTestCI(t, graph), giaTestCI(t, graph)}
	function := GiaConst0
	for assignment := 0; assignment < 4; assignment++ {
		if truth>>uint(assignment)&1 == 0 {
			continue
		}
		cube := GiaConst1
		for variable, input := range inputs {
			literal := input.NotIf(assignment>>uint(variable)&1 == 0)
			var err error
			cube, err = graph.HashAnd(cube, literal)
			if err != nil {
				t.Fatal(err)
			}
		}
		var err error
		function, err = graph.HashOr(function, cube)
		if err != nil {
			t.Fatal(err)
		}
	}
	giaTestCO(t, graph, function)
	giaTestCheck(t, graph)
	return graph
}

func ecoOldGraph(t testing.TB) (*GiaManager, []GiaID) {
	t.Helper()
	graph := NewGiaManager()
	giaTestHash(t, graph)
	a, b := giaTestCI(t, graph), giaTestCI(t, graph)
	root := giaTestAnd(t, graph, a, b)
	giaTestCO(t, graph, root)
	giaTestCO(t, graph, root)
	giaTestCheck(t, graph)
	return graph, []GiaID{a.ID(), b.ID()}
}

func TestEcoMiterExactSemanticsAndOwnership(t *testing.T) {
	golden := ecoTruthGraph(t, 8) // a & b
	old, fans := ecoOldGraph(t)
	goldObjects, oldObjects := slices.Clone(golden.objects), slices.Clone(old.objects)
	miter, err := BuildEcoMiter(golden, old, fans, EcoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if miter.PrimaryInputCount() != 3 || miter.PrimaryOutputCount() != 4 || miter.Name != golden.Name || miter.Spec != golden.Spec {
		t.Fatal("miter shape", miter.PrimaryInputCount(), miter.PrimaryOutputCount())
	}
	for assignment := 0; assignment < 8; assignment++ {
		a, b, replacement := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		outputs, err := miter.Evaluate([]bool{a, b, replacement})
		if err != nil {
			t.Fatal(err)
		}
		gold := a && b
		want := []bool{gold != replacement, gold != !replacement, a, b}
		if !slices.Equal(outputs, want) {
			t.Fatal(assignment, outputs, want)
		}
	}
	if !slices.Equal(goldObjects, golden.objects) || !slices.Equal(oldObjects, old.objects) {
		t.Fatal("miter construction mutated a source")
	}
}

func TestEcoMiterFlattensSequentialTerminalsLikeSource(t *testing.T) {
	build := func(extra bool) (*GiaManager, []GiaID) {
		graph := NewGiaManager()
		giaTestHash(t, graph)
		a, latch := giaTestCI(t, graph), giaTestCI(t, graph)
		root := giaTestAnd(t, graph, a, latch)
		giaTestCO(t, graph, root)
		if extra {
			giaTestCO(t, graph, root)
		}
		giaTestCO(t, graph, a)
		if err := graph.SetRegisterCount(1); err != nil {
			t.Fatal(err)
		}
		return graph, []GiaID{a.ID(), latch.ID()}
	}
	golden, _ := build(false)
	old, fans := build(true)
	miter, err := BuildEcoMiter(golden, old, fans, EcoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if miter.RegisterCount() != 0 || miter.InputCount() != 3 || miter.OutputCount() != 5 {
		t.Fatal("flattened shape", miter.RegisterCount(), miter.InputCount(), miter.OutputCount())
	}
	for assignment := 0; assignment < 8; assignment++ {
		a, latch, replacement := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		goldPO := a && latch
		got, err := miter.Evaluate([]bool{a, latch, replacement})
		want := []bool{goldPO != replacement, a != replacement, goldPO != !replacement || a != !replacement, a, latch}
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, err)
		}
	}
}

func TestEcoPatchComputedAndImpossible(t *testing.T) {
	for truth := uint8(0); truth < 16; truth++ {
		golden := ecoTruthGraph(t, truth)
		old, fans := ecoOldGraph(t)
		miter, err := BuildEcoMiter(golden, old, fans, EcoOptions{})
		if err != nil {
			t.Fatal(truth, err)
		}
		result, err := ComputeEcoPatch(miter, 2, 1, EcoOptions{})
		if err != nil || result.Status != EcoComputed || result.Reason != SatFinished {
			t.Fatal(truth, result.Status, result.Reason, err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			got, err := result.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
			want := truth>>uint(assignment)&1 != 0
			if err != nil || got != want {
				t.Fatal(truth, assignment, got, want, result.Cubes, err)
			}
		}
		if len(result.Cubes) > 4 {
			t.Fatal("cube bound", truth, result.Cubes)
		}
	}

	golden := ecoTruthGraph(t, 6) // a xor b cannot be expressed using only a
	old, fans := ecoOldGraph(t)
	miter, err := BuildEcoMiter(golden, old, fans[:1], EcoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := ComputeEcoPatch(miter, 1, 1, EcoOptions{})
	if err != nil || result.Status != EcoImpossible {
		t.Fatal(result.Status, result.Cubes, err)
	}
	if _, err = result.Evaluate([]bool{false}); err == nil {
		t.Fatal("incomplete patch must not evaluate")
	}
}

func TestEcoErrorsLimitsAndCancellation(t *testing.T) {
	golden := ecoTruthGraph(t, 8)
	old, fans := ecoOldGraph(t)
	tests := []struct {
		gold, old *GiaManager
		fans      []GiaID
		options   EcoOptions
		want      error
	}{
		{nil, old, fans, EcoOptions{}, ErrDimension},
		{golden, old, []GiaID{InvalidGiaID}, EcoOptions{}, ErrInvalidLiteral},
		{golden, old, fans, EcoOptions{MaxObjects: 1}, ErrLimit},
		{golden, old, fans, EcoOptions{MaxMappingEntries: 1}, ErrLimit},
		{golden, old, fans, EcoOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for index, test := range tests {
		if result, err := BuildEcoMiter(test.gold, test.old, test.fans, test.options); result != nil || !errors.Is(err, test.want) {
			t.Fatal(index, result, err)
		}
	}
	extended := NewGiaManager()
	giaTestHash(t, extended)
	a, b := giaTestCI(t, extended), giaTestCI(t, extended)
	x, err := extended.HashXorReal(a, b)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, extended, x)
	giaTestCO(t, extended, x)
	if _, err = BuildEcoMiter(golden, extended, []GiaID{a.ID()}, EcoOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}

	miter, err := BuildEcoMiter(golden, old, fans, EcoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ComputeEcoPatch(miter, 2, 1, EcoOptions{MaxVariables: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("variable limit", err)
	}
	if _, err = ComputeEcoPatch(miter, 2, 1, EcoOptions{MaxClauses: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("clause limit", err)
	}
	if _, err = ComputeEcoPatch(miter, 2, 1, EcoOptions{MaxLiterals: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("literal limit", err)
	}
	if _, err = ComputeEcoPatch(miter, 2, 1, EcoOptions{MaxCubeLiterals: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("cube literal limit", err)
	}
	limited, err := ComputeEcoPatch(miter, 2, 1, EcoOptions{MaxIterations: 1})
	if err != nil || limited.Status != EcoUnknown || len(limited.Cubes) != 1 {
		t.Fatal("iteration limit", limited, err)
	}
	xorMiter, err := BuildEcoMiter(ecoTruthGraph(t, 6), old, fans, EcoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ComputeEcoPatch(xorMiter, 2, 1, EcoOptions{MaxCubes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("cube limit", err)
	}
	if _, err = ComputeEcoPatch(miter, 2, 1, EcoOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err = ComputeEcoPatch(miter, 1, 1, EcoOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("layout", err)
	}
}

func TestEcoMiterDeepCone(t *testing.T) {
	build := func(outputs int) *GiaManager {
		graph := NewGiaManager()
		giaTestHash(t, graph)
		a, b := giaTestCI(t, graph), giaTestCI(t, graph)
		root := a
		for index := 0; index < 10000; index++ {
			var err error
			root, err = graph.HashAnd(root, b.NotIf(index&1 != 0))
			if err != nil {
				t.Fatal(err)
			}
		}
		for index := 0; index < outputs; index++ {
			giaTestCO(t, graph, root)
		}
		return graph
	}
	golden, old := build(1), build(2)
	miter, err := BuildEcoMiter(golden, old, []GiaID{old.cis[0], old.cis[1]}, EcoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if miter.PrimaryInputCount() != 3 || miter.PrimaryOutputCount() != 4 {
		t.Fatal(miter.PrimaryInputCount(), miter.PrimaryOutputCount())
	}
}

func FuzzEcoPatchTruth(f *testing.F) {
	for truth := byte(0); truth < 16; truth++ {
		f.Add(truth)
	}
	f.Fuzz(func(t *testing.T, raw byte) {
		truth := raw & 15
		golden := ecoTruthGraph(t, truth)
		old, fans := ecoOldGraph(t)
		miter, err := BuildEcoMiter(golden, old, fans, EcoOptions{})
		if err != nil {
			t.Fatal(err)
		}
		result, err := ComputeEcoPatch(miter, 2, 1, EcoOptions{})
		if err != nil || result.Status != EcoComputed {
			t.Fatal(result.Status, err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			values := []bool{assignment&1 != 0, assignment&2 != 0}
			got, err := result.Evaluate(values)
			if err != nil || got != (truth>>uint(assignment)&1 != 0) {
				t.Fatal(truth, assignment, got, result.Cubes, err)
			}
			for _, cube := range result.Cubes {
				seen := [2]bool{}
				for _, literal := range cube {
					if literal.Divisor < 0 || literal.Divisor >= 2 || seen[literal.Divisor] {
						t.Fatal("malformed cube", cube)
					}
					seen[literal.Divisor] = true
				}
			}
		}
	})
}
