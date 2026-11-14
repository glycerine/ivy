package pdr

import (
	"fmt"
	"sync"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

// Independently owned graphs and parameters must not race when PDR publishes
// successful invariants through ABC's shared global frame.
func TestConcurrentDirectSolves(t *testing.T) {
	const workers = 8
	var wg sync.WaitGroup
	errors := make(chan error, workers+1)
	done := make(chan struct{})
	var reader sync.WaitGroup
	reader.Add(1)
	go func() {
		defer reader.Done()
		defer func() {
			if p := recover(); p != nil {
				errors <- fmt.Errorf("invalid invariant snapshot: %v", p)
			}
		}()
		for {
			select {
			case <-done:
				return
			default:
			}
			frame := abcmain.Abc_FrameReadGlobalFrame()
			if frame == nil {
				continue
			}
			snapshot := frame.IntValues("pAbcWlcInv")
			if len(snapshot) == 0 {
				continue
			}
			packed := vec.Vec_IntAlloc(len(snapshot))
			for _, value := range snapshot {
				vec.Vec_IntPush(packed, value)
			}
			pdrInvCubes(packed)
			snapshot[0] = -100
		}
	}()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				g := aig.NewGraph()
				g.NewInput()
				regs := 1 + worker%4
				var latch aig.Edge
				for r := 0; r < regs; r++ {
					latch = g.NewInput()
				}
				g.AddOutput(latch)
				for r := 0; r < regs; r++ {
					g.AddOutput(g.Const0())
				}
				aig.Aig_ManSetRegNum(g, int32(regs))
				var pars Pdr_Par_t
				Pdr_ManSetDefaultParams(&pars)
				pars.FSilent = 1
				got := int32(-1)
				if worker%2 == 0 {
					got = Pdr_ManSolve(g, &pars)
				} else {
					got = IPdr_ManSolve(g, &pars)
				}
				if got != 1 {
					errors <- fmt.Errorf("got %d, want proved", got)
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(done)
	reader.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
