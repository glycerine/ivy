package wlc

import "sync/atomic"

// g_nRunIds is the generation of the most recently started prover. The C
// implementation uses a volatile int; atomic access also makes the callback
// safe when a solver polls it from another goroutine.
var g_nRunIds int32

func Wla_CallBackToStop(runID int32) int32 {
	current := atomic.LoadInt32(&g_nRunIds)
	assert(runID <= current, "RunId <= g_nRunIds")
	if runID < current {
		return 1
	}
	return 0
}

func Wla_GetGlobalRunId() int32 { return atomic.LoadInt32(&g_nRunIds) }

// The C implementation provides no-op functions when ABC_USE_PTHREADS is
// disabled. The Go port currently uses that same non-threaded configuration.
func Wla_ManJoinThread(_ *Wla_Man_t, _ int32) {}

// The C implementation leaves concurrent BMC unavailable without pthreads.
// any keeps this stub independent of the C-only AIG and counterexample types.
func Wla_ManConcurrentBmc3(_ *Wla_Man_t, _ any, _ any) {}
