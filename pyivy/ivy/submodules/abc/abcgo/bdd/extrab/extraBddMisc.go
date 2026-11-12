/**CFile****************************************************************

  FileName    [extraBddMisc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [DD-based utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraBddMisc.c,v 1.4 2005/10/04 00:19:54 alanmi Exp $]

***********************************************************************/

package extrab

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab/cascade"
)

func Extra_TransferPermute(source, destination *DdManager, f *DdNode, permutation []int) *DdNode {
	if !extraValidBDD(source, f) || destination == nil {
		return nil
	}
	return source.Transfer(f, destination, permutation)
}
func Extra_TransferLevelByLevel(source, destination *DdManager, f *DdNode) *DdNode {
	if source == nil || destination == nil {
		return nil
	}
	mapping := make([]int, len(source.Vars))
	for i := range mapping {
		mapping[i] = -1
	}
	for level := 0; level < min(len(source.Vars), len(destination.Vars)); level++ {
		mapping[source.InvPerm[level]] = destination.InvPerm[level]
	}
	return Extra_TransferPermute(source, destination, f, mapping)
}
func Extra_bddRemapUp(dd *DdManager, f *DdNode) *DdNode {
	if !extraValidBDD(dd, f) {
		return nil
	}
	permutation := make([]int, len(dd.Vars))
	for i := range permutation {
		permutation[i] = i
	}
	for rank, index := range dd.Support(f) {
		permutation[index] = dd.InvPerm[rank]
	}
	return cudd.Cudd_bddPermute(dd, f, permutation)
}
func Extra_bddMove(dd *DdManager, f *DdNode, distance int) *DdNode {
	if !extraValidBDD(dd, f) {
		return nil
	}
	permutation := make([]int, len(dd.Vars))
	for i := range permutation {
		permutation[i] = i
	}
	for _, i := range dd.Support(f) {
		target := i + distance
		if target < 0 || target >= len(dd.Vars) {
			return nil
		}
		permutation[i] = target
	}
	return cudd.Cudd_bddPermute(dd, f, permutation)
}
func extraBddMove(dd *DdManager, f, flag *DdNode) *DdNode {
	if !extraValidBDD(dd, flag) || !Extra_bddIsVar(flag) {
		return nil
	}
	distance := int(cudd.Cudd_Regular(flag).Index)
	if cudd.Cudd_IsComplement(flag) {
		distance = -distance
	}
	return Extra_bddMove(dd, f, distance)
}
func Extra_StopManager(dd *DdManager) {
	if dd != nil {
		cudd.Cudd_Quit(dd)
	}
}
func Extra_bddFormat(dd *DdManager, f *DdNode) string {
	if f == nil {
		return "NULL"
	}
	if !extraValidBDD(dd, f) {
		return "Invalid BDD"
	}
	if f == dd.LogicZero {
		return "Constant 0"
	}
	if f == dd.One {
		return "Constant 1"
	}
	terms := []string{}
	literals := map[int]bool{}
	var visit func(*DdNode)
	visit = func(f *DdNode) {
		if f == dd.LogicZero {
			return
		}
		if f == dd.One {
			ids := make([]int, 0, len(literals))
			for i := range literals {
				ids = append(ids, i)
			}
			sort.Ints(ids)
			var term strings.Builder
			for _, i := range ids {
				fmt.Fprintf(&term, "[%d]", i)
				if !literals[i] {
					term.WriteByte('\'')
				}
			}
			terms = append(terms, term.String())
			return
		}
		r := cudd.Cudd_Regular(f)
		lo, hi := extraBDDCofactors(f)
		literals[int(r.Index)] = false
		visit(lo)
		literals[int(r.Index)] = true
		visit(hi)
		delete(literals, int(r.Index))
	}
	visit(f)
	return strings.Join(terms, " + ")
}
func Extra_bddPrint(dd *DdManager, f *DdNode)        { fmt.Print(Extra_bddFormat(dd, f)) }
func Extra_bddPrintSupport(dd *DdManager, f *DdNode) { Extra_bddPrint(dd, cudd.Cudd_Support(dd, f)) }
func Extra_bddSuppSize(dd *DdManager, cube *DdNode) int {
	vars, valid := extraPositiveCube(dd, cube)
	if !valid {
		return -1
	}
	return len(vars)
}
func Extra_bddSuppContainVar(dd *DdManager, cube, variable *DdNode) bool {
	vars, valid := extraPositiveCube(dd, cube)
	if !valid || !extraValidBDD(dd, variable) || !Extra_bddIsVar(variable) {
		return false
	}
	index := int(cudd.Cudd_Regular(variable).Index)
	for _, v := range vars {
		if v == index {
			return true
		}
	}
	return false
}
func Extra_bddSuppOverlapping(dd *DdManager, a, b *DdNode) bool {
	x, ok := extraPositiveCube(dd, a)
	y, valid := extraPositiveCube(dd, b)
	if !ok || !valid {
		return false
	}
	set := map[int]bool{}
	for _, i := range x {
		set[i] = true
	}
	for _, i := range y {
		if set[i] {
			return true
		}
	}
	return false
}
func Extra_bddSuppDifferentVars(dd *DdManager, a, b *DdNode, limit int) int {
	if limit < 0 {
		return -1
	}
	x, ok := extraPositiveCube(dd, a)
	y, valid := extraPositiveCube(dd, b)
	if !ok || !valid {
		return -1
	}
	difference := map[int]bool{}
	for _, i := range x {
		difference[i] = true
	}
	for _, i := range y {
		if difference[i] {
			delete(difference, i)
		} else {
			difference[i] = true
		}
	}
	return min(len(difference), limit)
}
func Extra_bddSuppCheckContainment(dd *DdManager, a, b *DdNode, large, small **DdNode) bool {
	x, ok := extraPositiveCube(dd, a)
	y, valid := extraPositiveCube(dd, b)
	if !ok || !valid || large == nil || small == nil {
		return false
	}
	sx, sy := map[int]bool{}, map[int]bool{}
	for _, i := range x {
		sx[i] = true
	}
	for _, i := range y {
		sy[i] = true
	}
	xContains, yContains := true, true
	for _, i := range y {
		if !sx[i] {
			xContains = false
		}
	}
	for _, i := range x {
		if !sy[i] {
			yContains = false
		}
	}
	if xContains == yContains {
		return false
	}
	if xContains {
		*large, *small = a, b
	} else {
		*large, *small = b, a
	}
	return true
}
func extraValidDD(dd *DdManager, f *DdNode) bool {
	if extraValidBDD(dd, f) {
		return true
	}
	if dd == nil || f == nil {
		return false
	}
	if cudd.Cudd_IsAdd(f) {
		return cudd.Cudd_addFindMin(dd, f) != nil
	}
	return cudd.Cudd_zddUnion(dd, f, dd.Zero) != nil
}
func Extra_SupportArray(dd *DdManager, f *DdNode, support []int) []int {
	return Extra_VectorSupportArray(dd, []*DdNode{f}, 1, support)
}
func Extra_VectorSupportArray(dd *DdManager, functions []*DdNode, n int, support []int) []int {
	if dd == nil || n < 0 || n > len(functions) {
		return nil
	}
	size := max(len(dd.Vars), dd.ZddSize)
	if support == nil {
		support = make([]int, size)
	} else if len(support) < size {
		return nil
	}
	for _, f := range functions[:n] {
		if !extraValidDD(dd, f) {
			return nil
		}
	}
	clear(support)
	seen := map[*DdNode]bool{}
	var visit func(*DdNode)
	visit = func(f *DdNode) {
		r := cudd.Cudd_Regular(f)
		if cudd.Cudd_IsConstant(r) || seen[r] {
			return
		}
		seen[r] = true
		support[r.Index] = 1
		visit(r.T)
		visit(r.E)
	}
	for _, f := range functions[:n] {
		visit(f)
	}
	return support
}
func Extra_bddGetOneCube(dd *DdManager, f *DdNode) *DdNode {
	if !extraValidBDD(dd, f) {
		return nil
	}
	if cudd.Cudd_IsConstant(f) {
		return f
	}
	lo, hi := extraBDDCofactors(f)
	v := dd.Vars[cudd.Cudd_Regular(f).Index]
	if lo != dd.LogicZero {
		return cudd.Cudd_bddAnd(dd, cudd.Cudd_Not(v), Extra_bddGetOneCube(dd, lo))
	}
	return cudd.Cudd_bddAnd(dd, v, Extra_bddGetOneCube(dd, hi))
}
func Extra_bddFindOneCube(dd *DdManager, f *DdNode) *DdNode { return Extra_bddGetOneCube(dd, f) }
func Extra_bddComputeRangeCube(dd *DdManager, start, stop int) *DdNode {
	if dd == nil || start < 0 || start > stop || stop > len(dd.Vars) {
		return nil
	}
	return Extra_bddComputeCube(dd, dd.Vars[start:stop], stop-start)
}
func Extra_bddBitsToCube(dd *DdManager, code, width int, vars []*DdNode, msb bool) *DdNode {
	return cascade.Extra_bddBitsToCube(dd, code, width, vars, msb)
}
func Extra_bddSupportNegativeCube(dd *DdManager, f *DdNode) *DdNode {
	support := Extra_SupportArray(dd, f, nil)
	if support == nil {
		return nil
	}
	result := dd.One
	for i, present := range support {
		if present != 0 {
			result = cudd.Cudd_bddAnd(dd, result, cudd.Cudd_Not(cudd.Cudd_bddIthVar(dd, i)))
		}
	}
	return result
}
func Extra_bddIsVar(f *DdNode) bool {
	if f == nil {
		return false
	}
	r := cudd.Cudd_Regular(f)
	return !cudd.Cudd_IsConstant(r) && !cudd.Cudd_IsAdd(r) && cudd.Cudd_IsConstant(r.T) && cudd.Cudd_IsConstant(r.E)
}
func extraCreate(dd *DdManager, n, kind int) *DdNode {
	if dd == nil || n < 0 {
		return nil
	}
	f := dd.LogicZero
	if kind == 0 {
		f = dd.One
	}
	for i := 0; i < n; i++ {
		v := cudd.Cudd_bddIthVar(dd, i)
		switch kind {
		case 0:
			f = cudd.Cudd_bddAnd(dd, f, v)
		case 1:
			f = cudd.Cudd_bddOr(dd, f, v)
		case 2:
			f = cudd.Cudd_bddXor(dd, f, v)
		}
	}
	return f
}
func Extra_bddCreateAnd(dd *DdManager, n int) *DdNode  { return extraCreate(dd, n, 0) }
func Extra_bddCreateOr(dd *DdManager, n int) *DdNode   { return extraCreate(dd, n, 1) }
func Extra_bddCreateExor(dd *DdManager, n int) *DdNode { return extraCreate(dd, n, 2) }
func Extra_bddComputeCube(dd *DdManager, vars []*DdNode, n int) *DdNode {
	if dd == nil || n < 0 || n > len(vars) {
		return nil
	}
	f := dd.One
	for _, v := range vars[:n] {
		if !extraValidBDD(dd, v) {
			return nil
		}
		f = cudd.Cudd_bddAnd(dd, f, v)
	}
	return f
}
func Extra_bddPermuteArray(dd *DdManager, input, output []*DdNode, n int, permutation []int) bool {
	if dd == nil || n < 0 || n > len(input) || n > len(output) {
		return false
	}
	results := make([]*DdNode, n)
	for i, f := range input[:n] {
		results[i] = cudd.Cudd_bddPermute(dd, f, permutation)
		if results[i] == nil {
			return false
		}
	}
	for i, f := range results {
		output[i] = f
		cudd.Cudd_Ref(f)
	}
	return true
}
func Extra_bddChangePolarity(dd *DdManager, f, vars *DdNode) *DdNode {
	indices, valid := extraPositiveCube(dd, vars)
	if !valid || !extraValidBDD(dd, f) {
		return nil
	}
	vector := append([]*DdNode(nil), dd.Vars...)
	for _, i := range indices {
		vector[i] = cudd.Cudd_Not(vector[i])
	}
	return cudd.Cudd_bddVectorCompose(dd, f, vector)
}
func extraBddChangePolarity(dd *DdManager, f, vars *DdNode) *DdNode {
	return Extra_bddChangePolarity(dd, f, vars)
}
func Extra_bddVarIsInCube(cube *DdNode, index int) int {
	if cube == nil {
		return -1
	}
	for !cudd.Cudd_IsConstant(cube) {
		r := cudd.Cudd_Regular(cube)
		lo, hi := extraBDDCofactors(cube)
		zeroLo := cudd.Cudd_IsConstant(lo) && cudd.Cudd_IsComplement(lo)
		zeroHi := cudd.Cudd_IsConstant(hi) && cudd.Cudd_IsComplement(hi)
		if !zeroLo && !zeroHi {
			return -1
		}
		if int(r.Index) == index {
			if zeroLo {
				return 1
			}
			return 0
		}
		if zeroLo {
			cube = hi
		} else {
			cube = lo
		}
	}
	return -1
}

