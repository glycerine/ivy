package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCexRemapChecksReplay(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, x)
	pis := vec.Vec_IntStart(1)
	vec.Vec_IntWriteEntry(pis, 0, x>>1)
	abstract := util.Abc_CexAlloc(0, 1, 1)
	util.Abc_InfoSetBit(abstract.PData, 0)
	actual := Gia_ManCexRemap(p, abstract, pis)
	if actual == nil || !util.Abc_InfoHasBit(actual.PData, 0) {
		t.Fatal("valid input was not remapped")
	}
	abstract.PData[0] = 0
	if Gia_ManCexRemap(p, abstract, pis) != nil {
		t.Fatal("invalid witness accepted")
	}
}
