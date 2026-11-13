package pdr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestTxs2ReductionAndRepeatedCones(t *testing.T) {
	g := aig.NewGraph()
	a, b, ff0, ff1 := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.AND(a, ff0), g.AND(b, ff1)))
	g.AddOutput(ff0)
	g.AddOutput(ff1)
	aig.Aig_ManSetRegNum(g, 2)
	prio := vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	p := Txs_ManStart(nil, g, prio)
	defer Txs_ManStop(p)
	co := gia.Gia_ManCo(p.pGia, 0)
	vec.Vec_IntPush(p.vCoObjs, gia.Gia_ObjId(p.pGia, co))
	for mask := 0; mask < 16; mask++ {
		Txs_ManCollectCone(p.pGia, p.vCoObjs, p.vCiObjs, p.vNodes)
		vec.Vec_IntClear(p.vCiVals)
		for _, id := range vec.Vec_IntArray(p.vCiObjs) {
			o := gia.Gia_ManObj(p.pGia, id)
			vec.Vec_IntPush(p.vCiVals, int32(mask>>gia.Gia_ObjCioId(o)&1))
		}
		vec.Vec_IntClear(p.vCoVals)
		out := int32(0)
		if mask == 15 {
			out = 1
		}
		vec.Vec_IntPush(p.vCoVals, out)
		Txs_ManForwardPass(p.pGia, p.vPrio, p.vCiObjs, p.vCiVals, p.vNodes, p.vCoObjs, p.vCoVals)
		Txs_ManFindCiReduction(p.pGia, p.vPrio, p.vCiObjs, p.vNodes, p.vCoObjs, p.vPiLits, p.vFfLits, p.vTemp)
		Txs_ManVerify(p.pGia, p.vCiObjs, p.vNodes, p.vPiLits, p.vFfLits, p.vCoObjs, p.vCoVals)
		if mask == 15 && (vec.Vec_IntSize(p.vPiLits) != 2 || vec.Vec_IntSize(p.vFfLits) != 2) {
			t.Fatalf("mask=%d: true AND requires all inputs", mask)
		}
	}
}

func TestTxs2EmptyConeConstant(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	p := Txs_ManStart(nil, g, vec.Vec_IntAlloc(0))
	defer Txs_ManStop(p)
	vec.Vec_IntPush(p.vCoObjs, gia.Gia_ObjId(p.pGia, gia.Gia_ManCo(p.pGia, 0)))
	vec.Vec_IntPush(p.vCoVals, 1)
	for i := 0; i < 2; i++ {
		Txs_ManCollectCone(p.pGia, p.vCoObjs, p.vCiObjs, p.vNodes)
		Txs_ManForwardPass(p.pGia, p.vPrio, p.vCiObjs, p.vCiVals, p.vNodes, p.vCoObjs, p.vCoVals)
		Txs_ManFindCiReduction(p.pGia, p.vPrio, p.vCiObjs, p.vNodes, p.vCoObjs, p.vPiLits, p.vFfLits, p.vTemp)
		Txs_ManVerify(p.pGia, p.vCiObjs, p.vNodes, p.vPiLits, p.vFfLits, p.vCoObjs, p.vCoVals)
		if vec.Vec_IntSize(p.vFfLits) != 0 {
			t.Fatal("constant requires no registers")
		}
	}
}
