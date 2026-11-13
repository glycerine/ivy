package abs

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/pdr"
)

var absProver absProverControl

// PDR currently has global invariant state, so serialize abstraction solves.
var absPDRMutex sync.Mutex

func absRunPDR(id int32, graph *aig.Graph, verbose int32) int32 {
	if verbose != 0 {
		fmt.Printf("\nTrying to prove abstraction %d.\n", id)
	}
	absPDRMutex.Lock()
	defer absPDRMutex.Unlock()
	if absProver.stopped(id) {
		return -1
	}
	pars := pdr.DefaultOptions()
	pars.Silent, pars.RunID = true, id
	pars.Stop = absProver.stopped
	result, err := pdr.Solve(graph, &pars)
	if err != nil {
		panic(err)
	}
	if verbose != 0 {
		switch result.Status {
		case pdr.Proved:
			fmt.Printf("Proved abstraction %d.\n", id)
		case pdr.Failed:
			fmt.Printf("Disproved abstraction %d.\n", id)
		default:
			fmt.Printf("Cancelled abstraction %d.\n", id)
		}
	}
	return int32(result.Status)
}
