package res

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
)

func TestResubstituteNetworkPreservesTruthAndNames(t *testing.T) {
	source := abc.NewNetwork("source")
	source.Type = uint8(abc.ABC_NTK_STRASH)
	source.Function = uint8(abc.ABC_FUNC_AIG)
	manager := abc.Abc_AigAlloc(source)
	source.FunctionManager = manager
	a := source.NewObject(abc.ObjectPI)
	b := source.NewObject(abc.ObjectPI)
	c := source.NewObject(abc.ObjectPI)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")

	ab := abc.Abc_AigAnd(manager, a.Signal(), b.Signal())
	left := abc.Abc_AigAnd(manager, ab, c.Signal())
	bc := abc.Abc_AigAnd(manager, b.Signal(), c.Signal())
	right := abc.Abc_AigAnd(manager, a.Signal(), bc)
	addNetworkOutput(t, source, left, "left")
	addNetworkOutput(t, source, right, "right")

	beforeNodes := len(source.ObjectsOfKind(abc.ObjectNode))
	result, stats, err := ResubstituteNetwork(source, GraphParams{MaxSupport: 6, MaxDivisors: 32})
	if err != nil {
		t.Fatal(err)
	}
	if len(source.ObjectsOfKind(abc.ObjectNode)) != beforeNodes {
		t.Fatal("ResubstituteNetwork mutated its input network")
	}
	if got := result.Name; got != "source" {
		t.Fatalf("network name %q, want %q", got, "source")
	}
	inputs := result.Inputs()
	outputs := result.Outputs()
	if len(inputs) != 3 || len(outputs) != 2 {
		t.Fatalf("result has %d inputs and %d outputs, want 3 and 2", len(inputs), len(outputs))
	}
	for i, want := range []string{"a", "b", "c"} {
		if got := inputs[i].Name(); got != want {
			t.Fatalf("input %d name %q, want %q", i, got, want)
		}
	}
	for i, want := range []string{"left", "right"} {
		if got := outputs[i].Name(); got != want {
			t.Fatalf("output %d name %q, want %q", i, got, want)
		}
	}
	if stats.AndsAfter > stats.AndsBefore {
		t.Fatalf("resubstitution increased AND count: %+v", stats)
	}
	for mask := 0; mask < 8; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		if got, want := evaluateNetworkOutputs(t, result, values), evaluateNetworkOutputs(t, source, values); !equalBoolSlice(got, want) {
			t.Fatalf("assignment %03b: got %v want %v", mask, got, want)
		}
	}
}

func TestResNetworkFromGraphEncodesExplicitXorWithoutDeadGate(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.NewXor(a.Not(), b))

	network, err := ResNetworkFromGraph(graph)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(network.ObjectsOfKind(abc.ObjectNode)); got != 3 {
		t.Fatalf("explicit XOR produced %d AND nodes, want 3", got)
	}
	inputs, output := network.Inputs(), network.Outputs()[0]
	for assignment := 0; assignment < 4; assignment++ {
		values := map[*abc.Object]bool{inputs[0]: assignment&1 != 0, inputs[1]: assignment&2 != 0}
		got, err := network.Evaluate(output.Fanins()[0], values)
		if err != nil {
			t.Fatal(err)
		}
		want := (assignment&1 == 0) != (assignment&2 != 0)
		if got != want {
			t.Fatalf("assignment %02b: got %v want %v", assignment, got, want)
		}
	}
}

func TestResGraphFromNetworkPreservesComplementedFanins(t *testing.T) {
	network := abc.NewNetwork("phased")
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	manager := abc.Abc_AigAlloc(network)
	network.FunctionManager = manager
	a := network.NewObject(abc.ObjectPI)
	b := network.NewObject(abc.ObjectPI)
	root := abc.Abc_AigAnd(manager, a.Signal().Not(), b.Signal())
	addNetworkOutput(t, network, root, "f")

	graph, err := ResGraphFromNetwork(network)
	if err != nil {
		t.Fatal(err)
	}
	inputs, outputs := graph.Inputs(), graph.Outputs()
	for assignment := 0; assignment < 4; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0}
		got, err := graph.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		want := !values[0] && values[1]
		if len(outputs) != 1 || got[0] != want {
			t.Fatalf("assignment %02b inputs=%v outputs=%v want %v", assignment, inputs, got, want)
		}
	}
}

func addNetworkOutput(t *testing.T, n *abc.Network, edge abc.Edge, name string) {
	t.Helper()
	output := n.NewObject(abc.ObjectPO)
	output.SetName(name)
	if err := n.AddFanin(output, edge); err != nil {
		t.Fatal(err)
	}
}

func evaluateNetworkOutputs(t *testing.T, n *abc.Network, values []bool) []bool {
	t.Helper()
	inputs := n.Inputs()
	assignment := make(map[*abc.Object]bool, len(inputs))
	for i, input := range inputs {
		assignment[input] = values[i]
	}
	outputs := n.Outputs()
	result := make([]bool, len(outputs))
	for i, output := range outputs {
		fanin := output.Fanins()[0]
		value, err := n.Evaluate(fanin, assignment)
		if err != nil {
			t.Fatal(err)
		}
		result[i] = value
	}
	return result
}

func equalBoolSlice(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
