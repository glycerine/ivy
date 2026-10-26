package abc

// Exact small-function NPN class enumeration from opt/dau/dauNpn.c:Dau_TruthEnum.

import "slices"

type DauNPNClassOptions struct {
	MaxVariables, MaxClasses int
	MaxTransforms            uint64
	Stop                     func() bool
}

type DauNPNClasses struct {
	variables       int
	representatives []uint64
	representative  []uint64
}

func (classes *DauNPNClasses) Variables() int {
	if classes == nil {
		return 0
	}
	return classes.variables
}

func (classes *DauNPNClasses) Representatives() []uint64 {
	if classes == nil {
		return nil
	}
	return slices.Clone(classes.representatives)
}

func (classes *DauNPNClasses) Canonical(truth uint64) (uint64, error) {
	if classes == nil || classes.variables < 0 || classes.variables > 4 || len(classes.representative) == 0 {
		return 0, ErrDimension
	}
	bits := 1 << uint(classes.variables)
	mask := uint64(1)<<uint(bits) - 1
	if truth&^mask != 0 {
		return 0, ErrDimension
	}
	if truth&(1<<uint(bits-1)) != 0 {
		truth = (^truth) & mask
	}
	return classes.representative[truth], nil
}

func dauNPNPermutations(variables int) [][]int {
	current := make([]int, variables)
	for i := range current {
		current[i] = i
	}
	result := make([][]int, 0)
	var generate func(int)
	generate = func(position int) {
		if position == variables {
			result = append(result, slices.Clone(current))
			return
		}
		for i := position; i < variables; i++ {
			current[position], current[i] = current[i], current[position]
			generate(position + 1)
			current[position], current[i] = current[i], current[position]
		}
	}
	generate(0)
	return result
}

func transformDauNPNTruth(truth uint64, variables int, permutation []int, phase int) uint64 {
	assignments := 1 << uint(variables)
	var result uint64
	for target := 0; target < assignments; target++ {
		source := 0
		for targetVariable, sourceVariable := range permutation {
			bit := (target >> uint(targetVariable)) & 1
			bit ^= (phase >> uint(targetVariable)) & 1
			source |= bit << uint(sourceVariable)
		}
		if truth&(1<<uint(source)) != 0 {
			result |= 1 << uint(target)
		}
	}
	return result
}

// EnumerateDauNPNClasses exhaustively enumerates output-negation, input-phase
// and input-permutation classes in ascending representative order. Four inputs
// require 32,768 normalized functions; larger exhaustive tables are rejected.
func EnumerateDauNPNClasses(variables int, options DauNPNClassOptions) (*DauNPNClasses, error) {
	if options.MaxVariables < 0 || options.MaxClasses < 0 {
		return nil, ErrDimension
	}
	maximumVariables := options.MaxVariables
	if maximumVariables == 0 {
		maximumVariables = 4
	}
	if variables < 0 || variables > maximumVariables || variables > 4 {
		return nil, ErrDimension
	}
	maximumClasses := options.MaxClasses
	if maximumClasses == 0 {
		maximumClasses = 1 << 20
	}
	maximumTransforms := options.MaxTransforms
	if maximumTransforms == 0 {
		maximumTransforms = ^uint64(0)
	}
	truthBits := 1 << uint(variables)
	mask := uint64(1)<<uint(truthBits) - 1
	normalizedFunctions := 1 << uint(truthBits-1)
	representative := make([]uint64, normalizedFunctions)
	marked := make([]bool, normalizedFunctions)
	permutations := dauNPNPermutations(variables)
	result := &DauNPNClasses{variables: variables, representative: representative}
	used := uint64(0)
	for candidate := 0; candidate < normalizedFunctions; candidate++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if marked[candidate] {
			continue
		}
		if len(result.representatives) >= maximumClasses {
			return nil, ErrLimit
		}
		representativeTruth := uint64(candidate)
		result.representatives = append(result.representatives, representativeTruth)
		for _, permutation := range permutations {
			for phase := 0; phase < 1<<uint(variables); phase++ {
				if used >= maximumTransforms {
					return nil, ErrLimit
				}
				used++
				transformed := transformDauNPNTruth(representativeTruth, variables, permutation, phase)
				if transformed&(1<<uint(truthBits-1)) != 0 {
					transformed = (^transformed) & mask
				}
				marked[transformed] = true
				representative[transformed] = representativeTruth
			}
		}
	}
	return result, nil
}
