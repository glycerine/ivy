package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/bits"
)

const MAJ_NOBJSV2 = 32

type Maj_Man_tV2 struct {
	nVars, nNodes, nObjs, nWords, iVar              int32
	fUseConst, fUseLine, fUseRand, nRands, fVerbose int32
	vInfo                                           *vec.Vec_Wrd_t
	VarMarks                                        [MAJ_NOBJSV2][3][MAJ_NOBJSV2]int32
	VarVals                                         [MAJ_NOBJSV2 + 2]int32
	vOutLits                                        *vec.Vec_Wec_t
	pSat                                            *bsat.Solver
}

func Maj_ManTruthV2(p *Maj_Man_tV2, v int32) []uint64 {
	if p == nil || v < 0 || v > p.nObjs {
		panic("bmc: invalid majority truth table")
	}
	start := int(p.nWords * v)
	return vec.Vec_WrdArray(p.vInfo)[start : start+int(p.nWords)]
}
func Maj_ManTruthTablesV2(p *Maj_Man_tV2) *vec.Vec_Wrd_t {
	p.vInfo = vec.Vec_WrdStart(int(p.nWords * (p.nObjs + 1)))
	util.Abc_TtFill(Maj_ManTruthV2(p, 1), p.nWords)
	for i := int32(0); i < p.nVars; i++ {
		util.Abc_TtIthVar(Maj_ManTruthV2(p, i+2), i, p.nVars)
	}
	nMints := int32(64)
	if (int32(1) << uint(p.nVars)) > nMints {
		nMints = int32(1) << uint(p.nVars)
	}
	for i := int32(0); i < nMints; i++ {
		if Maj_ManValueV2(i, p.nVars) != 0 {
			util.Abc_TtSetBit(Maj_ManTruthV2(p, p.nObjs), i)
		}
	}
	return p.vInfo
}

func Maj_ManMarkupV2(p *Maj_Man_tV2) int32 {
	var connections [MAJ_NOBJSV2][3]int32
	if p.nObjs > MAJ_NOBJSV2 || p.nVars < 3 || p.nNodes < 1 {
		panic("bmc: invalid majority synthesis size")
	}
	if p.fUseRand != 0 {
		Maj_ManConnect(connections[:], p.nVars, p.nObjs, p.nRands, p.fVerbose)
	}
	p.iVar = 1
	add := func(i, k, j int32) {
		vec.Vec_WecPush(p.vOutLits, int(j), p.iVar<<1)
		p.VarMarks[i][k][j] = p.iVar
		p.iVar++
	}
	first := p.nVars + 2
	for k := int32(0); k < 3; k++ {
		add(first, k, 4-k)
	}
	for i := first + 1; i < p.nObjs; i++ {
		for k := int32(0); k < 3; k++ {
			if p.fUseLine != 0 && k == 0 {
				add(i, k, i-1)
				continue
			}
			if p.fUseRand != 0 && connections[i][k] > 0 {
				add(i, k, connections[i][k])
				continue
			}
			start := int32(2)
			if p.fUseConst != 0 && k == 2 {
				start = 0
			}
			limit := i - k
			if p.fUseRand != 0 {
				limit = p.nVars + 2 - k
			}
			for j := start; j < limit; j++ {
				add(i, k, j)
			}
		}
	}
	return p.iVar
}
func Maj_ManAllocV2(nVars, nNodes, useConst, useLine, useRand, nRands, verbose int32) *Maj_Man_tV2 {
	if nVars < 3 || nVars > 20 || nNodes < 1 || nVars+nNodes+2 > MAJ_NOBJSV2 {
		panic("bmc: invalid majority synthesis size")
	}
	p := &Maj_Man_tV2{nVars: nVars, nNodes: nNodes, nObjs: 2 + nVars + nNodes, nWords: util.Abc_TtWordNum(nVars), fUseConst: useConst, fUseLine: useLine, fUseRand: useRand, nRands: nRands, fVerbose: verbose, pSat: bsat.NewSolver()}
	p.vOutLits = vec.Vec_WecStart(int(p.nObjs))
	Maj_ManMarkupV2(p)
	p.VarVals[1] = 1
	Maj_ManTruthTablesV2(p)
	for p.pSat.NumVars() < p.iVar {
		p.pSat.NewVariable()
	}
	return p
}
func Maj_ManFreeV2(p *Maj_Man_tV2) {
	if p != nil {
		p.vInfo = nil
		p.vOutLits = nil
		p.pSat = nil
	}
}

