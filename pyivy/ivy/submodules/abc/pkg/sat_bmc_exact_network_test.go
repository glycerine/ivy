package abc

// Go-only models for the topology and semantic foundations in bmcMaj.c.

import (
	"errors"
	"math/bits"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestExactMajorityTruth(t *testing.T) {
	for variables := 1; variables <= 10; variables++ {
		truth, err := ExactMajorityTruth(variables, ExactSynthesisLimits{})
		if err != nil || len(truth) != 1<<variables {
			t.Fatal(variables, len(truth), err)
		}
		for assignment, value := range truth {
			want := bits.OnesCount(uint(assignment)) > variables/2
			got, err := ExactMajorityValue(assignment, variables)
			if err != nil || got != want || value != want {
				t.Fatal(variables, assignment, value, got, want, err)
			}
		}
	}
}

func TestBuildMajorityExactTopology(t *testing.T) {
	topology, err := BuildMajorityExactTopology(3, 2, true, false, ExactSynthesisLimits{})
	if err != nil || topology.VariableCount != 15 || len(topology.Nodes) != 2 || len(topology.FanoutVariables) != 7 {
		t.Fatal(topology, err)
	}
	for pin, want := range [][]int{{4}, {3}, {2}} {
		if !slices.Equal(topology.Nodes[0].Pins[pin].Candidates, want) {
			t.Fatal(pin, topology.Nodes[0].Pins[pin], want)
		}
	}
	for pin, want := range [][]int{{2, 3, 4, 5}, {2, 3, 4}, {0, 1, 2, 3}} {
		if !slices.Equal(topology.Nodes[1].Pins[pin].Candidates, want) {
			t.Fatal(pin, topology.Nodes[1].Pins[pin], want)
		}
	}
	cascade, err := BuildMajorityExactTopology(3, 3, false, true, ExactSynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for row := 1; row < len(cascade.Nodes); row++ {
		if !slices.Equal(cascade.Nodes[row].Pins[0].Candidates, []int{cascade.Nodes[row].Object - 1}) {
			t.Fatal(row, cascade.Nodes[row].Pins[0])
		}
	}
	clone := topology.Clone()
	clone.Nodes[1].Pins[0].Candidates[0] = 99
	clone.FanoutVariables[2][0] = 99
	if topology.Nodes[1].Pins[0].Candidates[0] == 99 || topology.FanoutVariables[2][0] == 99 {
		t.Fatal("topology clone aliases source storage")
	}
}

func TestBuildLUTExactTopology(t *testing.T) {
	topology, err := BuildLUTExactTopology(3, 2, 2, false, ExactSynthesisLimits{})
	if err != nil || topology.VariableCount != 19 || len(topology.Nodes) != 2 {
		t.Fatal(topology, err)
	}
	if !slices.Equal(topology.Nodes[0].FunctionVariables, []SatVar{1, 2, 3}) || !slices.Equal(topology.Nodes[1].FunctionVariables, []SatVar{4, 5, 6}) {
		t.Fatal(topology.Nodes)
	}
	for row, wants := range [][][]int{{{0, 1, 2}, {0, 1}}, {{0, 1, 2, 3}, {0, 1, 2}}} {
		for pin, want := range wants {
			if !slices.Equal(topology.Nodes[row].Pins[pin].Candidates, want) {
				t.Fatal(row, pin, topology.Nodes[row].Pins[pin], want)
			}
		}
	}
	fewer, err := BuildLUTExactTopology(3, 2, 2, true, ExactSynthesisLimits{})
	if err != nil || !slices.Equal(fewer.Nodes[0].Pins[0].Candidates, []int{1, 2}) || !slices.Equal(fewer.Nodes[0].Pins[1].Candidates, []int{0, 1}) || !slices.Equal(fewer.Nodes[1].Pins[0].Candidates, []int{3}) {
		t.Fatal(fewer, err)
	}
	three, err := BuildLUTExactTopology(4, 2, 3, true, ExactSynthesisLimits{})
	if err != nil || len(three.Nodes[0].FunctionVariables) != 7 || !slices.Equal(three.Nodes[0].Pins[0].Candidates, []int{2, 3}) || !slices.Equal(three.Nodes[0].Pins[2].Candidates, []int{0, 1}) {
		t.Fatal(three, err)
	}
}

func TestExactNetworkEvaluationAndMismatch(t *testing.T) {
	majority := ExactNetwork{Inputs: 3, Gates: []ExactGate{{Fanins: []int{0, 1, 2}, Majority: true}}}
	truth, err := majority.Truth(ExactSynthesisLimits{})
	want, _ := ExactMajorityTruth(3, ExactSynthesisLimits{})
	if err != nil || !slices.Equal(truth, want) {
		t.Fatal(truth, want, err)
	}
	if mismatch, err := majority.FirstMismatch(want, true, ExactSynthesisLimits{}); err != nil || mismatch != -1 {
		t.Fatal(mismatch, err)
	}
	broken := majority.Clone()
	broken.Gates[0] = ExactGate{Fanins: []int{0, 1, 2}, Truth: 0}
	if mismatch, err := broken.FirstMismatch(want, true, ExactSynthesisLimits{}); err != nil || mismatch != 3 {
		t.Fatal(mismatch, err)
	}
	network := ExactNetwork{Inputs: 3, Gates: []ExactGate{
		{Fanins: []int{0, 1}, Truth: 0b0110},
		{Fanins: []int{2, 3}, Truth: 0b1000},
	}}
	for assignment := 0; assignment < 8; assignment++ {
		input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, err := network.Evaluate(input, ExactSynthesisLimits{})
		want := input[2] && (input[0] != input[1])
		if err != nil || got != want {
			t.Fatal(assignment, got, want, err)
		}
	}
	complement := network.Clone()
	complement.ComplementOutput = true
	base, _ := network.Truth(ExactSynthesisLimits{})
	inverted, _ := complement.Truth(ExactSynthesisLimits{})
	for i := range base {
		if base[i] == inverted[i] {
			t.Fatal(i, base[i], inverted[i])
		}
	}
}

func TestExactNetworkDescriptionAndBLIF(t *testing.T) {
	network := ExactNetwork{Inputs: 2, Gates: []ExactGate{{Fanins: []int{ExactConst1, 0}, Truth: 0b1000}, {Fanins: []int{1, 2}, Truth: 0b0110}}}
	description, err := network.Description(ExactSynthesisLimits{})
	if err != nil || description != "n1 = 4'b0110(n0 i1)\nn0 = 4'b1000(i0 const1)\n" {
		t.Fatal(description, err)
	}
	blif, err := network.MarshalBLIF("exact", ExactSynthesisLimits{})
	text := string(blif)
	for _, fragment := range []string{".model exact\n", ".inputs i0 i1\n", ".outputs F\n", ".names const1\n1\n", ".names const1 i0 n0\n", ".names i1 n0 F\n", ".end\n"} {
		if err != nil || !strings.Contains(text, fragment) {
			t.Fatal(fragment, text, err)
		}
	}
	if !strings.Contains(text, "11 1\n") || !strings.Contains(text, "10 1\n") || !strings.Contains(text, "01 1\n") {
		t.Fatal(text)
	}
}

func TestExactNetworkValidationLimitsAndCancellation(t *testing.T) {
	valid := ExactNetwork{Inputs: 2, Gates: []ExactGate{{Fanins: []int{0, 1}, Truth: 8}}}
	for _, invalid := range []ExactNetwork{
		{},
		{Inputs: 2, Gates: []ExactGate{{Fanins: []int{0, 0}, Truth: 8}}},
		{Inputs: 2, Gates: []ExactGate{{Fanins: []int{0, 2}, Truth: 8}}},
		{Inputs: 2, Gates: []ExactGate{{Fanins: []int{0, 1}, Truth: 16}}},
		{Inputs: 2, Gates: []ExactGate{{Fanins: []int{0, 1}, Majority: true}}},
	} {
		if err := invalid.Check(ExactSynthesisLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatal(invalid, err)
		}
	}
	if err := valid.Check(ExactSynthesisLimits{MaxObjects: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := valid.Truth(ExactSynthesisLimits{MaxAssignments: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := valid.Description(ExactSynthesisLimits{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := valid.MarshalBLIF("x", ExactSynthesisLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := BuildLUTExactTopology(3, 2, 2, false, ExactSynthesisLimits{MaxVariables: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := BuildMajorityExactTopology(3, 2, false, false, ExactSynthesisLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	before := valid.Clone()
	_, _ = valid.Truth(ExactSynthesisLimits{})
	if !reflect.DeepEqual(before, valid) {
		t.Fatal("evaluation mutated network")
	}
}

func FuzzExactNetworkTruth(f *testing.F) {
	for _, seed := range []uint16{0, 1, 0x96, 0xff} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, functions uint16) {
		first, second := uint64(functions&15), uint64(functions>>4&15)
		network := ExactNetwork{Inputs: 3, Gates: []ExactGate{
			{Fanins: []int{0, 1}, Truth: first},
			{Fanins: []int{2, 3}, Truth: second},
		}}
		truth, err := network.Truth(ExactSynthesisLimits{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment, got := range truth {
			a, b, c := assignment&1, assignment>>1&1, assignment>>2&1
			inner := first >> (a | b<<1) & 1
			want := second>>(c|int(inner)<<1)&1 != 0
			if got != want {
				t.Fatal(functions, assignment, got, want)
			}
		}
	})
}
