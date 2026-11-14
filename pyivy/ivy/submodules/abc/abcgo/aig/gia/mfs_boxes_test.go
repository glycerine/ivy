package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestGiaMfsBoxGroups(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendCo(g, a)
	y := Gia_ManAppendCo(g, b)
	tm := tim.Tim_ManStart(2, 2)
	tim.Tim_ManCreateBox(tm, 0, 1, 1, 1, -1, 0)
	g.PManTime = tm
	copies := make([]int32, g.nObjs)
	copies[a>>1] = 1
	copies[b>>1] = 2
	copies[x>>1] = 3
	copies[y>>1] = 4
	groups, which := giaMfsBoxGroups(g, copies, 5)
	if len(groups) != 1 || len(groups[0]) != 2 || groups[0][0] != 3 || groups[0][1] != 2 || which[3] != 0 || which[2] != 1 || which[1] != -1 {
		t.Fatalf("groups %v map %v", groups, which)
	}
}
