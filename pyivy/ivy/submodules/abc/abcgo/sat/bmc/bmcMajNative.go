package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/bits"
	"time"
)

const MAJ_NOBJS = 32

// Maj_Man_t stores the structural and minterm SAT variables for MAJ3 synthesis.
type Maj_Man_t struct {
	nVars, nNodes, nObjs, nWords, iVar, fUseConst, fUseLine int32
	vInfo                                                   [][]uint64
	VarMarks                                                [MAJ_NOBJS][3][MAJ_NOBJS]int32
	VarVals                                                 [MAJ_NOBJS + 2]int32
	vOutLits                                                [][]bsat.Literal
	pSat                                                    *bsat.Solver
}

func Maj_ManTruth(p *Maj_Man_t, v int32) []uint64 { return p.vInfo[v] }
func Maj_ManValue(m, n int32) int32 {
	if bits.OnesCount32(uint32(m)&((uint32(1)<<n)-1)) > int(n/2) {
		return 1
	}
	return 0
}
func Maj_ManTruthTables(p *Maj_Man_t) [][]uint64 {
	info := make([][]uint64, p.nObjs+1)
	for i := range info {
		info[i] = make([]uint64, p.nWords)
	}
	p.vInfo = info
	for i := range info[1] {
		info[1][i] = ^uint64(0)
	}
	nbits := int32(1) << p.nVars
	if nbits < 64 {
		nbits = 64
	}
	for i := int32(0); i < nbits; i++ {
		for v := int32(0); v < p.nVars; v++ {
			if i&(1<<v) != 0 {
				info[v+2][i>>6] |= uint64(1) << (i & 63)
			}
		}
		if Maj_ManValue(i, p.nVars) != 0 {
			info[p.nObjs][i>>6] |= uint64(1) << (i & 63)
		}
	}
	return info
}
func Maj_ManMarkup(p *Maj_Man_t) int32 {
	p.iVar = 1
	i := p.nVars + 2
	for k := int32(0); k < 3; k++ {
		j := int32(4) - k
		p.vOutLits[j] = append(p.vOutLits[j], bsat.Positive(p.iVar))
		p.VarMarks[i][k][j] = p.iVar
		p.iVar++
	}
	for i = p.nVars + 3; i < p.nObjs; i++ {
		for k := int32(0); k < 3; k++ {
			if p.fUseLine != 0 && k == 0 {
				j := i - 1
				p.vOutLits[j] = append(p.vOutLits[j], bsat.Positive(p.iVar))
				p.VarMarks[i][k][j] = p.iVar
				p.iVar++
				continue
			}
			start := int32(2)
			if p.fUseConst != 0 && k == 2 {
				start = 0
			}
			for j := start; j < i-k; j++ {
				p.vOutLits[j] = append(p.vOutLits[j], bsat.Positive(p.iVar))
				p.VarMarks[i][k][j] = p.iVar
				p.iVar++
			}
		}
	}
	return p.iVar
}
func Maj_ManAlloc(vars, nodes, useConst, useLine int32) *Maj_Man_t {
	if vars < 3 || vars > 20 || nodes < 1 || 2+vars+nodes > MAJ_NOBJS {
		panic("bmc: invalid majority synthesis size")
	}
	p := &Maj_Man_t{nVars: vars, nNodes: nodes, nObjs: 2 + vars + nodes, fUseConst: useConst, fUseLine: useLine, nWords: 1, pSat: bsat.NewSolver()}
	if vars > 6 {
		p.nWords = 1 << (vars - 6)
	}
	p.vOutLits = make([][]bsat.Literal, p.nObjs)
	Maj_ManMarkup(p)
	p.VarVals[1] = 1
	Maj_ManTruthTables(p)
	for p.pSat.NumVars() < p.iVar {
		p.pSat.NewVariable()
	}
	return p
}
func Maj_ManFree(p *Maj_Man_t) {
	if p != nil {
		p.vInfo = nil
		p.vOutLits = nil
		p.pSat = nil
	}
}
func Maj_ManFindFanin(p *Maj_Man_t, i, k int32) int32 {
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
func Maj_ManEval(p *Maj_Man_t) int32 {
	for i := p.nVars + 2; i < p.nObjs; i++ {
		a, b, c := Maj_ManTruth(p, Maj_ManFindFanin(p, i, 0)), Maj_ManTruth(p, Maj_ManFindFanin(p, i, 1)), Maj_ManTruth(p, Maj_ManFindFanin(p, i, 2))
		for w := range p.vInfo[i] {
			p.vInfo[i][w] = (a[w] & b[w]) | (a[w] & c[w]) | (b[w] & c[w])
		}
	}
	for i := int32(0); i < 1<<p.nVars; i++ {
		ones := int32(bits.OnesCount32(uint32(i)))
		if ones < p.nVars/2 || ones > p.nVars/2+1 {
			continue
		}
		if ((p.vInfo[p.nObjs][i>>6]^p.vInfo[p.nObjs-1][i>>6])>>(i&63))&1 != 0 {
			return i
		}
	}
	return -1
}
func Maj_ManPrintSolution(p *Maj_Man_t) {
	fmt.Printf("Realization of %d-input majority using %d MAJ3 gates:\n", p.nVars, p.nNodes)
	for i := p.nObjs - 1; i >= p.nVars+2; i-- {
		fmt.Printf("%02d = MAJ(", i-2)
		for k := int32(2); k >= 0; k-- {
			v := Maj_ManFindFanin(p, i, k)
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
func Maj_ManAddCnfStart(p *Maj_Man_t) int32 {
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
			if k == 2 {
				break
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
		if !p.pSat.AddClause(p.vOutLits[i]...) {
			return 0
		}
	}
	return 1
}
func Maj_ManAddCnf(p *Maj_Man_t, mint int32) int32 {
	value := Maj_ManValue(mint, p.nVars)
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
func Maj_ManExactSynthesis(vars, nodes, useConst, useLine, verbose int32) int32 {
	start := time.Now()
	p := Maj_ManAlloc(vars, nodes, useConst, useLine)
	defer Maj_ManFree(p)
	if Maj_ManAddCnfStart(p) == 0 {
		return 0
	}
	fmt.Printf("Running exact synthesis for %d-input majority with %d MAJ3 gates...\n", vars, nodes)
	for iteration, mint := int32(0), int32(0); ; iteration++ {
		if Maj_ManAddCnf(p, mint) == 0 {
			return 0
		}
		status := p.pSat.Solve()
		if verbose != 0 {
			fmt.Printf("Iter %d : minterm %d, variables %d\n", iteration, mint, p.iVar)
		}
		if status != bsat.Satisfiable {
			return 0
		}
		mint = Maj_ManEval(p)
		if mint < 0 {
			Maj_ManPrintSolution(p)
			fmt.Printf("Total runtime: %s\n", time.Since(start))
			return 1
		}
	}
}
