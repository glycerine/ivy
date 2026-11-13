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

func TestTransformFlopsAcrossWordBoundary(t *testing.T) {
	const nregs = 34
	p := gia.Gia_ManStart(2*nregs + 1)
	regs := make([]int32, nregs)
	for i := range regs {
		regs[i] = gia.Gia_ManAppendCi(p)
	}
	for _, q := range regs {
		gia.Gia_ManAppendCo(p, q)
	}
	for _, q := range regs {
		gia.Gia_ManAppendCo(p, q)
	}
	gia.Gia_ManSetRegNum(p, nregs)
	flops, init := vec.Vec_IntAlloc(3), vec.Vec_IntAlloc(3)
	for _, entry := range []struct{ index, value int32 }{{31, 1}, {32, 2}, {31, 0}} {
		vec.Vec_IntPush(flops, regs[entry.index]>>1)
		vec.Vec_IntPush(init, entry.value)
	}
	transformed := Gia_ManTransformFlops(p, flops, init)
	state := make([]bool, nregs)
	for i := range state {
		state[i] = i%2 == 0
	}
	got, err := transformed.Evaluate(state)
	if err != nil {
		t.Fatal(err)
	}
	for i := range state {
		wantPO := state[i] != (i == 31 || i == 32)
		if got[i] != wantPO || got[nregs+i] != state[i] {
			t.Errorf("register %d: output/next = %v/%v, want %v/%v", i, got[i], got[nregs+i], wantPO, state[i])
		}
	}
}
