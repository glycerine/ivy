package gia

import (
	"math/rand"
	"testing"
)

func TestLutBalanceRandomSemantics(t *testing.T) {
	r := rand.New(rand.NewSource(881))
	for trial := 0; trial < 100; trial++ {
		p := Gia_ManStart(200)
		Gia_ManHashStart(p)
		es := []int32{0, 1}
		for i := 0; i < 5; i++ {
			es = append(es, Gia_ManAppendCi(p))
		}
		for i := 0; i < 40; i++ {
			a, b := es[r.Intn(len(es))]^int32(r.Intn(2)), es[r.Intn(len(es))]^int32(r.Intn(2))
			switch i % 4 {
			case 0:
				es = append(es, Gia_ManHashXor(p, a, b))
			case 1:
				c := es[r.Intn(len(es))]
				es = append(es, Gia_ManHashMux(p, c, a, b))
			default:
				es = append(es, Gia_ManHashAnd(p, a, b))
			}
		}
		for i := 0; i < 4; i++ {
			Gia_ManAppendCo(p, es[r.Intn(len(es))])
		}
		Gia_ManHashStop(p)
		Gia_ManSetRegNum(p, 1)
		p.nConstrs = 1
		q := Gia_ManBalanceLut(p, int32(3+trial%4), int32(2+trial%7), 0)
		if q.nRegs != 1 || q.nConstrs != 1 {
			t.Fatal("sequential metadata")
		}
		for mask := 0; mask < 32; mask++ {
			in := make([]bool, 5)
			for i := range in {
				in[i] = mask&(1<<i) != 0
			}
			a, e := p.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			b, e := q.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			for i := range a {
				if a[i] != b[i] {
					t.Fatalf("trial%d mask%d output%d", trial, mask, i)
				}
			}
		}
	}
}
func TestLutCutMergesAndAntichains(t *testing.T) {
	for a := 1; a < 64; a++ {
		for b := 1; b < 64; b++ {
			var x, y Bal_Cut_t
			for i := int32(0); i < 6; i++ {
				if a&(1<<i) != 0 {
					x.Leaves = append(x.Leaves, i)
				}
				if b&(1<<i) != 0 {
					y.Leaves = append(y.Leaves, i)
				}
			}
			x.Sign = Bal_CutGetSign(x.Leaves)
			y.Sign = Bal_CutGetSign(y.Leaves)
			x.Delay = 2
			y.Delay = 3
			var c Bal_Cut_t
			if !Bal_CutMergeOrder(&x, &y, &c, 6) || c.Delay != 3 || c.Sign != x.Sign|y.Sign {
				t.Fatal("ordered union")
			}
			want := 0
			for i := 0; i < 6; i++ {
				if (a|b)&(1<<i) != 0 {
					want++
				}
			}
			if len(c.Leaves) != want {
				t.Fatal("union cardinality")
			}
			if Bal_CutCheck(&x, &y) != (a&b == b) {
				t.Fatal("subset")
			}
		}
	}
}
func TestLutBalanceChainAndConstants(t *testing.T) {
	p := Gia_ManStart(100)
	Gia_ManHashStart(p)
	in := make([]int32, 12)
	for i := range in {
		in[i] = Gia_ManAppendCi(p)
	}
	l := in[0]
	for _, x := range in[1:] {
		l = Gia_ManHashAnd(p, l, x)
	}
	Gia_ManAppendCo(p, l)
	Gia_ManAppendCo(p, 0)
	Gia_ManAppendCo(p, 1)
	Gia_ManHashStop(p)
	q := Gia_ManBalanceLut(p, 6, 8, 0)
	for mask := 0; mask < 1<<12; mask++ {
		bits := make([]bool, 12)
		for i := range bits {
			bits[i] = mask&(1<<i) != 0
		}
		v, e := q.Evaluate(bits)
		if e != nil {
			t.Fatal(e)
		}
		if v[0] != (mask == 4095) || v[1] || !v[2] {
			t.Fatal("chain or constants changed")
		}
	}
}

func TestLutBinarySizeAndValidation(t *testing.T) {
	p := Gia_ManStart(15)
	Gia_ManHashStart(p)
	in := make([]int32, 4)
	for i := range in {
		in[i] = Gia_ManAppendCi(p)
	}
	Gia_ManAppendCo(p, Gia_ManHashAnd(p, Gia_ManHashAnd(p, in[0], in[1]), Gia_ManHashAnd(p, in[2], in[3])))
	Gia_ManHashStop(p)
	q := Gia_ManBalanceLut(p, 2, 2, 0)
	for mask := 0; mask < 16; mask++ {
		v, e := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0})
		if e != nil || v[0] != (mask == 15) {
			t.Fatal("binary LUT")
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("cut capacity validation missing")
		}
	}()
	Gia_ManBalanceLut(p, 6, 1, 0)
}
func TestLutCutInsertionTieAndDominance(t *testing.T) {
	makeCut := func(leaves ...int32) Bal_Cut_t {
		return Bal_Cut_t{Sign: Bal_CutGetSign(leaves), Delay: 2, Leaves: leaves}
	}
	a, b, c := makeCut(1, 2, 3), makeCut(4, 5, 6), makeCut(1, 2)
	cuts := balLutAddCut(nil, a, 4)
	cuts = balLutAddCut(cuts, b, 4)
	if cuts[0].Leaves[0] != 4 {
		t.Fatal("new equal-cost cut should precede old cut")
	}
	cuts = balLutAddCut(cuts, c, 4)
	if len(cuts) != 2 || len(cuts[0].Leaves) != 2 || !Bal_SetCheckArray(cuts) {
		t.Fatal("dominated cut not removed")
	}
	cuts = balLutAddCut(cuts, a, 4)
	if len(cuts) != 2 {
		t.Fatal("superset cut admitted")
	}
}
