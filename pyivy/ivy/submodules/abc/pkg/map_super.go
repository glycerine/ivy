package abc

// Native port of src/map/super/{super.c,superAnd.c,superGate.c} and its headers.

import (
	"math"
	"slices"
	"strconv"
	"strings"
)

const superNoVariable = -9999.0
const superEpsilon = 0.001

// AndSuperLink is an owned, checked replacement for the complemented-pointer
// edges used by Super2_Gate_t.
type AndSuperLink struct {
	Gate       int
	Complement bool
}

// AndSupergate describes a constant, input, or two-input AND/INV tree.
type AndSupergate struct {
	Truth uint32
	Input int
	Left  AndSuperLink
	Right AndSuperLink
}

// AndSuperLibrary is the deterministic result of Super2_Precompute. Gates from
// earlier levels remain in the library, as in Super2_LibDup.
type AndSuperLibrary struct {
	Inputs int
	Levels int
	Tried  int
	Gates  []AndSupergate
}

// AndSuperOptions bounds computation. A zero MaxGates/MaxAttempts is unbounded.
type AndSuperOptions struct {
	MaxGates    int
	MaxAttempts int
	Stop        func() bool
}

func superMask(bits int) uint32 {
	if bits == 32 {
		return ^uint32(0)
	}
	return uint32(1)<<uint(bits) - 1
}

func normalizeAndTruth(truth, mask, top uint32) uint32 {
	if truth&top != 0 {
		return mask &^ truth
	}
	return truth
}

