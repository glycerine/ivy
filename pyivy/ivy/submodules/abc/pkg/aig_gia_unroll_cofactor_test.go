package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func giaUnrollCofactorTestGraph(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	r0, _ := g.AppendCI()
	r1, _ := g.AppendCI()
	n, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{r0, r1, n, n} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return g
}
func TestGiaUnrollCofactorNativeAndSequentialSemantics(t *testing.T) {
	g := giaUnrollCofactorTestGraph(t)
	g.Name = "pipeline"
	g.Spec = "source.aig"
	raw, err := os.ReadFile("test_vectors/oracle/gia_unroll_cofactor.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]uint32
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 2 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for index, threshold := range []int{0, 2} {
		graph, err := g.UnrollAndCofactor(GiaUnrollCofactorOptions{GiaUnrollOptions: GiaUnrollOptions{Frames: 2}, FanoutThreshold: threshold})
		if err != nil {
			t.Fatal(err)
		}
		if graph.InputCount() != 4 || graph.OutputCount() != 4 || graph.RegisterCount() != 0 || graph.Name != g.Name || graph.Spec != g.Spec {
			t.Fatal("unrolled metadata")
		}
		masks := make([]uint32, 4)
		for assignment := 0; assignment < 16; assignment++ {
			values, err := graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			if err != nil {
				t.Fatal(err)
			}
			expected := assignment&3 == 3
			if values[0] || values[1] || values[2] != expected || values[3] != expected {
				t.Fatal("sequential truth", threshold, assignment, values)
			}
			for output, value := range values {
				if value {
					masks[output] |= 1 << assignment
				}
			}
		}
		if !slices.Equal(masks, native[index]) {
			t.Fatal("native pipeline truth", masks, native[index])
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutated")
	}
}
func TestGiaUnrollCofactorLimitsAndCancellation(t *testing.T) {
	g := giaUnrollCofactorTestGraph(t)
	for _, options := range []GiaUnrollCofactorOptions{{}, {GiaUnrollOptions: GiaUnrollOptions{Frames: 2}, FanoutThreshold: -1}, {GiaUnrollOptions: GiaUnrollOptions{Frames: 2, MaxObjects: 1}}, {GiaUnrollOptions: GiaUnrollOptions{Frames: 2, MaxMappingEntries: 1}}, {GiaUnrollOptions: GiaUnrollOptions{Frames: 2}, MaxCofactorPasses: 1}} {
		if graph, err := g.UnrollAndCofactor(options); err == nil || graph != nil {
			t.Fatal("invalid/resource options", options, err)
		}
	}
	for _, threshold := range []int{1, 10, 30} {
		calls := 0
		graph, err := g.UnrollAndCofactor(GiaUnrollCofactorOptions{GiaUnrollOptions: GiaUnrollOptions{Frames: 3, Stop: func() bool { calls++; return calls >= threshold }}})
		if !errors.Is(err, ErrCancelled) || graph != nil {
			t.Fatal("pipeline cancellation", threshold, err)
		}
	}
}

func TestGiaUnrollCofactorExplicitInitializationAndConstraints(t *testing.T) {
	for _, initial := range [][]LatchInit{{LatchInitOne, LatchInitZero}, {LatchInitUnknown, LatchInitUnknown}} {
		g := giaUnrollCofactorTestGraph(t)
		if err := g.SetConstraintCount(1); err != nil {
			t.Fatal(err)
		}
		graph, err := g.UnrollAndCofactor(GiaUnrollCofactorOptions{GiaUnrollOptions: GiaUnrollOptions{Frames: 2, Initial: initial}})
		if err != nil {
			t.Fatal(err)
		}
		unknowns := 0
		for _, init := range initial {
			if init == LatchInitUnknown {
				unknowns++
			}
		}
		if graph.InputCount() != unknowns+4 || graph.ConstraintCount() != 0 {
			t.Fatal("initial CI/constraint layout")
		}
		for assignment := 0; assignment < 1<<graph.InputCount(); assignment++ {
			bits := make([]bool, graph.InputCount())
			for i := range bits {
				bits[i] = assignment&(1<<i) != 0
			}
			values, err := graph.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			index := 0
			for reg, init := range initial {
				expected := init == LatchInitOne
				if init == LatchInitUnknown {
					expected = bits[index]
					index++
				}
				if values[reg] != expected {
					t.Fatal("initial output", initial, assignment, values)
				}
			}
			expected := bits[unknowns] && bits[unknowns+1]
			if values[2] != expected || values[3] != expected {
				t.Fatal("initialization/constraint frame layout", values)
			}
		}
	}
}

func TestGiaUnrollCofactorIndependentTruthMachines(t *testing.T) {
	for trial := 0; trial < 64; trial++ {
		truth := bmcTruthCase{bad: uint8(trial * 53), constraint: uint8(trial * 37), next0: uint8(trial * 73), next1: uint8(trial * 131), initial: [2]LatchInit{LatchInit(trial & 1), LatchInit((trial >> 1) & 1)}}
		model := bmcTruthModel(t, truth)
		source, _, err := GiaFromAig(model.Graph())
		if err != nil {
			t.Fatal(err)
		}
		graph, err := source.UnrollAndCofactor(GiaUnrollCofactorOptions{GiaUnrollOptions: GiaUnrollOptions{Frames: 2, Initial: truth.initial[:]}})
		if err != nil {
			t.Fatal(trial, err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			values, err := graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
			if err != nil {
				t.Fatal(err)
			}
			state := int(truth.initial[0]) | int(truth.initial[1])<<1
			for frame := 0; frame < 2; frame++ {
				index := 2*state + ((assignment >> frame) & 1)
				if values[2*frame] != (truth.bad&(1<<index) != 0) || values[2*frame+1] != (truth.constraint&(1<<index) != 0) {
					t.Fatal("independent pipeline truth", trial, assignment, frame, values)
				}
				state = 0
				if truth.next0&(1<<index) != 0 {
					state |= 1
				}
				if truth.next1&(1<<index) != 0 {
					state |= 2
				}
			}
		}
	}
}
