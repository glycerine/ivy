package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestChoiceLevelsWithSibling(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	deep := Gia_ManAppendAnd(p, x, c)
	shallow := Gia_ManAppendAnd(p, a, c)
	Gia_ManAppendCo(p, shallow)
	p.pSibls = make([]int32, p.nObjs)
	p.pSibls[shallow>>1] = deep >> 1
	if got := Gia_ManChoiceLevel(p); got != 2 || Gia_ObjLevelId(p, shallow>>1) != 2 {
		t.Fatalf("level=%d", got)
	}
}
func TestChoiceLevelsWithBox(t *testing.T) {
	p := Gia_ManStart(8)
	a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, root)
	tm := tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(tm, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(tm, 0, 1, 1, 1, 0, 0)
	p.PManTime = tm
	if got := Gia_ManChoiceLevel(p); got != 2 {
		t.Fatalf("level=%d", got)
	}
}
