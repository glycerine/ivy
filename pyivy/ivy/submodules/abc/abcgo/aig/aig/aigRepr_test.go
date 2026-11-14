package aig

import "testing"

func reprFixture() *Graph {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	inverse := Aig_ObjCreate(g, &Node{Kind: ExclusiveOr, Fanins: [2]Edge{g.Const1(), root}})
	g.AddOutput(inverse)
	Aig_ManReprStart(g, int32(len(g.nodes)))
	Aig_ObjCreateRepr(g, Aig_Regular(root), Aig_Regular(inverse))
	return g
}
func TestRepresentativeConstantAndChains(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a, b)
	Aig_ManReprStart(g, 100)
	Aig_ObjCreateRepr(g, g.Const1(), a)
	Aig_ObjCreateRepr(g, a, b)
	Aig_ObjCreateRepr(g, b, n)
	if got := Aig_ObjFindRepr(g, a); got != g.Const1() {
		t.Fatal("lost constant representative", got)
	}
	if got := Aig_ObjFindReprTransitive(g, n); got != g.Const1() {
		t.Fatal(got)
	}
	if Aig_ManCountReprs(g) != 3 {
		t.Fatal(Aig_ManCountReprs(g))
	}
	Aig_ObjClearRepr(g, b)
	if Aig_ObjFindRepr(g, b) != InvalidEdge {
		t.Fatal("not cleared")
	}
	Aig_ManReprStop(g)
	if Aig_ManCountReprs(g) != 0 {
		t.Fatal("not stopped")
	}
}
func TestRepresentativeDuplicationAndPhase(t *testing.T) {
	for _, ordered := range []int32{0, 1} {
		g := reprFixture()
		copy := Aig_ManDupRepr(g, ordered)
		equalDup(t, g, copy)
	}
}
func TestRepresentativeRehash(t *testing.T) {
	g := reprFixture()
	oracle := Aig_ManDupSimple(g)
	out := Aig_ManRehash(g)
	equalDup(t, oracle, out)
	if Aig_ManRemapRepr(out) != 0 {
		t.Fatal("critical represented fanouts remain")
	}
}
func TestRepresentativeTransfer(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	other := g.NewXor(a, g.AND(a, b.Not()))
	g.AddOutput(other)
	Aig_ManReprStart(g, int32(len(g.nodes)))
	Aig_ObjCreateRepr(g, root, other)
	q := Aig_ManDupSimple(g)
	Aig_ManReprStart(q, int32(len(q.nodes)))
	Aig_ManTransferRepr(q, g)
	if Aig_ManCountReprs(q) != 1 {
		t.Fatal(Aig_ManCountReprs(q))
	}
	target := NewGraph()
	for i := 1; i < len(g.nodes); i++ {
		target.NewInput()
	}
	Aig_ManReprStart(target, int32(len(target.nodes)))
	mapping := make([]int32, len(g.nodes))
	for i := range mapping {
		mapping[i] = int32(i)
	}
	if Aig_TransferMappedClasses(target, g, mapping) != 1 {
		t.Fatal("class transfer count")
	}
}
func TestRepresentativeChoiceValidity(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	other := g.NewXor(a, g.AND(a, b.Not()))
	g.AddOutput(root)
	Aig_ManReprStart(g, int32(len(g.nodes)))
	Aig_ObjCreateRepr(g, root, other)
	if Aig_ObjCheckTfi(g, other, root) != 0 {
		t.Fatal("false dependency")
	}
	Aig_ManMarkValidChoices(g)
	if got := g.equivs[root.NodeID()]; got != other {
		t.Fatal("missing choice", got)
	}
	if Aig_ObjCheckTfi(g, root, other) != 1 {
		t.Fatal("choice link not traversed")
	}
	g = reprFixture()
	Aig_ManMarkValidChoices(g)
	if Aig_ManCountReprs(g) != 0 {
		t.Fatal("retained cyclic or referenced choice")
	}
}
func TestRepresentativeCycleRejected(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	Aig_ManReprStart(g, 10)
	g.reprs[a.NodeID()] = b
	g.reprs[b.NodeID()] = a
	defer func() {
		if recover() == nil {
			t.Fatal("accepted representative cycle")
		}
	}()
	Aig_ObjFindReprTransitive(g, a)
}
func TestRepresentativeBasicRetainsRegisters(t *testing.T) {
	g := NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.nRegs = 1
	g.AddOutput(pi)
	g.AddOutput(g.AND(pi, lo))
	Aig_ManReprStart(g, int32(len(g.nodes)))
	q := Aig_ManDupReprBasic(g)
	if q.nRegs != 1 || q.InputCount() != 2 || q.OutputCount() != 2 {
		t.Fatal("register interface changed")
	}
	if q.outputDriver(q.outputs[1]) != q.Const0() {
		t.Fatal("unreachable latch not disconnected")
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestRepresentativeRemapAndConstantPhase(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	y := g.NewXor(a, g.AND(a, b.Not()))
	z := g.NewXor(b, g.AND(a.Not(), b))
	g.AddOutput(z)
	Aig_ManReprStart(g, int32(len(g.nodes)))
	Aig_ObjCreateRepr(g, x, y)
	Aig_ObjCreateRepr(g, y, z)
	if Aig_ManRemapRepr(g) != 1 || Aig_ObjFindRepr(g, z) != x {
		t.Fatal("representative chain not compressed")
	}
	equalDup(t, g, Aig_ManDupRepr(g, 0))
	constant := NewGraph()
	i := constant.NewInput()
	constant.AddOutput(i)
	Aig_ManReprStart(constant, 2)
	Aig_ObjCreateRepr(constant, constant.Const1(), i)
	q := Aig_ManDupRepr(constant, 0)
	for _, value := range []bool{false, true} {
		out, err := q.Evaluate([]bool{value})
		if err != nil || out[0] {
			t.Fatal("phase-correct constant representative", err, out)
		}
	}
}
func TestRepresentativeInvalidOperands(t *testing.T) {
	for _, kind := range []string{"decreasing", "complement", "duplicate start", "bad mapping", "bad mapped index"} {
		t.Run(kind, func(t *testing.T) {
			g := NewGraph()
			a, b := g.NewInput(), g.NewInput()
			Aig_ManReprStart(g, 3)
			defer func() {
				if recover() == nil {
					t.Fatal("accepted invalid state")
				}
			}()
			switch kind {
			case "decreasing":
				Aig_ObjCreateRepr(g, b, a)
			case "complement":
				Aig_ObjCreateRepr(g, a.Not(), b)
			case "duplicate start":
				Aig_ManReprStart(g, 3)
			case "bad mapping":
				Aig_ObjCreateRepr(g, a, b)
				Aig_TransferMappedClasses(g, g, []int32{0})
			case "bad mapped index":
				Aig_ObjCreateRepr(g, a, b)
				Aig_TransferMappedClasses(g, g, []int32{0, 100, 200})
			}
		})
	}
}

func TestRepresentativeCIAtGateFanout(t *testing.T) {
	for _, ordered := range []int32{0, 1} {
		g := NewGraph()
		a, b := g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(a, b))
		Aig_ManReprStart(g, int32(len(g.nodes)))
		Aig_ObjCreateRepr(g, a, b)
		q := Aig_ManDupRepr(g, ordered)
		out, err := q.Evaluate([]bool{true, false})
		if err != nil || !out[0] {
			t.Fatalf("ordered=%d: CI representative ignored at gate: %v %v", ordered, out, err)
		}
	}
}

func TestRepresentativeChainsPreserveAlternatingPhases(t *testing.T) {
	for _, ordered := range []int32{0, 1} {
		g := NewGraph()
		a, b := g.NewInput(), g.NewInput()
		root := g.AND(a, b.Not())
		inverse := Aig_ObjCreate(g, &Node{Kind: ExclusiveOr, Fanins: [2]Edge{g.Const1(), root}})
		positive := Aig_ObjCreate(g, &Node{Kind: ExclusiveOr, Fanins: [2]Edge{g.Const1(), inverse}})
		for _, e := range []Edge{root, inverse, positive, positive.Not()} {
			g.AddOutput(e)
		}
		Aig_ManReprStart(g, int32(len(g.nodes)))
		Aig_ObjCreateRepr(g, root, inverse)
		Aig_ObjCreateRepr(g, inverse, positive)
		copy := Aig_ManDupRepr(g, ordered)
		for mask := 0; mask < 4; mask++ {
			want := mask&1 != 0 && mask&2 == 0
			values, err := copy.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
			if err != nil || len(values) != 4 || values[0] != want || values[1] == want || values[2] != want || values[3] == want {
				t.Fatalf("ordered=%d mask=%d outputs=%v err=%v", ordered, mask, values, err)
			}
		}
		if err := copy.Check(); err != nil {
			t.Fatal(err)
		}
	}
}
