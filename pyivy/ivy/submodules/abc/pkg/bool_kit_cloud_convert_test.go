package abc

// Standalone tests for src/bool/kit/kitCloud.c conversion semantics.

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func kitTruthFromBits(t testing.TB, variables int, bitsValue uint64) *TruthTable {
	t.Helper()
	truth, err := NewConstantTruthTable(variables, false, 1024)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if bitsValue&(uint64(1)<<uint(assignment)) != 0 {
			if err = truth.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return truth
}

func TestKitCompiledSourceInventory(t *testing.T) {
	want := []string{"kitAig.c", "kitBdd.c", "kitCloud.c", "cloud.c", "kitDsd.c", "kitFactor.c", "kitGraph.c", "kitHop.c", "kitIsop.c", "kitPla.c", "kitSop.c", "kitTruth.c"}
	got := KitCompiledSourceInventory()
	if !slices.Equal(got, want) {
		t.Fatalf("inventory=%v", got)
	}
	got[0] = "changed"
	if KitCompiledSourceInventory()[0] != "kitAig.c" {
		t.Fatal("inventory aliases caller")
	}
}

func TestKitCloudConversionExhaustiveSmall(t *testing.T) {
	for variables := 0; variables <= 3; variables++ {
		functions := uint64(1) << uint(uint64(1)<<uint(variables))
		for function := uint64(0); function < functions; function++ {
			truth := kitTruthFromBits(t, variables, function)
			manager, err := NewCloudManager(variables, 1<<12)
			if err != nil {
				t.Fatal(err)
			}
			root, err := KitTruthToCloud(manager, truth, CloudOptions{MaxOperations: 1 << 16})
			if err != nil {
				t.Fatalf("vars=%d function=%x: %v", variables, function, err)
			}
			roundTrip, err := KitCloudToTruth(manager, root, 1024, CloudOptions{MaxOperations: 1 << 16})
			if err != nil || !truth.Equal(roundTrip) {
				t.Fatalf("vars=%d function=%x equal=%v err=%v", variables, function, truth.Equal(roundTrip), err)
			}
		}
	}
}

func TestKitCloudComplementsSharingAndOwnership(t *testing.T) {
	truth := kitTruthFromBits(t, 4, 0x6996)
	words := truth.Words()
	manager, err := NewCloudManager(4, 256)
	if err != nil {
		t.Fatal(err)
	}
	root, err := KitTruthToCloud(manager, truth, CloudOptions{})
	if err != nil {
		t.Fatal(err)
	}
	notRoot, err := KitTruthToCloud(manager, truth.Not(), CloudOptions{})
	if err != nil || notRoot != root.Not() {
		t.Fatalf("root=%v not=%v err=%v", root, notRoot, err)
	}
	words[0] = 0
	roundTrip, err := KitCloudToTruth(manager, root, 16, CloudOptions{})
	if err != nil || !truth.Equal(roundTrip) {
		t.Fatalf("ownership equal=%v err=%v", truth.Equal(roundTrip), err)
	}
	copyWords := roundTrip.Words()
	copyWords[0] = 0
	if !truth.Equal(roundTrip) {
		t.Fatal("round-trip words alias caller")
	}
}

func TestKitCloudConversionLimitsRollbackAndCancellation(t *testing.T) {
	truth := kitTruthFromBits(t, 5, 0x96696996)
	manager, err := NewCloudManager(5, 256)
	if err != nil {
		t.Fatal(err)
	}
	before := len(manager.nodes)
	for _, options := range []CloudOptions{{MaxOperations: -1}, {MaxOperations: 3}, {Stop: func() bool { return true }}} {
		root, conversionErr := KitTruthToCloud(manager, truth, options)
		if conversionErr == nil || root != 0 || len(manager.nodes) != before {
			t.Fatalf("options=%+v root=%v err=%v nodes=%d/%d", options, root, conversionErr, len(manager.nodes), before)
		}
	}
	root, err := KitTruthToCloud(manager, truth, CloudOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := KitCloudToTruth(manager, root, -1, CloudOptions{}); !errors.Is(err, ErrDimension) || got != nil {
		t.Fatalf("word limit got=%v err=%v", got, err)
	}
	if got, err := KitCloudToTruth(manager, root, 16, CloudOptions{MaxOperations: 2}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatalf("work limit got=%v err=%v", got, err)
	}
	calls := 0
	if got, err := KitCloudToTruth(manager, root, 16, CloudOptions{Stop: func() bool {
		calls++
		return calls == 7
	}}); !errors.Is(err, ErrCancelled) || got != nil {
		t.Fatalf("cancel got=%v err=%v calls=%d", got, err, calls)
	}
	other, err := NewCloudManager(4, 64)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = KitTruthToCloud(other, truth, CloudOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("dimension err=%v", err)
	}
}

func TestKitCloudConversionConcurrentIndependentManagers(t *testing.T) {
	truth := kitTruthFromBits(t, 6, 0x6996966996696996)
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager, err := NewCloudManager(6, 1024)
			if err != nil {
				t.Error(err)
				return
			}
			root, err := KitTruthToCloud(manager, truth, CloudOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			got, err := KitCloudToTruth(manager, root, 16, CloudOptions{})
			if err != nil || !truth.Equal(got) {
				t.Errorf("equal=%v err=%v", truth.Equal(got), err)
			}
		}()
	}
	wg.Wait()
}

func FuzzKitCloudConversion(f *testing.F) {
	for _, seed := range []struct {
		variables byte
		truth     uint64
	}{{0, 0}, {1, 2}, {3, 0x96}, {6, 0x6996966996696996}} {
		f.Add(seed.variables, seed.truth)
	}
	f.Fuzz(func(t *testing.T, rawVariables byte, bitsValue uint64) {
		variables := int(rawVariables % 7)
		truth := kitTruthFromBits(t, variables, bitsValue)
		manager, err := NewCloudManager(variables, 4096)
		if err != nil {
			t.Fatal(err)
		}
		root, err := KitTruthToCloud(manager, truth, CloudOptions{MaxOperations: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			values := make([]bool, variables)
			for variable := range values {
				values[variable] = assignment&(1<<uint(variable)) != 0
			}
			got, evalErr := manager.Evaluate(root, values)
			want, truthErr := truth.Get(uint64(assignment))
			if evalErr != nil || truthErr != nil || got != want {
				t.Fatalf("assignment=%d got=%v want=%v eval=%v truth=%v", assignment, got, want, evalErr, truthErr)
			}
		}
		roundTrip, err := KitCloudToTruth(manager, root, 1024, CloudOptions{MaxOperations: 1 << 16})
		if err != nil || !truth.Equal(roundTrip) {
			t.Fatalf("equal=%v err=%v", truth.Equal(roundTrip), err)
		}
	})
}
