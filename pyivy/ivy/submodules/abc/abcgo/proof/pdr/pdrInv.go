package pdr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [pdrInv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Invariant computation, printing, verification.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrInv.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"
//#include "base/abc/abc.h"      // for Abc_NtkCollectCioNames()
//#include "base/main/main.h"    // for Abc_FrameReadGlobalFrame()
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManPrintProgress(p *Pdr_Man_t, closeLine int32, elapsed abctime) {
	if len(p.vSolvers) < 2 {
		fmt.Println("Frame Clauses                                                     Max Queue Flops Cex      Time")
		return
	}
	if abcmain.Abc_FrameIsBatchMode() != 0 && closeLine == 0 {
		return
	}
	sizes := make([]string, vec.Vec_VecSize(p.vClauses))
	length := 0
	for i := range sizes {
		sizes[i] = fmt.Sprintf(" %d", vec.Vec_PtrSize(vec.Vec_VecEntry(p.vClauses, i)))
		length += len(sizes[i])
	}
	skip := max(0, length-60)
	fmt.Printf("%3d :", len(p.vSolvers)-1)
	width, seen := 5, 0
	if skip > 0 {
		fmt.Print(" ...")
		width += 4
	}
	for _, size := range sizes {
		if seen < skip {
			seen += len(size)
			continue
		}
		fmt.Print(size)
		width += len(size)
	}
	for ; width < 70; width++ {
		fmt.Print(" ")
	}
	flops := p.nAbsFlops
	if p.vAbsFlops != nil {
		flops = int32(vec.Vec_IntCountPositive(p.vAbsFlops))
	}
	fmt.Printf("%5d%6d", p.nQueMax, flops)
	if p.pPars.fUseAbs != 0 {
		fmt.Printf("%4d", p.nCexes)
	}
	fmt.Printf("%10.2f sec", time.Duration(elapsed).Seconds())
	if p.pPars.fSolveAll != 0 {
		fmt.Printf("  CEX =%4d", p.pPars.nFailOuts)
	}
	if p.pPars.nTimeOutOne != 0 {
		fmt.Printf("  T/O =%3d", p.pPars.nDropOuts)
	}
	if closeLine != 0 {
		fmt.Println()
		p.nQueMax, p.nCexes = 0, 0
	} else {
		fmt.Print("\r")
	}
}

// Pdr_ManCountFlops counts register literals in live cubes.
func Pdr_ManCountFlops(p *Pdr_Man_t, cubes *Vec_Ptr_t) *Vec_Int_t {
	nRegs := int(aig.Aig_ManRegNum(p.pAig))
	counts := vec.Vec_IntStart(nRegs)
	for _, item := range vec.Vec_PtrArray(cubes) {
		cube := item.(*Pdr_Set_t)
		if cube.nRefs == -1 {
			continue
		}
		for _, lit := range cube.Lits[:cube.nLits] {
			if lit < 0 || int(lit>>1) >= nRegs {
				panic("pdr: cube literal out of register range")
			}
			vec.Vec_IntAddToEntry(counts, int(lit>>1), 1)
		}
	}
	return counts
}
func Pdr_ManFindInvariantStart(p *Pdr_Man_t) int32 {
	last := len(p.vSolvers) - 1
	for k := 1; k <= last && k < vec.Vec_VecSize(p.vClauses); k++ {
		if vec.Vec_PtrSize(vec.Vec_VecEntry(p.vClauses, k)) == 0 {
			return int32(k)
		}
	}
	return int32(last)
}
func Pdr_ManCollectCubes(p *Pdr_Man_t, start int32) *Vec_Ptr_t {
	if start < 0 {
		panic("pdr: invalid invariant start")
	}
	result := vec.Vec_PtrAlloc(100)
	for i := int(start); i < vec.Vec_VecSize(p.vClauses); i++ {
		for _, cube := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, i)) {
			vec.Vec_PtrPush(result, cube)
		}
	}
	return result
}
func Pdr_ManCountFlopsInv(p *Pdr_Man_t) *Vec_Int_t {
	return Pdr_ManCountFlops(p, Pdr_ManCollectCubes(p, Pdr_ManFindInvariantStart(p)))
}
func Pdr_ManCountVariables(p *Pdr_Man_t, start int32) int32 {
	cubes := p.vInfCubes
	if cubes == nil {
		cubes = Pdr_ManCollectCubes(p, start)
	}
	var count int32
	for _, n := range vec.Vec_IntArray(Pdr_ManCountFlops(p, cubes)) {
		if n > 0 {
			count++
		}
	}
	return count
}
func Pdr_ManPrintClauses(p *Pdr_Man_t, start int32) {
	if start < 0 {
		panic("pdr: invalid frame")
	}
	var count int
	for k := int(start); k < vec.Vec_VecSize(p.vClauses); k++ {
		level := vec.Vec_VecEntry(p.vClauses, k)
		vec.Vec_PtrSort(level, func(a, b *any) int {
			x, y := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
			return int(Pdr_SetCompare(&x, &y))
		})
		for _, item := range vec.Vec_PtrArray(level) {
			fmt.Printf("C=%4d. F=%4d ", count, k)
			Pdr_SetPrint(os.Stdout, item.(*Pdr_Set_t), aig.Aig_ManRegNum(p.pAig), nil)
			fmt.Println()
			count++
		}
	}
}
func Pdr_SetPrintOne(p *Pdr_Set_t) {
	fmt.Print("Clause: {")
	for _, lit := range p.Lits[:p.nLits] {
		if lit&1 != 0 {
			fmt.Printf(" !%d", lit>>1)
		} else {
			fmt.Printf(" %d", lit>>1)
		}
	}
	fmt.Println(" }")
}

