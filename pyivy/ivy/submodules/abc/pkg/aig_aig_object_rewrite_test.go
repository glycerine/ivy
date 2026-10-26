package abc

// Stand-alone tests for buffer propagation and replacement from aigObj.c.

import (
	"errors"
	"reflect"
	"testing"
)

func aigObjectRewriteHasBuffer(graph *AigManager) bool {
	for _, object := range graph.objects {
		if object.Type == AigBuffer {
			return true
		}
	}
	return false
}

func TestAigObjectPropagateBuffers(t *testing.T) {
	graph := NewAigManager()
	graph.Name, graph.Spec = "buffered", "owned"
	v := aigObjectVariables(t, graph, 4)
	base := aigObjectAnd(t, graph, v[0].Not(), v[1])
	first, err := graph.CreateBuffer(base.Not())
	if err != nil {
		t.Fatal(err)
	}
	second, err := graph.CreateBuffer(first.Not())
	if err != nil {
		t.Fatal(err)
	}
	root := aigObjectAnd(t, graph, second, v[2].Not())
	if _, err = graph.CreateBuffer(v[3]); err != nil { // unrelated live buffer
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(root.Not()); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(first); err != nil {
		t.Fatal(err)
	}
	result, err := graph.PropagateBuffers(AigRewriteOptions{})
	if err != nil || result.Graph == nil || result.BypassedBuffers != 3 || aigObjectRewriteHasBuffer(result.Graph) {
		t.Fatal(result.BypassedBuffers, err, result.Graph)
	}
	if result.Graph.Name != graph.Name || result.Graph.Spec != graph.Spec || result.Graph.Check() != nil {
		t.Fatal(result.Graph.Name, result.Graph.Spec, result.Graph.Check())
	}
	if result.ObjectMap[first.ID()] != result.ObjectMap[base.ID()].Not() || result.ObjectMap[second.ID()] != result.ObjectMap[base.ID()] {
		t.Fatal(result.ObjectMap[first.ID()], result.ObjectMap[second.ID()], result.ObjectMap[base.ID()])
	}
	for bits := 0; bits < 16; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		before, beforeErr := graph.Evaluate(inputs)
		after, afterErr := result.Graph.Evaluate(inputs)
		if beforeErr != nil || afterErr != nil || !reflect.DeepEqual(before, after) {
			t.Fatal(bits, before, after, beforeErr, afterErr)
		}
	}
}

func TestAigObjectReplaceForwardAndPartitions(t *testing.T) {
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 5)
	old := aigObjectAnd(t, graph, v[0].Not(), v[1])
	firstUse := aigObjectAnd(t, graph, old, v[2])
	replacement := aigObjectAnd(t, graph, v[3], v[4].Not())
	secondUse := aigObjectAnd(t, graph, old.Not(), replacement)
	if _, err := graph.CreateCO(firstUse); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.CreateCO(old.Not()); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.CreateCO(secondUse); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	result, err := graph.ReplaceObject(old.ID(), replacement.Not(), AigRewriteOptions{})
	if err != nil || result.Graph == nil || result.Graph.Check() != nil || result.BypassedBuffers != 0 {
		t.Fatal(err, result.BypassedBuffers, result.Graph)
	}
	if result.Graph.RegisterCount() != 1 || result.Graph.ConstraintCount() != 1 || result.Graph.OutputCount() != 3 {
		t.Fatal(result.Graph.RegisterCount(), result.Graph.ConstraintCount(), result.Graph.OutputCount())
	}
	if result.ObjectMap[old.ID()] != result.ObjectMap[replacement.ID()].Not() {
		t.Fatal(result.ObjectMap[old.ID()], result.ObjectMap[replacement.ID()])
	}
	for bits := 0; bits < 32; bits++ {
		input := func(index int) bool { return bits&(1<<index) != 0 }
		replaced := !(input(3) && !input(4))
		want := []bool{
			replaced && input(2),
			!replaced,
			!replaced && (input(3) && !input(4)),
		}
		got, evalErr := result.Graph.Evaluate([]bool{input(0), input(1), input(2), input(3), input(4)})
		if evalErr != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(bits, got, want, evalErr)
		}
	}
}

