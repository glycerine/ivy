package pdr

import (
	"fmt"
	"math"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Status describes a property's verification result.
type Status int8

const (
	// Unknown means no conclusion was reached, including cancellation or a limit.
	Unknown Status = -1
	// Failed means the property can be violated by a reachable state.
	Failed Status = 0
	// Proved means the property holds for all reachable states.
	Proved Status = 1
)

// Options controls a PDR run. Use DefaultOptions to obtain ABC's defaults.
// A nil Options argument to Solve also selects those defaults. Zero-valued
// limits disable the corresponding limit. Time limits are rounded up to the
// solver's resolution (one second, or one millisecond for PerOutputTimeout).
type Options struct {
	MaxFrames            int32
	ConflictLimit        int32
	RestartLimit         int32
	RecycleLimit         int32
	Timeout              time.Duration
	GapTimeout           time.Duration
	PerOutputTimeout     time.Duration
	SolveAll             bool
	StoreCounterexamples bool
	UseAbstraction       bool
	SimpleRefinement     bool
	UseDown              bool
	UseCTGs              bool
	TwoRounds            bool
	FlopOrder            bool
	Shortest             bool
	Verbose              bool
	Silent               bool
	// RunID is passed to Stop.
	RunID int32
	// Stop is called during solving; returning true cancels the run.
	Stop func(runID int32) bool
	// OnFailure is called in SolveAll mode. Returning true cancels the run.
	// The witness is nil when StoreCounterexamples is false.
	OnFailure func(output int32, witness *util.Abc_Cex_t) bool
}

// DefaultOptions returns the defaults used by Pdr_ManSetDefaultParams.
func DefaultOptions() Options { return Options{MaxFrames: 10000, RecycleLimit: 300} }

// OutputResult describes one output. A Failed output can have a nil witness
// when witness storage was not requested. Unknown is distinct from Failed.
type OutputResult struct {
	Status  Status
	Witness *util.Abc_Cex_t
}

// Result reports the overall result and, in SolveAll mode, each output's
// result. In the default mode, Outputs identifies the failing output when a
// witness exists; other outputs remain Unknown unless all were proved.
// Witnesses are borrowed from the graph and should be treated as read-only.
type Result struct {
	Status                                 Status
	Outputs                                []OutputResult
	Frame                                  int32
	FailedCount, DroppedCount, ProvedCount int32
}

func pdrDuration(d, unit time.Duration) (int32, error) {
	if d < 0 {
		return 0, fmt.Errorf("pdr: negative time limit")
	}
	if d == 0 {
		return 0, nil
	}
	n := int64(d / unit)
	if d%unit != 0 {
		n++
	}
	if n > math.MaxInt32 {
		return 0, fmt.Errorf("pdr: time limit too large")
	}
	return int32(n), nil
}
func pdrFlag(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

// Solve checks graph's sequential outputs using PDR. A result of Failed
// includes a counterexample unless SolveAll is set without witness storage.
// The graph's sequential counterexample storage is replaced by the run.
// Successful runs publish their invariant to ABC's shared global frame.
// Concurrent publications are synchronized but last-writer-wins; callers
// needing to associate an invariant with a run must serialize those runs
// and read the invariant before starting another publisher. General concurrent
// solver safety beyond independently owned inputs is not established.
func Solve(graph *aig.Graph, opts *Options) (Result, error) {
	if graph == nil {
		return Result{}, fmt.Errorf("pdr: nil graph")
	}
	options := DefaultOptions()
	if opts != nil {
		options = *opts
	}
	if options.MaxFrames < 0 || options.ConflictLimit < 0 || options.RestartLimit < 0 || options.RecycleLimit < 0 {
		return Result{}, fmt.Errorf("pdr: negative limit")
	}
	timeout, err := pdrDuration(options.Timeout, time.Second)
	if err != nil {
		return Result{}, err
	}
	gap, err := pdrDuration(options.GapTimeout, time.Second)
	if err != nil {
		return Result{}, err
	}
	each, err := pdrDuration(options.PerOutputTimeout, time.Millisecond)
	if err != nil {
		return Result{}, err
	}
	if each != 0 && !options.SolveAll {
		return Result{}, fmt.Errorf("pdr: per-output timeout requires SolveAll")
	}
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.nFrameMax = options.MaxFrames
	pars.nConfLimit = options.ConflictLimit
	pars.nRestLimit = options.RestartLimit
	pars.nRecycle = options.RecycleLimit
	if pars.nRecycle == 0 {
		pars.nRecycle = math.MaxInt32
	}
	pars.nTimeOut = timeout
	pars.nTimeOutGap = gap
	pars.nTimeOutOne = each
	if each != 0 && timeout == 0 {
		outputs := int64(len(graph.Outputs())) - int64(aig.Aig_ManRegNum(graph))
		total := (int64(each)*outputs + 999) / 1000
		if total > math.MaxInt32 {
			return Result{}, fmt.Errorf("pdr: combined output timeout too large")
		}
		pars.nTimeOut = int32(total)
	}
	pars.fSolveAll = pdrFlag(options.SolveAll)
	pars.fStoreCex = pdrFlag(options.StoreCounterexamples)
	pars.fUseAbs = pdrFlag(options.UseAbstraction)
	pars.fUseSimpleRef = pdrFlag(options.SimpleRefinement)
	pars.fSkipDown = pdrFlag(!options.UseDown)
	pars.fCtgs = pdrFlag(options.UseCTGs)
	pars.fTwoRounds = pdrFlag(options.TwoRounds)
	pars.fFlopOrder = pdrFlag(options.FlopOrder)
	pars.fShortest = pdrFlag(options.Shortest)
	pars.fVerbose = pdrFlag(options.Verbose)
	pars.fSilent = pdrFlag(options.Silent)
	pars.fNotVerbose = pdrFlag(options.Silent)
	pars.RunId = options.RunID
	if options.Stop != nil {
		pars.pFuncStop = func(id int32) int32 { return pdrFlag(options.Stop(id)) }
	}
	if options.OnFailure != nil {
		pars.pFuncOnFail = func(i int32, c *util.Abc_Cex_t) int32 { return pdrFlag(options.OnFailure(i, c)) }
	}
	graph.SetSequentialCounterexamples(nil)
	status := Status(Pdr_ManSolve(graph, &pars))
	n := len(graph.Outputs()) - int(aig.Aig_ManRegNum(graph))
	results := make([]OutputResult, n)
	for i := range results {
		results[i].Status = Unknown
	}
	if options.SolveAll {
		models := graph.SequentialCounterexamples()
		for i := range results {
			results[i].Status = Status(vec.Vec_IntEntry(pars.vOutMap, i))
			if models != nil {
				results[i].Witness, _ = vec.Vec_PtrEntry(models, i).(*util.Abc_Cex_t)
			}
		}
	} else if status == Proved {
		for i := range results {
			results[i].Status = Proved
		}
	} else if status == Failed {
		if cex := graph.SequentialCounterexample(); cex != nil && cex.IPo >= 0 && int(cex.IPo) < len(results) {
			results[cex.IPo] = OutputResult{Status: Failed, Witness: cex}
		}
	}
	return Result{Status: status, Outputs: results, Frame: pars.iFrame, FailedCount: pars.nFailOuts, DroppedCount: pars.nDropOuts, ProvedCount: pars.nProveOuts}, nil
}
