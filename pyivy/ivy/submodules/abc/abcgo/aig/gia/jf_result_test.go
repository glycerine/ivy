package gia

import (
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestJfDeriveStructuralMapping(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	var pars Jf_Par_t
	Jf_ManSetDefaultPars(&pars)
	pars.NLutSize = 3
	pars.FAreaOnly = 0
	p := Jf_ManAlloc(g, &pars)
	defer Jf_ManFree(p)
	p.pCutCmp = Jf_CutCompareDelay
	Jf_ManComputeCuts(p, 0)
	Jf_ManComputeRefs(p)
	Jf_ManDeriveMapping(p)
	if Gia_ManLutNum(g) != 1 || Gia_ObjIsLut(g, y>>1) == 0 || Gia_ObjLutSize(g, y>>1) != 3 {
		t.Fatalf("mapping LUTs %d size %d", Gia_ManLutNum(g), Gia_ObjLutSize(g, y>>1))
	}
	Gia_ManMappingVerify(g)
	for mask := 0; mask < 8; mask++ {
		out, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		if err != nil || out[0] != (mask == 7) {
			t.Fatalf("truth %d %v %v", mask, out, err)
		}
	}
}

func TestJfDeriveMinimizedMapping(t *testing.T) {
	for _, mode := range []struct {
		name     string
		dsd, cnf int32
	}{{"truth", 0, 0}, {"dsd", 1, 0}, {"cnf", 1, 1}} {
		t.Run(mode.name, func(t *testing.T) {
			g := Gia_ManStart(12)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b^1)
			y := Gia_ManAppendAnd(g, x, c)
			Gia_ManAppendCo(g, y^1)
			pars := &Jf_Par_t{NLutSize: 3, NCutNum: 4, FCutMin: 1, FFuncDsd: mode.dsd, FGenCnf: mode.cnf}
			p := Jf_ManAlloc(g, pars)
			defer Jf_ManFree(p)
			p.pCutCmp = Jf_CutCompareDelay
			Jf_ManComputeCuts(p, 0)
			Jf_ManComputeRefs(p)
			mapped := Jf_ManDeriveMappingGia(p)
			Gia_ManMappingVerify(mapped)
			for mask := 0; mask < 8; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				want, e1 := g.Evaluate(in)
				got, e2 := mapped.Evaluate(in)
				if e1 != nil || e2 != nil || want[0] != got[0] {
					t.Fatalf("input %d want %v got %v (%v %v)", mask, want, got, e1, e2)
				}
			}
			if mode.cnf != 0 {
				data, ok := mapped.pData.(*cnf.CNF)
				if !ok {
					t.Fatal("missing CNF")
				}
				if err := data.Validate(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
