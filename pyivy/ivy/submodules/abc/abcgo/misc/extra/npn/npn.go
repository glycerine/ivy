// Package npn computes ABC's four-input NPN canonicalization tables.
// The enumeration order matches extraUtilMisc.c, including phase/permutation
// representatives and the numbering of all 222 equivalence classes.
package npn

// Permutations returns the 24 variable permutations in ABC's original order.
func Permutations() [24][4]byte {
	var result [24][4]byte
	array := [4]byte{0, 1, 2, 3}
	var fill func(int, int, int)
	fill = func(base, fact, n int) {
		if n == 1 {
			result[base][0] = array[0]
			return
		}
		next := fact / n
		last := n - 1
		for cur := 0; cur < n; cur++ {
			array[cur], array[last] = array[last], array[cur]
			offset := base + (n-1-cur)*next
			for k := 0; k < next; k++ {
				result[offset+k][last] = array[last]
			}
			fill(offset, next, n-1)
			array[cur], array[last] = array[last], array[cur]
		}
	}
	fill(0, 24, 4)
	return result
}

// Transform applies input negations, variable permutation, then optional output
// negation. Phase bits 0..3 negate inputs; bit 4 negates the output.
func Transform(truth uint16, phase byte, permutation [4]byte) uint16 {
	var variables byte
	for _, v := range permutation {
		if v >= 4 || variables&(1<<v) != 0 {
			panic("invalid four-variable permutation")
		}
		variables |= 1 << v
	}
	if phase&^31 != 0 {
		panic("invalid four-variable phase")
	}
	var result uint16
	for mint := 0; mint < 16; mint++ {
		if truth&(uint16(1)<<uint(mint^int(phase&15))) == 0 {
			continue
		}
		dest := 0
		for v := 0; v < 4; v++ {
			if mint&(1<<uint(v)) != 0 {
				dest |= 1 << permutation[v]
			}
		}
		result |= uint16(1) << uint(dest)
	}
	if phase&16 != 0 {
		result = ^result
	}
	return result
}

// Extra_Truth4VarNPN returns newly allocated canonical, phase, permutation-index
// and class-index tables, each indexed by an unsigned 16-bit truth table.
// Transform(canons[f], phases[f], Permutations()[perms[f]]) reconstructs f.
func Extra_Truth4VarNPN() (canons []uint16, phases, perms, classes []byte) {
	canons = make([]uint16, 65536)
	phases = make([]byte, 65536)
	perms = make([]byte, 65536)
	classes = make([]byte, 65536)
	permutations := Permutations()
	nClasses := 1
	for truth := 1; truth < 32768; truth++ {
		if canons[truth] != 0 {
			classes[truth] = classes[canons[truth]]
			classes[65535^truth] = classes[truth]
			continue
		}
		classes[truth] = byte(nClasses)
		nClasses++
		// Filling both outputs below makes the C complemented-input loop redundant.
		for phase := byte(0); phase < 16; phase++ {
			for k, permutation := range permutations {
				transformed := Transform(uint16(truth), phase, permutation)
				if canons[transformed] == 0 {
					canons[transformed] = uint16(truth)
					phases[transformed] = phase
					perms[transformed] = byte(k)
					opposite := ^transformed
					canons[opposite] = uint16(truth)
					phases[opposite] = phase | 16
					perms[opposite] = byte(k)
				} else if canons[transformed] != uint16(truth) {
					panic("overlapping NPN classes")
				}
			}
		}
	}
	// Populate class IDs for upper-half complements of canonical representatives
	// too. The C loop fills IDs lazily only for visited lower-half members.
	for truth := range classes {
		classes[truth] = classes[canons[truth]]
	}
	phases[65535] = 16
	if nClasses != 222 {
		panic("incorrect four-input NPN class count")
	}
	return
}
