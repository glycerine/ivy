package cudd

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"unsafe"

	"github.com/glycerine/abcgo/bdd/cudd/core"
	epdpkg "github.com/glycerine/abcgo/bdd/epd"
)

type EpDouble = epdpkg.EpDouble

/**CFile***********************************************************************

  FileName    [cuddUtil.c]

  PackageName [cudd]

  Synopsis    [Utility functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_PrintMinterm()
                <li> Cudd_bddPrintCover()
                <li> Cudd_PrintDebug()
                <li> Cudd_DagSize()
                <li> Cudd_EstimateCofactor()
                <li> Cudd_EstimateCofactorSimple()
                <li> Cudd_SharingSize()
                <li> Cudd_CountMinterm()
                <li> Cudd_EpdCountMinterm()
                <li> Cudd_CountPath()
                <li> Cudd_CountPathsToNonZero()
                <li> Cudd_Support()
                <li> Cudd_SupportIndex()
                <li> Cudd_SupportSize()
                <li> Cudd_VectorSupport()
                <li> Cudd_VectorSupportIndex()
                <li> Cudd_VectorSupportSize()
                <li> Cudd_ClassifySupport()
                <li> Cudd_CountLeaves()
                <li> Cudd_bddPickOneCube()
                <li> Cudd_bddPickOneMinterm()
                <li> Cudd_bddPickArbitraryMinterms()
                <li> Cudd_SubsetWithMaskVars()
                <li> Cudd_FirstCube()
                <li> Cudd_NextCube()
                <li> Cudd_bddComputeCube()
                <li> Cudd_addComputeCube()
                <li> Cudd_FirstNode()
                <li> Cudd_NextNode()
                <li> Cudd_GenFree()
                <li> Cudd_IsGenEmpty()
                <li> Cudd_IndicesToCube()
                <li> Cudd_PrintVersion()
                <li> Cudd_AverageDistance()
                <li> Cudd_Random()
                <li> Cudd_Srandom()
                <li> Cudd_Density()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddP()
                <li> cuddStCountfree()
                <li> cuddCollectNodes()
                <li> cuddNodeArray()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> dp2()
                <li> ddPrintMintermAux()
                <li> ddDagInt()
                <li> ddCountMintermAux()
                <li> ddEpdCountMintermAux()
                <li> ddCountPathAux()
                <li> ddSupportStep()
                <li> ddClearFlag()
                <li> ddLeavesInt()
                <li> ddPickArbitraryMinterms()
                <li> ddPickRepresentativeCube()
                <li> ddEpdFree()
                </ul>]

  Author      [Fabio Somenzi]

  Copyright   [Copyright (c) 1995-2004, Regents of the University of Colorado

  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:

  Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright
  notice, this list of conditions and the following disclaimer in the
  documentation and/or other materials provided with the distribution.

  Neither the name of the University of Colorado nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
  FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
  BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
  LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
  ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.]

******************************************************************************/

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START

/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/* Random generator constants. */
const MODULUS1 = 2147483563
const LEQA1 = 40014
const LEQQ1 = 53668
const LEQR1 = 12211
const MODULUS2 = 2147483399
const LEQA2 = 40692
const LEQQ2 = 52774
const LEQR2 = 3791
const STAB_SIZE = 64
const STAB_DIV = 1 + (MODULUS1-1)/STAB_SIZE

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

var background, zero *DdNode

var cuddRand int64
var cuddRand2 int64
var shuffleSelect int64
var shuffleTable [STAB_SIZE]int64

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

// Original macro:
// #define bang(f) ((Cudd_IsComplement(f)) ? '!' : ' ')

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int dp2 (DdManager *dd, DdNode *f, st__table *t);

// static void ddPrintMintermAux (DdManager *dd, DdNode *node, int *list);

// static int ddDagInt (DdNode *n);

// static int cuddNodeArrayRecur (DdNode *f, DdNodePtr *table, int index);

// static int cuddEstimateCofactor (DdManager *dd, st__table *table, DdNode * node, int i, int phase, DdNode ** ptr);

// static DdNode * cuddUniqueLookup (DdManager * unique, int  index, DdNode * T, DdNode * E);

// static int cuddEstimateCofactorSimple (DdNode * node, int i);

// static double ddCountMintermAux (DdNode *node, double max, DdHashTable *table);

// static int ddEpdCountMintermAux (DdNode *node, EpDouble *max, EpDouble *epd, st__table *table);

// static double ddCountPathAux (DdNode *node, st__table *table);

// static double ddCountPathsToNonZero (DdNode * N, st__table * table);

// static void ddSupportStep (DdNode *f, int *support);

// static void ddClearFlag (DdNode *f);

// static int ddLeavesInt (DdNode *n);

// static int ddPickArbitraryMinterms (DdManager *dd, DdNode *node, int nvars, int nminterms, char **string);

// static int ddPickRepresentativeCube (DdManager *dd, DdNode *node, double *weight, char *string);

// static      st__retval ddEpdFree (char * key, char * value, char * arg);

/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/*
*Function********************************************************************

	Synopsis    [Prints a disjoint sum of products.]

	Description [Prints a disjoint sum of product cover for the function
	rooted at node. Each product corresponds to a path from node to a
	leaf node different from the logical zero, and different from the
	background value. Uses the package default output file.  Returns 1
	if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_PrintDebug Cudd_bddPrintCover]

*****************************************************************************
*/
func Cudd_PrintMinterm(manager *DdManager, node *DdNode) int32 {
	if manager == nil || node == nil {
		return 0
	}
	out := core.Cudd_Output(manager)
	background = manager.Background
	zero = Cudd_Not(manager.One)
	path := make([]int, len(manager.Vars))
	for i := range path {
		path[i] = 2 // Unvisited variables are don't-cares in each path cube.
	}
	var printPaths func(*DdNode)
	printPaths = func(edge *DdNode) {
		regular := Cudd_Regular(edge)
		if Cudd_IsConstant(regular) {
			if edge == background || edge == zero {
				return
			}
			for _, value := range path {
				if value == 0 {
					_, _ = fmt.Fprint(out, "0")
				} else if value == 1 {
					_, _ = fmt.Fprint(out, "1")
				} else {
					_, _ = fmt.Fprint(out, "-")
				}
			}
			_, _ = fmt.Fprintf(out, " % g\n", regular.Value)
			return
		}
		index := int(regular.Index)
		if index < 0 || index >= len(path) {
			return
		}
		path[index] = 0
		printPaths(Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(edge)))
		path[index] = 1
		printPaths(Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(edge)))
		path[index] = 2
	}
	printPaths(node)
	return 1

} /* end of Cudd_PrintMinterm */

/*
*Function********************************************************************

	Synopsis    [Prints a sum of prime implicants of a BDD.]

	Description [Prints a sum of product cover for an incompletely
	specified function given by a lower bound and an upper bound.  Each
	product is a prime implicant obtained by expanding the product
	corresponding to a path from node to the constant one.  Uses the
	package default output file.  Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_PrintMinterm]

*****************************************************************************
*/
func Cudd_bddPrintCover(dd *DdManager, l *DdNode, u *DdNode) int32 {
	if dd == nil || l == nil || u == nil {
		return 0
	}
	out := core.Cudd_Output(dd)
	array := make([]int32, len(dd.Vars))
	lb := l
	for lb != dd.LogicZero {
		var length int32
		implicant := Cudd_LargestCube(dd, lb, &length)
		if implicant == nil {
			return 0
		}
		prime := Cudd_bddMakePrime(dd, implicant, u)
		if prime == nil {
			return 0
		}
		lb = core.Cudd_bddAnd(dd, lb, core.Cudd_Not(prime))
		if lb == nil {
			return 0
		}
		if Cudd_BddToCubeArray(dd, prime, array) == 0 {
			return 0
		}
		for _, value := range array {
			switch value {
			case 0:
				_, _ = fmt.Fprint(out, "0")
			case 1:
				_, _ = fmt.Fprint(out, "1")
			case 2:
				_, _ = fmt.Fprint(out, "-")
			default:
				_, _ = fmt.Fprint(out, "?")
			}
		}
		_, _ = fmt.Fprint(out, " 1\n")
	}
	_, _ = fmt.Fprint(out, "\n")
	return 1

} /* end of Cudd_bddPrintCover */

