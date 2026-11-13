package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/tim"
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

func TestBalanceTimingMetadataIsCopiedAndUpdatedOnDestination(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(a, b), c)
	g.AddOutput(root)
	manager := tim.Tim_ManStart(3, 1)
	tim.Tim_ManInitPiArrival(manager, 0, 4)
	tim.Tim_ManInitPiArrival(manager, 1, 1)
	tim.Tim_ManInitPiArrival(manager, 2, 0)
	tim.Tim_ManIncrementTravId(manager)
	tim.Tim_ManSetCoArrival(manager, 0, 99)
	g.Timing = manager

	balanced, err := Balance(context.Background(), g, false)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := balanced.Timing.(*tim.Tim_Man_t)
	if !ok || got == manager {
		t.Fatalf("timing manager was not duplicated: %T", balanced.Timing)
	}
	for i, want := range []int{4, 1, 0} {
		if level := balanced.Node(balanced.Inputs()[i]).Level; level != want {
			t.Fatalf("timed CI %d level=%d want %d", i, level, want)
		}
	}
	if gotTime := tim.Tim_ManGetCoArrival(got, 0); gotTime != float32(balanced.Node(balanced.Node(balanced.Outputs()[0]).Fanins[0]).Level) {
		t.Fatalf("destination CO arrival=%v", gotTime)
	}
	if sourceTime := tim.Tim_ManGetCoArrival(manager, 0); sourceTime != 99 {
		t.Fatalf("source timing metadata mutated: %v", sourceTime)
	}
}

func TestBalanceBarrierOrderAndInputLevel(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(g.AND(g.AND(a, b), c))
	g.BarrierBuffers = 1
	balanced, err := Balance(context.Background(), g, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := balanced.Node(balanced.Inputs()[3]).Level; got != 2 {
		t.Fatalf("barrier CI level=%d want %d", got, 2)
	}
	for mask := 0; mask < 16; mask++ {
		values := make([]bool, 4)
		for i := range values {
			values[i] = mask&(1<<i) != 0
		}
		before, _ := g.Evaluate(values)
		after, _ := balanced.Evaluate(values)
		if before[0] != after[0] || before[1] != after[1] {
			t.Fatalf("barrier output order/function changed for input mask %04b", mask)
		}
	}
}

func TestBalanceReusesPreviouslyBuiltSharedPair(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(g.AND(g.AND(a, c), g.AND(b, d)))

	balanced, err := Balance(context.Background(), g, false)
	if err != nil {
		t.Fatal(err)
	}
	shared := balanced.Node(balanced.Outputs()[0]).Fanins[0]
	var containsShared func(aig.Edge) bool
	containsShared = func(edge aig.Edge) bool {
		if edge.NodeID() == shared.NodeID() {
			return true
		}
		n := balanced.Node(edge)
		return (n.Kind == aig.And || n.Kind == aig.ExclusiveOr) && (containsShared(n.Fanins[0]) || containsShared(n.Fanins[1]))
	}
	second := balanced.Node(balanced.Outputs()[1]).Fanins[0]
	if !containsShared(second) {
		t.Fatal("balanced second cone did not reuse the first output's pair")
	}
}

func TestBalanceChecksDeadlineDuringLargeTraversal(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 260)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	root := inputs[0]
	for _, e := range inputs[1:] {
		root = g.AND(root, e)
	}
	g.AddOutput(root)
	g.Time2Quit = time.Now().Add(-time.Second)
	if _, err := Balance(context.Background(), g, false); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error=%v", err)
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
