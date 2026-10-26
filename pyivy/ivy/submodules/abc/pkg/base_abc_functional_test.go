package abc

// Tests for the Go port of src/base/abc/{abcFunc.c,abcAig.c,abcNetlist.c,abcMinBase.c}.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func abcFunctionalTestNetwork(t *testing.T, variables int, truth uint64) *ABCNetwork {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	inputs := make([]ABCObjectEdge, variables)
	for i := range inputs {
		id, createErr := network.CreatePI()
		if createErr != nil {
			t.Fatal(createErr)
		}
		inputs[i] = ABCObjectEdge{ID: id}
	}
	words := []uint64{truth}
	sop, err := ABCSOPFromTruth(variables, words, ABCMetadataOptions{})
	if err != nil {
		t.Fatal(err)
	}
	text, err := sop.Format(ABCMetadataOptions{})
	if err != nil {
		t.Fatal(err)
	}
	node, err := network.CreateLogicNode(text, inputs...)
	if err != nil {
		t.Fatal(err)
	}
	po, err := network.CreatePO()
	if err == nil {
		err = network.AddFanin(po, ABCObjectEdge{ID: node})
	}
	if err != nil {
		t.Fatal(err)
	}
	return network
}

func abcFunctionalIndependentEvaluate(t *testing.T, network *ABCNetwork, inputs []bool) []bool {
	t.Helper()
	if len(inputs) != len(network.pis) || len(network.boxes) != 0 {
		t.Fatal("independent evaluator dimension")
	}
	values := make([]bool, len(network.objects))
	known := make([]bool, len(network.objects))
	if len(network.objects) > 0 && network.objects[0].Alive && network.objects[0].Type == ABCObjectConst1 {
		values[0], known[0] = true, true
	}
	for i, id := range network.pis {
		values[id], known[id] = inputs[i], true
	}
	var evaluate func(int) bool
	evaluate = func(id int) bool {
		if known[id] {
			return values[id]
		}
		object := network.objects[id]
		switch object.Type {
		case ABCObjectNet, ABCObjectPO, ABCObjectBI:
			edge := object.Fanins[0]
			values[id] = evaluate(edge.ID) != edge.Complemented
		case ABCObjectNode:
			sop, err := ParseABCSOP(object.Function, ABCMetadataOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var assignment uint64
			for i, edge := range object.Fanins[:sop.Variables] {
				if evaluate(edge.ID) != edge.Complemented {
					assignment |= uint64(1) << uint(i)
				}
			}
			values[id], err = sop.Evaluate(assignment)
			if err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unsupported object type %v", object.Type)
		}
		known[id] = true
		return values[id]
	}
	outputs := make([]bool, len(network.pos))
	for i, id := range network.pos {
		outputs[i] = evaluate(id)
	}
	return outputs
}

func TestABCFunctionalSOPToAIGExhaustive(t *testing.T) {
	for variables := 0; variables <= 3; variables++ {
		functions := uint64(1) << (uint64(1) << uint(variables))
		for truth := uint64(0); truth < functions; truth++ {
			network := abcFunctionalTestNetwork(t, variables, truth)
			converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{})
			if err != nil {
				t.Fatalf("vars=%d truth=%x: %v", variables, truth, err)
			}
			for assignment := 0; assignment < 1<<variables; assignment++ {
				inputs := make([]bool, variables)
				for i := range inputs {
					inputs[i] = assignment&(1<<i) != 0
				}
				got, err := converted.Graph.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				want := truth&(uint64(1)<<uint(assignment)) != 0
				if len(got) != 1 || got[0] != want {
					t.Fatalf("vars=%d truth=%x assignment=%d got=%v want=%v", variables, truth, assignment, got, want)
				}
			}
		}
	}
	for _, truth := range []uint64{0, 1, 0xffff, 0x6996, 0xa55a, 0x8001} {
		network := abcFunctionalTestNetwork(t, 4, truth)
		converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
			got, _ := converted.Graph.Evaluate(inputs)
			want := truth&(uint64(1)<<uint(assignment)) != 0
			if len(got) != 1 || got[0] != want {
				t.Fatalf("truth=%x assignment=%d got=%v want=%v", truth, assignment, got, want)
			}
		}
	}
}

