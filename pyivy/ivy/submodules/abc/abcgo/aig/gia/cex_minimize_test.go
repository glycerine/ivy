package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCounterexampleSATCare(t *testing.T) {
	p := Gia_ManStart(20)
	x, y, s := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, x, s))
	Gia_ManAppendCo(p, y)
	Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 2, 2)
	// frame 0 y=1; frame 1 x=1. The other PI bits are irrelevant.
	c.PData[0] = (1 << 2) | (1 << 3)
	c.IFrame = 1
	if Gia_ManVerifyCex(p, c, 0) != 1 {
		t.Fatal("bad test witness")
	}
	unrolled := Gia_ManFramesForCexMin(p, 2)
	outs, err := unrolled.Evaluate([]bool{false, true, true, false})
	if err != nil || !outs[0] {
		t.Fatalf("unroll: %v %v", outs, err)
	}
	mask := Bmc_CexCareSatBasedMinimizeAig(p, c, 1, 0)
	if mask.PData[0] != (1<<2)|(1<<3) {
		t.Fatalf("unexpected care mask %b", mask.PData[0])
	}
}

func TestCounterexampleCareForcesOutput(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, s := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendXor(p, a, s))
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	Gia_ManSetRegNum(p, 1)
	for inputBits := uint32(0); inputBits < 16; inputBits++ {
		c := util.Abc_CexAlloc(1, 2, 2)
		c.IFrame = 1
		c.PData[0] = inputBits << 1
		if Gia_ManVerifyCex(p, c, 0) == 0 {
			continue
		}
		mask := Bmc_CexCareSatBasedMinimizeAig(p, c, 1, 0).PData[0]
		for other := uint32(0); other < 16; other++ {
			if ((other<<1)^c.PData[0])&mask != 0 {
				continue
			}
			candidate := *c
			candidate.PData = []uint32{other << 1}
			if Gia_ManVerifyCex(p, &candidate, 0) == 0 {
				t.Fatalf("mask %05b does not force output for %05b", mask, c.PData[0])
			}
		}
	}
}
