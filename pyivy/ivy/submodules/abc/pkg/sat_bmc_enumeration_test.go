package abc

// Stand-alone tests for assignment justification and enumeration from bmcEnum.c.

import (
	"errors"
	"reflect"
	"testing"
)

func giaEnumerationGraph(t testing.TB) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	graph.Name, graph.Spec = "enumeration", "source"
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	c, _ := graph.AppendCI()
	first, err := graph.AppendAnd(a, b.Not())
	if err != nil {
		t.Fatal(err)
	}
	second, err := graph.AppendAnd(first.Not(), c.Not())
	if err != nil {
		t.Fatal(err)
	}
	graph.AppendCO(first)
	graph.AppendCO(second.Not())
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return graph
}

func giaBoolAssignment(count, assignment int) []bool {
	values := make([]bool, count)
	for index := range values {
		values[index] = assignment>>index&1 != 0
	}
	return values
}

func TestGiaAssignmentJustificationExhaustiveImplication(t *testing.T) {
	graph := giaEnumerationGraph(t)
	before := graph.Clone()
	for assignment := 0; assignment < 1<<graph.InputCount(); assignment++ {
		inputs := giaBoolAssignment(graph.InputCount(), assignment)
		baseOutputs, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		justification, err := graph.DeriveAssignmentJustification(inputs, GiaCofactorOptions{})
		if err != nil || justification.Name != graph.Name || justification.Spec != graph.Spec || justification.InputCount() != graph.InputCount() || justification.OutputCount() != graph.OutputCount() || justification.RegisterCount() != graph.RegisterCount() || justification.Check() != nil {
			t.Fatal(assignment, justification, err)
		}
		atBase, err := justification.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for output, value := range atBase {
			if !value {
				t.Fatal("justification false at originating assignment", assignment, output)
			}
		}
		for candidate := 0; candidate < 1<<graph.InputCount(); candidate++ {
			candidateInputs := giaBoolAssignment(graph.InputCount(), candidate)
			original, originalErr := graph.Evaluate(candidateInputs)
			derived, derivedErr := justification.Evaluate(candidateInputs)
			if originalErr != nil || derivedErr != nil {
				t.Fatal(originalErr, derivedErr)
			}
			for output := range derived {
				if derived[output] && original[output] != baseOutputs[output] {
					t.Fatal("invalid implication", assignment, candidate, output, baseOutputs, original, derived)
				}
			}
		}
	}
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("source graph mutated")
	}
	zero, err := graph.DeriveZeroAssignmentJustification(GiaCofactorOptions{})
	want, wantErr := graph.DeriveAssignmentJustification(make([]bool, graph.InputCount()), GiaCofactorOptions{})
	if err != nil || wantErr != nil || !reflect.DeepEqual(zero, want) {
		t.Fatal(err, wantErr)
	}
}

func giaEnumerationXorGraph(t testing.TB) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	left, _ := graph.AppendAnd(a, b.Not())
	right, _ := graph.AppendAnd(a.Not(), b)
	joined, _ := graph.AppendAnd(left.Not(), right.Not())
	graph.AppendCO(joined.Not())
	return graph
}