// PrecomputeAndSupergates enumerates one representative of every function up
// to output complementation obtainable from distinct prior trees at each level.
func PrecomputeAndSupergates(inputs, levels int, options AndSuperOptions) (*AndSuperLibrary, error) {
	if inputs < 0 || inputs >= 6 || levels < 0 || levels > 5 || options.MaxGates < 0 || options.MaxAttempts < 0 {
		return nil, ErrDimension
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	mints := 1 << uint(inputs)
	mask := superMask(mints)
	top := uint32(1) << uint(mints-1)
	lib := &AndSuperLibrary{Inputs: inputs, Levels: levels}
	lib.Gates = append(lib.Gates, AndSupergate{Input: -1})
	for variable := 0; variable < inputs; variable++ {
		var truth uint32
		for assignment := 0; assignment < mints; assignment++ {
			if assignment&(1<<uint(variable)) != 0 {
				truth |= uint32(1) << uint(assignment)
			}
		}
		lib.Gates = append(lib.Gates, AndSupergate{Truth: truth, Input: variable})
	}
	if options.MaxGates > 0 && len(lib.Gates) > options.MaxGates {
		return nil, ErrLimit
	}
	attempts := 0
	for level := 1; level <= levels; level++ {
		oldCount := len(lib.Gates)
		// Super2_LibCompute resets nTried to the number of gates already in
		// the duplicated library before counting the four phases of each
		// candidate pair. Keep the resource-limit count separate so
		// MaxAttempts remains a whole-operation bound.
		lib.Tried = oldCount
		seen := make(map[uint32]struct{}, len(lib.Gates))
		for _, gate := range lib.Gates {
			seen[normalizeAndTruth(gate.Truth, mask, top)] = struct{}{}
		}
		for first := 0; first < oldCount; first++ {
			for second := 0; second < first; second++ {
				for phase := 0; phase < 4; phase++ {
					if options.Stop != nil && options.Stop() {
						return nil, ErrCancelled
					}
					if options.MaxAttempts > 0 && attempts >= options.MaxAttempts {
						return nil, ErrLimit
					}
					left, right := lib.Gates[first].Truth, lib.Gates[second].Truth
					if phase&1 != 0 {
						left = mask &^ left
					}
					if phase&2 != 0 {
						right = mask &^ right
					}
					truth := left & right
					canonical := normalizeAndTruth(truth, mask, top)
					lib.Tried++
					attempts++
					if _, exists := seen[canonical]; exists {
						continue
					}
					if options.MaxGates > 0 && len(lib.Gates) >= options.MaxGates {
						return nil, ErrLimit
					}
					seen[canonical] = struct{}{}
					lib.Gates = append(lib.Gates, AndSupergate{
						Truth: truth, Input: -2,
						Left:  AndSuperLink{Gate: first, Complement: phase&1 != 0},
						Right: AndSuperLink{Gate: second, Complement: phase&2 != 0},
					})
				}
			}
		}
	}
	return lib, nil
}

func (lib *AndSuperLibrary) Check() error {
	if lib == nil || lib.Inputs < 0 || lib.Inputs >= 6 || lib.Levels < 0 || len(lib.Gates) < lib.Inputs+1 {
		return ErrDimension
	}
	mints := 1 << uint(lib.Inputs)
	mask, top := superMask(mints), uint32(1)<<uint(mints-1)
	seen := make(map[uint32]struct{}, len(lib.Gates))
	for index, gate := range lib.Gates {
		if gate.Truth&^mask != 0 {
			return ErrInvariant
		}
		if index == 0 {
			if gate.Truth != 0 || gate.Input != -1 {
				return ErrInvariant
			}
		} else if index <= lib.Inputs {
			if gate.Input != index-1 {
				return ErrInvariant
			}
		} else if gate.Input != -2 || gate.Left.Gate < 0 || gate.Left.Gate >= index || gate.Right.Gate < 0 || gate.Right.Gate >= index || gate.Left.Gate == gate.Right.Gate {
			return ErrInvariant
		}
		key := normalizeAndTruth(gate.Truth, mask, top)
		if _, exists := seen[key]; exists {
			return ErrInvariant
		}
		seen[key] = struct{}{}
	}
	return nil
}

func andSuperExpression(lib *AndSuperLibrary, index int, complement bool) (string, error) {
	if lib == nil || index < 0 || index >= len(lib.Gates) {
		return "", ErrDimension
	}
	gate := lib.Gates[index]
	if gate.Input >= -1 {
		if gate.Input == -1 {
			if complement {
				return "1$", nil
			}
			return "0$", nil
		}
		letter := byte('a' + gate.Input)
		if complement {
			letter = byte('A' + gate.Input)
		}
		return string(letter), nil
	}
	left, err := andSuperExpression(lib, gate.Left.Gate, gate.Left.Complement)
	if err != nil {
		return "", err
	}
	right, err := andSuperExpression(lib, gate.Right.Gate, gate.Right.Complement)
	if err != nil {
		return "", err
	}
	// The C writer orders branches by phase and then by a punctuation-insensitive
	// variable comparison. This stable key gives the same canonical purpose.
	if gate.Left.Complement != gate.Right.Complement {
		if !gate.Left.Complement {
			left, right = right, left
		}
	} else if super2WriteCompare(left, right) != 1 {
		left, right = right, left
	}
	if complement {
		return "<" + left + right + ">", nil
	}
	return "(" + left + right + ")", nil
}

func super2WriteCompare(left, right string) int {
	i, j := 0, 0
	for {
		for i < len(left) && left[i] < 'A' {
			i++
		}
		for j < len(right) && right[j] < 'A' {
			j++
		}
		if i == len(left) || j == len(right) {
			if j < len(right) {
				return 1
			}
			return -1
		}
		if left[i] < right[j] {
			return 1
		}
		if left[i] > right[j] {
			return -1
		}
		i++
		j++
	}
}

func superWriteInt(out *strings.Builder, value int) {
	out.WriteString(strconv.Itoa(value))
}

func superWritePaddedInt(out *strings.Builder, value, width int, zero bool) {
	text := strconv.Itoa(value)
	pad := ' '
	if zero {
		pad = '0'
	}
	for count := len(text); count < width; count++ {
		out.WriteRune(pad)
	}
	out.WriteString(text)
}

func superWriteBinary(out *strings.Builder, value uint64, width int) {
	text := strconv.FormatUint(value, 2)
	for count := len(text); count < width; count++ {
		out.WriteByte('0')
	}
	out.WriteString(text)
}

func superWriteFloat2(out *strings.Builder, value float64, width int) {
	text := strconv.FormatFloat(value, 'f', 2, 64)
	for count := len(text); count < width; count++ {
		out.WriteByte(' ')
	}
	out.WriteString(text)
}

// MarshalText returns deterministic standalone data rather than writing the
// process-global filename used by Super2_LibWrite.
func (lib *AndSuperLibrary) MarshalText() ([]byte, error) {
	if err := lib.Check(); err != nil {
		return nil, err
	}
	mints := 1 << uint(lib.Inputs)
	mask, top := superMask(mints), uint32(1)<<uint(mints-1)
	order := make([]int, len(lib.Gates))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(left, right int) int {
		a := normalizeAndTruth(lib.Gates[left].Truth, mask, top)
		b := normalizeAndTruth(lib.Gates[right].Truth, mask, top)
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return left - right
	})
	var out strings.Builder
	out.WriteString("# AND2/INV supergates\n# inputs = ")
	superWriteInt(&out, lib.Inputs)
	out.WriteString("\n# levels = ")
	superWriteInt(&out, lib.Levels)
	out.WriteString("\n# gates = ")
	superWriteInt(&out, len(lib.Gates))
	out.WriteByte('\n')
	superWriteInt(&out, len(lib.Gates))
	out.WriteByte('\n')
	for _, index := range order {
		gate := lib.Gates[index]
		invert := gate.Truth&top != 0
		truth := gate.Truth
		if invert {
			truth = mask &^ truth
		}
		expression, err := andSuperExpression(lib, index, invert)
		if err != nil {
			return nil, err
		}
		superWriteBinary(&out, uint64(truth), mints)
		out.WriteString("   ")
		out.WriteString(expression)
		out.WriteByte('\n')
	}
	return []byte(out.String()), nil
}

