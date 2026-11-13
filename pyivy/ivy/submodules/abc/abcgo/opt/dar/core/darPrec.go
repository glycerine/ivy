package core
import "math"

func Dar_ArrayAlloc(columns, rows, size int) [][]byte {
	if columns <= 0 || rows <= 0 || size <= 0 || rows > math.MaxInt/size || columns > math.MaxInt/(rows*size) {
		panic("dar: invalid matrix dimensions")
	}
	data := make([]byte, columns*rows*size)
	result := make([][]byte, columns)
	for i := range result {
		result[i] = data[i*rows*size : (i+1)*rows*size]
	}
	return result
}
func Dar_Factorial(n int) int {
	if n < 0 {
		panic("dar: negative factorial")
	}
	result := 1
	for i := 2; i <= n; i++ {
		if result > math.MaxInt/i {
			panic("dar: factorial overflow")
		}
		result *= i
	}
	return result
}
func Dar_Permutations_rec(result [][]int, nFact, n int, values []int) {
	if n == 0 {
		return
	}
	if n == 1 {
		result[0][0] = values[0]
		return
	}
	part := nFact / n
	last := n - 1
	for current := 0; current < n; current++ {
		values[current], values[last] = values[last], values[current]
		next := result[(n-1-current)*part : (n-current)*part]
		for _, row := range next {
			row[last] = values[last]
		}
		Dar_Permutations_rec(next, part, n-1, values)
		values[current], values[last] = values[last], values[current]
	}
}
func Dar_Permutations(n int) [][]int {
	if n < 0 || n > 9 {
		panic("dar: permutation dimension")
	}
	count := Dar_Factorial(n)
	result := make([][]int, count)
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}
	for i := range result {
		result[i] = make([]int, n)
	}
	Dar_Permutations_rec(result, count, n, values)
	return result
}
func Dar_TruthPermute_int(minterms []int, count int, permutation []int, variables int, result []int) {
	if variables < 0 || variables > 5 || len(permutation) < variables {
		panic("dar: truth permutation dimensions")
	}
	for i, m := range minterms[:count] {
		mapped := 0
		for v := 0; v < variables; v++ {
			if permutation[v] < 0 || permutation[v] >= variables {
				panic("dar: invalid permutation")
			}
			if m&(1<<v) != 0 {
				mapped |= 1 << permutation[v]
			}
		}
		result[i] = mapped
	}
}
func Dar_TruthPermute(truth uint32, permutation []int, variables int, reverse bool) uint32 {
	if variables < 0 || variables > 5 || len(permutation) < variables {
		panic("dar: truth permutation dimensions")
	}
	seen := uint32(0)
	for _, v := range permutation[:variables] {
		if v < 0 || v >= variables || seen&(1<<v) != 0 {
			panic("dar: invalid permutation")
		}
		seen |= 1 << v
	}
	var result uint32
	for m := 0; m < 1<<variables; m++ {
		mapped := 0
		for v := 0; v < variables; v++ {
			if m&(1<<v) != 0 {
				mapped |= 1 << permutation[v]
			}
		}
		from, to := m, mapped
		if reverse {
			from, to = to, from
		}
		if truth&(1<<from) != 0 {
			result |= 1 << to
		}
	}
	return result
}
func Dar_TruthPolarize(truth uint32, polarity, variables int) uint32 {
	if variables < 0 || variables > 5 {
		panic("dar: truth polarization dimension")
	}
	masks := [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	for v := 0; v < variables; v++ {
		if polarity&(1<<v) != 0 {
			shift := uint(1 << v)
			truth = (truth&^masks[v])<<shift | (truth&masks[v])>>shift
		}
	}
	return truth
}

// NPNMatch describes f(x) = canonical(y) XOR OutputComplement, where
// y[i] = x[Permutation[i]] XOR the i-th bit of InputPhase.
type NPNMatch struct {
	Canonical        uint16
	Class            uint8
	Permutation      [4]int
	InputPhase       uint8
	OutputComplement bool
}
type CanonicalTable struct {
	canonical                 [65536]uint16
	phase, permutation, class [65536]uint8
	permutations              [][4]int
	representatives           []uint16
}

func (t *CanonicalTable) Lookup(truth uint16) NPNMatch {
	return NPNMatch{Canonical: t.canonical[truth], Class: t.class[truth], Permutation: t.permutations[t.permutation[truth]], InputPhase: t.phase[truth] & 15, OutputComplement: t.phase[truth]&16 != 0}
}
func (t *CanonicalTable) Representatives() []uint16 {
	return append([]uint16(nil), t.representatives...)
}

// Dar_Truth4VarNPN retains the source's deterministic representative, phase and
// permutation order. One pass over each unvisited orbit covers all 65,536
// functions; the second complement pass in C is redundant because both output
// phases are assigned together.
func Dar_Truth4VarNPN() *CanonicalTable {
	table := new(CanonicalTable)
	permutations := Dar_Permutations(4)
	for _, p := range permutations {
		table.permutations = append(table.permutations, [4]int(p))
	}
	table.representatives = []uint16{0}
	for canonical := uint32(1); canonical < 32768; canonical++ {
		if table.canonical[canonical] != 0 {
			continue
		}
		class := uint8(len(table.representatives))
		table.representatives = append(table.representatives, uint16(canonical))
		for phase := 0; phase < 16; phase++ {
			polarized := Dar_TruthPolarize(canonical, phase, 4)
			for index, p := range permutations {
				truth := uint16(Dar_TruthPermute(polarized, p, 4, false))
				if table.canonical[truth] != 0 {
					continue
				}
				table.canonical[truth] = uint16(canonical)
				table.phase[truth] = uint8(phase)
				table.permutation[truth] = uint8(index)
				table.class[truth] = class
				truth = ^truth
				table.canonical[truth] = uint16(canonical)
				table.phase[truth] = uint8(phase) | 16
				table.permutation[truth] = uint8(index)
				table.class[truth] = class
			}
		}
	}
	table.phase[65535] = 16
	if len(table.representatives) != 222 {
		panic("dar: NPN table must have 222 classes")
	}
	return table
}
