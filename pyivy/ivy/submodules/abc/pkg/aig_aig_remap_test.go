package abc

// Tests for CI substitution from aig/aig/aigScl.c:Aig_ManRemap.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigRemapOracleRow struct {
	Seed, Registers, Constraints, Check int
	Cis, Cos                            []AigID
	Objects                             []aigOperationObject
}

func aigRemapGraph(t testing.TB, seed int) (*AigManager, []AigLit) {
	t.Helper()
	graph := NewAigManager()
	variables := aigObjectVariables(t, graph, 5)
	p0, p1, p2, s0, s1 := variables[0], variables[1], variables[2], variables[3], variables[4]
	n0 := aigObjectAnd(t, graph, p0.NotIf(seed&1 != 0), p1.NotIf(seed&2 != 0))
	n1 := aigObjectAnd(t, graph, n0.NotIf(seed&4 != 0), s0.NotIf(seed&8 != 0))
	n2 := aigObjectAnd(t, graph, p2.NotIf(seed&16 != 0), s1.NotIf(seed&1 != 0))
	for _, driver := range []AigLit{n1.NotIf(seed&2 != 0), n2.NotIf(seed&4 != 0), p2.NotIf(seed&8 != 0), n2, n0} {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	replacements := []AigLit{p0, p1, p2, s0, s1}
	if seed&1 != 0 {
		replacements[1] = p0.Not()
	}
	if seed&2 != 0 {
		replacements[2] = AigConst0
	}
	switch seed >> 2 & 3 {
	case 1:
		replacements[3] = AigConst0
	case 2:
		replacements[3] = s1
	case 3:
		replacements[3] = s1.Not()
	}
	if seed&16 != 0 {
		replacements[4] = AigConst1
	}
	return graph, replacements
}

func TestAigRemapNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_remap.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigRemapOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		graph, replacements := aigRemapGraph(t, row.Seed)
		before := graph.Clone()
		result, err := graph.RemapInputs(replacements, AigRewriteOptions{})
		if err != nil || (result.Check() == nil) != (row.Check != 0) || result.RegisterCount() != row.Registers || result.ConstraintCount() != row.Constraints || !slices.Equal(result.Inputs(), row.Cis) || !slices.Equal(result.Outputs(), row.Cos) || !slices.Equal(aigOperationObjects(result), row.Objects) || !reflect.DeepEqual(graph, before) {
			t.Fatal(row.Seed, err, result.RegisterCount(), row.Registers, result.ConstraintCount(), row.Constraints, result.Inputs(), row.Cis, result.Outputs(), row.Cos, aigOperationObjects(result), row.Objects, reflect.DeepEqual(graph, before))
		}
	}
}

func TestAigRemapErrorsBuffersAndDeep(t *testing.T) {
	graph, replacements := aigRemapGraph(t, 0)
	internal := AigLit(6) << 1
	invalid := slices.Clone(replacements)
	invalid[0] = AigLit(graph.ObjectCount()) << 1
	nonCI := slices.Clone(replacements)
	nonCI[0] = internal
	choice := graph.Clone()
	if err := choice.SetChoiceChains([][]AigID{{6, 7}}, AigTraversalOptions{}); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		graph        *AigManager
		replacements []AigLit
		options      AigRewriteOptions
		want         error
	}{
		{nil, replacements, AigRewriteOptions{}, ErrDimension},
		{graph, replacements[:4], AigRewriteOptions{}, ErrDimension},
		{graph, invalid, AigRewriteOptions{}, ErrInvalidLiteral},
		{graph, nonCI, AigRewriteOptions{}, ErrDimension},
		{choice, replacements, AigRewriteOptions{}, ErrUnsupported},
		{graph, replacements, AigRewriteOptions{MaxObjects: 1}, ErrLimit},
		{graph, replacements, AigRewriteOptions{MaxVisits: 1}, ErrLimit},
		{graph, replacements, AigRewriteOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := test.graph.RemapInputs(test.replacements, test.options); result != nil || !errors.Is(err, test.want) {
			t.Fatal(result, err, test.want)
		}
	}
	buffered := NewAigManager()
	input := aigObjectVariables(t, buffered, 1)[0]
	buffer, err := buffered.CreateBuffer(input.Not())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = buffered.CreateCO(buffer); err != nil {
		t.Fatal(err)
	}
	remapped, err := buffered.RemapInputs([]AigLit{input}, AigRewriteOptions{})
	if err != nil || remapped.ObjectCount() != 3 {
		t.Fatal(remapped.ObjectCount(), err)
	}
	outputs, err := remapped.Evaluate([]bool{true})
	if err != nil || !slices.Equal(outputs, []bool{false}) {
		t.Fatal(outputs, err)
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
	deepResult, err := deep.RemapInputs(variables, AigRewriteOptions{MaxObjects: deep.ObjectCount(), MaxVisits: 30_000})
	if err != nil || deepResult.Check() != nil {
		t.Fatal(err, deepResult.Check())
	}
}

func FuzzAigRemapInputs(f *testing.F) {
	for seed := 0; seed < 32; seed++ {
		f.Add(uint8(seed), uint8(seed*7), uint8(seed*13))
	}
	f.Fuzz(func(t *testing.T, structure, choices, phases uint8) {
		graph := NewAigManager()
		variables := aigObjectVariables(t, graph, 5)
		n0 := aigObjectAnd(t, graph, variables[0].NotIf(structure&1 != 0), variables[1].NotIf(structure&2 != 0))
		n1 := aigObjectAnd(t, graph, variables[2].NotIf(structure&4 != 0), variables[3].NotIf(structure&8 != 0))
		n2 := aigObjectAnd(t, graph, n0.NotIf(structure&16 != 0), n1.NotIf(structure&32 != 0))
		for _, driver := range []AigLit{n2, aigObjectAnd(t, graph, n2, variables[4].NotIf(structure&64 != 0))} {
			if _, err := graph.CreateCO(driver); err != nil {
				t.Fatal(err)
			}
		}
		replacements := make([]AigLit, len(variables))
		candidates := append([]AigLit{AigConst1, AigConst0}, variables...)
		for index := range replacements {
			selector := uint8(int(choices) + index*int(structure|1))
			replacements[index] = candidates[int(selector)%len(candidates)].NotIf(phases&(1<<uint(index)) != 0)
		}
		before := graph.Clone()
		result, err := graph.RemapInputs(replacements, AigRewriteOptions{})
		if err != nil || !reflect.DeepEqual(graph, before) {
			t.Fatal(err, reflect.DeepEqual(graph, before))
		}
		ciOrdinal := make([]int, graph.ObjectCount())
		for index := range ciOrdinal {
			ciOrdinal[index] = -1
		}
		for index, id := range graph.cis {
			ciOrdinal[id] = index
		}
		for assignment := 0; assignment < 32; assignment++ {
			targetInputs := make([]bool, len(variables))
			for index := range targetInputs {
				targetInputs[index] = assignment&(1<<uint(index)) != 0
			}
			sourceInputs := make([]bool, len(variables))
			for index, replacement := range replacements {
				value := true
				if replacement.ID() != 0 {
					value = targetInputs[ciOrdinal[replacement.ID()]]
				}
				sourceInputs[index] = value != replacement.Complemented()
			}
			want, sourceErr := graph.Evaluate(sourceInputs)
			got, resultErr := result.Evaluate(targetInputs)
			if sourceErr != nil || resultErr != nil || !slices.Equal(got, want) {
				t.Fatal(assignment, got, want, sourceErr, resultErr, replacements)
			}
		}
	})
}