// SuperPrimitiveGate is an owned technology-library root gate. Truth uses the
// low 2^Fanins bits, indexed by the root's fanin values. PinDelays contains one
// maximum delay per root pin.
type SuperPrimitiveGate struct {
	Name      string
	Fanins    int
	Truth     uint64
	Area      float64
	PinDelays []float64
}

// SuperGate is either an elementary variable (Root == -1) or a composition.
type SuperGate struct {
	Root      int
	Input     int
	Fanins    []int
	Truth     uint64
	Area      float64
	PinDelays []float64
	DelayMax  float64
	IsLibrary bool
}

// SuperLibrary owns all generated gates and the primitive roots used by them.
type SuperLibrary struct {
	Name       string
	Variables  int
	Levels     int
	Tried      int
	Added      int
	Removed    int
	Lookups    int
	Aliases    int
	Roots      []SuperPrimitiveGate
	Gates      []SuperGate
	LibraryIDs []int
}

// SuperOptions corresponds to the useful, non-process-global parts of the
// super command. Time limits become deterministic attempt limits or Stop.
type SuperOptions struct {
	Name               string
	Variables          int
	Levels             int
	MaxGates           int
	MaxAttempts        int
	DelayMax           float64
	AreaMax            float64
	SkipInputInverters bool
	Stop               func() bool
}

func validateSuperRoots(roots []SuperPrimitiveGate, options SuperOptions) error {
	if options.Variables < 2 || options.Variables > 6 || options.Levels < 0 || options.MaxGates < 0 || options.MaxGates > 0 && options.MaxGates < options.Variables || options.MaxAttempts < 0 || options.DelayMax < 0 || options.AreaMax < 0 || len(roots) == 0 {
		return ErrDimension
	}
	for _, root := range roots {
		if root.Name == "" || root.Fanins < 1 || root.Fanins > 6 || root.Fanins > options.Variables || len(root.PinDelays) != root.Fanins || root.Area < 0 || math.IsNaN(root.Area) || math.IsInf(root.Area, 0) {
			return ErrDimension
		}
		bits := 1 << uint(root.Fanins)
		if bits < 64 && root.Truth>>uint(bits) != 0 {
			return ErrDimension
		}
		for _, delay := range root.PinDelays {
			if delay < 0 || math.IsNaN(delay) || math.IsInf(delay, 0) {
				return ErrDimension
			}
		}
	}
	return nil
}

