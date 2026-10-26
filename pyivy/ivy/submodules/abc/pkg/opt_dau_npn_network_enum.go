package abc

// Exact small-input network enumeration from opt/dau/dauNpn.c.

import (
	"slices"

	"github.com/glycerine/insdict"
)

type DauNetworkEnumerationOptions struct {
	Variables     int
	MaxNodes      int
	UseTwo        bool
	MaxClasses    int
	MaxCandidates uint64
	Stop          func() bool
}

type DauNetworkLevel struct {
	Nodes, Start, End, NewProper, TwoNew int
	Candidates                           uint64
}

type DauNetworkEnumeration struct {
	variables int
	classes   []DauFunctionClass
	levels    []DauNetworkLevel
}

func (enumeration *DauNetworkEnumeration) Variables() int {
	if enumeration == nil {
		return 0
	}
	return enumeration.variables
}

func (enumeration *DauNetworkEnumeration) Classes() []DauFunctionClass {
	if enumeration == nil {
		return nil
	}
	return slices.Clone(enumeration.classes)
}

func (enumeration *DauNetworkEnumeration) Levels() []DauNetworkLevel {
	if enumeration == nil {
		return nil
	}
	return slices.Clone(enumeration.levels)
}

func dauNetworkTwoCandidate(truth uint64, inputs, first, second int, reverse bool) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(inputs); assignment++ {
		cofactor := func(firstValue, secondValue bool) bool {
			source := assignment &^ (1 << uint(first)) &^ (1 << uint(second))
			if firstValue {
				source |= 1 << uint(first)
			}
			if secondValue {
				source |= 1 << uint(second)
			}
			return truth&(1<<uint(source)) != 0
		}
		inner := dauGateValue(4, first, second, assignment)
		cofactor0 := cofactor(false, !inner)
		cofactor1 := cofactor(true, !inner)
		outer := dauGateValue(0, first, second, assignment)
		if reverse {
			outer = !outer
		}
		value := cofactor0
		if outer {
			value = cofactor1
		}
		if value {
			result |= 1 << uint(assignment)
		}
	}
	return result
}

// EnumerateDauNetworks preserves Dau_NetworkEnum's source frontier and gate
// order for one through four variables. Exact in-memory NPN tables replace the
// native tableW*.data dependency, and owned level records replace printing.
func EnumerateDauNetworks(options DauNetworkEnumerationOptions) (*DauNetworkEnumeration, error) {
	if options.Variables < 1 || options.Variables > 4 || options.MaxNodes < 0 || options.MaxNodes > 31 || options.MaxClasses < 0 {
		return nil, ErrDimension
	}
	maxClasses := options.MaxClasses
	if maxClasses == 0 {
		maxClasses = 1 << 20
	}
	maxCandidates := options.MaxCandidates
	if maxCandidates == 0 {
		maxCandidates = ^uint64(0)
	}
	tables := make([]*DauNPNClasses, options.Variables+1)
	for variables := range tables {
		var err error
		tables[variables], err = EnumerateDauNPNClasses(variables, DauNPNClassOptions{})
		if err != nil {
			return nil, err
		}
	}
	result := &DauNetworkEnumeration{variables: options.Variables}
	index := insdict.NewDict[uint64, int]()
	insert := func(truth uint64, nodes int) (bool, error) {
		canonical, support, err := canonicalizeDauFunctionTruth(truth, options.Variables, tables)
		if err != nil {
			return false, err
		}
		if _, found := index.Get2(canonical); found {
			return false, nil
		}
		if len(result.classes) >= maxClasses {
			return false, ErrLimit
		}
		index.Set(canonical, len(result.classes))
		result.classes = append(result.classes, DauFunctionClass{Truth: canonical, Nodes: nodes, Support: support})
		return true, nil
	}
	if _, err := insert(0, 0); err != nil {
		return nil, err
	}
	variable := uint64(0)
	for assignment := 0; assignment < 1<<uint(options.Variables); assignment++ {
		if assignment&1 == 0 {
			variable |= 1 << uint(assignment)
		}
	}
	if _, err := insert(variable, 0); err != nil {
		return nil, err
	}
	limits := make([]int, options.MaxNodes+2)
	limits[0], limits[1] = 1, 2
	candidates := uint64(0)
	take := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if candidates >= maxCandidates {
			return ErrLimit
		}
		candidates++
		return nil
	}
	for nodes := 1; nodes <= options.MaxNodes; nodes++ {
		start := len(result.classes)
		try := func(candidate uint64) (bool, error) {
			if err := take(); err != nil {
				return false, err
			}
			return insert(candidate, nodes)
		}
		for entry := limits[nodes-1]; entry < limits[nodes]; entry++ {
			class := result.classes[entry]
			for variable := 0; variable < class.Support; variable++ {
				if class.Support < options.Variables {
					newVariable := class.Support
					for _, candidate := range []uint64{
						dauFunctionGateCandidate(class.Truth, options.Variables, variable, 0, variable, newVariable, false),
						dauFunctionGateCandidate(class.Truth, options.Variables, variable, 0, variable, newVariable, true),
						dauFunctionGateCandidate(class.Truth, options.Variables, variable, 4, variable, newVariable, false),
					} {
						if _, err := try(candidate); err != nil {
							return nil, err
						}
					}
				}
				for other := 0; other < class.Support; other++ {
					if other == variable {
						continue
					}
					for _, spec := range [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {4, 0}} {
						candidate := dauFunctionGateCandidate(class.Truth, options.Variables, variable, spec[0], variable, other, spec[1] != 0)
						if _, err := try(candidate); err != nil {
							return nil, err
						}
					}
				}
				for first := 0; first < class.Support; first++ {
					if first == variable {
						continue
					}
					for second := first + 1; second < class.Support; second++ {
						if second == variable {
							continue
						}
						for _, spec := range [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {2, 0}, {2, 1}, {3, 0}, {3, 1}, {4, 0}, {4, 1}} {
							candidate := dauFunctionGateCandidate(class.Truth, options.Variables, variable, spec[0], second, first, spec[1] != 0)
							if _, err := try(candidate); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
		twoNew := 0
		if options.UseTwo && nodes > 1 {
			for entry := limits[nodes-2]; entry < limits[nodes-1]; entry++ {
				class := result.classes[entry]
				for first := 0; first < class.Support; first++ {
					for second := 0; second < class.Support; second++ {
						if second == first {
							continue
						}
						for _, reverse := range []bool{false, true} {
							added, err := try(dauNetworkTwoCandidate(class.Truth, options.Variables, first, second, reverse))
							if err != nil {
								return nil, err
							}
							if added {
								twoNew++
							}
						}
					}
				}
			}
		}
		limits[nodes+1] = len(result.classes)
		newProper := 0
		for _, class := range result.classes[start:] {
			if class.Support < options.Variables {
				newProper++
			}
		}
		result.levels = append(result.levels, DauNetworkLevel{Nodes: nodes, Start: start, End: len(result.classes), NewProper: newProper, TwoNew: twoNew, Candidates: candidates})
		if start == len(result.classes) {
			break
		}
	}
	return result, nil
}
