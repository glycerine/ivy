package gia

import "testing"

func TestCec4ChoicesWithoutClasses(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	p.pNexts = make([]int32, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
	q := Gia_ManDup(p)
	result := Cec4_ManSatSolverChoices(p, q)
	for bits := 0; bits < 4; bits++ {
		got, err := result.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if err != nil || got[0] != (bits == 3) {
			t.Fatalf("bits=%d got=%v err=%v", bits, got, err)
		}
	}
}

func TestCec4ChoicesIndependentClassFailsExplicitly(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, Gia_ManAppendAnd(p, a, b), c)
	y := Gia_ManAppendAnd(p, a, Gia_ManAppendAnd(p, b, c))
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.pNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManDup(p)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || len(out) != 2 || out[0] != (bits == 7) || out[1] != (bits == 7) {
			t.Fatalf("bits=%d outputs=%v error=%v", bits, out, err)
		}
	}
	if Cec4_ManMarkIndependentClasses(p, q) != 1 {
		t.Fatal("expected independent equivalent alternatives")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("source builder must fail on uninitialized output driver")
		}
	}()
	Cec4_ManSatSolverChoices(p, q)
}

func TestCec4ChoicesDependentClassPreservesFunction(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	or := Gia_ManAppendAnd(p, a^1, b^1) ^ 1
	y := Gia_ManAppendAnd(p, x, or)
	Gia_ManAppendCo(p, y)
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.pNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManDup(p)
	if Cec4_ManMarkIndependentClasses(p, q) != 0 {
		t.Fatal("dependent alternative was marked independent")
	}
	result := Cec4_ManSatSolverChoices(p, q)
	for bits := 0; bits < 4; bits++ {
		out, err := result.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if err != nil || len(out) != 1 || out[0] != (bits == 3) {
			t.Fatalf("bits=%d outputs=%v error=%v", bits, out, err)
		}
	}
}

func TestCec4ChoiceAndBuilderSourceStub(t *testing.T) {
	if got := Cec4_ManSatSolverAnd_rec(nil, nil, nil, 0); got != 0 {
		t.Fatalf("stub returned %d", got)
	}
}

// The C choice-recursion path invokes the unfinished AND builder for each
// independent class member; each returned literal is zero.
func TestCec4ChoiceRecursionUsesSourceZeroStub(t *testing.T) {
	src := Gia_ManStart(8)
	Gia_ManAppendCi(src)
	Gia_ManAppendCi(src)
	src.StartRepresentatives()
	Gia_ObjSetRepr(src, 2, 1)
	Gia_ObjSetNext(src, 1, 2)
	src.pReprs[1].fColorA = 1
	src.pReprs[2].fColorA = 1
	dst := Gia_ManStart(8)
	dst.pSibls = make([]int32, 3)
	if got := Cec4_ManSatSolverChoices_rec(dst, src, nil, 1); got != 0 {
		t.Fatalf("class literal %d", got)
	}
	if got := Cec4_ManSatSolverChoices_rec(dst, src, nil, 0); got != 0 {
		t.Fatalf("nonclass literal %d", got)
	}
}
