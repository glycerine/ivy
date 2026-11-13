package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCounterexampleStateAndSuffix(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(p)
	q := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, q)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 0) // C replay deliberately ignores witness initial register bits
	util.Abc_InfoSetBit(c.PData, 1) // x in frame 0
	state := Gia_ManGetStateAndCheckCex(p, c, 1)
	if state == nil || vec.Vec_IntEntry(state, 0) != 1 {
		t.Fatal("expected true state before frame 1")
	}
	initial := Gia_ManGetStateAndCheckCex(p, c, 0)
	if initial == nil || vec.Vec_IntEntry(initial, 0) != 0 {
		t.Fatal("expected zero initial state")
	}
	if _, valid := replayCexFromZero(p, c, 1, 1); valid {
		t.Fatal("suffix must restart with zero register state")
	}
	c.PData[0] = 0
	if Gia_ManGetStateAndCheckCex(p, c, 1) != nil {
		t.Fatal("accepted a non-failing witness")
	}
}
