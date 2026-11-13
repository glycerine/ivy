package cec

// Cofactoring-based CEC, translated from cecSplit.c by Alan Mishchenko.
import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
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
	aigGraph := gia.Gia_ManToAigSimple(p)
	aig.Aig_ManSetRegNum(aigGraph, 0)
	data := cnf.Cnf_Derive(aigGraph, 0)
	if data == nil {
		return splitResult{status: -1}
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
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
	lits := make([]int32, len(part.fixes))
	for i, fix := range part.fixes {
		lits[i] = fix.variable << 1
		if !fix.value {
			lits[i] |= 1
		}
	}
	p.SetSplitCofactors(lits)
	model := Cec_SplitDeriveModel(p, data, solver)
	result.model = make([]bool, model.NPis)
	for i := range result.model {
		result.model[i] = model.PData[i>>5]&(1<<uint(i&31)) != 0
	}
	return result
}

// Cec_SplitDeriveModel follows cecSplit.c: SAT PI values are overwritten by
// the cofactor literals and stored in a zero-register combinational model.
func Cec_SplitDeriveModel(p *gia.Gia_Man_t, data *cnf.Cnf_Dat_t, solver *bsat.Solver) *util.Abc_Cex_t {
	model := util.Abc_CexAlloc(0, gia.Gia_ManPiNum(p), 1)
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManPi(p, i))
		v, _ := solver.Value(data.VarNums[id])
		if v {
			model.PData[i>>5] |= 1 << uint(i&31)
		}
	}
	for _, lit := range p.SplitCofactors() {
		i := lit >> 1
		mask := uint32(1) << uint(i&31)
		if lit&1 == 0 {
			model.PData[i>>5] |= mask
		} else {
			model.PData[i>>5] &^= mask
		}
	}
	return model
}

// splitChoose excludes variables already fixed along this branch.
func splitChoose(part splitPart, lookAhead int32) (int32, int32, int32) {
	var refs, cost int32
	variable := Gia_SplitCofVar(part.graph, lookAhead, &refs, &cost)
	return variable, refs, cost
}

func splitRun(p *gia.Gia_Man_t, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32, solve func(splitPart, int32) splitResult) (int32, []bool) {
	return splitRunMode(p, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, workers != 1, solve)
}
func splitRunMode(p *gia.Gia_Man_t, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32, parallel bool, solve func(splitPart, int32) splitResult) (int32, []bool) {
	started := time.Now()
	initial := solve(splitPart{graph: p}, timeout)
	if verbose != 0 && (!parallel || initial.status != -1) {
		Cec_GiaSplitPrint(0, 0, initial.vars, initial.conflicts, initial.status, 0, time.Since(started))
	}
	if initial.status != -1 {
		if silent == 0 {
			if initial.status == 0 {
				fmt.Println("The problem is SAT without cofactoring.")
			} else {
				fmt.Println("The problem is UNSAT without cofactoring.")
			}
		}
		return initial.status, initial.model
	}
	if !parallel {
		return splitSerial(p, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, started, solve)
	}
	return splitParallel(p, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, started, solve)
}
func splitFinish(ret int32, iter int32, started time.Time, silent int32) {
	if silent != 0 {
		return
	}
	name := "UNDECIDED"
	if ret == 0 {
		name = "SAT"
	} else if ret == 1 {
		name = "UNSAT"
	}
	fmt.Printf("Problem is %s after %d case-splits. Time: %.2f sec\n", name, iter, time.Since(started).Seconds())
}