/*
*Function********************************************************************

	Synopsis    [Prints to the standard output a DD and its statistics.]

	Description [Prints to the standard output a DD and its statistics.
	The statistics include the number of nodes, the number of leaves, and
	the number of minterms. (The number of minterms is the number of
	assignments to the variables that cause the function to be different
	from the logical zero (for BDDs) and from the background value (for
	ADDs.) The statistics are printed if pr &gt; 0. Specifically:
	<ul>
	<li> pr = 0 : prints nothing
	<li> pr = 1 : prints counts of nodes and minterms
	<li> pr = 2 : prints counts + disjoint sum of product
	<li> pr = 3 : prints counts + list of nodes
	<li> pr &gt; 3 : prints counts + disjoint sum of product + list of nodes
	</ul>
	For the purpose of counting the number of minterms, the function is
	supposed to depend on n variables. Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_DagSize Cudd_CountLeaves Cudd_CountMinterm
	Cudd_PrintMinterm]

*****************************************************************************
*/
func Cudd_PrintDebug(dd *DdManager, f *DdNode, n int32, pr int32) int32 {
	if dd == nil {
		return 0
	}
	out := core.Cudd_Output(dd)
	if f == nil {
		_, _ = fmt.Fprintln(out, ": is the NULL DD")
		cuddFlushOutput(out)
		return 0
	}
	if pr == 0 {
		return 1
	}
	if f == dd.LogicZero || f == Cudd_Not(dd.One) {
		_, _ = fmt.Fprintln(out, ": is the zero DD")
		cuddFlushOutput(out)
		return 1
	}
	nodes := int32(Cudd_DagSize(f))
	leaves := Cudd_CountLeaves(f)
	minterms := Cudd_CountMinterm(dd, f, n)
	_, _ = fmt.Fprintf(out, ": %d nodes %d leaves %g minterms\n", nodes, leaves, minterms)
	if minterms == float64(CUDD_OUT_OF_MEM) {
		cuddFlushOutput(out)
		return 0
	}
	if pr > 2 && cuddP(dd, f) == 0 {
		cuddFlushOutput(out)
		return 0
	}
	if pr == 2 || pr > 3 {
		if Cudd_PrintMinterm(dd, f) == 0 {
			cuddFlushOutput(out)
			return 0
		}
		_, _ = fmt.Fprintln(out)
	}
	cuddFlushOutput(out)
	return 1

} /* end of Cudd_PrintDebug */

/**Function********************************************************************

  Synopsis    [Counts the number of nodes in a DD.]

  Description [Counts the number of nodes in a DD. Returns the number
  of nodes in the graph rooted at node.]

  SideEffects [None]

  SeeAlso     [Cudd_SharingSize Cudd_PrintDebug]

******************************************************************************/
/**Function********************************************************************

  Synopsis    [Estimates the number of nodes in a cofactor of a DD.]

  Description [Estimates the number of nodes in a cofactor of a DD.
  Returns an estimate of the number of nodes in a cofactor of
  the graph rooted at node with respect to the variable whose index is i.
  In case of failure, returns CUDD_OUT_OF_MEM.
  This function uses a refinement of the algorithm of Cabodi et al.
  (ICCAD96). The refinement allows the procedure to account for part
  of the recombination that may occur in the part of the cofactor above
  the cofactoring variable. This procedure does no create any new node.
  It does keep a small table of results; therefore it may run out of memory.
  If this is a concern, one should use Cudd_EstimateCofactorSimple, which
  is faster, does not allocate any memory, but is less accurate.]

  SideEffects [None]

  SeeAlso     [Cudd_DagSize Cudd_EstimateCofactorSimple]

******************************************************************************/
func Cudd_EstimateCofactor(dd *DdManager, f *DdNode, i int32, phase int32) int32 {
	if dd == nil || f == nil || i < 0 {
		return CUDD_OUT_OF_MEM
	}
	type nodeKey struct {
		index uint32
		then  uint64
		else_ uint64
	}
	const falseID, trueID uint64 = 0, 1
	ids := make(map[nodeKey]uint64)
	memo := make(map[*DdNode]uint64)
	var nextID uint64 = 2
	var cofactor func(*DdNode) uint64
	cofactor = func(edge *DdNode) uint64 {
		if id, ok := memo[edge]; ok {
			return id
		}
		regular := Cudd_Regular(edge)
		if Cudd_IsConstant(regular) {
			if edge == dd.One {
				memo[edge] = trueID
				return trueID
			}
			memo[edge] = falseID
			return falseID
		}
		complement := Cudd_IsComplement(edge)
		var id uint64
		if int32(regular.Index) == i {
			child := Cudd_E(regular)
			if phase != 0 {
				child = Cudd_T(regular)
			}
			id = cofactor(Cudd_NotCond(child, complement))
		} else {
			thenID := cofactor(Cudd_NotCond(Cudd_T(regular), complement))
			elseID := cofactor(Cudd_NotCond(Cudd_E(regular), complement))
			if thenID == elseID {
				id = thenID
			} else {
				phase := thenID & 1
				thenID ^= phase
				elseID ^= phase
				key := nodeKey{index: regular.Index, then: thenID, else_: elseID}
				id, ok := ids[key]
				if !ok {
					id = nextID
					nextID += 2
					ids[key] = id
				}
				id |= phase
			}
		}
		memo[edge] = id
		return id
	}
	cofactor(f)
	return int32(len(ids) + 1) // Include the single canonical Boolean terminal.
}

/*
*Function********************************************************************

	Synopsis    [Estimates the number of nodes in a cofactor of a DD.]

	Description [Estimates the number of nodes in a cofactor of a DD.
	Returns an estimate of the number of nodes in the positive cofactor of
	the graph rooted at node with respect to the variable whose index is i.
	This procedure implements with minor changes the algorithm of Cabodi et al.
	(ICCAD96). It does not allocate any memory, it does not change the
	state of the manager, and it is fast. However, it has been observed to
	overestimate the size of the cofactor by as much as a factor of 2.]

	SideEffects [None]

	SeeAlso     [Cudd_DagSize]

*****************************************************************************
*/
func Cudd_EstimateCofactorSimple(node *DdNode, i int32) int32 {
	if node == nil {
		return 0
	}
	visited := make(map[*DdNode]struct{})
	var estimate func(*DdNode) int32
	estimate = func(edge *DdNode) int32 {
		current := Cudd_Regular(edge)
		if _, ok := visited[current]; ok {
			return 0
		}
		visited[current] = struct{}{}
		if Cudd_IsConstant(current) {
			return 1
		}
		thenCount := estimate(Cudd_T(current))
		if int32(current.Index) == i {
			return thenCount
		}
		elseCount := estimate(Cudd_E(current))
		return 1 + thenCount + elseCount
	}
	return estimate(Cudd_Regular(node))

} /* end of Cudd_EstimateCofactorSimple */

/*
*Function********************************************************************

	Synopsis    [Counts the number of nodes in an array of DDs.]

	Description [Counts the number of nodes in an array of DDs. Shared
	nodes are counted only once.  Returns the total number of nodes.]

	SideEffects [None]

	SeeAlso     [Cudd_DagSize]

*****************************************************************************
*/
func Cudd_SharingSize(nodeArray []*DdNode, n int32) int32 {
	if n < 0 {
		return 0
	}
	if int(n) > len(nodeArray) {
		n = int32(len(nodeArray))
	}
	seen := make(map[*DdNode]struct{})
	stack := make([]*DdNode, 0, n)
	for j := int32(0); j < n; j++ {
		if nodeArray[j] != nil {
			stack = append(stack, nodeArray[j])
		}
	}
	for len(stack) != 0 {
		last := len(stack) - 1
		current := Cudd_Regular(stack[last])
		stack = stack[:last]
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}
		if !Cudd_IsConstant(current) {
			stack = append(stack, Cudd_T(current), Cudd_E(current))
		}
	}
	return int32(len(seen))

} /* end of Cudd_SharingSize */

