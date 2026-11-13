package fra

// Class APIs use the canonical shared FRAIG implementation.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/framgr"
)

func Fra_ClassesStart(graph *aig.Graph) *Fra_Cla_t { return framgr.Fra_ClassesStart(graph) }
func Fra_ClassesStop(p *Fra_Cla_t)                 { framgr.Fra_ClassesStop(p) }
func Fra_ClassesCopyReprs(p *Fra_Cla_t, failed *vec.Vec_Ptr_t) {
	framgr.Fra_ClassesCopyReprs(p, failed)
}
func Fra_ClassCount(class []int) int32                 { return framgr.Fra_ClassCount(class) }
func Fra_ClassesCountLits(p *Fra_Cla_t) int32          { return framgr.Fra_ClassesCountLits(p) }
func Fra_ClassesCountPairs(p *Fra_Cla_t) int32         { return framgr.Fra_ClassesCountPairs(p) }
func Fra_ClassesPrint(p *Fra_Cla_t, veryVerbose int32) { framgr.Fra_ClassesPrint(p, veryVerbose) }
func Fra_ClassesPrepare(p *Fra_Cla_t, latchCorr, maxLevels int32) {
	framgr.Fra_ClassesPrepare(p, latchCorr, maxLevels)
}
func Fra_ClassesRefine(p *Fra_Cla_t) int32 { return framgr.Fra_ClassesRefine(p) }
func Fra_ClassesRefine1(p *Fra_Cla_t, refineNew int32, skipped *int32) int32 {
	return framgr.Fra_ClassesRefine1(p, refineNew, skipped)
}
func Fra_ClassesTest(p *Fra_Cla_t, id1, id2 int32) { framgr.Fra_ClassesTest(p, id1, id2) }
func Fra_ClassesLatchCorr(p *Fra_Man_t)            { framgr.Fra_ClassesLatchCorr(p) }
func Fra_ClassesPostprocess(p *Fra_Cla_t)          { framgr.Fra_ClassesPostprocess(p) }
func Fra_ClassesSelectRepr(p *Fra_Cla_t)           { framgr.Fra_ClassesSelectRepr(p) }
func Fra_ClassesDeriveAig(p *Fra_Cla_t, frames int32) *aig.Graph {
	return framgr.Fra_ClassesDeriveAig(p, frames)
}
