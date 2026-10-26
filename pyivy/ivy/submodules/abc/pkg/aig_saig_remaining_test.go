package abc

// Standalone Go tests for deterministic SAIG simulation and autonomous orbits.

import (
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
)

func saigRemainingTraceGraph(t testing.TB) (*AigManager, AigLit, AigLit) {
	t.Helper()
	graph := NewAigManager()
	primary := aigTestCI(t, graph)
	register := aigTestCI(t, graph)
	conjunction := aigTestAnd(t, graph, primary, register.Not())
	aigTestCO(t, graph, register)
	aigTestCO(t, graph, conjunction)
	aigTestCO(t, graph, primary)
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return graph, primary, register
}

func TestSaigCompiledSourceInventory(t *testing.T) {
	units := SaigCompiledSourceUnits()
	if len(units) != 26 || units[0] != "saigCone.c" || units[len(units)-1] != "saigWnd.c" {
		t.Fatal(units)
	}
	seen := make(map[string]bool, len(units))
	for _, unit := range units {
		if unit == "" || seen[unit] {
			t.Fatal("invalid source inventory", unit)
		}
		seen[unit] = true
	}
	units[0] = "changed"
	if SaigCompiledSourceUnits()[0] != "saigCone.c" {
		t.Fatal("inventory storage escaped")
	}
}