/*
*Function********************************************************************

	Synopsis    [Counts the number of minterms of a DD.]

	Description [Counts the number of minterms of a DD. The function is
	assumed to depend on nvars variables. The minterm count is
	represented as a double, to allow for a larger number of variables.
	Returns the number of minterms of the function rooted at node if
	successful; (double) CUDD_OUT_OF_MEM otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_PrintDebug Cudd_CountPath]

*****************************************************************************
*/
func Cudd_CountMinterm(manager *DdManager, node *DdNode, nvars int32) float64 {
	count := cuddMintermCountExact(manager, node, int(nvars))
	if count == nil {
		return float64(CUDD_OUT_OF_MEM)
	}
	value, _ := new(big.Float).SetInt(count).Float64()
	return value

} /* end of Cudd_CountMinterm */

func cuddMintermCountExact(manager *DdManager, node *DdNode, nvars int) *big.Int {
	if manager == nil || node == nil || nvars < 0 {
		return nil
	}
	type state struct {
		edge  *DdNode
		level int
	}
	memo := make(map[state]*big.Int)
	var count func(*DdNode, int) *big.Int
	count = func(edge *DdNode, level int) *big.Int {
		if level > nvars {
			return new(big.Int)
		}
		key := state{edge: edge, level: level}
		if cached, ok := memo[key]; ok {
			return new(big.Int).Set(cached)
		}
		regular := Cudd_Regular(edge)
		if Cudd_IsConstant(regular) {
			result := new(big.Int)
			if edge != manager.Background && edge != Cudd_Not(manager.One) {
				result.Lsh(big.NewInt(1), uint(nvars-level))
			}
			memo[key] = new(big.Int).Set(result)
			return result
		}
		index := int(regular.Index)
		variableLevel := index
		if index < len(manager.Perm) && manager.Perm[index] >= 0 {
			variableLevel = manager.Perm[index]
		}
		if variableLevel < level || variableLevel >= nvars {
			return new(big.Int)
		}
		complement := Cudd_IsComplement(edge)
		thenCount := count(Cudd_NotCond(Cudd_T(regular), complement), variableLevel+1)
		elseCount := count(Cudd_NotCond(Cudd_E(regular), complement), variableLevel+1)
		result := new(big.Int).Add(thenCount, elseCount)
		result.Lsh(result, uint(variableLevel-level))
		memo[key] = new(big.Int).Set(result)
		return result
	}
	result := count(node, 0)
	return result
}

/*
*Function********************************************************************

	Synopsis    [Counts the number of paths of a DD.]

	Description [Counts the number of paths of a DD.  Paths to all
	terminal nodes are counted. The path count is represented as a
	double, to allow for a larger number of variables.  Returns the
	number of paths of the function rooted at node if successful;
	(double) CUDD_OUT_OF_MEM otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_CountMinterm]

*****************************************************************************
*/
func Cudd_CountPath(node *DdNode) float64 {
	if node == nil {
		return 0
	}
	memo := make(map[*DdNode]float64)
	var count func(*DdNode) float64
	count = func(edge *DdNode) float64 {
		current := Cudd_Regular(edge)
		if paths, ok := memo[current]; ok {
			return paths
		}
		if Cudd_IsConstant(current) {
			memo[current] = 1
			return 1
		}
		paths := count(Cudd_T(current)) + count(Cudd_E(current))
		memo[current] = paths
		return paths
	}
	return count(node)

} /* end of Cudd_CountPath */

/*
*Function********************************************************************

	Synopsis    [Counts the number of minterms of a DD with extended precision.]

	Description [Counts the number of minterms of a DD with extended precision.
	The function is assumed to depend on nvars variables. The minterm count is
	represented as an EpDouble, to allow any number of variables.
	Returns 0 if successful; CUDD_OUT_OF_MEM otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_PrintDebug Cudd_CountPath]

*****************************************************************************
*/
func Cudd_EpdCountMinterm(manager *DdManager, node *DdNode, nvars int32, epd *EpDouble) int32 {
	if epd == nil {
		return CUDD_OUT_OF_MEM
	}
	count := cuddMintermCountExact(manager, node, int(nvars))
	if count == nil {
		*epd = EpDouble{}
		return CUDD_OUT_OF_MEM
	}
	if count.Sign() == 0 {
		*epd = EpDouble{}
		return 0
	}
	fraction := new(big.Float).SetPrec(64)
	exponent := new(big.Float).SetPrec(64).SetInt(count).MantExp(fraction)
	value, _ := fraction.Float64()
	*epd = EpDouble{Value: value, Exponent: exponent}
	return 0

} /* end of Cudd_EpdCountMinterm */

/*
*Function********************************************************************

	Synopsis    [Counts the number of paths to a non-zero terminal of a DD.]

	Description [Counts the number of paths to a non-zero terminal of a
	DD.  The path count is
	represented as a double, to allow for a larger number of variables.
	Returns the number of paths of the function rooted at node.]

	SideEffects [None]

	SeeAlso     [Cudd_CountMinterm Cudd_CountPath]

*****************************************************************************
*/
func Cudd_CountPathsToNonZero(node *DdNode) float64 {
	if node == nil {
		return 0
	}
	memo := make(map[*DdNode]float64)
	var count func(*DdNode) float64
	count = func(edge *DdNode) float64 {
		regular := Cudd_Regular(edge)
		if paths, ok := memo[edge]; ok {
			return paths
		}
		if Cudd_IsConstant(regular) {
			if !Cudd_IsComplement(edge) && regular.Value != 0 {
				memo[edge] = 1
				return 1
			}
			memo[edge] = 0
			return 0
		}
		paths := count(Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(edge))) +
			count(Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(edge)))
		memo[edge] = paths
		return paths
	}
	return count(node)

} /* end of Cudd_CountPathsToNonZero */

/**Function********************************************************************

  Synopsis    [Finds the variables on which a DD depends.]

  Description [Finds the variables on which a DD depends.
  Returns a BDD consisting of the product of the variables if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_VectorSupport Cudd_ClassifySupport]

******************************************************************************/

func Cudd_SupportIndex(dd *DdManager, f *DdNode) []int32 {
	return cuddSupportIndex(dd, []*DdNode{f})
}

func cuddSupportIndex(dd *DdManager, roots []*DdNode) []int32 {
	if dd == nil {
		return nil
	}
	result := make([]int32, len(dd.Vars))
	for _, root := range roots {
		for _, index := range dd.Support(root) {
			if index >= 0 && index < len(result) {
				result[index] = 1
			}
		}
	}
	return result
}

func Cudd_VectorSupport(dd *DdManager, roots []*DdNode, n int32) *DdNode {
	if dd == nil || n < 0 || int(n) > len(roots) {
		return nil
	}
	indices := cuddSupportIndex(dd, roots[:n])
	result := dd.One
	for level := len(dd.InvPerm) - 1; level >= 0; level-- {
		index := dd.InvPerm[level]
		if index >= 0 && index < len(indices) && indices[index] != 0 {
			result = core.Cudd_bddAnd(dd, result, core.Cudd_bddIthVar(dd, index))
			if result == nil {
				return nil
			}
		}
	}
	return result
}

func Cudd_VectorSupportIndex(dd *DdManager, roots []*DdNode, n int32) []int32 {
	if n < 0 || int(n) > len(roots) {
		return nil
	}
	return cuddSupportIndex(dd, roots[:n])
}

func Cudd_VectorSupportSize(dd *DdManager, roots []*DdNode, n int32) int32 {
	indices := Cudd_VectorSupportIndex(dd, roots, n)
	var result int32
	for _, present := range indices {
		result += present
	}
	return result
}

// Cudd_ClassifySupport partitions the variables shared by f and g from
// those unique to each function. The three result pointers receive BDD cubes.
func Cudd_ClassifySupport(dd *DdManager, f, g *DdNode, common, onlyF, onlyG **DdNode) int32 {
	if dd == nil || f == nil || g == nil || common == nil || onlyF == nil || onlyG == nil {
		return 0
	}
	supportF := cuddSupportIndex(dd, []*DdNode{f})
	supportG := cuddSupportIndex(dd, []*DdNode{g})
	sets := [3][]int32{make([]int32, len(supportF)), make([]int32, len(supportF)), make([]int32, len(supportF))}
	for i := range supportF {
		switch {
		case supportF[i] != 0 && supportG[i] != 0:
			sets[0][i] = 1
		case supportF[i] != 0:
			sets[1][i] = 1
		case supportG[i] != 0:
			sets[2][i] = 1
		}
	}
	cubes := [3]*DdNode{dd.One, dd.One, dd.One}
	for level := len(dd.InvPerm) - 1; level >= 0; level-- {
		index := dd.InvPerm[level]
		if index < 0 || index >= len(supportF) {
			continue
		}
		for set := range sets {
			if sets[set][index] != 0 {
				cubes[set] = core.Cudd_bddAnd(dd, cubes[set], core.Cudd_bddIthVar(dd, index))
				if cubes[set] == nil {
					return 0
				}
			}
		}
	}
	*common, *onlyF, *onlyG = cubes[0], cubes[1], cubes[2]
	return 1
}

