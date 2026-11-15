package gia

import "testing"

func cofReviewMuxGraph() *Gia_Man_t {
	g := Gia_ManStart(16)
	c, a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendMux(g, c, a, b))
	return g
}

func TestCofactorStructuralMuxReferences(t *testing.T) {
	g := cofReviewMuxGraph()
	defer Gia_ManStop(g)
	p := Cof_ManCreateLogicSimple(g)
	defer Cof_ManStop(p)
	control := p.objects[Gia_ManCi(g, 0).id]
	if control.nFanoutsM != 1 {
		t.Fatalf("structural mux control refs%d want1", control.nFanoutsM)
	}
	for i := int32(1); i < 3; i++ {
		if p.objects[Gia_ManCi(g, i).id].nFanoutsM != 0 {
			t.Fatal("data input incorrectly counted as a control")
		}
	}
}

func TestCofactorFanoutAccessorTracksReferences(t *testing.T) {
	g := cofReviewMuxGraph()
	defer Gia_ManStop(g)
	p := Cof_ManCreateLogicSimple(g)
	defer Cof_ManStop(p)
	root := p.objects[Gia_ObjFaninId0(Gia_ManCo(g, 0), Gia_ManCo(g, 0).id)]
	before := make([]int32, len(p.objects))
	for i, obj := range p.objects {
		before[i] = Cof_ObjFanoutNum(obj)
	}
	if removed := Cof_NodeDeref_rec(root); removed != 3 {
		t.Fatalf("deref removed%d want3", removed)
	}
	if got := Cof_ObjFanoutNum(root); got != 0 {
		t.Errorf("dereferenced root fanouts%d want0", got)
	}
	for _, child := range root.fanins {
		if got := Cof_ObjFanoutNum(child); got != 0 {
			t.Errorf("dereferenced child%d fanouts%d want0", child.Id, got)
		}
	}
	if restored := Cof_NodeRef_rec(root); restored != 3 {
		t.Fatalf("ref restored%d want3", restored)
	}
	for i, obj := range p.objects {
		if got := Cof_ObjFanoutNum(obj); got != before[i] {
			t.Fatalf("object%d refs%d want%d", i, got, before[i])
		}
	}
}

func TestCofactorMuxReferencesKeepTypedNodesAtomic(t *testing.T) {
	for _, kind := range []string{"xor root", "xor branch", "real mux"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(24)
			defer Gia_ManStop(g)
			c, d, r := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			left, right := Gia_ManAppendAnd(g, c, r), Gia_ManAppendAnd(g, c^1, d)
			var root int32
			want := [3]uint32{}
			switch kind {
			case "xor root":
				root = Gia_ManAppendXorReal(g, left^1, right^1)
			case "xor branch":
				left = Gia_ManAppendXorReal(g, c, r)
				root = Gia_ManAppendAnd(g, left^1, right^1)
			case "real mux":
				g.EnableMuxes()
				root = Gia_ManAppendMuxReal(g, d, left^1, right^1)
				want[1] = 1
			}
			Gia_ManAppendCo(g, root)
			p := Cof_ManCreateLogicSimple(g)
			defer Cof_ManStop(p)
			for i, expected := range want {
				if got := p.objects[Gia_ManCi(g, int32(i)).id].nFanoutsM; got != expected {
					t.Fatalf("CI%d muxrefs%d want%d", i, got, expected)
				}
			}
		})
	}
}
