package gia

import "testing"

func TestForceGraphConstruction(t *testing.T) {
	for _, cluster := range []bool{false, true} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, y)
		var p *Frc_Man_t
		if cluster {
			p = Frc_ManStart(g)
		} else {
			p = Frc_ManStartSimple(g)
		}
		if Frc_ManCiNum(p) != 3 || Frc_ManCoNum(p) != 1 || Frc_ObjIsCi(Frc_ManCi(p, 0)) == 0 {
			t.Fatal("invalid interface")
		}
		root := Frc_ObjFanin(Frc_ManCo(p, 0), 0)
		want := int32(2)
		if cluster {
			want = 3
		}
		if Frc_ObjFaninNum(root) != want {
			t.Fatalf("cluster=%v fanins=%d", cluster, Frc_ObjFaninNum(root))
		}
		Frc_ManStop(p)
	}
}

func TestForcePlacementAndCrossCut(t *testing.T) {
	g := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	Gia_ManAppendCo(g, x)
	p := Frc_ManStartSimple(g)
	order := Frc_ManCollectCos(p)
	if Frc_ManCrossCut(p, order, 0) <= 0 || Frc_ManCrossCut(p, order, 1) <= 0 {
		t.Fatal("missing cross cut")
	}
	Frc_ManPlacementRefine(p, 3, 0)
	seen := make([]bool, p.nObjs)
	for _, o := range p.objects {
		if o.pPlace < 0 || o.pPlace >= p.nObjs || seen[o.pPlace] {
			t.Fatalf("invalid placement %d", o.pPlace)
		}
		seen[o.pPlace] = true
	}
	Frc_ManStop(p)
}
