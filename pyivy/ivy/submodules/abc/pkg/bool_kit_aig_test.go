package abc

// Standalone tests for src/bool/kit/kitAig.c graph translation.

import (
	"errors"
	"sync"
	"testing"
)

func kitAigEvaluateRoot(t testing.TB, manager *AigManager, root AigLit, inputs []bool) bool {
	t.Helper()
	if _, err := manager.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	outputs, err := manager.Evaluate(inputs)
	if err != nil || len(outputs) != 1 {
		t.Fatalf("outputs=%v err=%v", outputs, err)
	}
	return outputs[0]
}

func TestKitGraphToAigExact(t *testing.T) {
	graph, err := NewKitGraph(3, KitGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := graph.Leaf(0, false)
	b, _ := graph.Leaf(1, true)
	c, _ := graph.Leaf(2, false)
	xor, err := graph.AddXor(a, b, 1)
	if err != nil {
		t.Fatal(err)
	}
	root, err := graph.AddMux(c, xor, a, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRoot(root.Not()); err != nil {
		t.Fatal(err)
	}
	truth, err := graph.Truth32(KitGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	source := NewAigManager()
	fanins := make([]AigLit, 3)
	for index := range fanins {
		fanins[index], err = source.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	before := source.ObjectCount()
	for assignment := 0; assignment < 8; assignment++ {
		target, aigRoot, conversionErr := KitGraphToAig(source, fanins, graph, KitAigOptions{})
		if conversionErr != nil {
			t.Fatal(conversionErr)
		}
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got := kitAigEvaluateRoot(t, target, aigRoot, inputs)
		want := truth&(uint32(1)<<uint(assignment)) != 0
		if got != want {
			t.Fatalf("assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	if source.ObjectCount() != before || source.OutputCount() != 0 {
		t.Fatal("source manager was mutated")
	}
}

func TestKitGraphToAigConstantsAndLiteral(t *testing.T) {
	source := NewAigManager()
	for _, value := range []bool{false, true} {
		target, root, err := KitGraphToAig(source, nil, NewKitConstantGraph(value), KitAigOptions{})
		if err != nil || kitAigEvaluateRoot(t, target, root, nil) != value {
			t.Fatalf("value=%v root=%v err=%v", value, root, err)
		}
	}
	input, err := source.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	graph, err := NewKitLeafGraph(0, 1, true, KitGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []bool{false, true} {
		target, root, conversionErr := KitGraphToAig(source, []AigLit{input}, graph, KitAigOptions{})
		if conversionErr != nil || kitAigEvaluateRoot(t, target, root, []bool{value}) != !value {
			t.Fatalf("value=%v root=%v err=%v", value, root, conversionErr)
		}
	}
}

func TestKitGraphToAigLimitsAndAtomicity(t *testing.T) {
	source := NewAigManager()
	aigInput, err := source.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	graph, err := NewKitGraph(1, KitGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	leaf, _ := graph.Leaf(0, false)
	root, err := graph.AddAnd(leaf, leaf.Not())
	if err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRoot(root); err != nil {
		t.Fatal(err)
	}
	before := source.ObjectCount()
	for _, options := range []KitAigOptions{{MaxObjects: before}, {MaxWork: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		target, result, conversionErr := KitGraphToAig(source, []AigLit{aigInput}, graph, options)
		if conversionErr == nil || target != nil || result != InvalidAigLit || source.ObjectCount() != before {
			t.Fatalf("options=%+v target=%v result=%v err=%v", options, target, result, conversionErr)
		}
	}
	if target, result, err := KitGraphToAig(source, nil, graph, KitAigOptions{}); !errors.Is(err, ErrDimension) || target != nil || result != InvalidAigLit {
		t.Fatalf("shape target=%v result=%v err=%v", target, result, err)
	}
	if target, result, err := KitGraphToAig(source, []AigLit{InvalidAigLit}, graph, KitAigOptions{}); !errors.Is(err, ErrInvalidLiteral) || target != nil || result != InvalidAigLit {
		t.Fatalf("literal target=%v result=%v err=%v", target, result, err)
	}
}

func TestKitGraphToAigConcurrentIndependentResults(t *testing.T) {
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
	source := NewAigManager()
	first, _ := source.CreateCI()
	second, _ := source.CreateCI()
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			target, aigRoot, conversionErr := KitGraphToAig(source, []AigLit{first, second}, graph, KitAigOptions{})
			if conversionErr != nil {
				t.Error(conversionErr)
				return
			}
			if !kitAigEvaluateRoot(t, target, aigRoot, []bool{false, true}) {
				t.Error("wrong concurrent result")
			}
		}()
	}
	wg.Wait()
}

func FuzzKitGraphToAig(f *testing.F) {
	f.Add(byte(0), byte(0), byte(0))
	f.Add(byte(3), byte(7), byte(19))
	f.Fuzz(func(t *testing.T, rawOperation, rawComplements, rawAssignment byte) {
		graph, err := NewKitGraph(3, KitGraphOptions{})
		if err != nil {
			t.Fatal(err)
		}
		edges := make([]KitEdge, 3)
		for index := range edges {
			edges[index], _ = graph.Leaf(index, rawComplements&(1<<uint(index)) != 0)
		}
		var root KitEdge
		switch rawOperation % 4 {
		case 0:
			root, err = graph.AddAnd(edges[0], edges[1])
		case 1:
			root, err = graph.AddOr(edges[0], edges[1])
		case 2:
			root, err = graph.AddXor(edges[0], edges[1], int(rawOperation>>7))
		case 3:
			root, err = graph.AddMux(edges[0], edges[1], edges[2], int(rawOperation>>7))
		}
		if err != nil {
			t.Fatal(err)
		}
		if rawComplements&8 != 0 {
			root = root.Not()
		}
		if err = graph.SetRoot(root); err != nil {
			t.Fatal(err)
		}
		truth, err := graph.Truth32(KitGraphOptions{})
		if err != nil {
			t.Fatal(err)
		}
		source := NewAigManager()
		fanins := make([]AigLit, 3)
		for index := range fanins {
			fanins[index], _ = source.CreateCI()
		}
		target, aigRoot, err := KitGraphToAig(source, fanins, graph, KitAigOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assignment := int(rawAssignment & 7)
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got := kitAigEvaluateRoot(t, target, aigRoot, inputs)
		want := truth&(uint32(1)<<uint(assignment)) != 0
		if got != want {
			t.Fatalf("operation=%d complements=%x assignment=%d got=%v want=%v", rawOperation, rawComplements, assignment, got, want)
		}
	})
}