func Cudd_CountLeaves(node *DdNode) int32 {
	if node == nil {
		return 0
	}
	visited := make(map[*DdNode]struct{})
	leaves := make(map[*DdNode]struct{})
	stack := []*DdNode{node}
	for len(stack) != 0 {
		last := len(stack) - 1
		current := Cudd_Regular(stack[last])
		stack = stack[:last]
		if Cudd_IsConstant(current) {
			leaves[current] = struct{}{}
			continue
		}
		if _, ok := visited[current]; ok {
			continue
		}
		visited[current] = struct{}{}
		stack = append(stack, Cudd_T(current), Cudd_E(current))
	}
	return int32(len(leaves))

} /* end of Cudd_CountLeaves */

/*
*Function********************************************************************

	Synopsis    [Picks one on-set cube randomly from the given DD.]

	Description [Picks one on-set cube randomly from the given DD. The
	cube is written into an array of characters.  The array must have at
	least as many entries as there are variables. Returns 1 if
	successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_bddPickOneMinterm]

*****************************************************************************
*/
func Cudd_bddPickOneCube(dd *DdManager, f *DdNode, cube []byte) int32 {
	if dd == nil || f == nil || len(cube) < len(dd.Vars) || f == dd.LogicZero {
		return 0
	}
	for i := range dd.Vars {
		cube[i] = 2
	}
	for !Cudd_IsConstant(Cudd_Regular(f)) {
		regular := Cudd_Regular(f)
		thenEdge := Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(f))
		elseEdge := Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(f))
		index := int(regular.Index)
		if index >= len(cube) {
			return 0
		}
		switch {
		case thenEdge == dd.LogicZero:
			cube[index] = 0
			f = elseEdge
		case elseEdge == dd.LogicZero:
			cube[index] = 1
			f = thenEdge
		default:
			if Cudd_Random()&0x2000 != 0 {
				cube[index] = 1
				f = thenEdge
			} else {
				cube[index] = 0
				f = elseEdge
			}
		}
	}
	return 1
}

func Cudd_bddPickOneMinterm(dd *DdManager, f *DdNode, vars []*DdNode, n int32) *DdNode {
	if dd == nil || f == nil || n < 0 || int(n) > len(vars) {
		return nil
	}
	cube := make([]byte, len(dd.Vars))
	if Cudd_bddPickOneCube(dd, f, cube) == 0 {
		return nil
	}
	result := dd.One
	for i := int32(0); i < n; i++ {
		variable := vars[i]
		if variable == nil {
			return nil
		}
		index := int(Cudd_NodeReadIndex(variable))
		if index >= len(cube) {
			return nil
		}
		phase := cube[index]
		if phase == 2 {
			if Cudd_Random()&0x20 != 0 {
				phase = 1
			} else {
				phase = 0
			}
		}
		literal := variable
		if phase == 0 {
			literal = core.Cudd_Not(variable)
		}
		result = core.Cudd_bddAnd(dd, result, literal)
		if result == nil {
			return nil
		}
	}
	return result
}

/*
*Function********************************************************************

	Synopsis    [Picks k on-set minterms evenly distributed from given DD.]

	Description [Picks k on-set minterms evenly distributed from given DD.
	The minterms are in terms of <code>vars</code>. The array
	<code>vars</code> should contain at least all variables in the
	support of <code>f</code>; if this condition is not met the minterms
	built by this procedure may not be contained in
	<code>f</code>. Builds an array of BDDs for the minterms and returns a
	pointer to it if successful; NULL otherwise. There are three reasons
	why the procedure may fail:
	<ul>
	<li> It may run out of memory;
	<li> the function <code>f</code> may be the constant 0;
	<li> the minterms may not be contained in <code>f</code>.
	</ul>]

	SideEffects [None]

	SeeAlso     [Cudd_bddPickOneMinterm Cudd_bddPickOneCube]

*****************************************************************************
*/
func Cudd_bddPickArbitraryMinterms(dd *DdManager, f *DdNode, vars []*DdNode, n, k int32) []*DdNode {
	if dd == nil || f == nil || n < 0 || k < 0 || int(n) > len(vars) {
		return nil
	}
	if k == 0 {
		return []*DdNode{}
	}
	for i := int32(0); i < n; i++ {
		if vars[i] == nil {
			return nil
		}
	}
	if float64(k) > Cudd_CountMinterm(dd, f, n) {
		return nil
	}
	indices := make([]int, n)
	for i := range indices {
		indices[i] = int(Cudd_NodeReadIndex(vars[i]))
	}
	build := func(values []bool) *DdNode {
		result := dd.One
		for i, variable := range vars[:n] {
			literal := variable
			if !values[i] {
				literal = core.Cudd_Not(literal)
			}
			result = core.Cudd_bddAnd(dd, result, literal)
			if result == nil {
				return nil
			}
		}
		return result
	}
	picked := make([]*DdNode, 0, k)
	seen := make(map[*DdNode]struct{}, k)
	assignment := make([]bool, len(dd.Vars))
	if n <= 20 {
		limit := uint64(1) << uint(n)
		for mask := uint64(0); mask < limit && len(picked) < int(k); mask++ {
			clear(assignment)
			values := make([]bool, n)
			for i, index := range indices {
				values[i] = mask&(uint64(1)<<uint(i)) != 0
				if index >= 0 && index < len(assignment) {
					assignment[index] = values[i]
				}
			}
			satisfied, err := dd.Evaluate(f, assignment)
			if err != nil || !satisfied {
				continue
			}
			cube := build(values)
			if cube != nil {
				if _, duplicate := seen[cube]; !duplicate {
					seen[cube] = struct{}{}
					picked = append(picked, cube)
				}
			}
		}
		if len(picked) == int(k) {
			return picked
		}
		return nil
	}
	for attempt := 0; attempt < int(k)*128 && len(picked) < int(k); attempt++ {
		cube := Cudd_bddPickOneMinterm(dd, f, vars, n)
		if cube == nil {
			return nil
		}
		if _, duplicate := seen[cube]; !duplicate {
			seen[cube] = struct{}{}
			picked = append(picked, cube)
		}
	}
	if len(picked) != int(k) {
		return nil
	}
	return picked
}

