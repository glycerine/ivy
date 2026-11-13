package saig

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// TemporalBMC performs bounded checking and reports the last completed frame.
type TemporalBMC func(*aig.Graph, int32, int32, int32, int32, *int32) int32

var temporalBMC struct {
	sync.RWMutex
	run TemporalBMC
}

// RegisterTemporalBMC connects the bounded model checking implementation.
func RegisterTemporalBMC(run TemporalBMC) {
	temporalBMC.Lock()
	temporalBMC.run = run
	temporalBMC.Unlock()
}

func Vec_IntLastNonZeroBeforeLimit(values *vec.Vec_Int_t, limit int32) int32 {
	if values == nil || limit <= 0 {
		return -1
	}
	i := min(int(limit), vec.Vec_IntSize(values)) - 1
	for ; i >= 0; i-- {
		if vec.Vec_IntEntry(values, i) != 0 {
			return int32(i)
		}
	}
	return -1
}

func Saig_ManTempor(g *aig.Graph, frames, timeout, conflicts, useBMC, useTransient, verbose, veryVerbose int32) *aig.Graph {
	if g == nil || frames < 0 {
		return nil
	}
	var transient *vec.Vec_Int_t
	if frames == 0 {
		frames = Saig_ManPhasePrefixLength(g, verbose, veryVerbose, &transient)
		if frames <= 1 {
			fmt.Printf("The leading sequence has length %d. Temporal decomposition is not performed.\n", frames)
			return nil
		}
		if useTransient != 0 {
			last := Vec_IntLastNonZeroBeforeLimit(transient, int32(vec.Vec_IntSize(transient)))
			if last > 0 && last < frames {
				fmt.Printf("Reducing frame count from %d to %d to fit the last transient.\n", frames, last)
				frames = last
			}
		}
		fmt.Printf("Using computed frame number (%d).\n", frames)
	} else {
		fmt.Printf("Using user-given frame number (%d).\n", frames)
	}
	if useBMC != 0 {
		temporalBMC.RLock()
		run := temporalBMC.run
		temporalBMC.RUnlock()
		if run == nil {
			fmt.Println("Bounded model checking is unavailable; temporal decomposition is not performed.")
			return nil
		}
		finished := int32(-1)
		result := run(g, frames, timeout, conflicts, verbose, &finished)
		if result == 0 {
			fmt.Printf("A cex found in the first %d frames.\n", frames)
			return nil
		}
		if finished+1 < frames {
			last := Vec_IntLastNonZeroBeforeLimit(transient, finished)
			if last < 1 || useTransient == 0 {
				fmt.Printf("BMC for %d frames could not be completed. A cex may exist!\n", frames)
				return nil
			}
			fmt.Printf("BMC succeeded to frame %d. Adjusting frame count to be (%d) based on the last transient signal.\n", finished, last)
			frames = last
		}
	}
	return Saig_ManTemporDecompose(g, frames)
}
