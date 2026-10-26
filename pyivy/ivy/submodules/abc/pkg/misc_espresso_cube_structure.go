package abc

// Immutable cube-layout lifecycle and derived masks from src/misc/espresso/cubestr.c.

// NewEspressoCubeLayoutFromSource applies cube_setup's input rules: binary
// part sizes become two, and negative multiple-valued sizes mark symbolic
// fields while their absolute values determine storage.
func NewEspressoCubeLayoutFromSource(partSizes []int, binaryVariables int) (*EspressoCubeLayout, error) {
	if binaryVariables < 0 || binaryVariables > len(partSizes) {
		return nil, ErrDimension
	}
	normalized := make([]int, len(partSizes))
	symbolic := make([]bool, len(partSizes))
	for variable, size := range partSizes {
		if variable < binaryVariables {
			normalized[variable] = 2
			continue
		}
		if size < 0 {
			symbolic[variable] = true
			size = -size
		}
		if size <= 0 {
			return nil, ErrDimension
		}
		normalized[variable] = size
	}
	layout, err := NewEspressoCubeLayout(normalized, binaryVariables)
	if err != nil {
		return nil, err
	}
	copy(layout.symbolic, symbolic)
	return layout, nil
}

func (layout *EspressoCubeLayout) Clone() (*EspressoCubeLayout, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	result := &EspressoCubeLayout{
		partSizes: slicesCloneInts(layout.partSizes), firstParts: slicesCloneInts(layout.firstParts),
		symbolic: make([]bool, len(layout.symbolic)), binaryVariables: layout.binaryVariables, bitCount: layout.bitCount,
	}
	copy(result.symbolic, layout.symbolic)
	return result, nil
}

func (layout *EspressoCubeLayout) BinaryVariableCount() int {
	if layout == nil {
		return 0
	}
	return layout.binaryVariables
}

func (layout *EspressoCubeLayout) MultipleValuedVariableCount() int {
	if layout == nil {
		return 0
	}
	return len(layout.partSizes) - layout.binaryVariables
}

func (layout *EspressoCubeLayout) OutputVariable() int {
	if layout == nil || len(layout.partSizes) == layout.binaryVariables {
		return -1
	}
	return len(layout.partSizes) - 1
}

func (layout *EspressoCubeLayout) SymbolicVariables() ([]bool, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	result := make([]bool, len(layout.symbolic))
	copy(result, layout.symbolic)
	return result, nil
}

func (layout *EspressoCubeLayout) VariableRange(variable int) (first, end int, err error) {
	if err = layout.check(); err != nil {
		return 0, 0, err
	}
	if variable < 0 || variable >= len(layout.partSizes) {
		return 0, 0, ErrDimension
	}
	first = layout.firstParts[variable]
	return first, first + layout.partSizes[variable], nil
}

// VariableWordRange returns inclusive zero-based Go packed-word indexes. The C
// representation adds one header word, which EspressoSet stores as fields.
func (layout *EspressoCubeLayout) VariableWordRange(variable int) (first, last int, err error) {
	start, end, err := layout.VariableRange(variable)
	if err != nil {
		return 0, 0, err
	}
	return start >> 5, (end - 1) >> 5, nil
}

func (layout *EspressoCubeLayout) VariableMask(variable int) (*EspressoSet, error) {
	start, end, err := layout.VariableRange(variable)
	if err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(layout.bitCount)
	for part := start; part < end; part++ {
		_ = result.Set(part, true)
	}
	return result, nil
}

func (layout *EspressoCubeLayout) BinaryMask() (*EspressoSet, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(layout.bitCount)
	for part := 0; part < 2*layout.binaryVariables; part++ {
		_ = result.Set(part, true)
	}
	return result, nil
}

func (layout *EspressoCubeLayout) MultipleValuedMask() (*EspressoSet, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(layout.bitCount)
	for part := 2 * layout.binaryVariables; part < layout.bitCount; part++ {
		_ = result.Set(part, true)
	}
	return result, nil
}

func (layout *EspressoCubeLayout) EmptyCube() (*EspressoSet, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	return NewEspressoSet(layout.bitCount)
}

func (layout *EspressoCubeLayout) DefaultSparseVariables() ([]bool, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	result := make([]bool, len(layout.partSizes))
	for variable := layout.binaryVariables; variable < len(result); variable++ {
		result[variable] = true
	}
	return result, nil
}

// BinaryInputWordMask is cube.inword/inmask in zero-based EspressoSet words.
func (layout *EspressoCubeLayout) BinaryInputWordMask() (word int, mask uint32, err error) {
	if err = layout.check(); err != nil {
		return -1, 0, err
	}
	if layout.binaryVariables == 0 {
		return -1, 0, nil
	}
	word = (2*layout.binaryVariables - 1) >> 5
	binary, _ := layout.BinaryMask()
	return word, binary.words[word] & 0x55555555, nil
}
