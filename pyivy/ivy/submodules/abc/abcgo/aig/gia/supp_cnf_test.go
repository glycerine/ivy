package gia

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestMin2CNFEncoding(t *testing.T) {
	for _, mux := range []bool{false, true} {
		p := Gia_ManStart(20)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		root := Gia_ManAppendAnd(p, Gia_ManAppendAnd(p, a, b^1), c)
		if mux {
			root = Gia_ManAppendMux(p, a, b^1, c)
			Gia_ManObj(p, root>>1).fMark0 = 1
		}
		Gia_ManAppendCo(p, root&^1)
		m := Gia_Man2SuppStart(p)
		Vec_IntFillExtra(&p.vCopies, int(Gia_ManObjNum(p)), -1)
		r := Gia_Min2ObjGetCnfVar(m, root>>1)
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			out, err := p.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			assumptions := make([]int, 0, 4)
			for i, lit := range []int32{a, b, c} {
				v := Gia_Min2ObjSatId(p, Gia_ManObj(p, lit>>1))
				if v < 0 {
					t.Fatal("input not encoded")
				}
				sign := 0
				if !inputs[i] {
					sign = 1
				}
				assumptions = append(assumptions, int(v)*2+sign)
			}
			for value := 0; value < 2; value++ {
				sign := 1 - value
				status := m.pSat.SolveAssumptions(append(assumptions, int(r)*2+sign))
				if (status == satoko.SATOKO_SAT) != (out[0] == (value == 1)) {
					t.Fatalf("mux=%v mask=%d value=%d", mux, mask, value)
				}
			}
		}
		Gia_Man2SuppStop(m)
	}
}
