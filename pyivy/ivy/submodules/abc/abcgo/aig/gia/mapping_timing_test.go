package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestMappedLutLevelWithBoxes(t *testing.T) {
	p := Gia_ManStart(8)
	a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, root)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	Vec_IntWriteEntry(p.VMapping, int(root>>1), int32(Vec_IntSize(p.VMapping)))
	for _, v := range []int32{2, a >> 1, box >> 1, root >> 1} {
		Vec_IntPush(p.VMapping, v)
	}
	timing := tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0)
	p.PManTime = timing
	var luts, edges, levels int32
	Gia_ManLutParams(p, &luts, &edges, &levels)
	if luts != 1 || edges != 2 || levels != 2 {
		t.Fatalf("luts=%d edges=%d levels=%d", luts, edges, levels)
	}
}
