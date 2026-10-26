package abc

// Stand-alone tests for src/bool/deco/deco.h.

import (
	"errors"
	"slices"
	"testing"
)

func decoInputs(assignment, variables int) []bool {
	inputs := make([]bool, variables)
	for variable := range inputs {
		inputs[variable] = assignment&(1<<uint(variable)) != 0
	}
	return inputs
}

func TestDecoEdgeCodecAndBounds(t *testing.T) {
	for node := 0; node < 5000; node++ {
		for _, complement := range []bool{false, true} {
			edge, err := NewDecoEdge(node, complement)
			if err != nil {
				t.Fatal(err)
			}
			packed, err := edge.Uint32()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecoEdgeFromUint32(packed)
			if err != nil || decoded != edge {
				t.Fatalf("%+v %08x %+v %v", edge, packed, decoded, err)
			}
			if edge.Not().Not() != edge {
				t.Fatal(edge)
			}
		}
	}
	for _, node := range []int{-1, decoEdgeNodeLimit} {
		if _, err := NewDecoEdge(node, false); !errors.Is(err, ErrDimension) {
			t.Fatalf("node %d: %v", node, err)
		}
	}
	if _, err := (DecoEdge{Node: -1}).Uint32(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := DecoEdgeFromUint32(1 << 31); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	maximum := DecoEdge{Node: decoEdgeNodeLimit - 1, Complement: true}
	packed, err := maximum.Uint32()
	if err != nil || packed != 1<<31-1 {
		t.Fatalf("maximum edge: %#x %v", packed, err)
	}
	if decoded, decodeErr := DecoEdgeFromUint32(packed); decodeErr != nil || decoded != maximum {
		t.Fatalf("maximum decode: %+v %v", decoded, decodeErr)
	}
}

func TestDecoFunctionUnionBounds(t *testing.T) {
	var node DecoNode
	for _, literal := range []int{-1 << 31, -1, 0, 1, 1<<31 - 1} {
		if err := node.SetFunctionLiteral(literal); err != nil {
			t.Fatalf("literal %d: %v", literal, err)
		}
		if got := node.FunctionLiteral(); got != literal {
			t.Fatalf("literal %d became %d", literal, got)
		}
	}
	for _, literal := range []int{-1<<31 - 1, 1 << 31} {
		if err := node.SetFunctionLiteral(literal); !errors.Is(err, ErrDimension) {
			t.Fatalf("out-of-range literal %d: %v", literal, err)
		}
	}
	token := ^uintptr(0)
	if err := node.SetFunctionToken(token); err != nil {
		t.Fatal(err)
	}
	if node.FunctionToken() != token || node.FunctionLiteral() != -1 {
		t.Fatalf("token union view: %#x %d", node.FunctionToken(), node.FunctionLiteral())
	}
	var nilNode *DecoNode
	if err := nilNode.SetFunctionLiteral(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := nilNode.SetFunctionToken(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestDecoConstantsLeavesCapacityAndTraversal(t *testing.T) {
	zero, err := NewConstantDecoGraph(false, DecoGraphLimits{})
	if err != nil {
		t.Fatal(err)
	}
	one, _ := NewConstantDecoGraph(true, DecoGraphLimits{})
	if got, ok := zero.ConstantValue(); !ok || got {
		t.Fatal(got, ok)
	}
	if got, ok := one.ConstantValue(); !ok || !got {
		t.Fatal(got, ok)
	}
	if err = zero.Complement(); err != nil {
		t.Fatal(err)
	}
	if got, ok := zero.ConstantValue(); !ok || !got {
		t.Fatal(got, ok)
	}
	leaf, err := NewLeafDecoGraph(2, 3, true, DecoGraphLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if capacity, _ := leaf.Capacity(); capacity != 56 {
		t.Fatalf("capacity %d", capacity)
	}
	if count, _ := leaf.LeafCount(); count != 3 {
		t.Fatal(count)
	}
	if count, _ := leaf.InternalCount(); count != 0 {
		t.Fatal(count)
	}
	if variable, _ := leaf.Variable(); variable != 2 {
		t.Fatal(variable)
	}
	leaves, err := leaf.Leaves()
	if err != nil || len(leaves) != 3 {
		t.Fatal(leaves, err)
	}
	for index, record := range leaves {
		if record.Index != index || record.Node != (DecoNode{}) {
			t.Fatal(record)
		}
	}
	leaves[0].Node.Level = 10
	if node, _ := leaf.Node(0); node.Level != 0 {
		t.Fatal("snapshot aliases graph")
	}
	if got, err := leaf.Evaluate([]bool{false, false, true}); err != nil || got {
		t.Fatalf("complemented leaf = %v %v", got, err)
	}
	if err = leaf.Complement(); err != nil {
		t.Fatal(err)
	}
	if got, _ := leaf.Evaluate([]bool{false, false, true}); !got {
		t.Fatal("double complement")
	}
	growing, _ := NewDecoGraph(2, DecoGraphLimits{MaxNodes: 200})
	for index := 0; index < 53; index++ {
		if _, err = growing.AddAND(DecoEdge{Node: 0}, DecoEdge{Node: 1}); err != nil {
			t.Fatal(index, err)
		}
	}
	if capacity, _ := growing.Capacity(); capacity != 108 {
		t.Fatalf("doubled capacity %d", capacity)
	}
}

func TestDecoANDORXORMUXTruthAndMetadata(t *testing.T) {
	graph, err := NewDecoGraph(3, DecoGraphLimits{})
	if err != nil {
		t.Fatal(err)
	}
	x := DecoEdge{Node: 0}
	y := DecoEdge{Node: 1}
	z := DecoEdge{Node: 2}
	and, err := graph.AddAND(x.Not(), y)
	if err != nil {
		t.Fatal(err)
	}
	node, _ := graph.Node(and.Node)
	if !node.OriginalComplement0 || node.OriginalComplement1 || node.NodeOr || node.Edge0 != x.Not() || node.Edge1 != y {
		t.Fatalf("AND metadata %+v", node)
	}
	if err = node.SetFunctionLiteral(-17); err != nil {
		t.Fatal(err)
	}
	if node.FunctionLiteral() != -17 || node.FunctionToken() != uintptr(^uint32(16)) {
		t.Fatalf("literal union view: %d %#x", node.FunctionLiteral(), node.FunctionToken())
	}
	if err = node.SetFunctionToken(23); err != nil {
		t.Fatal(err)
	}
	if node.FunctionLiteral() != 23 || node.FunctionToken() != 23 {
		t.Fatalf("token union view: %d %#x", node.FunctionLiteral(), node.FunctionToken())
	}
	node.Level = 31
	node.Latches0, node.Latches1, node.Latches2 = 1, 2, 3
	if err = graph.WriteNode(and.Node, node); err != nil {
		t.Fatal(err)
	}
	if written, _ := graph.Node(and.Node); written != node {
		t.Fatalf("written metadata %+v", written)
	}
	or, err := graph.AddOR(x, y.Not())
	if err != nil {
		t.Fatal(err)
	}
	node, _ = graph.Node(or.Node)
	if !node.NodeOr || node.OriginalComplement0 || !node.OriginalComplement1 || node.Edge0 != x.Not() || node.Edge1 != y {
		t.Fatalf("OR metadata %+v edge=%+v", node, or)
	}
	xor0, err := graph.AddXOR(x, y, false)
	if err != nil {
		t.Fatal(err)
	}
	xor1, err := graph.AddXOR(x, y, true)
	if err != nil {
		t.Fatal(err)
	}
	mux0, err := graph.AddMUX(x, y, z, false)
	if err != nil {
		t.Fatal(err)
	}
	mux1, err := graph.AddMUX(x, y, z, true)
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		edge DecoEdge
		want func(bool, bool, bool) bool
	}{
		{and, func(a, b, _ bool) bool { return !a && b }},
		{or, func(a, b, _ bool) bool { return a || !b }},
		{xor0, func(a, b, _ bool) bool { return a != b }},
		{xor1, func(a, b, _ bool) bool { return a != b }},
		{mux0, func(c, a, b bool) bool {
			if c {
				return a
			}
			return b
		}},
		{mux1, func(c, a, b bool) bool {
			if c {
				return a
			}
			return b
		}},
	}
	for _, check := range checks {
		if err = graph.SetRoot(check.edge); err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := decoInputs(assignment, 3)
			got, evaluateErr := graph.Evaluate(inputs)
			want := check.want(inputs[0], inputs[1], inputs[2])
			if evaluateErr != nil || got != want {
				t.Fatalf("edge=%+v input=%v got=%v want=%v err=%v", check.edge, inputs, got, want, evaluateErr)
			}
		}
	}
	if internals, _ := graph.Internals(); len(internals) != 14 {
		t.Fatalf("internal count %d", len(internals))
	}
}

func TestDecoXORMUXExactLowering(t *testing.T) {
	type operation struct {
		name string
		add  func(*DecoGraph, DecoEdge, DecoEdge, DecoEdge) (DecoEdge, error)
		want []DecoNode
		root DecoEdge
	}
	x, y, z := DecoEdge{Node: 0}, DecoEdge{Node: 1}, DecoEdge{Node: 2}
	operations := []operation{
		{
			name: "xor-type-0",
			add:  func(g *DecoGraph, a, b, _ DecoEdge) (DecoEdge, error) { return g.AddXOR(a, b, false) },
			want: []DecoNode{
				{Edge0: x.Not(), Edge1: y, OriginalComplement0: true},
				{Edge0: x, Edge1: y.Not(), OriginalComplement1: true},
				{Edge0: DecoEdge{Node: 3, Complement: true}, Edge1: DecoEdge{Node: 4, Complement: true}, NodeOr: true},
			},
			root: DecoEdge{Node: 5, Complement: true},
		},
		{
			name: "xor-type-1",
			add:  func(g *DecoGraph, a, b, _ DecoEdge) (DecoEdge, error) { return g.AddXOR(a, b, true) },
			want: []DecoNode{
				{Edge0: x, Edge1: y},
				{Edge0: x.Not(), Edge1: y.Not(), OriginalComplement0: true, OriginalComplement1: true},
				{Edge0: DecoEdge{Node: 3, Complement: true}, Edge1: DecoEdge{Node: 4, Complement: true}, NodeOr: true},
			},
			root: DecoEdge{Node: 5},
		},
		{
			name: "mux-type-0",
			add:  func(g *DecoGraph, a, b, c DecoEdge) (DecoEdge, error) { return g.AddMUX(a, b, c, false) },
			want: []DecoNode{
				{Edge0: x, Edge1: y},
				{Edge0: x.Not(), Edge1: z, OriginalComplement0: true},
				{Edge0: DecoEdge{Node: 3, Complement: true}, Edge1: DecoEdge{Node: 4, Complement: true}, NodeOr: true},
			},
			root: DecoEdge{Node: 5, Complement: true},
		},
		{
			name: "mux-type-1",
			add:  func(g *DecoGraph, a, b, c DecoEdge) (DecoEdge, error) { return g.AddMUX(a, b, c, true) },
			want: []DecoNode{
				{Edge0: x, Edge1: y.Not(), OriginalComplement1: true},
				{Edge0: x.Not(), Edge1: z.Not(), OriginalComplement0: true, OriginalComplement1: true},
				{Edge0: DecoEdge{Node: 3, Complement: true}, Edge1: DecoEdge{Node: 4, Complement: true}, NodeOr: true},
			},
			root: DecoEdge{Node: 5},
		},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			graph, err := NewDecoGraph(3, DecoGraphLimits{})
			if err != nil {
				t.Fatal(err)
			}
			root, err := operation.add(graph, x, y, z)
			if err != nil {
				t.Fatal(err)
			}
			if root != operation.root {
				t.Fatalf("root %+v want %+v", root, operation.root)
			}
			internals, err := graph.Internals()
			if err != nil {
				t.Fatal(err)
			}
			if len(internals) != len(operation.want) {
				t.Fatalf("node count %d", len(internals))
			}
			for index, record := range internals {
				if record.Index != index+3 || record.Node != operation.want[index] {
					t.Fatalf("node %d: %+v want %+v", record.Index, record.Node, operation.want[index])
				}
			}
		})
	}
}

func TestDecoLimitsCancellationAtomicityCloneAndClose(t *testing.T) {
	if _, err := NewDecoGraph(-1, DecoGraphLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewDecoGraph(3, DecoGraphLimits{MaxLeaves: 2, MaxNodes: 10}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := NewDecoGraph(1, DecoGraphLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	graph, err := NewDecoGraph(2, DecoGraphLimits{MaxNodes: 4})
	if err != nil {
		t.Fatal(err)
	}
	x, y := DecoEdge{Node: 0}, DecoEdge{Node: 1}
	if _, err = graph.AddAND(x, y); err != nil {
		t.Fatal(err)
	}
	before, _ := graph.Internals()
	if _, err = graph.AddXOR(x, y, false); !errors.Is(err, ErrLimit) {
		t.Fatalf("XOR limit %v", err)
	}
	after, _ := graph.Internals()
	if !slices.Equal(before, after) {
		t.Fatal("failed atomic XOR mutated graph")
	}
	clone, err := graph.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clone.AddAND(x, y); err != nil {
		t.Fatal(err)
	}
	if original, _ := graph.InternalCount(); original != 1 {
		t.Fatal(original)
	}
	if err = graph.SetRoot(DecoEdge{Node: 99}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = graph.Evaluate([]bool{true}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = graph.WriteNode(2, DecoNode{Edge0: DecoEdge{Node: 2}, Edge1: y}); !errors.Is(err, ErrDimension) {
		t.Fatalf("forward edge %v", err)
	}
	if err = graph.WriteNode(0, DecoNode{Level: 1 << 14}); !errors.Is(err, ErrDimension) {
		t.Fatalf("level bitfield %v", err)
	}
	graph.Close()
	graph.Close()
	if _, err = graph.LeafCount(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	var nilGraph *DecoGraph
	if _, err = nilGraph.Leaves(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	calls := 0
	cancelled, _ := NewDecoGraph(2, DecoGraphLimits{Stop: func() bool { calls++; return calls >= 2 }})
	if _, err = cancelled.AddAND(x, y); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel mutation %v", err)
	}
	if count, _ := cancelled.InternalCount(); count != 0 {
		t.Fatal("cancelled mutation was not atomic")
	}
	composite, _ := NewDecoGraph(3, DecoGraphLimits{MaxNodes: 20})
	compositeBefore, _ := composite.Internals()
	compositeCalls := 0
	composite.limits.Stop = func() bool {
		compositeCalls++
		return compositeCalls == 3
	}
	if _, err = composite.AddMUX(DecoEdge{Node: 0}, DecoEdge{Node: 1}, DecoEdge{Node: 2}, false); !errors.Is(err, ErrCancelled) {
		t.Fatalf("composite cancellation %v", err)
	}
	composite.limits.Stop = nil
	compositeAfter, _ := composite.Internals()
	if !slices.Equal(compositeBefore, compositeAfter) {
		t.Fatal("cancelled composite mutation was not atomic")
	}
	limited, _ := NewDecoGraph(3, DecoGraphLimits{MaxNodes: 5})
	limitedBefore, _ := limited.Internals()
	if _, err = limited.AddMUX(DecoEdge{Node: 0}, DecoEdge{Node: 1}, DecoEdge{Node: 2}, true); !errors.Is(err, ErrLimit) {
		t.Fatalf("composite limit %v", err)
	}
	limitedAfter, _ := limited.Internals()
	if !slices.Equal(limitedBefore, limitedAfter) {
		t.Fatal("limited composite mutation was not atomic")
	}
	evaluate, _ := NewLeafDecoGraph(0, 1, false, DecoGraphLimits{})
	evaluate.limits.Stop = func() bool { return true }
	if _, err = evaluate.Evaluate([]bool{true}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("evaluation cancellation %v", err)
	}
}

func runDecoModel(data []byte) error {
	const variables = 4
	graph, err := NewDecoGraph(variables, DecoGraphLimits{MaxNodes: 2048})
	if err != nil {
		return err
	}
	patterns := make([]uint16, variables)
	for variable := 0; variable < variables; variable++ {
		var pattern uint16
		for assignment := 0; assignment < 1<<variables; assignment++ {
			if assignment&(1<<uint(variable)) != 0 {
				pattern |= 1 << uint(assignment)
			}
		}
		patterns[variable] = pattern
	}
	edges := []DecoEdge{{Node: 0}, {Node: 1}, {Node: 2}, {Node: 3}}
	edgePatterns := slices.Clone(patterns)
	for offset := 0; offset+3 < len(data) && len(edges) < 300; offset += 4 {
		firstIndex := int(data[offset]) % len(edges)
		secondIndex := int(data[offset+1]) % len(edges)
		thirdIndex := int(data[offset+2]) % len(edges)
		first, second, third := edges[firstIndex], edges[secondIndex], edges[thirdIndex]
		firstToggle := data[offset]&0x80 != 0
		secondToggle := data[offset+1]&0x80 != 0
		thirdToggle := data[offset+2]&0x80 != 0
		if firstToggle {
			first = first.Not()
		}
		if secondToggle {
			second = second.Not()
		}
		if thirdToggle {
			third = third.Not()
		}
		pattern := func(index int, toggle bool) uint16 {
			value := edgePatterns[index]
			if toggle {
				return ^value
			}
			return value
		}
		a, b, c := pattern(firstIndex, firstToggle), pattern(secondIndex, secondToggle), pattern(thirdIndex, thirdToggle)
		var edge DecoEdge
		switch data[offset+3] % 4 {
		case 0:
			edge, err = graph.AddAND(first, second)
			edgePatterns = append(edgePatterns, a&b)
		case 1:
			edge, err = graph.AddOR(first, second)
			edgePatterns = append(edgePatterns, a|b)
		case 2:
			edge, err = graph.AddXOR(first, second, data[offset+3]&0x80 != 0)
			edgePatterns = append(edgePatterns, a^b)
		case 3:
			edge, err = graph.AddMUX(first, second, third, data[offset+3]&0x80 != 0)
			edgePatterns = append(edgePatterns, (a&b)|(^a&c))
		}
		if err != nil {
			return err
		}
		edges = append(edges, edge)
		if err = graph.SetRoot(edge); err != nil {
			return err
		}
		for assignment := 0; assignment < 1<<variables; assignment++ {
			got, evaluateErr := graph.Evaluate(decoInputs(assignment, variables))
			if evaluateErr != nil {
				return evaluateErr
			}
			want := edgePatterns[len(edgePatterns)-1]&(1<<uint(assignment)) != 0
			if got != want {
				return ErrInvariant
			}
		}
	}
	return nil
}

func TestDecoIndependentModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	data := make([]byte, 80)
	for execution := 0; execution < 5000; execution++ {
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := runDecoModel(data); err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
	}
}

func FuzzDecoGraphModel(f *testing.F) {
	f.Add([]byte{0, 1, 2, 0, 3, 2, 1, 1})
	f.Add([]byte{0x80, 0x81, 2, 2, 3, 2, 1, 0x83})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			return
		}
		if err := runDecoModel(data); err != nil {
			t.Fatal(err)
		}
	})
}
