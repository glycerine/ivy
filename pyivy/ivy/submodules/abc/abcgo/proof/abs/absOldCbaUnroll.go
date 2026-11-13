package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManCbaUnrollWithCex also records each latch input's unrolled literal
// by frame. If synthesis changed the register count, initialization is zero.
func Saig_ManCbaUnrollWithCex(g *aig.Graph, cex *util.Abc_Cex_t, nInputs int32, mapping **vec.Vec_Int_t, regFrames **vec.Vec_Vec_t) *aig.Graph {
	return oldUnrollWithCex(g, cex, nInputs, mapping, regFrames, true)
}
