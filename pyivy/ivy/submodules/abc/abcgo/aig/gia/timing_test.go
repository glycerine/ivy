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
func TestTimingTransferOwnership(t *testing.T) {
	source, dest := Gia_ManStart(4), Gia_ManStart(4)
	source.vCiArrs = Vec_IntAlloc(1)
	Vec_IntPush(source.vCiArrs, 7)
	source.vNamesIn = Vec_PtrAlloc(1)
	Vec_PtrPush(source.vNamesIn, "input")
	source.pCellStr = "cell data"
	source.pManTime = tim.Tim_ManStart(0, 0)
	source.nAnd2Delay = 3
	arrivals, names, timing := source.vCiArrs, source.vNamesIn, source.pManTime
	Gia_ManTransferTiming(dest, source)
	if source.vCiArrs != nil || source.vNamesIn != nil || source.pCellStr != "" || source.pManTime != nil || source.nAnd2Delay != 0 {
		t.Fatal("source ownership not cleared")
	}
	if dest.vCiArrs != arrivals || dest.vNamesIn != names || dest.pManTime != timing || Vec_IntEntry(arrivals, 0) != 7 {
		t.Fatal("transfer invalidated aliases")
	}
}

func TestTimingBoxObjectBoundary(t *testing.T) {
	p := Gia_ManStart(5)
	pi, boxOut := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, pi)
	Gia_ManAppendCo(p, boxOut)
	timing := tim.Tim_ManStart(2, 2)
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0)
	p.pManTime = timing
	if p.BoxInputObjects(pi>>1) != nil {
		t.Fatal("PI is not a box output")
	}
	inputs := p.BoxInputObjects(boxOut >> 1)
	if len(inputs) != 1 || inputs[0] != Gia_ObjId(p, Gia_ManCo(p, 0)) {
		t.Fatal("box CO object IDs", inputs)
	}
	inputs[0] = 0
	if p.BoxInputObjects(boxOut >> 1)[0] == 0 {
		t.Fatal("box boundary aliases caller storage")
	}
}

func TestBoxLevelBoundaryOmitsOnlyDisabledArcs(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, out := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	Gia_ManAppendCo(p, out)
	timing := tim.Tim_ManStart(3, 3)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 2, 1, 2, -float32(giaInfinity)}})
	tim.Tim_ManCreateBox(timing, 0, 2, 2, 1, 0, 0)
	p.pManTime = timing
	if len(p.BoxInputObjects(out>>1)) != 2 {
		t.Fatal("structural box inputs filtered")
	}
	enabled := p.BoxLevelInputObjects(out >> 1)
	if len(enabled) != 1 || enabled[0] != Gia_ObjId(p, Gia_ManCo(p, 0)) {
		t.Fatal("enabled timing arcs", enabled)
	}
}
