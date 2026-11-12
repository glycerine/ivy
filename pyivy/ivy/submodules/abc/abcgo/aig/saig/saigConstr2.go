package saig

import (
	"fmt"
	"math/rand"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ManCreateIndMiter retains ABC's entry-point spelling while accepting
// native candidate edge groups and returning a native AIG.
func Saig_ManCreateIndMiter(src *aig.Graph, candidates FunctionalCandidates) *aig.Graph {
	graph, err := InductionMiter(src, candidates)
	if err != nil {
		panic(err)
	}
	return graph
}

// Saig_ManFilterUsingInd returns only candidates proved inductive by SAT.
func Saig_ManFilterUsingInd(src *aig.Graph, candidates FunctionalCandidates, conflicts, propagations int32, verbose int32) FunctionalCandidates {
	filtered, err := FilterInductiveCandidates(src, candidates, int64(conflicts), int64(propagations))
	if err != nil {
		panic(err)
	}
	if verbose != 0 {
		fmt.Printf("Filtered %d candidates to %d inductive constraints.\n", candidateCount(candidates), candidateCount(filtered))
	}
	return filtered
}

// Saig_ManUnrollCOI returns a combinational unrolling of output zero at the
// final frame. The native implementation copies graph state locally.
func Saig_ManUnrollCOI(src *aig.Graph, frames int32) *aig.Graph {
	graph, err := UnrollProperty(src, int(frames))
	if err != nil {
		panic(err)
	}
	return graph
}

// Saig_ManUnrollCOI_ preserves the historical alternate entry point. Both
// source implementations unroll the selected property cone.
func Saig_ManUnrollCOI_(src *aig.Graph, frames int32) *aig.Graph {
	return Saig_ManUnrollCOI(src, frames)
}

// Ssw_ManFindDirectImplications finds functional constraints with SAT and
// filters them through the two-frame induction miter.
func Ssw_ManFindDirectImplications(src *aig.Graph, frames, conflicts, propagations, verbose int32) FunctionalCandidates {
	return detectFunctionalOrPanic(src, frames, conflicts, propagations, verbose)
}

// Saig_ManDetectConstrFunc is the SAT implication detector used by the older
// pattern-guided source path. The Go port proves implications directly.
func Saig_ManDetectConstrFunc(src *aig.Graph, frames, conflicts, propagations, verbose int32) FunctionalCandidates {
	return detectFunctionalOrPanic(src, frames, conflicts, propagations, verbose)
}

func detectFunctionalOrPanic(src *aig.Graph, frames, conflicts, propagations, verbose int32) FunctionalCandidates {
	candidates, err := DetectFunctionalConstraints(src, int(frames), int64(conflicts), int64(propagations))
	if err != nil {
		panic(err)
	}
	if verbose != 0 {
		fmt.Printf("Found %d inductive functional constraints across %d frames.\n", candidateCount(candidates), frames)
	}
	return candidates
}

// Saig_ManDetectConstrFuncTest runs the selected detector for its diagnostics.
func Saig_ManDetectConstrFuncTest(src *aig.Graph, frames, conflicts, propagations, oldAlgorithm, verbose int32) {
	if oldAlgorithm != 0 {
		_ = Saig_ManDetectConstrFunc(src, frames, conflicts, propagations, verbose)
		return
	}
	_ = Ssw_ManFindDirectImplications(src, frames, conflicts, propagations, verbose)
}

// Saig_ManDupUnfoldConstrsFunc detects and pipelines functional constraints.
func Saig_ManDupUnfoldConstrsFunc(src *aig.Graph, frames, conflicts, propagations, oldAlgorithm, verbose int32) *aig.Graph {
	var candidates FunctionalCandidates
	if oldAlgorithm != 0 {
		candidates = Saig_ManDetectConstrFunc(src, frames, conflicts, propagations, verbose)
	} else {
		candidates = Ssw_ManFindDirectImplications(src, frames, conflicts, propagations, verbose)
	}
	graph, err := UnfoldFunctionalConstraints(src, candidates)
	if err != nil {
		panic(err)
	}
	return graph
}

// Saig_ManDupFoldConstrsFunc folds trailing constraint outputs into a sticky
// failure state and gates property outputs with the healthy-state condition.
func Saig_ManDupFoldConstrsFunc(src *aig.Graph, complemented, verbose int32) *aig.Graph {
	graph, err := FoldFunctionalConstraints(src, complemented != 0)
	if err != nil {
		panic(err)
	}
	return graph
}

// Ssw_ManProfileConstraints samples sequential traces and reports output-one
// rates. It returns one when the requested simulation completes.
func Ssw_ManProfileConstraints(src *aig.Graph, words, frames, verbose int32) int32 {
	if src == nil || words <= 0 || frames <= 0 {
		return 0
	}
	inputs, outputs := src.Inputs(), src.Outputs()
	regs := int(aig.Aig_ManRegNum(src))
	piCount := len(inputs) - regs
	poCount := len(outputs) - regs
	if regs < 0 || piCount < 0 || poCount < 0 {
		return 0
	}
	rng := rand.New(rand.NewSource(1))
	counts := make([]int64, poCount)
	samples := int64(words) * 32 * int64(frames)
	state := make([]bool, regs)
	values := make([]bool, len(inputs))
	for sample := 0; sample < int(words)*32; sample++ {
		for i := range state {
			state[i] = rng.Intn(2) != 0
		}
		for frame := int32(0); frame < frames; frame++ {
			for i := 0; i < piCount; i++ {
				values[i] = rng.Intn(2) != 0
			}
			copy(values[piCount:], state)
			result, err := src.Evaluate(values)
			if err != nil {
				return 0
			}
			for i := 0; i < poCount; i++ {
				if result[i] {
					counts[i]++
				}
			}
			copy(state, result[poCount:])
		}
	}
	if verbose != 0 {
		for i, count := range counts {
			fmt.Printf("Output %d ProbOne = %.6f\n", i, float64(count)/float64(samples))
		}
	}
	return 1
}

// Saig_ManDupUnfoldConstrsFuncNative and Saig_ManDupFoldConstrsFuncNative
// expose error-returning variants for Go callers.
func Saig_ManDupUnfoldConstrsFuncNative(src *aig.Graph, frames int, conflicts, propagations int64) (*aig.Graph, error) {
	candidates, err := DetectFunctionalConstraints(src, frames, conflicts, propagations)
	if err != nil {
		return nil, err
	}
	return UnfoldFunctionalConstraints(src, candidates)
}

func Saig_ManDupFoldConstrsFuncNative(src *aig.Graph, complemented bool) (*aig.Graph, error) {
	return FoldFunctionalConstraints(src, complemented)
}
