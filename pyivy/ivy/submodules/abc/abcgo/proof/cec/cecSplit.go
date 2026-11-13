package cec

// Cofactoring-based CEC, translated from cecSplit.c by Alan Mishchenko.
import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bmc"
	"github.com/glycerine/abcgo/sat/bsat"
)

type splitFix struct {
	variable int32
	value    bool
}
type splitPart struct {
	graph *gia.Gia_Man_t
	fixes []splitFix
}
type splitResult struct {
	status    int32
	model     []bool
	vars      int32
	conflicts int64
}

// splitSolve asserts the single output and extracts a primary-input model.
func splitSolve(part splitPart, timeout int32) splitResult {
	p := part.graph
	data := bmc.Cnf_DeriveGiaRemapped(p)
	if data == nil || gia.Gia_ManCoNum(p) != 1 {
		return splitResult{status: -1}
	}
	solver := bsat.NewSolver()
	for _, clause := range data.Clauses {
		lits := make([]bsat.Literal, len(clause))
		for i, lit := range clause {
			lits[i] = bsat.Literal(lit)
		}
		if !solver.AddClause(lits...) {
			return splitResult{status: 1}
		}
	}
	outputID := gia.Gia_ObjId(p, gia.Gia_ManCo(p, 0))
	if outputID < 0 || int(outputID) >= len(data.VarNums) {
		return splitResult{status: -1}
	}
	output := data.VarNums[outputID]
	if output < 0 {
		return splitResult{status: -1}
	}
	if !solver.AddClause(bsat.Positive(output)) {
		return splitResult{status: 1}
	}
	if timeout > 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	status := solver.Solve()
	result := splitResult{status: -1, vars: solver.NumVars(), conflicts: solver.ConflictCount()}
	if status == bsat.Unsatisfiable {
		result.status = 1
		return result
	}
	if status != bsat.Satisfiable {
		return result
	}
	result.status = 0
	result.model = make([]bool, gia.Gia_ManCiNum(p))
	for i := range result.model {
		id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, int32(i)))
		if id >= 0 && int(id) < len(data.VarNums) {
			result.model[i], _ = solver.Value(data.VarNums[id])
		}
	}
	for _, fix := range part.fixes {
		result.model[fix.variable] = fix.value
	}
	return result
}

// splitChoose excludes variables already fixed along this branch.
func splitChoose(part splitPart, lookAhead int32) (int32, int32, int32) {
	p := part.graph
	order := Gia_PermuteSpecialOrder(p)
	fixed := make(map[int32]bool, len(part.fixes))
	for _, f := range part.fixes {
		fixed[f.variable] = true
	}
	best, bestCost, bestRefs := int32(-1), int32(math.MaxInt32), int32(0)
	if lookAhead < 1 {
		lookAhead = 1
	}
	tried := int32(0)
	for _, variable := range order {
		if fixed[variable] {
			continue
		}
		refs := gia.Gia_ObjRefNum(p, gia.Gia_ManPi(p, variable))
		if refs == 0 {
			continue
		}
		if lookAhead == 1 {
			return variable, refs, -1
		}
		zero := gia.Gia_ManDupCofactorVar(p, variable, 0)
		one := gia.Gia_ManDupCofactorVar(p, variable, 1)
		cost := gia.Gia_ManAndNum(zero) + gia.Gia_ManAndNum(one)
		gia.Gia_ManStop(zero)
		gia.Gia_ManStop(one)
		if cost < bestCost {
			best, bestCost, bestRefs = variable, cost, refs
		}
		tried++
		if tried == lookAhead {
			break
		}
	}
	return best, bestRefs, bestCost
}

