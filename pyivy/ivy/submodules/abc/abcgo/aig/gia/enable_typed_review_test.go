package gia

import "testing"

func TestEnableTransformsPreserveTypedLogic(t *testing.T) {
	for _, operation := range []struct {
		name string
		copy func(*Gia_Man_t) *Gia_Man_t
	}{
		{"unroll", func(p *Gia_Man_t) *Gia_Man_t { return Gia_ManUnrollInit(p, 1) }},
		{"remove", Gia_ManRemoveEnables},
		{"remove2", Gia_ManRemoveEnables2},
	} {
		for _, kind := range []string{"xor", "mux", "buffer", "simple opposite"} {
			t.Run(operation.name+"/"+kind, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("valid graph transformation panicked: %v", r)
					}
				}()
				p := Gia_ManStart(16)
				defer Gia_ManStop(p)
				p.pName, p.PSpec = "typed", "typed spec"
				a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
				var lit int32
				switch kind {
				case "xor":
					lit = Gia_ManAppendXorReal(p, a, b^1)
				case "mux":
					p.EnableMuxes()
					lit = Gia_ManAppendMuxReal(p, c, a^1, b)
				case "buffer":
					lit = Gia_ManAppendBuf(p, a^1)
				case "simple opposite":
					p.fGiaSimple = 1
					lit = Gia_ManAppendAnd(p, a, a^1)
				}
				Gia_ManAppendCo(p, lit)
				q := operation.copy(p)
				defer Gia_ManStop(q)
				if q.pName != p.pName || q.PSpec != p.PSpec {
					t.Fatal("graph identity was lost")
				}
				for mask := 0; mask < 8; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					want := false
					switch kind {
					case "xor":
						want = inputs[0] != !inputs[1]
					case "mux":
						want = inputs[1]
						if inputs[2] {
							want = !inputs[0]
						}
					case "buffer":
						want = !inputs[0]
					}
					got, err := q.Evaluate(inputs)
					if err != nil || len(got) != 1 || got[0] != want {
						t.Errorf("mask%d: outputs %v, error %v, want %t", mask, got, err, want)
					}
				}
			})
		}
	}
}
