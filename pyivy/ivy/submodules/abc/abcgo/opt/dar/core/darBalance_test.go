package core
import (
	"context"
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestBalanceDeepCone(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 8)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	root := inputs[0]
	for _, e := range inputs[1:] {
		root = g.AND(root, e)
	}
	g.AddOutput(root)
	before := g.Node(root).Level
	balanced, err := Balance(context.Background(), g, true)
	if err != nil {
		t.Fatal(err)
	}
	output := balanced.Node(balanced.Outputs()[0]).Fanins[0]
	if balanced.Node(output).Level != 3 {
		t.Fatalf("depth got %d, was %d", balanced.Node(output).Level, before)
	}
	for n := 0; n < 256; n++ {
		values := make([]bool, 8)
		for i := range values {
			values[i] = n&(1<<i) != 0
		}
		a, _ := g.Evaluate(values)
		b, _ := balanced.Evaluate(values)
		if a[0] != b[0] {
			t.Fatal(n)
		}
	}
	if g.Node(root).Level != before {
		t.Fatal("input mutated")
	}
}
func TestBalanceLiterals(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	if Dar_BalanceBuildSuper(g, []aig.Edge{a, a, a.Not(), b}) != 1 {
		t.Fatal("contradiction")
	}
	if Dar_BalanceBuildSuper(g, []aig.Edge{a, a, 0}) != a {
		t.Fatal("idempotence")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g.AddOutput(g.AND(a, b))
	if _, err := Balance(ctx, g, true); err != context.Canceled {
		t.Fatal(err)
	}
}
func TestBalanceNativeXor(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 8)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	root := inputs[0]
	for _, e := range inputs[1:] {
		root = g.NewXor(root, e)
	}
	g.AddOutput(root.Not())
	result, err := Balance(nil, g, true)
	if err != nil {
		t.Fatal(err)
	}
	driver := result.Node(result.Outputs()[0]).Fanins[0]
	if result.Node(driver).Level != 3 {
		t.Fatal("XOR not balanced", result.Node(driver).Level)
	}
	for n := 0; n < 256; n++ {
		values := make([]bool, 8)
		for i := range values {
			values[i] = n&(1<<i) != 0
		}
		a, _ := g.Evaluate(values)
		b, _ := result.Evaluate(values)
		if a[0] != b[0] {
			t.Fatal(n)
		}
	}
	if balanceBuildKind(g, []aig.Edge{inputs[0], inputs[0].Not()}, aig.ExclusiveOr) != 0 {
		t.Fatal("XOR phase cancellation")
	}
}
