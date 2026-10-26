package abc

// Tests for constant-register reduction from aig/aig/aigTsim.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigConstantReduceOracleRow struct {
	Seed, Registers, Constraints, Check int
	Cis, Cos                            []AigID
	Objects                             [][7]int
}

func aigConstantReduceGraph(t testing.TB, seed int) *AigManager {
	t.Helper()
	graph := NewAigManager()
	variables := aigObjectVariables(t, graph, 7)
	p0, p1 := variables[0], variables[1]
	s0, s1, s2, s3, s4 := variables[2], variables[3], variables[4], variables[5], variables[6]
	_ = s1
	n0 := aigObjectAnd(t, graph, s0.NotIf(seed&8 != 0), s2.NotIf(seed&16 != 0))
	n1 := aigObjectAnd(t, graph, s3.NotIf(seed&16 != 0), s4.NotIf(seed&8 != 0))
	d0 := AigConst0.NotIf(seed&1 != 0)
	d1 := s0.NotIf(seed&2 != 0)
	d2 := s2.NotIf(seed&4 == 0)
	d3 := aigObjectAnd(t, graph, p0.NotIf(seed&8 != 0), s0)
	d4 := aigObjectAnd(t, graph, p1.NotIf(seed&16 != 0), s2)
	for _, driver := range []AigLit{n0, n1, s2.NotIf(seed&1 != 0), s4.NotIf(seed&2 != 0), d0, d1, d2, d3, d4} {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.SetRegisterCount(5); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	return graph
}

func aigLiveObjectRecords(graph *AigManager) [][7]int {
	records := make([][7]int, 0, len(graph.objects))
	for id, object := range graph.objects {
		if object.Type == AigNone {
			continue
		}
		fanin0, fanin1 := -1, -1
		if object.Fanin0.Valid() {
			fanin0 = int(object.Fanin0)
		}
		if object.Fanin1.Valid() {
			fanin1 = int(object.Fanin1)
		}
		phase := 0
		if object.Phase {
			phase = 1
		}
		records = append(records, [7]int{id, int(object.Type), fanin0, fanin1, int(object.RefCount), int(object.Level), phase})
	}
	return records
}

func TestAigConstantReduceNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_constant_reduce.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigConstantReduceOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		graph := aigConstantReduceGraph(t, row.Seed)
		before := graph.Clone()
		result, err := graph.ReduceTernaryConstants(AigConstantReduceOptions{})
		if err != nil || result.Graph.Check() != nil || result.Graph.RegisterCount() != row.Registers || result.Graph.ConstraintCount() != row.Constraints || !slices.Equal(result.Graph.Inputs(), row.Cis) || !slices.Equal(result.Graph.Outputs(), row.Cos) || !slices.Equal(aigLiveObjectRecords(result.Graph), row.Objects) || !reflect.DeepEqual(graph, before) {
			t.Fatal(row.Seed, err, row.Check, result.Graph.RegisterCount(), row.Registers, result.Graph.ConstraintCount(), row.Constraints, result.Graph.Inputs(), row.Cis, result.Graph.Outputs(), row.Cos, aigLiveObjectRecords(result.Graph), row.Objects, reflect.DeepEqual(graph, before))
		}
		if result.RemovedRegisters != 5-row.Registers || len(result.Registers) != row.Registers {
			t.Fatal(row.Seed, result)
		}
	}
}

func TestAigConstantReduceErrorsOwnershipAndDeep(t *testing.T) {
	graph := aigConstantReduceGraph(t, 0)
	choice := graph.Clone()
	choice.equivs = make([]AigID, len(choice.objects))
	for index := range choice.equivs {
		choice.equivs[index] = InvalidAigID
	}
	for _, test := range []struct {
		graph   *AigManager
		options AigConstantReduceOptions
		want    error
	}{
		{nil, AigConstantReduceOptions{}, ErrDimension},
		{graph, AigConstantReduceOptions{MaxPasses: -1}, ErrDimension},
		{graph, AigConstantReduceOptions{Multivalued: true}, ErrUnsupported},
		{choice, AigConstantReduceOptions{}, ErrUnsupported},
		{graph, AigConstantReduceOptions{Ternary: AigTernaryStateOptions{MaxRounds: 1}}, ErrLimit},
		{graph, AigConstantReduceOptions{Remap: AigRewriteOptions{MaxObjects: 1}}, ErrLimit},
		{graph, AigConstantReduceOptions{Cleanup: AigSequentialCleanupOptions{Stop: func() bool { return true }}}, ErrCancelled},
	} {
		if result, err := test.graph.ReduceTernaryConstants(test.options); result.Graph != nil || !errors.Is(err, test.want) {
			t.Fatal(result, err, test.want)
		}
	}
	combinational := NewAigManager()
	input := aigObjectVariables(t, combinational, 1)[0]
	if _, err := combinational.CreateCO(input); err != nil {
		t.Fatal(err)
	}
	result, err := combinational.ReduceTernaryConstants(AigConstantReduceOptions{})
	if err != nil || result.Graph == combinational || result.Passes != 0 || result.RemovedRegisters != 0 || result.Registers == nil || len(result.Registers) != 0 {
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
	result, err = deep.ReduceTernaryConstants(AigConstantReduceOptions{
		Ternary: AigTernaryStateOptions{MaxObjects: deep.ObjectCount()},
		Remap:   AigRewriteOptions{MaxObjects: deep.ObjectCount(), MaxVisits: 30_000},
		Cleanup: AigSequentialCleanupOptions{MaxObjects: deep.ObjectCount()},
	})
	if err != nil || result.Graph.RegisterCount() != 0 || result.RemovedRegisters != 1 {
		t.Fatal(result, err)
	}
}

func FuzzAigConstantReduce(f *testing.F) {
	for seed := 0; seed < 32; seed++ {
		f.Add(uint8(seed), uint64(seed)*0x9e3779b97f4a7c15)
	}
	f.Fuzz(func(t *testing.T, seed uint8, inputs uint64) {
		graph := aigConstantReduceGraph(t, int(seed&31))
		before := graph.Clone()
		result, err := graph.ReduceTernaryConstants(AigConstantReduceOptions{})
		if err != nil || !reflect.DeepEqual(graph, before) {
			t.Fatal(result, err, reflect.DeepEqual(graph, before))
		}
		originalState := make([]bool, graph.RegisterCount())
		reducedState := make([]bool, result.Graph.RegisterCount())
		for frame := 0; frame < 32; frame++ {
			primary := []bool{inputs&(1<<uint((2*frame)%64)) != 0, inputs&(1<<uint((2*frame+1)%64)) != 0}
			originalOutputs, nextOriginal, originalErr := graph.Step(primary, originalState)
			reducedOutputs, nextReduced, reducedErr := result.Graph.Step(primary, reducedState)
			if originalErr != nil || reducedErr != nil || !slices.Equal(originalOutputs, reducedOutputs) {
				t.Fatal(frame, originalOutputs, reducedOutputs, originalErr, reducedErr, result)
			}
			originalState, reducedState = nextOriginal, nextReduced
		}
	})
}
