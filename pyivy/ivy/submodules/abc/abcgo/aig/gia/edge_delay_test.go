package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEdgeDelayAssignmentAndEvaluation(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.vMapping2 = vec.Vec_WecStart(int(p.nObjs))
	for _, f := range []int32{a >> 1, b >> 1} {
		vec.Vec_WecPush(p.vMapping2, int(x>>1), f)
	}
	for _, f := range []int32{x >> 1, c >> 1} {
		vec.Vec_WecPush(p.vMapping2, int(y>>1), f)
	}
	if got := Gia_ManComputeEdgeDelay(p, 1); got != 12 {
		t.Fatalf("assigned delay %d", got)
	}
	if Gia_ManEvalEdgeCount(p) != 1 || Gia_ObjCheckEdge(p, x>>1, y>>1) == 0 {
		t.Fatal("critical edge missing")
	}
	if got := Gia_ManEvalEdgeDelay(p); got != 12 {
		t.Fatalf("evaluated delay %d", got)
	}
	if got := Gia_ManComputeEdgeDelay2(p); got != 1 {
		t.Fatalf("reverse assigned delay %d", got)
	}
	if Gia_ManEvalEdgeCount(p) != 1 || Gia_ObjCheckEdge(p, x>>1, y>>1) == 0 {
		t.Fatal("reverse assignment missing critical edge")
	}
}
func TestEdgeDelayWithTimingBox(t *testing.T) {
	p := Gia_ManStart(8)
	a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, root)
	timing := tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0)
	tim.Tim_ManInitPiArrival(timing, 0, 6)
	p.PManTime = timing
	p.vMapping2 = vec.Vec_WecStart(int(p.nObjs))
	vec.Vec_WecPush(p.vMapping2, int(root>>1), a>>1)
	vec.Vec_WecPush(p.vMapping2, int(root>>1), box>>1)
	if got := Gia_ManComputeEdgeDelay(p, 0); got != 16 {
		t.Fatalf("box edge delay %d", got)
	}
}