// Extra_bddAndPermute fuses cross-manager transfer with conjunction. The
// recursion splits in the destination order and memoizes both source handles.
func Extra_bddAndPermute(ddF *DdManager, f *DdNode, ddG *DdManager, g *DdNode, permutation []int) *DdNode {
	if !extraValidBDD(ddF, f) || !extraValidBDD(ddG, g) {
		return nil
	}
	target := func(index int) int {
		if permutation == nil {
			return index
		}
		if index >= len(permutation) {
			return -1
		}
		return permutation[index]
	}
	highest := -1
	for _, i := range ddG.Support(g) {
		j := target(i)
		if j < 0 || uint64(j) >= uint64(cudd.CUDD_MAXINDEX) {
			return nil
		}
		highest = max(highest, j)
	}
	if highest >= 0 && cudd.Cudd_bddIthVar(ddF, highest) == nil {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, g *DdNode) *DdNode {
		if f == ddF.LogicZero || g == ddG.LogicZero {
			return ddF.LogicZero
		}
		if g == ddG.One {
			return f
		}
		key := [2]*DdNode{f, g}
		if r := memo[key]; r != nil {
			return r
		}
		lf := math.MaxInt
		if !cudd.Cudd_IsConstant(f) {
			lf = ddF.Perm[cudd.Cudd_Regular(f).Index]
		}
		lg := ddF.Perm[target(int(cudd.Cudd_Regular(g).Index))]
		level := min(lf, lg)
		loF, hiF := f, f
		if lf == level {
			loF, hiF = extraBDDCofactors(f)
		}
		loG, hiG := g, g
		if lg == level {
			loG, hiG = extraBDDCofactors(g)
		}
		r := cudd.Cudd_bddIte(ddF, ddF.Vars[ddF.InvPerm[level]], visit(hiF, hiG), visit(loF, loG))
		memo[key] = r
		return r
	}
	return visit(f, g)
}
func Extra_bddCountCubes(dd *DdManager, functions []*DdNode, n, mode, limit int, guide []int) int {
	if dd == nil || n < 0 || n > len(functions) || n > len(guide) || (mode != -1 && mode != 0 && mode != 1) || limit < 0 {
		return -1
	}
	baseline := cudd.Cudd_ReadNodeCount(dd)
	total := 0
	count := func(f *DdNode, bound int) (int, bool) {
		remaining := 1000000 - (cudd.Cudd_ReadNodeCount(dd) - baseline)
		if remaining < 0 {
			return -1, false
		}
		result, nCubes := cudd.CountISOPBounded(dd, f, f, bound, remaining)
		return nCubes, result != nil
	}
	for i, f := range functions[:n] {
		if f == nil {
			continue
		}
		if !extraValidBDD(dd, f) {
			return -1
		}
		pos, neg := 0, 0
		posOK, negOK := false, false
		if mode != 0 {
			pos, posOK = count(f, limit)
		}
		if mode != 1 {
			bound := limit
			if posOK {
				bound = pos
			}
			neg, negOK = count(cudd.Cudd_Not(f), bound)
		}
		chosen := 0
		if posOK && (!negOK || pos <= neg) {
			guide[i] = 1
			chosen = pos
		} else if negOK {
			guide[i] = 0
			chosen = neg
		} else {
			return -1
		}
		if chosen > int(^uint(0)>>1)-total {
			return -1
		}
		total += chosen
	}
	return total
}
func Extra_bddTuples(dd *DdManager, k int, vars *DdNode) *DdNode {
	indices, valid := extraPositiveCube(dd, vars)
	if !valid || k < 0 || k > len(indices) {
		return nil
	}
	memo := map[[2]int]*DdNode{}
	var visit func(int, int) *DdNode
	visit = func(pos, k int) *DdNode {
		if k < 0 || k > len(indices)-pos {
			return dd.LogicZero
		}
		if pos == len(indices) {
			return dd.One
		}
		key := [2]int{pos, k}
		if f := memo[key]; f != nil {
			return f
		}
		r := cudd.Cudd_bddIte(dd, dd.Vars[indices[pos]], visit(pos+1, k-1), visit(pos+1, k))
		memo[key] = r
		return r
	}
	return visit(0, k)
}
func extraBddTuples(dd *DdManager, kVars, vars *DdNode) *DdNode {
	indices, valid := extraPositiveCube(dd, kVars)
	if !valid {
		return nil
	}
	return Extra_bddTuples(dd, len(indices), vars)
}
func Extra_DecomposeCover(dd *DdManager, cover *DdNode) (negative, positive, absent *DdNode, ok bool) {
	if cover == nil || cudd.Cudd_IsConstant(cover) {
		return
	}
	positive, negative, absent, ok = cudd.ZDDCoverCofactors(dd, cover, int(cover.Index))
	return
}
func extraDecomposeCover(dd *DdManager, cover *DdNode, negative, positive, absent **DdNode) {
	n, p, a, ok := Extra_DecomposeCover(dd, cover)
	if !ok || negative == nil || positive == nil || absent == nil {
		return
	}
	*negative, *positive, *absent = n, p, a
}
func Extra_zddCombination(dd *DdManager, values []int, n int) *DdNode {
	if dd == nil || n < 0 || n > len(values) {
		return nil
	}
	f := dd.One
	for i, value := range values[:n] {
		if value == 1 {
			f = cudd.Cudd_zddChange(dd, f, i)
			if f == nil {
				return nil
			}
		}
	}
	return f
}
func extraZddCombination(dd *DdManager, values []int, n int) *DdNode {
	return Extra_zddCombination(dd, values, n)
}
func extraRandomSet(dd *DdManager, n, k int, density float64, rng *rand.Rand) *DdNode {
	if dd == nil || n <= 0 || k <= 0 || math.IsNaN(density) || density < 0 || density > 1 || (n < 63 && uint64(k) > uint64(1)<<uint(n)) || ((density == 0 || density == 1) && k > 1) {
		return nil
	}
	f := dd.Zero
	values := make([]int, n)
	for count := 0; count < k; {
		for i := range values {
			values[i] = 0
			if rng.Float64() < density {
				values[i] = 1
			}
		}
		cube := Extra_zddCombination(dd, values, n)
		if cudd.Cudd_zddIntersect(dd, f, cube) != dd.Zero {
			continue
		}
		f = cudd.Cudd_zddUnion(dd, f, cube)
		count++
	}
	return f
}
func Extra_zddRandomSet(dd *DdManager, n, k int, density float64) *DdNode {
	return extraRandomSet(dd, n, k, density, rand.New(rand.NewSource(time.Now().UnixNano())))
}
func extraPlaValid(dd *DdManager, f *DdNode, n int) error {
	if dd == nil || f == nil || n < 0 || cudd.Cudd_zddUnion(dd, f, dd.Zero) == nil {
		return fmt.Errorf("invalid PLA cover")
	}
	indices := map[int]bool{}
	for node := range Extra_CollectNodes(f) {
		if cudd.Cudd_IsConstant(node) {
			continue
		}
		if int(node.Index)/2 >= n {
			return fmt.Errorf("PLA cover exceeds input count")
		}
		indices[int(node.Index)] = true
	}
	for index := range indices {
		if index&1 == 0 && indices[index+1] {
			positive := cudd.Cudd_zddSubset1(dd, f, index)
			if cudd.Cudd_zddSubset1(dd, positive, index+1) != dd.Zero {
				return fmt.Errorf("contradictory PLA literal pair")
			}
		}
	}
	return nil
}
func extraPlaWrite(w io.Writer, dd *DdManager, f *DdNode, n int) error {
	var failure error
	write := func(format string, args ...any) {
		if failure == nil {
			_, failure = fmt.Fprintf(w, format, args...)
		}
	}
	write("# PLA file dumped by Extra_zddDumpPla() in ABC\n.i %d\n.o 1\n", n)
	cube := make([]byte, n)
	for i := range cube {
		cube[i] = '-'
	}
	var visit func(*DdNode)
	visit = func(f *DdNode) {
		if f == dd.Zero || failure != nil {
			return
		}
		if f == dd.One {
			write("%s 1\n", cube)
			return
		}
		visit(f.E)
		index := int(f.Index) / 2
		old := cube[index]
		if f.Index&1 == 0 {
			cube[index] = '1'
		} else {
			cube[index] = '0'
		}
		visit(f.T)
		cube[index] = old
	}
	visit(f)
	write(".e\n\n")
	return failure
}
func Extra_zddWritePla(w io.Writer, dd *DdManager, f *DdNode, n int) error {
	if w == nil {
		return fmt.Errorf("nil PLA writer")
	}
	if err := extraPlaValid(dd, f, n); err != nil {
		return err
	}
	return extraPlaWrite(w, dd, f, n)
}
func Extra_zddDumpPla(dd *DdManager, f *DdNode, n int, fileName string) error {
	if err := extraPlaValid(dd, f, n); err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	err = extraPlaWrite(file, dd, f, n)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Extra_TestAndPerm(dd *DdManager, f, g *DdNode) {
	if !extraValidBDD(dd, f, g) {
		return
	}
	other := cudd.NewBDDManager(len(dd.Vars))
	defer cudd.Cudd_Quit(other)
	cudd.Cudd_ShuffleHeap(other, dd.InvPerm)
	transferred := Extra_TransferLevelByLevel(dd, other, g)
	cudd.Cudd_Ref(transferred)
	defer cudd.Cudd_RecursiveDeref(other, transferred)
	cudd.Cudd_ReduceHeap(other, cudd.CUDD_REORDER_SYMM_SIFT, 1)
	start := time.Now()
	want := cudd.Cudd_bddAnd(dd, f, g)
	fmt.Printf("Runtime of Cudd_bddAnd: %s\n", time.Since(start))
	start = time.Now()
	got := Extra_bddAndPermute(dd, f, other, transferred, nil)
	fmt.Printf("Runtime of fused transfer/conjunction: %s\nResult verified: %t\n", time.Since(start), got == want)
}
func Extra_GraphExperiment() {
	dd := cudd.NewBDDManager(0)
	defer cudd.Cudd_Quit(dd)
	family := dd.Zero
	for _, edge := range [][]int{{1, 3, 4}, {1, 5}, {2, 3, 5}, {2, 4}} {
		cube := dd.One
		for _, v := range edge {
			cube = cudd.Cudd_zddChange(dd, cube, v)
		}
		family = cudd.Cudd_zddUnion(dd, family, cube)
	}
	fmt.Printf("Graph hyperedges: %d; ZDD nodes: %d\n", cudd.Cudd_zddCount(dd, family), cudd.Cudd_DagSize(family))
}
func Extra_ZddTest() {
	dd := cudd.NewBDDManager(0)
	defer cudd.Cudd_Quit(dd)
	for i := 1; i <= 10; i++ {
		n, k := 64, i*1000
		f := Extra_zddRandomSet(dd, n, k, .5)
		size := cudd.Cudd_DagSize(f)
		fmt.Printf("N = %5d K = %5d ZDD size = %6d MemBDD = %.3f MB MemBit = %.3f MB\n", n, k, size, 20.*float64(size)/(1<<20), .125*float64(n*k)/(1<<20))
	}
}

// Live source internal entrypoint shares the authoritative transfer primitive.
func extraTransferPermute(source, destination *DdManager, f *DdNode, permutation []int) *DdNode {
	return Extra_TransferPermute(source, destination, f, permutation)
}
