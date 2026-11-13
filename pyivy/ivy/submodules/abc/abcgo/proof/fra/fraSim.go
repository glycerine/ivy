package fra

// Simulation entry points use the canonical shared FRAIG implementation.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/framgr"
)

func fraBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Fra_SmlNodeHash(s *Fra_Sml_t, obj aig.Edge, size int32) int32 {
	return framgr.Fra_SmlNodeHash(s, obj, size)
}
func Fra_SmlNodeIsConst(s *Fra_Sml_t, obj aig.Edge) int32 { return framgr.Fra_SmlNodeIsConst(s, obj) }
func Fra_SmlNodesAreEqual(s *Fra_Sml_t, a, b aig.Edge) int32 {
	return framgr.Fra_SmlNodesAreEqual(s, a, b)
}
func Fra_SmlNodeNotEquWeight(s *Fra_Sml_t, a, b int32) int32 {
	return framgr.Fra_SmlNodeNotEquWeight(s, a, b)
}
func Fra_SmlNodeIsZero(s *Fra_Sml_t, obj aig.Edge) int32 { return framgr.Fra_SmlNodeIsZero(s, obj) }
func Fra_SmlNodeCountOnes(s *Fra_Sml_t, obj aig.Edge) int32 {
	return framgr.Fra_SmlNodeCountOnes(s, obj)
}
func Fra_SmlSavePattern0(p *Fra_Man_t, init int32) { framgr.Fra_SmlSavePattern0(p, init) }
func Fra_SmlSavePattern1(p *Fra_Man_t, init int32) { framgr.Fra_SmlSavePattern1(p, init) }
func Fra_SmlSavePattern(p *Fra_Man_t)              { framgr.Fra_SmlSavePattern(p) }
func Fra_SmlCheckOutputSavePattern(p *Fra_Man_t, output aig.Edge) {
	framgr.Fra_SmlCheckOutputSavePattern(p, output)
}
func Fra_SmlCheckOutput(p *Fra_Man_t) int32          { return framgr.Fra_SmlCheckOutput(p) }
func Fra_SmlAssignRandom(s *Fra_Sml_t, obj aig.Edge) { framgr.Fra_SmlAssignRandom(s, obj) }
func Fra_SmlAssignConst(s *Fra_Sml_t, obj aig.Edge, value, frame int32) {
	framgr.Fra_SmlAssignConst(s, obj, value, frame)
}
func Fra_SmlInitialize(s *Fra_Sml_t, init int32)        { framgr.Fra_SmlInitialize(s, init) }
func Fra_SmlAssignDist1(s *Fra_Sml_t, pattern []uint32) { framgr.Fra_SmlAssignDist1(s, pattern) }
func Fra_SmlNodeSimulate(s *Fra_Sml_t, obj aig.Edge, frame int32) {
	framgr.Fra_SmlNodeSimulate(s, obj, frame)
}
func Fra_SmlNodesCompareInFrame(s *Fra_Sml_t, a, b aig.Edge, frameA, frameB int32) int32 {
	return framgr.Fra_SmlNodesCompareInFrame(s, a, b, frameA, frameB)
}
func Fra_SmlNodeCopyFanin(s *Fra_Sml_t, obj aig.Edge, frame int32) {
	framgr.Fra_SmlNodeCopyFanin(s, obj, frame)
}
func Fra_SmlNodeTransferNext(s *Fra_Sml_t, out, in aig.Edge, frame int32) {
	framgr.Fra_SmlNodeTransferNext(s, out, in, frame)
}
func Fra_SmlCheckNonConstOutputs(s *Fra_Sml_t) int32 { return framgr.Fra_SmlCheckNonConstOutputs(s) }
func Fra_SmlSimulateOne(s *Fra_Sml_t)                { framgr.Fra_SmlSimulateOne(s) }
func Fra_SmlResimulate(p *Fra_Man_t)                 { framgr.Fra_SmlResimulate(p) }
func Fra_SmlSimulate(p *Fra_Man_t, init int32)       { framgr.Fra_SmlSimulate(p, init) }
func Fra_SmlStart(g *aig.Graph, prefix, frames, words int32) *Fra_Sml_t {
	return framgr.Fra_SmlStart(g, prefix, frames, words)
}
func Fra_SmlStop(s *Fra_Sml_t) { framgr.Fra_SmlStop(s) }
func Fra_SmlSimulateComb(g *aig.Graph, words, check int32) *Fra_Sml_t {
	return framgr.Fra_SmlSimulateComb(g, words, check)
}
func Fra_SmlSimulateReadFile(name string) *vec.Vec_Str_t { return framgr.Fra_SmlSimulateReadFile(name) }
func Fra_SmlInitializeGiven(s *Fra_Sml_t, values *vec.Vec_Str_t) {
	framgr.Fra_SmlInitializeGiven(s, values)
}
func Fra_SmlPrintOutputs(s *Fra_Sml_t, patterns int32) { framgr.Fra_SmlPrintOutputs(s, patterns) }
func Fra_SmlSimulateCombGiven(g *aig.Graph, name string, check, verbose int32) *Fra_Sml_t {
	return framgr.Fra_SmlSimulateCombGiven(g, name, check, verbose)
}
func Fra_SmlSimulateSeq(g *aig.Graph, prefix, frames, words, check int32) *Fra_Sml_t {
	return framgr.Fra_SmlSimulateSeq(g, prefix, frames, words, check)
}
func Fra_SmlGetCounterExample(s *Fra_Sml_t) *util.Abc_Cex_t {
	return framgr.Fra_SmlGetCounterExample(s)
}
func Fra_SmlCopyCounterExample(g, frames *aig.Graph, model []int32) *util.Abc_Cex_t {
	return framgr.Fra_SmlCopyCounterExample(g, frames, model)
}
