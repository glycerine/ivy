package abc

// Tests for timing-box topology and choice levels from aig/aig/aigDfs.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigBoxDFSFixture(t testing.TB, seed int) (*AigManager, *TimManager) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	leaf := make([]AigLit, 4)
	for i := range leaf {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed>>i&1 != 0)
	}
	first, _ := g.And(leaf[0], leaf[1])
	firstDeep, _ := g.And(first, leaf[2])
	if seed&1 != 0 {
		ci[4], _ = g.CreateCI()
	}
	g.CreateCO(firstDeep.NotIf(seed>>4&1 != 0))
	if seed&1 == 0 {
		ci[4], _ = g.CreateCI()
	}
	middle, _ := g.And(ci[4].NotIf(seed>>1&1 != 0), leaf[3])
	middleDeep, _ := g.And(middle, leaf[0])
	if seed&2 != 0 {
		ci[5], _ = g.CreateCI()
	}
	g.CreateCO(middleDeep.NotIf(seed>>3&1 != 0))
	if seed&2 == 0 {
		ci[5], _ = g.CreateCI()
	}
	other := firstDeep
	if seed&4 != 0 {
		other = middleDeep
	}
	last, _ := g.And(ci[5].NotIf(seed>>2&1 != 0), other)
	g.CreateCO(last.NotIf(seed>>4&1 != 0))
	timing, err := NewTimManager(6, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = timing.AddBox(TimBox{FirstInput: 0, Inputs: 1, FirstOutput: 4, Outputs: 1}, TimBoxOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err = timing.AddBox(TimBox{FirstInput: 1, Inputs: 1, FirstOutput: 5, Outputs: 1, DelayTable: 1}, TimBoxOptions{}); err != nil {
		t.Fatal(err)
	}
	return g, timing
}

func TestAigBoxDFSNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_box_dfs.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, Objects, Topological, Maximum int
		Levels                              []uint32
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g, timing := aigBoxDFSFixture(t, row.Seed)
		beforeGraph, beforeTiming := g.Clone(), timSnapshot(timing)
		err = g.VerifyTopologicalOrderWithBoxes(timing, AigTraversalOptions{})
		if (err == nil) != (row.Topological == 1) || err != nil && !errors.Is(err, ErrInvariant) {
			t.Fatal(row.Seed, row.Topological, err)
		}
		levels, levelErr := g.ComputeChoiceLevelsWithBoxes(timing, AigTraversalOptions{})
		if levelErr != nil || g.ObjectCount() != row.Objects || int(levels.Maximum) != row.Maximum || !slices.Equal(levels.Levels, row.Levels) {
			t.Fatal(row.Seed, g.ObjectCount(), levels, row, levelErr)
		}
		if !reflect.DeepEqual(g, beforeGraph) || !slices.Equal(timSnapshot(timing), beforeTiming) {
			t.Fatal(row.Seed, "source changed")
		}
	}
}

