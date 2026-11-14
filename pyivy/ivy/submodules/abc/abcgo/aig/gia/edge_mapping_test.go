package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEdgeMappingDelaysAndWindowSwap(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.vMapping2 = vec.Vec_WecStart(int(p.nObjs))
	p.vFanouts2 = vec.Vec_WecStart(int(p.nObjs))
	for _, pair := range [][2]int32{{x >> 1, a >> 1}, {x >> 1, b >> 1}, {y >> 1, x >> 1}, {y >> 1, c >> 1}} {
		vec.Vec_WecPush(p.vMapping2, int(pair[0]), pair[1])
		vec.Vec_WecPush(p.vFanouts2, int(pair[1]), pair[0])
	}
	p.VEdge1 = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge2 = vec.Vec_IntStart(int(p.nObjs))
	if got := Edg_ManEvalEdgeDelay(p); got != 2 {
		t.Fatalf("forward delay = %d", got)
	}
	if got := Edg_ManEvalEdgeDelayR(p); got != 1 {
		t.Fatalf("reverse delay = %d", got)
	}
	levels := vec.Vec_WecStart(0)
	Edg_ManCollectCritEdges(p, levels, 2)
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(levels, 1)); len(got) != 1 || got[0] != x>>1 {
		t.Fatalf("first critical level = %v", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(levels, 2)); len(got) != 1 || got[0] != y>>1 {
		t.Fatalf("second critical level = %v", got)
	}
	Gia_ObjEdgeAdd(x>>1, y>>1, p.VEdge1, p.VEdge2)
	Gia_ObjEdgeAdd(y>>1, x>>1, p.VEdge1, p.VEdge2)
	if got := Edg_ManEvalEdgeDelay(p); got != 1 {
		t.Fatalf("assigned forward delay = %d", got)
	}
	if got := Edg_ManEvalEdgeDelayR(p); got != 0 {
		t.Fatalf("assigned reverse delay = %d", got)
	}
	nodes := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(nodes, x>>1)
	window := vec.Vec_WecStart(1)
	vec.Vec_WecPush(window, 0, c>>1)
	Gia_ManUpdateMapping(p, nodes, window)
	if got := vec.Vec_IntArray(Gia_ObjLutFanins2(p, x>>1)); len(got) != 1 || got[0] != c>>1 {
		t.Fatalf("swapped mapping = %v", got)
	}
	Gia_ManUpdateMapping(p, nodes, window)
	if got := vec.Vec_IntArray(Gia_ObjLutFanins2(p, x>>1)); len(got) != 2 || got[0] != a>>1 || got[1] != b>>1 {
		t.Fatalf("restored mapping = %v", got)
	}
}

func TestEdgeAssignmentOptimizer(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.VMapping = vec.Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id     int32
		fanins []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		vec.Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(vec.Vec_IntSize(p.VMapping)))
		vec.Vec_IntPush(p.VMapping, int32(len(lut.fanins)))
		for _, fanin := range lut.fanins {
			vec.Vec_IntPush(p.VMapping, fanin)
		}
		vec.Vec_IntPush(p.VMapping, lut.id)
	}
	if got := Edg_ManAssignEdgeNew(p, 2, 0); got != 1 {
		t.Fatalf("optimized delay = %d", got)
	}
	if Gia_ObjCheckEdge(p, x>>1, y>>1) == 0 {
		t.Fatal("optimizer omitted critical edge")
	}
	if got := Edg_ManEvalEdgeDelay(p); got != 1 {
		t.Fatalf("restored assignment delay = %d", got)
	}
}
