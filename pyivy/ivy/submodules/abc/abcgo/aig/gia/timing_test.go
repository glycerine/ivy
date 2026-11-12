package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestBoxTimingArrivalAndScaledArc(t *testing.T) {
	p := Gia_ManStart(8)
	a, boxOutput := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, boxOutput)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, root)
	timing := tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0)
	tim.Tim_ManInitPiArrival(timing, 0, 6)
	p.pManTime = timing
	p.nAnd2Delay = 2
	if level := Gia_ManLevelWithBoxes(p); level != 5 || Gia_ObjLevelId(p, boxOutput>>1) != 4 {
		t.Fatal("scaled box level", level, Gia_ObjLevelId(p, boxOutput>>1))
	}
}
func TestTimingTransferSelfAndOwnership(t *testing.T) {
	source, dest := Gia_ManStart(4), Gia_ManStart(4)
	source.vCiArrs = Vec_IntAlloc(1)
	Vec_IntPush(source.vCiArrs, 7)
	source.vNamesIn = Vec_PtrAlloc(1)
	Vec_PtrPush(source.vNamesIn, "input")
	source.pCellStr = "cell data"
	source.pManTime = tim.Tim_ManStart(0, 0)
	source.nAnd2Delay = 3
	arrivals, names, timing := source.vCiArrs, source.vNamesIn, source.pManTime
	Gia_ManTransferTiming(source, source)
	if source.vCiArrs != arrivals || source.vNamesIn != names || source.pCellStr != "cell data" || source.pManTime != timing {
		t.Fatal("self transfer destroyed metadata")
	}
	Gia_ManTransferTiming(dest, source)
	if source.vCiArrs != nil || source.vNamesIn != nil || source.pCellStr != "" || source.pManTime != nil || source.nAnd2Delay != 0 {
		t.Fatal("source ownership not cleared")
	}
	if dest.vCiArrs != arrivals || dest.vNamesIn != names || dest.pManTime != timing || Vec_IntEntry(arrivals, 0) != 7 {
		t.Fatal("transfer invalidated aliases")
	}
}
