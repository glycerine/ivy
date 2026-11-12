/**CFile****************************************************************

FileName    [extraBddThresh.c]

SystemName  [ABC: Logic synthesis and verification system.]

PackageName [extra]

Synopsis    [Dealing with threshold functions.]

Author      [Augusto Neutzling, Jody Matos, and Alan Mishchenko (UC Berkeley).]

Affiliation [Federal University of Rio Grande do Sul, Brazil]

Date        [Ver. 1.0. Started - October 7, 2014.]

Revision    [$Id: extraBddThresh.c,v 1.0 2014/10/07 00:00:00 alanmi Exp $]

***********************************************************************/

package extrab

import (
	"fmt"
	"sort"
	"strings"
)

func thresholdWords(n int) int {
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func thresholdValid(t []uint64, n int) bool   { return n >= 1 && n <= 16 && len(t) >= thresholdWords(n) }
func thresholdBit(t []uint64, index int) bool { return t[index>>6]&(uint64(1)<<uint(index&63)) != 0 }
func thresholdSet(t []uint64, index int, value bool) {
	mask := uint64(1) << uint(index&63)
	if value {
		t[index>>6] |= mask
	} else {
		t[index>>6] &= ^mask
	}
}
func thresholdPolarity(t []uint64, n, v int) (positive, negative bool) {
	positive, negative = true, true
	for a := 0; a < 1<<uint(n); a++ {
		if a&(1<<uint(v)) != 0 {
			continue
		}
		lo, hi := thresholdBit(t, a), thresholdBit(t, a|1<<uint(v))
		if lo && !hi {
			positive = false
		}
		if hi && !lo {
			negative = false
		}
	}
	return
}
func thresholdMakePositive(t []uint64, n int) bool {
	polarities := make([]bool, n)
	for v := 0; v < n; v++ {
		pos, neg := thresholdPolarity(t, n, v)
		if !pos && !neg {
			return false
		}
		polarities[v] = neg
	}
	for v, flip := range polarities {
		if !flip {
			continue
		}
		for a := 0; a < thresholdWords(n)*64; a++ {
			if a&(1<<uint(v)) != 0 {
				continue
			}
			b := a | 1<<uint(v)
			lo, hi := thresholdBit(t, a), thresholdBit(t, b)
			thresholdSet(t, a, hi)
			thresholdSet(t, b, lo)
		}
	}
	return true
}
func thresholdSwap(t []uint64, n, v int) {
	for a := 0; a < thresholdWords(n)*64; a++ {
		if a&(1<<uint(v)) != 0 || a&(1<<uint(v+1)) == 0 {
			continue
		}
		b := a ^ (3 << uint(v))
		lo, hi := thresholdBit(t, a), thresholdBit(t, b)
		thresholdSet(t, a, hi)
		thresholdSet(t, b, lo)
	}
}
func Extra_ThreshPrintChow(chow0 int, chow []int, n int) {
	for _, v := range chow[:n] {
		fmt.Printf("%d ", v)
	}
	fmt.Printf("  %d\n", chow0)
}
func Extra_ThreshComputeChow(t []uint64, n int, chow []int) int {
	if !thresholdValid(t, n) || len(chow) < n {
		return 0
	}
	clear(chow[:n])
	count := 0
	for a := 0; a < 1<<uint(n); a++ {
		if !thresholdBit(t, a) {
			continue
		}
		count++
		for v := 0; v < n; v++ {
			if a&(1<<uint(v)) != 0 {
				chow[v]++
			}
		}
	}
	for v := 0; v < n; v++ {
		chow[v] = 2*chow[v] - count
	}
	return count - (1 << uint(n-1))
}
func thresholdSort(t []uint64, n int, chow []int, inverted bool) {
	if !thresholdValid(t, n) || len(chow) < n {
		return
	}
	for {
		changed := false
		for v := 0; v < n-1; v++ {
			swap := chow[v] < chow[v+1]
			if inverted {
				swap = chow[v] > chow[v+1]
			}
			if swap {
				chow[v], chow[v+1] = chow[v+1], chow[v]
				thresholdSwap(t, n, v)
				changed = true
			}
		}
		if !changed {
			return
		}
	}
}
func Extra_ThreshSortByChow(t []uint64, n int, chow []int)         { thresholdSort(t, n, chow, false) }
func Extra_ThreshSortByChowInverted(t []uint64, n int, chow []int) { thresholdSort(t, n, chow, true) }
func Extra_ThreshInitializeChow(n int, chow []int) int {
	if n <= 0 || len(chow) < n {
		return 0
	}
	groups := make([]int, n)
	group := 0
	for i := 1; i < n; i++ {
		if chow[i] != chow[i-1] {
			group++
		}
		groups[i] = group
	}
	copy(chow, groups)
	return group + 1
}
func Extra_ThreshWeightedSum(weights []int, n, assignment int) int {
	sum := 0
	for i, w := range weights[:n] {
		if assignment&(1<<uint(i)) != 0 {
			sum += w
		}
	}
	return sum
}
func Extra_ThreshCubeWeightedSum1(weights, chow []int, sop string, n, start int) int {
	sum := 0
	for v := 0; v < n; v++ {
		if sop[start+v] == '1' {
			sum += weights[chow[v]]
		}
	}
	return sum
}
func Extra_ThreshCubeWeightedSum2(weights, chow []int, sop string, n, start int) int {
	sum := 0
	for v := 0; v < n; v++ {
		if sop[start+v] == '-' {
			sum += weights[chow[v]]
		}
	}
	return sum
}
func Extra_ThreshCubeWeightedSum3(weights []int, n int, greater [][]uint64, row int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += weights[i] * int(greater[row][i])
	}
	return sum
}
func Extra_ThreshCubeWeightedSum4(weights []int, n int, smaller [][]uint64, row int) int {
	return Extra_ThreshCubeWeightedSum3(weights, n, smaller, row)
}

