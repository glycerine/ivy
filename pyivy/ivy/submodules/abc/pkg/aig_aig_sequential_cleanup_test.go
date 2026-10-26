package abc

// Tests for sequential cleanup from aig/aig/aigScl.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigSequentialCleanupSnapshot struct {
	Removed, Registers, Check int
	Cis, Cos                  []AigID
	Objects                   []aigOperationObject
}

type aigSequentialCleanupRow struct {
	Seed        int
	Full, Basic aigSequentialCleanupSnapshot
}

func aigSequentialCleanupGraph(t testing.TB, seed int) *AigManager {
	t.Helper()
	graph := NewAigManager()
	p := aigObjectVariables(t, graph, 2)
	state := aigObjectVariables(t, graph, 4)
	root := (seed >> 3) & 3
	node := aigObjectAnd(t, graph, p[0].NotIf(seed&1 != 0), state[root].NotIf(seed&2 != 0))
	if _, err := graph.CreateCO(node); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 4; index++ {
		driver := p[0]
		if seed&(1<<index) != 0 {
			driver = state[(index+1)&3]
		}
		leftPhase := seed&(1<<uint((index+1)&4)) != 0
		rightPhase := seed&(1<<uint((index+2)&4)) != 0
		node = aigObjectAnd(t, graph, p[1].NotIf(leftPhase), driver.NotIf(rightPhase))
		if _, err := graph.CreateCO(node); err != nil {
			t.Fatal(err)
		}
	}
	_ = aigObjectAnd(t, graph, p[0].Not(), p[1])
	if err := graph.SetRegisterCount(4); err != nil {
		t.Fatal(err)
	}
	return graph
}

func aigSequentialRegisterIndices(cis []AigID) []int {
	result := make([]int, 0, len(cis)-2)
	for _, id := range cis[2:] {
		result = append(result, int(id)-3)
	}
	return result
}

func TestAigSequentialCleanupNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_sequential_cleanup.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigSequentialCleanupRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		source := aigSequentialCleanupGraph(t, row.Seed)
		before := source.Clone()
		full, err := source.SequentialCleanup(AigSequentialCleanupOptions{})
		if err != nil || full.Graph.Check() != nil || full.RemovedObjects != row.Full.Removed || full.Graph.RegisterCount() != row.Full.Registers || !slices.Equal(full.Graph.Inputs(), row.Full.Cis) || !slices.Equal(full.Graph.Outputs(), row.Full.Cos) || !slices.Equal(aigOperationObjects(full.Graph), row.Full.Objects) || !slices.Equal(full.Registers, aigSequentialRegisterIndices(row.Full.Cis)) {
			t.Fatal("full", row.Seed, err, full.RemovedObjects, row.Full.Removed, full.Graph.RegisterCount(), row.Full.Registers, full.Graph.Inputs(), row.Full.Cis, full.Graph.Outputs(), row.Full.Cos, aigOperationObjects(full.Graph), row.Full.Objects, full.Registers, full.Graph.Check())
		}
		basic, err := source.SequentialCleanupBasic(AigSequentialCleanupOptions{})
		if err != nil || basic.Graph.Check() != nil || basic.RemovedObjects != row.Basic.Removed || basic.Graph.RegisterCount() != row.Basic.Registers || !slices.Equal(basic.Graph.Inputs(), row.Basic.Cis) || !slices.Equal(basic.Graph.Outputs(), row.Basic.Cos) || !slices.Equal(aigOperationObjects(basic.Graph), row.Basic.Objects) {
			t.Fatal("basic", row.Seed, err, basic.RemovedObjects, row.Basic.Removed, basic.Graph.RegisterCount(), row.Basic.Registers, basic.Graph.Inputs(), row.Basic.Cis, basic.Graph.Outputs(), row.Basic.Cos, aigOperationObjects(basic.Graph), row.Basic.Objects, basic.Graph.Check())
		}
		if !reflect.DeepEqual(source, before) {
			t.Fatal(row.Seed, "source mutated")
		}
	}
}

