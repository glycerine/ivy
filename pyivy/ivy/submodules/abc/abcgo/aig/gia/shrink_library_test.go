package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDarCutReconstruction(t *testing.T) {
	truths := []uint32{0x8888, 0x6996, 0xf0aa, 0xaaaa, 0x0000, 0xffff}
	for x := uint32(0); x < 256; x++ {
		truths = append(truths, (x*251+37)&0xffff)
	}
	for _, truth := range truths {
		p := Gia_ManStart(32)
		leaves := vec.Vec_IntAlloc(4)
		for i := 0; i < 4; i++ {
			vec.Vec_IntPush(leaves, Gia_ManAppendCi(p))
		}
		Gia_ManCleanLevels(p, 32)
		Gia_ManHashStart(p)
		lit := Dar_LibEvalBuild(p, leaves, truth, 0, nil)
		Gia_ManHashStop(p)
		Gia_ManAppendCo(p, lit)
		got := Gia_LutComputeTruth66_rec(p, Gia_ManCo(p, 0))
		if uint16(got) != uint16(truth) {
			t.Fatalf("truth %04x got %04x", truth, uint16(got))
		}
	}
}

func TestMappedShrink4Truth(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c^1)
	Gia_ManAppendCo(p, y)
	p.VMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.VMapping, int(y>>1), int32(p.nObjs))
	vec.Vec_IntPush(p.VMapping, 3)
	for _, id := range []int32{a >> 1, b >> 1, c >> 1} {
		vec.Vec_IntPush(p.VMapping, id)
	}
	vec.Vec_IntPush(p.VMapping, y>>1)
	for _, keep := range []int32{0, 1} {
		q := Gia_ManMapShrink4(p, keep, 0)
		if q == nil || Gia_ManVerifyTwoTruths(p, q) != 1 {
			t.Fatalf("shrink keep=%d changed function", keep)
		}
	}
}
