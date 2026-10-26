package abc

// Owned cube-cover manipulation from src/map/cov. The Abc_Ntk_t conversion
// routines in covBuild.c, covCore.c, covMan.c, and covTest.c remain an explicit
// integration boundary until the general ABC network representation is ported.

import (
	"fmt"
	"sort"
	"strings"
)

var CovSourceInventory = [...]string{
	"cov.h", "covInt.h", "covBuild.c", "covCore.c", "covMan.c",
	"covMinEsop.c", "covMinMan.c", "covMinSop.c", "covMinUtil.c",
	"covTest.c", "module.make",
}

// CovValue preserves Min_Cube_t's two-bit encoding: 01 means zero, 10 means
// one, 11 is don't-care, and 00 is the invalid empty value.
type CovValue uint8

const (
	CovInvalid  CovValue = 0
	CovZero     CovValue = 1
	CovOne      CovValue = 2
	CovDontCare CovValue = 3
)

type CovCube struct{ values []CovValue }

func NewCovCube(values []CovValue) (CovCube, error) {
	c := CovCube{values: append([]CovValue(nil), values...)}
	if err := c.check(len(values)); err != nil {
		return CovCube{}, err
	}
	return c, nil
}

func CovCubeFromPattern(pattern string) (CovCube, error) {
	values := make([]CovValue, len(pattern))
	for i := range pattern {
		switch pattern[i] {
		case '0':
			values[i] = CovZero
		case '1':
			values[i] = CovOne
		case '-':
			values[i] = CovDontCare
		default:
			return CovCube{}, fmt.Errorf("COV cube column %d: %w", i, ErrDimension)
		}
	}
	return CovCube{values: values}, nil
}

func (c CovCube) check(variables int) error {
	if variables < 0 || len(c.values) != variables {
		return ErrDimension
	}
	for _, value := range c.values {
		if value < CovZero || value > CovDontCare {
			return ErrInvariant
		}
	}
	return nil
}

func (c CovCube) Values() []CovValue { return append([]CovValue(nil), c.values...) }

func (c CovCube) Pattern() string {
	b := make([]byte, len(c.values))
	for i, value := range c.values {
		switch value {
		case CovZero:
			b[i] = '0'
		case CovOne:
			b[i] = '1'
		case CovDontCare:
			b[i] = '-'
		default:
			b[i] = '?'
		}
	}
	return string(b)
}

func (c CovCube) LiteralCount() int {
	count := 0
	for _, value := range c.values {
		if value != CovDontCare {
			count++
		}
	}
	return count
}

func (c CovCube) Check() error { return c.check(len(c.values)) }

func (c CovCube) Contains(inner CovCube) (bool, error) {
	if err := c.Check(); err != nil {
		return false, err
	}
	if err := inner.check(len(c.values)); err != nil {
		return false, err
	}
	return covCubeContains(c, inner), nil
}

func (c CovCube) Product(other CovCube) (CovCube, bool, error) {
	if err := c.Check(); err != nil {
		return CovCube{}, false, err
	}
	if err := other.check(len(c.values)); err != nil {
		return CovCube{}, false, err
	}
	product, ok := covCubeProduct(c, other)
	return product, ok, nil
}

func (c CovCube) Disjoint(other CovCube) (bool, error) {
	_, intersects, err := c.Product(other)
	return !intersects, err
}

// DistanceVariables returns the ordered variables whose two-bit cube values
// differ. It is the checked counterpart of Min_CubesDistOne/DistTwo.
func (c CovCube) DistanceVariables(other CovCube) ([]int, error) {
	if err := c.Check(); err != nil {
		return nil, err
	}
	if err := other.check(len(c.values)); err != nil {
		return nil, err
	}
	result := make([]int, 0)
	for i := range c.values {
		if c.values[i] != other.values[i] {
			result = append(result, i)
		}
	}
	return result, nil
}

func (c CovCube) Matches(assignment uint64) bool {
	for variable, value := range c.values {
		bit := CovZero
		if assignment>>uint(variable)&1 != 0 {
			bit = CovOne
		}
		if value != CovDontCare && value != bit {
			return false
		}
	}
	return true
}

func covCubeEqual(a, b CovCube) bool {
	if len(a.values) != len(b.values) {
		return false
	}
	for i := range a.values {
		if a.values[i] != b.values[i] {
			return false
		}
	}
	return true
}