func giaCheckJustificationEnumeration(t testing.TB, graph *GiaManager, desired bool, result GiaJustificationEnumeration) {
	t.Helper()
	if !result.Complete || result.Status != SatUnsatisfiable || result.Iterations != len(result.Records)+1 || result.Variables == 0 || result.Clauses == 0 || result.Literals == 0 {
		t.Fatal("incomplete enumeration", result)
	}
	covered := make([]bool, 1<<graph.InputCount())
	for recordIndex, record := range result.Records {
		if len(record.Inputs) != graph.InputCount() || record.Justification == nil || record.Justification.Check() != nil || record.Observed[0] != desired {
			t.Fatal(recordIndex, record)
		}
		atOrigin, err := record.Justification.Evaluate(record.Inputs)
		if err != nil || !atOrigin[0] {
			t.Fatal(recordIndex, atOrigin, err)
		}
		for previous := 0; previous < recordIndex; previous++ {
			prior, err := result.Records[previous].Justification.Evaluate(record.Inputs)
			if err != nil || prior[0] {
				t.Fatal("record was not excluded by prior root", recordIndex, previous, prior, err)
			}
		}
		for assignment := range covered {
			inputs := giaBoolAssignment(graph.InputCount(), assignment)
			formula, formulaErr := record.Justification.Evaluate(inputs)
			original, originalErr := graph.Evaluate(inputs)
			if formulaErr != nil || originalErr != nil {
				t.Fatal(formulaErr, originalErr)
			}
			if formula[0] {
				if original[0] != desired {
					t.Fatal("enumerated formula covers wrong phase", desired, recordIndex, assignment)
				}
				covered[assignment] = true
			}
		}
	}
	for assignment := range covered {
		original, _ := graph.Evaluate(giaBoolAssignment(graph.InputCount(), assignment))
		if covered[assignment] != (original[0] == desired) {
			t.Fatal("coverage mismatch", desired, assignment, covered[assignment], original)
		}
	}
}

func TestGiaAssignmentJustificationEnumerationCoverage(t *testing.T) {
	graph := giaEnumerationXorGraph(t)
	before := graph.Clone()
	for _, desired := range []bool{false, true} {
		result, err := graph.EnumerateAssignmentJustifications(GiaJustificationEnumerationOptions{Output: 0, Desired: desired})
		if err != nil {
			t.Fatal(desired, err)
		}
		giaCheckJustificationEnumeration(t, graph, desired, result)
	}
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("source graph mutated")
	}
	limited, err := graph.EnumerateAssignmentJustifications(GiaJustificationEnumerationOptions{Output: 0, Desired: true, MaxAssignments: 1})
	if !errors.Is(err, ErrLimit) || limited.Complete || limited.Status != SatSatisfiable || len(limited.Records) != 1 {
		t.Fatal(limited, err)
	}
}

