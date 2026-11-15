package gia

import "testing"

func TestGenWordUsesActualXor(t *testing.T) {
	for _, direct := range []bool{false, true} {
		name := "driver"
		if direct {
			name = "gate"
		}
		t.Run(name, func(t *testing.T) {
			p := Gia_ManStart(16)
			defer Gia_ManStop(p)
			a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			x := Gia_ManAppendXorReal(p, a, b)
			Gia_ManAppendCo(p, x)
			if direct {
				p.NSimWords = 1
				p.vSims = Vec_WrdStart(int(p.nObjs))
				Gia_ManObjSim(p, a>>1)[0] = 0xaaaaaaaaaaaaaaaa
				Gia_ManObjSim(p, b>>1)[0] = 0xcccccccccccccccc
				Gia_ManObjSimAnd(p, x>>1)
			} else {
				in := Vec_WrdAllocArrayCopy([]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc}, 2)
				defer Vec_WrdFree(in)
				Gia_ManSimulateWordsInit(p, in)
			}
			if got := Gia_ManObjSim(p, x>>1)[0]; got != 0x6666666666666666 {
				t.Fatalf("XOR got%016x want6666666666666666", got)
			}
		})
	}
}

func TestGenWordUsesMuxControl(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	p.EnableMuxes()
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendMuxReal(p, c, b, a)
	Gia_ManAppendCo(p, root)
	in := Vec_WrdAllocArrayCopy([]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0}, 3)
	defer Vec_WrdFree(in)
	Gia_ManSimulateWordsInit(p, in)
	if got := Gia_ManObjSim(p, root>>1)[0]; got != 0xcacacacacacacaca {
		t.Fatalf("MUX got%016x wantcacacacacacacaca", got)
	}
}

func TestGenRepeatedWordSizesAndRejectedInputs(t *testing.T) {
	p := genReviewGraph(true, 125, true)
	defer Gia_ManStop(p)
	for _, words := range []int32{5, 1, 0, 3} {
		in := Vec_WrdStart(int(4 * words))
		for i := int32(0); i < 4; i++ {
			for w := int32(0); w < words; w++ {
				Vec_WrdWriteEntry(in, int(i*words+w), uint64(i+1)*0x9e3779b97f4a7c15^uint64(w+1)*0xd1b54a32d192ed03)
			}
		}
		Gia_ManSimulateWordsInit(p, in)
		if p.NSimWords != words || Vec_WrdSize(p.vSims) != int(p.nObjs*words) {
			t.Fatal("resized simulation dimensions wrong")
		}
		for w := int32(0); w < words; w++ {
			for bit := uint(0); bit < 64; bit++ {
				assignment := 0
				for i := int32(0); i < 4; i++ {
					if Vec_WrdEntry(in, int(i*words+w))>>bit&1 != 0 {
						assignment |= 1 << i
					}
				}
				values := cofReviewValues(p, assignment)
				for id := int32(0); id < p.nObjs; id++ {
					if (Gia_ManObjSim(p, id)[w]>>bit&1 != 0) != values[id] {
						t.Fatalf("resized words%d word%d bit%d obj%d", words, w, bit, id)
					}
				}
			}
		}
		Vec_WrdFree(in)
	}
	old, words := p.vSims, p.NSimWords
	bad := Vec_WrdStart(5)
	defer Vec_WrdFree(bad)
	cexReviewMustPanic(t, func() { Gia_ManSimulateWordsInit(p, bad) })
	cexReviewMustPanic(t, func() { Gia_ManSimulateWordsInit(p, nil) })
	cexReviewMustPanic(t, func() { Gia_ManSimulateWords(p, -1) })
	cexReviewMustPanic(t, func() { Gia_ManObjSim(p, -1) })
	cexReviewMustPanic(t, func() { Gia_ManObjSim(p, p.nObjs) })
	if p.vSims != old || p.NSimWords != words {
		t.Fatal("rejected operation modified simulation state")
	}
	view := Gia_ManObjSim(p, Gia_ManCi(p, 0).id)
	next := Gia_ManObjSim(p, Gia_ManCi(p, 1).id)
	before := append([]uint64(nil), next...)
	_ = append(view, ^uint64(0))
	for i, v := range before {
		if next[i] != v {
			t.Fatal("appending object view overwrote another object's storage")
		}
	}
}