func Maj_ManFindFaninV2(p *Maj_Man_tV2, i, k int32) int32 {
	found := int32(-1)
	for j := int32(0); j < p.nObjs; j++ {
		v := p.VarMarks[i][k][j]
		if v != 0 {
			value, ok := p.pSat.Value(v)
			if !ok {
				panic("bmc: missing SAT model")
			}
			if value {
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
func Maj_ManEvalV2(p *Maj_Man_tV2) int32 {
	for i := p.nVars + 2; i < p.nObjs; i++ {
		a, b, c := Maj_ManTruthV2(p, Maj_ManFindFaninV2(p, i, 0)), Maj_ManTruthV2(p, Maj_ManFindFaninV2(p, i, 1)), Maj_ManTruthV2(p, Maj_ManFindFaninV2(p, i, 2))
		for w := range Maj_ManTruthV2(p, i) {
			Maj_ManTruthV2(p, i)[w] = (a[w] & b[w]) | (a[w] & c[w]) | (b[w] & c[w])
		}
	}
	for i := int32(0); i < 1<<p.nVars; i++ {
		ones := int32(bits.OnesCount32(uint32(i)))
		if ones < p.nVars/2 || ones > p.nVars/2+1 {
			continue
		}
		if ((Maj_ManTruthV2(p, p.nObjs)[i>>6]^Maj_ManTruthV2(p, p.nObjs-1)[i>>6])>>(i&63))&1 != 0 {
			return i
		}
	}
	return -1
}
func Maj_ManPrintSolutionV2(p *Maj_Man_tV2) {
	fmt.Printf("Realization of %d-input majority using %d MAJ3 gates:\n", p.nVars, p.nNodes)
	for i := p.nObjs - 1; i >= p.nVars+2; i-- {
		fmt.Printf("%02d = MAJ(", i-2)
		for k := int32(2); k >= 0; k-- {
			v := Maj_ManFindFaninV2(p, i, k)
			if v >= 2 && v < p.nVars+2 {
				fmt.Printf(" %c", 'a'+v-2)
			} else if v < 2 {
				fmt.Printf(" %d", v)
			} else {
				fmt.Printf(" %02d", v-2)
			}
		}
		fmt.Println(" )")
	}
}
func Maj_ManAddCnfStartV2(p *Maj_Man_tV2) int32 {
	for i := p.nVars + 2; i < p.nObjs; i++ {
		for k := int32(0); k < 3; k++ {
			var lits []bsat.Literal
			for j := int32(0); j < p.nObjs; j++ {
				if v := p.VarMarks[i][k][j]; v != 0 {
					lits = append(lits, bsat.Positive(v))
				}
			}
			if !p.pSat.AddClause(lits...) {
				return 0
			}
			for n := range lits {
				for m := n + 1; m < len(lits); m++ {
					if !p.pSat.AddClause(lits[n]^1, lits[m]^1) {
						return 0
					}
				}
			}
			if k == 2 || p.VarMarks[i][k][2] == 0 || p.VarMarks[i][k+1][2] == 0 {
				continue
			}
			for j := int32(0); j < p.nObjs; j++ {
				a := p.VarMarks[i][k][j]
				if a == 0 {
					continue
				}
				for n := j; n < p.nObjs; n++ {
					b := p.VarMarks[i][k+1][n]
					if b != 0 && !p.pSat.AddClause(bsat.Negative(a), bsat.Negative(b)) {
						return 0
					}
				}
			}
		}
	}
	for i := int32(2); i < p.nObjs-1; i++ {
		if !p.pSat.AddClause(vec.Vec_IntArray(vec.Vec_WecEntry(p.vOutLits, int(i)))...) {
			return 0
		}
	}
	return 1
}
func Maj_ManAddCnfV2(p *Maj_Man_tV2, mint int32) int32 {
	value := Maj_ManValueV2(mint, p.nVars)
	for i := int32(0); i < p.nVars; i++ {
		p.VarVals[i+2] = (mint >> i) & 1
	}
	for p.pSat.NumVars() < p.iVar+4*p.nNodes {
		p.pSat.NewVariable()
	}
	for i := p.nVars + 2; i < p.nObjs; i++ {
		base := p.iVar + 4*(i-p.nVars-2)
		for k := int32(0); k < 3; k++ {
			for j := int32(0); j < p.nObjs; j++ {
				mark := p.VarMarks[i][k][j]
				if mark == 0 {
					continue
				}
				other := p.iVar + 4*(j-p.nVars-2)
				for n := int32(0); n < 2; n++ {
					clause := []bsat.Literal{bsat.Negative(mark), bsat.Positive(base+k) ^ n}
					if j >= p.nVars+2 {
						clause = append(clause, bsat.Positive(other+3)^(1-n))
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
			for k := int32(0); k < 3; k++ {
				var clause []bsat.Literal
				for j := int32(0); j < 3; j++ {
					if j != k {
						clause = append(clause, bsat.Positive(base+j)^n)
					}
				}
				if i != p.nObjs-1 {
					clause = append(clause, bsat.Positive(base+3)^(1-n))
				}
				if !p.pSat.AddClause(clause...) {
					return 0
				}
			}
		}
	}
	p.iVar += 4 * p.nNodes
	return 1
}

func Maj_ManExactSynthesis2(vars, nodes, useConst, useLine, useRand, nRands, verbose int32) int32 {
	p := Maj_ManAllocV2(vars, nodes, useConst, useLine, useRand, nRands, verbose)
	defer Maj_ManFreeV2(p)
	if Maj_ManAddCnfStartV2(p) == 0 {
		return 0
	}
	for mint := int32(0); ; {
		if Maj_ManAddCnfV2(p, mint) == 0 || p.pSat.Solve() != bsat.Satisfiable {
			return 0
		}
		mint = Maj_ManEvalV2(p)
		if mint < 0 {
			Maj_ManPrintSolutionV2(p)
			return 1
		}
	}
}
func Maj_ManExactSynthesisTest() int32 {
	for Maj_ManExactSynthesis2(9, 10, 0, 0, 1, 3, 0) == 0 {
	}
	return 1
}
