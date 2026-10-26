package abc

// Cover/PLA verification, named permutation and consistency from src/misc/espresso/verify.c.

import "slices"

type EspressoVerifyOptions struct {
	MaxCubes, MaxOperations int
	Stop                    func() bool
}

func newEspressoVerifyWork(options EspressoVerifyOptions) (*espressoSharpWork, error) {
	return newEspressoSharpWork(EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
}

type EspressoCoverVerification struct {
	Equivalent                      bool
	CurrentOutside, OriginalOutside *EspressoSetFamily
}

func (layout *EspressoCubeLayout) coverDifference(source, allowed *EspressoSetFamily, work *espressoSharpWork) (*EspressoSetFamily, error) {
	if err := source.check(); err != nil {
		return nil, err
	}
	if err := allowed.check(); err != nil {
		return nil, err
	}
	if source.bits != layout.bitCount || allowed.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(source.sets) > work.options.MaxCubes || len(allowed.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	cubes := make([]*EspressoSet, 0)
	for _, cube := range source.sets {
		pieces, err := layout.sharpCubeByCover(cube, allowed, work)
		if err != nil {
			return nil, err
		}
		cubes = append(cubes, pieces.sets...)
	}
	return layout.sharpFamily(cubes, true, work)
}

// VerifyCovers checks current <= original union DC and original <= current
// union DC, returning owned counterexample covers in both directions.
func (layout *EspressoCubeLayout) VerifyCovers(current, original, dontCare *EspressoSetFamily, options EspressoVerifyOptions) (EspressoCoverVerification, error) {
	work, err := newEspressoVerifyWork(options)
	if err != nil {
		return EspressoCoverVerification{}, err
	}
	if err = layout.check(); err != nil {
		return EspressoCoverVerification{}, err
	}
	if err = dontCare.check(); err != nil {
		return EspressoCoverVerification{}, err
	}
	if dontCare.bits != layout.bitCount {
		return EspressoCoverVerification{}, ErrDimension
	}
	oldAllowed, err := JoinEspressoSetFamilies(original, dontCare)
	if err != nil {
		return EspressoCoverVerification{}, err
	}
	currentOutside, err := layout.coverDifference(current, oldAllowed, work)
	if err != nil {
		return EspressoCoverVerification{}, err
	}
	newAllowed, err := JoinEspressoSetFamilies(current, dontCare)
	if err != nil {
		return EspressoCoverVerification{}, err
	}
	originalOutside, err := layout.coverDifference(original, newAllowed, work)
	if err != nil {
		return EspressoCoverVerification{}, err
	}
	return EspressoCoverVerification{Equivalent: len(currentOutside.sets) == 0 && len(originalOutside.sets) == 0, CurrentOutside: currentOutside, OriginalOutside: originalOutside}, nil
}

type EspressoPLA struct {
	Layout            *EspressoCubeLayout
	On, DontCare, Off *EspressoSetFamily
	Labels            []string
	Filename          string
	Type              int
}

func NewEspressoPLA(layout *EspressoCubeLayout, on, dontCare, off *EspressoSetFamily, labels []string) (*EspressoPLA, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	for _, family := range []*EspressoSetFamily{on, dontCare, off} {
		if err := family.check(); err != nil {
			return nil, err
		}
		if family.bits != layout.bitCount {
			return nil, ErrDimension
		}
	}
	if len(labels) != 0 && len(labels) != layout.bitCount {
		return nil, ErrDimension
	}
	layoutClone, _ := layout.Clone()
	onClone, _ := on.Clone()
	dontCareClone, _ := dontCare.Clone()
	offClone, _ := off.Clone()
	return &EspressoPLA{Layout: layoutClone, On: onClone, DontCare: dontCareClone, Off: offClone, Labels: slices.Clone(labels)}, nil
}

func (pla *EspressoPLA) check() error {
	if pla == nil || pla.Layout == nil {
		return ErrDimension
	}
	if err := pla.Layout.check(); err != nil {
		return err
	}
	for _, family := range []*EspressoSetFamily{pla.On, pla.DontCare, pla.Off} {
		if err := family.check(); err != nil || family.bits != pla.Layout.bitCount {
			return ErrInvariant
		}
	}
	if len(pla.Labels) != 0 && len(pla.Labels) != pla.Layout.bitCount {
		return ErrInvariant
	}
	return nil
}

func (pla *EspressoPLA) Clone() (*EspressoPLA, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	result, _ := NewEspressoPLA(pla.Layout, pla.On, pla.DontCare, pla.Off, pla.Labels)
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}

// PermuteTo reorders all three covers by reference labels and returns a PLA
// using the reference variable grouping. Missing or ambiguous labels reject.
func (pla *EspressoPLA) PermuteTo(reference *EspressoPLA) (*EspressoPLA, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	if err := reference.check(); err != nil {
		return nil, err
	}
	if len(pla.Labels) == 0 || len(reference.Labels) == 0 {
		return nil, ErrDimension
	}
	positions := make(map[string]int, len(pla.Labels))
	for index, label := range pla.Labels {
		if label == "" {
			return nil, ErrDimension
		}
		if _, duplicate := positions[label]; duplicate {
			return nil, ErrInvariant
		}
		positions[label] = index
	}
	permutation := make([]int, len(reference.Labels))
	seen := make([]bool, len(pla.Labels))
	for index, label := range reference.Labels {
		source, exists := positions[label]
		if !exists || seen[source] {
			return nil, ErrDimension
		}
		seen[source] = true
		permutation[index] = source
	}
	on, err := pla.On.Permute(permutation)
	if err != nil {
		return nil, err
	}
	dontCare, err := pla.DontCare.Permute(permutation)
	if err != nil {
		return nil, err
	}
	off, err := pla.Off.Permute(permutation)
	if err != nil {
		return nil, err
	}
	result, err := NewEspressoPLA(reference.Layout, on, dontCare, off, reference.Labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}

func (pla *EspressoPLA) Verify(reference *EspressoPLA, options EspressoVerifyOptions) (EspressoCoverVerification, error) {
	permuted, err := pla.PermuteTo(reference)
	if err != nil {
		return EspressoCoverVerification{}, err
	}
	return reference.Layout.VerifyCovers(reference.On, permuted.On, permuted.DontCare, options)
}

type EspressoConsistency struct {
	Consistent                     bool
	OnDontCare, OnOff, DontCareOff *EspressoSetFamily
	Unspecified                    *EspressoSetFamily
}

func (pla *EspressoPLA) CheckConsistency(options EspressoVerifyOptions) (EspressoConsistency, error) {
	if err := pla.check(); err != nil {
		return EspressoConsistency{}, err
	}
	work, err := newEspressoVerifyWork(options)
	if err != nil {
		return EspressoConsistency{}, err
	}
	onDontCare, err := pla.Layout.intersectFamilies(pla.On, pla.DontCare, work)
	if err != nil {
		return EspressoConsistency{}, err
	}
	onOff, err := pla.Layout.intersectFamilies(pla.On, pla.Off, work)
	if err != nil {
		return EspressoConsistency{}, err
	}
	dontCareOff, err := pla.Layout.intersectFamilies(pla.DontCare, pla.Off, work)
	if err != nil {
		return EspressoConsistency{}, err
	}
	union, err := JoinEspressoSetFamilies(pla.On, pla.DontCare)
	if err != nil {
		return EspressoConsistency{}, err
	}
	union, err = JoinEspressoSetFamilies(union, pla.Off)
	if err != nil {
		return EspressoConsistency{}, err
	}
	full, _ := pla.Layout.FullCube()
	unspecified, err := pla.Layout.sharpCubeByCover(full, union, work)
	if err != nil {
		return EspressoConsistency{}, err
	}
	consistent := len(onDontCare.sets) == 0 && len(onOff.sets) == 0 && len(dontCareOff.sets) == 0 && len(unspecified.sets) == 0
	return EspressoConsistency{Consistent: consistent, OnDontCare: onDontCare, OnOff: onOff, DontCareOff: dontCareOff, Unspecified: unspecified}, nil
}
