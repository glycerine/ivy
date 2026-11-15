package gia

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
)

func TestCollapseTypedGatesHonorAllocationLimit(t *testing.T) {
	for _, kind := range []string{"xor", "mux"} {
		t.Run(kind, func(t *testing.T) {
			p := Gia_ManStart(16)
			a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			var root int32
			if kind == "xor" {
				root = Gia_ManAppendXorReal(p, a, b)
			} else {
				p.EnableMuxes()
				root = Gia_ManAppendMuxReal(p, a, b, c)
			}
			Gia_ManAppendCo(p, root)
			dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
			defer cudd.Cudd_Quit(dd)
			roots := Gia_ManCollapse(p, dd, 0, 0)
			if roots != nil {
				Gia_ManCollapseDeref(dd, roots)
				t.Fatal("fresh typed gate bypassed zero-node allocation limit")
			}
			if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
				t.Fatalf("failed collapse leaked %d roots", count)
			}
			roots = Gia_ManCollapse(p, dd, 100, 0)
			if roots == nil {
				t.Fatal("failed collapse prevented retry")
			}
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				want, err := p.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				got, err := dd.Evaluate(Vec_PtrEntry(roots, 0).(*cudd.DdNode), inputs)
				if err != nil || got != want[0] {
					t.Fatalf("assignment%d: %v %v, want%v", mask, got, err, want)
				}
			}
			Gia_ManCollapseDeref(dd, roots)
			roots = Gia_ManCollapse(p, dd, 0, 0)
			if roots == nil {
				t.Fatal("cached gate rejected despite requiring no allocation")
			}
			Gia_ManCollapseDeref(dd, roots)
			if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
				t.Fatalf("successful collapse leaked %d roots", count)
			}
		})
	}
}

func TestCollapseReleasesSimplifiedAndFailedFanins(t *testing.T) {
	for _, fail := range []bool{false, true} {
		p := Gia_ManStart(16)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		f := Gia_ManAppendAnd(p, a, b)
		root := Gia_ManAppendAnd(p, f, a^1)
		if fail {
			root = Gia_ManAppendAnd(p, root^1, Gia_ManAppendAnd(p, b, c))
		}
		Gia_ManAppendCo(p, root)
		dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
		before := []uint32{dd.Vars[0].Ref, dd.Vars[1].Ref, dd.Vars[2].Ref}
		// Prepopulate the first conjunction so limit zero can pass it, then fail
		// at the later independent conjunction. It has no external owner.
		if fail {
			cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
		}
		limit := int32(100)
		if fail {
			limit = 0
		}
		roots := Gia_ManCollapse(p, dd, limit, 0)
		if fail && roots != nil {
			t.Fatal("expected allocation failure")
		}
		if !fail && roots == nil {
			t.Fatal("unexpected collapse failure")
		}
		if roots != nil {
			Gia_ManCollapseDeref(dd, roots)
		}
		// Both successful and failed collapse must release temporary child edges.
		for i, n := range dd.Vars {
			if n.Ref != before[i] {
				t.Fatalf("failure%v projection%d retained temporary fanin references: %d want%d", fail, i, n.Ref, before[i])
			}
		}
		if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
			t.Fatalf("collapse failure%v leaked %d roots", fail, count)
		}
		cudd.Cudd_Quit(dd)
	}
}

func TestCollapseSharedOutputOwnershipAndReordering(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCi(p) // An unused projection must not acquire a reference.
	f := Gia_ManAppendAnd(p, a, b)
	buffer := Gia_ManAppendBuf(p, f)
	for _, lit := range []int32{f, f ^ 1, buffer, 0, 1} {
		Gia_ManAppendCo(p, lit)
	}
	dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	held := cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	cudd.Cudd_Ref(held)
	roots := Gia_ManCollapse(p, dd, 100, 0)
	if roots == nil || Vec_PtrSize(roots) != 5 {
		t.Fatal("shared collapse failed")
	}
	if got := cudd.Cudd_ExternalRefCount(dd, held); got != 4 {
		t.Fatalf("shared root owns %d references, want four", got)
	}
	if cudd.Cudd_ShuffleHeap(dd, []int{2, 0, 1}) == 0 {
		t.Fatal("reorder failed")
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want := []bool{inputs[0] && inputs[1], !(inputs[0] && inputs[1]), inputs[0] && inputs[1], false, true}
		for i, w := range want {
			got, err := dd.Evaluate(Vec_PtrEntry(roots, i).(*cudd.DdNode), inputs)
			if err != nil || got != w {
				t.Fatalf("assignment%d output%d: %v %v want%v", mask, i, got, err, w)
			}
		}
	}
	Gia_ManCollapseDeref(dd, roots)
	if got := cudd.Cudd_ExternalRefCount(dd, held); got != 1 {
		t.Fatalf("borrowed root damaged: references %d want one", got)
	}
	cudd.Cudd_RecursiveDeref(dd, held)
	if count := cudd.Cudd_CheckZeroRef(dd); count != 0 {
		t.Fatalf("shared collapse leaked %d roots", count)
	}
}