// PrecomputeSupergates composes technology roots over up to six variables,
// retaining the same area/delay Pareto frontier per exact truth table.
func PrecomputeSupergates(roots []SuperPrimitiveGate, options SuperOptions) (*SuperLibrary, error) {
	if err := validateSuperRoots(roots, options); err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	lib := &SuperLibrary{Name: options.Name, Variables: options.Variables, Roots: make([]SuperPrimitiveGate, len(roots))}
	if lib.Name == "" {
		lib.Name = "library"
	}
	for i, root := range roots {
		root.PinDelays = slices.Clone(root.PinDelays)
		lib.Roots[i] = root
	}
	for variable := 0; variable < options.Variables; variable++ {
		truth := uint64(0)
		for assignment := 0; assignment < 1<<uint(options.Variables); assignment++ {
			if assignment&(1<<uint(variable)) != 0 {
				truth |= uint64(1) << uint(assignment)
			}
		}
		delays := make([]float64, options.Variables)
		for i := range delays {
			delays[i] = superNoVariable
		}
		delays[variable] = 0
		lib.Gates = append(lib.Gates, SuperGate{Root: -1, Input: variable, Truth: truth, PinDelays: delays})
	}
	frontiers := make(map[uint64][]int)
	for id, gate := range lib.Gates {
		frontiers[gate.Truth] = append(frontiers[gate.Truth], id)
		lib.Added++
	}
	for level := 1; level <= options.Levels; level++ {
		snapshot := len(lib.Gates)
		order := make([]int, snapshot)
		for i := range order {
			order[i] = i
		}
		slices.SortStableFunc(order, func(left, right int) int {
			a, b := lib.Gates[left], lib.Gates[right]
			if a.DelayMax != b.DelayMax {
				if a.DelayMax < b.DelayMax {
					return -1
				}
				return 1
			}
			if a.Area != b.Area {
				if a.Area < b.Area {
					return -1
				}
				return 1
			}
			return left - right
		})
		stopLevel := false
		for rootID, root := range lib.Roots {
			if level > 1 && root.Fanins >= 4 {
				continue
			}
			chosen := make([]int, root.Fanins)
			used := make([]bool, snapshot)
			var enumerate func(int) error
			enumerate = func(pin int) error {
				if pin == root.Fanins {
					if options.Stop != nil && options.Stop() {
						return ErrCancelled
					}
					if options.MaxAttempts > 0 && lib.Tried >= options.MaxAttempts {
						return ErrLimit
					}
					lib.Tried++
					candidate, err := composeSuperGate(lib, rootID, chosen)
					if err != nil {
						return err
					}
					truthMask := superTruthMask(options.Variables)
					if options.AreaMax > 0 && candidate.Area > options.AreaMax+superEpsilon || options.DelayMax > 0 && candidate.DelayMax > options.DelayMax+superEpsilon || candidate.Truth == 0 || candidate.Truth == truthMask {
						return nil
					}
					keep, remove := compareSuperCandidate(lib, frontiers[candidate.Truth], candidate)
					if !keep {
						return nil
					}
					newClass := len(frontiers[candidate.Truth]) == 0
					if len(remove) != 0 {
						removeSet := make(map[int]struct{}, len(remove))
						for _, id := range remove {
							removeSet[id] = struct{}{}
						}
						kept := frontiers[candidate.Truth][:0]
						for _, id := range frontiers[candidate.Truth] {
							if _, gone := removeSet[id]; !gone {
								kept = append(kept, id)
							}
						}
						frontiers[candidate.Truth] = kept
						lib.Removed += len(remove)
					}
					id := len(lib.Gates)
					lib.Gates = append(lib.Gates, candidate)
					frontiers[candidate.Truth] = append(frontiers[candidate.Truth], id)
					lib.Added++
					// The source tests nClasses after insertion and retains the gate
					// that first exceeds the requested class count.
					if options.MaxGates > 0 && newClass && len(frontiers) > options.MaxGates {
						stopLevel = true
					}
					return nil
				}
				for pos, id := range order {
					if stopLevel {
						break
					}
					if used[id] {
						continue
					}
					gate := lib.Gates[id]
					if level > 1 && len(gate.Fanins) >= 3 {
						continue
					}
					if options.SkipInputInverters && root.Fanins == 1 && gate.Root < 0 {
						continue
					}
					if options.AreaMax > 0 && root.Area+gate.Area > options.AreaMax+superEpsilon {
						continue
					}
					if options.DelayMax > 0 && root.PinDelays[pin]+gate.DelayMax > options.DelayMax+superEpsilon {
						continue
					}
					_ = pos
					used[id] = true
					chosen[pin] = id
					if err := enumerate(pin + 1); err != nil {
						return err
					}
					used[id] = false
				}
				return nil
			}
			if err := enumerate(0); err != nil {
				return nil, err
			}
			if stopLevel {
				break
			}
		}
		lib.Levels = level
		if stopLevel {
			break
		}
	}
	for truth, ids := range frontiers {
		_ = truth
		for _, id := range ids {
			if lib.Gates[id].Root >= 0 && superConsecutiveSupport(lib.Gates[id].PinDelays) {
				lib.LibraryIDs = append(lib.LibraryIDs, id)
			}
		}
	}
	slices.SortFunc(lib.LibraryIDs, func(left, right int) int {
		a, b := lib.Gates[left], lib.Gates[right]
		if a.Truth != b.Truth {
			if a.Truth < b.Truth {
				return -1
			}
			return 1
		}
		if a.Area != b.Area {
			if a.Area < b.Area {
				return -1
			}
			return 1
		}
		if a.DelayMax != b.DelayMax {
			if a.DelayMax < b.DelayMax {
				return -1
			}
			return 1
		}
		return left - right
	})
	for _, id := range lib.LibraryIDs {
		lib.Gates[id].IsLibrary = true
	}
	return lib, nil
}

