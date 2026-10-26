package abc

// Stand-alone tests for the BBR transition-system port.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func bbrCounterSystem() BBRTransitionSystem {
	// Two-bit incrementer. Input 1 increments; input 0 holds.
	partitions := make([]BBRPartition, 2)
	for bit := range partitions {
		partitions[bit].NextBit = bit
		partitions[bit].Values = make([]bool, 8)
	}
	bad := make([]bool, 8)
	for state := 0; state < 4; state++ {
		for input := 0; input < 2; input++ {
			next := state
			if input == 1 {
				next = (state + 1) & 3
			}
			row := state*2 + input
			for bit := 0; bit < 2; bit++ {
				partitions[bit].Values[row] = next&(1<<bit) != 0
			}
			bad[row] = state == 3
		}
	}
	return BBRTransitionSystem{StateBits: 2, InputBits: 1, Initial: []uint64{0}, Partitions: partitions, Outputs: []BBROutput{{Values: bad}}}
}

func TestBBRImagePreimageAndOwnership(t *testing.T) {
	system := bbrCounterSystem()
	tree, err := NewBBRImageTree(system, BBROptions{})
	if err != nil {
		t.Fatal(err)
	}
	// Mutating caller storage must not change the tree.
	system.Partitions[0].Values[1] = false
	image, err := tree.Image([]uint64{0, 1}, BBROptions{})
	if err != nil || !slices.Equal(image, []uint64{0, 1, 2}) {
		t.Fatalf("image=%v err=%v", image, err)
	}
	read, err := tree.Read()
	if err != nil || !slices.Equal(read, image) {
		t.Fatalf("read=%v err=%v", read, err)
	}
	read[0] = 99
	again, _ := tree.Read()
	if again[0] == 99 {
		t.Fatal("read result aliases tree")
	}
	preimage, err := tree.Preimage([]uint64{2}, BBROptions{})
	if err != nil || !slices.Equal(preimage, []uint64{1, 2}) {
		t.Fatalf("preimage=%v err=%v", preimage, err)
	}
	if err := tree.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tree.Image(nil, BBROptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed tree err=%v", err)
	}
}

