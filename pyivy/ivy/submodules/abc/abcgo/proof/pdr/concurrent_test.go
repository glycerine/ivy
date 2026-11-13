package pdr

import (
	"fmt"
	"sync"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Independently owned graphs and parameters must not race when PDR publishes
// successful invariants through ABC's shared global frame.
func TestConcurrentDirectSolves(t *testing.T) {
	const workers = 8
	var wg sync.WaitGroup
	errors := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				g := aig.NewGraph()
				g.NewInput()
				latch := g.NewInput()
				g.AddOutput(latch)
				g.AddOutput(g.Const0())
				aig.Aig_ManSetRegNum(g, 1)
				var pars Pdr_Par_t
				Pdr_ManSetDefaultParams(&pars)
				pars.fSilent = 1
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
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