/*
*Function********************************************************************

	Synopsis    [Extracts a subset from a BDD.]

	Description [Extracts a subset from a BDD in the following procedure.
	1. Compute the weight for each mask variable by counting the number of
	   minterms for both positive and negative cofactors of the BDD with
	   respect to each mask variable. (weight = #positive - #negative)
	2. Find a representative cube of the BDD by using the weight. From the
	   top variable of the BDD, for each variable, if the weight is greater
	   than 0.0, choose THEN branch, othereise ELSE branch, until meeting
	   the constant 1.
	3. Quantify out the variables not in maskVars from the representative
	   cube and if a variable in maskVars is don't care, replace the
	   variable with a constant(1 or 0) depending on the weight.
	4. Make a subset of the BDD by multiplying with the modified cube.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_SubsetWithMaskVars(dd *DdManager, f *DdNode, vars []*DdNode, nvars int32, maskVars []*DdNode, mvars int32) *DdNode {
	if dd == nil || f == nil || nvars < 0 || mvars < 0 || int(nvars) > len(vars) || int(mvars) > len(maskVars) {
		return nil
	}
	assignment := make([]byte, len(dd.Vars))
	if Cudd_bddPickOneCube(dd, f, assignment) == 0 {
		return nil
	}
	mask := make(map[int]bool, mvars)
	for _, variable := range maskVars[:mvars] {
		if variable == nil {
			return nil
		}
		mask[int(Cudd_NodeReadIndex(variable))] = true
	}
	cube := dd.One
	for _, variable := range vars[:nvars] {
		index := int(Cudd_NodeReadIndex(variable))
		if !mask[index] {
			continue
		}
		phase := assignment[index]
		if phase == 2 {
			if Cudd_Random()&0x20 != 0 {
				phase = 1
			} else {
				phase = 0
			}
		}
		literal := variable
		if phase == 0 {
			literal = core.Cudd_Not(variable)
		}
		cube = core.Cudd_bddAnd(dd, cube, literal)
		if cube == nil {
			return nil
		}
	}
	return core.Cudd_bddAnd(dd, f, cube)
}

/*
*Function********************************************************************

	Synopsis    [Finds the first cube of a decision diagram.]

	Description [Defines an iterator on the onset of a decision diagram
	and finds its first cube. Returns a generator that contains the
	information necessary to continue the enumeration if successful; NULL
	otherwise.<p>
	A cube is represented as an array of literals, which are integers in
	{0, 1, 2}; 0 represents a complemented literal, 1 represents an
	uncomplemented literal, and 2 stands for don't care. The enumeration
	produces a disjoint cover of the function associated with the diagram.
	The size of the array equals the number of variables in the manager at
	the time Cudd_FirstCube is called.<p>
	For each cube, a value is also returned. This value is always 1 for a
	BDD, while it may be different from 1 for an ADD.
	For BDDs, the offset is the set of cubes whose value is the logical zero.
	For ADDs, the offset is the set of cubes whose value is the
	background value. The cubes of the offset are not enumerated.]

	SideEffects [The first cube and its value are returned as side effects.]

	SeeAlso     [Cudd_ForeachCube Cudd_NextCube Cudd_GenFree Cudd_IsGenEmpty
	Cudd_FirstNode]

*****************************************************************************
*/
// cuddCubePush mirrors CUDD's fixed nvars+1 stack: a backtracked slot is
// overwritten on the next descent rather than appended after the slice tail.
func cuddCubePush(gen *DdGen, edge *DdNode) bool {
	if gen.stack.sp < 0 || int(gen.stack.sp) >= len(gen.stack.stack) {
		gen.status = CUDD_GEN_EMPTY
		return false
	}
	gen.stack.stack[gen.stack.sp] = edge
	gen.stack.sp++
	return true
}

// cuddNextCube implements both the initial descent and Cudd_NextCube's
// backtrack-then-descent sequence. The order and pointer identity checks match
// cuddUtil.c; in particular, zero terminals are skipped before continuing.
func cuddNextCube(gen *DdGen, first bool) bool {
	if gen == nil || gen.manager == nil || gen.stack.sp <= 0 {
		return false
	}
	dd := gen.manager
	if !first {
		for {
			if gen.stack.sp == 1 {
				gen.status = CUDD_GEN_EMPTY
				gen.stack.sp--
				return false
			}
			top := gen.stack.stack[gen.stack.sp-1]
			prev := gen.stack.stack[gen.stack.sp-2]
			preg := Cudd_Regular(prev)
			if preg == nil {
				gen.status = CUDD_GEN_EMPTY
				return false
			}
			next := Cudd_NotCond(Cudd_T(preg), Cudd_IsComplement(prev))
			if next != top {
				index := int(preg.Index)
				if index < 0 || index >= len(gen.data.cubes.cube) {
					gen.status = CUDD_GEN_EMPTY
					return false
				}
				gen.data.cubes.cube[index] = 1
				gen.stack.stack[gen.stack.sp-1] = next
				break
			}
			index := int(preg.Index)
			if index < 0 || index >= len(gen.data.cubes.cube) {
				gen.status = CUDD_GEN_EMPTY
				return false
			}
			gen.data.cubes.cube[index] = 2
			gen.stack.sp--
		}
	}

	for gen.stack.sp > 0 {
		top := gen.stack.stack[gen.stack.sp-1]
		treg := Cudd_Regular(top)
		if treg == nil {
			gen.status = CUDD_GEN_EMPTY
			return false
		}
		if !Cudd_IsConstant(treg) {
			index := int(treg.Index)
			if index < 0 || index >= len(gen.data.cubes.cube) {
				gen.status = CUDD_GEN_EMPTY
				return false
			}
			gen.data.cubes.cube[index] = 0
			next := Cudd_NotCond(Cudd_E(treg), Cudd_IsComplement(top))
			if !cuddCubePush(gen, next) {
				return false
			}
			continue
		}
		if top == Cudd_Not(dd.One) || top == dd.Background {
			// Backtrack until a parent whose then edge is not the edge just
			// explored; that edge is the next disjoint cube's branch.
			for {
				if gen.stack.sp == 1 {
					gen.status = CUDD_GEN_EMPTY
					gen.stack.sp--
					return false
				}
				prev := gen.stack.stack[gen.stack.sp-2]
				preg := Cudd_Regular(prev)
				if preg == nil {
					gen.status = CUDD_GEN_EMPTY
					return false
				}
				next := Cudd_NotCond(Cudd_T(preg), Cudd_IsComplement(prev))
				index := int(preg.Index)
				if index < 0 || index >= len(gen.data.cubes.cube) {
					gen.status = CUDD_GEN_EMPTY
					return false
				}
				if next != top {
					gen.data.cubes.cube[index] = 1
					gen.stack.stack[gen.stack.sp-1] = next
					break
				}
				gen.data.cubes.cube[index] = 2
				gen.stack.sp--
				top = gen.stack.stack[gen.stack.sp-1]
			}
			continue
		}
		gen.status = CUDD_GEN_NONEMPTY
		gen.data.cubes.value = treg.Value
		return true
	}
	gen.status = CUDD_GEN_EMPTY
	return false
}

func Cudd_FirstCube(dd *DdManager, f *DdNode, cube **int32, value *CUDD_VALUE_TYPE) *DdGen {
	if dd == nil || f == nil || cube == nil || value == nil {
		return nil
	}
	gen := &DdGen{manager: dd, genType: CUDD_GEN_CUBES, status: CUDD_GEN_EMPTY}
	gen.data.cubes.cube = make([]int32, len(dd.Vars))
	for i := range gen.data.cubes.cube {
		gen.data.cubes.cube[i] = 2
	}
	gen.stack.stack = make([]*DdNode, len(dd.Vars)+1)
	gen.stack.stack[0] = f
	gen.stack.sp = 1
	cuddNextCube(gen, true)
	// The C routine assigns both out parameters even when the onset is empty.
	if len(gen.data.cubes.cube) > 0 {
		*cube = &gen.data.cubes.cube[0]
	} else {
		*cube = nil
	}
	*value = gen.data.cubes.value
	return gen
}

func Cudd_NextCube(gen *DdGen, cube **int32, value *CUDD_VALUE_TYPE) int32 {
	if gen == nil || cube == nil || value == nil || gen.genType != CUDD_GEN_CUBES {
		return 0
	}
	if !cuddNextCube(gen, false) {
		return 0
	}
	if len(gen.data.cubes.cube) > 0 {
		*cube = &gen.data.cubes.cube[0]
	} else {
		*cube = nil
	}
	*value = gen.data.cubes.value
	return 1
}

/*
*Function********************************************************************

	Synopsis    [Finds the first prime of a Boolean function.]

	Description [Defines an iterator on a pair of BDDs describing a
	(possibly incompletely specified) Boolean functions and finds the
	first cube of a cover of the function.  Returns a generator
	that contains the information necessary to continue the enumeration
	if successful; NULL otherwise.<p>

	The two argument BDDs are the lower and upper bounds of an interval.
	It is a mistake to call this function with a lower bound that is not
	less than or equal to the upper bound.<p>

	A cube is represented as an array of literals, which are integers in
	{0, 1, 2}; 0 represents a complemented literal, 1 represents an
	uncomplemented literal, and 2 stands for don't care. The enumeration
	produces a prime and irredundant cover of the function associated
	with the two BDDs.  The size of the array equals the number of
	variables in the manager at the time Cudd_FirstCube is called.<p>

	This iterator can only be used on BDDs.]

	SideEffects [The first cube is returned as side effect.]

	SeeAlso     [Cudd_ForeachPrime Cudd_NextPrime Cudd_GenFree Cudd_IsGenEmpty
	Cudd_FirstCube Cudd_FirstNode]

*****************************************************************************
*/
func cuddPrimeCubeValues(dd *DdManager, cube *DdNode) []int32 {
	values := make([]int32, len(dd.Vars))
	for i := range values {
		values[i] = 2
	}
	for !Cudd_IsConstant(Cudd_Regular(cube)) {
		regular := Cudd_Regular(cube)
		index := int(regular.Index)
		if index < 0 || index >= len(values) {
			return values
		}
		thenEdge := Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(cube))
		elseEdge := Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(cube))
		if thenEdge == dd.LogicZero {
			values[index] = 0
			cube = elseEdge
		} else if elseEdge == dd.LogicZero {
			values[index] = 1
			cube = thenEdge
		} else {
			break
		}
	}
	return values
}

