package abc

// Go-only exhaustive tests for the complete bmcFault.c workflow.

import (
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func faultWorkflowCombinational(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	a, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	gate, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(gate); err != nil {
		t.Fatal(err)
	}
	return g
}

func faultWorkflowSequential(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	pi, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	ro, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	gate, err := g.AppendAnd(pi, ro)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(gate); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(gate); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestPrepareGiaFaultCampaignAllAlgorithms(t *testing.T) {
	combinational := faultWorkflowCombinational(t)
	for _, test := range []struct {
		algorithm, parameters, functional int
		formula                           string
	}{
		{0, 1, 2, "(p?(a&b):~(a&b))"},
		{2, 2, 2, ""},
		{3, 1, 2, ""},
		{4, 4, 2, ""},
	} {
		options := GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: test.algorithm, Formula: test.formula}}
		prepared, err := PrepareGiaFaultCampaign(combinational, combinational.Clone(), options)
		if err != nil || !prepared.Ready || prepared.FunctionalInputs != test.functional || len(prepared.ParameterVariables) != test.parameters || prepared.Miter.PrimaryInputCount() != test.functional+test.parameters {
			t.Fatal(test, prepared, err)
		}
	}
	sequential := faultWorkflowSequential(t)
	prepared, err := PrepareGiaFaultCampaign(sequential, sequential.Clone(), GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 1}})
	if err != nil || !prepared.Ready || prepared.FunctionalInputs != 3 || len(prepared.ParameterVariables) != 1 || prepared.Miter.PrimaryInputCount() != 4 {
		t.Fatal(prepared, err)
	}
}

func TestPrepareGiaFaultCampaignPatternsAndCardinality(t *testing.T) {
	design := faultWorkflowCombinational(t)
	patterns := []bool{false, false, true, true}
	prepared, err := PrepareGiaFaultCampaign(design, design.Clone(), GiaFaultPreparationOptions{
		Parameters: GiaFaultParameters{Algorithm: 2, Basic: true}, Patterns: patterns, WarmUp: true,
	})
	if err != nil || prepared.Ready || prepared.Status != SatUnsatisfiable || len(prepared.Patterns) != 2 {
		t.Fatal(prepared, err)
	}
	if !slices.Equal(patterns, []bool{false, false, true, true}) {
		t.Fatal("caller patterns mutated")
	}
	started, err := PrepareGiaFaultCampaign(design, design.Clone(), GiaFaultPreparationOptions{
		Parameters: GiaFaultParameters{Algorithm: 3, StartPatterns: true},
	})
	if err != nil || started.Status != SatUnsatisfiable || !slices.Equal(started.Patterns, []bool{false, false}) {
		t.Fatal(started, err)
	}
}

func TestFindGiaFaultNewAssignments(t *testing.T) {
	design := faultWorkflowCombinational(t)
	prepared, err := PrepareGiaFaultCampaign(design, design.Clone(), GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3}})
	if err != nil {
		t.Fatal(err)
	}
	newest := []bool{true, true}
	result, err := FindGiaFaultNewAssignments(prepared.Miter, 2, nil, newest, GiaFaultNewOptions{})
	if err != nil || result.Status != SatUnsatisfiable || !reflect.DeepEqual(result.Assignments, [][]int{{0}}) {
		t.Fatal(result, err)
	}
	covered, err := FindGiaFaultNewAssignments(prepared.Miter, 2, newest, newest, GiaFaultNewOptions{})
	if err != nil || covered.Status != SatUnsatisfiable || len(covered.Assignments) != 0 {
		t.Fatal(covered, err)
	}
}

func TestRunGiaFaultCampaign(t *testing.T) {
	design := faultWorkflowCombinational(t)
	before := design.Clone()
	result, err := RunGiaFaultCampaign(design, design, GiaFaultCampaignOptions{
		GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{
			Algorithm: 3, CheckUntestable: true, DumpUntestable: true, Dump: true,
		}},
	})
	if err != nil || result.Status != SatUnsatisfiable || result.Generated != 1 || len(result.Patterns) != 2 || len(result.PatternData) != 3 || result.UntestableStatus != SatUnsatisfiable || len(result.Untestable) != 0 {
		t.Fatal(result, err)
	}
	if !reflect.DeepEqual(before, design) {
		t.Fatal("campaign mutated source graph")
	}
	newFaults, err := RunGiaFaultCampaign(design, design.Clone(), GiaFaultCampaignOptions{
		GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3, DumpNewFaults: true}},
	})
	if err != nil || newFaults.Status != SatSatisfiable || !reflect.DeepEqual(newFaults.NewFaults, [][]int{{0}}) || len(newFaults.Patterns) != 0 {
		t.Fatal(newFaults, err)
	}
}

