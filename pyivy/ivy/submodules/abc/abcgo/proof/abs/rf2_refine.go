package abs

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rf2_ManRefine retains the experimental C routine's nil result: it computes
// and prints candidate sets but does not choose a refinement.
func Rf2_ManRefine(p *Rf2_Man_t, cex *util.Abc_Cex_t, mapping *vec.Vec_Int_t, propFanout, verbose int32) *vec.Vec_Int_t {
	start := time.Now()
	p.nCalls++
	p.pCex, p.vMap, p.fPropFanout, p.fVerbose = cex, mapping, propFanout, verbose
	Rf2_ManCollect(p)
	Rf2_ManAssignJustIds(p)
	just := Rf2_ManPropagate(p, 32)
	Rf2_ManPrintVectorSpecial(p, just)
	if vec.Vec_IntSize(just) == 0 {
		fmt.Println("Empty set of justifying subsets.")
		return nil
	}
	clock := time.Now()
	p.timeVer += time.Since(clock)
	p.timeTotal += time.Since(start)
	return nil
}
