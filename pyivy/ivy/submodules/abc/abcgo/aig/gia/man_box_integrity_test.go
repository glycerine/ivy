package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestGiaCarryOuts(t *testing.T) {
	g := Gia_ManStart(8)
	Gia_ManAppendCi(g)
	box := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, box)
	Gia_ManAppendCo(g, box)
	tm := tim.Tim_ManStart(2, 2)
	tim.Tim_ManCreateBox(tm, 0, 1, 1, 1, 0, 0)
	g.PManTime = tm
	if ids := vec.Vec_IntArray(Gia_ManComputeCarryOuts(g)); !reflect.DeepEqual(ids, []int32{box >> 1}) {
		t.Fatal(ids)
	}
	Gia_ManCheckIntegrityWithBoxes(g)
	if g.Refs != nil {
		t.Fatal("references retained")
	}
}
