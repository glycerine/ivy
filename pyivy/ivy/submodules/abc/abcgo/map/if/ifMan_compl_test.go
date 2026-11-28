package if_

import "testing"

// If_ManCreateAndCompl(p, a, ca, b, cb) must build what
// If_ManCreateAnd(p, If_NotCond(a, ca), If_NotCond(b, cb)) builds: the same
// trivial results (constants, equal and opposite fanins) and the same new
// node (fanins, complements, phase, level).
func TestCreateAndComplMatchesProxies(t *testing.T) {
	describe := func(p *If_Man_t, o *If_Obj_t) [6]int32 {
		r := If_Regular(o)
		if r.Type != IF_AND || r.pFanin0 == nil {
			return [6]int32{r.Id, If_IsComplement(o), -1, -1, -1, -1}
		}
		return [6]int32{-2, If_IsComplement(o), r.pFanin0.Id<<1 | int32(r.fCompl0), r.pFanin1.Id<<1 | int32(r.fCompl1), int32(r.fPhase), int32(r.Level)}
	}
	build := func(compl bool, i, j int, ci, cj int32) [6]int32 {
		var pars If_Par_t
		If_ManSetDefaultPars(&pars)
		pars.NLutSize = 6
		p := If_ManStart(&pars)
		x, y := If_ManCreateCi(p), If_ManCreateCi(p)
		y.fPhase = 1
		ops := []*If_Obj_t{p.pConst1, x, y}
		a, b := ops[i], ops[j]
		if compl {
			return describe(p, If_ManCreateAndCompl(p, a, ci, b, cj))
		}
		return describe(p, If_ManCreateAnd(p, If_NotCond(a, ci), If_NotCond(b, cj)))
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for ci := int32(0); ci < 2; ci++ {
				for cj := int32(0); cj < 2; cj++ {
					if got, want := build(true, i, j, ci, cj), build(false, i, j, ci, cj); got != want {
						t.Errorf("ops %d,%d compl %d,%d: got %v, want %v", i, j, ci, cj, got, want)
					}
				}
			}
		}
	}
}