func thresholdSelect(t []uint64, n int, weights []int) int {
	if !thresholdValid(t, n) || len(weights) < n {
		return 0
	}
	if n <= 2 {
		if t[0]&15 != 6 && t[0]&15 != 9 {
			return 1
		}
		return 0
	}
	if n > 8 {
		return 0
	}
	limit := n
	switch n {
	case 6:
		limit = n + 3
	case 7:
		limit = n + 6
	case 8:
		limit = n + 1
	}
	var search func(int, int) int
	search = func(index, minimum int) int {
		if index < 0 {
			lower, upper := 10000, 0
			for a := 0; a < 1<<uint(n); a++ {
				sum := Extra_ThreshWeightedSum(weights, n, a)
				if thresholdBit(t, a) {
					lower = min(lower, sum)
				} else {
					upper = max(upper, sum)
				}
				if upper >= lower {
					return 0
				}
			}
			return lower
		}
		for weights[index] = minimum; weights[index] <= limit; weights[index]++ {
			if threshold := search(index-1, weights[index]); threshold != 0 {
				return threshold
			}
		}
		return 0
	}
	return search(n-1, 1)
}
func Extra_ThreshSelectWeights(t []uint64, n int, weights []int) int {
	return thresholdSelect(t, n, weights)
}
func Extra_ThreshSelectWeights3(t []uint64, n int, weights []int) int {
	if n != 3 {
		return 0
	}
	return thresholdSelect(t, n, weights)
}
func Extra_ThreshSelectWeights4(t []uint64, n int, weights []int) int {
	if n != 4 {
		return 0
	}
	return thresholdSelect(t, n, weights)
}
func Extra_ThreshSelectWeights5(t []uint64, n int, weights []int) int {
	if n != 5 {
		return 0
	}
	return thresholdSelect(t, n, weights)
}
func Extra_ThreshSelectWeights6(t []uint64, n int, weights []int) int {
	if n != 6 {
		return 0
	}
	return thresholdSelect(t, n, weights)
}
func Extra_ThreshSelectWeights7(t []uint64, n int, weights []int) int {
	if n != 7 {
		return 0
	}
	return thresholdSelect(t, n, weights)
}
func Extra_ThreshSelectWeights8(t []uint64, n int, weights []int) int {
	if n != 8 {
		return 0
	}
	return thresholdSelect(t, n, weights)
}
func Extra_ThreshIncrementWeights(n int, weights []int, i int) {
	for k := i; k < n; k++ {
		weights[k]++
	}
}
func Extra_ThreshDecrementWeights(n int, weights []int, i int) {
	for k := i; k < n; k++ {
		weights[k]--
	}
}
func Extra_ThreshPrintInequalities(greater, smaller [][]uint64, n, rows int) {
	for row := 0; row < rows; row++ {
		fmt.Printf("\n Inequality [%d] = ", row)
		for _, v := range greater[row][:n] {
			fmt.Printf("%d ", v)
		}
		fmt.Print(" > ")
		for _, v := range smaller[row][:n] {
			fmt.Printf("%d ", v)
		}
	}
}
func thresholdSOPValid(sop string, n int) bool {
	if len(sop)%(n+3) != 0 {
		return false
	}
	for start := 0; start < len(sop); start += n + 3 {
		for _, c := range sop[start : start+n] {
			if c != '0' && c != '1' && c != '-' {
				return false
			}
		}
	}
	return true
}
func Extra_ThreshCreateInequalities(on, off string, n int, weights, chow []int, groups, rows int, greater, smaller [][]uint64) {
	for row := 0; row < rows; row++ {
		clear(greater[row][:groups])
		clear(smaller[row][:groups])
	}
	row := 0
	for a := 0; a < len(on); a += n + 3 {
		for b := 0; b < len(off); b += n + 3 {
			for v := 0; v < n; v++ {
				if on[a+v] == '1' {
					greater[row][chow[v]]++
				}
				if off[b+v] == '-' {
					smaller[row][chow[v]]++
				}
			}
			row++
		}
	}
}
func Extra_ThreshSimplifyInequalities(rows, groups int, greater, smaller [][]uint64) {
	for row := 0; row < rows; row++ {
		for g := 0; g < groups; g++ {
			common := min(greater[row][g], smaller[row][g])
			greater[row][g] -= common
			smaller[row][g] -= common
		}
	}
}
func Extra_ThreshAssignWeights(t []uint64, on, off string, n int, weights, chow []int, groups, minimum int) int {
	if n <= 0 || n > 16 || len(weights) < n || len(chow) < n || groups < 1 || !thresholdSOPValid(on, n) || !thresholdSOPValid(off, n) {
		return 0
	}
	for _, g := range chow[:n] {
		if g < 0 || g >= groups {
			return 0
		}
	}
	rows := (len(on) / (n + 3)) * (len(off) / (n + 3))
	greater, smaller := make([][]uint64, rows), make([][]uint64, rows)
	for row := 0; row < rows; row++ {
		greater[row] = make([]uint64, groups)
		smaller[row] = make([]uint64, groups)
	}
	groupWeights := make([]int, groups)
	Extra_ThreshCreateInequalities(on, off, n, groupWeights, chow, groups, rows, greater, smaller)
	Extra_ThreshSimplifyInequalities(rows, groups, greater, smaller)
	for i := range groupWeights {
		groupWeights[i] = minimum + i
	}
	row, oldDelta, incremented := 0, -1000, false
	for group := 0; group < groups && groupWeights[groups-1] <= n*2; group++ {
		for row < rows {
			if greater[row][group] == 0 {
				row++
				continue
			}
			lower := Extra_ThreshCubeWeightedSum3(groupWeights, groups, greater, row)
			upper := Extra_ThreshCubeWeightedSum4(groupWeights, groups, smaller, row)
			delta := lower - upper
			if delta > 0 {
				if incremented {
					row = 0
					incremented = false
					oldDelta = -1000
				} else {
					row++
				}
				continue
			}
			if delta > oldDelta {
				Extra_ThreshIncrementWeights(groups, groupWeights, group)
				oldDelta = delta
				incremented = true
			} else if incremented {
				Extra_ThreshDecrementWeights(groups, groupWeights, group)
				row++
				oldDelta = -1000
				incremented = false
			} else {
				row++
			}
		}
		row = 0
	}
	lower, upper := 1000, 0
	for start := 0; start < len(on); start += n + 3 {
		lower = min(lower, Extra_ThreshCubeWeightedSum1(groupWeights, chow, on, n, start))
	}
	for start := 0; start < len(off); start += n + 3 {
		upper = max(upper, Extra_ThreshCubeWeightedSum2(groupWeights, chow, off, n, start))
	}
	for v := 0; v < n; v++ {
		weights[v] = groupWeights[chow[v]]
	}
	if lower > upper {
		return lower
	}
	return 0
}
func Extra_ThreshPrintWeights(threshold int, weights []int, n int) {
	if threshold == 0 {
		fmt.Print("\nHeuristic method: is not TLF\n\n")
		return
	}
	fmt.Print("\nHeuristic method: Weights and threshold value:\n")
	Extra_ThreshPrintChow(threshold, weights, n)
}
func Extra_ThreshCheck(t []uint64, n int, weights []int) int {
	if !thresholdValid(t, n) || len(weights) < n || !thresholdMakePositive(t, n) {
		return 0
	}
	chow := make([]int, n)
	Extra_ThreshComputeChow(t, n, chow)
	Extra_ThreshSortByChow(t, n, chow)
	return Extra_ThreshSelectWeights(t, n, weights)
}

