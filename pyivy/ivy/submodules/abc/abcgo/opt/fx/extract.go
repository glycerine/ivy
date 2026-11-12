package fx

import (
	"fmt"
	"slices"
)

// Options controls the classical cube divisor extraction.
type Options struct {
	MaxNewNodes   int32
	LiteralLimit  int32
	CanonicalOnly bool
}

// Extract repeatedly chooses the highest positive-weight divisor and rewrites
// its occurrences. It returns independently owned cubes and the number of new
// nodes. This implementation rebuilds the candidate catalog after each step;
// its priority tie-breaking and levels may differ from ABC's incremental queue.
func Extract(cubes [][]int32, nVars int32, options Options) ([][]int32, int32, error) {
	if options.MaxNewNodes < 0 || options.LiteralLimit < 0 {
		return nil, 0, fmt.Errorf("fx: negative extraction limit")
	}
	current := make([][]int32, len(cubes))
	for i := range cubes {
		current[i] = slices.Clone(cubes[i])
	}
	current = removeContained(current)
	var created int32
	// A simplification can remove pairs without creating a node, so the
	// source's iteration bound counts selected divisors rather than nodes.
	for iteration := int32(0); iteration < options.MaxNewNodes; iteration++ {
		catalog, contained, err := Catalog(current, nVars, options.LiteralLimit, options.CanonicalOnly)
		if err != nil {
			return nil, created, err
		}
		best := -1
		for i := range catalog {
			if catalog[i].Weight > 0 && (best < 0 || catalog[i].Weight > catalog[best].Weight) {
				best = i
			}
		}
		if best < 0 {
			break
		}
		next, added, err := ApplyDivisor(current, catalog[best].Literals, nVars)
		if err != nil {
			return nil, created, err
		}
		// Containment recorded before a rewrite may refer to removed or moved
		// cubes. Recompute it on the new covers before deleting redundant cubes.
		if len(contained) != 0 {
			_, redundant, problem := Catalog(next, nVars+boolInt(added), options.LiteralLimit, options.CanonicalOnly)
			if problem != nil {
				return nil, created, problem
			}
			if len(redundant) > 0 {
				marked := make([]bool, len(next))
				for _, i := range redundant {
					marked[i] = true
				}
				kept := make([][]int32, 0, len(next))
				for i, cube := range next {
					if !marked[i] {
						kept = append(kept, cube)
					}
				}
				next = kept
			}
		}
		current = next
		if added {
			created++
			nVars++
		}
	}
	return current, created, nil
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

func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
