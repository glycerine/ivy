package abc

// Input/group symmetry from opt/dau/dauCanon.c:
// Abc_TtIsSymmetric and Abc_TtIsSymmetricHigh.

// TruthSymmetry is a bit mask: phased symmetry swaps the inputs/groups and
// complements selected inputs; direct symmetry swaps them without complements.
type TruthSymmetry uint8

const (
	TruthSymmetryNone   TruthSymmetry = 0
	TruthSymmetryPhased TruthSymmetry = 1
	TruthSymmetryDirect TruthSymmetry = 2
	TruthSymmetryBoth   TruthSymmetry = 3
)

// InputSymmetry checks invariance under swapping two distinct inputs. If
// includePhase, it also checks swapping and complementing both inputs.
func (t *TruthTable) InputSymmetry(first, second int, includePhase bool) (TruthSymmetry, error) {
	return t.GroupSymmetry([]int{first}, []int{second}, nil, includePhase)
}

// GroupSymmetry swaps corresponding inputs in two nonempty, disjoint groups
// of equal size. It optionally checks the same swap followed by complementing
// selected inputs in these groups. phaseVariables is indexed by input; nil
// selects every group input. Inputs outside the groups are never complemented.
// Results match C's symmetry bits, with owned scratch instead of static buffers.
func (t *TruthTable) GroupSymmetry(first, second []int, phaseVariables []bool, includePhase bool) (TruthSymmetry, error) {
	if e := t.Check(); e != nil {
		return 0, e
	}
	if len(first) == 0 || len(first) != len(second) || (phaseVariables != nil && len(phaseVariables) != t.variables) {
		return 0, ErrDimension
	}
	seen := make([]bool, t.variables)
	for _, group := range [][]int{first, second} {
		for _, variable := range group {
			if variable < 0 || variable >= t.variables || seen[variable] {
				return 0, ErrDimension
			}
			seen[variable] = true
		}
	}
	r := t.Clone()
	for i, variable := range first {
		r.swapVariables(variable, second[i])
	}
	result := TruthSymmetryNone
	if r.Equal(t) {
		result |= TruthSymmetryDirect
	}
	if !includePhase {
		return result, nil
	}
	for _, group := range [][]int{first, second} {
		for _, variable := range group {
			if phaseVariables == nil || phaseVariables[variable] {
				r.flipVariable(variable)
			}
		}
	}
	if r.Equal(t) {
		result |= TruthSymmetryPhased
	}
	return result, nil
}
