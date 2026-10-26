package abc

// Variable-aware packed-cube operations from src/misc/espresso/setc.c.

type EspressoCubeLayout struct {
	partSizes       []int
	firstParts      []int
	symbolic        []bool
	binaryVariables int
	bitCount        int
}

func NewEspressoCubeLayout(partSizes []int, binaryVariables int) (*EspressoCubeLayout, error) {
	if binaryVariables < 0 || binaryVariables > len(partSizes) {
		return nil, ErrDimension
	}
	if len(partSizes) > 1<<20 {
		return nil, ErrLimit
	}
	layout := &EspressoCubeLayout{
		partSizes: slicesCloneInts(partSizes), firstParts: make([]int, len(partSizes)),
		symbolic: make([]bool, len(partSizes)), binaryVariables: binaryVariables,
	}
	for variable, size := range partSizes {
		if size <= 0 || variable < binaryVariables && size != 2 {
			return nil, ErrDimension
		}
		if size > espressoSetMaximumBits-layout.bitCount {
			return nil, ErrLimit
		}
		layout.firstParts[variable] = layout.bitCount
		layout.bitCount += size
	}
	return layout, nil
}

func slicesCloneInts(values []int) []int {
	result := make([]int, len(values))
	copy(result, values)
	return result
}

func (layout *EspressoCubeLayout) check() error {
	if layout == nil || layout.binaryVariables < 0 || layout.binaryVariables > len(layout.partSizes) || len(layout.firstParts) != len(layout.partSizes) || len(layout.symbolic) != len(layout.partSizes) || layout.bitCount < 0 || layout.bitCount > espressoSetMaximumBits {
		return ErrInvariant
	}
	position := 0
	for variable, size := range layout.partSizes {
		if size <= 0 || variable < layout.binaryVariables && (size != 2 || layout.symbolic[variable]) || layout.firstParts[variable] != position {
			return ErrInvariant
		}
		position += size
	}
	if position != layout.bitCount {
		return ErrInvariant
	}
	return nil
}

func (layout *EspressoCubeLayout) BitCount() int {
	if layout == nil {
		return 0
	}
	return layout.bitCount
}

func (layout *EspressoCubeLayout) VariableCount() int {
	if layout == nil {
		return 0
	}
	return len(layout.partSizes)
}

func (layout *EspressoCubeLayout) PartSizes() []int {
	if layout == nil {
		return nil
	}
	return slicesCloneInts(layout.partSizes)
}

func (layout *EspressoCubeLayout) checkSet(set *EspressoSet) error {
	if err := layout.check(); err != nil {
		return err
	}
	if err := set.check(); err != nil {
		return err
	}
	if set.bits != layout.bitCount {
		return ErrDimension
	}
	return nil
}

func (layout *EspressoCubeLayout) variableIntersection(first, second *EspressoSet, variable int) bool {
	start := layout.firstParts[variable]
	end := start + layout.partSizes[variable]
	for part := start; part < end; part++ {
		mask := uint32(1) << uint(part&31)
		if first.words[part>>5]&second.words[part>>5]&mask != 0 {
			return true
		}
	}
	return false
}

func (layout *EspressoCubeLayout) variableFull(set *EspressoSet, variable int) bool {
	start := layout.firstParts[variable]
	end := start + layout.partSizes[variable]
	for part := start; part < end; part++ {
		if set.words[part>>5]&(uint32(1)<<uint(part&31)) == 0 {
			return false
		}
	}
	return true
}

func (layout *EspressoCubeLayout) FullCube() (*EspressoSet, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	return NewFullEspressoSet(layout.bitCount)
}

func (layout *EspressoCubeLayout) FullRow(cube, cofactor *EspressoSet) (bool, error) {
	if err := layout.checkSet(cube); err != nil {
		return false, err
	}
	if err := layout.checkSet(cofactor); err != nil {
		return false, err
	}
	for index, word := range cube.words {
		want := ^uint32(0)
		if index == len(cube.words)-1 && layout.bitCount&31 != 0 {
			want = uint32(1)<<uint(layout.bitCount&31) - 1
		}
		if layout.bitCount == 0 {
			want = 0
		}
		if word|cofactor.words[index] != want {
			return false, nil
		}
	}
	return true, nil
}

func (layout *EspressoCubeLayout) Distance(first, second *EspressoSet) (int, error) {
	if err := layout.checkSet(first); err != nil {
		return 0, err
	}
	if err := layout.checkSet(second); err != nil {
		return 0, err
	}
	distance := 0
	for variable := range layout.partSizes {
		if !layout.variableIntersection(first, second, variable) {
			distance++
		}
	}
	return distance, nil
}

func (layout *EspressoCubeLayout) DistanceZero(first, second *EspressoSet) (bool, error) {
	distance, err := layout.Distance(first, second)
	return distance == 0, err
}

func (layout *EspressoCubeLayout) DistanceAtMostOne(first, second *EspressoSet) (int, error) {
	if err := layout.checkSet(first); err != nil {
		return 0, err
	}
	if err := layout.checkSet(second); err != nil {
		return 0, err
	}
	distance := 0
	for variable := range layout.partSizes {
		if !layout.variableIntersection(first, second, variable) {
			distance++
			if distance > 1 {
				return 2, nil
			}
		}
	}
	return distance, nil
}

