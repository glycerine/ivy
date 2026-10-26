package abc

// Standalone semantic, SAT, limit, concurrency and fuzz tests for abcSense.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func abcSensitivityFixture(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	c, _ := graph.CreateCI()
	first, _ := graph.And(a, b)
	second, _ := graph.Xor(b, c)
	constraint, _ := graph.Or(a, c)
	graph.CreateCO(first)
	graph.CreateCO(second)
	graph.CreateCO(constraint)
	graph.SetConstraintCount(1)
	return graph
}

func abcSensitivityModel(source *AigManager, variable int) bool {
	inputs := source.InputCount()
	outputs := source.OutputCount()
	values := make([]bool, inputs)
	for assignment := 0; assignment < 1<<inputs; assignment++ {
		for index := range values {
			values[index] = assignment&(1<<index) != 0
		}
		values[variable] = false
		low, _ := source.Evaluate(values)
		values[variable] = true
		high, _ := source.Evaluate(values)
		changed := false
		for output := 0; output < outputs-1; output++ {
			changed = changed || low[output] != high[output]
		}
		if changed && low[outputs-1] && high[outputs-1] {
			return true
		}
	}
	return false
}

func TestABCSensitivityMiterExactSemanticsAndOwnership(t *testing.T) {
	source := abcSensitivityFixture(t)
	before := slices.Clone(source.objects)
	for variable := 0; variable < source.InputCount(); variable++ {
		miter, err := ABCSensitivityMiter(source, variable, ABCSensitivityOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if miter.Name != source.Name || miter.Spec != source.Spec || miter.InputCount() != source.InputCount() || miter.OutputCount() != 1 {
			t.Fatalf("variable=%d miter inputs/outputs=%d/%d", variable, miter.InputCount(), miter.OutputCount())
		}
		inputs := make([]bool, source.InputCount())
		for assignment := 0; assignment < 1<<len(inputs); assignment++ {
			for index := range inputs {
				inputs[index] = assignment&(1<<index) != 0
			}
			inputs[variable] = false
			low, _ := source.Evaluate(inputs)
			inputs[variable] = true
			high, _ := source.Evaluate(inputs)
			want := low[len(low)-1] && high[len(high)-1]
			changed := false
			for output := 0; output < len(low)-1; output++ {
				changed = changed || low[output] != high[output]
			}
			want = want && changed
			inputs[variable] = assignment&(1<<variable) != 0
			got, evalErr := miter.Evaluate(inputs)
			if evalErr != nil || len(got) != 1 || got[0] != want {
				t.Fatalf("variable=%d assignment=%d got=%v want=%v err=%v", variable, assignment, got, want, evalErr)
			}
		}
	}
	if !slices.Equal(before, source.objects) {
		t.Fatal("source graph was mutated")
	}
}

func TestABCSensitivityClassificationModelsConstraintAndUnknown(t *testing.T) {
	source := abcSensitivityFixture(t)
	result, err := ABCSensitivity(source, ABCSensitivityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := make([]int, 0)
	for variable := 0; variable < source.InputCount(); variable++ {
		if abcSensitivityModel(source, variable) {
			want = append(want, variable)
		}
	}
	if !slices.Equal(result.Sensitive, want) || result.SAT != len(want) || result.UNSAT != source.InputCount()-len(want) || result.Unknown != 0 {
		t.Fatalf("result=%+v want=%v", result, want)
	}
	models := make([][]bool, len(result.Sensitive))
	for index, variable := range result.Sensitive {
		entry := result.Inputs[variable]
		if entry.Status != SatSatisfiable || len(entry.Model) != source.InputCount() {
			t.Fatalf("entry=%+v", entry)
		}
		models[index] = slices.Clone(entry.Model)
	}
	if len(result.Sensitive) > 1 {
		first := result.Sensitive[0]
		second := result.Sensitive[1]
		result.Inputs[first].Model[0] = !result.Inputs[first].Model[0]
		if !slices.Equal(result.Inputs[second].Model, models[1]) {
			t.Fatal("SAT models alias each other")
		}
		again, againErr := ABCSensitivity(source, ABCSensitivityOptions{})
		if againErr != nil || !slices.Equal(again.Inputs[first].Model, models[0]) {
			t.Fatal("SAT model mutation leaked into a subsequent call")
		}
	}
	limited, err := ABCSensitivity(source, ABCSensitivityOptions{SAT: SatLimits{Propagations: 1}})
	if err != nil || limited.Unknown == 0 || len(limited.Sensitive) < limited.Unknown {
		t.Fatalf("limited=%+v err=%v", limited, err)
	}

	constrained := NewAigManager()
	a, _ := constrained.CreateCI()
	b, _ := constrained.CreateCI()
	constrained.CreateCO(a)
	constrained.CreateCO(a)
	constrained.SetConstraintCount(1)
	got, err := ABCSensitivity(constrained, ABCSensitivityOptions{})
	if err != nil || len(got.Sensitive) != 0 || got.UNSAT != 2 || b.ID() == 0 {
		t.Fatalf("constraint result=%+v err=%v", got, err)
	}
}

func TestABCSensitivityErrorsLimitsDepthAndConcurrency(t *testing.T) {
	source := abcSensitivityFixture(t)
	for _, row := range []struct {
		variable int
		options  ABCSensitivityOptions
		want     error
	}{
		{-1, ABCSensitivityOptions{}, ErrDimension},
		{source.InputCount(), ABCSensitivityOptions{}, ErrDimension},
		{0, ABCSensitivityOptions{MaxObjects: 1}, ErrLimit},
		{0, ABCSensitivityOptions{MaxWork: 1}, ErrLimit},
		{0, ABCSensitivityOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if got, err := ABCSensitivityMiter(source, row.variable, row.options); got != nil || !errors.Is(err, row.want) {
			t.Fatalf("variable=%d options=%+v got=%v err=%v want=%v", row.variable, row.options, got, err, row.want)
		}
	}
	sequential := source.Clone()
	sequential.SetRegisterCount(1)
	if got, err := ABCSensitivity(sequential, ABCSensitivityOptions{}); got.Inputs != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("sequential got=%+v err=%v", got, err)
	}
	empty := NewAigManager()
	if got, err := ABCSensitivity(empty, ABCSensitivityOptions{}); got.Inputs != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("empty got=%+v err=%v", got, err)
	}
	for _, options := range []ABCSensitivityOptions{
		{MaxInputs: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }},
	} {
		if got, err := ABCSensitivity(source, options); got.Inputs != nil || (err == nil && got.Unknown == 0) {
			t.Fatalf("options=%+v got=%+v err=%v", options, got, err)
		}
	}

	deep := NewAigManager()
	a, _ := deep.CreateCI()
	b, _ := deep.CreateCI()
	root := a
	for index := 0; index < 10000; index++ {
		root, _ = deep.And(root.NotIf(index%3 == 0), b.NotIf(index%5 == 0))
	}
	if deep.ObjectCount() < 10000 {
		t.Fatalf("deep fixture collapsed to %d objects", deep.ObjectCount())
	}
	deep.CreateCO(root)
	deep.CreateCO(AigConst1)
	deepResult, err := ABCSensitivity(deep, ABCSensitivityOptions{MaxObjects: 30000, MaxWork: 100000, MaxVariables: 30000, MaxClauses: 100000, MaxLiterals: 300000})
	if err != nil || len(deepResult.Inputs) != 2 {
		t.Fatalf("deep=%+v err=%v", deepResult, err)
	}

	baseline, _ := ABCSensitivity(source, ABCSensitivityOptions{})
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, callErr := ABCSensitivity(source, ABCSensitivityOptions{})
			if callErr != nil || !slices.Equal(got.Sensitive, baseline.Sensitive) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzABCSensitivity(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6})
	f.Add([]byte{0xff, 0, 17, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph := NewAigManager()
		pool := make([]AigLit, 4)
		for index := range pool {
			pool[index], _ = graph.CreateCI()
		}
		gates := 1 + int(data[0]%12)
		for gate := 0; gate < gates; gate++ {
			left := pool[int(data[(2*gate+1)%len(data)])%len(pool)].NotIf(data[(gate+1)%len(data)]&1 != 0)
			right := pool[int(data[(2*gate+2)%len(data)])%len(pool)].NotIf(data[(gate+2)%len(data)]&2 != 0)
			var root AigLit
			var err error
			if data[gate%len(data)]&4 != 0 {
				root, err = graph.Xor(left, right)
			} else {
				root, err = graph.And(left, right)
			}
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, root)
		}
		regular := 1 + int(data[0]%3)
		for output := 0; output < regular; output++ {
			graph.CreateCO(pool[(len(pool)-1-output+len(pool))%len(pool)].NotIf(data[output%len(data)]&8 != 0))
		}
		graph.CreateCO(pool[int(data[len(data)-1])%len(pool)].NotIf(data[0]&16 != 0))
		graph.SetConstraintCount(1)
		result, err := ABCSensitivity(graph, ABCSensitivityOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := make([]int, 0, graph.InputCount())
		for variable := 0; variable < graph.InputCount(); variable++ {
			if abcSensitivityModel(graph, variable) {
				want = append(want, variable)
			}
		}
		if !slices.Equal(result.Sensitive, want) || result.Unknown != 0 {
			t.Fatalf("got=%v want=%v", result.Sensitive, want)
		}
	})
}
