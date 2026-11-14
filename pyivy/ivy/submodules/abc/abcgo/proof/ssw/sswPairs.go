package ssw

// Pair-guided sequential correspondence, ported from sswPairs.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Ssw_TransferSignalPairs reads copy metadata already installed on the source
// graphs. When using Go's miter constructor, prefer Ssw_TransferSignalPairsMapped
// with the maps returned by saig.Saig_ManCreateMiterWithCopies.
func Ssw_TransferSignalPairs(miter, first, second *aig.Graph, ids1, ids2 *vec.Vec_Int_t) *vec.Vec_Int_t {
	copies1 := make([]aig.Edge, aig.Aig_ManObjNumMax(first))
	copies2 := make([]aig.Edge, aig.Aig_ManObjNumMax(second))
	for i := range copies1 {
		copies1[i] = aig.Aig_ObjCopy(first, aig.Aig_ManObj(first, int32(i)))
	}
	for i := range copies2 {
		copies2[i] = aig.Aig_ObjCopy(second, aig.Aig_ManObj(second, int32(i)))
	}
	return Ssw_TransferSignalPairsMapped(ids1, ids2, copies1, copies2)
}
func Ssw_SignalCorrespondenceWithPairs(first, second *aig.Graph, ids1, ids2 *vec.Vec_Int_t, pars *Ssw_Pars_t) *aig.Graph {
	if pars == nil {
		pars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pars)
	}

	if vec.Vec_IntSize(ids1) != vec.Vec_IntSize(ids2) {
		panic("ssw: mismatched pair vectors")
	}
	miter, copies1, copies2 := saig.Saig_ManCreateMiterWithCopies(first, second, 0)
	aig.Aig_ManCleanup(miter)
	pairs := Ssw_TransferSignalPairsMapped(ids1, ids2, copies1, copies2)
	classes := Ssw_TransformPairsIntoTempClasses(pairs, aig.Aig_ManObjNumMax(miter))
	vec.Vec_IntFree(pairs)
	p := Ssw_ManCreate(miter, pars)
	p.ppClasses = Ssw_ClassesPreparePairs(miter, classes)
	p.pSml = Ssw_SmlStart(miter, 0, p.nFrames+pars.NFramesAddSim, 1)
	Ssw_ClassesSetData(p.ppClasses, p.pSml, func(data any, node aig.Edge) uint32 { return Ssw_SmlObjHashWord(data.(*Ssw_Sml_t), node) }, func(data any, node aig.Edge) int32 { return Ssw_SmlObjIsConstWord(data.(*Ssw_Sml_t), node) }, func(data any, a, b aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(data.(*Ssw_Sml_t), a, b) })
	result := Ssw_SignalCorrespondenceRefine(p)
	Ssw_FreeTempClasses(classes, aig.Aig_ManObjNumMax(miter))
	Ssw_ManStop(p)
	aig.Aig_ManStop(miter)
	return result
}
func Ssw_SignalCorrespondeceTestPairs(graph *aig.Graph) *aig.Graph {
	start := time.Now()
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.FVerbose = 1
	reduced := Ssw_SignalCorrespondence(graph, &pars)
	ids1, ids2 := vec.Vec_IntAlloc(int(aig.Aig_ManObjNumMax(graph))), vec.Vec_IntAlloc(int(aig.Aig_ManObjNumMax(graph)))
	live := make(map[int]bool)
	for _, node := range reduced.LiveEdges() {
		live[node.NodeID()] = true
	}
	for _, node := range graph.LiveEdges() {
		copy := aig.Aig_ObjCopy(graph, node)
		if copy == aig.InvalidEdge || !live[aig.Aig_Regular(copy).NodeID()] {
			continue
		}
		vec.Vec_IntPush(ids1, int32(node.NodeID()))
		vec.Vec_IntPush(ids2, int32(aig.Aig_Regular(copy).NodeID()))
	}
	fmt.Printf("Recorded %d pairs (before: %d after: %d).\n", vec.Vec_IntSize(ids1), aig.Aig_ManObjNumMax(graph), aig.Aig_ManObjNumMax(reduced))
	result := Ssw_SignalCorrespondenceWithPairs(graph, reduced, ids1, ids2, &pars)
	vec.Vec_IntFree(ids1)
	vec.Vec_IntFree(ids2)
	sswReportPairResult(result, aig.Aig_ManRegNum(graph)+aig.Aig_ManRegNum(reduced), start)
	aig.Aig_ManStop(reduced)
	return result
}
func sswReportPairResult(graph *aig.Graph, total int32, start time.Time) int32 {
	result := Ssw_MiterStatus(graph, 1)
	switch result {
	case 1:
		fmt.Print("Verification successful.  ")
	case 0:
		fmt.Print("Verification failed with a counter-example.  ")
	default:
		fmt.Printf("Verification UNDECIDED. The number of remaining regs = %d (total = %d).  ", aig.Aig_ManRegNum(graph), total)
	}
	fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
	return result
}
func Ssw_SecWithPairs(first, second *aig.Graph, ids1, ids2 *vec.Vec_Int_t, pars *Ssw_Pars_t) int32 {
	if ids1 == nil || ids2 == nil {
		panic("ssw: missing pair vectors")
	}
	start := time.Now()
	fmt.Println("Performing specialized verification with node pairs.")
	result := Ssw_SignalCorrespondenceWithPairs(first, second, ids1, ids2, pars)
	if result == nil {
		return -1
	}
	defer aig.Aig_ManStop(result)
	return sswReportPairResult(result, aig.Aig_ManRegNum(first)+aig.Aig_ManRegNum(second), start)
}
func Ssw_SecGeneral(first, second *aig.Graph, pars *Ssw_Pars_t) int32 {
	start := time.Now()
	fmt.Println("Performing general verification without node pairs.")
	miter := saig.Saig_ManCreateMiter(first, second, 0)
	aig.Aig_ManCleanup(miter)
	result := Ssw_SignalCorrespondence(miter, pars)
	aig.Aig_ManStop(miter)
	if result == nil {
		return -1
	}
	defer aig.Aig_ManStop(result)
	return sswReportPairResult(result, aig.Aig_ManRegNum(first)+aig.Aig_ManRegNum(second), start)
}
func Ssw_SecGeneralMiter(miter *aig.Graph, pars *Ssw_Pars_t) int32 {
	start := time.Now()
	result := Ssw_SignalCorrespondence(miter, pars)
	if result == nil {
		return -1
	}
	defer aig.Aig_ManStop(result)
	return sswReportPairResult(result, aig.Aig_ManRegNum(miter), start)
}
