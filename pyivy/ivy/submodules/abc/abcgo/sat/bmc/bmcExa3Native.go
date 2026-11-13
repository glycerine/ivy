package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Exa3_Man_t represents an exact synthesis instance for normal LUT networks.
// Every LUT maps the all-zero input assignment to zero; the output can be
// complemented globally when the target is not normal.
type Exa3_Man_t struct {
	pPars                                                 *Bmc_EsPar_t
	nVars, nNodes, nLutSize, LutMask, nObjs, nWords, iVar int32
	pTruth                                                []uint64
	info                                                  [][]uint64
	marks                                                 [][][]int32
	outLits                                               [][]int32
	used                                                  map[[5]int32]bool
	nUsed                                                 [2]int
	pSat                                                  *bsat.Solver
}

func Exa3_ManTruth(p *Exa3_Man_t, v int32) []uint64 { return p.info[v] }
func Exa3_ManIsUsed2(p *Exa3_Man_t, m, n, i, j int32) int32 {
	return exa3Used(p, [5]int32{m, n, i, j, -1})
}
func Exa3_ManIsUsed3(p *Exa3_Man_t, m, n, i, j, k int32) int32 {
	return exa3Used(p, [5]int32{m, n, i, j, k})
}
func exa3Used(p *Exa3_Man_t, key [5]int32) int32 {
	p.nUsed[0]++
	if p.used[key] {
		return 1
	}
	p.used[key] = true
	p.nUsed[1]++
	return 0
}
func Exa3_ManMarkup(p *Exa3_Man_t) int32 {
	p.iVar = 1 + p.LutMask*p.nNodes
	for i := p.nVars; i < p.nObjs; i++ {
		for k := int32(0); k < p.nLutSize; k++ {
			if p.pPars.fFewerVars != 0 && i == p.nObjs-1 && k == 0 {
				j := p.nObjs - 2
				p.marks[i][k][j] = p.iVar
				p.outLits[j] = append(p.outLits[j], bsat.Positive(p.iVar))
				p.iVar++
				continue
			}
			start := int32(0)
			if p.pPars.fFewerVars != 0 {
				start = p.nLutSize - 1 - k
			}
			for j := start; j < i-k; j++ {
				p.marks[i][k][j] = p.iVar
				p.outLits[j] = append(p.outLits[j], bsat.Positive(p.iVar))
				p.iVar++
			}
		}
	}
	return p.iVar
}
func Exa3_ManAlloc(params *Bmc_EsPar_t, truth []uint64) *Exa3_Man_t {
	if params == nil || params.nVars < 1 || params.nVars > 10 || params.nNodes < 1 || params.nLutSize < 1 || params.nLutSize > 6 || params.nLutSize > params.nVars {
		panic("bmc: invalid LUT synthesis parameters")
	}
	p := &Exa3_Man_t{pPars: params, nVars: params.nVars, nNodes: params.nNodes, nLutSize: params.nLutSize, LutMask: (1 << params.nLutSize) - 1, nObjs: params.nVars + params.nNodes, nWords: int32((1<<params.nVars + 63) / 64), pTruth: append([]uint64(nil), truth...), used: make(map[[5]int32]bool), pSat: bsat.NewSolver()}
	if len(p.pTruth) < int(p.nWords) {
		panic("bmc: short truth table")
	}
	p.marks = make([][][]int32, p.nObjs)
	p.outLits = make([][]int32, p.nObjs)
	p.info = make([][]uint64, p.nObjs+1)
	for i := range p.marks {
		p.marks[i] = make([][]int32, p.nLutSize)
		for k := range p.marks[i] {
			p.marks[i][k] = make([]int32, p.nObjs)
		}
	}
	for i := range p.info {
		p.info[i] = make([]uint64, p.nWords)
	}
	for i := int32(0); i < p.nVars; i++ {
		for m := int32(0); m < 1<<p.nVars; m++ {
			if m&(1<<i) != 0 {
				p.info[i][m>>6] |= uint64(1) << uint(m&63)
			}
		}
	}
	Exa3_ManMarkup(p)
	for p.pSat.NumVars() < p.iVar {
		p.pSat.NewVariable()
	}
	if params.RuntimeLim > 0 {
		p.pSat.SetDeadline(time.Now().Add(time.Duration(params.RuntimeLim) * time.Second))
	}
	return p
}
func Exa3_ManFree(p *Exa3_Man_t) {
	if p != nil {
		p.pSat = nil
		p.info = nil
		p.marks = nil
		p.outLits = nil
		p.used = nil
	}
}
func Exa3_ManFindFanin(p *Exa3_Man_t, i, k int32) int32 {
	found := int32(-1)
	for j, v := range p.marks[i][k] {
		if v == 0 {
			continue
		}
		value, ok := p.pSat.Value(v)
		if !ok {
			panic("bmc: missing synthesis model")
		}
		if value {
			if found >= 0 {
				panic("bmc: multiple LUT fanins")
			}
			found = int32(j)
		}
	}
	if found < 0 {
		panic("bmc: no LUT fanin")
	}
	return found
}
func Exa3_ManEval(p *Exa3_Man_t) int32 {
	for i := p.nVars; i < p.nObjs; i++ {
		clear(p.info[i])
		fans := make([]int32, p.nLutSize)
		for k := range fans {
			fans[k] = Exa3_ManFindFanin(p, i, int32(k))
		}
		start := 1 + p.LutMask*(i-p.nVars)
		for m := int32(0); m < 1<<p.nVars; m++ {
			lut := int32(0)
			for k, j := range fans {
				if p.info[j][m>>6]&(uint64(1)<<uint(m&63)) != 0 {
					lut |= 1 << k
				}
			}
			if lut > 0 {
				value, _ := p.pSat.Value(start + lut - 1)
				if value {
					p.info[i][m>>6] |= uint64(1) << uint(m&63)
				}
			}
		}
	}
	for m := int32(0); m < 1<<p.nVars; m++ {
		if (p.info[p.nObjs-1][m>>6]^p.pTruth[m>>6])&(uint64(1)<<uint(m&63)) != 0 {
			return m
		}
	}
	return -1
}
func exa3Clause(p *Exa3_Man_t, lits ...int32) bool { return p.pSat.AddClause(lits...) }
func exa3Lit(v int32, neg bool) int32 {
	if neg {
		return bsat.Negative(v)
	}
	return bsat.Positive(v)
}
func Exa3_ManAddCnfStart(p *Exa3_Man_t, onlyAnd int32) int32 {
	for i := p.nVars; i < p.nObjs; i++ {
		start := 1 + p.LutMask*(i-p.nVars)
		for k := int32(0); k < p.nLutSize; k++ {
			var choices []int32
			for _, v := range p.marks[i][k] {
				if v != 0 {
					choices = append(choices, bsat.Positive(v))
				}
			}
			if len(choices) == 0 || !exa3Clause(p, choices...) {
				return 0
			}
			for a := range choices {
				for b := a + 1; b < len(choices); b++ {
					if !exa3Clause(p, choices[a]^1, choices[b]^1) {
						return 0
					}
				}
			}
			if k == p.nLutSize-1 {
				continue
			}
			for j, v := range p.marks[i][k] {
				if v == 0 {
					continue
				}
				for n := j; n < int(p.nObjs); n++ {
					v2 := p.marks[i][k+1][n]
					if v2 != 0 && !exa3Clause(p, bsat.Negative(v), bsat.Negative(v2)) {
						return 0
					}
				}
			}
		}
		if p.pPars.fUseIncr != 0 {
			for j := p.nVars; j < i; j++ {
				for n, v := range p.marks[i][0] {
					if v == 0 {
						continue
					}
					for m := n + 1; m < int(p.nObjs); m++ {
						v2 := p.marks[j][0][m]
						if v2 != 0 && !exa3Clause(p, bsat.Negative(v), bsat.Negative(v2)) {
							return 0
						}
					}
				}
			}
		}
		if p.nLutSize == 2 {
			for k := 0; k < 3; k++ {
				if !exa3Clause(p, exa3Lit(start, k == 1), exa3Lit(start+1, k == 2), exa3Lit(start+2, k != 0)) {
					return 0
				}
			}
			if onlyAnd != 0 && !exa3Clause(p, bsat.Negative(start), bsat.Negative(start+1), bsat.Positive(start+2)) {
				return 0
			}
		}
	}
	for i := int32(0); i < p.nObjs-1; i++ {
		if len(p.outLits[i]) == 0 || !exa3Clause(p, p.outLits[i]...) {
			return 0
		}
	}
	return 1
}
func Exa3_ManAddCnf(p *Exa3_Man_t, mint int32) int32 {
	value := int32((p.pTruth[mint>>6] >> uint(mint&63)) & 1)
	end := p.iVar + (p.nLutSize+1)*p.nNodes
	for p.pSat.NumVars() < end {
		p.pSat.NewVariable()
	}
	for i := p.nVars; i < p.nObjs; i++ {
		start := 1 + p.LutMask*(i-p.nVars)
		base := p.iVar + (p.nLutSize+1)*(i-p.nVars)
		for k := int32(0); k < p.nLutSize; k++ {
			for j, v := range p.marks[i][k] {
				if v == 0 {
					continue
				}
				for n := int32(0); n < 2; n++ {
					clause := []int32{bsat.Negative(v), exa3Lit(base+k, n != 0)}
					if int32(j) >= p.nVars {
						jb := p.iVar + (p.nLutSize+1)*(int32(j)-p.nVars)
						clause = append(clause, exa3Lit(jb+p.nLutSize, n == 0))
					} else if (mint>>j)&1 == n {
						continue
					}
					if !exa3Clause(p, clause...) {
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
				clause := make([]int32, 0, p.nLutSize+2)
				for j := int32(0); j < p.nLutSize; j++ {
					clause = append(clause, exa3Lit(base+j, (k>>j)&1 != 0))
				}
				if i != p.nObjs-1 {
					clause = append(clause, exa3Lit(base+p.nLutSize, n == 0))
				}
				if k > 0 {
					clause = append(clause, exa3Lit(start+k-1, n != 0))
				}
				if !exa3Clause(p, clause...) {
					return 0
				}
			}
		}
	}
	p.iVar = end
	return 1
}
func Exa3_ManAddCnf2(p *Exa3_Man_t, mint int32) int32 {
	base := p.iVar + p.nNodes*mint
	allFans := make([][]int32, p.nNodes)
	for i := p.nVars; i < p.nObjs; i++ {
		allFans[i-p.nVars] = make([]int32, p.nLutSize)
		for j := range allFans[i-p.nVars] {
			allFans[i-p.nVars][j] = Exa3_ManFindFanin(p, i, int32(j))
		}
	}
	for p.pSat.NumVars() < base+p.nNodes {
		p.pSat.NewVariable()
	}
	value := int32((p.pTruth[mint>>6] >> uint(mint&63)) & 1)
	for i := p.nVars; i < p.nObjs; i++ {
		start := 1 + p.LutMask*(i-p.nVars)
		fans := allFans[i-p.nVars]
		if p.nLutSize == 2 && Exa3_ManIsUsed2(p, mint, i, fans[1], fans[0]) != 0 {
			continue
		}
		if p.nLutSize == 3 && Exa3_ManIsUsed3(p, mint, i, fans[2], fans[1], fans[0]) != 0 {
			continue
		}
		for n := int32(0); n < 2; n++ {
			if i == p.nObjs-1 && n == value {
				continue
			}
			for k := int32(0); k <= p.LutMask; k++ {
				if k == 0 && n == 1 {
					continue
				}
				var clause []int32
				valid := true
				for j, fanin := range fans {
					clause = append(clause, bsat.Negative(p.marks[i][j][fanin]))
					bit := (k >> j) & 1
					if fanin >= p.nVars {
						clause = append(clause, exa3Lit(base+fanin-p.nVars, bit != 0))
					} else if (mint>>fanin)&1 != bit {
						valid = false
						break
					}
				}
				if !valid {
					continue
				}
				if i != p.nObjs-1 {
					clause = append(clause, exa3Lit(base+i-p.nVars, n == 0))
				}
				if k > 0 {
					clause = append(clause, exa3Lit(start+k-1, n != 0))
				}
				if !exa3Clause(p, clause...) {
					return 0
				}
			}
		}
	}
	return 1
}
func Exa3_ManPrintSolution(p *Exa3_Man_t, complement int32) {
	fmt.Printf("Realization of given %d-input function using %d %d-input LUTs:\n", p.nVars, p.nNodes, p.nLutSize)
	for i := p.nObjs - 1; i >= p.nVars; i-- {
		start := 1 + p.LutMask*(i-p.nVars)
		fmt.Printf("%02d = %d'b", i, 1<<p.nLutSize)
		for k := p.LutMask - 1; k >= 0; k-- {
			value, _ := p.pSat.Value(start + k)
			if i == p.nObjs-1 && complement != 0 {
				value = !value
			}
			if value {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
		}
		if i == p.nObjs-1 && complement != 0 {
			fmt.Print("1(")
		} else {
			fmt.Print("0(")
		}
		for k := p.nLutSize - 1; k >= 0; k-- {
			fan := Exa3_ManFindFanin(p, i, k)
			if fan < p.nVars {
				fmt.Printf(" %c", 'a'+fan)
			} else {
				fmt.Printf(" %02d", fan)
			}
		}
		fmt.Println(" )")
	}
}
func Exa3_ManPrint(p *Exa3_Man_t, iteration, mint int32, elapsed int64) {
	stats := p.pSat.Statistics()
	fmt.Printf("Iter%6d : %0*b Var =%5d Cla =%6d Conf =%9d Time = %.2f sec\n", iteration, p.nVars, mint, stats.Variables, stats.Clauses, stats.Conflicts, time.Duration(elapsed).Seconds())
}
func Exa3_ManExactSynthesis(params *Bmc_EsPar_t) {
	if params == nil || params.nVars < 1 || params.nVars > 10 {
		fmt.Println("Invalid LUT synthesis parameters.")
		return
	}
	truth := make([]uint64, 16)
	util.Abc_TtReadHex(truth, params.pTtStr)
	complement := int32(0)
	if truth[0]&1 != 0 {
		complement = 1
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	p := Exa3_ManAlloc(params, truth)
	defer Exa3_ManFree(p)
	start := time.Now()
	if Exa3_ManAddCnfStart(p, params.fOnlyAnd) == 0 {
		fmt.Println("The problem has no solution.")
		return
	}
	fmt.Printf("Running exact synthesis for %d-input function with %d %d-input LUTs...\n", p.nVars, p.nNodes, p.nLutSize)
	status := bsat.Unknown
	mint := int32(1)
	iteration := int32(0)
	if params.fUseIncr != 0 {
		for p.pSat.NumVars() < p.iVar+p.nNodes*(1<<p.nVars) {
			p.pSat.NewVariable()
		}
		status = p.pSat.Solve()
		if status != bsat.Satisfiable {
			fmt.Println("The problem has no solution.")
			return
		}
	}
	for ; mint != -1; iteration++ {
		added := int32(0)
		if params.fUseIncr != 0 {
			added = Exa3_ManAddCnf2(p, mint)
		} else {
			added = Exa3_ManAddCnf(p, mint)
		}
		if added == 0 {
			status = bsat.Unsatisfiable
			break
		}
		status = p.pSat.Solve()
		if params.fVerbose != 0 && (params.fUseIncr == 0 || iteration%100 == 0) {
			Exa3_ManPrint(p, iteration, mint, int64(time.Since(start)))
		}
		if status != bsat.Satisfiable {
			break
		}
		mint = Exa3_ManEval(p)
	}
	if mint == -1 {
		Exa3_ManPrintSolution(p, complement)
	} else if status == bsat.Unknown {
		fmt.Printf("The problem timed out after %d sec.\n", params.RuntimeLim)
	} else {
		fmt.Println("The problem has no solution.")
	}
	fmt.Printf("Added = %d. Tried = %d. Total runtime: %s\n", p.nUsed[1], p.nUsed[0], time.Since(start))
}
