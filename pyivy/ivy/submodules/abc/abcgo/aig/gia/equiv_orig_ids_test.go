package gia

import "testing"

func TestOriginalIDEquivalencePairs(t *testing.T) {
	pairs := permutation(5, 3, 4, 5, 2, 4, 3, 2)
	m := Gia_ManOrigIdsRemapPairs(pairs, 7)
	for _, id := range []int{3, 4, 5} {
		if got := Vec_IntEntry(m, id); got != 2 {
			t.Fatalf("representative[%d]=%d", id, got)
		}
	}
	if Vec_IntEntry(m, 2) != -1 || Vec_IntEntry(m, 6) != -1 {
		t.Fatalf("unexpected mapping %v", Vec_IntArray(m))
	}
}
func TestOriginalIDReduction(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	q := Gia_ManOrigIdsReduce(p, permutation(x>>1, y>>1))
	if Gia_ObjRepr(p, y>>1) != x>>1 || p.PNexts[x>>1] != y>>1 {
		t.Fatal("equivalence class missing")
	}
	out, err := q.Evaluate([]bool{true, true})
	if err != nil || !out[0] || !out[1] {
		t.Fatalf("reduced outputs %v %v", out, err)
	}
}

func TestEquivalenceInspectionAndChoiceLevels(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, y)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	if Gia_ManEquivCountClasses(p) != 1 || Gia_ManEquivCountLitsAll(p) != 1 || Gia_ManEquivCountOne(p, x>>1) != 2 {
		t.Fatal("incorrect equivalence counts")
	}
	if Gia_ManCheckTopoOrder(p) != 0 {
		t.Fatal("unvisited representative should fail topology check")
	}
	mapping := Gia_ManChoiceMinLevel(p)
	if Vec_IntEntry(mapping, int(x>>1)) != x>>1 || Gia_ObjLevelId(p, y>>1) != 1 {
		t.Fatalf("choice mapping %v", Vec_IntArray(mapping))
	}
}

func TestEquivalenceReduction(t *testing.T) {
	makeGraph := func() *Gia_Man_t {
		p := Gia_ManStart(12)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
		Gia_ManAppendCo(p, y)
		p.PReprs = make([]Gia_Rpr_t, p.nObjs)
		for i := range p.PReprs {
			p.PReprs[i].iRepr = GIA_VOID
		}
		Gia_ObjSetRepr(p, y>>1, x>>1)
		Gia_ObjSetProved(p, y>>1)
		p.PNexts = Gia_ManDeriveNexts(p)
		return p
	}
	for _, mode := range []int{0, 1} {
		p := makeGraph()
		var q *Gia_Man_t
		if mode == 0 {
			q = Gia_ManEquivReduce(p, 0, 0, 0, 0)
		} else {
			q = Gia_ManEquivReduce2(p)
		}
		if q == nil {
			t.Fatal("missing reduced graph")
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			got, err := q.Evaluate(in)
			if err != nil || got[0] != (in[0] && in[1]) {
				t.Fatalf("mode %d inputs %v outputs %v error %v", mode, in, got, err)
			}
		}
	}
}

func TestEquivalenceRemapDFS(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManEquivRemapDfs(p)
	if Gia_ManEquivCountClasses(q) != 1 || Gia_ManEquivCountLitsAll(q) != 1 {
		t.Fatal("equivalence class lost during DFS remapping")
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && in[1]) || out[1] != out[0] {
			t.Fatalf("remapped output %v %v", out, err)
		}
	}
}

func TestSpeculativeEquivalenceReduction(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, c)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	trace, mapping := Vec_IntAlloc(1), Vec_IntAlloc(1)
	q := Gia_ManSpecReduceTrace(p, trace, mapping)
	if Gia_ManPoNum(q) != 3 || Vec_IntSize(trace) != 1 || Vec_IntEntry(trace, 0) != 1 || Vec_IntEntry(mapping, 0) != y>>1 {
		t.Fatalf("unexpected speculative trace: %v map %v", Vec_IntArray(trace), Vec_IntArray(mapping))
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := q.Evaluate(in)
		xVal, yVal := in[0] && in[1], in[0] && in[2]
		if err != nil || got[0] != xVal || got[1] != xVal || got[2] != (xVal != yVal) {
			t.Fatalf("inputs %v outputs %v error %v", in, got, err)
		}
	}
}