func TestAigBoxDFSErrorsCyclesAndNoBoxes(t *testing.T) {
	g, timing := aigBoxDFSFixture(t, 0)
	for _, options := range []AigTraversalOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if err := g.VerifyTopologicalOrderWithBoxes(timing, options); !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, err)
		}
		if result, err := g.ComputeChoiceLevelsWithBoxes(timing, options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	if err := g.VerifyTopologicalOrderWithBoxes(nil, AigTraversalOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if result, err := g.ComputeChoiceLevelsWithBoxes(nil, AigTraversalOptions{}); result == nil || err != nil {
		t.Fatal(result, err)
	}
	wrong, _ := NewTimManager(5, 3, 0)
	if err := g.VerifyTopologicalOrderWithBoxes(wrong, AigTraversalOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	overlap, _ := NewTimManager(6, 3, 0)
	overlap.AddBox(TimBox{Inputs: 2, FirstOutput: 4, Outputs: 1}, TimBoxOptions{})
	overlap.AddBox(TimBox{FirstInput: 1, Inputs: 1, FirstOutput: 5, Outputs: 1}, TimBoxOptions{})
	if _, err := g.ComputeChoiceLevelsWithBoxes(overlap, AigTraversalOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	corrupt := timing.Clone()
	corrupt.cis[4].Position = 1
	if err := g.VerifyTopologicalOrderWithBoxes(corrupt, AigTraversalOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	cycle := NewAigManager()
	boxOutput, _ := cycle.CreateCI()
	cycle.CreateCO(boxOutput)
	cyclicTiming, _ := NewTimManager(1, 1, 0)
	cyclicTiming.AddBox(TimBox{Inputs: 1, Outputs: 1}, TimBoxOptions{})
	if err := cycle.VerifyTopologicalOrderWithBoxes(cyclicTiming, AigTraversalOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if levels, err := cycle.ComputeChoiceLevelsWithBoxes(cyclicTiming, AigTraversalOptions{}); levels != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(levels, err)
	}

	noBoxes, _ := NewTimManager(g.InputCount(), g.OutputCount(), 0)
	plain, err := g.ComputeChoiceLevels(AigTraversalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	boxed, err := g.ComputeChoiceLevelsWithBoxes(noBoxes, AigTraversalOptions{})
	if err != nil || !reflect.DeepEqual(plain, boxed) {
		t.Fatal(plain, boxed, err)
	}
}

func TestAigBoxDFSMultipleInputsAndOutputs(t *testing.T) {
	g := NewAigManager()
	a, _ := g.CreateCI()
	b, _ := g.CreateCI()
	c, _ := g.CreateCI()
	first, _ := g.And(a, b)
	g.CreateCO(first)
	second, _ := g.And(first, c)
	g.CreateCO(second)
	out0, _ := g.CreateCI()
	out1, _ := g.CreateCI()
	root, _ := g.And(out0, out1)
	g.CreateCO(root)
	timing, _ := NewTimManager(5, 3, 0)
	if _, err := timing.AddBox(TimBox{Inputs: 2, FirstOutput: 3, Outputs: 2}, TimBoxOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := g.VerifyTopologicalOrderWithBoxes(timing, AigTraversalOptions{}); err != nil {
		t.Fatal(err)
	}
	levels, err := g.ComputeChoiceLevelsWithBoxes(timing, AigTraversalOptions{})
	if err != nil || levels.Levels[out0.ID()] != 3 || levels.Levels[out1.ID()] != 3 || levels.Levels[root.ID()] != 4 || levels.Maximum != 4 {
		t.Fatal(levels, err)
	}
}

func TestAigBoxDFSDeepIterative(t *testing.T) {
	const boxes = 10000
	g := NewAigManager()
	current, _ := g.CreateCI()
	for i := 0; i < boxes; i++ {
		g.CreateCO(current)
		current, _ = g.CreateCI()
	}
	timing, err := NewTimManager(boxes+1, boxes, 0)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < boxes; i++ {
		if _, err = timing.AddBox(TimBox{FirstInput: i, Inputs: 1, FirstOutput: i + 1, Outputs: 1}, TimBoxOptions{}); err != nil {
			t.Fatal(i, err)
		}
	}
	options := AigTraversalOptions{MaxObjects: 30000, MaxVisits: 100000}
	if err = g.VerifyTopologicalOrderWithBoxes(timing, options); err != nil {
		t.Fatal(err)
	}
	levels, err := g.ComputeChoiceLevelsWithBoxes(timing, options)
	if err != nil || levels.Maximum != boxes || levels.Levels[current.ID()] != boxes {
		t.Fatal(levels, err)
	}
}

func FuzzAigBoxDFS(f *testing.F) {
	f.Add(uint8(0), uint8(1))
	f.Add(uint8(31), uint8(12))
	f.Fuzz(func(t *testing.T, code, lengthByte uint8) {
		length := int(lengthByte%32) + 1
		g := NewAigManager()
		current, _ := g.CreateCI()
		valid := true
		for i := 0; i < length; i++ {
			if code>>uint(i%8)&1 != 0 {
				previous := current
				current, _ = g.CreateCI()
				g.CreateCO(previous)
				valid = false
			} else {
				g.CreateCO(current)
				current, _ = g.CreateCI()
			}
		}
		timing, _ := NewTimManager(length+1, length, 0)
		for i := 0; i < length; i++ {
			timing.AddBox(TimBox{FirstInput: i, Inputs: 1, FirstOutput: i + 1, Outputs: 1}, TimBoxOptions{})
		}
		err := g.VerifyTopologicalOrderWithBoxes(timing, AigTraversalOptions{})
		if (err == nil) != valid {
			t.Fatal(valid, err)
		}
		levels, err := g.ComputeChoiceLevelsWithBoxes(timing, AigTraversalOptions{})
		if err != nil || int(levels.Maximum) != length || int(levels.Levels[current.ID()]) != length {
			t.Fatal(length, levels, err)
		}
	})
}
