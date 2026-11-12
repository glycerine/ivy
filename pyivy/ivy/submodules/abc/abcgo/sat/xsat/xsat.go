// Package xsat ports Bruno Schmitt's xSAT solver from ABC to Go-owned storage.
// Search uses xSAT's integer activities, binary watches and LBD heuristics.
package xsat

import "time"

const (
	Var0       = 1
	Var1       = 0
	VarX       = 3
	LBoolUndef = 0
	LBoolTrue  = 1
	LBoolFalse = -1
	VarUndef   = -1
	LitUndef   = -2
)
const CRefUndef uint32 = 0xffffffff

type Options struct {
	Verbose                                              bool
	ConflictLimit, PropagationLimit                      int64
	Deadline                                             time.Time
	K, R                                                 float64
	FirstBlockRestart, LBDWindow, TrailWindow            int
	FirstReduce, ReduceIncrement, SpecialReduceIncrement int
	FrozenLBD                                            uint32
}
type xSAT_SolverOptions_t = Options

var DefaultConfig = Options{K: 0.8, R: 1.4, FirstBlockRestart: 10000, LBDWindow: 50, TrailWindow: 5000, FirstReduce: 2000, ReduceIncrement: 300, SpecialReduceIncrement: 1000, FrozenLBD: 30}

type Stats struct {
	Starts, Reductions                                                              uint32
	Decisions, Propagations, Inspections, Conflicts, ClauseLiterals, LearntLiterals int64
}
type xSAT_Stats_t = Stats
