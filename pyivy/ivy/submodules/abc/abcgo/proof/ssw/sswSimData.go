package ssw

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ssw_Sml_t_ owns object-major simulation words for all frames. Prefix words
// are retained for sequential initialization but excluded from comparisons.
type Ssw_Sml_t_ struct {
	pAig                                                 *aig.Graph
	nObjects                                             int
	nPref, nFrames, nWordsFrame, nWordsTotal, nWordsPref int32
	fNonConstOut, nSimRounds                             int32
	timeSim                                              time.Duration
	pData                                                []uint32
}

func Ssw_SmlStart(graph *aig.Graph, prefix, frames, wordsPerFrame int32) *Ssw_Sml_t {
	if graph == nil || prefix < 0 || frames < 0 || wordsPerFrame < 0 {
		panic("ssw: invalid simulation dimensions")
	}
	frameCount := int64(prefix) + int64(frames)
	total := frameCount * int64(wordsPerFrame)
	if frameCount > int64(^uint32(0)>>1) || total > int64(^uint32(0)>>1) {
		panic("ssw: simulation dimensions overflow")
	}
	nObjects := len(graph.Nodes())
	return &Ssw_Sml_t_{
		pAig: graph, nPref: prefix, nFrames: prefix + frames,
		nObjects:    nObjects,
		nWordsFrame: wordsPerFrame, nWordsTotal: int32(total),
		nWordsPref: prefix * wordsPerFrame,
		pData:      make([]uint32, nObjects*int(total)),
	}
}

// Ssw_ObjSim returns the live word slice belonging to an AIG object ID.
func Ssw_ObjSim(p *Ssw_Sml_t, id int32) []uint32 {
	if id < 0 || int(id) >= p.nObjects {
		panic("ssw: simulation object out of range")
	}
	start := int(id) * int(p.nWordsTotal)
	return p.pData[start : start+int(p.nWordsTotal)]
}

func Ssw_ObjRandomSim() uint32 { return aig.Aig_ManRandom(0) }

func Ssw_SmlObjHashWord(p *Ssw_Sml_t, obj aig.Edge) uint32 {
	var hash uint32
	words := Ssw_ObjSim(p, int32(obj.NodeID()))
	for i := p.nWordsPref; i < p.nWordsTotal; i++ {
		hash ^= words[i] * sswSimulationPrimes[i&127]
	}
	return hash
}
func Ssw_SmlObjIsConstWord(p *Ssw_Sml_t, obj aig.Edge) int32 {
	for _, value := range Ssw_ObjSim(p, int32(obj.NodeID()))[p.nWordsPref:] {
		if value != 0 {
			return 0
		}
	}
	return 1
}
func Ssw_SmlObjsAreEqualWord(p *Ssw_Sml_t, a, b aig.Edge) int32 {
	x, y := Ssw_ObjSim(p, int32(a.NodeID())), Ssw_ObjSim(p, int32(b.NodeID()))
	for i := p.nWordsPref; i < p.nWordsTotal; i++ {
		if x[i] != y[i] {
			return 0
		}
	}
	return 1
}
func Ssw_SmlClean(p *Ssw_Sml_t) { clear(p.pData) }
func Ssw_SmlStop(p *Ssw_Sml_t) {
	if p != nil {
		p.pData = nil
		p.pAig = nil
	}
}
func Ssw_SmlNumFrames(p *Ssw_Sml_t) int32     { return p.nFrames }
func Ssw_SmlNumWordsTotal(p *Ssw_Sml_t) int32 { return p.nWordsTotal }
func Ssw_SmlSimInfo(p *Ssw_Sml_t, obj aig.Edge) []uint32 {
	if obj.Complemented() {
		panic("ssw: simulation info requires regular edge")
	}
	return Ssw_ObjSim(p, int32(obj.NodeID()))
}
func Ssw_SmlSimDataPointers(p *Ssw_Sml_t) *vec.Vec_Ptr_t {
	result := vec.Vec_PtrStart(p.nObjects)
	for _, edge := range p.pAig.LiveEdges() {
		vec.Vec_PtrWriteEntry(result, edge.NodeID(), Ssw_ObjSim(p, int32(edge.NodeID())))
	}
	return result
}

var sswSimulationPrimes [128]uint32 = [128]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}
