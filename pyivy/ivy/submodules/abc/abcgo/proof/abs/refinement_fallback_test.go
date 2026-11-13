package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRefinementKeepsSelectionWhenFilterIsEmpty(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	size := int(gia.Gia_ManObjNum(g))
	cuts := vec.Vec_IntStart(size + 3)
	vec.Vec_IntWriteEntry(cuts, int(gate>>1), int32(size))
	vec.Vec_IntWriteEntry(cuts, size, 2)
	vec.Vec_IntWriteEntry(cuts, size+1, a>>1)
	vec.Vec_IntWriteEntry(cuts, size+2, b>>1)
	gia.Gia_ManSetMapping(g, cuts)
	p := Rnm_ManStart(g)
	defer Rnm_ManStop(p, 0)
	mapping := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(mapping, gate>>1)
	cex := util.Abc_CexAlloc(0, 1, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	result := vec.Vec_IntArray(Rnm_ManRefine(p, cex, mapping, 0, 0, 0))
	if len(result) != 1 || result[0] != gate>>1 {
		t.Fatalf("filtered result=%v", result)
	}
}
