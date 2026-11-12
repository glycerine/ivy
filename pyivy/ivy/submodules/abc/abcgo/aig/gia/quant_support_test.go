package gia

import "testing"

func TestQuantificationSupportExpansion(t *testing.T) {
	p := Gia_ManStart(100)
	Gia_ManQuantSetSuppStart(p)
	var first, last int32
	for i := 0; i < 65; i++ {
		lit := Gia_ManAppendCi(p)
		Gia_ManQuantSetSuppZero(p)
		Gia_ManQuantSetSuppCi(p, Gia_ManObj(p, lit>>1))
		if i == 0 {
			first = lit
		}
		last = lit
	}
	root := Gia_ManAppendAnd(p, first, last)
	if p.nSuppWords != 2 || Gia_ManQuantCheckSupp(p, root>>1, 0) != 1 || Gia_ManQuantCheckSupp(p, root>>1, 64) != 1 || Gia_ManQuantCheckSupp(p, root>>1, 1) != 0 {
		t.Fatal("support expansion incorrect")
	}
	Gia_ManQuantMarkUsedCis(p, func(_ any, id int32) int32 { return giaBool(id != 64) }, nil)
	if Gia_ManQuantCheckOverlap(p, root>>1) != 1 || Gia_ManQuantCheckOverlap(p, first>>1) != 0 {
		t.Fatal("support overlap incorrect")
	}
}

func TestExistentialQuantification(t *testing.T) {
	for _, mode := range []int{0, 1} {
		p := Gia_ManStart(30)
		Gia_ManQuantSetSuppStart(p)
		inputs := make([]int32, 3)
		for i := range inputs {
			inputs[i] = Gia_ManAppendCi(p)
			Gia_ManQuantSetSuppZero(p)
			Gia_ManQuantSetSuppCi(p, Gia_ManObj(p, inputs[i]>>1))
		}
		Gia_ManHashAlloc(p)
		root := Gia_ManHashMux(p, inputs[0], inputs[1], inputs[2])
		keep := func(_ any, id int32) int32 { return giaBool(id != 0) }
		var result int32
		if mode == 0 {
			result = Gia_ManQuantExist(p, root, keep, nil)
		} else {
			result = Gia_ManQuantExist2(p, root, keep, nil)
		}
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			values, err := giaEvaluateNodes(p, in)
			if err != nil {
				t.Fatal(err)
			}
			got := values[result>>1] != (result&1 != 0)
			want := in[1] || in[2]
			if got != want {
				t.Fatalf("mode=%d inputs=%v result=%v want=%v", mode, in, got, want)
			}
		}
	}
}

func TestQuantificationConstantAndInput(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManQuantSetSuppStart(p)
	a := Gia_ManAppendCi(p)
	Gia_ManQuantSetSuppZero(p)
	Gia_ManQuantSetSuppCi(p, Gia_ManObj(p, a>>1))
	keep := func(_ any, _ int32) int32 { return 0 }
	if Gia_ManQuantExist(p, 0, keep, nil) != 0 || Gia_ManQuantExist(p, 1, keep, nil) != 1 || Gia_ManQuantExist(p, a^1, keep, nil) != 1 {
		t.Fatal("constant or input quantification incorrect")
	}
}
