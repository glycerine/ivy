package abc

// Exact small-input function-by-node enumeration from opt/dau/dauNpn.c.

import (
	"slices"

	"github.com/glycerine/insdict"
)

type DauFunctionEnumerationOptions struct {
	Inputs, TargetVariables, MaxNodes int
	UseTwo, Reduce                    bool
	MaxClasses                        int
	MaxCandidates                     uint64
	MaxCanonicalCandidates            uint64
	MaxCanonicalScheduleSteps         int
	Stop                              func() bool
}

type DauFunctionClass struct {
	Truth          uint64
	Nodes, Support int
}

type DauFunctionLevel struct {
	Nodes, Start, End, NewTarget, TwoNew int
	Candidates                           uint64
}

type DauFunctionEnumeration struct {
	inputs, target int
	classes        []DauFunctionClass
	levels         []DauFunctionLevel
}

func (enumeration *DauFunctionEnumeration) Inputs() int {
	if enumeration == nil {
		return 0
	}
	return enumeration.inputs
}

func (enumeration *DauFunctionEnumeration) TargetVariables() int {
	if enumeration == nil {
		return 0
	}
	return enumeration.target
}

func (enumeration *DauFunctionEnumeration) Classes() []DauFunctionClass {
	if enumeration == nil {
		return nil
	}
	return slices.Clone(enumeration.classes)
}

func (enumeration *DauFunctionEnumeration) Levels() []DauFunctionLevel {
	if enumeration == nil {
		return nil
	}
	return slices.Clone(enumeration.levels)
}

func logicalDauTruthTable(variables int, truth uint64) (*TruthTable, error) {
	if variables < 0 || variables > 6 {
		return nil, ErrDimension
	}
	assignments := 1 << uint(variables)
	mask := uint64(1)<<uint(assignments) - 1
	if truth&^mask != 0 {
		return nil, ErrDimension
	}
	table, err := NewConstantTruthTable(variables, false, 0)
	if err != nil {
		return nil, err
	}
	for assignment := 0; assignment < assignments; assignment++ {
		if truth&(1<<uint(assignment)) != 0 {
			if err := table.Set(uint64(assignment), true); err != nil {
				return nil, err
			}
		}
	}
	return table, nil
}

func dauLogicalSupport(truth uint64, inputs int) []int {
	result := make([]int, 0, inputs)
	assignments := 1 << uint(inputs)
	for variable := 0; variable < inputs; variable++ {
		used := false
		for assignment := 0; assignment < assignments; assignment++ {
			other := assignment ^ (1 << uint(variable))
			if (truth>>uint(assignment))&1 != (truth>>uint(other))&1 {
				used = true
				break
			}
		}
		if used {
			result = append(result, variable)
		}
	}
	return result
}

func canonicalizeDauFunctionTruth(truth uint64, inputs int, tables []*DauNPNClasses) (uint64, int, error) {
	canonicalizer := dauFunctionCanonicalizer{inputs: inputs, tables: tables}
	return canonicalizer.canonicalize(truth)
}

type dauFunctionCanonicalizer struct {
	inputs           int
	tables           []*DauNPNClasses
	remaining        uint64
	limited          bool
	maxScheduleSteps int
	stop             func() bool
}