func cuddWritePrimeCube(dd *DdManager, prime *DdNode, out **int32, cube []int32) {
	copy(cube, cuddPrimeCubeValues(dd, prime))
	if len(cube) == 0 {
		*out = nil
	} else {
		*out = &cube[0]
	}
}

func Cudd_FirstPrime(dd *DdManager, l *DdNode, u *DdNode, cube **int32) *DdGen {
	if dd == nil || l == nil || u == nil || cube == nil {
		return nil
	}
	gen := &DdGen{manager: dd, genType: CUDD_GEN_PRIMES, status: CUDD_GEN_EMPTY}
	gen.data.primes.ub = u
	seen := make(map[*DdNode]struct{})
	addPrime := func(minterm *DdNode) {
		prime := Cudd_bddMakePrime(dd, minterm, u)
		if prime == nil {
			return
		}
		if _, ok := seen[prime]; !ok {
			seen[prime] = struct{}{}
			gen.stack.stack = append(gen.stack.stack, prime)
		}
	}
	if len(dd.Vars) <= 20 {
		assignment := make([]bool, len(dd.Vars))
		limit := uint64(1) << uint(len(dd.Vars))
		for mask := uint64(0); mask < limit; mask++ {
			for i := range assignment {
				assignment[i] = mask&(uint64(1)<<uint(i)) != 0
			}
			satisfies, err := dd.Evaluate(l, assignment)
			if err != nil || !satisfies {
				continue
			}
			minterm := dd.One
			for index, value := range assignment {
				literal := core.Cudd_bddIthVar(dd, index)
				if !value {
					literal = core.Cudd_Not(literal)
				}
				minterm = core.Cudd_bddAnd(dd, minterm, literal)
				if minterm == nil {
					return nil
				}
			}
			addPrime(minterm)
		}
	} else {
		var length int32
		candidate := Cudd_LargestCube(dd, l, &length)
		addPrime(candidate)
	}
	if len(gen.stack.stack) == 0 {
		return gen
	}
	gen.stack.sp = 1
	gen.status = CUDD_GEN_NONEMPTY
	gen.data.primes.cube = cuddPrimeCubeValues(dd, gen.stack.stack[0])
	if len(gen.data.primes.cube) == 0 {
		*cube = nil
	} else {
		*cube = &gen.data.primes.cube[0]
	}
	return gen
}

func Cudd_NextPrime(gen *DdGen, cube **int32) int32 {
	if gen == nil || cube == nil || gen.genType != CUDD_GEN_PRIMES || gen.stack.sp >= int32(len(gen.stack.stack)) {
		if gen != nil {
			gen.status = CUDD_GEN_EMPTY
		}
		return 0
	}
	prime := gen.stack.stack[gen.stack.sp]
	gen.stack.sp++
	gen.data.primes.cube = cuddPrimeCubeValues(gen.manager, prime)
	if len(gen.data.primes.cube) == 0 {
		*cube = nil
	} else {
		*cube = &gen.data.primes.cube[0]
	}
	gen.status = CUDD_GEN_NONEMPTY
	return 1
}

/*
*Function********************************************************************

	Synopsis    [Builds a cube of BDD variables from an array of indices.]

	Description [Builds a cube of BDD variables from an array of indices.
	Returns a pointer to the result if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_bddComputeCube Cudd_CubeArrayToBdd]

*****************************************************************************
*/
func Cudd_IndicesToCube(dd *DdManager, array *int32, n int32) *DdNode {
	if dd == nil || n < 0 || (n > 0 && array == nil) {
		return nil
	}
	indices := unsafe.Slice(array, int(n))
	cube := dd.One
	for i := n - 1; i >= 0; i-- {
		index := int(indices[i])
		if index < 0 || index >= len(dd.Vars) {
			return nil
		}
		cube = core.Cudd_bddAnd(dd, cube, core.Cudd_bddIthVar(dd, index))
		if cube == nil {
			return nil
		}
	}
	return cube

} /* end of Cudd_IndicesToCube */

func Cudd_bddComputeCube(dd *DdManager, vars []*DdNode, phase []int32, n int32) *DdNode {
	if dd == nil || n < 0 || int(n) > len(vars) || int(n) > len(phase) {
		return nil
	}
	result := dd.One
	for i := int32(0); i < n; i++ {
		literal := vars[i]
		if literal == nil {
			return nil
		}
		if phase[i] == 0 {
			literal = core.Cudd_Not(literal)
		}
		result = core.Cudd_bddAnd(dd, result, literal)
		if result == nil {
			return nil
		}
	}
	return result
}

func Cudd_addComputeCube(dd *DdManager, vars []*DdNode, phase []int32, n int32) *DdNode {
	cube := Cudd_bddComputeCube(dd, vars, phase, n)
	if cube == nil {
		return nil
	}
	return core.Cudd_BddToAdd(dd, cube)
}

func Cudd_CubeArrayToBdd(dd *DdManager, array []int32) *DdNode {
	if dd == nil || len(array) < len(dd.Vars) {
		return nil
	}
	result := dd.One
	for index, phase := range array[:len(dd.Vars)] {
		if phase == 2 {
			continue
		}
		literal := core.Cudd_bddIthVar(dd, index)
		if phase == 0 {
			literal = core.Cudd_Not(literal)
		} else if phase != 1 {
			return nil
		}
		result = core.Cudd_bddAnd(dd, result, literal)
		if result == nil {
			return nil
		}
	}
	return result
}

func Cudd_BddToCubeArray(dd *DdManager, cube *DdNode, array []int32) int32 {
	if dd == nil || cube == nil || len(array) < len(dd.Vars) {
		return 0
	}
	for i := range dd.Vars {
		array[i] = 2
	}
	for !Cudd_IsConstant(Cudd_Regular(cube)) {
		regular := Cudd_Regular(cube)
		index := int(regular.Index)
		if index >= len(dd.Vars) {
			return 0
		}
		thenEdge := Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(cube))
		elseEdge := Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(cube))
		if thenEdge == dd.LogicZero {
			array[index] = 0
			cube = elseEdge
		} else if elseEdge == dd.LogicZero {
			array[index] = 1
			cube = thenEdge
		} else {
			return 0
		}
	}
	if cube == dd.LogicZero {
		return 0
	}
	return 1
}

func Cudd_FirstNode(dd *DdManager, f *DdNode, node **DdNode) *DdGen {
	if dd == nil || f == nil || node == nil {
		return nil
	}
	gen := &DdGen{manager: dd, genType: CUDD_GEN_NODES, status: CUDD_GEN_EMPTY}
	gen.stack.stack = cuddNodeArray(f, &gen.data.nodes.size)
	if len(gen.stack.stack) == 0 {
		return gen
	}
	gen.stack.sp = 1
	gen.status = CUDD_GEN_NONEMPTY
	*node = gen.stack.stack[0]
	return gen
}

func Cudd_NextNode(gen *DdGen, node **DdNode) int32 {
	if gen == nil || node == nil || gen.genType != CUDD_GEN_NODES || gen.stack.sp >= int32(len(gen.stack.stack)) {
		if gen != nil {
			gen.status = CUDD_GEN_EMPTY
		}
		return 0
	}
	*node = gen.stack.stack[gen.stack.sp]
	gen.stack.sp++
	return 1
}

