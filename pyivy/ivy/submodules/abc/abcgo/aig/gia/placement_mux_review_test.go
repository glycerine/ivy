package gia

import "testing"

func placementMuxReviewGraph() (*Gia_Man_t, int32) {
	g := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	control := Gia_ManAppendAnd(g, c, d)
	g.EnableMuxes()
	root := Gia_ManAppendMuxReal(g, control, a, b)
	Gia_ManAppendCo(g, root)
	return g, root >> 1
}

func TestEmbeddingPlacementKeepsMuxControlCone(t *testing.T) {
	for _, cluster := range []bool{false, true} {
		g, root := placementMuxReviewGraph()
		start := Emb_ManStartSimple
		want := 3
		if cluster {
			start, want = Emb_ManStart, 4
		}
		p := start(g)
		obj := Emb_ManObj(p, Gia_ObjValue(Gia_ManObj(g, root)))
		if len(obj.fanins) != want {
			t.Fatalf("cluster=%t: MUX has %d fanins, want %d including control cone", cluster, len(obj.fanins), want)
		}
		Emb_ManSetValue(p)
		starts := Vec_IntAlloc(1)
		Vec_IntPush(starts, Emb_ManCo(p, 0).hHandle)
		Emb_ManFindDistances(p, starts, nil)
		for i := int32(0); i < 4; i++ {
			if Emb_ObjIsTravIdCurrent(p, Emb_ManCi(p, i)) == 0 {
				t.Fatalf("cluster=%t: control/data input %d is disconnected", cluster, i)
			}
		}
	}
}

func TestForcePlacementKeepsMuxControlCone(t *testing.T) {
	for _, cluster := range []bool{false, true} {
		g, root := placementMuxReviewGraph()
		start := Frc_ManStartSimple
		want := 3
		if cluster {
			start, want = Frc_ManStart, 4
		}
		p := start(g)
		obj := Frc_ManObj(p, Gia_ObjValue(Gia_ManObj(g, root)))
		if len(obj.fanins) != want {
			t.Fatalf("cluster=%t: MUX has %d fanins, want %d including control cone", cluster, len(obj.fanins), want)
		}
		seen := make(map[*Frc_Obj_t]bool)
		var visit func(*Frc_Obj_t)
		visit = func(o *Frc_Obj_t) {
			if seen[o] {
				return
			}
			seen[o] = true
			for _, fanin := range o.fanins {
				visit(fanin)
			}
		}
		visit(Frc_ManCo(p, 0))
		for i := int32(0); i < 4; i++ {
			if !seen[Frc_ManCi(p, i)] {
				t.Fatalf("cluster=%t: control/data input %d is disconnected", cluster, i)
			}
		}
		Frc_ManPlacementRefine(p, 3, 0)
		positions := make([]bool, p.nObjs)
		for _, node := range p.objects {
			if node.pPlace < 0 || node.pPlace >= p.nObjs || positions[node.pPlace] {
				t.Fatal("MUX placement is not a permutation")
			}
			positions[node.pPlace] = true
		}
	}
}
