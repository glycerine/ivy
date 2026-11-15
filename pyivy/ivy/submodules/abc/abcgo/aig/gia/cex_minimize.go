package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"slices"
	"time"
)

// Gia_ManFramesForCexMin unrolls a zero-initialized transition system.
func Gia_ManFramesForCexMin(p *Gia_Man_t, frames int32) *Gia_Man_t {
	if Gia_ManPoNum(p) != 1 || frames <= 0 {
		panic("gia: invalid counterexample unrolling")
	}
	out := Gia_ManStart(p.nObjs)
	out.pName, out.PSpec = p.pName, p.PSpec
	Gia_ManHashAlloc(out)
	mapped := make([]int32, p.nObjs)
	state := make([]int32, Gia_ManRegNum(p))
	for frame := int32(0); frame < frames; frame++ {
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			mapped[Gia_ObjId(p, Gia_ManCi(p, i))] = Gia_ManAppendCi(out)
		}
		for i := range state {
			mapped[Gia_ObjId(p, Gia_ManCi(p, Gia_ManPiNum(p)+int32(i)))] = state[i]
		}
		lit := func(v int32) int32 { return mapped[v>>1] ^ (v & 1) }
		for id := int32(1); id < p.nObjs; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 1:
			case 2:
				mapped[id] = Gia_ManHashAnd(out, lit(a), lit(b))
			case 3:
				mapped[id] = Gia_ManHashXor(out, lit(a), lit(b))
			case 4:
				mapped[id] = Gia_ManHashMux(out, lit(a), lit(b), lit(c))
			case 5, 6:
				mapped[id] = lit(a)
			default:
				panic("gia: invalid gate during counterexample unrolling")
			}
		}
		for i := range state {
			state[i] = mapped[Gia_ObjId(p, Gia_ManCo(p, Gia_ManPoNum(p)+int32(i)))]
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManAppendCo(out, mapped[Gia_ObjId(p, Gia_ManCo(p, i))])
	}
	clean := Gia_ManCleanup(out)
	Gia_ManStop(out)
	return clean
}

// Bmc_CexCareDeriveCex builds a mask from the selected input SAT literals.
func Bmc_CexCareDeriveCex(cex *util.Abc_Cex_t, first int32, lits []int32, count int32) *util.Abc_Cex_t {
	if count < 0 || int(count) > len(lits) {
		panic("gia: invalid care literals")
	}
	out := util.Abc_CexAlloc(cex.NRegs, cex.NPis, cex.IFrame+1)
	out.IPo, out.IFrame = cex.IPo, cex.IFrame
	for _, lit := range lits[:count] {
		index := (lit >> 1) - first
		if index < 0 || index >= cex.NBits-cex.NRegs {
			panic("gia: care literal out of range")
		}
		index += out.NRegs
		out.PData[index>>5] |= 1 << uint(index&31)
	}
	return out
}

// Bmc_CexCareSatBasedMinimizeAig follows the source priority prepass, MF
// CNF, final conflict analysis in both orders, and optional LEXUNSAT passes.
func Bmc_CexCareSatBasedMinimizeAig(p *Gia_Man_t, cex *util.Abc_Cex_t, highEffort, verbose int32) *util.Abc_Cex_t {
	if Gia_ManVerifyCex(p, cex, 0) == 0 || cex.IPo != 0 || Gia_ManPoNum(p) != 1 {
		panic("gia: expected a valid single-output counterexample")
	}
	for i := int32(0); i < cex.NRegs; i++ {
		if util.Abc_InfoHasBit(cex.PData, i) {
			panic("gia: expected zero initial state")
		}
	}
	started := time.Now()
	best := Gia_ManCexCareMinimize(p, Gia_ManPiNum(p), cex, 1, 1, verbose)
	countBest := util.Abc_CexCountOnes(best)
	if verbose != 0 {
		fmt.Printf("Care bits = %d. Non-SAT-based CEX minimization = %.2f sec\n", countBest, time.Since(started).Seconds())
	}
	started = time.Now()
	solver, first, assumptions, negated := giaCexCareSATInstance(p, cex)
	defer solver.Close()
	if !negated {
		// A constant-true output is a supported Go extension. The priority proof
		// already establishes that no input assignments need to be retained.
		if countBest == 0 {
			return best
		}
		panic("gia: counterexample output cannot be negated")
	}

	if verbose != 0 {
		fmt.Printf("Constructing SAT solver = %.2f sec\n", time.Since(started).Seconds())
	}
	update := func(lits []bsat.Literal) {
		if int32(len(lits)) >= countBest {
			return
		}
		plain := make([]int32, len(lits))
		for i, lit := range lits {
			plain[i] = int32(lit)
		}
		countBest = int32(len(lits))
		best = Bmc_CexCareDeriveCex(cex, first, plain, int32(len(plain)))
	}
	for pass := 0; pass < 2; pass++ {
		lits := append([]bsat.Literal(nil), assumptions...)
		if pass != 0 {
			slices.Reverse(lits)
		}
		started = time.Now()
		status := solver.Solve(lits...)
		core := solver.UnsatCore()
		if status != bsat.Unsatisfiable {
			panic("gia: counterexample does not force output")
		}
		if verbose != 0 {
			fmt.Printf("Status OK    Selected %5d assumptions out of %5d.  Analyze_final = %.2f sec\n", len(core), len(lits), time.Since(started).Seconds())
		}
		update(core)
		if highEffort == 0 {
			continue
		}
		started = time.Now()
		count := solver.MinimizeAssumptions(lits, 0)
		if verbose != 0 {
			fmt.Printf("Status OK    Selected %5d assumptions out of %5d.  LEXUNSAT = %.2f sec\n", count, len(lits), time.Since(started).Seconds())
		}
		update(lits[:count])
	}
	if verbose != 0 {
		fmt.Print("Final    :    ")
		util.Abc_CexPrintStatsInputs(best, best.NPis)
	}
	return best
}

