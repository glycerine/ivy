package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaClassHelpers(t *testing.T) {
	p := gia.Gia_ManStart(10)
	gia.Gia_ManAppendCi(p)
	obj := gia.Gia_ManObj(p, 0)
	p.SetGateClass(0, 1)
	if Gia_ObjIsInGla(p, obj) != 1 {
		t.Fatal("initial class")
	}
	Gia_ObjRemFromGla(p, obj)
	if Gia_ObjIsInGla(p, obj) != 0 {
		t.Fatal("remove")
	}
	Gia_ObjAddToGla(p, obj)
	if Gia_ObjIsInGla(p, obj) != 1 {
		t.Fatal("restore")
	}
}

func TestShrinkGlaKeepsNecessaryInitialState(t *testing.T) {
	g := gia.Gia_ManStart(10)
	x := gia.Gia_ManAppendCi(g)
	q := gia.Gia_ManAppendCi(g)
	out := gia.Gia_ManAppendAnd(g, x, q)
	gia.Gia_ManAppendCo(g, out)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	for _, id := range []int32{0, q >> 1, out >> 1} {
		vec.Vec_IntWriteEntry(classes, int(id), 1)
	}
	gia.Gia_ManSetGateClasses(g, classes)
	if result := Gia_ManShrinkGla(g, 2, 0, 0, 0, 0, 0); result != nil {
		t.Fatal("C entry point returns nil")
	}
	if vec.Vec_IntEntry(gia.Gia_ManGateClasses(g), int(q>>1)) != 1 {
		t.Fatal("removed latch needed to keep frame zero safe")
	}
}

func TestShrinkGlaRemovesIrrelevantLatch(t *testing.T) {
	g := gia.Gia_ManStart(10)
	x := gia.Gia_ManAppendCi(g)
	q := gia.Gia_ManAppendCi(g)
	out := int32(0)
	gia.Gia_ManAppendCo(g, out)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	for _, id := range []int32{0, q >> 1, out >> 1} {
		vec.Vec_IntWriteEntry(classes, int(id), 1)
	}
	gia.Gia_ManSetGateClasses(g, classes)
	Gia_ManShrinkGla(g, 2, 0, 0, 0, 0, 0)
	if vec.Vec_IntEntry(gia.Gia_ManGateClasses(g), int(q>>1)) != 0 {
		t.Fatal("retained latch that cannot affect property")
	}
}
