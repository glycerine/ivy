package core

import (
	"math/rand"
	"sort"
)

type geneticOrder struct {
	order []int
	cost  int
}

// Cudd_GeneticReorder applies a bounded PMX genetic search to a variable
// interval and installs the best order found.
func Cudd_GeneticReorder(dd *DdManager, lower, upper int) bool {
	if dd == nil || lower < 0 || upper < lower || upper >= len(dd.InvPerm) {
		return false
	}
	width := upper - lower + 1
	if width < 2 {
		return true
	}
	base := append([]int(nil), dd.InvPerm...)
	rng := rand.New(rand.NewSource(1))
	populationSize := 3 * width
	if populationSize > 120 {
		populationSize = 120
	}
	targetSize := 1
	for i := 2; i <= width && targetSize < populationSize; i++ {
		targetSize *= i
	}
	if targetSize > populationSize {
		targetSize = populationSize
	}
	population := make([]geneticOrder, 0, targetSize)
	seen := make(map[string]bool)
	add := func(order []int) {
		key := orderKey(order[lower : upper+1])
		if seen[key] {
			return
		}
		if Cudd_ShuffleHeap(dd, order) == 0 {
			return
		}
		seen[key] = true
		population = append(population, geneticOrder{append([]int(nil), order...), len(dd.unique)})
	}
	add(base)
	for len(population) < targetSize {
		candidate := append([]int(nil), base...)
		rng.Shuffle(width, func(i, j int) { candidate[lower+i], candidate[lower+j] = candidate[lower+j], candidate[lower+i] })
		add(candidate)
	}
	for generation := 0; generation < 50; generation++ {
		parent1 := population[tournament(population, rng)].order
		parent2 := population[tournament(population, rng)].order
		child := pmxOrder(parent1, parent2, lower, upper, rng)
		add(child)
		sort.Slice(population, func(i, j int) bool { return population[i].cost < population[j].cost })
		if len(population) > targetSize {
			removed := population[len(population)-1]
			delete(seen, orderKey(removed.order[lower:upper+1]))
			population = population[:len(population)-1]
		}
	}
	if len(population) == 0 {
		Cudd_ShuffleHeap(dd, base)
		return false
	}
	sort.Slice(population, func(i, j int) bool { return population[i].cost < population[j].cost })
	return Cudd_ShuffleHeap(dd, population[0].order) != 0
}

func tournament(population []geneticOrder, rng *rand.Rand) int {
	a, b := rng.Intn(len(population)), rng.Intn(len(population))
	if population[a].cost <= population[b].cost {
		return a
	}
	return b
}

func pmxOrder(parent1, parent2 []int, lower, upper int, rng *rand.Rand) []int {
	child := append([]int(nil), parent1...)
	a, b := rng.Intn(upper-lower+1), rng.Intn(upper-lower+1)
	if a > b {
		a, b = b, a
	}
	a, b = a+lower, b+lower
	for i := a; i <= b; i++ {
		child[i] = parent2[i]
	}
	used := make(map[int]bool, b-a+1)
	for i := a; i <= b; i++ {
		used[child[i]] = true
	}
	for i := lower; i <= upper; i++ {
		if i >= a && i <= b {
			continue
		}
		candidate := parent1[i]
		for used[candidate] {
			mapped := false
			for j := a; j <= b; j++ {
				if parent2[j] == candidate {
					candidate = parent1[j]
					mapped = true
					break
				}
			}
			if !mapped {
				break
			}
		}
		child[i] = candidate
		used[candidate] = true
	}
	return child
}

func orderKey(order []int) string {
	bytes := make([]byte, len(order)*4)
	for i, value := range order {
		bytes[i*4] = byte(value)
		bytes[i*4+1] = byte(value >> 8)
		bytes[i*4+2] = byte(value >> 16)
		bytes[i*4+3] = byte(value >> 24)
	}
	return string(bytes)
}
