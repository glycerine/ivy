package cov

import (
	"container/heap"
	"math/bits"
)

type esopQueueItem struct {
	state uint32
	cost  int
}
type esopPriorityQueue []esopQueueItem

func (q esopPriorityQueue) Len() int           { return len(q) }
func (q esopPriorityQueue) Less(i, j int) bool { return q[i].cost < q[j].cost }
func (q esopPriorityQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *esopPriorityQueue) Push(value any)    { *q = append(*q, value.(esopQueueItem)) }
func (q *esopPriorityQueue) Pop() any {
	old := *q
	last := old[len(old)-1]
	*q = old[:len(old)-1]
	return last
}

// Min_EsopMinimize computes an exact ESOP from the represented parity cover.
// Up to four variables it searches all cube combinations for minimum literal
// count then term count. Larger supports use the exact algebraic-normal-form
// transform, which is linear in the truth-table size and avoids exponential
// parity-state search.
func Min_EsopMinimize(p *Min_Man_t) {
	if p == nil || p.nVars < 0 || p.nVars > 16 {
		panic("cov: ESOP minimization supports 0..16 variables")
	}
	truth := esopTruth(p)
	var terms []minimizeCube
	if p.nVars <= 4 {
		terms = minimizeESOPExact(int(p.nVars), truth)
	} else {
		terms = algebraicNormalForm(int(p.nVars), truth)
	}
	Min_ManClean(p, p.nVars)
	for _, term := range terms {
		cube := Min_CubeAlloc(p)
		for variable := int32(0); variable < p.nVars; variable++ {
			bit := uint64(1) << variable
			if term.mask&bit == 0 {
				continue
			}
			if term.bits&bit != 0 {
				Min_CubeXorBit(cube, 2*variable)
			} else {
				Min_CubeXorBit(cube, 2*variable+1)
			}
		}
		cube.nLits = uint32(bits.OnesCount64(term.mask))
		Min_EsopAddCube(p, cube)
	}
}

func Min_EsopRewrite(p *Min_Man_t) { Min_EsopMinimize(p) }

type minimizeCube struct{ mask, bits uint64 }

func esopTruth(p *Min_Man_t) []bool {
	limit := 1 << p.nVars
	truth := make([]bool, limit)
	for _, head := range p.ppStore {
		for cube := head; cube != nil; cube = cube.pNext {
			for assignment := 0; assignment < limit; assignment++ {
				if cubeMatches(cube, uint64(assignment)) {
					truth[assignment] = !truth[assignment]
				}
			}
		}
	}
	return truth
}

func minimizeESOPExact(variables int, truth []bool) []minimizeCube {
	limit := 1 << variables
	stateCount := 1 << limit
	target := uint32(0)
	for minterm, value := range truth {
		if value {
			target |= 1 << minterm
		}
	}
	if target == 0 {
		return nil
	}
	candidates := make([]minimizeCube, 0, 3*3*3*3)
	signatures := make([]uint32, 0, cap(candidates))
	for mask := 0; mask < 1<<variables; mask++ {
		for value := 0; value < 1<<variables; value++ {
			if value&^mask != 0 {
				continue
			}
			var signature uint32
			for minterm := 0; minterm < limit; minterm++ {
				if (minterm^value)&mask == 0 {
					signature |= 1 << minterm
				}
			}
			candidates = append(candidates, minimizeCube{uint64(mask), uint64(value)})
			signatures = append(signatures, signature)
		}
	}
	const infinity = int(^uint(0) >> 2)
	distance := make([]int, stateCount)
	previous := make([]uint32, stateCount)
	via := make([]int, stateCount)
	for i := range distance {
		distance[i], via[i] = infinity, -1
	}
	distance[0] = 0
	queue := &esopPriorityQueue{{state: 0, cost: 0}}
	heap.Init(queue)
	for queue.Len() != 0 {
		item := heap.Pop(queue).(esopQueueItem)
		if item.cost != distance[item.state] {
			continue
		}
		if item.state == target {
			break
		}
		for i, signature := range signatures {
			next := item.state ^ signature
			weight := bits.OnesCount64(candidates[i].mask)*128 + 1
			if item.cost+weight < distance[next] {
				distance[next] = item.cost + weight
				previous[next], via[next] = item.state, i
				heap.Push(queue, esopQueueItem{next, distance[next]})
			}
		}
	}
	var result []minimizeCube
	for state := target; state != 0; state = previous[state] {
		index := via[state]
		if index < 0 {
			panic("cov: no ESOP representation found")
		}
		result = append(result, candidates[index])
	}
	return result
}

func algebraicNormalForm(variables int, truth []bool) []minimizeCube {
	coefficients := append([]bool(nil), truth...)
	for variable := 0; variable < variables; variable++ {
		bit := 1 << variable
		for mask := range coefficients {
			if mask&bit != 0 {
				coefficients[mask] = coefficients[mask] != coefficients[mask^bit]
			}
		}
	}
	result := make([]minimizeCube, 0)
	for monomial, enabled := range coefficients {
		if enabled {
			result = append(result, minimizeCube{uint64(monomial), uint64(monomial)})
		}
	}
	return result
}

// Min_EsopAddCubeInt toggles an identical term (XOR cancellation) or appends
// the term to its literal-count bucket.
func Min_EsopAddCubeInt(p *Min_Man_t, cube *Min_Cube_t) int32 {
	if p == nil || cube == nil || int(cube.nLits) >= len(p.ppStore) {
		panic("cov: invalid ESOP cube")
	}
	previous := &p.ppStore[cube.nLits]
	for current := *previous; current != nil; current = *previous {
		if Min_CubesAreEqual(cube, current) != 0 {
			*previous = current.pNext
			Min_CubeRecycle(p, current)
			Min_CubeRecycle(p, cube)
			p.nCubes--
			return 0
		}
		previous = &current.pNext
	}
	cube.pNext = p.ppStore[cube.nLits]
	p.ppStore[cube.nLits] = cube
	p.nCubes++
	return 0
}

func Min_EsopAddCube(p *Min_Man_t, cube *Min_Cube_t) {
	if cube == nil || Min_CubeCheck(cube) == 0 || int32(cube.nLits) != Min_CubeCountLits(cube) {
		panic("cov: invalid ESOP cube encoding")
	}
	Min_EsopAddCubeInt(p, cube)
}
