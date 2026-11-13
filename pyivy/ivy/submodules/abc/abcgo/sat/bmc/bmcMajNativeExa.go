package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"strings"
	"time"
)

type Exa_Man_t struct {
	pPars                              *Bmc_EsPar_t
	nVars, nNodes, nObjs, nWords, iVar int32
	pTruth                             []uint64
	vInfo                              [][]uint64
	VarMarks                           [MAJ_NOBJS][2][MAJ_NOBJS]int32
	VarVals                            [MAJ_NOBJS]int32
	vOutLits                           [][]bsat.Literal
	pSat                               *bsat.Solver
}

func Exa_ManTruth(p *Exa_Man_t, v int32) []uint64 { return p.vInfo[v] }
func Exa_ManTruthTables(p *Exa_Man_t) [][]uint64 {
	p.vInfo = make([][]uint64, p.nObjs+1)
	for i := range p.vInfo {
		p.vInfo[i] = make([]uint64, p.nWords)
	}
	for i := int32(0); i < p.nVars; i++ {
		util.Abc_TtIthVar(p.vInfo[i], i, p.nVars)
	}
	return p.vInfo
}
func Exa_ManMarkup(p *Exa_Man_t) int32 {
	p.iVar = 1 + p.nNodes*3
	for i := p.nVars; i < p.nObjs; i++ {
		for k := int32(0); k < 2; k++ {
			if p.pPars.fFewerVars != 0 && i == p.nObjs-1 && k == 0 {
				j := p.nObjs - 2
				p.vOutLits[j] = append(p.vOutLits[j], bsat.Positive(p.iVar))
				p.VarMarks[i][k][j] = p.iVar
				p.iVar++
				continue
			}
			start := int32(0)
			if p.pPars.fFewerVars != 0 {
				start = 1 - k
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
func Exa_ManAlloc(pars *Bmc_EsPar_t, truth []uint64) *Exa_Man_t {
	if pars == nil || pars.nVars < 2 || pars.nVars > 10 || pars.nNodes < 1 || pars.nVars+pars.nNodes > MAJ_NOBJS {
		panic("bmc: invalid exact synthesis size")
	}
	p := &Exa_Man_t{pPars: pars, nVars: pars.nVars, nNodes: pars.nNodes, nObjs: pars.nVars + pars.nNodes, nWords: util.Abc_TtWordNum(pars.nVars), pTruth: truth, pSat: bsat.NewSolver()}
	if len(truth) < int(p.nWords) {
		panic("bmc: short truth table")
	}
	p.vOutLits = make([][]bsat.Literal, p.nObjs)
	Exa_ManMarkup(p)
	Exa_ManTruthTables(p)
	for p.pSat.NumVars() < p.iVar {
		p.pSat.NewVariable()
	}
	if pars.RuntimeLim > 0 {
		p.pSat.SetDeadline(time.Now().Add(time.Duration(pars.RuntimeLim) * time.Second))
	}
	return p
}
func Exa_ManFree(p *Exa_Man_t) {
	if p != nil {
		p.pSat = nil
		p.vInfo = nil
		p.vOutLits = nil
	}
}
func Exa_ManFindFanin(p *Exa_Man_t, i, k int32) int32 {
	found := int32(-1)
	for j := int32(0); j < p.nObjs; j++ {
		v := p.VarMarks[i][k][j]
		if v != 0 {
			value, ok := p.pSat.Value(v)
			if !ok {
				panic("bmc: missing solver model")
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
func Exa_ManEval(p *Exa_Man_t) int32 {
	for i := p.nVars; i < p.nObjs; i++ {
		a, b := p.vInfo[Exa_ManFindFanin(p, i, 0)], p.vInfo[Exa_ManFindFanin(p, i, 1)]
		clear(p.vInfo[i])
		base := 1 + 3*(i-p.nVars)
		for k := int32(1); k < 4; k++ {
			v, ok := p.pSat.Value(base + k - 1)
			if !ok {
				panic("bmc: missing solver model")
			}
			if !v {
				continue
			}
			for w := range a {
				x, y := a[w], b[w]
				if k&1 == 0 {
					x = ^x
				}
				if k&2 == 0 {
					y = ^y
				}
				p.vInfo[i][w] |= x & y
			}
		}
	}
	for i := int32(0); i < 1<<p.nVars; i++ {
		if ((p.vInfo[p.nObjs-1][i>>6]^p.pTruth[i>>6])>>(i&63))&1 != 0 {
			return i
		}
	}
	return -1
}
func exaFunction(p *Exa_Man_t, i int32) [4]bool {
	var f [4]bool
	for k := int32(1); k < 4; k++ {
		value, ok := p.pSat.Value(1 + 3*(i-p.nVars) + k - 1)
		if !ok {
			panic("bmc: missing solver model")
		}
		f[k] = value
	}
	return f
}
func Exa_ManPrintSolution(p *Exa_Man_t, compl int32) {
	fmt.Printf("Realization of given %d-input function using %d two-input gates:\n", p.nVars, p.nNodes)
	for i := p.nObjs - 1; i >= p.nVars; i-- {
		f := exaFunction(p, i)
		if i == p.nObjs-1 && compl != 0 {
			for k := range f {
				f[k] = !f[k]
			}
		}
		fmt.Printf("%02d = 4'b", i)
		for k := 3; k >= 0; k-- {
			if f[k] {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
		}
		fmt.Print("(")
		for k := int32(1); k >= 0; k-- {
			v := Exa_ManFindFanin(p, i, k)
			if v < p.nVars {
				fmt.Printf(" %c", 'a'+v)
			} else {
				fmt.Printf(" %02d", v)
			}
		}
		fmt.Println(" )")
	}
}
func Exa_ManDumpBlif(p *Exa_Man_t, compl int32) {
	truth := append([]uint64(nil), p.pTruth...)
	if compl != 0 {
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	var sb strings.Builder
	util.Abc_TtPrintHexRev(&sb, truth, p.nVars)
	name := strings.TrimSpace(sb.String())
	filename := fmt.Sprintf("%s_2_%d.blif", name, p.nNodes)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "# Realization of the %d-input function %s using %d two-input gates:\n.model %s_2_%d\n.inputs", p.nVars, name, p.nNodes, name, p.nNodes)
	for i := int32(0); i < p.nVars; i++ {
		fmt.Fprintf(f, " %c", 'a'+i)
	}
	fmt.Fprint(f, "\n.outputs F\n")
	for i := p.nObjs - 1; i >= p.nVars; i-- {
		fmt.Fprint(f, ".names")
		for k := int32(1); k >= 0; k-- {
			v := Exa_ManFindFanin(p, i, k)
			if v < p.nVars {
				fmt.Fprintf(f, " %c", 'a'+v)
			} else {
				fmt.Fprintf(f, " %02d", v)
			}
		}
		if i == p.nObjs-1 {
			fmt.Fprintln(f, " F")
		} else {
			fmt.Fprintf(f, " %02d\n", i)
		}
		fun := exaFunction(p, i)
		if i == p.nObjs-1 && compl != 0 {
			for k := range fun {
				fun[k] = !fun[k]
			}
		}
		for k, on := range fun {
			if on {
				fmt.Fprintf(f, "%d%d 1\n", k>>1, k&1)
			}
		}
	}
	fmt.Fprintln(f, ".end")
}
func Exa_ManAddCnfStart(p *Exa_Man_t, onlyAnd int32) int32 {
	for i := p.nVars; i < p.nObjs; i++ {
		base := 1 + 3*(i-p.nVars)
		for k := int32(0); k < 2; k++ {
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
			if k == 1 {
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
		for k := int32(0); k < 3; k++ {
			a, b, c := bsat.Positive(base), bsat.Positive(base+1), bsat.Positive(base+2)
			if k == 1 {
				a ^= 1
			}
			if k == 2 {
				b ^= 1
			}
			if k != 0 {
				c ^= 1
			}
			if !p.pSat.AddClause(a, b, c) {
				return 0
			}
		}
		if onlyAnd != 0 && !p.pSat.AddClause(bsat.Negative(base), bsat.Negative(base+1), bsat.Positive(base+2)) {
			return 0
		}
	}
	for i := int32(0); i < p.nObjs-1; i++ {
		if !p.pSat.AddClause(p.vOutLits[i]...) {
			return 0
		}
	}
	return 1
}
func Exa_ManAddCnf(p *Exa_Man_t, mint int32) int32 {
	value := util.Abc_TtGetBit(p.pTruth, mint)
	for i := int32(0); i < p.nVars; i++ {
		p.VarVals[i] = (mint >> i) & 1
	}
	for p.pSat.NumVars() < p.iVar+3*p.nNodes {
		p.pSat.NewVariable()
	}
	for i := p.nVars; i < p.nObjs; i++ {
		base := p.iVar + 3*(i-p.nVars)
		function := 1 + 3*(i-p.nVars)
		for k := int32(0); k < 2; k++ {
			for j := int32(0); j < p.nObjs; j++ {
				mark := p.VarMarks[i][k][j]
				if mark == 0 {
					continue
				}
				other := p.iVar + 3*(j-p.nVars)
				for n := int32(0); n < 2; n++ {
					clause := []bsat.Literal{bsat.Negative(mark), bsat.Positive(base+k) ^ n}
					if j >= p.nVars {
						clause = append(clause, bsat.Positive(other+2)^(1-n))
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
			for k := int32(0); k < 4; k++ {
				if k == 0 && n == 1 {
					continue
				}
				clause := []bsat.Literal{bsat.Positive(base) ^ (k & 1), bsat.Positive(base+1) ^ (k >> 1)}
				if i != p.nObjs-1 {
					clause = append(clause, bsat.Positive(base+2)^(1-n))
				}
				if k > 0 {
					clause = append(clause, bsat.Positive(function+k-1)^n)
				}
				if !p.pSat.AddClause(clause...) {
					return 0
				}
			}
		}
	}
	p.iVar += 3 * p.nNodes
	return 1
}
func Exa_ManExactSynthesis(pars *Bmc_EsPar_t) { exaManExactSynthesis(pars, true, true) }

func exaManExactSynthesis(pars *Bmc_EsPar_t, dumpBlif, useTimeout bool) {
	if pars == nil || pars.nVars > 10 {
		return
	}
	truth := make([]uint64, util.Abc_TtWordNum(pars.nVars))
	util.Abc_TtReadHex(truth, pars.pTtStr)
	compl := int32(truth[0] & 1)
	if compl != 0 {
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	localPars := *pars
	if !useTimeout {
		localPars.RuntimeLim = 0
	}
	p := Exa_ManAlloc(&localPars, truth)
	defer Exa_ManFree(p)
	if Exa_ManAddCnfStart(p, pars.fOnlyAnd) == 0 {
		fmt.Println("The problem has no solution.")
		return
	}
	for mint := int32(1); ; {
		if Exa_ManAddCnf(p, mint) == 0 {
			fmt.Println("The problem has no solution.")
			return
		}
		status := p.pSat.Solve()
		if status != bsat.Satisfiable {
			fmt.Printf("Synthesis ended with status %v.\n", status)
			return
		}
		mint = Exa_ManEval(p)
		if mint < 0 {
			Exa_ManPrintSolution(p, compl)
			if dumpBlif {
				Exa_ManDumpBlif(p, compl)
			}
			return
		}
	}
}