func TestGiaAssignmentJustificationErrorsLimitsCancellationAndDepth(t *testing.T) {
	graph := giaEnumerationXorGraph(t)
	for index, test := range []struct {
		graph   *GiaManager
		values  []bool
		options GiaCofactorOptions
		want    error
	}{
		{nil, nil, GiaCofactorOptions{}, ErrDimension},
		{graph, []bool{false}, GiaCofactorOptions{}, ErrDimension},
		{graph, []bool{false, false}, GiaCofactorOptions{MaxObjects: -1}, ErrDimension},
		{graph, []bool{false, false}, GiaCofactorOptions{MaxMappingEntries: -1}, ErrDimension},
		{graph, []bool{false, false}, GiaCofactorOptions{MaxObjects: 1}, ErrLimit},
		{graph, []bool{false, false}, GiaCofactorOptions{MaxMappingEntries: 1}, ErrLimit},
		{graph, []bool{false, false}, GiaCofactorOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		result, err := test.graph.DeriveAssignmentJustification(test.values, test.options)
		if result != nil || !errors.Is(err, test.want) {
			t.Fatal(index, result, err, test.want)
		}
	}
	constant := NewGiaManager()
	constant.AppendCI()
	constant.AppendCO(GiaConst0)
	if result, err := constant.DeriveAssignmentJustification([]bool{false}, GiaCofactorOptions{}); result != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}
	buffered := NewGiaManager()
	input, _ := buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(input)
	buffered.AppendCO(buffer)
	if result, err := buffered.DeriveAssignmentJustification([]bool{false}, GiaCofactorOptions{}); result != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	badOptions := []GiaJustificationEnumerationOptions{
		{Output: -1}, {Output: 0, MaxAssignments: -1}, {Output: 0, MaxTotalObjects: -1},
		{Output: 0, MaxObjects: -1}, {Output: 0, MaxMappingEntries: -1},
		{Output: 0, MaxVariables: -1}, {Output: 0, MaxClauses: -1}, {Output: 0, MaxLiterals: -1},
	}
	for index, options := range badOptions {
		if result, err := graph.EnumerateAssignmentJustifications(options); !reflect.DeepEqual(result, GiaJustificationEnumeration{}) || !errors.Is(err, ErrDimension) {
			t.Fatal(index, result, err)
		}
	}
	for index, options := range []GiaJustificationEnumerationOptions{
		{Output: 1},
		{Output: 0, MaxVariables: 1},
		{Output: 0, MaxClauses: 1},
		{Output: 0, MaxLiterals: 1},
		{Output: 0, MaxObjects: 1},
		{Output: 0, MaxMappingEntries: 1},
		{Output: 0, MaxTotalObjects: 1},
		{Output: 0, Stop: func() bool { return true }},
	} {
		result, err := graph.EnumerateAssignmentJustifications(options)
		if !reflect.DeepEqual(result, GiaJustificationEnumeration{}) || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(index, result, err)
		}
	}
	unknown, err := graph.EnumerateAssignmentJustifications(GiaJustificationEnumerationOptions{Output: 0, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || unknown.Status != SatUnknown || unknown.Reason != SatCancelled || unknown.Complete || len(unknown.Records) != 0 {
		t.Fatal(unknown, err)
	}

	deep := NewGiaManager()
	inputs := make([]GiaLit, 10001)
	for index := range inputs {
		inputs[index], _ = deep.AppendCI()
	}
	root := inputs[0]
	for index := 1; index < len(inputs); index++ {
		root, _ = deep.AppendAnd(root, inputs[index].NotIf(index&1 != 0))
	}
	deep.AppendCO(root)
	values := make([]bool, len(inputs))
	for index := range values {
		values[index] = index&1 == 0
	}
	justification, err := deep.DeriveAssignmentJustification(values, GiaCofactorOptions{MaxObjects: 30000, MaxMappingEntries: 30000})
	if err != nil || justification.Check() != nil || justification.AndCount() != 10000 {
		t.Fatal(justification, err)
	}
}

func FuzzGiaAssignmentJustification(f *testing.F) {
	f.Add([]byte{1, 3, 5, 7, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph := NewGiaManager()
		signals := make([]GiaLit, 3)
		for index := range signals {
			signals[index], _ = graph.AppendCI()
		}
		for index := 0; index < len(data) && index < 10; index++ {
			left := int(data[index]) % len(signals)
			right := (left + 1 + int(data[index]>>4)) % len(signals)
			if signals[left].ID() == signals[right].ID() {
				continue
			}
			gate, err := graph.AppendAnd(signals[left].NotIf(data[index]&1 != 0), signals[right].NotIf(data[index]&2 != 0))
			if err == nil {
				signals = append(signals, gate)
			}
		}
		graph.AppendCO(signals[int(data[0])%len(signals)].NotIf(data[0]&4 != 0))
		origin := giaBoolAssignment(3, int(data[len(data)-1]))
		base, err := graph.Evaluate(origin)
		if err != nil {
			t.Fatal(err)
		}
		justification, err := graph.DeriveAssignmentJustification(origin, GiaCofactorOptions{MaxObjects: 1000, MaxMappingEntries: 1000})
		if err != nil {
			t.Fatal(err)
		}
		if atOrigin, evalErr := justification.Evaluate(origin); evalErr != nil || !atOrigin[0] {
			t.Fatal(atOrigin, evalErr)
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := giaBoolAssignment(3, assignment)
			formula, formulaErr := justification.Evaluate(inputs)
			original, originalErr := graph.Evaluate(inputs)
			if formulaErr != nil || originalErr != nil || formula[0] && original[0] != base[0] {
				t.Fatal(assignment, formula, original, base, formulaErr, originalErr)
			}
		}
	})
}