func superConsecutiveSupport(delays []float64) bool {
	absent := false
	for _, delay := range delays {
		if delay <= superNoVariable+superEpsilon {
			absent = true
		} else if absent {
			return false
		}
	}
	return true
}

func superTruthMask(variables int) uint64 {
	bits := 1 << uint(variables)
	if bits == 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(bits) - 1
}

func composeSuperGate(lib *SuperLibrary, rootID int, fanins []int) (SuperGate, error) {
	root := lib.Roots[rootID]
	gate := SuperGate{Root: rootID, Input: -1, Fanins: slices.Clone(fanins), Area: root.Area, PinDelays: make([]float64, lib.Variables)}
	for variable := range gate.PinDelays {
		gate.PinDelays[variable] = superNoVariable
	}
	for _, id := range fanins {
		gate.Area += lib.Gates[id].Area
	}
	mints := 1 << uint(lib.Variables)
	for assignment := 0; assignment < mints; assignment++ {
		rootAssignment := 0
		for pin, id := range fanins {
			if lib.Gates[id].Truth&(uint64(1)<<uint(assignment)) != 0 {
				rootAssignment |= 1 << uint(pin)
			}
		}
		if root.Truth&(uint64(1)<<uint(rootAssignment)) != 0 {
			gate.Truth |= uint64(1) << uint(assignment)
		}
	}
	for variable := 0; variable < lib.Variables; variable++ {
		for pin, id := range fanins {
			delay := lib.Gates[id].PinDelays[variable]
			if delay <= superNoVariable+superEpsilon {
				continue
			}
			delay += root.PinDelays[pin]
			if gate.PinDelays[variable] <= superNoVariable+superEpsilon || delay > gate.PinDelays[variable] {
				gate.PinDelays[variable] = delay
			}
		}
		if gate.PinDelays[variable] > gate.DelayMax {
			gate.DelayMax = gate.PinDelays[variable]
		}
	}
	return gate, nil
}

func compareSuperCandidate(lib *SuperLibrary, ids []int, candidate SuperGate) (bool, []int) {
	var remove []int
	for _, id := range ids {
		lib.Lookups++
		gate := lib.Gates[id]
		oldBetter, newBetter := gate.Area+superEpsilon < candidate.Area, gate.Area > candidate.Area+superEpsilon
		for pin := 0; pin < lib.Variables; pin++ {
			a, b := gate.PinDelays[pin], candidate.PinDelays[pin]
			if a <= superNoVariable+superEpsilon || b <= superNoVariable+superEpsilon {
				continue
			}
			if a+superEpsilon < b {
				oldBetter = true
			} else if a > b+superEpsilon {
				newBetter = true
			}
		}
		switch {
		case oldBetter && newBetter:
		case newBetter:
			remove = append(remove, id)
		case oldBetter:
			return false, nil
		default:
			return false, nil
		}
	}
	return true, remove
}

func (lib *SuperLibrary) Check() error {
	if lib == nil || lib.Variables < 2 || lib.Variables > 6 || len(lib.Roots) == 0 || len(lib.Gates) < lib.Variables {
		return ErrDimension
	}
	for id, gate := range lib.Gates {
		if len(gate.PinDelays) != lib.Variables || gate.Area < 0 || gate.DelayMax < 0 {
			return ErrInvariant
		}
		if id < lib.Variables {
			if gate.Root != -1 || gate.Input != id || len(gate.Fanins) != 0 {
				return ErrInvariant
			}
		} else {
			if gate.Root < 0 || gate.Root >= len(lib.Roots) || len(gate.Fanins) != lib.Roots[gate.Root].Fanins {
				return ErrInvariant
			}
			seen := make(map[int]struct{}, len(gate.Fanins))
			for _, fanin := range gate.Fanins {
				if fanin < 0 || fanin >= id {
					return ErrInvariant
				}
				if _, duplicate := seen[fanin]; duplicate {
					return ErrInvariant
				}
				seen[fanin] = struct{}{}
			}
		}
	}
	for _, id := range lib.LibraryIDs {
		if id < 0 || id >= len(lib.Gates) || !lib.Gates[id].IsLibrary {
			return ErrInvariant
		}
	}
	return nil
}