// splitSerial follows source depth-first order, including its empty-stack bug.
func splitSerial(p *gia.Gia_Man_t, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32, started time.Time, solve func(splitPart, int32) splitResult) (int32, []bool) {
	stack := []splitPart{{graph: gia.Gia_ManDup(p)}}
	ret := int32(-1)
	var model []bool
	iter := int32(1)
	progress := 0.0
	for ; len(stack) > 0; iter++ {
		last := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		variable, fanouts, cost := splitChoose(last, lookAhead)
		if variable < 0 {
			gia.Gia_ManStop(last.graph)
			break
		}
		found := false
		for value := int32(0); value < 2; value++ {
			child := splitPart{graph: gia.Gia_ManDupCofactorVar(last.graph, variable, value), fixes: append(append([]splitFix(nil), last.fixes...), splitFix{variable, value != 0})}
			if veryVerbose != 0 && value == 0 {
				fmt.Printf("Var = %5d. Fanouts = %5d. Cost = %8d.  AndBefore = %6d.  AndAfter = %6d.\n", variable, fanouts, cost, gia.Gia_ManAndNum(last.graph), gia.Gia_ManAndNum(child.graph))
			}
			result := solve(child, timeout)
			if result.status == 1 {
				progress += math.Ldexp(1, -len(child.fixes))
			}
			if verbose != 0 {
				Cec_GiaSplitPrint(iter, int32(len(child.fixes)), result.vars, result.conflicts, result.status, progress, time.Since(started))
			}
			if result.status == 0 {
				ret = 0
				model = result.model
				gia.Gia_ManStop(child.graph)
				found = true
				break
			}
			if result.status == 1 {
				gia.Gia_ManStop(child.graph)
			} else {
				stack = append(stack, child)
			}
		}
		gia.Gia_ManStop(last.graph)
		if found || maxIter != 0 && iter >= maxIter {
			break
		}
	}
	if len(stack) == 0 {
		ret = 1
	}
	for _, part := range stack {
		gia.Gia_ManStop(part.graph)
	}
	splitFinish(ret, iter, started, silent)
	return ret, model
}

// The source checks completed workers in slot order and records the negative
// cofactor literal for both children, including the positive branch.
type splitSlot struct {
	part *splitPart
	done chan splitResult
}

func splitParallel(p *gia.Gia_Man_t, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32, started time.Time, solve func(splitPart, int32) splitResult) (int32, []bool) {
	return splitParallelWithScanHook(p, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, started, solve, nil)
}

// The hook permits tests to synchronize worker completion at a scan boundary.
func splitParallelWithScanHook(p *gia.Gia_Man_t, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32, started time.Time, solve func(splitPart, int32) splitResult, beforeScan func([]splitSlot)) (int32, []bool) {
	if workers < 1 || workers > 100 {
		panic("cec: invalid split worker count")
	}
	slots := make([]splitSlot, workers)
	stack := []splitPart{{graph: gia.Gia_ManDup(p)}}
	iter := int32(0)
	ret := int32(-1)
	var model []bool
	progress := 0.0
	finished := false
	for !finished {
		if beforeScan != nil {
			beforeScan(slots)
		}
		work := len(stack) > 0
		for i := range slots {
			slot := &slots[i]
			if slot.part != nil {
				select {
				case result := <-slot.done:
					part := *slot.part
					slot.part = nil
					slot.done = nil
					if verbose != 0 {
						Cec_GiaSplitPrint(int32(i+1), int32(len(part.fixes)), result.vars, result.conflicts, result.status, progress, time.Since(started))
					}
					if result.status == 0 {
						ret = 0
						model = result.model
						gia.Gia_ManStop(part.graph)
						finished = true
						break
					}
					if result.status == -1 {
						variable, fanouts, cost := splitChoose(part, lookAhead)
						if variable < 0 {
							gia.Gia_ManStop(part.graph)
							finished = true
							break
						}
						for value := int32(0); value < 2; value++ {
							child := splitPart{graph: gia.Gia_ManDupCofactorVar(part.graph, variable, value), fixes: append(append([]splitFix(nil), part.fixes...), splitFix{variable, false})}
							stack = append(stack, child)
							if value == 0 && veryVerbose != 0 {
								fmt.Printf("Var = %5d. Fanouts = %5d. Cost = %8d.  AndBefore = %6d.  AndAfter = %6d.\n", variable, fanouts, cost, gia.Gia_ManAndNum(part.graph), gia.Gia_ManAndNum(child.graph))
							}
						}
						iter++
						work = true
					} else {
						progress += math.Ldexp(1, -len(part.fixes))
					}
					gia.Gia_ManStop(part.graph)
				default:
					work = true
					continue
				}
			}
			if finished {
				break
			}
			if len(stack) > 0 {
				part := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				slot.part = &part
				slot.done = make(chan splitResult, 1)
				ch := slot.done
				go func() { ch <- solve(part, timeout) }()
			}
		}
		if finished {
			break
		}
		if maxIter != 0 && iter >= maxIter {
			break
		}
		if !work {
			ret = 1
			break
		}
		runtime.Gosched()
	}
	for i := range slots {
		if slots[i].part != nil {
			<-slots[i].done
			gia.Gia_ManStop(slots[i].part.graph)
		}
	}
	for _, part := range stack {
		gia.Gia_ManStop(part.graph)
	}
	splitFinish(ret, iter, started, silent)
	return ret, model
}

