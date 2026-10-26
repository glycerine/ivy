package abc

// Standalone tests for src/bool/kit/kitGraph.c graph semantics.

import (
	"errors"
	"sync"
	"testing"
)

func TestKitGraphConstantsAndLeaves(t *testing.T) {
	for _, tc := range []struct {
		value bool
		want  uint32
	}{{false, 0}, {true, ^uint32(0)}} {
		got, err := NewKitConstantGraph(tc.value).Truth32(KitGraphOptions{})
		if err != nil || got != tc.want {
			t.Fatalf("value=%v got=%08x err=%v", tc.value, got, err)
		}
	}
	graph, err := NewKitLeafGraph(1, 3, true, KitGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := graph.Truth32(KitGraphOptions{})
	if err != nil || got != ^uint32(0xcccccccc) {
		t.Fatalf("truth=%08x err=%v", got, err)
	}
}

func TestKitGraphOperationsExact(t *testing.T) {
	for form := 0; form <= 1; form++ {
		graph, err := NewKitGraph(3, KitGraphOptions{MaxNodes: 32})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := graph.Leaf(0, false)
		b, _ := graph.Leaf(1, false)
		c, _ := graph.Leaf(2, false)
		xor, err := graph.AddXor(a, b, form)
		if err != nil {
			t.Fatal(err)
		}
		mux, err := graph.AddMux(c, xor, a.Not(), form)
		if err != nil {
			t.Fatal(err)
		}
		if err = graph.SetRoot(mux); err != nil {
			t.Fatal(err)
		}
		truth, err := graph.Truth32(KitGraphOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			aValue := assignment&1 != 0
			bValue := assignment&2 != 0
			cValue := assignment&4 != 0
			want := !aValue
			if cValue {
				want = aValue != bValue
			}
			got := truth&(uint32(1)<<uint(assignment)) != 0
			if got != want {
				t.Fatalf("form=%d assignment=%d got=%v want=%v", form, assignment, got, want)
			}
		}
		if graph.NodeCount() != 9 {
			t.Fatalf("form=%d nodes=%d", form, graph.NodeCount())
		}
		level, err := graph.Level(KitGraphOptions{})
		if err != nil || level != 4 {
			t.Fatalf("level=%d err=%v", level, err)
		}
		if depth, err := graph.LeafDepth(2, KitGraphOptions{}); err != nil || depth != 2 {
			t.Fatalf("control depth=%d err=%v", depth, err)
		}
	}
}

func TestKitGraphLimitsValidationAndAtomicity(t *testing.T) {
	graph, err := NewKitGraph(2, KitGraphOptions{MaxNodes: 4})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := graph.Leaf(0, false)
	b, _ := graph.Leaf(1, false)
	if edge, err := graph.AddXor(a, b, 0); !errors.Is(err, ErrLimit) || edge != (KitEdge{}) || graph.NodeCount() != 2 {
		t.Fatalf("atomic edge=%+v err=%v nodes=%d", edge, err, graph.NodeCount())
	}
	if _, err := graph.AddAnd(a, KitEdge{Node: 99}); !errors.Is(err, ErrInvalidLiteral) || graph.NodeCount() != 2 {
		t.Fatalf("invalid err=%v nodes=%d", err, graph.NodeCount())
	}
	and, err := graph.AddAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	root, err := graph.AddAnd(and, a)
	if err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRoot(root); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.Truth32(KitGraphOptions{MaxWork: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("work err=%v", err)
	}
	if _, err = graph.Truth32(KitGraphOptions{MaxBytes: 8}); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte err=%v", err)
	}
	if _, err = graph.Level(KitGraphOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel err=%v", err)
	}
	if _, err = NewKitLeafGraph(2, 2, false, KitGraphOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("leaf err=%v", err)
	}
}

func TestKitGraphConcurrentFrozenReaders(t *testing.T) {
	graph, err := NewKitGraph(2, KitGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := graph.Leaf(0, false)
	b, _ := graph.Leaf(1, false)
	root, err := graph.AddOr(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRoot(root); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				truth, truthErr := graph.Truth32(KitGraphOptions{})
				level, levelErr := graph.Level(KitGraphOptions{})
				if truthErr != nil || levelErr != nil || truth != 0xeeeeeeee || level != 1 {
					t.Errorf("truth=%08x level=%d errors=%v/%v", truth, level, truthErr, levelErr)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func FuzzKitGraphOperations(f *testing.F) {
	f.Add(byte(2), []byte{0, 1, 2, 3})
	f.Add(byte(5), []byte{3, 7, 11, 19, 23})
	f.Fuzz(func(t *testing.T, rawLeaves byte, operations []byte) {
		leaves := int(rawLeaves%5) + 1
		if len(operations) > 24 {
			t.Skip()
		}
		graph, err := NewKitGraph(leaves, KitGraphOptions{MaxNodes: leaves + 3*len(operations) + 1})
		if err != nil {
			t.Fatal(err)
		}
		edges := make([]KitEdge, leaves)
		models := make([][]bool, leaves)
		assignments := 1 << uint(leaves)
		for leaf := 0; leaf < leaves; leaf++ {
			edges[leaf], _ = graph.Leaf(leaf, false)
			models[leaf] = make([]bool, assignments)
			for assignment := range models[leaf] {
				models[leaf][assignment] = assignment&(1<<uint(leaf)) != 0
			}
		}
		for _, operation := range operations {
			firstIndex := int(operation) % len(edges)
			secondIndex := int(operation/3+1) % len(edges)
			first, second := edges[firstIndex], edges[secondIndex]
			firstModel, secondModel := models[firstIndex], models[secondIndex]
			var edge KitEdge
			model := make([]bool, assignments)
			switch operation % 4 {
			case 0:
				edge, err = graph.AddAnd(first, second)
				for i := range model {
					model[i] = firstModel[i] && secondModel[i]
				}
			case 1:
				edge, err = graph.AddOr(first, second)
				for i := range model {
					model[i] = firstModel[i] || secondModel[i]
				}
			case 2:
				edge, err = graph.AddXor(first, second, int(operation>>7))
				for i := range model {
					model[i] = firstModel[i] != secondModel[i]
				}
			case 3:
				controlIndex := int(operation/5+2) % len(edges)
				edge, err = graph.AddMux(edges[controlIndex], first, second, int(operation>>7))
				for i := range model {
					if models[controlIndex][i] {
						model[i] = firstModel[i]
					} else {
						model[i] = secondModel[i]
					}
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			edges = append(edges, edge)
			models = append(models, model)
		}
		root := edges[len(edges)-1]
		if err = graph.SetRoot(root); err != nil {
			t.Fatal(err)
		}
		truth, err := graph.Truth32(KitGraphOptions{MaxWork: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		want := models[len(models)-1]
		for assignment, value := range want {
			if got := truth&(uint32(1)<<uint(assignment)) != 0; got != value {
				t.Fatalf("assignment=%d got=%v want=%v", assignment, got, value)
			}
		}
	})
}
