// Package minimize provides an exact, bounded SOP minimizer for small covers.
// It is the native Go core for the legacy map/cov minimizer and deliberately
// uses value types instead of the C implementation's mutable cube pool.
package minimize

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"
)

// Cube represents a product term. A set bit in Mask means that the matching
// bit in Bits is required; all other variables are don't-cares.
type Cube struct {
	Mask uint64
	Bits uint64
}

// Matches reports whether minterm is covered by c.
func (c Cube) Matches(minterm uint64) bool {
	return (minterm^c.Bits)&c.Mask == 0
}

// Minimize returns a minimum-literal, then minimum-cube SOP for on. dontCare
// values may be used to enlarge implicants but are not required to be covered.
// Exact minimization is exponential, so this implementation bounds inputs to
// sixteen variables, 256 on-set minterms, and 512 prime implicants.
func Minimize(variables int, on, dontCare []uint64) ([]Cube, error) {
	if variables < 0 || variables > 16 {
		return nil, fmt.Errorf("minimize: variable count %d outside [0,16]", variables)
	}
	limit := uint64(1) << variables
	seen := make(map[uint64]struct{}, len(on)+len(dontCare))
	for _, values := range [][]uint64{on, dontCare} {
		for _, m := range values {
			if m >= limit {
				return nil, fmt.Errorf("minimize: minterm %d exceeds %d variables", m, variables)
			}
			seen[m] = struct{}{}
		}
	}
	if len(on) == 0 {
		return []Cube{}, nil
	}
	if len(on) > 256 {
		return nil, fmt.Errorf("minimize: on-set has %d minterms (limit 256)", len(on))
	}
	onSet := make(map[uint64]struct{}, len(on))
	for _, m := range on {
		onSet[m] = struct{}{}
	}
	all := make([]uint64, 0, len(seen))
	for m := range seen {
		all = append(all, m)
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })

	current := make([]Cube, len(all))
	fullMask := limit - 1
	for i, m := range all {
		current[i] = Cube{Mask: fullMask, Bits: m}
	}
	primeSet := make(map[Cube]struct{})
	for len(current) > 0 {
		combined := make([]bool, len(current))
		nextSet := make(map[Cube]struct{})
		for i := range current {
			for j := i + 1; j < len(current); j++ {
				a, b := current[i], current[j]
				if a.Mask != b.Mask {
					continue
				}
				diff := (a.Bits ^ b.Bits) & a.Mask
				if bits.OnesCount64(diff) != 1 {
					continue
				}
				combined[i], combined[j] = true, true
				next := Cube{Mask: a.Mask &^ diff, Bits: a.Bits &^ diff}
				nextSet[next] = struct{}{}
			}
		}
		for i, c := range current {
			if combined[i] {
				continue
			}
			for m := range onSet {
				if c.Matches(m) {
					primeSet[c] = struct{}{}
					break
				}
			}
		}
		current = current[:0]
		for c := range nextSet {
			current = append(current, c)
		}
	}
	primes := make([]Cube, 0, len(primeSet))
	for c := range primeSet {
		primes = append(primes, c)
	}
	if len(primes) > 512 {
		return nil, fmt.Errorf("minimize: generated %d prime implicants (limit 512)", len(primes))
	}
	sort.Slice(primes, func(i, j int) bool {
		if bits.OnesCount64(primes[i].Mask) != bits.OnesCount64(primes[j].Mask) {
			return bits.OnesCount64(primes[i].Mask) < bits.OnesCount64(primes[j].Mask)
		}
		if primes[i].Mask != primes[j].Mask {
			return primes[i].Mask < primes[j].Mask
		}
		return primes[i].Bits < primes[j].Bits
	})

	coverers := make(map[uint64][]int, len(onSet))
	for m := range onSet {
		for i, c := range primes {
			if c.Matches(m) {
				coverers[m] = append(coverers[m], i)
			}
		}
	}
	bestLits, bestCount := int(^uint(0)>>1), int(^uint(0)>>1)
	var best []int
	selected := make([]bool, len(primes))
	covered := make(map[uint64]bool, len(onSet))
	chosen := make([]int, 0, len(on))
	var search func(lits int)
	search = func(lits int) {
		if lits > bestLits || lits == bestLits && len(chosen) >= bestCount {
			return
		}
		var pivot uint64
		fewest := int(^uint(0) >> 1)
		found := false
		for m := range onSet {
			if covered[m] {
				continue
			}
			options := 0
			for _, idx := range coverers[m] {
				if !selected[idx] {
					options++
				}
			}
			if options < fewest {
				pivot, fewest, found = m, options, true
			}
		}
		if !found {
			if lits < bestLits || lits == bestLits && len(chosen) < bestCount {
				bestLits, bestCount = lits, len(chosen)
				best = append(best[:0], chosen...)
			}
			return
		}
		for _, idx := range coverers[pivot] {
			if selected[idx] {
				continue
			}
			selected[idx] = true
			newly := make([]uint64, 0, len(on))
			for m := range onSet {
				if !covered[m] && primes[idx].Matches(m) {
					covered[m] = true
					newly = append(newly, m)
				}
			}
			chosen = append(chosen, idx)
			search(lits + bits.OnesCount64(primes[idx].Mask))
			chosen = chosen[:len(chosen)-1]
			for _, m := range newly {
				delete(covered, m)
			}
			selected[idx] = false
		}
	}
	search(0)
	if best == nil {
		return nil, errors.New("minimize: no cover found")
	}
	result := make([]Cube, len(best))
	for i, idx := range best {
		result[i] = primes[idx]
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Mask != result[j].Mask {
			return result[i].Mask < result[j].Mask
		}
		return result[i].Bits < result[j].Bits
	})
	return result, nil
}