func TestBBRReachabilityAndCounterexample(t *testing.T) {
	tree, err := NewBBRImageTree(bbrCounterSystem(), BBROptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := tree.Reachability(true, BBROptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Counterexample == nil || result.Counterexample.Frame != 3 || result.Counterexample.Output != 0 {
		t.Fatalf("counterexample=%+v", result.Counterexample)
	}
	if !slices.Equal(result.Counterexample.States, []uint64{0, 1, 2, 3}) || !slices.Equal(result.Counterexample.Inputs, []uint64{1, 1, 1, 0}) {
		t.Fatalf("trace=%+v", result.Counterexample)
	}
	valid, err := tree.VerifyCounterexample(result.Counterexample, BBROptions{})
	if err != nil || !valid {
		t.Fatalf("verify=%v err=%v", valid, err)
	}
	corrupt := *result.Counterexample
	corrupt.States = slices.Clone(corrupt.States)
	corrupt.States[2] = 0
	valid, err = tree.VerifyCounterexample(&corrupt, BBROptions{})
	if err != nil || valid {
		t.Fatalf("corrupt verify=%v err=%v", valid, err)
	}
	if !slices.Equal(result.Rings[0], []uint64{0}) || !slices.Equal(result.Rings[3], []uint64{3}) {
		t.Fatalf("rings=%v", result.Rings)
	}

	noCheck, err := tree.Reachability(false, BBROptions{})
	if err != nil || !noCheck.Complete || noCheck.Counterexample != nil || !slices.Equal(noCheck.Reached, []uint64{0, 1, 2, 3}) {
		t.Fatalf("noCheck=%+v err=%v", noCheck, err)
	}
}

func TestBBRDefaultParametersAndValidation(t *testing.T) {
	parameters := DefaultBBRParameters()
	if parameters.MaximumBDDNodes != 50000 || parameters.MaximumIterations != 1000 || !parameters.Partition || !parameters.Reorder || !parameters.ReorderImage || parameters.ExploredFrame != -1 {
		t.Fatalf("defaults=%+v", parameters)
	}
	if BBRBuildSources[0] != "src/bdd/bbr/bbrCex.c" || len(BBRBuildSources) != 4 {
		t.Fatalf("sources=%v", BBRBuildSources)
	}
	bad := bbrCounterSystem()
	bad.Partitions[1].NextBit = 0
	if tree, err := NewBBRImageTree(bad, BBROptions{}); tree != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("duplicate partition tree=%v err=%v", tree, err)
	}
	bad = bbrCounterSystem()
	bad.Outputs[0].Values = bad.Outputs[0].Values[:7]
	if _, err := NewBBRImageTree(bad, BBROptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad output err=%v", err)
	}
}

func TestBBRLimitsCancellationAndAtomicRead(t *testing.T) {
	tree, err := NewBBRImageTree(bbrCounterSystem(), BBROptions{})
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := tree.Image([]uint64{0}, BBROptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result, err := tree.Image([]uint64{0, 1, 2, 3}, BBROptions{MaxOperations: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("limited result=%v err=%v", result, err)
	}
	read, _ := tree.Read()
	if !slices.Equal(read, baseline) {
		t.Fatalf("failed image changed last: %v vs %v", read, baseline)
	}
	if result, err := tree.Preimage([]uint64{1}, BBROptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel result=%v err=%v", result, err)
	}
	if result, err := tree.Reachability(false, BBROptions{MaxBytes: 1}); result.Reached != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("memory result=%+v err=%v", result, err)
	}
	limited, err := tree.Reachability(false, BBROptions{MaxIterations: 1})
	if err != nil || limited.Complete || len(limited.Rings) != 2 || !slices.Equal(limited.Rings[1], []uint64{1}) {
		t.Fatalf("iteration limit result=%+v err=%v", limited, err)
	}
	fromRings := make([]uint64, 0)
	for _, ring := range limited.Rings {
		fromRings = append(fromRings, ring...)
	}
	slices.Sort(fromRings)
	if !slices.Equal(fromRings, limited.Reached) {
		t.Fatalf("rings %v do not partition reached %v", limited.Rings, limited.Reached)
	}
}

func TestBBRConstructionBudgetsAndCancellation(t *testing.T) {
	system := bbrCounterSystem()
	if tree, err := NewBBRImageTree(system, BBROptions{MaxOperations: 1}); tree != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("limited construction tree=%v err=%v", tree, err)
	}
	polls := 0
	if tree, err := NewBBRImageTree(system, BBROptions{Stop: func() bool {
		polls++
		return polls == 2
	}}); tree != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled construction tree=%v err=%v polls=%d", tree, err, polls)
	}
}

func TestBBRStateSetBDD(t *testing.T) {
	manager, err := NewExtraBDDManager(4, 128)
	if err != nil {
		t.Fatal(err)
	}
	system := bbrCounterSystem()
	root, err := system.StateSetBDD(manager, []uint64{1, 3}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		bits := make([]bool, 4)
		for bit := range bits {
			bits[bit] = assignment&(1<<bit) != 0
		}
		value, evalErr := manager.Evaluate(root, bits)
		want := assignment&3 == 1 || assignment&3 == 3
		if evalErr != nil || value != want {
			t.Fatalf("assignment=%d value=%v want=%v err=%v", assignment, value, want, evalErr)
		}
	}
	before := len(manager.nodes)
	if root, err := system.StateSetBDD(manager, []uint64{0, 1, 2, 3}, ExtraBDDOptions{MaxOperations: 1}); root != 0 || !errors.Is(err, ErrLimit) {
		t.Fatalf("limited root=%v err=%v", root, err)
	}
	if len(manager.nodes) != before {
		t.Fatalf("failed build retained nodes: %d -> %d", before, len(manager.nodes))
	}
	if root, err := system.StateSetBDD(manager, []uint64{0, 1}, ExtraBDDOptions{MaxBytes: 20}); root != 0 || !errors.Is(err, ErrLimit) {
		t.Fatalf("preallocation root=%v err=%v", root, err)
	}
	if root, err := system.StateSetBDD(manager, []uint64{0}, ExtraBDDOptions{Stop: func() bool { return true }}); root != 0 || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled root=%v err=%v", root, err)
	}
	closed, err := NewExtraBDDManager(4, 128)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if root, err := system.StateSetBDD(closed, []uint64{0}, ExtraBDDOptions{}); root != 0 || !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed-manager root=%v err=%v", root, err)
	}
}

func TestBBRIndependentRandomModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0xBB12))
	for trial := 0; trial < 5000; trial++ {
		stateBits := 1 + rng.Intn(4)
		inputBits := rng.Intn(3)
		states, inputs := 1<<stateBits, 1<<inputBits
		parts := make([]BBRPartition, stateBits)
		nextTable := make([]uint64, states*inputs)
		for row := range nextTable {
			nextTable[row] = uint64(rng.Intn(states))
		}
		for bit := range parts {
			parts[bit] = BBRPartition{NextBit: bit, Values: make([]bool, len(nextTable))}
			for row, next := range nextTable {
				parts[bit].Values[row] = next&(1<<bit) != 0
			}
		}
		system := BBRTransitionSystem{StateBits: stateBits, InputBits: inputBits, Initial: []uint64{0}, Partitions: parts}
		tree, err := NewBBRImageTree(system, BBROptions{})
		if err != nil {
			t.Fatal(err)
		}
		care := []uint64{uint64(rng.Intn(states)), uint64(rng.Intn(states))}
		got, err := tree.Image(care, BBROptions{})
		if err != nil {
			t.Fatal(err)
		}
		seen := make([]bool, states)
		for _, state := range care {
			for input := 0; input < inputs; input++ {
				seen[nextTable[int(state)*inputs+input]] = true
			}
		}
		want := make([]uint64, 0)
		for state, present := range seen {
			if present {
				want = append(want, uint64(state))
			}
		}
		if !slices.Equal(got, want) {
			t.Fatalf("trial=%d got=%v want=%v", trial, got, want)
		}
	}
}

func FuzzBBRImageAndReachability(f *testing.F) {
	f.Add([]byte{2, 1, 1, 2, 3, 0, 1, 2})
	f.Add([]byte{1, 0, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		stateBits := 1 + int(data[0]%4)
		inputBits := int(data[1] % 3)
		states, inputs := 1<<stateBits, 1<<inputBits
		parts := make([]BBRPartition, stateBits)
		for bit := range parts {
			parts[bit] = BBRPartition{NextBit: bit, Values: make([]bool, states*inputs)}
		}
		for row := 0; row < states*inputs; row++ {
			value := byte(row)
			if len(data) > 2 {
				value = data[2+row%(len(data)-2)]
			}
			next := int(value) & (states - 1)
			for bit := range parts {
				parts[bit].Values[row] = next&(1<<bit) != 0
			}
		}
		tree, err := NewBBRImageTree(BBRTransitionSystem{StateBits: stateBits, InputBits: inputBits, Initial: []uint64{0}, Partitions: parts}, BBROptions{MaxStates: 32, MaxBytes: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		result, err := tree.Reachability(false, BBROptions{MaxStates: 32, MaxBytes: 1 << 16, MaxOperations: 1 << 16, MaxIterations: 32})
		if err != nil {
			t.Fatal(err)
		}
		for _, state := range result.Reached {
			if state >= uint64(states) {
				t.Fatalf("state %d outside %d", state, states)
			}
		}
	})
}