func TestABCFunctionalAIGRoundTripSequential(t *testing.T) {
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	state, _ := graph.CreateCI()
	next, _ := graph.Xor(a, state)
	out, _ := graph.Mux(b, next, state)
	if _, err := graph.CreateCO(out); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	network, err := AIGToABCNetwork(graph, ABCFunctionalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if converted.Graph.RegisterCount() != 1 {
		t.Fatalf("registers=%d", converted.Graph.RegisterCount())
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, _ := graph.Evaluate(inputs)
		got, _ := converted.Graph.Evaluate(inputs)
		if !slices.Equal(got, want) {
			t.Fatalf("assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
}

func TestABCFunctionalMinimumBase(t *testing.T) {
	network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	a, _ := network.CreatePI()
	b, _ := network.CreatePI()
	node, err := network.CreateLogicNode("1-1 1\n", ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b}, ABCObjectEdge{ID: a})
	if err != nil {
		t.Fatal(err)
	}
	po, _ := network.CreatePO()
	_ = network.AddFanin(po, ABCObjectEdge{ID: node})
	changed, err := ABCMinimizeNodeBase(network, node, ABCFunctionalOptions{})
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	object, _ := network.Object(node)
	if len(object.Fanins) != 1 || object.Fanins[0].ID != a {
		t.Fatalf("fanins=%v", object.Fanins)
	}
	sop, err := ParseABCSOP(object.Function, ABCMetadataOptions{})
	if err != nil || sop.Variables != 1 {
		t.Fatalf("sop=%+v err=%v", sop, err)
	}
	for assignment := uint64(0); assignment < 2; assignment++ {
		got, _ := sop.Evaluate(assignment)
		if got != (assignment == 1) {
			t.Fatalf("assignment=%d got=%v", assignment, got)
		}
	}
}

func TestABCFunctionalLogicNetlistRoundTrip(t *testing.T) {
	network := abcFunctionalTestNetwork(t, 3, 0x96)
	_ = network.SetObjectName(network.pis[0], "a")
	_ = network.SetObjectName(network.pis[1], "b")
	_ = network.SetObjectName(network.pis[2], "c")
	_ = network.SetObjectName(network.pos[0], "out")
	for id, object := range network.objects {
		if object.Alive && object.Type == ABCObjectNode {
			if err := network.SetObjectName(id, "logic_node"); err != nil {
				t.Fatal(err)
			}
		}
	}
	netlist, err := ABCLogicToNetlist(network, ABCFunctionalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	logic, err := ABCNetlistToLogic(netlist, ABCFunctionalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want := abcFunctionalIndependentEvaluate(t, network, inputs)
		got := abcFunctionalIndependentEvaluate(t, logic, inputs)
		if !slices.Equal(got, want) {
			t.Fatalf("assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	if netlist.Type != ABCNetworkNetlist || logic.Type != ABCNetworkLogic {
		t.Fatalf("types=%v/%v", netlist.Type, logic.Type)
	}
	for i, name := range []string{"a", "b", "c"} {
		if logic.objects[logic.pis[i]].Name != name {
			t.Fatalf("PI %d name=%q want=%q", i, logic.objects[logic.pis[i]].Name, name)
		}
	}
	foundNodeName := false
	for _, object := range logic.objects {
		foundNodeName = foundNodeName || object.Alive && object.Type == ABCObjectNode && object.Name == "logic_node"
	}
	if !foundNodeName {
		t.Fatal("logic node output-net name was not restored")
	}
}

func TestABCFunctionalRejectsMalformedLatchAndBoundsDuplicateTruth(t *testing.T) {
	malformed, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = malformed.CreateObject(ABCObjectLatch); err != nil {
		t.Fatal(err)
	}
	if _, err = ABCNetworkToAIG(malformed, ABCFunctionalOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("malformed latch err=%v", err)
	}

	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	edges := make([]ABCObjectEdge, 0, 22)
	for i := 0; i < 21; i++ {
		id, createErr := network.CreatePI()
		if createErr != nil {
			t.Fatal(createErr)
		}
		edges = append(edges, ABCObjectEdge{ID: id})
	}
	edges = append(edges, edges[0])
	node, err := network.CreateLogicNode("1111111111111111111111 1\n", edges...)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := network.Object(node)
	if _, err = ABCMinimizeNodeBase(network, node, ABCFunctionalOptions{MaxOperations: 100, MaxTruthVariables: 32}); !errors.Is(err, ErrLimit) {
		t.Fatalf("duplicate truth bound err=%v", err)
	}
	after, _ := network.Object(node)
	if !slices.Equal(before.Fanins, after.Fanins) || before.Function != after.Function {
		t.Fatal("bounded duplicate reduction mutated the node")
	}
}

func TestABCFunctionalSharedRootIsVisitedOnce(t *testing.T) {
	network := abcFunctionalTestNetwork(t, 2, 0x8)
	shared := network.objects[network.pos[0]].Fanins[0]
	second, err := network.CreatePO()
	if err == nil {
		err = network.AddFanin(second, shared)
	}
	if err != nil {
		t.Fatal(err)
	}
	converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(converted.CombinationalIDs) != 1 {
		t.Fatalf("combinational order=%v", converted.CombinationalIDs)
	}
	netlist, err := ABCLogicToNetlist(network, ABCFunctionalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	nodes := 0
	for _, object := range netlist.objects {
		if object.Alive && object.Type == ABCObjectNode {
			nodes++
		}
	}
	if nodes != 1 {
		t.Fatalf("netlist duplicated shared root: nodes=%d", nodes)
	}
}

func TestABCFunctionalCancellationAtomic(t *testing.T) {
	network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	a, _ := network.CreatePI()
	b, _ := network.CreatePI()
	node, _ := network.CreateLogicNode("1- 1\n", ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b})
	before := network.Clone()
	beforeObject, _ := before.Object(node)
	calls := 0
	_, err := ABCMinimizeNodeBase(network, node, ABCFunctionalOptions{Stop: func() bool { calls++; return true }})
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("err=%v", err)
	}
	afterObject, _ := network.Object(node)
	if !slices.Equal(afterObject.Fanins, beforeObject.Fanins) || afterObject.Function != beforeObject.Function || network.edges != before.edges {
		t.Fatal("cancelled minimum-base mutation was not atomic")
	}
}

func TestABCFunctionalNetworkBudgetIsCumulative(t *testing.T) {
	network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	a, _ := network.CreatePI()
	b, _ := network.CreatePI()
	first, _ := network.CreateLogicNode("1- 1\n", ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b})
	second, _ := network.CreateLogicNode("1- 1\n", ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b})
	beforeFirst, _ := network.Object(first)
	beforeSecond, _ := network.Object(second)
	if _, err := ABCMinimumBase(network, ABCFunctionalOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("err=%v", err)
	}
	afterFirst, _ := network.Object(first)
	afterSecond, _ := network.Object(second)
	if !slices.Equal(beforeFirst.Fanins, afterFirst.Fanins) || beforeFirst.Function != afterFirst.Function ||
		!slices.Equal(beforeSecond.Fanins, afterSecond.Fanins) || beforeSecond.Function != afterSecond.Function {
		t.Fatal("network-wide budget failure did not roll back")
	}
}

func TestABCFunctionalIndependentModel(t *testing.T) {
	rng := rand.New(rand.NewSource(0xabcf00))
	for trial := 0; trial < 5000; trial++ {
		variables := rng.Intn(5)
		bits := 1 << variables
		truth := rng.Uint64()
		if bits < 64 {
			truth &= uint64(1)<<uint(bits) - 1
		}
		network := abcFunctionalTestNetwork(t, variables, truth)
		converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{})
		if err != nil {
			t.Fatalf("trial=%d: %v", trial, err)
		}
		for assignment := 0; assignment < 1<<variables; assignment++ {
			inputs := make([]bool, variables)
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			got, _ := converted.Graph.Evaluate(inputs)
			want := abcFunctionalIndependentEvaluate(t, network, inputs)
			if !slices.Equal(got, want) {
				t.Fatalf("trial=%d assignment=%d got=%v want=%v", trial, assignment, got, want)
			}
		}
	}
}

func FuzzABCFunctionalRoundTrip(f *testing.F) {
	f.Add(byte(3), uint64(0x96))
	f.Add(byte(0), uint64(1))
	f.Add(byte(4), uint64(0xa55a))
	f.Fuzz(func(t *testing.T, rawVariables byte, truth uint64) {
		variables := int(rawVariables % 5)
		bits := 1 << variables
		if bits < 64 {
			truth &= uint64(1)<<uint(bits) - 1
		}
		network := abcFunctionalTestNetwork(t, variables, truth)
		converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{MaxOperations: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<variables; assignment++ {
			inputs := make([]bool, variables)
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			got, err := converted.Graph.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			want := truth&(uint64(1)<<uint(assignment)) != 0
			if len(got) != 1 || got[0] != want {
				t.Fatalf("assignment=%d got=%v want=%v", assignment, got, want)
			}
		}
	})
}
