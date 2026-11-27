package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/bits"
)

const MAJ3_OBJS = 32

type Maj3_Man_t struct {
	nVars, nNodes, nObjs, nWords, iVar int32
	vInfo                              *vec.Vec_Wrd_t
	vLevels                            *vec.Vec_Int_t
	VarMarks                           [MAJ3_OBJS][MAJ3_OBJS]int32
	ObjVals                            [MAJ3_OBJS]int32
	pLits                              [2][MAJ3_OBJS]int32
	nLits                              [3]int32
	pSat                               *bsat.Solver
	model                              []bool
}

func Maj3_ManTruth(p *Maj3_Man_t, v int32) []uint64 {
	data := vec.Vec_WrdArray(p.vInfo)
	start := int(p.nWords * v)
	return data[start : start+int(p.nWords)]
}
func Maj3_ManTruthTables(p *Maj3_Man_t) *vec.Vec_Wrd_t {
	p.vInfo = vec.Vec_WrdStart(int(p.nWords * (p.nObjs + 1)))
	for i := int32(0); i < p.nVars; i++ {
		util.Abc_TtIthVar(Maj3_ManTruth(p, i), i, p.nVars)
	}
	count := int32(1) << p.nVars
	if count < 64 {
		count = 64
	}
	for i := int32(0); i < count; i++ {
		if Maj3_ManValue(i, p.nVars) != 0 {
			util.Abc_TtSetBit(Maj3_ManTruth(p, p.nObjs), i)
		}
	}
	return p.vInfo
}
func Maj3_ManMarkup(p *Maj3_Man_t) int32 {
	counts := vec.Vec_IntArray(p.vLevels)
	if len(counts) < 2 || counts[0] != 1 || p.nObjs > MAJ3_OBJS || p.nVars < 3 {
		panic("bmc: invalid majority synthesis structure")
	}
	var first, levels [MAJ3_OBJS]int32
	Maj3_ManFirstAndLevel(p.vLevels, first[:], levels[:], p.nVars, p.nObjs)
	for i := int32(0); i < p.nObjs; i++ {
		for k := int32(0); k < p.nObjs; k++ {
			p.VarMarks[i][k] = -1
		}
	}
	for k := int32(0); k < 3; k++ {
		p.VarMarks[p.nVars][k] = 1
	}
	second := counts[1]
	for k := int32(0); k < second; k++ {
		p.VarMarks[p.nObjs-1][p.nObjs-2-k] = 1
	}
	for k := 2; k < len(counts); k++ {
		p.VarMarks[first[k]][first[k-1]] = 1
	}
	end := p.nObjs
	if second == 3 {
		end--
	}
	next := int32(2)
	for i := p.nVars + 1; i < end; i++ {
		for k := int32(0); k < first[levels[i]]; k++ {
			if p.VarMarks[i][k] == -1 {
				p.VarMarks[i][k] = next
				next++
			}
		}
	}
	return next
}
func Maj3_ManVarMapPrint(p *Maj3_Man_t) {
	fmt.Printf("Variable map for %d inputs and %d nodes:\n", p.nVars, p.nNodes)
	for i := p.nObjs - 1; i >= p.nVars; i-- {
		fmt.Printf("%2d ", i)
		for k := int32(0); k < p.nObjs; k++ {
			mark := p.VarMarks[i][k]
			if mark < 0 {
				fmt.Print("  .  ")
			} else if mark == 1 {
				fmt.Print("  +  ")
			} else {
				val, _ := p.pSat.Value(mark)
				suffix := ' '
				if val {
					suffix = '+'
				}
				fmt.Printf("%3d%c ", mark, suffix)
			}
		}
		fmt.Println()
	}
}
func Maj3_ManFindFanin(p *Maj3_Man_t, i int32, fanins []int32) int32 {
	p.nLits = [3]int32{}
	n := int32(0)
	for f := int32(0); f < i; f++ {
		mark := p.VarMarks[i][f]
		if mark < 0 {
			continue
		}
		if mark == 1 {
			p.nLits[2]++
			fanins[n] = f
			n++
			continue
		}
		if int(mark) >= len(p.model) {
			panic("bmc: missing majority SAT model")
		}
		value := p.model[mark]
		if value {
			p.pLits[1][p.nLits[1]] = bsat.Negative(mark)
			p.nLits[1]++
			fanins[n] = f
			n++
		} else {
			p.pLits[0][p.nLits[0]] = bsat.Positive(mark)
			p.nLits[0]++
		}
	}
	return n
}
func Maj3_ManPrintSolution(p *Maj3_Man_t) {
	fmt.Printf("Realization of %d-input majority using %d MAJ3 gates:\n", p.nVars, p.nNodes)
	var fanins [MAJ3_OBJS]int32
	for i := p.nObjs - 1; i >= p.nVars; i-- {
		n := Maj3_ManFindFanin(p, i, fanins[:])
		fmt.Printf("%02d = MAJ(", i)
		for _, f := range fanins[:n] {
			if f < p.nVars {
				fmt.Printf(" %c", 'a'+f)
			} else {
				fmt.Printf(" %02d", f)
			}
		}
		fmt.Println(" )")
	}
}
func Maj3_ManEval(p *Maj3_Man_t) int32 {
	var fs [MAJ3_OBJS]int32
	for i := p.nVars; i < p.nObjs; i++ {
		if Maj3_ManFindFanin(p, i, fs[:]) != 3 {
			panic("bmc: majority gate has wrong fanin count")
		}
		a, b, c := Maj3_ManTruth(p, fs[0]), Maj3_ManTruth(p, fs[1]), Maj3_ManTruth(p, fs[2])
		dst := Maj3_ManTruth(p, i)
		for w := range dst {
			dst[w] = (a[w] & b[w]) | (a[w] & c[w]) | (b[w] & c[w])
		}
	}
	spec, actual := Maj3_ManTruth(p, p.nObjs), Maj3_ManTruth(p, p.nObjs-1)
	for i := int32(0); i < 1<<p.nVars; i++ {
		n := int32(bits.OnesCount32(uint32(i)))
		if n >= p.nVars/2 && n <= p.nVars/2+1 && util.Abc_TtGetBit(spec, i) != util.Abc_TtGetBit(actual, i) {
			return i
		}
	}
	return -1
}
func Maj3_PrintClause(lits []int32, n int32) {
	for _, lit := range lits[:n] {
		c := '+'
		if lit&1 != 0 {
			c = '-'
		}
		fmt.Printf("%c%d ", c, lit>>1)
	}
	fmt.Println()
}
func Maj3_ManAddCnfStart(p *Maj3_Man_t) int32 {
	for i := p.nVars; i < p.nObjs; i++ {
		fixed := 0
		var clause []bsat.Literal
		for k := int32(0); k < p.nObjs; k++ {
			m := p.VarMarks[i][k]
			if m == 1 {
				fixed++
			} else if m > 1 {
				clause = append(clause, bsat.Positive(m))
			}
		}
		if fixed != 3 && len(clause) > 0 && !p.pSat.AddClause(clause...) {
			return 0
		}
	}
	for k := int32(0); k < p.nObjs-1; k++ {
		fixed := 0
		var clause []bsat.Literal
		for i := int32(0); i < p.nObjs; i++ {
			m := p.VarMarks[i][k]
			if m == 1 {
				fixed++
			} else if m > 1 {
				clause = append(clause, bsat.Positive(m))
			}
		}
		if fixed == 0 && len(clause) > 0 && !p.pSat.AddClause(clause...) {
			return 0
		}
	}
	return 1
}
func Maj3_ManAddCnf(p *Maj3_Man_t, mint int32) int32 {
	vals := &p.ObjVals
	for i := int32(0); i < p.nVars; i++ {
		vals[i] = (mint >> i) & 1
	}
	vals[p.nVars] = (vals[0] & vals[1]) | (vals[0] & vals[2]) | (vals[1] & vals[2])
	vals[p.nObjs-1] = Maj3_ManValue(mint, p.nVars)
	for i := p.nVars + 1; i < p.nObjs-1; i++ {
		vals[i] = p.iVar
		p.iVar++
	}
	for p.pSat.NumVars() < p.iVar {
		p.pSat.NewVariable()
	}
	for phase := int32(0); phase < 2; phase++ {
		for i := p.nVars + 1; i < p.nObjs; i++ {
			for k := int32(0); k < p.nObjs; k++ {
				if p.VarMarks[i][k] < 1 || vals[k] == 1-phase {
					continue
				}
				var first []bsat.Literal
				if vals[k] > 1 {
					first = append(first, bsat.Positive(vals[k])^phase)
				}
				if p.VarMarks[i][k] > 1 {
					first = append(first, bsat.Negative(p.VarMarks[i][k]))
				}
				for j := k + 1; j < p.nObjs; j++ {
					if p.VarMarks[i][j] < 1 || vals[j] == 1-phase || vals[i] == phase {
						continue
					}
					clause := append([]bsat.Literal(nil), first...)
					if vals[j] > 1 {
						clause = append(clause, bsat.Positive(vals[j])^phase)
					}
					if p.VarMarks[i][j] > 1 {
						clause = append(clause, bsat.Negative(p.VarMarks[i][j]))
					}
					if vals[i] > 1 {
						clause = append(clause, bsat.Positive(vals[i])^(1-phase))
					}
					if !p.pSat.AddClause(clause...) {
						return 0
					}
				}
			}
		}
	}
	return 1
}
func Maj3_ManAddConstraintsLazy(p *Maj3_Man_t) int32 {
	var fanins [MAJ3_OBJS]int32
	count := int32(0)
	for i := p.nVars + 1; i < p.nObjs; i++ {
		n := Maj3_ManFindFanin(p, i, fanins[:])
		if n == 3 {
			continue
		}
		count++
		var clause []bsat.Literal
		if n < 3 {
			clause = p.pLits[0][:p.nLits[0]]
		} else {
			size := 4 - p.nLits[2]
			if size > p.nLits[1] {
				size = p.nLits[1]
			}
			clause = p.pLits[1][:size]
		}
		if !p.pSat.AddClause(clause...) {
			return -1
		}
	}
	return count
}
func Maj3_ManAlloc(nVars, nNodes int32, levels *vec.Vec_Int_t) *Maj3_Man_t {
	p := &Maj3_Man_t{nVars: nVars, nNodes: nNodes, nObjs: nVars + nNodes, nWords: util.Abc_TtWordNum(nVars), vLevels: levels, pSat: bsat.NewSolver()}
	p.iVar = Maj3_ManMarkup(p)
	Maj3_ManTruthTables(p)
	for p.pSat.NumVars() < p.iVar {
		p.pSat.NewVariable()
	}
	if Maj3_ManAddCnfStart(p) == 0 {
		return nil
	}
	return p
}
func Maj3_ManFree(p *Maj3_Man_t) {}
func Maj3_ManExactSynthesis(nVars, nNodes, verbose int32, levels *vec.Vec_Int_t) int32 {
	clkTotal := util.Abc_Clock()
	fmt.Printf("Running exact synthesis for %d-input majority with %d MAJ3 gates...\n", nVars, nNodes)
	p := Maj3_ManAlloc(nVars, nNodes, levels)
	if p == nil {
		return 0
	}
	mint, lazyAll := int32(0), int32(0)
	for i := int32(0); mint != -1; i++ {
		clk := util.Abc_Clock()
		if Maj3_ManAddCnf(p, mint) == 0 {
			break
		}
		var status bsat.Status
		for {
			status = p.pSat.Solve()
			if status != bsat.Satisfiable {
				break
			}
			p.model = p.pSat.Model()
			lazy := Maj3_ManAddConstraintsLazy(p)
			if lazy == -1 {
				fmt.Printf("Became UNSAT after adding lazy constraints.\n")
				status = bsat.Unsatisfiable
				break
			}
			if lazy == 0 {
				break
			}
			lazyAll += lazy
		}
		if verbose != 0 {
			stats := p.pSat.Statistics()
			fmt.Printf("Iter %3d : ", i)
			for k := p.nVars - 1; k >= 0; k-- {
				fmt.Printf("%d", mint>>uint(k)&1)
			}
			fmt.Printf("  Var =%5d  ", p.iVar)
			fmt.Printf("Cla =%6d  ", stats.Clauses)
			fmt.Printf("Conf =%9d  ", stats.Conflicts)
			fmt.Printf("Lazy =%9d  ", lazyAll)
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
		}
		if status != bsat.Satisfiable {
			fmt.Printf("The problem has no solution.\n")
			break
		}
		mint = Maj3_ManEval(p)
	}
	if mint == -1 {
		Maj3_ManPrintSolution(p)
	}
	util.Abc_PrintTime(1, "Total runtime", util.Abc_Clock()-clkTotal)
	if mint == -1 {
		return 1
	}
	return 0
}
func Maj3_ManTest() int32 {
	levels := vec.Vec_IntAllocArrayCopy([]int32{1, 2, 2, 2}, 4)
	Maj3_ManExactSynthesis(7, 7, 1, levels)
	return 0
}
