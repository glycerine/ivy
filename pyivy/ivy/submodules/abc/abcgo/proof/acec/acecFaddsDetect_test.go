package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDetectFullAdderAcrossPolarities(t *testing.T) {
	for signs := 0; signs < 8; signs++ {
		p := gia.Gia_ManStart(64)
		var in [3]int32
		for i := range in {
			in[i] = gia.Gia_ManAppendCi(p) ^ int32(signs>>i&1)
		}
		var out [2]int32
		Acec_InsertFadd(p, in[:], out[:])
		gia.Gia_ManAppendCo(p, out[0])
		gia.Gia_ManAppendCo(p, out[1])
		var xor2 *Vec_Int_t
		adds := Gia_ManDetectFullAdders(p, 0, &xor2)
		found := false
		for i := 0; i+4 < vec.Vec_IntSize(adds); i += 5 {
			if vec.Vec_IntEntry(adds, i+3) == out[0]>>1 && vec.Vec_IntEntry(adds, i+4) == out[1]>>1 {
				found = true
			}
		}
		if !found {
			t.Fatalf("signs %d: full adders %v", signs, vec.Vec_IntArray(adds))
		}
	}
}
