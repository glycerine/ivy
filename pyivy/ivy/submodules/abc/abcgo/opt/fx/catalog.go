package fx

import (
	"encoding/binary"
	"fmt"
)

// Divisor is a candidate with ABC's literal-saving priority score.
type Divisor struct {
	Literals []int32
	Weight   float32
}

// Catalog constructs the initial single- and two-cube divisor candidates.
// Cubes must be grouped by owner and have sorted, distinct literals. nVars is
// the first available new node number. Contained records are cube indices
// identified for subsequent single-cube containment removal.
func Catalog(cubes [][]int32, nVars, literalCountMax int32, canonical bool) (divisors []Divisor, contained []int, err error) {
	if nVars < 0 {
		return nil, nil, fmt.Errorf("fx: negative variable count")
	}
	levels := make([]int32, nVars)
	lastOwner, first := int32(-1), 0
	for i, cube := range cubes {
		if len(cube) == 0 || cube[0] < 0 || cube[0] >= nVars || cube[0] < lastOwner {
			return nil, nil, fmt.Errorf("fx: invalid cube owner at %d", i)
		}
		if cube[0] != lastOwner {
			if lastOwner >= 0 {
				levels[lastOwner] += int32(i - first)
			}
			first, lastOwner = i, cube[0]
		}
		for j, lit := range cube[1:] {
			if lit < 0 || lit>>1 >= nVars || j > 0 && lit <= cube[j] {
				return nil, nil, fmt.Errorf("fx: invalid literal in cube %d", i)
			}
			levels[cube[0]] = max(levels[cube[0]], levels[lit>>1])
		}
	}
	index := make(map[string]int)
	key := func(lits []int32) string {
		bytes := make([]byte, 4*len(lits))
		for i, lit := range lits {
			binary.LittleEndian.PutUint32(bytes[4*i:], uint32(lit))
		}
		return string(bytes)
	}
	add := func(lits []int32, single bool, gain float32) {
		k := key(lits)
		id, exists := index[k]
		if !exists {
			var level int32
			for _, lit := range lits {
				level = max(level, levels[lit>>2])
			}
			level = min(level, 800)
			initial := float32(-float64(len(lits)) + 0.9 - 0.0009*float64(level))
			if single {
				initial = float32(-2 + 0.9 - 0.001*float64(level))
			}
			id = len(divisors)
			index[k] = id
			divisors = append(divisors, Divisor{append([]int32(nil), lits...), initial})
		}
		divisors[id].Weight += gain
	}
	for _, cube := range cubes {
		for i := 1; i < len(cube); i++ {
			for j := i + 1; j < len(cube); j++ {
				add([]int32{(cube[i] ^ 1) << 1, (cube[j]^1)<<1 | 1}, true, 1)
			}
		}
	}
	for i, left := range cubes {
		for j := i + 1; j < len(cubes) && cubes[j][0] == left[0]; j++ {
			div, common, problem := CubeFreeDivisor(cubes[j], left)
			if problem != nil {
				return nil, nil, problem
			}
			if common == -1 {
				if len(cubes[j]) > len(left) {
					contained = append(contained, j)
				} else {
					contained = append(contained, i)
				}
				continue
			}
			if len(div) == 4 && NormalizeDivisor(div) < 0 && canonical {
				continue
			}
			if literalCountMax > 0 && int32(len(div)) > literalCountMax || canonical && len(div) == 3 {
				continue
			}
			add(div, false, float32(common+len(div)-1))
		}
	}
	return divisors, contained, nil
}
