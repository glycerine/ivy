package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

type testPinTiming struct{}

func (testPinTiming) VariablePinDelays() bool { return true }
func (testPinTiming) PinDelay(inputs, pin int32) float32 {
	if pin == 0 {
		return 1
	}
	return 4
}
func TestLutPinTimingAndRequired(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.vMapping, int(x>>1), int32(p.nObjs))
	for _, v := range []int32{2, a >> 1, b >> 1, x >> 1} {
		vec.Vec_IntPush(p.vMapping, v)
	}
	Gia_ManTimeStart(p)
	defer Gia_ManTimeStop(p)
	Gia_ObjSetTimeArrival(p, a>>1, 1)
	Gia_ObjSetTimeArrival(p, b>>1, 3)
	p.pLutLib = testPinTiming{}
	if Gia_ObjComputeArrival(p, x>>1, 1) != 5 || Gia_ObjComputeArrival(p, x>>1, 0) != 7 {
		t.Fatal("incorrect LUT arrival")
	}
	order := make([]int32, 2)
	delays := make([]float32, 2)
	Gia_LutDelayTraceSortPins(p, x>>1, order, delays)
	if Gia_LutWhereIsPin(p, x>>1, b>>1, order) != 0 {
		t.Fatal("incorrect pin permutation")
	}
	Gia_ObjSetTimeRequired(p, x>>1, 10)
	Gia_ObjPropagateRequired(p, x>>1, 1)
	if Gia_ObjTimeRequired(p, a>>1) != 6 || Gia_ObjTimeRequired(p, b>>1) != 9 {
		t.Fatal("incorrect required times")
	}
}