func TestSpeculativeInitializedFrames(t *testing.T) {
	p := Gia_ManStart(10)
	a, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, ro)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 1)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, ro>>1, 0)
	p.PNexts = Gia_ManDeriveNexts(p)
	init := &Abc_Cex_t{NRegs: 1, PData: []uint32{1}}
	q := Gia_ManSpecReduceInit(p, init, 2, 0)
	if q == nil || Gia_ManPoNum(q) != 2 || Gia_ManPiNum(q) != 2 {
		t.Fatal("incorrect initialized speculative interface")
	}
	for _, input := range []bool{false, true} {
		got, err := q.Evaluate([]bool{input, false})
		if err != nil || !got[0] || got[1] != input {
			t.Fatalf("input %v outputs %v error %v", input, got, err)
		}
	}
	if Vec_IntSize(&p.vCopies) != 0 {
		t.Fatal("frame copies retained")
	}
}

func TestEquivalenceTransferAfterReordering(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManChangeOrder(p)
	Gia_ManTransferEquivs(p, q)
	if Gia_ManEquivCountClasses(q) != 1 || Gia_ManEquivCountLitsAll(q) != 1 {
		t.Fatal("class lost in reorder")
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && in[1]) || out[1] != out[0] {
			t.Fatalf("reordered output %v %v", out, err)
		}
	}
}

func TestEquivalenceToChoices(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	ab := Gia_ManAppendAnd(p, a, b)
	x := Gia_ManAppendAnd(p, ab, c)
	bc := Gia_ManAppendAnd(p, b, c)
	y := Gia_ManAppendAnd(p, a, bc)
	Gia_ManAppendCo(p, y)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManEquivToChoices(p, 1)
	if Gia_ManCountChoices(q) != 1 {
		t.Fatalf("choice count = %d", Gia_ManCountChoices(q))
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && in[1] && in[2]) {
			t.Fatalf("choice graph output %v %v", out, err)
		}
	}
}

func TestEquivalenceFilterDisprovedOutput(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, c)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	Gia_ManEquivFilter(p, permutation(2), 0)
	if Gia_ObjHasRepr(p, y>>1) != 0 || Gia_ManEquivCountClasses(p) != 0 {
		t.Fatal("disproved equivalence retained")
	}
}
func TestEquivalenceTransformRemovesSharedColors(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, b>>1, a>>1)
	Gia_ObjSetRepr(p, c>>1, a>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	Gia_ObjSetColor(p, a>>1, 0)
	Gia_ObjSetColors(p, b>>1)
	Gia_ObjSetColor(p, c>>1, 1)
	Gia_ManEquivTransform(p, 0)
	if Gia_ObjHasRepr(p, b>>1) != 0 || Gia_ObjRepr(p, c>>1) != a>>1 || p.PNexts[a>>1] != c>>1 {
		t.Fatal("class was not filtered correctly")
	}
}

func TestEquivalenceFilterByLatches(t *testing.T) {
	p := Gia_ManStart(10)
	a, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 2)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, r0>>1, a>>1)
	Gia_ObjSetRepr(p, r1>>1, a>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	Gia_ManFilterEquivsUsingLatches(p, 1, 0, 0)
	if Gia_ObjHasRepr(p, r0>>1) != 0 || Gia_ObjRepr(p, r1>>1) != r0>>1 || p.PNexts[r0>>1] != r1>>1 {
		t.Fatal("latch-only class incorrect")
	}
}
func TestCombSpecReduceComparison(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, c)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManCombSpecReduce(p)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != ((in[0] && in[1]) != (in[0] && in[2])) {
			t.Fatalf("comparison output %v error %v", out, err)
		}
	}
}

func TestIndependentEquivalenceCandidates(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, c)
	z := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, z)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ObjSetRepr(p, y>>1, x>>1)
	Gia_ObjSetRepr(p, z>>1, x>>1)
	p.PNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManDup(p)
	if Cec4_ManMarkIndependentClasses(p, q) != 1 {
		t.Fatal("independent alternative not found")
	}
	if p.PReprs[x>>1].fColorA != 1 || p.PReprs[y>>1].fColorA != 1 || p.PReprs[z>>1].fColorA != 0 {
		t.Fatal("incorrect independent class marking")
	}
}
