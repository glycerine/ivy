package abc

// Standalone semantic tests for src/map/fpga/fpgaTruth.c.

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func fpgaTruthFixture() ([]*FPGACut, *FPGACut) {
	leaves := []*FPGACut{{}, {}, {}}
	andNot := &FPGACut{One: FPGACutEdge{Cut: leaves[0]}, Two: FPGACutEdge{Cut: leaves[1], Complement: true}}
	nandC := &FPGACut{One: FPGACutEdge{Cut: andNot}, Two: FPGACutEdge{Cut: leaves[2]}, Phase: true}
	root := &FPGACut{One: FPGACutEdge{Cut: andNot, Complement: true}, Two: FPGACutEdge{Cut: nandC}}
	return leaves, root
}

func TestFPGACutTruthAndVolume(t *testing.T) {
	leaves, root := fpgaTruthFixture()
	table, err := FPGACutTruth(root, leaves, FPGACutTruthOptions{})
	if err != nil || table.Variables() != 3 {
		t.Fatalf("truth: table=%v err=%v", table, err)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		a := assignment&1 != 0
		b := assignment&2 != 0
		c := assignment&4 != 0
		x := a && !b
		want := !x && !(x && c)
		got, getErr := table.Get(assignment)
		if getErr != nil || got != want {
			t.Fatalf("assignment %03b: got=%v want=%v err=%v", assignment, got, want, getErr)
		}
	}
	volume, err := FPGACutVolume(root, leaves, FPGACutTruthOptions{})
	if err != nil || volume != 3 {
		t.Fatalf("volume=%d err=%v", volume, err)
	}
	words := table.Words()
	words[0] = 0
	again, _ := FPGACutTruth(root, leaves, FPGACutTruthOptions{})
	if again.Words()[0] == 0 {
		t.Fatal("returned truth or input DAG was not independently owned")
	}
}

func TestFPGACutTruthValidationLimitsAndCancellation(t *testing.T) {
	leaves, root := fpgaTruthFixture()
	bad := []struct {
		root    *FPGACut
		leaves  []*FPGACut
		options FPGACutTruthOptions
		want    error
	}{
		{nil, leaves, FPGACutTruthOptions{}, ErrDimension},
		{root, leaves[:1], FPGACutTruthOptions{}, ErrDimension},
		{root, []*FPGACut{leaves[0], nil}, FPGACutTruthOptions{}, ErrDimension},
		{root, []*FPGACut{leaves[0], leaves[0]}, FPGACutTruthOptions{}, ErrDimension},
		{leaves[0], leaves, FPGACutTruthOptions{}, ErrDimension},
		{root, leaves, FPGACutTruthOptions{MaxCuts: -1}, ErrDimension},
		{root, leaves, FPGACutTruthOptions{MaxWords: -1}, ErrDimension},
		{root, leaves, FPGACutTruthOptions{MaxCuts: 2}, ErrLimit},
		{root, leaves, FPGACutTruthOptions{MaxWords: 3}, ErrLimit},
		{root, leaves, FPGACutTruthOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for i, row := range bad {
		if table, err := FPGACutTruth(row.root, row.leaves, row.options); table != nil || !errors.Is(err, row.want) {
			t.Fatalf("case %d: table=%v err=%v want=%v", i, table, err, row.want)
		}
	}
	missing := &FPGACut{One: FPGACutEdge{Cut: leaves[0]}}
	if _, err := FPGACutVolume(missing, leaves, FPGACutTruthOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("missing input: %v", err)
	}
	cycleA, cycleB := &FPGACut{}, &FPGACut{}
	cycleA.One, cycleA.Two = FPGACutEdge{Cut: cycleB}, FPGACutEdge{Cut: leaves[0]}
	cycleB.One, cycleB.Two = FPGACutEdge{Cut: cycleA}, FPGACutEdge{Cut: leaves[1]}
	if _, err := FPGACutVolume(cycleA, leaves, FPGACutTruthOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("cycle: %v", err)
	}
	deep := leaves[0]
	for i := 0; i < 10_000; i++ {
		deep = &FPGACut{One: FPGACutEdge{Cut: deep}, Two: FPGACutEdge{Cut: leaves[1]}, Phase: i&1 != 0}
	}
	if volume, err := FPGACutVolume(deep, leaves[:2], FPGACutTruthOptions{MaxCuts: 10_000}); err != nil || volume != 10_000 {
		t.Fatalf("deep volume=%d err=%v", volume, err)
	}
	for stopAt := 1; stopAt <= 8; stopAt++ {
		calls := 0
		if table, err := FPGACutTruth(root, leaves, FPGACutTruthOptions{Stop: func() bool { calls++; return calls == stopAt }}); table != nil || !errors.Is(err, ErrCancelled) {
			t.Fatalf("stop %d: table=%v err=%v calls=%d", stopAt, table, err, calls)
		}
	}
}

func TestFPGACutTruthConcurrentFrozenDAG(t *testing.T) {
	leaves, root := fpgaTruthFixture()
	want, _ := FPGACutTruth(root, leaves, FPGACutTruthOptions{})
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 64; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, err := FPGACutTruth(root, leaves, FPGACutTruthOptions{})
			volume, volumeErr := FPGACutVolume(root, leaves, FPGACutTruthOptions{})
			if err != nil || volumeErr != nil || volume != 3 || !got.Equal(want) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures: %d", failures.Load())
	}
}

func FuzzFPGACutTruth(f *testing.F) {
	f.Add([]byte{3, 0, 1, 0, 1, 2, 5, 2, 3, 4, 1})
	f.Add([]byte{2, 0, 1, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		leafCount := 2 + int(data[0]%5)
		leaves := make([]*FPGACut, leafCount)
		all := make([]*FPGACut, leafCount, leafCount+32)
		for i := range leaves {
			leaves[i] = &FPGACut{}
			all[i] = leaves[i]
		}
		for offset := 1; offset+2 < len(data) && len(all) < leafCount+32; offset += 3 {
			available := len(all)
			all = append(all, &FPGACut{
				One:   FPGACutEdge{Cut: all[int(data[offset])%available], Complement: data[offset+2]&1 != 0},
				Two:   FPGACutEdge{Cut: all[int(data[offset+1])%available], Complement: data[offset+2]&2 != 0},
				Phase: data[offset+2]&4 != 0,
			})
		}
		if len(all) == leafCount {
			all = append(all, &FPGACut{One: FPGACutEdge{Cut: leaves[0]}, Two: FPGACutEdge{Cut: leaves[1]}})
		}
		root := all[len(all)-1]
		table, err := FPGACutTruth(root, leaves, FPGACutTruthOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<uint(leafCount); assignment++ {
			values := make(map[*FPGACut]bool, len(all))
			for i, leaf := range leaves {
				values[leaf] = assignment&(1<<uint(i)) != 0
			}
			for _, cut := range all[leafCount:] {
				a, b := values[cut.One.Cut], values[cut.Two.Cut]
				if cut.One.Complement {
					a = !a
				}
				if cut.Two.Complement {
					b = !b
				}
				values[cut] = (a && b) != cut.Phase
			}
			got, getErr := table.Get(uint64(assignment))
			if getErr != nil || got != values[root] {
				t.Fatalf("assignment=%d got=%v want=%v err=%v", assignment, got, values[root], getErr)
			}
		}
		volume, err := FPGACutVolume(root, leaves, FPGACutTruthOptions{})
		if err != nil || volume < 1 || volume > len(all)-leafCount {
			t.Fatalf("volume=%d nodes=%d err=%v", volume, len(all)-leafCount, err)
		}
	})
}