func Cudd_GenFree(gen *DdGen) int32 {
	if gen == nil {
		return 0
	}
	gen.stack.stack = nil
	gen.data.cubes.cube = nil
	gen.data.primes.cube = nil
	gen.status = CUDD_GEN_EMPTY
	return 1
}

func Cudd_IsGenEmpty(gen *DdGen) int32 {
	if gen == nil || gen.status == CUDD_GEN_EMPTY {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Prints the package version number.]

	Description []

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_PrintVersion(fp *os.File) {
	_, _ = fmt.Fprintf(fp, "%s\n", CUDD_VERSION)

} /* end of Cudd_PrintVersion */

/*
*Function********************************************************************

	Synopsis    [Computes the average distance between adjacent nodes.]

	Description [Computes the average distance between adjacent nodes in
	the manager. Adjacent nodes are node pairs such that the second node
	is the then child, else child, or next node in the collision list.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
// Cudd_AverageDistance is an architectural fidelity gap. The C routine measures
// byte distances between nodes in unique-table buckets and collision chains.
// Go's manager stores nodes in maps without stable slot addresses or those
// chains, so this implementation can only approximate the report from graph
// edges; it must not be used as a C-equivalent statistic.
func Cudd_AverageDistance(dd *DdManager) float64 {
	if dd == nil || len(dd.Vars) == 0 {
		return 0
	}
	seen := make(map[*DdNode]struct{})
	stack := append([]*DdNode(nil), dd.Vars...)
	var distance float64
	var measured int64
	for len(stack) > 0 {
		last := len(stack) - 1
		node := Cudd_Regular(stack[last])
		stack = stack[:last]
		if Cudd_IsConstant(node) {
			continue
		}
		if _, ok := seen[node]; ok {
			continue
		}
		seen[node] = struct{}{}
		for _, child := range []*DdNode{Cudd_T(node), Cudd_Regular(Cudd_E(node))} {
			a, b := uintptr(unsafe.Pointer(node)), uintptr(unsafe.Pointer(child))
			if a > b {
				distance += float64(a - b)
			} else {
				distance += float64(b - a)
			}
			measured++
			stack = append(stack, child)
		}
	}
	if measured == 0 {
		return 0
	}
	return distance / float64(measured)

} /* end of Cudd_AverageDistance */

/*
*Function********************************************************************

	Synopsis    [Portable random number generator.]

	Description [Portable number generator based on ran2 from "Numerical
	Recipes in C." It is a long period (> 2 * 10^18) random number generator
	of L'Ecuyer with Bays-Durham shuffle. Returns a long integer uniformly
	distributed between 0 and 2147483561 (inclusive of the endpoint values).
	The random generator can be explicitly initialized by calling
	Cudd_Srandom. If no explicit initialization is performed, then the
	seed 1 is assumed.]

	SideEffects [None]

	SeeAlso     [Cudd_Srandom]

*****************************************************************************
*/
func Cudd_Random() int64 {
	var i int
	var w int64

	/* cuddRand == 0 if the geneartor has not been initialized yet. */
	if cuddRand == 0 {
		Cudd_Srandom(1)
	}

	/* Compute cuddRand = (cuddRand * LEQA1) % MODULUS1 avoiding
	 ** overflows by Schrage's method.
	 */
	w = cuddRand / LEQQ1
	cuddRand = LEQA1*(cuddRand-w*LEQQ1) - w*LEQR1
	if cuddRand < 0 {
		cuddRand += MODULUS1
	}

	/* Compute cuddRand2 = (cuddRand2 * LEQA2) % MODULUS2 avoiding
	 ** overflows by Schrage's method.
	 */
	w = cuddRand2 / LEQQ2
	cuddRand2 = LEQA2*(cuddRand2-w*LEQQ2) - w*LEQR2
	if cuddRand2 < 0 {
		cuddRand2 += MODULUS2
	}

	/* cuddRand is shuffled with the Bays-Durham algorithm.
	 ** shuffleSelect and cuddRand2 are combined to generate the output.
	 */

	/* Pick one element from the shuffle table; "i" will be in the range
	 ** from 0 to STAB_SIZE-1.
	 */
	i = int(shuffleSelect / STAB_DIV)
	/* Mix the element of the shuffle table with the current iterate of
	 ** the second sub-generator, and replace the chosen element of the
	 ** shuffle table with the current iterate of the first sub-generator.
	 */
	shuffleSelect = shuffleTable[i] - cuddRand2
	shuffleTable[i] = cuddRand
	if shuffleSelect < 1 {
		shuffleSelect += MODULUS1 - 1
	}
	/* Since shuffleSelect != 0, and we want to be able to return 0,
	 ** here we subtract 1 before returning.
	 */
	return (shuffleSelect - 1)

} /* end of Cudd_Random */

/*
*Function********************************************************************

	Synopsis    [Initializer for the portable random number generator.]

	Description [Initializer for the portable number generator based on
	ran2 in "Numerical Recipes in C." The input is the seed for the
	generator. If it is negative, its absolute value is taken as seed.
	If it is 0, then 1 is taken as seed. The initialized sets up the two
	recurrences used to generate a long-period stream, and sets up the
	shuffle table.]

	SideEffects [None]

	SeeAlso     [Cudd_Random]

*****************************************************************************
*/
func Cudd_Srandom(seed int64) {
	var i int

	if seed < 0 {
		cuddRand = -seed
	} else if seed == 0 {
		cuddRand = 1
	} else {
		cuddRand = seed
	}
	cuddRand2 = cuddRand
	/* Load the shuffle table (after 11 warm-ups). */
	for i = 0; i < STAB_SIZE+11; i++ {
		var w int64
		w = cuddRand / LEQQ1
		cuddRand = LEQA1*(cuddRand-w*LEQQ1) - w*LEQR1
		if cuddRand < 0 {
			cuddRand += MODULUS1
		}
		shuffleTable[i%STAB_SIZE] = cuddRand
	}
	shuffleSelect = shuffleTable[1%STAB_SIZE]

} /* end of Cudd_Srandom */

// Cudd_Density returns the minterm count divided by the DAG node count.
// A zero nvars argument uses the manager's current variable count.
func Cudd_Density(dd *DdManager, f *DdNode, nvars int32) float64 {
	if dd == nil {
		return float64(CUDD_OUT_OF_MEM)
	}
	if nvars == 0 {
		nvars = int32(len(dd.Vars))
	}
	minterms := Cudd_CountMinterm(dd, f, nvars)
	if minterms == float64(CUDD_OUT_OF_MEM) {
		return minterms
	}
	nodes := Cudd_DagSize(f)
	return minterms / float64(nodes)
}

// Cudd_OutOfMem reports a failed discretionary allocation.
func Cudd_OutOfMem(size int64) {
	_, _ = fmt.Fprintf(os.Stderr, "\nunable to allocate %d bytes\n", size)
}

/**Function********************************************************************

  Synopsis    [Computes the density of a BDD or ADD.]

  Description [Computes the density of a BDD or ADD. The density is
  the ratio of the number of minterms to the number of nodes. If 0 is
  passed as number of variables, the number of variables existing in
  the manager is used. Returns the density if successful; (double)
  CUDD_OUT_OF_MEM otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_CountMinterm Cudd_DagSize]

******************************************************************************/

/* Internal graph helpers. Go's garbage collector replaces C's flag bits
 * and temporary storage tables, so these routines use local maps and slices. */

func cuddP(dd *DdManager, f *DdNode) int32 {
	if dd == nil || f == nil {
		return 0
	}
	visited := make(map[*DdNode]struct{})
	retval := dp2(dd, f, visited)
	_, _ = fmt.Fprintln(core.Cudd_Output(dd))
	return retval
}

// cuddFlushOutput mirrors fflush(dd->out) when the configured Go writer
// exposes a flush operation. Ordinary io.Writer values need no extra action.
func cuddFlushOutput(out interface{}) {
	if flusher, ok := out.(interface{ Flush() error }); ok {
		_ = flusher.Flush()
	} else if flusher, ok := out.(interface{ Flush() }); ok {
		flusher.Flush()
	}
}

func cuddCollectNodes(f *DdNode, visited map[*DdNode]struct{}) int32 {
	if f == nil {
		return 0
	}
	stack := []*DdNode{f}
	for len(stack) > 0 {
		last := len(stack) - 1
		node := Cudd_Regular(stack[last])
		stack = stack[:last]
		if _, exists := visited[node]; exists {
			continue
		}
		visited[node] = struct{}{}
		if !Cudd_IsConstant(node) {
			stack = append(stack, Cudd_T(node), Cudd_E(node))
		}
	}
	return 1
}

func cuddNodeArray(f *DdNode, n *int32) []*DdNode {
	if f == nil || n == nil {
		return nil
	}
	visited := make(map[*DdNode]struct{})
	result := make([]*DdNode, 0)
	var visit func(*DdNode)
	visit = func(edge *DdNode) {
		node := Cudd_Regular(edge)
		if _, exists := visited[node]; exists {
			return
		}
		visited[node] = struct{}{}
		if !Cudd_IsConstant(node) {
			visit(Cudd_T(node))
			visit(Cudd_E(node))
		}
		result = append(result, node)
	}
	visit(f)
	*n = int32(len(result))
	return result
}

func dp2(dd *DdManager, f *DdNode, visited map[*DdNode]struct{}) int32 {
	if f == nil {
		return 0
	}
	visitedNode := Cudd_Regular(f)
	out := core.Cudd_Output(dd)
	if Cudd_IsConstant(visitedNode) {
		_, _ = fmt.Fprintf(out, "ID = %c0x%x\tvalue = %-9g\n", cuddBang(f), visitedNode.ID, visitedNode.Value)
		return 1
	}
	if _, exists := visited[visitedNode]; exists {
		return 1
	}
	visited[visitedNode] = struct{}{}
	_, _ = fmt.Fprintf(out, "ID = %c0x%x\tindex = %d\tr = %d\t", cuddBang(f), visitedNode.ID, visitedNode.Index, visitedNode.Ref)
	t := Cudd_T(visitedNode)
	if Cudd_IsConstant(Cudd_Regular(t)) {
		_, _ = fmt.Fprintf(out, "T = %-9g\t", Cudd_Regular(t).Value)
	} else {
		_, _ = fmt.Fprintf(out, "T = 0x%x\t", Cudd_Regular(t).ID)
	}
	e := Cudd_E(visitedNode)
	if Cudd_IsConstant(Cudd_Regular(e)) {
		_, _ = fmt.Fprintf(out, "E = %c%-9g\n", cuddBang(e), Cudd_Regular(e).Value)
	} else {
		_, _ = fmt.Fprintf(out, "E = %c0x%x\n", cuddBang(e), Cudd_Regular(e).ID)
	}
	if !Cudd_IsConstant(Cudd_Regular(e)) {
		if dp2(dd, Cudd_Regular(e), visited) == 0 {
			return 0
		}
	}
	if !Cudd_IsConstant(Cudd_Regular(t)) {
		if dp2(dd, Cudd_Regular(t), visited) == 0 {
			return 0
		}
	}
	return 1
}

func cuddBang(edge *DdNode) byte {
	if Cudd_IsComplement(edge) {
		return '!'
	}
	return ' '
}

func ddPrintMintermAux(dd *DdManager, node *DdNode, list []int32) {
	if dd == nil || node == nil {
		return
	}
	out := core.Cudd_Output(dd)
	if Cudd_IsConstant(Cudd_Regular(node)) {
		if node == dd.LogicZero || node == dd.Background {
			return
		}
		for _, value := range list {
			if value == 0 {
				_, _ = fmt.Fprint(out, "0")
			} else if value == 1 {
				_, _ = fmt.Fprint(out, "1")
			} else {
				_, _ = fmt.Fprint(out, "-")
			}
		}
		_, _ = fmt.Fprintf(out, " % g\n", Cudd_Regular(node).Value)
		return
	}
	regular := Cudd_Regular(node)
	index := int(regular.Index)
	if index < 0 || index >= len(list) {
		return
	}
	list[index] = 0
	ddPrintMintermAux(dd, Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(node)), list)
	list[index] = 1
	ddPrintMintermAux(dd, Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(node)), list)
	list[index] = 2
}

