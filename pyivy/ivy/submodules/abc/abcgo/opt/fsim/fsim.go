// Package fsim provides frontier-based, word-parallel sequential AIG simulation.
package fsim

import (
	"github.com/glycerine/abcgo/misc/util"
	"time"
)

type Fsim_ParSim_t struct {
	Words, Iterations                int
	TimeLimit                        time.Duration
	CheckMiter, Verbose, CompressAIG bool
}
type Fsim_ParSwitch_t struct {
	Words, Iterations, Prefix, RandomPI            int
	ProbabilityOne, ProbabilityTransition, Verbose bool
}
type SimulationResult struct {
	Frames          int
	TimedOut, Found bool
	Counterexample  *util.Abc_Cex_t
}
type Switching struct {
	One, Transition             []float32
	OneFrames, TransitionFrames int
}
