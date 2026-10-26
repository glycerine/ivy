package abc

// SOP and PLA operations from src/bool/kit/kitSop.c and kitPla.c.

import (
	"slices"
	"strings"
)

// KitCoverOptions bounds owned cover construction and iterative kernel work.
// Zero limits select conservative defaults.
type KitCoverOptions struct {
	MaxCubes, MaxCubeEntries, MaxSteps int
	Stop                               func() bool
}

func checkKitCoverOptions(o KitCoverOptions) (KitCoverOptions, error) {
	if o.MaxCubes < 0 || o.MaxCubeEntries < 0 || o.MaxSteps < 0 {
		return o, ErrDimension
	}
	if o.MaxCubes == 0 {
		o.MaxCubes = 1 << 20
	}
	if o.MaxCubeEntries == 0 {
		o.MaxCubeEntries = 16 << 20
	}
	if o.MaxSteps == 0 {
		o.MaxSteps = 16 << 20
	}
	return o, nil
}

type kitCoverBudget struct {
	options KitCoverOptions
	steps   int
}

func (b *kitCoverBudget) charge(count int) error {
	if count < 0 {
		return ErrDimension
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	next, err := checkedSize(b.steps, count, 1, b.options.MaxSteps)
	if err != nil {
		return err
	}
	b.steps = next
	return nil
}

// KitSOPCover owns literal-bit cubes. Bit i means literal i is present. This
// is the representation used by kitSop.c after its two-bit ISOP conversion.
type KitSOPCover struct {
	Cubes []uint32
}

func NewKitSOPCover(cubes []uint32, literals int, o KitCoverOptions) (KitSOPCover, error) {
	o, err := checkKitCoverOptions(o)
	if err != nil || literals < 0 || literals > 31 {
		return KitSOPCover{}, ErrDimension
	}
	if len(cubes) > o.MaxCubes || literals > 0 && len(cubes) > o.MaxCubeEntries/literals {
		return KitSOPCover{}, ErrLimit
	}
	mask := uint32(0)
	if literals == 31 {
		mask = 0x7fffffff
	} else if literals > 0 {
		mask = 1<<uint(literals) - 1
	}
	for _, cube := range cubes {
		if cube&^mask != 0 {
			return KitSOPCover{}, ErrDimension
		}
	}
	return KitSOPCover{Cubes: slices.Clone(cubes)}, nil
}

// NewInverseKitSOPCover preserves Kit_SopCreateInverse: each selected literal
// in a two-bit-per-variable ISOP cube is exchanged with its opposite literal.
func NewInverseKitSOPCover(cubes []uint32, literals int, o KitCoverOptions) (KitSOPCover, error) {
	cover, err := NewKitSOPCover(cubes, literals, o)
	if err != nil {
		return KitSOPCover{}, err
	}
	for i, cube := range cover.Cubes {
		mask := (cube | cube>>1) & 0x55555555
		mask |= mask << 1
		cover.Cubes[i] = cube ^ mask
	}
	return cover, nil
}

func (c KitSOPCover) Clone() KitSOPCover { return KitSOPCover{Cubes: slices.Clone(c.Cubes)} }

func kitCubeHasLiteral(cube uint32, literal int) bool {
	return literal >= 0 && literal < 32 && cube&(1<<uint(literal)) != 0
}

// DivideByLiteral returns the quotient made from cubes containing literal.
func (c KitSOPCover) DivideByLiteral(literal, literals int, o KitCoverOptions) (KitSOPCover, error) {
	o, err := checkKitCoverOptions(o)
	if err != nil || literal < 0 || literal >= literals || literals > 31 {
		return KitSOPCover{}, ErrDimension
	}
	b := kitCoverBudget{options: o}
	result := make([]uint32, 0, len(c.Cubes))
	for _, cube := range c.Cubes {
		if err := b.charge(1); err != nil {
			return KitSOPCover{}, err
		}
		if kitCubeHasLiteral(cube, literal) {
			result = append(result, cube&^(1<<uint(literal)))
		}
	}
	return NewKitSOPCover(result, literals, o)
}

// Divide returns the deterministic quotient and remainder selected by
// Kit_SopDivideInternal. Inputs are never used as marking storage.
func (c KitSOPCover) Divide(divisor KitSOPCover, literals int, o KitCoverOptions) (KitSOPCover, KitSOPCover, error) {
	o, err := checkKitCoverOptions(o)
	if err != nil || literals < 0 || literals > 31 || len(divisor.Cubes) == 0 || len(c.Cubes) < len(divisor.Cubes) {
		return KitSOPCover{}, KitSOPCover{}, ErrDimension
	}
	if _, err = NewKitSOPCover(c.Cubes, literals, o); err != nil {
		return KitSOPCover{}, KitSOPCover{}, err
	}
	if _, err = NewKitSOPCover(divisor.Cubes, literals, o); err != nil {
		return KitSOPCover{}, KitSOPCover{}, err
	}
	b := kitCoverBudget{options: o}
	used := make([]bool, len(c.Cubes))
	quotient := make([]uint32, 0, len(c.Cubes)/len(divisor.Cubes))
	for i, cube := range c.Cubes {
		if err := b.charge(1); err != nil {
			return KitSOPCover{}, KitSOPCover{}, err
		}
		if used[i] {
			continue
		}
		first := -1
		for k, div := range divisor.Cubes {
			if cube&div == div {
				first = k
				break
			}
		}
		if first < 0 {
			continue
		}
		quo := cube &^ divisor.Cubes[first]
		selected := []int{i}
		complete := true
		for k, div := range divisor.Cubes {
			if k == first {
				continue
			}
			found := -1
			for j, candidate := range c.Cubes {
				if err := b.charge(1); err != nil {
					return KitSOPCover{}, KitSOPCover{}, err
				}
				if used[j] || slices.Contains(selected, j) {
					continue
				}
				if candidate&div == div && candidate&^div == quo {
					found = j
					break
				}
			}
			if found < 0 {
				complete = false
				break
			}
			selected = append(selected, found)
		}
		if !complete {
			continue
		}
		quotient = append(quotient, quo)
		for _, index := range selected {
			used[index] = true
		}
	}
	remainder := make([]uint32, 0, len(c.Cubes)-len(quotient)*len(divisor.Cubes))
	for i, cube := range c.Cubes {
		if !used[i] {
			remainder = append(remainder, cube)
		}
	}
	q, err := NewKitSOPCover(quotient, literals, o)
	if err != nil {
		return KitSOPCover{}, KitSOPCover{}, err
	}
	r, err := NewKitSOPCover(remainder, literals, o)
	if err != nil {
		return KitSOPCover{}, KitSOPCover{}, err
	}
	return q, r, nil
}

func (c KitSOPCover) CommonCube() uint32 {
	if len(c.Cubes) == 0 {
		return ^uint32(0)
	}
	common := ^uint32(0)
	for _, cube := range c.Cubes {
		common &= cube
	}
	return common
}

func (c KitSOPCover) IsCubeFree() bool { return c.CommonCube() == 0 }

func (c KitSOPCover) MakeCubeFree() KitSOPCover {
	result := c.Clone()
	common := result.CommonCube()
	for i := range result.Cubes {
		result.Cubes[i] &^= common
	}
	return result
}

func (c KitSOPCover) CommonCubeCover() (KitSOPCover, error) {
	if len(c.Cubes) == 0 {
		return KitSOPCover{}, ErrDimension
	}
	return KitSOPCover{Cubes: []uint32{c.CommonCube()}}, nil
}

func (c KitSOPCover) literalCount(literal int) int {
	count := 0
	for _, cube := range c.Cubes {
		if kitCubeHasLiteral(cube, literal) {
			count++
		}
	}
	return count
}

func (c KitSOPCover) AnyRepeatedLiteral(literals int) (int, error) {
	if literals < 0 || literals > 31 {
		return -1, ErrDimension
	}
	for literal := 0; literal < literals; literal++ {
		if c.literalCount(literal) > 1 {
			return literal, nil
		}
	}
	return -1, nil
}

// WorstLiteral returns the least frequent repeated literal, with the first
// literal winning ties as in the enabled native branch.
func (c KitSOPCover) WorstLiteral(literals int) (int, error) {
	if literals < 0 || literals > 31 {
		return -1, ErrDimension
	}
	best, bestCount := -1, int(^uint(0)>>1)
	for literal := 0; literal < literals; literal++ {
		count := c.literalCount(literal)
		if count >= 2 && count < bestCount {
			best, bestCount = literal, count
		}
	}
	return best, nil
}

// BestLiteral returns the most frequent repeated literal selected by mask.
// The first literal wins ties.
func (c KitSOPCover) BestLiteral(literals int, mask uint32) (int, error) {
	if literals < 0 || literals > 31 || literals < 31 && mask>>uint(literals) != 0 {
		return -1, ErrDimension
	}
	best, bestCount := -1, -1
	for literal := 0; literal < literals; literal++ {
		if !kitCubeHasLiteral(mask, literal) {
			continue
		}
		count := c.literalCount(literal)
		if count >= 2 && count > bestCount {
			best, bestCount = literal, count
		}
	}
	return best, nil
}

func (c KitSOPCover) BestLiteralCover(literals int, mask uint32) (KitSOPCover, error) {
	literal, err := c.BestLiteral(literals, mask)
	if err != nil {
		return KitSOPCover{}, err
	}
	if literal < 0 {
		return KitSOPCover{}, ErrDimension
	}
	return KitSOPCover{Cubes: []uint32{1 << uint(literal)}}, nil
}

// Divisor computes the quick level-zero kernel. False reports the source's
// trivial-divisor cases and leaves the result empty.
func (c KitSOPCover) Divisor(literals int, o KitCoverOptions) (KitSOPCover, bool, error) {
	o, err := checkKitCoverOptions(o)
	if err != nil || literals < 0 || literals > 31 {
		return KitSOPCover{}, false, ErrDimension
	}
	if len(c.Cubes) <= 1 {
		return KitSOPCover{}, false, nil
	}
	repeated, err := c.AnyRepeatedLiteral(literals)
	if err != nil || repeated < 0 {
		return KitSOPCover{}, false, err
	}
	b := kitCoverBudget{options: o}
	result := c.Clone()
	for {
		if err := b.charge(1); err != nil {
			return KitSOPCover{}, false, err
		}
		literal, err := result.WorstLiteral(literals)
		if err != nil {
			return KitSOPCover{}, false, err
		}
		if literal < 0 {
			break
		}
		result, err = result.DivideByLiteral(literal, literals, KitCoverOptions{MaxCubes: o.MaxCubes, MaxCubeEntries: o.MaxCubeEntries, MaxSteps: o.MaxSteps - b.steps, Stop: o.Stop})
		if err != nil {
			return KitSOPCover{}, false, err
		}
		result = result.MakeCubeFree()
	}
	return result, true, nil
}

// KitPLA owns a uniform-output PLA cover. Each cube uses the ISOP two-bit
// encoding: 1 requires zero, 2 requires one, and 0 is don't-care.
type KitPLA struct {
	Variables  int
	Cubes      []uint32
	Output     byte // '1'/'x' direct, '0'/'n' complemented
	maxEntries int
	maxCubes   int
}

func NewKitPLA(variables int, cubes []uint32, output byte, o KitCoverOptions) (KitPLA, error) {
	o, err := checkKitCoverOptions(o)
	if err != nil || variables < 0 || variables > 16 || !strings.ContainsRune("10xn", rune(output)) {
		return KitPLA{}, ErrDimension
	}
	if len(cubes) == 0 || len(cubes) > o.MaxCubes || variables > 0 && len(cubes) > o.MaxCubeEntries/variables {
		return KitPLA{}, ErrLimit
	}
	mask := uint32(0)
	if variables == 16 {
		mask = ^uint32(0)
	} else if variables > 0 {
		mask = 1<<uint(2*variables) - 1
	}
	for _, cube := range cubes {
		if cube&^mask != 0 {
			return KitPLA{}, ErrDimension
		}
		for variable := 0; variable < variables; variable++ {
			if cube>>uint(2*variable)&3 == 3 {
				return KitPLA{}, ErrDimension
			}
		}
	}
	return KitPLA{Variables: variables, Cubes: slices.Clone(cubes), Output: output, maxEntries: o.MaxCubeEntries, maxCubes: o.MaxCubes}, nil
}

// StartKitPLA preserves Kit_PlaStart by creating all-don't-care cubes.
func StartKitPLA(variables, cubeCount int, o KitCoverOptions) (KitPLA, error) {
	if cubeCount < 0 {
		return KitPLA{}, ErrDimension
	}
	return NewKitPLA(variables, make([]uint32, cubeCount), '1', o)
}

func (p KitPLA) Clone() KitPLA {
	p.Cubes = slices.Clone(p.Cubes)
	return p
}

func (p KitPLA) IsConst0() bool {
	return p.Variables == 0 && len(p.Cubes) == 1 && p.Cubes[0] == 0 && p.Output == '0'
}
func (p KitPLA) IsConst1() bool {
	return p.Variables == 0 && len(p.Cubes) == 1 && p.Cubes[0] == 0 && p.Output == '1'
}
func (p KitPLA) IsComplemented() bool { return p.Output == '0' || p.Output == 'n' }

func (p KitPLA) IsBuffer() bool {
	return p.Variables == 1 && len(p.Cubes) == 1 && (p.Cubes[0] == 2 && p.Output == '1' || p.Cubes[0] == 1 && p.Output == '0')
}

func (p KitPLA) IsInverter() bool {
	return p.Variables == 1 && len(p.Cubes) == 1 && (p.Cubes[0] == 1 && p.Output == '1' || p.Cubes[0] == 2 && p.Output == '0')
}

func (p KitPLA) Complement() KitPLA {
	r := p.Clone()
	switch r.Output {
	case '0':
		r.Output = '1'
	case '1':
		r.Output = '0'
	case 'x':
		r.Output = 'n'
	case 'n':
		r.Output = 'x'
	}
	return r
}

func KitPLAFromISOP(variables int, cubes []uint32, complemented bool, o KitCoverOptions) (KitPLA, error) {
	output := byte('1')
	if complemented {
		output = '0'
	}
	return NewKitPLA(variables, cubes, output, o)
}

func (p KitPLA) ISOP() []uint32 { return slices.Clone(p.Cubes) }

func (p KitPLA) Format() (string, error) {
	if _, err := NewKitPLA(p.Variables, p.Cubes, p.Output, KitCoverOptions{MaxCubes: p.maxCubes, MaxCubeEntries: p.maxEntries}); err != nil {
		return "", err
	}
	var b strings.Builder
	length, err := checkedSize(0, len(p.Cubes), p.Variables+3, p.maxEntries+3*p.maxCubes)
	if err != nil {
		return "", err
	}
	b.Grow(length)
	for _, cube := range p.Cubes {
		for variable := 0; variable < p.Variables; variable++ {
			switch cube >> uint(2*variable) & 3 {
			case 0:
				b.WriteByte('-')
			case 1:
				b.WriteByte('0')
			case 2:
				b.WriteByte('1')
			}
		}
		b.WriteByte(' ')
		b.WriteByte(p.Output)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func ParseKitPLA(text string, o KitCoverOptions) (KitPLA, error) {
	o, err := checkKitCoverOptions(o)
	if err != nil || len(text) == 0 || strings.IndexByte(text, 0) >= 0 || text[len(text)-1] != '\n' {
		return KitPLA{}, ErrDimension
	}
	lines := strings.Split(text[:len(text)-1], "\n")
	if len(lines) == 0 || len(lines) > o.MaxCubes {
		return KitPLA{}, ErrLimit
	}
	if len(lines[0]) < 2 || lines[0][len(lines[0])-2] != ' ' {
		return KitPLA{}, ErrDimension
	}
	variables := len(lines[0]) - 2
	if variables > 16 || variables > 0 && len(lines) > o.MaxCubeEntries/variables {
		return KitPLA{}, ErrLimit
	}
	output := lines[0][len(lines[0])-1]
	cubes := make([]uint32, len(lines))
	for i, line := range lines {
		if len(line) != variables+2 || line[variables] != ' ' || line[variables+1] != output {
			return KitPLA{}, ErrDimension
		}
		for variable := 0; variable < variables; variable++ {
			switch line[variable] {
			case '-':
			case '0':
				cubes[i] |= 1 << uint(2*variable)
			case '1':
				cubes[i] |= 2 << uint(2*variable)
			default:
				return KitPLA{}, ErrDimension
			}
		}
	}
	return NewKitPLA(variables, cubes, output, o)
}

func (p KitPLA) Evaluate(assignment uint64) (bool, error) {
	if p.Variables < 0 || p.Variables > 16 || assignment >= uint64(1)<<uint(p.Variables) {
		return false, ErrDimension
	}
	matched := false
	for _, cube := range p.Cubes {
		match := true
		for variable := 0; variable < p.Variables; variable++ {
			switch cube >> uint(2*variable) & 3 {
			case 1:
				match = match && assignment>>uint(variable)&1 == 0
			case 2:
				match = match && assignment>>uint(variable)&1 != 0
			case 3:
				return false, ErrInvariant
			}
		}
		matched = matched || match
	}
	if p.IsComplemented() {
		matched = !matched
	}
	return matched, nil
}

func (p KitPLA) Truth() (*TruthTable, error) {
	inputs := make([]*TruthTable, p.Variables)
	for variable := range inputs {
		var err error
		inputs[variable], err = NewVariableTruthTable(p.Variables, variable, 1<<20)
		if err != nil {
			return nil, err
		}
	}
	return p.ComposeTruth(inputs)
}

// ComposeTruth preserves Kit_PlaToTruth with caller-supplied input functions.
// The returned table is independent of both the cover and the inputs.
func (p KitPLA) ComposeTruth(inputs []*TruthTable) (*TruthTable, error) {
	if len(inputs) != p.Variables {
		return nil, ErrDimension
	}
	variables := 0
	if len(inputs) > 0 {
		if err := inputs[0].Check(); err != nil {
			return nil, err
		}
		variables = inputs[0].variables
	}
	for _, input := range inputs {
		if err := input.Check(); err != nil || input.variables != variables {
			return nil, ErrDimension
		}
	}
	t, err := NewConstantTruthTable(variables, false, 1<<20)
	if err != nil {
		return nil, err
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
		matched := false
		for _, cube := range p.Cubes {
			match := true
			for inputIndex, input := range inputs {
				value, err := input.Get(assignment)
				if err != nil {
					return nil, err
				}
				switch cube >> uint(2*inputIndex) & 3 {
				case 1:
					match = match && !value
				case 2:
					match = match && value
				case 3:
					return nil, ErrInvariant
				}
			}
			matched = matched || match
		}
		if p.IsComplemented() {
			matched = !matched
		}
		if err = t.Set(assignment, matched); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func KitPLAFromTruth(t *TruthTable, o TruthISOPOptions, limits KitCoverOptions) (KitPLA, error) {
	if t == nil || t.variables > 16 {
		return KitPLA{}, ErrDimension
	}
	r, err := t.ISOP(o)
	if err != nil {
		return KitPLA{}, err
	}
	if len(r.Cubes) == 0 {
		return NewKitPLA(0, []uint32{0}, '0', limits)
	}
	if len(r.Cubes) == 1 && r.Cubes[0] == 0 && !r.Complemented {
		return NewKitPLA(0, []uint32{0}, '1', limits)
	}
	return KitPLAFromISOP(t.variables, r.Cubes, r.Complemented, limits)
}
