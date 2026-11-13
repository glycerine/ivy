package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// giaAig.c: Gia_ManReprFromAigRepr writes raw GIA object IDs into
// every non-CO AIG object's iData before transferring representatives.
func TestReprFromAigReprUpdatesSourceCopyData(t *testing.T) {
	a := aig.NewGraph()
	x, y := a.NewInput(), a.NewInput()
	a.AddOutput(y)
	aig.Aig_ManReprStart(a, aig.Aig_ManObjNumMax(a))
	aig.Aig_ObjCreateRepr(a, x, y)
	g := Gia_ManStart(8)
	gx, gy := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, gy)
	Gia_ObjSetValue(Gia_ManConst0(g), 0)
	Gia_ObjSetValue(Gia_ManObj(g, gx>>1), int32(x))
	Gia_ObjSetValue(Gia_ManObj(g, gy>>1), int32(y))
	Gia_ManReprFromAigRepr(a, g)
	if got := Gia_ObjRepr(g, gy>>1); got != gx>>1 {
		t.Fatalf("repr=%d", got)
	}
	if got := aig.Aig_ObjCopy(a, x); got != aig.Edge(gx>>1) {
		t.Fatalf("raw iData x=%d", got)
	}
	if got := aig.Aig_ObjCopy(a, y); got != aig.Edge(gy>>1) {
		t.Fatalf("raw iData y=%d", got)
	}
}
func TestReprFromAigReprRequiresSourceStorage(t *testing.T) {
	a := aig.NewGraph()
	a.NewInput()
	g := Gia_ManStart(4)
	Gia_ManAppendCi(g)
	defer func() {
		if got := recover(); got != "gia: missing AIG representative storage" {
			t.Fatalf("panic %v", got)
		}
	}()
	Gia_ManReprFromAigRepr(a, g)
}

func TestReprFromAigReprSourcePreconditions(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		setup      func() (*aig.Graph, *Gia_Man_t)
	}{
		{"existing GIA classes", "gia: representative storage already initialized", func() (*aig.Graph, *Gia_Man_t) {
			a := aig.NewGraph()
			a.NewInput()
			aig.Aig_ManReprStart(a, aig.Aig_ManObjNumMax(a))
			g := Gia_ManStart(4)
			Gia_ManAppendCi(g)
			g.StartRepresentatives()
			return a, g
		}},
		{"different object counts", "gia: AIG/GIA object counts differ", func() (*aig.Graph, *Gia_Man_t) {
			a := aig.NewGraph()
			aig.Aig_ManReprStart(a, aig.Aig_ManObjNumMax(a))
			g := Gia_ManStart(4)
			Gia_ManAppendCi(g)
			return a, g
		}},
		{"complemented copy", "gia: complemented GIA/AIG copy", func() (*aig.Graph, *Gia_Man_t) {
			a := aig.NewGraph()
			a.NewInput()
			aig.Aig_ManReprStart(a, aig.Aig_ManObjNumMax(a))
			g := Gia_ManStart(4)
			x := Gia_ManAppendCi(g)
			Gia_ObjSetValue(Gia_ManObj(g, x>>1), 3)
			return a, g
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, g := tc.setup()
			defer func() {
				if got := recover(); got != tc.want {
					t.Fatalf("panic %v want %s", got, tc.want)
				}
			}()
			Gia_ManReprFromAigRepr(a, g)
		})
	}
}