func splitSetModel(p *gia.Gia_Man_t, model []bool, output int32) error {
	c := util.Abc_CexAlloc(0, gia.Gia_ManPiNum(p), 1)
	c.IPo = output
	for i, v := range model {
		if v {
			c.PData[i>>5] |= 1 << uint(i&31)
		}
	}
	return p.SetSplitCounterexample(c)
}

func Cec_GiaSplitTestInt(p *gia.Gia_Man_t, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32) int32 {
	p.ClearCombinationalCounterexample()
	if err := cecRequireAIG(p); err != nil {
		fmt.Println(err)
		return -1
	}
	if procs == 1 {
		return Cec_GiaSplitTest2(p, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent)
	}
	workers := procs - 1
	if workers < 1 || workers > 100 {
		panic("cec: invalid split processor count")
	}
	status, model := splitRunMode(p, workers, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, true, splitSolve)
	if model != nil {
		if err := splitSetModel(p, model, 0); err != nil {
			return -1
		}
	}
	return status
}
func Cec_GiaSplitTest2(p *gia.Gia_Man_t, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32) int32 {
	if err := cecRequireAIG(p); err != nil {
		fmt.Println(err)
		return -1
	}
	status, model := splitRun(p, 1, timeout, maxIter, lookAhead, verbose, veryVerbose, silent, splitSolve)
	if model != nil {
		if err := splitSetModel(p, model, 0); err != nil {
			return -1
		}
	}
	return status
}
func splitOutput(p *gia.Gia_Man_t, output int32) *gia.Gia_Man_t {
	return gia.Gia_ManDupOutputGroup(p, output, output+1)
}
func Cec_GiaSplitTest(p *gia.Gia_Man_t, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent int32) int32 {
	p.ClearCombinationalCounterexample()
	if err := cecRequireAIG(p); err != nil {
		fmt.Println(err)
		return -1
	}
	result := int32(-1)
	unknown := false
	var witness *util.Abc_Cex_t
	for output := int32(0); output < gia.Gia_ManPoNum(p); output++ {
		one := splitOutput(p, output)
		if verbose != 0 {
			fmt.Printf("\nSolving output %d:\n", output)
		}
		status := Cec_GiaSplitTestInt(one, procs, timeout, maxIter, lookAhead, verbose, veryVerbose, silent)
		if status == 0 && result == -1 {
			witness = one.TakeCombinationalCounterexample()
			if witness != nil {
				witness.IPo = output
			}
			result = 0
		}
		gia.Gia_ManStop(one)
		if status == -1 {
			unknown = true
		}
	}
	if result == -1 {
		if unknown {
			return -1
		}
		return 1
	}
	if witness != nil {
		if err := p.SetSplitCounterexample(witness); err != nil {
			return -1
		}
	}
	return result
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
