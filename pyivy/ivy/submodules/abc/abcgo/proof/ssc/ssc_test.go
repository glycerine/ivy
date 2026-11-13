package ssc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
	"math/rand"
	"testing"
)

func fixture() (*gia.Gia_Man_t, *gia.Gia_Man_t) {
	p := newBuilder(3, 50)
	in := p.graph.InputLiterals()
	gia.Gia_ManAppendCo(p.graph, p.and(in[0], in[1]))
	gia.Gia_ManAppendCo(p.graph, p.xor(in[0], in[1]))
	gia.Gia_ManAppendCo(p.graph, p.xor(in[1], in[2]))
	c := newBuilder(3, 20)
	ci := c.graph.InputLiterals()
	gia.Gia_ManAppendCo(c.graph, c.xor(ci[0], ci[1]))
	return p.graph, c.graph
}
func checkCare(t *testing.T, a, b, c *gia.Gia_Man_t) {
	t.Helper()
	for mask := 0; mask < 1<<len(a.InputLiterals()); mask++ {
		in := make([]bool, len(a.InputLiterals()))
		for i := range in {
			in[i] = mask&(1<<i) != 0
		}
		care, err := c.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		bad := false
		for _, v := range care {
			bad = bad || v
		}
		if bad {
			continue
		}
		x, err := a.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		y, err := b.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(x) != len(y) {
			t.Fatal("interface")
		}
		for i := range x {
			if x[i] != y[i] {
				t.Fatalf("mask%d output%d changed", mask, i)
			}
		}
	}
}
func TestConstrainedSweeping(t *testing.T) {
	a, c := fixture()
	p := DefaultParams()
	p.Verify = true
	r := Ssc_PerformSweeping(a, c, &p)
	checkCare(t, a, r, c)
	if gia.Gia_ManAndNum(r) >= gia.Gia_ManAndNum(a) {
		t.Fatal("no actual reduction")
	}
	if Ssc_PerformVerification(a, r, c) != 1 {
		t.Fatal("verification")
	}
}
func TestUnsatisfiableAndZeroInputCare(t *testing.T) {
	b := newBuilder(0, 5)
	gia.Gia_ManAppendCo(b.graph, 1)
	care := newBuilder(0, 5)
	gia.Gia_ManAppendCo(care.graph, 1)
	r := Ssc_PerformSweeping(b.graph, care.graph, nil)
	v, _ := r.Evaluate(nil)
	if v[0] {
		t.Fatal("empty care should produce zero")
	}
	care2 := emptyCare(b.graph)
	r = Ssc_PerformSweeping(b.graph, care2, nil)
	v, _ = r.Evaluate(nil)
	if !v[0] {
		t.Fatal("constant output changed")
	}
}
func TestRandomSweeps(t *testing.T) {
	rng := rand.New(rand.NewSource(204))
	for trial := 0; trial < 150; trial++ {
		b := newBuilder(5, 100)
		es := append([]int32{0, 1}, b.graph.InputLiterals()...)
		for i := 0; i < 35; i++ {
			a, c := es[rng.Intn(len(es))], es[rng.Intn(len(es))]
			a ^= int32(rng.Intn(2))
			c ^= int32(rng.Intn(2))
			if i%3 == 0 {
				es = append(es, b.xor(a, c))
			} else {
				es = append(es, b.and(a, c))
			}
		}
		for i := 0; i < 4; i++ {
			gia.Gia_ManAppendCo(b.graph, es[rng.Intn(len(es))])
		}
		care := newBuilder(5, 20)
		ci := care.graph.InputLiterals()
		gia.Gia_ManAppendCo(care.graph, care.xor(ci[trial%5], ci[(trial+1)%5]))
		p := DefaultParams()
		p.SatVarMax = 1
		p.CallsRecycle = 0
		r := Ssc_PerformSweeping(b.graph, care.graph, &p)
		checkCare(t, b.graph, r, care.graph)
	}
}
func TestSATPatternAndClasses(t *testing.T) {
	a, c := fixture()
	p, status := Ssc_ManStart(a, c, nil)
	if status != satoko.SATOKO_SAT {
		t.Fatal(status)
	}
	in := a.InputLiterals()
	if Ssc_ManCheckEquivalence(p, in[0]>>1, in[1]>>1, 0) != satoko.SATOKO_UNSAT {
		t.Fatal("missed constrained equivalence")
	}
	if Ssc_ManCheckEquivalence(p, in[0]>>1, in[2]>>1, 0) != satoko.SATOKO_SAT {
		t.Fatal("false proof")
	}
	v, _ := c.Evaluate(p.Pattern)
	if v[0] || p.Pattern[0] == p.Pattern[2] {
		t.Fatal("invalid witness")
	}
}
func TestPackedSimulation(t *testing.T) {
	p, _ := fixture()
	s := NewSimulation(p, 1)
	for mask := 0; mask < 130; mask++ {
		pattern := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		Ssc_GiaSavePiPattern(s, pattern)
	}
	if s.Words != 4 {
		t.Fatal("pattern doubling")
	}
	Ssc_GiaSimRound(s)
	for bit := 0; bit < 130; bit++ {
		in := []bool{bit&1 != 0, bit&2 != 0, bit&4 != 0}
		v, _ := p.Evaluate(in)
		for i, l := range p.OutputLiterals() {
			x := s.Values[l>>1][bit/64]&(1<<uint(bit%64)) != 0
			if l&1 != 0 {
				x = !x
			}
			if x != v[i] {
				t.Fatal("packed simulation mismatch")
			}
		}
	}
}
func TestContainedAndRing(t *testing.T) {
	b := newBuilder(2, 20)
	in := b.graph.InputLiterals()
	gia.Gia_ManAppendCo(b.graph, b.and(in[0], in[1]))
	gia.Gia_ManAppendCo(b.graph, in[0])
	q := Gia_ManDropContained(b.graph)
	if len(q.OutputLiterals()) != 1 {
		t.Fatal("contained output retained")
	}
	r := Gia_ManOptimizeRing(b.graph)
	if len(r.OutputLiterals()) != 1 {
		t.Fatal("ring containment")
	}
}