func TestAigSequentialCleanupErrorsOwnershipAndDeep(t *testing.T) {
	graph := aigSequentialCleanupGraph(t, 9)
	for _, test := range []struct {
		graph   *AigManager
		options AigSequentialCleanupOptions
		want    error
	}{
		{nil, AigSequentialCleanupOptions{}, ErrDimension},
		{graph, AigSequentialCleanupOptions{MaxObjects: -1}, ErrDimension},
		{graph, AigSequentialCleanupOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{graph, AigSequentialCleanupOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := test.graph.SequentialCleanup(test.options); result.Graph != nil || !errors.Is(err, test.want) {
			t.Fatal(result, err, test.want)
		}
	}
	choice := NewAigManager()
	v := aigObjectVariables(t, choice, 5)
	a := aigObjectAnd(t, choice, v[0], v[1])
	b := aigObjectAnd(t, choice, v[2], v[3])
	choice.CreateCO(a)
	choice.CreateCO(v[4])
	choice.SetRegisterCount(1)
	if err := choice.SetChoiceChains([][]AigID{{a.ID(), b.ID()}}, AigTraversalOptions{}); err != nil {
		t.Fatal(err)
	}
	if result, err := choice.SequentialCleanup(AigSequentialCleanupOptions{}); result.Graph != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	full, err := graph.SequentialCleanup(AigSequentialCleanupOptions{})
	if err != nil {
		t.Fatal(err)
	}
	full.Registers[0] = 99
	again, err := graph.SequentialCleanup(AigSequentialCleanupOptions{})
	if err != nil || len(again.Registers) == 0 || again.Registers[0] == 99 || again.Graph == full.Graph {
		t.Fatal(err, again.Registers)
	}

	const depth = 10000
	deep := NewAigManager()
	primary := aigObjectVariables(t, deep, 1)[0]
	state := aigObjectVariables(t, deep, 1)[0]
	root := state
	for index := 0; index < depth; index++ {
		root = aigObjectAnd(t, deep, root, primary.NotIf(index&1 != 0))
	}
	deep.CreateCO(root)
	deep.CreateCO(primary)
	deep.SetRegisterCount(1)
	cleaned, err := deep.SequentialCleanup(AigSequentialCleanupOptions{MaxObjects: deep.ObjectCount()})
	if err != nil || cleaned.Graph.RegisterCount() != 1 || cleaned.Graph.AndCount() != depth || cleaned.Graph.Check() != nil {
		t.Fatal(err, cleaned.Graph.RegisterCount(), cleaned.Graph.AndCount(), cleaned.Graph.Check())
	}
}

func FuzzAigSequentialCleanup(f *testing.F) {
	f.Add(uint8(5), uint8(2), uint8(13))
	f.Fuzz(func(t *testing.T, links, selected, assignment uint8) {
		graph := NewAigManager()
		primary := aigObjectVariables(t, graph, 1)[0]
		state := aigObjectVariables(t, graph, 3)
		root := int(selected % 3)
		graph.CreateCO(state[root])
		for index := range state {
			driver := primary
			if links&(1<<index) != 0 {
				driver = state[(index+1)%3].Not()
			}
			graph.CreateCO(driver)
		}
		graph.SetRegisterCount(3)
		result, err := graph.SequentialCleanup(AigSequentialCleanupOptions{})
		if err != nil {
			t.Fatal(err)
		}
		used := uint8(1 << root)
		for round := 0; round < 3; round++ {
			for index := 0; index < 3; index++ {
				if used&(1<<index) != 0 && links&(1<<index) != 0 {
					used |= 1 << ((index + 1) % 3)
				}
			}
		}
		expected := make([]int, 0, 3)
		for index := 0; index < 3; index++ {
			if used&(1<<index) != 0 {
				expected = append(expected, index)
			}
		}
		if !slices.Equal(result.Registers, expected) || result.Graph.Check() != nil {
			t.Fatal(links, selected, result.Registers, expected, result.Graph.Check())
		}
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		inputs := []bool{values[0]}
		for _, index := range expected {
			inputs = append(inputs, values[1+index])
		}
		got, err := result.Graph.Evaluate(inputs)
		if err != nil || got[0] != values[1+root] {
			t.Fatal(got, err)
		}
		for index, sourceIndex := range expected {
			want := values[0]
			if links&(1<<sourceIndex) != 0 {
				want = !values[1+(sourceIndex+1)%3]
			}
			if got[index+1] != want {
				t.Fatal(got, sourceIndex, want)
			}
		}
	})
}
