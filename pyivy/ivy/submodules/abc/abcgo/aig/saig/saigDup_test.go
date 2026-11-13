package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"testing"
)

func dupFixture() *aig.Graph {
	p := aig.NewGraph()
	a := p.NewInput()
	q := p.NewInput()
	p.AddOutput(p.AND(a, q))
	p.AddOutput(p.NewXor(a, q))
	aig.Aig_ManSetRegNum(p, 1)
	return p
}
func dupInts(a ...int32) *vec.Vec_Int_t {
	v := vec.Vec_IntAlloc(len(a))
	for _, x := range a {
		vec.Vec_IntPush(v, x)
	}
	return v
}
func TestDupPhaseAndCones(t *testing.T) {
	p := dupFixture()
	q := Saig_ManDupWithPhase(p, dupInts(1))
	for a := 0; a < 2; a++ {
		for state := 0; state < 2; state++ {
			x, _ := p.Evaluate([]bool{a != 0, state != 0})
			y, _ := q.Evaluate([]bool{a != 0, state == 0})
			if x[0] != y[0] || x[1] == y[1] {
				t.Fatal("phase coordinates changed behavior")
			}
		}
	}
	cone := Saig_ManDupCones(p, []int32{0}, 1)
	if cone.RegisterCount() != 1 || len(cone.Inputs()) != 2 {
		t.Fatal("lost transitive register")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		a, _ := p.Evaluate(in)
		b, _ := cone.Evaluate(in)
		if a[0] != b[0] || a[1] != b[1] {
			t.Fatal("cone mismatch")
		}
	}
}
func TestDupOrAndEquivalence(t *testing.T) {
	p := dupFixture()
	q := Saig_ManDupOrpos(p)
	for m := 0; m < 4; m++ {
		in := []bool{m&1 != 0, m&2 != 0}
		a, _ := p.Evaluate(in)
		b, _ := q.Evaluate(in)
		if a[0] != b[0] || a[1] != b[1] {
			t.Fatal("OR mismatch")
		}
	}
	ins := p.Inputs()
	r := Saig_ManCreateEquivMiter(p, dupInts(int32(ins[0].NodeID()), int32(ins[1].NodeID())), 1)
	if r.RegisterCount() != 1 || len(r.Outputs()) != 2 {
		t.Fatal("miter interface")
	}
}
func TestDupAbstractionAndTrim(t *testing.T) {
	p := dupFixture()
	a := Saig_ManDupAbstraction(p, dupInts())
	if a.RegisterCount() != 0 || len(a.Inputs()) != 2 {
		t.Fatal("abstract state not PI")
	}
	b := Saig_ManDupAbstraction(p, dupInts(0))
	if b.RegisterCount() != 1 || len(b.InputNumbers) != 2 {
		t.Fatal("retained state")
	}
	q := aig.NewGraph()
	x := q.NewInput()
	q.NewInput()
	q.AddOutput(x)
	aig.Aig_ManSetRegNum(q, 0)
	q.InputNumbers = []int32{7, 9}
	r := Saig_ManTrimPis(q)
	if len(r.Inputs()) != 1 || r.InputNumbers[0] != 7 {
		t.Fatal("trim/provenance")
	}
	s := aig.NewGraph()
	s.NewInput()
	s.NewInput()
	s.AddOutput(s.Const0())
	aig.Aig_ManSetRegNum(s, 0)
	if len(Saig_ManTrimPis(s).Inputs()) != 2 {
		t.Fatal("all-unused source special case")
	}
}
func TestDupCounterexampleReplay(t *testing.T) {
	p := dupFixture()
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.IPo = 0
	c.PData[0] = 6 // initial0, inputs1,1 => state1 at frame1
	if Saig_ManVerifyCexNoClear(p, c) != 1 || !p.Node(p.Outputs()[0]).MarkB {
		t.Fatal("failure not replayed")
	}
	state := Saig_ManReturnFailingState(p, c, 0)
	next := Saig_ManReturnFailingState(p, c, 1)
	if vec.Vec_IntEntry(state, 0) != 1 || vec.Vec_IntEntry(next, 0) != 0 {
		t.Fatal("state timing")
	}
	extended := Saig_ManExtendCex(p, c)
	if extended.NRegs != 0 || extended.NPis != 2 || extended.PData[0] != 13 {
		t.Fatalf("extended %v", extended)
	}
	if Saig_ManFindFailedPoCex(p, c) != 0 || Saig_ManVerifyCex(p, c) != 1 {
		t.Fatal("failing PO")
	}
	for _, n := range p.Nodes() {
		if n.MarkB {
			t.Fatal("marks leaked")
		}
	}
}