func (lib *SuperLibrary) gateName(id int) (string, error) {
	if lib == nil || id < 0 || id >= len(lib.Gates) {
		return "", ErrDimension
	}
	gate := lib.Gates[id]
	if gate.Root < 0 {
		return string(rune('a' + gate.Input)), nil
	}
	var out strings.Builder
	out.WriteString(lib.Roots[gate.Root].Name)
	out.WriteByte('(')
	for i, fanin := range gate.Fanins {
		if i != 0 {
			out.WriteByte(',')
		}
		name, err := lib.gateName(fanin)
		if err != nil {
			return "", err
		}
		out.WriteString(name)
	}
	out.WriteByte(')')
	return out.String(), nil
}

// MarshalTree emits the standalone tree format produced by Super_Write.
func (lib *SuperLibrary) MarshalTree() ([]byte, error) {
	if err := lib.Check(); err != nil {
		return nil, err
	}
	var out strings.Builder
	unique := 0
	var previous uint64
	for i, id := range lib.LibraryIDs {
		truth := lib.Gates[id].Truth
		if i == 0 || truth != previous {
			unique++
			previous = truth
		}
	}
	out.WriteString("# Supergate library derived for ")
	out.WriteString(strconv.Quote(lib.Name))
	out.WriteString(".\n# inputs = ")
	superWriteInt(&out, lib.Variables)
	out.WriteString("\n# levels = ")
	superWriteInt(&out, lib.Levels)
	out.WriteString("\n# attempts = ")
	superWriteInt(&out, lib.Tried)
	out.WriteString("\n# supergates = ")
	superWriteInt(&out, len(lib.LibraryIDs))
	out.WriteString("\n# functions = ")
	superWriteInt(&out, unique)
	out.WriteByte('\n')
	out.WriteString(lib.Name)
	out.WriteByte('\n')
	superWriteInt(&out, lib.Variables)
	out.WriteByte('\n')
	linePos := out.Len()
	out.WriteString("         \n")
	numbers := make([]int, len(lib.Gates))
	for i := range numbers {
		numbers[i] = -1
	}
	for i := 0; i < lib.Variables; i++ {
		numbers[i] = i
	}
	counter := lib.Variables
	var writeGate func(int) error
	writeGate = func(id int) error {
		if numbers[id] >= 0 {
			return nil
		}
		gate := lib.Gates[id]
		for _, fanin := range gate.Fanins {
			if err := writeGate(fanin); err != nil {
				return err
			}
		}
		numbers[id] = counter
		counter++
		if gate.IsLibrary {
			out.WriteString("* ")
		}
		out.WriteString(lib.Roots[gate.Root].Name)
		for _, fanin := range gate.Fanins {
			out.WriteByte(' ')
			superWriteInt(&out, numbers[fanin])
		}
		out.WriteByte('\n')
		return nil
	}
	for _, id := range lib.LibraryIDs {
		if err := writeGate(id); err != nil {
			return nil, err
		}
	}
	text := []byte(out.String())
	count := []byte(strconv.Itoa(counter))
	copy(text[linePos:linePos+9], count)
	return text, nil
}

