package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSplitMappingRoundTrip(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.vMapping, int(x>>1), int32(p.nObjs))
	vec.Vec_IntPush(p.vMapping, 2)
	vec.Vec_IntPush(p.vMapping, a>>1)
	vec.Vec_IntPush(p.vMapping, b>>1)
	vec.Vec_IntPush(p.vMapping, x>>1)
	p.vMapping2 = Spl_ManToWecMapping(p)
	p.vMapping = Spl_ManFromWecMapping(p, p.vMapping2)
	got := Gia_ObjLutFanins(p, x>>1)
	if len(got) != 2 || got[0] != a>>1 || got[1] != b>>1 {
		t.Fatalf("round trip %v", got)
	}
}

func TestSplitManagerMappingLifecycle(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.vMapping, int(x>>1), int32(p.nObjs))
	vec.Vec_IntPush(p.vMapping, 2)
	vec.Vec_IntPush(p.vMapping, a>>1)
	vec.Vec_IntPush(p.vMapping, b>>1)
	vec.Vec_IntPush(p.vMapping, x>>1)
	m := Spl_ManAlloc(p, 8, 0)
	if p.vMapping != nil || p.vMapping2 == nil || p.vFanout == nil {
		t.Fatal("window setup failed")
	}
	Spl_ManStop(m)
	if p.vMapping == nil || p.vMapping2 != nil || p.vFanout != nil || Gia_ObjLutSize(p, x>>1) != 2 {
		t.Fatal("window cleanup failed")
	}
}
