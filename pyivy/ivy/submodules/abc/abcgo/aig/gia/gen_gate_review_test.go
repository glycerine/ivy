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
