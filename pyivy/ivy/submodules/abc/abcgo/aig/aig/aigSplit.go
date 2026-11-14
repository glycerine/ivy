// Property splitting, ported from ABC aigSplit.c (Alan Mishchenko).
package aig

import (
	"fmt"
	"time"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab/cascade"
)

func Aig_NodeBddToMuxes_rec(dd *cudd.DdManager, function *cudd.DdNode, destination *Graph, mapping map[*cudd.DdNode]Edge) Edge {
	if cudd.Cudd_IsComplement(function) {
		panic("aig: BDD-to-MUX recursion requires a regular node")
	}
	if result, ok := mapping[function]; ok {
		return result
	}
	low := cudd.Cudd_E(function)
	a := Aig_NodeBddToMuxes_rec(dd, cudd.Cudd_Regular(low), destination, mapping)
	if cudd.Cudd_IsComplement(low) {
		a = a.Not()
	}
	b := Aig_NodeBddToMuxes_rec(dd, cudd.Cudd_T(function), destination, mapping)
	variable, ok := mapping[cudd.Cudd_bddIthVar(dd, int(function.Index))]
	if !ok {
		panic("aig: unmapped BDD variable")
	}
	result := destination.Mux(variable, b, a)
	mapping[function] = result
	return result
}

// Typed local scratch replaces the source pData BDD cache, keeping BDD handles
// separate from the graph's AIG-copy maps and unrelated object payloads.
func Aig_ManBuildPoBdd_rec(p *Graph, e Edge, dd *cudd.DdManager, mapping map[int]*cudd.DdNode) *cudd.DdNode {
	result, ok := mapping[e.NodeID()]
	if !ok {
		n := *p.node(e)
		a := Aig_ManBuildPoBdd_rec(p, n.Fanins[0], dd, mapping)
		switch n.Kind {
		case Buffer, Output:
			result = a
		case And, ExclusiveOr:
			b := Aig_ManBuildPoBdd_rec(p, n.Fanins[1], dd, mapping)
			if n.Kind == And {
				result = cudd.Cudd_bddAnd(dd, a, b)
			} else {
				result = cudd.Cudd_bddXor(dd, a, b)
			}
		default:
			panic("aig: unmapped BDD terminal")
		}
		if result == nil {
			panic("aig: failed to build property BDD")
		}
		cudd.Cudd_Ref(result)
		mapping[e.NodeID()] = result
	}
	return cudd.Cudd_NotCond(result, e.Complemented())
}

// The returned function has one external BDD reference owned by the caller.
func Aig_ManBuildPoBdd(p *Graph, function **cudd.DdNode) *cudd.DdManager {
	if len(p.outputs)-int(p.nRegs) != 1 || function == nil {
		panic("aig: splitting requires one property output")
	}
	dd := cudd.Cudd_Init(uint32(len(p.inputs)), 0, cudd.CUDD_UNIQUE_SLOTS, cudd.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		panic("aig: failed to create BDD manager")
	}
	cudd.Cudd_AutodynEnable(dd, cudd.CUDD_REORDER_SYMM_SIFT)
	mapping := map[int]*cudd.DdNode{0: cudd.Cudd_ReadOne(dd)}
	cudd.Cudd_Ref(mapping[0])
	for i, input := range p.inputs {
		mapping[input.NodeID()] = cudd.Cudd_bddIthVar(dd, i)
		cudd.Cudd_Ref(mapping[input.NodeID()])
	}
	*function = Aig_ManBuildPoBdd_rec(p, p.outputDriver(p.outputs[0]), dd, mapping)
	cudd.Cudd_Ref(*function)
	for _, node := range mapping {
		cudd.Cudd_RecursiveDeref(dd, node)
	}
	cudd.Cudd_ReduceHeap(dd, cudd.CUDD_REORDER_SYMM_SIFT, 1)
	return dd
}

