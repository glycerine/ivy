package abc

// Derived from ABC src/base/exor/{exor.h,exorBits.c,exorCubes.c,
// exorLink.c,exorList.c,exorUtil.c,exor.c,module.make}.

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

// ExorValue is the two-bit literal encoding used by EXORCISM.  The unusual
// values are useful: XORing two different encodings gives the encoding of the
// residual literal in the distance-one identity.
type ExorValue uint8

const (
	ExorNegative ExorValue = 1
	ExorPositive ExorValue = 2
	ExorAbsent   ExorValue = 3
)

const (
	ExorMaxInputs  = 1000
	ExorMaxOutputs = 1 << 20
)

// ExorCube is one product term and the output mask toggled by that term.
// Slices are owned. Outputs are packed least-significant-bit first.
type ExorCube struct {
	Inputs  []ExorValue
	Outputs []uint64
}

// NewExorCube constructs an owned cube.
func NewExorCube(inputs []ExorValue, outputs []uint64) ExorCube {
	return ExorCube{Inputs: slices.Clone(inputs), Outputs: slices.Clone(outputs)}
}

func (c ExorCube) clone() ExorCube { return NewExorCube(c.Inputs, c.Outputs) }

// LiteralCount returns the number of present input literals.
func (c ExorCube) LiteralCount() int {
	n := 0
	for _, value := range c.Inputs {
		if value != ExorAbsent {
			n++
		}
	}
	return n
}

// NegativeCount returns the number of complemented input literals.
func (c ExorCube) NegativeCount() int {
	n := 0
	for _, value := range c.Inputs {
		if value == ExorNegative {
			n++
		}
	}
	return n
}

// OutputCount returns the number of output bits toggled by the cube.
func (c ExorCube) OutputCount() int {
	n := 0
	for _, word := range c.Outputs {
		n += bits.OnesCount64(word)
	}
	return n
}

// QCost is the source's reversible-circuit cost for the cube's controls.
func (c ExorCube) QCost() int { return ExorQCost(c.LiteralCount(), c.NegativeCount()) }

// TCount preserves ComputeQCostTcountBits for a caller-supplied line count.
func (c ExorCube) TCount(lines int) (int, error) {
	gates, err := ExorToffoliGateCount(c.LiteralCount(), lines)
	if err != nil {
		return 0, err
	}
	return 7 * gates, nil
}

// ExorQCost preserves GetQCost.
func ExorQCost(literals, negative int) int {
	if literals < 0 || negative < 0 || negative > literals {
		return -1
	}
	switch literals {
	case 0:
		return 1
	case 1:
		if negative == 0 {
			return 1
		}
		return 2
	case 2:
		if negative <= 1 {
			return 5
		}
		return 6
	case 3:
		return 14 + 2*max(0, negative-1)
	default:
		extra := negative - literals/2
		return 20 + 12*(literals-4) + 2*max(0, extra)
	}
}

// ExorToffoliGateCount preserves the source's gate-count estimate.
func ExorToffoliGateCount(controls, lines int) (int, error) {
	if controls < 0 || lines < 0 {
		return 0, ErrDimension
	}
	switch controls {
	case 0, 1:
		return 0, nil
	case 2:
		return 1, nil
	case 3:
		return 4, nil
	case 4:
		if (lines+1)/2 >= controls {
			return 8, nil
		}
		return 10, nil
	default:
		if (lines+1)/2 >= controls {
			return 4 * (controls - 2), nil
		}
		return 8 * (controls - 3), nil
	}
}

// ExorCover is an owned multi-output ESOP.
type ExorCover struct {
	Inputs  int
	Outputs int
	Cubes   []ExorCube
}

// ExorOptions bounds minimization work. Stop is polled without introducing a
// runtime interface. Zero limits select conservative defaults.
type ExorOptions struct {
	Quality       int
	UseQCost      bool
	MaxCubes      int
	MaxPairChecks int
	Stop          func() bool
}

// ExorStats describes a completed atomic minimization.
type ExorStats struct {
	CubesBefore, CubesAfter       int
	LiteralsBefore, LiteralsAfter int
	QCostBefore, QCostAfter       int
	PairChecks, LinkAttempts      int
}