func TestConstraintAppendAndSequentialOrdering(t *testing.T) {
	b := newBuilder(2, 30)
	a, q := b.graph.InputLiterals()[0], b.graph.InputLiterals()[1]
	gia.Gia_ManAppendCo(b.graph, b.and(a, q))
	gia.Gia_ManAppendCo(b.graph, b.xor(a, q))
	gia.Gia_ManAppendCo(b.graph, q^1)
	gia.Gia_ManSetRegNum(b.graph, 1)
	gia.Gia_ManSetConstrNum(b.graph, 1)
	params := DefaultParams()
	params.Append = true
	r := Ssc_PerformSweepingConstr(b.graph, &params)
	if r.RegisterCount() != 1 || gia.Gia_ManConstrNum(r) != 1 || len(r.OutputLiterals()) != 3 {
		t.Fatal("sequential/constraint ordering")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		a, _ := b.graph.Evaluate(in)
		z, _ := r.Evaluate(in)
		if a[1] != z[1] || !a[1] && (a[0] != z[0] || a[2] != z[2]) {
			t.Fatal("constraint or state transition changed")
		}
	}
	no := emptyCare(b.graph)
	r = Ssc_PerformSweepingConstr(no, &params)
	if gia.Gia_ManConstrNum(r) != 1 {
		t.Fatal("source append adds a trivial false constraint even for unconstrained inputs")
	}
}

func TestIndependentManagersDoNotRecycle(t *testing.T) {
	for run := 0; run < 8; run++ {
		t.Run(fmt.Sprint(run), func(t *testing.T) {
			t.Parallel()
			a, c := fixture()
			params := DefaultParams()
			params.SatVarMax = 1
			params.CallsRecycle = 0
			p, status := Ssc_ManStart(a, c, &params)
			if status != satoko.SATOKO_SAT {
				t.Fatal(status)
			}
			in := a.InputLiterals()
			Ssc_ManCheckEquivalence(p, in[0]>>1, in[2]>>1, 0)
			Ssc_ManCheckEquivalence(p, in[0]>>1, in[1]>>1, 0)
			if p.Stats.Recycles != 0 {
				t.Fatal("C SSC does not recycle its solver")
			}
		})
	}
}
func TestPublishedRepresentativesHavePivotPhase(t *testing.T) {
	a := newBuilder(1, 10)
	x := a.graph.InputLiterals()[0]
	gia.Gia_ManAppendCo(a.graph, x)
	care := newBuilder(1, 10)
	gia.Gia_ManAppendCo(care.graph, care.graph.InputLiterals()[0]^1)
	r := Ssc_PerformSweeping(a.graph, care.graph, nil)
	v, _ := r.Evaluate([]bool{true})
	if !v[0] || gia.Gia_ObjRepr(a.graph, x>>1) != 0 || gia.Gia_ObjPhase(gia.Gia_ManObj(a.graph, x>>1)) != 1 || gia.Gia_ObjNext(a.graph, 0) != 0 {
		t.Fatal("proved constant-one phase/class metadata lost")
	}
}
func TestConstraintSetterBounds(t *testing.T) {
	b := newBuilder(1, 5)
	gia.Gia_ManAppendCo(b.graph, 0)
	gia.Gia_ManSetConstrNum(b.graph, 1)
	if gia.Gia_ManConstrNum(b.graph) != 1 {
		t.Fatal("constraint setter")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("invalid constraint count accepted")
		}
	}()
	gia.Gia_ManSetConstrNum(b.graph, 2)
}

func TestRingPreservesOutputUnion(t *testing.T) {
	b := newBuilder(2, 15)
	in := b.graph.InputLiterals()
	gia.Gia_ManAppendCo(b.graph, b.and(in[0], in[1]))
	gia.Gia_ManAppendCo(b.graph, b.and(in[0]^1, in[1]))
	q := Gia_ManOptimizeRing(b.graph)
	if gia.Gia_ManConstrNum(q) != 0 {
		t.Fatal("temporary ring constraints leaked")
	}
	for mask := 0; mask < 4; mask++ {
		values, e := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if e != nil {
			t.Fatal(e)
		}
		union := false
		for _, v := range values {
			union = union || v
		}
		if union != (mask&2 != 0) {
			t.Fatal("ring changed union")
		}
	}
}
