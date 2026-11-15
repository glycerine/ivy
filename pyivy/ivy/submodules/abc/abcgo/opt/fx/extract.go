package fx

import (
	"fmt"
	"slices"

	"github.com/glycerine/abcgo/misc/vec"
)

// Options controls the classical cube divisor extraction.
type Options struct {
	MaxNewNodes   int32
	LiteralLimit  int32
	CanonicalOnly bool
}

// Extract runs abcFx.c's incremental fast-extract matrix and priority queue.
// It returns independently owned cubes and the number of newly defined nodes.
// MaxNewNodes bounds selected divisors, including simplifications which do not
// introduce a node, as in the original C loop. The existing Go API also
// removes duplicate/absorbed input cubes before starting the C engine; native
// Fx_FastExtract can assert on such redundant covers. Irredundant inputs pass
// through unchanged.
func Extract(cubes [][]int32, nVars int32, options Options) ([][]int32, int32, error) {
	if options.MaxNewNodes < 0 || options.LiteralLimit < 0 || nVars < 0 {
		return nil, 0, fmt.Errorf("fx: negative extraction limit or variable count")
	}
	if len(cubes) == 0 {
		return nil, 0, nil
	}
	p := &extraction{options: options, ids: make(map[string]int), first: make([]int, nVars), levels: make([]int32, nVars), lits: make([][]int, 2*nVars), marked: make([]bool, len(cubes))}
	for i := range p.first {
		p.first[i] = -1
	}
	lastOwner := int32(-1)
	for i, cube := range cubes {
		if len(cube) == 0 || cube[0] < 0 || cube[0] >= nVars || cube[0] < lastOwner {
			return nil, 0, fmt.Errorf("fx: invalid cube owner at %d", i)
		}
		for j, lit := range cube[1:] {
			if lit < 0 || lit>>1 >= nVars || j > 0 && lit <= cube[j] {
				return nil, 0, fmt.Errorf("fx: invalid literal in cube %d", i)
			}
		}
		p.cubes = append(p.cubes, slices.Clone(cube))
		lastOwner = cube[0]
	}
	p.cubes = removeContained(p.cubes)
	for i, cube := range p.cubes {
		if p.first[cube[0]] < 0 {
			p.first[cube[0]] = i
		}
		for _, lit := range cube[1:] {
			p.lits[lit] = append(p.lits[lit], i)
		}
	}
	owner, first := p.cubes[0][0], 0
	for i, cube := range p.cubes {
		if cube[0] != owner {
			p.levels[owner] += int32(i - first)
			owner, first = cube[0], i
		}
		p.levels[owner] = max(p.levels[owner], p.level(cube))
	}
	for i := range p.cubes {
		p.singles(i, false, false)
	}
	for i := range p.cubes {
		if err := p.doubles(i+1, i, false, false); err != nil {
			return nil, 0, err
		}
	}
	p.queue = vec.Vec_QueAlloc(len(p.weights))
	vec.Vec_QueSetPriority(p.queue, &p.weights)
	defer vec.Vec_QueFree(p.queue)
	for id, weight := range p.weights {
		if weight > 0 {
			vec.Vec_QuePush(p.queue, id)
		}
	}
	for i := int32(0); i < options.MaxNewNodes && vec.Vec_QueTopPriority(p.queue) > 0; i++ {
		if err := p.update(vec.Vec_QuePop(p.queue)); err != nil {
			return nil, 0, err
		}
	}
	p.cubes = slices.DeleteFunc(p.cubes, func(cube []int32) bool { return len(cube) == 0 })
	return p.cubes, int32(len(p.first)) - nVars, nil
}

// removeContained keeps a single copy of each cube and removes terms that
// are absorbed by a smaller cube of the same owner.
func removeContained(cubes [][]int32) [][]int32 {
	removed := make([]bool, len(cubes))
	for start := 0; start < len(cubes); {
		if len(cubes[start]) == 0 {
			start++ // Catalog reports malformed input.
			continue
		}
		end := start + 1
		for end < len(cubes) && len(cubes[end]) > 0 && cubes[end][0] == cubes[start][0] {
			end++
		}
		for i := start; i < end; i++ {
			a := cubes[i]
			for j := start; j < end; j++ {
				b := cubes[j]
				if i == j || len(b) > len(a) || len(a) == len(b) && j > i {
					continue
				}
				if isLiteralSubset(b[1:], a[1:]) {
					removed[i] = true
					break
				}
			}
		}
		start = end
	}
	result := make([][]int32, 0, len(cubes))
	for i, cube := range cubes {
		if !removed[i] {
			result = append(result, cube)
		}
	}
	return result
}

func isLiteralSubset(a, b []int32) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			i++
		}
		j++
	}
	return i == len(a)
}
