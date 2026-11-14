package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

type Exa3_Man_tV2 struct {
	pPars                                                 *Bmc_EsPar_t
	nVars, nNodes, nLutSize, LutMask, nObjs, nWords, iVar int32
	pTruth                                                []uint64
	vInfo                                                 *vec.Vec_Wrd_t
	VarMarks                                              [MAJ_NOBJSV2][6][MAJ_NOBJSV2]int32
	VarVals                                               [MAJ_NOBJSV2]int32
	vOutLits                                              *vec.Vec_Wec_t
	pSat                                                  *bsat.Solver
}

func Exa3_ManTruthV2(p *Exa3_Man_tV2, v int32) []uint64 {
	return vec.Vec_WrdArray(p.vInfo)[v*p.nWords : (v+1)*p.nWords]
}
func Exa3_ManTruthTablesV2(p *Exa3_Man_tV2) *vec.Vec_Wrd_t {
	p.vInfo = vec.Vec_WrdStart(int(p.nWords * (p.nObjs + 1)))
	for i := int32(0); i < p.nVars; i++ {
		util.Abc_TtIthVar(Exa3_ManTruthV2(p, i), i, p.nVars)
	}
	return p.vInfo
}
func Exa3_ManMarkupV2(p *Exa3_Man_tV2) int32 {
	p.iVar = 1 + p.LutMask*p.nNodes
	for i := p.nVars; i < p.nObjs; i++ {
		for k := int32(0); k < p.nLutSize; k++ {
			if p.pPars.FFewerVars != 0 && i == p.nObjs-1 && k == 0 {
				j := p.nObjs - 2
				vec.Vec_WecPush(p.vOutLits, int(j), bsat.Positive(p.iVar))
				p.VarMarks[i][k][j] = p.iVar
				p.iVar++
				continue
			}
			j := int32(0)
			if p.pPars.FFewerVars != 0 {
				j = p.nLutSize - 1 - k
			}
			for ; j < i-k; j++ {
				vec.Vec_WecPush(p.vOutLits, int(j), bsat.Positive(p.iVar))
				p.VarMarks[i][k][j] = p.iVar
				p.iVar++
			}
		}
	}
	fmt.Printf("The number of parameter variables = %d.\n", p.iVar)
	return p.iVar
}
func Exa3_ManInitPolarityFindVar(p *Exa3_Man_tV2, i, k int32, index *int32) int32 {
	for tries := int32(0); tries < p.nVars; tries++ {
		v := p.VarMarks[i][k][*index]
		*index = (*index + 1) % p.nVars
		if v > 0 {
			return v
		}
	}
	panic("bmc: no selectable LUT fanin")
}
func Exa3_ManInitPolarity(p *Exa3_Man_tV2) {
	for i := p.nVars; i < p.nObjs; i++ {
		p.pSat.SetPolarity(1+p.LutMask*(i-p.nVars)+p.LutMask-1, true)
	}
	var index int32
	for i := p.nVars; i < p.nObjs; i++ {
		min := int32(1)
		if i == p.nVars {
			min = 0
		}
		for k := p.nLutSize - 1; k >= min; k-- {
			p.pSat.SetPolarity(Exa3_ManInitPolarityFindVar(p, i, k, &index), true)
		}
		if min == 1 {
			v := p.VarMarks[i][0][i-1]
			if v <= 0 {
				panic("bmc: invalid fanin polarity")
			}
			p.pSat.SetPolarity(v, true)
		}
	}
}
func Exa3_ManAllocV2(pars *Bmc_EsPar_t, truth []uint64) *Exa3_Man_tV2 {
	if pars == nil || pars.NVars < 1 || pars.NVars > 10 || pars.NNodes < 1 || pars.NLutSize < 1 || pars.NLutSize > 6 || pars.NVars+pars.NNodes > MAJ_NOBJSV2 || int(util.Abc_TtWordNum(pars.NVars)) > len(truth) {
		panic("bmc: invalid LUT synthesis parameters")
	}
	p := &Exa3_Man_tV2{pPars: pars, nVars: pars.NVars, nNodes: pars.NNodes, nLutSize: pars.NLutSize, LutMask: (1 << pars.NLutSize) - 1, nObjs: pars.NVars + pars.NNodes, nWords: util.Abc_TtWordNum(pars.NVars), pTruth: truth}
	p.vOutLits = vec.Vec_WecStart(int(p.nObjs))
	Exa3_ManMarkupV2(p)
	Exa3_ManTruthTablesV2(p)
	p.pSat = bsat.NewSolver()
	for p.pSat.NumVars() < p.iVar {
		p.pSat.NewVariable()
	}
	return p
}
func Exa3_ManFreeV2(p *Exa3_Man_tV2) {
	if p != nil {
		p.pSat = nil
		p.vInfo = nil
		p.vOutLits = nil
	}
}
func Exa3_ManFindFaninV2(p *Exa3_Man_tV2, i, k int32) int32 {
	found := int32(-1)
	for j := int32(0); j < p.nObjs; j++ {
		v := p.VarMarks[i][k][j]
		if v != 0 {
			b, ok := p.pSat.Value(v)
			if !ok {
				panic("bmc: missing SAT model")
			}
			if b {
				if found >= 0 {
					panic("bmc: multiple fanins")
				}
				found = j
			}
		}
	}
	if found < 0 {
		panic("bmc: missing fanin")
	}
	return found
}
func Exa3_ManEvalV2(p *Exa3_Man_tV2) int32 {
	for i := p.nVars; i < p.nObjs; i++ {
		fans := make([][]uint64, p.nLutSize)
		for k := int32(0); k < p.nLutSize; k++ {
			fans[k] = Exa3_ManTruthV2(p, Exa3_ManFindFaninV2(p, i, k))
		}
		dst := Exa3_ManTruthV2(p, i)
		clear(dst)
		for k := int32(1); k <= p.LutMask; k++ {
			b, ok := p.pSat.Value(1 + p.LutMask*(i-p.nVars) + k - 1)
			if !ok {
				panic("bmc: missing SAT model")
			}
			if !b {
				continue
			}
			for w := range dst {
				term := ^uint64(0)
				for j := int32(0); j < p.nLutSize; j++ {
					v := fans[j][w]
					if k&(1<<j) == 0 {
						v = ^v
					}
					term &= v
				}
				dst[w] |= term
			}
		}
	}
	out := Exa3_ManTruthV2(p, p.nObjs-1)
	for bit := int32(0); bit < 1<<p.nVars; bit++ {
		if util.Abc_TtGetBit(out, bit) != util.Abc_TtGetBit(p.pTruth, bit) {
			return bit
		}
	}
	return -1
}
func Exa3_ManPrintSolutionV2(p *Exa3_Man_tV2, compl int32) {
	fmt.Printf("Realization of given %d-input function using %d %d-input LUTs:\n", p.nVars, p.nNodes, p.nLutSize)
	for i := p.nObjs - 1; i >= p.nVars; i-- {
		fmt.Printf("%02d = %d'b", i, 1<<p.nLutSize)
		for k := p.LutMask - 1; k >= 0; k-- {
			v, _ := p.pSat.Value(1 + p.LutMask*(i-p.nVars) + k)
			if i == p.nObjs-1 && compl != 0 {
				v = !v
			}
			if v {
				fmt.Print(1)
			} else {
				fmt.Print(0)
			}
		}
		if i == p.nObjs-1 && compl != 0 {
			fmt.Print("1(")
		} else {
			fmt.Print("0(")
		}
		for k := p.nLutSize - 1; k >= 0; k-- {
			j := Exa3_ManFindFaninV2(p, i, k)
			if j < p.nVars {
				fmt.Printf(" %c", 'a'+j)
			} else {
				fmt.Printf(" %02d", j)
			}
		}
		fmt.Println(" )")
	}
}

