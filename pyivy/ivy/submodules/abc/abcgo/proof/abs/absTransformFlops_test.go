package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestTransformFlopsChangesStateCoordinates(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(p)
	q := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, q)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManSetRegNum(p, 1)
	flops := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(flops, q>>1)
	init := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(init, 1)
	transformed := Gia_ManTransformFlops(p, flops, init)
	// Original q=1 corresponds to transformed q'=0; next q'=!x.
	for _, xval := range []bool{false, true} {
		out, err := transformed.Evaluate([]bool{xval, false})
		if err != nil || !out[0] || out[1] == xval {
			t.Fatalf("x=%v: outputs %v, error %v", xval, out, err)
		}
	}
}
