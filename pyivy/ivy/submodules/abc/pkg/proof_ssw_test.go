package abc

// Tests for the Go port of ABC src/proof/ssw.

import (
	"errors"
	"reflect"
	"testing"
)

func buildSSWToggler(t *testing.T, invertOutput bool) *SSWMachine {
	t.Helper()
	m, err := NewSSWMachine(1, 1, []bool{false}, SSWOptions{})
	if err != nil {
		t.Fatal(err)
	}
	in, _ := m.Input(0)
	state, _ := m.State(0)
	next, err := m.AddXor(state, in, SSWOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetNext([]SSWSignal{next}, SSWOptions{}); err != nil {
		t.Fatal(err)
	}
	out := state
	if invertOutput {
		out = out.Not()
	}
	if err = m.SetOutputs([]SSWSignal{out}, SSWOptions{}); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestSSWMachineSimulationAndLifecycle(t *testing.T) {
	m := buildSSWToggler(t, false)
	trace, err := m.Simulate([]bool{false}, [][]bool{{true}, {false}, {true}, {true}}, SSWOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantState := []bool{false, true, true, false}
	wantOut := []bool{false, true, true, false}
	for i := range trace.Frames {
		if trace.Frames[i].State[0] != wantState[i] || trace.Frames[i].Outputs[0] != wantOut[i] {
			t.Fatalf("frame %d: %#v", i, trace.Frames[i])
		}
	}
	nodes, outputs, next, err := m.Snapshot(SSWOptions{})
	if err != nil || len(nodes) != 4 || len(outputs) != 1 || len(next) != 1 {
		t.Fatalf("snapshot %v %d %d %d", err, len(nodes), len(outputs), len(next))
	}
	nodes[0].Kind = SSWNodeAnd
	outputs[0].Node = 99
	next[0].Node = 99
	if err = m.Validate(SSWOptions{}); err != nil {
		t.Fatal(err)
	}
	m.Close()
	m.Close()
	if _, err = m.Simulate(nil, nil, SSWOptions{}); !errors.Is(err, ErrSSWClosed) {
		t.Fatalf("closed: %v", err)
	}
}

func TestSSWAllTwoInputFunctions(t *testing.T) {
	for truth := 0; truth < 16; truth++ {
		m, err := NewSSWMachine(2, 0, nil, SSWOptions{})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := m.Input(0)
		b, _ := m.Input(1)
		na, nb := a.Not(), b.Not()
		terms := make([]SSWSignal, 0, 4)
		inputs := [][2]SSWSignal{{na, nb}, {a, nb}, {na, b}, {a, b}}
		for assignment := 0; assignment < 4; assignment++ {
			if truth&(1<<assignment) == 0 {
				continue
			}
			term, e := m.AddAnd(inputs[assignment][0], inputs[assignment][1], SSWOptions{})
			if e != nil {
				t.Fatal(e)
			}
			terms = append(terms, term)
		}
		out := m.Constant(false)
		for _, term := range terms {
			// OR(x,y) = !( !x & !y ).
			and, e := m.AddAnd(out.Not(), term.Not(), SSWOptions{})
			if e != nil {
				t.Fatal(e)
			}
			out = and.Not()
		}
		if err = m.SetOutputs([]SSWSignal{out}, SSWOptions{}); err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			trace, e := m.Simulate(nil, [][]bool{{assignment&1 != 0, assignment&2 != 0}}, SSWOptions{})
			if e != nil {
				t.Fatal(e)
			}
			got := trace.Frames[0].Outputs[0]
			want := truth&(1<<assignment) != 0
			if got != want {
				t.Fatalf("truth %x assignment %d: got %v", truth, assignment, got)
			}
		}
	}
}

func TestSSWSignaturesClassesAndSweep(t *testing.T) {
	m, err := NewSSWMachine(2, 0, nil, SSWOptions{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := m.Input(0)
	b, _ := m.Input(1)
	x, _ := m.AddAnd(a, b, SSWOptions{})
	y, _ := m.AddAnd(b, a, SSWOptions{})
	z, _ := m.AddBuffer(x.Not(), SSWOptions{})
	if err = m.SetOutputs([]SSWSignal{x, y, z}, SSWOptions{}); err != nil {
		t.Fatal(err)
	}
	patterns := []SSWPattern{{Inputs: []bool{false, false}}, {Inputs: []bool{true, false}}, {Inputs: []bool{false, true}}, {Inputs: []bool{true, true}}}
	sigs, err := m.Signatures(patterns, SSWOptions{})
	if err != nil || len(sigs) != m.NodeCount() {
		t.Fatalf("signatures %v %d", err, len(sigs))
	}
	wantWords := []struct {
		node int
		word uint64
	}{
		{node: a.Node, word: 0b1010},
		{node: b.Node, word: 0b1100},
		{node: x.Node, word: 0b1000},
		{node: y.Node, word: 0b1000},
		{node: z.Node, word: 0b0111},
	}
	for _, want := range wantWords {
		if len(sigs[want.node].Words) != 1 || sigs[want.node].Words[0] != want.word {
			t.Fatalf("signature node %d: got %#v want %#x", want.node, sigs[want.node].Words, want.word)
		}
	}
	classes, err := m.CandidateClasses(patterns, SSWOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(classes) == 0 {
		t.Fatal("missing candidate class")
	}
	result, err := m.Sweep(patterns, SSWOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Representative[y.Node] != x.Node || result.Inverted[y.Node] {
		t.Fatalf("and class %#v", result)
	}
	if result.Representative[z.Node] != x.Node || !result.Inverted[z.Node] {
		t.Fatalf("phase class %#v", result)
	}
}

func TestSSWEquivalenceAndReplay(t *testing.T) {
	left := buildSSWToggler(t, false)
	right := buildSSWToggler(t, false)
	result, err := CheckSSWEquivalent(left, right, 4, SSWOptions{MaxOperations: 1 << 20})
	if err != nil || !result.Equivalent || result.Counterexample != nil || result.CheckedCases != 32 {
		t.Fatalf("equiv %#v %v", result, err)
	}
	bad := buildSSWToggler(t, true)
	result, err = CheckSSWEquivalent(left, bad, 3, SSWOptions{MaxOperations: 1 << 20})
	if err != nil || result.Equivalent || result.Counterexample == nil {
		t.Fatalf("inequiv %#v %v", result, err)
	}
	ok, err := ReplaySSWCounterexample(left, bad, result.Counterexample, SSWOptions{})
	if err != nil || !ok {
		t.Fatalf("replay %v %v", ok, err)
	}
	result.Counterexample.Left = !result.Counterexample.Left
	ok, err = ReplaySSWCounterexample(left, bad, result.Counterexample, SSWOptions{})
	if err != nil || ok {
		t.Fatalf("corrupt replay %v %v", ok, err)
	}
}

func TestSSWLimitsCancellationAtomicityAndValidation(t *testing.T) {
	if _, err := NewSSWMachine(-1, 0, nil, SSWOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("dimension %v", err)
	}
	if _, err := NewSSWMachine(2, 0, nil, SSWOptions{MaxNodes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit %v", err)
	}
	m, _ := NewSSWMachine(1, 1, []bool{false}, SSWOptions{})
	in, _ := m.Input(0)
	before := m.NodeCount()
	if _, err := m.AddBuffer(in, SSWOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || m.NodeCount() != before {
		t.Fatalf("cancel add %v %d", err, m.NodeCount())
	}
	if err := m.SetOutputs([]SSWSignal{{Node: 999}}, SSWOptions{}); !errors.Is(err, ErrInvalidLiteral) || m.OutputCount() != 0 {
		t.Fatalf("atomic outputs %v", err)
	}
	checks := 0
	outputs := make([]SSWSignal, 16)
	for i := range outputs {
		outputs[i] = in
	}
	if err := m.SetOutputs(outputs, SSWOptions{Stop: func() bool {
		checks++
		return checks == 3
	}}); !errors.Is(err, ErrCancelled) || m.OutputCount() != 0 {
		t.Fatalf("cancel outputs %v count %d checks %d", err, m.OutputCount(), checks)
	}
	if _, err := m.Simulate([]bool{}, [][]bool{{true}}, SSWOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("state dimension %v", err)
	}
	if _, err := m.Signatures([]SSWPattern{{State: []bool{false}, Inputs: []bool{true}}}, SSWOptions{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("signature bytes %v", err)
	}
	if _, err := m.Sweep(nil, SSWOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("sweep limit %v", err)
	}
	corrupt := *m
	corrupt.nodes = slicesCloneSSW(m.nodes)
	corrupt.nodes[len(corrupt.nodes)-1] = SSWNode{Kind: SSWNodeBuffer, A: SSWSignal{Node: len(corrupt.nodes) - 1}, B: SSWSignal{Node: -1}}
	if err := corrupt.Validate(SSWOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("cycle %v", err)
	}
	wrongPartition := *m
	wrongPartition.nodes = slicesCloneSSW(m.nodes)
	wrongPartition.nodes[wrongPartition.inputs[0]].Kind = SSWNodeState
	if err := wrongPartition.Validate(SSWOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("input partition kind %v", err)
	}
}

func slicesCloneSSW(nodes []SSWNode) []SSWNode {
	result := make([]SSWNode, len(nodes))
	copy(result, nodes)
	return result
}

func TestSSWDeterministicPatternsAndInventory(t *testing.T) {
	a, err := SSWDeterministicPatterns(3, 2, 100, 7, SSWOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := SSWDeterministicPatterns(3, 2, 100, 7, SSWOptions{})
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatalf("patterns %v", err)
	}
	if _, err = SSWDeterministicPatterns(3, 2, 101, 7, SSWOptions{MaxPatterns: 100}); !errors.Is(err, ErrLimit) {
		t.Fatalf("pattern limit %v", err)
	}
	if _, err = SSWDeterministicPatterns(0, 0, 1, 7, SSWOptions{MaxBytes: sswPatternBytes - 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("pattern header bytes %v", err)
	}
	if got := SSWBuildInventory(); len(got) != 20 || got[0] != "sswAig.c" || got[len(got)-1] != "sswUnique.c" {
		t.Fatalf("build inventory %#v", got)
	}
	if got := SSWTrackedInventory(); len(got) != 24 || got[0] != "module.make" || got[17] != "sswRarity2.c" {
		t.Fatalf("tracked inventory %#v", got)
	}
}

func TestSSWIndependentFiveThousand(t *testing.T) {
	seed := uint64(0x123456789abcdef)
	for trial := 0; trial < 5000; trial++ {
		seed ^= seed << 7
		seed ^= seed >> 9
		seed ^= seed << 8
		m, err := NewSSWMachine(2, 1, []bool{seed&1 != 0}, SSWOptions{})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := m.Input(0)
		b, _ := m.Input(1)
		s, _ := m.State(0)
		x, _ := m.AddXor(a, b, SSWOptions{})
		y, _ := m.AddAnd(x, s, SSWOptions{})
		if err = m.SetOutputs([]SSWSignal{x, y}, SSWOptions{}); err != nil {
			t.Fatal(err)
		}
		if err = m.SetNext([]SSWSignal{x}, SSWOptions{}); err != nil {
			t.Fatal(err)
		}
		initial := []bool{seed&2 != 0}
		inputs := [][]bool{{seed&4 != 0, seed&8 != 0}, {seed&16 != 0, seed&32 != 0}}
		trace, e := m.Simulate(initial, inputs, SSWOptions{})
		if e != nil {
			t.Fatal(e)
		}
		state := initial[0]
		for f := range inputs {
			xWant := inputs[f][0] != inputs[f][1]
			yWant := xWant && state
			if trace.Frames[f].Outputs[0] != xWant || trace.Frames[f].Outputs[1] != yWant || trace.Frames[f].State[0] != state {
				t.Fatalf("trial %d frame %d", trial, f)
			}
			state = xWant
		}
	}
}

func FuzzSSWSequentialContracts(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 7, 31, 0x12345678, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		m, err := NewSSWMachine(2, 1, []bool{seed&1 != 0}, SSWOptions{MaxOperations: 1 << 18, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := m.Input(0)
		b, _ := m.Input(1)
		s, _ := m.State(0)
		x, _ := m.AddXor(a, b, SSWOptions{})
		n, _ := m.AddXor(s, x, SSWOptions{})
		if err = m.SetOutputs([]SSWSignal{s, x}, SSWOptions{}); err != nil {
			t.Fatal(err)
		}
		if err = m.SetNext([]SSWSignal{n}, SSWOptions{}); err != nil {
			t.Fatal(err)
		}
		inputs := make([][]bool, 4)
		for i := range inputs {
			inputs[i] = []bool{seed&(1<<uint(2*i+1)) != 0, seed&(1<<uint(2*i+2)) != 0}
		}
		trace, e := m.Simulate(m.InitialState(), inputs, SSWOptions{MaxOperations: 1 << 18})
		if e != nil {
			t.Fatal(e)
		}
		state := seed&1 != 0
		for i := range inputs {
			xWant := inputs[i][0] != inputs[i][1]
			if trace.Frames[i].Outputs[0] != state || trace.Frames[i].Outputs[1] != xWant {
				t.Fatalf("frame %d", i)
			}
			state = state != xWant
		}
		patterns, e := SSWDeterministicPatterns(2, 1, 16, seed, SSWOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Sweep(patterns, SSWOptions{MaxOperations: 1 << 18, MaxBytes: 1 << 20}); e != nil {
			t.Fatal(e)
		}
	})
}
