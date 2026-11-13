package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCounterexampleJustificationAcrossFrames(t *testing.T) {
	g := gia.Gia_ManStart(20)
	pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 1 << 1
	var values, just *vec.Vec_Bit_t
	Bmc_GiaGenerateJust(g, c, &values, &just)
	n := int(gia.Gia_ManObjNum(g))
	piID := int(pi >> 1)
	roID := int(ro >> 1)
	poID := int(gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0)))
	riID := int(gia.Gia_ObjId(g, gia.Gia_ManRi(g, 0)))
	if vec.Vec_BitEntry(values, poID) != 0 || vec.Vec_BitEntry(values, n+poID) != 1 || vec.Vec_BitEntry(values, n+roID) != 1 {
		t.Fatal("incorrect replay values")
	}
	if vec.Vec_BitEntry(just, n+poID) != 1 || vec.Vec_BitEntry(just, n+roID) != 1 || vec.Vec_BitEntry(just, riID) != 1 || vec.Vec_BitEntry(just, piID) != 1 {
		t.Fatal("missing sequential justification")
	}
}
