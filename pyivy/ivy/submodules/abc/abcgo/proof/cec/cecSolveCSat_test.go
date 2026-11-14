package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCbsSolveMarksOutputsAndStoresPattern(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, a^1, b)
	zero := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, zero)
	gia.Gia_ManAppendCo(g, x)
	pat := Cec_ManPatStart()
	defer Cec_ManPatStop(pat)
	Cec_ManSatSolveCSat(pat, g, &Cec_ParSat_t{NBTLimit: 1000})
	if gia.Gia_ObjMark1(gia.Gia_ManCo(g, 0)) != 1 || gia.Gia_ObjMark0(gia.Gia_ManCo(g, 1)) != 1 {
		t.Fatal("CBS output statuses incorrect")
	}
	if pat.vStorage == nil || vec.Vec_StrSize(pat.vStorage) == 0 {
		t.Fatal("CBS did not store satisfying pattern")
	}
}
