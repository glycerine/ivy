package live

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/big"
)

func countCombination(n, k int64) int64 {
	if n < 0 || k < 0 || k > n {
		panic("live: invalid combination dimensions")
	}
	value := new(big.Int).Binomial(n, k)
	if !value.IsInt64() {
		panic("live: combination count overflows int64")
	}
	return value.Int64()
}

// ForEachCombination streams source colexicographic order. The callback receives
// borrowed ascending indexes and may stop enumeration by returning false.
func ForEachCombination(n, k int, visit func([]int) bool) {
	if n < 0 || k < 0 || k > n {
		panic("live: invalid combination dimensions")
	}
	values := make([]int, k+2)
	for i := 0; i < k; i++ {
		values[i] = i
	}
	values[k] = n
	for {
		if !visit(values[:k]) {
			return
		}
		j := 0
		for j < k && values[j]+1 == values[j+1] {
			values[j] = j
			j++
		}
		if j == k {
			return
		}
		values[j]++
	}
}
func Combinations(n, k int) [][]int {
	result := [][]int{}
	ForEachCombination(n, k, func(values []int) bool { result = append(result, append([]int(nil), values...)); return true })
	return result
}
func listCombination(n, k int32) {
	counter := 0
	ForEachCombination(int(n), int(k), func(values []int) bool {
		counter++
		fmt.Printf("Comb-%3d : ", counter)
		for i := len(values) - 1; i >= 0; i-- {
			fmt.Printf("vC[%d] = %d ", i+1, values[i])
		}
		fmt.Println()
		return true
	})
}
func generateCombinatorialStabil(new, old *aig.Graph, candidates []int32, disjunctions *[]aig.Edge, n, k int32) int32 {
	if n < 0 || int(n) > len(candidates) {
		panic("live: candidate count mismatch")
	}
	count := int32(0)
	ForEachCombination(int(n), int(k), func(values []int) bool {
		signals := []aig.Edge{}
		for j := len(values) - 1; j >= 0; j-- {
			signals = append(signals, copiedDriver(old, int(candidates[values[j]])))
		}
		*disjunctions = append(*disjunctions, disjoin(new, signals))
		count++
		return true
	})
	return count
}
func generateCombinatorialStabilExhaust(new, old *aig.Graph, disjunctions *[]aig.Edge, n, k int32) int32 {
	if int(n) > registers(old) {
		panic("live: register candidate count")
	}
	count := int32(0)
	ForEachCombination(int(n), int(k), func(values []int) bool {
		signals := []aig.Edge{}
		for j := len(values) - 1; j >= 0; j-- {
			signals = append(signals, aig.Aig_ObjCopy(old, old.Inputs()[trueInputs(old)+values[j]]))
		}
		*disjunctions = append(*disjunctions, disjoin(new, signals))
		count++
		return true
	})
	return count
}
func copiedDriver(old *aig.Graph, index int) aig.Edge {
	driver := outputDriver(old, index)
	copy := aig.Aig_ObjCopy(old, aig.Aig_Regular(driver))
	if copy == aig.InvalidEdge {
		panic("live: output driver has no image")
	}
	return copy ^ aig.Edge(driver&1)
}
func disjunctiveTester(p *aig.Graph, candidates []aig.Edge, k int32) *aig.Graph {
	if k <= 0 || int(k) > len(candidates) {
		panic("live: invalid maximum combination width")
	}
	c := newSourceGraphCopy(p, "_nCk")
	c.copyInputs()
	groups := [][]int{}
	for width := 1; width <= int(k); width++ {
		ForEachCombination(len(candidates), width, func(values []int) bool { groups = append(groups, append([]int(nil), values...)); return true })
	}
	states := make([]aig.Edge, len(groups))
	for i := range states {
		states[i] = aig.Aig_ObjCreateCi(c.destination)
	}
	c.copySourceNodes()
	properties := make([]aig.Edge, len(groups))
	for i, group := range groups {
		signals := make([]aig.Edge, len(group))
		for j, index := range group {
			signals[j] = c.signal(candidates[index])
		}
		properties[i] = aig.Aig_Or(c.destination, disjoin(c.destination, signals).Not(), states[i])
	}
	for i := 0; i < propertyOutputs(p); i++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[i]))
	}
	for _, e := range properties {
		aig.Aig_ObjCreateCo(c.destination, e)
	}
	c.appendConstraints()
	for r := 0; r < registers(p); r++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[trueOutputs(p)+r]))
	}
	for _, e := range properties {
		aig.Aig_ObjCreateCo(c.destination, e)
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)+len(states)))
	c.setConstraints()
	return c.destination
}
func generateDisjunctiveTester(n *Network, p *aig.Graph, combN, combK int32) *aig.Graph {
	hints := findHintOutputs(n)
	signals := make([]aig.Edge, len(hints))
	for i, index := range hints {
		signals[i] = p.Outputs()[index]
	}

	return disjunctiveTester(p, signals, combK)
}
func generateGeneralDisjunctiveTester(n *Network, p *aig.Graph, k int32) *aig.Graph {
	return disjunctiveTester(p, p.Inputs()[trueInputs(p):], k)
}