func Exa3_ManAddCnfStartV2(p *Exa3_Man_tV2, onlyAnd int32) int32 {
	add := func(lits ...bsat.Literal) bool { return p.pSat.AddClause(lits...) }
	for i := p.nVars; i < p.nObjs; i++ {
		start := int32(1) + p.LutMask*(i-p.nVars)
		for k := int32(0); k < p.nLutSize; k++ {
			var lits []bsat.Literal
			for j := int32(0); j < p.nObjs; j++ {
				if v := p.VarMarks[i][k][j]; v != 0 {
					lits = append(lits, bsat.Positive(v))
				}
			}
			if len(lits) == 0 || !add(lits...) {
				return 0
			}
			for n := 0; n < len(lits); n++ {
				for m := n + 1; m < len(lits); m++ {
					if !add(lits[n]^1, lits[m]^1) {
						return 0
					}
				}
			}
			if k == p.nLutSize-1 {
				break
			}
			for j := int32(0); j < p.nObjs; j++ {
				if a := p.VarMarks[i][k][j]; a != 0 {
					for n := j; n < p.nObjs; n++ {
						if b := p.VarMarks[i][k+1][n]; b != 0 && !add(bsat.Negative(a), bsat.Negative(b)) {
							return 0
						}
					}
				}
			}
		}
		if p.nLutSize == 2 {
			for k := int32(0); k < 3; k++ {
				a, b, c := bsat.Positive(start), bsat.Positive(start+1), bsat.Positive(start+2)
				if k == 1 {
					a ^= 1
				}
				if k == 2 {
					b ^= 1
				}
				if k != 0 {
					c ^= 1
				}
				if !add(a, b, c) {
					return 0
				}
			}
			if onlyAnd != 0 && !add(bsat.Negative(start), bsat.Negative(start+1), bsat.Positive(start+2)) {
				return 0
			}
		}
	}
	for i := int32(0); i < p.nObjs-1; i++ {
		if !add(vec.Vec_IntArray(vec.Vec_WecEntry(p.vOutLits, int(i)))...) {
			return 0
		}
	}
	return 1
}
func Exa3_ManAddCnfV2(p *Exa3_Man_tV2, mint int32) int32 {
	if mint < 0 || mint >= 1<<p.nVars {
		panic("bmc: invalid LUT minterm")
	}
	value := util.Abc_TtGetBit(p.pTruth, mint)
	for i := int32(0); i < p.nVars; i++ {
		p.VarVals[i] = mint >> i & 1
	}
	for p.pSat.NumVars() < p.iVar+(p.nLutSize+1)*p.nNodes {
		p.pSat.NewVariable()
	}
	for i := p.nVars; i < p.nObjs; i++ {
		start := int32(1) + p.LutMask*(i-p.nVars)
		base := p.iVar + (p.nLutSize+1)*(i-p.nVars)
		for k := int32(0); k < p.nLutSize; k++ {
			for j := int32(0); j < p.nObjs; j++ {
				mark := p.VarMarks[i][k][j]
				if mark == 0 {
					continue
				}
				for n := int32(0); n < 2; n++ {
					clause := []bsat.Literal{bsat.Negative(mark), bsat.Positive(base+k) ^ n}
					if j >= p.nVars {
						other := p.iVar + (p.nLutSize+1)*(j-p.nVars) + p.nLutSize
						clause = append(clause, bsat.Positive(other)^(n^1))
					} else if p.VarVals[j] == n {
						continue
					}
					if !p.pSat.AddClause(clause...) {
						return 0
					}
				}
			}
		}
		for n := int32(0); n < 2; n++ {
			if i == p.nObjs-1 && n == value {
				continue
			}
			for k := int32(0); k <= p.LutMask; k++ {
				if k == 0 && n == 1 {
					continue
				}
				var clause []bsat.Literal
				for j := int32(0); j < p.nLutSize; j++ {
					clause = append(clause, bsat.Positive(base+j)^(k>>j&1))
				}
				if i != p.nObjs-1 {
					clause = append(clause, bsat.Positive(base+p.nLutSize)^(n^1))
				}
				if k > 0 {
					clause = append(clause, bsat.Positive(start+k-1)^n)
				}
				if !p.pSat.AddClause(clause...) {
					return 0
				}
			}
		}
	}
	p.iVar += (p.nLutSize + 1) * p.nNodes
	return 1
}
func Exa3_ManExactSynthesisV2(pars *Bmc_EsPar_t) {
	if pars == nil || pars.NVars < 1 || pars.NVars > 10 {
		fmt.Println("Invalid synthesis parameters.")
		return
	}
	start := time.Now()
	truth := make([]uint64, util.Abc_TtWordNum(pars.NVars))
	util.Abc_TtReadHex(truth, pars.PTtStr)
	complement := int32(0)
	if truth[0]&1 != 0 {
		complement = 1
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	p := Exa3_ManAllocV2(pars, truth)
	if Exa3_ManAddCnfStartV2(p, pars.FOnlyAnd) == 0 {
		fmt.Println("The problem has no solution.")
		return
	}
	fmt.Printf("Running exact synthesis for %d-input function with %d %d-input LUTs...\n", p.nVars, p.nNodes, p.nLutSize)
	for mint := int32(1); mint >= 0; {
		if Exa3_ManAddCnfV2(p, mint) == 0 {
			fmt.Println("The problem has no solution.")
			break
		}
		status := p.pSat.Solve()
		if status != bsat.Satisfiable {
			fmt.Println("The problem has no solution.")
			break
		}
		mint = Exa3_ManEvalV2(p)
		if mint < 0 {
			Exa3_ManPrintSolutionV2(p, complement)
		}
	}
	fmt.Printf("Total runtime: %s\n", time.Since(start))
}
