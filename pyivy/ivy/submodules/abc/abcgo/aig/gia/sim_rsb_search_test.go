package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubstitutionSearch(t *testing.T) {
	sims, transposed := vec.Vec_WrdStart(64), vec.Vec_WrdStart(64)
	vec.Vec_WrdWriteEntry(sims, 0, 0xc)
	vec.Vec_WrdWriteEntry(sims, 1, 0xa)
	for pat := 0; pat < 64; pat++ {
		var word uint64
		for cand := 0; cand < 2; cand++ {
			if vec.Vec_WrdEntry(sims, cand)&(uint64(1)<<uint(pat)) != 0 {
				word |= uint64(1) << uint(cand)
			}
		}
		vec.Vec_WrdWriteEntry(transposed, pat, word)
	}
	cands := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(cands, 0)
	vec.Vec_IntPush(cands, 1)
	result := Gia_RsbSetFind([]uint64{3}, []uint64{12}, sims, 1, transposed, 1, cands)
	if vec.Vec_IntSize(result) != 1 || vec.Vec_IntEntry(result, 0) != 0 {
		t.Fatalf("unexpected support %v", vec.Vec_IntArray(result))
	}
}
