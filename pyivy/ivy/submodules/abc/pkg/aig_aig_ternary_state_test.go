package abc

// Tests for sequential ternary-state simulation from aig/aig/aigTsim.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigTernaryStateOracle struct {
	Hashes []struct {
		Seed, Count, Table, Hash int
	}
	Simulations []struct {
		Seed    int
		Mapping []AigLit
	}
}

func aigTernaryStateGraph(t testing.TB, seed int) *AigManager {
	t.Helper()
	graph := NewAigManager()
	variables := aigObjectVariables(t, graph, 6)
	p0, p1 := variables[0], variables[1]
	s0, s2 := variables[2], variables[4]
	n0 := aigObjectAnd(t, graph, s0.NotIf(seed&8 != 0), p0.NotIf(seed&16 != 0))
	drivers := []AigLit{
		AigConst0.NotIf(seed&1 != 0),
		s0.NotIf(seed&2 != 0),
		s2.NotIf(seed&4 != 0),
		n0.NotIf(seed&8 != 0),
	}
	for _, driver := range []AigLit{p0.NotIf(seed&1 != 0), p1.NotIf(seed&2 != 0)} {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	for _, driver := range drivers {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.SetRegisterCount(4); err != nil {
		t.Fatal(err)
	}
	return graph
}

func aigTernaryHashWords(seed, count int) []uint32 {
	words := make([]uint32, count)
	for index := range words {
		words[index] = uint32(0x9e3779b9*uint32(index+1)) ^ uint32(seed*0x10203)
	}
	return words
}

func TestAigTernaryStateNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_ternary_state.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle aigTernaryStateOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Hashes) != 32 || len(oracle.Simulations) != 32 {
		t.Fatal(len(oracle.Hashes), len(oracle.Simulations), err)
	}
	for _, row := range oracle.Hashes {
		got, err := AigTernaryStateHash(aigTernaryHashWords(row.Seed, row.Count), row.Table)
		if err != nil || got != row.Hash {
			t.Fatal(row, got, err)
		}
	}
	for _, row := range oracle.Simulations {
		graph := aigTernaryStateGraph(t, row.Seed)
		before := graph.Clone()
		result, err := graph.SequentialTernaryConstants(AigTernaryStateOptions{})
		if err != nil || !slices.Equal(result.Mapping, row.Mapping) || !reflect.DeepEqual(graph, before) {
			t.Fatal(row.Seed, result, row.Mapping, err, reflect.DeepEqual(graph, before))
		}
		if (result.Constants == 0) != (row.Mapping == nil) || result.States != result.Rounds+1 || len(result.Registers) != graph.RegisterCount() {
			t.Fatal(row.Seed, result)
		}
		if result.Mapping != nil {
			for index, value := range result.Registers {
				mapping := result.Mapping[graph.PrimaryInputCount()+index]
				switch value {
				case AigTernaryZero:
					if mapping != AigConst0 {
						t.Fatal(row.Seed, index, value, mapping)
					}
				case AigTernaryOne:
					if mapping != AigConst1 {
						t.Fatal(row.Seed, index, value, mapping)
					}
				case AigTernaryUnknown:
					if mapping != AigLit(graph.cis[graph.PrimaryInputCount()+index])<<1 {
						t.Fatal(row.Seed, index, value, mapping)
					}
				default:
					t.Fatal(row.Seed, index, value)
				}
			}
		}
	}
}