// MarshalOld emits the truth/delay/area/formula view without filesystem I/O.
func (lib *SuperLibrary) MarshalOld() ([]byte, error) {
	if err := lib.Check(); err != nil {
		return nil, err
	}
	var out strings.Builder
	out.WriteString(lib.Name)
	out.WriteByte('\n')
	superWriteInt(&out, lib.Variables)
	out.WriteByte('\n')
	superWriteInt(&out, len(lib.LibraryIDs))
	out.WriteByte('\n')
	width := 1 << uint(lib.Variables)
	for number, id := range lib.LibraryIDs {
		gate := lib.Gates[id]
		name, err := lib.gateName(id)
		if err != nil {
			return nil, err
		}
		superWritePaddedInt(&out, number+1, 4, true)
		out.WriteString("  ")
		superWriteBinary(&out, gate.Truth, width)
		out.WriteString("   ")
		superWriteFloat2(&out, gate.DelayMax, 5)
		for _, delay := range gate.PinDelays {
			if delay <= superNoVariable+superEpsilon {
				delay = 0
			}
			out.WriteByte(' ')
			superWriteFloat2(&out, delay, 5)
		}
		out.WriteString("   ")
		superWriteFloat2(&out, gate.Area, 5)
		out.WriteString("   ")
		out.WriteString(name)
		out.WriteByte('\n')
	}
	return []byte(out.String()), nil
}

// Super2Command and SuperCommandOptions preserve command parsing without a
// command registry, process-global getopt state, or implicit files.
type SuperCommandDescriptor struct {
	Group string
	Name  string
}

// SuperCommandCatalog is the owned replacement for Super_Init's mutation of
// the process-global ABC command registry. Super_End has no source-side work,
// and an owned catalog needs no matching teardown.
func SuperCommandCatalog() []SuperCommandDescriptor {
	return []SuperCommandDescriptor{{Group: "SC mapping", Name: "super"}, {Group: "SC mapping", Name: "super2"}}
}

type Super2Command struct {
	Inputs, Levels int
	Verbose        bool
}

func ParseSuper2Command(args []string) (Super2Command, error) {
	result := Super2Command{Inputs: 4, Levels: 3}
	for i := 0; i < len(args); i++ {
		token := args[i]
		if len(token) < 2 || token[0] != '-' {
			return result, ErrDimension
		}
		for pos := 1; pos < len(token); pos++ {
			option := token[pos]
			switch option {
			case 'v':
				result.Verbose = !result.Verbose
			case 'I', 'L':
				text := token[pos+1:]
				if text == "" {
					if i+1 >= len(args) {
						return result, ErrDimension
					}
					i++
					text = args[i]
				}
				value, err := strconv.Atoi(text)
				if err != nil || value < 0 {
					return result, ErrDimension
				}
				if option == 'I' {
					result.Inputs = value
				} else {
					result.Levels = value
				}
				pos = len(token)
			default:
				return result, ErrDimension
			}
		}
	}
	return result, nil
}

type SuperCommandOptions struct {
	InputName                             string
	ExcludeName                           string
	Variables, Levels, MaxGates           int
	TimeLimitSeconds                      int
	DelayMax, AreaMax                     float64
	SkipInputInverters, WriteOld, Verbose bool
}

func ParseSuperCommand(args []string) (SuperCommandOptions, error) {
	result := SuperCommandOptions{Variables: 5, Levels: 2, SkipInputInverters: true}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			if result.InputName != "" {
				return result, ErrDimension
			}
			result.InputName = arg
			continue
		}
		for pos := 1; pos < len(arg); pos++ {
			option := arg[pos]
			switch option {
			case 's':
				result.SkipInputInverters = !result.SkipInputInverters
			case 'o':
				result.WriteOld = !result.WriteOld
			case 'v':
				result.Verbose = !result.Verbose
			case 'I', 'L', 'N', 'T', 'D', 'A', 'E':
				text := arg[pos+1:]
				if text == "" {
					if i+1 >= len(args) {
						return result, ErrDimension
					}
					i++
					text = args[i]
				}
				switch option {
				case 'E':
					if text == "" {
						return result, ErrDimension
					}
					result.ExcludeName = text
				case 'D', 'A':
					value, err := strconv.ParseFloat(text, 64)
					if err != nil || value <= 0 {
						return result, ErrDimension
					}
					if option == 'D' {
						result.DelayMax = value
					} else {
						result.AreaMax = value
					}
				default:
					value, err := strconv.Atoi(text)
					if err != nil || value < 0 {
						return result, ErrDimension
					}
					switch option {
					case 'I':
						result.Variables = value
					case 'L':
						result.Levels = value
					case 'N':
						result.MaxGates = value
					case 'T':
						result.TimeLimitSeconds = value
					}
				}
				pos = len(arg)
			default:
				return result, ErrDimension
			}
		}
	}
	if result.InputName == "" || result.Variables < 2 || result.Variables > 6 {
		return result, ErrDimension
	}
	return result, nil
}