func TestDupRandomCounterexamples(t *testing.T) {
	r := rand.New(rand.NewSource(148))
	for trial := 0; trial < 200; trial++ {
		p := aig.NewGraph()
		es := []aig.Edge{p.Const1()}
		for i := 0; i < 5; i++ {
			es = append(es, p.NewInput())
		}
		for i := 0; i < 25; i++ {
			a, b := es[r.Intn(len(es))], es[r.Intn(len(es))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			if r.Intn(2) == 0 {
				b = b.Not()
			}
			if i%3 == 0 {
				es = append(es, p.NewXor(a, b))
			} else {
				es = append(es, p.AND(a, b))
			}
		}
		for i := 0; i < 4; i++ {
			p.AddOutput(es[r.Intn(len(es))])
		}
		aig.Aig_ManSetRegNum(p, 2)
		c := util.Abc_CexAlloc(2, 3, 7)
		c.IFrame = 6
		c.IPo = 0
		for i := range c.PData {
			c.PData[i] = r.Uint32()
		}
		state := []bool{c.PData[0]&1 != 0, c.PData[0]&2 != 0}
		var expected []bool
		for frame := 0; frame < 7; frame++ {
			in := make([]bool, 5)
			for i := 0; i < 3; i++ {
				bit := 2 + frame*3 + i
				in[i] = c.PData[bit/32]&(1<<uint(bit%32)) != 0
			}
			copy(in[3:], state)
			expected, _ = p.Evaluate(in)
			if frame < 6 {
				copy(state, expected[2:])
			}
		}
		_, actualState, actual := replayCex(p, c)
		for i := range expected {
			if actual[i] != expected[i] {
				t.Fatalf("trial %d output %d", trial, i)
			}
		}
		for i := range state {
			if state[i] != actualState[i] {
				t.Fatal("state mismatch")
			}
		}
	}
}
func TestDupAbstractionCleanupProvenance(t *testing.T) {
	p := aig.NewGraph()
	a := p.NewInput()
	p.NewInput()
	p.NewInput()
	p.AddOutput(a)
	p.AddOutput(p.Const0())
	p.AddOutput(p.Const0())
	aig.Aig_ManSetRegNum(p, 2)
	q := Saig_ManDupAbstraction(p, dupInts(0, 1))
	if q.RegisterCount() != 0 || len(q.InputNumbers) != 1 || q.InputNumbers[0] != 0 {
		t.Fatalf("cleanup mapping %v", q.InputNumbers)
	}
}

func TestDupOrMultipleOutputsAndMiterTruth(t *testing.T) {
	p := aig.NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(a)
	p.AddOutput(b.Not())
	aig.Aig_ManSetRegNum(p, 0)
	q := Saig_ManDupOrpos(p)
	m := Saig_ManCreateEquivMiter(p, dupInts(int32(a.NodeID()), int32(b.NodeID())), 0)
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		out, _ := q.Evaluate(in)
		mit, _ := m.Evaluate(in)
		if out[0] != (in[0] || !in[1]) || mit[0] != (in[0] != in[1]) {
			t.Fatal("OR or miter function changed")
		}
	}
	aig.Aig_ManSetConstrNum(p, 1)
	if Saig_ManDupOrpos(p) != nil || Saig_ManCreateEquivMiter(p, dupInts(), 0) != nil {
		t.Fatal("constraints accepted by non-constraint miter")
	}
}
