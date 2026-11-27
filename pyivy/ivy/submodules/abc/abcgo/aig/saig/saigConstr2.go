package saig

import (
	"fmt"
	"math/bits"
	"time"

	"github.com/glycerine/abcgo/misc/util"

	"github.com/glycerine/abcgo/aig/aig"
)

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

// Saig_ManDetectConstrFunc is the SAT implication detector used by the older
// pattern-guided source path. The Go port proves implications directly.
func Saig_ManDetectConstrFunc(src *aig.Graph, frames, conflicts, propagations, verbose int32) FunctionalCandidates {
	return detectFunctionalOrPanic(src, frames, conflicts, propagations, verbose)
}

func detectFunctionalOrPanic(src *aig.Graph, frames, conflicts, propagations, verbose int32) FunctionalCandidates {
	if poNum := Saig_ManPoNum(src); poNum != 1 { // as Saig_ManDetectConstrFunc
		fmt.Printf("The number of outputs is different from 1.\n")
		return nil
	}
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

// Ssw_ManProfileConstraints simulates nWords random words for nFrames
// frames from the all-zero state (saigConstr2.c): POs report failing
// patterns not blocked by constraints, and with fVerbose each PO prints its
// probability of one, of one under a failing PO, and its phase.
func Ssw_ManProfileConstraints(src *aig.Graph, words, frames, verbose int32) int32 {
	if src == nil || words <= 0 || frames <= 0 {
		return 0
	}
	start := time.Now()
	nodes, inputs, outputs := src.Nodes(), src.Inputs(), src.Outputs()
	regs := int(aig.Aig_ManRegNum(src))
	piCount, poCount := len(inputs)-regs, len(outputs)-regs
	constrs := int(Saig_ManConstrNum(src))
	if verbose != 0 {
		fmt.Printf("Simulating %d nodes and %d flops for %d frames with %d words... ", aig.Aig_ManNodeNum(src), regs, frames, words)
	}
	aig.Aig_ManRandom(1)
	info := make([][]uint32, len(nodes))
	for id := range info {
		info[id] = make([]uint32, words)
	}
	probs, probs2 := make([]int32, poCount), make([]int32, poCount)
	for id, n := range nodes {
		if n.Kind == aig.Constant {
			for w := range info[id] {
				info[id][w] = ^uint32(0)
			}
		}
	}
	mask, mask2 := make([]uint32, words), make([]uint32, words)
	child := func(e aig.Edge) []uint32 { return info[e.NodeID()] }
	for f := int32(0); f < frames; f++ {
		for _, pi := range inputs[:piCount] {
			for w := range info[pi.NodeID()] {
				info[pi.NodeID()][w] = aig.Aig_ManRandom(0)
			}
		}
		for i := 0; i < regs; i++ {
			copy(info[inputs[piCount+i].NodeID()], info[outputs[poCount+i].NodeID()])
		}
		for id, n := range nodes {
			if !src.ValidEdge(aig.Edge(id)<<1) || n.Kind != aig.And && n.Kind != aig.ExclusiveOr {
				continue
			}
			a, b := child(n.Fanins[0]), child(n.Fanins[1])
			ca, cb := n.Fanins[0].Complemented(), n.Fanins[1].Complemented()
			for w := range info[id] {
				x, y := a[w], b[w]
				if ca {
					x = ^x
				}
				if cb {
					y = ^y
				}
				info[id][w] = x & y
			}
		}
		for w := range mask {
			mask[w], mask2[w] = 0, 0
		}
		for i, co := range outputs {
			out, in := info[co.NodeID()], child(aig.Aig_ObjChild0(src, co))
			compl := aig.Aig_ObjChild0(src, co).Complemented()
			constraint := i >= poCount-constrs && i < poCount
			for w := range out {
				v := in[w]
				if compl {
					v = ^v
				}
				if constraint {
					out[w] |= v
				} else {
					out[w] = v
				}
			}
			for w := range out {
				if i < poCount-constrs {
					mask[w] |= out[w]
				} else if i < poCount {
					mask2[w] |= out[w]
				}
			}
		}
		for i := 0; i < poCount; i++ {
			out := info[outputs[i].NodeID()]
			for w := range out {
				probs[i] += int32(bits.OnesCount32(out[w]))
			}
			if i < poCount-constrs {
				blocked := true
				for w := range out {
					if out[w]&^mask2[w] != 0 {
						blocked = false
						break
					}
				}
				if !blocked {
					fmt.Printf("Primary output %d fails on some input patterns.\n", i)
				}
			} else {
				for w := range out {
					probs2[i] += int32(bits.OnesCount32(out[w] & mask[w]))
				}
			}
		}
	}
	if verbose != 0 {
		util.Abc_PrintTime(1, "T", time.Since(start).Nanoseconds())
		total := 32 * words * frames
		for i := 0; i < poCount; i++ {
			if i < poCount-constrs {
				fmt.Printf("Primary output :  ")
			} else {
				fmt.Printf("Constraint %3d :  ", i-(poCount-constrs))
			}
			fmt.Printf("ProbOne = %f  ", float32(probs[i])/float32(total))
			fmt.Printf("ProbOneC = %f  ", float32(probs2[i])/float32(total))
			fmt.Printf("AllZeroValue = %d ", boolToInt(aig.Aig_ObjPhase(src, outputs[i])))
			fmt.Printf("\n")
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
