package sfm

// Port of src/opt/sfm/sfmCnf.c.

import (
	"fmt"

	"github.com/glycerine/abcgo/bool/kit"
)

func abcVar2Lit(Var int32, c int32) int32 { return Var + Var + c }
func abcLit2Var(Lit int32) int32          { return Lit >> 1 }
func abcLitIsCompl(Lit int32) bool        { return Lit&1 != 0 }
func abcLitNot(Lit int32) int32           { return Lit ^ 1 }
func abcLitNotCond(Lit int32, c bool) int32 {
	if c {
		return Lit ^ 1
	}
	return Lit
}

// Sfm_PrintCnf prints a CNF in the vCnf encoding: literals separated by -1.
func Sfm_PrintCnf(vCnf []int8) {
	for _, Entry := range vCnf {
		Lit := int32(Entry)
		if Lit == -1 {
			fmt.Printf("\n")
		} else {
			s := ""
			if abcLitIsCompl(Lit) {
				s = "-"
			}
			fmt.Printf("%s%d ", s, abcLit2Var(Lit))
		}
	}
}

// Sfm_TruthToCnf derives the CNF of a node with nVars fanins whose output is
// variable nVars. Truth is used when nVars <= 6, pTruth otherwise. The CNF
// is written to *vCnf as literals with -1 terminating each clause; the cover
// is computed in *vCover. Returns the number of cubes.
func Sfm_TruthToCnf(Truth uint64, pTruth []uint64, nVars int, vCover *[]uint32, vCnf *[]int8) int {
	nWords := 1
	if nVars > 6 {
		nWords = 1 << uint(nVars-6)
	}
	*vCnf = (*vCnf)[:0]
	if nVars <= 6 {
		if Truth == 0 || ^Truth == 0 {
			// the unit clause fixes the output (variable nVars); a node
			// simplified to a constant may still have fanins (nVars > 0).
			// C used variable 0 here (CBUGS_FIXED.md #3).
			*vCnf = append(*vCnf, int8(abcVar2Lit(int32(nVars), int32(boolInt8(Truth == 0)))), -1)
			return 1
		}
	} else {
		// const0
		w := 0
		for w = 0; w < nWords; w++ {
			if pTruth[w] != 0 {
				break
			}
		}
		if w == nWords {
			*vCnf = append(*vCnf, int8(abcVar2Lit(int32(nVars), 1)), -1)
			return 1
		}
		// const1
		for w = 0; w < nWords; w++ {
			if ^pTruth[w] != 0 {
				break
			}
		}
		if w == nWords {
			*vCnf = append(*vCnf, int8(abcVar2Lit(int32(nVars), 0)), -1)
			return 1
		}
	}
	{
		nCubes := 0
		assert(nVars > 0, "nVars > 0")
		var small [2]uint32
		var large []uint32
		for c := 0; c < 2; c++ {
			var RetValue int32
			if nVars <= 6 {
				if c != 0 {
					Truth = ^Truth
				}
				small[0], small[1] = uint32(Truth), uint32(Truth>>32)
				RetValue = kit.Kit_TruthIsop(small[:], int32(nVars), vCover, 0)
			} else {
				if large == nil {
					large = make([]uint32, 2*nWords)
				}
				for k := 0; k < nWords; k++ {
					w := pTruth[k]
					if c != 0 {
						w = ^w
					}
					large[2*k], large[2*k+1] = uint32(w), uint32(w>>32)
				}
				RetValue = kit.Kit_TruthIsop(large, int32(nVars), vCover, 0)
			}
			assert(RetValue == 0, "RetValue == 0")
			nCubes += len(*vCover)
			for _, Cube := range *vCover {
				for k := 0; k < nVars; k++ {
					Literal := 3 & (Cube >> uint(k<<1))
					if Literal == 1 { // '0'  -> pos lit
						*vCnf = append(*vCnf, int8(abcVar2Lit(int32(k), 0)))
					} else if Literal == 2 { // '1'  -> neg lit
						*vCnf = append(*vCnf, int8(abcVar2Lit(int32(k), 1)))
					} else if Literal != 0 {
						assert(false, "Literal != 0")
					}
				}
				*vCnf = append(*vCnf, int8(abcVar2Lit(int32(nVars), int32(c))), -1)
			}
		}
		return nCubes
	}
}

func boolInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// Sfm_CreateCnf derives the CNFs of all internal nodes.
func Sfm_CreateCnf(p *Sfm_Ntk_t) [][]int8 {
	var vCnf []int8
	vCnfs := make([][]int8, p.nObjs)
	for i := p.Pis; i < len(p.vTruths)-p.Pos; i++ {
		uTruth := p.vTruths[i]
		var pRes []uint64
		if len(p.vTruths2) > 0 {
			pRes = p.vTruths2[p.vStarts[i]:]
		}
		Sfm_TruthToCnf(uTruth, pRes, Sfm_ObjFaninNum(p, i), &p.vCover, &vCnf)
		vCnfs[i] = append([]int8(nil), vCnf...)
	}
	return vCnfs
}

// Sfm_TranslateCnf maps the node CNF vCnf into SAT literals using vFaninMap
// (the SAT variables of the fanins followed by that of the node). Literals of
// variable iPivotVar are complemented. The clauses are written into *vRes,
// which always ends with one empty clause, as Vec_WecPushLevel leaves in C.
func Sfm_TranslateCnf(vRes *[][]int32, vCnf []int8, vFaninMap []int32, iPivotVar int32) {
	res := (*vRes)[:0]
	push := func() {
		if len(res) < cap(res) {
			res = res[:len(res)+1]
			res[len(res)-1] = res[len(res)-1][:0]
		} else {
			res = append(res, make([]int32, 0, 8))
		}
	}
	push()
	for _, Entry := range vCnf {
		if int32(Entry) == -1 {
			push()
			continue
		}
		Lit := int32(Entry)
		Lit = abcVar2Lit(vFaninMap[abcLit2Var(Lit)], 0) ^ (Lit & 1) // Abc_Lit2LitV
		Lit = abcLitNotCond(Lit, abcLit2Var(Lit) == iPivotVar)
		res[len(res)-1] = append(res[len(res)-1], Lit)
	}
	*vRes = res
}
