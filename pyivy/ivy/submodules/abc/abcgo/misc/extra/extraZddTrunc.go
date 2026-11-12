package extra

import (
	"fmt"
	"math"

	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

const TEST_VAR_MAX = 10
const TEST_SET_MAX = 10

// Extra_zddTruncate_rec enumerates else before then, retaining a subset when
// the product of included-variable probabilities is at least the limit.
// Omitted variables contribute factor one, matching the original procedure.
func Extra_zddTruncate_rec(dd *core.DdManager, function *core.DdNode, probabilities []float64, limit, current float64, subset, result *vec.Vec_Int_t) {
	if current < limit || function == core.Cudd_ReadZero(dd) {
		return
	}
	if function == core.Cudd_ReadOne(dd) {
		vec.Vec_IntWriteEntry(result, 0, vec.Vec_IntEntry(result, 0)+1)
		for _, v := range vec.Vec_IntArray(subset) {
			vec.Vec_IntPush(result, v)
		}
		vec.Vec_IntPush(result, -1)
		return
	}
	if function == nil || core.Cudd_IsConstant(function) {
		panic("invalid ZDD terminal")
	}
	index := core.Cudd_NodeReadIndex(function)
	Extra_zddTruncate_rec(dd, core.Cudd_E(function), probabilities, limit, current, subset, result)
	vec.Vec_IntPush(subset, int32(index))
	Extra_zddTruncate_rec(dd, core.Cudd_T(function), probabilities, limit, current*probabilities[index], subset, result)
	vec.Vec_IntPop(subset)
}
func Extra_zddTruncate(dd *core.DdManager, function *core.DdNode, probabilities []float64, limit float64) []int32 {
	if dd == nil || function == nil || math.IsNaN(limit) || limit <= 0 || limit > 1 || len(probabilities) < core.Cudd_ReadZddSize(dd) {
		panic("invalid ZDD probability dimensions")
	}
	for _, p := range probabilities[:core.Cudd_ReadZddSize(dd)] {
		if math.IsNaN(p) || p <= 0 || p > 1 {
			panic("ZDD probability outside (0,1]")
		}
	}
	subset, result := vec.Vec_IntAlloc(core.Cudd_ReadZddSize(dd)), vec.Vec_IntAlloc(10*core.Cudd_ReadZddSize(dd))
	vec.Vec_IntPush(result, 0)
	Extra_zddTruncate_rec(dd, function, probabilities, limit, 1, subset, result)
	return vec.Vec_IntArray(result)
}
func Extra_zddVariable(dd *core.DdManager, index int32) *core.DdNode {
	if index < 0 || int(index) >= core.Cudd_ReadZddSize(dd) {
		panic("invalid ZDD variable")
	}
	return core.Cudd_zddChange(dd, core.Cudd_ReadOne(dd), int(index))
}
func Extra_zddCreateSubsets(dd *core.DdManager, subsets [][]int32, count int32) *core.DdNode {
	result := core.Cudd_ReadZero(dd)
	for _, subset := range subsets[:count] {
		one := core.Cudd_ReadOne(dd)
		for _, v := range subset {
			if v == -1 {
				break
			}
			one = core.Cudd_zddUnateProduct(dd, one, Extra_zddVariable(dd, v))
			if one == nil {
				panic("ZDD subset construction failed")
			}
		}
		result = core.Cudd_zddUnion(dd, result, one)
		if result == nil {
			panic("ZDD union failed")
		}
	}
	return result
}
func Extra_zddPrintSubsets(subsets []int32) {
	if len(subsets) == 0 || subsets[0] < 0 {
		panic("invalid packed subsets")
	}
	fmt.Printf("The set contains %d subsets:\n", subsets[0])
	at := 1
	for i := int32(0); i < subsets[0]; i++ {
		fmt.Printf("Subset %3d : {", i)
		for {
			if at >= len(subsets) {
				panic("unterminated packed subset")
			}
			v := subsets[at]
			at++
			if v == -1 {
				break
			}
			fmt.Printf(" %d", v)
		}
		fmt.Println(" }")
	}
}
func Extra_zddTruncateTest() {
	dd := core.Cudd_Init(0, TEST_VAR_MAX, 256, 256, 0)
	defer core.Cudd_Quit(dd)
	subsets := [][]int32{{0, 3, 5, -1}, {1, 2, 3, 6, 9, -1}, {1, 5, 7, 8, -1}, {2, 4, -1}, {0, 5, 6, 9, -1}}
	function := Extra_zddCreateSubsets(dd, subsets, int32(len(subsets)))
	probabilities := make([]float64, TEST_VAR_MAX)
	for i := range probabilities {
		probabilities[i] = 0.1
	}
	fmt.Printf("The initial ZDD representing %d subsets:\n", len(subsets))
	Extra_zddPrintSubsets(Extra_zddTruncate(dd, function, probabilities, math.SmallestNonzeroFloat64))
	Extra_zddPrintSubsets(Extra_zddTruncate(dd, function, probabilities, 0.001))
}