// ForceLower returns a clone of lower with first's parts added for every
// variable whose first and second parts are disjoint.
func (layout *EspressoCubeLayout) ForceLower(lower, first, second *EspressoSet) (*EspressoSet, error) {
	if err := layout.checkSet(lower); err != nil {
		return nil, err
	}
	if err := layout.checkSet(first); err != nil {
		return nil, err
	}
	if err := layout.checkSet(second); err != nil {
		return nil, err
	}
	result, _ := lower.Clone()
	for variable := range layout.partSizes {
		if layout.variableIntersection(first, second, variable) {
			continue
		}
		start := layout.firstParts[variable]
		for part := start; part < start+layout.partSizes[variable]; part++ {
			mask := uint32(1) << uint(part&31)
			result.words[part>>5] |= first.words[part>>5] & mask
		}
	}
	return result, nil
}

// Consensus computes per-variable intersection, substituting union whenever
// the intersection for that variable is empty.
func (layout *EspressoCubeLayout) Consensus(first, second *EspressoSet) (*EspressoSet, error) {
	if err := layout.checkSet(first); err != nil {
		return nil, err
	}
	if err := layout.checkSet(second); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(layout.bitCount)
	for variable := range layout.partSizes {
		intersects := layout.variableIntersection(first, second, variable)
		start := layout.firstParts[variable]
		for part := start; part < start+layout.partSizes[variable]; part++ {
			mask := uint32(1) << uint(part&31)
			word := first.words[part>>5] & second.words[part>>5]
			if !intersects {
				word = first.words[part>>5] | second.words[part>>5]
			}
			result.words[part>>5] |= word & mask
		}
	}
	return result, nil
}

// ActiveVariable returns the sole variable that is not full, or -1 when no
// variable or more than one variable is active.
func (layout *EspressoCubeLayout) ActiveVariable(cube *EspressoSet) (int, error) {
	if err := layout.checkSet(cube); err != nil {
		return -1, err
	}
	active := -1
	for variable := range layout.partSizes {
		if layout.variableFull(cube, variable) {
			continue
		}
		if active >= 0 {
			return -1, nil
		}
		active = variable
	}
	return active, nil
}

func (layout *EspressoCubeLayout) CommonActive(first, second, cofactor *EspressoSet) (bool, error) {
	if err := layout.checkSet(first); err != nil {
		return false, err
	}
	if err := layout.checkSet(second); err != nil {
		return false, err
	}
	if err := layout.checkSet(cofactor); err != nil {
		return false, err
	}
	for variable := range layout.partSizes {
		firstFull, secondFull := true, true
		start := layout.firstParts[variable]
		for part := start; part < start+layout.partSizes[variable]; part++ {
			mask := uint32(1) << uint(part&31)
			if first.words[part>>5]&mask == 0 && cofactor.words[part>>5]&mask == 0 {
				firstFull = false
			}
			if second.words[part>>5]&mask == 0 && cofactor.words[part>>5]&mask == 0 {
				secondFull = false
			}
		}
		if !firstFull && !secondFull {
			return true, nil
		}
	}
	return false, nil
}

func espressoCompareWords(first, second *EspressoSet, descending bool, mask *EspressoSet) (int, error) {
	if err := espressoSetPair(first, second); err != nil {
		return 0, err
	}
	if mask != nil {
		if err := espressoSetPair(first, mask); err != nil {
			return 0, err
		}
	}
	for index := len(first.words) - 1; index >= 0; index-- {
		firstWord, secondWord := first.words[index], second.words[index]
		if mask != nil {
			firstWord |= mask.words[index]
			secondWord |= mask.words[index]
		}
		if firstWord == secondWord {
			continue
		}
		if descending == (firstWord > secondWord) {
			return -1, nil
		}
		return 1, nil
	}
	return 0, nil
}

func EspressoSetCompareDescending(first, second *EspressoSet) (int, error) {
	if err := espressoSetPair(first, second); err != nil {
		return 0, err
	}
	if first.Rank > second.Rank {
		return -1, nil
	}
	if first.Rank < second.Rank {
		return 1, nil
	}
	return espressoCompareWords(first, second, true, nil)
}

func EspressoSetCompareAscending(first, second *EspressoSet) (int, error) {
	if err := espressoSetPair(first, second); err != nil {
		return 0, err
	}
	if first.Rank < second.Rank {
		return -1, nil
	}
	if first.Rank > second.Rank {
		return 1, nil
	}
	return espressoCompareWords(first, second, false, nil)
}

func EspressoSetCompareLexical(first, second *EspressoSet) (int, error) {
	return espressoCompareWords(first, second, true, nil)
}

func EspressoSetCompareDistanceOne(first, second, mask *EspressoSet) (int, error) {
	return espressoCompareWords(first, second, true, mask)
}

func EspressoSetCompareDescendingNullable(first, second *EspressoSet) (int, error) {
	if first == nil {
		if second == nil {
			return 0, nil
		}
		return 1, nil
	}
	if second == nil {
		return -1, nil
	}
	return EspressoSetCompareDescending(first, second)
}

func (layout *EspressoCubeLayout) Validate() error { return layout.check() }
