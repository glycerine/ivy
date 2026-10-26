package abc

// Standalone semantic, limit, concurrency and fuzz tests for llb2Bad.c.

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func llbBadTruthAIG(t testing.TB, firstTruth, secondTruth uint16) *AigManager {
	t.Helper()
	graph := NewAigManager()
	inputs := make([]AigLit, 4)
	for index := range inputs {
		var err error
		inputs[index], err = graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	build := func(truth uint16) AigLit {
		result := AigConst0
		for assignment := 0; assignment < 16; assignment++ {
			if truth&(1<<uint(assignment)) == 0 {
				continue
			}
			term := AigConst1
			for variable, input := range inputs {
				var err error
				term, err = graph.And(term, input.NotIf(assignment&(1<<uint(variable)) == 0))
				if err != nil {
					t.Fatal(err)
				}
			}
			var err error
			result, err = graph.Or(result, term)
			if err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	for _, root := range []AigLit{build(firstTruth), build(secondTruth), inputs[2], inputs[3]} {
		if _, err := graph.CreateCO(root); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return graph
}

func llbBadBDDInputs(aigInputs []bool) []bool {
	return []bool{aigInputs[2], aigInputs[3], aigInputs[0], aigInputs[1]}
}

func llbBadCheckSemantics(t testing.TB, graph *AigManager, manager *ExtraBDDManager, root ExtraBDDLit) {
	t.Helper()
	for assignment := 0; assignment < 16; assignment++ {
		aigInputs := make([]bool, 4)
		for variable := range aigInputs {
			aigInputs[variable] = assignment&(1<<uint(variable)) != 0
		}
		outputs, aigErr := graph.Evaluate(aigInputs)
		got, bddErr := manager.Evaluate(root, llbBadBDDInputs(aigInputs))
		want := outputs[0] || outputs[1]
		if aigErr != nil || bddErr != nil || got != want {
			t.Fatalf("assignment=%04b got=%v want=%v aigErr=%v bddErr=%v", assignment, got, want, aigErr, bddErr)
		}
	}
}

func TestLLBComputeBadAndQuantifyPIsSemantics(t *testing.T) {
	graph := llbBadTruthAIG(t, 0x6996, 0x0f30)
	manager, err := NewExtraBDDManager(4, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	bad, err := LLBComputeBad(graph, manager, LLBBadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	llbBadCheckSemantics(t, graph, manager, bad)
	quantified, err := LLBQuantifyPIs(graph, manager, bad, LLBBadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for state := 0; state < 4; state++ {
		want := false
		for primary := 0; primary < 4; primary++ {
			assignment := primary | state<<2
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
			outputs, _ := graph.Evaluate(inputs)
			want = want || outputs[0] || outputs[1]
		}
		for primary := 0; primary < 4; primary++ {
			got, evalErr := manager.Evaluate(quantified, []bool{state&1 != 0, state&2 != 0, primary&1 != 0, primary&2 != 0})
			if evalErr != nil || got != want {
				t.Fatalf("state=%02b primary=%02b got=%v want=%v err=%v", state, primary, got, want, evalErr)
			}
		}
	}
}

func TestLLBBadConstantsBuffersLimitsAndErrors(t *testing.T) {
	empty := NewAigManager()
	manager0, _ := NewExtraBDDManager(0, 8)
	root, err := LLBComputeBad(empty, manager0, LLBBadOptions{})
	if err != nil || root != ExtraBDDFalse {
		t.Fatalf("empty root=%d err=%v", root, err)
	}

	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer := input
	for index := 0; index < 10000; index++ {
		buffer, err = buffered.CreateBuffer(buffer)
		if err != nil {
			t.Fatal(err)
		}
	}
	buffered.CreateCO(buffer.Not())
	manager1, _ := NewExtraBDDManager(1, 32)
	root, err = LLBComputeBad(buffered, manager1, LLBBadOptions{MaxObjects: 20000, MaxVisits: 30000})
	if err != nil {
		t.Fatal(err)
	}
	for value := 0; value < 2; value++ {
		got, evalErr := manager1.Evaluate(root, []bool{value != 0})
		if evalErr != nil || got != (value == 0) {
			t.Fatalf("buffer value=%d got=%v err=%v", value, got, evalErr)
		}
	}

	graph := llbBadTruthAIG(t, 0x6996, 0x0100)
	for _, options := range []LLBBadOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxOperations: -1}, {MaxBytes: -1}} {
		if result, callErr := LLBComputeBad(graph, manager1, options); result != 0 || !errors.Is(callErr, ErrDimension) {
			t.Fatalf("negative %+v result=%d err=%v", options, result, callErr)
		}
	}
	manager4, _ := NewExtraBDDManager(4, 1<<12)
	checks := []struct {
		options LLBBadOptions
		want    error
	}{
		{LLBBadOptions{MaxObjects: 1}, ErrLimit},
		{LLBBadOptions{MaxVisits: 1}, ErrLimit},
		{LLBBadOptions{MaxOperations: 1}, ErrLimit},
		{LLBBadOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, check := range checks {
		if result, callErr := LLBComputeBad(graph, manager4, check.options); result != 0 || !errors.Is(callErr, check.want) {
			t.Fatalf("options=%+v result=%d err=%v want=%v", check.options, result, callErr, check.want)
		}
	}
	if result, callErr := LLBComputeBad(nil, manager4, LLBBadOptions{}); result != 0 || !errors.Is(callErr, ErrDimension) {
		t.Fatalf("nil graph result=%d err=%v", result, callErr)
	}
	if result, callErr := LLBComputeBad(graph, nil, LLBBadOptions{}); result != 0 || !errors.Is(callErr, ErrDimension) {
		t.Fatalf("nil manager result=%d err=%v", result, callErr)
	}
	if result, callErr := LLBComputeBad(graph, manager1, LLBBadOptions{}); result != 0 || !errors.Is(callErr, ErrDimension) {
		t.Fatalf("mismatch result=%d err=%v", result, callErr)
	}
	bad, _ := LLBComputeBad(graph, manager4, LLBBadOptions{})
	if result, callErr := LLBQuantifyPIs(graph, manager4, ExtraBDDLit(99999), LLBBadOptions{}); result != 0 || !errors.Is(callErr, ErrInvalidLiteral) {
		t.Fatalf("invalid root result=%d err=%v", result, callErr)
	}
	if result, callErr := LLBQuantifyPIs(graph, manager4, bad, LLBBadOptions{MaxOperations: 1}); result != 0 || !errors.Is(callErr, ErrLimit) {
		t.Fatalf("quantify limit result=%d err=%v", result, callErr)
	}
	if result, callErr := LLBQuantifyPIs(graph, manager4, bad, LLBBadOptions{Stop: func() bool { return true }}); result != 0 || !errors.Is(callErr, ErrCancelled) {
		t.Fatalf("quantify cancel result=%d err=%v", result, callErr)
	}
}

func TestLLBBadConcurrentIndependentManagers(t *testing.T) {
	graph := llbBadTruthAIG(t, 0x4bd2, 0xa35c)
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			manager, err := NewExtraBDDManager(4, 1<<12)
			if err != nil {
				failures.Add(1)
				return
			}
			bad, err := LLBComputeBad(graph, manager, LLBBadOptions{})
			if err != nil {
				failures.Add(1)
				return
			}
			quantified, err := LLBQuantifyPIs(graph, manager, bad, LLBBadOptions{})
			if err != nil || quantified == ExtraBDDFalse {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzLLBBadSemantics(f *testing.F) {
	f.Add(uint16(0x6996), uint16(0x0f30))
	f.Add(uint16(0), uint16(0xffff))
	f.Fuzz(func(t *testing.T, firstTruth, secondTruth uint16) {
		graph := llbBadTruthAIG(t, firstTruth, secondTruth)
		manager, err := NewExtraBDDManager(4, 1<<14)
		if err != nil {
			t.Fatal(err)
		}
		bad, err := LLBComputeBad(graph, manager, LLBBadOptions{MaxObjects: 1 << 14, MaxVisits: 1 << 16, MaxOperations: 1 << 20, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		llbBadCheckSemantics(t, graph, manager, bad)
		quantified, err := LLBQuantifyPIs(graph, manager, bad, LLBBadOptions{MaxObjects: 1 << 14, MaxVisits: 1 << 16, MaxOperations: 1 << 20, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		for state := 0; state < 4; state++ {
			want := false
			for primary := 0; primary < 4; primary++ {
				assignment := primary | state<<2
				want = want || firstTruth&(1<<uint(assignment)) != 0 || secondTruth&(1<<uint(assignment)) != 0
			}
			got, evalErr := manager.Evaluate(quantified, []bool{state&1 != 0, state&2 != 0, false, true})
			if evalErr != nil || got != want {
				t.Fatalf("state=%d got=%v want=%v err=%v", state, got, want, evalErr)
			}
		}
	})
}
