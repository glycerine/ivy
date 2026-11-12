package gia

import "testing"

func TestDupNoMuxesTruthAndBufferPolicy(t *testing.T) {
	for phase := int32(0); phase < 8; phase++ {
		p := Gia_ManStart(16)
		p.EnableMuxes()
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendXorReal(p, a^(phase&1), b^((phase>>1)&1))
		m := Gia_ManAppendMuxReal(p, c^((phase>>2)&1), x, b)
		buf := Gia_ManAppendBuf(p, m)
		Gia_ManAppendCo(p, buf)
		for skip := int32(0); skip < 2; skip++ {
			q := Gia_ManDupNoMuxes(p, skip)
			for _, obj := range q.pObjs {
				_, _, _, kind := q.Gate(obj.id)
				if kind == 3 || kind == 4 || skip != 0 && kind == 5 {
					t.Fatal("unlowered gate", kind)
				}
			}
			for mask := 0; mask < 8; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				want, e := p.Evaluate(in)
				if e != nil {
					t.Fatal(e)
				}
				got, e := q.Evaluate(in)
				if e != nil {
					t.Fatal(e)
				}
				if got[0] != want[0] {
					t.Fatal("truth mismatch", phase, skip, mask)
				}
			}
		}
	}
}