func TestSaigDeterministicTraceAndActivity(t *testing.T) {
	graph, primary, register := saigRemainingTraceGraph(t)
	frames := [][]bool{{true}, {false}, {true}}
	result, err := graph.SimulateSaigTrace(frames, []bool{false}, 1, SaigSimulationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantOutputs := [][]bool{{false, true}, {true, false}, {false, true}}
	wantStates := [][]bool{{false}, {true}, {false}, {true}}
	if !slices.EqualFunc(result.Outputs(), wantOutputs, slices.Equal) || !slices.EqualFunc(result.States(), wantStates, slices.Equal) {
		t.Fatal(result.Outputs(), result.States())
	}
	activity := result.Activity()
	if activity[0].Samples != 2 || activity[0].Ones != 2 || activity[0].ProbabilityOne != 1 || activity[0].Switching != 0 {
		t.Fatal(activity[0])
	}
	if activity[primary.ID()].Ones != 1 || activity[primary.ID()].ProbabilityOne != 0.5 || activity[primary.ID()].Switching != 0.5 {
		t.Fatal(activity[primary.ID()])
	}
	if activity[register.ID()].Ones != 1 || activity[register.ID()].ProbabilityOne != 0.5 {
		t.Fatal(activity[register.ID()])
	}

	// Result and input ownership are independent in both directions.
	frames[0][0] = false
	outputs := result.Outputs()
	outputs[0][0] = true
	states := result.States()
	states[0][0] = true
	copyActivity := result.Activity()
	copyActivity[0].Ones = 0
	if !slices.Equal(result.Outputs()[0], wantOutputs[0]) || !slices.Equal(result.States()[0], wantStates[0]) || result.Activity()[0].Ones != 2 {
		t.Fatal("simulation result aliases caller storage")
	}
}

func TestSaigTraceExhaustiveAgainstStep(t *testing.T) {
	graph, _, _ := saigRemainingTraceGraph(t)
	for initialMask := 0; initialMask < 2; initialMask++ {
		for pattern := 0; pattern < 1<<4; pattern++ {
			frames := make([][]bool, 4)
			for frame := range frames {
				frames[frame] = []bool{pattern&(1<<frame) != 0}
			}
			result, err := graph.SimulateSaigTrace(frames, []bool{initialMask != 0}, 0, SaigSimulationOptions{})
			if err != nil {
				t.Fatal(err)
			}
			state := []bool{initialMask != 0}
			ones := make([]int, graph.ObjectCount())
			for frame, row := range frames {
				outputs, next, stepErr := graph.Step(row, state)
				if stepErr != nil || !slices.Equal(outputs, result.outputs[frame]) || !slices.Equal(next, result.states[frame+1]) {
					t.Fatal(initialMask, pattern, frame, outputs, next, stepErr)
				}
				allInputs := append(slices.Clone(row), state...)
				values := make([]bool, graph.ObjectCount())
				values[0] = true
				for index, id := range graph.cis {
					values[id] = allInputs[index]
				}
				for id, object := range graph.objects {
					switch object.Type {
					case AigAND:
						values[id] = saigReadValue(values, object.Fanin0) && saigReadValue(values, object.Fanin1)
					case AigBuffer, AigCO:
						values[id] = saigReadValue(values, object.Fanin0)
					}
					if object.Type != AigNone && values[id] {
						ones[id]++
					}
				}
				state = next
			}
			for id, activity := range result.activity {
				if activity.Live && (activity.Samples != 4 || activity.Ones != ones[id]) {
					t.Fatal(initialMask, pattern, id, activity, ones[id])
				}
			}
		}
	}
}

func TestSaigAutonomousOrbit(t *testing.T) {
	graph := NewAigManager()
	register := aigTestCI(t, graph)
	aigTestCO(t, graph, register)
	aigTestCO(t, graph, register.Not())
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	orbit, err := graph.AnalyzeSaigAutonomousOrbit([]bool{false}, SaigSimulationOptions{MaxFrames: 8})
	if err != nil {
		t.Fatal(err)
	}
	if !orbit.Complete || orbit.Prefix != 0 || orbit.Period != 2 || !slices.EqualFunc(orbit.States, [][]bool{{false}, {true}}, slices.Equal) || !slices.EqualFunc(orbit.Outputs, [][]bool{{false}, {true}}, slices.Equal) {
		t.Fatal(orbit)
	}
	bounded, err := graph.AnalyzeSaigAutonomousOrbit([]bool{false}, SaigSimulationOptions{MaxFrames: 1})
	if err != nil || bounded.Complete || bounded.Period != 0 || len(bounded.States) != 1 {
		t.Fatal(bounded, err)
	}
	// Returned state rows are owned, including the initial row.
	initial := []bool{true}
	owned, err := graph.AnalyzeSaigAutonomousOrbit(initial, SaigSimulationOptions{MaxFrames: 8})
	if err != nil {
		t.Fatal(err)
	}
	initial[0] = false
	if !owned.States[0][0] {
		t.Fatal("orbit aliases initial state")
	}
}

func TestSaigAutonomousOrbitTransientAndEmptyState(t *testing.T) {
	transient := NewAigManager()
	first := aigTestCI(t, transient)
	second := aigTestCI(t, transient)
	aigTestCO(t, transient, first)
	aigTestCO(t, transient, second)
	aigTestCO(t, transient, second)
	if err := transient.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	orbit, err := transient.AnalyzeSaigAutonomousOrbit([]bool{true, false}, SaigSimulationOptions{MaxFrames: 4})
	if err != nil || !orbit.Complete || orbit.Prefix != 1 || orbit.Period != 1 ||
		!slices.EqualFunc(orbit.States, [][]bool{{true, false}, {false, false}}, slices.Equal) ||
		!slices.EqualFunc(orbit.Outputs, [][]bool{{true}, {false}}, slices.Equal) {
		t.Fatal(orbit, err)
	}

	empty := NewAigManager()
	aigTestCO(t, empty, AigConst1)
	orbit, err = empty.AnalyzeSaigAutonomousOrbit(nil, SaigSimulationOptions{MaxFrames: 2})
	if err != nil || !orbit.Complete || orbit.Prefix != 0 || orbit.Period != 1 || len(orbit.States) != 1 || len(orbit.States[0]) != 0 || !slices.Equal(orbit.Outputs[0], []bool{true}) {
		t.Fatal(orbit, err)
	}
}

func TestSaigConcurrentReadOnlyCallsAndSnapshots(t *testing.T) {
	graph, _, _ := saigRemainingTraceGraph(t)
	frames := [][]bool{{true}, {false}, {true}, {true}}
	const workers = 8
	var wait sync.WaitGroup
	wait.Add(workers)
	failures := make(chan int, workers)
	for worker := 0; worker < workers; worker++ {
		go func(id int) {
			defer wait.Done()
			for iteration := 0; iteration < 20; iteration++ {
				result, err := graph.SimulateSaigTrace(frames, []bool{id&1 != 0}, 1, SaigSimulationOptions{})
				if err != nil || len(result.Outputs()) != len(frames) || len(result.States()) != len(frames)+1 || len(result.Activity()) != graph.ObjectCount() {
					failures <- id
					return
				}
			}
		}(worker)
	}
	wait.Wait()
	close(failures)
	for worker := range failures {
		t.Fatalf("worker %d failed", worker)
	}
}

func TestSaigSimulationErrorsLimitsAndAtomicity(t *testing.T) {
	graph, _, _ := saigRemainingTraceGraph(t)
	frames := [][]bool{{true}, {false}}
	tests := []struct {
		graph   *AigManager
		frames  [][]bool
		initial []bool
		prefix  int
		options SaigSimulationOptions
		want    error
	}{
		{nil, frames, []bool{false}, 0, SaigSimulationOptions{}, ErrDimension},
		{graph, [][]bool{{true, false}}, []bool{false}, 0, SaigSimulationOptions{}, ErrDimension},
		{graph, frames, nil, 0, SaigSimulationOptions{}, ErrDimension},
		{graph, frames, []bool{false}, -1, SaigSimulationOptions{}, ErrDimension},
		{graph, frames, []bool{false}, 3, SaigSimulationOptions{}, ErrDimension},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{MaxFrames: 1}, ErrLimit},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{MaxWork: 2*graph.ObjectCount() - 1}, ErrLimit},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{MaxStoredValues: 1}, ErrLimit},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{MaxBytes: 1}, ErrLimit},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{Stop: func() bool { return true }}, ErrCancelled},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{MaxFrames: -1}, ErrDimension},
		{graph, frames, []bool{false}, 0, SaigSimulationOptions{MaxBytes: -1}, ErrDimension},
	}
	for index, test := range tests {
		result, err := test.graph.SimulateSaigTrace(test.frames, test.initial, test.prefix, test.options)
		if result != nil || !errors.Is(err, test.want) {
			t.Fatalf("case %d: result=%v err=%v want=%v", index, result, err, test.want)
		}
	}
	requiredWork := 2*graph.ObjectCount() + len(frames) + len(frames)*2*graph.ObjectCount()
	if result, err := graph.SimulateSaigTrace(frames, []bool{false}, 0, SaigSimulationOptions{MaxWork: requiredWork}); err != nil || result == nil {
		t.Fatal(result, err)
	}
	if result, err := graph.SimulateSaigTrace(frames, []bool{false}, 0, SaigSimulationOptions{MaxWork: requiredWork - 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	stopCalls := 0
	if result, err := graph.SimulateSaigTrace(frames, []bool{false}, 0, SaigSimulationOptions{Stop: func() bool {
		stopCalls++
		return stopCalls == 3
	}}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err, stopCalls)
	}
	if result, err := graph.SimulateSaigTrace(nil, []bool{false}, 0, SaigSimulationOptions{}); err != nil || len(result.outputs) != 0 || len(result.states) != 1 || result.activity[0].Samples != 0 || math.IsNaN(result.activity[0].ProbabilityOne) {
		t.Fatal(result, err)
	}
	if orbit, err := graph.AnalyzeSaigAutonomousOrbit([]bool{false}, SaigSimulationOptions{}); len(orbit.States) != 0 || !errors.Is(err, ErrDimension) {
		t.Fatal(orbit, err)
	}
	autonomous := NewAigManager()
	reg := aigTestCI(t, autonomous)
	aigTestCO(t, autonomous, reg)
	aigTestCO(t, autonomous, reg)
	if err := autonomous.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if orbit, err := autonomous.AnalyzeSaigAutonomousOrbit([]bool{false}, SaigSimulationOptions{MaxFrames: 2, MaxWork: autonomous.ObjectCount()}); len(orbit.States) != 0 || !errors.Is(err, ErrLimit) {
		t.Fatal(orbit, err)
	}
	if orbit, err := autonomous.AnalyzeSaigAutonomousOrbit([]bool{false}, SaigSimulationOptions{Stop: func() bool { return true }}); len(orbit.States) != 0 || !errors.Is(err, ErrCancelled) {
		t.Fatal(orbit, err)
	}
	if orbit, err := autonomous.AnalyzeSaigAutonomousOrbit([]bool{false}, SaigSimulationOptions{MaxFrames: 2, MaxBytes: 1}); len(orbit.States) != 0 || !errors.Is(err, ErrLimit) {
		t.Fatal(orbit, err)
	}
}