// thresholdPrimeSOP enumerates all prime cubes of a monotone function: minimal
// true assignments or maximal false assignments. Literal-first order matches
// Abc_ConvertZddToSop_rec, without a global ABC frame/BDD manager dependency.
func thresholdPrimeSOP(t []uint64, n int, on bool) string {
	cubes := []string{}
	for a := 0; a < 1<<uint(n); a++ {
		if thresholdBit(t, a) != on {
			continue
		}
		prime := true
		for v := 0; v < n; v++ {
			bit := 1 << uint(v)
			if on && a&bit != 0 && thresholdBit(t, a^bit) {
				prime = false
				break
			}
			if !on && a&bit == 0 && !thresholdBit(t, a|bit) {
				prime = false
				break
			}
		}
		if !prime {
			continue
		}
		cube := make([]byte, n+3)
		for v := 0; v < n; v++ {
			cube[v] = '-'
			if on && a&(1<<uint(v)) != 0 {
				cube[v] = '1'
			}
			if !on && a&(1<<uint(v)) == 0 {
				cube[v] = '0'
			}
		}
		copy(cube[n:], " 1\n")
		cubes = append(cubes, string(cube))
	}
	sort.Slice(cubes, func(i, j int) bool {
		for v := 0; v < n; v++ {
			if cubes[i][v] != cubes[j][v] {
				return cubes[i][v] != '-'
			}
		}
		return false
	})
	return strings.Join(cubes, "")
}
func Extra_ThreshHeuristic(t []uint64, n int, weights []int) int {
	if n >= 0 && n <= 1 {
		return 1
	}
	if !thresholdValid(t, n) || len(weights) < n || !thresholdMakePositive(t, n) {
		return 0
	}
	chow := make([]int, n)
	Extra_ThreshComputeChow(t, n, chow)
	Extra_ThreshSortByChowInverted(t, n, chow)
	groups := Extra_ThreshInitializeChow(n, chow)
	on, off := thresholdPrimeSOP(t, n, true), thresholdPrimeSOP(t, n, false)
	// Preserve the source's complement side effect, but use the 64-bit word count
	// instead of its erroneous 32-bit word count (which writes beyond the buffer).
	for i := 0; i < thresholdWords(n); i++ {
		t[i] = ^t[i]
	}
	threshold := Extra_ThreshAssignWeights(t, on, off, n, weights, chow, groups, 1)
	for minimum := 2; minimum < 4 && threshold == 0 && n >= 6; minimum++ {
		threshold = Extra_ThreshAssignWeights(t, on, off, n, weights, chow, groups, minimum)
	}
	return threshold
}
func thresholdExample() ([]uint64, int) {
	t := []uint64{0}
	for a := 0; a < 64; a++ {
		x := func(v int) bool { return a&(1<<uint(v)) != 0 }
		if x(0) && x(1) && x(2) || x(0) && x(1) && x(3) || x(0) && x(1) && x(4) && x(5) || x(0) && x(2) && x(3) && x(4) && x(5) {
			thresholdSet(t, a, true)
		}
	}
	return t, 6
}
func Extra_ThreshCheckTest() {
	t, n := thresholdExample()
	weights := make([]int, n)
	threshold := Extra_ThreshCheck(t, n, weights)
	if threshold != 0 {
		Extra_ThreshPrintChow(threshold, weights, n)
	} else {
		fmt.Println("No threshold")
	}
}
func Extra_ThreshHeuristicTest() {
	t, n := thresholdExample()
	weights := make([]int, n)
	Extra_ThreshPrintWeights(Extra_ThreshHeuristic(t, n, weights), weights, n)
}