// covCubeContains reports whether every minterm of inner is in outer.
func covCubeContains(outer, inner CovCube) bool {
	if len(outer.values) != len(inner.values) {
		return false
	}
	for i, value := range outer.values {
		if value != CovDontCare && value != inner.values[i] {
			return false
		}
	}
	return true
}

func covCubeProduct(a, b CovCube) (CovCube, bool) {
	if len(a.values) != len(b.values) {
		return CovCube{}, false
	}
	values := make([]CovValue, len(a.values))
	for i := range values {
		switch {
		case a.values[i] == CovDontCare:
			values[i] = b.values[i]
		case b.values[i] == CovDontCare || a.values[i] == b.values[i]:
			values[i] = a.values[i]
		default:
			return CovCube{}, false
		}
	}
	return CovCube{values: values}, true
}

func covCubeMerge(a, b CovCube) (CovCube, bool) {
	if len(a.values) != len(b.values) {
		return CovCube{}, false
	}
	difference := -1
	values := append([]CovValue(nil), a.values...)
	for i := range values {
		if a.values[i] == b.values[i] {
			continue
		}
		if difference >= 0 || a.values[i] == CovDontCare || b.values[i] == CovDontCare {
			return CovCube{}, false
		}
		difference = i
	}
	if difference < 0 {
		return CovCube{}, false
	}
	values[difference] = CovDontCare
	return CovCube{values: values}, true
}

type CovCover struct {
	variables int
	exclusive bool
	cubes     []CovCube
}

func NewCovCover(variables int, exclusive bool, cubes []CovCube) (CovCover, error) {
	if variables < 0 {
		return CovCover{}, ErrDimension
	}
	result := CovCover{variables: variables, exclusive: exclusive, cubes: make([]CovCube, len(cubes))}
	for i := range cubes {
		if err := cubes[i].check(variables); err != nil {
			return CovCover{}, err
		}
		result.cubes[i] = CovCube{values: append([]CovValue(nil), cubes[i].values...)}
	}
	return result, nil
}

func (c CovCover) Variables() int  { return c.variables }
func (c CovCover) Exclusive() bool { return c.exclusive }
func (c CovCover) CubeCount() int  { return len(c.cubes) }

func (c CovCover) Cubes() []CovCube {
	result := make([]CovCube, len(c.cubes))
	for i := range c.cubes {
		result[i] = CovCube{values: append([]CovValue(nil), c.cubes[i].values...)}
	}
	return result
}

func (c CovCover) check() error {
	if c.variables < 0 {
		return ErrDimension
	}
	for _, cube := range c.cubes {
		if err := cube.check(c.variables); err != nil {
			return err
		}
	}
	return nil
}

func (c CovCover) Evaluate(assignment uint64) bool {
	value := false
	for _, cube := range c.cubes {
		if !cube.Matches(assignment) {
			continue
		}
		if c.exclusive {
			value = !value
		} else {
			return true
		}
	}
	return value
}

func (c CovCover) LiteralCount() int {
	count := 0
	for _, cube := range c.cubes {
		count += cube.LiteralCount()
	}
	return count
}

func (c CovCover) Support() []int {
	used := make([]bool, c.variables)
	for _, cube := range c.cubes {
		for i, value := range cube.values {
			used[i] = used[i] || value != CovDontCare
		}
	}
	result := make([]int, 0, c.variables)
	for i := range used {
		if used[i] {
			result = append(result, i)
		}
	}
	return result
}

