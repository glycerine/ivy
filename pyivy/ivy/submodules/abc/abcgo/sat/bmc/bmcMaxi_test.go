package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func maxiGraph() *gia.Gia_Man_t {
	p := gia.Gia_ManStart(16)
	x := gia.Gia_ManAppendCi(p)
	r := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, x, r))
	gia.Gia_ManSetRegNum(p, 1)
	return p
}
func TestMaxiUnfold(t *testing.T) {
	p := maxiGraph()
	for init := int32(0); init <= 5; init++ {
		for use := int32(0); use < 2; use++ {
			v := vec.Vec_IntAllocArrayCopy([]int32{init}, 1)
			q := Gia_ManMaxiUnfold(p, 2, use, v)
			for bits := 0; bits < 16; bits++ {
				in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
				var initial bool
				switch init {
				case 0:
					initial = use != 0 && in[0] && in[1]
				case 1:
					initial = use == 0 || !in[0] || in[1]
				case 2:
					initial = in[0] && in[1]
				case 3:
					initial = !in[0] || in[1]
				case 4:
					initial = false
				case 5:
					initial = true
				}
				want := initial && in[2] && in[3]
				got, err := q.Evaluate(in)
				if err != nil || len(got) != 1 || got[0] != want {
					t.Fatalf("init=%d use=%d in=%v got=%v err=%v want=%v", init, use, in, got, err, want)
				}
			}
		}
	}
}
func TestMaxiPerformPreservesNecessaryState(t *testing.T) {
	p := maxiGraph()
	for _, init := range []int32{0, 1} {
		input := vec.Vec_IntAllocArrayCopy([]int32{init}, 1)
		got := Gia_ManMaxiPerform(p, input, 1, 0, 0)
		if vec.Vec_IntEntry(got, 0) != init {
			t.Fatalf("initial %d result %v", init, vec.Vec_IntArray(got))
		}
	}
}
func TestMaxiPerformWeakensUnusedState(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManSetRegNum(p, 1)
	got := Gia_ManMaxiTest(p, nil, 1, 0, 0, 0, 0)
	if vec.Vec_IntEntry(got, 0) != 2 {
		t.Fatalf("got %v", vec.Vec_IntArray(got))
	}
}
