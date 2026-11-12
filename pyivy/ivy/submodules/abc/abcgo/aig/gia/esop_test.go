package gia

import (
	"math/rand"
	"testing"
)

func TestEsopGraphConversion(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b^1)
	y := Gia_ManAppendAnd(p, x^1, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x^1)
	q := Eso_ManCompute(p, 0, nil)
	if q == nil {
		t.Fatal("missing converted graph")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, err := p.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		got, err := q.Evaluate(in)
		if err != nil || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("inputs %v got %v want %v error %v", in, got, want, err)
		}
	}
	var cover *Vec_Wec_t
	if Eso_ManCompute(p, 0, &cover) != nil || cover == nil || Vec_WecSize(cover) == 0 {
		t.Fatal("missing extracted cover")
	}
}

func TestEsopConversionRandomSmallGraphs(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for trial := 0; trial < 50; trial++ {
		p := Gia_ManStart(40)
		lits := []int32{0, 1}
		for i := 0; i < 4; i++ {
			lits = append(lits, Gia_ManAppendCi(p))
		}
		for i := 0; i < 12; i++ {
			a, b := lits[rng.Intn(len(lits))]^int32(rng.Intn(2)), lits[rng.Intn(len(lits))]^int32(rng.Intn(2))
			lits = append(lits, Gia_ManAppendAnd2(p, a, b))
		}
		for i := 0; i < 3; i++ {
			Gia_ManAppendCo(p, lits[rng.Intn(len(lits))]^int32(rng.Intn(2)))
		}
		q := Eso_ManCompute(p, 0, nil)
		for bits := 0; bits < 16; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
			want, e := p.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			got, e := q.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("trial %d bits %d output %d got %v want %v", trial, bits, i, got, want)
				}
			}
		}
	}
}
