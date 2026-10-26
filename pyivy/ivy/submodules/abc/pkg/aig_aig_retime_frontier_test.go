package abc

// Tests for forward-frontier retiming from aig/aig/aigRetF.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigRetimeOracleRow struct {
	Seed, Maximum, Registers, Check int
	Cis, Cos                        []AigID
	Objects                         []aigOperationObject
}

func aigRetimeGraph(t testing.TB, seed int) *AigManager {
	t.Helper()
	graph := NewAigManager()
	p := aigObjectVariables(t, graph, 2)
	state := aigObjectVariables(t, graph, 3)
	n0 := aigObjectAnd(t, graph, state[0].NotIf(seed&1 != 0), state[1].NotIf(seed&2 != 0))
	n1 := aigObjectAnd(t, graph, n0.NotIf(seed&4 != 0), state[2].NotIf(seed&8 != 0))
	third := aigObjectAnd(t, graph, p[0].NotIf(seed&1 != 0), p[1].NotIf(seed&2 != 0))
	for _, driver := range []AigLit{
		n0.NotIf(seed&2 != 0), n1.NotIf(seed&4 != 0), p[0],
		func() AigLit {
			if seed&16 != 0 {
				return AigConst1
			}
			return p[1]
		}(), third,
	} {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.SetRegisterCount(3); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestAigRetimeFrontierCorrectedNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_retime.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigRetimeOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		source := aigRetimeGraph(t, row.Seed)
		before := source.Clone()
		result, err := source.RetimeFrontier(AigRetimeOptions{MaxSteps: row.Maximum})
		if err != nil || result.Steps != min(row.Maximum, 2) || (result.Graph.Check() == nil) != (row.Check != 0) || result.Graph.RegisterCount() != row.Registers || !slices.Equal(result.Graph.Inputs(), row.Cis) || !slices.Equal(result.Graph.Outputs(), row.Cos) || !slices.Equal(aigOperationObjects(result.Graph), row.Objects) {
			t.Fatal(row.Seed, err, result.Steps, row.Maximum, result.Graph.RegisterCount(), row.Registers, result.Graph.Inputs(), row.Cis, result.Graph.Outputs(), row.Cos, aigOperationObjects(result.Graph), row.Objects, result.Graph.Check())
		}
		if !reflect.DeepEqual(source, before) {
			t.Fatal(row.Seed, "source mutated")
		}
	}
}

func aigTraceLanguage(t testing.TB, graph *AigManager, frames int) map[uint64]map[uint64]bool {
	t.Helper()
	pis, regs, pos := graph.PrimaryInputCount(), graph.RegisterCount(), graph.PrimaryOutputCount()
	result := make(map[uint64]map[uint64]bool)
	totalFrames := frames + 1 // one transition establishes the retimed state relation
	inputSequences := 1 << (pis * totalFrames)
	states := 1 << regs
	for inputsCode := 0; inputsCode < inputSequences; inputsCode++ {
		outputs := make(map[uint64]bool)
		for stateCode := 0; stateCode < states; stateCode++ {
			state := make([]bool, regs)
			for index := range state {
				state[index] = stateCode&(1<<index) != 0
			}
			var outputCode uint64
			for frame := 0; frame < totalFrames; frame++ {
				primary := make([]bool, pis)
				for index := range primary {
					primary[index] = inputsCode&(1<<(frame*pis+index)) != 0
				}
				output, next, err := graph.Step(primary, state)
				if err != nil {
					t.Fatal(err)
				}
				if frame != 0 {
					for index := 0; index < pos; index++ {
						if output[index] {
							outputCode |= 1 << uint((frame-1)*pos+index)
						}
					}
				}
				state = next
			}
			outputs[outputCode] = true
		}
		result[uint64(inputsCode)] = outputs
	}
	return result
}

func TestAigRetimeFrontierErrorsAndTraceLanguage(t *testing.T) {
	graph := aigRetimeGraph(t, 7)
	for _, test := range []struct {
		graph   *AigManager
		options AigRetimeOptions
		want    error
	}{
		{nil, AigRetimeOptions{}, ErrDimension},
		{graph, AigRetimeOptions{MaxSteps: -1}, ErrDimension},
		{graph, AigRetimeOptions{MaxObjects: -1}, ErrDimension},
		{graph, AigRetimeOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{graph, AigRetimeOptions{MaxObjects: graph.ObjectCount(), MaxSteps: 1}, ErrLimit},
		{graph, AigRetimeOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		before := (*AigManager)(nil)
		if test.graph != nil {
			before = test.graph.Clone()
		}
		result, err := test.graph.RetimeFrontier(test.options)
		if result.Graph != nil || !errors.Is(err, test.want) || (test.graph != nil && !reflect.DeepEqual(test.graph, before)) {
			t.Fatal(result, err, test.want)
		}
	}
	combinational := NewAigManager()
	input := aigObjectVariables(t, combinational, 1)[0]
	combinational.CreateCO(input)
	if result, err := combinational.RetimeFrontier(AigRetimeOptions{}); result.Graph != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	for seed := 0; seed < 32; seed++ {
		source := aigRetimeGraph(t, seed)
		result, err := source.RetimeFrontier(AigRetimeOptions{MaxSteps: 2})
		if err != nil {
			t.Fatal(seed, err)
		}
		before := aigTraceLanguage(t, source, 2)
		after := aigTraceLanguage(t, result.Graph, 2)
		if !reflect.DeepEqual(before, after) {
			t.Fatal(seed, before, after)
		}
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
	deepResult, err := deep.RetimeFrontier(AigRetimeOptions{MaxObjects: deep.ObjectCount()})
	if err != nil || deepResult.Steps != 0 || deepResult.Graph.AndCount() != depth || deepResult.Graph.Check() != nil {
		t.Fatal(err, deepResult.Steps, deepResult.Graph.AndCount(), deepResult.Graph.Check())
	}
}

func FuzzAigRetimeFrontier(f *testing.F) {
	for seed := uint8(0); seed < 32; seed++ {
		f.Add(seed, uint8(1))
	}
	f.Fuzz(func(t *testing.T, seed, steps uint8) {
		seed &= 31
		maximum := 1 + int(steps%2)
		source := aigRetimeGraph(t, int(seed))
		result, err := source.RetimeFrontier(AigRetimeOptions{MaxSteps: maximum})
		if err != nil || result.Steps != maximum || result.Graph.Check() != nil {
			t.Fatal(seed, maximum, result.Steps, err, result.Graph.Check())
		}
		if !reflect.DeepEqual(aigTraceLanguage(t, source, 2), aigTraceLanguage(t, result.Graph, 2)) {
			t.Fatal(seed, maximum, "trace language")
		}
	})
}