func (canonicalizer *dauFunctionCanonicalizer) canonicalize(truth uint64) (uint64, int, error) {
	support := dauLogicalSupport(truth, canonicalizer.inputs)
	compact := uint64(0)
	for assignment := 0; assignment < 1<<uint(len(support)); assignment++ {
		source := 0
		for position, variable := range support {
			source |= ((assignment >> uint(position)) & 1) << uint(variable)
		}
		if truth&(1<<uint(source)) != 0 {
			compact |= 1 << uint(assignment)
		}
	}
	canonical := uint64(0)
	if len(support) <= 4 {
		if len(canonicalizer.tables) <= len(support) || canonicalizer.tables[len(support)] == nil {
			return 0, 0, ErrDimension
		}
		var err error
		canonical, err = canonicalizer.tables[len(support)].Canonical(compact)
		if err != nil {
			return 0, 0, err
		}
	} else {
		if canonicalizer.stop != nil && canonicalizer.stop() {
			return 0, 0, ErrCancelled
		}
		if canonicalizer.limited && canonicalizer.remaining == 0 {
			return 0, 0, ErrLimit
		}
		table, err := logicalDauTruthTable(len(support), compact)
		if err != nil {
			return 0, 0, err
		}
		candidateLimit := int(^uint(0) >> 1)
		if canonicalizer.limited && canonicalizer.remaining < uint64(candidateLimit) {
			candidateLimit = int(canonicalizer.remaining)
		}
		minimum, err := table.MinimizeNPN(TruthExactOptions{MaxScheduleSteps: canonicalizer.maxScheduleSteps, MaxCandidates: candidateLimit, Stop: canonicalizer.stop})
		if err != nil {
			return 0, 0, err
		}
		if canonicalizer.limited {
			canonicalizer.remaining -= uint64(minimum.Candidates)
		}
		canonical = minimum.Table.words[0] & dauTruthMask(len(support))
	}
	expanded := uint64(0)
	lowMask := (1 << uint(len(support))) - 1
	for assignment := 0; assignment < 1<<uint(canonicalizer.inputs); assignment++ {
		if canonical&(1<<uint(assignment&lowMask)) != 0 {
			expanded |= 1 << uint(assignment)
		}
	}
	return expanded, len(support), nil
}

func dauFunctionCofactorValue(truth uint64, assignment, variable int, one bool) bool {
	assignment &^= 1 << uint(variable)
	if one {
		assignment |= 1 << uint(variable)
	}
	return truth&(1<<uint(assignment)) != 0
}

func dauGateValue(kind, first, second, assignment int) bool {
	a := assignment&(1<<uint(first)) != 0
	b := assignment&(1<<uint(second)) != 0
	switch kind {
	case 0:
		return a && b
	case 1:
		return a && !b
	case 2:
		return !a && b
	case 3:
		return !a && !b
	case 4:
		return a != b
	case 5:
		return a || b
	}
	return false
}

func dauFunctionGateCandidate(truth uint64, inputs, variable, gateKind, first, second int, reverse bool) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(inputs); assignment++ {
		gate := dauGateValue(gateKind, first, second, assignment)
		if reverse {
			gate = !gate
		}
		value := dauFunctionCofactorValue(truth, assignment, variable, gate)
		if value {
			result |= 1 << uint(assignment)
		}
	}
	return result
}

func dauFunctionTwoGateCandidate(truth uint64, inputs, firstVariable, secondVariable, firstGate, secondGate int) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(inputs); assignment++ {
		cofactor := func(first, second bool) bool {
			source := assignment &^ (1 << uint(firstVariable)) &^ (1 << uint(secondVariable))
			if first {
				source |= 1 << uint(firstVariable)
			}
			if second {
				source |= 1 << uint(secondVariable)
			}
			return truth&(1<<uint(source)) != 0
		}
		gate1 := dauGateValue(firstGate, firstVariable, secondVariable, assignment)
		cofactor0 := cofactor(false, gate1)
		cofactor1 := cofactor(true, gate1)
		gate2 := dauGateValue(secondGate, firstVariable, secondVariable, assignment)
		value := cofactor0
		if gate2 {
			value = cofactor1
		}
		if value {
			result |= 1 << uint(assignment)
		}
	}
	return result
}

