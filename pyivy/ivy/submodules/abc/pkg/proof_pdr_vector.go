package abc

// Compact invariant vectors from src/proof/pdr/pdrInv.c:
// Pdr_ManDeriveInfinityClauses and Pdr_ForEachCube.

// ABCVector writes [cube count, cube size, literals..., ..., register count].
// Literals encode excluded-state cubes; importing/exporting is not proof
// verification. maxEntries zero selects one million signed 32-bit entries.
func (invariant *PdrInvariant) ABCVector(registers, maxEntries int) ([]int32, error) {
	if invariant == nil || registers < 0 || uint64(registers) > 1<<30 || maxEntries < 0 {
		return nil, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if maxEntries > 1<<31-1 {
		maxEntries = 1<<31 - 1
	}
	total := 2
	for _, cube := range invariant.clauses {
		if err := cube.Check(); err != nil {
			return nil, err
		}
		for _, l := range cube.state {
			if uint64(l.Index()) >= uint64(registers) {
				return nil, ErrDimension
			}
		}
		var err error
		total, err = checkedSize(total, 1, len(cube.state)+1, maxEntries)
		if err != nil {
			return nil, err
		}
	}
	if total > maxEntries {
		return nil, ErrLimit
	}
	result := make([]int32, 0, total)
	result = append(result, int32(len(invariant.clauses)))
	for _, cube := range invariant.clauses {
		result = append(result, int32(len(cube.state)))
		for _, l := range cube.state {
			result = append(result, int32(l))
		}
	}
	return append(result, int32(registers)), nil
}

// ReadPdrABCVector checks every count, boundary and register reference before
// allocating cubes. The returned invariant owns its data. Duplicate equal
// literals are canonicalized; contradictory literals are rejected.
func ReadPdrABCVector(data []int32, maxEntries int) (*PdrInvariant, int, error) {
	if maxEntries < 0 {
		return nil, 0, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if maxEntries > 1<<31-1 {
		maxEntries = 1<<31 - 1
	}
	if len(data) > maxEntries {
		return nil, 0, ErrLimit
	}
	if len(data) < 2 || data[0] < 0 {
		return nil, 0, ErrDimension
	}
	registers := int(data[len(data)-1])
	if registers < 0 || uint64(registers) > 1<<30 {
		return nil, 0, ErrDimension
	}
	count := int(data[0])
	if count > len(data)-2 {
		return nil, 0, ErrDimension
	}
	// Validate the entire envelope before allocating a count-sized slice.
	position := 1
	for i := 0; i < count; i++ {
		if position >= len(data)-1 || data[position] < 0 {
			return nil, 0, ErrDimension
		}
		size := int(data[position])
		position++
		if size > len(data)-1-position {
			return nil, 0, ErrDimension
		}
		for _, raw := range data[position : position+size] {
			if raw < 0 || uint64(uint32(raw)>>1) >= uint64(registers) {
				return nil, 0, ErrInvalidLiteral
			}
		}
		position += size
	}
	if position != len(data)-1 {
		return nil, 0, ErrDimension
	}
	cubes := make([]*PdrCube, 0, count)
	position = 1
	for i := 0; i < count; i++ {
		size := int(data[position])
		position++
		literals := make([]PdrLit, size)
		for j := range literals {
			literals[j] = PdrLit(data[position+j])
		}
		cube, err := NewPdrCube(literals, nil)
		if err != nil {
			return nil, 0, err
		}
		cubes = append(cubes, cube)
		position += size
	}
	// Cubes already own their slices; avoid a second complete allocation.
	return &PdrInvariant{clauses: cubes}, registers, nil
}
