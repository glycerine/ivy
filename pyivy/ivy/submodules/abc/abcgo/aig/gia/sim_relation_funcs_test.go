package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationRelationDeriveFunctions(t *testing.T) {
	rel := vec.Vec_WrdStart(2)
	data := vec.Vec_WrdArray(rel)
	for pat := int32(0); pat < 64; pat++ {
		mint := int32(0)
		if pat >= 32 {
			mint = 1
		}
		bit := 2*pat + mint
		data[bit>>6] |= uint64(1) << uint(bit&63)
		if pat == 0 {
			data[0] |= 2
		}
	}
	exact := vec.Vec_WrdArray(Gia_ManSimRelDeriveFuncs(nil, rel, 1))
	if exact[0] != 0xffffffff || exact[1] != 0xffffffff00000000 {
		t.Fatalf("wrong exact functions %x", exact)
	}
	dc := vec.Vec_WrdArray(Gia_ManSimRelDeriveFuncs2(nil, rel, 1))
	if dc[0] != 0xfffffffe || dc[1] != 0xffffffff00000000 {
		t.Fatalf("wrong don't-care functions %x", dc)
	}
}