func FuzzSaigDeterministicTrace(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3}, true, uint8(1))
	f.Add([]byte{}, false, uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, initial bool, prefixByte uint8) {
		if len(data) > 24 {
			data = data[:24]
		}
		graph, _, _ := saigRemainingTraceGraph(t)
		frames := make([][]bool, len(data))
		for index, value := range data {
			frames[index] = []bool{value&1 != 0}
		}
		prefix := 0
		if len(frames) != 0 {
			prefix = int(prefixByte) % (len(frames) + 1)
		}
		result, err := graph.SimulateSaigTrace(frames, []bool{initial}, prefix, SaigSimulationOptions{})
		if err != nil {
			t.Fatal(err)
		}
		state := []bool{initial}
		for frame, row := range frames {
			outputs, next, stepErr := graph.Step(row, state)
			if stepErr != nil || !slices.Equal(outputs, result.outputs[frame]) || !slices.Equal(next, result.states[frame+1]) {
				t.Fatal(frame, outputs, next, stepErr)
			}
			state = next
		}
		for _, activity := range result.activity {
			if activity.Live && (activity.Samples != len(frames)-prefix || activity.Ones < 0 || activity.Ones > activity.Samples || activity.ProbabilityOne < 0 || activity.ProbabilityOne > 1 || activity.Switching < 0 || activity.Switching > 0.5) {
				t.Fatal(activity)
			}
		}
	})
}