func exorWordCount(outputs int) int { return (outputs + 63) / 64 }

// Validate checks dimensions, literal encodings, output padding, and limits.
func (c ExorCover) Validate(maxCubes int) error {
	if c.Inputs <= 0 || c.Inputs > ExorMaxInputs || c.Outputs <= 0 {
		return ErrDimension
	}
	if c.Outputs > ExorMaxOutputs {
		return ErrLimit
	}
	if maxCubes > 0 && len(c.Cubes) > maxCubes {
		return ErrLimit
	}
	words := exorWordCount(c.Outputs)
	for _, cube := range c.Cubes {
		if len(cube.Inputs) != c.Inputs || len(cube.Outputs) != words {
			return ErrDimension
		}
		for _, value := range cube.Inputs {
			if value < ExorNegative || value > ExorAbsent {
				return ErrDimension
			}
		}
		if rem := c.Outputs & 63; rem != 0 && cube.Outputs[words-1]>>rem != 0 {
			return ErrDimension
		}
	}
	return nil
}

// Clone returns a deep independent cover.
func (c ExorCover) Clone() ExorCover {
	r := ExorCover{Inputs: c.Inputs, Outputs: c.Outputs, Cubes: make([]ExorCube, len(c.Cubes))}
	for i := range c.Cubes {
		r.Cubes[i] = c.Cubes[i].clone()
	}
	return r
}

// Evaluate evaluates one assignment and returns an owned output mask.
func (c ExorCover) Evaluate(assignment []bool) ([]uint64, error) {
	if len(assignment) != c.Inputs {
		return nil, ErrDimension
	}
	if err := c.Validate(0); err != nil {
		return nil, err
	}
	result := make([]uint64, exorWordCount(c.Outputs))
	for _, cube := range c.Cubes {
		matches := true
		for variable, value := range cube.Inputs {
			if value == ExorNegative && assignment[variable] || value == ExorPositive && !assignment[variable] {
				matches = false
				break
			}
		}
		if matches {
			for w := range result {
				result[w] ^= cube.Outputs[w]
			}
		}
	}
	return result, nil
}

func exorInputKey(inputs []ExorValue) string {
	b := make([]byte, len(inputs))
	for i := range inputs {
		b[i] = byte(inputs[i])
	}
	return string(b)
}

func exorOutputsZero(outputs []uint64) bool {
	for _, word := range outputs {
		if word != 0 {
			return false
		}
	}
	return true
}

func exorOutputsEqual(a, b []uint64) bool { return slices.Equal(a, b) }

// Normalize combines equal products by XORing output masks, removes zero
// products, and preserves first-encounter order. It does not mutate c.
func (c ExorCover) Normalize() (ExorCover, error) {
	if err := c.Validate(0); err != nil {
		return ExorCover{}, err
	}
	r := ExorCover{Inputs: c.Inputs, Outputs: c.Outputs}
	positions := make(map[string]int, len(c.Cubes)) // lookup only; order is in Cubes
	for _, cube := range c.Cubes {
		key := exorInputKey(cube.Inputs)
		if pos, ok := positions[key]; ok {
			for w := range cube.Outputs {
				r.Cubes[pos].Outputs[w] ^= cube.Outputs[w]
			}
			continue
		}
		positions[key] = len(r.Cubes)
		r.Cubes = append(r.Cubes, cube.clone())
	}
	out := r.Cubes[:0]
	for _, cube := range r.Cubes {
		if !exorOutputsZero(cube.Outputs) {
			out = append(out, cube)
		}
	}
	r.Cubes = out
	return r, nil
}

// ExorDistance reports differing input variables plus one when output masks
// differ. Values above cutoff are returned as cutoff+1.
func ExorDistance(a, b ExorCube, cutoff int) (int, error) {
	if len(a.Inputs) != len(b.Inputs) || len(a.Outputs) != len(b.Outputs) || cutoff < 0 {
		return 0, ErrDimension
	}
	distance := 0
	for i := range a.Inputs {
		if a.Inputs[i] != b.Inputs[i] {
			distance++
			if distance > cutoff {
				return cutoff + 1, nil
			}
		}
	}
	if !exorOutputsEqual(a.Outputs, b.Outputs) {
		distance++
	}
	if distance > cutoff {
		return cutoff + 1, nil
	}
	return distance, nil
}

