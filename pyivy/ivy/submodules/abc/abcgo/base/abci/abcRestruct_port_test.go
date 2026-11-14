package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRestructNodeDivisorsRequiresBothFaninsAndClearsMarks(t *testing.T) {
	n := abc.NewNetwork("divisors")
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	ab := n.NewObject(abc.ObjectNode)
	ac := n.NewObject(abc.ObjectNode)
	if err := n.AddFanin(ab, a.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := n.AddFanin(ab, b.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := n.AddFanin(ac, a.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := n.AddFanin(ac, c.Signal()); err != nil {
		t.Fatal(err)
	}
	po := n.NewObject(abc.ObjectPO)
	if err := n.AddFanin(po, ab.Signal()); err != nil {
		t.Fatal(err)
	}
	// A fanout whose two fanins are already selected is collected once.
	shared := n.NewObject(abc.ObjectNode)
	if err := n.AddFanin(shared, ab.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := n.AddFanin(shared, a.Signal()); err != nil {
		t.Fatal(err)
	}

	manager := &Abc_ManRst_t{vLeaves: vec.Vec_PtrAlloc(2), vDecs: vec.Vec_PtrAlloc(4)}
	vec.Vec_PtrPush(manager.vLeaves, a)
	vec.Vec_PtrPush(manager.vLeaves, b)
	Abc_RestructNodeDivisors(manager, shared, 0)

	got := make([]*abc.Object, vec.Vec_PtrSize(manager.vDecs))
	for i := range got {
		got[i], _ = vec.Vec_PtrEntry(manager.vDecs, i).(*abc.Object)
	}
	want := []*abc.Object{a, b, ab, shared}
	if len(got) != len(want) {
		t.Fatalf("divisor count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("divisor[%d] = %p, want %p", i, got[i], want[i])
		}
	}
	for _, node := range []*abc.Object{a, b, c, ab, ac, shared, po} {
		if node.MarkC {
			t.Fatalf("temporary mark leaked on object %d", node.ID)
		}
	}
}

func TestNtkRestructureRunsFourLeafDsdPass(t *testing.T) {
	n := abc.NewNetwork("restructure-four-leaves")
	aig := abc.Abc_AigAlloc(n)
	n.FunctionManager = aig
	a, b, c, d := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	d.SetName("d")
	ab := abc.Abc_AigAnd(aig, a.Signal(), b.Signal())
	cd := abc.Abc_AigAnd(aig, c.Signal(), d.Signal())
	root := abc.Abc_AigAnd(aig, ab, cd)
	po := n.NewObject(abc.ObjectPO)
	po.SetName("f")
	if err := n.AddFanin(po, root); err != nil {
		t.Fatal(err)
	}

	if status := Abc_NtkRestructure(n, 4, 1, 0, 0); status != 1 {
		t.Fatalf("restructure status = %d", status)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("restructured network failed its integrity check")
	}
}

func TestNtkRestructureZeroGainCanReplaceBalancedAig(t *testing.T) {
	n := abc.NewNetwork("restructure-or4")
	aig := abc.Abc_AigAlloc(n)
	n.FunctionManager = aig
	a, b, c, d := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	for i, input := range []*abc.Object{a, b, c, d} {
		input.SetName(string(rune('a' + i)))
	}
	or := func(left, right abc.Edge) abc.Edge {
		return abc.Abc_AigAnd(aig, left.Not(), right.Not()).Not()
	}
	// The deliberately unbalanced OR4 has the same AIG size as its balanced
	// DSD realization; fUseZeros permits that zero-gain restructuring.
	root := or(or(or(a.Signal(), b.Signal()), c.Signal()), d.Signal())
	po := n.NewObject(abc.ObjectPO)
	po.SetName("f")
	if err := n.AddFanin(po, root); err != nil {
		t.Fatal(err)
	}
	if status := Abc_NtkRestructure(n, 4, 1, 1, 0); status != 1 {
		t.Fatalf("restructure status = %d", status)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("restructured network failed its integrity check")
	}
	for assignment := 0; assignment < 16; assignment++ {
		inputs := map[*abc.Object]bool{a: assignment&1 != 0, b: assignment&2 != 0, c: assignment&4 != 0, d: assignment&8 != 0}
		got, err := n.Evaluate(po.Signal(), inputs)
		if err != nil || got != (inputs[a] || inputs[b] || inputs[c] || inputs[d]) {
			t.Fatalf("assignment %#x: got=%t err=%v", assignment, got, err)
		}
	}
}

func TestRestructMffcSimulationUsesComplementedFanins(t *testing.T) {
	n := abc.NewNetwork("mffc-simulation")
	a, b := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	and := n.NewObject(abc.ObjectNode)
	if err := n.AddFanin(and, a.Signal().Not()); err != nil {
		t.Fatal(err)
	}
	if err := n.AddFanin(and, b.Signal()); err != nil {
		t.Fatal(err)
	}
	manager := &Abc_ManRst_t{vDecs: vec.Vec_PtrAlloc(3)}
	vec.Vec_PtrPush(manager.vDecs, a)
	vec.Vec_PtrPush(manager.vDecs, b)
	vec.Vec_PtrPush(manager.vDecs, and)
	randoms, sims := vec.Vec_IntAlloc(2), vec.Vec_IntAlloc(3)
	var random0, random1 uint32 = 0xaaaa5555, 0xcccc3333
	vec.Vec_IntPush(randoms, int32(random0))
	vec.Vec_IntPush(randoms, int32(random1))
	Abc_NodeMffcSimulate(manager.vDecs, 2, randoms, sims)
	if got, want := uint32(vec.Vec_IntEntry(sims, 2)), ^uint32(0xaaaa5555)&uint32(0xcccc3333); got != want {
		t.Fatalf("simulated AIG word = %#08x, want %#08x", got, want)
	}
}

func TestRestructRandomWordsMatchGlibcRandSequence(t *testing.T) {
	rng := newGlibcRand(1)
	first := restructRandomWords(rng)
	second := restructRandomWords(rng)
	if got, want := uint32(first[0]), uint32(0xb100f869); got != want {
		t.Fatalf("first simulation word = %#08x, want %#08x", got, want)
	}
	if got, want := uint32(first[1]), uint32(0x678c4cff); got != want {
		t.Fatalf("second simulation word = %#08x, want %#08x", got, want)
	}
	// Each vector consumes exactly 60 libc rand() values. The first word of
	// the next vector therefore starts at output 61, not at a reseeded stream.
	if got, want := uint32(second[0]), uint32(0x89a85ddc); got != want {
		t.Fatalf("next-vector first word = %#08x, want %#08x", got, want)
	}
}