// Pdr_ManDupAigWithClauses creates cube outputs followed by the original
// next-state outputs, retaining the original input and register ordering.
func Pdr_ManDupAigWithClauses(src *aig.Graph, cubes *Vec_Ptr_t) *aig.Graph {
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	inputs, outputs := src.Inputs(), src.Outputs()
	nRegs := int(aig.Aig_ManRegNum(src))
	copies := map[int]aig.Edge{0: dst.Const1()}
	for _, in := range inputs {
		copies[in.NodeID()] = dst.NewInput()
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		result, ok := copies[edge.NodeID()]
		if !ok {
			node := src.Node(edge)
			switch node.Kind {
			case aig.And:
				result = dst.AND(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.ExclusiveOr:
				result = dst.NewXor(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.Buffer, aig.Output:
				result = copyEdge(node.Fanins[0])
			default:
				panic("pdr: unexpected AIG node in invariant duplication")
			}
			copies[edge.NodeID()] = result
		}
		if edge.Complemented() {
			return result.Not()
		}
		return result
	}
	for i := 0; i < vec.Vec_PtrSize(cubes); i++ {
		cube := vec.Vec_PtrEntry(cubes, i).(*Pdr_Set_t)
		term := dst.Const1()
		for _, lit := range cube.Lits[:cube.nLits] {
			reg := int(lit >> 1)
			if lit < 0 || reg >= nRegs {
				panic("pdr: cube literal out of range")
			}
			input := copyEdge(inputs[len(inputs)-nRegs+reg])
			if lit&1 != 0 {
				input = input.Not()
			}
			term = dst.AND(term, input)
		}
		dst.AddOutput(term)
	}
	for _, out := range outputs[len(outputs)-nRegs:] {
		dst.AddOutput(copyEdge(out))
	}
	aig.Aig_ManSetRegNum(dst, int32(nRegs))
	return dst
}

func Pdr_ManDumpAig(p *aig.Graph, cubes *Vec_Ptr_t) {
	if err := ioa.Ioa_WriteAiger(Pdr_ManDupAigWithClauses(p, cubes), "aig_with_clauses.aig", 0, 0); err != nil {
		fmt.Printf("Could not dump modified AIG: %v\n", err)
		return
	}
	fmt.Println("Dumped modified AIG into file \"aig_with_clauses.aig\".")
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManDumpClauses(p *Pdr_Man_t, fileName string, proved int32) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Cannot open file %q for writing invariant: %v\n", fileName, err)
		return
	}
	defer file.Close()
	start := Pdr_ManFindInvariantStart(p)
	var cubes *Vec_Ptr_t
	if proved != 0 {
		cubes = Pdr_ManCollectCubes(p, start)
	} else {
		cubes = vec.Vec_PtrAlloc(vec.Vec_PtrSize(p.vInfCubes))
		for _, item := range vec.Vec_PtrArray(p.vInfCubes) {
			vec.Vec_PtrPush(cubes, item)
		}
	}
	vec.Vec_PtrSort(cubes, func(a, b *any) int {
		x, y := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
		return int(Pdr_SetCompare(&x, &y))
	})
	count := 0
	for _, item := range vec.Vec_PtrArray(cubes) {
		if item.(*Pdr_Set_t).nRefs != -1 {
			count++
		}
	}
	var counts *Vec_Int_t
	nInputs := aig.Aig_ManRegNum(p.pAig)
	if p.pPars.fUseSupp != 0 {
		counts = Pdr_ManCountFlops(p, cubes)
		nInputs = Pdr_ManCountVariables(p, start)
	}
	if proved != 0 {
		fmt.Fprintf(file, "# Inductive invariant for %q\n", p.pAig.Name)
	} else {
		fmt.Fprintf(file, "# Clauses of the last timeframe for %q\n", p.pAig.Name)
	}
	fmt.Fprintf(file, "# generated by PDR in ABC on %s\n.i %d\n.o 1\n.p %d\n", time.Now().Format(time.ANSIC), nInputs, count)
	var network *abc.Network
	if frame := abcmain.Abc_FrameReadGlobalFrame(); frame != nil {
		network, _ = abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	}
	if network != nil {
		names := abc.Abc_NtkCollectCioNames(network, 0)
		offset := len(p.pAig.Inputs()) - int(aig.Aig_ManRegNum(p.pAig))
		if len(names) >= offset+int(aig.Aig_ManRegNum(p.pAig)) {
			fmt.Fprint(file, ".ilb")
			for i := 0; i < int(aig.Aig_ManRegNum(p.pAig)); i++ {
				if counts == nil || vec.Vec_IntEntry(counts, i) != 0 {
					fmt.Fprintf(file, " %s", names[offset+i])
				}
			}
			fmt.Fprint(file, "\n.ob inv\n")
		}
	}
	for _, item := range vec.Vec_PtrArray(cubes) {
		cube := item.(*Pdr_Set_t)
		if cube.nRefs == -1 {
			continue
		}
		Pdr_SetPrint(file, cube, aig.Aig_ManRegNum(p.pAig), counts)
		fmt.Fprint(file, " 1\n")
	}
	fmt.Fprint(file, ".e\n\n")
	if proved != 0 {
		fmt.Printf("Inductive invariant was written into file %q.\n", fileName)
	} else {
		fmt.Printf("Clauses of the last timeframe were written into file %q.\n", fileName)
	}
}

func Pdr_ManDumpString(p *Pdr_Man_t) *Vec_Str_t {
	cubes := p.vInfCubes
	if cubes == nil {
		cubes = Pdr_ManCollectCubes(p, Pdr_ManFindInvariantStart(p))
	} else {
		copyCubes := vec.Vec_PtrAlloc(vec.Vec_PtrSize(cubes))
		for _, item := range vec.Vec_PtrArray(cubes) {
			vec.Vec_PtrPush(copyCubes, item)
		}
		cubes = copyCubes
	}
	vec.Vec_PtrSort(cubes, func(a, b *any) int {
		x, y := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
		return int(Pdr_SetCompare(&x, &y))
	})
	var counts *Vec_Int_t
	if p.pPars.fUseSupp != 0 {
		counts = Pdr_ManCountFlops(p, cubes)
	}
	result := vec.Vec_StrAlloc(1000)
	for _, item := range vec.Vec_PtrArray(cubes) {
		cube := item.(*Pdr_Set_t)
		if cube.nRefs != -1 {
			Pdr_SetPrintStr(result, cube, aig.Aig_ManRegNum(p.pAig), counts)
		}
	}
	vec.Vec_StrPush(result, 0)
	return result
}

func Pdr_ManReportInvariant(p *Pdr_Man_t) {
	start := Pdr_ManFindInvariantStart(p)
	cubes := Pdr_ManCollectCubes(p, start)
	average := float64(0)
	if p.nXsimRuns != 0 {
		average = float64(p.nXsimLits) / float64(p.nXsimRuns)
	}
	fmt.Printf("Invariant F[%d] : %d clauses with %d flops (out of %d) (cex = %d, ave = %.2f)\n", start, vec.Vec_PtrSize(cubes), Pdr_ManCountVariables(p, start), aig.Aig_ManRegNum(p.pAig), p.nCexesTotal, average)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManVerifyInvariant(p *Pdr_Man_t) {
	started := time.Now()
	cubes := Pdr_ManCollectCubes(p, Pdr_ManFindInvariantStart(p))
	frame := int32(len(p.vSolvers))
	solver := Pdr_ManCreateSolver(p, frame)
	if solver == nil {
		fmt.Println("Could not create invariant verification solver.")
		return
	}
	for _, item := range vec.Vec_PtrArray(cubes) {
		cube := item.(*Pdr_Set_t)
		if !solver.AddClause(pdrLiterals(Pdr_ManCubeToLits(p, frame, cube, 1, 0))...) || !solver.Compress() {
			panic("pdr: inconsistent invariant clauses")
		}
	}
	failures := 0
	for i, item := range vec.Vec_PtrArray(cubes) {
		cube := item.(*Pdr_Set_t)
		if solver.Solve(pdrLiterals(Pdr_ManCubeToLits(p, frame, cube, 0, 1))...) != bsat.Unsatisfiable {
			fmt.Printf("Verification of clause %d failed.\n", i)
			failures++
		}
	}
	if failures != 0 {
		fmt.Printf("Verification of %d clauses has failed.\n", failures)
	} else {
		fmt.Printf("Verification of invariant with %d clauses was successful. Time = %.2f sec\n", vec.Vec_PtrSize(cubes), time.Since(started).Seconds())
	}
}

// Pdr_ManDeriveMarkNonInductive marks clauses not preserved by one transition.
func Pdr_ManDeriveMarkNonInductive(p *Pdr_Man_t, cubes *Vec_Ptr_t) int32 {
	frame := int32(len(p.vSolvers))
	solver := Pdr_ManCreateSolver(p, frame)
	if solver == nil {
		panic("pdr: could not create invariant verification solver")
	}
	for _, item := range vec.Vec_PtrArray(cubes) {
		cube := item.(*Pdr_Set_t)
		if cube.nRefs == -1 {
			continue
		}
		if !solver.AddClause(pdrLiterals(Pdr_ManCubeToLits(p, frame, cube, 1, 0))...) || !solver.Compress() {
			panic("pdr: inconsistent invariant clauses")
		}
	}
	var changed int32
	for _, item := range vec.Vec_PtrArray(cubes) {
		cube := item.(*Pdr_Set_t)
		if cube.nRefs == -1 {
			continue
		}
		if solver.Solve(pdrLiterals(Pdr_ManCubeToLits(p, frame, cube, 0, 1))...) != bsat.Unsatisfiable {
			cube.nRefs = -1
			changed = 1
		}
	}
	return changed
}
func Pdr_ManDeriveInfinityClauses(p *Pdr_Man_t, reduce int32) *Vec_Int_t {
	cubes := Pdr_ManCollectCubes(p, Pdr_ManFindInvariantStart(p))
	if reduce != 0 {
		for Pdr_ManDeriveMarkNonInductive(p, cubes) != 0 {
		}
	}
	result := vec.Vec_IntAlloc(1000)
	vec.Vec_IntPush(result, 0)
	for i := 0; i < vec.Vec_PtrSize(cubes); i++ {
		cube := vec.Vec_PtrEntry(cubes, i).(*Pdr_Set_t)
		if cube.nRefs == -1 {
			continue
		}
		vec.Vec_IntAddToEntry(result, 0, 1)
		vec.Vec_IntPush(result, cube.nLits)
		for _, lit := range cube.Lits[:cube.nLits] {
			vec.Vec_IntPush(result, lit)
		}
	}
	p.vInfCubes = cubes
	vec.Vec_IntPush(result, aig.Aig_ManRegNum(p.pAig))
	return result
}

/**Function*************************************************************

  Synopsis    [Remove clauses while maintaining the invariant.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Original macro:
// #define Pdr_ForEachCube( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 1 )

// pdrInvCubes decodes ABC's packed invariant vector: a cube count, each
// cube's literal count and literals, followed by the original flop count.
func pdrInvCubes(vInv *Vec_Int_t) ([][]int32, int) { return pdrInvDecode(vInv, false) }

func pdrInvDecode(vInv *Vec_Int_t, allowRemoved bool) ([][]int32, int) {
	data := vec.Vec_IntArray(vInv)
	if len(data) < 2 || data[0] < 0 || data[len(data)-1] < 0 {
		panic("pdr: invalid invariant header")
	}
	nRegs := int(data[len(data)-1])
	cubes := make([][]int32, 0, data[0])
	pos := 1
	for i := int32(0); i < data[0]; i++ {
		if pos >= len(data)-1 || data[pos] < 0 {
			panic("pdr: invalid invariant cube")
		}
		n := int(data[pos])
		pos++
		if n > len(data)-1-pos {
			panic("pdr: truncated invariant cube")
		}
		cube := data[pos : pos+n]
		for _, lit := range cube {
			if lit == -1 && allowRemoved {
				continue
			}
			if lit < 0 || int(lit>>1) >= nRegs {
				panic("pdr: invariant literal out of range")
			}
		}
		cubes = append(cubes, cube)
		pos += n
	}
	if pos != len(data)-1 {
		panic("pdr: unexpected invariant data")
	}
	return cubes, nRegs
}

func Pdr_InvMap(vCounts *Vec_Int_t) *Vec_Int_t {
	counts := vec.Vec_IntArray(vCounts)
	result := vec.Vec_IntStart(len(counts))
	var next int32
	for i, count := range counts {
		if count != 0 {
			vec.Vec_IntWriteEntry(result, i, next)
			next++
		}
	}
	return result
}
func Pdr_InvCounts(vInv *Vec_Int_t) *Vec_Int_t {
	cubes, nRegs := pdrInvCubes(vInv)
	counts := vec.Vec_IntStart(nRegs)
	for _, cube := range cubes {
		for _, lit := range cube {
			vec.Vec_IntAddToEntry(counts, int(lit>>1), 1)
		}
	}
	return counts
}
func Pdr_InvUsedFlopNum(vInv *Vec_Int_t) int32 {
	counts := Pdr_InvCounts(vInv)
	return int32(vec.Vec_IntSize(counts) - vec.Vec_IntCountZero(counts))
}
func Pdr_InvPrintStr(vInv *Vec_Int_t, vCounts *Vec_Int_t) *Vec_Str_t {
	cubes, nRegs := pdrInvCubes(vInv)
	if vec.Vec_IntSize(vCounts) != nRegs {
		panic("pdr: invariant count size mismatch")
	}
	mapping := Pdr_InvMap(vCounts)
	nVars := nRegs - vec.Vec_IntCountZero(vCounts)
	result := vec.Vec_StrAlloc(1000)
	for _, cube := range cubes {
		pattern := make([]byte, nVars)
		for i := range pattern {
			pattern[i] = '-'
		}
		for _, lit := range cube {
			v := int(lit >> 1)
			if vec.Vec_IntEntry(vCounts, v) == 0 {
				panic("pdr: invariant literal absent from count map")
			}
			pattern[vec.Vec_IntEntry(mapping, v)] = '1' - byte(lit&1)
		}
		vec.Vec_StrPushBuffer(result, pattern, len(pattern))
		vec.Vec_StrPushBuffer(result, []byte(" 1\n"), 3)
	}
	vec.Vec_StrPush(result, 0)
	return result
}
func Pdr_InvPrint(vInv *Vec_Int_t, fVerbose int32) {
	cubes, nRegs := pdrInvCubes(vInv)
	literalCount := 0
	for _, cube := range cubes {
		literalCount += len(cube)
	}
	fmt.Printf("Invariant contains %d clauses with %d literals and %d flops (out of %d).\n", len(cubes), literalCount, Pdr_InvUsedFlopNum(vInv), nRegs)
	if fVerbose != 0 {
		text := vec.Vec_StrArray(Pdr_InvPrintStr(vInv, Pdr_InvCounts(vInv)))
		if len(text) > 0 && text[len(text)-1] == 0 {
			text = text[:len(text)-1]
		}
		fmt.Print(string(text))
	}
}

func Pdr_InvCheck_int(p *gia.Gia_Man_t, inv *Vec_Int_t, verbose int32, solver *bsat.Solver, skip int32) int32 {
	cubes, nRegs := pdrInvDecode(inv, true)
	if nRegs != int(gia.Gia_ManRegNum(p)) {
		panic("pdr: invariant register count mismatch")
	}
	inputBase := int32(1) + gia.Gia_ManPoNum(p)
	outputBase := solver.NumVars() - gia.Gia_ManRegNum(p)
	for _, cube := range cubes {
		clause := make([]bsat.Literal, 0, len(cube))
		for _, lit := range cube {
			if lit != -1 {
				clause = append(clause, (outputBase+(lit>>1))<<1|((lit&1)^1))
			}
		}
		if len(clause) == 0 {
			return 1
		}
		if !solver.AddClause(clause...) {
			panic("pdr: invariant contradicts transition CNF")
		}
	}
	var failed int32
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		status := solver.Solve(bsat.Positive(1 + i))
		if status == bsat.Unknown {
			// Inconclusive SAT results must never certify an invariant.
			return failed + 1
		}
		if status == bsat.Satisfiable {
			if verbose != 0 {
				fmt.Printf("Coverage check failed for output %d.\n", i)
			}
			failed++
			if skip != 0 {
				return 1
			}
		}
	}
	for i, cube := range cubes {
		assumptions := make([]bsat.Literal, 0, len(cube))
		for _, lit := range cube {
			if lit != -1 {
				assumptions = append(assumptions, (inputBase+(lit>>1))<<1|(lit&1))
			}
		}
		status := solver.Solve(assumptions...)
		if status == bsat.Unknown {
			// Inconclusive SAT results must never certify an invariant.
			return failed + 1
		}
		if status == bsat.Satisfiable {
			if verbose != 0 {
				fmt.Printf("Inductiveness check failed for clause %d.\n", i)
			}
			failed++
			if skip != 0 {
				return 1
			}
		}
	}
	return failed
}

func pdrInvGiaSolver(p *gia.Gia_Man_t) *bsat.Solver {
	cnf := gia.Mf_ManGenerateCnf(p, 8, 0, 0, 0, 0)
	return pdrInvSolverFromCNF(cnf)
}

func pdrInvSolverFromCNF(cnf *mf.CNF) *bsat.Solver {
	if cnf == nil {
		return nil
	}
	solver := bsat.NewSolver()
	for solver.NumVars() < int32(cnf.Variables) {
		solver.NewVariable()
	}
	for _, clause := range cnf.Clauses {
		if !solver.AddClause(clause...) {
			return nil
		}
	}
	return solver
}

func Pdr_InvCheck(p *gia.Gia_Man_t, inv *Vec_Int_t, verbose int32) int32 {
	solver := pdrInvGiaSolver(p)
	if solver == nil {
		panic("pdr: could not generate invariant verification CNF")
	}
	return Pdr_InvCheck_int(p, inv, verbose, solver, 0)
}

// Pdr_InvMinimize removes whole clauses while preserving coverage and induction.
// A nil return means no clause was removed or a SAT call was inconclusive.
func Pdr_InvMinimize(p *gia.Gia_Man_t, inv *Vec_Int_t, verbose int32) *Vec_Int_t {
	cubes, nRegs := pdrInvCubes(inv)
	if nRegs != int(gia.Gia_ManRegNum(p)) {
		panic("pdr: invariant register count mismatch")
	}
	solver := pdrInvGiaSolver(p)
	if solver == nil {
		panic("pdr: could not generate invariant verification CNF")
	}
	outputBase := solver.NumVars() - int32(nRegs)
	inputBase := int32(1) + gia.Gia_ManPoNum(p)
	auxBase := solver.NumVars()
	for range cubes {
		solver.NewVariable()
	}
	for i, cube := range cubes {
		clause := []bsat.Literal{bsat.Negative(auxBase + int32(i))}
		for _, lit := range cube {
			clause = append(clause, (outputBase+(lit>>1))<<1|((lit&1)^1))
		}
		if !solver.AddClause(clause...) {
			panic("pdr: inconsistent invariant clause")
		}
	}
	removed := make([]bool, len(cubes))
	count := 0
	for i := range cubes {
		assumptions := make([]bsat.Literal, 0, len(cubes)+nRegs)
		for k := range cubes {
			if k != i && !removed[k] {
				assumptions = append(assumptions, bsat.Positive(auxBase+int32(k)))
			}
		}
		prefix := len(assumptions)
		cannot := false
		for k := int32(0); k < gia.Gia_ManPoNum(p); k++ {
			status := solver.Solve(append(assumptions[:prefix], bsat.Positive(1+k))...)
			if status == bsat.Unknown {
				return nil
			}
			if status == bsat.Satisfiable {
				cannot = true
				break
			}
		}
		if cannot {
			continue
		}
		for k, cube := range cubes {
			if k == i || removed[k] {
				continue
			}
			assumptions = assumptions[:prefix]
			for _, lit := range cube {
				assumptions = append(assumptions, (inputBase+(lit>>1))<<1|(lit&1))
			}
			status := solver.Solve(assumptions...)
			if status == bsat.Unknown {
				return nil
			}
			if status == bsat.Satisfiable {
				cannot = true
				break
			}
		}
		if !cannot {
			removed[i] = true
			count++
			if verbose != 0 {
				fmt.Printf("Removing clause %d.\n", i)
			}
		}
	}
	if count == 0 {
		return nil
	}
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(inv))
	vec.Vec_IntPush(result, int32(len(cubes)-count))
	for i, cube := range cubes {
		if removed[i] {
			continue
		}
		vec.Vec_IntPush(result, int32(len(cube)))
		for _, lit := range cube {
			vec.Vec_IntPush(result, lit)
		}
	}
	vec.Vec_IntPush(result, int32(nRegs))
	return result
}

// Pdr_InvMinimizeLits greedily drops individual cube literals.
func Pdr_InvMinimizeLits(p *gia.Gia_Man_t, inv *Vec_Int_t, verbose int32) *Vec_Int_t {
	// Work on a copy: temporary -1 markers are not valid in a packed invariant.
	data := vec.Vec_IntArray(inv)
	working := vec.Vec_IntAllocArrayCopy(data, len(data))
	cubes, nRegs := pdrInvCubes(working)
	if nRegs != int(gia.Gia_ManRegNum(p)) {
		panic("pdr: invariant register count mismatch")
	}
	cnf := gia.Mf_ManGenerateCnf(p, 8, 0, 0, 0, 0)
	if cnf == nil {
		panic("pdr: could not generate invariant verification CNF")
	}
	count := 0
	for i, cube := range cubes {
		for k, literal := range cube {
			cube[k] = -1
			solver := pdrInvSolverFromCNF(cnf)
			if solver == nil {
				cube[k] = literal
				panic("pdr: could not generate invariant verification CNF")
			}
			if Pdr_InvCheck_int(p, working, 0, solver, 1) != 0 {
				cube[k] = literal
			} else {
				count++
				if verbose != 0 {
					fmt.Printf("Removing lit %d from clause %d.\n", k, i)
				}
			}
		}
	}
	if count == 0 {
		return nil
	}
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(inv))
	vec.Vec_IntPush(result, int32(len(cubes)))
	for _, cube := range cubes {
		size := int32(0)
		for _, lit := range cube {
			if lit != -1 {
				size++
			}
		}
		vec.Vec_IntPush(result, size)
		for _, lit := range cube {
			if lit != -1 {
				vec.Vec_IntPush(result, lit)
			}
		}
	}
	vec.Vec_IntPush(result, int32(nRegs))
	return result
}