// ExorDifferentVariables returns -1 first when outputs differ, followed by
// ascending differing input indices. Results are bounded by cutoff.
func ExorDifferentVariables(a, b ExorCube, cutoff int) ([]int, error) {
	if len(a.Inputs) != len(b.Inputs) || len(a.Outputs) != len(b.Outputs) || cutoff < 0 {
		return nil, ErrDimension
	}
	diff := make([]int, 0, min(cutoff+1, 5))
	if !exorOutputsEqual(a.Outputs, b.Outputs) {
		diff = append(diff, -1)
	}
	for i := range a.Inputs {
		if a.Inputs[i] != b.Inputs[i] {
			diff = append(diff, i)
			if len(diff) > cutoff {
				return nil, ErrLimit
			}
		}
	}
	return diff, nil
}

func exorAdjacent(a, b ExorCube) (ExorCube, bool) {
	if !exorOutputsEqual(a.Outputs, b.Outputs) || len(a.Inputs) != len(b.Inputs) {
		return ExorCube{}, false
	}
	diff := -1
	for i := range a.Inputs {
		if a.Inputs[i] != b.Inputs[i] {
			if diff >= 0 {
				return ExorCube{}, false
			}
			diff = i
		}
	}
	if diff < 0 {
		return ExorCube{}, false
	}
	r := a.clone()
	r.Inputs[diff] = a.Inputs[diff] ^ b.Inputs[diff]
	if r.Inputs[diff] < ExorNegative || r.Inputs[diff] > ExorAbsent {
		return ExorCube{}, false
	}
	return r, true
}

func exorReduceAdjacent(c ExorCover, stop func() bool, checks *int, maxChecks int) (ExorCover, error) {
	normalized, err := c.Normalize()
	if err != nil {
		return ExorCover{}, err
	}
	for {
		changed := false
		for i := 0; i < len(normalized.Cubes) && !changed; i++ {
			for j := i + 1; j < len(normalized.Cubes); j++ {
				*checks++
				if maxChecks > 0 && *checks > maxChecks {
					return ExorCover{}, ErrLimit
				}
				if stop != nil && stop() {
					return ExorCover{}, ErrCancelled
				}
				merged, ok := exorAdjacent(normalized.Cubes[i], normalized.Cubes[j])
				if !ok {
					continue
				}
				next := normalized.Clone()
				next.Cubes[i] = merged
				next.Cubes = append(next.Cubes[:j], next.Cubes[j+1:]...)
				normalized, err = next.Normalize()
				if err != nil {
					return ExorCover{}, err
				}
				changed = true
				break
			}
		}
		if !changed {
			return normalized, nil
		}
	}
}

var exorLinkGroups = map[int][][]int{
	2: {{0, 3}, {2, 1}},
	3: {{0, 6, 11}, {0, 7, 10}, {4, 2, 11}, {4, 3, 9}, {8, 1, 7}, {8, 3, 5}},
	4: {
		{0, 12, 22, 31}, {0, 12, 23, 30}, {0, 20, 14, 31}, {0, 20, 15, 29}, {0, 28, 13, 23}, {0, 28, 15, 21},
		{8, 4, 22, 31}, {8, 4, 23, 30}, {8, 18, 6, 31}, {8, 18, 7, 27}, {8, 26, 5, 23}, {8, 26, 7, 19},
		{16, 2, 14, 31}, {16, 2, 15, 29}, {16, 10, 6, 31}, {16, 10, 7, 27}, {16, 25, 3, 15}, {16, 25, 7, 11},
		{24, 1, 13, 23}, {24, 1, 15, 21}, {24, 9, 5, 23}, {24, 9, 7, 19}, {24, 17, 3, 15}, {24, 17, 7, 11},
	},
}