func splitRun(p *gia.Gia_Man_t, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32, solve func(splitPart, int32) splitResult) (int32, []bool) {
	start := time.Now()
	initial := solve(splitPart{graph: p}, timeout)
	if verbose != 0 {
		Cec_GiaSplitPrint(0, 0, initial.vars, initial.conflicts, initial.status, 0, time.Since(start))
	}
	if initial.status != -1 {
		return initial.status, initial.model
	}
	if workers < 1 {
		workers = 1
	}
	if workers > 100 {
		workers = 100
	}
	stack := []splitPart{{graph: gia.Gia_ManDup(p)}}
	splits := int32(0)
	progress := 0.0
	unknown := false
	first := true
	for len(stack) > 0 {
		count := int(workers)
		if count > len(stack) {
			count = len(stack)
		}
		batch := make([]splitPart, count)
		for i := range batch {
			batch[i] = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		}
		results := make([]splitResult, count)
		var wg sync.WaitGroup
		for i := range batch {
			if first {
				results[i] = initial
				continue
			}
			wg.Add(1)
			go func(i int) { defer wg.Done(); results[i] = solve(batch[i], timeout) }(i)
		}
		wg.Wait()
		first = false
		for i, part := range batch {
			result := results[i]
			if verbose != 0 {
				Cec_GiaSplitPrint(splits, int32(len(part.fixes)), result.vars, result.conflicts, result.status, progress, time.Since(start))
			}
			if result.status == 0 {
				for _, pending := range stack {
					gia.Gia_ManStop(pending.graph)
				}
				for _, item := range batch {
					gia.Gia_ManStop(item.graph)
				}
				return 0, result.model
			}
			if result.status == 1 {
				progress += math.Ldexp(1, -len(part.fixes))
				gia.Gia_ManStop(part.graph)
				continue
			}
			if maxIter > 0 && splits >= maxIter {
				unknown = true
				gia.Gia_ManStop(part.graph)
				continue
			}
			variable, fanouts, cost := splitChoose(part, lookAhead)
			if variable < 0 {
				unknown = true
				gia.Gia_ManStop(part.graph)
				continue
			}
			for value := int32(0); value < 2; value++ {
				child := gia.Gia_ManDupCofactorVar(part.graph, variable, value)
				fixes := append(append([]splitFix(nil), part.fixes...), splitFix{variable, value != 0})
				stack = append(stack, splitPart{child, fixes})
				if value == 0 && veryVerbose != 0 {
					fmt.Printf("Var = %5d. Fanouts = %5d. Cost = %8d. AndBefore = %6d. AndAfter = %6d.\n", variable, fanouts, cost, gia.Gia_ManAndNum(part.graph), gia.Gia_ManAndNum(child))
				}
			}
			splits++
			gia.Gia_ManStop(part.graph)
		}
	}
	status := int32(1)
	if unknown {
		status = -1
	}
	if silent == 0 {
		fmt.Printf("Problem status %d after %d case-splits. Time: %s\n", status, splits, time.Since(start))
	}
	return status, nil
}

func Cec_GiaSplitTestInt(p *gia.Gia_Man_t, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32) int32 {
	p.ClearCombinationalCounterexample()
	workers := procs
	if workers > 1 {
		workers--
	}
	status, model := splitRun(p, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, splitSolve)
	if status == 0 {
		if err := p.SetCombinationalCounterexample(model, 0); err != nil {
			return -1
		}
	}
	return status
}
func Cec_GiaSplitTest2(p *gia.Gia_Man_t, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32) int32 {
	return Cec_GiaSplitTestInt(p, 1, timeout, maxIter, lookAhead, verbose, veryVerbose, silent)
}

// splitOutput retains all CIs in their original order: the SAT assignment must
// remain meaningful in the original graph even when the output cone omits or
// visits some inputs in a different order. Register outputs are combinational
// inputs for this single-output SAT query.
func splitOutput(p *gia.Gia_Man_t, output int32) *gia.Gia_Man_t {
	q := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	if gia.Gia_ManHasMuxData(p) {
		q.EnableMuxes()
	}
	copies := make([]int32, gia.Gia_ManObjNum(p))
	for _, lit := range p.InputLiterals() {
		copies[lit>>1] = gia.Gia_ManAppendCi(q)
	}
	copyLit := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 1, 6:
		case 2:
			copies[id] = gia.Gia_ManAppendAnd(q, copyLit(a), copyLit(b))
		case 3:
			copies[id] = gia.Gia_ManAppendXorReal(q, copyLit(a), copyLit(b))
		case 4:
			copies[id] = gia.Gia_ManAppendMuxReal(q, copyLit(a), copyLit(b), copyLit(c))
		case 5:
			copies[id] = gia.Gia_ManAppendBuf(q, copyLit(a))
		default:
			panic("cec: invalid gate in split output")
		}
	}
	gia.Gia_ManAppendCo(q, copyLit(p.OutputLiterals()[output]))
	return q
}

