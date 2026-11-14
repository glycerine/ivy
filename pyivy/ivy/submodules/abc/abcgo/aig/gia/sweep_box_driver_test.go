package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/proof/dch"
	"testing"
)

func TestSweepWithBoxesCombinational(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	Gia_ManAppendCo(p, b)
	Gia_ManAppendCo(p, root)
	timing := tim.Tim_ManStart(3, 2)
	tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, -1, 0)
	p.PManTime = timing
	extra := Gia_ManStart(5)
	e := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, e^1)
	p.PAigExtra = extra
	params := dch.DefaultParams()
	q := Gia_ManSweepWithBoxes(p, &params, nil, 0, 0, 0, 0)
	if q == nil || Gia_ManBoxNum(q) != 0 || Gia_ManCiNum(q) != 2 || Gia_ManCoNum(q) != 1 {
		t.Fatal("unused box was not removed")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && !in[1]) {
			t.Fatalf("input=%v output=%v err=%v", in, out, err)
		}
	}
}

type noDomainEquivalences struct{ calls int }

func (e *noDomainEquivalences) SweepCorrespondence(p *Gia_Man_t) {
	e.calls++
	if Gia_ManRegNum(p) != 2 {
		panic("wrong domain register count")
	}
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
}

func TestSweepWithBoxesMultipleDomains(t *testing.T) {
	p := Gia_ManStart(24)
	a := Gia_ManAppendCi(p)
	r0, r1, r2, r3 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	box := Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	for _, r := range []int32{r0, r1, r2, r3} {
		root = Gia_ManAppendAnd(p, root, r)
	}
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, root)
	Gia_ManAppendCo(p, r0)
	Gia_ManAppendCo(p, r1)
	Gia_ManAppendCo(p, r2)
	Gia_ManAppendCo(p, r3)
	timing := tim.Tim_ManStart(6, 6)
	tim.Tim_ManCreateBox(timing, 0, 1, 5, 1, -1, 0)
	p.PManTime = timing
	extra := Gia_ManStart(5)
	e := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, e)
	p.PAigExtra = extra
	p.vRegClasses = Vec_IntAlloc(4)
	for _, class := range []int32{1, 1, 2, 2} {
		Vec_IntPush(p.vRegClasses, class)
	}
	engine := new(noDomainEquivalences)
	q := Gia_ManSweepWithBoxes(p, nil, engine, 0, 0, 0, 0)
	if q == nil || engine.calls != 2 {
		t.Fatalf("result=%v domain calls=%d", q != nil, engine.calls)
	}
	if Gia_ManRegBoxNum(q) != 4 {
		t.Fatalf("register count %d", Gia_ManRegBoxNum(q))
	}
	if Gia_ManBoxNum(q) != 0 {
		t.Fatal("identity box should be removed")
	}
	for mask := 0; mask < 32; mask++ {
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		out, err := q.Evaluate(inputs)
		if err != nil || len(out) != 5 {
			t.Fatalf("input=%v output=%v err=%v", inputs, out, err)
		}
		if out[0] != (inputs[0] && inputs[1] && inputs[2] && inputs[3] && inputs[4]) {
			t.Fatalf("bad output %v", out)
		}
		for i := 0; i < 4; i++ {
			if out[i+1] != inputs[i+1] {
				t.Fatalf("bad register output %v", out)
			}
		}
	}
}
