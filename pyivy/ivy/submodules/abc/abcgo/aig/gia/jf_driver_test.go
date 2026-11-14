package gia

import (
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestJfDriver(t *testing.T) {
	for _, mode := range []struct {
		name                string
		min, dsd, pure, gen int32
	}{{"mapping", 0, 0, 0, 0}, {"truth", 1, 0, 0, 0}, {"dsd", 1, 1, 0, 0}, {"aig", 0, 0, 1, 0}, {"truth-aig", 1, 0, 1, 0}, {"dsd-aig", 1, 1, 1, 0}, {"cnf", 0, 0, 0, 1}} {
		t.Run(mode.name, func(t *testing.T) {
			g := Gia_ManStart(12)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b^1)
			y := Gia_ManAppendAnd(g, x, c)
			Gia_ManAppendCo(g, y^1)
			var pars Jf_Par_t
			Jf_ManSetDefaultPars(&pars)
			pars.NLutSize = 3
			pars.FCutMin = mode.min
			pars.FFuncDsd = mode.dsd
			pars.FPureAig = mode.pure
			pars.FGenCnf = mode.gen
			out := Jf_ManPerformMapping(g, &pars)
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				a, e1 := g.Evaluate(inputs)
				b, e2 := out.Evaluate(inputs)
				if e1 != nil || e2 != nil || a[0] != b[0] {
					t.Fatalf("truth %d %v %v %v %v", mask, a, b, e1, e2)
				}
			}
			if mode.pure == 0 {
				Gia_ManMappingVerify(out)
			}
			if mode.gen != 0 {
				data, ok := out.pData.(*cnf.CNF)
				if !ok {
					t.Fatal("no CNF")
				}
				if err := data.Validate(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