func TestCollapseEmptyAndConstantInterfaces(t *testing.T) {
	for _, outputs := range [][]int32{nil, {0}, {1}, {0, 1, 0, 1}} {
		p := Gia_ManStart(8)
		for _, lit := range outputs {
			Gia_ManAppendCo(p, lit)
		}
		dd := cudd.Cudd_Init(0, 0, 0, 0, 0)
		roots := Gia_ManCollapse(p, dd, 0, 0)
		if roots == nil || Vec_PtrSize(roots) != len(outputs) {
			t.Fatal("constant/empty collapse failed")
		}
		out := Gia_ManCollapseTest(p, 0)
		want := make([]bool, len(outputs))
		for i, lit := range outputs {
			want[i] = lit != 0
			got, err := dd.Evaluate(Vec_PtrEntry(roots, i).(*cudd.DdNode), nil)
			if err != nil || got != want[i] {
				t.Fatalf("constant output%d: %v %v", i, got, err)
			}
		}
		checkQuantificationOutputs(t, out, nil, want)
		Gia_ManCollapseDeref(dd, roots)
		cudd.Cudd_Quit(dd)
	}
}

func TestCollapseSharingSize(t *testing.T) {
	dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	a, b := dd.Vars[0], dd.Vars[1]
	f := cudd.Cudd_bddAnd(dd, a, b)
	g := cudd.Cudd_bddAnd(dd, a, cudd.Cudd_Not(b))
	for _, tc := range []struct {
		roots []*cudd.DdNode
		want  int
	}{
		{nil, 0},
		{[]*cudd.DdNode{dd.One, dd.LogicZero}, 1},
		{[]*cudd.DdNode{f, cudd.Cudd_Not(f), f}, 3},
		{[]*cudd.DdNode{f, g}, 4},
		{[]*cudd.DdNode{f, g, a}, 5},
	} {
		if got := giaClpSharingSize(tc.roots); got != tc.want {
			t.Fatalf("shared size %d want%d", got, tc.want)
		}
	}
}

func TestCollapseDerefWaitsForFinalFanout(t *testing.T) {
	p := Gia_ManStart(8)
	id := Gia_ManAppendCi(p) >> 1
	p.Refs = make([]int32, p.nObjs)
	p.Refs[id] = 2
	dd := cudd.Cudd_Init(2, 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	f := cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	cudd.Cudd_Ref(f)
	funcs := Vec_PtrStart(int(p.nObjs))
	Vec_PtrWriteEntry(funcs, int(id), f)
	Gia_ObjCollapseDeref(p, dd, funcs, id)
	if p.Refs[id] != 1 || Vec_PtrEntry(funcs, int(id)) != f || cudd.Cudd_ExternalRefCount(dd, f) != 1 {
		t.Fatal("nonfinal fanout released a live root")
	}
	Gia_ObjCollapseDeref(p, dd, funcs, id)
	if p.Refs[id] != 0 || Vec_PtrEntry(funcs, int(id)) != nil || cudd.Cudd_CheckZeroRef(dd) != 0 {
		t.Fatal("final fanout retained the root")
	}
	Gia_ManCollapseDeref(dd, funcs)
	Gia_ManCollapseDeref(dd, nil)
}

func TestCollapseGeneratedNames(t *testing.T) {
	for _, tc := range []struct {
		n           int32
		first, last string
	}{
		{0, "", ""}, {1, "a", "a"}, {25, "a", "y"}, {26, "a0", "z0"}, {27, "a0", "a1"}, {52, "a0", "z1"}, {260, "a0", "z9"}, {261, "a0", "a10"},
	} {
		names := Gia_GetFakeNames(tc.n)
		if Vec_PtrSize(names) != int(tc.n) {
			t.Fatal("name count")
		}
		seen := map[string]bool{}
		for i, v := range Vec_PtrArray(names) {
			name := v.(string)
			if seen[name] {
				t.Fatalf("duplicate generated name %q", name)
			}
			seen[name] = true
			if i == 0 && name != tc.first || i == int(tc.n)-1 && name != tc.last {
				t.Fatalf("n%d position%d: %q", tc.n, i, name)
			}
		}
	}
}