// Each result retains its own reference, including zero and duplicate cofactors.
func Aig_ManCofactorBdds(p *Graph, subset []*cudd.DdNode, dd *cudd.DdManager, function *cudd.DdNode) []*cudd.DdNode {
	if len(subset) > 16 {
		panic("aig: at most 16 cofactoring variables")
	}
	seen := make(map[*cudd.DdNode]bool)
	for _, variable := range subset {
		if variable == nil || int(variable.Index) >= len(dd.Vars) || variable != dd.Vars[variable.Index] || seen[variable] {
			panic("aig: invalid cofactoring variables")
		}
		seen[variable] = true
	}
	results := make([]*cudd.DdNode, 0, 1<<uint(len(subset)))
	for code := 0; code < 1<<uint(len(subset)); code++ {
		cube := cascade.Extra_bddBitsToCube(dd, code, len(subset), subset, true)
		cudd.Cudd_Ref(cube)
		cofactor := cudd.Cudd_Cofactor(dd, function, cube)
		cudd.Cudd_Ref(cofactor)
		result := cudd.Cudd_bddAnd(dd, cofactor, cube)
		cudd.Cudd_Ref(result)
		cudd.Cudd_RecursiveDeref(dd, cofactor)
		cudd.Cudd_RecursiveDeref(dd, cube)
		results = append(results, result)
	}
	return results
}

func Aig_ManConvertBddsToAigs(p *Graph, dd *cudd.DdManager, cofactors []*cudd.DdNode) *Graph {
	q := Aig_ManStartFrom(p)
	mapping := map[*cudd.DdNode]Edge{cudd.Cudd_ReadOne(dd): q.Const1()}
	if len(dd.Vars) != len(q.inputs) {
		panic("aig: inconsistent BDD variable count")
	}
	for i, input := range q.inputs {
		mapping[cudd.Cudd_bddIthVar(dd, i)] = input
	}
	for _, function := range cofactors {
		if function == cudd.Cudd_ReadLogicZero(dd) {
			continue
		}
		result := Aig_NodeBddToMuxes_rec(dd, cudd.Cudd_Regular(function), q, mapping)
		if cudd.Cudd_IsComplement(function) {
			result = result.Not()
		}
		q.AddOutput(result)
	}
	for _, output := range p.outputs[1:] {
		q.AddOutput(Aig_ManDup_rec(q, p, p.outputDriver(output)))
	}
	Aig_ManCleanup(q)
	Aig_ManSetRegNum(q, p.nRegs)
	return q
}

func Aig_ManVecRandSubset(values []Edge, count int32) []Edge {
	if count < 0 {
		panic("aig: negative subset size")
	}
	result := append([]Edge(nil), values...)
	for len(result) > int(count) {
		value := result[int(Aig_ManRandom(0)%uint32(len(result)))]
		// Vec_PtrRemove removes the last equal entry, retaining survivor order.
		for i := len(result) - 1; i >= 0; i-- {
			if result[i] == value {
				result = append(result[:i], result[i+1:]...)
				break
			}
		}
	}
	return result
}

func Aig_ManSplit(p *Graph, variables, verbose int32) *Graph {
	start := time.Now()
	if len(p.outputs)-int(p.nRegs) != 1 {
		fmt.Println("Currently works only for one primary output.")
		return nil
	}
	if variables < 1 {
		fmt.Println("The number of cofactoring variables should be a positive number.")
		return nil
	}
	if variables > 16 {
		fmt.Println("The number of cofactoring variables should be less than 17.")
		return nil
	}
	support := Aig_Support(p, Aig_Regular(p.outputDriver(p.outputs[0])))
	if len(support) == 0 {
		fmt.Println("Property output function is a constant.")
		return nil
	}
	var function *cudd.DdNode
	dd := Aig_ManBuildPoBdd(p, &function)
	defer cudd.Cudd_Quit(dd)
	defer cudd.Cudd_RecursiveDeref(dd, function)
	if verbose != 0 {
		fmt.Printf("Support =%5d.  BDD size =%6d.  ", len(support), cudd.Cudd_DagSize(function))
	}
	selected := Aig_ManVecRandSubset(support, variables)
	positions := make(map[int]int, len(p.inputs))
	for i, input := range p.inputs {
		positions[input.NodeID()] = i
	}
	subset := make([]*cudd.DdNode, len(selected))
	for i, input := range selected {
		subset[i] = cudd.Cudd_bddIthVar(dd, positions[input.NodeID()])
	}
	cofactors := Aig_ManCofactorBdds(p, subset, dd, function)
	defer func() {
		for _, cofactor := range cofactors {
			cudd.Cudd_RecursiveDeref(dd, cofactor)
		}
	}()
	result := Aig_ManConvertBddsToAigs(p, dd, cofactors)
	if verbose != 0 {
		fmt.Printf("Created %d cofactors (out of %d).  Time=%s\n", len(result.outputs)-int(result.nRegs), len(cofactors), time.Since(start))
	}
	return result
}