func ddDagInt(n *DdNode) int32 {
	return int32(Cudd_DagSize(n))
}

func cuddNodeArrayRecur(f *DdNode, table []*DdNode, index int32) int32 {
	if f == nil {
		return index
	}
	items := cuddNodeArray(f, new(int32))
	for _, node := range items {
		if int(index) >= len(table) {
			return index
		}
		table[index] = node
		index++
	}
	return index
}

func cuddEstimateCofactor(dd *DdManager, node *DdNode, i, phase int32) int32 {
	return Cudd_EstimateCofactor(dd, node, i, phase)
}

func cuddUniqueLookup(unique *DdManager, index int32, T, E *DdNode) *DdNode {
	if unique == nil || index < 0 || int(index) >= len(unique.Vars) {
		return nil
	}
	return core.Cudd_bddIte(unique, core.Cudd_bddIthVar(unique, int(index)), T, E)
}

func cuddEstimateCofactorSimple(node *DdNode, i int32) int32 {
	return Cudd_EstimateCofactorSimple(node, i)
}

func ddCountMintermAux(node *DdNode, max float64) float64 {
	if node == nil || Cudd_IsConstant(Cudd_Regular(node)) {
		if node != nil && !Cudd_IsComplement(node) && Cudd_Regular(node).Value != 0 {
			return max
		}
		return 0
	}
	regular := Cudd_Regular(node)
	thenCount := ddCountMintermAux(Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(node)), max/2)
	elseCount := ddCountMintermAux(Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(node)), max/2)
	return thenCount + elseCount
}

func ddCountPathAux(node *DdNode, visited map[*DdNode]float64) float64 {
	if node == nil {
		return 0
	}
	regular := Cudd_Regular(node)
	if value, ok := visited[regular]; ok {
		return value
	}
	if Cudd_IsConstant(regular) {
		visited[regular] = 1
		return 1
	}
	value := ddCountPathAux(Cudd_T(regular), visited) + ddCountPathAux(Cudd_E(regular), visited)
	visited[regular] = value
	return value
}

func ddEpdCountMintermAux(node *DdNode, max *EpDouble, epd *EpDouble) int32 {
	if max == nil || epd == nil {
		return CUDD_OUT_OF_MEM
	}
	value := ddCountMintermAux(node, max.Value*math.Pow(2, float64(max.Exponent)))
	epdpkg.EpdConvert(value, epd)
	return 0
}

func ddCountPathsToNonZero(node *DdNode, visited map[*DdNode]float64) float64 {
	if node == nil {
		return 0
	}
	if value, ok := visited[node]; ok {
		return value
	}
	regular := Cudd_Regular(node)
	if Cudd_IsConstant(regular) {
		if !Cudd_IsComplement(node) && regular.Value != 0 {
			visited[node] = 1
			return 1
		}
		visited[node] = 0
		return 0
	}
	value := ddCountPathsToNonZero(Cudd_NotCond(Cudd_T(regular), Cudd_IsComplement(node)), visited) +
		ddCountPathsToNonZero(Cudd_NotCond(Cudd_E(regular), Cudd_IsComplement(node)), visited)
	visited[node] = value
	return value
}

func ddSupportStep(f *DdNode, support []int32) {
	visited := make(map[*DdNode]struct{})
	var visit func(*DdNode)
	visit = func(edge *DdNode) {
		node := Cudd_Regular(edge)
		if Cudd_IsConstant(node) {
			return
		}
		if _, ok := visited[node]; ok {
			return
		}
		visited[node] = struct{}{}
		if int(node.Index) < len(support) {
			support[node.Index] = 1
		}
		visit(Cudd_T(node))
		visit(Cudd_E(node))
	}
	visit(f)
}

func ddClearFlag(*DdNode) {}

func ddLeavesInt(node *DdNode) int32 { return Cudd_CountLeaves(node) }

func ddPickArbitraryMinterms(dd *DdManager, node *DdNode, vars []*DdNode, n, k int32) []*DdNode {
	return Cudd_bddPickArbitraryMinterms(dd, node, vars, n, k)
}

func ddPickRepresentativeCube(dd *DdManager, node *DdNode, cube []byte) int32 {
	return Cudd_bddPickOneCube(dd, node, cube)
}

//ABC_NAMESPACE_IMPL_END