func exorSelectValue(a, b ExorValue, selector int) ExorValue {
	if selector == 0 {
		return a
	}
	if selector == 1 {
		return b
	}
	return a ^ b
}

func exorSelectOutputs(a, b []uint64, selector int) []uint64 {
	if selector == 0 {
		return slices.Clone(a)
	}
	if selector == 1 {
		return slices.Clone(b)
	}
	r := make([]uint64, len(a))
	for i := range a {
		r[i] = a[i] ^ b[i]
	}
	return r
}

func exorGeneratedCube(a, b ExorCube, diff []int, index int) ExorCube {
	distance := len(diff)
	xPosition := index / (1 << (distance - 1))
	bitsValue := index & ((1 << (distance - 1)) - 1)
	r := a.clone()
	for _, variable := range diff {
		if variable >= 0 {
			r.Inputs[variable] = ExorAbsent
		}
	}
	other := 0
	for position, variable := range diff {
		selector := 2
		if position != xPosition {
			shift := distance - 2 - other
			selector = (bitsValue >> shift) & 1
			other++
		}
		if variable < 0 {
			r.Outputs = exorSelectOutputs(a.Outputs, b.Outputs, selector)
		} else {
			r.Inputs[variable] = exorSelectValue(a.Inputs[variable], b.Inputs[variable], selector)
		}
	}
	return r
}

// ExorLinkCandidates returns the source's deterministic distance-2 through
// distance-4 replacement groups. Each group has distance cubes and is owned.
func ExorLinkCandidates(a, b ExorCube) ([][]ExorCube, error) {
	diff, err := ExorDifferentVariables(a, b, 4)
	if err != nil {
		return nil, err
	}
	groups, ok := exorLinkGroups[len(diff)]
	if !ok {
		return nil, ErrDimension
	}
	generated := make([]ExorCube, len(diff)*(1<<(len(diff)-1)))
	for i := range generated {
		generated[i] = exorGeneratedCube(a, b, diff, i)
	}
	result := make([][]ExorCube, len(groups))
	for i, rule := range groups {
		result[i] = make([]ExorCube, len(rule))
		for j, cube := range rule {
			result[i][j] = generated[cube].clone()
		}
	}
	return result, nil
}

func exorCoverCounts(c ExorCover) (literals, qcost int) {
	for _, cube := range c.Cubes {
		literals += cube.LiteralCount()
		qcost += cube.QCost()
	}
	return literals, qcost
}

func exorBetter(a, b ExorCover, useQ bool) bool {
	if len(a.Cubes) != len(b.Cubes) {
		return len(a.Cubes) < len(b.Cubes)
	}
	al, aq := exorCoverCounts(a)
	bl, bq := exorCoverCounts(b)
	if useQ {
		if aq != bq {
			return aq < bq
		}
		return al < bl
	}
	if al != bl {
		return al < bl
	}
	return aq < bq
}