func TestRunGiaFaultCampaignAnalysisAndIterationLimit(t *testing.T) {
	design := faultWorkflowCombinational(t)
	result, err := RunGiaFaultCampaign(design, design, GiaFaultCampaignOptions{
		GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 2, CheckIterations: 1}},
		MaxIterations:              8,
	})
	if err != nil || result.Status != SatUnsatisfiable || result.Generated != 2 || len(result.Patterns) != 4 || !slices.Equal(result.Active, []bool{true, false}) {
		t.Fatal(result, err)
	}
	limited, err := RunGiaFaultCampaign(design, design, GiaFaultCampaignOptions{
		GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 2}},
		MaxIterations:              1,
	})
	if err != nil || limited.Status != SatUnknown || !limited.IterationLimitReached || limited.Generated != 1 || len(limited.Patterns) != 2 {
		t.Fatal(limited, err)
	}
}

func TestRunGiaFaultCampaignPatternFile(t *testing.T) {
	design := faultWorkflowCombinational(t)
	path := t.TempDir() + "/patterns.txt"
	if err := os.WriteFile(path, []byte("00\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := RunGiaFaultCampaign(design, design, GiaFaultCampaignOptions{GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3, FileName: path}}})
	if err != nil || result.Status != SatUnsatisfiable || result.Generated != 0 || !slices.Equal(result.Patterns, []bool{false, false}) {
		t.Fatal(result, err)
	}
	_, err = RunGiaFaultCampaign(design, design, GiaFaultCampaignOptions{GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3, FileName: path}, Patterns: []bool{false, false}}})
	if !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestGiaFaultWorkflowErrorsAndCancellation(t *testing.T) {
	design := faultWorkflowCombinational(t)
	for _, test := range []struct {
		options GiaFaultPreparationOptions
		want    error
	}{
		{GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 5}}, ErrDimension},
		{GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3}, Active: []bool{true, false}}, ErrDimension},
		{GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3}, Patterns: []bool{true}}, ErrDimension},
		{GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3}, CNF: GiaFaultCNFOptions{MaxVariables: 1}}, ErrLimit},
		{GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3}, Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := PrepareGiaFaultCampaign(design, design, test.options); result != nil || !errors.Is(err, test.want) {
			t.Fatal(test, result, err)
		}
	}
	if result, err := FindGiaFaultNewAssignments(design, 2, nil, []bool{false}, GiaFaultNewOptions{}); !errors.Is(err, ErrDimension) || result.Assignments != nil {
		t.Fatal(result, err)
	}
	if result, err := RunGiaFaultCampaign(design, design, GiaFaultCampaignOptions{GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 3}, Stop: func() bool { return true }}, MaxIterations: 1}); !errors.Is(err, ErrCancelled) || result.Patterns != nil {
		t.Fatal(result, err)
	}
}

func FuzzGiaFaultCampaignCoverage(f *testing.F) {
	for _, seed := range []byte{0, 1, 2, 3} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, flags byte) {
		design := faultWorkflowCombinational(t)
		basic := flags&1 != 0
		result, err := RunGiaFaultCampaign(design, design, GiaFaultCampaignOptions{
			GiaFaultPreparationOptions: GiaFaultPreparationOptions{Parameters: GiaFaultParameters{Algorithm: 2, Basic: basic}},
			MaxIterations:              8,
		})
		if err != nil || result.Status != SatUnsatisfiable || len(result.Patterns)%2 != 0 {
			t.Fatal(result, err)
		}
		for stuckZero := 0; stuckZero < 2; stuckZero++ {
			for stuckOne := 0; stuckOne < 2; stuckOne++ {
				if stuckZero+stuckOne == 0 || basic && stuckZero+stuckOne > 1 {
					continue
				}
				detectable, covered := false, false
				for assignment := 0; assignment < 4; assignment++ {
					a, b := assignment&1 != 0, assignment&2 != 0
					nominal := a && b
					faulty := nominal && stuckZero == 0 || stuckOne != 0
					if nominal != faulty {
						detectable = true
					}
				}
				for row := 0; row < len(result.Patterns)/2; row++ {
					a, b := result.Patterns[2*row], result.Patterns[2*row+1]
					nominal := a && b
					faulty := nominal && stuckZero == 0 || stuckOne != 0
					covered = covered || nominal != faulty
				}
				if detectable != covered {
					t.Fatal(stuckZero, stuckOne, basic, result.Patterns, detectable, covered)
				}
			}
		}
	})
}
