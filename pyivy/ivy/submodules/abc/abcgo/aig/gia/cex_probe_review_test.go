package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCounterexampleProbeOwnsSourceScratch(t *testing.T) {
	p := Gia_ManStart(16)
	x, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, x, r))
	Gia_ManAppendCo(p, x)
	Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 1, 37)
	c.IFrame = 36
	for i := int32(0); i < 37; i++ {
		bit := 1 + i
		c.PData[bit>>5] |= 1 << uint(bit&31)
	}
	Gia_ManCounterExampleValueStart(p, c)
	if p.pData2 == nil {
		t.Fatal("probe did not occupy source scratch storage")
	}
	for frame := int32(0); frame < 37; frame++ {
		want := int32(0)
		if frame != 0 {
			want = 1
		}
		if got := Gia_ManCounterExampleValueLookup(p, r>>1, frame); got != want {
			t.Fatalf("frame%d state %d want%d", frame, got, want)
		}
	}
	Gia_ManCounterExampleValueStop(p)
	if p.pData2 != nil {
		t.Fatal("probe scratch not released")
	}
}