func TestAigObjectRewriteErrorsAndOwnership(t *testing.T) {
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 3)
	old := aigObjectAnd(t, graph, v[0], v[1])
	dependent := aigObjectAnd(t, graph, old, v[2])
	buffer, err := graph.CreateBuffer(v[2])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(dependent); err != nil {
		t.Fatal(err)
	}
	before := graph.Clone()
	tests := []struct {
		name        string
		old         AigID
		replacement AigLit
		options     AigRewriteOptions
		want        error
	}{
		{"invalid old", AigID(graph.ObjectCount() + 1), v[0], AigRewriteOptions{}, ErrInvalidLiteral},
		{"terminal old", v[0].ID(), v[1], AigRewriteOptions{}, ErrUnsupported},
		{"invalid replacement", old.ID(), InvalidAigLit, AigRewriteOptions{}, ErrInvalidLiteral},
		{"same object", old.ID(), old.Not(), AigRewriteOptions{}, ErrInvariant},
		{"buffer replacement", old.ID(), buffer, AigRewriteOptions{}, ErrUnsupported},
		{"dependent replacement", old.ID(), dependent, AigRewriteOptions{}, ErrInvariant},
		{"source limit", old.ID(), v[2], AigRewriteOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{"visit limit", old.ID(), v[2], AigRewriteOptions{MaxVisits: 1}, ErrLimit},
		{"cancelled", old.ID(), v[2], AigRewriteOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, rewriteErr := graph.ReplaceObject(test.old, test.replacement, test.options)
			if result.Graph != nil || !errors.Is(rewriteErr, test.want) || !reflect.DeepEqual(graph, before) {
				t.Fatal(result, rewriteErr, test.want, reflect.DeepEqual(graph, before))
			}
		})
	}
	if result, rewriteErr := (*AigManager)(nil).PropagateBuffers(AigRewriteOptions{}); result.Graph != nil || !errors.Is(rewriteErr, ErrDimension) {
		t.Fatal(result, rewriteErr)
	}

	choice := NewAigManager()
	cv := aigObjectVariables(t, choice, 4)
	a := aigObjectAnd(t, choice, cv[0], cv[1])
	b := aigObjectAnd(t, choice, cv[2], cv[3])
	if err = choice.SetChoiceChains([][]AigID{{a.ID(), b.ID()}}, AigTraversalOptions{}); err != nil {
		t.Fatal(err)
	}
	if result, rewriteErr := choice.PropagateBuffers(AigRewriteOptions{}); result.Graph != nil || !errors.Is(rewriteErr, ErrUnsupported) {
		t.Fatal(result, rewriteErr)
	}
}

func TestAigObjectRewriteDeep(t *testing.T) {
	const depth = 10000
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, depth+2)
	old := aigObjectAnd(t, graph, v[0], v[1])
	root := old
	for index := 2; index < depth; index++ {
		root = aigObjectAnd(t, graph, root, v[index])
	}
	replacement := aigObjectAnd(t, graph, v[depth], v[depth+1])
	if _, err := graph.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	result, err := graph.ReplaceObject(old.ID(), replacement, AigRewriteOptions{MaxObjects: graph.ObjectCount(), MaxVisits: 8 * graph.ObjectCount()})
	if err != nil || result.Graph.Check() != nil || result.ObjectMap[old.ID()] != result.ObjectMap[replacement.ID()] {
		t.Fatal(err, result.Graph.Check())
	}
}

func FuzzAigObjectReplace(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x123456789abcdef0, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph := NewAigManager()
		v := aigObjectVariables(t, graph, 5)
		phase := func(bit uint) bool { return seed&(uint64(1)<<bit) != 0 }
		old := aigObjectAnd(t, graph, v[0].NotIf(phase(0)), v[1].NotIf(phase(1)))
		first := aigObjectAnd(t, graph, old.NotIf(phase(2)), v[2].NotIf(phase(3)))
		replacement := aigObjectAnd(t, graph, v[3].NotIf(phase(4)), v[4].NotIf(phase(5)))
		second := aigObjectAnd(t, graph, old.NotIf(phase(6)), replacement.NotIf(phase(7)))
		if _, err := graph.CreateCO(first.NotIf(phase(8))); err != nil {
			t.Fatal(err)
		}
		if _, err := graph.CreateCO(second.NotIf(phase(9))); err != nil {
			t.Fatal(err)
		}
		replacementEdge := replacement.NotIf(phase(10))
		result, err := graph.ReplaceObject(old.ID(), replacementEdge, AigRewriteOptions{})
		if err != nil || result.Graph.Check() != nil || aigObjectRewriteHasBuffer(result.Graph) {
			t.Fatal(seed, err, result.Graph.Check())
		}
		for bits := 0; bits < 32; bits++ {
			value := func(index int, inverted bool) bool { return (bits&(1<<index) != 0) != inverted }
			repl := value(3, phase(4)) && value(4, phase(5))
			substitute := repl != phase(10)
			wantFirst := ((substitute != phase(2)) && value(2, phase(3))) != phase(8)
			wantSecond := ((substitute != phase(6)) && (repl != phase(7))) != phase(9)
			got, evalErr := result.Graph.Evaluate([]bool{
				bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0, bits&16 != 0,
			})
			if evalErr != nil || len(got) != 2 || got[0] != wantFirst || got[1] != wantSecond {
				t.Fatal(seed, bits, got, wantFirst, wantSecond, evalErr)
			}
		}
	})
}