func Cec_GiaSplitTest(p *gia.Gia_Man_t, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32) int32 {
	p.ClearCombinationalCounterexample()
	unknown := false
	for output := int32(0); output < gia.Gia_ManPoNum(p); output++ {
		one := splitOutput(p, output)
		workers := procs
		if workers > 1 {
			workers--
		}
		status, model := splitRun(one, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, splitSolve)
		gia.Gia_ManStop(one)
		if status == 0 {
			if p.SetCombinationalCounterexample(model, int(output)) != nil {
				return -1
			}
			return 0
		}
		if status < 0 {
			unknown = true
		}
	}
	if unknown {
		return -1
	}
	return 1
}
func Cec_GiaSplitPrint(iter, depth, vars int32, conflicts int64, status int32, progress float64, elapsed time.Duration) {
	label := "UNDECIDED"
	if status == 0 {
		label = "SAT"
	}
	if status == 1 {
		label = "UNSAT"
	}
	fmt.Printf("%4d : Depth =%3d SatVar =%7d SatConf =%7d %s Solved %8.4f %% Time: %s\n", iter, depth, vars, conflicts, label, 100*progress, elapsed)
}
func Cec_GiaSplitPrintRefs(p *gia.Gia_Man_t) {
	if p.Refs == nil {
		gia.Gia_ManCreateRefs(p)
	}
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		fmt.Printf("%d ", gia.Gia_ObjRefNum(p, gia.Gia_ManPi(p, i)))
	}
	fmt.Println()
}

// Cec_GiaSplitExplore prints reference counts around recognized XORs.
func Cec_GiaSplitExplore(p *gia.Gia_Man_t) {
	p.Refs = nil
	gia.Gia_ManCreateRefs(p)
	count := 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		obj := gia.Gia_ManObj(p, id)
		_, _, _, kind := p.Gate(id)
		// Structural XOR recognition assumes an AND-inverter cone. Native
		// MUX, XOR and buffer nodes are not such cones.
		if kind != 2 {
			continue
		}
		_, _, _, k0 := p.Gate(gia.Gia_ObjId(p, gia.Gia_ObjFanin0(obj)))
		_, _, _, k1 := p.Gate(gia.Gia_ObjId(p, gia.Gia_ObjFanin1(obj)))
		if k0 != 2 || k1 != 2 {
			continue
		}
		var a, b *gia.Gia_Obj_t
		if gia.Gia_ObjRecognizeExor(obj, &a, &b) == 0 {
			continue
		}
		if gia.Gia_ObjRefNum(p, gia.Gia_ObjFanin0(obj)) > 1 && gia.Gia_ObjRefNum(p, gia.Gia_ObjFanin1(obj)) > 1 {
			continue
		}
		fmt.Printf("%5d : %2d %2d    %2d %2d\n", count, gia.Gia_ObjRefNum(p, gia.Gia_Regular(a)), gia.Gia_ObjRefNum(p, gia.Gia_Regular(b)), gia.Gia_ObjRefNum(p, gia.Gia_ObjFanin0(obj)), gia.Gia_ObjRefNum(p, gia.Gia_ObjFanin1(obj)))
		count++
	}
}

func Cec_GiaPrintCofStats(p *gia.Gia_Man_t) {
	marks := make([]int32, gia.Gia_ManObjNum(p))
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(obj) == 0 || gia.Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var a, b *gia.Gia_Obj_t
		if gia.Gia_ObjRecognizeExor(obj, &a, &b) != 0 {
			continue
		}
		ctrl := gia.Gia_Regular(gia.Gia_ObjRecognizeMux(obj, &a, &b))
		marks[gia.Gia_ObjId(p, ctrl)]++
	}
	unique := 0
	for _, count := range marks {
		if count > 0 {
			unique++
		}
	}
	fmt.Printf("The AIG with %d candidate nodes (PI+AND) has %d unique MUX control drivers:\n", gia.Gia_ManCandNum(p), unique)
	gia.Gia_ManLevelNum(p)
	count := 0
	for id, refs := range marks {
		if refs == 0 {
			continue
		}
		zero := gia.Gia_ManDupCofactorObj(p, int32(id), 0)
		one := gia.Gia_ManDupCofactorObj(p, int32(id), 1)
		fmt.Printf("%6d : Obj = %6d MUX refs = %5d Level = %5d Cof0 = %7d Cof1 = %7d\n", count, id, refs, gia.Gia_ObjLevelId(p, int32(id)), gia.Gia_ManAndNum(zero), gia.Gia_ManAndNum(one))
		gia.Gia_ManStop(zero)
		gia.Gia_ManStop(one)
		count++
	}
}
func Cec_GiaPrintCofStats2(p *gia.Gia_Man_t) {
	gia.Gia_ManLevelNum(p)
	if p.Refs == nil {
		gia.Gia_ManCreateRefs(p)
	}
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		zero := gia.Gia_ManDupCofactorVar(p, i, 0)
		one := gia.Gia_ManDupCofactorVar(p, i, 1)
		fmt.Printf("PI %5d : Refs = %5d Cof0 = %7d Cof1 = %7d\n", i, gia.Gia_ObjRefNum(p, gia.Gia_ManPi(p, i)), gia.Gia_ManAndNum(zero), gia.Gia_ManAndNum(one))
		gia.Gia_ManStop(zero)
		gia.Gia_ManStop(one)
	}
}