// giaCexCareSATInstance uses source MF variable ordering: COs first, inputs last.
func giaCexCareSATInstance(p *Gia_Man_t, cex *util.Abc_Cex_t) (*bsat.Solver, int32, []bsat.Literal, bool) {
	graph := Gia_ManFramesForCexMin(p, cex.IFrame+1)
	defer Gia_ManStop(graph)
	cnf := Mf_ManGenerateCnf(graph, 8, 0, 0, 0, 0)
	if cnf == nil {
		panic("gia: counterexample CNF generation failed")
	}
	solver := bsat.NewSolver()
	solver.SetNumVars(int32(cnf.Variables))
	for _, row := range cnf.Clauses {
		clause := make([]bsat.Literal, len(row))
		for i, lit := range row {
			clause[i] = bsat.Literal(lit)
		}
		if !solver.AddClause(clause...) {
			panic("gia: inconsistent counterexample CNF")
		}
	}
	// Cnf_DataWriteIntoSolver finishes with level-zero simplification before
	// the caller adds its output assumption.
	if !solver.Simplify() {
		panic("gia: inconsistent counterexample CNF")
	}
	negated := solver.AddClause(bsat.Negative(1))
	first := int32(cnf.Variables) - (cex.IFrame+1)*cex.NPis
	assumptions := make([]bsat.Literal, cex.NBits-cex.NRegs)
	for i := range assumptions {
		assumptions[i] = bsat.Positive(first + int32(i))
		if !util.Abc_InfoHasBit(cex.PData, cex.NRegs+int32(i)) {
			assumptions[i] ^= 1
		}
	}
	return solver, first, assumptions, negated
}

func Gia_ManMinCex(p *Gia_Man_t, cex *util.Abc_Cex_t) {
	if Gia_ManVerifyCex(p, cex, 0) == 0 || cex.IPo != 0 || Gia_ManPoNum(p) != 1 {
		panic("gia: expected a valid single-output counterexample")
	}
	for i := int32(0); i < cex.NRegs; i++ {
		if util.Abc_InfoHasBit(cex.PData, i) {
			panic("gia: expected zero initial state")
		}
	}
	started := time.Now()
	care := Gia_ManCexCareMinimize(p, Gia_ManPiNum(p), cex, 1, 1, 1)
	fmt.Printf("Care bits = %d. CEX minimization = %.2f sec\n", util.Abc_CexCountOnes(care), time.Since(started).Seconds())
	started = time.Now()
	solver, _, lits, _ := giaCexCareSATInstance(p, cex)
	defer solver.Close()
	fmt.Printf("SAT solver = %.2f sec\n", time.Since(started).Seconds())
	for pass := 0; pass < 2; pass++ {
		if pass != 0 {
			slices.Reverse(lits)
		}
		started = time.Now()
		status := solver.Solve(lits...)
		core := solver.UnsatCore()
		if status != bsat.Unsatisfiable {
			panic("gia: counterexample does not force output")
		}
		fmt.Printf("Status %d.  Selected %d assumptions out of %d.  Analyze_final = %.2f sec\n", status, len(core), len(lits), time.Since(started).Seconds())
		started = time.Now()
		count := solver.MinimizeAssumptions(lits, 0)
		fmt.Printf("Status %d.  Selected %d assumptions out of %d.  LEXUNSAT = %.2f sec\n", status, count, len(lits), time.Since(started).Seconds())
	}
}