// MinimizeExorCover applies exact duplicate/distance-one reduction followed by
// deterministic distance-2, -3 and -4 EXOR-link trials. Failure is atomic.
func MinimizeExorCover(input ExorCover, options ExorOptions) (ExorCover, ExorStats, error) {
	stats := ExorStats{CubesBefore: len(input.Cubes)}
	stats.LiteralsBefore, stats.QCostBefore = exorCoverCounts(input)
	maxCubes := options.MaxCubes
	if maxCubes == 0 {
		maxCubes = 1 << 20
	}
	maxChecks := options.MaxPairChecks
	if maxChecks == 0 {
		maxChecks = 10_000_000
	}
	if options.Quality < 0 || options.Quality > 1000 {
		return ExorCover{}, stats, ErrDimension
	}
	if err := input.Validate(maxCubes); err != nil {
		return ExorCover{}, stats, err
	}
	if options.Stop != nil && options.Stop() {
		return ExorCover{}, stats, ErrCancelled
	}
	current, err := exorReduceAdjacent(input.Clone(), options.Stop, &stats.PairChecks, maxChecks)
	if err != nil {
		return ExorCover{}, stats, err
	}
	noImprovement := 0
	for noImprovement < 1+options.Quality {
		improved := false
		for distance := 2; distance <= 4 && !improved; distance++ {
			for i := 0; i < len(current.Cubes) && !improved; i++ {
				for j := i + 1; j < len(current.Cubes); j++ {
					stats.PairChecks++
					if stats.PairChecks > maxChecks {
						return ExorCover{}, stats, ErrLimit
					}
					if options.Stop != nil && options.Stop() {
						return ExorCover{}, stats, ErrCancelled
					}
					d, e := ExorDistance(current.Cubes[i], current.Cubes[j], 4)
					if e != nil || d != distance {
						continue
					}
					groups, e := ExorLinkCandidates(current.Cubes[i], current.Cubes[j])
					if e != nil {
						return ExorCover{}, stats, e
					}
					for _, group := range groups {
						stats.LinkAttempts++
						candidate := ExorCover{Inputs: current.Inputs, Outputs: current.Outputs}
						candidate.Cubes = make([]ExorCube, 0, len(current.Cubes)-2+len(group))
						for k, cube := range current.Cubes {
							if k != i && k != j {
								candidate.Cubes = append(candidate.Cubes, cube.clone())
							}
						}
						for _, cube := range group {
							candidate.Cubes = append(candidate.Cubes, cube.clone())
						}
						if len(candidate.Cubes) > maxCubes {
							continue
						}
						candidate, e = exorReduceAdjacent(candidate, options.Stop, &stats.PairChecks, maxChecks)
						if e != nil {
							return ExorCover{}, stats, e
						}
						if exorBetter(candidate, current, options.UseQCost) {
							current = candidate
							improved = true
							break
						}
					}
					if improved {
						break
					}
				}
			}
		}
		if improved {
			noImprovement = 0
		} else {
			noImprovement++
		}
	}
	stats.CubesAfter = len(current.Cubes)
	stats.LiteralsAfter, stats.QCostAfter = exorCoverCounts(current)
	return current, stats, nil
}

// MarshalExorPLA returns deterministic ESOP PLA text without timestamps or
// process-global output files.
func (c ExorCover) MarshalExorPLA() ([]byte, error) {
	if err := c.Validate(0); err != nil {
		return nil, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, ".i %d\n.o %d\n.p %d\n.type esop\n", c.Inputs, c.Outputs, len(c.Cubes))
	for _, cube := range c.Cubes {
		for _, value := range cube.Inputs {
			switch value {
			case ExorNegative:
				b.WriteByte('0')
			case ExorPositive:
				b.WriteByte('1')
			default:
				b.WriteByte('-')
			}
		}
		b.WriteByte(' ')
		for output := 0; output < c.Outputs; output++ {
			if c.Cubes != nil && cube.Outputs[output>>6]&(uint64(1)<<(output&63)) != 0 {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		}
		b.WriteByte('\n')
	}
	b.WriteString(".e\n")
	return []byte(b.String()), nil
}

// ExorCoverDigest provides a stable content digest input without exposing
// internal backing arrays. It is intentionally not a cryptographic hash.
func (c ExorCover) ExorCoverDigest() ([]byte, error) {
	if err := c.Validate(0); err != nil {
		return nil, err
	}
	perCube := c.Inputs + 8*exorWordCount(c.Outputs)
	if len(c.Cubes) > 0 && perCube > (int(^uint(0)>>1)-8)/len(c.Cubes) {
		return nil, ErrLimit
	}
	size := 8 + len(c.Cubes)*perCube
	b := make([]byte, 8, size)
	binary.LittleEndian.PutUint32(b[0:4], uint32(c.Inputs))
	binary.LittleEndian.PutUint32(b[4:8], uint32(c.Outputs))
	for _, cube := range c.Cubes {
		for _, value := range cube.Inputs {
			b = append(b, byte(value))
		}
		for _, word := range cube.Outputs {
			var tmp [8]byte
			binary.LittleEndian.PutUint64(tmp[:], word)
			b = append(b, tmp[:]...)
		}
	}
	return b, nil
}
