package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestFaultUnfold(t *testing.T) {
	p := gia.Gia_ManStart(16)
	x := gia.Gia_ManAppendCi(p)
	r := gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, x, r)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManSetRegNum(p, 1)
	for _, ff := range []int32{0, 1} {
		for _, mux := range []int32{0, 1} {
			q := Gia_ManFaultUnfold(p, mux, ff)
			if len(q.InputLiterals()) != 4 || len(q.OutputLiterals()) != 2 || q.RegisterCount() != 0 {
				t.Fatalf("ff=%d mux=%d: incorrect interface", ff, mux)
			}
			for bits := 0; bits < 16; bits++ {
				in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
				first := in[0] && in[1]
				second := in[2] && first
				po, ri := second, second
				if mux != 0 && in[3] {
					ri = first
					if ff == 0 {
						po = first
					}
				}
				got, err := q.Evaluate(in)
				if err != nil || got[0] != po || got[1] != ri {
					t.Fatalf("ff=%d mux=%d inputs=%v: got %v (%v), want %v %v", ff, mux, in, got, err, po, ri)
				}
			}
		}
	}
}

func TestFaultControls(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	for _, stuck := range []bool{false, true} {
		n := 1
		if stuck {
			n = 2
		}
		for mask := 0; mask < 1<<n; mask++ {
			enabled := make([]int32, n)
			for i := range enabled {
				if mask&(1<<i) != 0 {
					enabled[i] = 1
				}
			}
			m := vec.Vec_IntAllocArrayCopy(enabled, n)
			var q *gia.Gia_Man_t
			if stuck {
				q = Gia_ManStuckAtUnfold(p, m)
			} else {
				q = Gia_ManFlipUnfold(p, m)
			}
			for bits := 0; bits < 1<<(2+n); bits++ {
				in := make([]bool, 2+n)
				for i := range in {
					in[i] = bits&(1<<i) != 0
				}
				want := in[0] && in[1]
				if stuck {
					if enabled[0] != 0 && in[2] {
						want = false
					}
					if enabled[1] != 0 && in[3] {
						want = true
					}
				} else if enabled[0] != 0 && in[2] {
					want = !want
				}
				out, err := q.Evaluate(in)
				if err != nil || out[0] != want {
					t.Fatalf("stuck=%v enabled=%v in=%v got=%v err=%v want=%v", stuck, enabled, in, out, err, want)
				}
			}
		}
	}
}

func TestFunctionFaultControls(t *testing.T) {
	for polarity := 0; polarity < 4; polarity++ {
		p := gia.Gia_ManStart(8)
		a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a^int32(polarity&1), b^int32((polarity>>1)&1)))
		for mask := 0; mask < 16; mask++ {
			enabled := make([]int32, 4)
			for i := range enabled {
				if mask&(1<<i) != 0 {
					enabled[i] = 1
				}
			}
			q := Gia_ManFOFUnfold(p, vec.Vec_IntAllocArrayCopy(enabled, 4))
			for bits := 0; bits < 64; bits++ {
				in := make([]bool, 6)
				for i := range in {
					in[i] = bits&(1<<i) != 0
				}
				row := 0
				if in[0] {
					row++
				}
				if in[1] {
					row += 2
				}
				want := (in[0] != (polarity&1 != 0)) && (in[1] != (polarity&2 != 0))
				if row < 3 && enabled[row] != 0 && in[2+row] {
					want = !want
				}
				out, err := q.Evaluate(in)
				if err != nil || out[0] != want {
					t.Fatalf("polarity=%d mask=%d in=%v got=%v err=%v want=%v", polarity, mask, in, out, err, want)
				}
			}
		}
	}
}

func TestFaultCofactor(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b)^1)
	for fixed := int32(0); fixed <= 1; fixed++ {
		q := Gia_ManFaultCofactor(p, vec.Vec_IntAllocArrayCopy([]int32{fixed}, 1))
		if len(q.InputLiterals()) != 2 {
			t.Fatal("cofactor changed input interface")
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			got, err := q.Evaluate(in)
			want := !(fixed != 0 && in[1])
			if err != nil || got[0] != want {
				t.Fatalf("fixed=%d in=%v got=%v err=%v want=%v", fixed, in, got, err, want)
			}
		}
	}
}
