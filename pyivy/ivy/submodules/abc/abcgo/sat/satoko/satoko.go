// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
// Package satoko implements incremental conflict-driven Boolean SAT solving.
package satoko

const (
	SATOKO_ERR          = 0
	SATOKO_OK           = 1
	SATOKO_UNDEC        = 0
	SATOKO_SAT          = 1
	SATOKO_UNSAT        = -1
	SATOKO_LIT_FALSE    = 1
	SATOKO_LIT_TRUE     = 0
	SATOKO_VAR_UNASSING = 3
)

type Options struct {
	ConflictLimit, PropagationLimit                      int64
	VariableDecay                                        float64
	ClauseDecay                                          float32
	FirstBlockRestart                                    uint32
	RestartWindow, TrailWindow                           uint32
	RestartFactor, BlockFactor                           float64
	FirstReduce, ReduceIncrement, SpecialReduceIncrement uint32
	FreezeLBD                                            uint32
	LearntRatio                                          float32
	VariableRescale                                      uint32
	VariableLimit                                        uint64
	BinaryResolutionSize, BinaryResolutionLBD            uint32
	GarbageRatio                                         float32
	Verbose, NoSimplify                                  bool
}
type Stats struct {
	Starts, Reductions                                                             uint32
	Decisions, Propagations, AllPropagations, Inspections, Conflicts, AllConflicts int64
	OriginalLiterals, LearntLiterals                                               int64
}
type satoko_opts_t = Options
type satoko_stats_t = Stats
type satoko_t = Solver
type solver_t = Solver

func DefaultOptions() Options {
	return Options{ConflictLimit: 0, PropagationLimit: 0, VariableDecay: 0.95, ClauseDecay: 0.999, FirstBlockRestart: 10000, RestartWindow: 50, TrailWindow: 5000, RestartFactor: 0.8, BlockFactor: 1.4, FirstReduce: 2000, ReduceIncrement: 300, SpecialReduceIncrement: 1000, FreezeLBD: 30, LearntRatio: 0.5, VariableRescale: 200, VariableLimit: 0x014c924d692ca61b, BinaryResolutionSize: 30, BinaryResolutionLBD: 6, GarbageRatio: 0.3}
}
func New() *Solver {
	s := &Solver{opts: DefaultOptions(), ok: true, actInc: 1 << 47, clauseInc: 1 << 11}
	s.initOrder()
	return s
}
func satoko_create() *Solver                 { return New() }
func satoko_destroy(s *Solver)               { *s = Solver{} }
func (s *Solver) Destroy()                   { satoko_destroy(s) }
func satoko_default_opts(o *Options)         { *o = DefaultOptions() }
func satoko_configure(s *Solver, o *Options) { s.opts = *o }
func (s *Solver) Options() *Options          { return &s.opts }
func (s *Solver) Stats() *Stats              { return &s.stats }
