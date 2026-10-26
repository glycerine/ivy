package abc

// NPN configuration composition and recovery from opt/dau/dauNpn2.c.

type DttNPNTransform struct {
	variables   uint8
	permutation [5]uint8
	phases      [5]bool
}

func IdentityDttNPNTransform(variables int) (DttNPNTransform, error) {
	if variables < 2 || variables > 5 {
		return DttNPNTransform{}, ErrDimension
	}
	transform := DttNPNTransform{variables: uint8(variables)}
	for index := 0; index < variables; index++ {
		transform.permutation[index] = uint8(index)
	}
	return transform, nil
}

func NewDttNPNTransform(permutation []int, phases []bool) (DttNPNTransform, error) {
	variables := len(permutation)
	if variables < 2 || variables > 5 || len(phases) != variables {
		return DttNPNTransform{}, ErrDimension
	}
	if err := checkPermutation(permutation, variables); err != nil {
		return DttNPNTransform{}, err
	}
	transform := DttNPNTransform{variables: uint8(variables)}
	for index := range permutation {
		transform.permutation[index] = uint8(permutation[index])
		transform.phases[index] = phases[index]
	}
	return transform, nil
}

func (transform DttNPNTransform) Variables() int { return int(transform.variables) }

func (transform DttNPNTransform) Permutation() []int {
	result := make([]int, transform.variables)
	for index := range result {
		result[index] = int(transform.permutation[index])
	}
	return result
}

func (transform DttNPNTransform) InputPhases() []bool {
	return append([]bool(nil), transform.phases[:transform.variables]...)
}

func (transform DttNPNTransform) check() error {
	variables := int(transform.variables)
	if variables < 2 || variables > 5 {
		return ErrDimension
	}
	seen := [5]bool{}
	for index := 0; index < variables; index++ {
		value := int(transform.permutation[index])
		if value >= variables || seen[value] {
			return ErrInvariant
		}
		seen[value] = true
	}
	return nil
}

// ComposeDttNPNTransforms implements Dtt_ComposeNP: result[i] is
// first[second[i]], with the corresponding phase xor.
func ComposeDttNPNTransforms(first, second DttNPNTransform) (DttNPNTransform, error) {
	if err := first.check(); err != nil {
		return DttNPNTransform{}, err
	}
	if err := second.check(); err != nil {
		return DttNPNTransform{}, err
	}
	if first.variables != second.variables {
		return DttNPNTransform{}, ErrDimension
	}
	result := DttNPNTransform{variables: first.variables}
	for index := 0; index < int(first.variables); index++ {
		middle := second.permutation[index]
		result.permutation[index] = first.permutation[middle]
		result.phases[index] = first.phases[middle] != second.phases[index]
	}
	return result, nil
}

// Variable returns the target variable and phase spelling of one source input,
// matching Dtt_MakePI's inverse lookup.
func (transform DttNPNTransform) Variable(source int) (int, bool, error) {
	if err := transform.check(); err != nil {
		return 0, false, err
	}
	if source < 0 || source >= int(transform.variables) {
		return 0, false, ErrDimension
	}
	for target := 0; target < int(transform.variables); target++ {
		if int(transform.permutation[target]) == source {
			return target, transform.phases[target], nil
		}
	}
	return 0, false, ErrInvariant
}

func (transform DttNPNTransform) ApplyTruth(truth uint64) (uint64, error) {
	if err := transform.check(); err != nil {
		return 0, err
	}
	variables := int(transform.variables)
	truth &= dauTruthMask(variables)
	var result uint64
	for targetAssignment := 0; targetAssignment < 1<<uint(variables); targetAssignment++ {
		sourceAssignment := 0
		for target := 0; target < variables; target++ {
			bit := (targetAssignment >> uint(target)) & 1
			if transform.phases[target] {
				bit ^= 1
			}
			sourceAssignment |= bit << transform.permutation[target]
		}
		if truth&(1<<uint(sourceAssignment)) != 0 {
			result |= 1 << uint(targetAssignment)
		}
	}
	return result, nil
}

type DttTransformSearchOptions struct {
	MaxTransforms uint64
	Stop          func() bool
}

func FindDttNPNTransform(source, target uint64, variables int, options DttTransformSearchOptions) (DttNPNTransform, bool, error) {
	walker, err := newDttOrbitWalker(variables, options.MaxTransforms, options.Stop)
	if err != nil {
		return DttNPNTransform{}, false, err
	}
	entries, err := walker.entries(source)
	if err != nil {
		return DttNPNTransform{}, false, err
	}
	target &= walker.mask
	for _, entry := range entries {
		actual, err := entry.Transform.ApplyTruth(source)
		if err != nil {
			return DttNPNTransform{}, false, err
		}
		if actual == target {
			return entry.Transform, false, nil
		}
		if (^actual)&walker.mask == target {
			return entry.Transform, true, nil
		}
	}
	return DttNPNTransform{}, false, ErrUnsupported
}

func ProcessDttType(kind uint8, fanin int) (uint8, error) {
	if kind > 9 || fanin < 1 || fanin > 3 {
		return 0, ErrDimension
	}
	if fanin == 3 {
		if kind < 5 {
			return kind + 5, nil
		}
		return kind - 5, nil
	}
	if kind == 0 || kind == 5 {
		return kind + uint8(fanin), nil
	}
	if int(kind) == fanin {
		return 0, nil
	}
	if int(kind)+fanin == 3 {
		return 8, nil
	}
	if kind == 3 {
		if fanin == 1 {
			return 7, nil
		}
		return 6, nil
	}
	if kind == 4 {
		return 9, nil
	}
	if int(kind) == fanin+5 {
		return 5, nil
	}
	if int(kind)+fanin == 8 {
		return 3, nil
	}
	if kind == 8 {
		if fanin == 1 {
			return 2, nil
		}
		return 1, nil
	}
	if kind == 9 {
		return 4, nil
	}
	return 0, ErrInvariant
}

func ApplyDttType(kind uint8, first, second, mask uint64) (uint64, error) {
	first &= mask
	second &= mask
	switch kind {
	case 0:
		return first & second, nil
	case 1:
		return (^first) & second & mask, nil
	case 2:
		return first & (^second) & mask, nil
	case 3:
		return first | second, nil
	case 4:
		return first ^ second, nil
	case 5:
		return (^(first & second)) & mask, nil
	case 6:
		return (^(^first & second)) & mask, nil
	case 7:
		return (^(first & ^second)) & mask, nil
	case 8:
		return (^first) & (^second) & mask, nil
	case 9:
		return (^(first ^ second)) & mask, nil
	}
	return 0, ErrDimension
}