func TestAigTernaryStateErrorsAndDeep(t *testing.T) {
	graph := aigTernaryStateGraph(t, 0)
	for _, test := range []struct {
		graph   *AigManager
		options AigTernaryStateOptions
		want    error
	}{
		{nil, AigTernaryStateOptions{}, ErrDimension},
		{graph, AigTernaryStateOptions{MaxRounds: -1}, ErrDimension},
		{graph, AigTernaryStateOptions{MaxObjects: -1}, ErrDimension},
		{graph, AigTernaryStateOptions{MaxStateWords: -1}, ErrDimension},
		{graph, AigTernaryStateOptions{MaxRounds: 1}, ErrLimit},
		{graph, AigTernaryStateOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{graph, AigTernaryStateOptions{MaxStateWords: 1}, ErrLimit},
		{graph, AigTernaryStateOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := test.graph.SequentialTernaryConstants(test.options); !reflect.DeepEqual(result, AigTernaryConstants{}) || !errors.Is(err, test.want) {
			t.Fatal(result, err, test.want)
		}
	}
	if _, err := AigTernaryStateHash(nil, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	buffered := NewAigManager()
	input := aigObjectVariables(t, buffered, 1)[0]
	buffer, err := buffered.CreateBuffer(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = buffered.CreateCO(buffer); err != nil {
		t.Fatal(err)
	}
	if _, err = buffered.SequentialTernaryConstants(AigTernaryStateOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	combinational := NewAigManager()
	input = aigObjectVariables(t, combinational, 1)[0]
	if _, err = combinational.CreateCO(input); err != nil {
		t.Fatal(err)
	}
	if result, err := combinational.SequentialTernaryConstants(AigTernaryStateOptions{}); err != nil || result.Constants != 0 || result.Mapping != nil {
		t.Fatal(result, err)
	}

	deep := NewAigManager()
	variables := aigObjectVariables(t, deep, 2)
	root := variables[0]
	for index := 0; index < 10_000; index++ {
		root = aigObjectAnd(t, deep, root, variables[1].NotIf(index&1 != 0))
	}
	if _, err = deep.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	if _, err = deep.CreateCO(AigConst0); err != nil {
		t.Fatal(err)
	}
	if err = deep.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if result, err := deep.SequentialTernaryConstants(AigTernaryStateOptions{MaxObjects: deep.ObjectCount()}); err != nil || result.Constants != 1 || !slices.Equal(result.Mapping, []AigLit{2, AigConst0}) {
		t.Fatal(result, err)
	}
}

func aigTernaryFuzzGraph(t testing.TB, selectors [3]uint8, phases uint8) *AigManager {
	t.Helper()
	graph := NewAigManager()
	variables := aigObjectVariables(t, graph, 5)
	candidates := []AigLit{AigConst0, AigConst1, variables[0], variables[1], variables[2], variables[3], variables[4]}
	drivers := make([]AigLit, 3)
	for index := range drivers {
		left := candidates[int(selectors[index])%len(candidates)].NotIf(phases&(1<<uint(index)) != 0)
		right := candidates[(int(selectors[index])/len(candidates)+index+1)%len(candidates)].NotIf(phases&(1<<uint(index+3)) != 0)
		if selectors[index]&0x80 != 0 {
			drivers[index] = aigObjectAnd(t, graph, left, right)
		} else {
			drivers[index] = left
		}
	}
	for _, driver := range drivers {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.SetRegisterCount(3); err != nil {
		t.Fatal(err)
	}
	return graph
}

func FuzzAigTernaryState(f *testing.F) {
	for seed := 0; seed < 32; seed++ {
		f.Add(uint8(seed), uint8(seed*7), uint8(seed*13), uint8(seed*29))
	}
	f.Fuzz(func(t *testing.T, a, b, c, phases uint8) {
		graph := aigTernaryFuzzGraph(t, [3]uint8{a, b, c}, phases)
		before := graph.Clone()
		result, err := graph.SequentialTernaryConstants(AigTernaryStateOptions{})
		if err != nil || !reflect.DeepEqual(graph, before) {
			t.Fatal(result, err, reflect.DeepEqual(graph, before))
		}
		reached := map[uint8]bool{0: true}
		queue := []uint8{0}
		for len(queue) != 0 {
			stateBits := queue[0]
			queue = queue[1:]
			state := []bool{stateBits&1 != 0, stateBits&2 != 0, stateBits&4 != 0}
			for inputBits := uint8(0); inputBits < 4; inputBits++ {
				_, next, stepErr := graph.Step([]bool{inputBits&1 != 0, inputBits&2 != 0}, state)
				if stepErr != nil {
					t.Fatal(stepErr)
				}
				var encoded uint8
				for index, bit := range next {
					if bit {
						encoded |= 1 << uint(index)
					}
				}
				if !reached[encoded] {
					reached[encoded] = true
					queue = append(queue, encoded)
				}
			}
		}
		for index, value := range result.Registers {
			if value == AigTernaryUnknown {
				continue
			}
			want := value == AigTernaryOne
			for state := range reached {
				if (state&(1<<uint(index)) != 0) != want {
					t.Fatal(index, value, reached, result)
				}
			}
		}
	})
}