// EnumerateDauFunctions preserves the source candidate order and first-class
// insertion cost for the exact three/four-input domain. It returns owned data
// instead of writing tables or progress reports.
func EnumerateDauFunctions(options DauFunctionEnumerationOptions) (*DauFunctionEnumeration, error) {
	if options.Inputs < 3 || options.Inputs > 6 || options.TargetVariables < 3 || options.TargetVariables > options.Inputs || options.MaxNodes < 0 || options.MaxNodes > 31 || options.MaxClasses < 0 || options.MaxCanonicalScheduleSteps < 0 {
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
	tables := make([]*DauNPNClasses, options.Inputs+1)
	for variables := 0; variables <= min(options.Inputs, 4); variables++ {
		var err error
		tables[variables], err = EnumerateDauNPNClasses(variables, DauNPNClassOptions{})
		if err != nil {
			return nil, err
		}
	}
	canonicalizer := dauFunctionCanonicalizer{inputs: options.Inputs, tables: tables, maxScheduleSteps: options.MaxCanonicalScheduleSteps, stop: options.Stop}
	if options.MaxCanonicalCandidates != 0 {
		canonicalizer.limited = true
		canonicalizer.remaining = options.MaxCanonicalCandidates
	}
	index := insdict.NewDict[uint64, int]()
	result := &DauFunctionEnumeration{inputs: options.Inputs, target: options.TargetVariables}
	insert := func(truth uint64, nodes int) (bool, error) {
		canonical, support, err := canonicalizer.canonicalize(truth)
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
	variableTruth := uint64(0)
	for assignment := 0; assignment < 1<<uint(options.Inputs); assignment++ {
		if assignment&1 == 0 {
			variableTruth |= 1 << uint(assignment)
		}
	}
	if _, err := insert(variableTruth, 0); err != nil {
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
		expand := !(options.Reduce && nodes == options.MaxNodes)
		twoNew := 0
		for entry := limits[nodes-1]; entry < limits[nodes]; entry++ {
			class := result.classes[entry]
			table, err := logicalDauTruthTable(class.Support, class.Truth&((uint64(1)<<uint(1<<uint(class.Support)))-1))
			if err != nil {
				return nil, err
			}
			analysis, err := AnalyzeDauNPN(table, DauNPNAnalysisOptions{Stop: options.Stop})
			if err != nil {
				return nil, err
			}
			for variable := 0; variable < class.Support; variable++ {
				if analysis.SymmetricSecondVariables[variable] {
					continue
				}
				try := func(candidate uint64) error {
					if err := take(); err != nil {
						return err
					}
					_, err := insert(candidate, nodes)
					return err
				}
				if class.Support < options.Inputs && expand {
					newVariable := class.Support
					for _, candidate := range []uint64{
						dauFunctionGateCandidate(class.Truth, options.Inputs, variable, 0, variable, newVariable, false),
						dauFunctionGateCandidate(class.Truth, options.Inputs, variable, 0, variable, newVariable, true),
						dauFunctionGateCandidate(class.Truth, options.Inputs, variable, 4, variable, newVariable, false),
					} {
						if err := try(candidate); err != nil {
							return nil, err
						}
					}
				}
				if expand {
					for other := 0; other < class.Support; other++ {
						if other == variable || analysis.SymmetricSecondVariables[other] && other != variable+1 {
							continue
						}
						for _, spec := range [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {4, 0}} {
							if err := try(dauFunctionGateCandidate(class.Truth, options.Inputs, variable, spec[0], variable, other, spec[1] != 0)); err != nil {
								return nil, err
							}
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
						specs := [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {2, 0}, {2, 1}, {3, 0}, {3, 1}, {4, 0}, {4, 1}}
						for _, spec := range specs {
							if err := try(dauFunctionGateCandidate(class.Truth, options.Inputs, variable, spec[0], second, first, spec[1] != 0)); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
		if options.UseTwo && nodes > 2 && expand {
			for entry := limits[nodes-2]; entry < limits[nodes-1]; entry++ {
				class := result.classes[entry]
				for first := 0; first < class.Support; first++ {
					for second := 0; second < class.Support; second++ {
						if second == first {
							continue
						}
						for gate1 := 0; gate1 < 5; gate1++ {
							for gate2 := gate1 + 1; gate2 < 5; gate2++ {
								if err := take(); err != nil {
									return nil, err
								}
								added, err := insert(dauFunctionTwoGateCandidate(class.Truth, options.Inputs, first, second, gate1, gate2), nodes)
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
		}
		limits[nodes+1] = len(result.classes)
		newTarget := 0
		for _, class := range result.classes[start:] {
			if class.Support <= options.TargetVariables {
				newTarget++
			}
		}
		result.levels = append(result.levels, DauFunctionLevel{Nodes: nodes, Start: start, End: len(result.classes), NewTarget: newTarget, TwoNew: twoNew, Candidates: candidates})
		if newTarget == 0 {
			break
		}
	}
	return result, nil
}
