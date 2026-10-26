package abc

// SCC phase filtering from opt/dau/dauCanon.c: Abc_TgReorderFGrps,
// Abc_TgPermCostScc, Abc_TgRecordPhase and Abc_TgRecordPhase1.

import (
	"math"
	"math/bits"
	"slices"
)

type TruthSCCPhaseMode uint8

const (
	TruthSCCPhaseTotal     TruthSCCPhaseMode = iota // AdjSE: minimize total coefficient only
	TruthSCCPhaseSignature                          // minimize total, then sorted group coefficients
	TruthSCCPhaseCost                               // minimize permutation cost, then total and group coefficients
)

type TruthSCCPhaseOptions struct {
	Mode              TruthSCCPhaseMode
	HighSymmetry      bool
	MaxConfigurations int // zero: one million
	Stop              func() bool
}

// TruthSCCPhaseFilter describes the Gray-order masks retained by C's filter.
// PhaseGroups contains source input labels ordered stably by symmetry weight.
// Mask bits select these representatives' nonzero-weight symmetry chains.
// Fixed phases produce no masks. CandidateLog2 is floor(log2(len(Masks))).
// PermutationCost uses C's rounded 100*log2(run-size!) tie costs. Total-only
// mode leaves Coefficients nil and PermutationCost zero. PhaseInputs is the
// number of phase groups, and Visited counts signatures evaluated. All slices
// own storage; masks apply after the initial normalization/reduction/permutation.
type TruthSCCPhaseFilter struct {
	PhaseGroups                                          []int
	Masks                                                []uint64
	Total                                                uint64
	Coefficients                                         []uint64
	PermutationCost, CandidateLog2, PhaseInputs, Visited int
}

func truthSCCPermutationCost(groups []TruthVariableGroup, coefficients []uint64) int {
	native := [17]int{0, 0, 100, 258, 458, 691, 949, 1230, 1530, 1847, 2179, 2525, 2884, 3254, 3634, 4025, 4425}
	factorialCost := func(n int) int {
		if n < len(native) {
			return native[n]
		}
		sum := 0.0
		for i := 2; i <= n; i++ {
			sum += math.Log2(float64(i))
		}
		return int(math.Round(sum * 100))
	}
	cost := 0
	for _, group := range groups {
		run := 1
		for j := group.Start + 1; j < group.Start+group.Size; j++ {
			if coefficients[j] == coefficients[j-1] {
				run++
			} else {
				cost += factorialCost(run)
				run = 1
			}
		}
		if group.Size > 0 {
			cost += factorialCost(run)
		}
	}
	return cost
}

// FilterAdaptivePhases visits first-group phase assignments using C's stable
// weight order and Gray walk, retaining all minimum signatures in visit order.
// It ports phase selection only; it does not perform permutation enumeration.
// Cancellation or budget exhaustion returns no partial result. uint64 signature
// accumulation extends C's signed-int arithmetic without overflow.
func (t *TruthTable) FilterAdaptivePhases(options TruthSCCPhaseOptions) (TruthSCCPhaseFilter, error) {
	fail := func(e error) (TruthSCCPhaseFilter, error) { return TruthSCCPhaseFilter{}, e }
	if options.Mode > TruthSCCPhaseCost || options.MaxConfigurations < 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	reduction, e := t.ReduceInputSymmetry(options.HighSymmetry)
	if e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return newTruthAdaptiveState(reduction).filterSCCPhases(options)
}

func (s *truthAdaptiveState) filterSCCPhases(options TruthSCCPhaseOptions) (TruthSCCPhaseFilter, error) {
	fail := func(e error) (TruthSCCPhaseFilter, error) { return TruthSCCPhaseFilter{}, e }
	if options.Mode > TruthSCCPhaseCost || options.MaxConfigurations < 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	work := s.clone()
	result := TruthSCCPhaseFilter{}
	if !work.phased && len(work.groups) > 0 {
		g := work.groups[0]
		result.PhaseGroups = slices.Clone(work.representatives[g.Start : g.Start+g.Size])
		for i := 1; i < len(result.PhaseGroups); i++ {
			v := result.PhaseGroups[i]
			j := i
			for j > 0 && work.weights[result.PhaseGroups[j-1]] > work.weights[v] {
				result.PhaseGroups[j] = result.PhaseGroups[j-1]
				j--
			}
			result.PhaseGroups[j] = v
		}
	}
	result.PhaseInputs = len(result.PhaseGroups)
	limit := options.MaxConfigurations
	if limit == 0 {
		limit = 1 << 20
	}
	for step := uint64(0); step < uint64(1)<<uint(result.PhaseInputs); step++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		if result.Visited >= limit {
			return fail(ErrLimit)
		}
		if step > 0 {
			work.flipChainBySource(result.PhaseGroups[bits.TrailingZeros64(step)])
		}
		result.Visited++
		coefficients, e := work.table.ShiftedCofactorCoefficients()
		if e != nil {
			return fail(e)
		}
		var groupCoefficients []uint64
		cost := 0
		if options.Mode != TruthSCCPhaseTotal {
			groupCoefficients, e = work.shiftedGroupCoefficients(true)
			if e != nil {
				return fail(e)
			}
			cost = truthSCCPermutationCost(work.groups, groupCoefficients)
		}
		compare := 0
		if step > 0 {
			if options.Mode == TruthSCCPhaseCost {
				if cost < result.PermutationCost {
					compare = -1
				} else if cost > result.PermutationCost {
					compare = 1
				}
			}
			if compare == 0 {
				if coefficients.Total < result.Total {
					compare = -1
				} else if coefficients.Total > result.Total {
					compare = 1
				}
			}
			if compare == 0 && options.Mode != TruthSCCPhaseTotal {
				compare = slices.Compare(groupCoefficients, result.Coefficients)
			}
		}
		if step == 0 || compare < 0 {
			result.Total = coefficients.Total
			result.Coefficients = groupCoefficients
			result.PermutationCost = cost
			result.Masks = result.Masks[:0]
		}
		if !work.phased && (step == 0 || compare <= 0) {
			result.Masks = append(result.Masks, step^(step>>1))
		}
	}
	if len(result.Masks) > 0 {
		result.CandidateLog2 = bits.Len(uint(len(result.Masks))) - 1
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}