func (c CovCover) Format(output byte) (string, error) {
	if err := c.check(); err != nil || output != '0' && output != '1' {
		if err != nil {
			return "", err
		}
		return "", ErrDimension
	}
	var b strings.Builder
	for _, cube := range c.cubes {
		b.WriteString(cube.Pattern())
		b.WriteByte(' ')
		b.WriteByte(output)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

type CovOptions struct {
	MaxVariables, MaxCubes, MaxOperations int
	Stop                                  func() bool
}

type CovManager struct {
	variables int
	options   CovOptions
}

func NewCovManager(variables int, options CovOptions) (*CovManager, error) {
	if variables < 0 || options.MaxVariables < 0 || options.MaxCubes < 0 || options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1024
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 24
	}
	if variables > options.MaxVariables {
		return nil, ErrLimit
	}
	return &CovManager{variables: variables, options: options}, nil
}

type covWork struct {
	options CovOptions
	steps   int
}

func (w *covWork) charge(amount int) error {
	if w.options.Stop != nil && w.options.Stop() {
		return ErrCancelled
	}
	if amount < 0 || w.steps > w.options.MaxOperations-amount {
		return ErrLimit
	}
	w.steps += amount
	return nil
}

func (m *CovManager) accept(c CovCover) error {
	if m == nil || c.variables != m.variables {
		return ErrDimension
	}
	if err := c.check(); err != nil {
		return err
	}
	if len(c.cubes) > m.options.MaxCubes {
		return ErrLimit
	}
	return nil
}

func covSort(cubes []CovCube) {
	sort.SliceStable(cubes, func(i, j int) bool {
		li, lj := cubes[i].LiteralCount(), cubes[j].LiteralCount()
		if li != lj {
			return li < lj
		}
		return cubes[i].Pattern() < cubes[j].Pattern()
	})
}

func (m *CovManager) MinimizeSOP(cover CovCover) (CovCover, error) {
	if err := m.accept(cover); err != nil || cover.exclusive {
		if err != nil {
			return CovCover{}, err
		}
		return CovCover{}, ErrDimension
	}
	work := covWork{options: m.options}
	cubes := cover.Cubes()
	changed := true
	for changed {
		changed = false
		covSort(cubes)
		keep := make([]CovCube, 0, len(cubes))
		for i, cube := range cubes {
			if err := work.charge(1); err != nil {
				return CovCover{}, err
			}
			contained := false
			for j, other := range cubes {
				if i != j && (other.LiteralCount() < cube.LiteralCount() || other.LiteralCount() == cube.LiteralCount() && j < i) && covCubeContains(other, cube) {
					contained = true
					break
				}
			}
			if contained {
				changed = true
				continue
			}
			keep = append(keep, cube)
		}
		cubes = keep
		merged := false
		for i := 0; i < len(cubes) && !merged; i++ {
			for j := i + 1; j < len(cubes); j++ {
				if err := work.charge(1); err != nil {
					return CovCover{}, err
				}
				if cube, ok := covCubeMerge(cubes[i], cubes[j]); ok {
					cubes = append(append(cubes[:i:i], cubes[i+1:j]...), cubes[j+1:]...)
					cubes = append(cubes, cube)
					changed, merged = true, true
					break
				}
			}
		}
		if len(cubes) > m.options.MaxCubes {
			return CovCover{}, ErrLimit
		}
	}
	covSort(cubes)
	return NewCovCover(m.variables, false, cubes)
}

func (m *CovManager) MinimizeESOP(cover CovCover) (CovCover, error) {
	if err := m.accept(cover); err != nil || !cover.exclusive {
		if err != nil {
			return CovCover{}, err
		}
		return CovCover{}, ErrDimension
	}
	work := covWork{options: m.options}
	cubes := cover.Cubes()
	for {
		covSort(cubes)
		cancelled := false
		for i := 0; i < len(cubes) && !cancelled; i++ {
			for j := i + 1; j < len(cubes); j++ {
				if err := work.charge(1); err != nil {
					return CovCover{}, err
				}
				if covCubeEqual(cubes[i], cubes[j]) {
					cubes = append(append(cubes[:i:i], cubes[i+1:j]...), cubes[j+1:]...)
					cancelled = true
					break
				}
			}
		}
		if cancelled {
			continue
		}
		merged := false
		for i := 0; i < len(cubes) && !merged; i++ {
			for j := i + 1; j < len(cubes); j++ {
				if err := work.charge(1); err != nil {
					return CovCover{}, err
				}
				if cube, ok := covCubeMerge(cubes[i], cubes[j]); ok {
					cubes = append(append(cubes[:i:i], cubes[i+1:j]...), cubes[j+1:]...)
					cubes = append(cubes, cube)
					merged = true
					break
				}
			}
		}
		if !merged {
			break
		}
		if len(cubes) > m.options.MaxCubes {
			return CovCover{}, ErrLimit
		}
	}
	covSort(cubes)
	return NewCovCover(m.variables, true, cubes)
}

func (m *CovManager) Product(first, second CovCover, exclusive bool) (CovCover, error) {
	if err := m.accept(first); err != nil {
		return CovCover{}, err
	}
	if err := m.accept(second); err != nil {
		return CovCover{}, err
	}
	work := covWork{options: m.options}
	cubes := make([]CovCube, 0)
	for _, a := range first.cubes {
		for _, b := range second.cubes {
			if err := work.charge(1); err != nil {
				return CovCover{}, err
			}
			if cube, ok := covCubeProduct(a, b); ok {
				if len(cubes) == m.options.MaxCubes {
					return CovCover{}, ErrLimit
				}
				cubes = append(cubes, cube)
			}
		}
	}
	result, _ := NewCovCover(m.variables, exclusive, cubes)
	if exclusive {
		return m.MinimizeESOP(result)
	}
	return m.MinimizeSOP(result)
}

func (m *CovManager) Union(first, second CovCover, exclusive bool) (CovCover, error) {
	if err := m.accept(first); err != nil {
		return CovCover{}, err
	}
	if err := m.accept(second); err != nil {
		return CovCover{}, err
	}
	if len(first.cubes) > m.options.MaxCubes-len(second.cubes) {
		return CovCover{}, ErrLimit
	}
	cubes := append(first.Cubes(), second.Cubes()...)
	result, _ := NewCovCover(m.variables, exclusive, cubes)
	if exclusive {
		return m.MinimizeESOP(result)
	}
	return m.MinimizeSOP(result)
}

func covSharpCube(first, blocked CovCube) []CovCube {
	if _, intersects := covCubeProduct(first, blocked); !intersects {
		return []CovCube{first}
	}
	if covCubeContains(blocked, first) {
		return nil
	}
	prefix := CovCube{values: append([]CovValue(nil), first.values...)}
	result := make([]CovCube, 0)
	for i := range prefix.values {
		if prefix.values[i] != CovDontCare || blocked.values[i] == CovDontCare {
			continue
		}
		outside := CovCube{values: append([]CovValue(nil), prefix.values...)}
		if blocked.values[i] == CovZero {
			outside.values[i] = CovOne
		} else {
			outside.values[i] = CovZero
		}
		result = append(result, outside)
		prefix.values[i] = blocked.values[i]
	}
	return result
}

func (m *CovManager) ComplementSOP(cover CovCover) (CovCover, error) {
	if err := m.accept(cover); err != nil || cover.exclusive {
		if err != nil {
			return CovCover{}, err
		}
		return CovCover{}, ErrDimension
	}
	tautology := CovCube{values: make([]CovValue, m.variables)}
	for i := range tautology.values {
		tautology.values[i] = CovDontCare
	}
	cubes := []CovCube{tautology}
	work := covWork{options: m.options}
	for _, blocked := range cover.cubes {
		next := make([]CovCube, 0)
		for _, cube := range cubes {
			if err := work.charge(1); err != nil {
				return CovCover{}, err
			}
			parts := covSharpCube(cube, blocked)
			if len(next) > m.options.MaxCubes-len(parts) {
				return CovCover{}, ErrLimit
			}
			next = append(next, parts...)
		}
		cubes = next
	}
	result, _ := NewCovCover(m.variables, false, cubes)
	return m.MinimizeSOP(result)
}

func ParseCovCover(text string, variables int, exclusive bool, options CovOptions) (CovCover, error) {
	manager, err := NewCovManager(variables, options)
	if err != nil {
		return CovCover{}, err
	}
	lines := strings.Split(text, "\n")
	cubes := make([]CovCube, 0)
	for lineNumber, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || len(fields[0]) != variables || fields[1] != "1" && fields[1] != "0" {
			return CovCover{}, fmt.Errorf("COV line %d: %w", lineNumber+1, ErrDimension)
		}
		if len(cubes) == manager.options.MaxCubes {
			return CovCover{}, ErrLimit
		}
		if manager.options.Stop != nil && manager.options.Stop() {
			return CovCover{}, ErrCancelled
		}
		cube, err := CovCubeFromPattern(fields[0])
		if err != nil {
			return CovCover{}, fmt.Errorf("COV line %d: %w", lineNumber+1, err)
		}
		cubes = append(cubes, cube)
	}
	return NewCovCover(variables, exclusive, cubes)
}
